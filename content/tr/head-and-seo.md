---
description: Title'ları, meta tag'leri, stylesheet'leri ve structured data'yı onları bilen fragment tanımlar, layout yerleştirir.
reference: RenderContext, FragmentBuilder.WithTitle, Effect, ErrNoPathInLocale
---

# Head ve SEO

Bir page'in title'ını bilen fragment, `<head>`'i yazan fragment'le nadiren aynıdır. Post
kendi title'ını ve özetini bilir. Gallery, `gallery.css`'e ihtiyacı olduğunu bilir.
Layout ise `<head>`'i bunlar render edilmeden önce yazmıştır; ikisini de bilmez.

collage bu sorunu **hoisting** ile çözer. Fragment, page içinde nerede durursa
dursun head'e neyin gireceğini *tanımlar*. Tanımların nereye yerleşeceğini
ise layout belirler.

## Nereye yerleşir

Layout, `<head>`'inin içine bir marker koyar:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  {{hoist "head"}}
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
</head>
<body>
  {{slot "content"}}
</body>
</html>
```

`{{hoist "head"}}` içerik yazmaz, bir marker yazar. Çünkü o noktada altındaki hiçbir
şey henüz render edilmemiştir.

Page'in tamamı bittiğinde collage bu marker'ı, ağacın herhangi bir yerinde `"head"`
alanı için tanımlanmış her şeyle değiştirir. Bu iş tek geçişte yapılır ve konumu
yine layout belirler.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 300" role="img" aria-labelledby="hs-hoist-title">
<title id="hs-hoist-title">Bir blog post page'inde hoisting. Layout, The Wire title'ını tanımlar. İçindeki post, data handler'ından kendi title'ını ve bir description meta'sını tanımlar. Post'un içindeki iki gallery fragment'inin her biri gallery.css ister. Tanımlar her key için tek bir kayıt olarak toplanır: en içteki tanım kazandığı için title post'unkidir, description post'unkidir, iki kez istenen gallery.css bir kez link'lenir. Page'in tamamı render edildikten sonra bu kayıtlar layout'un hoist head marker'ının yerine yazılır.</title>
<defs><marker id="hs-hoist-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="4" width="312" height="140" rx="8" />
<text x="14" y="23">layout</text>
<g class="d-dropped"><rect class="d-chip" x="196" y="10" width="110" height="18" rx="9" /><text class="d-small" x="251" y="22" text-anchor="middle">title: The Wire</text></g>
<rect class="d-frame d-frame-a" x="14" y="36" width="292" height="100" rx="6" />
<text class="d-slot-text-a" x="24" y="53">post</text><text class="d-small" x="60" y="53">data handler</text>
<rect class="d-chip-hit" x="24" y="60" width="120" height="18" rx="9" /><text class="d-hit-text d-small" x="84" y="72" text-anchor="middle">title: post'unki</text>
<rect class="d-chip-hit" x="152" y="60" width="144" height="18" rx="9" /><text class="d-hit-text d-small" x="224" y="72" text-anchor="middle">meta:description</text>
<rect class="d-c" x="24" y="86" width="130" height="42" rx="5" /><text class="d-on-c" x="89" y="103" text-anchor="middle">gallery</text><text class="d-on-c d-small" x="89" y="119" text-anchor="middle">gallery.css</text>
<rect class="d-c" x="166" y="86" width="130" height="42" rx="5" /><text class="d-on-c" x="231" y="103" text-anchor="middle">gallery</text><text class="d-on-c d-small" x="231" y="119" text-anchor="middle">gallery.css</text>
<line class="d-flow" x1="160" y1="144" x2="160" y2="166" marker-end="url(#hs-hoist-arrow)" />
<text class="d-small" x="168" y="159">her key için bir kayıt</text>
<rect class="d-frame" x="4" y="170" width="312" height="70" rx="6" />
<text class="d-small" x="14" y="190">title</text><text class="d-small" x="186" y="190">post'unki (en içteki)</text>
<text class="d-small" x="14" y="208">meta:description</text><text class="d-small" x="186" y="208">post'unki</text>
<text class="d-small" x="14" y="226">stylesheet:/static/gallery.css</text><text class="d-small" x="186" y="226">iki istek, tek link</text>
<line class="d-flow" x1="160" y1="240" x2="160" y2="262" marker-end="url(#hs-hoist-arrow)" />
<text class="d-small" x="168" y="255">page render edilince</text>
<rect class="d-cmd" x="40" y="266" width="240" height="28" rx="6" /><text class="d-cmd-text" x="160" y="284" text-anchor="middle">hoist "head" marker'ı dolar</text>
</svg>
<figcaption>Her fragment bildiğini, bulunduğu yerden tanımlar: layout varsayılan bir title'ı, post kendi title'ını ve bir description'ı, her gallery ihtiyaç duyduğu stylesheet'i. Tanımlar her key için tek bir kayıt olarak toplanır. Bir key'in en içteki tanımı kazandığı için post'un title'ı layout'unkinin yerine geçer, <code>gallery.css</code> için gelen iki istek de tek bir link olur. Page'in tamamı render edildikten sonra bu kayıtlar, her key'in ilk tanımının sırasıyla layout'un <code>{{hoist "head"}}</code> marker'ının yerine yazılır.</figcaption>
</figure>

Marker yoksa tanımlar hiçbir yere yerleşmez. Bir title ya da bir plugin'in
çıktısı eksikse ilk kontrol etmeniz gereken şey budur.

## Handler olmadan bir title

Program başlarken bilinen bir title, örneğin layout'taki site adı ya da "Hakkında"
page'inin adı, hiç kod gerektirmez. `WithTitle` onu fragment üzerinde tanımlar:

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithTitle("The Wire").
	Build()
```

