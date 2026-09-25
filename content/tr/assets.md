---
description: Stylesheet'leri, script'leri, görselleri ve indirilecek dosyaları bir dizinden ya da gömülü bir dosya sisteminden, bir yıl boyunca cache'lenebilen URL'lerle sunmak.
reference: Mount, MountOption, WithCacheControl, WithoutBuildCopy, ErrUnknownAsset
---

# Static asset'ler

Stylesheet'ler, script'ler, görseller, font'lar ve indirilecek dosyalar bir **asset
mount** tarafından sunulur. Asset mount, bir URL prefix'i altında sunulan bir dosya
sistemidir ve içindekileri dosya olarak sunar.

```go
root, err := os.OpenRoot("static")
if err != nil {
	log.Fatal(err)
}
if err := app.Mount("/static/", root.FS()); err != nil {
	log.Fatal(err)
}
```

`app.Mount` herhangi bir `fs.FS` kabul eder. Böylece önemli olan iki kaynağın ikisi
de kapsanır: diskteki bir dizin ve binary'nin içine derlenmiş bir `embed.FS`.
Artık `static/app.css` dosyası `/static/app.css` adresinden sunulur.

Mount'lar page değildir. [Page cache](/docs/caching)'e hiç girmezler, tag
almazlar ve `InvalidateTags` onlara ulaşmaz. Page'lere göre boyutlandırılmış bir
cache'e 50 MB'lık bir video girseydi, binlerce page'i dışarı iterdi. Bu yüzden
dosyalar dosya olarak sunulur.

## os.DirFS değil, os.OpenRoot kullanın

Bir dizini sunmanın akla gelen ilk yolu `os.DirFS` gibi görünür, ama internete açık
bir sunucu için yanlış yoldur. Kendi dokümantasyonu, symlink traversal'ı
engellemediğini söyler. Dizinin içinde olup dışını gösteren bir symlink takip
edilir ve gösterdiği şey her neyse sunulur.

`os.OpenRoot` ise bu sınırı işletim sistemi düzeyinde uygular. Açtığı her path,
tuttuğu dizinin içinde çözümlenir. Dizinin dışına çıkan bir symlink ise hiç
açılamaz. `os.OpenRoot`, Go 1.24'ten beri vardır.

collage bu açığı sizin yerinize kapatamaz. collage'a bir `fs.FS` verilir ve collage
onun üzerinde `Open`'ı çağırır. Dosya sisteminin kendi dizini içinde kalıp
kalmaması, dosya sisteminin bir özelliğidir. collage'ın yaptığı şey şudur: `Open`
çağrılmadan önce URL'deki path'i `path.Clean` ile temizler. Mount'un içinde kalan
bir `..` ya da `.` normalize edilir ve dosya sunulur; örneğin
`/static/css/../app.css`, `/static/app.css` olarak ele alınır. Mount'un dışına
çıkan bir path ya da boş bir elemanla başlayan bir path (`/static//…`) ise `404`
ile reddedilir. Bu kontrol URL'ye yazılmış bir traversal'ı durdurur, ama diskteki
symlink'lerle kurulmuş bir traversal'ı durdurmaz.

`*os.Root`'u process boyunca açık tutun, çünkü onun dosya sistemi her request'i
sunar.

### Embed etmek

Bir `embed.FS` içinde symlink yoktur, dolayısıyla yapısı gereği güvenlidir. Ayrıca
binary'nin herhangi bir çalışma dizininden çalışabilmesini sağlar. Dizin adını
çıkarmak için `fs.Sub` kullanın. Aksi hâlde stylesheet `/static/static/app.css`
adresinden sunulur:

```go
//go:embed all:static
var staticFS embed.FS

func mountAssets(app *collage.App) error {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	return app.Mount("/static/", static)
}
```

Development'ta ise diskteki dizini kullanmak istersiniz. Gömülü kopya
binary build edildiği anda sabitlenmiştir, bu yüzden bir dosyayı düzenlemek hiçbir
şeyi değiştirmez. Scaffold edilmiş bir proje bunu şöyle yapar:

```go
func staticFiles(devMode bool) (fs.FS, error) {
	if devMode {
		if root, err := os.OpenRoot("static"); err == nil {
			return root.FS(), nil
		}
	}
	return fs.Sub(staticFS, "static")
}
```

collage bu seçimi template'ler için kendisi yapar, çünkü `Template.Root`
template'lerin diskte nerede olduğunu ona söyler. Bir mount'a ise yalnızca bir
`fs.FS` verilir. Dosyalarının nereden geldiğini yalnızca sizin kodunuz bilir.

## Content-addressed URL'ler

