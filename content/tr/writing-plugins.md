---
description: Plugin sözleşmesi, Host ve ConfigHost'un sundukları, her hook ve neyi değiştirebileceği, testleriyle birlikte eksiksiz bir plugin.
reference: Plugin, Host, ConfigHost, ConfigReader, PluginConfig, Configurer, Command, BeforeRenderHook, BeforeActionHook, BeforeActionEvent, AfterRenderHook, CacheInvalidateHook, FragmentRequest, FragmentRender, HoistItem, StreamCloser, DrainHook, ServeHook, PageURL, FragmentReport, PathTag, Finding, FindingLevel, FindingWarning, FindingError, ErrBuildFindings, BuildFinishedHook, BuildFinishedEvent, BuiltFile, BuiltRedirect, RedirectSource, IsCapture, RequestHook, ErrStreamingBody, RouteInfo, Route, ClientIP, IsHTTPS, RouteKindPage, BuiltPage, StageErrorPage, Duration, Key, NewKey, RenderValues
---

# Plugin yazmak

Plugin, dört metodu olan bir Go tipidir. Geri kalan her şey isteğe bağlıdır: bir
render'a tepki vermek, çıktıyı yeniden yazmak ya da bir template fonksiyonu
eklemek. İstediğiniz hook'un interface'ini implement edersiniz, framework de onu
type assertion ile bulur.

Bu sayfa, bu yüzeyin referansıdır. [Plugin kullanmak](/docs/plugins) ise işin öbür
tarafını anlatır: bir uygulama, sizin yazdığınız plugin'i nasıl register eder ve
nasıl yapılandırır. `collage plugin add`'in plugin'inizi kurabilmesi için
[repository sözleşmesine](/docs/plugin-authors) uyun.

Bir plugin'in hayatı, uygulamanın kurduğu değerden `Shutdown`'ına kadar sürer.
Aşağıdaki her aşamanın kendi bölümü vardır; [Yaşam döngüsü](#lifecycle) onları
sırasıyla listeler.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 506" role="img" aria-labelledby="wp-life-title">
<title id="wp-life-title">Uygulama plugin'i New(Options) ile kurar ve Config.Plugins'e koyar. collage.New, Configure'ı bir ConfigHost ile register sırasıyla çalıştırır: config, logger, template fonksiyonları, mount wrapper'ları. Uygulama page'lerini register eder, ardından ilk Handler, Start ya da build Init'i bir Host ile register sırasıyla çalıştırır: config'i ve page'leri okur; document, mount, handler, middleware ve komut ekler. Hizmet verirken ListenAndServe OnServe'ü çağırır, her request plugin'in middleware'inden ve hook'larından geçer, invalidation ve static build ise OnCacheInvalidate ile OnBuildFinished'i tetikler. Dururken, yalnızca Init'i başarılı olmuş plugin'ler için önce OnDrain, sonra CloseStreams, ardından ters sırayla Shutdown çağrılır.</title>
<defs><marker id="wp-life-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="95" y="8" width="130" height="26" rx="13" /><text x="160" y="25" text-anchor="middle">New(Options{})</text>
<line class="d-flow" x1="160" y1="34" x2="160" y2="50" marker-end="url(#wp-life-arrow)" />
<text class="d-small" x="168" y="46">Config.Plugins</text>
<rect class="d-slot d-slot-a" x="6" y="54" width="308" height="74" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="70">collage.New içinde</text><text class="d-small" x="302" y="70" text-anchor="end">register sırasıyla</text>
<rect class="d-a" x="18" y="78" width="284" height="40" rx="5" /><text class="d-on-a" x="160" y="94" text-anchor="middle">Configure(ctx, ConfigHost)</text><text class="d-on-a d-small" x="160" y="109" text-anchor="middle">config · logger · template func'ları · WrapMount</text>
<line class="d-flow" x1="160" y1="128" x2="160" y2="144" marker-end="url(#wp-life-arrow)" />
<text class="d-small" x="166" y="140">page'ler register edilir</text>
<rect class="d-slot d-slot-b" x="6" y="148" width="308" height="88" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="164">ilk Handler, Start, build</text><text class="d-small" x="302" y="164" text-anchor="end">register sırasıyla</text>
<rect class="d-a" x="18" y="172" width="284" height="54" rx="5" /><text class="d-on-a" x="160" y="189" text-anchor="middle">Init(ctx, Host)</text><text class="d-on-a d-small" x="160" y="204" text-anchor="middle">config · logger · page ve URL'leri okur</text><text class="d-on-a d-small" x="160" y="217" text-anchor="middle">document · mount · Handle · Use · komut ekler</text>
<line class="d-flow" x1="160" y1="236" x2="160" y2="252" marker-end="url(#wp-life-arrow)" />
<rect class="d-slot d-slot-c" x="6" y="256" width="308" height="130" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="272">hizmet verirken</text><text class="d-small" x="302" y="272" text-anchor="end">hook'lar register sırasıyla</text>
<rect class="d-frame" x="18" y="280" width="284" height="24" rx="5" /><text x="160" y="296" text-anchor="middle">OnServe · yalnızca ListenAndServe</text>
<rect class="d-b" x="18" y="312" width="138" height="64" rx="5" /><text class="d-on-b" x="87" y="329" text-anchor="middle">her request</text><text class="d-on-b d-small" x="87" y="344" text-anchor="middle">OnRequest · middleware</text><text class="d-on-b d-small" x="87" y="357" text-anchor="middle">render/action hook'ları</text><text class="d-on-b d-small" x="87" y="370" text-anchor="middle">OnPersonalise · OnError</text>
<rect class="d-c" x="164" y="312" width="138" height="64" rx="5" /><text class="d-on-c" x="233" y="329" text-anchor="middle">request dışında</text><text class="d-on-c d-small" x="233" y="348" text-anchor="middle">OnCacheInvalidate</text><text class="d-on-c d-small" x="233" y="364" text-anchor="middle">OnBuildFinished</text>
<line class="d-flow" x1="160" y1="386" x2="160" y2="402" marker-end="url(#wp-life-arrow)" />
<rect class="d-slot d-slot-a" x="6" y="406" width="308" height="92" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="422">dururken</text><text class="d-small" x="302" y="422" text-anchor="end">App.Shutdown</text>
<rect class="d-frame" x="17" y="432" width="86" height="28" rx="5" /><text x="60" y="450" text-anchor="middle">OnDrain</text>
<line class="d-flow" x1="103" y1="446" x2="115" y2="446" marker-end="url(#wp-life-arrow)" />
<rect class="d-frame" x="117" y="432" width="86" height="28" rx="5" /><text x="160" y="450" text-anchor="middle">CloseStreams</text>
<line class="d-flow" x1="203" y1="446" x2="215" y2="446" marker-end="url(#wp-life-arrow)" />
<rect class="d-a" x="217" y="432" width="86" height="28" rx="5" /><text class="d-on-a" x="260" y="450" text-anchor="middle">Shutdown</text>
<text class="d-small" x="160" y="476" text-anchor="middle">yalnızca Init'i başarılı plugin'ler;</text>
<text class="d-small" x="160" y="489" text-anchor="middle">Shutdown sunucu durunca, ters sırayla</text>
</svg>
<figcaption>Plugin'i kodunuz New(Options) ile kurar ve Config.Plugins'e verir. collage.New, template'ler parse edilmeden önce onun Configure'ını çalıştırır; bir template fonksiyonunun eklenebileceği tek an budur. Ardından uygulama page'lerini register eder ve onu başlatan ilk çağrı (Handler, ListenAndServe, Start, RenderPath, static bir build) Init'i çalıştırır. Init config'i ve page'leri okur; document, mount, handler, middleware ve komut ekler. Başarısız olan bir Init uygulamanın başlamasını durdurur ve ondan önce initialize edilmiş plugin'ler ters sırayla shutdown edilir. Hizmet verirken ListenAndServe OnServe'ü bir kez çağırır, her request plugin'in middleware'inden ve hook'larından geçer; tag'leri invalidate etmek ya da static bir build'i bitirmek kendi hook'unu tetikler. Dururken önce OnDrain, sonra CloseStreams çağrılır, sunucunun request'leri beklenir ve Shutdown ters register sırasıyla çağrılır. Bunların her biri yalnızca Init'i başarılı olmuş plugin'ler için çağrılır.</figcaption>
</figure>

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
- **`Shutdown`**, `Init`'in edindiği her şeyi serbest bırakır. Bir kez ve
  yalnızca `Init`'i başarılı olmuş bir plugin için çağrılır (v0.61.0'dan beri); yine
  de idempotent tutun. Ayrıntılar için
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
bir plugin onları okuyabilir, kendi document ve mount'larını da ekleyebilir.

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
| `collage.FindPlugin[T](host)` | evet | — |
| `Pages`, `Page` | — | evet |
| `URL`, `Locales`, `PageURLs`, `OriginFor`, `MultiOrigin` | — | evet |
| `BuildID`, `ServeStatus` | — | evet |
| `RegisterDocument`, `Mount`, `OpenMounted`, `Handle`, `Use` | — | evet |
| `RenderFragment` | — | evet |
| `RegisterCommand` | — | evet |

`ConfigHost` bilerek daha dar tutulmuştur. `Configure` sırasında uygulama henüz
hiçbir şey register etmemiştir. Page listesi boş olurdu. Sitenin origin'leri de,
aralarında bir `OriginResolver` olan bütün plugin'ler bilinmeden belli olmaz.

v0.78.0, v1'den önce ikisini de daraltıp hiçbir plugin'in kullanmadığını kaldırdı:

- `Host.InvalidateTags` kaldırıldı. Invalidate etmek uygulamanın işidir
  (`App.InvalidateTags`, `ActionResult.InvalidateTags`); bir plugin bunu
  [`OnCacheInvalidate`](#cacheinvalidatehook) ile duyar.
- `ConfigHost.BaseURL`, `OriginFor` ve `MultiOrigin` kaldırıldı. `Host` üçünü de
  tutar: onları `Init`'te okuyun.
- `Init`'e verilen `Host` plugin'leri listelemez, bu yüzden `Init`'ten çağrılan
  `collage.FindPlugin` false döner. Onu `Configure`'dan çağırın ve bulduğunu
  saklayın.
- `ErrorEvent.Page` ve `BeforeActionEvent.Page` kaldırıldı: hook'a framework'ün
  canlı page'i veriliyordu ve onu okuyan yoktu.

### ConfigHost

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın development modunda çalışıp çalışmadığını söyler. |
| `Logger() *slog.Logger` | Uygulamanın logger'ını döner. |
| `ConfigReader` (embed edilmiş) | `collage.PluginConfig(host, defaults)` bu plugin'in config bölümünü bunun üzerinden okur. Bkz. [Yapılandırma](#configuration). |
| `AddTemplateFunc(name, fn) error` | Bir template fonksiyonu ekler. Ad daha önce eklenmişse `ErrDuplicateTemplateFunc` döner; ekleyen başka bir plugin de olabilir, aynı plugin'in önceki bir çağrısı da. Ad bir built-in'inkiyse (v0.77.0'dan beri) ya da template dilininse (v0.78.0'dan beri) de aynı hatayı döner. |
| `WrapMount(wrap func(fs.FS) fs.FS)` | Mount edilen her dosya sistemine uygulanacak bir dönüşümü register eder. Wrapper'lar register edildikleri sırayla uygulanır. |
| `AddRenderFunc(name, factory) error` | Her render için o render'ın `*RenderContext`'inden yeniden üretilen bir template fonksiyonu ekler. Bkz. [Template fonksiyonları](#template-functions) (v0.21.0'dan beri). |

Bir `ConfigHost`, plugin'in `Configure`'ı döndüğünde kapanır (v0.78.0'dan beri).
Onu saklayan bir plugin, sonraki bir `AddTemplateFunc` ya da `AddRenderFunc`
çağrısından `collage.ErrConfigureClosed` alır; sonraki bir `WrapMount` hiçbir şeyi
wrap etmez ve log'lanır; onun üzerinden `collage.FindPlugin` de hiçbir şey bulmaz.
Template'ler yapım bitince parse edilir ve request'ler o ana kadar register
edilenlerle render edilir. Önceden geç gelen bir render fonksiyonu, her render'ın
okuduğu map'e yazılıyordu; Go runtime'ı bunun üzerine process'i sonlandırır.

### Host

| Metot | Ne yapar |
| --- | --- |
| `DevMode() bool` | Uygulamanın development modunda çalışıp çalışmadığını söyler. |
| `Logger() *slog.Logger` | Uygulamanın logger'ını döner. |
| `ConfigReader` (embed edilmiş) | `collage.PluginConfig(host, defaults)` bu plugin'in config bölümünü bunun üzerinden okur. |
| `Pages() []*collage.Page` | Register edilmiş bütün page'leri döner. Her biri bir defensive copy'dir. |
| `Page(name) (*collage.Page, bool)` | Adı verilen page'i defensive copy olarak döner. |
| `RegisterDocument(doc) error` | Plugin'in eklediği bir document'ı register eder. |
| `Mount(prefix, fsys, opts...) error` | Bir dosya sistemini bir URL prefix'i altında sunar. |
| `OpenMounted(urlPath) (fs.File, error)` | Sitenin kendi dosyalarından birini sunulduğu URL ile (`/static/icon.png`), prefix'i altında bulunduğu mount üzerinden açar: development'ta diskten, build edilmiş bir binary'de embed edilmiş kopyadan. Eksik bir dosya `fs.ErrNotExist`'tir (v0.68.0'dan beri). |
| `Handle(prefix, handler) error` | `App.Handle` gibi, bir `http.Handler`'ı `/` ile biten bir URL prefix'i altında ya da sonunda `/` olmayan tek bir tam path'te (`/metrics`, v0.24.0'dan beri) sunar. Örneğin bir event stream ya da bir WebSocket (v0.18.0'dan beri). v0.67.0'dan beri düz bir `App.Handle` mount'unun aksine middleware'in içinde çalışır; v0.78.0'dan beri request body'si, bir page'inki gibi, middleware'den önce `Server.MaxBodyBytes` ile sınırlanır. |
| `RenderFragment(r, req) (*collage.FragmentRender, error)` | Bir page'in `WithFragmentPath` ile açtığı bir fragment'i kısımlar hâlinde render eder. Bkz. [Fragment göndermek](#pushing-fragments) (v0.18.0'dan beri). |
| `RegisterCommand(cmd) error` | Bir komut ekler. Bkz. [Komutlar](#commands). |
| `Use(middleware) error` | Her request'i sarmalar (v0.21.0'dan beri), plugin'in register edildiği yerde: `Config.Plugins` içindeki bir plugin için uygulamanın kendi middleware'inin dışında (v0.38.0'dan beri; önceden içindeydi). v0.67.0'dan beri uygulamanın düz `App.Handle` ile mount ettiği bir handler'a gelen request'leri sarmalamaz. |
| `URL(name, locale, params) (string, error)` | Bir page'in ya da document'ın path'ini `App.URL`'in oluşturduğu gibi döner (v0.21.0'dan beri). |
| `BaseURL() string` | Sonunda slash olmadan `Config.BaseURL`'i döner; ayarlanmamışsa boştur. |
| `OriginFor(ctx, host) string` | Bir host'un public origin'ini döner: bir `OriginResolver`'ınkini, yoksa `Config.BaseURL`'i. Bkz. [Bir host'un origin'i](#which-origin-a-host-has-originresolver) (v0.60.0'dan beri). |
| `MultiOrigin() bool` | Bir `OriginResolver`'ın register edilip edilmediğini, yani origin'lerin host'a göre değişip değişmediğini söyler (v0.60.0'dan beri). |
| `Locales() (default, supported)` | Varsayılan locale'i ve varsayılan dahil desteklenen her locale'i döner (v0.21.0'dan beri). |
| `PageURLs(ctx, name) ([]collage.PageURL, error)` | Bir page'in her locale'de cevap verdiği her URL'yi döner; bir pattern, `WithStaticParams`'ı üzerinden açılır. Bir sitemap'in içeriği budur (v0.21.0'dan beri). |
| `BuildID() string` | Sunan build'i adlandırır: `Config.Cache.Version` ya da executable'ın bir parmak izi. Tarayıcının deploy'lar boyunca sakladıklarını sürümlemek içindir; örneğin bir service worker'ın cache'leri ya da bir asset'in query string'i (v0.24.0'dan beri). |
| `ServeStatus(w, r, status)` | Request'e status'la ve sitenin o status için kendi page'iyle cevap verir: 404 ve 410 için not-found page'i, diğerleri için error page'i. Bir request'e kendisi cevap veren ve cevabı sitenin geri kalanı gibi görünmesi gereken bir plugin içindir (v0.24.0'dan beri). |

v0.60.0, `RegisterPage` ve `FragmentURL`'i `Host`'tan kaldırdı; uygulama ikisini de
`App.RegisterPage` ve `App.FragmentURL` olarak tutar. Bir plugin'in type assertion ile
ulaştığı `collage.Origins` capability'si de `OriginFor` ve `MultiOrigin` metotlarına
dönüştü.

`Init`'e gelen değer `*App` değildir. Yalnızca bu metotları ileten dar bir değerdir.
Bu yüzden bir plugin, type assertion ile `ListenAndServe`'e, `Shutdown`'a,
router'a, cache'e ya da template kümesine ulaşamaz.

`Handle` ve `RenderFragment`
v0.18.0'da eklendi. `Use`, `URL`, `Locales` ve `PageURLs` ise
v0.21.0'da, `ConfigHost`'taki `AddRenderFunc` ile birlikte eklendi. `BuildID` ve
`ServeStatus` ise v0.24.0'da geldi. **Bunların her biri bir test double'ı için
breaking change'dir:** `Host`'u ya da `ConfigHost`'u implement eden bir test
double'ının da yeni metotlara ihtiyacı vardır. v0.24.0'dan beri bunlar `BuildID` ve
`ServeStatus`'tur.

v0.50.0'dan beri ikisi de `collage.ConfigReader`'ı içerir. Onun
tek metodu export edilmemiştir; bu yüzden onu yalnızca collage'ın kendi host'ları
karşılar. Plugin'i bunun yerine gerçek bir uygulama üzerinden test edin. Bkz.
[Bir plugin'i test etmek](#testing-a-plugin).

**`Host`, bir plugin'in neye ulaşabileceğini sınırlar; neyi değiştirebileceğini
sınırlamaz.** `Pages` ve `Page`, page struct'ının bir kopyasını döner. Tuttuğu
her slice ve map (`Paths`, `CacheParams`, `FragmentPaths` ve locale başına map'leri,
`DependencyTags`, `LayoutChain`) ile `Redirect` ve `Action` değerleri de kopyalanır.
Bu yüzden bunları düzenlemek uygulamanın kendi page'ine dokunmaz (v0.78.0'dan beri;
önceden `CacheParams`, `FragmentPaths` ve action'lar paylaşılıyordu, bu yüzden
kopyasına yazan bir plugin cache key'ini ve fragment URL'lerini değiştiriyordu).
Ancak kopyanın içindeki fragment'ler hâlâ paylaşılır. Aşağıdaki event'ler de kopyayı değil, *canlı*
page'i taşır, çünkü her request'te bir page'i ve fragment ağacını kopyalamak hot
path'e pahalıya mal olurdu.

Bir event'in `Page`'i üzerinden yazarsanız, eşzamanlı
bütün request'lerin okuduğu page'i değiştirmiş olursunuz. Bu bir data race'tir ve
`go test -race` bunu raporlar. Page'leri salt okunur kabul edin. Plugin'ler güvenilen
koddur; bir sandbox içinde çalışmazlar.

Bir plugin'in register ettiği document ve mount'lar, uygulamanınkilerle aynı
kurallara tabidir. Zaten kullanılan bir ad ya da path bir startup hatasıdır. Hangi
kaydın kazanacağı register sırasına bırakılmaz.

### Fragment göndermek

`RenderFragment`, fragment'lerin tarayıcı tarafından istenmesini beklemek yerine
onları kendi sahip olduğu bir bağlantı üzerinden gönderen plugin'ler içindir. Bu
bağlantı bir event stream ya da bir WebSocket olabilir. Metot, fragment'in path'ine
gelen bir request'in render edeceği şeyin tam olarak aynısını render eder. Bir
response yerine kısımları döner. Bir fark vardır: `collage.Cached` ile okunan değerler
tazedir. Bir request'e, değer yenilenirken stale hali sunulabilir. Ama push edilen
bir fragment bir daha istenmez. Bu yüzden `RenderFragment` upstream'i bekler ve upstream
başarısız olursa hata döner:

| Field | |
| --- | --- |
| `HTML` | Markup. İçindeki her form okuyucunun forgery token'ını taşır |
| `Head` | Fragment'in marker koymadığı bir alana hoist ettikleri; area'ları ve key'leriyle birlikte `HoistItem` olarak |
| `DependencyTags` | Render'ın bağlı olduğu tag'ler. Neyin gönderileceğini bilmek için bunları `CacheInvalidateEvent.Tags` ile eşleştirin |
| `Shared` | Render bu anda aynı host'taki her okuyucu için aynıdır: page herkes için cache'lenir ya da alt ağaçtaki her handler `Static()` veya `Shared()` olarak tanımlanmıştır ve slot resolver yoktur. Ayrıca form token'ı ve (v0.78.0'dan beri) bir `PersonaliseHook`'un kişisel yaptığı bir şey de yoktur (v0.19.0'dan beri; öncesinde `Static()` olmayan her handler request'i okuyor sayılırdı). v0.71.0'dan beri böyle bir render, page cache'inki gibi, okuyucunun cookie'leri, header'ları ve context değerleri çıkarılmış bir request'ten üretilir |
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
ulaştığı yere ulaşır. v0.78.0'dan beri her çağrı, fragment'in URL'sine gelen bir
request gibi sunulur: ona giden, `r`'nin host'unu, client adresini, header'larını ve
cookie'lerini taşıyan (path değerlerini, body'sini ya da method'unu taşımayan) bir
`GET`, fragment'in path'i için uygulamanın ve her plugin'in middleware'inden, bir
request'teki sırayla geçer. Ardından fragment'in kendi
[guard](/docs/pages-and-layouts#private-pages-guards)'ı ve page'inki gelir
(v0.28.0'dan beri). Bu guard'lar fragment URL'sinin parametrelerini `r.PathValue`
ile okur, çağıranınkileri asla.

- **Bir ret `collage.ErrGuardRefused`'tır**, render değil: reddeden bir guard ya da
  next'i çağırmak yerine cevap veren bir middleware (bir basicauth 401'i, bir
  ratelimit 429'u, bir login redirect'i, bir timeout'un 503'ü). O bağlantıya hiçbir
  şey göndermeyin. Middleware'in yazdığı ve ayarladığı header ve cookie'ler atılır;
  context'e koyduğu değerler guard'a ve render'a ulaşır. Panic'e düşen bir
  middleware, `ErrPanic`'i saran bir hatadır.
- **Çağıranın context değerleri taşınmaz.** Stream'in kendi middleware'inin oraya
  koyduğu şey stream'in path'iyle ilgiliydi. Render, `r` bittiğinde biter; `r`'nin
  taşıdığı bir trace span'i onun parent'ı değildir.
- **Her `PersonaliseHook` çalışır**, markup'ta ve hoist edilen her öğede. Böylece
  bir CSP nonce'u ya da bir honeypot marker'ı bir page'deki gibi doldurulur; bir
  hook'un kişisel yaptığı render `Shared` değildir.
- **Her çağrı middleware'e giden bir request'tir**: bir rate limit her push'u
  sayar.
- Bir plugin'in `Init`'inden, uygulama başlarken çağrılırsa middleware olmadan
  render etmek yerine hata döner.

`Page` ve `Fragment` ile istendiğinde URL, `FragmentURL`'in kurduğu gibi page'in
pattern'inden, istenen locale'de kurulur (v0.78.0'dan beri): hiçbir request'in
taşıyamayacağı bir parametre (tanımlanmamış bir ad, catch-all dışında `/` içeren,
`.`, `..` ya da boş bir değer), fragment'in path'i olmayan bir locale ve bir
request'in temiz yazımına redirect edileceği bir path `ErrUnknownFragmentPath`'tir.

Her okuyucunun middleware'ine ve guard'larına bir değişikliği sormak için onu her
okuyucu için bir kez, o okuyucunun request'iyle çağırın ve her bağlantıya kendisi
için yapılan render'ı gönderin. `Shared` olmayan bir render tek bir okuyucunun
verisini taşıyabilir. `Shared` bir render ise kaç okuyucuya ulaşırsa ulaşsın her
değişiklikte bir render'a mal olur (v0.78.0'dan beri): bir okuyucunun middleware'i
ve guard'ları geçtiğinde, aynı host, URL, locale ve cache'lenen query için daha
önce yapılmış bir `Shared` render, o okuyucunun kendi forgery token'ı ve
`PersonaliseHook` geçişiyle bir kopya olarak verilir; render sürerken gelen
okuyucular onu bekler. Bağlı olduğu bir tag (aralarında URL'sinin `PathTag`'i ve
host'unki) invalidate ya da purge edildiğinde, page'in cache süresi geçtiğinde ve
her request'te render edilen bir page için, cache kapalıyken ya da instance'ların
paylaştığı bir store'la, yapıldıktan bir saniye sonra yeniden yapılır. En çok 4096
render ya da 32 MiB (daha düşükse `Cache.MaxBytes`) tutulur; development'ta hiç
tutulmaz, request'i cache'i atlayan bir okuyucu için de tutulmaz: bir
middleware'in `SkipCache`'i, HTTP'de olduğu gibi, render'ı o okuyucunun kendisine
ait yapar.

Handler'ı belli bir anda herkes için
aynı şeyi döndüren bir fragment (bir ölçüm gibi)
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

Sunucunun trafik almayı bırakmak üzere olduğunu bilmesi gereken bir plugin
(readiness'i false'a çeviren, yeni işleri reddeden bir plugin) `DrainHook`'u
implement eder (v0.53.0'dan beri):

```go
var _ collage.DrainHook = (*Plugin)(nil)

func (p *Plugin) OnDrain() { p.draining.Store(true) }
```

`OnDrain`, drain başladığında bir kez çağrılır:
[`Server.DrainDelay`](/docs/deployment#graceful-shutdown-and-draining) beklenmeden
ve `CloseStreams`'ten önce. `DrainDelay` 0 olduğunda bile çağrılır. Bloklamamalıdır:
drain sürdüğü sürece sunucu hizmet vermeye devam eder ve yavaş bir hook kendinden
sonraki her adımı geciktirir. İçindeki bir panic yakalanır ve plugin'in adıyla Warn
seviyesinde loglanır. `Shutdown` ve `CloseStreams` gibi yalnızca `Init`'i
başarılı olmuş plugin'lere ulaşır (v0.61.4'ten beri).

`Start`'tan önce ya da
başarısız bir `Start`'tan sonra yapılan bir `Shutdown` hiçbirine haber vermez.
v0.61.4'ten önce hiç çalışmamış plugin'lerde de `OnDrain` ve `CloseStreams`
çağrılırdı. Bu yüzden ikinci bir başlatma için yeniden kullanılan bir plugin
(`collagetest.TrySite`'ın ardından aynı plugin değeriyle `NewSite` gibi) daha en
baştan drain ediyor olurdu.

Yalnızca uygulama hizmet verirken iş yapan bir plugin (bir scheduler, bir queue
worker'ı) `ServeHook`'u implement eder (v0.55.0'dan beri):

```go
var _ collage.ServeHook = (*Plugin)(nil)

func (p *Plugin) OnServe(ctx context.Context) { go p.run(ctx) }
```

`OnServe`'ü register sırasına göre, bir kez `ListenAndServe` çağırır: port
bağlandıktan sonra, yani `collage: listening` loglandığında ve hizmet vermeye
başlamadan önce. `Start()` ya da `Handler()` onu çağırmaz. Bu yüzden static bir
build'de, bir plugin komutunda, `app.Handler()` üzerinde kendi `http.Server`'ını
çalıştıran bir uygulamada ve `Shutdown`, `ListenAndServe`'den önce çalıştığında
çağrılmaz. `collage dev`, `ListenAndServe` üzerinden hizmet verir. Bu yüzden
development'ta da, her restart'ta yeniden çağrılır.

`ctx`, `context.Background()`'dan türer ve drain başladığında, `OnDrain`'in
çalıştığı anda iptal edilir. Durdurmanın bir sinyalden mi yoksa
`App.Shutdown(ctx)`'ten mi geldiği fark etmez.

Bu iptalin anlamı "yeni iş başlatma"dır. Zaten
çalışan işi bitirmek plugin'in `Shutdown`'ının işidir. `Shutdown` sunucu durduktan
sonra çalışır ve shutdown ctx'iyle sınırlıdır. `OnServe` çağrıldığında ctx zaten
bitmiş olabilir (bir an önce başlamış bir drain). Bu yüzden iş başlatmadan önce onu
kontrol edin.

`OnServe` bloklamamalıdır: goroutine'leri başlatın ve dönün. İçindeki
bir panic yakalanır ve plugin'in adıyla Warn seviyesinde loglanır. Sunucu yine de
hizmet verir.

## Hook'lar

| Interface | Metot | Event | Ne zaman çalışır | Neyi değiştirebilir |
| --- | --- | --- | --- | --- |
| `RequestHook` | `OnRequest` | `*http.Request` | Her request'te ilk olarak; collage'ın request span'inden, middleware'den ve routing'den önce (v0.25.0'dan beri) | request'in context'ini |
| `BeforeRenderHook` | `OnBeforeRender` | `BeforeRenderEvent` | Yeni bir page render'ından önce | event'te hiçbir şeyi; `ev.Context` üzerinden hoist edebilir |
| `BeforeActionHook` | `OnBeforeAction` | `BeforeActionEvent` | Bir action'ın handler'ından önce, guard'larından, body sınırından ve forgery kontrolünden sonra (v0.31.0'dan beri) | `ev.Result`; handler'ın yerine cevap verir |
| `AfterRenderHook` | `OnAfterRender` | `AfterRenderEvent` | Bir page render'ı başarıyla bittikten sonra. `ev.Values`'u taşır (bir key'in `In` metoduyla okunur) | `ev.HTML`; `ev.Warn` ve `ev.Error` ile raporlar |
| `PersonaliseHook` | `OnPersonalise` | `PersonaliseEvent` | Cache'ten sonra, bir okuyucuya giden her HTML response'u için (v0.43.0'dan beri) | `ev.Body`, `ev.Header`, `ev.Personal` |
| `DocumentRenderedHook` | `OnDocumentRendered` | `DocumentRenderedEvent` | Bir document handler'ı body'sini ürettikten sonra | `ev.Body` |
| `CacheInvalidateHook` | `OnCacheInvalidate` | `CacheInvalidateEvent` | Entry'ler tag ile invalidate edildikten sonra | hiçbir şeyi |
| `OriginResolver` | `Origin` | string olarak host | `collage.BaseURL` ya da `Host.OriginFor` bir host'un origin'ini sorduğunda (v0.42.0'dan beri) | dönen origin |
| `ErrorHook` | `OnError` | `ErrorEvent` | Bir request sunulurken bir hata oluştuğunda | hiçbir şeyi |
| `BuildFinishedHook` | `OnBuildFinished` | `BuildFinishedEvent` | Bir static build her dosyayı yazıp header'larını yakaladığında, bir kez (v0.21.0'dan beri) | `ev.Warn` ve `ev.Error` ile raporlar |

Bir request'te hook'lar bu sırayla çalışır:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 552" role="img" aria-labelledby="wp-req-title">
<title id="wp-req-title">Bir request önce OnRequest'ten, sonra plugin'lerin middleware'inden, sonra app.Use'dan, sonra router'dan geçer. Bir page cache'e bakar: hit doğrudan OnPersonalise'a gider; miss OnBeforeRender'ı, render'ı ve OnAfterRender'ı çalıştırır, bir GET OnPersonalise'dan önce cache'e yazılır. Bir action guard'larından ve forgery kontrolünden geçer, OnBeforeAction'ı ve handler'ını çalıştırır; page ile verilen bir cevap render hook'larını cache'e uğramadan çalıştırır ve OnPersonalise'dan geçer. Bir hata OnError'a ulaşır, bir document OnDocumentRendered'ı çalıştırır.</title>
<defs><marker id="wp-req-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="wp-req-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">request</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="50" y="50" width="220" height="26" rx="5" /><text class="d-on-a" x="160" y="67" text-anchor="middle">OnRequest</text>
<line class="d-flow" x1="160" y1="76" x2="160" y2="88" marker-end="url(#wp-req-arrow)" />
<rect class="d-frame" x="50" y="92" width="220" height="26" rx="6" /><text x="160" y="109" text-anchor="middle">plugin middleware'i</text>
<rect class="d-frame" x="50" y="122" width="220" height="26" rx="6" /><text x="160" y="139" text-anchor="middle">app.Use</text>
<line class="d-flow" x1="160" y1="148" x2="160" y2="162" marker-end="url(#wp-req-arrow)" />
<rect class="d-cmd" x="110" y="166" width="100" height="24" rx="12" /><text class="d-cmd-text" x="160" y="182" text-anchor="middle">router</text>
<path class="d-flow" d="M135 190C135 206 84 206 84 222" marker-end="url(#wp-req-arrow)" />
<path class="d-flow" d="M185 190C185 206 238 206 238 222" marker-end="url(#wp-req-arrow)" />
<text class="d-small" x="88" y="204" text-anchor="end">page</text>
<text class="d-small" x="232" y="204">action</text>
<rect class="d-chip-hit" x="34" y="226" width="100" height="22" rx="11" /><text class="d-hit-text d-small" x="84" y="241" text-anchor="middle">cache'e bakış</text>
<path class="d-hitline" d="M34 237H12V445H22" marker-end="url(#wp-req-hit)" />
<text class="d-hit-text d-small" x="14" y="231">hit</text>
<line class="d-flow" x1="84" y1="248" x2="84" y2="262" marker-end="url(#wp-req-arrow)" />
<text class="d-small" x="90" y="259">miss</text>
<rect class="d-a" x="24" y="266" width="120" height="26" rx="5" /><text class="d-on-a" x="84" y="283" text-anchor="middle">OnBeforeRender</text>
<line class="d-flow" x1="84" y1="292" x2="84" y2="304" marker-end="url(#wp-req-arrow)" />
<rect class="d-soft-a" x="24" y="308" width="120" height="26" rx="6" /><text x="84" y="325" text-anchor="middle">render</text>
<line class="d-flow" x1="84" y1="334" x2="84" y2="346" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="24" y="350" width="120" height="26" rx="5" /><text class="d-on-a" x="84" y="367" text-anchor="middle">OnAfterRender</text>
<line class="d-flow" x1="84" y1="376" x2="84" y2="388" marker-end="url(#wp-req-arrow)" />
<rect class="d-chip-hit" x="29" y="392" width="110" height="22" rx="11" /><text class="d-hit-text d-small" x="84" y="407" text-anchor="middle">cache'e yaz · GET</text>
<line class="d-flow" x1="84" y1="414" x2="84" y2="428" marker-end="url(#wp-req-arrow)" />
<rect class="d-b" x="24" y="432" width="120" height="26" rx="5" /><text class="d-on-b" x="84" y="449" text-anchor="middle">OnPersonalise</text>
<rect class="d-frame" x="170" y="226" width="136" height="22" rx="5" /><text class="d-small" x="238" y="241" text-anchor="middle">guard'lar · forgery</text>
<line class="d-flow" x1="238" y1="248" x2="238" y2="262" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="170" y="266" width="136" height="26" rx="5" /><text class="d-on-a" x="238" y="283" text-anchor="middle">OnBeforeAction</text>
<line class="d-flow" x1="238" y1="292" x2="238" y2="304" marker-end="url(#wp-req-arrow)" />
<rect class="d-soft-a" x="170" y="308" width="136" height="26" rx="6" /><text x="238" y="325" text-anchor="middle">action handler'ı</text>
<line class="d-flow" x1="238" y1="334" x2="238" y2="346" marker-end="url(#wp-req-arrow)" />
<rect class="d-frame" x="170" y="350" width="136" height="46" rx="5" /><text class="d-small" x="238" y="364" text-anchor="middle">page ile cevaplarsa:</text><text class="d-small" x="238" y="376" text-anchor="middle">render hook'ları</text><text class="d-small" x="238" y="388" text-anchor="middle">cache'e uğramadan</text>
<line class="d-flow" x1="238" y1="396" x2="238" y2="428" marker-end="url(#wp-req-arrow)" />
<rect class="d-b" x="170" y="432" width="136" height="26" rx="5" /><text class="d-on-b" x="238" y="449" text-anchor="middle">OnPersonalise</text>
<path class="d-flow" d="M84 458C84 476 140 474 140 486" marker-end="url(#wp-req-arrow)" />
<path class="d-flow" d="M238 458C238 476 180 474 180 486" marker-end="url(#wp-req-arrow)" />
<rect class="d-pill" x="110" y="490" width="100" height="24" rx="12" /><text x="160" y="506" text-anchor="middle">response</text>
<line class="d-rule" x1="6" y1="524" x2="314" y2="524" />
<text class="d-small" x="160" y="542" text-anchor="middle">hata: OnError · document: OnDocumentRendered</text>
</svg>
<figcaption>İlk olarak, collage'ın request span'inden önce OnRequest çalışır. Sonra Config.Plugins'teki plugin'lerin middleware'i, uygulamanın app.Use'unun dışında çalışır. GET ya da HEAD için cache'in tuttuğu bir page render hook'larını atlar ve doğrudan OnPersonalise'a gider. Yeni bir render OnBeforeRender'ı ve OnAfterRender'ı çalıştırır ve yalnızca render'ı degraded olmayan bir GET için tag'iyle cache'e yazılır; hiç cache'lenmeyen bir page'de ne bakış ne yazma vardır. Bir action, guard'larından, body sınırından ve forgery kontrolünden sonra OnBeforeAction'ı çalıştırır; ev.Result'ı ayarlayan bir plugin handler'ın yerine cevap verir. Page'iyle cevap veren bir action render hook'larını cache'e uğramadan çalıştırır ve her HTML cevabı OnPersonalise'dan geçer; redirect ile cevap veren bir action ikisine de ulaşmaz. Bir document render hook'ları yerine OnDocumentRendered'ı çalıştırır; sunulurken oluşan bir hata OnError'a ulaşır.</figcaption>
</figure>

`OnRequest` dışında her hook metodunun imzası
`func(ctx context.Context, ev *Event) error` biçimindedir. v0.60.0, hiçbir şeyin
kullanmadığı `PageResolvedHook` ve `CacheWriteHook`'u kaldırdı: bir plugin'in görmek
istediği request `OnRequest`'in, çalışmak üzere olan bir render ise
`OnBeforeRender`'ın konusudur.

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
		if route := collage.RouteInfo(ctx); route.Kind != "" {
			span.SetName(r.Method + " " + route.Kind + ":" + route.Name)
		}
		span.SetAttributes(attribute.Int("http.response.status_code", status))
		span.End()
	}
}
```

`collage.RouteInfo(ctx)`, request'in neye resolve edildiğini bir `collage.Route`
olarak bildirir: `Kind`'ı (`collage.RouteKindPage`, `RouteKindDocument`,
`RouteKindAction`, `RouteKindMount` ya da `RouteKindHandler`; v0.61.0'dan beri.
String olarak `"page"`, `"document"`, `"action"`, `"mount"`, `"handler"`'dır ve
küme büyüyebilir) ve register edildiği adı ya da prefix'i olan `Name`'i. Bu ad page'in,
document'ın ya da action'ın adıdır; mount'ta ve handler'da ise prefix'tir. Request
henüz resolve edilmemişken ve 404 gibi hiçbir şeye resolve edilmeyen bir request'te
sıfır `Route`'tur. (v0.60.0'dan önce `collage.RouteOf(ctx)` yalnızca türü ve adı
dönüyordu.)

`collage.RouteInfo`, collage'ın request'i sunduğu
context'i okur. Finish fonksiyonunun context'i de bu context'i paylaşır,
`Metrics.HTTPResponse` da bu context'i alır. Böylece bir span ya da metric, ham
path yerine route ile etiketlenebilir. Bu, URL başına bir değer değil, sınırlı bir
kümedir.

`Route` iki alan daha taşır: `Pattern`,
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
Bunlar çoğu zaman bir tarayıcı botunun yoklamalarıdır.

Böyle bir request için finish
fonksiyonu yazılan status ile çağrılır (request sunulurken bir panic olmadıkça bir
kez), `collage.RouteInfo` sıfır `Route`'tur ve `r.URL` ham path'tir.
`collage.ClientIP(r)` `OnRequest` içinde çalışır; bu yüzden istemcileri
sayan ya da banlayan bir plugin bu request'leri görebilir. `Metrics.HTTPResponse` ve
collage'ın kendi request span'i bu request'leri hâlâ atlar.

Hook'lar static bir build'in header yakalamasını da görür (v0.52.0'dan beri).
`BuildFinishedHook` implement eden bir plugin register edilmişse build, yazdığı her
dosyayı handler'a sorar ve bu request'lerde `collage.IsCapture(r.Context())`
true'dur. İstemcileri sayan ya da banlayan bir plugin bunları atlar. Bkz.
[BuildFinishedHook](#buildfinishedhook).

İstemcinin adresi için `r.RemoteAddr` yerine `collage.ClientIP(r)` kullanın. Bu bir
`netip.Addr`'dir: `RemoteAddr`'in host'u ya da uygulamanın `Server.TrustedProxies`'te
listelediği proxy'lerin arkasında proxy'nin bildirdiği istemci (bkz.
[Deployment](/docs/deployment#behind-a-proxy-trustedproxies)). Adres unmapped'dır,
zone taşımaz ve verilecek bir adres olmadığında sıfır `netip.Addr`'dir: `RemoteAddr`'i
olmayan paylaşılan bir page render'ında ve `X-Forwarded-For`'u kullanılabilir bir
istemci bildirmeyen (`unknown`) güvenilen bir proxy'nin arkasında. Onu asla
saymayın.

`collage.IsHTTPS(r)` (v0.61.0'dan beri), request'in TLS üzerinden gelip gelmediğini
söyler: ya kendi bağlantısında TLS vardır ya da bunu `X-Forwarded-Proto` ile bildiren
bir proxy'den geçmiştir. `Server.TrustedProxies` doluysa header yalnızca oradaki
proxy'lerden geldiğinde dikkate alınır; böylece bir istemci sahip olmadığı bir TLS
bağlantısını iddia edemez. Liste boşsa header kimden gelirse gelsin kabul edilir.
Forgery guard kendi cookie'sinin `Secure` bayrağına da bu kararla karar verir.
Bir cookie'ye `Secure` koyan ya da `Strict-Transport-Security` gönderen bir
plugin, header'ı kendisi okumak yerine bu fonksiyonu kullanır.

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

Yeni bir render'dan hemen önce çalışır; cache hit'te **çalışmaz**. Page'in
[guard](/docs/pages-and-layouts#private-pages-guards)'larından sonra çalışır;
onlardan birinin engellediği bir request için hiç çalışmaz. Page'ler, error
page'ler ve bir action'ın `RenderPage` ile cevap olarak döndüğü page için çalışır.
Static build'in render ettiği her page için de çalışır.

Render context'ini alan tek hook budur ve sebebi hoisting'dir. Page'e bir şey
ekleyen plugin, bunu ağaç render edilmeden önce tanımlar. Böylece bir fragment onu
hâlâ override edebilir:

```go
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.opts.Generator)
	return nil
}
```

Burada yapılan bir tanım sıfır derinliğinde durur. Bu yüzden aynı key'i tanımlayan
herhangi bir fragment onun yerini alır. Varsayılan değeri plugin, özel değeri page
verir. Tanım yalnızca layout'un `{{hoist "head"}}` çağırdığı yere yerleşir. Dönen
bir hata, request'i `"before_render"` altında 500 ile başarısız kılar.

`Static` (v0.22.0'dan beri), page'in bir request için değil, `App.RenderPath`
üzerinden bir static build için render edildiğini söyler. `AfterRenderEvent`'te de
bulunur. Build'in header yakalaması için yapılan render bir request'tir, bu yüzden
onda `Static` false'tur; onu `collage.IsCapture(ctx)` ayırt eder.

### BeforeActionHook

```go
type BeforeActionEvent struct {
	Action  *collage.Action
	Locale  string
	Request *http.Request
	Result  *collage.ActionResult // set it to answer in the handler's place
}

func (ev *BeforeActionEvent) Form() (url.Values, error)
```

`Page` v0.78.0'dan beri yok: onu okuyan yoktu.

Her action request'inde bir kez, handler'dan hemen önce çalışır. Page'in
guard'larından, action'ın body sınırı konduktan ve forgery kontrolünden sonra
gelir (v0.31.0'dan beri). Bir plugin'in gönderimi kontrol ettiği yer, örneğin bir
spam filtresi ya da bir kota için, burasıdır, çünkü buradaki body action'ın kendi
sınırıyla sınırlıdır; body'yi daha önce, middleware'de okuyan bir plugin'in ise
kendi sınırına ihtiyacı olurdu.

`ev.Form()` form'u parse eder, URL-encoded ya da multipart fark etmez, ve alanlarını
döndürür. Parse ettiği parse edilmiş kalır. Böylece handler aynı form'u okur ve body
iki kez okunmaz. Tek başına `ParseMultipartForm` yerine bunu kullanın. O, URL-encoded
bir body'de okuma hatasını, sınırı aşan bir body'ninkini de, yok sayar.

`WithStreamingBody()` ile kurulmuş bir action için (v0.57.0'dan beri) `ev.Form()`
hiçbir şey okumaz ve `collage.ErrStreamingBody` döndürür. Böylece bir plugin,
handler'ın okuyacağı stream'i tüketemez. Form'ları inceleyen bir plugin bu hatayı
"bu action'ın kontrol edilecek form'u yok" diye yorumlamalı ve request'i
geçirmelidir. Bu hatayı döndüren bir plugin her upload'u başarısız kılar.
`ev.Request.ParseForm()`'u kendisi çağıran bir plugin ise body'yi yine okur ve
stream'i handler'ın elinden alır.

```go
func (p *Plugin) OnBeforeAction(_ context.Context, ev *collage.BeforeActionEvent) error {
	form, err := ev.Form()
	if errors.Is(err, collage.ErrStreamingBody) {
		return nil // a streaming action: no form to check
	}
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
sarıyorsa `413`, aksi halde `500` olur. v0.57.0'dan beri body, plugin onu okurken
yarıda kesildiyse cevap `400` (okuma deadline'ı için `408`) olur. Bu durum debug
seviyesinde log'lanır ve hiçbir error hook'una verilmez.

### AfterRenderHook

```go
type AfterRenderEvent struct {
	Page     *collage.Page
	Locale   string
	Degraded bool   // some fragment failed, fallback or not, or stale cached data was used
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
çalışır. Çıktıyı post-process etmek için `ev.HTML`'i değiştirin. Orada
bıraktığınız içerik sunulan içeriktir; cache'e yazılan içerik de odur. Sonraki
plugin'ler, öncekilerin ürettiği içeriği görür.

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

Hook, forgery token paylaşılan body'ye girdikten sonra, her middleware'in içinde,
yani sıkıştırmadan önce ve dev reload script'inden önce çalışır. Bir page'de dev
overlay'inden önce çalışır; bir error page'de ya da bir action'ın cevap olarak
verdiği bir page'de ise overlay body'de zaten vardır.
Plugin'ler register sırasına göre çalışır ve her biri öncekinin body'sini görür. Bir
page'i (cache'ten ya da yeni), bir fragment path'ini ya da fragment read'ini, bir
action'ın HTML cevabını, bir error page'i ve v0.78.0'dan beri `RenderFragment`'in
döndürdüğünü (markup'ı ve hoist edilen her öğeyi) kapsar; document'ı asla. Bir hook
`Personal`'ı bir kez ayarladığında sonraki bir hook onu geri alamaz (v0.78.0'dan
beri).

`Personal`, response'u bir forgery token'ın yaptığı gibi ele alır: ETag'i gönderilen
body'den yeniden hesaplanır, page `private, no-store` cevabı verir; ne cache'ten
gelen bir page ne de bir fragment read, `If-None-Match: *`'a bile, `304` alır.

Fragment read yeniden
hesaplanan ETag ile `private, no-cache`, action'ın HTML'i ise `private, no-store`
cevabı verir; handler'ın kendi ayarladığı `Cache-Control` da geçersiz kalır.
elagoht/compress gibi sıkıştırılmış body'leri ETag başına saklayan bir middleware
kişisel bir response'tan bir şey kazanmaz: ETag'i her seferinde yenidir.
elagoht/compress onu her okuyucu için sıkıştırır ve hiçbir şey saklamaz (v0.1.3'ten
beri).

`Personal` olmadan değişen bir body da gönderilen byte'ları adlandıran bir ETag
alır. Hata dönen ya da panic eden bir hook log'lanır; page, fragment ya da action
500 cevabı verir, error page'de ise yerine o status'un built-in page'i gönderilir.
Böyle bir plugin'i olmayan bir site eskisiyle aynı response'ları gönderir.

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

Bir document, `OnBeforeRender` ya da `OnAfterRender` dispatch
etmez, çünkü page'i yoktur ve template render etmez.

### CacheInvalidateHook

```go
type CacheInvalidateEvent struct {
	Tags    []string
	Entries []InvalidatedEntry // the cached entries invalidated, each with its host, sorted
}

type InvalidatedEntry struct {
	Host string
	Path string
}
```

`InvalidateTags` ya da `PurgeTags` bazı tag'lerin entry'lerine ulaştıktan sonra
çalışır. Onu uygulamanın ya da bir action'ın çağırması fark etmez. Hook bir
request'ten değil, o çağrının içinden dispatch edilir. Onu implement eden her
plugin, kendisinden önceki biri başarısız olsa ya da panic'e düşse bile çağrılır
(v0.78.0'dan beri; başarısız olan tek bir plugin, bir CDN purge'ünün ya da bir live
güncellemesinin çalışmasını engelliyordu). Her hata log'lanır ve `InvalidateTags`'in
döndüğü hatayla join edilir. Plugin invalidate etmez: `Host.InvalidateTags`
v0.78.0'dan beri yok.

`Entries`, invalidation'ın ulaştığı cache'lenmiş page ve document'ları listeler.
Bir CDN'in purge etmesi ve bir arama motoruna değiştiği bildirilmesi gereken
budur. Built-in store'larda `InvalidateTags`'in ulaştığı bir entry silinmez, stale
olarak işaretlenir (v0.63.0'dan beri). Artık taze diye sunulmaz. Yalnızca yeniden
render'ı başarısız ya da degraded olursa `STALE` olarak yeniden sunulur.
Cache'lenmemiş bir page burada asla yer almaz, çünkü onun için ulaşılacak bir
entry yoktur. Daha önceki bir invalidation'ın stale olarak işaretlediği bir page
de yeniden render edilip saklanana kadar burada yer almaz. Bir filonun paylaştığı
bir store, yani bir [`collage.SharedCache`](/docs/caching#a-cache-a-fleet-shares)
(v0.66.0'dan beri) bunun istisnasıdır: onun kendi tag index'i herhangi bir
instance'ın yazdığı her page'i listeler ve invalidate edilip henüz yeniden render
edilmemiş bir page'i de tutar. Bu yüzden ikinci bir invalidation o page'i yine
listeler.

Cache'lenen her entry ayrıca
`collage.PathTag(path)` tag'ine de bağlıdır. Bu yüzden tag'i değil path'i bilen bir
uygulama, oradaki cache'i `app.InvalidateTags(ctx, collage.PathTag("/blog"))` ile
invalidate eder. Bkz. [Caching](/docs/caching#invalidating-by-path).

Her entry bir host ve bir URL path'tir; önce host'a, sonra path'e göre sıralıdır.
İki host altında cache'lenmiş bir path iki entry'dir ve request bir port içeriyorsa
host portunu korur. Bir CDN'i purge eden ya da bir arama motoruna haber veren plugin,
her URL'yi doğru origin'e göre kurmak için bunu okur. (Path'ler v0.23.0'da `Paths`
olarak, host'lar v0.42.0'da `Entries` olarak geldi; v0.60.0 yalnızca `Entries`'i
tuttu.) Host, `collage:path:`'in yanında bir `collage:host:<host>` dependency tag'i
olarak kaydedilir. Host tag'leri tracker'ın tag başına sınırından muaftır;
`MaxEntries` hâlâ belleği sınırlar.

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
ve development modu bunu plugin ve origin başına bir kez log'lar.

Hiçbir plugin'in
tanımadığı host `Config.BaseURL`'ü alır; panic eden bir resolver log'lanır ve
atlanır. Uygulama kodu sonucu `collage.BaseURL(rc)` ile okur. Host cache key'in
içinde olduğu için bu, cache'lenmiş bir render'da güvenlidir; request'i olmayan bir
static build de `Config.BaseURL`'ü alır. Resolver'ı olmayan bir site hiçbir değişiklik
görmez.

Mutlak URL'leri kendisi kuran bir plugin, `Init`'te aldığı `Host`'a sorar
(v0.60.0'dan beri; önceden type assertion ile `collage.Origins` capability'si
üzerinden). v0.78.0'dan beri `ConfigHost`'ta `BaseURL`, `OriginFor` ve
`MultiOrigin` yoktur: `Configure` sırasında her resolver bilinmez.

```go
origin := host.OriginFor(ctx, "acme.test") // a resolver's, else Config.BaseURL
varies := host.MultiOrigin()               // is a resolver registered?
```

Origin'leri kendi config'inde alan bir plugin, kuralın bir kopyası yerine bunları
`collage.ParseOrigin` ile doğrular; onun reddettiği şey `collage.ErrInvalidOrigin`'dir
(v0.60.0'dan önce `ErrInvalidBaseURL`). Yalnızca bir context'i olan kod,
`collage.Vary` ile bildirilmiş bir header'ı `collage.Varied(ctx, header)` üzerinden
okur.

### ErrorHook

```go
type ErrorEvent struct {
	Err     error
	Path    string
	Stage   string
	Status  int           // the HTTP status the failure is answered with; 0 for none
	Request *http.Request // the reader's request, read-only; nil outside a request
}
```

Bir request sunulurken bir hata oluştuğunda çalışır. Hata bir page'de, bir
document'ta, bir action'da, bir mount'ta ya da `App.Handle` ile register edilmiş bir
handler'da olabilir. `Stage`, hatanın nerede olduğunu söyler.

v0.78.0'dan beri event bir `Page` taşımaz; neyin başarısız olduğunu `Path`'ten
okuyun.

Framework'ün kullandığı stage'lerin her birinin bir sabiti vardır (v0.61.0'dan
beri). Literal string yerine bunlarla karşılaştırın:

| Stage | Sabit |
| --- | --- |
| `"route"` | `collage.StageRoute` |
| `"not_found"` | `collage.StageNotFound` |
| `"guard"` (v0.28.0'dan beri: bir page'in guard'ı başarısız olmuş ya da yazılamayacak bir kararla cevap vermiştir) | `collage.StageGuard` |
| `"before_render"` | `collage.StageBeforeRender` |
| `"before_action"` | `collage.StageBeforeAction` |
| `"render"` | `collage.StageRender` |
| `"after_render"` | `collage.StageAfterRender` |
| `"cache_write"` | `collage.StageCacheWrite` |
| `"error_page"` | `collage.StageErrorPage` |
| `"asset"` | `collage.StageAsset` |
| `"plugin"` | `collage.StagePlugin` |
| `"handler"` | `collage.StageHandler` |
| `"panic"` | `collage.StagePanic` |

Bu küme kapalı bir enum değildir ve
büyüyebilir; bu yüzden hepsini gördüğünüzü varsaymak yerine ele aldığınız
stage'leri adlarıyla seçin.

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
`errors.As` panic değerine ve panic'in oluştuğu stack'e ulaşır, kod bir hata ile
panic ettiyse `errors.Is` değerin kendisine ulaşır. Stack artık hatanın mesajında
değildir; yalnızca framework'ün `collage: request failed` log satırındaki ayrı bir
`stack` attribute'unda bulunur. Stack'i mesajda arayan kod bunun yerine
`PanicError.Stack`'i okumalıdır.

`Status` (v0.45.0'dan beri), hatanın cevaplandığı HTTP status'tür: var olmayan bir
asset için 404, başarısız bir render ya da panic için 500. Hata için hiçbir cevap
yazılmıyorsa 0'dır; page sunulduktan sonra başarısız olan bir cache yazımı ya da
yerine stale bir page sunulan başarısız bir render gibi (v0.63.0'dan beri). Bu
sonuncusu başarısız olduğu stage'i korur, örneğin `"render"`. Onu kaç okuyucu
beklemiş ve stale page'i almış olursa olsun, `OnError`'a başarısız her render için
bir kez ulaşır. Bkz. [Caching](/docs/caching#stale-pages-and-x-collage-cache).
`"error_page"` event'i her zaman 500 der.

`Request`, okuyucunun canlı request'idir ve
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
	OutDir    string                  // the directory the build wrote into
	Files     []collage.BuiltFile     // every file it wrote, in no particular order
	Redirects []collage.BuiltRedirect // every redirect the site declares
	// Findings: what ev.Warn and ev.Error reported so far
}

type BuiltFile struct {
	Kind     string // collage.BuiltPage, BuiltDocument or BuiltAsset: "page", "document", "asset"
	Name     string // the page's or document's name; empty for an asset
	Locale   string
	Path     string // the URL path the file answers
	File     string // its absolute path on disk
	Captured bool        // the build asked the application for this path
	Status   int         // what it answered with; 0 when not captured or failed
	Headers  http.Header // the headers a static host can carry
}

type BuiltRedirect struct {
	From   string // "/old/{slug}", as registered
	To     string // "/new/{slug}", an absolute URL, or empty for a 410
	Status int    // 301, 302, 307, 308 or 410
	Source string // "page:<name>", "document:<name>" or a plugin's name
}
```

Bir static build her page'i, document'ı ve asset'i yazdığında bir kez çalışır
(v0.21.0'dan beri). Page'ler arası denetimler içindir. Bir dosyanın içeriği
gerektiğinde onu `os.ReadFile` ile okuyun. `ev.Warn(path, rule, message)` ve
`ev.Error`, `path`'teki page hakkında bir finding raporlar; `path` boşsa finding
build'in bütünü hakkındadır.

Hook'tan dönen bir hata da build'i başarısız kılar;
zaten yazılmış dosyalar yerinde kalır. Bir sunucuda hiçbir zaman çalışmaz.
Context'i sona ermiş bir build de ona verilmez.

**Header'lar.** v0.52.0'dan beri build, hook çalışmadan önce her dosyanın path'ini
uygulamanın handler'ına iki kez sorar ve cevabı dosyaya koyar: `Status` ve
`Headers`. Tek bir response'a ait header'lar (`Date`, `ETag`, `Set-Cookie`, …) ve
iki cevap arasında değişenler, örneğin bir CSP nonce'ı, bunlara girmez. `Captured`,
build'in sorduğu her dosya için true'dur; yakalaması başarısız olan dosya da
(`Status` 0) buna dahildir. Build'in kendisinin ürettiği `404.html` page'leri için
ise false'tur. `PrefixDefault` açıkken home page'in kökteki kopyası, home page'in
yakalama sonucunu taşır.

Yakalama yalnızca `BuildFinishedHook` implement
eden bir plugin register edilmişse çalışır. Bulduğu şeyler uyarı olarak raporlanır;
bkz. [Static export](/docs/static-export#headers-and-redirects).

**Redirect'ler.** `ev.Redirects`, sitenin tanımladığı her redirect'i taşır: önce
register sırasıyla her page'inkiler (`Source` `"page:<name>"`), sonra her
document'ınkiler (`"document:<name>"`), en son da her plugin'inkiler. `Status`,
redirect'in status'udur; status vermemiş bir redirect için `302`'dir. Yalnızca
`RegisterNotFoundPage` ya da `RegisterErrorPage` ile register edilen bir page
hiçbir zaman eşleşmez, bu yüzden onun redirect'leri listede yer almaz.

Bir deploy adapter'ı tam olarak bu ikisine ihtiyaç duyar:
[elagoht/deploy](/docs/plugins#elagohtdeploy) onları bir host'un `_headers`,
`_redirects` ya da `vercel.json` dosyası olarak yazar.

**Yakalama bir request'tir.** Middleware ve request hook'ları onu görür, page de
`OnBeforeRender`, `OnAfterRender` ve `OnDocumentRendered` ile yeniden render edilir.
Trafiği sayan ya da sınırlayan bir plugin de, bir render hook'undan dosya yazan ya
da sayım tutan bir plugin de `collage.IsCapture(ctx)`'e bakar ve onu atlar. Bir
yakalama request'i response cache'ine hiçbir zaman yazmaz.

### RedirectSource

```go
type RedirectSource interface {
	Redirects() []collage.BuiltRedirect
}
```

Kendi redirect'lerini (bir dosyadan, bir veritabanından) sunan bir plugin,
static export'un bu kuralları host'a taşıması için `RedirectSource` implement eder
(v0.52.0'dan beri). Build, her dosya yazıldıktan sonra ona bir kez sorar ve
kurallarını `Source` alanına plugin'in adını koyarak `ev.Redirects`'e ekler. Her
kural, register edilmiş bir redirect gibi kontrol edilir: `From` tek bir `/` ile
başlamalı ve bir route pattern'i olarak parse edilebilmeli, `To`'daki her
placeholder da onun tarafından yakalanmalıdır. Bir page'in redirect'inden farklı
olarak `To` mutlak bir `http` ya da `https` URL'i olabilir. `410`'un (artık olmayan
bir path) ise `To`'su yoktur; 301, 302, 307 ve 308 birine ihtiyaç duyar.

Başka
herhangi bir `Status` build'i `collage.ErrInvalidRedirectStatus` ile, bozuk ya da
control karakteri içeren bir kural ise `collage.ErrInvalidRedirect` ile başarısız
kılar. Hata, kuralı ve plugin'i adıyla söyler. v0.78.0'dan beri bir host'un redirect
dosyasının alan ayırıcı olarak okuduğu bir boşluk içeren `From` ya da `To` da
bozuktur ve her bozuk kural, her plugin'inki, raporlanır; önceden tek bir kural
bütün plugin'lerin kurallarını düşürüyordu. Başka bir redirect ile aynı path'ten
gelen ya da yazılmış bir dosyanın üstüne düşen bir plugin kuralı da build'i diğerleri
gibi başarısız kılar ([Neler başarısız olur](/docs/static-export#what-fails)).
[elagoht/redirects](/docs/plugins#elagohtredirects) böyle bir plugin'dir.

### Dispatch kuralları

- Hook'lar **register sırasına göre** çalışır.
- Her çağrı **panic'e karşı korunur**. Panic'e düşen bir hook, hata dönmüş bir hook
  gibi başarısız sayılır; process'i çökertmez.
- `OnBeforeRender`, `OnAfterRender` ve `OnDocumentRendered` için
  **ilk hata dispatch'i durdurur** ve request'i başarısız kılar.
- `OnPersonalise` için ilk hata dispatch'i durdurur; page, fragment ya da action 500
  cevabı verir, error page ise o status'un built-in page'ine döner.
- `OnBeforeAction` için ilk hata dispatch'i durdurur ve request'i başarısız kılar.
  `ev.Result`'ı ilk ayarlayan plugin de dispatch'i durdurur ve request'i cevaplar.
- `OnCacheInvalidate` için her plugin, biri başarısız olsa ya da panic'e düşse bile
  çağrılır; hatalar log'lanır ve `InvalidateTags`'ten join edilmiş olarak döner
  (v0.78.0'dan beri; önceden ilk hata dispatch'i durduruyordu).
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
- Bunun dışında **uygulama kazanır**. `Config.Template.Funcs` içinde bir
  plugin'in eklediği adla bir kayıt varsa, plugin'in fonksiyonunun yerini alır:
  uygulama ikisini de görebilir ve kararı o verir; `New` bu adların her birini bir
  kez log'lar. v0.78.0'dan beri bu `AddRenderFunc` için de geçerlidir. Onun
  fonksiyonu önceden her render'da, uygulamanınkilerden sonra yeniden bağlanıyordu;
  bu yüzden çalışan oydu.
- **Built-in bir ad plugin'in alabileceği bir ad değildir.** v0.77.0'dan beri
  `AddTemplateFunc`, herhangi bir [built-in fonksiyonun](/docs/template-functions)
  adını, `pageURL` kadar `json` ya da `truncate`'i de, `ErrDuplicateTemplateFunc`
  ile reddeder. Önceden plugin'in fonksiyonu, o built-in'e göre yazılmış her
  template'in altında built-in'in yerini sessizce alıyordu; built-in her render'da
  bağlanan bir fonksiyonsa da hiç çalışmıyordu. Kendinize ait bir ad seçin. Bir
  built-in'i değiştirmek uygulamanın işidir ve `Config.Template.Funcs` ile yapılır.
- **Template dilinin bir adı da alınamaz.** v0.78.0'dan beri `AddTemplateFunc` ve
  `AddRenderFunc`, Go'nun builtin'lerini (`eq`, `len`, `index`, `print`…) ve
  escaper'larını (`html`, `js`, `urlquery`, html/template'in kendi escaper'ları)
  `ErrDuplicateTemplateFunc` ile reddeder: bir plugin'in `html`'i, ondan geçirilen
  bir değerin aldığı tek escaping'di.
- `AddTemplateFunc`, `Configure` içinden çağrılmalıdır. Sonradan eklemenin bir yolu
  yoktur, çünkü parse işleminden sonra eklenen bir fonksiyonu hiçbir template
  çağıramaz; `Configure` döndükten sonra saklanan bir `ConfigHost` üzerinden yapılan
  çağrı `ErrConfigureClosed`'dır (v0.78.0'dan beri).

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
plugin'in eklediği bir ad ya da bir built-in'in adı `ErrDuplicateTemplateFunc`
hatasıdır: bir render fonksiyonu, framework'ün
[Template fonksiyonları](/docs/template-functions) sayfasında listelenen kendi
fonksiyonlarından birinin yerini alamaz.

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
dosyanın yanlış kısmını döner. Dosyaları *zaten* dönüştürülmüş hâlde sunan bir
wrapper ise bu hesabı doğru tutar. Wrapper'lar register edildikleri sırayla çalışır;
`nil` bir wrapper yok sayılır. `Configure` içinden çağrılır; v0.78.0'dan beri daha
sonraki bir çağrı hiçbir şeyi wrap etmez ve log'lanır.

## Document ve mount eklemek

Bir plugin, `Init` içinden kendi route'larını ekleyebilir. Bunun için
`Host.RegisterDocument` ve `Host.Mount` kullanılır. Bu route'lar, bir uygulamanın
kullandığı builder'larla oluşturulur. v0.60.0'dan beri bir plugin page register
etmez: `Host.RegisterPage` kaldırıldı, page'ler uygulamanındır. `Host.Handle` ise
page olmayan şeyler için, örneğin bir event stream ya da bir WebSocket için, bir
prefix altında düz bir `http.Handler` sunar. `/` ile biten bir prefix, altındaki her
path'i üstlenir. v0.24.0'dan beri sonunda `/` olmayan bir prefix, örneğin
`/metrics`, tek bir tam path'tir. Her zaman middleware'in, yani uygulamanınkinin ve
her plugin'inkinin içinde çalışır: sitenin bir parçasıdır, session'ı okuyabilir ya
da önündeki bir basicauth'a dayanabilir. v0.78.0'dan beri request body'si, bir
page'inki gibi, middleware'den önce `Server.MaxBodyBytes` ile sınırlanır; böylece
orada hiçbir middleware sınırsız bir body okumaz. `App.Handle` değişmedi.

`Host.Use` (v0.21.0'dan beri),
`App.Use` gibi her request'i, plugin'in register edildiği yerde sarmalar.
v0.38.0'dan beri `Config.Plugins` içindeki bir plugin, uygulamanın kendi
middleware'inin dışındadır. Böylece request context'ine koyduğu şey, örneğin bir
session, uygulamanın middleware'i tarafından okunabilir; header'ları ve
cookie'leri de o middleware'in ürettiğini sarar. `App.RegisterPlugin` ile
register edilen bir plugin ise uygulamanın middleware'leri arasında, çağrının
yapıldığı yerde durur. v0.67.0'dan beri uygulamanın düz `App.Handle` ile mount
ettiği, kendi middleware'ini getiren bir handler için çalışmaz: ban'layan,
sınırlayan, koruyan ya da log'layan bir plugin böyle bir mount'un sorumluluğunu
üstlenemez. Bkz.
[İstemedikçe middleware yok](/docs/middleware-and-apis#no-middleware-unless-you-ask).

Page'lere link veren bir plugin, onların nerede olduğunu tahmin etmez, sorar.
`Host.URL`, bir path'i `App.URL` gibi oluşturur. `Host.Locales` varsayılan
locale'i ve desteklenen her locale'i döner.
`Host.PageURLs(ctx, name)` ise bir page'in cevap verdiği her URL'yi listeler: her
locale için ve bir pattern'in `WithStaticParams`'ının listelediği her parametre
kümesi için bir `collage.PageURL` (`Locale`, `Path`, `Params`).
`WithStaticParams`'ı olmayan bir pattern'in, bir plugin'in bilebileceği hiçbir
URL'si yoktur. Bir sitemap bundan oluşur. `App.Locales` ve `App.PageURLs` aynı
metotlardır ve uygulamanın kendi kodu içindir.

Sitenin kendi dosyalarından birini, örneğin bir kaynak görseli ya da bir stylesheet'i
*okuyan* bir plugin, onu diskteki bir path ile değil, sitenin onu sunduğu URL ile
`Host.OpenMounted` üzerinden (v0.68.0'dan beri) ister:

```go
f, err := host.OpenMounted("/static/icon.png")
if errors.Is(err, fs.ErrNotExist) {
	// no icon yet: fall back, and say where to put one
}
```

Bu çağrı uygulamanın mount'u üzerinden geçer. Bu yüzden development'ta
`static/icon.png`'yi diskten, build edilmiş bir binary'de ise embed edilmiş kopyadan
okur. Binary'nin yanına hiçbir şeyin kopyalanması gerekmez. Yalnızca bir
request'in ulaşabileceği dosyalar açılır: bir dotfile ya da hiçbir mount'un altında
olmayan bir path `fs.ErrNotExist`'tir. elagoht/favicon kaynağını bu şekilde okur.

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

`RegisterCommand`, hiçbir zaman çalıştıramayacağı bir adı reddeder (v0.78.0'dan
beri):

- boş bir ad;
- tek bir shell kelimesi olmayan bir ad (herhangi bir yazı sisteminin harf ve
  rakamları, `-` `_` `.` `:`, başta bir harf ya da rakam; boşluk yok, shell sözdizimi
  yok) `ErrInvalidCommandName`'dir;
- `DispatchCommands`'ın kendisinin cevapladığı bir `collage-` adı (`collage-check`,
  `collage-inspect`) ve başka bir komutun zaten kullandığı bir ad
  `ErrDuplicateCommand`'dır;
- bir plugin'in komutu için `collage` binary'sinin kendi adlarından biri de (`new`,
  `dev`, `build`, `export`, `serve`, `version`, `help`…) öyledir: `go run . build` ile
  `collage build`'in farklı şeyler yapması bir tuzaktır. Uygulamanın kendi komutu
  yine de böyle bir ad alabilir.

Bütün plugin'lerin `Init`'i çalışırken açık kalır ve uygulama başladıktan sonra
`ErrAppStarted` ile kapanır (v0.78.0'dan beri; önceden açık kalıyordu ve o sırada
register edilen bir komut hiç dispatch edilmiyordu).

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
`Init`'tir. Exit code'lar şöyledir:

- başarı için `0`;
- startup hatası, çalışıp başarısız olan bir komut ya da `Run`'ı olmayan bir komut
  için `1`;
- nil bir app, argüman verilmemesi ya da hiçbir komutun sahiplenmediği bir ad
  (`ErrUnknownCommand`) için `2`.

Plugin'inizin README'sinde, komutlarının uygulama üzerinden çalıştığını
belirtin. v0.10.0'dan önce scaffold edilmiş bir projenin bu bloğu kendisinin
eklemesi gerekir. Bkz. [collage CLI](/docs/cli#plugin-commands).

## Yapılandırma

Bir plugin, `Config.PluginConfig` içindeki kendi bölümünü
`collage.PluginConfig(host, defaults)` ile tipli bir struct'a okur. Bu fonksiyon
`Host`'u da `ConfigHost`'u da kabul eder ve varsayılan değerlerin tipini döner.
Uygulamanın bölümünü varsayılan değerlerin bir kopyasının üzerine decode eder:

```go
type Options struct {
	Generator string           `json:"generator"`
	Disabled  bool             `json:"disabled"`
	Timeout   collage.Duration `json:"timeout"` // "5s", "1m30s"
}

func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts) // what New was given, overlaid by the application's section
	p.opts = opts
	return err
}
```

- **Bölüm yoksa, boşsa ya da `null` ise varsayılan değerler olduğu gibi döner.**
  Böylece "yapılandırılmamış" ile "zero value'ya yapılandırılmış" iki farklı durum
  olarak kalır.
- **Bölüm varsa ama bozuksa bu, plugin'in adını veren bir hatadır.** Hataya yol
  açan bir üyenin değeriyse hata onun key'ini de verir
  (`collage: acme/stamp: "timeout": …`) ve varsayılan değerlerle birlikte döner.
  Operatör bir şey yazmıştır. Sessizce varsayılan değerlerle çalışmak bunu
  gizlerdi. Bölüm bir kopyaya decode edildiği için başarısız bir decode varsayılan
  değerlere dokunmaz.
- **Options'ınızda olmayan bir key de hatadır** (v0.61.0'dan beri). Bu, hangi
  derinlikte olursa olsun geçerlidir; `json:"-"` tag'li bir field'ın adını taşıyan
  bir key de hatadır: `collage: acme/stamp: unknown key "generater" in
  plugins-config.json; remove it, or rename it to a key the plugin documents`.
  Uygulama `Configure` ya da `Init` aşamasında durur. Key'ler field'larla,
  `encoding/json`'ın yaptığı gibi büyük küçük harf ayırmadan eşleşir. Tam `Options`
  tipini tek bir `PluginConfig` çağrısıyla decode edin: kontrol, decode edilen tipe
  göre yapılır ve kendi `UnmarshalJSON`'ı olan bir tip kendi key'lerini kendisi
  kontrol eder (ya da hiç etmez). Adı değişen ya da kaldırılan bir key böylece
  kendiliğinden yüksek sesle başarısız olur.
- JSON'daki bir scalar, varsayılan değerinizin yerini alır. **En üst düzeydeki bir
  liste varsayılan listenin yerini bütünüyle alır** (v0.61.2'den beri): elemanları
  yenidir, yani bir elemanın yazmadığı field o field'ın zero value'su olur,
  varsayılan elemanın değeri değil. En üst düzeydeki bir map, kendi kayıtlarını
  sizinkinin bir kopyasına ekler. İkisi de sizin slice'larınıza ya da map'lerinize
  yazmaz. Daha derinde ise bir JSON object'i varsayılanın üzerine decode edilir ve
  yalnızca adını verdiği field'ları atar; iç içe bir map hâlâ yerinde doldurulur. Bu
  yüzden varsayılan değerleri çağrılar arasında paylaşılan tek bir değer olarak
  değil, bir literal olarak yazın.
- **Bir süre `collage.Duration`'dır** (v0.61.0'dan beri), `time.Duration` değil.
  `time.ParseDuration`'ın anladığı bir string'i — `"10s"`, `"1h30m"` — okur ve geri
  yine string yazar. Çıplak bir sayı key'inin adı verilerek reddedilir;
  `time.Duration` olsaydı `3600`'ü tek kelime etmeden 3.6µs olarak okurdu. Negatif
  bir süre de reddedilir; `null` varsayılanı bırakır. Kullandığınız yerde
  dönüştürün: `time.Duration(p.opts.Timeout)`.
- Register edilmiş hiçbir plugin'e karşılık gelmeyen bir key, uygulama için bir startup
  hatasıdır (`ErrUnknownPluginConfig`). Bu yüzden config'inizin adresi tamamen
  `Name`'inizden ibarettir. Onu değiştirmek bir breaking change'dir.

v0.50.0'dan önce bir plugin varsayılan değerlerini bir field'a atar ve
`host.Config(&cfg)` çağırırdı.

README'nizde her key'i, tipini ve varsayılan değerini belgeleyin. Aynı ayarları
Go'da tek bir constructor ile, `func New(opts Options) *Plugin` ile alın ve bölümü
onların üzerine decode edin: böylece bir uygulama sizi Go'da, JSON'da ya da ikisiyle
birden yapılandırabilir. `collage plugin add` `New(Options{})` yazar; bu yüzden sıfır
`Options` çalışan bir plugin olmalıdır (ya da `Init`'te neyin eksik olduğunu söyleyen
bir mesajla başarısız olmalıdır).

Tipin adını `Config` değil `Options` koyun;
`Config`, çağrıldığı yerde `collage.Config` gibi okunur. Boolean'ların varsayılanı
zero value'larıdır: bir şeyi kapatan ayar, varsayılanı true olan bir key değil, bir
`noX` key'idir.

Bir secret asla `plugins-config.json`'a girmez: bir `<x>Env` key'i
değeri tutan environment variable'ın adını verir, `json:"-"` tag'li bir field da
değeri Go'dan alır.

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

// Options is the plugin's configuration. The zero value is the defaults.
type Options struct {
	// Generator is what the generator meta tag says; "collage" when empty.
	Generator string `json:"generator"`
}

// Plugin is the stamp plugin. Construct it with New.
type Plugin struct {
	opts   Options
	logger *slog.Logger
}

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

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
	opts, err := collage.PluginConfig(host, p.opts) // the options New was given
	if err != nil {
		return err // a section that is malformed or names an unknown key
	}
	if opts.Generator == "" {
		opts.Generator = "collage"
	}
	p.opts = opts
	return host.AddTemplateFunc("generator", func() string { return p.opts.Generator })
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
	ev.Context.HoistMeta("generator", p.opts.Generator)
	return nil
}

// OnError reports the one failure nobody would otherwise notice: the error page
// itself failing.
func (p *Plugin) OnError(_ context.Context, ev *collage.ErrorEvent) error {
	if ev.Stage == collage.StageErrorPage {
		p.logger.Error("stamp: the error page failed", "path", ev.Path, "error", ev.Err)
	}
	return nil
}
```

Bir uygulama bu plugin'i diğer plugin'ler gibi kullanır:

```go
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{stamp.New(stamp.Options{})},
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
4. **OnServe**, bir `ServeHook` için, `ListenAndServe` port'unu bağladıktan sonra
   ve hizmet vermeye başlamadan önce register sırasına göre çalışır (v0.55.0'dan
   beri). Onu yalnızca `ListenAndServe` çağırır. ctx'i drain başladığında iptal
   edilir. Bkz. [Stream'ler ve shutdown](#streams-and-shutdown).
5. **Shutdown**, `App.Shutdown` ile ters register sırasına göre çalışır.
   `ListenAndServe`, `App.Shutdown`'ı `SIGINT` ya da `SIGTERM` geldiğinde çağırır.

   - v0.61.0'dan beri `App.Shutdown`, `Shutdown`'ı **yalnızca `Init`'i başarılı olmuş
     plugin'ler için ve bir kez** çağırır: `Init`'i hiç çalışmamış ya da başarısız
     olmuş bir plugin çağrılmaz, zaten shutdown edilmiş bir plugin de tekrar
     çağrılmaz.
   - Her `Init` başarılı olduktan *sonra* başarısız olan bir başlatma —
     bir page'i yutan bir mount, hiç register edilmemiş bir error page — plugin'leri
     orada hemen shutdown eder; başarısız bir `Init`'in kendinden öncekileri geri
     alması gibi.
   - `Start`'tan önce gelen bir `Shutdown` hiçbir plugin'i initialize
     etmez (v0.61.1'den beri): `Start` başarısız olur, `ListenAndServe` de nil döner.
     `Shutdown`'dan sonra gelen bir `Start`, plugin register edilmiş olsa da olmasa da
     başarısız olur (v0.61.3'ten beri).
   - Biri başarısız olsa bile her plugin sırasını
     alır ve hatalar join edilir.
   - `StreamCloser`'ı implement eden bir plugin'in
     `CloseStreams`'i daha önce, shutdown başlarken çağrılır. `DrainHook`'u
     implement eden bir plugin'in `OnDrain`'i ise ondan da önce, drain başlarken
     çağrılır. İkisi de yalnızca `Init`'i başarılı olmuş plugin'ler için çağrılır
     (v0.61.4'ten beri); `Start`'tan önceki bir `Shutdown` hiçbirine haber vermez.
     Bkz. [Stream'ler ve shutdown](#streams-and-shutdown).

   `ListenAndServe` kullanıyorsanız, plugin'ler drain'in `Server.DrainDelay`'inden
   ve sunucu request'lerini bitirdikten sonra shutdown edilir. Sunucu bunu
   bitiremezse, `Server.ShutdownTimeout` dolduğunda shutdown edilirler. Bu süre
   dolduktan sonra hâlâ çalışan bir request olabilir.

   Sunucuyu kendiniz yönetiyorsanız, `App`'in bu sunucudan haberi yoktur.
   `ServeHook` plugin'lerini kendiniz başlatın, önce sunucunuzu durdurun, sonra
   `App.Shutdown`'ı çağırın. Aksi hâlde bir plugin, hâlâ devam eden bir request'in
   altından çekilip alınabilir (bkz.
   [Deployment](/docs/deployment#serving-with-your-own-server)).

   Static bir build ya
   da bir plugin komutu hizmet vermez. Bu yüzden onlar için de `App.Shutdown`'ı
   çağıran bir şey yoktur. Scaffold edilmiş bir `main.go`, build ya da komut
   bittiğinde onu çağırır (v0.55.0'dan beri).

## Bir plugin'i test etmek

Bir plugin'i, bir uygulamanın onu kullanacağı şekilde test edin: gerçek bir
uygulamanın içinde, onun handler'ı üzerinden. `collagetest.NewSite`
(v0.61.0'dan beri) bir testin ihtiyaç duyduğu en az yapılandırmayla böyle bir
uygulama kurar (`{{hoist "head"}}` ve `{{slot "content"}}` içeren bir layout, `/`
adresinde bir `home` page ve kayıtlarını saklayan debug seviyesinde bir logger),
plugin'inizi içine koyar ve başlatır. Böylece `Configure` ile `Init` hemen çalışır
ve hataları testi başarısız kılar.

Test bitince uygulama kapatılır. Hiçbir sunucu
dinlemez ve hiçbir port seçilmez.

```go
package stamp_test

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"example.com/stamp"
)

// newSite builds a one-page application with the plugin in it.
func newSite(t *testing.T, options ...collagetest.SiteOption) *collagetest.Site {
	t.Helper()
	return collagetest.NewSite(t, append([]collagetest.SiteOption{
		collagetest.WithPlugins(stamp.New(stamp.Options{})),
		collagetest.WithTemplates(map[string]string{"home.html": `<p>{{generator}}</p>`}),
	}, options...)...)
}

func TestStamp_UsesItsDefaults(t *testing.T) {
	site := newSite(t)
	body := site.Get("/").WantStatus(http.StatusOK).Body
	if !strings.Contains(body, `<meta name="generator" content="collage">`) {
		t.Errorf("no default meta tag in %s", body)
	}
	if site.Logs.Has(slog.LevelError) {
		t.Errorf("the plugin logged an error:\n%s", site.Logs)
	}
}

func TestStamp_ReadsItsConfiguration(t *testing.T) {
	site := newSite(t, collagetest.WithPluginConfig(stamp.Name, `{"generator": "my site"}`))
	body := site.Get("/").WantStatus(http.StatusOK).Body
	if !strings.Contains(body, `<meta name="generator" content="my site">`) {
		t.Errorf("no configured meta tag in %s", body)
	}
	if !strings.Contains(body, "<p>my site</p>") {
		t.Errorf("template function not applied in %s", body)
	}
}

func TestStamp_RefusesAnUnknownKey(t *testing.T) {
	_, err := collagetest.TrySite(t,
		collagetest.WithPlugins(stamp.New(stamp.Options{})),
		collagetest.WithPluginConfig(stamp.Name, `{"generater": "typo"}`),
	)
	if err == nil || !strings.Contains(err.Error(), `unknown key "generater"`) {
		t.Errorf("err = %v, want the unknown key named", err)
	}
}

func TestStamp_RegistersItsCommand(t *testing.T) {
	commands := newSite(t).App.Commands()
	if len(commands) != 1 || commands[0].Name != "pages" {
		t.Errorf("commands = %v, want one named pages", commands)
	}
}

func TestStamp_RefusesLateRegistration(t *testing.T) {
	app, err := collage.New(&collage.Config{Template: collage.TemplateConfig{Root: t.TempDir()}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := app.RegisterPlugin(stamp.New(stamp.Options{})); !errors.Is(err, collage.ErrConfigurerRegisteredLate) {
		t.Errorf("RegisterPlugin: err = %v, want ErrConfigurerRegisteredLate", err)
	}
}
```

Bir `Site` üç şeyi bir arada sunar: başlatılmış `App`; onun handler'ı üzerinde
çalışan bir `collagetest.Client` (yani `Get`, `Submit`, `Follow` ve diğer bütün
[client](/docs/testing#the-client) metotları onun üzerinde çalışır); ve uygulamanın
debug seviyesinde ve üstünde yazdığı her kaydı tutan `Logs`.

`Logs.Has(level, parts...)`, o seviyede ya da
üstünde, verilen her kısmı içeren bir kayıt olup olmadığını söyler ve uygulama
başka goroutine'lerden request'lere cevap verirken çağrılması güvenlidir. Bir test,
plugin'inin ihtiyaç duyduğunu ekler:

| Seçenek | Etkisi |
| --- | --- |
| `WithPlugins(plugins...)` | Plugin'leri sırasıyla ekler |
| `WithPluginConfig(name, json)` | Bir plugin'in bölümünü `plugins-config.json`'ın yapacağı gibi atar. Geçersiz JSON testi başarısız kılar |
| `WithTemplates(files)` | Root altındaki template'leri adlarıyla ekler ya da değiştirir: `"layout.html"`, `"home.html"` ya da register edilen bir page'in render ettiği bir template |
| `WithRegister(items...)` | Home page'den sonra page, document ve action register eder |
| `WithoutHome()` | Home page'i dışarıda bırakır; `/` testin olur |
| `WithConfig(func(*collage.Config))` | Uygulama kurulmadan önce yapılandırmayı düzenler: locale'ler, bir cache, `BaseURL`, development modu, ne gerekirse |
| `WithSetup(func(*collage.App) error)` | Uygulama başlamadan önce doğrudan uygulamanın üzerinde register eder — `Handle`, `Mount`, `Use`, `RegisterNotFoundPage`, `RegisterErrorPage` (v0.61.1'den beri). Varsayılan register'lardan sonra çalışır; birden fazlası sırayla çalışır; bir hata site'ı başarısız kılar |

`collagetest.TrySite`, kurulmasının ya da başlamasının başarısız olması beklenen —
örneğin kötü bir yapılandırmayı reddeden bir plugin'i deneyen — bir test için
`NewSite`'tır: testi başarısız kılmak yerine hatayı döner. Site kurulurken bir
`WithRegister` öğesi, bir `WithSetup` ya da `Start` başarısız olursa önce uygulamayı
kapatır (v0.61.3'ten beri); böylece aynı plugin değeri bir sonraki site için yeniden
kullanılabilir.

Bir `collagetest.Site` kopyalanabilir (`s := *site`); `go vet`'in
copylocks denetimi artık itiraz etmez.

Uygulamayı kendisi kapatan bir test — sizin `Shutdown`'ınızın ne döndüğünü kontrol
etmek için — `site.App.Shutdown(ctx)` değil `site.Shutdown(ctx)` çağırır
(v0.61.1'den beri). `App.Shutdown` bir kez çalışır ve sonraki her çağrıda ilk
hatasını döner; bu yüzden cleanup, testin beklediği hatayı yeniden bildirirdi.
`site.Shutdown` site'ı kapatır ve cleanup başka bir şey yapmaz.

Test etmeye değer ve kolayca unutulan birkaç şey:

- **Yapılandırılmamış durum.** Çoğu uygulama sizin plugin'iniz için hiçbir zaman bir
  config bölümü yazmaz.
- **Document'lar**, `OnDocumentRendered` implement ediyorsanız. Bir document
  register edin ve ona request atın, çünkü yalnızca page'lerle test edilen bir plugin
  hiç document görmez. `OnError` implement ediyorsanız başarısız olan bir document'a
  da request atın.
- **`-race` ile çalıştırın.** Bir event'in canlı `Page`'i üzerinden yazan bir hook
  bir data race'tir. Bunu race detector raporlar, başka hiçbir şey raporlamaz.
- **Static export**, dosya üretiyorsanız. Bir `t.TempDir()` içine
  `collage.Build(ctx, app, ...)` ile export alın. Bu, export edilen sitede sunulan
  sitedeki her şeyin olup olmadığını gösterir. Bkz. [Test yazmak](/docs/testing) ve
  [Static export](/docs/static-export).

## Örnek: bir kataloğun referrer policy'si, sıfırdan

Bir ürün kataloğu tedarikçilerine, bir ödeme sağlayıcısına ve yorum sitelerine link
verir. URL'leri bir müşterinin neye baktığını ele verir
(`/products/{slug}?colour=red&campaign=autumn`). Ekip, her page'in bir
`Referrer-Policy` header'ıyla tarayıcıya bunun ne kadarını aktarabileceğini
söylemesini istiyor. Policy sunucuda da sitenin export edilmiş kopyasında da aynı
olmalıdır. Footer da bir güvenlik açığının bildirileceği adresi gösterir.

`acme/policy` bunun için yazılmış küçük ama eksiksiz bir plugin'dir:

- **Options**, Go'da ya da `plugins-config.json`'da ayarlanır: policy ve iletişim
  adresi.
- **Middleware**, `Init` içinde `Host.Use` ile eklenir ve header'ı her response'a
  koyar.
- **Bir template fonksiyonu**, `{{securityContact}}`, `Configure` içinde eklenir.
- **Bir build finding'i**, `OnBuildFinished`'ten gelir: static build'in, ayarlanan
  header olmadan yakaladığı bir page.

Güvenlik header'larının tamamını isteyen bir site
[elagoht/secure](/docs/plugins#elagohtsecure) kullanır; `Referrer-Policy` de onun
gönderdikleri arasındadır. Bu plugin, bir plugin'in her kısmı görünsün diye küçük
tutulmuştur.

### Plugin

```go
// Package policy sends a Referrer-Policy header on every response, offers the
// site's security contact to templates, and checks a static build for pages
// that went out without the header.
package policy

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key of its section in plugins-config.json.
const Name = "acme/policy"

// DefaultReferrerPolicy is what is sent when Options.ReferrerPolicy is empty.
const DefaultReferrerPolicy = "strict-origin-when-cross-origin"

// policies are the values a Referrer-Policy header may hold.
var policies = []string{
	"no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin",
	"same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url",
}

// Options is the plugin's configuration. The zero value is a working plugin.
type Options struct {
	// ReferrerPolicy is the header's value; DefaultReferrerPolicy when empty.
	ReferrerPolicy string `json:"referrerPolicy"`
	// Contact is where to report a vulnerability, offered as {{securityContact}}.
	Contact string `json:"contact"`
}

// Plugin is the policy plugin. Construct it with New.
type Plugin struct {
	opts Options
}

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

var (
	_ collage.Plugin            = (*Plugin)(nil)
	_ collage.Configurer        = (*Plugin)(nil)
	_ collage.BuildFinishedHook = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure reads the options and adds the template function: it runs inside
// collage.New, before the templates are parsed.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts)
	if err != nil {
		return err
	}
	if opts.ReferrerPolicy == "" {
		opts.ReferrerPolicy = DefaultReferrerPolicy
	}
	if !slices.Contains(policies, opts.ReferrerPolicy) {
		return fmt.Errorf("%s: %q is not a referrer policy", Name, opts.ReferrerPolicy)
	}
	p.opts = opts
	return host.AddTemplateFunc("securityContact", func() string { return p.opts.Contact })
}

// Init adds the middleware. A plugin in Config.Plugins wraps the application's
// own middleware, so a header the application sets wins over this one.
func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	return host.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Referrer-Policy", p.opts.ReferrerPolicy)
			next.ServeHTTP(w, r)
		})
	})
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBuildFinished checks what a static host will be told to send: the build
// asked the application for every file it wrote, and kept the headers.
func (p *Plugin) OnBuildFinished(_ context.Context, ev *collage.BuildFinishedEvent) error {
	if p.opts.Contact == "" {
		ev.Warn("", "no-contact", "no security contact is configured; set contact in plugins-config.json")
	}
	for _, f := range ev.Files {
		if f.Kind != collage.BuiltPage || f.Status < 200 || f.Status > 299 {
			continue // not a page, or not captured with a 2xx
		}
		if got := f.Headers.Get("Referrer-Policy"); got != p.opts.ReferrerPolicy {
			ev.Error(f.Path, "referrer-policy", fmt.Sprintf("Referrer-Policy is %q, want %q", got, p.opts.ReferrerPolicy))
		}
	}
	return nil
}
```

Her kısmın nerede durduğu yukarıdaki bölümlerden çıkar:

- **Options `Configure` içinde kontrol edilir.** Header'ın taşıyamayacağı bir
  policy, her response'ta gönderilmek yerine bilinmeyen bir key gibi
  `collage.New`'u başarısız kılar.
- **Middleware header'ı `next`'i çağırmadan önce koyar.** `Config.Plugins` içindeki
  bir plugin uygulamanın kendi middleware'ini sarar. Bu yüzden uygulamanın
  middleware'inin ya da bir handler'ın sonradan koyduğu header bunun yerini alır:
  son sözü uygulama söyler.
- **Kontrol `f.Headers`'ı okur.** Build bunu yalnızca `BuildFinishedHook` implement
  eden bir plugin register edildiği için doldurur; bu yüzden sunucuya hiçbir
  maliyeti yoktur. Build'in yakalamadığı ya da 2xx ile cevaplanmamış bir dosya
  atlanır.
- **Eksik iletişim adresi bir uyarı, yanlış header bir hatadır.** Uyarı raporda
  listelenir; hata build'i `collage.ErrBuildFindings` ile başarısız kılar ve
  page'ler her durumda yazılır.

### Uygulamada register edilmesi

Kataloğun `plugins-config.json`'ı:

```json
{
  "acme/policy": {
    "referrerPolicy": "same-origin",
    "contact": "security@catalogue.example"
  }
}
```

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	log.Fatal(err)
}
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{policy.New(policy.Options{})},
	PluginConfig: pluginConfig,
})
if err != nil {
	log.Fatal(err)
}
```

Ardından footer template'i iletişim adresini gösterir:

```html
<footer>Found a vulnerability? Write to {{securityContact}}.</footer>
```

### Testleri

```go
package policy_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"example.com/policy"
)

// newSite builds a one-page application with the plugin in it.
func newSite(t *testing.T, options ...collagetest.SiteOption) *collagetest.Site {
	t.Helper()
	return collagetest.NewSite(t, append([]collagetest.SiteOption{
		collagetest.WithPlugins(policy.New(policy.Options{})),
		collagetest.WithTemplates(map[string]string{"home.html": `<footer>{{securityContact}}</footer>`}),
	}, options...)...)
}

func TestPolicy_SendsTheDefault(t *testing.T) {
	res := newSite(t).Get("/").WantStatus(http.StatusOK)
	if got := res.Header.Get("Referrer-Policy"); got != policy.DefaultReferrerPolicy {
		t.Errorf("Referrer-Policy = %q, want %q", got, policy.DefaultReferrerPolicy)
	}
}

func TestPolicy_ReadsItsConfiguration(t *testing.T) {
	site := newSite(t, collagetest.WithPluginConfig(policy.Name,
		`{"referrerPolicy": "no-referrer", "contact": "security@example.com"}`))
	res := site.Get("/").WantStatus(http.StatusOK)
	if got := res.Header.Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("Referrer-Policy = %q, want no-referrer", got)
	}
	if !strings.Contains(res.Body, "<footer>security@example.com</footer>") {
		t.Errorf("no contact in %s", res.Body)
	}
}

func TestPolicy_RefusesAnUnknownPolicy(t *testing.T) {
	_, err := collagetest.TrySite(t,
		collagetest.WithPlugins(policy.New(policy.Options{})),
		collagetest.WithPluginConfig(policy.Name, `{"referrerPolicy": "never"}`),
	)
	if err == nil || !strings.Contains(err.Error(), "not a referrer policy") {
		t.Errorf("err = %v, want the policy refused", err)
	}
}

func TestPolicy_BuildFindsAnOverriddenHeader(t *testing.T) {
	site := newSite(t,
		collagetest.WithPluginConfig(policy.Name, `{"contact": "security@example.com"}`),
		// The application's own middleware runs inside the plugin's, so its
		// header replaces the plugin's.
		collagetest.WithSetup(func(app *collage.App) error {
			return app.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Referrer-Policy", "unsafe-url")
					next.ServeHTTP(w, r)
				})
			})
		}),
	)
	report, err := collage.Build(context.Background(), site.App, collage.BuildOptions{OutDir: t.TempDir()})
	if !errors.Is(err, collage.ErrBuildFindings) {
		t.Fatalf("Build: err = %v, want ErrBuildFindings", err)
	}
	var found bool
	for _, f := range report.Findings {
		if f.Rule == "referrer-policy" && f.Path == "/" && f.Level == collage.FindingError {
			found = true
		}
	}
	if !found {
		t.Errorf("findings = %+v, want referrer-policy on /", report.Findings)
	}
}

func TestPolicy_BuildWarnsWithoutAContact(t *testing.T) {
	report, err := collage.Build(context.Background(), newSite(t).App, collage.BuildOptions{OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Build: %v", err) // a warning fails nothing
	}
	if len(report.Findings) != 1 || report.Findings[0].Rule != "no-contact" || report.Findings[0].Level != collage.FindingWarning {
		t.Errorf("findings = %+v, want one no-contact warning", report.Findings)
	}
}
```

Her test plugin'in bir sözünü kontrol eder:

| Test | Ne gösterir |
| --- | --- |
| `SendsTheDefault` | Sıfır `Options` çalışan bir plugin'dir |
| `ReadsItsConfiguration` | Config bölümü hem header'a hem template fonksiyonuna ulaşır |
| `RefusesAnUnknownPolicy` | Kötü bir değer `collage.New`'u başarısız kılar, `TrySite` da hatayı döner |
| `BuildFindsAnOverriddenHeader` | Uygulamanın `WithSetup` ile register edilen kendi middleware'i plugin'inkinin içinde çalışır ve kazanır; build page'i raporlar ve `ErrBuildFindings` ile başarısız olur |
| `BuildWarnsWithoutAContact` | Build'in bütünü hakkındaki bir uyarının `Path`'i boştur ve hiçbir şeyi başarısız kılmaz |
