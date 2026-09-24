---
description: collage.Config ve alt yapılarının her alanı, varsayılan değerleri ve doğrulamanın neyi denetlediği.
---

# Yapılandırma

Bir uygulama, `collage.New`'a verilen tek bir `collage.Config` değeriyle
yapılandırılır. Her alanın kullanılabilir bir sıfır değeri vardır; bu yüzden bir
yapılandırma yalnızca varsayılanlardan farklı olanı söyler:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
})
```

`New` bu değerle sırasıyla üç şey yapar:

1. **Varsayılanları doldurur.** `cfg.ApplyDefaults()`'u çağırır; bu da varsayılanı
   olan ve sıfır değerde duran her alanı ayarlar. Verdiğiniz değeri değiştirir; böylece
   uygulamanın tam olarak neyle kurulduğunu geri okuyabilirsiniz.
2. **Doğrular.** `cfg.Validate()`'i çağırır ve bulduğu ilk sorunu hata olarak döner —
   bkz. [Doğrulama](#validation).
3. **Uygulamayı kurar.** Var olmayan bir şablon kökü, ayrıştırılamayan bir şablon, bir
   plugin'in `Configure`'unun başarısız olması — her biri ilk istekte değil, burada
   bildirilir.

`nil` vermek `ErrNilConfig`'dir. nil bir yapılandırma varsayılanları istemek değil,
bir hatadır; çünkü şablonlarınızın nerede olduğunu yapılandırma söyler.

## Config

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `DevMode` | `bool` | `false` | Framework genelinde geliştirme modu. |
| `DevWatch` | `[]string` | yok | Değiştiğinde geliştirme sayfasını yenileyen ek dizinler. v0.10.0'dan itibaren. |
| `Logger` | `*slog.Logger` | başlangıçta seçilir | Framework'ün ve plugin'lerin yazdığı logger. |
| `Server` | `ServerConfig` | | HTTP sunucusu. |
| `Security` | `SecurityConfig` | | İstek sahteciliği koruması. |
| `Template` | `TemplateConfig` | | Şablon yükleme ve render etme. |
| `Cache` | `CacheConfig` | | Sayfa önbelleği. |
| `Locale` | `LocaleConfig` | | URL'lerin hangi locale'leri taşıdığı. |
| `TrailingSlash` | `bool` | `false` | Her sayfanın URL'sini `/` ile bitirir. v0.13.0'dan itibaren. |
| `Observability` | `ObservabilityConfig` | | Metrikler ve izleme (tracing). |
| `Plugins` | `[]Plugin` | yok | Uygulama kurulurken kaydedilen plugin'ler. |
| `PluginConfig` | `map[string]json.RawMessage` | yok | Plugin adına göre anahtarlanmış, her plugin'in kendi yapılandırması. |

### DevMode

Geliştirme modunu açar: şablonlar her render'dan önce diskten yeniden yüklenir,
sayfalar tarayıcıda kendilerini yeniler, başarısız fragment'ler sayfada gösterilir,
sayfa önbelleğinden hiç okunmaz, disk önbelleğinin yerini bellek alır, `DevWatch`'ta
adı geçen dizinler izlenir ve yerleşik hata sayfası hatanın başladığı fragment'in
adını verip hata zincirinin tamamını gösterir. **Production'da kapalı olmalıdır** —
bu tanılama bilgileri yolları, sunucu adlarını ve hata mesajlarının içerdiği her şeyi
taşır.

Geçerli değer `cfg.IsDevMode()`'dur; bu da `DevMode || Template.DevMode`'dur. İskelet
olarak oluşturulan bir proje bunu `COLLAGE_DEV=1`'den ayarlar; `collage dev` bunu
sizin için ayarlar.

### DevWatch

Şablonlar ve mount'lar dışında, değiştiğinde geliştirme sayfasını yenileyen
dizinler — uygulamanızın diskten kendisi okuduğu Markdown ya da JSON gibi içerikler
için. v0.10.0'dan itibaren.

```go
app, err := collage.New(&collage.Config{
	DevMode:  os.Getenv("COLLAGE_DEV") == "1",
	DevWatch: []string{"content"},
	// ...
})
```

Var olmayan bir dizin hata sayılmaz, atlanır; çünkü aynı yapılandırma binary nerede
başlatılırsa orada çalışır. Geliştirme dışında yok sayılır. Bir dizini izlemek yalnızca
tarayıcıyı yeniler; yeni içeriği bir sonraki render'da — başlangıçta yüklenmiş bir
kopyadan değil — okumak sizin kodunuzun işidir.

### Logger

Varsayılan olan `nil`, seçimi uygulamayı kurarken framework'e bırakır: bir
terminalde — ve yalnızca slog'un varsayılan handler'ını hiçbir şey değiştirmemişse —
bir insan için tasarlanmış, her kayıt için tek satır ve renkli bir seviye işaretiyle
yazan derli toplu bir handler. Diğer her yerde, değiştirilmeden `slog.Default()`.
`slog.SetDefault`'u çağırmış bir uygulama kendi handler'ını korur. Emin olmak için
bir logger verin; örneğin bir makinenin okuyacağı loglar için bir JSON handler'ı.
`ApplyDefaults` bu alanı `nil` bırakır.

### Plugins ve PluginConfig

`Plugins`, `New` çalışırken sırayla kaydedilir. Şablonlar ayrıştırılmadan önce
harekete geçmesi gereken — bir şablon fonksiyonu eklemek ya da mount'ları sarmalamak
için — bir plugin, `app.RegisterPlugin` üzerinden değil, buradan gelmelidir.

`PluginConfig` her plugin'in bölümünü ham JSON olarak tutar. Framework hiçbir dosya
okumaz: bunu dilediğiniz gibi doldurun ya da eksik bir dosya için `nil` dönen
`collage.LoadPluginConfig("plugins-config.json")`'u kullanın. Kayıtlı hiçbir plugin'i
adlandırmayan bir anahtar, uygulamanın `ErrUnknownPluginConfig` ile başlamasını
engeller. Bkz. [Plugin kullanmak](/docs/plugins).

### TrailingSlash

Her sayfanın tek bir adresi vardır; `TrailingSlash` bu adresin `/` ile bitip
bitmediğini söyler. Açıkken sayfa `/blog/hello/` adresindedir ve `/blog/hello` oraya
`301` ile yönlendirilir; varsayılan olan kapalı durumda tersi geçerlidir. Adla
kurulan bağlantılar — `pageURL`, `pageURLIn`, `localeURL`, `app.URL` — seçilen
yazımla çıkar; bir locale'in ana sayfası da buna dahildir: açıkken `/tr/`,
kapalıyken `/tr`.

Statik bir barındırma hizmetine [dışa aktardığınız](/docs/static-export#hosting) bir
site için açın. Dışa aktarma bir sayfayı `<path>/index.html` olarak yazar; barındırma
hizmeti bu dosyayı `/blog/hello/` adresinde sunar, `/blog/hello` adresinden ise
ancak bir yönlendirmeyle ulaşır. Yani ayar kapalıyken her canonical bağlantı, her
sitemap girdisi ve her iç bağlantı bir yönlendirmeyi gösterir.

Ayar sayfalara uygulanır. Bir [document](/docs/documents) bir dosyadır ve yolunu
yazıldığı gibi korur, iki durumda da `/sitemap.xml`; bir action ise hangi yazıma
gönderildiyse orada yanıtlanır.

## ServerConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Host` | `string` | `"localhost"` | Sunucunun dinlediği adres. |
| `Port` | `int` | `3000` | TCP portu, `1`–`65535`. |
| `ReadTimeout` | `time.Duration` | `15s` | Bir isteği okumanın ne kadar sürebileceği. |
| `WriteTimeout` | `time.Duration` | `30s` | Bir yanıtı yazmanın ne kadar sürebileceği. |
| `IdleTimeout` | `time.Duration` | `60s` | Bir keep-alive bağlantısının ne kadar boşta bekleyebileceği. |
| `ShutdownTimeout` | `time.Duration` | `10s` | Düzgün kapanışın süren istekleri ne kadar beklediği. |
| `MaxBodyBytes` | `int64` | 4 MiB | Action kendisi bir sınır koymadığında, action'ın istek gövdesine uygulanan sınır. |

