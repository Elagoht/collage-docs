---
description: Page nedir: layout'ları ve içeriği, kimin görebileceği, ona ulaşan path'ler, nasıl cache'lendiği, başarısız olunca ne gösterdiği ve register etmenin etkisi.
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
`WithSlot(collage.DefaultContentSlot, collage.SlotRequired, collage.SlotSingle)` hâlâ kabul edilir. Bu,
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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 150" role="img" aria-labelledby="pl-chain-title">
<title id="pl-chain-title">Register işleminden sonra login page'inin layout zinciri. En dıştaki layout olan Master'ın content slot'unda, ikinci ve en içteki layout olan auth layout'u durur. Auth layout'unun content slot'unda da login content fragment'i durur.</title>
<rect class="d-frame" x="4" y="4" width="312" height="142" rx="8" />
<text x="14" y="21">Master()</text>
<text class="d-small" x="306" y="21" text-anchor="end">1. · en dıştaki</text>
<rect class="d-slot d-slot-a" x="14" y="30" width="292" height="106" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="44">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="52" width="272" height="76" rx="6" />
<text class="d-slot-text-a" x="34" y="69">Auth()</text>
<text class="d-small" x="286" y="69" text-anchor="end">2. · en içteki</text>
<rect class="d-slot d-slot-c" x="34" y="78" width="252" height="42" rx="4" />
<text class="d-small" x="40" y="103">slot "content"</text>
<rect class="d-c" x="126" y="84" width="152" height="30" rx="4" />
<text class="d-on-c" x="202" y="103" text-anchor="middle">login · content</text>
</svg>
<figcaption><code>WithLayouts(layouts.Master(), layouts.Auth())</code> zinciri en dıştakinden başlayarak sıralar. İç içe yerleştirmeyi register işlemi yapar: login içeriği en içteki layout olan <code>Auth()</code>'un content slot'una, <code>Auth()</code> da <code>Master()</code>'ın content slot'una girer. Her biri, layout'un slot tablosunun page'e ait kopyası üzerinde yapılır.</figcaption>
</figure>

Tek bir layout içeriği nasıl render ediyorsa, zincirdeki her layout da
içindekini `{{slot "content"}}` ile öyle render eder. Sarmalamanın tamamını
register işlemi yapar. İçerik en içteki layout'un `content` slot'una, her layout
da kendisini saran layout'un `content` slot'una girer.

Bu yüzden içinde boşluk bırakan bir layout da bitmiş bir `*collage.Fragment`'tir.
Onu oluşturan bir yardımcı fonksiyon, page'in tamamlaması gereken bir builder değil,
build edilmiş fragment'i döner.

Tek bir layout'ta olduğu gibi her page, zincirdeki her layout'un slot tablosunun
kendine ait bir kopyasını alır. Bu yüzden zincirin layout'ları site genelinde
paylaşılabilir. Zincirdeki bir layout'un `content` slot'u boş gelmelidir, çünkü onu
register işlemi doldurur. Dolu gelen bir layout `ErrSlotOccupied` ile reddedilir.

`WithLayouts` şu durumlarda, register işleminin reddettiği bir hata kaydeder:

| Çağrı | Hata |
| --- | --- |
| hiç layout verilmeden | `ErrMissingLayout` |
| nil bir layout ile | `ErrNilFragment` |
| aynı layout iki kez verilerek | `ErrFragmentCycle` |
| aynı builder'da ikinci kez | `ErrConflictingLayout` |

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
verir.

