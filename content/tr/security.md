---
description: collage'ın request forgery'ye, open redirect'lere, path hilelerine, cache sızıntılarına ve aşırı büyük body'lere karşı yaptıkları ve size bıraktıkları.
reference: SecurityConfig, ErrCSRFCrossOrigin, Vary
---

# Güvenlik

Bir framework'teki bug, onunla kurulan her sitedeki bir bug'dır. Bu yüzden collage,
her sitenin aksi hâlde hatırlaması gerekecek güvenlik işlerini kendisi yapar. Bu
sayfa, collage'ın sizin için yaptıklarını tek bir yerde toplar ve her birinin
anlatıldığı yere link verir. Yapmadıkları ise sonda yer alır.

## Form'lar sahte olarak gönderilemez

Unsafe bir method'a sahip her action, handler'ı çalışmadan önce imzalı bir token'ı
kontrol eder. `{{csrfToken}}` içeren bir page yine de cache'lenir: her okuyucuya
kendi token'ı gönderilir. Token tek başına kontrol edilmez. Tarayıcının başka bir
origin'den geldiğini işaretlediği bir gönderim (`Sec-Fetch-Site` ile ya da `Host`'a
karşı `Origin` ile) ne taşırsa taşısın reddedilir. Kardeş subdomain'ler de buna
dahildir, bu yüzden yerleştirilmiş bir cookie içeri girmenin yolu değildir. Buraya
post etmesi amaçlanan bir origin `Security.CSRFTrustedOrigins`'te belirtilir.
Ayrıntılar için [Form'lar ve action'lar](/docs/forms-and-actions#forgery-protection)
sayfasına bakın.

Deploy etmeden önce `Security.CSRFKey`'i ayarlayın: üretilen bir key her process'te
farklıdır.

Bir token, üretildiği zamanı imzalı olarak taşır ve `Security.CSRFTokenTTL`'den daha
eski olduğunda reddedilir. Varsayılan on iki saattir, böylece sızan bir token sonsuza
kadar tekrar kullanılamaz. Uzun süredir açık bir form'u olan bir okuyucu reddedilmek
yerine yeni bir token alır. Bir token'ı imzası geçerli olduğu sürece geçerli tutmak
için `CSRFTokenTTL`'i negatif yapın.

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
sayfası), collage'ın onları kontrol etmemesinin sebebi budur. URL'den alınan bir
`next` için collage v0.44.0'dan beri kendi kontrolünü sunar:

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
yüzden render'a yalnızca key'in içerdikleri verilir: path, host, key'deki query
parametreleri, `collage.Vary` ile bildirilen header'lar. Bir okuyucunun cookie'leri,
credential'ları ya da adresi hiçbir zaman verilmez. Host key'de yer aldığı için
başka bir host adı veren bir request, diğer herkesin aldığı kopyayı zehirleyemez.
Degraded bir render, bir error page ve development'taki bir page hiçbir zaman public
olarak cache'lenemez. Ayrıntılar için
[Paylaşılan bir render neyi görür](/docs/caching#what-a-shared-render-sees)
bölümüne bakın.

## Body'ler sınırlıdır

Bir request body'si, middleware'iniz çalışmadan önce, yönlendirildiği action'ın
limitiyle sınırlanır. Aksini söylemediğiniz sürece bu limit 4 MiB'tır. Diske taşan
multipart dosyalar action cevap verdiğinde silinir. Forgery kontrolünün hiç
imzalamadığı bir cookie ise body daha hiç okunmadan reddedilir. Ayrıntılar için
[Request body'leri sınırlıdır](/docs/forms-and-actions#request-bodies-are-bounded)
bölümüne bakın.

## Bir dosya, uzantısının söylediği şeydir

Bir [asset mount](/docs/assets), bir dosyanın tipini uzantısına göre belirler. Adı
bir tip belirtmeyen dosyanın tipi içeriğinden çıkarılır; böylece uzantısız yüklenen
bir görsel yine görsel olarak gider. Ama hiçbir zaman çalıştırılabilen bir tipe
dönüşmez: HTML ya da XML'e benzeyen bir dosya `text/plain` olur ve `nosniff`
sayesinde tarayıcı buna uyar. Böylece markup içeren bir upload bir page olarak
sunulmaz. Mount hiçbir zaman bir dizini listelemez; `.well-known` dışında bir
dotfile'ı (`.env`, `.git/`) hiçbir zaman sunmaz ya da export etmez. `collage serve`
de aynısını yapar.

## Hatalar production'da hiçbir şey söylemez

Production'da collage'ın kendi error page'i status'tan başka bir şey göstermez: hata
metni yok, stack yok, path yok. Development overlay'i ve reload kanalı yalnızca
development'ta vardır. `collage dev` de yalnızca bu makineyi adlandıran bir `Host`'a
cevap verir. Bu yüzden başka bir sitedeki bir page, DNS rebinding yoluyla onları
okuyamaz. Bir log satırındaki request path'i control character içerdiğinde tırnak
içine alınır. Böylece sahte bir satır oluşturamaz ya da terminalinizi yeniden
yazamaz.

## Security header'ları

Her response varsayılan olarak `X-Content-Type-Options: nosniff` ve
`X-Frame-Options: SAMEORIGIN` taşır. Böylece bir site, hiçbir şey eklemeden
form'larında MIME-sniff ve clickjacking korumasına sahip olur. `Security.FrameOptions`
değeri belirler: `"-"` hiçbirini göndermez, başka her şey olduğu gibi gönderilir.
`Security.NoSniff` ise `false`'a işaret ettiğinde nosniff'i kapatır. Tam bir
`Content-Security-Policy`, HSTS ve geri kalanını
[elagoht/secure](/docs/plugins#elagohtsecure) plugin'i ekler ve onun header'ları
bunları geçersiz kılar. Production'daki bir page hiçbir framework script'i taşımadığı
için sıkı bir policy nonce gerektirmez.

## Size kalanlar

- **TLS**, `X-Forwarded-Proto`'yu ve tarayıcının `Host`'unu ileten bir proxy'de.
  Ayrıntılar için [Deployment](/docs/deployment#tls-behind-a-proxy) sayfasına bakın.
- **Rate limit'ler**: [elagoht/ratelimit](/docs/plugins#elagohtratelimit).
- **`app.Handle` ile mount edilmiş bir handler** tamamen sizindir: forgery kontrolü,
  body limiti ve cache yoktur.
- **Bir data handler'ın okudukları.** Cache'lenebilir bir page'de, middleware'in
  request context'ine koyduğu bir değer paylaşılan render'dan gizlenir, böylece bir
  okuyucunun değeri başka bir okuyucunun page'ine ulaşamaz. Bu değeri `collage.Vary`
  ile bildirin ve `collage.Varied` ile okuyun.

Bir güvenlik açığını bildirmek için
[collage repository'sinde](https://github.com/Elagoht/collage/security) GitHub'ın
private reporting özelliğini kullanın.