`Host`'un varsayılanı `localhost`'tur ve makinenin dışından erişilemez — bir
container'da bunu `0.0.0.0` yapın. `MaxBodyBytes`'ı `ApplyDefaults` doldurmaz: sıfır,
bir istek geldiğinde uygulanan yerleşik 4 MiB (`4 << 20` bayt) anlamına gelir; negatif
bir değer ise sınırsız demektir. Sınırsız bir gövde, boyutunu anonim bir çağıranın
seçtiği bellektir; bunu bilinçli olarak seçin. Bir action kendi sınırını
`WithMaxBodyBytes` ile koyabilir; bkz. [Formlar ve action'lar](/docs/forms-and-actions).

## SecurityConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `CSRFKey` | `[]byte` | her süreç için üretilir | Sahtecilik token'larını imzalayan anahtar. |
| `CSRFCookieName` | `string` | `"collage_csrf"` | Token'ın taşındığı cookie. |
| `CSRFFieldName` | `string` | `"_csrf"` | Token'ın gönderildiği form alanı. |
| `CSRFHeaderName` | `string` | `"X-CSRF-Token"` | Token'ın bunun yerine gönderilebileceği header. |
| `DisableCSRF` | `bool` | `false` | Sahtecilik denetimini tüm uygulama için kapatır. |

**Form içeren herhangi bir şeyi yayına almadan önce `CSRFKey`'i ayarlayın.** En az
32 rastgele bayt olmalı, diğer gizli bilgilerinizle birlikte saklanmalı ve her
instance'ta aynı olmalıdır. Boş bırakılırsa her süreç için bir anahtar üretilir: ilk
çalıştırma için sorun değildir, yayına almak içinse yanlıştır; çünkü yeniden
başlatmadan önce — ya da başka bir instance tarafından — verilmiş bir token reddedilir.
Uygulama bir anahtar ürettiğini loglar — geliştirme dışında uyarı, geliştirmede info
seviyesinde — ve bunu yalnızca güvenli olmayan bir metoda (`POST`, `PUT`, `PATCH`,
`DELETE`) yanıt veren ve `WithoutCSRF` ile muaf tutulmamış bir action varsa yapar;
çünkü token'ı yalnızca böyle bir action doğrular (`WithoutCSRF` muafiyeti v0.11.0'dan
itibaren). Anahtar disk önbelleğinin ad alanına dahil değildir, bu yüzden önbellek yeni
bir anahtardan etkilenmeden kalır; içinde form olan ve eski anahtarla saklanmış bir
önbellekteki sayfa ise sunulmaz, yeniden render edilir — bkz.
[Önbellekleme](/docs/caching#the-namespace). İskelet olarak oluşturulan `main.go` onu
`COLLAGE_CSRF_KEY`'den okur; `openssl rand -hex 32` bir tane üretir.

Yukarıdaki ad varsayılanlarını `ApplyDefaults` değil, sahtecilik koruması uygular; bu
yüzden alanlar `Config`'inizde boş kalır. `CSRFFieldName` alanı iki tarafta da yeniden
adlandırır: `{{csrfToken}}` denetimin okuduğu adı yazar, formlarda değişiklik gerekmez.
Token'ı formdan okuyan script'ler — `input[name="_csrf"]` — yeni adı kullanmak
zorundadır.

`DisableCSRF`, tarayıcıdan gönderilen hiçbir formu olmayan bir uygulama içindir —
kendi kimlik doğrulamasının arkasındaki bir API. Ayarlandığında `{{csrfToken}}`,
gönderimi hiçbir anlam taşımayacak bir form render etmek yerine render'ı başarısız
kılar. Bunun yerine tek bir action'ı muaf tutmak için o action'da `WithoutCSRF`
kullanın.

## TemplateConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `FS` | `fs.FS` | `nil` | Şablonların yüklendiği dosya sistemi; `nil` disk demektir. |
| `Root` | `string` | diskte `"./templates"` | Şablonların yüklendiği, her şablon adından çıkarılan dizin. |
| `Extension` | `string` | `".html"` | Şablon dosyalarının uzantısı. |
| `Funcs` | `template.FuncMap` | yok | Yerleşiklerin üzerine birleştirilen ek şablon fonksiyonları. |
| `DevMode` | `bool` | `false` | Şablonları her render'da diskten yeniden yükler. |
| `Timeout` | `time.Duration` | `5s` | Varsayılan data handler zaman aşımı ve her document handler'ının toplam süre bütçesi. |

### FS ve Root

`FS` `nil` iken `Root` çalışma dizinine göre diskteki bir yoldur ve varsayılanı
`./templates`'tir. `FS` ayarlıyken `Root` onun içinde eğik çizgiyle ayrılmış bir
dizindir ve varsayılanı **yoktur** — boş bir `Root`, `FS`'in kendi kökü demektir.

Bir binary'nin herhangi bir dizinden çalışabilmesini sağlayan şey gömmektir:

```go
//go:embed all:templates
var templates embed.FS

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{FS: templates, Root: "templates"},
})
```

Geliştirmede gömülü bir küme değişemeyeceği için collage, oradaysa diskteki dizini
tercih eder. Var olmayan bir `Root`, `New`'dan `ErrTemplateRootMissing` döner; altında
olup — bir symlink üzerinden — dışına çözümlenen bir şablon ise
`ErrTemplateEscapesRoot`'tur.

### Funcs

Şablonlar ayrıştırılırken yerleşik fonksiyonlara eklenir; yerleşik bir adla girilen
bir kayıt o yerleşiğin yerini alır, bir plugin'in eklediği adla girilen bir kayıt da
plugin'inkinin. `New`'dan önce ayarlanmalıdır; çünkü bir şablon yalnızca
ayrıştırıldığı sırada var olan bir fonksiyonu çağırabilir — bilinmeyen bir adı çağıran
şablon `New`'da başarısız olur. Her render'a bağlanan fonksiyonlar (`slot`, `hoist`,
`asset`, `stylesheet`, `csrfToken`, `pageURL`, `pageURLIn`, `localeURL`) her render'da
yeniden bağlanır; bu yüzden onları geçersiz kılmanın hiçbir etkisi yoktur. Bkz.
[Şablon fonksiyonları](/docs/template-functions).

### Timeout

Fragment'i `WithTimeout` ile bir süre koymamış bir data handler'ın son süresi. Aynı
zamanda kendi zaman aşımı olmayan bir [document](/docs/documents) handler'ının tek
sınırıdır — dolayısıyla yavaş tek bir fragment için onu artırmak, her sitemap ve feed
için de artırır. collage'daki her zaman aşımı gibi, handler'ın aldığı context'i
sınırlar; `ctx.Done()`'ı hiç denetlemeyen bir handler bu süreyi aşabilir.

## CacheConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Enabled` | `bool` | `false` | Sayfa önbelleğinin ana şalteri. |
| `Store` | `Cache` | `nil` | Kendi önbellek gerçeklemeniz. |
| `Type` | `string` | etkinse `"memory"` | `Store` `nil` iken yerleşik önbellek: `"memory"` ya da `"disk"`. |
| `DefaultTTL` | `time.Duration` | `5m` | Sayfa bir süre koymadığında kaydın ömrü. |
| `MaxEntries` | `int` | `10000` | Önbellek kaydı sayısının üst sınırı; negatif sınırsız demektir. |
| `Dir` | `string` | yok | Disk önbelleğinin kayıtları sakladığı yer. `"disk"` için zorunlu. |
| `Version` | `string` | binary'den türetilir | Disk önbelleğindeki kayıtların hangi build'e ait olduğunu belirler. |
| `MaxKeysPerTag` | `int` | `10000` | Tek bir etiket altında tutulan önbellek anahtarı sayısının üst sınırı; negatif sınırsız demektir. |

**Önbellekleme varsayılan olarak kapalıdır.** `Enabled` false iken, bir `Store`
ayarlı olsa bile hiçbir şey önbelleğe alınmaz. `Type`'ın varsayılanı yalnızca `Enabled`
true ve `Store` `nil` olduğunda `"memory"`'dir.

**Bir `Store`, doğrulama da dahil olmak üzere `Type`'ın yerini tamamen alır.**
Sayfaları Redis'e ya da başka bir yere koymak için `collage.Cache`'i (`Get`, `Set`,
`Invalidate`, `InvalidateKey`, `Clear`) gerçekleyin; ayrıca `collage.TaggedCache`'i de
gerçekleyen bir depoya her kaydın etiketleri yazma sırasında verilir. Somut bir tipin
nil pointer'ını atamak yerine alanı boş bırakın; o, ilk aramada panic'e yol açan nil
olmayan bir interface'tir.

**Disk önbelleği süreçten uzun yaşar.** Kayıtları `Dir`'in, `Version`'ın hash'iyle
adlandırılmış bir alt dizininde durur; böylece yeni bir build farklı bir dizini okur
ve bayat hiçbir şey bulmaz. `Version`'ı boş bırakırsanız çalışan executable'ın hash'i
olur; bu da tam olarak çıktının değişebileceği anda değişir. Sayfaların nasıl
görüneceğine binary'nin dışındaki bir şey karar veriyorsa onu ayarlayın — bir commit,
bir sürüm etiketi. Executable'ın hash'i alınamıyorsa bir uyarıyla birlikte bellek içi
bir önbellek kullanılır — v0.11.0'dan itibaren dizin oluşturulamadığında da (örneğin
salt okunur bir dosya sisteminde) böyledir: `collage.New` başarısız olmak yerine uyarır
ve bellekle devam eder. Uygulama çalışırken başarısız olan bir yazma loglanır ve sayfa
önbelleğe alınmadan sunulur. Geliştirmede disk önbelleği hiç kullanılmaz: yerini bellek
içi bir önbellek alır.