Başka bir fragment'teki guard yok sayılır: bir slot'a bağlanan, bir resolver'ın
döndüğü ya da fallback olan fragment'lerdekiler. Bir page'i kimin görebileceği,
onu oluşturan fragment'lerin değil, page'in kendi özelliğidir.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 306" role="img" aria-labelledby="pl-spine-title">
<title id="pl-spine-title">/admin/users için gelen bir request, page'in omurgasındaki guard'lardan en dıştakinden başlayarak geçer. Master'ın guard'ı yoktur. Account'un guard'ı giriş yapmamış bir okuyucuya /login'e 303 ile cevap verir. Admin'in guard'ı admin olmayan bir okuyucuya body'siz bir 403 ile cevap verir. Content fragment'i olan admin-users en son sorulur. Page ancak hiçbir guard cevap vermediğinde render edilir; cache'lenen bir page'de cache de ancak guard'lardan sonra okunur. Bir slot'a bağlanan menu fragment'i omurgada değildir, bu yüzden üzerindeki bir guard sorulmaz.</title>
<defs><marker id="pl-spine-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="40" y="6" width="140" height="24" rx="12" /><text x="110" y="22" text-anchor="middle">GET /admin/users</text>
<line class="d-flow" x1="110" y1="30" x2="110" y2="44" marker-end="url(#pl-spine-arrow)" />
<rect class="d-frame" x="20" y="48" width="180" height="34" rx="6" /><text x="110" y="62" text-anchor="middle">Master</text><text class="d-small" x="110" y="75" text-anchor="middle">1. layout · guard yok</text>
<line class="d-flow" x1="110" y1="82" x2="110" y2="96" marker-end="url(#pl-spine-arrow)" />
<rect class="d-a" x="20" y="100" width="180" height="34" rx="6" /><text class="d-on-a" x="110" y="114" text-anchor="middle">Account</text><text class="d-on-a d-small" x="110" y="127" text-anchor="middle">2. layout · RequireUser</text>
<line class="d-flow" x1="200" y1="117" x2="212" y2="117" marker-end="url(#pl-spine-arrow)" />
<rect class="d-chip" x="216" y="100" width="100" height="34" rx="6" /><text class="d-small" x="266" y="114" text-anchor="middle">giriş yapmamış</text><text class="d-small" x="266" y="127" text-anchor="middle">303 · /login</text>
<line class="d-flow" x1="110" y1="134" x2="110" y2="148" marker-end="url(#pl-spine-arrow)" />
<rect class="d-a" x="20" y="152" width="180" height="34" rx="6" /><text class="d-on-a" x="110" y="166" text-anchor="middle">Admin</text><text class="d-on-a d-small" x="110" y="179" text-anchor="middle">3. layout · requireRole</text>
<line class="d-flow" x1="200" y1="169" x2="212" y2="169" marker-end="url(#pl-spine-arrow)" />
<rect class="d-chip" x="216" y="152" width="100" height="34" rx="6" /><text class="d-small" x="266" y="166" text-anchor="middle">admin değil</text><text class="d-small" x="266" y="179" text-anchor="middle">403 · body'siz</text>
<line class="d-flow" x1="110" y1="186" x2="110" y2="200" marker-end="url(#pl-spine-arrow)" />
<rect class="d-frame" x="20" y="204" width="180" height="34" rx="6" /><text x="110" y="218" text-anchor="middle">admin-users</text><text class="d-small" x="110" y="231" text-anchor="middle">content · en son sorulur</text>
<g class="d-dropped"><rect class="d-frame" x="216" y="204" width="100" height="34" rx="6" /><text class="d-small" x="266" y="218" text-anchor="middle">menu · slot'ta</text><text class="d-small" x="266" y="231" text-anchor="middle">guard sorulmaz</text></g>
<line class="d-flow" x1="110" y1="238" x2="110" y2="252" marker-end="url(#pl-spine-arrow)" />
<rect class="d-c" x="20" y="256" width="180" height="24" rx="6" /><text class="d-on-c" x="110" y="272" text-anchor="middle">page render edilir</text>
<text class="d-small" x="160" y="298" text-anchor="middle">en dıştaki önce · ilk cevap karar verir</text>
</svg>
<figcaption><a href="#example-a-blog-with-a-private-area">Örnek: private bir bölümü olan bir blog</a> bölümündeki admin page'inin omurgası. Bir request önce layout'larının guard'larından, en dıştakinden başlayarak geçer, sonra content fragment'inkinden. Cevap veren ilk guard'ın cevabı gönderilir ve sonrasında hiçbir şey çalışmaz: giriş yapmamış bir okuyucu account layout'undan login page'ine bir redirect alır, giriş yapmış ama admin olmayan bir okuyucu da admin layout'undan body'siz bir 403 alır. Render'a ve cache'lenen bir page'de cache'e yalnızca hiçbir guard'ın cevap vermediği bir request ulaşır; cache guard'lardan sonra okunur. Admin menüsü gibi bir slot'a bağlanan bir fragment'teki guard sorulmaz.</figcaption>
</figure>

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
  her render hook'undan da önce çalışır: engellenen request page'e hiç
  ulaşmamıştır, bu yüzden render'ları izleyen plugin'lere haber verilmez.
