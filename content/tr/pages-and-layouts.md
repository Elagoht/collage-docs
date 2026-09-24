---
description: Sayfa nedir, layout'u ve içeriği nasıl bir araya gelir, ona hangi yollar ulaşır, nasıl önbelleğe alınır, başarısız olduğunda ne gösterir ve kayıt ona ne yapar.
---

# Sayfalar ve layout'lar

Bir **sayfa**, adı olan bir render yapılandırmasıdır. Sayfayı hangi fragment'lerin
oluşturduğunu — bir layout ve içindeki içerik —, ona hangi URL'lerin ulaştığını,
çıktısının nasıl önbelleğe alındığını ve render edilemediğinde onun yerine hangi
sayfanın gösterileceğini söyler. Kendine ait bir şablonu ya da verisi yoktur; bunlar
fragment'lerine aittir.

```go
page := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()

if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

## Sayfa kurmak

`collage.NewPage(name)` bir builder başlatır, her `WithX` çağrısı tek bir şeyi
ayarlar ve `Build()` `*collage.Page`'i döndürür.

Ad, sayfanın kimliğidir. Bağlantılar ondan kurulur
(`{{pageURL "blog-post" "slug" .Slug}}`), testler ve plugin'ler sayfaları onunla
arar ve iki sayfa aynı adı paylaşamaz. Adı sayfanın nerede durduğuna göre değil, ne
olduğuna göre seçin; böylece URL değiştiğinde de geçerli kalır.

Builder'lar bir hata döndürmek için zinciri asla kesmez. İsteneni yapamayan bir
çağrı hatayı not eder ve devam eder; `BuildErr()` not edilen her şeyi döndürür:

```go
builder := collage.NewPage("blog-post").WithLayout(layout).WithPath("en", "/blog/{slug}")
page := builder.Build()
if err := builder.BuildErr(); err != nil {
	return err // collage.ErrMissingContent: there is no WithContent
}
```

`BuildErr`'ü görmezden gelmek bir hatanın geçip gitmesine izin vermez. Bir builder'ın
not ettiği şey kurduğu değerin üzerinde kalır ve `RegisterPage`, böyle bir hata
taşıyan sayfayı — hata ister sayfanın kendisine, ister ağacındaki herhangi bir
fragment'e ait olsun, örneğin iki kez bildirilmiş bir slot — sayfanın adını veren bir
hatayla reddeder; birinin `BuildErr`'ü çağırıp çağırmadığından bağımsız olarak.
`BuildErr`'ü, kontrolünüzde olmayan bir girdiden kurulan ve hatayı nedenine daha
yakın bir yerde görmek istediğiniz sayfalarda denetleyin.

## Layout ve içerik

Bir sitedeki çoğu sayfa dış kısmını — `<head>`, üst bilgi, alt bilgi — paylaşır ve
içeride farklılaşır. Dış kısım **layout**'tur: `content` adında bir slot'u olan bir
fragment. İç kısım ise **içerik fragment'i**dir: sayfanın göstermek için var olduğu
fragment.

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithSlot(collage.DefaultContentSlot, true, false). // "content": required, one fragment
	Build()
```

```html
<!-- templates/layouts/default.html -->
<!doctype html>
<html lang="en">
<head>{{hoist "head"}}</head>
<body>
  <header>…</header>
  {{slot "content"}}
  <footer>…</footer>
</body>
</html>
```

`WithLayout(layout)` ve `WithContent(post)` bu ikisini adlandırır ve **içeriği
layout'un `content` slot'una kayıt yerleştirir**. Bağlamayı kendiniz yapmazsınız.
Layout'un `content` slot'unu boş bırakın: onu kayıt doldurur ve tek fragment tutan
bir slot ikincisini reddeder (`ErrSlotOccupied`).

Bir sayfanın içeriği olmalıdır — yoksa `ErrMissingContent`. Layout ise isteğe
bağlıdır: layout'u olmayan bir sayfa, içerik fragment'ini yanıtın tamamı olarak
render eder; şablonu eksiksiz bir HTML belgesi olan bir sayfanın istediği de budur.

### Tek layout, çok sayfa

Layout, bir sitede en çok yeniden kullanılan şeydir; bu yüzden tek bir
`*collage.Fragment` değerini hata sayfaları dahil bütün sayfalar arasında paylaşmak
güvenlidir. Kayıt, içeriği bağlamadan önce her sayfaya **layout'un slot tablosunun
özel bir kopyasını** verir; böylece A sayfasının içeriği asla B sayfasında görünmez.