Bu, `rc.HoistTitle`'ın yaptığı tanımın aynısıdır. Bu yüzden
[aşağıdaki](#keys-and-the-innermost-wins) kurallara uyar: en içteki kazanır ve aynı
fragment'in data handler'ının hoist ettiği bir title onun yerine geçer. Handler'ın
aksine, strateji tanımlamayan bir page'i static bırakır. Bkz.
[Caching](/docs/caching#a-page-that-declares-none).

## Data handler'dan tanımlamak

Geri kalan her şey, yani içerikten gelen bir title, bir description ya da bir
stylesheet, `RenderContext` üzerindeki helper'larla tanımlanır. Bunları bir data
handler'dan çağırın:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}

	rc.HoistTitle(post.Title + " — The Wire")
	rc.HoistMeta("description", post.Summary)
	rc.HoistProperty("og:title", post.Title)
	rc.HoistProperty("og:image", post.CoverURL)
	rc.HoistLink("canonical", siteOrigin+"/blog/"+post.Slug)
	if err := rc.HoistStylesheet("/static/post.css"); err != nil {
		return nil, nil, err
	}

	return post, []string{"post:" + post.Slug}, nil
}
```

| Çağrı | Ürettiği | Key |
| --- | --- | --- |
| `rc.HoistTitle(text)` | `<title>text</title>` | `title` |
| `rc.HoistMeta(name, content)` | `<meta name="…" content="…">` | `meta:<name>` |
| `rc.HoistProperty(property, content)` | `<meta property="…" content="…">` (Open Graph) | `property:<property>` |
| `rc.HoistLink(rel, href)` | `<link rel="…" href="…">` | `link:<rel>` |
| `rc.HoistAlternate(hreflang, href)` | `<link rel="alternate" hreflang="…" href="…">` (v0.10.0'dan beri) | `alternate:<hreflang>` |
| `rc.HoistStylesheet(path)` | Dosyanın [content-addressed URL'siyle](/docs/assets) `<link rel="stylesheet" href="…">` | `stylesheet:<path>` |

Her helper kendisine verilen değeri escape eder. Bu yüzden bir post'un adından
oluşturulan title, post'un adı ne olursa olsun güvenlidir.

`HoistStylesheet`, dosyanın mount edilmiş path'ini alır. O dosya hiçbir mount'ta
yoksa hata döndürür. Var olmayan bir stylesheet'e link veren page bozuk bir page'dir
ve bunu açıkça söylemesi gerekir.

Bu helper'ları senkron olarak, data handler'ın kendi goroutine'inden çağırın.
Handler'ın başlattığı başka bir goroutine'den yapılan tanımın page içinde bir
konumu yoktur.

### Template'ten

Bir fragment'in template'i stylesheet'i kendisi de isteyebilir:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

`{{stylesheet}}` bulunduğu yerde hiçbir şey render etmez, `HoistStylesheet` ile aynı
işi yapar. Böylece gallery'nin CSS'i, gallery hangi page'deyse onunla birlikte oraya
gider.

## Key'ler ve en içtekinin kazanması

Her tanımın bir key'i vardır. Neyin aynı şey sayılacağını bu key belirler:

- **Farklı key'lerin hepsi görünür.** Her biri, page sırasındaki ilk
  tanımının yerinde durur.
- **Aynı key iki kez tanımlanırsa en içteki tanım kalır.**

Bu tek kural iki işi birden görür. Stylesheet'in key'i path'idir. Bu yüzden
`gallery.css` isteyen beş fragment tek bir `<link>` üretir, farklı stylesheet'ler ise
birikir. Title'ın ise tek bir key'i vardır. Bu yüzden daha spesifik bir fragment'in
title'ı, daha genel olanın title'ının yerine geçer. Bu da varsayılan değerleri
kolaylaştırır. Varsayılanları layout'ta tanımlayın, her page istediğinde onları
override etsin:

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithTitle("The Wire").
	WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistMeta("description", "News about the sea, and the people who live beside it.")
		return nil
	})).
	Build()
```

Bu layout'un içindeki bir post kendi title'ını ve description'ını tanımlar.
Page'de görünenler de bunlar olur. Hiçbir şey tanımlamayan bir page layout'unkileri
korur.

Layout'un template'ine ayrıca elle bir `<title>` yazmayın. Yazarsanız o title,
hoist edilen title'ın yanında durur. İki title'ı olan bir page'de tarayıcı bunlardan
birini yok sayar. `collage new`'un oluşturduğu layout da sitenin adını `WithTitle`
ile verir. Böylece bir page'in kendi title'ı onun yerine geçer.

Layout'taki bir handler her page'deki bir handler demektir. Bu yüzden yukarıdaki
description, strateji tanımlamayan her page'i dynamic yapar. Page'lerinin
cache'lenmesini isteyen bir site, page'lerinde `Static()` ya da `Incremental(ttl)`
belirtir ya da layout'ta yalnızca `WithTitle` bırakır.

Sürpriz olmasınlar diye üç ayrıntıyı belirtelim:

- **Konumu kazanan tanım değil, ilk tanım belirler.** Head, iç taraftaki
  bir fragment'in title'ı override edip etmemesine göre kendini yeniden sıralamaz.
- **Aynı derinlikte, sonra tanımlayan fragment kazanır.** Aynı key'i tanımlayan
  iki kardeş fragment gerçek bir çakışmadır. Bu çakışma, fragment'lerin page içindeki
  sırasıyla çözülür. Hangi data handler'ın önce bittiği belirleyici değildir, öyle
  olsaydı head request'ten request'e değişirdi.
- **Sırayı saat değil, page belirler.** Kardeş data handler'lar eşzamanlı çalışır.
  Ama bir key, önce onu ilk tanımlayan fragment'in page içindeki yerine, sonra o
  fragment'in tanımlama sırasına göre yerleştirilir. Böylece head, birbirini
  override eden stylesheet'ler de dahil olmak üzere her render'da aynı olur.
  (v0.12.0'dan önce bir key, ilk tanımı nereye denk gelirse oraya
  yerleşiyordu ve kardeş fragment'lerin key'leri request'ler arasında yer
  değiştirebiliyordu.)

## Geri kalan her şey: rc.Hoist

Helper'ların hepsi tek bir method üzerine kuruludur:

```go
func (rc *RenderContext) Hoist(area, key string, html template.HTML)
```

Bu method `html`'i `key` altında, `area`'ya **tam yazıldığı gibi** ekler. Zaten var
olma sebebi budur: helper'ların kapsamadığı markup. Bu da escape işinin size düştüğü
anlamına gelir. Markup'ı kontrolünüzdeki değerlerden oluşturun ya da değerleri escape
edin:

```go
rc.Hoist("head", "preload:hero", template.HTML(
	`<link rel="preload" as="image" href="`+html.EscapeString(post.CoverURL)+`">`))
```

Aynı key kuralları burada da geçerlidir. Bu yüzden eklediğiniz şeyin ne olduğunu
anlatan bir key seçin. Bir kez görünmesi gereken her şey için tek bir key kullanın.
Biriken şeyler için ise her değere ayrı bir key verin.

Alanın adı serbesttir ve bir layout'ta birden fazla marker bulunabilir.
Örneğin fragment'lerin ihtiyaç duyduğu script'ler için `</body>`'den önce bir `{{hoist
"scripts"}}` koyabilirsiniz:

```go
rc.Hoist("scripts", "script:map", template.HTML(`<script src="`+mapJS+`" defer></script>`))
```

## Canonical ve alternate link'ler

Canonical link, aynı içeriğe birden fazla URL ulaştığında (örneğin tracking
query'siyle ve onsuz) arama motorlarına hangisinin asıl URL olduğunu söyler. Bu
link'e mutlak bir URL verin. collage origin değil path üretir. Bu yüzden sitenizin
origin'ini kendi config'inizde tutun ve ikisini birleştirin:

```go
rc.HoistLink("canonical", siteOrigin+rc.Request.URL.Path)
```

Birden fazla dildeki bir site, her çevirinin nerede olduğunu da belirtmelidir. Bunun
için her dile bir `<link rel="alternate" hreflang="…">` gerekir. `HoistLink` her
`rel` için tek bir link tuttuğundan bunu ifade edemez. `rc.HoistAlternate` ise her
link'in key'ini diline göre belirler, böylece her çeviri bir kez görünür. Bu link'leri
layout'tan, o an hangi page render ediliyorsa onun için,
[`app.URL`](/docs/links-and-locales#links-from-go) ile tanımlayın:

```go
func Layout(app *collage.App) *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
			for _, locale := range []string{"en", "tr"} {
				href, err := app.URL(rc.Page.Name, locale, rc.PathParams)
				if errors.Is(err, collage.ErrNoPathInLocale) {
					continue // not translated into this one
				}
				if err != nil {
					return err
				}
				rc.HoistAlternate(locale, siteOrigin+href)
			}
			return nil
		})).
		Build()
}
```

Bazı page'lerin çevirisi farklı bir slug kullanır ve bunu yalnızca o page'in içeriği
bilir. Böyle bir page, o dil için kendi `rc.HoistAlternate` çağrısını yapar. Daha
içte olduğu için de layout'un tanımının yerine geçer.

Handler'ı olmayan bir layout için template'teki karşılığı
[`localeURL`](/docs/links-and-locales#a-language-switcher) fonksiyonudur. Böyle bir
layout, strateji tanımlamayan page'leri static bırakır. `localeURL`, page'in var
olmadığı bir dil için boş döner:

```html
{{with localeURL "tr"}}<link rel="alternate" hreflang="tr" href="https://thewire.example{{.}}">{{end}}
```

## Structured data

Arama motorları schema.org structured data'sını head'deki JSON-LD bloklarından okur.
[`elagoht/jsonld` plugin'i](/docs/plugins#elagohtjsonld) bu blokları sizin yerinize
yazar. Bunu elle oluşturulmuş JSON'dan değil, tipli değerlerden yapar:

```go
import "github.com/Elagoht/collage-jsonld"

