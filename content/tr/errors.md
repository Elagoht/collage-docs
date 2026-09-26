---
description: collage'ın export ettiği bütün error değerleri, nereden geldiklerine göre gruplanmış hâlde; her birinin ne anlama geldiği ve ne yapmanız gerektiğiyle birlikte.
reference: PanicError, ErrUnknownSlot, ErrConflictingData, ErrNoDocumentHandler, ErrRouteParams, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Hatalar

collage hataları sentinel error değerleriyle bildirir. Bu değerler `pkg/collage`
paketinden export edilir ve ayrıntılarla (page, path, alan) wrap edilir. Onları
mesajları karşılaştırarak değil, her zaman `errors.Is` ile eşleştirin:

```go
if err := app.RegisterPage(page); errors.Is(err, collage.ErrDuplicateRoute) {
	// two pages claim one path
}
```

Bunların çoğu **uygulama açılırken** bildirilir. `New` config'i doğrular ve bütün
template'leri parse eder. Register işlemi her page'i doğrular. Uygulama başlarken de
ancak kümenin tamamına bakınca görülebilecek sorunlar kontrol edilir. Böylece sitenin
bir araya getirilişindeki bir hata, onu ilk bulan okuyucuda patlayan bir page olarak
değil, hiç başlamayan bir program olarak ortaya çıkar. `collage dev` altında bu ret,
programın yazdıklarını gösteren bir 503 page'i olarak tarayıcıda görünür. Bkz.
[CLI](/docs/cli#collage-dev).

Aşağıdaki tablolar her hatanın nereden geldiğine göre gruplanmıştır. Mesaj,
sentinel'in wrap işlemi ayrıntı eklemeden önce taşıdığı metindir.

v0.10.0'dan beri aşağıdaki bütün hatalar export edilir. Daha önce action, request
forgery, method, asset, `{{dict}}` ve template-escapes-root hataları ile `ErrNotStatic`
vardı, ama birbirinden ancak mesajlarıyla ayırt edilebiliyordu.
(`ErrTemplateRootMissing` zaten export ediliyordu.)

## Config

Bu hataları `collage.New` döner. `Config.Validate` içinden ya da uygulama kurulurken
ortaya çıkarlar. Bkz. [Config](/docs/configuration#validation).

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilConfig` | `collage: nil config` | `New(nil)` çağrılmıştır. | Bir `*Config` verin; zero value'su da yeterlidir. |
| `ErrInvalidPort` | `collage: invalid port` | `Server.Port`, `1`–`65535` aralığının dışındadır. | `3000` için sıfır bırakın ya da geçerli bir port verin. |
| `ErrEmptyTemplateRoot` | `collage: empty template root` | `Template.Root` boştur ve `Template.FS` nil'dir. | Bu hataya yalnızca `Validate`'i kendiniz çağırırsanız ulaşılır; `New` önce `Root` için varsayılan değeri atar. |
| `ErrTemplateRootMissing` | `collage: template root missing` | `Template.Root` yoktur ya da bir dizin değildir. | En yaygın başlangıç hatasıdır. Çalışma dizinini kontrol edin ya da template'leri embed edin. |
| `ErrTemplateEscapesRoot` | `collage: template escapes root` | `Template.Root` altındaki bir template, onun dışına çıkan bir yere çözülür. Bu, dizinin dışına giden bir symlink'tir. | Dosyaya link vermek yerine onu dizine kopyalayın. |
| `ErrInvalidCacheType` | `collage: invalid cache type` | `Cache.Enabled` açıktır, `Store` yoktur ve `Type` ne `"memory"` ne de `"disk"`'tir. | Yazımı düzeltin ya da bir `Store` verin. |
| `ErrEmptyCacheDir` | `collage: disk cache needs a directory` | `Cache.Type` `"disk"`'tir ve `Cache.Dir` boştur. | `Dir`'i ayarlayın. |
| `ErrEmptyCacheVersion` | `collage: disk cache needs a version` | Bir disk cache'i version olmadan kurulmuştur. | Normalde bu hataya ulaşılmaz: boş bir `Version` binary'den türetilir. |
| `ErrUnsupportedCache` | `collage: unsupported cache type` | Cache tipi framework'ün kurabildiği hiçbir şeye karşılık gelmez. | Normalde bu durum daha önce `ErrInvalidCacheType` olarak yakalanır. |
| `ErrEmptyLocaleDefault` | `collage: empty default locale` | `Locale.Default` boştur. | Bu hataya yalnızca `Validate` üzerinden ulaşılır; `New` onu varsayılan olarak `"en"` yapar. |
| `ErrLocaleDefaultNotSupported` | `collage: default locale not in supported locales` | `Locale.Supported`, `Locale.Default`'u içermez. | Varsayılan locale'i `Supported`'a ekleyin. |
| `ErrNegativeDuration` | `collage: negative duration` | Altı süre alanından biri negatiftir; mesaj hangisi olduğunu söyler. | Varsayılan değer için sıfır kullanın. |

## Plugin'ler ve komutlar

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilPlugin` | `collage: nil plugin` | `nil` bir plugin register edilmiştir. | — |
| `ErrEmptyPluginName` | `collage: empty plugin name` | Bir plugin'in `Name()`'i boştur. | Ona `acme/stamp` gibi bir isim verin. |
| `ErrDuplicatePlugin` | `collage: duplicate plugin` | İki plugin aynı ismi taşır. | Her plugin'i bir kez register edin. |
| `ErrConfigurerRegisteredLate` | `collage: plugin needs Configure and must be supplied in Config.Plugins` | `RegisterPlugin`'e `Configure` aşaması olan bir plugin verilmiştir, oysa bu aşama çoktan geçmiştir. | Plugin'i `Config.Plugins`'e taşıyın. |
| `ErrDuplicateTemplateFunc` | `collage: duplicate plugin template function` | `AddTemplateFunc`, daha önce eklenmiş bir isim için çağrılmıştır. İsmi başka bir plugin ya da aynı plugin ikinci kez eklemiş olabilir. | `AddTemplateFunc` bu hatayı plugin'in `Configure`'una döner; `New` ancak `Configure` onu dönerse başarısız olur. Plugin'ler çakışmaktadır ve `Template.Funcs` bunu önleyemez. Bu yüzden birini kaldırın ya da ismini değiştirin. |
| `ErrUnknownPluginConfig` | `collage: plugin configuration names no registered plugin` | Bir `PluginConfig` key'i register edilmiş hiçbir plugin'le eşleşmez. Bu, uygulama başlarken kontrol edilir. | Neredeyse her zaman key'de bir yazım hatası vardır. |
| `ErrEmptyCommandName` | `collage: empty command name` | `RegisterCommand` isim olmadan çağrılmıştır. | — |
| `ErrDuplicateCommand` | `collage: duplicate command name` | İki komut aynı ismi taşır. | — |
| `ErrNilApp` | `collage: nil app` | `DispatchCommands`'a nil bir `*App` verilmiştir. | — |
| `ErrUnknownCommand` | `collage: unknown command` | `DispatchCommands` hiç argüman almamıştır ya da hiçbir plugin'in register etmediği bir isim almıştır. | Scaffold edilen `main.go` bu hatada bir kullanım hatası olarak `2` koduyla çıkar. Sunucu olarak çalışmayı tercih eden bir program ise bu hatada bir sonraki adıma geçebilir. |

Bkz. [Plugin yazmak](/docs/writing-plugins).

## Fragment'ler ve page'ler

Bu hataları iki yer bildirir. Birkaçını builder'lar zincir çalışırken kaydeder:
`WithSlot`, `WithSlotResolver` ve `WithSlotFragment` `ErrDuplicateSlot` ve
`ErrSlotResolved`'u kaydeder (`WithSlotFragment` ayrıca `Bind`'ın
`ErrNilFragment` ve `ErrSlotOccupied` hatalarını da kaydeder). `WithTimeout`
`ErrInvalidTimeout`'u, bir page'in `Build`'i `ErrMissingContent`'i, bir document'ın
`Build`'i de `ErrNoDocumentHandler`'ı kaydeder. Bunları `BuildErr()` ile
okuyabilirsiniz. Bir builder'ın kaydettiği hatalar, kurduğu değerin üzerinde kalır.
`RegisterPage` ve `RegisterDocument`, böyle bir hata taşıyan değeri reddeder. Hata
page'in kendisine ya da ağacındaki herhangi bir fragment'e ait olabilir. Red,
`collage: page %q was built with errors: %w` olarak wrap edilir (document için
`collage: document %q was built with errors: %w`). Bu, `BuildErr()` çağrılmış olsun
ya da olmasın geçerlidir.

Geri kalanları, page register edilirken validation bulur. `RegisterPage` ağaçtaki
her fragment'i kontrol eder: isimlerini, template'lerini, slot'larını (her bağlamayı
template'in çağırdığı slot'larla karşılaştırarak), verilerini, timeout'larını,
TTL'lerini, path'lerini ve error page'lerini. İlk hatayı döner. `WithFragmentPath`
ile açılan bir fragment de bu açıdan page'in bir parçası sayılır (v0.11.0'dan beri).
Onun template'i, builder hataları ve validation'ı da diğerleri gibi register sırasında
kontrol edilir. Her iki durumda da hatalı bir page, hiçbir şey serve etmeden
reddedilir.

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrEmptyName` | `collage: empty name` | Bir fragment'in ya da page'in ismi yoktur. |
| `ErrEmptyTemplatePath` | `collage: empty template path` | Bir fragment hiçbir template belirtmez. |
| `ErrNilFragment` | `collage: nil fragment` | Fragment gereken bir yerde `nil` bir fragment kullanılmıştır. Bu bir slot'a bağlanmış ya da bir slot resolver tarafından döndürülmüş olabilir. |
| `ErrDuplicateSlot` | `collage: slot already declared` | `WithSlot` aynı isimle iki kez çağrılmıştır. |
| `ErrUnknownSlot` | `collage: unknown slot` | Bir fragment, template'inin hiç çağırmadığı bir slot'a bağlanmıştır; bağlamanın iki tarafından birinde yazım hatası vardır. Mesaj, slot'u ve template'in gerçekten çağırdığı slot'ları adlarıyla belirtir. Template'in çağırdığı ama hiçbir şeyin doldurmadığı bir slot hata değildir: boş render edilir. |
| `ErrInvalidSlotDefinition` | `collage: invalid slot definition` | Bir slot'un ismi boştur ya da map key'i slot'un kendi ismiyle eşleşmez. |
| `ErrSlotOccupied` | `collage: slot already occupied` | Zaten bir fragment taşıyan bir slot'a ikinci bir fragment bağlanmıştır. Ya da bir resolver o slot için birden fazla fragment dönmüştür. |
| `ErrSlotResolved` | `collage: slot is filled by a resolver` | Aynı slot'a hem bir resolver hem de bağlanmış fragment'ler verilmiştir. |
| `ErrRequiredSlotUnfilled` | `collage: required slot has no fill` | Required olarak tanımlanmış bir slot'a hiçbir şey bağlanmamıştır. |
| `ErrFragmentCycle` | `collage: fragment cycle detected` | Bir fragment'e kendisinden ulaşılabilir. |
| `ErrMissingContent` | `collage: missing content` | Bir page'in content fragment'i yoktur. |
| `ErrConflictingData` | `collage: fixed data and a handler are both set` | Bir fragment hem `WithData` hem `WithDataHandler` ayarlamıştır. |
| `ErrInvalidTimeout` | `collage: invalid timeout` | Bir fragment'in timeout'u negatiftir. |
| `ErrMissingTTL` | `collage: missing cache ttl for incremental strategy` | `Incremental`'a sıfır bir TTL verilmiştir. |
| `ErrInvalidTTL` | `collage: invalid cache ttl` | Bir page'in TTL'i negatiftir. |
| `ErrInvalidPath` | `collage: invalid path` | Bir path pattern'i `/` ile başlamaz. |
| `ErrInvalidRedirectStatus` | `collage: invalid redirect status code` | Redirect status'u `0`, `301`, `302`, `307` ya da `308` dışında bir değerdir. |
| `ErrSelfErrorPage` | `collage: page cannot reference itself as an error page` | Bir page kendi not-found ya da error page'idir. |

Bkz. [Page'ler ve layout'lar](/docs/pages-and-layouts) ve
[Fragment'ler ve slot'lar](/docs/fragments-and-slots).

## Register ve routing

Bu hataları `RegisterPage`, `RegisterNotFoundPage`, `RegisterErrorPage` ve
`RegisterDocument` döner. Bir kısmı da uygulama başlarken ortaya çıkar.

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrAppStarted` | `collage: application already started` | Uygulama başladıktan sonra bir register metodu çağrılmıştır. Başarısız olan herhangi bir başlatmadan sonra çağrılan `RegisterPlugin` de buna dahildir (v0.12.0'dan beri; bir plugin'in `Init`'inde başarısız olan başlatmadan sonrası için v0.11.0'dan beri). | `Handler`, `ListenAndServe`, `Start`, `RenderPath` ya da `DispatchCommands`'tan önce her şeyi register edin. |
| `ErrNilPage` | `collage: nil page` | `nil` bir page register edilmiştir. | — |
| `ErrDuplicatePage` | `collage: duplicate page name` | İki page aynı ismi taşır. | Link'ler page'leri isimleriyle bulur; isimleri benzersiz yapın. |
| `ErrTemplateNotFound` | `collage: template not found` | Bir page'in fragment'i, yüklenmemiş bir template belirtir. | Path'i `Template.Root`'a göre, uzantısıyla birlikte kontrol edin. |
| `ErrUnregisteredErrorPage` | `collage: error page not registered` | Bir page, hiç register edilmemiş bir not-found ya da error page belirtir. Bu, başlarken kontrol edilir. | Onu `RegisterNotFoundPage` ya da `RegisterErrorPage` ile register edin. Register edilmemiş bir error page, gerektiğinde boş render edilirdi. |
| `ErrInvalidPattern` | `collage: invalid pattern` | Bir path ya da redirect kaynağı hatalıdır: başta `/` yoktur, boş bir segment vardır, placeholder ismi boştur, catch-all en sonda değildir ya da (v0.11.0'dan beri) bir segment'in içinde placeholder vardır, örneğin `/feeds/{category}.xml`. | Pattern'i düzeltin. Bir placeholder segment'in tamamını kaplar: `/feeds/{category}/rss.xml`. |
| `ErrDuplicateRoute` | `collage: duplicate route` | Bir path ya da redirect kaynağı o locale'de zaten register edilmiştir. v0.18.0'dan beri bir action'ın bir page'in ya da bir document'ın path'inde `GET` veya `HEAD`'e cevap vermesi de bu hatayı verir; register sırası fark etmez. | Genellikle sebep, page'inin path'iyle aynı yazılmış bir fragment path'idir. Ona kendine ait bir path verin. |
| `ErrAmbiguousParameterName` | `collage: ambiguous parameter name` | İki pattern aynı pozisyonda farklı parametre isimleri kullanır, örneğin `/blog/{slug}` ve `/blog/{id}/edit`. | O pozisyonda tek bir isim kullanın. |
| `ErrRedirectShadowsPage` | `collage: redirect shadows a registered page` | Bir redirect'in kaynağı aynı zamanda bir page'in path'idir. | İkisinden birine ulaşılamazdı; birini kaldırın. |
| `ErrUnsubstitutedPlaceholder` | `collage: redirect placeholder not captured by from pattern` | Bir redirect'in hedefi, kaynağının capture etmediği bir `{name}` kullanır. | Onu kaynakta capture edin ya da kaldırın. |

## Document'lar

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrNilDocument` | `collage: nil document` | `nil` bir document register edilmiştir. |
| `ErrEmptyContentType` | `collage: empty content type` | Bir document hiçbir content type tanımlamaz. Content type zorunludur ve asla tahmin edilmez. |
| `ErrNoDocumentHandler` | `collage: document has no handler or body` | Bir document'ın ne handler'ı ne de sabit bir body'si (`WithBody`) vardır. Page'in aksine, geri düşebileceği bir template'i yoktur. |
| `ErrConflictingData` | bkz. [Fragment'ler ve page'ler](#fragments-and-pages) | Bir document'ta hem `WithHandler` hem `WithBody` vardır. |
| `ErrDuplicateDocument` | `collage: duplicate document name` | İki document aynı ismi taşır. |
| `ErrDocumentNotFound` | `collage: no document at path` | `RenderDocumentPath` path'te hiçbir document bulamamıştır. O path'te bir page ya da redirect olması da buna dahildir. |
| `ErrEmptyDocumentBody` | `collage: document handler produced an empty body` | Bir handler başarılı olmuş ama boş bir body dönmüştür. Serve edilirken bu bir 500'dür ve build bu document'ı yazmaz. Gerçekten "boş" demek isteyen bir handler tek bir newline dönebilir. |

Bkz. [Document'lar](/docs/documents).

## Mount'lar ve handler'lar

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrUnknownAsset` | `collage: unknown asset` | `rc.Asset`, `rc.HoistStylesheet`, `{{asset}}` ya da `{{stylesheet}}`'e hiçbir mount'un serve etmediği bir path verilmiştir. Bir template'te bu durumda render başarısız olur. |
| `ErrNoMountForAsset` | `collage: no mount serves that asset` | Hiçbir mount'un prefix'i path'i hiç kapsamadığında `ErrUnknownAsset`'in içinde wrap edilir. Bu, mount'un olup böyle bir dosyanın olmadığı durumdan farklıdır. |
| `ErrInvalidPrefix` | `collage: invalid mount prefix` | Bir mount prefix'i `/` ile başlamaz ya da bitmez, tek başına `/`'dır ya da `//` ile başlar. `/`'daki bir mount bütün route'ları yutardı. |
| `ErrNilFS` | `collage: nil mount file system` | Bir mount'a hiçbir filesystem verilmemiştir. |
| `ErrMountConflict` | `collage: mount prefixes overlap` | İki mount ya da bir mount ile bir `App.Handle` prefix'i çakışan prefix'ler talep eder. |
| `ErrMountShadowsRoute` | `collage: mount shadows a route` | Bir mount prefix'i ya da bir `App.Handle` prefix'i bir route'un path'ini yutardı. Bu route bir page'in, bir document'ın, bir redirect'in ya da bir action'ın olabilir. Register sırası ne olursa olsun, başlarken kontrol edilir. |
| `ErrInvalidHandlerPrefix` | `collage: handler prefix must begin and end with "/" and not be "/"` | `App.Handle`'a hatalı bir prefix verilmiştir. |
| `ErrNilHandler` | `collage: nil handler` | `App.Handle`'a ya da `App.Use`'a çalıştırılacak hiçbir şey verilmemiştir. |

Bkz. [Static asset'ler](/docs/assets) ve
[Middleware ve kendi API'niz](/docs/middleware-and-apis).

## Render

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNotFound` | `collage: not found` | **Bu hatayı siz dönersiniz.** Onu wrap eden bir data handler, içeriğin yüklenemediğini değil, var olmadığını söyler. | `fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)`, required bir fragment'in hatasını 500 yerine not-found page'li bir 404'e çevirir. |
| `ErrRequiredSlotEmpty` | `collage: required slot is empty` | Render sırasında required bir slot boştur. Bu genellikle bir resolver hiç fragment dönmediğinde olur. | Fragment'in failure policy'sine tabidir. |
| `ErrMaxDepthExceeded` | `collage: max fragment depth exceeded` | Fragment ağacı engine'in izin verdiğinden daha derin iç içe geçer. | Neredeyse her zaman sebep, doğrudan ya da dolaylı olarak kendi slot'una bağlanmış bir fragment'tir. |
| `ErrNoRootFragment` | `collage: page has no root fragment` | Bir page'in ne layout'u ne de içeriği vardır. | — |
| `ErrPageNotFound` | `collage: no page at path` | `App.RenderPath` path'te hiçbir page bulamamıştır. | — |
| `ErrOnceTypeMismatch` | `collage: once key fetched as two different types` | Tek bir render içinde iki `collage.Once` çağrısı aynı key'i farklı tiplerle istemiştir. | Key'ler çakışmaktadır; onlara namespace verin. |
| `ErrCachedTypeMismatch` | `collage: Cached key holds a value of a different type` | İki `collage.Cached` çağrısı aynı key'i farklı tiplerle istemiştir. | Yukarıdakiyle aynı. |
| `ErrDictOddArgs` | `collage: dict requires an even number of arguments` | `{{dict}}`'e değeri olmayan bir key verilmiştir. | Her key'i bir değerle eşleştirin. |
| `ErrDictKeyNotString` | `collage: dict key must be a string` | Bir `{{dict}}` key'i string değildir. | Key'i tırnak içine alın. |
| `ErrCSRFDisabled` | `collage: csrfToken used but request-forgery protection is disabled` | `Security.DisableCSRF` ayarlanmış bir uygulamada bir template `{{csrfToken}}` çağırır. | Çağrıyı kaldırın ya da korumayı yeniden açın. |

Bir data handler'daki, bir slot resolver'daki ya da bir template fonksiyonundaki panic
process'i çökertmez. Panic bir `collage.PanicError`'a dönüşür ve fragment diğer her
hatada olduğu gibi başarısız olur. Panic değerine ve stack'ine ulaşmak için ona
`errors.As` ile erişin:

```go
var panicked *collage.PanicError
if errors.As(err, &panicked) {
	log.Printf("panic: %v", panicked.Value)
}
```

Bir hatanın nasıl 404'e, 500'e, fallback'e ya da hiçbir şeye dönüştüğünü görmek için
[Data handler'lar](/docs/data-handlers) sayfasına bakın.

Development'ta built-in 500 page'i hatanın sebebiyle başlar. Template kaynaklı bir
hatada, başarısız olan çağrının template'ini, satırını, sütununu ve ne döndüğünü
gösterir. Hatanın yukarı çıkarken geçtiği template zinciri bunun altında
yer alır. Production page'i değişmemiştir ve sebep hakkında hiçbir şey söylemez.

## Link'ler ve URL'ler

Bu hataları `App.URL` ve `App.FragmentURL` döner. `{{pageURL}}`, `{{pageURLIn}}`,
`{{localeURL}}`, `{{fragmentURL}}` ve `{{fragmentURLIn}}` ise başarısız olan
render'larla bildirir. Bkz. [Link'ler ve locale'ler](/docs/links-and-locales).

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrUnknownRoute` | `collage: no page or document by that name` | O isimle register edilmiş hiçbir page ya da document yoktur. Ya da ikisi birden vardır ve link belirsizdir. |
| `ErrNoPathInLocale` | `collage: no path in that locale` | Route'un istenen locale'de bir path'i yoktur. `{{pageURL}}` bu durumda varsayılan locale'e fallback yapar, `{{localeURL}}` ise boş string render eder. |
| `ErrRouteParams` | `collage: route parameters do not match the pattern` | Bir parametre eksik ya da boştur, hiçbir placeholder'a karşılık gelmez, `.` ya da `..`'dır, ya da template tek sayıda argüman vermiştir. |
| `ErrLocaleUnreachable` | `collage: no URL reaches that locale` | Locale `Locale.Supported`'da yoktur ya da varsayılan locale değildir ve path locale'leri kapalıdır. |
| `ErrUnknownFragmentPath` | `collage: the page opened no fragment path by that name` | Page, `WithFragmentPath` ile o adda bir fragment açmamıştır (v0.18.0'dan beri). |
| `ErrAmbiguousFragmentPath` | `collage: the fragment is opened at more than one path` | Page fragment'i o locale'de iki path'te açmıştır, bu yüzden link birini seçemez (v0.18.0'dan beri). |

## Action'lar

Bu hataları `RegisterAction` döner. `RegisterPage` de bir page'e eklenmiş action'ı
kontrol eder: `ErrNilAction`, `ErrNoMethods` ve v0.10.0'dan beri `ErrNoActionHandler`
ile `ErrDuplicateAction`. `nil` bir fragment path'ini ise `ErrNilFragmentPath` ile
reddeder. v0.11.0'dan beri tek bir `ErrNoActionHandler` vardır; register sırasında
ve request'te aynı değerdir.

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilAction` | `collage: nil action` | `nil` bir action register edilmiş ya da bir page'e eklenmiştir. | — |
| `ErrEmptyActionName` | `collage: action has no name` | Bir action'ın ismi yoktur. | Ona bir isim verin; log'larda ve hatalarda action'ı bu isim tanımlar. |
| `ErrDuplicateAction` | `collage: duplicate action` | Alınmış bir isimle ikinci bir action register edilmiştir. | — |
| `ErrNoActionPaths` | `collage: action has no paths` | Bağımsız bir action'ın `WithPath`'i yoktur. | Bir page'e bağlı action page'in path'lerini alır; tek başına duran bir action'ın kendi path'leri olmalıdır. |
| `ErrNoActionHandler` | `collage: action has no handler` | Bir action'ın `WithHandler`'ı yoktur. | — |
| `ErrNoMethods` | `collage: action declares no methods` | Bir action hiçbir method'a cevap vermez. | `WithMethods(http.MethodPost)` kullanın ya da page'de `WithAction` kullanın. |
| `ErrNilFragmentPath` | `collage: fragment path has no fragment` | `WithFragmentPath`'e `nil` bir fragment verilmiştir. | — |
| `ErrUnregisteredPage` | `collage: action answered with a page that was never registered` | Bir action'ın `RenderPage`'i register edilmemiş bir page dönmüştür. Request 500 ile başarısız olur. | Page'i register edin ve handler içinde yeni bir page kurmak yerine aynı değerle cevap verin. |

Bkz. [Form'lar ve action'lar](/docs/forms-and-actions).

## Vary ve SkipCache

Bu hataları `collage.Vary` ve `collage.SkipCache` döner. Bu iki fonksiyon middleware
içinden çağrılmalıdır. Bkz. [Caching](/docs/caching) ve [Preview'lar](/docs/previews).

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrVaryTooLate` | `collage: Vary or SkipCache called after routing; call it from middleware` | `Vary` ya da `SkipCache` routing başladıktan sonra çağrılmıştır, örneğin bir data handler'dan. v0.11.0'dan beri bu her route'ta geçerlidir. Öncesinde yalnızca cache'lenen bir page'de geçerliydi ve başka yerlerdeki geç bir çağrı hiçbir şey yapmıyordu. | Onu `App.Use` ile register edilmiş bir middleware'den çağırın. |
| `ErrVaryOutsideRequest` | `collage: Vary called on a request collage is not serving` | `Vary` ya da `SkipCache`, collage'ın handler'ından geçmemiş bir request üzerinde çağrılmıştır. | — |

## Request forgery

Forgery kontrolünden geçemeyen bir gönderim, action'ın handler'ı çalışmadan önce
**403** ile cevaplanır. Error hook'ları sebebi `"route"` stage'i altında, action'ın
ismiyle wrap edilmiş olarak alır:

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrCSRFMissing` | `collage: no csrf token` | Gönderimde token ya da cookie yoktur. |
| `ErrCSRFMismatch` | `collage: csrf token does not match` | Token, cookie'deki token değildir. |
| `ErrCSRFInvalid` | `collage: csrf token is not valid` | Token bu uygulamanın key'iyle imzalanmamıştır. |

Yaygın sebepler iki tanedir: `{{csrfToken}}` içermeyen bir form ve her process'te
yeniden üretilen bir key. Token'ların restart'lardan sonra da geçerli kalması ve
instance'lar arasında çalışması için `Security.CSRFKey`'i ayarlayın. Boyut limitini
aşan bir body ise **413** ile cevaplanır; limite takılan şey token'ın okunması olsa
bile böyledir. Bkz. [Form'lar ve action'lar](/docs/forms-and-actions#forgery-protection).

`Security.DisableCSRF` ayarlanmış bir uygulamada `{{csrfToken}}` kullanmak
`ErrCSRFDisabled` hatasına yol açar; bkz. [Render](#rendering).

## Error hook'larına bildirilenler

Bu hatalar bir plugin'in `OnError`'una `ErrorEvent.Err` içinde ulaşır. Böylece
plugin, mesajları okumadan hataları birbirinden ayırabilir. Bkz.
[Plugin yazmak](/docs/writing-plugins#errorhook).

| Hata | Mesaj | Stage | Anlamı |
| --- | --- | --- | --- |
| `ErrNoRoute` | `collage: no route matched the request` | `not_found` | Hiçbir route request'le eşleşmemiştir; bu bir link ya da routing sorunudur. Bilerek `ErrNotFound` değildir: ikisi de 404'tür, ama sebepleri farklıdır. |
| `ErrNotFound` | bkz. [Render](#rendering) | `render` | Required bir fragment'in içeriği yoktur; bu bir içerik sorunudur. |
| `ErrMethodNotAllowed` | `collage: method not allowed` | `route` | Path vardır ama o method'a cevap vermez. Sonuç, cevap verdiği method'ları listeleyen bir `Allow` header'ıyla birlikte 405'tir. Bir document'ın URL'sinde bu 405 düz metindir (v0.11.0'dan beri). |
| `ErrEmptyRender` | `collage: page rendered no markup` | `render` | Bir page başarıyla render edilmiş ama hiç markup üretmemiştir. Page ister serve edilsin ister bir action'ın cevabı olsun, sonuç 500'dür. Static build'in kaydettiği sentinel de budur. |
| `ErrCSRFMissing`, `ErrCSRFMismatch`, `ErrCSRFInvalid` | bkz. [yukarıda](#request-forgery) | `route` | Forgery kontrolünün reddettiği bir gönderim. |
| `ErrEmptyErrorPage` | `collage: error page rendered empty` | `error_page` | Register edilmiş bir error page başarıyla render edilmiş ama hiç markup üretmemiştir. Bu yüzden onun yerine built-in page serve edilmiştir. |
| `ErrPanic` | `collage: panic recovered while serving the request` | `panic` | Serve sırasında bir şey panic etmiş ve bu panic recover edilip 500'e çevrilmiştir. Bu bir `Cache`, `Metrics` ya da `Tracer` implementasyonu, bir router veya bir plugin hook'u olabilir. Data handler'lardaki ve template'lerdeki panic'ler ise `PanicError` olur. |
| `ErrAssetFailed` | `collage: asset request failed` | `asset` | Mount edilmiş bir dosya request'i 400 ya da üstü bir status ile cevaplanmıştır. Bu tür bütün status'lar için tek bir sentinel vardır. |
| `ErrHandlerFailed` | `collage: mounted handler failed` | `handler` | `App.Handle` ile mount edilmiş bir handler server error ile cevap vermiştir. |
| `ErrUnsafeRedirectTarget` | `collage: unsafe redirect target` | `route` | Bir redirect'in hedefi, substitution'dan sonra tek slash'le başlayan relative bir path değildir: `//host`, `/\host` ya da control character içeren bir path'tir. Bir `Location` header'ıyla değil, 500 ile cevaplanır. |

Alarm kurmaya değer stage `"error_page"`'dir. Hataları bildiren page'in kendisi başarısız
olmuştur, ama okuyucu yine de makul görünen bir page görmüştür. Bu yüzden size bunu
başka hiçbir şey söylemez.

## Static build'ler

Bu hataları `collage.NewBuilder` ve `Builder.Build` döner ya da `BuildReport`'a
kaydedilirler. Bkz. [Static export](/docs/static-export).

| Hata | Mesaj | Nerede | Anlamı |
| --- | --- | --- | --- |
| `ErrNilRenderer` | `collage: nil renderer` | `NewBuilder` | App `nil`'dir. |
| `ErrInvalidOutDir` | `collage: invalid output directory` | `NewBuilder` | `BuildOptions.OutDir` boştur. |
| `ErrDangerousOutDir` | `collage: refusing to use a dangerous output directory` | `Build` | `OutDir` bir filesystem root'una çözülür. `Clean` açıksa bir repository root'una çözülmesi de bu hatayı verir. |
| `ErrOutputPathCollision` | `collage: two builds target one output path` | `Build` | İki page aynı dosyaya yazılacaktır. Sebep yalnızca sondaki slash'le ayrılan pattern'ler ya da aynı değerleri iki kez listeleyen bir `WithStaticParams` olabilir. Hiçbir page render edilmeden önce bildirilir ve ardından hiçbir page render edilmez. Document'lar, `404.html` ve asset'ler yine de yazılır. |
| `ErrPathEscapesOutDir` | `collage: resolved path escapes the output directory` | report error | Bir output path'i ya da ona giden yoldaki bir symlink `OutDir`'in dışına çıkar. Bu, o dosya yazılmadan önce kontrol edilir; yalnızca o path başarısız olur. |
| `ErrDynamicPathUnresolved` | `collage: a path pattern with a {param} needs WithStaticParams to be built` | skip | Bir page'in ya da document'ın path'inde bir `{param}` vardır ve `WithStaticParams` yoktur. |
| `ErrRouteParams` | bkz. [Link'ler ve URL'ler](#links-and-urls) | report error | `WithStaticParams`'ın döndüğü bir map pattern'i tam olarak doldurmaz: bir isim eksiktir ya da pattern'de olmayan bir isim vardır. Yalnızca o dosya başarısız olur, geri kalanlar build edilir. |
| `ErrNotStatic` | `collage: a Dynamic() route cannot be built statically` | skip | Bir page ya da document dynamic'tir: ya `Dynamic()` olarak tanımlanmıştır ya da strateji tanımlamayıp data handler'ı olan bir şey render eder. Dolayısıyla export edilecek bir şey yoktur. |
| `ErrDuplicateOutputPath` | `collage: two build tasks write the same output path` | skip | İki document görevi aynı dosyaya çözülür; `WithStaticParams` aynı değerleri iki kez listelemiştir. İlki build edilir, diğerleri atlanır. |
| `ErrDegradedRender` | `collage: refusing to write a degraded render` | report error | Bir page başarısız olan bir fragment'le render edilmiştir ve `AllowDegraded` kapalıdır. Hiçbir dosya yazılmaz. |
| `ErrEmptyRender` | `collage: page rendered no markup` | report error | Bir page hiç markup render etmemiştir. `AllowDegraded` açık olsa bile reddedilir. Yukarıda serve için anlatılanla aynı sentinel'dir. |
| `ErrUnresolvedToken` | `collage: refusing to write a page whose forgery token was never resolved` | report error, skip | Not-found page bir `{{csrfToken}}` taşıyorsa bu bir report error'dır. Token taşıyan diğer page'ler ise atlanır, çünkü onlar bir sunucuya ihtiyaç duyar. |
| `ErrBuildPanic` | `collage: panic while building a page` | report error | Bir page'in render edilmesi ya da yazılması ya da bir `WithStaticParams` fonksiyonu panic etmiştir. İkincisi o route'un o locale'ini başarısız kılar. Build bunu recover etmiş ve diğerleriyle devam etmiştir. |
| `ErrEmptyDocumentBody` | bkz. [Document'lar](#documents) | report error | Bir document boş bir body üretmiştir. |

Report error'lar `BuildReport.Errors` içindedir. Bu, `errors.Is` ile eşleşen
hatalardan oluşan bir slice'tır. Bir atlama ise `BuildReport.Skipped` içindeki bir
`SkipRecord`'dur. Onun `Reason`'ı insanlar için yazılmış bir cümledir. `Err`'i ise
(v0.10.0'dan beri) kod için sentinel'dir: `ErrNotStatic`, `ErrDynamicPathUnresolved`,
`ErrUnresolvedToken` ya da `ErrDuplicateOutputPath`. Bu yüzden `Reason`'ı okumak
yerine onu `errors.Is(skip.Err, …)` ile eşleştirin.
