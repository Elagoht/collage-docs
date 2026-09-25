---
description: collage.Config ve alt struct'larının her alanı, varsayılan değerleri ve validation'ın neleri kontrol ettiği.
reference: Config, CacheConfig, ServerConfig, SecurityConfig, TemplateConfig, LocaleConfig, ObservabilityConfig, Config.Validate
---

# Config

Bir uygulamanın config'i tek bir `collage.Config` değeridir ve bu değer
`collage.New`'a verilir. Her alanın kullanılabilir bir sıfır değeri vardır. Bu yüzden
bir config'e yalnızca varsayılanlardan farklı olanları yazarsınız:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
})
```

`New` bu değerle sırasıyla üç şey yapar:

1. **Varsayılanları doldurur.** `cfg.ApplyDefaults()`'u çağırır. Bu metot, varsayılanı
   olan ve sıfır değerde duran her alanı ayarlar. Verdiğiniz değeri yerinde
   değiştirdiği için uygulamanın tam olarak hangi değerlerle kurulduğunu sonradan
   okuyabilirsiniz.
2. **Validation yapar.** `cfg.Validate()`'i çağırır ve bulduğu ilk sorunu hata olarak
   döner. Ayrıntılar için [Validation](#validation) bölümüne bakın.
3. **Uygulamayı kurar.** Var olmayan bir template root'u, parse edilemeyen bir
   template ya da bir plugin'in `Configure` metodunun başarısız olması bu adımda
   bildirilir. Bu hataların hiçbiri ilk request'e kadar beklemez.

`nil` verirseniz `ErrNilConfig` alırsınız. nil bir config varsayılanları istemek
anlamına gelmez, bir hatadır. Çünkü template'lerinizin nerede olduğunu config söyler.

## Config

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `DevMode` | `bool` | `false` | Framework genelinde development modu. |
| `DevWatch` | `[]string` | yok | Değiştiğinde development'taki page'i yeniden yükleyen ek dizinler. v0.10.0'dan beri. |
| `Logger` | `*slog.Logger` | başlangıçta seçilir | Framework'ün ve plugin'lerin log yazdığı logger. |
| `Server` | `ServerConfig` | | HTTP sunucusu. |
| `Security` | `SecurityConfig` | | Request forgery koruması. |
| `Template` | `TemplateConfig` | | Template'lerin yüklenmesi ve render edilmesi. |
| `Cache` | `CacheConfig` | | Page cache. |
| `Locale` | `LocaleConfig` | | URL'lerin hangi locale'leri taşıdığı. |
| `TrailingSlash` | `bool` | `false` | Her page'in URL'si `/` ile biter. v0.13.0'dan beri. |
| `Observability` | `ObservabilityConfig` | | Metric'ler ve tracing. |
| `Plugins` | `[]Plugin` | yok | Uygulama kurulurken register edilen plugin'ler. |
| `PluginConfig` | `map[string]json.RawMessage` | yok | Her plugin'in kendi config'i, plugin adına göre key'lenmiş. |

### DevMode

Development modunu açar. Bu modda template'ler her render'dan önce diskten yeniden
yüklenir ve page'ler tarayıcıda kendilerini yeniler. Başarısız fragment'ler page'in
üzerinde gösterilir, page cache'ten hiç okunmaz ve disk cache'in yerini memory alır.
`DevWatch`'ta adı geçen dizinler izlenir. Built-in error page de hatanın hangi
fragment'te başladığını söyler, hatanın sebebiyle başlar (template kaynaklı bir
hatada başarısız olan çağrının dosyası, satırı ve sütunu; v0.15.0'dan beri) ve hata
zincirinin tamamını gösterir. **Production'da kapalı olmalıdır.** Bu tanı bilgileri
dosya yollarını, host adlarını ve hata mesajlarında ne varsa hepsini içerir.

Geçerli değeri `cfg.IsDevMode()` verir. Bu değer `DevMode || Template.DevMode`'dur.
Scaffold edilmiş bir proje bu alanı `COLLAGE_DEV=1`'e göre ayarlar. `collage dev` bu
değişkeni sizin yerinize set eder.

### DevWatch

Template'ler ve mount'lar dışında, değiştiğinde development'taki page'i yeniden
yükleyen dizinlerdir. Uygulamanızın diskten kendisinin okuduğu Markdown ya da JSON
gibi içerikler için kullanılır. v0.10.0'dan beri vardır.

```go
app, err := collage.New(&collage.Config{
	DevMode:  os.Getenv("COLLAGE_DEV") == "1",
	DevWatch: []string{"content"},
	// ...
})
```

Var olmayan bir dizin hata sayılmaz, atlanır. Çünkü aynı config, binary nerede
başlatılırsa orada çalışır. Development dışında bu alan dikkate alınmaz. Bir dizini
izlemek yalnızca tarayıcıyı yeniler. Yeni içeriği başlangıçta yüklenmiş bir kopyadan
değil de bir sonraki render'da okumak sizin kodunuzun işidir.

### Logger

Varsayılan değer olan `nil`, seçimi framework'e bırakır ve framework uygulamayı
kurarken seçer. Terminalde, ve yalnızca slog'un varsayılan handler'ını hiçbir şey
değiştirmemişse, insanların okuması için tasarlanmış kompakt bir handler kullanılır.
Bu handler her kaydı tek satırda ve renkli bir level işaretiyle yazar. Diğer bütün
ortamlarda `slog.Default()` olduğu gibi kullanılır. `slog.SetDefault`'u çağırmış bir
uygulama kendi handler'ını korur. Sonuçtan emin olmak istiyorsanız bir logger verin,
örneğin makinelerin okuyacağı loglar için bir JSON handler. `ApplyDefaults` bu alanı
`nil` bırakır.

### Plugins ve PluginConfig

`Plugins` içindeki plugin'ler, `New` çalışırken sırayla register edilir. Template'ler
parse edilmeden önce devreye girmesi gereken bir plugin, örneğin template fonksiyonu
ekleyen ya da mount'ları saran bir plugin, `app.RegisterPlugin` ile değil bu alanla
verilmelidir.

`PluginConfig` her plugin'in bölümünü ham JSON olarak tutar. Framework hiçbir dosya
okumaz. Bu alanı istediğiniz gibi doldurabilir ya da
`collage.LoadPluginConfig("plugins-config.json")`'u kullanabilirsiniz. Bu fonksiyon
dosya yoksa `nil` döner. Register edilmiş hiçbir plugin'in adına karşılık gelmeyen bir
key, uygulamanın başlamasını `ErrUnknownPluginConfig` ile engeller. Ayrıntılar için
[Plugin kullanmak](/docs/plugins) sayfasına bakın.

### TrailingSlash

Her page'in tek bir adresi vardır. `TrailingSlash` bu adresin `/` ile bitip
bitmeyeceğini belirler. Açıksa page `/blog/hello/` adresindedir ve `/blog/hello`
oraya `301` ile redirect edilir. Varsayılan olan kapalı durumda bunun tersi geçerlidir.
İsimle üretilen link'ler (`pageURL`, `pageURLIn`, `localeURL`, `app.URL`) seçilen
yazımla çıkar. Bir locale'in ana sayfası da buna dahildir: açıkken `/tr/`, kapalıyken
`/tr`.

Static bir host'a [export](/docs/static-export#hosting) ettiğiniz bir sitede bu
ayarı açın. Export bir page'i `<path>/index.html` olarak yazar. Host bu dosyayı
`/blog/hello/` adresinden sunar, `/blog/hello` adresinden ise ancak bir redirect ile
ulaşılır. Bu yüzden ayar kapalıyken her canonical link, her sitemap girdisi ve her
iç link bir redirect'e işaret eder.

Bu ayar yalnızca page'lere uygulanır. Bir [document](/docs/documents) bir dosyadır ve
path'ini yazıldığı gibi korur, yani iki durumda da `/sitemap.xml` olarak kalır. Bir
action ise hangi yazıma post edildiyse o adreste cevap verir.

## ServerConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Host` | `string` | `"localhost"` | Sunucunun dinlediği adres. |
| `Port` | `int` | `3000` | TCP port'u, `1`–`65535`. |
| `ReadTimeout` | `time.Duration` | `15s` | Bir request'i okumanın en fazla ne kadar sürebileceği. |
| `WriteTimeout` | `time.Duration` | `30s` | Bir response'u yazmanın en fazla ne kadar sürebileceği. |
| `IdleTimeout` | `time.Duration` | `60s` | Bir keep-alive bağlantısının en fazla ne kadar boşta bekleyebileceği. |
| `ShutdownTimeout` | `time.Duration` | `10s` | Graceful shutdown'ın devam eden request'leri ne kadar beklediği. |
| `MaxBodyBytes` | `int64` | 4 MiB | Action kendi sınırını koymadığında action'ın request body'sine uygulanan sınır. |

