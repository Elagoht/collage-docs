---
description: Başlıklar, meta etiketleri, stil dosyaları ve structured data; onları bilen fragment bildirir, layout yerleştirir.
---

# Head ve SEO

Bir sayfanın başlığını bilen fragment, `<head>`'i yazan fragment nadiren olur. Yazı
kendi başlığını ve özetini bilir; galeri `gallery.css`'e ihtiyaç duyduğunu bilir;
`<head>`'i ikisi de render edilmeden önce yazmış olan layout ise ikisini de bilmez.

collage bunu **hoisting** (yukarı taşıma) ile çözer: bir fragment, sayfada nerede
durursa dursun head'e neyin ait olduğunu *bildirir*, layout da bildirimlerin nereye
düşeceğini söyler.

## Nereye düşer

Layout, `<head>`'inin içine bir işaret koyar:

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

`{{hoist "head"}}` içerik değil bir işaret yazar — altındaki hiçbir şey henüz render
edilmemiştir. Sayfanın tamamı bittiğinde collage, işareti ağacın herhangi bir yerinde
`"head"` alanı için bildirilmiş her şeyle değiştirir. Tek geçiş; konuma yine layout
karar verir.

İşaret olmadan bildirimler hiçbir yere gitmez. Bir başlık ya da bir plugin'in
çıktısı eksik olduğunda kontrol edilecek ilk şey budur.

## Bir data handler'dan bildirmek

`RenderContext` üzerindeki yardımcılar neredeyse her sayfanın ihtiyacını karşılar.
Onları bir data handler'dan çağırın:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return postView{}, nil, err
	}

	rc.HoistTitle(post.Title + " — The Wire")
	rc.HoistMeta("description", post.Summary)
	rc.HoistProperty("og:title", post.Title)
	rc.HoistProperty("og:image", post.CoverURL)
	rc.HoistLink("canonical", siteOrigin+"/blog/"+post.Slug)
	if err := rc.HoistStylesheet("/static/post.css"); err != nil {
		return postView{}, nil, err
	}

	return postView{Post: post}, []string{"post:" + post.Slug}, nil
}
```

| Çağrı | Ürettiği | Anahtar |
| --- | --- | --- |
| `rc.HoistTitle(text)` | `<title>text</title>` | `title` |
| `rc.HoistMeta(name, content)` | `<meta name="…" content="…">` | `meta:<name>` |
| `rc.HoistProperty(property, content)` | `<meta property="…" content="…">` (Open Graph) | `property:<property>` |
| `rc.HoistLink(rel, href)` | `<link rel="…" href="…">` | `link:<rel>` |
| `rc.HoistAlternate(hreflang, href)` | `<link rel="alternate" hreflang="…" href="…">` (v0.10.0'dan itibaren) | `alternate:<hreflang>` |
| `rc.HoistStylesheet(path)` | Dosyanın [içerik adresli URL'sinde](/docs/assets) `<link rel="stylesheet" href="…">` | `stylesheet:<path>` |

Her yardımcı kendisine verileni escape eder; bu yüzden bir yazının adından kurulan
başlık, yazının adı ne olursa olsun güvenlidir. `HoistStylesheet` dosyanın mount
edilmiş yolunu alır ve hiçbir mount o dosyaya sahip değilse bir hata döndürür — var
olmayan bir stil dosyasına bağlanan sayfa bozuk bir sayfadır ve bunu söylemelidir.

Onları senkron olarak, data handler'ın kendi goroutine'inden çağırın. Handler'ın
başlattığı bir goroutine'den yapılan bildirimin sayfada bir konumu yoktur.

### Bir şablondan

Bir fragment'in şablonu stil dosyasını kendisi isteyebilir:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

`{{stylesheet}}` durduğu yerde hiçbir şey render etmez; `HoistStylesheet`'in
yaptığını yapar. Galerinin CSS'i galeriyle birlikte, galeri hangi sayfadaysa oraya
gider.

## Anahtarlar ve en içteki kazanır

Her bildirimin bir anahtarı vardır ve neyin aynı şey sayılacağına anahtar karar
verir:

- **Farklı anahtarların hepsi görünür**, her biri sayfa sırasındaki en erken
  bildiriminin yerinde.
- **Aynı anahtar iki kez bildirilirse en içteki bildirim kalır.**

Bu tek kural iki işi birden görür. Stil dosyası yoluna göre anahtarlanır; böylece
`gallery.css` isteyen beş fragment tek bir `<link>` üretir, farklı stil dosyaları da
birikir. Başlığın tek bir anahtarı vardır; böylece daha özel bir fragment'in başlığı
daha genel olanınkinin yerini alır. Bu, varsayılanları kolaylaştırır — onları
layout'ta bildirin ve herhangi bir sayfanın üzerine yazmasına izin verin:

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistTitle("The Wire")
		rc.HoistMeta("description", "News about the sea, and the people who live beside it.")
		return nil
	})).
	WithSlot("content", true, false).
	Build()
```

