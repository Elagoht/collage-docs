---
description: Page nedir, layout'ları ve içeriği nasıl bir araya gelir, onu kim görebilir, ona hangi path'ler ulaşır, nasıl cache'lenir, başarısız olduğunda ne gösterir ve register edilmek onu nasıl değiştirir.
reference: Registrable, ErrNilRegistrable, NewPage, PageBuilder, PageBuilder.WithLayouts, Page, RenderPage, DefaultContentSlot, StrategyAuto, FragmentBuilder.WithGuard, GuardFunc, GuardDecision
---

# Page'ler ve layout'lar

Bir **page**, adı olan bir render yapılandırmasıdır. Page'i hangi fragment'lerin
oluşturduğunu belirler: bir layout ve onun içindeki içerik. Ayrıca page'e hangi
URL'lerin ulaştığını, çıktısının nasıl cache'lendiğini ve render edilemediğinde
yerine hangi page'in gösterileceğini de belirler. Page'in kendine ait bir template'i
ya da verisi yoktur. Bunlar fragment'lerine aittir.

```go
page := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()

if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

## Page oluşturmak

`collage.NewPage(name)` bir builder başlatır. Her `WithX` çağrısı tek bir şeyi
ayarlar, `Build()` da `*collage.Page`'i döner.

Ad, page'in kimliğidir. Link'ler bu addan üretilir
(`{{pageURL "blog-post" "slug" .Slug}}`), testler ve plugin'ler page'leri bu adla
bulur ve iki page aynı adı taşıyamaz. Adı page'in nerede durduğuna göre değil, ne
olduğuna göre seçin. Böylece URL değişse de ad geçerliliğini korur.

Builder'lar hata dönmek için zinciri hiçbir zaman kesmez. İstenen şeyi yapamayan bir
çağrı hatayı kaydeder ve devam eder. `BuildErr()` kaydedilen bütün hataları döner:

```go
builder := collage.NewPage("blog-post").WithLayouts(layout).WithPath("en", "/blog/{slug}")
page := builder.Build()
if err := builder.BuildErr(); err != nil {
	return err // collage.ErrMissingContent: there is no WithContent
}
```

`BuildErr`'ü görmezden gelmek bir hatanın gözden kaçmasına yol açmaz. Builder'ın
kaydettiği hatalar, ürettiği değerin üzerinde kalır. `RegisterPage` bu hatalardan
birini taşıyan bir page'i reddeder ve page'in adını içeren bir hata döner. Hatanın
page'in kendisine ya da ağacındaki herhangi bir fragment'e ait olması fark etmez;
örneğin iki kez tanımlanmış bir slot da buna dahildir. Birinin `BuildErr`'ü çağırıp
çağırmamış olması da sonucu değiştirmez. `BuildErr`'ü, kontrolünüzde olmayan bir
girdiden oluşturulan page'lerde kontrol edin. Orada hatayı kaynağına daha yakın bir
yerde görmek istersiniz.

## Layout ve içerik

Bir sitedeki page'lerin çoğu dış kısmı ortak kullanır: `<head>`, header ve footer.
Farklı olan iç kısımdır. Dış kısım **layout**'tur. Layout, template'i `content`
adında bir slot'u çağıran bir fragment'tir. İç kısım ise **content fragment**'tir.
Page, bu fragment'i göstermek için vardır.

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithTitle("My site").
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

Layout hiçbir slot tanımlamaz. Template'indeki `{{slot "content"}}` çağrısı
tanımın kendisidir. Çağırdığı diğer slot'lar için de durum aynıdır; bkz.
[Fragment'ler ve slot'lar](/docs/fragments-and-slots#slots). `WithTitle`, layout'u
kullanan her page'e, içerideki bir şey daha iyisini belirtene kadar bir `<title>`
verir.

`WithLayouts(layout)` ve `WithContent(post)` bu ikisini belirtir. **İçeriği
layout'un `content` slot'una (`collage.DefaultContentSlot`) register işlemi
yerleştirir.** Bu bağlamayı kendiniz yapmazsınız. Template'i hiç
`{{slot "content"}}` çağırmayan bir layout ise register sırasında `ErrUnknownSlot`
ile reddedilir, çünkü içeriğin render edileceği bir yer yoktur. Layout'ta
`WithSlot(collage.DefaultContentSlot, true, false)` hâlâ kabul edilir. Bu,
`content`'e page'in kendi içeriği dışında bağlanan her fragment'i reddetmesi
gereken bir layout içindir (`ErrSlotOccupied`).

Her page'in bir içeriği olmalıdır, yoksa `ErrMissingContent` alırsınız. Layout ise
isteğe bağlıdır. Layout'u olmayan bir page, content fragment'ini response'un
tamamı olarak render eder. Template'i baştan sona eksiksiz bir HTML belgesi olan bir
page'in istediği de tam olarak budur.

### Tek layout, birçok page

Layout, bir sitede en çok tekrar kullanılan şeydir. Bu yüzden tek bir
`*collage.Fragment` değerini error page'ler dahil bütün page'ler arasında paylaşmak
güvenlidir. Register işlemi, içeriği bağlamadan önce her page'e **layout'un slot
tablosunun kendine ait bir kopyasını** verir. Böylece A page'inin içeriği hiçbir
zaman B page'inde görünmez.

Yalnızca slot tablosu kopyalanır. Layout'un template'i, data handler'ı, fallback'i
ve diğer slot'larına önceden bağlanmış fragment'ler ortak kalır. Bu da register
işleminin o bağlamaların bir anlık görüntüsünü aldığı anlamına gelir. Bir page
register edildikten sonra ortak layout'a bağlanan bir fragment o page'de görünmez.
Önce layout'u eksiksiz oluşturun, sonra page'leri onunla register edin.

Scaffold, layout'unu her çağrıda yeni bir fragment dönen bir fonksiyon olarak
yazar: `layouts.Master()`. Bu yöntem de aynı şekilde çalışır. Tek bir değeri
paylaşmak yalnızca izin verilen bir seçenektir.

### Layout içinde layout

Bir page birden fazla layout'un içinde durabilir: dışta `<head>`'iyle HTML
iskeleti, onun içinde de giriş page'lerinin paylaştığı daha dar bir çerçeve.
`WithLayouts` zincirin tamamını **en dıştakinden başlayarak** alır. v0.28.0'dan
beri vardır ve `WithLayout`'un yerini almıştır; tek bir layout
`WithLayouts(layout)` olarak yazılır:

```go
func Master() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("My site").
		Build()
}