`Host`'un varsayılanı `localhost`'tur ve bu adrese makinenin dışından erişilemez.
Container içinde bu alanı `0.0.0.0` yapın. Scaffold edilen bir proje `Host` ve
`Port`'u `HOST` ve `PORT` ortam değişkenlerinden doldurur. `collage dev` altında bu
değişkenler, `collage dev`'in seçtiği bir loopback adresidir; sizin adresinizde ise
`collage dev`'in kendisi dinler. Bu yüzden onları ortamdan okumaya devam edin. Bkz.
[CLI](/docs/cli#collage-dev). `MaxBodyBytes`'ı `ApplyDefaults` doldurmaz. Sıfır,
request geldiğinde uygulanan yerleşik 4 MiB (`4 << 20` byte) demektir. Negatif bir
değer ise sınır olmadığı anlamına gelir. Sınırsız bir body, boyutunu anonim bir
çağıranın belirlediği bir bellek kullanımıdır. Bu yüzden bunu bilerek seçin. Bir
action kendi sınırını `WithMaxBodyBytes` ile koyabilir. Ayrıntılar için [Form'lar ve
action'lar](/docs/forms-and-actions) sayfasına bakın.

## SecurityConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `CSRFKey` | `[]byte` | her process için üretilir | Forgery token'larını imzalayan key. |
| `CSRFCookieName` | `string` | `"collage_csrf"` | Token'ı taşıyan cookie. |
| `CSRFFieldName` | `string` | `"_csrf"` | Token'ın gönderildiği form alanı. |
| `CSRFHeaderName` | `string` | `"X-CSRF-Token"` | Token'ın alternatif olarak gönderilebileceği header. |
| `DisableCSRF` | `bool` | `false` | Forgery kontrolünü bütün uygulama için kapatır. |

**Form içeren herhangi bir şeyi deploy etmeden önce `CSRFKey`'i ayarlayın.** Key en
az 32 byte rastgele veri olmalı, diğer secret'larınızla birlikte saklanmalı ve her
instance'ta aynı olmalıdır. Boş bırakılırsa her process için yeni bir key üretilir.
Bu ilk çalıştırma için yeterlidir ama deploy için yanlıştır. Çünkü restart'tan önce ya
da başka bir instance tarafından üretilmiş bir token reddedilir. Uygulama key
ürettiğini log'lar. Development dışında bu bir uyarı, development'ta ise info
seviyesinde bir kayıttır. Bu log yalnızca unsafe bir method'a (`POST`, `PUT`, `PATCH`,
`DELETE`) cevap veren ve `WithoutCSRF` ile muaf tutulmamış bir action varsa yazılır.
Çünkü token'ı yalnızca böyle bir action doğrular (`WithoutCSRF` muafiyeti v0.11.0'dan
beri vardır). Key disk cache'in namespace'ine dahil değildir, bu yüzden cache yeni bir
key'e rağmen korunur. İçinde form olan ve eski key ile saklanmış bir cache'lenmiş page ise
sunulmaz, yeniden render edilir. Ayrıntılar için
[Caching](/docs/caching#the-namespace) sayfasına bakın. Scaffold edilmiş `main.go`
key'i `COLLAGE_CSRF_KEY`'den okur. `openssl rand -hex 32` ile bir key
üretebilirsiniz.

Yukarıdaki isim varsayılanlarını `ApplyDefaults` değil forgery guard uygular. Bu
yüzden bu alanlar `Config`'inizde boş kalır. `CSRFFieldName` alanın adını iki tarafta
birden değiştirir. `{{csrfToken}}` kontrolün okuduğu adı yazdığı için form'larınızda
değişiklik gerekmez. Ancak token'ı form'dan okuyan script'lerin
(`input[name="_csrf"]`) yeni adı kullanması gerekir.

`DisableCSRF`, tarayıcıdan gönderilen hiçbir form'u olmayan uygulamalar içindir,
örneğin kendi authentication'ının arkasında duran bir API. Bu ayar açıkken
`{{csrfToken}}`, gönderilmesi anlamsız olacak bir form'u render etmek yerine render'ı
başarısız kılar. Yalnızca tek bir action'ı muaf tutmak istiyorsanız o action'da
`WithoutCSRF` kullanın.

## TemplateConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `FS` | `fs.FS` | `nil` | Template'lerin yüklendiği dosya sistemi. `nil` disk demektir. |
| `Root` | `string` | diskte `"./templates"` | Template'lerin yüklendiği dizin. Her template adının başından çıkarılır. |
| `Extension` | `string` | `".html"` | Template dosyalarının uzantısı. |
| `Funcs` | `template.FuncMap` | yok | Built-in fonksiyonların üzerine eklenen ek template fonksiyonları. |
| `DevMode` | `bool` | `false` | Template'leri her render'da diskten yeniden yükler. |
| `Timeout` | `time.Duration` | `5s` | Data handler'ların varsayılan timeout'u ve her document handler'ın toplam süre bütçesi. |

### FS ve Root

`FS` `nil` ise `Root` diskte, çalışma dizinine göre bir path'tir ve varsayılanı
`./templates`'tir. `FS` ayarlıysa `Root` onun içinde slash ile ayrılmış bir dizindir
ve **varsayılan değer almaz**. Boş bir `Root`, `FS`'in kendi kökü demektir.

Bir binary'nin herhangi bir dizinden çalışabilmesini sağlayan şey embed etmektir:

```go
//go:embed all:templates
var templates embed.FS

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{FS: templates, Root: "templates"},
})
```

Development'ta embed edilmiş bir template seti değişemez. Bu yüzden collage, diskteki
dizin varsa onu tercih eder. Var olmayan bir `Root` için `New`
`ErrTemplateRootMissing` döner. Altındaki bir template bir symlink üzerinden dışarıya
çözümleniyorsa `ErrTemplateEscapesRoot` döner.

### Funcs

Bu fonksiyonlar template'ler parse edilirken built-in fonksiyonlara eklenir. Built-in
bir fonksiyonla aynı adı taşıyan bir kayıt o built-in fonksiyonun yerini alır. Bir
plugin'in eklediği bir adı taşıyan kayıt da plugin'in fonksiyonunun yerini alır. Bu
alan `New`'dan önce ayarlanmalıdır. Çünkü bir template yalnızca parse edildiği anda
var olan bir fonksiyonu çağırabilir. Bilinmeyen bir adı çağıran template `New`'da
hata verir. Her render'a bağlanan fonksiyonlar (`slot`, `hoist`, `asset`,
`stylesheet`, `csrfToken`, `pageURL`, `pageURLIn`, `localeURL`) her render'da yeniden
bağlanır. Bu yüzden onları override etmenin hiçbir etkisi olmaz. Ayrıntılar için
[Template fonksiyonları](/docs/template-functions) sayfasına bakın.

