---
description: Gerçek uygulamayı app.Handler() ve net/http/httptest ile test edin. Page'leri, forgery token'ıyla birlikte form'ları, document'ları ve static export'u bu şekilde sınayabilirsiniz.
reference: NewBuilder, BuildOptions, App
---

# Test yazmak

Bir collage uygulaması bir `http.Handler`'dır. `app.Handler()` bu handler'ı döndürür.
Dinleyen bir sunucu yoktur, seçmeniz gereken bir port da yoktur. Bu yüzden
`net/http/httptest`, sıradan bir Go testinden bütün siteyi çalıştırabilir. Routing,
data handler'lar, template'ler, cache, form'lar, document'lar ve middleware testte de
production'daki gibi çalışır.

`collage new` komutunun scaffold ettiği proje, bu şekilde kurulmuş bir test dosyasıyla
gelir. Bu sayfa o dosyadaki kalıbı adım adım anlatır.

## `main`'in kurduğu uygulamayı test edin

Scaffold edilen `main.go`, uygulamayı ayrı bir fonksiyonda kurar:

```go
func newApp(devMode bool, port int) (*collage.App, error)
```

`main` bu fonksiyonu siteyi serve etmek için, testler ise test etmek için çağırır.
Test dosyasındaki en önemli karar budur. Testler, gerçekten çalışan siteyi gerçek
config'i, route'ları ve mount'larıyla birlikte sınar. Zamanla ondan uzaklaşan ikinci
bir kurulumu sınamaz.

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

`newApp(false, 0)` production config'ini kurar. Development mode kapalıdır ve port
`0`'dır. collage bu `0` değerini kendi varsayılan portuyla değiştirir. Zaten hiçbir
şey bu portu dinlemez.

Bundan sonra her test, bir request atmaktan ve response'a bakmaktan ibarettir:

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
yapılan ilk çağrı uygulamayı başlatır ve register aşamasını kapatır. Handler her test
için yeniden kurulduğundan her test boş bir cache ile başlar.

Başlatma başarısız olabilir. Örneğin bir plugin'in `Init`'i hata döner, bir page
register edilmemiş bir error page'i belirtir ya da bir mount bir route'u gölgeler. Bu
durumda `app.Handler()` her request'e 503 ile cevap verir ve nedenini log'a yazar. Bu
hatayı testte bir hata olarak almak istiyorsanız önce `app.Start()`'ı çağırın:

```go
if err := app.Start(); err != nil {
	t.Fatalf("start: %v", err)
}
```

## Disk cache'ini izole edin

Scaffold'daki cache dizini bir package değişkenidir. `handler`, uygulamayı kurmadan
önce bu değişkeni yeni bir dizine yönlendirir:

```go
// cacheDir is where rendered pages are kept between restarts. A variable so the
// tests can point it at a directory of their own.
var cacheDir = ".cache"
```

Development mode kapalıyken uygulama disk cache'ini kullanır. Aynı dizini ve aynı
build'i kullanan her şey bu disk cache'ini paylaşır. `cacheDir = t.TempDir()` satırı
olmasaydı, bir teste önceki bir testin render ettiği bir page sunulabilirdi. Cache
process bittikten sonra da kaldığı için bu page önceki bir çalıştırmadan da
gelebilirdi. Test de kodla hiçbir ilgisi olmayan nedenlerle geçer ya da kalırdı.
Ayrıca package'ınızın içinde bir `.cache` dizini bırakırdı.

`cacheDir` bütün package tarafından paylaşıldığı için bu testler `t.Parallel()`
çağırmaz. Paralel test istiyorsanız dizini bunun yerine `newApp`'e parametre olarak
verin.

Caching'i test eden bir test bu durumdan yararlanabilir. Aynı handler'a iki request
atın ve ikincisinin cache'ten sunulduğunu kontrol edin. Ya da iki request arasında
`app.InvalidateTags`'i çağırın ve ikincisinin cache'ten sunulmadığını kontrol edin.

## Form'lar ve forgery token

Güvenli olmayan bir HTTP method'unun arkasındaki her action, bir request forgery
token'ı kontrol eder. Form post'ları da `fetch()` çağrıları da buna dahildir. Testin de
bu token'ı bir tarayıcının yaptığı gibi göndermesi gerekir:

1. Form'un bulunduğu page'e `GET` atın. Response `collage_csrf` cookie'sini set eder.
   Page'deki `{{csrfToken}}` de aynı token'ı taşıyan gizli bir `_csrf` alanı render
   eder.
2. Token'ı page'in içinden okuyun.
3. Form'u `POST` edin. Token'ı `_csrf` alanında ya da bir `fetch()`'in yaptığı gibi
   `X-CSRF-Token` header'ında gönderin. 1. adımdaki cookie'yi de ekleyin.

