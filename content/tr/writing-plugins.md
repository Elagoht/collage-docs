---
description: Plugin sözleşmesi, Host ve ConfigHost'un sundukları, her hook ve neyi değiştirebileceği, testleriyle birlikte eksiksiz bir plugin.
reference: Plugin, Host, ConfigHost, ConfigReader, PluginConfig, Configurer, Command, BeforeRenderHook, BeforeActionHook, BeforeActionEvent, AfterRenderHook, CacheInvalidateHook, FragmentRequest, FragmentRender, HoistItem, StreamCloser, PageURL, FragmentReport, PathTag, Finding, FindingLevel, FindingWarning, FindingError, ErrBuildFindings, BuildFinishedHook, BuildFinishedEvent, BuiltFile, RequestHook, RouteOf, RouteInfo, Route, ClientIP, Key, NewKey, RenderValues
---

# Plugin yazmak

Plugin, dört metodu olan bir Go tipidir. Geri kalan her şey isteğe bağlıdır: bir
render'a tepki vermek, çıktıyı yeniden yazmak ya da bir template fonksiyonu
eklemek. İstediğiniz hook'un interface'ini implement edersiniz, framework de onu
type assertion ile bulur.

Bu sayfa, bu yüzeyin referansıdır. [Plugin kullanmak](/docs/plugins) ise işin öbür
tarafını anlatır: bir uygulama, sizin yazdığınız plugin'i nasıl register eder ve
nasıl yapılandırır.

## Sözleşme

```go
type Plugin interface {
	Name() string
	Version() string
	Init(ctx context.Context, host collage.Host) error
	Shutdown(ctx context.Context) error
}
```

- **`Name`** plugin'i tanımlar. Boş olmamalı ve uygulama içinde benzersiz olmalıdır.
  Aynı zamanda plugin'in config bölümünün key'idir. Bu yüzden bir module path
  gibi okunacak bir ad seçin: `acme/stamp`.
- **`Version`** plugin'inizin kendi versiyonudur ve teşhis amaçlı kullanılır.
- **`Init`** uygulama başlarken bir kez çalışır. Bu, uygulama kendi page'lerini
  register ettikten sonra ve ilk request'ten önce olur.
