---
description: Gerçek uygulamayı app.Handler() ve pkg/collagetest ile test edin. pkg/collagetest, cookie'leri saklayan ve form'ları bir tarayıcının yaptığı gibi gönderen bir client'tır. Page'leri, forgery token'ıyla birlikte form'ları, document'ları ve static export'u bu şekilde sınayabilirsiniz.
reference: NewBuilder, BuildOptions, App
---

# Test yazmak

Bir collage uygulaması bir `http.Handler`'dır. `app.Handler()` bu handler'ı döndürür.
Dinleyen bir sunucu yoktur, seçmeniz gereken bir port da yoktur. Bu yüzden sıradan
bir Go testi bütün siteyi çalıştırabilir. Routing, data handler'lar, template'ler,
cache, form'lar, document'lar, plugin'ler ve middleware testte de production'daki gibi
çalışır. `pkg/collagetest` (v0.40.0'dan beri) bunu bir tarayıcının yaptığı gibi yapan
client'tır: bir response'un set ettiği cookie'leri saklar ve bir form'u, page'in
içine koyduğu bütün gizli alanlarla birlikte gönderir.

```go
import "github.com/Elagoht/collage/pkg/collagetest"

func TestLogin(t *testing.T) {
	c := client(t) // below: a collagetest.Client on the application main builds

	page := c.Get("/login").WantStatus(http.StatusOK)
	res := c.Submit(page, "/login", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct horse"},
	}).WantStatus(http.StatusSeeOther)

	if res.Location() != "/panel" {
		t.Errorf("Location = %q, want /panel", res.Location())
	}
	c.Follow(res).WantStatus(http.StatusOK) // the page the login's cookie opens
}
```

`collage new --template demo` komutunun scaffold ettiği demo projesi, bu şekilde
kurulmuş bir test dosyasıyla gelir. Bu sayfa o dosyadaki kalıbı adım adım anlatır.

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
// client returns a browser of its own on a fresh copy of the site.
func client(t *testing.T) *collagetest.Client {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v, want nil", err)
	}
	return collagetest.New(t, app.Handler())
}
```

`newApp(false, 0)` production config'ini kurar. Development mode kapalıdır ve port
`0`'dır. collage bu `0` değerini kendi varsayılan portuyla değiştirir. Zaten hiçbir
şey bu portu dinlemez.

Bundan sonra her test, bir request atmaktan ve response'a bakmaktan ibarettir:

```go
func TestPagesRender(t *testing.T) {
	c := client(t)
	for target, want := range map[string]string{
		"/":         "Explore the features",
		"/features": "Four live demos",
		"/hello":    "Hello, stranger!",
	} {
		res := c.Get(target)
		if res.Status != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, res.Status)
			continue
		}
		if !strings.Contains(res.Body, want) {
			t.Errorf("GET %s does not contain %q", target, want)
		}
		// Counted, not merely found: a layout writing one and a page hoisting
		// another is two titles, which is a page that looks fine and is not.
		if n := strings.Count(res.Body, "<title>"); n != 1 {
			t.Errorf("GET %s has %d titles, want exactly 1", target, n)
		}
	}
}
```

`client(t)`'nin yaptığı gibi her test için yeni bir client kurun. `app.Handler()`'a
yapılan ilk çağrı uygulamayı başlatır ve register aşamasını kapatır. Client her test
için yeniden kurulduğundan her test boş bir cache ve boş bir cookie jar ile başlar.
İki okuyucu, iki client demektir.

Başlatma başarısız olabilir. Örneğin bir plugin'in `Init`'i hata döner, bir page
register edilmemiş bir error page'i belirtir ya da bir mount bir route'u gölgeler. Bu
durumda `app.Handler()` her request'e 503 ile cevap verir ve nedenini log'a yazar. Bu
hatayı testte bir hata olarak almak istiyorsanız önce `app.Start()`'ı çağırın:

```go
if err := app.Start(); err != nil {
	t.Fatalf("start: %v", err)
}
```

## Client

| | |
| --- | --- |
| `collagetest.New(t, h)` | `h` için boş bir cookie jar'ı olan bir client |
| `c.Get(target)` | Bir path'e ya da mutlak bir URL'ye `GET` |
| `c.Submit(page, action, values)` | `page` içinde action'ı `action` olan form'u gönderir |
| `c.Follow(res)` | Bir redirect'in belirttiği `Location`'a `GET` |
| `c.Request(method, target, body)` / `c.Do(req)` | Diğer her request: bir JSON body, kendine ait bir header. `Do`, jar'daki cookie'leri ekler ve response'un set ettiklerini saklar |

Bir `*Response`, `Status`, `Header`, `Body`, `URL` ve `Method` taşır.
`WantStatus(code)`, status farklıysa testi body ile birlikte başarısız kılar ve
response'u döner. Böylece bir request ve onun kontrolü tek bir satır gibi okunur.
`Location()`, `Location` header'ıdır. `CSRFToken()` ise page'deki ilk gizli `_csrf`
input'unun değeridir.

Jar bir `net/http/cookiejar`'dır ve cookie'leri bir tarayıcının yaptığı gibi path'e
ve süresine göre kapsar. Tek başına bir path, `net/http/httptest`'in kullandığı host
olan `http://example.com`'a gönderilir. Cookie'lerini `Secure` olarak işaretleyen bir
site mutlak `https://example.com/...` hedefleri kullanır. Bu request'ler TLS
üzerinden gelmiş gibi ulaşır, böylece jar o cookie'leri geri gönderir.