func Auth() *collage.Fragment {
	return collage.NewFragment("auth-layout", "layouts/auth.html").Build()
}

page := collage.NewPage("login").
	WithLayouts(layouts.Master(), layouts.Auth()).
	WithContent(login).
	WithPath("en", "/login").
	Build()
```

Tek bir layout içeriği nasıl render ediyorsa, zincirdeki her layout da
içindekini `{{slot "content"}}` ile öyle render eder. Sarmalamanın tamamını
register işlemi yapar. İçerik en içteki layout'un `content` slot'una, her layout
da kendisini saran layout'un `content` slot'una girer. Bu yüzden içinde boşluk
bırakan bir layout da bitmiş bir `*collage.Fragment`'tir. Onu oluşturan bir
yardımcı fonksiyon, page'in tamamlaması gereken bir builder değil, build edilmiş
fragment'i döner.

Tek bir layout'ta olduğu gibi her page, zincirdeki her layout'un slot tablosunun
kendine ait bir kopyasını alır. Bu yüzden zincirin layout'ları site genelinde
paylaşılabilir. Zincirdeki bir layout'un `content` slot'u boş gelmelidir, çünkü onu
register işlemi doldurur. Dolu gelen bir layout `ErrSlotOccupied` ile reddedilir.

`WithLayouts`, hiç layout verilmeden çağrılırsa `ErrMissingLayout`, nil bir layout
için `ErrNilFragment`, aynı layout iki kez verilirse `ErrFragmentCycle`, aynı
builder'da ikinci kez çağrılırsa `ErrConflictingLayout` kaydeder.

## Private page'ler: guard'lar

Sitenin yalnızca giriş yapmış okuyucuların görebileceği bir bölümü, bunu söyleyen
bir layout'tur. `WithGuard` (v0.28.0'dan beri) bir fragment'e bir fonksiyon
verir. Page sunulmadan önce bu fonksiyona, bu request'in page'i alıp alamayacağı
sorulur:

```go
func requireUser(ctx context.Context, r *http.Request) (*collage.GuardDecision, error) {
	if session.FromContext(ctx).Get("user") != "" {
		return nil, nil // allowed
	}
	return &collage.GuardDecision{
		Status:   http.StatusSeeOther,
		Location: "/login?next=" + url.QueryEscape(r.URL.RequestURI()),
	}, nil
}

