---
description: Stil dosyalarını, script'leri, görselleri ve indirmeleri bir dizinden ya da gömülü bir dosya sisteminden, bir yıl önbellekte tutulabilen URL'lerle sunmak.
---

# Statik dosyalar

Stil dosyaları, script'ler, görseller, font'lar ve indirmeler bir **asset
mount**'u tarafından sunulur: bir URL önekinin altında, dosya olarak sunulan bir
dosya sistemi.

```go
root, err := os.OpenRoot("static")
if err != nil {
	log.Fatal(err)
}
if err := app.Mount("/static/", root.FS()); err != nil {
	log.Fatal(err)
}
```

`app.Mount` herhangi bir `fs.FS` alır; bu da önemli olan iki kaynağı kapsar:
diskteki bir dizin ve binary'nin içine derlenmiş bir `embed.FS`. `static/app.css`
artık `/static/app.css`'tir.

Mount'lar sayfa değildir. [Sayfa önbelleğine](/docs/caching) hiç girmezler,
etiketlenmezler ve `InvalidateTags` onlara ulaşmaz. Sayfalara göre boyutlandırılmış
bir önbellekteki 50 MB'lık bir video, binlerce sayfayı dışarı iterdi; bu yüzden
dosyalar dosya olarak sunulur.

## os.DirFS değil, os.OpenRoot kullanın

`os.DirFS`, bir dizini sunmanın akla gelen ilk yolu gibi görünür ve internet için
yanlış olanıdır. Kendi belgeleri, sembolik bağlantı üzerinden dizin dışına çıkmayı
engellemediğini söyler: dizinin içinde olup dışını gösteren bir sembolik bağlantı
izlenir ve gösterdiği şey her neyse sunulur.

`os.OpenRoot` ise işletim sistemi tarafından uygulanır. Açtığı her yol, tuttuğu
dizinin içinde çözümlenir ve dizinden çıkan bir sembolik bağlantı hiç açılamaz.
Go'da 1.24'ten beri vardır.

collage bu açığı sizin yerinize kapatamaz. Kendisine bir `fs.FS` verilir ve onun
üzerinde `Open`'ı çağırır; dosya sisteminin kendi dizini içinde kalıp kalmadığı,
dosya sisteminin bir özelliğidir. collage'ın yaptığı şey, `Open` çağrılmadan önce
URL'deki yolu `path.Clean` ile temizlemektir. Mount'un içinde kalan bir `..` ya da
`.` normalleştirilir ve sunulur — `/static/css/../app.css`, `/static/app.css`'tir —
mount'un dışına tırmanan bir yol ya da boş bir parçayla başlayan bir yol
(`/static//…`) ise `404` ile reddedilir. Bu, URL'de yazılmış bir dizin dışına çıkma
girişimini durdurur; diskteki sembolik bağlantılarla kurulmuş olanı değil.

`*os.Root`'u sürecin ömrü boyunca açık tutun: dosya sistemi her isteği sunar.

### Gömmek

Bir `embed.FS`'te sembolik bağlantı yoktur ve yapısı gereği güvenlidir. Ayrıca
binary'nin herhangi bir çalışma dizininden çalışmasını sağlar. Dizin adını çıkarmak
için `fs.Sub` kullanın; aksi hâlde stil dosyası `/static/static/app.css` adresinde
yanıt verir:

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

Geliştirmede ise bunun yerine diskteki dizini istersiniz — gömülü kopya binary
derlendiğinde sabitlenmiştir, dolayısıyla bir dosyayı düzenlemek hiçbir şeyi
değiştirmezdi. İskeleti oluşturulmuş bir proje bunu yapar:

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

collage bu seçimi şablonlar için kendisi yapar, çünkü `Template.Root` ona
şablonların diskte nerede olduğunu söyler. Bir mount'a ise yalnızca bir `fs.FS`
verilir; dosyalarının nereden geldiğini yalnızca sizin kodunuz bilir.

## İçerik adresli URL'ler

