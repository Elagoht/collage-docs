---
description: collage'ın request forgery'ye, open redirect'lere, path hilelerine, cache sızıntılarına ve aşırı büyük body'lere karşı yaptıkları ve size bıraktıkları.
reference: SecurityConfig, ErrCSRFCrossOrigin, Vary
---

# Güvenlik

Bir framework'teki bug, onunla kurulan her sitedeki bir bug'dır. Bu yüzden collage,
her sitenin aksi hâlde hatırlaması gerekecek güvenlik işlerini kendisi yapar.

Bu sayfa, collage'ın sizin için yaptıklarını tek bir yerde toplar ve her birinin
anlatıldığı yere link verir. Yapmadıkları ise sonda yer alır.

Her request collage'ın kontrolleriyle hep aynı sırayla karşılaşır ve her kontrol,
kendisinden sonra gelen çalışmadan önce cevap verebilir:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 490" role="img" aria-labelledby="sec-flow-title">
<title id="sec-flow-title">Bir request önce security header'larını alır, sonra path'i temizlenir: kirli bir path redirect edilir, encode edilmiş bir slash 404, bozuk bir escape 400 olur. Ardından düz app.Handle ile mount edilmiş bir path tamamen sizindir ve onun için hiçbir middleware çalışmaz. Geri kalan her şey middleware'lerle karşılaşır: sizinkiler ve plugin'lerinki. Bir page, action'ları ve fragment path'leri cache'ten önce page'in guard'larıyla karşılaşır; bir guard'ın kendi cevabı no-store'dur. Bir page cache'ten sunulur ya da cookie içermeyen bir key ile render edilir; guard'lıysa private, no-cache ile gider. Bir fragment path, GET'e cevap veren bir action'dır ve hiçbir zaman cache'lenmez. Bir action önce body limitiyle (413), sonra unsafe bir method için origin ve token ile (403), sonra handler'ınızla karşılaşır. Production'da her error page yalnızca status'u gösterir.</title>
<defs><marker id="sec-flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">request</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="52" width="184" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">security header'ları</text>
<line class="d-flow" x1="207" y1="65" x2="218" y2="65" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="54" width="96" height="22" rx="11" /><text class="d-small" x="270" y="68" text-anchor="middle">her response</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="98" width="184" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">path temizlenir</text>
<line class="d-flow" x1="207" y1="111" x2="218" y2="111" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="94" width="96" height="34" rx="6" /><text class="d-small" x="270" y="108" text-anchor="middle">redirect</text><text class="d-small" x="270" y="121" text-anchor="middle">404 · 400</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="23" y="144" width="184" height="36" rx="6" /><text class="d-on-c" x="115" y="160" text-anchor="middle">middleware</text><text class="d-on-c d-small" x="115" y="173" text-anchor="middle">sizin ve plugin'lerin</text>
<path class="d-flow" d="M115 132H270V140" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="222" y="144" width="96" height="36" rx="6" /><text class="d-on-c" x="270" y="160" text-anchor="middle">app.Handle</text><text class="d-on-c d-small" x="270" y="173" text-anchor="middle">tamamen sizin</text>
<line class="d-flow" x1="115" y1="180" x2="115" y2="196" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="23" y="200" width="184" height="36" rx="6" /><text x="115" y="216" text-anchor="middle">guard'lar, cache'ten önce</text><text class="d-small" x="115" y="229" text-anchor="middle">page, action, fragment path</text>
<line class="d-flow" x1="207" y1="218" x2="218" y2="218" marker-end="url(#sec-flow-arrow)" />
<rect class="d-chip" x="222" y="201" width="96" height="34" rx="6" /><text class="d-small" x="270" y="215" text-anchor="middle">guard'ın cevabı</text><text class="d-small" x="270" y="228" text-anchor="middle">no-store</text>
<path class="d-flow" d="M95 236C95 248 80 246 80 258" marker-end="url(#sec-flow-arrow)" />
<path class="d-flow" d="M135 236C135 248 240 246 240 258" marker-end="url(#sec-flow-arrow)" />
<text class="d-small" x="50" y="254" text-anchor="middle">page</text>
<text class="d-small" x="272" y="254" text-anchor="middle">action</text>
<rect class="d-frame" x="4" y="262" width="152" height="36" rx="6" /><text x="80" y="278" text-anchor="middle">cache ya da render</text><text class="d-small" x="80" y="291" text-anchor="middle">key'de cookie yok</text>
<line class="d-flow" x1="80" y1="298" x2="80" y2="310" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="4" y="314" width="152" height="36" rx="6" /><text x="80" y="330" text-anchor="middle">guard'lıysa</text><text class="d-small" x="80" y="343" text-anchor="middle">private, no-cache</text>
<rect class="d-frame" x="164" y="262" width="152" height="36" rx="6" /><text x="240" y="278" text-anchor="middle">body limiti</text><text class="d-small" x="240" y="291" text-anchor="middle">413 · çok büyük</text>
<line class="d-flow" x1="240" y1="298" x2="240" y2="310" marker-end="url(#sec-flow-arrow)" />
<rect class="d-frame" x="164" y="314" width="152" height="36" rx="6" /><text x="240" y="330" text-anchor="middle">origin, sonra token</text><text class="d-small" x="240" y="343" text-anchor="middle">403 · unsafe method'lar</text>
<line class="d-flow" x1="240" y1="350" x2="240" y2="362" marker-end="url(#sec-flow-arrow)" />
<rect class="d-c" x="164" y="366" width="152" height="26" rx="6" /><text class="d-on-c" x="240" y="383" text-anchor="middle">handler'ınız</text>
<line class="d-rule" x1="23" y1="408" x2="297" y2="408" />
<rect class="d-frame" x="23" y="418" width="274" height="36" rx="6" /><text x="160" y="434" text-anchor="middle">production'da her hata</text><text class="d-small" x="160" y="447" text-anchor="middle">yalnızca status: metin yok, stack yok</text>
<rect class="d-frame" x="96" y="468" width="12" height="12" rx="2" /><text class="d-small" x="114" y="478">collage</text>
<rect class="d-c" x="180" y="468" width="12" height="12" rx="2" /><text class="d-small" x="198" y="478">sizin</text>
</svg>
<figcaption>collage'ın kendi adımları çerçeveli, size ait olanlar dolu çizilmiştir. Her request, herhangi bir middleware çalışmadan önce security header'larını ve temiz bir path'i alır. Ardından düz app.Handle ile mount edilmiş bir path tamamen sizindir: middleware, guard, forgery kontrolü, body limiti ve cache yoktur. Geri kalan her şey middleware'lerle karşılaşır: sizinkiler ve plugin'lerinki. Bir page, action'ları ve fragment path'leri cache'ten önce page'in guard'larıyla karşılaşır. Ardından bir action, handler'ınızdan önce body limitiyle (body herhangi bir middleware çalışmadan önce sınırlanır; body'si çok büyük bir action burada reddedilir), sonra origin ve token ile karşılaşır; bir fragment path, GET'e cevap veren bir action'dır ve hiçbir zaman cache'lenmez. Production'da bir error page status'tan başka bir şey göstermez.</figcaption>
</figure>