- **Page'in kendi URL'sindeki action'lar.** Bir form bulunduğu page'e post eder.
  Okuyucunun göremediği bir page, form'unu da gönderemeyeceği bir page'dir. Guard,
  body okunmadan ve forgery kontrolünden önce çalışır.
- **Page'in fragment path'leri.**
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile kendi
  URL'sinde açılan bir fragment page'in bir kısmıdır. v0.34.0'dan beri
  fragment'in kendi guard'ından önce page'in guard'larından geçer. Öncesinde bir
  fragment path yalnızca fragment'inin guard'ını çalıştırırdı. Private bir
  page'deki fragment path de, o fragment bir guard taşımıyorsa public'ti.
- **Kendi URL'sinde register edilen bir action ve page'in not-found ile error
  page'leri kapsanmaz.** Aksi hâlde private bir error page, hatayla karşılaşan
  okuyucuyu redirect ederdi.

Guard'ın izin verdiği okuyucular page'in cache'ini paylaşır; bu cache sunucunun
kendi cache'idir. Sunucunun önündeki bir CDN ya da proxy guard çalıştırmaz. Bu
yüzden guard'lı bir page, stratejisi ne derse desin `Cache-Control: private,
no-cache` ile gönderilir.

v0.34.0'dan beri guard'ın kendi cevabı (giriş yapmaya yönlendiren bir redirect, bir
ret) `no-store` ile gönderilir. Böylece varsayılan olarak bir 404'ü ya da 308'i
saklayan bir CDN, bir okuyucuya verilen cevabı sonrakine vermez.

İçeriği okuyucudan okuyucuya değişen bir page ise guard değil kişiselleştirme
konusudur; bkz. [Caching](/docs/caching#render-strategies).
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
| `{slug}.md`, `post-{id}`, `v{version}.json` | Önünde ve arkasında bu metni taşıyan, arada en az bir karakter olan bir segment'le. Yakalanan, aradaki kısımdır (v0.41.0'dan beri): `/blog/hello.md` için `slug` = `hello` |
| `{rest...}` | Geriye kalan her şeyle, `rest` adıyla yakalanır. Yalnızca son segment olabilir ve etrafında metin olamaz |

Bir data handler yakalanan değeri `rc.Param("slug")` ya da `rc.PathParams["slug"]`
ile okur. Değerler percent-decode edilmiş olarak ve segment segment gelir.

v0.34.0'dan beri encode edilmiş bir slash (`%2F`) içeren bir segment 404 döner.
URL standardına göre `%2F` bir ayırıcı değildir: `/public%2Fsecret`, değeri
`public/secret` olan tek bir segment'tir.

Ama Go, path'i hiçbir middleware okumadan önce decode edip `r.URL.Path`'e yazar. Orada `/public%2Fsecret` artık
`/public/secret` olmuştur ve encode edilmiş slash gerçek bir slash'ten ayırt
edilemez. Yani router'ın tek segment gördüğü yerde middleware iki segment görür.
Tek segment'i eşleyen bir router (collage'ın eskiden yaptığı gibi), middleware'in
başka bir şey olarak değerlendirdiği bir request'i sunmuş olurdu. Bu yüzden collage
iki okumadan hiçbirini sunmaz.

Kendisi bir path olan bir değer, örneğin `guide/intro`, bir catch-all'da yer
almalıdır: `{rest...}`. `{{pageURL}}` ve `BuildPath`, tek bir segment'in
değerindeki `/`'ı reddeder.

Her seviyede önce static bir segment denenir. Ardından etrafında metin olan
placeholder'lar gelir, en özelinden başlayarak. Sonra düz bir `{param}`, en son da
`{rest...}` denenir. Eşleştirme backtracking ile yapılır. Böylece iki route da
eşleşebilecek olsa bile `/blog/archive`, `/blog/{slug}`'a karşı kazanır. `/blog` ve
`/blog/` aynı route'tur.

### Placeholder'ın etrafında metin

Bir yazı ile onun Markdown'ı ya da kategori başına bir feed yan yana durabilir:

```go
collage.NewPage("post").WithPath("en", "/blog/{slug}")            // /blog/hello
collage.NewDocument("post-md", "text/markdown; charset=utf-8").
	WithPath("en", "/blog/{slug}.md")                               // /blog/hello.md
