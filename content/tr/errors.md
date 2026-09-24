---
description: collage'ın dışa açtığı her hata değeri, nereden geldiğine göre gruplanmış hâlde; ne anlama geldiği ve ne yapmanız gerektiğiyle birlikte.
---

# Hatalar

collage hataları `pkg/collage`'dan dışa açılan sabit (sentinel) hata değerleriyle
bildirir ve onları ayrıntılarla — sayfa, yol, alan — sarmalar. Bunları mesajları
karşılaştırarak değil, her zaman `errors.Is` ile eşleştirin:

```go
if err := app.RegisterPage(page); errors.Is(err, collage.ErrDuplicateRoute) {
	// two pages claim one path
}
```

Çoğu **başlangıçta** bildirilir: `New` yapılandırmayı doğrular ve her şablonu
ayrıştırır, kayıt her sayfayı doğrular, uygulamanın başlatılması da ancak bütün
kümenin ortaya çıkarabileceği şeyleri denetler. Sitenin bir araya getirilişindeki bir
hata, onu ilk bulan okuyucuda patlayan bir sayfa değil, başlamayı reddeden bir
programdır.

Aşağıdaki tablolar her hatanın nereden geldiğine göre gruplanmıştır. Mesaj, herhangi
bir sarmalama ayrıntı eklemeden önce sentinel'in taşıdığı metindir.

v0.10.0'dan itibaren aşağıdaki her hata dışa açıktır. Öncesinde action, istek
sahteciliği, metot, asset, `{{dict}}` ve şablonun kök dizinden kaçması hataları ile
`ErrNotStatic` vardı, ama birbirinden yalnızca mesajlarıyla ayırt edilebiliyordu.
(`ErrTemplateRootMissing` zaten dışa açıktı.)

## Yapılandırma