Yalnızca slot tablosu kopyalanır. Layout'un şablonu, data handler'ı, yedeği ve diğer
slot'larına zaten bağlanmış fragment'ler paylaşımlı kalır — yani kayıt bu
bağlamaların anlık bir görüntüsünü alır. Bir sayfa kaydedildikten sonra paylaşılan
layout'a bağlanan bir fragment o sayfada görünmez. Layout'u eksiksiz kurun, sonra
sayfaları onunla kaydedin.

İskelet (scaffold) layout'unu, her çağrıda yeni bir fragment döndüren
`layouts.Layout()` fonksiyonu olarak yazar. Bu da aynı derecede iyi çalışır; tek bir
değeri paylaşmak yalnızca izin verilen bir seçenektir.

## Yollar

`WithPath(locale, pattern)` sayfaya tek bir locale'de ulaşan URL'yi kaydeder. Tek
dilli bir site tek bir locale kullanır; bir sayfanın her locale'de farklı bir yolu
olabilir:

```go
collage.NewPage("about").
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda")
```

Bir pattern segment'lerden oluşur:

| Segment | Eşleştiği |
| --- | --- |
| `blog` | Tam olarak bu metin |
| `{slug}` | Tam olarak bir segment, `slug` olarak yakalanır |
| `{rest...}` | Geriye kalan her şey, `rest` olarak yakalanır. Yalnızca son segment olarak |

Bir data handler yakalananı `rc.Param("slug")` ya da `rc.PathParams["slug"]` ile
okur. Değerler percent-decode edilmiş olarak, her seferinde bir segment gelir; bu
yüzden bir segmentin içindeki kodlanmış bir `/` yeni bir segment değil, değerin
parçasıdır.

Her düzeyde statik bir segment `{param}`'dan önce, `{param}` da `{rest...}`'ten önce
denenir — geri izlemeyle (backtracking); böylece ikisi de eşleşebilecek olsa bile
`/blog/archive`, `/blog/{slug}`'a üstün gelir. `/blog` ve `/blog/` aynı route'tur.

Pattern'lerdeki hatalar, istek anında çıkan sürprizler değil, kayıt sırasındaki
hatalardır:

- Bir pattern `/` ile başlamalı, boş segment ve boş placeholder adı içermemeli ve
  catch-all'u yalnızca en sona koymalıdır — `ErrInvalidPath` ya da
  `ErrInvalidPattern`.
- Placeholder bütün bir segmenttir. v0.11.0'dan itibaren bir segmentin içine
  yazılmış bir placeholder, örneğin `/feeds/{category}.xml` ya da `/post-{id}`,
  `ErrInvalidPattern`'dir; bunun yerine `/feeds/{category}/rss.xml` yazın.
- Tek bir locale'de aynı yolda iki route — `ErrDuplicateRoute`. Bir sayfa ile bir
  [document](/docs/documents)'ın çakışması da buna dahildir, çünkü ikisi tek bir
  ağacı paylaşır.
- Aynı konumda iki parametre adı, örneğin `/blog/{slug}` ve `/blog/{id}/edit` —
  `ErrAmbiguousParameterName`.

Bir sayfa `GET` ve `HEAD`'e, `OPTIONS`'a da `Allow` header'ı URL'nin kabul ettiği
metotları listeleyen bir `204` ile yanıt verir. Diğer her metot, sayfanın o metot
için bir [action](/docs/forms-and-actions)'ı yoksa, aynı `Allow` header'ıyla bir
405'tir — bir form da üzerinde durduğu sayfaya böyle gönderilir.