func Private() *collage.Fragment {
	return collage.NewFragment("private", "layouts/private.html").
		WithGuard(requireUser).
		Build()
}

page := collage.NewPage("dashboard").
	WithLayouts(layouts.Master(), layouts.Private()).
	WithContent(dashboard).
	WithPath("en", "/dashboard").
	Build()
```

`Private()` ile sarılan her page private'tır; bir page'i public yapmak için
`Private()`'ı zincirine eklememek yeterlidir. Route'larla senkron tutulacak bir
korumalı path listesi yoktur. Guard, page'e hangi URL ulaşırsa ulaşsın, her locale'de ve her
parametreyle çalışır.

Bir guard, page'in **omurgasında** olduğunda o page için sorulur: page'in
zincirindeki bir layout ya da page'in content fragment'i. Guard'lar en dıştakinden
başlayarak çalışır, content fragment'inki en son sorulur ve ilk cevap veren karar
verir. Başka bir fragment'teki guard yok sayılır: bir slot'a bağlanan, bir
resolver'ın döndüğü ya da fallback olan fragment'lerdekiler. Bir page'i kimin
görebileceği, onu oluşturan fragment'lerin değil, page'in kendi özelliğidir.

Guard üç şekilde cevap verir:

| Dönüş | Okuyucunun aldığı |
| --- | --- |
| `nil, nil` | Page |
| bir `3xx` status ve bir `Location` | Oraya body'siz bir redirect. Location verilip status sıfır bırakılırsa `303 See Other` olur. `Collage-Fetch` ile işaretlenmiş bir request, bir action'ın redirect'inde olduğu gibi `204` alır; hedef `Collage-Location` içinde gelir. |
| bir `4xx` ya da `5xx` status, location yok | O status, body'siz |

Bunların dışındaki her şey (gidecek yeri olmayan bir redirect, bir `200`)
`ErrInvalidGuardDecision`'dır. Guard'ın bir hata dönmesinde olduğu gibi request
500 ile başarısız olur. Ne demek istediğini söyleyemeyen bir guard'ın niyeti
tahmin edilmez: location'ı olmayan bir redirect, aksi hâlde tam da saklaması
gereken page'e düşerdi.

Guard'ın kapsadıkları:

- **Page'in render'ları**, `GET` ve `HEAD`. Guard routing'den sonra ve **page'in
  cache'i okunmadan önce** çalışır. Böylece engellenen bir okuyucu cache'lenmiş bir
  render'a hiç ulaşmaz ve private bir page `Static()` olabilir. Guard
  `PageResolvedHook`'tan da önce çalışır: engellenen request page'e hiç
  ulaşmamıştır, bu yüzden page'leri izleyen plugin'lere haber verilmez.
- **Page'in kendi URL'sindeki action'lar.** Bir form bulunduğu page'e post eder.
  Okuyucunun göremediği bir page, form'unu da gönderemeyeceği bir page'dir. Guard,
  body okunmadan ve forgery kontrolünden önce çalışır.
- **Page'in fragment path'leri.**
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile kendi
  URL'sinde açılan bir fragment page'in bir parçasıdır. v0.34.0'dan beri
  fragment'in kendi guard'ından önce page'in guard'larından geçer. Öncesinde bir
  fragment path yalnızca fragment'inin guard'ını çalıştırırdı. Private bir
  page'deki fragment path de, o fragment bir guard taşımıyorsa public'ti.
- **Kendi URL'sinde register edilen bir action ve page'in not-found ile error
  page'leri kapsanmaz.** Aksi hâlde private bir error page, hatayla karşılaşan
  okuyucuyu redirect ederdi.

Guard'ın izin verdiği okuyucular page'in cache'ini paylaşır; bu cache sunucunun
kendi cache'idir. Sunucunun önündeki bir CDN ya da proxy guard çalıştırmaz. Bu
yüzden guard'lı bir page, stratejisi ne derse desin `Cache-Control: private,
no-cache` ile gönderilir. v0.34.0'dan beri guard'ın kendi cevabı (giriş yapmaya
yönlendiren bir redirect, bir ret) `no-store` ile gönderilir. Böylece varsayılan
olarak bir 404'ü ya da 308'i saklayan bir CDN, bir okuyucuya verilen cevabı
sonrakine vermez. İçeriği okuyucudan okuyucuya değişen bir page ise guard
değil kişiselleştirme konusudur; bkz. [Caching](/docs/caching#render-strategies).
[Static export](/docs/static-export#what-is-skipped) guard'lı page'leri hiç yazmaz.

Guard'ın neyi kontrol ettiği framework'ün bilebileceği bir şey değildir. Yukarıdaki
`requireUser` session'ı okur; bir guard aynı şekilde bir header'ı, bir rolü, bir
feature flag'i, kısacası request'in taşıdığı her şeyi okuyabilir.
[Session plugin'i](/docs/plugins#elagohtsession) bu guard'ı hazır olarak,
`session.RequireUser("/login")` şeklinde sunar. `collage inspect` her page için
guard taşıyan fragment'leri listeler.

## Path'ler

`WithPath(locale, pattern)`, bir locale'de page'e ulaşan URL'yi register eder. Tek
dilli bir site tek bir locale kullanır. Bir page'in her locale'de farklı bir path'i
olabilir:

```go
collage.NewPage("about").
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda")
```

Bir pattern segment'lerden oluşur:

| Segment | Neyle eşleşir |
| --- | --- |
| `blog` | Tam olarak bu metinle |
| `{slug}` | Tam olarak bir segment'le, `slug` adıyla yakalanır |
| `{rest...}` | Geriye kalan her şeyle, `rest` adıyla yakalanır. Yalnızca son segment olabilir |

Bir data handler yakalanan değeri `rc.Param("slug")` ya da `rc.PathParams["slug"]`
ile okur. Değerler percent-decode edilmiş olarak ve segment segment gelir.

v0.34.0'dan beri encode edilmiş bir slash (`%2F`) içeren bir segment 404 döner.
Middleware decode edilmiş path'i okur ve orada `/public%2Fsecret` iki segment'tir.
Onu tek segment olarak okuyan bir router (collage'ın eskiden yaptığı gibi),
middleware'in başka bir şey olarak değerlendirdiği bir request'i sunmuş olurdu.
Kendisi bir path olan bir değer, örneğin `guide/intro`, bir catch-all'da yer
almalıdır: `{rest...}`. `{{pageURL}}` ve `BuildPath`, tek bir segment'in
değerindeki `/`'ı reddeder.

Her seviyede static bir segment `{param}`'dan önce, `{param}` da `{rest...}`'ten
önce denenir. Eşleştirme backtracking ile yapılır. Böylece iki route da
eşleşebilecek olsa bile `/blog/archive`, `/blog/{slug}`'a karşı kazanır. `/blog` ve
`/blog/` aynı route'tur.

Pattern'lerdeki hatalar request anında sürpriz olarak çıkmaz. Register sırasında
hata verirler:

- Bir pattern `/` ile başlamalıdır. Boş segment ve boş placeholder adı içeremez,
  catch-all'u da yalnızca en sona koyabilir. Aksi hâlde `ErrInvalidPath` ya da
  `ErrInvalidPattern` alırsınız.
- Bir placeholder bütün bir segment'i kaplar. v0.11.0'dan itibaren bir segment'in
  içine yazılmış bir placeholder, örneğin `/feeds/{category}.xml` ya da `/post-{id}`,
  `ErrInvalidPattern` verir. Bunun yerine `/feeds/{category}/rss.xml` yazın.
- Bir locale'de aynı path'te iki route olursa `ErrDuplicateRoute` alırsınız. Bir
  page ile bir [document](/docs/documents)'ın çakışması da buna dahildir, çünkü
  ikisi aynı ağacı kullanır.
- Aynı konumda iki farklı parametre adı, örneğin `/blog/{slug}` ve
  `/blog/{id}/edit`, `ErrAmbiguousParameterName` verir.

Bir page `GET` ve `HEAD` request'lerine cevap verir. `OPTIONS` request'ine de
`204` ile cevap verir. Bu response'un `Allow` header'ı, URL'nin kabul ettiği
metotları listeler. Diğer bütün metotlar aynı `Allow` header'ıyla 405 alır. Tek
istisna, page'in o metot için bir [action](/docs/forms-and-actions)'ı olmasıdır.
Bir form, üzerinde bulunduğu page'e bu sayede post eder.

Locale'ler, `/tr/hakkinda` gibi locale prefix'leri, `/en/about`'u `/about`'a
gönderen redirect ve page adlarından link üretmek
[Link'ler ve locale'ler](/docs/links-and-locales) sayfasında anlatılıyor. Bir page
ayrıca eski URL'lerden redirect'ler de taşıyabilir. Bunun için
`WithRedirect(from, to, status)` ve `WithPermanentRedirect(from, to)` kullanılır.
v0.34.0'dan beri bir redirect, request'in query string'ini hedefine taşır. Yakaladığı
bir değer, düştüğü yere göre escape edilir: bir path segment'i ya da `?` işaretinden
sonraki bir query değeri. Böylece `/login?next={slug}`'a ikinci bir `next`
verilemez. Siteden çıkacak bir hedef (`//host` ya da `/\host`) ise register sırasında
`ErrInvalidPath` olur.