### Timeout

Fragment'i `WithTimeout` ile kendi süresini koymamış bir data handler'ın deadline'ıdır.
Bir [document](/docs/documents) handler'ın kendi timeout'u yoktur ve bu değer onun
tek sınırıdır. Bu yüzden yavaş tek bir fragment için değeri yükseltirseniz her sitemap
ve feed için de yükseltmiş olursunuz. collage'daki bütün timeout'lar gibi bu da
handler'a verilen context'i sınırlar. `ctx.Done()`'ı hiç kontrol etmeyen bir handler
bu süreyi aşabilir.

## CacheConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Enabled` | `bool` | `false` | Page cache'in ana şalteri. |
| `Store` | `Cache` | `nil` | Kendi cache implementasyonunuz. |
| `Type` | `string` | açıksa `"memory"` | `Store` `nil` olduğunda kullanılan built-in cache: `"memory"` ya da `"disk"`. |
| `DefaultTTL` | `time.Duration` | `5m` | Page bir süre koymadığında entry'nin ömrü. |
| `MaxEntries` | `int` | `10000` | Cache entry sayısının üst sınırı. Negatif değer sınırsız demektir. |
| `Dir` | `string` | yok | Disk cache'in entry'leri sakladığı yer. `"disk"` için zorunludur. |
| `Version` | `string` | binary'den türetilir | Disk cache'teki entry'lerin hangi build'e ait olduğunu belirler. |
| `MaxKeysPerTag` | `int` | `10000` | Tek bir tag altında tutulan cache key sayısının üst sınırı. Negatif değer sınırsız demektir. |