Mount edilmiş bir dosyaya link verirken path'ini elle yazmayın, `asset` kullanın:

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<script src="{{asset "/static/app.js"}}" defer></script>
```

Page'e yazılan şey, içinde dosya içeriğinin hash'i bulunan bir dosya adıdır:

```html
<link rel="stylesheet" href="/static/app.0d5f2b53aebf6c72.css">
```

Bu URL şu header ile sunulur:

```
Cache-Control: public, max-age=31536000, immutable
```

Bu header bir yıl anlamına gelir ve dosyanın hiç değişmeyeceğine dair bir söz
verir. Söz dürüsttür, çünkü ad byte'lardan türetilir. Dosya değiştiğinde ad da
değişir, page'ler yeni ada link verir ve eski URL bir daha hiç istenmez.
Tarayıcılar ve CDN'ler dosyayı bir yıl boyunca tutar ve hiç revalidate etmez.
Kimseye de page ile uyuşmayan bir stylesheet sunulmaz.

Düz bir ad ise hangi süreyle olursa olsun bu şekilde cache'lenemez. Er ya da geç
birine bugünün HTML'iyle birlikte dünün stylesheet'ini verir.

Üç ayrıntı:

- **Var olmayan bir dosya URL üretmez, hata üretir.** `{{asset
  "/static/typo.css"}}`, fragment'in `collage.ErrUnknownAsset` ile başarısız olmasına
  yol açar. Aksi hâlde page "başarıyla" render edilir, ama 404 dönen bir
  stylesheet'e link verir.
- **Request'teki hash kontrol edilir.** Hash'i dosyanın güncel hash'iyle
  eşleşmeyen bir URL, dosyayı değil 404 döner. O dosyayı sunmak, shared cache'in
  arkasındaki herkese bir yıl boyunca yanlış bir cevabı sabitlemek olurdu.
- **Düz adlar da çalışır.** `/static/app.css` de sunulur, ama mount'un kendi, daha
  kısa `Cache-Control` değeriyle. URL'nin sabit kalması gereken yerlerde bunu
  kullanın. Örneğin site dışından referans verilen bir favicon ya da bir
  e-postadan link verilen bir dosya için.

### Go'dan

URL'ye ihtiyaç duyan bir data handler, örneğin bir Open Graph görseli ya da bir
preload hint'i için, URL'yi render context'ten ister:

```go
cover, err := rc.Asset("/static/covers/" + post.Slug + ".jpg")
if err != nil {
	return postView{}, nil, err
}
rc.HoistProperty("og:image", siteOrigin+cover)
```

`rc.Asset`, `{{asset}}`'in render ettiği değerin aynısını döner. Hiçbir mount'un
sunmadığı bir path için `collage.ErrUnknownAsset` döner. v0.10.0'dan beri bir
[document'ın](/docs/documents#the-handler) handler'ında da çalışır; örneğin bir web
manifest'indeki icon URL'si için. Stylesheet'ler için `rc.HoistStylesheet` ve
`{{stylesheet}}`, lookup'ı ve `<link>`'i tek adımda yapar. Ayrıntılar için
[Head ve SEO](/docs/head-and-seo) sayfasına bakın.

## Seçenekler

```go
err := app.Mount("/media/", media.FS(),
	collage.WithCacheControl("public, max-age=86400"),
	collage.WithoutBuildCopy(),
)
```

| Seçenek | Etkisi |
| --- | --- |
| `collage.WithCacheControl(value)` | Düz adlarıyla istenen dosyalar için `Cache-Control` değeri. Varsayılanı `public, max-age=3600` |
| `collage.WithoutBuildCopy()` | [Static export](/docs/static-export) bu mount'u çıktısına kopyalamaz |

`WithCacheControl` yalnızca düz adları etkiler. Content-addressed URL'ler her zaman
bir yıl ve `immutable` ile sunulur, çünkü adları bunu hak eder. Düz bir adın ne
kadar süre tutulabileceği, o adın altındaki dosyanın ne sıklıkla değiştiğine
bağlıdır. Bunu yalnızca siz bilirsiniz, bu yüzden ayarlamak da size kalır.

`WithoutBuildCopy`, production'da başka bir yerden (örneğin bir CDN'den) sunulan
ya da her build'e kopyalanamayacak kadar büyük olan bir mount içindir.

## Bir dosya response'u neleri içerir

Bir mount, `http.FileServer` üzerine değil, `fs.Open` ve Go'nun
`http.ServeContent`'i üzerine kurulmuş ince bir katmandır:

| | |
| --- | --- |
| Method'lar | `GET` ve `HEAD`. Diğer her method `Allow: GET, HEAD` ile birlikte `405` alır |
| `Content-Type` | Dosya uzantısından belirlenir. Belirlenemezse ilk 512 byte sniff edilir |
| `ETag` | Dosya içeriğinin strong hash'idir. İlk request'te hesaplanır ve hatırlanır |
| `Range` | `If-Range` ve `206 Partial Content` ile desteklenir |
| Dizinler | Hiçbir zaman listelenmez. Bir dizin ya da yalnızca prefix 404 döner |
| `index.html` | Hiçbir zaman örtük olarak sunulmaz |
| Olmayan bir dosya | Düz metin bir `404`, `no-store` ile döner. Sizin HTML not-found page'iniz asla dönmez |

**`Range` request'leri**, tarayıcının bir ses ya da video dosyasında dosyayı baştan
indirmeden ileri sarabilmesini ya da bir indirmeye kaldığı yerden devam
edebilmesini sağlar. Bu destek `http.ServeContent`'ten gelir. Dosyaların
[document'lardan](/docs/documents) ayrı bir mekanizma olmasının nedeni de budur,
çünkü document'ların body'si bellekte bir bütün olarak üretilir.

**ETag bir içerik hash'idir.** Bunun nedeni, bir Go programında asset'lerin en
yaygın kaynağı olan `embed.FS`'in her dosya için sıfır modification time
bildirmesidir. Bu yüzden `Last-Modified` ve boyut ile tarihe dayalı validation
onun için hiç çalışmaz. Dosyanın kendisi değil, yalnızca hash'i hatırlanır. Bu
yüzden harcanan bellek dosyaların boyutuyla değil, sayısıyla artar.

Olmayan bir stylesheet için sitenizin not-found page'i yerine düz metin bir 404
dönmesi bilinçli bir tercihtir. CSS isteyen bir tarayıcıya HTML bir error page
vermek, hangi taraftan bakılırsa bakılsın bir hatadır.

## Prefix'ler ve route'lar

Bir prefix `/` ile başlamalı ve `/` ile bitmelidir. Prefix tek başına `/` olamaz,
çünkü kökteki bir mount bütün page'leri yutardı (`collage.ErrInvalidPrefix`). Nil
bir dosya sistemi `collage.ErrNilFS` hatasını verir.

Bir mount'un prefix'i altına düşen bir request'e mount cevap verir ve request
router'a hiç ulaşmaz. Bu güvenlidir, çünkü uygulama başlarken bir page'in,
document'ın ya da redirect'in zaten cevap verdiği bir URL'yi gölgeleyecek bir
prefix reddedilir. Locale prefix'li URL'ler de buna dahildir. Örneğin Türkçe
page'leri olan bir sitede `/tr/` altındaki bir mount başarısız olur
(`collage.ErrMountShadowsRoute`). Prefix'leri çakışan iki mount da reddedilir
(`collage.ErrMountConflict`). İki kontrol de register aşaması kapandığında çalışır. Bu
yüzden `Mount` ve `RegisterPage`'i hangi sırayla çağırdığınızın bir önemi yoktur.

Bu kontrol prefix'leri register edilmiş pattern'lerle karşılaştırır. Bu yüzden prefix'in
üst seviyesindeki bir catch-all'u göremez. Örneğin `/{rest...}` adresindeki bir
page `/static/…` ile de eşleşirdi, ama bu URL'leri mount alır. Bir prefix'i
sahiplenmenin bedeli budur.

## Development'ta

Content-addressed bir URL ile düzenlemekte olduğunuz bir dosya birbiriyle çelişir.
Development dışında hash bir kez hesaplanır ve hatırlanır, URL için de bir yıllık
söz verilir. Çalışan bir process'in altında dosyayı düzenlerseniz page eski ada
link vermeye devam ederdi. Oysa tarayıcıya o adın hiç değişmeyeceği söylenmiştir.

Bu yüzden `DevMode` açıkken her mount, bir dosyanın hash'ini her request'te
yeniden hesaplar ve her şeyi `no-store` ile sunar. Düzenlenen bir dosya yeni bir URL
alır, page bu URL'ye link verir ve tarayıcı dosyayı yeniden çeker. Ayrıca
development'taki reload script'i, mount edilmiş bir dosya değiştiğinde
page'i yeniler. (Handler'larınızın okuduğu Markdown dosyaları gibi mount edilmemiş
bir dizin, [`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtildiğinde
page'i yeniler.) Development dışında ise diskte değişen bir dosya, process yeniden
başlatılana kadar hatırlanan hash'ini korur. Mount'lar, çalışan bir sunucunun
altında düzenlenen dosyalar için değil, deploy edilen dosyalar içindir.

