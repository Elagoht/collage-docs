---
description: Gerçek uygulamayı app.Handler() ve net/http/httptest üzerinden test edin — sayfalar, sahtecilik token'larıyla formlar, document'lar ve statik dışa aktarma.
---

# Test

Bir collage uygulaması bir `http.Handler`'dır. `app.Handler()` onu döndürür; dinleyen
bir sunucu, seçilecek bir port yoktur. Böylece `net/http/httptest`, sitenin tamamını
sıradan bir Go testinden sürer: route'lar, data handler'lar, şablonlar, önbellek,
formlar, document'lar ve middleware, tam olarak üretim ortamında çalıştıkları gibi.

`collage new`'in iskeletini kurduğu proje, bu şekilde yazılmış bir test dosyasıyla
gelir. Bu sayfa o dosyanın kalıbını adım adım anlatır.

## `main`'in kurduğu uygulamayı test edin

İskeletteki `main.go`, uygulamayı kendine ait bir fonksiyonda kurar:

```go
func newApp(devMode bool, port int) (*collage.App, error)
```

`main` onu sunmak için, testler de test etmek için çağırır. Test dosyasındaki en
önemli karar budur: testlerin sınadığı şey, gerçek yapılandırması, route'ları ve
mount'larıyla gerçekten çalışan sitedir; ondan yavaş yavaş uzaklaşan ikinci bir
kurulum değil.

```go
func handler(t *testing.T) http.Handler {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v, want nil", err)
	}
	return app.Handler()
}

// get returns the response to a GET of target, and its body.
func get(t *testing.T, h http.Handler, target string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec, rec.Body.String()
}
```

`newApp(false, 0)` üretim yapılandırmasını kurar: geliştirme modu kapalı ve port
`0`; collage bunu kendi varsayılanıyla değiştirir — zaten hiçbir şey dinlemez.

Bundan sonra bir test, bir istek ve yanıta bir bakıştan ibarettir:

```go
func TestPagesRender(t *testing.T) {
	h := handler(t)
	for target, want := range map[string]string{
		"/":         "Explore the features",
		"/features": "Four live demos",
		"/hello":    "Hello, stranger!",
	} {
		rec, body := get(t, h, target)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, rec.Code)
			continue
		}
		if !strings.Contains(body, want) {
			t.Errorf("GET %s does not contain %q", target, want)
		}
		// Counted, not merely found: a layout writing one and a page hoisting
		// another is two titles, which is a page that looks fine and is not.
		if n := strings.Count(body, "<title>"); n != 1 {
			t.Errorf("GET %s has %d titles, want exactly 1", target, n)
		}
	}
}
```

`handler(t)`'nin yaptığı gibi her test için yeni bir handler kurun. `app.Handler()`'a
yapılan ilk çağrı uygulamayı başlatır ve kaydı kapatır; böylece her test boş bir
önbellekle başlar.

Başlatma başarısız olursa — bir plugin'in `Init`'i hata verirse, bir sayfa kayıtlı
olmayan bir hata sayfasını adlandırırsa, bir mount bir route'u gölgelerse —
`app.Handler()` her isteğe 503 ile yanıt verir ve nedenini loglar. Bu başarısızlığı
testte bir hata olarak almak istediğinizde önce `app.Start()`'ı çağırın:

```go
if err := app.Start(); err != nil {
	t.Fatalf("start: %v", err)
}
```

## Disk önbelleğini yalıtın

İskeletin önbellek dizini bir paket değişkenidir ve `handler`, uygulamayı kurmadan
önce onu yeni bir dizine yönlendirir:

```go
// cacheDir is where rendered pages are kept between restarts. A variable so the
// tests can point it at a directory of their own.
var cacheDir = ".cache"
```

Geliştirme modu kapalıyken uygulama disk önbelleğini kullanır ve bir disk önbelleği,
aynı dizini ve aynı build'i kullanan her şey tarafından paylaşılır. `cacheDir =
t.TempDir()` olmadan bir teste, daha önceki bir testin — ya da önbellek süreçten
sağ çıktığı için daha önceki bir çalıştırmanın — render ettiği bir sayfa sunulabilir
ve test, kodla hiçbir ilgisi olmayan nedenlerle geçebilir ya da kalabilir. Ayrıca
paketinizde bir `.cache` dizini bırakırdı.

`cacheDir` tüm paket tarafından paylaşıldığı için bu testler `t.Parallel()`
çağırmaz. Paralel testler istiyorsanız dizini bunun yerine `newApp`'e bir parametre
olarak verin.

Önbellekle ilgili bir test bundan yararlanabilir: aynı handler'a iki istek yapın ve
ikincisinin önbellekten sunulduğunu doğrulayın ya da aralarında
`app.InvalidateTags`'i çağırın ve sunulmadığını doğrulayın.

## Formlar ve sahtecilik token'ı

Güvenli olmayan bir metodun arkasındaki her action — bir form gönderimi, bir
`fetch()` — bir istek sahteciliği token'ı denetler ve bir testin de bunu bir
tarayıcının yaptığı gibi göndermesi gerekir:

1. Formun bulunduğu sayfaya `GET` yapın. Yanıt `collage_csrf` çerezini ayarlar ve
   sayfadaki `{{csrfToken}}`, aynı token'ı taşıyan gizli bir `_csrf` alanı render
   eder.
2. Token'ı sayfadan okuyun.
3. Formu, token `_csrf` alanında — ya da bir `fetch()`'in yaptığı gibi
   `X-CSRF-Token` header'ında — ve 1. adımdaki çerezle birlikte `POST` edin.

İskeletin yardımcıları:

```go
// token reads the forgery token out of a rendered page, the way a browser does.
func token(t *testing.T, body string) string {
	t.Helper()
	m := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no forgery token in the page:\n%s", body)
	}
	return m[1]
}