**`MaxEntries`** ayrıca `collage.Cached`'in istekler arasında tuttuğu değerleri de
sınırlar. **`MaxKeysPerTag`** ise framework'ün bağımlılık izleyicisini, yani bir
etiketi önbellek anahtarlarına geri eşleyen süreç başına dizini sınırlar. Her farklı
query string farklı bir anahtardır ve önbellek bir kaydı çıkardığında ya da kaydın
süresi dolduğunda hiçbir şey anahtarı izleyiciden silmez; bu yüzden bir üst sınır
olmadan bir istemci bu dizini sınırsızca büyütebilir. Bir etiket sınıra ulaştığında
en eski anahtarı yalnızca izleyiciden düşürülür, önbellekten değil.

Bunun bedeli depoya bağlıdır. Yerleşik bellek ve disk önbellekleri etiketleri kendileri
dizinler (`TaggedCache`'i gerçeklerler); bu yüzden `InvalidateTags` tuttukları her
kayda yine ulaşır; yalnızca `InvalidateTagsN`'in bildirdiği sayı — izleyicinin
çözümlediği sayı — daha düşük çıkabilir. `TaggedCache`'i gerçeklemeyen özel bir `Store`
yalnızca izleyiciye dayanır ve onun için düşürülen bir anahtar, `InvalidateTags`'in
artık ulaşamadığı bir kayıttır — süresi dolana kadar sunulur. Böyle bir depoyla, sınırı
herhangi bir etiketin kapsayabileceği canlı kayıt sayısının üzerine ayarlayın.

İki sınır için de sıfır varsayılan demektir; yalnızca negatif bir değer sınırsız
demektir. Bkz. [Önbellekleme](/docs/caching).

## LocaleConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Default` | `string` | `"en"` | Locale öneki olmayan bir URL'nin locale'i. |
| `Supported` | `[]string` | `[Default]` | Uygulamanın sunduğu her locale. |
| `DisablePathLocale` | `bool` | `false` | Locale'i yoldan çözümlemeyi bırakır; her istek `Default`'tadır. |

Locale'i seçen tek şey URL'dir: `/about` `Default`'tadır, `/tr/hakkinda` ise `"tr"`'de.
collage hiçbir zaman `Accept-Language`'ten ya da bir cookie'den locale seçmez; çünkü
farklı okuyuculara farklı şeyler ifade eden bir URL'yi önbellekler, crawler'lar ve
paylaşılan bağlantılar yanlış anlar. Varsayılan locale'in kendi öneki olan `/en/about`,
`/about`'a kalıcı olarak yönlendirilir. Dil seçimini müzakere etmek istiyorsanız bunu
middleware'de yapın: Türkçe bir tarayıcıyı `/tr`'ye yönlendirin ya da tek bir URL'yi
her dil için render edip bunu `collage.Vary` ile bildirin. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales#negotiating-a-language-yourself).

## ObservabilityConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Metrics` | `Metrics` | işlemsiz | Sayaçları ve süreleri alır. |
| `Tracer` | `Tracer` | işlemsiz | İstekler, render'lar ve fragment'ler etrafında span başlatır. |

İkisi de collage'ı kendi altyapınıza bağlamak için gerçeklediğiniz interface'lerdir;
`nil` hiçbir şey yapmayan (no-op) bir gerçekleme demektir.

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

`CacheEvent`; `collage.CacheHit`, `CacheMiss`, `CacheSet`, `CacheEvict`,
`CacheInvalidate` ve `CacheCoalesced`'ten biridir — sonuncusu, bir isteğin aynı
anahtar için zaten çalışmakta olan bir render tarafından karşılandığı anlamına gelir.

## Doğrulama

`Validate`, `ApplyDefaults` çalıştıktan sonra bunlardan bulduğu ilkini döner:

| Koşul | Hata |
| --- | --- |
| `Server.Port` `1`–`65535` aralığının dışında | `ErrInvalidPort` |
| `Template.Root` boş ve `Template.FS` nil | `ErrEmptyTemplateRoot` |
| `Cache.Enabled`, `Store` yok ve `Type` ne `"memory"` ne `"disk"` | `ErrInvalidCacheType` |
| `Cache.Enabled`, `Store` yok, `Type` `"disk"` ve `Cache.Dir` boş | `ErrEmptyCacheDir` |
| `Locale.Default` boş | `ErrEmptyLocaleDefault` |
| `Locale.Default`, `Locale.Supported` içinde değil | `ErrLocaleDefaultNotSupported` |
| Negatif bir `Server.ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`, `Template.Timeout` ya da `Cache.DefaultTTL` | `ErrNegativeDuration` |

Negatif süre hatası altı alanın hepsi için tek bir sentinel'dir; mesaj başarısız olan
alanı adlandırır, örneğin `server.read_timeout`. Varsayılanlar önce uygulandığı için
sıfır bir port ya da boş bir varsayılan locale `New`'dan doğrulamaya hiç ulaşmaz —
yalnızca açıkça ayarladığınız bir değer başarısız olabilir. Bunları `errors.Is` ile
eşleştirin; tam liste [Hatalar](/docs/errors#configuration) sayfasındadır.

İkisini de kendiniz çağırabilirsiniz — örneğin bir testte bir yapılandırmayı denetlemek
için:

```go
cfg := collage.Config{Server: collage.ServerConfig{Port: 70000}}
cfg.ApplyDefaults()
err := cfg.Validate() // wraps collage.ErrInvalidPort
```

`ApplyDefaults` idempotent'tir: varsayılanları zaten olan bir yapılandırmaya
uygulamak hiçbir şeyi değiştirmez.

## Eksiksiz bir yapılandırma

İskelet olarak oluşturulan bir projenin başladığı yapılandırma; ortama göre değişen
kısımlar ortamdan okunur:

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

`envString` ve `envInt`, iskelet `main.go`'daki, bir değişkeni okuyup yoksa bir
varsayılana dönen iki küçük yardımcıdır.