Redirect'ler takip edilmez: bir test genellikle `303`'ü ve nereyi gösterdiğini görmek
ister. Test, arkasındaki page ile ilgiliyse redirect'i `Follow` takip eder.

## Disk cache'ini izole edin

Scaffold'daki cache dizini bir package değişkenidir. `client`, uygulamayı kurmadan
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

Caching'i test eden bir test bu durumdan yararlanabilir. Aynı client ile iki request
atın ve ikincisinin cache'ten sunulduğunu kontrol edin. Ya da iki request arasında
`app.InvalidateTags`'i çağırın ve ikincisinin cache'ten sunulmadığını kontrol edin.

## Form'lar ve forgery token

Güvenli olmayan bir HTTP method'unun arkasındaki her action, bir request forgery
token'ı kontrol eder. Form post'ları da `fetch()` çağrıları da buna dahildir. Test bu
token'ı bir tarayıcının yaptığı gibi gönderir: form'un bulunduğu page'e `GET` atar.
Bu page `collage_csrf` cookie'sini set eder ve `{{csrfToken}}`'ı gizli bir `_csrf`
alanı olarak render eder. Test ardından form'u bu alan ve bu cookie ile post eder.
`Submit` bunların hepsini yapar:

```go
func TestHelloGreetsTheSubmittedName(t *testing.T) {
	c := client(t)

	res := c.Submit(c.Get("/features"), "/hello", url.Values{"name": {"Ada"}}).WantStatus(http.StatusOK)
	if !strings.Contains(res.Body, "Hello, Ada!") {
		t.Errorf("the response does not greet the name:\n%s", res.Body)
	}
}
```

`Submit` form'u bulur, içindeki **bütün gizli input'ları** taşır (forgery token'ı ve
bir plugin'in form'a koyduğu her şeyi, örneğin bir honeypot'un imzalı zaman
damgasını) ve `values`'u bunların üzerine koyar: `values` içindeki bir ad, aynı adlı
gizli bir input'un yerine geçer. Görünür alanları doldurmak testin işidir. Bir bot'un
dolduracağı tuzak alan boş kalır, böylece gönderim bir okuyucunun gönderimidir.

- `action`, page'in URL'sine göre çözümlenir ve form'unki gibi decode edilmiş
  hâliyle karşılaştırılır: `"/login"`, yanındaki bir page'den `"login"`, mutlak bir
  URL ve action'ı `"/giri%c5%9f"` olan bir form için `"/giriş"` aynı form'u belirtir.
  `action`'ı olmayan bir form kendi page'ine gönderilir. Boş bir `action`, page'in tek
  form'u demektir. Hiç eşleşme olmaması ya da birden fazla eşleşme olması testi
  başarısız kılar ve page'deki form'ların action'larını listeler.
- Form'un `method`'u ve `enctype`'ı dikkate alınır: `GET` form'u değerlerini query'de
  gönderir, `multipart/form-data` multipart olarak gönderilir, diğer her şey
  `application/x-www-form-urlencoded` olarak gönderilir.
- Form'lar bir HTML parser'ıyla değil, bir scanner ile bulunur: yorumlar ve
  `<script>`, `<style>`, `<template>` ve `<textarea>` gövdeleri atlanır. Böylece
  oralara metin olarak yazılmış markup form sanılmaz. Disabled bir gizli input,
  bir tarayıcının göndermediği gibi gönderilmez.

Bir JSON endpoint'i token'ı, bir `fetch()`'in gönderdiği gibi `X-CSRF-Token`
header'ında alır:

```go
func TestCountAnswersWithTheNewCount(t *testing.T) {
	c := client(t)
	page := c.Get("/features")

	req := c.Request(http.MethodPost, "/api/count", nil)
	req.Header.Set("X-CSRF-Token", page.CSRFToken())
	res := c.Do(req).WantStatus(http.StatusOK)

	var answer struct{ Count int64 }
	if err := json.Unmarshal([]byte(res.Body), &answer); err != nil {
		t.Fatalf("body = %q, want JSON: %v", res.Body, err)
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
	c := client(t)
	for _, target := range []string{"/api/count", "/hello"} {
		req := c.Request(http.MethodPost, target, strings.NewReader(url.Values{"name": {"Ada"}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if res := c.Do(req); res.Status != http.StatusForbidden {
			t.Errorf("POST %s with no token = %d, want 403", target, res.Status)
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
	res := client(t).Get("/healthz").WantStatus(http.StatusOK)

	var health struct{ Status string }
	if err := json.Unmarshal([]byte(res.Body), &health); err != nil || health.Status != "ok" {
		t.Errorf("body = %q, want a status of ok", res.Body)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
```

Not-found page, `Allow` header'ıyla dönen bir 405, bir redirect'in `Location`'ı ve
bir `Cache-Control` header'ı response üzerinde birer alandır:

```go
func TestNotFoundPage(t *testing.T) {
	res := client(t).Get("/there-is-nothing-here").WantStatus(http.StatusNotFound)

	if !strings.Contains(res.Body, "There is nothing at this address") {
		t.Errorf("body = %q, want this site's own not-found page", res.Body)
	}
}
```

`app.Use` ile register edilen middleware handler'ın bir parçasıdır. Bu yüzden bu
testlerde de çalışır. Bir preview'ı ya da bir `collage.Vary` boyutunu test etmek
için `c.Do`'dan önce bir `c.Request`'e cookie'yi ya da header'ı ekleyin. Client bir
kolaylıktır, zorunluluk değildir: `app.Handler()`, diğer her handler gibi bir
`net/http/httptest` recorder'ı ile de çalışır.

## Her page render edilir

Bütün page'leri ziyaret eden bir test, yalnızca birinde hata veren template'i
yakalar. collage-docs içeriğini yükler ve her page'e request atar. Aşağıdaki örnek
bütün tur için tek bir client, dolayısıyla tek bir uygulama kullanır:

```go
// Every page of the documentation renders, with its own title.
func TestEveryDocRenders(t *testing.T) {
	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatalf("site.Load: %v", err)
	}
	c := client(t)
	for _, page := range loaded.Pages() {
		res := c.Get(page.URL())
		if res.Status != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), res.Status)
			continue
		}
		if !strings.Contains(res.Body, "<title>"+page.Title+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
	}
}
```

Page'leri bir CMS'ten gelen bir sitede aynı test, page'lerinizin
[`WithStaticParams`](/docs/static-export#dynamic-paths-withstaticparams)'ının
listelediği değerleri dolaşabilir.

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
çağırır. Böylece test page'lerin yanında `WithStaticParams`'ı ve builder
seçeneklerini de kapsar. `staticBuild` build'in hatasını döner. Degraded bir page,
boş bir render ya da bir panic bu hata yüzünden testi başarısız kılar.

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
		t.Errorf("%s has a {param} and no WithStaticParams", skip.Page)
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
`app.Handler()` ve `collagetest` daha iyi bir seçimdir, çünkü okuyucunun gerçekte ne
aldığını test ederler. Bu iki metot tek bir render'a ayrıntılı bakmak içindir.

## Her link çözümlenir

Adla kurulan bir link (`{{pageURL "post" "slug" .Slug}}`, `{{actionURL "logout"}}`)
yalnızca template render edildiğinde ve yalnızca ona ulaşan page'de hata verir.
`app.Check()` (v0.40.0'dan beri) hiçbir şey render etmeden her template'in link'lerini
tek seferde kontrol eder. [`collage check`](/docs/cli#collage-check) komutunun
çalıştırdığı da budur:

```go
func TestLinks(t *testing.T) {
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if findings := app.Check(); len(findings) > 0 {
		t.Errorf("broken links: %v", findings)
	}
}
```

## Testleri çalıştırmak

```sh
go test ./...
```

Bu komutu build ya da export almadan önce CI'da çalıştırın. Gerçek `newApp`'i
kullanan bir test, sitenin başlamasını engelleyecek her sorunda CI'da başarısız olur.
Bunu CI'da görmek, sunucuda öğrenmekten daha ucuzdur.