**Caching varsayılan olarak kapalıdır.** `Enabled` false iken bir `Store` ayarlı olsa
bile hiçbir şey cache'lenmez. `Type`'ın varsayılanı yalnızca `Enabled` true ve `Store`
`nil` olduğunda `"memory"` olur.

**Bir `Store`, validation dahil `Type`'ın yerini tamamen alır.** Page'leri Redis'e ya
da başka bir yere koymak için `collage.Cache`'i (`Get`, `Set`, `Invalidate`,
`InvalidateKey`, `Clear`) implement edin. `collage.TaggedCache`'i de implement eden bir
store'a her entry yazılırken o entry'nin tag'leri de verilir. Somut bir tipin nil
pointer'ını atamak yerine alanı boş bırakın. Böyle bir pointer nil olmayan bir
interface'tir ve ilk lookup'ta panic'e yol açar.

**Disk cache process'ten uzun yaşar.** Entry'leri `Dir` altında, adı `Version`'ın
hash'i olan bir alt dizinde durur. Böylece yeni bir build farklı bir dizini okur ve
stale hiçbir şey bulmaz. `Version`'ı boş bırakırsanız değeri çalışan executable'ın
hash'i olur. Bu hash tam da çıktının değişebileceği durumlarda değişir. Page'lerin
nasıl görüneceğini binary dışındaki bir şey belirliyorsa `Version`'ı kendiniz
ayarlayın, örneğin bir commit ya da release tag'i ile. Executable'ın hash'i
alınamazsa bir uyarıyla birlikte memory cache kullanılır. v0.11.0'dan beri dizin
oluşturulamadığında da (örneğin salt okunur bir dosya sisteminde) aynısı olur:
`collage.New` başarısız olmak yerine bir uyarı verir ve memory ile devam eder.
Uygulama çalışırken başarısız olan bir yazma log'lanır ve page cache'lenmeden sunulur.
Development'ta disk cache hiç kullanılmaz, onun yerine bir memory cache devreye
girer.

