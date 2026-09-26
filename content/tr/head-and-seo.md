---
description: Title'ları, meta tag'leri, stylesheet'leri ve structured data'yı onları bilen fragment tanımlar, layout yerleştirir.
reference: RenderContext, FragmentBuilder.WithTitle, Effect, ErrNoPathInLocale
---

# Head ve SEO

Bir page'in title'ını bilen fragment, `<head>`'i yazan fragment nadiren olur. Post
kendi title'ını ve özetini bilir. Gallery, `gallery.css`'e ihtiyacı olduğunu bilir.
Layout ise `<head>`'i ikisi de render edilmeden önce yazmıştır ve ikisini de bilmez.

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
şey henüz render edilmemiştir. Page'in tamamı bittiğinde collage bu marker'ı, ağacın
herhangi bir yerinde `"head"` alanı için tanımlanmış her şeyle değiştirir. Bu iş
tek geçişte yapılır ve konumu yine layout belirler.

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
func loadPost(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
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
oluşturulan title, post'un adı ne olursa olsun güvenlidir. `HoistStylesheet`,
dosyanın mount edilmiş path'ini alır. O dosya hiçbir mount'ta yoksa hata döner.
Var olmayan bir stylesheet'e link veren page bozuk bir page'dir ve bunu açıkça
söylemesi gerekir.

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
	WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistMeta("description", "News about the sea, and the people who live beside it.")
		return nil
	})).
	Build()
```

Bu layout'un içindeki bir post kendi title'ını ve description'ını tanımlar.
Page'de görünenler de bunlar olur. Hiçbir şey tanımlamayan bir page layout'unkileri
korur. Layout'un template'ine ayrıca elle bir `<title>` yazmayın. Yazarsanız o title,
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

Area herhangi bir isim olabilir ve bir layout'ta birden fazla marker bulunabilir.
Örneğin fragment'lerin ihtiyaç duyduğu script'ler için `</body>`'den önce bir `{{hoist
"scripts"}}` koyabilirsiniz:

```go
rc.Hoist("scripts", "script:map", template.HTML(`<script src="`+mapJS+`" defer></script>`))
```

## Canonical ve alternate link'ler

Aynı içeriğe birden fazla URL ile ulaşılabiliyorsa, örneğin tracking query'siyle ve
onsuz, canonical link arama motorlarına hangisinin asıl URL olduğunu söyler. Bu link'e
mutlak bir URL verin. collage origin değil path üretir. Bu yüzden sitenizin
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
		WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
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
içte olduğu için de layout'un tanımının yerine geçer. Handler'ı olmayan bir
layout için (böyle bir layout, strateji tanımlamayan page'leri static bırakır)
template'teki karşılığı
[`localeURL`](/docs/links-and-locales#a-language-switcher) fonksiyonudur. Page'in var
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

func loadPost(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
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

`jsonld.Emit`, `"head"` area'sına hoist eder. Bu yüzden bu sayfada anlatılan diğer
her şeyle aynı marker'a ihtiyaç duyar ve aynı kurallara uyar: her schema.org tipi için
bir key vardır ve en içteki kazanır. Plugin'i register ettiğinizde (`Config.Plugins`
içinde `jsonld.New()`), bir site adıyla yapılandırılmışsa site genelinde bir
`WebSite` node'u ekler. Page'e özel veri ise her zaman data handler'larınızdan gelir.
Çünkü page'in ne hakkında olduğunu yalnızca onlar bilir.

## Hoisting ve collage'ın geri kalanı

- **Cache'lenen page'ler head'lerini korur.** Marker, page saklanmadan önce
  değiştirilir. Bu yüzden cache'lenmiş bir page, tanımlanmış her şeyle birlikte
  sunulur.
- **Kendi URL'sinde sunulan bir fragment'in head'i yoktur.**
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile sunulan
  bir fragment'in layout'u yoktur. v0.18.0'dan beri marker koymadığı bir alana
  hoist ettiği şeyler markup'ından önce, her item için etkisiz bir
  `<template data-collage-hoist>` olarak gelir. Bir script bunları page'e ekleyebilir.
  Bkz. [Hoist ettikleri](/docs/forms-and-actions#what-it-hoists).
- **Hoist işlemini `rc.Page` üzerinden değil, data handler'lardan yapın.** `rc.Page`
  register edilmiş page'dir ve her request tarafından paylaşılır. Ona yazmak data
  race'e yol açar. Request'e göre değişen her şey `rc` üzerinden tanımlanır.
