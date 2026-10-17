---
description: Stylesheet'leri, script'leri, görselleri ve indirilecek dosyaları bir dizinden ya da gömülü bir dosya sisteminden, bir yıl cache'lenebilen URL'lerle sunmak.
reference: App.Mount, MountOption, WithCacheControl, WithoutBuildCopy, ErrUnknownAsset
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
kalmaması, dosya sisteminin bir özelliğidir.

collage'ın yaptığı şey şudur: `Open` çağrılmadan önce URL'deki path'i `path.Clean`
ile temizler:

- Mount'un içinde kalan bir `..` ya da `.` normalize edilir ve dosya sunulur;
  örneğin `/static/css/../app.css`, `/static/app.css` olarak ele alınır.
- Mount'un dışına çıkan bir path ya da boş bir elemanla başlayan bir path
  (`/static//…`) ise `404` ile reddedilir.

Bu kontrol URL'ye yazılmış bir traversal'ı durdurur, ama diskteki symlink'lerle
kurulmuş bir traversal'ı durdurmaz.

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
verir.

Söz dürüsttür, çünkü ad byte'lardan türetilir. Dosya değiştiğinde ad da
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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 264" role="img" aria-labelledby="assets-mount-title">
<title id="assets-mount-title">Kodunuz app.Mount'a bir dosya sistemi verir: binary'ye derlenmiş bir embed.FS ya da development'ta os.OpenRoot ile açılmış, diskteki static dizini. Mount ardından /static/ altında üç tür ada cevap verir. asset fonksiyonunun linklediği hash'li ad app.hash.css, bir yıl cache'lenen ve immutable olan bir 200'dür. Düz ad app.css, mount'un kendi max-age=3600 değeriyle bir 200'dür. Geri kalan her şey, yani eski ya da yanlış bir hash veya olmayan bir dosya, no-store ile düz metin bir 404'tür. DevMode açıkken hash her request'te yeniden hesaplanır ve her 200 no-store olur.</title>
<defs><marker id="assets-mount-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="6" y="8" width="150" height="40" rx="6" /><text x="81" y="24" text-anchor="middle">embed.FS</text><text class="d-small" x="81" y="39" text-anchor="middle">binary'nin içinde</text>
<rect class="d-frame" x="164" y="8" width="150" height="40" rx="6" /><text x="239" y="24" text-anchor="middle">os.OpenRoot</text><text class="d-small" x="239" y="39" text-anchor="middle">dev: diskteki static/</text>
<path class="d-flow" d="M81 48C81 62 130 60 130 72" marker-end="url(#assets-mount-arrow)" />
<path class="d-flow" d="M239 48C239 62 190 60 190 72" marker-end="url(#assets-mount-arrow)" />
<rect class="d-a" x="50" y="76" width="220" height="26" rx="5" /><text class="d-on-a" x="160" y="93" text-anchor="middle">app.Mount("/static/", fsys)</text>
<path class="d-flow" d="M100 102C100 114 56 110 56 122" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="160" y1="102" x2="160" y2="122" marker-end="url(#assets-mount-arrow)" />
<path class="d-flow" d="M220 102C220 114 264 110 264 122" marker-end="url(#assets-mount-arrow)" />
<rect class="d-frame" x="6" y="126" width="100" height="48" rx="6" /><text x="56" y="141" text-anchor="middle">hash'li ad</text><text class="d-small" x="56" y="155" text-anchor="middle">app.&lt;hash&gt;.css</text><text class="d-small" x="56" y="167" text-anchor="middle">asset'in linki</text>
<rect class="d-frame" x="110" y="126" width="100" height="48" rx="6" /><text x="160" y="141" text-anchor="middle">düz ad</text><text class="d-small" x="160" y="155" text-anchor="middle">app.css</text><text class="d-small" x="160" y="167" text-anchor="middle">sabit bir URL</text>
<rect class="d-frame" x="214" y="126" width="100" height="48" rx="6" /><text x="264" y="141" text-anchor="middle">geri kalan</text><text class="d-small" x="264" y="155" text-anchor="middle">eski bir hash,</text><text class="d-small" x="264" y="167" text-anchor="middle">olmayan dosya</text>
<line class="d-flow" x1="56" y1="174" x2="56" y2="190" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="160" y1="174" x2="160" y2="190" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="264" y1="174" x2="264" y2="190" marker-end="url(#assets-mount-arrow)" />
<rect class="d-b" x="6" y="194" width="100" height="40" rx="5" /><text class="d-on-b" x="56" y="210" text-anchor="middle">200 immutable</text><text class="d-on-b d-small" x="56" y="225" text-anchor="middle">max-age=31536000</text>
<rect class="d-soft-a" x="110" y="194" width="100" height="40" rx="5" /><text x="160" y="210" text-anchor="middle">200</text><text class="d-small" x="160" y="225" text-anchor="middle">max-age=3600</text>
<g class="d-dropped"><rect class="d-frame" x="214" y="194" width="100" height="40" rx="5" /><text x="264" y="210" text-anchor="middle">404 no-store</text><text class="d-small" x="264" y="225" text-anchor="middle">düz metin</text></g>
<text class="d-small" x="160" y="254" text-anchor="middle">DevMode: her request'te hash, her 200 no-store</text>
</svg>
<figcaption>Mount'un hangi dosya sistemini sunacağını kodunuz seçer: production'da gömülü kopyayı, development'ta diskteki dizini. Prefix altında mount üç tür ada cevap verir. <code>{{asset}}</code>'in linklediği hash'li ad bir yıl boyunca, <code>immutable</code> ile sunulur. Düz ad, mount'un kendi <code>Cache-Control</code> değeriyle sunulur; <code>WithCacheControl</code> başka bir değer vermedikçe bu <code>public, max-age=3600</code>'dür. Dosyanın güncel hash'i olmayan bir hash ya da var olmayan bir dosya, <code>no-store</code> ile düz metin bir 404'tür. <code>DevMode</code> açıkken hash her request'te yeniden hesaplanır ve hiçbir şey cache'lenebilir değildir.</figcaption>
</figure>

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
| `Content-Type` | Dosya uzantısından belirlenir. Adı bir tip belirtmeyen dosya ilk 512 byte'ından sniff edilir, ama hiçbir zaman çalıştırılabilen bir tipe dönüşmez: HTML ya da XML'e benzeyen bir dosya `text/plain` olur (v0.34.2'den beri). Her zaman `X-Content-Type-Options: nosniff` ile gider |
| `ETag` | Dosya içeriğinin strong hash'idir. İlk request'te hesaplanır ve hatırlanır |
| `Range` | `If-Range` ve `206 Partial Content` ile desteklenir |
| Dizinler | Hiçbir zaman listelenmez. Bir dizin ya da yalnızca prefix 404 döner |
| `index.html` | Hiçbir zaman örtük olarak sunulmaz |
| Dotfile'lar | Hiçbir zaman sunulmaz ve export tarafından da kopyalanmaz. Bir elemanı `.` ile başlayan path (`.env`, `.git/`, bir editörün swap dosyası) 404 döner. Tek istisna `.well-known`'dur (v0.34.0'dan beri) |
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
page'i yeniler.)

Development dışında ise diskte değişen bir dosya, process yeniden
başlatılana kadar hatırlanan hash'ini korur. Mount'lar, çalışan bir sunucunun
altında düzenlenen dosyalar için değil, deploy edilen dosyalar içindir.

## Static export

[Static export](/docs/static-export), her mount'u kendi prefix'i altında çıktıya
kopyalar. Örneğin `/static/app.css`, `dist/static/app.css` olur. Buna ek olarak, bir
page'in `{{asset}}` ile gerçekten link verdiği her dosyanın content-addressed bir
kopyasını da çıktıya ekler. Yalnızca bu dosyaların kopyası eklenir. Böylece bir
medya dizini, hiçbir page'in kullanmadığı adlar için ikiye katlanmaz.
`collage.WithoutBuildCopy()` bir mount'u bu kopyalamanın dışında bırakır.

## Örnek: bir blogun stylesheet'i, script'i ve görselleri

[Template'ler](/docs/templates#example-a-blog-post) sayfasındaki blogun sunacağı
üç tür dosya vardır:

| Dosya | Nereden link verilir | Ne zaman değişir |
| --- | --- | --- |
| `static/app.css` | Layout'tan, her page'de | Bir yeniden tasarımla |
| `static/app.js` | Layout'tan, her page'de | Nadiren |
| `static/img/harbour.jpg` | Bir yazının template'inden | Yayımlandıktan sonra hiç |

### Mount etmek

Dosyalar embed edilir, böylece binary herhangi bir dizinden çalışır. `collage dev`
ise bunun yerine diskteki dizini sunar. `staticFiles`,
[Embed etmek](#embedding) bölümündeki fonksiyondur:

```go
//go:embed all:static
var staticFS embed.FS

func main() {
	devMode := os.Getenv("COLLAGE_DEV") == "1"

	app, err := collage.New(&collage.Config{
		DevMode:  devMode,
		Template: collage.TemplateConfig{Root: "templates"},
	})
	if err != nil {
		log.Fatal(err)
	}

	static, err := staticFiles(devMode)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Mount("/static/", static); err != nil {
		log.Fatal(err)
	}

	// register pages, then serve
}
```

### Link vermek

Layout stylesheet'e ve script'e, yazı da kendi görseline link verir. Her biri düz
path'iyle yazılır:

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<script src="{{asset "/static/app.js"}}" defer></script>
```

```html
<img src="{{asset "/static/img/harbour.jpg"}}" alt="The harbour at dawn">
```

Okurun tarayıcısına ulaşan HTML'de her dosyanın hash'i bulunur:

```html
<link rel="stylesheet" href="/static/app.b988912693d9423d.css">
<script src="/static/app.dc18125931ff948b.js" defer></script>
<img src="/static/img/harbour.9e00fefaa8d22370.jpg" alt="The harbour at dawn">
```

Tarayıcı ardından bu adları ister ve mount şöyle cevap verir:

| Request | Status | `Cache-Control` |
| --- | --- | --- |
| `/static/app.b988912693d9423d.css` | `200` | `public, max-age=31536000, immutable` |
| `/static/app.css` | `200` | `public, max-age=3600` |
| `/static/app.0000000000000000.css` | `404`, düz metin | `no-store` |
| `POST /static/app.css` | `405`, `Allow: GET, HEAD` | `no-store` |

`collage dev` altında da aynı adlara link verilir ve aynı adlar sunulur, ama her
`200` `no-store` ile gider.

### Değişen bir stylesheet'i deploy etmek

Bir yeniden tasarım `static/app.css`'i değiştirir ve yeni byte'ları gömülü yeni bir
binary deploy edilir. Ardından:

1. Yeni process'in render ettiği ilk page yeni dosyanın hash'ini hesaplar ve
   `/static/app.0a044e6f01a45071.css` adresine link verir.
2. Okurun tarayıcısı bu adı hiç görmemiştir. Stylesheet'i bir kez çeker ve bir yıl
   boyunca tutar.
3. Eski `app.b988912693d9423d.css` o tarayıcının cache'inde kullanılmadan kalır,
   çünkü artık hiçbir page ona link vermez.
4. Eski ad istendiğinde yeni process `404`, `no-store` ile cevap verir. Ne artık
   elinde olmayan eski byte'ları sunar ne de yeni byte'ları, onları tarif etmeyen
   bir adın altında.

Hiçbir şeyin purge edilmesi gerekmedi ve hiçbir okura yeni HTML eski stylesheet'le
birlikte gösterilmedi.

Bu güvence düz adda biter. Deploy'dan kısa süre önce `/static/app.css`'i çekmiş
bir okura, çektiği andan itibaren bir saat dolana kadar eski kopyası verilebilir.
Page'lerin `{{asset}}` ile link vermesinin ve düz adın sizin kontrolünüz dışındaki
URL'ler için ayrılmasının nedeni budur.

### Hiçbir mount'un sunmadığı bir path

Yazının template'inde bir yazım hatası:

```html
<img src="{{asset "/static/img/hrabour.jpg"}}" alt="The harbour at dawn">
```

Böyle bir dosya yoktur, bu yüzden `{{asset}}` yazı fragment'ini
`collage.ErrUnknownAsset` ile başarısız kılar. Log, fragment'i, template'in
satırını ve path'i gösterir. Okurun ne alacağını fragment'in
[failure policy'si](/docs/fragments-and-slots#when-a-fragment-fails) belirler:

- **Yazı `Required()` olduğu için** page başarısız olur: bir `500` ve error page
  döner, kırık görselli bir yazı asla dönmez.
- **Opsiyonel bir fragment'te**, örneğin bir sidebar'da, page o fragment olmadan
  sunulur, `200` döner, `no-store` ile gider ve cache'lenmez. Hata bir warning
  olarak log'lanır.

Her iki durumda da hata, okurun network panelinde bir `404` olarak değil, ona
ulaşan ilk render'da görünür. Bu render genellikle development'taki sizin
render'ınızdır.

Bir request ise başka bir konudur. `/static/` altında hiçbir dosyayı adlandırmayan
bir URL, mount'tan düz metin bir `404` alır. Hiçbir mount'un prefix'i altına
düşmeyen bir URL ise hiçbir mount'a ulaşmaz: ona router, sitenin kendi not-found
page'iyle cevap verir.

## Mount'ların yapmadıkları

- **Sıkıştırma yapmazlar.** Gzip ya da Brotli yoktur, önceden sıkıştırılmış
  sidecar dosyalar da desteklenmez. Önüne bir reverse proxy ya da CDN koyun,
  `app.Handler()`'ı wrap edin ya da [elagoht/compress](/docs/plugins#elagohtcompress)'i
  register edin.
- **Bundling, minification ya da görsel işleme yapmazlar.** Bir mount kendisine
  verilen byte'ları olduğu gibi sunar. Daha fazlası için [plugin'ler](/docs/plugins)
  kullanılabilir: [elagoht/bundle](/docs/plugins#elagohtbundle) script'leri ve
  stylesheet'leri esbuild ile bundle eder,
  [elagoht/minimizer](/docs/plugins#elagohtminimizer) bir mount'un sunduklarını
  minify eder, [elagoht/opti-image](/docs/plugins#elagohtopti-image) görselleri
  yeniden boyutlandırır, [elagoht/favicon](/docs/plugins#elagohtfavicon) bir sitenin
  ikonlarını tek bir görselden üretir, [elagoht/offline](/docs/plugins#elagohtoffline)
  ise page'leri ve static dosyaları ağ olmadan okumak için saklar.
- **Yalnızca `fs.FS` kaynaklarıyla çalışırlar.** S3 gibi bir object store bir
  `fs.FS` değildir. Onu [`app.Handle`](/docs/middleware-and-apis) ile kendiniz
  sunun ya da doğrudan ona link verin.