`collage.New` tarafından, `Config.Validate`'ten ya da uygulama kurulurken döner.
Bkz. [Yapılandırma](/docs/configuration#validation).

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilConfig` | `collage: nil config` | `New(nil)`. | Bir `*Config` verin; sıfır değeri de olur. |
| `ErrInvalidPort` | `collage: invalid port` | `Server.Port`, `1`–`65535` aralığının dışında. | `3000` için sıfır bırakın ya da geçerli bir port verin. |
| `ErrEmptyTemplateRoot` | `collage: empty template root` | `Template.Root` boş ve `Template.FS` nil. | Yalnızca `Validate`'i kendiniz çağırırsanız karşılaşılır; `New` önce `Root`'a varsayılan değer atar. |
| `ErrTemplateRootMissing` | `collage: template root missing` | `Template.Root` yok ya da bir dizin değil. | En sık görülen başlangıç hatası: çalışma dizinini kontrol edin ya da şablonları gömün. |
| `ErrTemplateEscapesRoot` | `collage: template escapes root` | `Template.Root` altındaki bir şablon, dizinin dışına çözümleniyor — dizinden dışarı çıkan bir sembolik bağlantı. | Dosyaya bağlantı vermek yerine dosyayı içeri kopyalayın. |
| `ErrInvalidCacheType` | `collage: invalid cache type` | `Cache.Enabled` açık, `Store` yok ve `Type` ne `"memory"` ne `"disk"`. | Yazımı düzeltin ya da bir `Store` verin. |
| `ErrEmptyCacheDir` | `collage: disk cache needs a directory` | `Cache.Type` `"disk"` ve `Cache.Dir` boş. | `Dir`'i ayarlayın. |
| `ErrEmptyCacheVersion` | `collage: disk cache needs a version` | Bir disk önbelleği sürüm olmadan kuruldu. | Normalde karşılaşılmaz: boş bir `Version` binary'den türetilir. |
| `ErrUnsupportedCache` | `collage: unsupported cache type` | Önbellek tipi, framework'ün kurabileceği hiçbir şeyi adlandırmıyor. | Normalde daha önce `ErrInvalidCacheType` olarak yakalanır. |
| `ErrEmptyLocaleDefault` | `collage: empty default locale` | `Locale.Default` boş. | Yalnızca `Validate` üzerinden karşılaşılır; `New` varsayılan olarak `"en"` atar. |
| `ErrLocaleDefaultNotSupported` | `collage: default locale not in supported locales` | `Locale.Supported`, `Locale.Default`'u içermiyor. | Varsayılanı `Supported`'a ekleyin. |
| `ErrNegativeDuration` | `collage: negative duration` | Altı süre alanından biri negatif; mesaj hangisi olduğunu söyler. | Varsayılan için sıfır kullanın. |

## Plugin'ler ve komutlar

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilPlugin` | `collage: nil plugin` | `nil` bir plugin kaydedildi. | — |
| `ErrEmptyPluginName` | `collage: empty plugin name` | Bir plugin'in `Name()`'i boş. | `acme/stamp` gibi bir ad verin. |
| `ErrDuplicatePlugin` | `collage: duplicate plugin` | İki plugin aynı adı taşıyor. | Her plugin'i bir kez kaydedin. |
| `ErrConfigurerRegisteredLate` | `collage: plugin needs Configure and must be supplied in Config.Plugins` | `RegisterPlugin`'e `Configure` aşaması olan bir plugin verildi; o aşama çoktan geçti. | Onu `Config.Plugins`'e taşıyın. |
| `ErrDuplicateTemplateFunc` | `collage: duplicate plugin template function` | `AddTemplateFunc` zaten eklenmiş bir ad için çağrıldı — başka bir plugin tarafından ya da aynı plugin tarafından ikinci kez. | `AddTemplateFunc` bunu plugin'in `Configure`'una döndürür; `New` yalnızca `Configure` bunu döndürürse başarısız olur. Plugin'ler çakışıyor: `Template.Funcs` bunu önleyemez, bu yüzden birini çıkarın ya da yeniden adlandırın. |
| `ErrUnknownPluginConfig` | `collage: plugin configuration names no registered plugin` | Bir `PluginConfig` anahtarı kayıtlı hiçbir plugin'le eşleşmiyor. Uygulama başlarken denetlenir. | Neredeyse her zaman anahtardaki bir yazım hatasıdır. |
| `ErrEmptyCommandName` | `collage: empty command name` | Adsız bir `RegisterCommand`. | — |
| `ErrDuplicateCommand` | `collage: duplicate command name` | İki komut aynı adı taşıyor. | — |
| `ErrNilApp` | `collage: nil app` | `DispatchCommands`'a nil bir `*App` verildi. | — |
| `ErrUnknownCommand` | `collage: unknown command` | `DispatchCommands` hiç argüman almadı ya da hiçbir plugin'in kaydetmediği bir ad aldı. | İskelet olarak üretilen `main.go` bunda, bir kullanım hatası olarak `2` ile çıkar; bunun yerine sunmayı tercih eden bir program bu hatada akışı sürdürebilir. |

Bkz. [Plugin yazmak](/docs/writing-plugins).

## Fragment'ler ve sayfalar

Bunları iki yer bildirir. Birkaçını builder'lar zincir çalışırken kaydeder —
`WithSlot`, `WithSlotResolver` ve `WithSlotFragment`'tan `ErrDuplicateSlot`,
`ErrUnknownSlot` ve `ErrSlotResolved` (`WithSlotFragment` ayrıca `Bind`'ın
`ErrNilFragment` ve `ErrSlotOccupied` hatalarını da kaydeder), `WithTimeout`'tan
`ErrInvalidTimeout`, bir sayfanın `Build`'inden `ErrMissingContent` ve bir
document'ın `Build`'inden `ErrNoDocumentHandler` — ve bunları `BuildErr()` ile
okuyabilirsiniz. Bir builder'ın kaydettiği hata, kurduğu değerin üzerinde kalır;
`RegisterPage` ve `RegisterDocument` de böyle bir hata taşıyan — sayfanın kendisine
ya da ağacındaki herhangi bir fragment'e ait — bir değeri, `BuildErr()` çağrılmış
olsun ya da olmasın, `collage: page %q was built with errors: %w` (bir document için
`collage: document %q was built with
errors: %w`) olarak sarmalayıp reddeder.

Geri kalanları, sayfa kaydedilirken doğrulama bulur: `RegisterPage` ağaçtaki her
fragment'i; adlarını, şablonlarını, slot'larını, zaman aşımlarını, TTL'lerini,
yollarını ve hata sayfalarını denetler ve ilk hatayı döndürür. `WithFragmentPath` ile
açılan bir fragment bu açıdan sayfanın bir parçasıdır (v0.11.0'dan itibaren): şablonu,
builder'ının hataları ve doğrulaması, diğerleri gibi kayıt sırasında denetlenir. Her
iki durumda da hatalı kurulmuş bir sayfa, herhangi bir şey sunmadan önce reddedilir.

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrEmptyName` | `collage: empty name` | Bir fragment'in ya da sayfanın adı yok. |
| `ErrEmptyTemplatePath` | `collage: empty template path` | Bir fragment hiçbir şablon adlandırmıyor. |
| `ErrNilFragment` | `collage: nil fragment` | Fragment gereken bir yerde `nil` bir fragment kullanıldı — bir slot'a bağlandı ya da bir slot resolver'ı tarafından döndürüldü. |
| `ErrDuplicateSlot` | `collage: slot already declared` | `WithSlot` aynı adla iki kez çağrıldı. |
| `ErrUnknownSlot` | `collage: unknown slot` | Fragment'in hiç bildirmediği bir slot'a bir fragment bağlandı ya da bir şablon onun için `{{slot}}` çağırdı. |
| `ErrInvalidSlotDefinition` | `collage: invalid slot definition` | Bir slot'un adı boş ya da map anahtarı kendi adıyla eşleşmiyor. |
| `ErrSlotOccupied` | `collage: slot already occupied` | Zaten bir fragment tutan bir slot'a ikinci bir fragment bağlandı — ya da bir resolver onun için birkaç tane döndürdü. |
| `ErrSlotResolved` | `collage: slot is filled by a resolver` | Bir slot'a hem bir resolver hem de bağlı fragment'ler verildi. |
| `ErrRequiredSlotUnfilled` | `collage: required slot has no fill` | Zorunlu olarak bildirilen bir slot'a hiçbir şey bağlanmamış. |
| `ErrFragmentCycle` | `collage: fragment cycle detected` | Bir fragment'e kendisinden ulaşılabiliyor. |
| `ErrMissingContent` | `collage: missing content` | Bir sayfanın içerik fragment'i yok. |
| `ErrInvalidTimeout` | `collage: invalid timeout` | Bir fragment'in zaman aşımı negatif. |
| `ErrMissingTTL` | `collage: missing cache ttl for incremental strategy` | `Incremental`'a sıfır bir TTL verildi. |
| `ErrInvalidTTL` | `collage: invalid cache ttl` | Bir sayfanın TTL'i negatif. |
| `ErrInvalidPath` | `collage: invalid path` | Bir yol pattern'i `/` ile başlamıyor. |
| `ErrInvalidRedirectStatus` | `collage: invalid redirect status code` | `0`, `301`, `302`, `307` ya da `308` dışında bir yönlendirme durum kodu. |
| `ErrSelfErrorPage` | `collage: page cannot reference itself as an error page` | Bir sayfa kendi bulunamadı ya da hata sayfası. |

Bkz. [Sayfalar ve layout'lar](/docs/pages-and-layouts) ve
[Fragment'ler ve slot'lar](/docs/fragments-and-slots).

## Kayıt ve yönlendirme

`RegisterPage`, `RegisterNotFoundPage`, `RegisterErrorPage` ve `RegisterDocument`
tarafından ya da uygulama başlarken döner.

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrAppStarted` | `collage: application already started` | Uygulama başladıktan sonra bir kayıt metodu çağrıldı — başarısız olan herhangi bir başlatmadan sonra çağrılan `RegisterPlugin` dahil (v0.12.0'dan itibaren; bir plugin'in `Init`'inde başarısız olan bir başlatmadan sonra v0.11.0'dan itibaren). | Her şeyi `Handler`, `ListenAndServe`, `Start`, `RenderPath` ya da `DispatchCommands`'tan önce kaydedin. |
| `ErrNilPage` | `collage: nil page` | `nil` bir sayfa kaydedildi. | — |
| `ErrDuplicatePage` | `collage: duplicate page name` | İki sayfa aynı adı taşıyor. | Bağlantılar sayfaları adlarıyla bulur; adları benzersiz yapın. |
| `ErrTemplateNotFound` | `collage: template not found` | Bir sayfanın fragment'i yüklenmemiş bir şablonu adlandırıyor. | Yolu, uzantısıyla birlikte, `Template.Root`'a göre kontrol edin. |
| `ErrUnregisteredErrorPage` | `collage: error page not registered` | Bir sayfa hiç kaydedilmemiş bir bulunamadı ya da hata sayfasını adlandırıyor. Başlangıçta denetlenir. | Onu `RegisterNotFoundPage` ya da `RegisterErrorPage` ile kaydedin; kaydedilmemiş olan, gerektiğinde boş render edilirdi. |
| `ErrInvalidPattern` | `collage: invalid pattern` | Bir yol ya da yönlendirme kaynağı hatalı: başta `/` yok, boş bir segment var, boş bir yer tutucu adı var, catch-all en sonda değil ya da — v0.11.0'dan itibaren — `/feeds/{category}.xml` gibi bir segmentin içinde bir yer tutucu var. | Pattern'i düzeltin. Bir yer tutucu bütün bir segmenttir: `/feeds/{category}/rss.xml`. |
| `ErrDuplicateRoute` | `collage: duplicate route` | Bir yol ya da yönlendirme kaynağı o locale'de zaten kayıtlı. | — |
| `ErrAmbiguousParameterName` | `collage: ambiguous parameter name` | İki pattern aynı konumda farklı parametre adları kullanıyor, örneğin `/blog/{slug}` ve `/blog/{id}/edit`. | O konumda tek bir ad kullanın. |
| `ErrRedirectShadowsPage` | `collage: redirect shadows a registered page` | Bir yönlendirmenin kaynağı aynı zamanda bir sayfanın yolu. | İkisinden birine ulaşılamazdı; birini kaldırın. |
| `ErrUnsubstitutedPlaceholder` | `collage: redirect placeholder not captured by from pattern` | Bir yönlendirmenin hedefi, kaynağının yakalamadığı bir `{name}` kullanıyor. | Onu kaynakta yakalayın ya da kaldırın. |

## Document'lar

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrNilDocument` | `collage: nil document` | `nil` bir document kaydedildi. |
| `ErrEmptyContentType` | `collage: empty content type` | Bir document hiçbir içerik tipi bildirmiyor. İçerik tipi zorunludur ve asla tahmin edilmez. |
| `ErrNoDocumentHandler` | `collage: document has no handler` | Bir document'ın handler'ı yok. Bir sayfanın aksine, geri çekilebileceği bir şablonu yoktur. |
| `ErrDuplicateDocument` | `collage: duplicate document name` | İki document aynı adı taşıyor. |
| `ErrDocumentNotFound` | `collage: no document at path` | `RenderDocumentPath` yolda hiçbir document bulamadı — orada bir sayfa ya da yönlendirme olduğu durum dahil. |
| `ErrEmptyDocumentBody` | `collage: document handler produced an empty body` | Bir handler boş bir gövdeyle başarılı oldu: sunulurken 500'dür ve bir build tarafından yazılmaz. Gerçekten "boş" demek isteyen bir handler tek bir satır sonu döndürebilir. |

Bkz. [Document'lar](/docs/documents).

## Mount'lar ve handler'lar

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrUnknownAsset` | `collage: unknown asset` | `rc.Asset`, `rc.HoistStylesheet`, `{{asset}}` ya da `{{stylesheet}}`'e hiçbir mount'un sunmadığı bir yol verildi. Bir şablonda render başarısız olur. |
| `ErrNoMountForAsset` | `collage: no mount serves that asset` | Hiçbir mount'un önekinin yolu hiç kapsamadığı durumda `ErrUnknownAsset`'in içine sarmalanır — böyle bir dosyası olmayan bir mount'un aksine. |
| `ErrInvalidPrefix` | `collage: invalid mount prefix` | Bir mount öneki `/` ile başlayıp bitmiyor, yalnızca `/`'dan ibaret ya da `//` ile başlıyor. `/`'daki bir mount her route'u yutardı. |
| `ErrNilFS` | `collage: nil mount file system` | Bir mount'a hiçbir dosya sistemi verilmedi. |
| `ErrMountConflict` | `collage: mount prefixes overlap` | İki mount — ya da bir mount ile bir `App.Handle` öneki — çakışan önekler talep ediyor. |
| `ErrMountShadowsRoute` | `collage: mount shadows a route` | Bir mount öneki ya da bir `App.Handle` öneki, bir route'un yolunu — bir sayfanın, bir document'ın, bir yönlendirmenin ya da bir action'ın — yutardı. Kayıt sırası ne olursa olsun başlangıçta denetlenir. |
| `ErrInvalidHandlerPrefix` | `collage: handler prefix must begin and end with "/" and not be "/"` | `App.Handle`'a hatalı bir önek verildi. |
| `ErrNilHandler` | `collage: nil handler` | `App.Handle`'a ya da `App.Use`'a çalıştıracak hiçbir şey verilmedi. |

Bkz. [Statik dosyalar](/docs/assets) ve
[Middleware ve kendi API'niz](/docs/middleware-and-apis).

## Render

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNotFound` | `collage: not found` | **Bunu siz döndürürsünüz.** Bunu sarmalayan bir data handler, içeriğin yüklenemediğini değil, var olmadığını söyler. | `fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)`, zorunlu bir fragment'in hatasını 500 yerine bulunamadı sayfasıyla birlikte bir 404 yapar. |
| `ErrRequiredSlotEmpty` | `collage: required slot is empty` | Render sırasında zorunlu bir slot hiçbir şey tutmuyor — genellikle bir resolver hiç fragment döndürmemiştir. | Fragment'in hata politikasına tabidir. |
| `ErrMaxDepthExceeded` | `collage: max fragment depth exceeded` | Fragment ağacı, motorun izin verdiğinden daha derin iç içe geçiyor. | Neredeyse her zaman, doğrudan ya da dolaylı olarak kendi slot'una bağlanmış bir fragment'tir. |
| `ErrNoRootFragment` | `collage: page has no root fragment` | Bir sayfanın ne layout'u ne de içeriği var. | — |
| `ErrPageNotFound` | `collage: no page at path` | `App.RenderPath` yolda hiçbir sayfa bulamadı. | — |
| `ErrOnceTypeMismatch` | `collage: once key fetched as two different types` | İki `collage.Once` çağrısı, tek bir render içinde aynı anahtarı farklı tiplerle istedi. | Anahtarlar çakışıyor; onlara ad alanı verin. |
| `ErrCachedTypeMismatch` | `collage: Cached key holds a value of a different type` | İki `collage.Cached` çağrısı aynı anahtarı farklı tiplerle istedi. | Yukarıdaki gibi. |
| `ErrDictOddArgs` | `collage: dict requires an even number of arguments` | `{{dict}}`'e değeri olmayan bir anahtar verildi. | Her anahtarı bir değerle eşleştirin. |
| `ErrDictKeyNotString` | `collage: dict key must be a string` | Bir `{{dict}}` anahtarı string değil. | Anahtarı tırnak içine alın. |
| `ErrCSRFDisabled` | `collage: csrfToken used but request-forgery protection is disabled` | `Security.DisableCSRF` ayarlanmış bir uygulamada bir şablon `{{csrfToken}}` çağırıyor. | Çağrıyı kaldırın ya da korumayı yeniden açın. |

Bir data handler'daki, bir slot resolver'ındaki ya da bir şablon fonksiyonundaki
panic süreci çökertmez: bir `collage.PanicError`'a dönüşür ve fragment, diğer her
hatada olduğu gibi başarısız olur. Panic değerini ve yığınını geri almak için ona
`errors.As` ile ulaşın:

```go
var panicked *collage.PanicError
if errors.As(err, &panicked) {
	log.Printf("panic: %v", panicked.Value)
}
```

Bir hatanın nasıl 404'e, 500'e, bir yedeğe ya da hiçliğe dönüştüğü için bkz.
[Data handler'lar](/docs/data-handlers).

## Bağlantılar ve URL'ler

`App.URL` tarafından ve `{{pageURL}}`, `{{pageURLIn}}` ve `{{localeURL}}`'den
kaynaklanan başarısız render'larda döner. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales).

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrUnknownRoute` | `collage: no page or document by that name` | O adla kayıtlı hiçbir sayfa ya da document yok — ya da ikisi de kayıtlı ve bağlantı belirsiz. |
| `ErrNoPathInLocale` | `collage: no path in that locale` | Route'un istenen locale'de yolu yok. `{{pageURL}}` bunun yerine varsayılan locale'e geri döner, `{{localeURL}}` ise boş string render eder. |
| `ErrRouteParams` | `collage: route parameters do not match the pattern` | Bir parametre eksik ya da boş, hiçbir yer tutucuyu adlandırmıyor, `.` ya da `..`, ya da şablon tek sayıda argüman geçti. |
| `ErrLocaleUnreachable` | `collage: no URL reaches that locale` | Locale `Locale.Supported` içinde değil ya da varsayılan değil ve yol locale'leri kapalı. |

## Action'lar

`RegisterAction` tarafından döner. `RegisterPage` bir sayfaya bağlı bir action'ı da
denetler — `ErrNilAction`, `ErrNoMethods` ve v0.10.0'dan itibaren
`ErrNoActionHandler` ile `ErrDuplicateAction` — ve `nil` bir fragment yolunu
`ErrNilFragmentPath` ile reddeder. v0.11.0'dan itibaren tek bir `ErrNoActionHandler`
vardır; kayıtta da bir istekte de aynı değerdir.

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrNilAction` | `collage: nil action` | `nil` bir action kaydedildi ya da bir sayfaya bağlandı. | — |
| `ErrEmptyActionName` | `collage: action has no name` | Bir action'ın adı yok. | Bir ad verin; ad, action'ı loglarda ve hatalarda tanımlar. |
| `ErrDuplicateAction` | `collage: duplicate action` | Zaten alınmış bir adla ikinci bir action kaydedildi. | — |
| `ErrNoActionPaths` | `collage: action has no paths` | Bağımsız bir action'ın `WithPath`'i yok. | Bir sayfadaki action sayfanın yollarını alır; kendi başına olan bir action'ın kendi yollarına ihtiyacı vardır. |
| `ErrNoActionHandler` | `collage: action has no handler` | Bir action'ın `WithHandler`'ı yok. | — |
| `ErrNoMethods` | `collage: action declares no methods` | Bir action hiçbir metoda yanıt vermiyor. | `WithMethods(http.MethodPost)` ya da bir sayfada `WithAction`. |
| `ErrNilFragmentPath` | `collage: fragment path has no fragment` | `WithFragmentPath`'e `nil` bir fragment verildi. | — |
| `ErrUnregisteredPage` | `collage: action answered with a page that was never registered` | Bir action'ın `RenderPage`'i kaydedilmemiş bir sayfa döndürdü. İstek 500 ile başarısız olur. | Sayfayı kaydedin ve handler içinde yeni bir tane kurmak yerine o aynı değerle yanıt verin. |

Bkz. [Formlar ve action'lar](/docs/forms-and-actions).

## Vary ve SkipCache

Middleware'den çağrılması gereken `collage.Vary` ve `collage.SkipCache` tarafından
döner. Bkz. [Önbellekleme](/docs/caching) ve [Önizlemeler](/docs/previews).

| Hata | Mesaj | Anlamı | Ne yapmalı |
| --- | --- | --- | --- |
| `ErrVaryTooLate` | `collage: Vary or SkipCache called after routing; call it from middleware` | `Vary` ya da `SkipCache` yönlendirme başladıktan sonra çağrıldı — örneğin bir data handler'dan. v0.11.0'dan itibaren her route'ta; öncesinde yalnızca önbelleğe alınan bir sayfada döner, başka yerdeki geç bir çağrı hiçbir şey yapmazdı. | Onu `App.Use` ile kaydedilmiş middleware'den çağırın. |
| `ErrVaryOutsideRequest` | `collage: Vary called on a request collage is not serving` | `Vary` ya da `SkipCache`, collage'ın handler'ından geçmemiş bir istek üzerinde çağrıldı. | — |

## İstek sahteciliği

Sahtecilik denetiminden geçemeyen bir gönderim, action'ın handler'ı çalışmadan önce
bir **403** ile yanıtlanır; hata hook'ları da nedeni, action'ın adıyla sarmalanmış
olarak `"route"` aşamasında alır:

| Hata | Mesaj | Anlamı |
| --- | --- | --- |
| `ErrCSRFMissing` | `collage: no csrf token` | Gönderim hiçbir token ya da hiçbir cookie taşımıyordu. |
| `ErrCSRFMismatch` | `collage: csrf token does not match` | Token, cookie'deki token değil. |
| `ErrCSRFInvalid` | `collage: csrf token is not valid` | Token bu uygulamanın anahtarıyla imzalanmamış. |

Olağan nedenler, `{{csrfToken}}` içermeyen bir form ve süreç başına üretilen bir
anahtardır — token'ların yeniden başlatmalardan sağ çıkması ve instance'lar arasında
çalışması için `Security.CSRFKey`'i ayarlayın. Boyut sınırını aşan bir gövde ise,
sınıra takılan şey token'ın okunması olsa bile, bunun yerine bir **413**'tür. Bkz.
[Formlar ve action'lar](/docs/forms-and-actions#forgery-protection).

`Security.DisableCSRF` ayarlanmış bir uygulamadaki `{{csrfToken}}`,
`ErrCSRFDisabled`'dır; bkz. [Render](#rendering).

## Hata hook'larına bildirilenler

Bunlar bir plugin'in `OnError`'ına `ErrorEvent.Err` içinde ulaşır; böylece plugin,
mesajları okumadan hataları birbirinden ayırt edebilir. Bkz.
[Plugin yazmak](/docs/writing-plugins#errorhook).

| Hata | Mesaj | Aşama | Anlamı |
| --- | --- | --- | --- |
| `ErrNoRoute` | `collage: no route matched the request` | `not_found` | İstekle hiçbir route eşleşmedi — bir bağlantı ya da yönlendirme sorunu. Bilerek `ErrNotFound` değildir: ikisi de 404'tür ama nedenleri farklıdır. |
| `ErrNotFound` | bkz. [Render](#rendering) | `render` | Zorunlu bir fragment'in içeriği yok — bir içerik sorunu. |
| `ErrMethodNotAllowed` | `collage: method not allowed` | `route` | Yol var ama böyle bir metoda yanıt vermiyor: yanıt verdiği metotları adlandıran bir `Allow` header'ıyla bir 405. Bir document'ın URL'sinde 405 düz metindir (v0.11.0'dan itibaren). |
| `ErrEmptyRender` | `collage: page rendered no markup` | `render` | Bir sayfa başarıyla render edildi ama sunulurken ya da bir action tarafından yanıt olarak verilirken hiçbir markup üretmedi: bir 500. Statik bir build'in kaydettiği sentinel'in aynısıdır. |
| `ErrCSRFMissing`, `ErrCSRFMismatch`, `ErrCSRFInvalid` | bkz. [yukarısı](#request-forgery) | `route` | Sahtecilik denetiminin reddettiği bir gönderim. |
| `ErrEmptyErrorPage` | `collage: error page rendered empty` | `error_page` | Kayıtlı bir hata sayfası başarıyla render edildi ama hiçbir markup üretmedi; bu yüzden yerine yerleşik sayfa sunuldu. |
| `ErrPanic` | `collage: panic recovered while serving the request` | `panic` | Sunum sırasında bir şey panic'ledi — bir `Cache`, `Metrics` ya da `Tracer` implementasyonu, bir router, bir plugin hook'u — ve bir 500'e dönüştürülerek kurtarıldı. Data handler'lardaki ve şablonlardaki panic'ler ise `PanicError`'dır. |
| `ErrAssetFailed` | `collage: asset request failed` | `asset` | Mount edilmiş bir dosya isteği 400 ya da üzeri bir durum koduyla yanıtlandı: bu tür her durum için tek bir sentinel. |
| `ErrHandlerFailed` | `collage: mounted handler failed` | `handler` | `App.Handle` ile mount edilmiş bir handler bir sunucu hatasıyla yanıt verdi. |
| `ErrUnsafeRedirectTarget` | `collage: unsafe redirect target` | `route` | Bir yönlendirmenin hedefi, yer değiştirmeden sonra tek eğik çizgiyle başlayan göreli bir yol değil — `//host`, `/\host` ya da kontrol karakteri içeren bir yol. Bir `Location` header'ıyla değil, bir 500 ile yanıtlanır. |

Alarm kurmaya değer aşama `"error_page"`'dir: hataları bildiren sayfanın kendisi
başarısız oldu ve okuyucu yine de makul görünen bir sayfa gördü; yani bunu size
başka hiçbir şey söylemezdi.

## Statik build'ler

`collage.NewBuilder` ve `Builder.Build` tarafından döner ya da `BuildReport`'a
kaydedilir. Bkz. [Statik dışa aktarma](/docs/static-export).

| Hata | Mesaj | Nerede | Anlamı |
| --- | --- | --- | --- |
| `ErrNilRenderer` | `collage: nil renderer` | `NewBuilder` | App `nil`. |
| `ErrInvalidOutDir` | `collage: invalid output directory` | `NewBuilder` | `BuildOptions.OutDir` boş. |
| `ErrDangerousOutDir` | `collage: refusing to use a dangerous output directory` | `Build` | `OutDir` bir dosya sistemi köküne — ya da `Clean` ile birlikte bir depo köküne — çözümleniyor. |
| `ErrOutputPathCollision` | `collage: two builds target one output path` | `Build` | İki sayfa tek bir dosyaya yazılacaktı — yalnızca sondaki eğik çizgiyle ayrılan pattern'ler ya da bir yolu iki kez döndüren bir path provider. Herhangi bir sayfa render edilmeden önce bildirilir ve ardından hiçbir sayfa render edilmez: document'lar, `404.html` ve asset'ler yine de yazılır. |
| `ErrPathEscapesOutDir` | `collage: resolved path escapes the output directory` | rapor hatası | Bir çıktı yolu ya da ona giden yoldaki bir sembolik bağlantı `OutDir`'in dışına çıkıyor. O tek dosya yazılmadan önce denetlenir; yalnızca o yol başarısız olur. |
| `ErrDynamicPathUnresolved` | `collage: dynamic path pattern requires a path provider` | atlama | Bir sayfanın ya da document'ın yolunda bir `{param}` var ve path provider yok. |
| `ErrNotStatic` | `collage: a Dynamic() route cannot be built statically` | atlama | Bir sayfa ya da document `Dynamic()`; dolayısıyla dışa aktarılacak bir şey yok. |
| `ErrDuplicateOutputPath` | `collage: two build tasks write the same output path` | atlama | İki document görevi tek bir dosyaya çözümleniyor — bir yolu iki kez döndüren bir `DocumentPathProvider`. İlki derlenir, diğerleri atlanır. |
| `ErrDegradedRender` | `collage: refusing to write a degraded render` | rapor hatası | Bir sayfa başarısız bir fragment'le render edildi ve `AllowDegraded` kapalı. Hiçbir dosya yazılmaz. |
| `ErrEmptyRender` | `collage: page rendered no markup` | rapor hatası | Bir sayfa hiç markup render etmedi. `AllowDegraded` açıkken bile reddedilir. Yukarıdaki sunum sentinel'iyle aynıdır. |
| `ErrUnresolvedToken` | `collage: refusing to write a page whose forgery token was never resolved` | rapor hatası, atlama | Bulunamadı sayfası bir `{{csrfToken}}` taşıyor: bir rapor hatası. Token taşıyan diğer her sayfa ise atlanır: bir sunucuya ihtiyacı vardır. |
| `ErrBuildPanic` | `collage: panic while building a page` | rapor hatası | Bir sayfanın render edilmesi ya da yazılması panic'ledi; build kurtarıldı ve diğerleriyle devam etti. |
| `ErrEmptyDocumentBody` | bkz. [Document'lar](#documents) | rapor hatası | Bir document boş bir gövde üretti. |

Rapor hataları, `errors.Is` ile eşleşen hatalardan oluşan bir dilim olan
`BuildReport.Errors`'tadır. Bir atlama, `BuildReport.Skipped` içindeki bir
`SkipRecord`'dur: `Reason`'ı insanlar için bir cümle, `Err`'i ise (v0.10.0'dan
itibaren) kod için sentinel'dir — `ErrNotStatic`, `ErrDynamicPathUnresolved`,
`ErrUnresolvedToken` ya da `ErrDuplicateOutputPath` — bu yüzden onu `Reason`'ı okuyarak
değil, `errors.Is(skip.Err, …)` ile eşleştirin.