Mount edilmiş bir dosyaya yolunu yazarak değil, `asset` ile bağlantı verin:

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<script src="{{asset "/static/app.js"}}" defer></script>
```

Sayfaya giren, içinde içeriğinin hash'i bulunan dosya adıdır:

```html
<link rel="stylesheet" href="/static/app.0d5f2b53aebf6c72.css">
```

ve bu URL şununla sunulur:

```
Cache-Control: public, max-age=31536000, immutable
```

— bir yıl ve hiç değişmeyeceğine dair bir söz. Söz dürüsttür, çünkü ad baytlardan
gelir: dosya değiştiğinde ad değişir, sayfalar yenisine bağlantı verir ve eski URL
bir daha hiç istenmez. Tarayıcılar ve CDN'ler dosyayı bir yıl tutar ve hiç yeniden
doğrulamaz; kimseye de sayfayla uyuşmayan bir stil dosyası sunulmaz.

Düz bir ad hiçbir süreyle bu şekilde önbelleğe alınamaz. Er ya da geç birine
bugünün HTML'iyle dünün stil dosyasını verir.

Üç ayrıntı:

- **Var olmayan bir dosya bir URL değil, bir hatadır.** `{{asset
  "/static/typo.css"}}` fragment'i `collage.ErrUnknownAsset` ile başarısız kılar.
  Alternatifi, 404 veren bir stil dosyasına bağlantı verirken "başarıyla" render
  edilen bir sayfadır.
- **İstekteki hash denetlenir.** Hash'i dosyanın güncel hash'i olmayan bir URL,
  dosya değil 404 döner. Onu sunmak, paylaşılan bir önbelleğin arkasındaki herkesin
  önüne bir yıl boyunca yanlış bir yanıt çakmak olurdu.
- **Düz adlar yine çalışır.** `/static/app.css` de sunulur; mount'un kendi, daha
  kısa `Cache-Control`'üyle. Bir URL'nin aynı kalması gereken yerlerde kullanın —
  sitenin dışından başvurulan bir favicon, bir e-postadan bağlantı verilen bir dosya.

### Go'dan

URL'ye ihtiyaç duyan bir data handler — bir Open Graph görseli, bir preload ipucu
için — render context'e sorar:

```go
cover, err := rc.Asset("/static/covers/" + post.Slug + ".jpg")
if err != nil {
	return postView{}, nil, err
}
rc.HoistProperty("og:image", siteOrigin+cover)
```

`rc.Asset`, tam olarak `{{asset}}`'in render ettiğini döndürür; hiçbir mount'un
sunmadığı bir yol için ise `collage.ErrUnknownAsset` döner. v0.10.0'dan itibaren
bir [document'ın](/docs/documents#the-handler) handler'ında da çalışır — bir web
manifest'indeki bir ikon URL'si gibi. Stil dosyaları için `rc.HoistStylesheet` ve
`{{stylesheet}}`, aramayı ve `<link>`'i tek adımda yapar — bkz.
[Head ve SEO](/docs/head-and-seo).

## Seçenekler

```go
err := app.Mount("/media/", media.FS(),
	collage.WithCacheControl("public, max-age=86400"),
	collage.WithoutBuildCopy(),
)
```

| Seçenek | Etkisi |
| --- | --- |
| `collage.WithCacheControl(value)` | Düz adlarıyla istenen dosyalar için `Cache-Control`. Varsayılanı `public, max-age=3600` |
| `collage.WithoutBuildCopy()` | [Statik dışa aktarma](/docs/static-export) bu mount'u çıktısına kopyalamaz |

`WithCacheControl` yalnızca düz adlarla ilgilidir; içerik adresli URL'ler her zaman
bir yıl ve `immutable` ile sunulur, çünkü adları bunu hak eder. Düz bir adın ne
kadar tutulabileceği, altındaki dosyanın ne sıklıkla değiştiğine bağlıdır ve bunu
yalnızca siz bilirsiniz — bu yüzden ayarlamak size kalmıştır.

`WithoutBuildCopy`, production'ın başka bir yerden, örneğin bir CDN'den sunduğu ya
da her build'e kopyalanamayacak kadar büyük olan bir mount içindir.

## Bir dosya yanıtı neler içerir

Bir mount, `http.FileServer` değil, `fs.Open` ve Go'nun `http.ServeContent`'i
üzerinde ince bir katmandır:

| | |
| --- | --- |
| Metotlar | `GET` ve `HEAD`. Başka her şey `Allow: GET, HEAD` ile bir `405`'tir |
| `Content-Type` | Dosya uzantısından; olmazsa ilk 512 bayt koklanarak |
| `ETag` | Dosya içeriğinin güçlü bir hash'i; ilk istekte hesaplanır ve hatırlanır |
| `Range` | `If-Range` ve `206 Partial Content` ile desteklenir |
| Dizinler | Hiçbir zaman listelenmez. Bir dizin ya da çıplak önek bir 404'tür |
| `index.html` | Hiçbir zaman örtük olarak sunulmaz |
| Eksik bir dosya | Düz metin bir `404`, `no-store` — asla sizin HTML bulunamadı sayfanız değil |

**`Range` istekleri**, bir tarayıcının bir ses ya da video dosyasında baştan çekmeden
ileri sarmasını ya da bir indirmeye kaldığı yerden devam etmesini sağlar.
`http.ServeContent`'ten gelirler ve dosyaların, gövdeleri bellekte bütün olarak
üretilen [document'lardan](/docs/documents) ayrı bir mekanizma olmasının nedeni de
onlardır.

**ETag bir içerik hash'idir**, çünkü bir Go programında statik dosyaların en yaygın
kaynağı olan `embed.FS`, her dosya için sıfır değişiklik zamanı bildirir — dolayısıyla
`Last-Modified` ve boyut-tarih doğrulaması onun için hiç çalışmaz. Dosya değil,
yalnızca hash hatırlanır; bu yüzden harcadığı bellek dosyaların boyutuyla değil,
sayısıyla büyür.

Eksik bir stil dosyası bilerek sitenizin bulunamadı sayfasını değil, düz metin bir
404 alır: CSS isteyen bir tarayıcıya verilen bir HTML hata sayfası, hangi yönden
bakılsa bir hatadır.

## Önekler ve route'lar

Bir önek `/` ile başlamalı ve bitmelidir; tek başına `/` olamaz — kökteki bir mount
her sayfayı yutardı (`collage.ErrInvalidPrefix`). Nil bir dosya sistemi
`collage.ErrNilFS`'tir.

Bir mount'un öneki altındaki bir isteğe mount yanıt verir ve istek router'a hiç
ulaşmaz. Bu güvenlidir, çünkü başlangıç, bir sayfanın, document'ın ya da
yönlendirmenin zaten yanıt verdiği bir URL'yi gizleyecek bir öneki reddeder —
locale önekli olanlar dahil; dolayısıyla Türkçe sayfaları olan bir sitede `/tr/`
adresindeki bir mount başarısız olur (`collage.ErrMountShadowsRoute`). Önekleri
örtüşen iki mount da reddedilir (`collage.ErrMountConflict`). İki denetim de kayıt
kapandığında çalışır; bu yüzden `Mount` ve `RegisterPage`'i hangi sırayla
çağırdığınızın önemi yoktur.

Denetim önekleri kayıtlı pattern'lerle karşılaştırır; bu yüzden önekin üstündeki
her şeyi yakalayan bir pattern'i göremez: `/{rest...}` adresindeki bir sayfa
`/static/…` ile de eşleşirdi ve bu URL'leri mount alır. Bir öneki sahiplenmenin
bedeli budur.

## Geliştirmede

İçerik adresli bir URL ile düzenlemekte olduğunuz bir dosya birbiriyle çelişir.
Geliştirme dışında hash bir kez hesaplanır ve hatırlanır, URL de bir yıl için söz
verilir; çalışan bir sürecin altında dosyayı düzenlerseniz sayfa, tarayıcıya hiç
değişmeyeceği söylenmiş olan eski ada bağlantı vermeye devam ederdi.

Bu yüzden `DevMode` açıkken her mount, bir dosyanın hash'ini her istekte yeniden
hesaplar ve her şeyi `no-store` ile sunar. Düzenlenen bir dosya yeni bir URL alır,
sayfa ona bağlantı verir ve tarayıcı onu çeker — geliştirmedeki yenileme script'i de
mount edilmiş bir dosya değiştiğinde sayfayı yeniler. (Handler'larınızın okuduğu
Markdown gibi mount edilmemiş bir dizin,
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtildiğinde sayfayı
yeniler.) Geliştirme dışında, diskte değişen bir dosya, süreç yeniden başlatılana
kadar hatırlanan hash'ini korur: mount'lar, çalışan bir sunucunun altında
düzenlenen dosyalar için değil, yayına alınan dosyalar içindir.

## Statik dışa aktarma

Bir [statik dışa aktarma](/docs/static-export), her mount'u kendi önekinin altında
çıktısına kopyalar — `/static/app.css`, `dist/static/app.css` olur — ve bir sayfanın
`{{asset}}` ile gerçekten bağlantı verdiği her dosyanın içerik adresli bir
kopyasını da ekler. Yalnızca bunların: bir medya dizini, hiçbir sayfanın
kullanmadığı adlar için ikiye katlanmaz. `collage.WithoutBuildCopy()` bir mount'u
dışarıda bırakır.

## Mount'ların yapmadıkları

- **Sıkıştırma yok.** Gzip ya da Brotli yok, önceden sıkıştırılmış yan dosyalar da
  yok. Önüne bir reverse proxy ya da CDN koyun veya `app.Handler()`'ı sarın.
- **Paketleme, küçültme ya da görsel işleme yok.** Bir mount kendisine verilen
  baytları sunar; [plugin'ler](/docs/plugins) daha fazlasını yapabilir.
- **Yalnızca `fs.FS` kaynakları.** S3 gibi bir nesne deposu bunlardan biri değildir.
  Onu [`app.Handle`](/docs/middleware-and-apis) ile kendiniz sunun ya da doğrudan
  ona bağlantı verin.