- **`Shutdown`**, `Init`'in edindiği her şeyi serbest bırakır. Register edilmiş her
  plugin için çağrılır; o plugin'in `Init`'inin çalışıp çalışmadığı ya da başarılı
  olup olmadığı fark etmez. Birden fazla kez de çağrılabilir. Bu yüzden `Init`
  çalışmadan da güvenli olmalı ve idempotent olmalıdır. Ayrıntılar için
  [Yaşam döngüsü](#lifecycle) bölümüne bakın.

Hook'lar type assertion ile bulunduğu için, adı yanlış yazılmış bir hook metodu
derleme hatası vermez. Sadece hiç çalışmayan bir hook olur. Implement etmek
istediğiniz her interface için bir assertion yazın:

```go
var (
	_ collage.Plugin          = (*Plugin)(nil)
	_ collage.AfterRenderHook = (*Plugin)(nil)
)
```

## İki aşama: Configure ve Init

Bazı işlerin template'ler parse edilmeden önce yapılması gerekir. `html/template`,
yalnızca template parse edildiği anda function map'inde bulunan bir fonksiyonu
çağırabilir. Parse işlemi de `collage.New` içinde gerçekleşir. Bu yüzden isteğe
bağlı, daha erken bir aşama vardır:

```go
type Configurer interface {
	Configure(ctx context.Context, host collage.ConfigHost) error
}
```

`Configure`, `New` içinde çalışır. Her plugin için bir kez, register sırasına göre ve
template'ler parse edilmeden önce çağrılır. Dönen bir hata `New`'u iptal eder. Henüz
hiçbir kaynak edinilmediği için geri alınacak bir şey de yoktur.

`Init` daha sonra, uygulama başlarken çalışır. Uygulama, `Handler`, `ListenAndServe`,
`Start`, `RenderPath`, `RenderDocumentPath` ya da `DispatchCommands` ilk kez
çağrıldığında başlar. Static build de buna dahildir, çünkü builder `RenderPath`
üzerinden render eder. Bu noktada uygulama page'lerini register etmiş olur. Böylece
bir plugin onları okuyabilir ya da kendi page'lerini ekleyebilir.

İkisine de ihtiyaç duyan bir plugin ikisini de implement eder. **`Configurer`'ı
implement eden bir plugin, `Config.Plugins` içinde verilmelidir.** `RegisterPlugin`,
`New` template'leri parse ettikten sonra çağrılır. Bu yüzden böyle bir plugin'in
`Configure`'ını sessizce atlamaz; plugin'i `ErrConfigurerRegisteredLate` ile
reddeder.

### Her aşama nelere erişebilir

| | `ConfigHost` (Configure) | `Host` (Init) |
| --- | --- | --- |
| `DevMode`, `Logger`, `collage.PluginConfig(host, …)` | evet | evet |
| `AddTemplateFunc`, `AddRenderFunc`, `WrapMount` | evet | — |
| `Pages`, `Page`, `InvalidateTags` | — | evet |
| `URL`, `FragmentURL`, `Locales`, `PageURLs` | — | evet |
| `BuildID`, `ServeStatus` | — | evet |
| `RegisterPage`, `RegisterDocument`, `Mount`, `Handle`, `Use` | — | evet |
| `RenderFragment` | — | evet |
| `RegisterCommand` | — | evet |

`ConfigHost` bilerek daha dar tutulmuştur. `Configure` sırasında uygulama henüz
hiçbir şey register etmemiştir. Page listesi boş olurdu, invalidation'ın da
ulaşabileceği bir cache olmazdı.

### ConfigHost

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın development modunda çalışıp çalışmadığını söyler. |
| `Logger() *slog.Logger` | Uygulamanın logger'ını döner. |
| `ConfigReader` (embed edilmiş) | `collage.PluginConfig(host, defaults)` bu plugin'in config bölümünü bunun üzerinden okur. Bkz. [Yapılandırma](#configuration). |
| `AddTemplateFunc(name, fn) error` | Bir template fonksiyonu ekler. Ad daha önce eklenmişse `ErrDuplicateTemplateFunc` döner; ekleyen başka bir plugin de olabilir, aynı plugin'in önceki bir çağrısı da. |
| `WrapMount(wrap func(fs.FS) fs.FS)` | Mount edilen her dosya sistemine uygulanacak bir dönüşümü register eder. Wrapper'lar register edildikleri sırayla uygulanır. |
| `AddRenderFunc(name, factory) error` | Her render için o render'ın `*RenderContext`'inden yeniden üretilen bir template fonksiyonu ekler. Bkz. [Template fonksiyonları](#template-functions) (v0.21.0'dan beri). |

### Host

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın development modunda çalışıp çalışmadığını söyler. |
| `Logger() *slog.Logger` | Uygulamanın logger'ını döner. |
| `ConfigReader` (embed edilmiş) | `collage.PluginConfig(host, defaults)` bu plugin'in config bölümünü bunun üzerinden okur. |
| `Pages() []*collage.Page` | Register edilmiş bütün page'leri döner. Her biri bir defensive copy'dir. |
| `Page(name) (*collage.Page, bool)` | Adı verilen page'i defensive copy olarak döner. |
| `InvalidateTags(ctx, tags...) error` | Bu tag'lerden herhangi biriyle oluşturulmuş bütün cache entry'lerini düşürür. |
| `RegisterPage(page) error` | Plugin'in eklediği bir page'i register eder. |
| `RegisterDocument(doc) error` | Plugin'in eklediği bir document'ı register eder. |
| `Mount(prefix, fsys, opts...) error` | Bir dosya sistemini bir URL prefix'i altında sunar. |
| `Handle(prefix, handler) error` | `App.Handle` gibi, bir `http.Handler`'ı `/` ile biten bir URL prefix'i altında ya da sonunda `/` olmayan tek bir tam path'te (`/metrics`, v0.24.0'dan beri) sunar. Örneğin bir event stream ya da bir WebSocket (v0.18.0'dan beri). |
| `RenderFragment(r, req) (*collage.FragmentRender, error)` | Bir page'in `WithFragmentPath` ile açtığı bir fragment'i parçalar hâlinde render eder. Bkz. [Fragment göndermek](#pushing-fragments) (v0.18.0'dan beri). |
| `RegisterCommand(cmd) error` | Bir komut ekler. Bkz. [Komutlar](#commands). |
| `Use(middleware) error` | Her request'i sarmalar (v0.21.0'dan beri), plugin'in register edildiği yerde: `Config.Plugins` içindeki bir plugin için uygulamanın kendi middleware'inin dışında (v0.38.0'dan beri; önceden içindeydi). |
| `URL(name, locale, params) (string, error)` | Bir page'in ya da document'ın path'ini `App.URL`'in oluşturduğu gibi döner (v0.21.0'dan beri). |
| `FragmentURL(page, fragment, locale, params) (string, error)` | Bir fragment path'inin path'ini `App.FragmentURL`'in oluşturduğu gibi döner (v0.21.0'dan beri). |
| `Locales() (default, supported)` | Varsayılan locale'i ve varsayılan dahil desteklenen her locale'i döner (v0.21.0'dan beri). |
| `PageURLs(ctx, name) ([]collage.PageURL, error)` | Bir page'in her locale'de cevap verdiği her URL'yi döner; bir pattern, `WithStaticParams`'ı üzerinden açılır. Bir sitemap'in içeriği budur (v0.21.0'dan beri). |
| `BuildID() string` | Sunan build'i adlandırır: `Config.Cache.Version` ya da executable'ın bir parmak izi. Tarayıcının deploy'lar boyunca sakladıklarını sürümlemek içindir; örneğin bir service worker'ın cache'leri ya da bir asset'in query string'i (v0.24.0'dan beri). |
| `ServeStatus(w, r, status)` | Request'e status'la ve sitenin o status için kendi page'iyle cevap verir: 404 ve 410 için not-found page'i, diğerleri için error page'i. Bir request'e kendisi cevap veren ve cevabı sitenin geri kalanı gibi görünmesi gereken bir plugin içindir (v0.24.0'dan beri). |

`Host` ayrıca isteğe bağlı bir capability sunar: `collage.Origins`. Ona metot olarak
değil type assertion ile ulaşılır; böylece `Host`'u implement eden bir test double'ı
derlenmeye devam eder. Bkz. [Bir host'un origin'i](#which-origin-a-host-has-originresolver)
(v0.42.0'dan beri).

`Init`'e gelen değer `*App` değildir. Yalnızca bu metotları ileten dar bir değerdir.
Bu yüzden bir plugin, type assertion ile `ListenAndServe`'e, `Shutdown`'a,
router'a, cache'e ya da template kümesine ulaşamaz. `Handle` ve `RenderFragment`
v0.18.0'da eklendi. `Use`, `URL`, `FragmentURL`, `Locales` ve `PageURLs` ise
v0.21.0'da, `ConfigHost`'taki `AddRenderFunc` ile birlikte eklendi. `BuildID` ve
`ServeStatus` ise v0.24.0'da geldi. **Bunların her biri bir test double'ı için
breaking change'dir:** `Host`'u ya da `ConfigHost`'u implement eden bir test
double'ının da yeni metotlara ihtiyacı vardır. v0.24.0'dan beri bunlar `BuildID` ve
`ServeStatus`'tur. v0.50.0'dan beri ikisi de `collage.ConfigReader`'ı içerir. Onun
tek metodu export edilmemiştir; bu yüzden onu yalnızca collage'ın kendi host'ları
karşılar. Plugin'i bunun yerine gerçek bir uygulama üzerinden test edin. Bkz.
[Bir plugin'i test etmek](#testing-a-plugin).

**`Host`, bir plugin'in neye ulaşabileceğini sınırlar; neyi değiştirebileceğini
sınırlamaz.** `Pages` ve `Page`, page struct'ının bir kopyasını döner. `Paths`,
`Redirects` ve `DependencyTags` container'ları da kopyalanır. Bu yüzden
bunları düzenlemek uygulamanın kendi page'ine dokunmaz. Ancak kopyanın içindeki
fragment pointer'ları hâlâ paylaşılır. Aşağıdaki event'ler de kopyayı değil, *canlı*
page'i taşır, çünkü her request'te bir page'i ve fragment ağacını kopyalamak hot
path'e pahalıya mal olurdu. Bir event'in `Page`'i üzerinden yazarsanız, eşzamanlı
bütün request'lerin okuduğu page'i değiştirmiş olursunuz. Bu bir data race'tir ve
`go test -race` bunu raporlar. Page'leri salt okunur kabul edin. Plugin'ler güvenilen
koddur; bir sandbox içinde çalışmazlar.

Bir plugin'in register ettiği page, document ve mount'lar, uygulamanınkilerle aynı
kurallara tabidir. Zaten kullanılan bir ad ya da path bir startup hatasıdır. Hangi
kaydın kazanacağı register sırasına bırakılmaz.

### Fragment göndermek

`RenderFragment`, fragment'lerin tarayıcı tarafından istenmesini beklemek yerine
onları kendi sahip olduğu bir bağlantı üzerinden gönderen plugin'ler içindir. Bu
bağlantı bir event stream ya da bir WebSocket olabilir. Metot, fragment'in path'ine
gelen bir request'in render edeceği şeyin tam olarak aynısını render eder. Bir
response yerine parçaları döner:

| Field | |
| --- | --- |
| `HTML` | Markup. İçindeki her form okuyucunun forgery token'ını taşır |
| `Head` | Fragment'in marker koymadığı bir alana hoist ettikleri; area'ları ve key'leriyle birlikte `HoistItem` olarak |
| `DependencyTags` | Render'ın bağlı olduğu tag'ler. Neyin gönderileceğini bilmek için bunları `CacheInvalidateEvent.Tags` ile eşleştirin |
| `Shared` | Render bu anda her okuyucu için aynıdır: page herkes için cache'lenir ya da alt ağaçtaki her handler `Static()` veya `Shared()` olarak tanımlanmıştır ve slot resolver yoktur. Ayrıca form token'ı da yoktur (v0.19.0'dan beri; öncesinde `Static()` olmayan her handler request'i okuyor sayılırdı) |
| `ETag` | Aynı body için fragment'in path'ine gelen bir request'in alacağı ETag. Böylece bir client, gönderilen bir kopyayı zaten gösterdiği kopyadan ayırt edebilir ya da ikisinin aynı olduğunu anlayabilir (v0.19.0'dan beri) |
| `Cookie` | Request hiç cookie taşımıyorsa `HTML`'deki form'ların ihtiyaç duyduğu forgery cookie'si |

Bir `FragmentRequest`, fragment'i ya `Page`, `Fragment`, `Locale` ve `Params` ile ya
da `Path` ile belirtir. `Path`, bir page'in `{{fragmentURL}}` ile link verdiği
URL'dir; query de buna dahildir. Bu URL, ona gelen bir request gibi çözümlenir.
Gösterdiği element'lere abone olan bir client yalnızca onların URL'lerini bilir. Bu
yüzden bir stream'in elinde genellikle `Path` vardır.

```go
out, err := host.RenderFragment(r, collage.FragmentRequest{Path: "/live/cpu"})
```

Yalnızca page'in açtığı fragment'ler render edilir: bir stream tam olarak HTTP'nin
ulaştığı yere ulaşır. Buna fragment'in kendi
[guard](/docs/pages-and-layouts#private-pages-guards)'ı da dahildir (v0.28.0'dan
beri): guard'a, fragment'in URL'sine gelen bir request'te olduğu gibi `r` hakkında
sorulur. Reddettiği bir okuyucu render yerine `collage.ErrGuardRefused` alır; o
bağlantıya hiçbir şey göndermeyin. `Shared` olmayan bir render tek bir okuyucunun
verisini taşıyabilir. Bu yüzden herkes için bir kez değil, her bağlantı için o
bağlantının request'iyle render edilmelidir. Handler'ı belli bir anda herkes için aynı şeyi
döndüren bir fragment (bir ölçüm gibi)
[`Shared()`](/docs/caching#a-page-that-declares-none) olarak işaretlenir. Böylece
render'ı, page'i static yapmadan `Shared` sayılır. `App.RenderFragment` aynı
metottur ve uygulamanın kendi kodu içindir.

### Stream'ler ve shutdown

Bir plugin'in `Shutdown`'ı sunucu durduktan sonra çalışır. Sunucu da açık her
request'in bitmesini bekleyerek durur. Bir event stream ya da bir WebSocket ise
kendiliğinden hiç bitmez. Böyle bir bağlantı sunan plugin `StreamCloser`'ı
implement eder (v0.18.0'dan beri):

```go
var _ collage.StreamCloser = (*Plugin)(nil)

func (p *Plugin) CloseStreams() { p.hub.close() }
```

`CloseStreams` shutdown başlarken, sunucu beklemeye başlamadan önce çalışır.
Stream'leri onları beklemeden sonlandırmalıdır. `Handle` ile sunulan bir handler,
write deadline'ını `http.NewResponseController(w).SetWriteDeadline` ile ileri
alabilir ve bağlantıyı `Hijack` ile devralabilir. Bunları çıplak bir `net/http`
sunucusunda da yapabilirdi.

## Hook'lar

| Interface | Metot | Event | Ne zaman çalışır | Neyi değiştirebilir |
| --- | --- | --- | --- | --- |
| `RequestHook` | `OnRequest` | `*http.Request` | Her request'te ilk olarak; collage'ın request span'inden, middleware'den ve routing'den önce (v0.25.0'dan beri) | request'in context'ini |
| `PageResolvedHook` | `OnPageResolved` | `PageResolvedEvent` | Her page request'inde bir kez, routing'in ve page'in guard'larının hemen ardından; cache hit'ler dahil | hiçbir şeyi |
| `BeforeRenderHook` | `OnBeforeRender` | `BeforeRenderEvent` | Yeni bir page render'ından önce | event'te hiçbir şeyi; `ev.Context` üzerinden hoist edebilir |
| `BeforeActionHook` | `OnBeforeAction` | `BeforeActionEvent` | Bir action'ın handler'ından önce, guard'larından, body sınırından ve forgery kontrolünden sonra (v0.31.0'dan beri) | `ev.Result`; handler'ın yerine cevap verir |
| `AfterRenderHook` | `OnAfterRender` | `AfterRenderEvent` | Bir page render'ı başarıyla bittikten sonra. `ev.Values`'u taşır (bir key'in `In` metoduyla okunur) | `ev.HTML`; `ev.Warn` ve `ev.Error` ile raporlar |
| `PersonaliseHook` | `OnPersonalise` | `PersonaliseEvent` | Cache'ten sonra, bir okuyucuya giden her HTML response'u için (v0.43.0'dan beri) | `ev.Body`, `ev.Header`, `ev.Personal` |
| `DocumentRenderedHook` | `OnDocumentRendered` | `DocumentRenderedEvent` | Bir document handler'ı body'sini ürettikten sonra | `ev.Body` |
| `CacheWriteHook` | `OnCacheWrite` | `CacheWriteEvent` | Bir page ya da document cache'e yazılmadan önce | `ev.Skip`, `ev.TTL`, `ev.Tags` |
| `CacheInvalidateHook` | `OnCacheInvalidate` | `CacheInvalidateEvent` | Entry'ler tag ile invalidate edildikten sonra | hiçbir şeyi |
| `OriginResolver` | `Origin` | string olarak host | `collage.BaseURL` ya da `Origins` bir host'un origin'ini sorduğunda (v0.42.0'dan beri) | dönen origin |
| `ErrorHook` | `OnError` | `ErrorEvent` | Bir request sunulurken bir hata oluştuğunda | hiçbir şeyi |
| `BuildFinishedHook` | `OnBuildFinished` | `BuildFinishedEvent` | Bir static build her dosyayı yazdığında, bir kez (v0.21.0'dan beri) | `ev.Warn` ve `ev.Error` ile raporlar |

`OnRequest` dışında her hook metodunun imzası
`func(ctx context.Context, ev *Event) error` biçimindedir.

### RequestHook

```go
type RequestHook interface {
	OnRequest(r *http.Request) (context.Context, func(status int))
}
```

Bir request'te her şeyden önce çalışır (v0.25.0'dan beri): collage kendi request
span'ini başlatmadan, middleware'den ve routing'den önce. Request'in sunulacağı
context'i döner; bu context `r.Context()`'ten türetilir. Bir de response yazıldıktan
sonra collage'ın status ile çağırdığı bir fonksiyon döner: request sunulurken bir
panic olmadıkça bir kez çağrılır. Bu fonksiyon `nil` olabilir ve response'a
yazmamalıdır.

Bir tracing plugin'inin ihtiyaç duyduğu şey budur. Çağırandan gelen bir trace,
collage'ın kendi `collage.http` span'inin parent'ı olmalıdır. `Host.Use` ile eklenen
middleware bunun için geç kalır; collage'ın çoktan başlattığı span'in içinde
çalışır:

```go
func (p *Plugin) OnRequest(r *http.Request) (context.Context, func(status int)) {
	ctx := p.propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := p.tracer.Start(ctx, r.Method, trace.WithSpanKind(trace.SpanKindServer))
	return ctx, func(status int) {
		if kind, name := collage.RouteOf(ctx); kind != "" {
			span.SetName(r.Method + " " + kind + ":" + name)
		}
		span.SetAttributes(attribute.Int("http.response.status_code", status))
		span.End()
	}
}
```

`collage.RouteOf(ctx)`, request'in neye resolve edildiğini bildirir: türünü (`"page"`,
`"document"`, `"action"`, `"mount"` ya da `"handler"`) ve register edildiği adı ya
da prefix'i. Bu ad page'in, document'ın ya da action'ın adıdır; mount'ta ve
handler'da ise prefix'tir. Request henüz resolve edilmemişken ve 404 gibi hiçbir şeye
resolve edilmeyen bir request'te ikisi de boştur. collage'ın request'i sunduğu
context'i okur. Finish fonksiyonunun context'i de bu context'i paylaşır,
`Metrics.HTTPResponse` da bu context'i alır. Böylece bir span ya da metric, ham
path yerine route ile etiketlenebilir. Bu, URL başına bir değer değil, sınırlı bir
kümedir.

`collage.RouteInfo(ctx)` aynı bilgiyi iki alanla birlikte bildirir: `Pattern`,
route'un register edildiği path pattern'idir (locale prefix'i olmadan
`/blog/{slug}`, ya da mount'un veya handler'ın prefix'i); `Locale` ise locale'dir.
Ad yetmediğinde etiket olarak bunu kullanın; örneğin bir trace'in `http.route`'u
için, page kadar document'ta da.

Hook'lar register sırasına göre çalışır. Her biri request'i kendinden öncekinin
oluşturduğu context ile alır. Finish fonksiyonları ters sırada çalışır. Panic'e
düşen bir hook atlanır ve request ondan önceki context ile sunulur.

v0.47.0'dan beri hook'lar, collage'ın routing'den önce yanıtladığı request'leri de
görür: encoded slash içeren bir path (`/a%2f..%2fb`) 404 ile, kirli bir path
(`/a/../b`) ise 301 ile yanıtlanır (`GET` ve `HEAD` dışındaki bir metotta 308).
Bunlar çoğu zaman bir tarayıcı botunun yoklamalarıdır. Böyle bir request için finish
fonksiyonu yazılan status ile çağrılır (request sunulurken bir panic olmadıkça bir
kez), `collage.RouteOf` boştur ve `r.URL` ham path'tir. `collage.ClientIP(r)` `OnRequest` içinde çalışır; bu yüzden istemcileri
sayan ya da banlayan bir plugin bu request'leri görebilir. `Metrics.HTTPResponse` ve
collage'ın kendi request span'i bu request'leri hâlâ atlar.

İstemcinin adresi için `r.RemoteAddr` yerine `collage.ClientIP(r)` kullanın. Bu bir
`netip.Addr`'dir: `RemoteAddr`'in host'u ya da uygulamanın `Server.TrustedProxies`'te
listelediği proxy'lerin arkasında proxy'nin bildirdiği istemci (bkz.
[Deployment](/docs/deployment#behind-a-proxy-trustedproxies)). Adres unmapped'dır,
zone taşımaz ve verilecek bir adres olmadığında sıfır `netip.Addr`'dir: `RemoteAddr`'i
olmayan paylaşılan bir page render'ında ve `X-Forwarded-For`'u kullanılabilir bir
istemci bildirmeyen (`unknown`) güvenilen bir proxy'nin arkasında. Onu asla
saymayın.

### PageResolvedHook

```go
type PageResolvedEvent struct {
	Page   *collage.Page // live — do not write through it
	Locale string
	Path   string
}
```

Bir page'e route edilen her request'te, cache'e bakılmadan önce bir kez çalışır. Bu
yüzden yeni render'ları da, cache hit'leri de görür. Page'in
[guard](/docs/pages-and-layouts#private-pages-guards)'larından sonra çalışır;
onlardan birinin engellediği bir request için hiç çalışmaz, çünkü o request page'e
hiç ulaşmamıştır. Bir document için hiçbir zaman çalışmaz. Static build
sırasında da çalışmaz: build bir request değildir ve request'leri sayan bir
plugin, kimsenin istemediği render'ları da saymış olurdu.
Dönen bir hata, request'i `"page_resolved"` stage'i altında 500 ile başarısız kılar.

### BeforeRenderHook

```go
type BeforeRenderEvent struct {
	Context *collage.RenderContext // the render about to run
	Page    *collage.Page
	Locale  string
	Path    string
	Static  bool // rendered for a static build, not for a request
}
```

Yeni bir render'dan hemen önce çalışır; cache hit'te **çalışmaz**. `PageResolvedHook`
ile farkı budur. Page'ler, error page'ler ve bir action'ın `RenderPage` ile cevap
olarak döndüğü page için çalışır. Static build'in render ettiği her page için de
çalışır.

Render context'ini alan tek hook budur ve sebebi hoisting'dir. Page'e bir şey
ekleyen plugin, bunu ağaç render edilmeden önce tanımlar. Böylece bir fragment onu
hâlâ override edebilir:

```go
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}
```

Burada yapılan bir tanım sıfır derinliğinde durur. Bu yüzden aynı key'i tanımlayan
herhangi bir fragment onun yerini alır. Varsayılan değeri plugin, özel değeri page
verir. Tanım yalnızca layout'un `{{hoist "head"}}` çağırdığı yere yerleşir. Dönen
bir hata, request'i `"before_render"` altında 500 ile başarısız kılar.

`Static` (v0.22.0'dan beri), page'in bir request için değil, `App.RenderPath`
üzerinden bir static build için render edildiğini söyler. `AfterRenderEvent`'te de
bulunur.

### BeforeActionHook

```go
type BeforeActionEvent struct {
	Action  *collage.Action
	Page    *collage.Page // the page whose URL the action answers on, or nil
	Locale  string
	Request *http.Request
	Result  *collage.ActionResult // set it to answer in the handler's place
}

func (ev *BeforeActionEvent) Form() (url.Values, error)
```

Her action request'inde bir kez, handler'dan hemen önce çalışır. Page'in guard'larından,
action'ın body sınırı konduktan ve forgery kontrolünden sonra gelir (v0.31.0'dan
beri). Bir plugin'in gönderimi kontrol ettiği yer burasıdır. Spam filtresi ya da
kota bunun örnekleridir. Buradaki body action'ın kendi sınırıyla sınırlıdır.
Body'yi daha önce, middleware'de okuyan bir plugin'in ise kendi sınırına ihtiyacı
olurdu.

`ev.Form()` form'u parse eder, URL-encoded ya da multipart fark etmez, ve alanlarını
döndürür. Parse ettiği parse edilmiş kalır. Böylece handler aynı form'u okur ve body
iki kez okunmaz. Tek başına `ParseMultipartForm` yerine bunu kullanın. O, URL-encoded
bir body'de okuma hatasını, sınırı aşan bir body'ninkini de, yok sayar.

```go
func (p *Plugin) OnBeforeAction(_ context.Context, ev *collage.BeforeActionEvent) error {
	form, err := ev.Form()
	if err != nil {
		return err // past the action's limit: 413
	}
	if form.Get("website") != "" {
		ev.Result = &collage.ActionResult{Status: http.StatusBadRequest}
	}
	return nil
}
```

`ev.Result` ayarlanırsa request, handler onu döndürmüş gibi onunla cevaplanır ve
dispatch durur. Sonraki hiçbir plugin ve handler çalışmaz. Dönen bir hata,
request'i `"before_action"` altında başarısız kılar. Hata `*http.MaxBytesError`
sarıyorsa `413`, aksi halde `500` olur.

### AfterRenderHook

```go
type AfterRenderEvent struct {
	Page     *collage.Page
	Locale   string
	Degraded bool   // some fragment failed, fallback or not
	Static   bool   // rendered for a static build, not for a request
	HTML     []byte // replace it to post-process the page
	// Fragments: each fragment's time and failure, as collage.FragmentReport
	// DependencyTags: the tags the render depended on
	// Values: the render's shared values, read with a collage.Key's In
	// Findings: what ev.Warn and ev.Error reported so far
}
```

Bir page render'ı başarıyla bittikten sonra çalışır. Page'ler, error page'ler, bir
action'ın `RenderPage` ile cevap olarak döndüğü page ve static build'ler için
çalışır. Çıktıyı post-process etmek için `ev.HTML`'i değiştirin. Orada bıraktığınız içerik sunulan
içeriktir. Bir cache-write hook'u yazmayı atlamadıkça, cache'e yazılan içerik de
odur. Sonraki plugin'ler, öncekilerin ürettiği içeriği görür.

`ev.Values` (bir `*collage.RenderValues`), page'in fragment'lerinin key'ler
üzerinden paylaştığı değerlerdir. Yani page'in *neyden* oluşturulduğunu gösterir.
Markup'ı geri parse etmek yerine doğrudan makalenin kendisini isteyen bir plugin
bunu kullanır. Hook'un bir `RenderContext`'i yoktur. Bu yüzden değerleri key'in
`In` metoduyla okur:

```go
var ArticleKey = collage.NewKey[Article]("jsonld:article")

func (p *Plugin) OnAfterRender(_ context.Context, ev *collage.AfterRenderEvent) error {
	article, ok := ArticleKey.In(ev.Values)
	if !ok {
		return nil // this page stored no article
	}
	ev.Hoist("head", "jsonld", p.script(article))
	return nil
}
```

Değerler opaktır. Bir plugin yalnızca key'ine sahip olduğu değeri okur; değerleri
listelemenin bir yolu yoktur. Uygulamanın kendisine bir şey vermesini isteyen bir
plugin key'i export eder, uygulama da değeri onun altında saklar. Kendi render
fonksiyonlarından hook'una state taşıyan bir plugin ise key'i export etmez.
Framework oraya hiçbir şey koymaz. Değerler render'ın kendisine aittir, bir kopya
değildir. Hook bittikten sonra onları elde tutarsanız request state'ini tutmuş
olursunuz.

Key'in adına plugin'in kendi adıyla başlayın: `"jsonld:article"`,
`"validate:errors"`. Key yalnızca ismi ve tipinden ibarettir. Bu yüzden her biri
`NewKey[string]("errors")` tanımlayan iki plugin aynı değeri paylaşır ve onu en son
set eden kazanır. v0.50.0'dan önce bunun yerine render'ın paylaşılan map'i
`ev.Data` vardı.

`ev.Fragments` ve `ev.DependencyTags` (v0.24.0'dan beri), render'ın nasıl geçtiğini
bildirir. Bir development aracının page'in yanında göstermesi içindir. Her
`collage.FragmentReport` şunları taşır: fragment'in adı `Name`; slot'larındaki
fragment'ler dahil süresi `Duration`; bir fallback yerine geçse bile render'ının
başarısız olduğunu söyleyen `Failed`; `UsedFallback`; ve neyle başarısız olduğunu
söyleyen `Err`. `DependencyTags`, render'ın bağlı olduğu tag'lerdir. İkisi de
event'in kendi kopyalarıdır.

Bulunduğu yerin iki sonucu vardır:

- **Cache hit'te tekrar çalışmaz.** Cache'e yazılan, zaten onun çıktısıdır. Her
  request'te çalışması gereken bir hook, cache'lenen bir page ile birlikte
  kullanılamaz.
- **Boş bir sonuç hata sayılır.** Dispatch'ten sonra `ev.HTML` boşsa, request boş
  bir page sunmak yerine 500 ile başarısız olur.

Bir action'ın `RenderPage` ile cevap olarak döndüğü page de bu hook'u çalıştırır.
Bu, v0.10.0'dan beri böyledir; öncesinde yalnızca `BeforeRender` çalışıyordu. Böylece
bir validation page'i de diğer page'ler gibi minify edilir. Dönen bir hata,
request'i `"after_render"` altında 500 ile başarısız kılar.

### Render'dan sonra head'e eklemek

Page'in neye ihtiyaç duyduğunu yalnızca bitmiş markup'tan öğrenen bir plugin
vardır: highlight ettiği kod blokları bir stylesheet ister. Böyle bir plugin bunu
`OnAfterRender` içinde `ev.Hoist(area, key, html)` ile ekler (v0.25.0'dan beri):

```go
if highlighted {
	ev.Hoist("head", "highlight:css", `<link rel="stylesheet" href="/_highlight/style.css">`)
}
```

Eklenen içerik, layout'un `{{hoist "head"}}` koyduğu yere, render'ın orada
tanımladıklarının arkasına yerleşir. Render'ın o alanda zaten tanımladığı bir
key'e dokunulmaz, çünkü page sonradan gelen bir plugin'den daha özeldir. İki kez
hoist edilen bir key'e de dokunulmaz. `ev.Hoist`, HTML'i ekleyip eklemediğini
döner.

Layout'un yerini yalnızca render'ın ürettiği HTML'de bulur. Daha önceki bir plugin
`ev.HTML`'i değiştirdiyse (ya da aynı hook'ta siz değiştirdiyseniz), `"head"`
bunun yerine `</head>`'in önüne yerleşir; başka bir alan ise false döner. Önce
hoist edin, sonra yeniden yazın.

### Çıktıyı denetlemek: finding'ler

Bir page'in render ettiği çıktıyı denetleyen bir plugin, örneğin atlanmış bir
heading seviyesini, `alt`'ı olmayan bir görseli ya da label'ı olmayan bir form
alanını bulan bir plugin, render'ı başarısız kılmak yerine bulduğunu raporlar
(v0.21.0'dan beri):

```go
func (p *Plugin) OnAfterRender(_ context.Context, ev *collage.AfterRenderEvent) error {
	if !bytes.Contains(ev.HTML, []byte("<h1")) {
		ev.Error("one-h1", "the page has no <h1>")
	}
	return nil
}
```

`ev.Warn(rule, message)` bir `collage.FindingWarning` raporlar; bu, düzeltmeye
değer ama hiçbir şeyi durdurmaz. `ev.Error` ise bir `collage.FindingError`
raporlar. Bir finding'in (`collage.Finding`) `Level`'ı, onu bulan `Rule`, bir
`Message`, `Plugin` ve page'in `Path`'i vardır. Son ikisini framework doldurur.
Finding'in nereye gideceği, page'in nerede render edildiğine bağlıdır:

- **Development'ta** page'in üzerinde, başarısız bir fragment'in kullandığı panelde
  gösterilir. Page olduğu gibi sunulur.
- **Static build'de** raporda, ilgili olduğu page'in altında listelenir
  (`BuildReport.Findings`). Error seviyesindeki bir finding build'i
  `collage.ErrBuildFindings` ile başarısız kılar; page'ler her durumda yazılır.
  Bkz. [Static export](/docs/static-export#reading-the-report).
- **Production'da** onunla hiçbir şey yapılmaz. Her render'da yeniden çalışan bir
  denetim, sunucunun zamanını build'in zaten bildiği şeye harcar. Bu yüzden
  denetleyen bir plugin orada kendini kapatır: `ev.Static` bir render'ın static
  build'e ait olduğunu, `host.DevMode()` ise sunucunun development sunucusu
  olduğunu söyler.

```go
if !ev.Static && !p.dev { // p.dev from host.DevMode() in Init
	return nil
}
```

Tek bir render'ın söyleyemeyeceği şeyler, örneğin aynı title'ı taşıyan iki page ya
da build'in yazmadığı bir page'e giden bir link,
[`OnBuildFinished`](#buildfinishedhook) içinde denetlenir.
[elagoht/htmlcheck](/docs/plugins#elagohthtmlcheck) ikisinin üzerine kurulmuş,
denetleyen bir plugin'dir.

### Her response'u yeniden yazmak: `PersonaliseHook`

`OnAfterRender` cache'ten önce, yalnızca bir kez çalışır; bu yüzden her okuyucuya
kendine özel bir şey veremez. Bunu yapması gereken bir plugin, örneğin bir
Content-Security-Policy nonce'u ya da ziyaretçiye özel bir token ekleyen,
`PersonaliseHook`'u implement eder (v0.43.0'dan beri). Hook, bir okuyucuya giden her
HTML response'u için çağrılır:

```go
func (p *Plugin) OnPersonalise(_ context.Context, ev *collage.PersonaliseEvent) error {
	if !bytes.Contains(ev.Body, []byte(p.marker)) {
		return nil
	}
	nonce, err := newNonce()
	if err != nil {
		return err
	}
	ev.Body = bytes.ReplaceAll(ev.Body, []byte(p.marker), []byte(nonce))
	ev.Header.Set("Content-Security-Policy", "script-src 'nonce-"+nonce+"'")
	ev.Personal = true
	return nil
}
```

Event şunları taşır:

- `ev.Request`: okuyucunun kendi request'i; paylaşılan bir render'ın sadeleştirilmiş
  request'i değil.
- `ev.Header`: response header'ı; body ile uyuşması gereken her şeyi burada ayarlayın.
- `ev.Body`: yazılacak olan body; okuyucunun forgery token'ı zaten içindedir. Hook
  onu değiştirir (yeni bir slice atar); içine asla yazmamalıdır: slice cache'in
  olabilir ve eşzamanlı okuyucularla paylaşılıyor olabilir.
- `ev.Personal`: body bu okuyucuya özel hale geldiğinde ayarlayın.

Forgery token paylaşılan body'ye girdikten sonra, her middleware'in içinde, yani
sıkıştırmadan önce ve dev reload script'inden önce çalışır. Bir page'de dev
overlay'inden önce çalışır; bir hata sayfasında ya da bir action'ın cevap olarak
verdiği bir page'de ise overlay body'de zaten vardır.
Plugin'ler register sırasına göre çalışır ve her biri öncekinin body'sini görür. Bir
page'i (cache'ten ya da yeni), bir fragment path'ini ya da fragment read'ini, bir
action'ın HTML cevabını ve bir hata sayfasını kapsar; document'ı asla.

`Personal`, response'u bir forgery token'ın yaptığı gibi ele alır: ETag'i gönderilen
body'den yeniden hesaplanır, page `private, no-store` cevabı verir; ne cache'ten
gelen bir page ne de bir fragment read, `If-None-Match: *`'a bile, `304` alır. Fragment read yeniden
hesaplanan ETag ile `private, no-cache`, action'ın HTML'i ise `private, no-store`
cevabı verir; handler'ın kendi ayarladığı `Cache-Control` da geçersiz kalır.
elagoht/compress gibi sıkıştırılmış body'leri ETag başına saklayan bir middleware
kişisel bir response'tan bir şey kazanmaz: ETag'i her seferinde yenidir.
elagoht/compress onu her okuyucu için sıkıştırır ve hiçbir şey saklamaz (v0.1.3'ten
beri). `Personal` olmadan değişen bir
body da gönderilen byte'ları adlandıran bir ETag alır. Hata dönen ya da panic eden
bir hook log'lanır; page, fragment ya da action 500 cevabı verir, hata sayfasında ise
o status'un yerleşik sayfası gönderilir. Böyle bir plugin'i olmayan bir site eskisiyle
aynı response'ları gönderir.

### DocumentRenderedHook

```go
type DocumentRenderedEvent struct {
	Document    *collage.Document
	ContentType string
	Locale      string
	Path        string
	Body        []byte // replace it to transform the document
}
```

[Document'lar](/docs/documents) için `AfterRenderHook`'un karşılığıdır: sitemap'ler,
feed'ler, JSON. Bu hook olmasaydı, çıktıyı post-process eden bir plugin page'leri
kapsar, geri kalan her şeyi sessizce atlardı. ETag hesaplanmadan ve body cache'e
yazılmadan önce çalışır. Yani sizin ürettiğiniz içerik saklanır ve ETag de onu
tanımlar. Bir body'ye dokunup dokunmayacağınıza karar vermek için `ContentType`'a
bakın. Dispatch'ten sonra body boşsa ya da bir hata dönerse, request 500 ile
başarısız olur.

Bir document, `OnPageResolved`, `OnBeforeRender` ya da `OnAfterRender` dispatch
etmez. Çünkü page'i yoktur ve template render etmez.

### CacheWriteHook

```go
type CacheWriteEvent struct {
	Key  string        // the cache key; changing it changes nothing
	Page *collage.Page // nil for a document
	TTL  time.Duration // may be adjusted
	Tags []string      // may be adjusted
	Skip bool          // set true to suppress the write
}
```

Render edilmiş bir page ya da document saklanmadan önce çalışır. **Document için
`Page` `nil`'dir.** Onu kontrol etmeden dereference eden bir hook, her document
request'inde panic'e düşer. Panic kontrol altına alınır, ama hook'un dispatch
edildiği yazma işlemi bırakılır. Sonuçta document hiçbir zaman cache'lenmez:

```go
func (p *Plugin) OnCacheWrite(_ context.Context, ev *collage.CacheWriteEvent) error {
	if ev.Page == nil {
		return nil // a document; Key, TTL and Tags are still valid
	}
	if ev.Page.Name == "home" {
		ev.TTL = time.Minute
	}
	return nil
}
```

Bir hata ya da `Skip`, yazmayı engeller; request yine de başarılı olur. Page zaten
render edilmiştir. Onu cache'lemeden sunmak, bir cache sorununu 500'e çevirmekten
iyidir. Hata, error hook'larına `"cache_write"` altında raporlanır.

### CacheInvalidateHook

```go
type CacheInvalidateEvent struct {
	Tags  []string
	Paths   []string           // the URL paths of the cached entries dropped, sorted
	Entries []InvalidatedEntry // the same, with each one's host (since v0.42.0)
}

type InvalidatedEntry struct {
	Host string
	Path string
}
```

`InvalidateTags` bazı tag'lerin entry'lerini düşürdükten sonra çalışır. Onu uygulamanın,
bir action'ın ya da bir plugin'in çağırması fark etmez. Hook bir request'ten değil,
o çağrının içinden dispatch edilir. Dönen bir hata, `InvalidateTags`'in döndüğü
hatayla join edilir. Invalidation'ı kendiniz tetiklemek için `Host.InvalidateTags`'i
çağırın.

`Paths` (v0.23.0'dan beri), invalidation'ın düşürdüğü cache'lenmiş page ve
document'ların URL path'lerini listeler. Bir CDN'in purge etmesi ve bir arama
motoruna değiştiği bildirilmesi gereken budur. Cache'lenmemiş bir page burada asla
yer almaz, çünkü ondan düşürülen bir şey yoktur. Cache'lenen her entry ayrıca
`collage.PathTag(path)` tag'ine de bağlıdır. Bu yüzden tag'i değil path'i bilen bir
plugin, oradaki cache'i `host.InvalidateTags(ctx, collage.PathTag("/blog"))` ile
düşürür. Bkz. [Caching](/docs/caching#invalidating-by-path).

`Entries` (v0.42.0'dan beri) aynı listedir, ama her entry'nin host'uyla birlikte ve
önce host'a, sonra path'e göre sıralıdır. İki host altında cache'lenmiş bir path iki
entry'dir ve request bir port içeriyorsa host portunu korur. Bir CDN'i purge eden ya
da bir arama motoruna haber veren plugin, her URL'yi doğru origin'e göre kurmak için
bunu okur; `Paths` değişmedi. Host, `collage:path:`'in yanında bir
`collage:host:<host>` dependency tag'i olarak kaydedilir. Host tag'leri tracker'ın
`MaxKeysPerTag` sınırından muaftır; `MaxKeys` hâlâ belleği sınırlar.

### Bir host'un origin'i: `OriginResolver`

Tek bir process'ten birden fazla siteye hizmet veren, yani müşteri başına bir host
kullanan bir plugin, `OriginResolver`'ı implement ederek bir host için mutlak
URL'lerin hangi public origin'e göre kurulacağını söyler (v0.42.0'dan beri):

```go
func (p *Plugin) Origin(ctx context.Context, host string) (string, bool) {
	origin, ok := p.origins[host]
	return origin, ok
}
```

Host küçük harfe çevrilmiş, portu atılmış olarak gelir; bir IPv6 literal'i köşeli
parantezsiz (`::1`) gelir. Register sırasında `ok` dönen ilk plugin kazanır. Bir
origin, şeması `http` ya da `https` olan yalın bir `scheme://host[:port]` olmalıdır;
bunu `collage.ParseOrigin` denetler ve normalize eder. Origin bir host da
adlandırmalıdır (`https://:8080` adlandırmaz). Geçersiz bir origin bilinmiyor sayılır
ve development modu bunu plugin ve origin başına bir kez log'lar. Hiçbir plugin'in
tanımadığı host `Config.BaseURL`'ü alır; panic eden bir resolver log'lanır ve
atlanır. Uygulama kodu sonucu `collage.BaseURL(rc)` ile okur. Host cache key'in
içinde olduğu için bu, cache'lenmiş bir render'da güvenlidir; request'i olmayan bir
static build de `Config.BaseURL`'ü alır. Resolver'ı olmayan bir site hiçbir değişiklik
görmez.

Mutlak URL'leri kendisi kuran bir plugin aynı cevaba, aldığı `Host` üzerinde type
assertion ile `Origins` capability'si üzerinden ulaşır. `Host`'a bunun için bir
metot eklenmez:

```go
if o, ok := host.(collage.Origins); ok {
	origin := o.OriginFor(ctx, "acme.test") // a resolver's, else Config.BaseURL
	varies := o.Dynamic()                   // is a resolver registered?
}
```

Origin'leri kendi config'inde alan bir plugin, kuralın bir kopyası yerine bunları
`collage.ParseOrigin` ile doğrular. `collage.ErrInvalidBaseURL`, `plugin.ErrInvalidOrigin`
ile aynı hatadır, dolayısıyla `errors.Is` ikisini de eşler. Yalnızca bir context'i
olan kod, `collage.Vary` ile bildirilmiş bir header'ı `collage.VariedContext(ctx, header)`
üzerinden okur.

### ErrorHook

```go
type ErrorEvent struct {
	Err     error
	Page    *collage.Page // nil unless the failure was a page's own; see below
	Path    string
	Stage   string
	Status  int           // the HTTP status the failure is answered with; 0 for none
	Request *http.Request // the reader's request, read-only; nil outside a request
}
```

Bir request sunulurken bir hata oluştuğunda çalışır. Hata bir page'de, bir
document'ta, bir action'da, bir mount'ta ya da `App.Handle` ile register edilmiş bir
handler'da olabilir. `Stage`, hatanın nerede olduğunu söyler.

`Page`, yalnızca routing'in resolve ettiği bir page'in kendi hatasında doludur.
Hiçbir page resolve edilmediğinde `nil`'dir. Document, action, mount ve
`App.Handle` handler'ı için de `nil`'dir; action'ın `RenderPage` ile cevap olarak
döndüğü page de buna dahildir. Bunları birbirinden ayırmak için `Path`'e bakın ve
`Page`'i her kullanışınızda önce kontrol edin. Framework'ün kullandığı stage'ler
şunlardır: `"route"`, `"not_found"`, `"guard"` (v0.28.0'dan beri: bir page'in
guard'ı başarısız olmuş ya da yazılamayacak bir kararla cevap vermiştir),
`"page_resolved"`, `"before_render"`,
`"render"`, `"after_render"`, `"cache_write"`, `"error_page"`, `"asset"`,
`"handler"` ve `"panic"`. Bu küme kapalı bir enum değildir.

Alarm kurmaya değer olan `"error_page"`'dir. Bu stage, hataları raporlayan page'in
kendisinin başarısız olduğunu gösterir. Client yine de makul görünen built-in bir
page alır. Bu yüzden başka türlü kimse durumu fark etmez.

`Err`'i `errors.Is` ile sınıflandırın. Hiçbir şeyle eşleşmeyen bir URL için
`collage.ErrNoRoute`, var olmayan içerik için `collage.ErrNotFound`, 405 için
`collage.ErrMethodNotAllowed` kullanılır. Reddedilen bir form gönderimi için
`collage.ErrCSRFMissing` ve benzerleri, 4xx ya da 5xx dönen bir mount için
`collage.ErrAssetFailed`, recover edilmiş bir panic için `collage.ErrPanic` vardır.
Geri kalanlar [Hatalar](/docs/errors#reported-to-error-hooks) sayfasında listelenir.

Bir request sunulurken recover edilen panic (stage `"panic"`) v0.45.0'dan beri bir
`*collage.PanicError`'dır. `errors.Is(err, collage.ErrPanic)` yine eşleşir;
`errors.As` panic değerine ve panic'in oluştuğu stack'e ulaşır, kod bir error ile
panic ettiyse `errors.Is` değerin kendisine ulaşır. Stack artık error'ın mesajında
değildir; yalnızca framework'ün `collage: request failed` log satırındaki ayrı bir
`stack` attribute'unda bulunur. Stack'i mesajda arayan kod bunun yerine
`PanicError.Stack`'i okumalıdır.

`Status` (v0.45.0'dan beri), hatanın cevaplandığı HTTP status'tür: var olmayan bir
asset için 404, başarısız bir render ya da panic için 500. Hata için hiçbir cevap
yazılmıyorsa 0'dır; page sunulduktan sonra başarısız olan bir cache yazımı gibi.
`"error_page"` event'i her zaman 500 der. `Request`, okuyucunun canlı request'idir ve
yalnızca okumak içindir: URL'si, method'u ve header'ları; body'si daha önce okunmuş
olabilir. Hata bir request'ten gelmediyse `nil`'dir. İhtiyacınız olanı `OnError`
sırasında okuyun; saklamayın ve başka bir goroutine'e vermeyin. İkisi birlikte, bir
plugin'in kendi middleware'ine ihtiyaç duymadan 5xx'i 4xx'ten ayırmasını ve URL'yi
adlandırmasını sağlar.

Bir page'in başarısız cache yazımı (`"cache_write"`), aynı cache key'ini bekleyen
bütün okuyucuların paylaştığı render'ın içinde olur. Bu yüzden onun `Request`'i
belirli bir okuyucunun değil, o paylaşılan render'ın sadeleştirilmiş request'idir:
cookie yoktur, yalnızca middleware'in `collage.Vary` ile bildirdiği header'lar vardır,
`RemoteAddr` yoktur. Bir document'ın cache yazımı ise paylaşılan render'ı başlatan
okuyucunun, yani cache'te bulamayanın request'ini taşır.

`OnError`'dan dönen bir hata log'lanır ve yutulur. Kalan plugin'ler event'i yine de
alır. Başarısız olan bir error handler, yeni bir error handling turu başlatmamalıdır.

### BuildFinishedHook

```go
type BuildFinishedEvent struct {
	OutDir string              // the directory the build wrote into
	Files  []collage.BuiltFile // every file it wrote, in no particular order
	// Findings: what ev.Warn and ev.Error reported so far
}

type BuiltFile struct {
	Kind   string // "page", "document" or "asset"
	Name   string // the page's or document's name; empty for an asset
	Locale string
	Path   string // the URL path the file answers
	File   string // its absolute path on disk
}
```

Bir static build her page'i, document'ı ve asset'i yazdığında bir kez çalışır
(v0.21.0'dan beri). Page'ler arası denetimler içindir. Bir dosyanın içeriği
gerektiğinde onu `os.ReadFile` ile okuyun. `ev.Warn(path, rule, message)` ve
`ev.Error`, `path`'teki page hakkında bir finding raporlar; `path` boşsa finding
build'in bütünü hakkındadır. Hook'tan dönen bir hata da build'i başarısız kılar;
zaten yazılmış dosyalar yerinde kalır. Bir sunucuda hiçbir zaman çalışmaz.

### Dispatch kuralları

- Hook'lar **register sırasına göre** çalışır.
- Her çağrı **panic'e karşı korunur**. Panic'e düşen bir hook, hata dönmüş bir hook
  gibi başarısız sayılır; process'i çökertmez.
- `OnPageResolved`, `OnBeforeRender`, `OnAfterRender` ve `OnDocumentRendered` için
  **ilk hata dispatch'i durdurur** ve request'i başarısız kılar.
- `OnPersonalise` için ilk hata dispatch'i durdurur; page, fragment ya da action 500
  cevabı verir, hata sayfası ise o status'un yerleşik sayfasına döner.
- `OnBeforeAction` için ilk hata dispatch'i durdurur ve request'i başarısız kılar.
  `ev.Result`'ı ilk ayarlayan plugin de dispatch'i durdurur ve request'i cevaplar.
- `OnCacheWrite` için ilk hata dispatch'i durdurur ve yazmayı engeller.
- `OnCacheInvalidate` için ilk hata dispatch'i durdurur ve `InvalidateTags`'ten
  döner.
- `OriginResolver` için, geçerli bir origin'le `ok` dönen ilk plugin kazanır; panic
  eden bir resolver log'lanır ve atlanır.
- `OnError` için hatalar log'lanır ve dispatch devam eder.
- `OnRequest` hata dönmez. Panic'e düşen bir `OnRequest` log'lanır ve atlanır;
  request ondan önceki context ile devam eder.
- `OnBuildFinished` için ilk hata dispatch'i durdurur ve build'i başarısız kılar.
- Bir finding hata değildir: dispatch'i hiçbir zaman durdurmaz ve sonraki her
  plugin yine çalışır.

## Template fonksiyonları

Bir plugin, template fonksiyonunu `Configure` içinden ekler:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	return host.AddTemplateFunc("readingTime", func(words int) string {
		return fmt.Sprintf("%d min read", max(1, words/200))
	})
}
```

Bundan sonra her template `{{readingTime .Words}}` çağırabilir. Fonksiyon,
`html/template`'in function map'inde kabul ettiği herhangi bir değer olabilir.

- İki kez eklenen bir ad `ErrDuplicateTemplateFunc` hatasıdır. Adı iki plugin de
  eklemiş olabilir, aynı plugin iki kez de. Hatayı ikinci `AddTemplateFunc` çağrısı
  döner. `Configure`'ınız bu hatayı dönerse `New` başarısız olur; hatayı yukarı
  iletirseniz olan da budur. Uygulama bunu `Template.Funcs` ile çözemez. Çakışma
  plugin'ler arasındadır ve birinin geri adım atması gerekir.
- Bunun dışında **uygulama kazanır**. `Config.Template.Funcs` içinde bir plugin'in
  eklediği adla bir kayıt varsa, plugin'in fonksiyonunun yerini alır. Uygulama
  ikisini de görebilir ve kararı o verir.
- Built-in bir adla eklenen plugin fonksiyonu, built-in fonksiyonun yerini alır. Her
  render'da bağlanan fonksiyonlar bunun istisnasıdır (`slot`, `hoist`, `asset`,
  `stylesheet`, `csrfToken`, `pageURL`, `pageURLIn`, `localeURL`). Render engine
  bunları her seferinde yeniden bağlar. Bkz.
  [Template fonksiyonları](/docs/template-functions).
- `AddTemplateFunc`, `Configure` içinden çağrılmalıdır. Sonradan eklemenin bir yolu
  yoktur, çünkü parse işleminden sonra eklenen bir fonksiyonu hiçbir template
  çağıramaz.

`AddTemplateFunc`'ın fonksiyonu, uygulamanın ömrü boyunca tek bir değerdir.
v0.21.0'dan beri `AddRenderFunc` ise bir factory alır. Factory her render için o
render'ın `*RenderContext`'iyle çağrılır. Böylece döndürdüğü fonksiyon, o render'ın
taşıdığı şeyleri okuyabilir: bir `BeforeRender` hook'unun ayarladığı nonce ya da
render'ın locale'i gibi:

```go
var nonceKey = collage.NewKey[string]("csp:nonce")

host.AddRenderFunc("nonce", func(rc *collage.RenderContext) any {
	nonce, _ := nonceKey.Get(rc)
	return func() string { return nonce }
})
```

`AddTemplateFunc`'ın kurallarına uyar: `Configure` içinden çağrılır ve başka bir
plugin'in eklediği bir ad `ErrDuplicateTemplateFunc` hatasıdır.

## Mount'ları wrap etmek

`WrapMount`, mount edilen her dosya sistemine uygulanacak bir fonksiyonu register
eder:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	host.WrapMount(func(fsys fs.FS) fs.FS {
		return minifyingFS{inner: fsys} // your own fs.FS
	})
	return nil
}
```

Response'u değil de dosya sistemini wrap etmesinin sebebi şudur: mount'lar
`http.ServeContent` üzerinden sunulur ve bu fonksiyon `Range`, `If-Range` ve partial
response'ları destekler. Byte'ları her response'ta değiştirmek bütün offset'leri
kaydırır. Bu durumda bir range request'i, ilan edilen uzunluğu artık tutmayan bir
dosyanın yanlış parçasını döner. Dosyaları *zaten* dönüştürülmüş hâlde sunan bir
wrapper ise bu hesabı doğru tutar. Wrapper'lar register edildikleri sırayla çalışır;
`nil` bir wrapper yok sayılır.

## Page, document ve mount eklemek

Bir plugin, `Init` içinden kendi route'larını ekleyebilir. Bunun için
`Host.RegisterPage`, `Host.RegisterDocument` ve `Host.Mount` kullanılır. Bu
route'lar, bir uygulamanın kullandığı builder'larla oluşturulur. `Host.Handle` ise
page olmayan şeyler için, örneğin bir event stream ya da bir WebSocket için, bir
prefix altında düz bir `http.Handler` sunar. `/` ile biten bir prefix, altındaki her
path'i üstlenir. v0.24.0'dan beri sonunda `/` olmayan bir prefix, örneğin
`/metrics`, tek bir tam path'tir. `Host.Use` (v0.21.0'dan beri),
`App.Use` gibi her request'i, plugin'in register edildiği yerde sarmalar.
v0.38.0'dan beri `Config.Plugins` içindeki bir plugin, uygulamanın kendi
middleware'inin dışındadır. Böylece request context'ine koyduğu şey, örneğin bir
session, uygulamanın middleware'i tarafından okunabilir; header'ları ve
cookie'leri de o middleware'in ürettiğini sarar. `App.RegisterPlugin` ile
register edilen bir plugin ise uygulamanın middleware'leri arasında, çağrının
yapıldığı yerde durur.

Page'lere link veren bir plugin, onların nerede olduğunu tahmin etmez, sorar.
`Host.URL` ve `Host.FragmentURL`, bir path'i `App.URL` ve `App.FragmentURL` gibi
oluşturur. `Host.Locales` varsayılan locale'i ve desteklenen her locale'i döner.
`Host.PageURLs(ctx, name)` ise bir page'in cevap verdiği her URL'yi listeler: her
locale için ve bir pattern'in `WithStaticParams`'ının listelediği her parametre
kümesi için bir `collage.PageURL` (`Locale`, `Path`, `Params`).
`WithStaticParams`'ı olmayan bir pattern'in, bir plugin'in bilebileceği hiçbir
URL'si yoktur. Bir sitemap bundan oluşur. `App.Locales` ve `App.PageURLs` aynı
metotlardır ve uygulamanın kendi kodu içindir.

*Dosya üreten* bir plugin, örneğin yeniden boyutlandırılmış görseller ya da
üretilmiş ikonlar, bu dosyaları bir route yerine bir mount'tan sunmalıdır. Static
build, bütün page'ler render edildikten sonra her mount'u çıktısına kopyalar. Bu
yüzden page'lerin ne istediğini kaydeden bir dosya sistemi, builder'a tam olarak
doğru dosya kümesini verir. Export edilen sitenin arkasında da hiçbir şeyin
çalışmasına gerek kalmaz. Dinamik bir path'teki document ise bu şekilde
listelenemez.

Bir request'e kendisi cevap veren bir plugin, örneğin artık olmayan bir page için ya
da reddettiği bir request için, tek satırlık bir metin yerine sitenin kendi page'iyle
cevap verebilir. `host.ServeStatus(w, r, http.StatusGone)`, 404 ve 410 için
not-found page'ini, diğer her status için error page'ini sunar (v0.24.0'dan beri).
Tarayıcının deploy'lar boyunca sakladıklarını sürümleyen bir plugin, örneğin bir
service worker'ın cache'lerini ya da bir asset'in query string'ini, `host.BuildID()`
okur. Bu, ayarlanmışsa `Config.Cache.Version`, değilse executable'ın bir parmak
izidir. `App.BuildID` ve `App.ServeStatus` aynı metotlardır ve uygulamanın kendi
kodu içindir.

## Komutlar

Bir plugin, `Init` içinden bir komut ekler:

```go
type Command struct {
	Name  string // as typed on the command line
	Usage string // for your program's own help; the framework never prints it
	Short string // one line
	Run   func(ctx context.Context, args []string) error
}
```

`RegisterCommand`, boş bir adı (`ErrEmptyCommandName`) ve başka bir komutun zaten
kullandığı bir adı (`ErrDuplicateCommand`) reddeder. `ErrAppStarted` ile hiçbir
zaman kapanmaz. Diğer `Host` register çağrıları gibi `Init` sırasında çalışır,
startup'tan sonra da çalışmaya devam eder. Ancak `DispatchCommands` çalıştıktan
sonra register edilen bir komutu kimse dispatch etmez.

`Usage` ve `Short` yalnızca veridir. Bunları ne framework ne de `collage` binary'si
yazdırır. Help listesi isteyen bir program, listeyi `app.Commands()` üzerinden
kendisi oluşturur.

`collage` CLI, plugin komutlarını çalıştırmaz, çünkü uygulamanızı hiçbir zaman
yüklemez. Komutları uygulamanın kendi `main`'i `collage.DispatchCommands` ile
dispatch eder. Scaffold edilmiş bir `main.go`, flag'lerden sonra kalan her kelime
için bunu yapar. Böylece plugin'inizin kullanıcısı `go run . <command>` çalıştırır:

```go
flag.Parse()

// ... build app and register everything ...

if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
log.Fatal(app.ListenAndServe())
```

`DispatchCommands` önce uygulamayı başlatır, çünkü komutları register eden
`Init`'tir. Exit code'lar şöyledir: başarı için `0`; startup hatası, çalışıp
başarısız olan bir komut ya da `Run`'ı olmayan bir komut için `1`; nil bir app,
argüman verilmemesi ya da hiçbir komutun sahiplenmediği bir ad (`ErrUnknownCommand`)
için `2`. Plugin'inizin README'sinde, komutlarının uygulama üzerinden çalıştığını
belirtin. v0.10.0'dan önce scaffold edilmiş bir projenin bu bloğu kendisinin
eklemesi gerekir. Bkz. [collage CLI](/docs/cli#plugin-commands).

## Yapılandırma

Bir plugin, `Config.PluginConfig` içindeki kendi bölümünü
`collage.PluginConfig(host, defaults)` ile tipli bir struct'a okur. Bu fonksiyon
`Host`'u da `ConfigHost`'u da kabul eder ve varsayılan değerlerin tipini döner.
Uygulamanın bölümünü varsayılan değerlerin bir kopyasının üzerine decode eder:

```go
type Config struct {
	Generator string `json:"generator"`
	Disabled  bool   `json:"disabled"`
}

func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	cfg, err := collage.PluginConfig(host, Config{Generator: "collage"}) // defaults, overlaid by the application's section
	p.cfg = cfg
	return err
}
```

- **Bölüm yoksa, boşsa ya da `null` ise varsayılan değerler olduğu gibi döner.**
  Böylece "yapılandırılmamış" ile "zero value'ya yapılandırılmış" iki farklı durum
  olarak kalır.
- **Bölüm varsa ama bozuksa bu, plugin'in adını veren bir hatadır**
  (`collage: plugin "acme/stamp" configuration: …`) ve varsayılan değerlerle
  birlikte döner. Operatör bir şey yazmıştır. Sessizce varsayılan değerlerle
  çalışmak bunu gizlerdi.
- Varsayılan değerlerin içindeki bir map ya da slice'a yerinde decode edilir. Bu
  yüzden varsayılan değerleri çağrılar arasında paylaşılan tek bir değer olarak
  değil, çağrının içinde bir literal olarak yazın.
- Bölüm, varsayılan değerlerinizin üzerine `json.Unmarshal` ile decode edilir ve onun
  kurallarına uyar. JSON'daki bir scalar ya da slice, varsayılan değerinizin yerini
  alır; slice'lar merge edilmez. Bir map'e decode edilen JSON object'i, kendi
  kayıtlarını sizin atadığınız map'e ekler ve diğer kayıtları korur. İç içe bir
  struct'a decode edilen JSON object'i yalnızca adını verdiği field'ları atar,
  gerisini varsayılan değerlerinizde bırakır.
- Register edilmiş hiçbir plugin'e karşılık gelmeyen bir key, uygulama için bir startup
  hatasıdır (`ErrUnknownPluginConfig`). Bu yüzden config'inizin adresi tamamen
  `Name`'inizden ibarettir. Onu değiştirmek bir breaking change'dir.

v0.50.0'dan önce bir plugin varsayılan değerlerini bir field'a atar ve
`host.Config(&cfg)` çağırırdı.

README'nizde her key'i, tipini ve varsayılan değerini belgeleyin. `New()`'un yanında
bir `NewWith(Config)` constructor'ı da sunarsanız, bir uygulama sizi Go kodunda da
yapılandırabilir.

## Editör desteği: collage.json

v0.27.0'dan beri bir plugin, module'ünün kökündeki bir `collage.json` ile kendini
editörlere tanıtabilir. Dosya, module'deki diğer dosyalar gibi module ile birlikte
gelir. Bir editör extension'ı onu projenin module grafiği üzerinden, yani
`go list -m -json all` ile bulur. Bu yüzden projenin bağımlı olduğu bir plugin,
yayımlanmış olsun ya da olmasın tanınır. Extension, dosyanın listelediklerini
önerir: imzaları ve dokümantasyonlarıyla template fonksiyonlarını, attribute'ları,
snippet'leri ve plugin'in `plugins-config.json` içindeki bölümünün şemasını.

```json
{
  "$schema": "https://raw.githubusercontent.com/Elagoht/collage-snippets-highlighter/main/schemas/collage-plugin-manifest.schema.json",
  "name": "you/greeting",
  "description": "Greets the reader by name.",
  "templateFunctions": [
    {
      "name": "greet",
      "signature": "greet name",
      "insert": "greet ${1:.Name}",
      "doc": "\"Hello, name\" in the render's locale."
    }
  ],
  "attributes": [],
  "snippets": {
    "greeting": { "language": "html", "prefix": "cgreet", "body": ["<p>{{greet .Name}}</p>"], "description": "A greeting" }
  },
  "config": {
    "type": "object",
    "properties": { "formal": { "type": "boolean", "description": "Use the formal greeting." } }
  }
}
```

| Alan | Ne içindir |
| --- | --- |
| `$schema` | Manifest'in [JSON Schema'sı](https://raw.githubusercontent.com/Elagoht/collage-snippets-highlighter/main/schemas/collage-plugin-manifest.schema.json). Böylece editör, dosyayı siz yazarken denetler |
| `name` | Plugin'in `Name()`'i; config'inin `plugins-config.json` içinde bulunduğu key |
| `description` | Plugin'in ne yaptığı, tek cümleyle |
| `repository` | Plugin'in dokümantasyonunun bulunduğu yer |
| `templateFunctions` | Plugin'in `AddTemplateFunc` ya da `AddRenderFunc` ile eklediği fonksiyonlar. Her birinin `name`'i, bir template'in çağırdığı hâliyle `signature`'ı, completion'ın eklediği metin (`insert`, varsayılan olarak ad) ve Markdown ile yazılmış `doc`'u |
| `attributes` | Plugin'in okuduğu ve tag'lerde önerilen HTML attribute'ları. Her birinin `name`'i, completion'ın eklediği `value` (boolean bir attribute için `null`) ve `doc`'u |
| `snippets` | Ada göre snippet'ler. Her birinin `language`'ı (`html` ya da `go`), `prefix`'i, `body`'si ve `description`'ı vardır |
| `config` | Plugin'in `plugins-config.json` içindeki bölümünün JSON Schema'sı |

`name` dışındaki her alan isteğe bağlıdır. `insert`, `value` ve snippet body'leri VS
Code'un snippet sözdizimini kullanır. Editör, bu fonksiyonlardan hangilerinin
uygulamada gerçekten bulunduğunu ve her page'i, fragment'i, slot'u ve mount edilmiş
dosyayı adıyla [`collage inspect`](/docs/cli#collage-inspect) üzerinden öğrenir.

## Eksiksiz bir plugin

`acme/stamp`, her page'in head'ine generator'ın adını yazar ve aynı adı
template'lere de sunar. Page'leri listeleyen bir komut ekler ve başarısız olan bir
error page'i raporlar. Her iki aşamayı, bir render hook'unu, bir error hook'unu ve
bir komutu kullanır.

```go
// Package stamp names the generator in every page's head, offers the same name
// to templates, and adds a command that lists the application's pages.
package stamp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key of its section in plugins-config.json.
const Name = "acme/stamp"

// Config is the plugin's configuration.
type Config struct {
	// Generator is what the generator meta tag says.
	Generator string `json:"generator"`
}

// Plugin is the stamp plugin. Construct it with New.
type Plugin struct {
	cfg    Config
	logger *slog.Logger
}

// New returns the plugin with its defaults.
func New() *Plugin { return &Plugin{} }

var (
	_ collage.Plugin           = (*Plugin)(nil)
	_ collage.Configurer       = (*Plugin)(nil)
	_ collage.BeforeRenderHook = (*Plugin)(nil)
	_ collage.ErrorHook        = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure runs inside collage.New, before templates are parsed: the one
// moment a template function can still be added.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	cfg, err := collage.PluginConfig(host, Config{Generator: "collage"}) // the defaults
	if err != nil {
		return err // a section that is present but malformed
	}
	p.cfg = cfg
	return host.AddTemplateFunc("generator", func() string { return p.cfg.Generator })
}

// Init runs when the application starts, after it has registered its pages.
func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	p.logger = host.Logger()
	return host.RegisterCommand(collage.Command{
		Name:  "pages",
		Usage: "pages",
		Short: "List every registered page",
		Run: func(_ context.Context, _ []string) error {
			for _, page := range host.Pages() {
				fmt.Println(page.Name)
			}
			return nil
		},
	})
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBeforeRender declares the meta tag before the page renders, at depth zero,
// so a fragment that declares its own generator replaces this one.
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}

// OnError reports the one failure nobody would otherwise notice: the error page
// itself failing.
func (p *Plugin) OnError(_ context.Context, ev *collage.ErrorEvent) error {
	if ev.Stage == "error_page" {
		p.logger.Error("stamp: the error page failed", "path", ev.Path, "error", ev.Err)
	}
	return nil
}
```

Bir uygulama bu plugin'i diğer plugin'ler gibi kullanır:

```go
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{stamp.New()},
	PluginConfig: pluginConfig, // {"acme/stamp": {"generator": "The Wire"}}
})
```

## Yaşam döngüsü

1. **Register.** Plugin'ler `New` içinde `Config.Plugins` ile ya da uygulama
   başlamadan önce `RegisterPlugin` ile register edilir. Uygulama başladıktan sonra
   `RegisterPlugin`, `ErrAppStarted` döner. Başarısız olan bir başlatma denemesinden
   sonra da aynı hatayı döner (v0.12.0'dan beri).
2. **Configure**, `New` içinde, bu metodu implement eden plugin'ler için çalışır.
   Register sırasına göre çalışır ve ilk hatada durur.
3. **Init**, uygulama başlarken register sırasına göre çalışır. Biri başarısız
   olursa startup iptal edilir. O ana kadar initialize edilmiş bütün plugin'ler ters
   sırayla shutdown edilir. Başarısız olan plugin shutdown edilmez, çünkü
   initialize işlemini hiç tamamlamamıştır.
4. **Shutdown**, `App.Shutdown` ile ters register sırasına göre çalışır.
   `ListenAndServe`, `App.Shutdown`'ı `SIGINT` ya da `SIGTERM` geldiğinde çağırır.
   `App.Shutdown`, **register edilmiş her plugin'in** `Shutdown`'ını çağırır; o
   plugin'in `Init`'inin çalışıp çalışmadığı ya da başarılı olup olmadığı fark
   etmez. Hiç başlamamış bir uygulamanın, başlatılması başarısız olmuş bir
   uygulamanın ve başarısız başlatmanın zaten geri aldığı plugin'lerin hepsi bu
   çağrıyı alır. Bu yüzden `Shutdown`, `Init` olmadan ve birden fazla kez
   çağrıldığında güvenli olmalıdır. Biri başarısız olsa bile her plugin sırasını
   alır ve hatalar join edilir. `StreamCloser`'ı implement eden bir plugin'in
   `CloseStreams`'i daha önce, shutdown başlarken çağrılır. Bkz.
   [Stream'ler ve shutdown](#streams-and-shutdown).

   `ListenAndServe` kullanıyorsanız, plugin'ler sunucu request'lerini boşalttıktan
   sonra shutdown edilir. Sunucu bunu bitiremezse, `Server.ShutdownTimeout`
   dolduğunda shutdown edilirler. Bu süre dolduktan sonra hâlâ çalışan bir request
   olabilir. Sunucuyu kendiniz yönetiyorsanız, `App`'in bu sunucudan haberi yoktur.
   Önce sunucunuzu durdurun, sonra `App.Shutdown`'ı çağırın. Aksi hâlde bir plugin,
   hâlâ devam eden bir request'in altından çekilip alınabilir.

## Bir plugin'i test etmek

Bir plugin'i, bir uygulamanın onu kullanacağı şekilde test edin. Plugin'in
`Config.Plugins` içinde olduğu gerçek bir `App` oluşturun. Template'i plugin'i
kullanan bir page register edin. Uygulamayı `httptest` ile `app.Handler()`
üzerinden çalıştırın. Bu yöntemde hiçbir sunucu dinlemez ve hiçbir port seçilmez.

```go
package stamp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"

	"example.com/stamp"
)

// newApp builds a one-page application with the plugin in it.
func newApp(t *testing.T, config string) *collage.App {
	t.Helper()

	root := t.TempDir()
	page := `<html><head>{{hoist "head"}}</head><body>{{generator}}</body></html>`
	if err := os.MkdirAll(filepath.Join(root, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pages", "home.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &collage.Config{
		Template: collage.TemplateConfig{Root: root},
		Plugins:  []collage.Plugin{stamp.New()},
	}
	if config != "" {
		cfg.PluginConfig = map[string]json.RawMessage{stamp.Name: json.RawMessage(config)}
	}
	app, err := collage.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	home := collage.NewPage("home").
		WithContent(collage.NewFragment("home", "pages/home.html").Build()).
		WithPath("en", "/").
		Build()
	if err := app.RegisterPage(home); err != nil {
		t.Fatalf("RegisterPage: %v", err)
	}
	return app
}

func get(t *testing.T, app *collage.App, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d: %s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestStamp_UsesItsDefaults(t *testing.T) {
	body := get(t, newApp(t, ""), "/")
	if !strings.Contains(body, `<meta name="generator" content="collage">`) {
		t.Errorf("no default meta tag in %s", body)
	}
}

func TestStamp_ReadsItsConfiguration(t *testing.T) {
	body := get(t, newApp(t, `{"generator": "my site"}`), "/")
	if !strings.Contains(body, `<meta name="generator" content="my site">`) {
		t.Errorf("no configured meta tag in %s", body)
	}
	if !strings.Contains(body, "<body>my site</body>") {
		t.Errorf("template function not applied in %s", body)
	}
}

func TestStamp_RegistersItsCommand(t *testing.T) {
	app := newApp(t, "")
	if err := app.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	commands := app.Commands()
	if len(commands) != 1 || commands[0].Name != "pages" {
		t.Errorf("commands = %v, want one named pages", commands)
	}
}

func TestStamp_RefusesLateRegistration(t *testing.T) {
	app, err := collage.New(&collage.Config{Template: collage.TemplateConfig{Root: t.TempDir()}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := app.RegisterPlugin(stamp.New()); !errors.Is(err, collage.ErrConfigurerRegisteredLate) {
		t.Errorf("RegisterPlugin: err = %v, want ErrConfigurerRegisteredLate", err)
	}
}
```

Test etmeye değer ve kolayca unutulan birkaç şey:

- **Yapılandırılmamış durum.** Çoğu uygulama sizin plugin'iniz için hiçbir zaman bir
  config bölümü yazmaz.
- **Document'lar**, `OnCacheWrite` ya da `OnDocumentRendered` implement ediyorsanız.
  Bir document register edin ve ona request atın. Böylece `nil` bir `Page`, hiç
  cache'lenmeyen bir document olarak değil, bir testte yakalanır.
- **`-race` ile çalıştırın.** Bir event'in canlı `Page`'i üzerinden yazan bir hook
  bir data race'tir. Bunu race detector raporlar, başka hiçbir şey raporlamaz.
- **Static export**, dosya üretiyorsanız. Bir `t.TempDir()` içine
  `collage.NewBuilder(app, ...)` ile export alın. Bu, export edilen sitede sunulan
  sitedeki her şeyin olup olmadığını gösterir. Bkz. [Test yazmak](/docs/testing) ve
  [Static export](/docs/static-export).