func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	jsonld.Emit(rc, jsonld.BlogPosting{
		Headline:      post.Title,
		Description:   post.Summary,
		DatePublished: post.PublishedAt,
		AuthorName:    post.Author,
	})
	return post, []string{"post:" + post.Slug}, nil
}
```

`jsonld.Emit`, `"head"` alanına hoist eder. Bu yüzden bu sayfada anlatılan diğer
her şeyle aynı marker'a ihtiyaç duyar ve aynı kurallara uyar: her schema.org tipi ve
node'un `Key`'i için bir key vardır ve en içteki kazanır.

Plugin'i register ettiğinizde (`Config.Plugins` içinde
`jsonld.New(jsonld.Options{})`), plugin bir site adıyla yapılandırılmışsa site
genelinde bir `WebSite` node'u ekler. Page'e
özel veri ise her zaman data handler'larınızdan gelir. Çünkü page'in ne hakkında
olduğunu yalnızca onlar bilir.

Crawler'lara ve arama motorlarına söylenenlerin geri kalanını altı plugin daha
karşılar. İlk üçü, bir crawler'ın head dışında okuduklarıyla ilgilidir:

| Plugin | Ne yapar |
| --- | --- |
| [elagoht/sitemap](/docs/plugins#elagohtsitemap) | Register ettiğiniz page'lerden `/sitemap.xml`'i sunar |
| [elagoht/robots](/docs/plugins#elagohtrobots) | `/robots.txt`'yi sunar |
| [elagoht/feed](/docs/plugins#elagohtfeed) | RSS ve Atom feed'leri sunar ve bunları her page'in head'inde duyurur |
| [elagoht/meta](/docs/plugins#elagohtmeta) | Bu sayfada anlatılan Open Graph ve Twitter tag'lerini, canonical URL'yi ve `hreflang` link'lerini site genelindeki varsayılanlardan ve her page'in söylediklerinden yazar |
| [elagoht/redirects](/docs/plugins#elagohtredirects) | Bir taşımadan kalan eski adresleri bir dosyadan okuyup sunar ve bunları static build'e aktarır |
| [elagoht/indexnow](/docs/plugins#elagohtindexnow) | Bir invalidation'ın hangi URL'leri değiştirdiğini arama motorlarına bildirir |

## Hoisting ve collage'ın geri kalanı

- **Cache'lenen page'ler head'lerini korur.** Marker, page saklanmadan önce
  değiştirilir. Bu yüzden cache'lenmiş bir page, tanımlanmış her şeyle birlikte
  sunulur.
- **Kendi URL'sinde sunulan bir fragment'in head'i yoktur.**
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile sunulan
  bir fragment'in layout'u yoktur. v0.18.0'dan beri marker koymadığı bir alana
  hoist ettiği şeyler markup'ından önce, her öğe için etkisiz bir
  `<template data-collage-hoist>` olarak gelir. Bir script bunları page'e ekleyebilir.
  Bkz. [Hoist ettikleri](/docs/forms-and-actions#what-it-hoists).
- **Hoist işlemini `rc.Page` üzerinden değil, data handler'lardan yapın.** `rc.Page`
  register edilmiş page'dir ve her request tarafından paylaşılır. Ona yazmak data
  race'e yol açar. Request'e göre değişen her şey `rc` üzerinden tanımlanır.

## Örnek: bir katalog ürün page'i

Bir mağazanın `/products/{id}` adresindeki ürün page'i, yani
[Fragment'ler ve slot'lar](/docs/fragments-and-slots#example-a-product-page) sayfasında
kurulanın küçük bir sürümü, bir arama motorunun ve bir link preview'ının
okuyabileceği bir head'e ihtiyaç duyar. Head'deki her şeyi, onu bilen fragment
tanımlar:

| Head'de | Tanımlayan | Key |
| --- | --- | --- |
| Title, `Brass deck lamp — The Shop` | Ürünün handler'ı, layout'un `WithTitle`'ının yerine | `title` |
| Description | Ürünün handler'ı, layout'unkinin yerine | `meta:description` |
| `og:title`, `og:description`, `og:image` | Ürünün handler'ı | `property:og:…` |
| Canonical link | Ürünün handler'ı | `link:canonical` |
| Yalnızca bu page'in ihtiyaç duyduğu `product.css` | Ürünün handler'ı, `HoistStylesheet` ile | `stylesheet:/static/product.css` |
| `stock.css` | Stok kutusunun template'i, `{{stylesheet}}` ile | `stylesheet:/static/stock.css` |

Layout'un template'i, [bu sayfanın başındaki](#where-it-lands) template'tir. Layout,
kendisi hakkında hiçbir şey söylemeyen her page için mağazanın adını ve kısa bir
açıklamasını verir:

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithTitle("The Shop").
	WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistMeta("description", "Lamps, rope and brass fittings for small boats.")
		return nil
	})).
	Build()
```