Scaffold'daki helper'lar şunlardır:

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

`(&http.Response{Header: page.Header()}).Cookies()`, recorder'daki `Set-Cookie`
header'larını standart kütüphanenin kendi cookie parser'ıyla parse eder. Böylece
test, bir tarayıcının geri göndereceği şeyin aynısını gönderir.

Token'ın form alanında gittiği bir form post'u:

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

Token'ın header'da gittiği bir JSON endpoint'i:

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

Korumanın açık kalmasını güvenceye alan test de şudur:

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

Token `Security.CSRFKey` ile imzalanır. Bu key ayarlanmamışsa uygulama için üretilen
bir key kullanılır. Her iki durumda da bir testteki page ve post aynı uygulamaya
gider. Bu yüzden testlerin kendilerine ait bir key'e ihtiyacı yoktur. `_csrf` ve
`collage_csrf` varsayılan isimlerdir. Bunları `Security.CSRFFieldName` ya da
`CSRFCookieName` ile değiştiren bir uygulama, testlerinde de aynı değişikliği yapar.

## Document'lar, redirect'ler ve status code'lar

Bir document, bir page gibi test edilir. Body'nin yanında content type'ı da kontrol
edin. Bir document'ı document yapan şeyin yarısı content type'ıdır:

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

Not-found page, `Allow` header'ıyla dönen bir 405, bir redirect'in `Location`'ı ve
bir `Cache-Control` header'ı recorder üzerinde birer alandır:

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

`app.Use` ile register edilen middleware handler'ın bir parçasıdır. Bu yüzden bu
testlerde de çalışır. Bir preview'ı ya da bir `collage.Vary` boyutunu test etmek
için `ServeHTTP`'den önce request'e cookie'yi ya da header'ı ekleyin.

## Her page render edilir

Bütün page'leri ziyaret eden bir test, yalnızca birinde hata veren template'i
yakalar. collage-docs içeriğini yükler ve her page'e request atar. Aşağıdaki örnek
yukarıdaki `handler` ve `get` helper'larını kullanır ve bütün tur için tek bir
uygulama kurar:

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

Page'leri bir CMS'ten gelen bir sitede aynı test,
[path provider'ınızın](/docs/static-export#dynamic-paths-pathprovider) döndürdüğü
listeyi dolaşabilir.

## Export'u test etmek

Static export da koddur. Atlanan bir page'in production'da fark edilmeden geçtiği tek
yer de burasıdır. Export'u `t.TempDir()` içine çalıştırın ve beklediğiniz dosyaların
orada olduğunu kontrol edin:

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

Bu test, `collage export`'un çalıştırdığı `staticBuild` fonksiyonunun aynısını
çağırır. Böylece test page'lerin yanında path provider'ı ve builder seçeneklerini de
kapsar. `staticBuild` build'in hatasını döner. Degraded bir page, boş bir render ya
da bir panic bu hata yüzünden testi başarısız kılar.

Atlanan page'leri ve uyarıları da kontrol etmek için builder'ı doğrudan çağırın ve raporu
okuyun:

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

`skip.Err`, v0.10.0'dan beri her atlamanın arkasındaki sentinel error'dır. Ayrıntılar
için [Hatalar](/docs/errors#static-builds) sayfasına bakın.

## HTTP olmadan render etmek

`app.RenderPath` tek bir page'i export'un yaptığı gibi render eder. Bu sırada
request, page cache ve middleware yoktur. Sonuç olarak HTML'i ve render'ın ne
yaptığını döner. Plugin'lerinizin render hook'ları yine çalışır, data cache de
çalışır. [`collage.Cached`](/docs/caching#caching-data-across-pages) bir render ya da
serve edilen bir request sırasında bir değer saklarsa, aynı uygulamadaki bir sonraki
`RenderPath` çağrısı o değeri alır. Bir test taze veriye ihtiyaç duyuyorsa yeni bir
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

`app.RenderDocumentPath` aynı işi bir document için yapar. Çoğu test için
`app.Handler()` daha iyi bir seçimdir, çünkü okuyucunun gerçekte ne aldığını test
eder. Bu iki metot tek bir render'a ayrıntılı bakmak içindir.

## Testleri çalıştırmak

```sh
go test ./...
```

Bu komutu build ya da export almadan önce CI'da çalıştırın. Gerçek `newApp`'i
kullanan bir test, sitenin başlamasını engelleyecek her sorunda CI'da başarısız olur.
Bunu CI'da görmek, sunucuda öğrenmekten daha ucuzdur.