## Form'lar sahte olarak gönderilemez

Unsafe bir method'a sahip her action, handler'ı çalışmadan önce imzalı bir token'ı
kontrol eder. `{{csrfToken}}` içeren bir page yine de cache'lenir: her okuyucuya
kendi token'ı gönderilir.

Token tek başına kontrol edilmez. Tarayıcının başka bir origin'den geldiğini
işaretlediği bir gönderim (`Sec-Fetch-Site` ile ya da `Host`'a karşı `Origin` ile)
ne taşırsa taşısın reddedilir. Kardeş subdomain'ler de buna dahildir, bu yüzden
yerleştirilmiş bir cookie içeri girmenin yolu değildir. Buraya post etmesi
amaçlanan bir origin `Security.CSRFTrustedOrigins`'te belirtilir. Ayrıntılar için
[Form'lar ve action'lar](/docs/forms-and-actions#forgery-protection) sayfasına
bakın.

Deploy etmeden önce `Security.CSRFKey`'i ayarlayın: üretilen bir key her process'te
farklıdır.

Bir token, üretildiği zamanı imzalı olarak taşır ve `Security.CSRFTokenTTL`'den daha
eski olduğunda reddedilir. Varsayılan on iki saattir, böylece sızan bir token sonsuza
kadar tekrar kullanılamaz.

Uzun süredir açık bir form'u olan bir okuyucu reddedilmek yerine yeni bir token
alır. Bir token'ı imzası geçerli olduğu sürece geçerli tutmak için
`CSRFTokenTTL`'i negatif yapın.

v0.72.0'dan beri bir form'daki token her response'ta yeniden maskelenir: her
seferinde yeni, rastgele bir pad ile XOR'lanır ve başına `~` konur. Request'in
gönderdiğini yansıtan sıkıştırılmış bir response, yansıyan metin token'ın bir
kısmıyla eşleştiğinde daha kısa olur. Bir okuyucunun tarayıcısına request
gönderttirebilen bir saldırgan da token'ı bu uzunluklardan okuyabilirdi (BREACH).
Maskelenmiş token her seferinde farklı byte'lardır. Kontrol maskeyi hem form'dan
hem header'dan gelen değerden kaldırır ve maskelenmemiş bir token'ı da kabul eder.

## Redirect'ler sitede kalır

collage'ın yazdığı hiçbir redirect, okuyucuyu başka bir siteye gönderemez:

- `/a/./b` ya da `/a//b`'yi temiz yazımına gönderen redirect, path'i escape ederek
  yazar. Böylece `/./%5Cevil.com`, `/\evil.com`'a değil `/%5Cevil.com`'a gider.
  Tarayıcı `/\evil.com`'u `//evil.com` olarak okur.
- Locale ve trailing slash redirect'leri hiçbir zaman `//` ile başlamaz.
- `//host` ya da `/\host`'a giden register edilmiş bir redirect, register sırasında
  reddedilir. Yakaladığı bir değer de düştüğü yere göre escape edilir: path'te ya da
  tek bir query değeri olarak. Böylece `/login?next={slug}`'a ikinci bir `next`
  verilemez.

Bir action'ın `Location`'ı ve bir guard'ınki sizindir: collage onları sizin
ayarladığınız gibi yazar. Bu yüzden request'ten, kontrol etmeden bir `Location`
oluşturmayın. Bazıları siteden çıkmak zorundadır (bir giriş sağlayıcısının
sayfası), collage'ın onları kontrol etmemesinin sebebi budur.

URL'den alınan bir `next` için collage v0.44.0'dan beri kendi kontrolünü sunar:

```go
loc := collage.SafeRedirect(rc.Request.URL.Query().Get("next"), "/")
```

`SafeRedirect`, `next` bu sitede bir path ise onu döndürür: `/` ile başlar, `//`
ile başlamaz, ters bölü ve kontrol karakteri içermez, temizlendikten sonra da tek
bir `/` ile başlar. Değilse fallback'i döndürür (fallback da path değilse `/`).

Temizleme önemlidir çünkü `http.Redirect` köklü bir path'i temizler:
`/./\evil.example`, `/\evil.example` olarak çıkar ve tarayıcı bunu
`//evil.example` diye okur. Mutlak bir URL, bu sitenin kendi origin'indeki
bile olsa hiçbir zaman kabul edilmez.

## Bir path tek bir anlama gelir

Middleware decode edilmiş `r.URL.Path`'i okur, router ise escape edilmiş olanı. İkisi
arasında bir fark çıkabilecek her yerde request hiçbir yere ulaşmaz:

- Dot segment'ler ya da çift slash içeren bir path, middleware çalışmadan önce temiz
  yazımına redirect edilir. Böylece `/admin/` üzerindeki bir kontrol `/x/../admin`
  ile hiç karşılaşmaz.
- Encode edilmiş bir slash (`%2F`) içeren bir path 404 döner. `%2F` bir ayırıcı
  değildir, ama middleware decode edilmiş path'i okur ve orada `/public%2Fsecret`,
  `/public/secret`'tir: middleware için iki segment, router için tek segment.
- Bozuk bir escape 500 değil 400 döner.

Ayrıntılar için
[Path'ler önce temizlenir](/docs/middleware-and-apis#paths-are-cleaned-first)
bölümüne bakın.

## Private page'ler private kalır

Bir [guard](/docs/pages-and-layouts#private-pages-guards), page'in render'ları,
URL'sindeki action'lar ve fragment path'leri için cache okunmadan önce çalışır.

Guard'lı bir page `private, no-cache` ile, bir guard'ın kendi cevabı ise `no-store`
ile gider. Böylece bir CDN, bir okuyucunun page'ini ya da aldığı reddi hiçbir zaman
sonrakine vermez. Static export guard'lı page'leri hiç yazmaz.

## Cache'lenmiş bir page kimseye özel değildir

Cache'te saklanan bir render, request'i aynı key'e sahip her okuyucuya sunulur. Bu
yüzden render'a yalnızca key'in içerdikleri verilir:

- path ve host;
- key'deki query parametreleri;
- `collage.Vary` ile bildirilen header'lar.

Bir okuyucunun cookie'leri, credential'ları ya da adresi hiçbir zaman verilmez. Host
key'de yer aldığı için başka bir host adı veren bir request, diğer herkesin aldığı
kopyayı zehirleyemez.

Degraded bir render, bir error page ve development'taki bir page hiçbir zaman public
olarak cache'lenemez. Ayrıntılar için
[Paylaşılan bir render neyi görür](/docs/caching#what-a-shared-render-sees)
bölümüne bakın.

## Body'ler sınırlıdır

Bir request body'si, middleware'iniz çalışmadan önce, yönlendirildiği action'ın
limitiyle sınırlanır. Aksini söylemediğiniz sürece bu limit 4 MiB'tır. v0.71.0'dan
beri middleware yokken de böyledir, yani bir page'in ya da action'ın guard'ı body'yi
sınırlı olarak okur. Daha büyük olduğunu söyleyen bir body, handler çalışmadan önce
reddedilir.

Diske taşan multipart dosyalar, form'u kim parse etmiş olursa olsun request
cevaplandığında silinir. Forgery kontrolünün
hiç imzalamadığı bir cookie ise body daha hiç okunmadan reddedilir. Ayrıntılar için
[Request body'leri sınırlıdır](/docs/forms-and-actions#request-bodies-are-bounded)
bölümüne bakın.

## Bir dosya, uzantısının söylediği şeydir

Bir [asset mount](/docs/assets), bir dosyanın tipini uzantısına göre belirler. Adı
bir tip belirtmeyen dosyanın tipi içeriğinden çıkarılır; böylece uzantısız yüklenen
bir görsel yine görsel olarak gider.

Ama hiçbir zaman çalıştırılabilen bir tipe dönüşmez: HTML ya da XML'e benzeyen bir
dosya `text/plain` olur ve `nosniff` sayesinde tarayıcı buna uyar. Böylece markup
içeren bir upload bir page olarak sunulmaz.

Mount hiçbir zaman bir dizini listelemez; `.well-known` dışında bir dotfile'ı
(`.env`, `.git/`) hiçbir zaman sunmaz ya da export etmez. `collage serve` de
aynısını yapar.

## Hatalar production'da hiçbir şey söylemez

Production'da collage'ın kendi error page'i status'tan başka bir şey göstermez: hata
metni yok, stack yok, path yok.

Development overlay'i ve reload kanalı yalnızca development'ta vardır. `collage dev`
de yalnızca bu makineyi adlandıran bir `Host`'a cevap verir. Bu yüzden başka bir
sitedeki bir page, DNS rebinding yoluyla onları okuyamaz.

Bir log satırındaki request path'i kontrol karakteri içerdiğinde tırnak içine
alınır. Böylece sahte bir satır oluşturamaz ya da terminalinizi yeniden yazamaz.

## Security header'ları

Her response varsayılan olarak `X-Content-Type-Options: nosniff` ve
`X-Frame-Options: SAMEORIGIN` taşır. Böylece bir site, hiçbir şey eklemeden
form'larında MIME-sniff ve clickjacking korumasına sahip olur.

`Security.FrameOptions` değeri belirler: `"-"` hiçbirini göndermez, başka her şey
olduğu gibi gönderilir. `Security.DisableNoSniff` ise nosniff'i kapatır (v0.60.0'dan
önce `NoSniff *bool`).

Tam bir `Content-Security-Policy`, HSTS ve geri kalanını
[elagoht/secure](/docs/plugins#elagohtsecure) plugin'i ekler ve onun header'ları
bunların yerine geçer. Production'daki bir page hiçbir framework script'i
taşımadığı için sıkı bir policy nonce gerektirmez.

## Size kalanlar

- **Güvenlik düzeltmelerini içeren bir Go toolchain'i.** Bir program, derlendiği
  standart kütüphaneyi içinde taşır. v0.72.0'dan beri collage Go 1.26.10 ister.
  Böylece varsayılan `GOTOOLCHAIN=auto` ile go komutu, daha eski bir Go ile
  yapılacak bir derleme için bu sürümü kendisi indirir. Bir binary'yi
  `go version -m` ile kontrol edin. Go 1.26.10, ele geçirilmiş eMudhra ara
  sertifikalarını reddeder (CVE-2026-111943).
- **TLS**, `X-Forwarded-Proto`'yu ve tarayıcının `Host`'unu ileten bir proxy'de.
  Ayrıntılar için [Deployment](/docs/deployment#tls-behind-a-proxy) sayfasına bakın.
- **Rate limit'ler**: [elagoht/ratelimit](/docs/plugins#elagohtratelimit).
- **`app.Handle` ile mount edilmiş bir handler** tamamen sizindir: forgery kontrolü,
  body limiti ve cache yoktur, v0.67.0'dan beri middleware de yoktur.
  `collage.WithMiddleware()` ile mount edilmedikçe hiçbir plugin'in authentication'ı,
  rate limit'i, ban'ı ya da access log'u onu kapsamaz. Ayrıntılar için
  [İstemedikçe middleware yok](/docs/middleware-and-apis#no-middleware-unless-you-ask)
  bölümüne bakın.
- **Bir data handler'ın okudukları.** Cache'lenebilir bir page'de, middleware'in
  request context'ine koyduğu bir değer paylaşılan render'dan gizlenir, böylece bir
  okuyucunun değeri başka bir okuyucunun page'ine ulaşamaz. Bu değeri `collage.Vary`
  ile bildirin ve `collage.Varied` ile okuyun.

Bir güvenlik açığını bildirmek için
[collage repository'sinde](https://github.com/Elagoht/collage/security) GitHub'ın
private reporting özelliğini kullanın.

## Örnek: bir ekibin iş listesi, yayına girmeden önce

[Middleware ve kendi API'niz](/docs/middleware-and-apis#example-a-teams-task-list)
ile [Form'lar ve action'lar](/docs/forms-and-actions#example-a-team-task-list)
sayfalarındaki ekip iş listesi yayına girmek üzere:

- her üye giriş yapar ve bir rolü vardır: bir admin ekibin her görevini, bir üye
  ise yalnızca kendisine atanmış görevleri değiştirebilir;
- ekibin listesi `/teams/{team}/tasks`'teki bir page'dir ve bir "görev ekle"
  form'u taşır;
- ekibin komut satırı aracı için bir JSON API `app.Handle` ile, `WithMiddleware()`
  kullanılarak mount edilmiştir.

Aşağıdaki her riski ya collage, ya uygulamanın eklediği bir plugin ya da
uygulamanın kendisi karşılar:

| Risk | collage, built-in | Plugin | Hâlâ uygulamanın işi |
| --- | --- | --- | --- |
| Başka bir site "görev ekle" form'unu post eder | Unsafe method'lu her action'da önce origin, sonra token | — | Form'da `{{csrfToken}}`; ayarlanmış bir `Security.CSRFKey` |
| Kimin giriş yaptığı | — | elagoht/session: imzalı, `HttpOnly`, `SameSite=Lax` bir cookie; `Encrypt` ile şifreli | Giriş action'ı; okuyucu giriş yaptığında `Regenerate` |
| Giriş yapmamış bir okuyucu listede | Guard'lar cache'ten önce çalışır: page, action'ları ve fragment path'leri için | Guard olarak `session.RequireUser("/login")` | Guard'lı layout'u her private page'e koymak |
| Bir üyenin başka bir üyenin ya da başka bir ekibin görevini değiştirmesi | — | — | Rol ve ekip kontrolü, verinin olduğu yerde: `requireRole` gibi bir guard, ayrıca action ve API |
| Girişten sonraki `next` okuyucuyu siteden çıkarır | `collage.SafeRedirect` | `RequireUser`, `next`'i URL'ye koyar | `next`'i `SafeRedirect`'ten geçirmek |
| `/x/../admin`'in bir path kontrolünü atlatması | Path middleware'den önce temizlenir | — | — |
| Bir üyenin page'inin başka birine sunulması | Guard'lı bir page `private, no-cache`'tir; cache key'i cookie içermez | elagoht/session: session'ı olan bir request taze render edilir, `private, no-store` | — |
| Çok büyük bir upload | Varsayılan olarak action başına 4 MiB; `WithMaxBodyBytes` | — | API'nin kendi sınırı: `http.MaxBytesReader` |
| `/login`'de şifre tahmini | — | elagoht/ratelimit: bir `/login` kuralı | Bir proxy arkasında `Server.TrustedProxies`; bucket'lar process başınadır |
| Bir form'u dolduran bot'lar | — | elagoht/honeypot: `{{honeypot}}`, forgery kontrolünden sonra, handler'dan önce kontrol edilir | elagoht/ratelimit ile birlikte kullanmak |
| Enjekte edilen script'ler, framing, düz HTTP | `nosniff`; `X-Frame-Options: SAMEORIGIN` | elagoht/secure: nonce'lu bir CSP, HSTS | Policy'nin kendisi; proxy'de TLS |
| Bir veritabanı adresini gösteren bir hata | Production error page'i yalnızca status'u gösterir | — | API'nin kendi yazdığı hata metni |
| Çalınmış bir session cookie'si | — | elagoht/session: bir session iptal edilemez, cookie'de yaşar | Kısa bir `maxAge` ya da `idleTimeout` |

Plugin'ler `Config.Plugins`'e girer: elagoht/secure ve elagoht/honeypot template
fonksiyonları ekler (`{{cspNonce}}`, `{{honeypot}}`). Bir plugin'in key'i hiçbir
zaman `plugins-config.json`'dan değil, bir ortam değişkeninden okunur:
`SESSION_KEY`, `SECURE_KEY`, `HONEYPOT_KEY` ya da `KeyEnv`'inin adını verdiği başka
bir değişken:

```go
app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{
		// The proxy that ends TLS, so ClientIP and IsHTTPS believe it.
		TrustedProxies: []string{"10.0.1.10"},
	},
	Security: collage.SecurityConfig{
		CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	},
	Plugins: []collage.Plugin{
		session.New(session.Options{Encrypt: true}),
		secure.New(secure.Options{
			CSP: "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
		}),
		ratelimit.New(ratelimit.Options{Rules: []ratelimit.Rule{
			{PathPrefix: "/login", Rate: 0.1, Burst: 5},
			{Rate: 0.5, Burst: 10},
		}}),
		honeypot.New(honeypot.Options{}),
	},
})
```

İkinci kuralın ne path'i ne de method'u vardır. Bu yüzden kendi skip listesi
dışındaki her yerde `GET`, `HEAD` ve `OPTIONS` dışındaki her method'u sayar. API'nin
`PATCH`'i de buna dahildir, çünkü API `WithMiddleware()` ile mount edilmiştir. Düz
bir `app.Handle` mount'unda hiçbir middleware çalışmaz. Bu yüzden ne bu rate limit
ne de başka bir plugin'in middleware'i onu kapsardı.

Roller uygulamanın işi olarak kalır. Yalnızca admin'leri geçiren guard,
[Örnek: private bir bölümü olan bir blog](/docs/pages-and-layouts#example-a-blog-with-a-private-area)
bölümündeki `requireRole`'dür. API ise ekibi ve atanan kişiyi kendisi kontrol eder;
[Olduğu gibi mount edilen API](/docs/middleware-and-apis#the-api-mounted-as-it-is)
bölümü bunu gösterir. collage kimin neyi değiştirebileceği hakkında hiçbir şey
söylemez.