// post sends a POST to target, carrying the cookies from page.
func post(t *testing.T, h http.Handler, target string, form url.Values, header http.Header, page *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for name, values := range header {
		req.Header[name] = values
	}
	if page != nil {
		for _, c := range (&http.Response{Header: page.Header()}).Cookies() {
			req.AddCookie(c)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
```

`(&http.Response{Header: page.Header()}).Cookies()`, recorder'ın `Set-Cookie`
header'larını standart kütüphanenin kendi çerez ayrıştırıcısıyla ayrıştırır; böylece
test tam olarak bir tarayıcının geri göndereceği şeyi gönderir.

Token'ı alanda taşıyan bir form gönderimi:

```go
func TestHelloGreetsTheSubmittedName(t *testing.T) {
	h := handler(t)
	page, body := get(t, h, "/features")

	rec := post(t, h, "/hello", url.Values{"_csrf": {token(t, body)}, "name": {"Ada"}}, nil, page)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /hello = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello, Ada!") {
		t.Errorf("the response does not greet the name:\n%s", rec.Body.String())
	}
}
```

Token'ı header'da taşıyan bir JSON uç noktası:

```go
func TestCountAnswersWithTheNewCount(t *testing.T) {
	h := handler(t)
	page, body := get(t, h, "/features")

	rec := post(t, h, "/api/count", nil, http.Header{"X-Csrf-Token": {token(t, body)}}, page)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/count = %d, want 200", rec.Code)
	}
	var answer struct{ Count int64 }
	if err := json.Unmarshal(rec.Body.Bytes(), &answer); err != nil {
		t.Fatalf("body = %q, want JSON: %v", rec.Body.String(), err)
	}
	if answer.Count < 1 {
		t.Errorf("count = %d, want at least 1 after a click", answer.Count)
	}
}
```

Ve korumayı açık tutan test:

```go
// Without a token a submission never reaches its handler. This is the test that
// fails if the protection is ever turned off by accident.
func TestASubmissionWithNoTokenIsRefused(t *testing.T) {
	h := handler(t)
	for _, target := range []string{"/api/count", "/hello"} {
		if rec := post(t, h, target, url.Values{"name": {"Ada"}}, nil, nil); rec.Code != http.StatusForbidden {
			t.Errorf("POST %s with no token = %d, want 403", target, rec.Code)
		}
	}
}
```

Token `Security.CSRFKey` ile, anahtar ayarlanmamışsa uygulama başına üretilen bir
anahtarla imzalanır. Her iki durumda da bir testteki sayfa ve gönderim aynı
uygulamaya gider; bu yüzden testlerin kendilerine ait bir anahtara ihtiyacı yoktur.
`_csrf` ve `collage_csrf` varsayılan adlardır; bunları `Security.CSRFFieldName` ya da
`CSRFCookieName` ile yeniden adlandıran bir uygulama, testlerinde de yeniden
adlandırır.

## Document'lar, yönlendirmeler ve durum kodları

Bir document, bir sayfa gibi test edilir. Gövdenin yanında içerik tipini de
denetleyin; çünkü içerik tipi, bir document'ın ne olduğunun yarısıdır:

```go
func TestHealthCheck(t *testing.T) {
	rec, body := get(t, handler(t), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz = %d, want 200", rec.Code)
	}
	var health struct{ Status string }
	if err := json.Unmarshal([]byte(body), &health); err != nil || health.Status != "ok" {
		t.Errorf("body = %q, want a status of ok", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
```

Bulunamadı sayfası, `Allow` header'ıyla bir 405, bir yönlendirmenin `Location`'ı,
bir `Cache-Control` header'ı — her biri recorder üzerindeki bir alandır:

```go
func TestNotFoundPage(t *testing.T) {
	rec, body := get(t, handler(t), "/there-is-nothing-here")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if !strings.Contains(body, "There is nothing at this address") {
		t.Errorf("body = %q, want this site's own not-found page", body)
	}
}
```

`app.Use` ile kaydedilen middleware da handler'ın parçası olduğu için bu testlerde
çalışır. Bir önizlemeyi ya da bir `collage.Vary` boyutunu test etmek için
`ServeHTTP`'den önce istekteki çerezi ya da header'ı ayarlayın.

## Her sayfa render edilir

Her sayfayı ziyaret eden bir test, yalnızca bunlardan birinde başarısız olan şablonu
yakalar. collage-docs içeriğini yükler ve her sayfayı ister — burada yukarıdaki
`handler` ve `get` yardımcılarıyla, tüm tur için tek bir uygulama:

```go
// Every page of the documentation renders, with its own title.
func TestEveryDocRenders(t *testing.T) {
	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatalf("site.Load: %v", err)
	}
	h := handler(t)
	for _, page := range loaded.Pages() {
		rec, body := get(t, h, page.URL())
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), rec.Code)
			continue
		}
		if !strings.Contains(body, "<title>"+page.Title+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
	}
}
```

Sayfaları bir CMS'ten gelen bir site için aynı test,
[path provider](/docs/static-export#dynamic-paths-pathprovider)'ınızın döndürdüğü
listeyi dolaşabilir.

## Dışa aktarmayı test etmek

Statik dışa aktarma da koddur ve atlanan bir sayfanın üretim ortamında sessiz kaldığı
tek yerdir. Onu `t.TempDir()` içine çalıştırın ve beklediğiniz dosyaların orada
olduğunu denetleyin:

```go
// The export writes the home page, every doc and the 404 page.
func TestExport(t *testing.T) {
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := staticBuild(app, out, false); err != nil {
		t.Fatalf("staticBuild: %v", err)
	}
	loaded, _ := site.Load(content.FS)
	want := []string{"index.html", "404.html"}
	for _, page := range loaded.Pages() {
		want = append(want, filepath.Join("docs", page.Slug, "index.html"))
	}
	for _, file := range want {
		if _, err := os.Stat(filepath.Join(out, file)); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}
```

`collage export`'un çalıştırdığı `staticBuild`'in aynısını çağırır; böylece test,
sayfaların yanında path provider'ı ve builder seçeneklerini de kapsar.
`staticBuild`, build'in hatasını döndürür; bu da bozulmuş (degraded) bir sayfa, boş
bir render ya da bir panic için testi başarısız kılar.

Atlamalar ve uyarılar üzerinde de doğrulama yapmak için builder'ı doğrudan çağırın ve
raporu okuyun:

```go
builder, err := collage.NewBuilder(app, collage.BuildOptions{OutDir: t.TempDir()})
if err != nil {
	t.Fatal(err)
}
report, err := builder.Build(context.Background())
if err != nil {
	t.Fatalf("build: %v", err)
}
for _, skip := range report.Skipped {
	if skip.Page == "home" {
		t.Errorf("home was skipped: %s", skip.Reason)
	}
	if errors.Is(skip.Err, collage.ErrDynamicPathUnresolved) {
		t.Errorf("%s has a {param} the path provider does not cover", skip.Page)
	}
}
if len(report.Warnings) > 0 {
	t.Errorf("warnings: %v", report.Warnings)
}
```

v0.10.0'dan itibaren `skip.Err`, her atlamanın arkasındaki sentinel hatadır; bkz.
[Hatalar](/docs/errors#static-builds).

## HTTP olmadan render etmek

`app.RenderPath`, tek bir sayfayı dışa aktarmanın yaptığı gibi render eder — istek
yok, sayfa önbelleği yok, middleware yok — ve HTML'i ve render'ın ne yaptığını
döndürür. Plugin'lerinizin render hook'ları yine çalışır, veri önbelleği de öyle:
[`collage.Cached`](/docs/caching#caching-data-across-pages)'ın bir render ya da
sunulan bir istek sırasında sakladığı değer, aynı uygulamadaki bir sonraki
`RenderPath`'e verilen değerdir. Bir test yeni veriye ihtiyaç duyduğunda yeni bir
uygulama kurun:

```go
result, err := app.RenderPath(context.Background(), "/blog/hello-world", "", nil)
if err != nil {
	t.Fatalf("render: %v", err)
}
if result.Degraded() {
	t.Errorf("a fragment failed: %+v", result.Metadata)
}
```

`app.RenderDocumentPath` aynısını bir document için yapar. Çoğu test için
`app.Handler()` daha uygundur, çünkü bir okuyucunun ne aldığını test eder; bunlar
tek bir render'a ayrıntılı bakmak içindir.

## Testleri çalıştırmak

```sh
go test ./...
```

Bunu CI'da, build almadan ya da dışa aktarmadan önce çalıştırın. Gerçek `newApp`'i
kullanan bir test, sitenin başlamasını engelleyecek her şeyde orada başarısız olur;
bu, sunucuda öğrenmekten daha ucuzdur.