## Static export

[Static export](/docs/static-export), her mount'u kendi prefix'i altında çıktıya
kopyalar. Örneğin `/static/app.css`, `dist/static/app.css` olur. Buna ek olarak, bir
page'in `{{asset}}` ile gerçekten link verdiği her dosyanın content-addressed bir
kopyasını da çıktıya ekler. Yalnızca bu dosyaların kopyası eklenir. Böylece bir
medya dizini, hiçbir page'in kullanmadığı adlar için ikiye katlanmaz.
`collage.WithoutBuildCopy()` bir mount'u bu kopyalamanın dışında bırakır.

## Mount'ların yapmadıkları

- **Sıkıştırma yapmazlar.** Gzip ya da Brotli yoktur, önceden sıkıştırılmış
  sidecar dosyalar da desteklenmez. Önüne bir reverse proxy ya da CDN koyun veya
  `app.Handler()`'ı wrap edin.
- **Bundling, minification ya da görsel işleme yapmazlar.** Bir mount kendisine
  verilen byte'ları olduğu gibi sunar. Daha fazlası için [plugin'ler](/docs/plugins)
  kullanılabilir.
- **Yalnızca `fs.FS` kaynaklarıyla çalışırlar.** S3 gibi bir object store bir
  `fs.FS` değildir. Onu [`app.Handle`](/docs/middleware-and-apis) ile kendiniz
  sunun ya da doğrudan ona link verin.