**`MaxEntries`**, `collage.Cached`'in request'ler arasında tuttuğu değerleri de
sınırlar. **`MaxKeysPerTag`** ise framework'ün dependency tracker'ını sınırlar. Bu
tracker, her process'te bir tag'i cache key'lerine eşleyen bir index'tir. Her farklı
query string ayrı bir key'dir. Cache bir entry'yi evict ettiğinde ya da entry'nin
süresi dolduğunda key tracker'dan silinmez. Bu yüzden bir üst sınır olmasa bir client
bu index'i sınırsızca büyütebilirdi. Bir tag sınıra ulaştığında o tag'in en eski key'i
yalnızca tracker'dan düşürülür, cache'ten silinmez.

Bunun bedeli store'a göre değişir. Built-in memory ve disk cache'ler tag'leri
kendileri index'ler (`TaggedCache`'i implement ederler). Bu yüzden `InvalidateTags`
tuttukları her entry'ye yine ulaşır. Yalnızca `InvalidateTagsN`'in raporladığı sayı,
yani tracker'ın çözümlediği sayı, daha düşük çıkabilir. `TaggedCache`'i implement
etmeyen custom bir `Store` ise yalnızca tracker'a güvenir. Onun için düşürülen bir
key, `InvalidateTags`'in artık ulaşamadığı bir entry demektir ve bu entry süresi
dolana kadar sunulmaya devam eder. Böyle bir store kullanıyorsanız sınırı, tek bir
tag'in kapsayabileceği canlı entry sayısının üzerinde bir değere ayarlayın.