O layout'un içine yerleşmiş bir yazı kendi başlığını ve açıklamasını bildirir ve
sayfada görünenler onlardır. Hiçbir şey bildirmeyen bir sayfa layout'unkileri korur.
Layout'un şablonuna ayrıca düz bir `<title>` yazmayın — hoist edilenin yanında
dururdu ve iki başlığı olan bir sayfanın başlıklarından birini tarayıcı yok sayar.
`collage new`'un iskeletini oluşturduğu layout sitenin adını bu şekilde bildirir;
böylece bir sayfanın `rc.HoistTitle`'ı onun yerini alır.

Sürpriz olmasınlar diye üç ayrıntı:

- **Konum, kazanan bildirimden değil ilk bildirimden gelir.** Head, iç içe bir şeyin
  bir başlığın üzerine yazıp yazmamasına göre kendini yeniden sıralamaz.
- **Eşit derinlikte, daha sonra bildiren fragment kazanır.** Aynı anahtarı bildiren
  iki kardeş fragment gerçek bir çakışmadır ve sayfadaki sıralarıyla çözülür —
  hangisinin data handler'ının önce bittiğiyle değil; öyle olsaydı head istekten
  isteğe değişirdi.
- **Sıra saatin değil sayfanındır.** Kardeş data handler'lar eşzamanlı çalışır, ama
  bir anahtar, önce onu ilk bildiren fragment'in sayfadaki yerine, sonra o
  fragment'in bildirim sırasına göre yerleştirilir — böylece head, birbirinin üzerine
  yazan stil dosyaları dahil, her render'da aynıdır. (v0.12.0'dan önce bir anahtar,
  ilk bildiriminin tesadüfen ulaştığı yere oturuyordu ve kardeşlerin anahtarları
  istekler arasında yer değiştirebiliyordu.)

## Geri kalan her şey: rc.Hoist

Yardımcılar tek bir metot üzerine kuruludur:

```go
func (rc *RenderContext) Hoist(area, key string, html template.HTML)
```

`html`'i `key` altında, `area`'ya **tam olarak yazıldığı gibi** ekler. Varlık
nedeni de budur — yardımcıların kapsamadığı markup — ve bu, escape etmenin sizin
işiniz olduğu anlamına gelir. Markup'ı kontrol ettiğiniz değerlerden kurun ya da
onları escape edin:

```go
rc.Hoist("head", "preload:hero", template.HTML(
	`<link rel="preload" as="image" href="`+html.EscapeString(post.CoverURL)+`">`))
```

Aynı anahtar kuralları geçerlidir; bu yüzden şeyin ne olduğunu söyleyen bir anahtar
seçin: bir kez görünmesi gereken her şey için bir anahtar, biriken şeyler için de
değer başına bir anahtar.

Alan herhangi bir ad olabilir ve bir layout'ta birden fazla işaret olabilir —
örneğin bir fragment'in ihtiyaç duyduğu script'ler için `</body>`'den önce bir `{{hoist "scripts"}}`:

```go
rc.Hoist("scripts", "script:map", template.HTML(`<script src="`+mapJS+`" defer></script>`))
```

## Canonical ve alternate bağlantılar

Canonical bağlantı, aynı içeriğe birden fazla URL ulaştığında — örneğin izleme
query'siyle ve onsuz — arama motorlarına hangisinin asıl URL olduğunu söyler. Ona
mutlak bir URL verin. collage origin'leri değil yolları kurar; bu yüzden sitenizin
origin'ini kendi yapılandırmanızda tutun ve ikisini birleştirin:

```go
rc.HoistLink("canonical", siteOrigin+rc.Request.URL.Path)
```

Birden fazla dilde olan bir site, her çevirinin nerede olduğunu da dil başına bir
`<link rel="alternate" hreflang="…">` ile söylemelidir. `HoistLink` her `rel` için
tek bağlantı tutar, bu yüzden bunu söyleyemez; `rc.HoistAlternate` ise her
bağlantıyı diline göre anahtarlar, böylece her çeviri bir kez görünür. Onları
layout'tan, hangi sayfa render ediliyorsa onun için,
[`app.URL`](/docs/links-and-locales#links-from-go) ile bildirin:

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
		WithSlot("content", true, false).
		Build()
}
```

Çevirisinin slug'ı farklı olan bir sayfa — bunu yalnızca içeriği bilir — o dil için
kendi `rc.HoistAlternate`'ini bildirir ve daha içte olduğu için layout'unkinin
yerini alır. Handler'ı olmayan bir layout için şablondaki karşılığı
[`localeURL`](/docs/links-and-locales#a-language-switcher)'dir; sayfanın var
olmadığı bir dil için boştur:

```html
{{with localeURL "tr"}}<link rel="alternate" hreflang="tr" href="https://thewire.example{{.}}">{{end}}
```

## Structured data

Arama motorları schema.org structured data'sını head'deki JSON-LD bloklarından okur.
[`elagoht/jsonld` plugin'i](/docs/plugins#elagohtjsonld) bunları sizin için, elle
kurulmuş JSON yerine tipli değerlerden yazar:

```go
import "github.com/Elagoht/collage-jsonld"

func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return postView{}, nil, err
	}
	jsonld.Emit(rc, jsonld.BlogPosting{
		Headline:      post.Title,
		Description:   post.Summary,
		DatePublished: post.PublishedAt,
		AuthorName:    post.Author,
	})
	return postView{Post: post}, []string{"post:" + post.Slug}, nil
}
```

`jsonld.Emit` `"head"`'e hoist eder; bu yüzden bu sayfadaki diğer her şeyle aynı
işarete ihtiyaç duyar ve aynı kurallara uyar: schema.org tipi başına bir anahtar, en
içteki kazanır. Plugin'i kaydetmek (`Config.Plugins` içinde `jsonld.New()`), bir site
adıyla yapılandırıldığında site genelinde bir `WebSite` düğümü ekler; sayfa başına
veri ise her zaman data handler'larınızdan gelir, çünkü sayfanın ne hakkında olduğunu
yalnızca onlar bilir.

## Hoisting ve collage'ın geri kalanı

- **Önbellekteki sayfalar head'lerini korur.** İşaret, sayfa saklanmadan önce
  değiştirilir; böylece önbellekteki bir sayfa bildirilmiş her şeyle birlikte sunulur.
- **Kendi URL'sindeki bir fragment'in head'i yoktur.**
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) üzerinden
  sunulan bir fragment'in layout'u yoktur; bu yüzden hoist ettiklerinin düşeceği bir
  yer yoktur.
- **`rc.Page`'den değil, data handler'lardan hoist edin.** `rc.Page` kaydedilmiş
  sayfadır ve her istek tarafından paylaşılır; ona yazmak bir data race'tir. İstek
  başına değişen şey `rc` üzerinden bildirilir.