Locale'ler, `/tr/hakkinda` gibi locale önekleri — ve `/en/about`'u `/about`'a
gönderen yönlendirme — ve sayfa adlarından bağlantı kurmak
[Bağlantılar ve locale'ler](/docs/links-and-locales) sayfasında anlatılıyor. Bir sayfa
ayrıca `WithRedirect(from, to, status)` ve `WithPermanentRedirect(from, to)` ile eski
URL'lerden yönlendirmeler taşıyabilir.

## Render stratejileri

Her sayfanın, çıktısının önbelleğe alınıp alınmayacağına karar veren üç stratejiden
biri vardır:

| Metot | Strateji | Davranış |
| --- | --- | --- |
| `Dynamic()` | `StrategyDynamic` | Her istekte render edilir, asla önbelleğe alınmaz. **Varsayılan** |
| `Static()` | `StrategyStatic` | Bir kez render edilir, etiketleri geçersiz kılınana kadar önbellekten sunulur |
| `Incremental(ttl)` | `StrategyIncremental` | Önbellekten sunulur, `ttl` geçtikten sonra yeniden render edilir |

`Incremental` pozitif bir TTL ister (`ErrMissingTTL`). `collage export`'un dosyalara
yazabildiği sayfalar da `Static` ve `Incremental` sayfalardır; `Dynamic` bir sayfa,
istek başına render edilmek için var olduğundan
[atlanır](/docs/static-export#what-is-skipped).

Önbelleğe alma yalnızca uygulama onu etkinleştirdiğinde gerçekleşir
(`Config.Cache.Enabled`, iskelette açık) ve geliştirme modunda asla gerçekleşmez;
orada önbellekteki bir sayfa az önce düzenlediğiniz şablonu gizlerdi.

İki builder metodu daha önbellekteki bir sayfayı şekillendirir.
`WithDependency(tags...)`, data handler'ların bildirdiği bağımlılık etiketlerine
sayfanın kendi etiketlerini ekler; `WithCacheParams(names...)` ise hangi query
parametrelerinin önbellek anahtarına katıldığını söyler. Bütün bunların — etiketler,
geçersiz kılma, query parametreleri, sayfalar yerine verinin önbelleğe alınması —
nasıl bir araya geldiği [Önbellekleme](/docs/caching) sayfasında anlatılıyor.

## Bulunamadı ve hata sayfaları

Bir sayfa gösterilemediğinde onun yerine başka bir sayfa gösterilir. İki durum ve
iki düzey vardır.

- **Bulunamadı — 404.** Hiçbir route URL ile eşleşmedi ya da `Required()` bir
  fragment'in data handler'ı `collage.ErrNotFound`'u sarmalayan bir hata döndürdü:
  URL'nin adlandırdığı içerik mevcut değil.
- **Hata — 500.** `Required()` bir fragment başka bir şekilde başarısız oldu ya da
  framework'ün kendi yolundaki bir şey başarısız oldu.

Bir sayfa kendi yerine geçecek sayfaları adlandırabilir, uygulama da site genelinde
geçerli olanları:

```go
post := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(postContent).
	WithPath("en", "/blog/{slug}").
	WithNotFoundPage(postNotFound). // "no such post", with a search box
	WithErrorPage(postError).
	Build()

app.RegisterNotFoundPage(siteNotFound)
app.RegisterErrorPage(siteError)
```

Bir hata durumunda önce sayfanın kendi `NotFoundPage`'i ya da `ErrorPage`'i, sonra
site genelindeki sayfa, sonra da framework'ün yerleşik sayfası kullanılır. Hiçbir
route ile eşleşmeyen bir URL'nin soracağı bir sayfa yoktur; bu yüzden doğrudan site
genelindeki bulunamadı sayfasına gider.

Bir hata sayfası diğerleri gibi bir sayfadır — bir layout, bir içerik fragment'i,
isterse data handler'lar — ama yolu yoktur. Yola ihtiyacı da yoktur; ona bir şeyin
başarısız olmasıyla ulaşılır.

```go
func NotFoundPage() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").Build()

	return collage.NewPage("not-found").
		WithLayout(layouts.Layout()).
		WithContent(content).
		Dynamic().
		Build()
}
```

Birkaç kural, hata sayfalarının en kötü anda başarısız olmasını önler:

- **`WithNotFoundPage` ya da `WithErrorPage` ile adlandırılan her sayfa kaydedilmiş
  olmalıdır**: `RegisterPage`, `RegisterNotFoundPage` ya da `RegisterErrorPage` ile.
  Aksi hâlde uygulama `ErrUnregisteredErrorPage` ile başlamayı reddeder — nedeni için
  [aşağıya](#why-the-registered-value-matters) bakın.
- Şablonları, onları adlandıran sayfa kaydedildiğinde denetlenir; böylece bir 500
  sayfasındaki yazım hatası, site zaten hata verirken keşfedilen bir şey değil, bir
  başlangıç hatası olur.
- Bir sayfa kendi hata sayfası olamaz (`ErrSelfErrorPage`).
- Bir hata sayfasının render'ı asla önbelleğe alınmaz; başarısız olursa ya da hiçbir
  şey render etmezse onun yerine yerleşik sayfa sunulur, hata da loglanır ve
  plugin'lere bildirilir. Kendi içine 500 verebilen bir 500 sayfası bir kesintidir.

Yerleşik sayfa, harici dosya kullanmayan, kendi kendine yeten bir HTML'dir; bu yüzden
bozulan şey asset'ler olsa bile render edilir. Geliştirme modunda hatanın başladığı
fragment'i adlandırır ve hata zincirini gösterir; production'da ise tek bir genel
cümle gösterir, çünkü hata metni host adlarını, dosya yollarını ve kimlik bilgilerini
sızdırır. [Hatalar](/docs/errors) framework'ün bildirdiği her hatayı listeler.

`collage export`, `RegisterNotFoundPage` ile kaydedilen sayfayı `404.html` olarak
yazar; statik barındırma hizmetlerinin eksik bir URL için sunduğu dosya budur.

## Kayıt

```go
if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

`RegisterPage`, bir sayfanın denetlendiği ve bir araya getirildiği yerdir; böylece
render sırasında başarısız olacak şey burada, başlangıçta, mesajda sayfanın adıyla
başarısız olur — sayfanın tuttuğu her fragment için, kendi URL'sinde açtığı fragment
dahil. Bir istek çalışırken bir
[slot resolver](/docs/fragments-and-slots#slots-filled-per-render)'ın döndürdüğü
fragment buradan görülemez; o, ilk render edildiğinde denetlenir.
Sırasıyla şunları yapar:

1. nil bir sayfayı, boş bir adı, başka bir sayfanın tuttuğu bir adı
   (`ErrDuplicatePage`) ve uygulama başladıktan sonra yapılan her kaydı
   (`ErrAppStarted`) reddeder;
2. builder'ı ya da ondan ulaşılabilen veya
   [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile açılan
   herhangi bir fragment'in builder'ı bir hata not etmiş olan sayfayı — yani
   `BuildErr()`'ün döndüreceği şeyi — reddeder;
3. layout'un slot tablosunu kopyalar ve içerik fragment'ini onun `content` slot'una
   bağlar;
4. sayfayı ve fragment yolları dahil bütün fragment ağacını doğrular: yollar,
   strateji ve TTL, yönlendirmeler, içi boş zorunlu slot'lar, kendisinden
   ulaşılabilen bir fragment;
5. her fragment'in şablonunun yüklendiğini denetler (`ErrTemplateNotFound`) —
   v0.11.0'dan itibaren ağacın geri kalanı gibi denetlenen, `WithFragmentPath` ile
   açılmış fragment'ler ve sayfanın kendi bulunamadı ve hata sayfaları dahil;
6. sayfanın yollarını, yönlendirmelerini, action'larını ve fragment yollarını
   router'a ekler.

Her hatayı ölümcül sayın. Yarıda başarısız olan bir kayıt geri alınmaz — bir
locale'de kabul edilen yol, bir sonrakinin reddedilmesi üzerine router'da kalır —
çünkü başarısız bir kayıt, toparlanılacak bir durum değil, başlamaması gereken bir
programdır.

### Kaydedilen değer neden önemli

Kayıt, kendisine verilen sayfayı değiştirir. 3. adım `page.LayoutFragment`'i,
layout'un sayfaya özel, bağlanmış kopyasıyla değiştirir ve render edilen şey —
slot'unda içerikle birlikte — o kopyadır. Aynı constructor yeniden çağrılarak kurulan
bir sayfa farklı bir değerdir ve layout'unun `content` slot'u boştur.

Dolayısıyla `RegisterPage`'e verdiğiniz değer o andan itibaren sayfanın ta
kendisidir ve bir sayfaya değer üzerinden başvuran her şey o değeri kullanmalıdır:

- **Hata sayfaları.** `WithNotFoundPage(p)`, kaydedilmiş olan `*collage.Page`'in ta
  kendisini göstermelidir. Aynı adla ayrıca kurulmuş bir sayfa
  `ErrUnregisteredErrorPage` ile reddedilir, çünkü layout'unu hiçbir şeyin etrafında
  render ederdi.
- **Sayfayla yanıt veren action'lar.** Bir action'daki `collage.RenderPage(p)`'ye
  kaydedilmiş değer verilmelidir; kaydedilmemiş bir değer boş render edilmek yerine
  `ErrUnregisteredPage` ile reddedilir. Olağan biçim, sayfayı bir kez kurmak ve
  action'ın closure'ının onu yakalamasına izin vermektir:

  ```go
  var page *collage.Page
  page = collage.NewPage("hello").
  	WithLayout(layouts.Layout()).
  	WithContent(content).
  	WithPath("en", "/hello").
  	WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
  		return collage.RenderPage(page), nil
  	}).
  	Build()
  return page
  ```

- **`rc.Page` de o değerdir**; sayfayı render eden her istek ve onlara hizmet eden
  her goroutine tarafından paylaşılır. Onu serbestçe okuyun; asla ona yazmayın.
  İstekten isteğe değişen her şeyin yeri, bir handler'ın döndürdüğü veri ya da
  render'ın paylaşılan verisidir — bkz. [Data handler'lar](/docs/data-handlers#the-render-context).

`app.Page(name)` ve `app.Pages()`, kaydedilmiş sayfaların kopyalarını onlara bakmak
için döndürür — her sayfanın yollarını listeleyen bir sitemap, bir stratejiyi
denetleyen bir test. Kopya, kaydedilmiş değer değildir; bu yüzden kaydedilmiş bir
sayfanın beklendiği yere kopya vermeyin.