Ürünün handler'ı, hangi ürün olduğunu öğrenir öğrenmez geri kalanını tanımlar.
Örnekteki catalog ve inventory paketleri sizin kendi veri katmanınızı temsil eder:

```go
const siteOrigin = "https://shop.example"

type product struct {
	ID       string
	Name     string
	Price    string
	Summary  string
	ImageURL string
}

type stockLevel struct {
	Count int
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (product, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Product(ctx, id)
	if errors.Is(err, catalog.ErrNoProduct) {
		return product{}, nil, fmt.Errorf("shop: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return product{}, nil, err
	}

	rc.HoistTitle(p.Name + " — The Shop")
	rc.HoistMeta("description", p.Summary)
	rc.HoistProperty("og:title", p.Name)
	rc.HoistProperty("og:description", p.Summary)
	rc.HoistProperty("og:image", p.ImageURL)
	rc.HoistLink("canonical", siteOrigin+"/products/"+id)
	if err := rc.HoistStylesheet("/static/product.css"); err != nil {
		return product{}, nil, err
	}

	return product{ID: p.ID, Name: p.Name, Price: p.Price, Summary: p.Summary, ImageURL: p.ImageURL},
		[]string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (stockLevel, error) {
	n, err := inventory.Count(ctx, rc.Param("id"))
	if err != nil {
		return stockLevel{}, err
	}
	return stockLevel{Count: n}, nil
}
```