## Render stratejileri

Her page'in üç stratejiden biri vardır. Strateji, page'in çıktısının cache'lenip
cache'lenmeyeceğini belirler:

| Metot | Strateji | Davranış |
| --- | --- | --- |
| `Dynamic()` | `StrategyDynamic` | Her request'te render edilir, hiçbir zaman cache'lenmez |
| `Static()` | `StrategyStatic` | Bir kez render edilir, tag'leri invalidate edilene kadar cache'ten sunulur |
| `Incremental(ttl)` | `StrategyIncremental` | Cache'ten sunulur, `ttl` dolduktan sonra yeniden render edilir |

Bu üç çağrıdan hiçbirini yapmayan bir page, register edilene kadar
`StrategyAuto`'dur. Register işlemi stratejiyi belirler: render ettiği herhangi bir
şeyin data handler'ı ya da slot resolver'ı varsa **dynamic**, yoksa **static**
olur. Template'lerden ve sabit değerlerden (`WithData`, `WithTitle`) oluşan bir page
herkes için aynı render edilir, bu yüzden söylemeye gerek kalmadan cache'lenir.
Handler ise request'i, bir cookie'yi ya da saati okuyabilir ve dışarıdan bu
anlaşılamaz. Bu yüzden handler'ı olan bir page, `Static()` ya da
`Incremental(ttl)` diyene kadar her request'te render edilir. Kuralın tamamı
[Caching](/docs/caching#a-page-that-declares-none) sayfasındadır.

`Incremental` pozitif bir TTL ister (`ErrMissingTTL`). `collage export`'un dosyaya
yazabildiği page'ler de static ve incremental page'lerdir. Dynamic bir page her
request'te render edilmek için vardır, bu yüzden
[atlanır](/docs/static-export#what-is-skipped).

Cache yalnızca uygulama onu açtığında devreye girer (`Config.Cache.Enabled`,
scaffold'da açıktır). Development'ta ise hiçbir zaman devreye girmez. Orada
cache'lenmiş bir page, az önce düzenlediğiniz template'i gizlerdi.

Cache'lenen bir page'i iki builder metodu daha etkiler. `WithDependency(tags...)`,
data handler'ların bildirdiği dependency tag'lerine page'in kendi tag'lerini ekler.
`WithCacheParams(names...)` ise cache key'ine hangi query parametrelerinin
katılacağını belirler. Bunların hepsinin nasıl bir araya geldiği
[Caching](/docs/caching) sayfasında anlatılıyor: tag'ler, invalidation, query
parametreleri ve page'ler yerine veriyi cache'lemek.

## Not-found ve error page'leri

Bir page gösterilemediğinde yerine başka bir page gösterilir. Bunun iki durumu ve
iki seviyesi vardır.

- **Not found, 404.** Ya hiçbir route URL ile eşleşmemiştir ya da `Required()` bir
  fragment'in data handler'ı `collage.ErrNotFound`'u wrap eden bir hata dönmüştür.
  Yani URL'nin işaret ettiği içerik yoktur.
- **Error, 500.** Ya `Required()` bir fragment başka bir sebeple başarısız olmuştur
  ya da framework'ün kendi akışında bir şey başarısız olmuştur.

Bir page kendi yerine geçecek page'leri belirleyebilir. Uygulama da site genelinde
geçerli olanları belirleyebilir:

```go
post := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(postContent).
	WithPath("en", "/blog/{slug}").
	WithNotFoundPage(postNotFound). // "no such post", with a search box
	WithErrorPage(postError).
	Build()

app.RegisterNotFoundPage(siteNotFound)
app.RegisterErrorPage(siteError)
```

Bir hata olduğunda önce page'in kendi `NotFoundPage`'i ya da `ErrorPage`'i
kullanılır. Sonra site genelindeki page, en son da framework'ün built-in page'i
devreye girer. Hiçbir route ile eşleşmeyen bir URL'nin sorabileceği bir page yoktur.
Bu yüzden doğrudan site genelindeki not-found page'ine gider.

Bir error page de diğerleri gibi bir page'dir. Bir layout'u, bir content fragment'i
ve isterse data handler'ları olur, ama path'i yoktur. Path'e ihtiyacı da yoktur,
çünkü ona bir şeyin başarısız olmasıyla ulaşılır.

```go
func NotFoundPage() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").Build()

	return collage.NewPage("not-found").
		WithLayouts(layouts.Master()).
		WithContent(content).
		Dynamic().
		Build()
}
```

Birkaç kural, error page'lerin en kötü anda başarısız olmasını önler:

- **`WithNotFoundPage` ya da `WithErrorPage` ile belirtilen her page register
  edilmiş olmalıdır.** Bunun için `RegisterPage`, `RegisterNotFoundPage` ya da
  `RegisterErrorPage` kullanılır. Aksi hâlde uygulama `ErrUnregisteredErrorPage` ile
  başlamayı reddeder. Nedenini [aşağıda](#why-the-registered-value-matters)
  bulabilirsiniz.
- Bu page'lerin template'leri, onları belirten page register edilirken kontrol
  edilir. Böylece bir 500 page'indeki yazım hatası, site zaten hata verirken fark
  edilmez. Uygulama başlarken hata olarak karşınıza çıkar.
- Bir page kendi error page'i olamaz (`ErrSelfErrorPage`).
- Bir error page'in render çıktısı hiçbir zaman cache'lenmez. Render başarısız olursa
  ya da hiçbir şey üretmezse yerine built-in page sunulur. Hata da log'lanır ve
  plugin'lere bildirilir. Kendi kendine 500 verebilen bir 500 page'i, sitenin
  tamamen çökmesi demektir.

Built-in page, hiçbir harici dosyaya ihtiyaç duymayan, kendi içinde tamamlanmış bir
HTML'dir. Bu yüzden bozulan şey asset'ler olsa bile render edilir. Development'ta
hatanın başladığı fragment'in adını verir ve hata zincirini gösterir.
Production'da ise tek bir genel cümle gösterir, çünkü hata metinleri host adlarını,
dosya yollarını ve credential'ları sızdırır. [Hatalar](/docs/errors) sayfası,
framework'ün bildirdiği bütün hataları listeler.

`collage export`, `RegisterNotFoundPage` ile register edilen page'i `404.html`
olarak yazar. Static host'lar, bulunamayan bir URL için bu dosyayı sunar.

## Register etmek

```go
if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

Bir page `RegisterPage`'te kontrol edilir ve bir araya getirilir. Böylece render
sırasında başarısız olacak bir şey burada, uygulama başlarken başarısız olur ve
hata mesajında page'in adı yer alır. Bu, page'in içerdiği bütün fragment'ler için
geçerlidir. Page'in kendi URL'sinde açtığı bir fragment de buna dahildir. Bir
request çalışırken
[slot resolver](/docs/fragments-and-slots#slots-filled-per-render)'ın döndüğü bir
fragment ise buradan görülemez. O fragment ilk render edildiğinde kontrol edilir.
`RegisterPage` sırasıyla şunları yapar:

1. nil bir page'i, boş bir adı, başka bir page'in kullandığı bir adı
   (`ErrDuplicatePage`) ve uygulama başladıktan sonra yapılan her register işlemini
   (`ErrAppStarted`) reddeder;
2. builder'ı bir hata kaydetmiş olan page'i reddeder. Page'den ulaşılabilen ya da
   [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile
   açılan herhangi bir fragment'in builder'ı için de aynısı geçerlidir. Bu hatalar,
   `BuildErr()`'ün döneceği hatalardır;
3. zincirdeki her layout'un slot tablosunu kopyalar, content fragment'i en içteki
   kopyanın `content` slot'una, her layout'u da kendisini saranınkine bağlar;
4. page'i ve fragment path'leri dahil bütün fragment ağacını doğrular: path'ler,
   strateji ve TTL, redirect'ler, içi boş kalmış required slot'lar ve kendisinden
   yine kendisine ulaşılabilen bir fragment;
5. her fragment'in template'inin yüklendiğini kontrol eder (`ErrTemplateNotFound`).
   `WithFragmentPath` ile açılan fragment'ler de buna dahildir; bunlar v0.11.0'dan
   beri ağacın geri kalanı gibi kontrol edilir. Page'in kendi not-found ve error
   page'leri de kontrol edilir. Ayrıca bir şey bağlanan her slot'un, template'in
   çağırdığı bir slot olduğunu kontrol eder (`ErrUnknownSlot`; hata slot'u ve
   template'in yaptığı çağrıları belirtir);
6. strateji tanımlamamış bir page'in stratejisini bütün ağaca bakarak belirler;
7. page'in path'lerini, redirect'lerini, action'larını ve fragment path'lerini
   router'a ekler.

Her hatayı fatal kabul edin. Yarıda başarısız olan bir register işlemi geri
alınmaz. Örneğin bir locale'de kabul edilen path, bir sonraki locale reddedildiğinde
router'da kalır. Bunun sebebi, başarısız bir register işleminin toparlanılacak bir
durum olmamasıdır. Bu, hiç başlamaması gereken bir programdır.

### Birkaçını birden register etmek

`app.Register` (v0.40.0'dan beri) page'leri, document'ları ve action'ları tek bir
çağrıda alır ve her birini sırayla `RegisterPage`, `RegisterDocument` ya da
`RegisterAction` üzerinden register eder. Scaffold'un oluşturduğu `routes.go`, ona
yapılan tek bir çağrıdan ibarettir:

```go
func register(app *collage.App) error {
	if err := app.Register(
		landingpages.Home(),
		blogpages.Post(),
		documents.Health(),
		actions.Logout(),
	); err != nil {
		return err
	}
	return app.RegisterNotFoundPage(errorspages.NotFound())
}
```

Reddedilen ilk değerde durur ve onun hatasını türü ve adıyla sarmalayarak döner
(`register page "post": collage: duplicate page ...`). Böylece `errors.Is` asıl
nedeni yine bulur. Ondan önce register edilenler, tekil metotlarda olduğu gibi
register edilmiş olarak kalır. Bir `nil`, `ErrNilRegistrable` olur. Not-found ve
error page'lerine bir path üzerinden değil, eşleşmenin başarısız olmasıyla ulaşılır.
Bu yüzden onlar `RegisterNotFoundPage` ve `RegisterErrorPage`'i kullanmaya devam
eder. `Register` bir `collage.Registrable` alır ve bunu yalnızca `*Page`,
`*Document` ve `*Action` karşılar. [`collage add`](/docs/cli#collage-add) yazdığı
şeyi bu çağrıya ekler.

### Register edilen değer neden önemli

Register işlemi kendisine verilen page'i değiştirir. 3. adım `page.LayoutFragment`'i,
en dıştaki layout'un bu page'e ait ve bağlanmış kopyasıyla değiştirir. Render
edilen şey, içinde zincirin geri kalanı ve içerik bulunan bu kopyadır. Aynı
constructor tekrar çağrılarak oluşturulan bir page ise farklı bir değerdir. Onun
layout'larının `content` slot'ları boştur.

Bu yüzden o andan itibaren page, `RegisterPage`'e verdiğiniz değerin ta kendisidir.
Bir page'e değer üzerinden başvuran her şey o değeri kullanmalıdır:

- **Error page'ler.** `WithNotFoundPage(p)`, register edilmiş olan `*collage.Page`
  değerinin ta kendisini göstermelidir. Aynı adla ayrıca oluşturulmuş bir page
  `ErrUnregisteredErrorPage` ile reddedilir, çünkü o page layout'unu boş bir
  içeriğin etrafında render ederdi.
- **Page ile cevap veren action'lar.** Bir action'da `collage.RenderPage(p)`'ye
  register edilmiş değer verilmelidir. Register edilmemiş bir değer boş render
  edilmez, `ErrUnregisteredPage` ile reddedilir. Page'in kendi URL'sindeki bir
  action o değeri `rc.Page` olarak alır (v0.33.0'dan beri):

  ```go
  return collage.NewPage("hello").
  	WithLayouts(layouts.Master()).
  	WithContent(content).
  	WithPath("en", "/hello").
  	WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
  		return collage.RenderPage(rc.Page), nil
  	}).
  	Build()
  ```

- **`rc.Page` de aynı değerdir.** Bu değeri, page'i render eden bütün request'ler
  ve onlara hizmet eden bütün goroutine'ler ortak kullanır. Onu dilediğiniz gibi
  okuyun, ama hiçbir zaman ona yazmayın. Request'ten request'e değişen her şey, bir
  handler'ın döndüğü veride ya da render'ın shared data'sında yer almalıdır.
  Ayrıntılar için [Data handler'lar](/docs/data-handlers#the-render-context)
  sayfasına bakın.

`app.Page(name)` ve `app.Pages()`, register edilmiş page'lerin kopyalarını döner.
Bu kopyalar incelemek içindir: örneğin her page'in path'lerini listeleyen bir
sitemap ya da bir stratejiyi kontrol eden bir test. Kopya, register edilmiş değer
değildir. Bu yüzden register edilmiş bir page beklenen yere kopya vermeyin.