collage.NewDocument("feed", "application/rss+xml").
	WithPath("en", "/feeds/{category}.xml")                         // /feeds/go.xml
```

Bir segment'in hangi route'a gittiği, route'ların kayıt sırasına hiçbir zaman
bağlı değildir. Aynı konumdaki iki placeholder üç durumdan birindedir:

| | Örnek | Sonuç |
| --- | --- | --- |
| Ayrık: hiçbir segment ikisine birden uymaz | `{slug}.md` ve `{slug}.json` | İkisi de kaydedilir, metin karar verir |
| İç içe: biri diğerinin daha özel hâlidir | `{slug}.min.md`, `{slug}.md`'nin içinde; ikisi de `{slug}`'ın içinde | İkisi de kaydedilir, daha özel olan kazanır |
| Kesişen: bir segment ikisine de uyar ama hiçbiri daha özel değildir | `a{x}` ve `{x}b`, ikisi de `aXb`'ye uyar | `ErrOverlappingPattern`, böyle bir segment'i örnek vererek |

Yani yukarıdaki route'larla `/blog/hello.md` document'e, `/blog/hello` sayfaya
gider. Static bir `/blog/index.md` ikisini de geçer. Metnin arasında hiçbir şey
olmayan `/blog/.md` ise sayfanındır ve `slug` = `.md` olur. Etrafında metin olan
bir placeholder `.` ya da `..` yakalamaz: `/blog/...md` de sayfaya gider,
document'e `..` slug'ını vermez. Link kurulurken yalnızca değer escape edilir:
`{{pageURL "post-md" "slug" "çay"}}`, `/blog/%C3%A7ay.md` olur. Export de
`blog/hello.md` dosyasını `blog/hello/index.html`'in yanına yazar.

Bir segment tek bir placeholder taşır. `{name}.{ext}` için `a.b.c`'nin tek bir
cevabı yoktur, bu yüzden `{name}.md` ve `{name}.json`'u iki ayrı route olarak
kaydedin. Regex de yoktur. Bir değeri handler'ı kontrol eder ve tanımadığı bir
değer için `collage.ErrNotFound` döner. Bu sayede her çakışma route kaydedilirken
kesin olarak belirlenebilir.

Pattern'lerdeki hatalar request anında sürpriz olarak çıkmaz. Register sırasında
hata verirler:

- Bir pattern `/` ile başlamalıdır. Boş segment ve boş placeholder adı içeremez,
  catch-all'u da yalnızca en sona koyabilir. Aksi hâlde `ErrInvalidPath` ya da
  `ErrInvalidPattern` alırsınız.
- Bir segment en fazla bir placeholder taşır, catch-all'un etrafında da metin
  olamaz: `/{name}.{ext}` ve `/files/{path...}.md` `ErrInvalidPattern` verir.
  (v0.11.0'dan v0.40.0'a kadar placeholder'ın etrafındaki her metin
  reddediliyordu. v0.41.0'dan beri `/feeds/{category}.xml` ve `/post-{id}` birer
  route'tur.)
- Etrafında metin olan iki placeholder kesişirse, örneğin `/a{x}` ve `/{x}b`,
  `ErrOverlappingPattern` alırsınız.
- Bir locale'de aynı path'te iki route olursa `ErrDuplicateRoute` alırsınız. Bir
  page ile bir [document](/docs/documents)'ın çakışması da buna dahildir, çünkü
  ikisi aynı ağacı kullanır.
- Aynı konumda iki farklı parametre adı, örneğin `/blog/{slug}` ve
  `/blog/{id}/edit` ya da `{slug}.md` ve `{id}.md`, `ErrAmbiguousParameterName`
  verir.

Bir page `GET` ve `HEAD` request'lerine cevap verir. `OPTIONS` request'ine de
`204` ile cevap verir. Bu response'un `Allow` header'ı, URL'nin kabul ettiği
metotları listeler. Diğer bütün metotlar aynı `Allow` header'ıyla 405 alır. Tek
istisna, page'in o metot için bir [action](/docs/forms-and-actions)'ı olmasıdır.
Bir form, üzerinde bulunduğu page'e bu sayede post eder.

Locale'ler, `/tr/hakkinda` gibi locale prefix'leri, `/en/about`'u `/about`'a
gönderen redirect ve page adlarından link üretmek
[Link'ler ve locale'ler](/docs/links-and-locales) sayfasında anlatılıyor.

Bir page ayrıca eski URL'lerden redirect'ler de taşıyabilir. Bunun için
`WithRedirect(from, to, status)` ve `WithPermanentRedirect(from, to)` kullanılır.
v0.34.0'dan beri:

- bir redirect, request'in query string'ini hedefine taşır;
- yakaladığı bir değer, düştüğü yere göre escape edilir: bir path segment'i ya da
  `?` işaretinden sonraki bir query değeri. Böylece `/login?next={slug}`'a ikinci
  bir `next` verilemez;
- siteden çıkacak bir hedef (`//host` ya da `/\host`) register sırasında
  `ErrInvalidPath` olur.