Stok kutusu kendi stylesheet'ini kendisi ister. Böylece hangi page'de render
edilirse edilsin stylesheet'i yanında olur:

```html
<!-- templates/fragments/stock.html -->
{{stylesheet "/static/stock.css"}}
<p class="stock">{{.Count}} in stock</p>
```

Page, sayıyı güncel tutan bir script için kutuyu kendi URL'sinde açar:

```go
stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

content := collage.NewFragment("product", "pages/product.html").
	WithData(collage.LoadTagged(loadProduct)).
	WithSlotFragment("stock", stock).
	Required().
	Build()

page := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(content).
	WithPath("en", "/products/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	Build()
```

### Page'in head'inde ne var

`/products/42` için gelen bir request 200 ile cevaplanır. Head'i aşağıdaki gibidir.
Burada her element ayrı bir satırda; collage ise hoist edilenleri yan yana yazar:

```html
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Brass deck lamp — The Shop</title>
  <meta name="description" content="A lamp for a boat&#39;s chart table.">
  <meta property="og:title" content="Brass deck lamp">
  <meta property="og:description" content="A lamp for a boat&#39;s chart table.">
  <meta property="og:image" content="https://shop.example/img/lamp.jpg">
  <link rel="canonical" href="https://shop.example/products/42">
  <link rel="stylesheet" href="/static/product.ab1d57cf7f5ec545.css">
  <link rel="stylesheet" href="/static/stock.1d101be0b97db6e6.css">
  <link rel="stylesheet" href="/static/app.2708d73bf31c36cd.css">
</head>
```