İki sınırda da sıfır varsayılan değer demektir. Sınırsız anlamına gelen tek değer
negatif bir değerdir. Ayrıntılar için [Caching](/docs/caching) sayfasına bakın.

## LocaleConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Default` | `string` | `"en"` | Locale prefix'i olmayan bir URL'nin locale'i. |
| `Supported` | `[]string` | `[Default]` | Uygulamanın sunduğu bütün locale'ler. |
| `DisablePathLocale` | `bool` | `false` | Locale path'ten çözümlenmez. Her request `Default` locale'indedir. |
| `PrefixDefault` | `bool` | `false` | `Default`'un page'leri de prefix alır: `/en/about`. v0.14.0'dan beri. |

Locale'i seçen tek şey URL'dir. `/about` `Default` locale'indedir, `/tr/hakkinda` ise
`"tr"` locale'indedir. collage locale'i hiçbir zaman `Accept-Language`'ten ya da bir
cookie'den seçmez. Çünkü farklı okuyuculara farklı şey ifade eden bir URL'yi cache'ler,
crawler'lar ve paylaşılan link'ler hep yanlış ele alır. Varsayılan locale'in kendi
prefix'i olan `/en/about`, `/about`'a kalıcı olarak redirect edilir. `PrefixDefault`
ayarlıysa durum tersine döner: adres `/en/about` olur ve `/about` oraya redirect
edilir. Ayrıntılar için
[Link'ler ve locale'ler](/docs/links-and-locales#the-url-decides-the-locale) sayfasına
bakın. Dili kendiniz belirlemek istiyorsanız bunu middleware'de yapın. Türkçe bir
tarayıcıyı `/tr`'ye redirect edebilir ya da her dil için ayrı bir URL render edip bunu
`collage.Vary` ile bildirebilirsiniz. Ayrıntılar için
[Link'ler ve locale'ler](/docs/links-and-locales#negotiating-a-language-yourself)
sayfasına bakın.

## ObservabilityConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Metrics` | `Metrics` | no-op | Counter'ları ve süre ölçümlerini alır. |
| `Tracer` | `Tracer` | no-op | Request'lerin, render'ların ve fragment'lerin etrafında span başlatır. |

İkisi de collage'ı kendi backend'inize bağlamak için implement ettiğiniz
interface'lerdir. `nil` no-op demektir.

```go
type Metrics interface {
	RenderDuration(ctx context.Context, page string, d time.Duration, cacheHit bool)
	FragmentDuration(ctx context.Context, page, fragment string, d time.Duration, err error)
	CacheEvent(ctx context.Context, event collage.CacheEvent, key string)
	HTTPResponse(ctx context.Context, status int, path string, d time.Duration)
	Invalidation(ctx context.Context, tags []string, keys int)
}

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, collage.Span)
}

type Span interface {
	SetAttribute(key, value string)
	RecordError(err error)
	End()
}
```

`CacheEvent` şu değerlerden biridir: `collage.CacheHit`, `CacheMiss`, `CacheSet`,
`CacheEvict`, `CacheInvalidate` ve `CacheCoalesced`. Sonuncusu, bir request'in aynı
key için zaten çalışmakta olan bir render tarafından karşılandığı anlamına gelir.

## Validation

`Validate`, `ApplyDefaults` çalıştıktan sonra aşağıdaki durumlardan bulduğu ilkini
döner:

| Koşul | Hata |
| --- | --- |
| `Server.Port` `1`–`65535` aralığının dışında | `ErrInvalidPort` |
| `Template.Root` boş ve `Template.FS` nil | `ErrEmptyTemplateRoot` |
| `Cache.Enabled` açık, `Store` yok ve `Type` ne `"memory"` ne `"disk"` | `ErrInvalidCacheType` |
| `Cache.Enabled` açık, `Store` yok, `Type` `"disk"` ve `Cache.Dir` boş | `ErrEmptyCacheDir` |
| `Locale.Default` boş | `ErrEmptyLocaleDefault` |
| `Locale.Default`, `Locale.Supported` içinde yok | `ErrLocaleDefaultNotSupported` |
| `Server.ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`, `Template.Timeout` ya da `Cache.DefaultTTL` negatif | `ErrNegativeDuration` |

Negatif süre hatası altı alanın hepsi için tek bir sentinel error'dır. Mesaj, hataya
yol açan alanın adını verir, örneğin `server.read_timeout`. Varsayılanlar önce
uygulandığı için `New` üzerinden gelen sıfır bir port ya da boş bir varsayılan locale
validation'a hiç ulaşmaz. Yalnızca açıkça set ettiğiniz bir değer hataya yol
açabilir. Bu hataları `errors.Is` ile karşılaştırın. Tam liste
[Hatalar](/docs/errors#configuration) sayfasındadır.

İki metodu da kendiniz çağırabilirsiniz. Örneğin bir testte bir config'i kontrol etmek
için:

```go
cfg := collage.Config{Server: collage.ServerConfig{Port: 70000}}
cfg.ApplyDefaults()
err := cfg.Validate() // wraps collage.ErrInvalidPort
```

`ApplyDefaults` idempotent'tir. Varsayılanları zaten uygulanmış bir config'e tekrar
uygulandığında hiçbir şeyi değiştirmez.

## Eksiksiz bir config

Aşağıdaki, scaffold edilmiş bir projenin başladığı config'dir. Ortama göre değişen
kısımlar ortam değişkenlerinden okunur:

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	return nil, fmt.Errorf("plugin configuration: %w", err)
}

app, err := collage.New(&collage.Config{
	DevMode: os.Getenv("COLLAGE_DEV") == "1",
	Server: collage.ServerConfig{
		Host: envString("HOST", "localhost"),
		Port: envInt("PORT", 3000),
	},
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "disk",
		Dir:        ".cache",
		DefaultTTL: 5 * time.Minute,
	},
	PluginConfig: pluginConfig,
	Security: collage.SecurityConfig{
		CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	},
})
```

`envString` ve `envInt`, scaffold edilmiş `main.go` içindeki iki küçük yardımcı
fonksiyondur. Bir environment değişkenini okurlar, değişken yoksa varsayılan bir
değere dönerler.