v0.52.0'dan beri iki path'ten birinde satır sonu gibi bir control karakteri olması
`ErrInvalidRedirect` olur. Static export da yazdığı bir page'in üstüne düşen bir
redirect'te başarısız olur; bkz. [Static export](/docs/static-export#what-fails).

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
olur. Template'lerden ve sabit değerlerden (`collage.Value`, `WithTitle`) oluşan bir page
herkes için aynı render edilir, bu yüzden söylemeye gerek kalmadan cache'lenir.
Handler ise request'i, bir cookie'yi ya da saati okuyabilir ve dışarıdan bu
anlaşılamaz. Bu yüzden handler'ı olan bir page, `Static()` ya da
`Incremental(ttl)` diyene kadar her request'te render edilir. Kuralın tamamı
[Caching](/docs/caching#a-page-that-declares-none) sayfasındadır.

`Incremental` pozitif bir TTL ister (`ErrInvalidTTL`). `collage export`'un dosyaya
yazabildiği page'ler de static ve incremental page'lerdir. Dynamic bir page her
request'te render edilmek için vardır, bu yüzden
[atlanır](/docs/static-export#what-is-skipped).

Cache yalnızca uygulama onu açtığında devreye girer (`Config.Cache.Enabled`,
scaffold'da açıktır). Development'ta ise hiçbir zaman devreye girmez. Orada
cache'lenmiş bir page, az önce düzenlediğiniz template'i gizlerdi.

Cache'lenen bir page'i iki builder metodu daha etkiler. `WithDependencyTags(tags...)`,
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
		pages.Home(),
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

Register işlemi kendisine verilen page'i değiştirir. 3. adım page'in en dıştaki
layout'unu, onun bu page'e ait ve bağlanmış bir kopyasıyla değiştirir. Render
edilen şey, içinde zincirin geri kalanı ve içerik bulunan bu kopyadır. Aynı
constructor tekrar çağrılarak oluşturulan bir page ise farklı bir değerdir. Onun
layout'larının `content` slot'ları boştur. v0.60.0'dan beri aynı builder'da ikinci
bir `Build()` da öyledir: her çağrı bağımsız bir kopya döner ve sonradan builder'a
yapılan `WithPath` ikisine de ulaşmaz. Bir kez build edin ve `*Page`'i saklayın.

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
  handler'ın döndüğü veride ya da bir `collage.Key` altında yer almalıdır.
  Ayrıntılar için [Data handler'lar](/docs/data-handlers#the-render-context)
  sayfasına bakın.

`app.Page(name)` ve `app.Pages()`, register edilmiş page'lerin kopyalarını döner.
Bu kopyalar incelemek içindir: örneğin her page'in path'lerini listeleyen bir
sitemap ya da bir stratejiyi kontrol eden bir test. Kopya, register edilmiş değer
değildir. Bu yüzden register edilmiş bir page beklenen yere kopya vermeyin.

## Örnek: private bir bölümü olan bir blog

Çok kullanıcılı bir blogda üç tür page vardır. Yazılar public'tir. Giriş yapmış her
yazarın taslaklarını listeleyen bir hesap page'i vardır. Admin'lerin ayrıca
kullanıcıları listeleyen bir page'i vardır. Üçü aynı iskeleti paylaşır; private
olanlar, kimin girebileceğini söyleyen layout'ların içinde durur:

| Page | Path | Layout'lar, en dıştakinden başlayarak | Strateji |
| --- | --- | --- | --- |
| `post` | `/blog/{slug}` | `Master` | `Incremental(10 * time.Minute)` |
| `account` | `/account` | `Master`, `Account` | tanımlanmamış: data handler'ı olduğu için dynamic |
| `admin-users` | `/admin/users` | `Master`, `Account`, `Admin` | tanımlanmamış: data handler'ı olduğu için dynamic |
| `not-found`, `error` | path yok | `Master` | `Dynamic()` |

Guard'ları layout'lar taşır. `Account`,
[session plugin'inin](/docs/plugins#elagohtsession) `RequireUser`'ını kullanır. Bu
guard, giriş yapmamış bir okuyucuyu gitmek istediği yeri `next` içinde taşıyarak
`/login`'e gönderir. `Admin` bir rolü kontrol eder. Rol, bu uygulamanın kendi login
action'ının yazdığı `"role"` session key'inde durur; plugin yalnızca kullanıcı
key'ini tanımlar.

```go
// Master is the site's shell: <head>, header, footer.
func Master() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("My blog").
		Build()
}

// Account is the signed-in area. Its guard sends everyone else to log in.
func Account() *collage.Fragment {
	return collage.NewFragment("account-layout", "layouts/account.html").
		WithGuard(session.RequireUser("/login")).
		Build()
}

// Admin sits inside Account and lets only admins through.
func Admin() *collage.Fragment {
	menu := collage.NewFragment("admin-menu", "fragments/admin-menu.html").Build()

	return collage.NewFragment("admin-layout", "layouts/admin.html").
		WithGuard(requireRole("admin")).
		WithSlotFragment("menu", menu).
		Build()
}

// requireRole lets through a reader whose session holds role under "role",
// which the login action sets. Anyone else gets a bare 403.
func requireRole(role string) collage.GuardFunc {
	return func(ctx context.Context, _ *http.Request) (*collage.GuardDecision, error) {
		if session.FromContext(ctx).Get("role") == role {
			return nil, nil
		}
		return &collage.GuardDecision{Status: http.StatusForbidden}, nil
	}
}
```

Admin layout'u iki slot'unu da çağırır: yukarıda bağlanan `menu`'yü ve register
işleminin doldurduğu `content`'i. `layouts/account.html` de aynıdır, yalnızca
`content`'i çağırır.

```html
<!-- templates/layouts/admin.html -->
<div class="admin">
  <nav>{{slot "menu"}}</nav>
  <main>{{slot "content"}}</main>
</div>
```

Bir page'in public ya da private olması yalnızca listelediği layout'lara bağlıdır:

```go
// Post is public and cached; an unknown slug is a 404.
func Post() *collage.Page {
	content := collage.NewFragment("post", "pages/post.html").
		WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) (store.Post, []string, error) {
			slug := rc.Param("slug")
			// PostBySlug wraps collage.ErrNotFound for a slug it does not know.
			post, err := store.PostBySlug(ctx, slug)
			return post, []string{"post:" + slug}, err
		})).
		Required().
		Build()

	return collage.NewPage("post").
		WithLayouts(layouts.Master()).
		WithContent(content).
		WithPath("en", "/blog/{slug}").
		Incremental(10 * time.Minute).
		Build()
}

// Account lists the signed-in reader's drafts.
func Account() *collage.Page {
	content := collage.NewFragment("account", "pages/account.html").
		WithData(collage.Load(func(ctx context.Context, rc *collage.RenderContext) ([]store.Draft, error) {
			return store.Drafts(ctx, session.Get(rc).Get(session.UserKey))
		})).
		Build()

	return collage.NewPage("account").
		WithLayouts(layouts.Master(), layouts.Account()).
		WithContent(content).
		WithPath("en", "/account").
		Build()
}

// AdminUsers is for admins only: it sits in both guarded layouts.
func AdminUsers() *collage.Page {
	content := collage.NewFragment("admin-users", "pages/admin-users.html").
		WithData(collage.Load(func(ctx context.Context, _ *collage.RenderContext) ([]store.User, error) {
			return store.Users(ctx)
		})).
		Build()

	return collage.NewPage("admin-users").
		WithLayouts(layouts.Master(), layouts.Account(), layouts.Admin()).
		WithContent(content).
		WithPath("en", "/admin/users").
		Build()
}

// NotFound and Error wrap the shell alone, so they look like the public site.
func NotFound() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").Build()

	return collage.NewPage("not-found").
		WithLayouts(layouts.Master()).
		WithContent(content).
		Dynamic().
		Build()
}

func Error() *collage.Page {
	content := collage.NewFragment("error-content", "pages/500.html").Build()

	return collage.NewPage("error").
		WithLayouts(layouts.Master()).
		WithContent(content).
		Dynamic().
		Build()
}
```

```go
func register(app *collage.App) error {
	if err := app.Register(
		pages.Post(),
		pages.Account(),
		pages.AdminUsers(),
	); err != nil {
		return err
	}
	if err := app.RegisterNotFoundPage(pages.NotFound()); err != nil {
		return err
	}
	return app.RegisterErrorPage(pages.Error())
}
```

Her okuyucunun aldığı:

| Request | Okuyucu | Cevap |
| --- | --- | --- |
| `/blog/hello` | herkes | Yazı; cache açıksa on dakika geçene kadar cache'ten |
| `/blog/no-such-post` | herkes | `404` ve sitenin not-found page'i, iskeletin içinde |
| `/account` | giriş yapmamış | `/login?next=%2Faccount`'a `303` |
| `/admin/users` | giriş yapmamış | `/login?next=%2Fadmin%2Fusers`'a `303`: önce account guard'ı cevap verir, rol hiç kontrol edilmez |
| `/admin/users` | giriş yapmış, admin değil | Body'siz bir `403` |
| `/admin/users` | bir admin | Page |

Yukarıdaki kurallardan birkaç sonuç çıkar:

- **Private path listesi yoktur.** `/admin` altındaki yeni bir page, path'inin
  nasıl başladığı yüzünden değil, `layouts.Admin()`'i listelediği için private'tır.
  Bunu unutan bir page public'tir. Bu yüzden `collage inspect`'e bakmaya değer:
  her page için guard taşıyan fragment'leri listeler.
- **Admin menüsünün guard'ı sorulmazdı.** Menü bir slot'a bağlıdır, omurgada
  değildir. Bölümü korumak admin layout'unun işidir.
- **Private page'ler ortak cache'lere girmez.** Account ve admin page'leri
  `Cache-Control: private, no-cache` ile, `303` ve `403` ise `no-store` ile
  gönderilir. Static export da iki page'i yazmaz.
- **Error page'ler herkes için render edilir.** Yalnızca `Master()`'ı sararlar ve
  başarısız olan page'in guard'ları onlar için sorulmaz. Böylece hatayla karşılaşan
  bir okuyucu, hatayı görmek yerine giriş yapmaya yönlendirilmez.