- **Hoist edilenler arasında title ve description en başta durur.** Layout'un onları
  tanımladığı yerdedirler ama ürünün değerlerini taşırlar: konum ilk tanımın, değer
  en içteki tanımındır.
- **Özetteki kesme işareti escape edilir.** Her helper kendisine verileni escape
  eder.
- **İki stylesheet de content-addressed URL'leriyle link'lenir.** Page onları hangi
  sırayla tanımladıysa o sırayla, önce ürününki, sonra stok kutusununki gelir.
  `app.css` onlardan sonra gelir, çünkü layout'un template'i onu marker'ın altına
  yazar.

### Stok kutusunun URL'si ne döner

`GET /products/42/stock` stok kutusunu tek başına render eder. Etrafında layout
yoktur, `"head"` için bir marker da yoktur. Hoist ettikleri,
[Hoist ettikleri](/docs/forms-and-actions#what-it-hoists) bölümünde anlatıldığı gibi
markup'ından önce gelir:

```html
<template data-collage-hoist="head" data-collage-key="stylesheet:/static/stock.css"><link rel="stylesheet" href="/static/stock.1d101be0b97db6e6.css"></template>
<p class="stock">12 in stock</p>
```

Kutuyla birlikte yalnızca kendi tanımları gelir. Ürünün handler'ı bu render'a
dahil değildir. Bu yüzden burada ne title ne de bir Open Graph tag'i vardır;
kutuyu çeken page kendi head'ini korur.

Stylesheet, page'in head'inin onu tekilleştirdiği key'i taşır. Bu page `stock.css`'i
ilk render'ında zaten link'lemiştir, yani bir script'in ekleyeceği bir şey yoktur.
Kutu olmadan render edilmiş bir page'de ise kutunun stylesheet'i client'a bu yolla
gelir ve bir script onu head'e ekleyene kadar etkisiz kalır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 170" role="img" aria-labelledby="hs-frag-title">
<title id="hs-frag-title">Aynı stok kutusu, iki farklı cevap. Ürün page'inin içinde, /products/42 için gelen bir request title'ı, description'ı, Open Graph tag'lerini, product.css'i ve stok kutusunun stock.css'ini layout'un head'ine, stok kutusunu da body'ye koyar. Kendi URL'sinde, /products/42/stock için gelen bir request önce stock.css'i taşıyan bir template elementi, sonra stok kutusunun markup'ıyla cevaplanır; etrafında layout yoktur ve ürünün handler'ının tanımladığı hiçbir şey yoktur.</title>
<rect class="d-pill" x="4" y="4" width="152" height="22" rx="11" /><text class="d-small" x="80" y="19" text-anchor="middle">GET /products/42</text>
<rect class="d-frame" x="4" y="34" width="152" height="130" rx="6" />
<text class="d-small" x="14" y="50">&lt;head&gt;</text>
<rect class="d-chip" x="14" y="56" width="132" height="16" rx="8" /><text class="d-small" x="80" y="67" text-anchor="middle">title · meta · og:*</text>
<rect class="d-chip" x="14" y="76" width="132" height="16" rx="8" /><text class="d-small" x="80" y="87" text-anchor="middle">product.css</text>
<rect class="d-chip-hit" x="14" y="96" width="132" height="16" rx="8" /><text class="d-hit-text d-small" x="80" y="107" text-anchor="middle">stock.css</text>
<text class="d-small" x="14" y="130">&lt;body&gt;</text>
<rect class="d-c" x="14" y="136" width="132" height="20" rx="4" /><text class="d-on-c d-small" x="80" y="150" text-anchor="middle">stok kutusu</text>
<rect class="d-pill" x="164" y="4" width="152" height="22" rx="11" /><text class="d-small" x="240" y="19" text-anchor="middle">GET /products/42/stock</text>
<rect class="d-frame" x="164" y="34" width="152" height="130" rx="6" />
<text class="d-small" x="174" y="50">response body</text>
<rect class="d-chip-hit" x="174" y="56" width="132" height="34" rx="8" /><text class="d-hit-text d-small" x="240" y="70" text-anchor="middle">&lt;template&gt;</text><text class="d-hit-text d-small" x="240" y="83" text-anchor="middle">stock.css</text>
<rect class="d-c" x="174" y="96" width="132" height="20" rx="4" /><text class="d-on-c d-small" x="240" y="110" text-anchor="middle">stok kutusu</text>
<text class="d-small" x="174" y="138">etrafında layout yok</text>
<text class="d-small" x="174" y="154">title yok, og:* yok</text>
</svg>
<figcaption>Aynı stok kutusu, iki farklı cevap. Ürün page'inin içinde kutunun <code>stock.css</code>'i; title, description, Open Graph tag'leri ve <code>product.css</code> ile birlikte layout'un head'ine katılır. Kendi URL'sinde layout yoktur: önce stylesheet, key'ini belirten etkisiz bir <code>&lt;template data-collage-hoist&gt;</code> içinde gelir, sonra kutunun markup'ı. Ürünün handler'ının tanımladığı hiçbir şey cevapta yer almaz.</figcaption>
</figure>
