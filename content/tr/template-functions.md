---
description: Yerleşik şablon fonksiyonlarının tamamı — slot, hoist, asset, stylesheet, csrfToken, URL fonksiyonları ve metin yardımcıları — imzaları, örnekleri ve sınır durumlarıyla.
---

# Şablon fonksiyonları

Şablonlar Go'nun `html/template` paketidir, dolayısıyla onun sunduğu her şey
elinizin altındadır: `if`, `range`, `with`, `define`, `block` ve standart
fonksiyonlar `and`, `or`, `not`, `len`, `index`, `slice`, `print`, `printf`,
`println`, `eq`, `ne`, `lt`, `le`, `gt`, `ge`, `call`, `html`, `js` ve `urlquery`.
collage bunlara bu sayfadaki fonksiyonları ekler.

| Fonksiyon | İmza | Döndürdüğü |
| --- | --- | --- |
| [`slot`](#slot) | `slot name` | bir slot'taki fragment'lerin işaretlemesi |
| [`hoist`](#hoist) | `hoist area` | yukarı taşınan içeriğin yerleşeceği yer |
| [`asset`](#asset) | `asset path` | mount edilmiş bir dosyanın içerik adresli URL'si |
| [`stylesheet`](#stylesheet) | `stylesheet path` | hiçbir şey; head için bir stil dosyası bildirir |
| [`csrfToken`](#csrftoken) | `csrfToken` | bir formun token'ını taşıyan gizli input |
| [`pageURL`](#pageurl) | `pageURL name [param value]...` | bir route'un bu render'ın locale'indeki URL'si |
| [`pageURLIn`](#pageurlin) | `pageURLIn locale name [param value]...` | bir route'un tam olarak o locale'deki URL'si |
| [`localeURL`](#localeurl) | `localeURL locale` | bu sayfanın başka bir locale'deki URL'si |
| [`safeHTML`](#safehtml) | `safeHTML string` | HTML olarak güvenilen metin |
| [`safeURL`](#safeurl) | `safeURL string` | URL olarak güvenilen metin |
| [`dict`](#dict) | `dict key value [key value]...` | çiftlerden kurulan bir map |
| [`default`](#default) | `default fallback value` | `value` ya da boşsa `fallback` |
| [`upper`](#upper-and-lower) | `upper string` | metnin büyük harfli hâli |
| [`lower`](#upper-and-lower) | `lower string` | metnin küçük harfli hâli |
| [`title`](#title) | `title string` | her kelimenin ilk harfi büyük |
| [`join`](#join) | `join sep items` | `sep` ile birleştirilmiş öğeler |
| [`formatTime`](#formattime) | `formatTime time layout` | biçimlendirilmiş zaman |

Hata döndüren bir fonksiyon şablonu başarısız kılar, fragment da onunla birlikte
başarısız olur — fragment'in kendi hata politikasına göre: zorunlu bir fragment
sayfayı başarısız kılar, isteğe bağlı olan yedeğini render eder, yedeği yoksa hiçbir
şey render etmez. Geliştirme modunda bu "hiçbir şey", fragment'i ve hatasını
adlandıran bir HTML yorumudur; sayfanın üstündeki geliştirme paneli de yedekli ya da
yedeksiz, başarısız olan her fragment'i adıyla gösterir.
Bkz. [Fragment'ler ve slot'lar](/docs/fragments-and-slots).

## Her render'da bağlananlar

İlk sekiz fonksiyon, parçası oldukları render'a ihtiyaç duyar — fragment'e, isteğe,
locale'e, uygulamanın route'larına ve mount'larına. Şablon çağırabilsin diye
şablonlar ayrıştırılırken yer tutucu olarak kaydedilirler; gerçek gerçekleştirimi
render motoru her render'da bağlar. Herhangi bir şekilde render dışında çalışan bir
yer tutucu tahmin yürütmek yerine hata döndürür.

Bunun bir sonucu var: `Config.Template.Funcs` içinde ya da bir plugin'den gelen, bu
adlardan birini taşıyan bir girdi ayrıştırılır ama hiçbir zaman çağrılmaz. Diğer
dokuzu değiştirilebilir.

### slot

```html
{{slot "name"}}
```

Bu fragment'in `name` adlı slot'una bağlanmış her fragment'i bağlanma sırasıyla
render eder ve işaretlemelerini kaçışlamadan yerleştirir. `slot` her zaman *bu*
fragment'in slot'u demektir; böylece iki fragment'in her biri, birbirine
karışmadan `"sidebar"` adlı bir slot'a sahip olabilir.

```html
<article>
  {{slot "content"}}
  <aside>{{slot "related"}}</aside>
</article>
```

- Fragment'in `WithSlot` ile hiç bildirmediği bir slot hatadır
  (`ErrUnknownSlot`) ve hata, fragment'in sahip olduğu slot'ları sayar. Hiçbir şey
  render etmek, bir yazım hatasını sessizce eksik kalan bir bölüme dönüştürürdü.
- Bildirilmiş ama boş bir slot hiçbir şey render etmez — zorunlu olarak
  bildirilmediyse. Hiçbir şey bağlanmamış ve resolver'ı olmayan zorunlu bir slot,
  sayfa kaydedilirken `ErrRequiredSlotUnfilled` ile reddedilir, dolayısıyla hiçbir
  zaman render'a ulaşmaz. Bir resolver'ın doldurduğu zorunlu slot ise render
  sırasında denetlenir: hiç fragment döndürmeyen bir resolver render'ı
  `ErrRequiredSlotEmpty` ile başarısız kılar.
- Bir layout'un içeriği `{{slot "content"}}`'e gider; bu ad
  `collage.DefaultContentSlot`'ta tutulur.

`slot` çalıştığında çocukların data handler'ları çoktan başlamıştır: hepsi, üst
fragment render edilmeden önce birlikte başlar. Bkz.
[Fragment'ler ve slot'lar](/docs/fragments-and-slots).

### hoist

```html
{{hoist "head"}}
```

`area`'ya taşınan içeriğin nereye yerleşeceğini işaretler. Fragment'ler bu içeriği
ağacın herhangi bir yerinden — genellikle bir data handler'dan — bildirir, nereye
gideceğine ise `hoist` karar verir:

```html
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
```

```go
rc.HoistTitle(post.Title)
rc.HoistMeta("description", post.Summary)
rc.HoistLink("canonical", canonicalURL)
```

İçerik değil, bir işaretçi yazar; çünkü altındaki hiçbir şey henüz render
edilmemiştir. Ağacın tamamı render edildiğinde her işaretçi, kendi alanı için
bildirilenle değiştirilir; böylece sayfanın derinliklerinde yapılan bir bildirim de
head'e ulaşır. Hiçbir şey bildirilmemiş bir alan, hiçbir şeyle değiştirilir.

`HoistTitle`, `HoistMeta`, `HoistProperty`, `HoistLink`, `HoistAlternate` ve
`HoistStylesheet` `"head"` alanına yazar; `rc.Hoist(area, key, html)` ise adını
verdiğiniz herhangi bir alana yazar. **Head'e katkıda bulunan bir plugin bu
işaretçiye ihtiyaç duyar**: o olmadan içeriğinin gidecek yeri yoktur. Bkz.
[Head ve SEO](/docs/head-and-seo).

### asset

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<img src="{{asset "/static/logo.svg"}}" alt="">
```

Bir mount'taki dosyanın içerik adresli URL'sini döndürür: uzantıdan önce içeriğinin
hash'i eklenmiş dosya yolu.

```text
/static/app.css  →  /static/app.41014ebb6c2d9f07.css
```

Baytlardan türetilen bir ad yalnızca o baytları ifade edebilir; onu bir yıllık,
`immutable` bir ömürle sunmayı güvenli kılan da budur: değişen bir dosya, değişen
bir addır ve eskisi bir daha asla istenmez.

- Yol, mount'un dosyayı sunduğu URL'dir; mount'un önekini de içerir.
- **Var olmayan bir dosya bir URL değil, hatadır**: `ErrUnknownAsset`. Alternatifi,
  sorunsuz render edilen ama 404 veren bir stil dosyasına bağlantı veren bir
  sayfadır.
- Hiçbir mount'un sunmadığı bir yol ya da hiç mount'u olmayan bir uygulama da
  `ErrUnknownAsset`'tir — hiçbir mount'un öneki yolu kapsamıyorsa
  `ErrNoMountForAsset`'i sarmalar.
- Statik dışa aktarma, sayfaların istediği parmak izli kopyaların tam olarak
  kendisini, orijinallerin yanına yazar.

Go'dan `rc.Asset(path)` aynı URL'yi döndürür. Bkz. [Statik dosyalar](/docs/assets).

### stylesheet

```html
{{stylesheet "/static/gallery.css"}}
```

Bu fragment'in bir stil dosyasına ihtiyaç duyduğunu bildirir ve çağrıldığı yerde
hiçbir şey render etmez. Sayfanın head'i, layout'un `{{hoist "head"}}` çağırdığı
yerde, dosyanın [`asset`](#asset) URL'siyle bir
`<link rel="stylesheet" href="…">` alır.

Böylece bir fragment, nerede kullanılırsa kullanılsın kendi stillerini
beraberinde taşıyabilir:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

Bildirim yola göre anahtarlanır; dolayısıyla birkaç fragment'in istediği bir stil
dosyası bir kez görünür. `asset` ile aynı koşullarda başarısız olur. Go'dan
karşılığı `rc.HoistStylesheet(path)`'tir.

### csrfToken

```html
<form method="post">
  {{csrfToken}}
  <input name="email" type="email">
  <button>Subscribe</button>
</form>
```

Bir formun istek sahteciliği token'ını taşıyan gizli input'u render eder:

```html
<input type="hidden" name="_csrf" value="…">
```

Çıplak değer değil de bütün bir input, çünkü çıplak değerin tam olarak doğru adı
taşıyan bir alana konması gerekir ve adı yanlış yazan bir form, nedenini açıklayan
hiçbir şey olmadan reddedilir.

Render sırasında yazılan değer bir yer tutucudur. Yanıt yazılırken her okuyucunun
kendi token'ıyla değiştirilir; **önbellekteki** bir sayfanın form taşıyabilmesinin
nedeni de budur: önbellekteki baytlar yer tutucuyu tutar ve her okuyucu kendi
token'ını alır.

- `Security.DisableCSRF` ayarlıyken `csrfToken`, render'ı `ErrCSRFDisabled` ile
  başarısız kılar: aksi hâlde token bekleyen bir form token'sız render edilirdi.
- `Security.CSRFFieldName` başka bir şey söylemedikçe alanın adı `_csrf`'tir ve her
  iki durumda da doğrulayıcının okuduğu ad budur; dolayısıyla adı değiştirilmiş bir
  alan şablonda hiçbir değişiklik gerektirmez.
- Statik dışa aktarmada token koyacak bir sunucu yoktur; bu yüzden token taşıyan
  bir sayfa dışa aktarılmaz ve rapor nedenini söyler.

Bkz. [Formlar ve action'lar](/docs/forms-and-actions).

### pageURL

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

`name` altında kaydedilmiş sayfanın ya da document'ın URL'sini, yol pattern'ini
`param value` çiftleriyle doldurarak döndürür. Ada göre bağlantı vermek, bir
sayfanın yolu değiştiğinde bağlantının da onu izlemesi demektir.

- **Bu render'ın locale'inde.** Türkçe bir sayfada `pageURL "blog-post"` Türkçe
  yoldur. Geçerli locale'de yolu olmayan bir route, onun yerine varsayılan
  locale'deki yoluna bağlanır; böylece yalnızca İngilizce olan bir sayfaya bağlantı
  veren Türkçe bir sayfa yine de render edilir.
- **Katı.** Bilinmeyen bir ad (`ErrUnknownRoute`), tek sayıda parametre argümanı,
  eksik ya da boş bir parametre veya pattern'de yer tutucusu olmayan bir parametre
  (`ErrRouteParams`) render'ı başarısız kılar. Kurulamayan bir bağlantı, okuyucu
  için bir 404 değil, geliştirme sırasında bulunacak bir hatadır.
- **Değerler metindir** ve kaçışlanır. Tarayıcının bir yol adımı olarak
  çözümleyeceği `.` ya da `..` değeri reddedilir. Bir sayıyı `printf` üzerinden
  geçirin:

```html
<a href="{{pageURL "user" "id" (printf "%d" .ID)}}">{{.Name}}</a>
```

- Hem sayfa hem document olarak kaydedilmiş bir ad, tahmin edilmek yerine
  reddedilir.

Go'dan karşılığı `app.URL(name, locale, params)`'tır. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales).

### pageURLIn

```html
<a href="{{pageURLIn "tr" "about"}}">Hakkımızda</a>
<a href="{{pageURLIn "en" "blog-post" "slug" .Slug}}">Read in English</a>
```

Tam olarak verilen locale'de, yedeksiz bir `pageURL`: o locale'de yolu olmayan bir
route `ErrNoPathInLocale`'dir; hiçbir URL'nin ulaşamadığı bir locale — ya
`Locale.Supported` içinde olmayan ya da `DisablePathLocale` ayarlıyken varsayılan
dışındaki herhangi biri — `ErrLocaleUnreachable`'dır.

### localeURL

```html
{{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
{{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
```

Render edilen sayfanın, aynı yol parametreleriyle başka bir locale'deki hâli — bir
dil değiştiricinin yapıldığı malzeme.

- **O locale'de yolu olmayan bir sayfa hata değil, boş metindir**; böylece
  `{{with}}`, sayfanın çevrilmediği bir dili atlar.
- Hiçbir URL'nin ulaşamadığı bir locale yine de hatadır (`ErrLocaleUnreachable`):
  bu eksik bir çeviri değil, şablondaki bir yanlıştır.
- Bir sayfaya ait olmayan bir render'da başarısız olur.
- Arama motorlarının okuduğu `<link rel="alternate" hreflang>` için bunları Go'dan
  `rc.HoistAlternate` ile bildirin — bkz.
  [Head ve SEO](/docs/head-and-seo#canonical-and-alternate-links).

### safeHTML

```html
{{safeHTML .RenderedMarkdown}}
```

Bir metni güvenilir HTML olarak işaretler; böylece `html/template` onu kaçışlamadan
yerleştirir. Bu bir kaçış kapısıdır: yalnızca kendi ürettiğiniz ya da kendiniz
temizlediğiniz işaretleme için kullanın, bir kullanıcının yazdığı hiçbir şey için
asla kullanmayın.

### safeURL

```html
<a href="{{safeURL .ExternalLink}}">Visit</a>
```

Bir metni güvenilir URL olarak işaretler ve `html/template`'in URL temizlemesini
atlar — bu temizleme, aksi hâlde güvenmediği bir şemayı, örneğin `javascript:`'i,
`#ZgotmplZ` ile değiştirir. Yalnızca doğruladığınız URL'ler için.

### dict

```html
{{template "card" dict "Title" .Title "URL" (pageURL "post" "slug" .Slug)}}
```

Bir alt şablona birkaç değer geçirmek için, sırayla gelen anahtar ve değerlerden bir
`map` kurar. Anahtarlar metin olmalıdır (`ErrDictKeyNotString`) ve argümanlar çift
çift gelmelidir (`ErrDictOddArgs`); iki yanlıştan her biri şablonu başarısız kılar.

### default

```html
<h2>{{default "Untitled" .Subtitle}}</h2>
<h2>{{.Subtitle | default "Untitled"}}</h2>
```

`value`'yu, `value` boş metinse `fallback`'i döndürür. İki argüman da metindir —
yedek önce gelir ki pipe'lı biçim doğal okunsun, çünkü bir pipeline değerini son
argüman olarak geçirir.

### upper ve lower

```html
<span class="badge">{{upper .Status}}</span>
<code>{{lower .Code}}</code>
```

`strings.ToUpper` ve `strings.ToLower`.

### title

```html
<h1>{{title .Name}}</h1>
```

Her kelimenin ilk harfini büyük, geri kalanını küçük harfe çevirir. Bir kelime
harflerden oluşan kesintisiz bir dizidir; dolayısıyla harf olmayan her şey — bir
boşluk, bir tire, bir kesme işareti — yeni bir kelime başlatır:

| Girdi | Çıktı |
| --- | --- |
| `hello world` | `Hello World` |
| `iPHONE case` | `Iphone Case` |
| `o'neil-smith` | `O'Neil-Smith` |

### join

```html
<p>Tags: {{join ", " .Tags}}</p>
```

Ayırıcı önde olmak üzere `strings.Join(items, sep)`. `items` bir `[]string`
olmalıdır.

### formatTime

```html
<time datetime="{{formatTime .Published "2006-01-02"}}">
  {{formatTime .Published "2 January 2006"}}
</time>
```

Bir [referans zamanı layout'u](https://pkg.go.dev/time#pkg-constants) ile
`t.Format(layout)`. `t` bir `time.Time`'dır. Taşıdığı konum neyse o konumda
biçimlendirilir; başka bir konum istiyorsanız data handler'da dönüştürün.

## Kendi fonksiyonlarınızı eklemek

Fonksiyonları `New`'dan önce `Config.Template.Funcs` ile ekleyin:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		Root: "templates",
		Funcs: template.FuncMap{
			"money": func(cents int64) string {
				return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
			},
		},
	},
})
```

Yerleşik fonksiyonların üzerine birleştirilirler; dolayısıyla yerleşik bir adı
taşıyan girdi — her render'da bağlananlar dışında — onun yerini alır. `New`'dan
önce orada olmaları gerekir, çünkü `html/template` yalnızca şablon ayrıştırılırken
fonksiyon map'inde bulunan bir adı çağırabilir; kimsenin kaydetmediği bir adı
çağıran şablon ilk istekte değil, `New`'da başarısız olur.

Bir plugin de fonksiyonları aynı şekilde, `Configure` aşamasından ekler; aynı adı
taşıyan uygulama girdisi kazanır. Bkz.
[Plugin yazmak](/docs/writing-plugins#template-functions).

İsteğe ihtiyaç duyan her şeyin yeri bir fonksiyon değil, bir data handler'dır:
sayfanın verisi zaten oradan gelir.
