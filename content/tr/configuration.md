---
description: collage.Config ve alt struct'larının her alanı, varsayılan değerleri ve validation'ın neleri kontrol ettiği.
reference: Config, CacheConfig, ServerConfig, SecurityConfig, TemplateConfig, LocaleConfig, ObservabilityConfig, CacheType, ParseOrigin
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

1. **Varsayılanları doldurur.** Varsayılanı olan ve sıfır değerde duran her alanı
   ayarlar. Verdiğiniz değeri yerinde değiştirdiği için uygulamanın tam olarak
   hangi değerlerle kurulduğunu sonradan okuyabilirsiniz.
2. **Validation yapar.** Sonucu kontrol eder ve bulduğu ilk sorunu hata olarak
   döner. Ayrıntılar için [Validation](#validation) bölümüne bakın.
3. **Uygulamayı kurar.** Var olmayan bir template root'u, parse edilemeyen bir
   template ya da bir plugin'in `Configure` metodunun başarısız olması bu adımda
   bildirilir. Bu hataların hiçbiri ilk request'e kadar beklemez.

`nil` verirseniz bir hata alırsınız. nil bir config varsayılanları istemek
anlamına gelmez, bir hatadır. Çünkü template'lerinizin nerede olduğunu config
söyler.

Scaffold edilmiş bir proje config'ini `main.go`'daki `newApp`'te kurar. Sizin
makinenizle bir sunucu arasında değişen şeyler ortamdan gelir; geri kalanı orada
yazılıdır:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 456" role="img" aria-labelledby="config-shape-title">
<title id="config-shape-title">Scaffold edilmiş main.go, collage.Config'in her alanını nereden alır. Ortamdan: DevMode COLLAGE_DEV=1'den, BaseURL BASE_URL'den, Server.Host HOST'tan, Server.Port varsayılanı PORT ya da 6060 olan -port flag'inden, Security.CSRFKey COLLAGE_CSRF_KEY'den. main.go'da ve yanındaki dosyalarda: PluginConfig plugins-config.json'dan, Plugins plugins.go'daki plugins()'ten, Template embed edilmiş templates dizininden, Cache ise .cache altında beş dakikalık TTL'li bir disk cache. Atanmayan ve varsayılanında kalanlar: Locale (yalnızca en), Logger, DevWatch, TrailingSlash ve Observability. Değerin tamamı collage.New'a gider; New varsayılanları doldurur, validation yapar ve uygulamayı kurar.</title>
<defs><marker id="config-shape-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<text class="d-small" x="78" y="14" text-anchor="middle">scaffold'un main.go'sunda</text>
<text class="d-small" x="242" y="14" text-anchor="middle">collage.Config alanı</text>
<text class="d-small" x="160" y="34" text-anchor="middle">ortamdan okunur</text>
<rect class="d-pill" x="6" y="40" width="144" height="24" rx="12" /><text class="d-small" x="78" y="56" text-anchor="middle">COLLAGE_DEV=1</text>
<line class="d-flow" x1="150" y1="52" x2="166" y2="52" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="40" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="56" text-anchor="middle">DevMode</text>
<rect class="d-pill" x="6" y="70" width="144" height="24" rx="12" /><text class="d-small" x="78" y="86" text-anchor="middle">BASE_URL</text>
<line class="d-flow" x1="150" y1="82" x2="166" y2="82" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="70" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="86" text-anchor="middle">BaseURL</text>
<rect class="d-pill" x="6" y="100" width="144" height="24" rx="12" /><text class="d-small" x="78" y="116" text-anchor="middle">HOST</text>
<line class="d-flow" x1="150" y1="112" x2="166" y2="112" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="100" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="116" text-anchor="middle">Server.Host</text>
<rect class="d-pill" x="6" y="130" width="144" height="24" rx="12" /><text class="d-small" x="78" y="146" text-anchor="middle">-port, PORT ya da 6060</text>
<line class="d-flow" x1="150" y1="142" x2="166" y2="142" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="130" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="146" text-anchor="middle">Server.Port</text>
<rect class="d-pill" x="6" y="160" width="144" height="24" rx="12" /><text class="d-small" x="78" y="176" text-anchor="middle">COLLAGE_CSRF_KEY</text>
<line class="d-flow" x1="150" y1="172" x2="166" y2="172" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="160" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="176" text-anchor="middle">Security.CSRFKey</text>
<text class="d-small" x="160" y="202" text-anchor="middle">main.go'da ve yanındaki dosyalarda</text>
<rect class="d-folder" x="6" y="208" width="144" height="24" rx="4" /><text class="d-small" x="78" y="224" text-anchor="middle">plugins-config.json</text>
<line class="d-flow" x1="150" y1="220" x2="166" y2="220" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="208" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="224" text-anchor="middle">PluginConfig</text>
<rect class="d-folder" x="6" y="238" width="144" height="24" rx="4" /><text class="d-small" x="78" y="254" text-anchor="middle">plugins.go'da plugins()</text>
<line class="d-flow" x1="150" y1="250" x2="166" y2="250" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="238" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="254" text-anchor="middle">Plugins</text>
<rect class="d-folder" x="6" y="268" width="144" height="24" rx="4" /><text class="d-small" x="78" y="284" text-anchor="middle">embed edilen templates/</text>
<line class="d-flow" x1="150" y1="280" x2="166" y2="280" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="268" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="284" text-anchor="middle">Template</text>
<rect class="d-folder" x="6" y="298" width="144" height="24" rx="4" /><text class="d-small" x="78" y="314" text-anchor="middle">disk, .cache, 5m</text>
<line class="d-flow" x1="150" y1="310" x2="166" y2="310" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="298" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="314" text-anchor="middle">Cache</text>
<text class="d-small" x="160" y="340" text-anchor="middle">atanmaz: her biri varsayılanında kalır</text>
<rect class="d-frame" x="6" y="346" width="308" height="40" rx="6" /><text class="d-small" x="160" y="362" text-anchor="middle">Locale (yalnızca en), Logger, DevWatch,</text><text class="d-small" x="160" y="376" text-anchor="middle">TrailingSlash, Observability</text>
<line class="d-flow" x1="160" y1="386" x2="160" y2="404" marker-end="url(#config-shape-arrow)" />
<rect class="d-cmd" x="90" y="408" width="140" height="24" rx="12" /><text class="d-cmd-text" x="160" y="424" text-anchor="middle">collage.New</text>
<text class="d-small" x="160" y="448" text-anchor="middle">varsayılanlar, validation, kurulum</text>
</svg>
<figcaption>Scaffold edilmiş <code>main.go</code>, <code>collage.Config</code>'in her alanını nereden alır. Development modu, public origin, sunucunun dinlediği adres ve forgery key'i ortamdan gelir; böylece tek bir binary hem sizin makinenizde hem bir sunucuda çalışır. Port'un <code>-port</code> flag'inin varsayılanı <code>PORT</code>, o da yoksa 6060'tır. Plugin'ler, config'leri, embed edilmiş template'ler ve disk cache <code>main.go</code>'da ve yanındaki dosyalarda yazılıdır. <code>Locale</code>, <code>Logger</code>, <code>DevWatch</code>, <code>TrailingSlash</code> ve <code>Observability</code> atanmaz, bu yüzden her biri varsayılanında kalır. <code>collage.New</code> değerin tamamını alır.</figcaption>
</figure>

Alanlar ve varsayılanları aşağıdaki tablolardadır.
[Eksiksiz bir config](#a-complete-configuration) scaffold'un kodunu,
[Örnek: development ve production için tek bir config](#example-one-configuration-for-development-and-production)
ise production'daki bir uygulamanın ona ne eklediğini gösterir.

## Config

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `DevMode` | `bool` | `false` | Framework genelinde development modu. |
| `DevWatch` | `[]string` | yok | Değiştiğinde development'taki page'i yeniden yükleyen ek dizinler. v0.10.0'dan beri. |
| `Logger` | `*slog.Logger` | başlangıçta seçilir | Framework'ün ve plugin'lerin log yazdığı logger. |
| `Server` | `ServerConfig` | | HTTP sunucusu. |
| `BaseURL` | `string` | yok | Sitenin public origin'i, `https://example.com`; bir render'ın başka türlü bilemeyeceği mutlak URL'ler için. |
| `Security` | `SecurityConfig` | | Request forgery koruması. |
| `Template` | `TemplateConfig` | | Template'lerin yüklenmesi ve render edilmesi. |
| `Cache` | `CacheConfig` | | Page cache. |
| `Locale` | `LocaleConfig` | | URL'lerin hangi locale'leri taşıdığı. |
| `TrailingSlash` | `bool` | `false` | Her page'in URL'si `/` ile biter. v0.13.0'dan beri. |
| `Observability` | `ObservabilityConfig` | | Metric'ler ve tracing. |
| `Plugins` | `[]Plugin` | yok | Uygulama kurulurken register edilen plugin'ler. |
| `PluginConfig` | `map[string]json.RawMessage` | yok | Her plugin'in kendi config'i, plugin adına göre key'lenmiş. |

### DevMode

Development modunu açar. Bu modda:

- template'ler her render'dan önce diskten yeniden yüklenir;
- page'ler tarayıcıda kendilerini yeniler;
- başarısız fragment'ler page'in üzerinde gösterilir;
- page cache'ten hiç okunmaz ve disk cache'in yerini memory alır;
- `DevWatch`'ta adı geçen dizinler izlenir;
- built-in error page hatanın hangi fragment'te başladığını söyler, hatanın
  sebebiyle başlar (template kaynaklı bir hatada başarısız olan çağrının dosyası,
  satırı ve sütunu; v0.15.0'dan beri) ve hata zincirinin tamamını gösterir.

**Production'da kapalı olmalıdır.** Bu tanı bilgileri dosya yollarını, host
adlarını ve hata mesajlarında ne varsa hepsini içerir.

Tek şalter `DevMode`'dur: v0.60.0'dan önce aynı işi yapan `Template.DevMode`
kaldırıldı. Bir plugin onu `host.DevMode()` ile okur. Scaffold edilmiş bir proje
bu alanı `COLLAGE_DEV=1`'e göre ayarlar. `collage dev` bu değişkeni sizin yerinize
set eder.

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
kurarken seçer:

- **Terminalde**, ve yalnızca slog'un varsayılan handler'ını hiçbir şey
  değiştirmemişse, insanların okuması için tasarlanmış kompakt bir handler
  kullanılır. Bu handler her kaydı tek satırda ve renkli bir level işaretiyle yazar.
- **Diğer bütün ortamlarda** `slog.Default()` olduğu gibi kullanılır. İstisnası,
  `NO_COLOR` ayarlı değilken `FORCE_COLOR`'ın ayarlı olmasıdır (v0.32.0'dan beri):
  [`collage dev`](/docs/cli#collage-dev) renkli bir terminalde çalıştırdığı program
  için bunu ayarlar.

`slog.SetDefault`'u çağırmış bir uygulama kendi handler'ını korur. Sonuçtan emin
olmak istiyorsanız bir logger verin, örneğin makinelerin okuyacağı loglar için bir
JSON handler. `New` bu alanı `nil` bırakır.

### Plugins ve PluginConfig

`Plugins` içindeki plugin'ler, `New` çalışırken sırayla register edilir. Template'ler
parse edilmeden önce devreye girmesi gereken bir plugin, örneğin template fonksiyonu
ekleyen ya da mount'ları saran bir plugin, `app.RegisterPlugin` ile değil bu alanla
verilmelidir.

`PluginConfig` her plugin'in bölümünü ham JSON olarak tutar. Framework hiçbir dosya
okumaz. Bu alanı istediğiniz gibi doldurabilir ya da
`collage.LoadPluginConfig("plugins-config.json")`'u kullanabilirsiniz. Bu fonksiyon
dosya yoksa `nil` döner. Register edilmiş hiçbir plugin'in adına karşılık gelmeyen bir
key, uygulamanın başlamasını `ErrUnknownPluginConfig` ile engeller; v0.61.0'dan beri
bir bölümün içinde plugin'in tanımadığı bir key de aynısını yapar. Ayrıntılar için
[Plugin'leri yapılandırmak](/docs/plugins#configuring-plugins) bölümüne bakın.

### TrailingSlash

Her page'in tek bir adresi vardır. `TrailingSlash` bu adresin `/` ile bitip
bitmeyeceğini belirler. Açıksa page `/blog/hello/` adresindedir ve `/blog/hello`
oraya `301` ile redirect edilir. Varsayılan olan kapalı durumda bunun tersi
geçerlidir. İsimle üretilen link'ler (`pageURL`, `pageURLIn`, `localeURL`,
`app.URL`) seçilen yazımla çıkar. Bir locale'in ana sayfası da buna dahildir:
açıkken `/tr/`, kapalıyken `/tr`.

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
| `Port` | `int` | `6060` (v0.32.0'dan önce `3000`) | TCP port'u, `1`–`65535`. |
| `ReadTimeout` | `time.Duration` | `15s` | Bir request'i okumanın en fazla ne kadar sürebileceği. |
| `ReadHeaderTimeout` | `time.Duration` | `0` | Yalnızca request header'larını okumanın en fazla ne kadar sürebileceği. Sıfır, `ReadTimeout`'u kullanır (v0.56.0'dan beri). |
| `WriteTimeout` | `time.Duration` | `30s` | Bir response'u yazmanın en fazla ne kadar sürebileceği. |
| `IdleTimeout` | `time.Duration` | `60s` | Bir keep-alive bağlantısının en fazla ne kadar boşta bekleyebileceği. |
| `ShutdownTimeout` | `time.Duration` | `10s` | Graceful shutdown'ın devam eden request'leri ne kadar beklediği. |
| `DrainDelay` | `time.Duration` | `0` | Bir shutdown başladıktan sonra, port kapanmadan önce sunucunun keep-alive'lar kapalı olarak ne kadar daha hizmet verdiği; development'ta yok sayılır (v0.53.0'dan beri). Bkz. [Deployment](/docs/deployment#graceful-shutdown-and-draining). |
| `MaxBodyBytes` | `int64` | 4 MiB | Request'in yönlendirildiği action kendi sınırını koymadığında request body'sine uygulanan sınır. v0.34.0'dan beri middleware'den önce uygulanır. |
| `TrustedProxies` | `[]string` | yok | Sunucunun önündeki proxy'lerin adresleri ve CIDR aralıkları; bunların `X-Forwarded-For`'una inanılır (v0.47.0'dan beri). Liste set edilmişse `X-Forwarded-Proto`'suna da inanılır (v0.56.0'dan beri). |

`Host`'un varsayılanı `localhost`'tur ve bu adrese makinenin dışından erişilemez.
Container içinde bu alanı `0.0.0.0` yapın.

Scaffold edilen bir proje `Host` ve `Port`'u `HOST` ve `PORT` ortam
değişkenlerinden doldurur. `collage dev` altında bu değişkenler, `collage dev`'in
seçtiği bir loopback adresidir; sizin adresinizde ise `collage dev`'in kendisi
dinler. Bu yüzden onları ortamdan okumaya devam edin. Bkz.
[CLI](/docs/cli#collage-dev).

`MaxBodyBytes`'ı `New` doldurmaz. Sıfır, request geldiğinde uygulanan yerleşik
4 MiB (`4 << 20` byte) demektir. Negatif bir değer ise sınır olmadığı anlamına
gelir. Sınırsız bir body, boyutunu anonim bir çağıranın belirlediği bir bellek
kullanımıdır. Bu yüzden bunu bilerek seçin. Bir action kendi sınırını
`WithMaxBodyBytes` ile koyabilir. Ayrıntılar için [Form'lar ve
action'lar](/docs/forms-and-actions) sayfasına bakın.

`TrustedProxies`, bir reverse proxy'nin arkasında istemcinin adresini bilmeyi
sağlayan alandır; orada her request'in `RemoteAddr`'i proxy'ninkidir. Her girdi bir
adres ya da CIDR aralığıdır (`"10.0.0.0/8"`, `"127.0.0.1"`). İkisinden de olmayan bir
girdi `collage.New`'ı, girdiyi adıyla anarak başarısız kılar. Boşken, yani
varsayılanda, `X-Forwarded-For` hiç okunmaz; bu yüzden istemci onu taklit edemez ve
`collage.ClientIP(r)`, `RemoteAddr`'in host'udur:

```go
Server: collage.ServerConfig{
	TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1"},
},
```

Güvenilen bir proxy'den gelen request, istemcisini `X-Forwarded-For`'da bildirir ve
header sağdan okunur: güvenilen proxy olmayan ilk adres istemcidir. Adreslerin
hepsine güveniliyorsa en soldaki adres istemcidir.

- Bir girdi port (`9.9.9.9:4567`, `[2001:db8::1]:443`) ya da köşeli parantez
  (`[2001:db8::1]`) taşıyabilir.
- Güvenilmeyen bir adresten önce gelen ve adres olmayan bir girdi (`unknown`)
  istemciyi bilinmez kılar: `ClientIP` sıfır değerli `netip.Addr` olur, hiçbir
  zaman proxy'nin adresi olmaz.

Bir CDN'in yayımladığı aralıklar dahil her hop'u listeye yazın. Yalnızca kendi
işlettiğiniz proxy'leri ya da platformunuzun belgelediği aralıkları yazın.
İstemcinin request gönderebildiği bir aralığa güvenmek, o istemcinin istediği
adresi kendine vermesine izin vermektir. Sıfır bitlik bir aralık (`0.0.0.0/0`,
`::/0`) herkese güvenir ve `collage.New` bunun için bir Warn log'lar.

v0.56.0'dan beri forgery cookie'sinin kimden gelen bir `X-Forwarded-Proto` ile
`Secure` olarak işaretleneceğine de aynı liste karar verir. Liste boşken bu header'a,
önceden olduğu gibi, herkesten inanılır.
Bkz. [Deployment](/docs/deployment#behind-a-proxy-trustedproxies).

## SecurityConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `CSRFKey` | `[]byte` | her process için üretilir | Forgery token'larını imzalayan key. |
| `CSRFCookieName` | `string` | `"collage_csrf"` | Token'ı taşıyan cookie. |
| `CSRFFieldName` | `string` | `"_csrf"` | Token'ın gönderildiği form alanı. |
| `CSRFHeaderName` | `string` | `"X-CSRF-Token"` | Token'ın alternatif olarak gönderilebileceği header. |
| `CSRFTrustedOrigins` | `[]string` | yok | Form'ları buraya post edebilecek diğer origin'ler. Her biri `scheme://host[:port]` biçimindedir (v0.34.0'dan beri). |
| `DisableCSRF` | `bool` | `false` | Forgery kontrolünü bütün uygulama için kapatır. |
| `CSRFTokenTTL` | `time.Duration` | `12h` | Bir forgery token'ının üretildikten sonra ne kadar geçerli kaldığı; negatif değer sona ermeyi kapatır (v0.39.0'dan beri). |
| `FrameOptions` | `string` | `"SAMEORIGIN"` | Her response'ta gönderilen `X-Frame-Options`; `"-"` hiçbirini göndermez (v0.39.0'dan beri). |
| `DisableNoSniff` | `bool` | `false` | Her response'ta gönderilen `X-Content-Type-Options: nosniff` header'ını durdurur (header v0.39.0'dan beri; v0.60.0'dan önce bu alan `NoSniff *bool` idi). |

**Form içeren herhangi bir şeyi deploy etmeden önce `CSRFKey`'i ayarlayın.** Key en
az 32 byte rastgele veri olmalı, diğer secret'larınızla birlikte saklanmalı ve her
instance'ta aynı olmalıdır.

Boş bırakılırsa her process için yeni bir key üretilir.
Bu ilk çalıştırma için yeterlidir ama deploy için yanlıştır. Çünkü restart'tan önce ya
da başka bir instance tarafından üretilmiş bir token reddedilir. Uygulama key
ürettiğini log'lar. Development dışında bu bir uyarı, development'ta ise info
seviyesinde bir kayıttır. Bu log yalnızca unsafe bir method'a (`POST`, `PUT`, `PATCH`,
`DELETE`) cevap veren ve `WithoutCSRF` ile muaf tutulmamış bir action varsa yazılır.
Çünkü token'ı yalnızca böyle bir action doğrular (`WithoutCSRF` muafiyeti v0.11.0'dan
beri vardır).

Key disk cache'in namespace'ine dahil değildir, bu yüzden cache yeni bir key'e
rağmen korunur. İçinde form olan ve eski key ile saklanmış cache'lenmiş bir page
ise sunulmaz, yeniden render edilir. Ayrıntılar için
[Caching](/docs/caching#the-namespace) sayfasına bakın. Scaffold edilmiş `main.go`
key'i `COLLAGE_CSRF_KEY`'den okur. `openssl rand -hex 32` ile bir key
üretebilirsiniz.

`CSRFTrustedOrigins`, bu origin dışında form'ları buraya post edebilecek origin'leri
belirtir. Diğer her origin, geçerli bir token taşısa bile reddedilir. Ayrıntılar için
[Form'lar ve action'lar](/docs/forms-and-actions#where-the-request-came-from)
sayfasına bakın.

Yukarıdaki isim varsayılanlarını `New` değil forgery guard uygular. Bu
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
| `Type` | `CacheType` | açıksa `CacheMemory` | `Store` `nil` olduğunda kullanılan built-in cache: `collage.CacheMemory` ya da `collage.CacheDisk`. |
| `DefaultTTL` | `time.Duration` | `5m` | Page bir süre koymadığında entry'nin ömrü. |
| `MaxEntries` | `int` | `10000` | Cache entry sayısının ve tek bir tag altında izlenen key sayısının üst sınırı; memory ve disk için de geçerli. Negatif değer sınırsız entry demektir (disk için v0.39.0'dan beri). |
| `MaxBytes` | `int64` | memory için 256 MiB, disk için 1 GiB | Built-in bir cache'in sakladığı toplam boyutun üst sınırı. Negatif değer sınırsız demektir (v0.40.0'dan beri). Bir disk cache'in memory katmanı 256 MiB tutar, `MaxBytes` bundan küçükse de `MaxBytes` kadar tutar (v0.63.0'dan beri). |
| `Dir` | `string` | yok | Disk cache'in entry'leri, `Dir/data/` altında ise `collage.Cached` değerlerini sakladığı yer. `CacheDisk` için zorunludur. Her instance'ın kendi dizini olmalıdır: bir dizini paylaşan instance'lar bir cache'i paylaşmaz. |
| `Version` | `string` | binary'den türetilir | Disk cache'teki entry'lerin hangi build'e ait olduğunu belirler. |
| `WarmLimit` | `int` | `1000` | Tek bir invalidation'ın arka planda yeniden render ettiği page sayısının üst sınırı; en yeniler önce gelir. Negatif değer warming'i kapatır (v0.64.0'dan beri). |

**Caching varsayılan olarak kapalıdır.** `Enabled` false iken bir `Store` ayarlı olsa
bile hiçbir şey cache'lenmez. `Type`'ın varsayılanı yalnızca `Enabled` true ve `Store`
`nil` olduğunda `collage.CacheMemory` olur. Tipsiz bir literal, `Type: "disk"`, yine
derlenir; bir değişkenden okunan string'in ise dönüştürülmesi gerekir:
`Type: collage.CacheType(v)`.

**Bir `Store`, validation dahil `Type`'ın yerini tamamen alır.** Page'leri Redis'e
ya da başka bir yere koymak için `collage.Cache`'i (`Get`, `Set`, `InvalidateKey`,
`Clear`) implement edin. `collage.TaggedCache`'i (`SetTagged`, `Invalidate`) de
implement eden bir store'a her entry yazılırken o entry'nin tag'leri de verilir ve
tag'e göre invalidate etmesi istenir. v0.60.0'dan önce `Invalidate`, `Cache`'in
kendi metoduydu. Ayrıca `collage.StaleCache`'i (`GetStale`, `MarkStaleKey`,
`MarkStale`, v0.63.0'dan beri) implement eden bir store'dan, invalidate edilen
entry'leri silmek yerine stale olarak işaretlemesi istenir. Böyle bir store
başarısız bir yeniden render'a eski page'le cevap verebilir. Bkz. [Caching](/docs/caching#stale-pages-in-your-own-store).
Somut bir tipin nil pointer'ını atamak yerine alanı boş bırakın. Böyle bir pointer
nil olmayan bir interface'tir ve ilk lookup'ta panic'e yol açar.

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

**Disk cache'in önünde memory vardır** (v0.63.0'dan beri). Bir hit, memory
katmanından sunulur. Bu katmanı `MaxEntries` ve 256 MiB (`MaxBytes` bundan küçükse
`MaxBytes`) sınırlar. Katman açılışta geçerli build'in dizininden, en yenilerden
başlanarak doldurulur. Böylece bir restart son page'lerini hemen memory'den sunar.
Her instance'ın kendi katmanı vardır, bu yüzden her instance'a kendi `Dir`'ini
verin. Paylaşılan bir dizinde, bir instance'taki invalidation'dan sonra diğerleri
kendi memory'lerindeki kopyayı taze diye sunmaya devam eder. Bkz.
[Caching](/docs/caching#memory-in-front-of-the-disk).

**`Dir/data/`, disk cache'in `collage.Cached` değerlerini tutar.** Böylece
uygulama restart'tan sonra boş bir data cache'le başlamaz. Entry'lerin aksine bu
değerler build'in alt dizininde durmaz ve şifrelenmez. Bkz.
[Caching](/docs/caching#across-a-restart).

**`MaxEntries`**, `collage.Cached`'in request'ler arasında tuttuğu değerleri ve
built-in bir cache'le dependency tracker'ın toplamda tuttuğu key sayısını da
sınırlar (v0.40.0'dan beri).

**`MaxBytes`**, built-in bir cache'in ne kadar yer kapladığını sınırlar: sıfır,
tipin varsayılanını seçer. Sınırın tamamından büyük bir page sunulur ama
saklanmaz. Kendi `Store`'unuz bunu yok sayar. Bkz.
[Caching](/docs/caching#bounded-in-bytes-too).

**`MaxEntries` framework'ün dependency tracker'ını tag başına da sınırlar.**
Bu tracker, her process'te bir tag'i cache key'lerine eşleyen bir index'tir.
v0.60.0'dan önce bu ayrı bir alandı: `MaxKeysPerTag`.

Her farklı query string ayrı bir key'dir. Cache bir entry'yi evict ettiğinde ya da
entry'nin süresi dolduğunda key tracker'dan silinmez. Bu yüzden bir üst sınır
olmasa bir client bu index'i sınırsızca büyütebilirdi. Bir tag sınıra ulaştığında
o tag'in en eski key'i yalnızca tracker'dan düşürülür, cache'ten silinmez. Negatif
bir `MaxEntries`, tag başına bu sınırı 10000'de tutar.

Bunun bedeli store'a göre değişir. Built-in memory ve disk cache'ler tag'leri
kendileri index'ler (`TaggedCache`'i implement ederler). Bu yüzden `InvalidateTags`
tuttukları her entry'ye yine ulaşır. Yalnızca `InvalidateTagsN`'in raporladığı sayı,
yani tracker'ın çözümlediği sayı, daha düşük çıkabilir. `TaggedCache`'i implement
etmeyen custom bir `Store` ise yalnızca tracker'a güvenir. Onun için düşürülen bir
key, `InvalidateTags`'in artık ulaşamadığı bir entry demektir ve bu entry süresi
dolana kadar sunulmaya devam eder. Böyle bir store kullanıyorsanız `MaxEntries`'i,
tek bir tag'in kapsayabileceği canlı entry sayısının üzerinde bir değere ayarlayın.

`MaxEntries` ve `MaxBytes`'ta sıfır varsayılan değer demektir. Sınırsız anlamına
gelen tek değer negatif bir değerdir. Ayrıntılar için [Caching](/docs/caching)
sayfasına bakın.

## LocaleConfig

| Alan | Tip | Varsayılan | Anlamı |
| --- | --- | --- | --- |
| `Default` | `string` | `"en"` | Locale prefix'i olmayan bir URL'nin locale'i. |
| `Supported` | `[]string` | `[Default]` | Uygulamanın sunduğu bütün locale'ler. |
| `PrefixDefault` | `bool` | `false` | `Default`'un page'leri de prefix alır: `/en/about`. v0.14.0'dan beri. |

Locale'i seçen tek şey URL'dir. `/about` `Default` locale'indedir, `/tr/hakkinda` ise
`"tr"` locale'indedir. collage locale'i hiçbir zaman `Accept-Language`'ten ya da bir
cookie'den seçmez. Çünkü farklı okuyuculara farklı şey ifade eden bir URL'yi cache'ler,
crawler'lar ve paylaşılan link'ler hep yanlış ele alır.

Varsayılan locale'in kendi prefix'i olan `/en/about`, `/about`'a kalıcı olarak
redirect edilir. `PrefixDefault` ayarlıysa durum tersine döner: adres `/en/about`
olur ve `/about` oraya redirect edilir. Ayrıntılar için
[Link'ler ve locale'ler](/docs/links-and-locales#the-url-decides-the-locale)
sayfasına bakın.

Dili kendiniz belirlemek istiyorsanız bunu middleware'de yapın. Türkçe bir
tarayıcıyı `/tr`'ye redirect edebilirsiniz. Ya da tek bir URL'yi her dil için ayrı
render edip bunu `collage.Vary` ile bildirebilirsiniz. Ayrıntılar için
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
`CacheEvict`, `CacheInvalidate`, `CacheCoalesced`, `CacheStale` ve `CacheWarm`.
`CacheCoalesced`, bir request'in aynı key için zaten çalışmakta olan bir render
tarafından karşılandığı anlamına gelir. `CacheStale` ise bir `collage.Cached`
değerinin ya da cache'lenmiş bir page'in stale olarak sunulduğu anlamına gelir.
Sunulan her değer ya da response için bir kez sayılır. `CacheWarm` (v0.64.0'dan beri) ise bir
invalidation'ın ulaştığı bir page ya da document'ın arka planda yeniden render
edilip saklandığı anlamına gelir. Her key için bir kez sayılır.

## Validation

`New`, varsayılanları doldurduktan sonra aşağıdaki durumlardan bulduğu ilkini
döner:

| Koşul | Hata |
| --- | --- |
| `Server.Port` `1`–`65535` aralığının dışında | `ErrInvalidPort` |
| `Cache.Enabled` açık, `Store` yok ve `Type` ne `CacheMemory` ne `CacheDisk` | `ErrInvalidCacheType` |
| `Cache.Enabled` açık, `Store` yok, `Type` `CacheDisk` ve `Cache.Dir` boş | `ErrEmptyCacheDir` |
| `Locale.Default`, `Locale.Supported` içinde yok | `ErrLocaleDefaultNotSupported` |
| `Server.ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`, `DrainDelay`, `Template.Timeout` ya da `Cache.DefaultTTL` negatif | `ErrNegativeDuration` |
| `BaseURL` set edilmiş ve yalın bir `scheme://host[:port]` değil | `ErrInvalidOrigin` |

Negatif süre hatası sekiz alanın hepsi için tek bir sentinel error'dır. Mesaj, hataya
yol açan alanın adını verir, örneğin `server.read_timeout`. Varsayılanlar önce
uygulandığı için `New` üzerinden gelen sıfır bir port ya da boş bir varsayılan locale
validation'a hiç ulaşmaz. Yalnızca açıkça set ettiğiniz bir değer hataya yol
açabilir. Bu hataları `errors.Is` ile karşılaştırın. Tam liste
[Hatalar](/docs/errors#configuration) sayfasındadır.

Varsayılanları doldurmak ve validation v0.60.0'dan beri ayrı ayrı export edilmez.
İkisinin çalıştığı tek yer `New`'dur, bu yüzden bir test bir config'i onu çağırarak
kontrol eder:

```go
_, err := collage.New(&collage.Config{Server: collage.ServerConfig{Port: 70000}})
// errors.Is(err, collage.ErrInvalidPort)
```

Varsayılanları doldurmak idempotent'tir. Varsayılanları zaten taşıyan bir config
olduğu gibi kalır.

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
		Port: envInt("PORT", 6060),
	},
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       collage.CacheDisk,
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
değere dönerler. Scaffold ayrıca `BaseURL`'yi `BASE_URL`'den okur ve `Plugins`'i,
`collage plugin add` ve `del` komutlarının yönettiği `plugins.go`'daki
`plugins()` ile doldurur.

## Örnek: development ve production için tek bir config

[Ekibin iş listesi](/docs/middleware-and-apis#example-a-teams-task-list) (her üye
giriş yapar, admin ekibin her görevini, member ise yalnızca kendi görevlerini
değiştirebilir) bir geliştiricinin makinesinde `collage dev` altında, production'da
ise [bir load balancer'ın arkasındaki iki sunucuda](/docs/deployment#example-deploying-a-teams-task-list)
çalışır. Page'leri İngilizce ve Türkçedir. Ekibin kendi subdomain'indeki admin
paneli de buraya form post eder.

İki ortama da tek bir binary hizmet verir. Bu binary'nin `newApp` fonksiyonu
scaffold'unkiyle aynıdır, yalnızca production ayarları eklenmiştir. İki ortam
arasında farklı olan her şey ya ortamdan okunur ya da `devMode`'a göre belirlenir:

```go
func newApp(devMode bool, port int) (*collage.App, error) {
	pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
	if err != nil {
		return nil, fmt.Errorf("plugin configuration: %w", err)
	}

	cfg := &collage.Config{
		DevMode: devMode,
		BaseURL: os.Getenv("BASE_URL"), // empty in development unless .env.development sets it
		Server: collage.ServerConfig{
			Host:              envString("HOST", "localhost"),
			Port:              port,
			ReadHeaderTimeout: 5 * time.Second,
			DrainDelay:        5 * time.Second, // ignored in development
		},
		Template: collage.TemplateConfig{
			FS:        templatesFS,
			Root:      "templates",
			Extension: ".html",
		},
		Cache: collage.CacheConfig{
			Enabled:    true,
			Type:       collage.CacheDisk,
			Dir:        cacheDir,
			DefaultTTL: 5 * time.Minute,
		},
		Locale: collage.LocaleConfig{
			Default:   "en",
			Supported: []string{"en", "tr"},
		},
		PluginConfig: pluginConfig,
		Plugins:      plugins(),
		Security: collage.SecurityConfig{
			// Empty in development: a key is generated per process.
			CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
		},
	}

	if !devMode {
		// The load balancer in front of both servers, and the admin
		// panel whose forms post here.
		cfg.Server.TrustedProxies = []string{"10.0.1.10"}
		cfg.Security.CSRFTrustedOrigins = []string{"https://admin.tasks.example.com"}
	}

	app, err := collage.New(cfg)
	if err != nil {
		return nil, err
	}
	// Routes and the static mount as scaffolded.
	return app, nil
}
```

Her ayarın iki ortamdaki değeri ve bunun nedeni şöyledir:

| Ayar | Development | Production | Neden |
| --- | --- | --- | --- |
| `DevMode` | açık: `collage dev`, `COLLAGE_DEV=1` set eder | kapalı: değişken set edilmez | Development'ın error page'leri dosya yollarını ve hata zincirlerini taşır. Bir development sunucusu ayrıca yalnızca `Host`'u bu makineyi adlandıran request'lere cevap verir; bkz. [Deployment](/docs/deployment#development-mode-stays-on-this-machine). |
| `Server.Host`, `Server.Port` | `collage dev`'in seçtiği loopback adresi | `HOST` ve `PORT`, ya da `-port` | Sizin adresinizde `collage dev`'in kendisi dinler ve request'leri iletir; bkz. [CLI](/docs/cli#collage-dev). |
| `BaseURL` | `.env.development` `BASE_URL`'i set etmedikçe boş | `BASE_URL`'den `https://tasks.example.com` | Boş olması hata değildir: hiç mutlak URL üretmeyen bir site için gerekmez. Set edilmişse yalın bir origin olmalıdır, yoksa `New` `ErrInvalidOrigin` ile başarısız olur. |
| `Security.CSRFKey` | boş: her process'e bir key, info seviyesinde log'lanır | en az 32 byte rastgele veri, iki sunucuda da aynı | Üretilmiş bir key, bir restart'tan önce ya da diğer sunucunun ürettiği her token'ı reddeder. |
| `Security.CSRFTrustedOrigins` | yok | admin panelinin origin'i | Diğer her origin, geçerli bir token taşısa bile reddedilir. |
| `Server.TrustedProxies` | yok | load balancer | Bu olmadan her istemcinin adresi load balancer'ınkidir. |
| `Server.DrainDelay` | 5s, yok sayılır | 5s | Load balancer, port kapanmadan önce trafik göndermeyi bırakır. |
| `Server.ReadHeaderTimeout` | 5s | 5s | Header'ları byte byte gönderen bir istemci erkenden düşürülür; bir body ise yine `ReadTimeout`'un tamamını kullanır. |
| `Cache` | aynı satırlar: yerine memory geçer, hiç okunmaz | `.cache` altında bir disk cache, her sunucuda bir tane | Development az önce düzenlediğiniz template'i gösterir. Bir disk cache restart'tan sonra sıcak açılır. |
| `Locale` | `en`, `tr` | `en`, `tr` | Locale'i URL seçer, bu yüzden binary'nin nerede çalıştığına bağlı hiçbir yanı yoktur. |
| `WriteTimeout`, `Template.Timeout` | varsayılanlar, 30s ve 5s | varsayılanlar, 30s ve 5s | `Template.Timeout`'u `WriteTimeout`'un epey altında tutmak, yavaş bir upstream'i page'in yarısında kapanan bir bağlantı yerine bir fragment'in fallback'ine çevirir; bkz. [Deployment](/docs/deployment#timeouts). |

Development tarafında hiçbir şey yazmanız gerekmez: `collage dev`, `COLLAGE_DEV`,
`HOST` ve `PORT`'u set eder. Go kodu değiştiğinde `collage dev`'in yaptığı
restart'larda bir form açık kalsın istiyorsanız development'a `.env.development`'ta
kendine ait bir key verin. Bu key hiçbir zaman production'daki key olmamalıdır;
bkz. [CLI](/docs/cli#environment-files).

Production tarafı, binary'nin çalıştığı yerde set edilen dört değişkendir: `HOST`,
`PORT`, `BASE_URL` ve `COLLAGE_CSRF_KEY`. Bunlara bir plugin'in `KeyEnv` ile
adlandırdığı değişkenler eklenir. Orada `.env` dosyalarını hiçbir şey
okumaz; bkz. [Deployment](/docs/deployment#environment).
