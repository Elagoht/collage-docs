---
description: Built-in template fonksiyonlarının tamamı (slot, hoist, asset, csrfToken, URL fonksiyonları, string yardımcıları), imzaları, örnekleri ve sınır durumları.
reference: TemplateConfig, DefaultContentSlot, ErrUnknownSlot, FragmentBuilder.WithTitle, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Template fonksiyonları

Template'ler Go'nun `html/template` paketiyle yazılır. Bu yüzden onun sunduğu her şey
kullanılabilir: `if`, `range`, `with`, `define`, `block` ve standart fonksiyonlar
olan `and`, `or`, `not`, `len`, `index`, `slice`, `print`, `printf`, `println`,
`eq`, `ne`, `lt`, `le`, `gt`, `ge`, `call`, `html`, `js` ve `urlquery`. collage
bunlara bu sayfadaki fonksiyonları ekler.

| Fonksiyon | İmza | Döndürdüğü |
| --- | --- | --- |
| [`slot`](#slot) | `slot name` | bir slot'taki fragment'lerin markup'ı |
| [`hoist`](#hoist) | `hoist area` | hoist edilen içeriğin yerleşeceği yer |
| [`asset`](#asset) | `asset path` | mount edilmiş bir dosyanın content-addressed URL'si |
| [`stylesheet`](#stylesheet) | `stylesheet path` | hiçbir şey; head için bir stylesheet tanımlar |
| [`csrfToken`](#csrftoken) | `csrfToken` | form'un token'ını taşıyan hidden input |
| [`pageURL`](#pageurl) | `pageURL name [param value]...` | bir route'un bu render'ın locale'indeki URL'si |
| [`pageURLIn`](#pageurlin) | `pageURLIn locale name [param value]...` | bir route'un tam olarak o locale'deki URL'si |
| [`localeURL`](#localeurl) | `localeURL locale` | bu page'in başka bir locale'deki URL'si |
| [`actionURL`](#actionurl) | `actionURL name [param value]...` | bir action'ın bu render'ın locale'indeki URL'si |
| [`fragmentURL`](#fragmenturl) | `fragmentURL page fragment [param value]...` | bir fragment path'inin bu render'ın locale'indeki URL'si |
| [`fragmentURLIn`](#fragmenturlin) | `fragmentURLIn locale page fragment [param value]...` | bir fragment path'inin tam olarak o locale'deki URL'si |
| [`safeHTML`](#safehtml) | `safeHTML string` | HTML olarak güvenilir sayılan string |
| [`safeURL`](#safeurl) | `safeURL string` | URL olarak güvenilir sayılan string |
| [`dict`](#dict) | `dict key value [key value]...` | çiftlerden oluşturulan bir map |
| [`default`](#default) | `default fallback value` | `value`, boşsa `fallback` |
| [`upper`](#upper-and-lower) | `upper string` | string'in büyük harfli hâli |
| [`lower`](#upper-and-lower) | `lower string` | string'in küçük harfli hâli |
| [`title`](#title) | `title string` | her kelimenin ilk harfi büyük |
| [`join`](#join) | `join sep items` | `sep` ile birleştirilmiş item'lar |
| [`formatTime`](#formattime) | `formatTime time layout` | formatlanmış zaman |

Hata dönen bir fonksiyon template'i başarısız kılar, fragment da onunla birlikte
başarısız olur. Bu durumda fragment'in kendi failure policy'si uygulanır:

- required bir fragment page'i başarısız kılar;
- optional bir fragment fallback'ini render eder;
- fallback'i yoksa hiçbir şey render etmez.

Development modunda bu "hiçbir şey", fragment'in adını ve hatasını içeren bir
HTML yorumudur. Page'in üstündeki development paneli de fallback'i olsun olmasın,
başarısız olan her fragment'i adıyla listeler. v0.36.0'dan beri hata her modda
ayrıca bir uyarı olarak log'lanır. Ayrıntılar için
[Fragment'ler ve slot'lar](/docs/fragments-and-slots) sayfasına bakın.

## Her render'da bağlananlar

İlk on bir fonksiyon, içinde çalıştıkları render'a ihtiyaç duyar: fragment'e,
request'e, locale'e, uygulamanın route'larına ve mount'larına. Her biri farklı bir
şey okur:

| Fonksiyon | Okuduğu |
| --- | --- |
| `slot` | bu fragment ve onun slot'larına bağlanmış fragment'ler |
| `hoist` | render'ın area'ları; page'in tamamı render edildikten sonra doldurulur |
| `asset` | uygulamanın mount'ları |
| `stylesheet` | bu render'ın tanım yaptığı head |
| `csrfToken` | response yazılırken her okuyucunun kendi token'ıyla değiştirilen bir placeholder |
| `pageURL`, `actionURL`, `fragmentURL` | render'ın locale'i ve uygulamanın route'ları |
| `pageURLIn`, `fragmentURLIn` | uygulamanın route'ları; locale bir argümandır |
| `localeURL` | render edilen page ve onun path parametreleri |

Template'ler parse edilirken bu fonksiyonlar placeholder olarak register edilir,
böylece template'ler onları çağırabilir. Gerçek implementasyonu ise render engine
her render'da bağlar. Bir placeholder bir şekilde render dışında çalışırsa tahmin
yürütmez, hata döner.

Bunun bir sonucu var: `Config.Template.Funcs` içinde ya da bir plugin'den gelen ve
bu adlardan birini taşıyan bir entry parse edilir, ama hiçbir zaman çağrılmaz. Geri
kalan dokuz fonksiyonu ise değiştirebilirsiniz.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 306" role="img" aria-labelledby="tf-bind-title">
<title id="tf-bind-title">Bir template'in fonksiyonları nereden gelir. collage.New template'leri tek bir function map ile parse eder: collage'ın yirmi built-in'i (on biri placeholder), sonra plugin'lerin AddTemplateFunc ile eklediği fonksiyonlar, sonra Config.Template.Funcs; her biri bir öncekinin üzerine merge edilir. Her render'da, her fragment için on bir per-render ad o render'dan yeniden bağlanır; plugin'lerin AddRenderFunc ile eklediği fonksiyonlar da öyle. Bu yüzden pageURL adlı bir Template.Funcs entry'si parse edilir ama hiç çalışmaz, upper adlı bir entry ise built-in'in yerini alır.</title>
<defs><marker id="tf-bind-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="40" y="4" width="240" height="22" rx="11" /><text x="160" y="19" text-anchor="middle">collage.New: template parse</text>
<line class="d-flow" x1="160" y1="26" x2="160" y2="36" marker-end="url(#tf-bind-arrow)" />
<rect class="d-frame" x="4" y="40" width="312" height="96" rx="6" />
<text class="d-small" x="14" y="55">tek bir function map</text>
<rect class="d-b" x="14" y="62" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="76" text-anchor="middle">20 built-in (11 placeholder)</text>
<rect class="d-b" x="14" y="86" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="100" text-anchor="middle">plugin'ler: AddTemplateFunc</text>
<rect class="d-b" x="14" y="110" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="124" text-anchor="middle">Config.Template.Funcs</text>
<line class="d-flow" x1="262" y1="66" x2="262" y2="126" marker-end="url(#tf-bind-arrow)" />
<text class="d-small" x="290" y="92" text-anchor="middle">sonraki</text><text class="d-small" x="290" y="104" text-anchor="middle">kazanır</text>
<line class="d-flow" x1="160" y1="136" x2="160" y2="148" marker-end="url(#tf-bind-arrow)" />
<rect class="d-pill" x="40" y="152" width="240" height="22" rx="11" /><text x="160" y="167" text-anchor="middle">her render, her fragment</text>
<line class="d-flow" x1="140" y1="174" x2="90" y2="188" marker-end="url(#tf-bind-arrow)" />
<line class="d-flow" x1="180" y1="174" x2="230" y2="188" marker-end="url(#tf-bind-arrow)" />
<rect class="d-a" x="12" y="192" width="146" height="40" rx="5" /><text class="d-on-a" x="85" y="208" text-anchor="middle">11 per-render</text><text class="d-on-a d-small" x="85" y="223" text-anchor="middle">bu render'dan bağlanır</text>
<rect class="d-c" x="162" y="192" width="146" height="40" rx="5" /><text class="d-on-c" x="235" y="208" text-anchor="middle">AddRenderFunc</text><text class="d-on-c d-small" x="235" y="223" text-anchor="middle">plugin'lerin, per-render</text>
<text class="d-small" x="160" y="252" text-anchor="middle">bir Template.Funcs entry'si…</text>
<g class="d-dropped"><rect class="d-frame" x="12" y="260" width="146" height="40" rx="6" /><text x="85" y="276" text-anchor="middle">"pageURL"</text><text class="d-small" x="85" y="291" text-anchor="middle">parse edilir, çalışmaz</text></g>
<rect class="d-chip-hit" x="162" y="260" width="146" height="40" rx="8" /><text class="d-hit-text" x="235" y="276" text-anchor="middle">"upper"</text><text class="d-hit-text d-small" x="235" y="291" text-anchor="middle">built-in'in yerini alır</text>
</svg>
<figcaption>Her template, <code>collage.New</code> içinde tek bir function map ile parse edilir: collage'ın yirmi built-in'i, sonra plugin'lerin <code>AddTemplateFunc</code> ile eklediği fonksiyonlar, sonra <code>Config.Template.Funcs</code>; her biri bir öncekinin üzerine gelir. Built-in'lerin on biri placeholder'dır. Her render'da, her fragment için bu on bir fonksiyon o render'dan yeniden bağlanır; plugin'lerin <code>AddRenderFunc</code> ile eklediği fonksiyonlar da öyle. Geri kalan dokuz addan birini taşıyan bir entry'niz, örneğin <code>upper</code>, built-in'in yerini alır; per-render bir adı taşıyan, örneğin <code>pageURL</code>, parse edilir ama hiç çalışmaz.</figcaption>
</figure>

### slot

```html
{{slot "name"}}
```

Bu fragment'in `name` adlı slot'una bağlanmış bütün fragment'leri bağlanma sırasıyla
render eder ve markup'larını escape etmeden yerleştirir. `slot` her zaman *bu*
fragment'in slot'unu ifade eder. Bu yüzden iki fragment'in her biri `"sidebar"` adlı
bir slot'a sahip olabilir ve bu slot'lar birbirine karışmaz.

```html
<article>
  {{slot "content"}}
  <aside>{{slot "related"}}</aside>
</article>
```

- Slot'u tanımlayan, `slot`'un çağrılmasıdır. Fragment'in Go kodunda bunun için
  `WithSlot` gerekmez. Template'in çağırdığı ama hiçbir şey bağlanmamış bir slot
  hiçbir şey render etmez.
- Kontrol ters yöndedir: bu template'in hiç çağırmadığı bir slot'a bağlanmış bir
  fragment, register sırasında `ErrUnknownSlot` ile başarısız olur. Hata, slot'u ve template'in gerçekten çağırdığı slot'ları söyler.
  Aksi hâlde bağlamanın iki tarafından birindeki bir yazım hatası, sessizce eksik
  kalan bir bölüm olurdu. Literal dışında bir şeyle adlandırılan bir slot
  (`{{slot .Which}}`), o fragment için kontrolü kapatır.
- Boş bir slot hiçbir şey render etmez. `WithSlot` onu required yaptıysa durum
  farklıdır:
  - Hiçbir şey bağlanmamış ve resolver'ı olmayan required bir slot, page register
    edilirken `ErrRequiredSlotUnfilled` ile reddedilir. Böylece hiçbir zaman bir
    render'a ulaşmaz.
  - Resolver'ın doldurduğu required bir slot ise render sırasında kontrol edilir:
    hiç fragment dönmeyen bir resolver render'ı `ErrRequiredSlotEmpty` ile başarısız
    kılar.
- Bir layout'un içeriği `{{slot "content"}}` içine gelir. Bu ad
  `collage.DefaultContentSlot` içinde tutulur.

`slot` çalıştığında child fragment'lerin data handler'ları çoktan başlamıştır. Hepsi
parent render edilmeden önce aynı anda başlar. Ayrıntılar için
[Fragment'ler ve slot'lar](/docs/fragments-and-slots) sayfasına bakın.

### hoist

```html
{{hoist "head"}}
```

`area`'ya hoist edilen içeriğin nereye yerleşeceğini işaretler. Fragment'ler bu
içeriği ağacın herhangi bir yerinden tanımlar: genellikle bir data handler'dan,
program başlarken belli olan bir title için ise `WithTitle` ile.
İçeriğin nereye gideceğine ise `hoist` karar verir:

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

`hoist` içerik değil, bir marker yazar, çünkü altındaki hiçbir şey henüz render
edilmemiştir. Ağacın tamamı render edildikten sonra her marker, kendi area'sı için
tanımlanan içerikle değiştirilir. Böylece page'in derinlerinde yapılan bir
tanım da head'e ulaşır. Hiçbir şey tanımlanmamış bir area'nın marker'ı ise iz
bırakmadan kaldırılır.

`HoistTitle`, `HoistMeta`, `HoistProperty`, `HoistLink`, `HoistAlternate` ve
`HoistStylesheet` `"head"` area'sına yazar. `rc.Hoist(area, key, html)` ise adını
verdiğiniz herhangi bir area'ya yazar. **Head'e içerik ekleyen bir plugin bu
marker'a ihtiyaç duyar**: marker yoksa plugin'in içeriğinin gidecek yeri yoktur.
Ayrıntılar için [Head ve SEO](/docs/head-and-seo) sayfasına bakın.

### asset

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<img src="{{asset "/static/logo.svg"}}" alt="">
```

Bir mount'taki dosyanın content-addressed URL'sini döner. Bu URL, dosyanın path'ine
uzantıdan hemen önce içeriğinin hash'i eklenerek elde edilir.

```text
/static/app.css  →  /static/app.41014ebb6c2d9f07.css
```

Byte'lardan türetilen bir ad yalnızca o byte'ları ifade edebilir. Dosyayı bir yıllık,
`immutable` bir ömürle serve etmeyi güvenli kılan da budur: dosya değişirse adı da
değişir ve eski ad bir daha hiç istenmez.

- Path, mount'un dosyayı serve ettiği URL'dir ve mount'un prefix'ini de içerir.
- **Var olmayan bir dosya bir URL değil, hata üretir**: `ErrUnknownAsset`. Aksi
  hâlde page sorunsuz render edilir, ama 404 veren bir stylesheet'e link verirdi.
- Hiçbir mount'un serve etmediği bir path ya da hiç mount'u olmayan bir uygulama da
  `ErrUnknownAsset` üretir. Hiçbir mount'un prefix'i path'i kapsamıyorsa mesaj bunu
  söyler.
- Static export, orijinallerin yanına yalnızca page'lerin istediği fingerprint'li
  kopyaları yazar.

Go tarafında `rc.Asset(path)` aynı URL'yi döner. Ayrıntılar için
[Static asset'ler](/docs/assets) sayfasına bakın.

### stylesheet

```html
{{stylesheet "/static/gallery.css"}}
```

Bu fragment'in bir stylesheet'e ihtiyaç duyduğunu tanımlar ve çağrıldığı yerde
hiçbir şey render etmez. Layout'un `{{hoist "head"}}` çağırdığı yerde, page'in
head'ine dosyanın [`asset`](#asset) URL'siyle bir
`<link rel="stylesheet" href="…">` eklenir.

Böylece bir fragment, nerede kullanılırsa kullanılsın kendi stillerini yanında
taşıyabilir:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

Tanım path'e göre key'lenir. Bu yüzden birkaç fragment'in istediği bir
stylesheet head'de yalnızca bir kez yer alır. `stylesheet`, `asset` ile aynı
durumlarda başarısız olur. Go tarafındaki karşılığı `rc.HoistStylesheet(path)`'tir.

### csrfToken

```html
<form method="post">
  {{csrfToken}}
  <input name="email" type="email">
  <button>Subscribe</button>
</form>
```

Form'un request forgery token'ını taşıyan hidden input'u render eder:

```html
<input type="hidden" name="_csrf" value="…">
```

Yalnızca değeri değil de input'un tamamını render etmesinin bir nedeni var. Değerin
tam olarak doğru ada sahip bir field'a konması gerekir. Field adını yanlış yazan bir
form ise nedenini açıklayan hiçbir şey olmadan reddedilir.

Render sırasında yazılan değer bir placeholder'dır. Response yazılırken bu
placeholder her okuyucunun kendi token'ıyla değiştirilir.

**Cache'lenmiş** bir page'in form taşıyabilmesinin nedeni budur: cache'teki
byte'larda placeholder durur ve her okuyucu kendi token'ını alır.

- `Security.DisableCSRF` set edildiğinde `csrfToken`, render'ı `ErrCSRFDisabled` ile
  başarısız kılar. Aksi hâlde token bekleyen bir form token'sız render edilirdi.
- `Security.CSRFFieldName` farklı bir ad belirtmedikçe field'ın adı `_csrf`'tir.
  Verifier da her iki durumda bu adı okur. Bu yüzden field'ın adını değiştirdiğinizde
  template'te bir şey değiştirmeniz gerekmez.
- Static export'ta token'ı yerleştirecek bir sunucu yoktur. Bu yüzden token taşıyan
  bir page export edilmez ve rapor bunun nedenini belirtir.

Ayrıntılar için [Form'lar ve action'lar](/docs/forms-and-actions) sayfasına bakın.

### pageURL

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

`name` altında register edilmiş page'in ya da document'ın URL'sini döner. Path
pattern'ini `param value` çiftleriyle doldurur. Ada göre link vermek, bir page'in
path'i değiştiğinde link'in de onu takip etmesini sağlar.

- **Bu render'ın locale'inde çalışır.** Türkçe bir page'de `pageURL "blog-post"`
  Türkçe path'i döner. Geçerli locale'de path'i olmayan bir route için bunun yerine
  varsayılan locale'deki path'e link verilir. Böylece yalnızca İngilizce olan bir
  page'e link veren Türkçe bir page yine de render edilir.
- **Katıdır.** Bilinmeyen bir ad (`ErrUnknownRoute`) render'ı başarısız kılar. Tek sayıda
  parametre argümanı, eksik ya da boş bir parametre ve pattern'de placeholder'ı
  olmayan bir parametre de (`ErrRouteParams`) aynı sonucu verir. Oluşturulamayan bir
  link, okuyucunun karşısına çıkan bir 404 değil, development'ta bulunması
  gereken bir bug'dır.
- **Değerler string, tam sayı ya da `fmt.Stringer` olabilir** (tam sayı ve
  Stringer v0.36.0'dan beri) ve escape edilir.
- **Bazı değerler reddedilir.** Tarayıcının bir path adımı olarak
  yorumlayacağı `.` ya da `..` değeri reddedilir. Float, bool ya da nil de
  reddedilir, çünkü hiçbirinin URL'de tek bir bariz yazımı yoktur. `ErrRouteParams`
  hatası değerin tipini söyler:

```html
<a href="{{pageURL "user" "id" .ID}}">{{.Name}}</a>
```

- Hem page hem document olarak register edilmiş bir ad için tahmin yürütülmez, ad
  reddedilir.

Go tarafındaki karşılığı `app.URL(name, locale, params)`'tır. Ayrıntılar için
[Link'ler ve locale'ler](/docs/links-and-locales) sayfasına bakın.

### pageURLIn

```html
<a href="{{pageURLIn "tr" "about"}}">Hakkımızda</a>
<a href="{{pageURLIn "en" "blog-post" "slug" .Slug}}">Read in English</a>
```

`pageURL` ile aynı işi tam olarak verilen locale'de yapar ve fallback uygulamaz. O
locale'de path'i olmayan bir route `ErrNoPathInLocale` üretir. Hiçbir URL'nin
ulaşamadığı bir locale ise `ErrLocaleUnreachable` üretir. Böyle bir locale ne
varsayılan locale'dir ne de `Locale.Supported` içinde yer alır.

### localeURL

```html
{{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
{{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
```

Render edilen page'in aynı path parametreleriyle başka bir locale'deki URL'sini döner.
Bir dil seçici bununla yapılır.

- **O locale'de path'i olmayan bir page hata değil, boş string döner.** Böylece
  `{{with}}`, page'in çevrilmediği bir dili atlar.
- Hiçbir URL'nin ulaşamadığı bir locale yine de hatadır (`ErrLocaleUnreachable`).
  Çünkü bu eksik bir çeviri değil, template'teki bir hatadır.
- Bir page'e ait olmayan bir render'da başarısız olur.
- Arama motorlarının okuduğu `<link rel="alternate" hreflang>` etiketlerini Go
  tarafında `rc.HoistAlternate` ile tanımlayın. Ayrıntılar için
  [Head ve SEO](/docs/head-and-seo#canonical-and-alternate-links) sayfasına bakın.

### actionURL

```html
<form method="post" action="{{actionURL "logout"}}">{{csrfToken}}…</form>
<form method="post" action="{{actionURL "vote" "id" .ID}}">{{csrfToken}}…</form>
```

`name` adıyla register edilmiş [action'ın](/docs/forms-and-actions#an-action-at-its-own-url)
URL'sini döner (v0.36.0'dan beri). URL, `pageURL`'in bir page'inkini oluşturduğu
gibi oluşturulur: path bir kez, Go tarafında yazılır ve ona post eden her form onu
takip eder.

- **Bu render'ın locale'inde çalışır.** `pageURL` gibi varsayılan locale'e fallback
  yapar.
- **`pageURL` kadar katıdır.** Bilinmeyen bir ad `ErrUnknownRoute`, pattern'i tam
  doldurmayan parametreler `ErrRouteParams` üretir.
- **Action'ların kendi adları vardır.** `actionURL "login"` action'ı, `pageURL
  "login"` ise page'i verir, çünkü bir `login` page'i ile bir `login` action'ı
  olağan bir ikilidir. Bir page'in kendi action'ı register edildiği adla bulunur:
  `contact` adlı bir page'de `WithAction("POST", …)` için bu ad `"contact:POST"`'tur.

Go tarafındaki karşılığı `app.ActionURL(name, locale, params)`'tır.

### fragmentURL

```html
<div data-live="{{fragmentURL "home" "cpu-usage"}}">{{slot "cpu-usage"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

Bir page'in fragment'lerinden biri için
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile açtığı
path'i döner (v0.18.0'dan beri). Path page'in adından ve fragment'in adından
oluşturulur. Path bir kez, Go tarafında yazılır ve ona giden her link onu takip
eder.

- **Bu render'ın locale'inde çalışır.** `pageURL` gibi varsayılan locale'e fallback
  yapar.
- **`pageURL` kadar katıdır.** Bilinmeyen bir page `ErrUnknownRoute`, page'in
  açmadığı bir fragment `ErrUnknownFragmentPath`, bir locale'de iki path'te açılmış
  bir fragment `ErrAmbiguousFragmentPath`, pattern'i doldurmayan parametreler ise
  `ErrRouteParams` olur. Her biri render'ı başarısız kılar.

Go tarafındaki karşılığı `app.FragmentURL(page, fragment, locale, params)`'tır.
Ayrıntılar için [Link'ler ve locale'ler](/docs/links-and-locales#a-fragments-url)
sayfasına bakın.

### fragmentURLIn

```html
<div data-live="{{fragmentURLIn "tr" "home" "cpu-usage"}}">…</div>
```

Tam olarak verilen locale'de, fallback olmadan çalışan `fragmentURL`'dir. O
locale'de path'i olmayan bir fragment `ErrNoPathInLocale`, hiçbir URL'nin
ulaşamadığı bir locale ise `ErrLocaleUnreachable` olur.

### safeHTML

```html
{{safeHTML .RenderedMarkdown}}
```

Bir string'i güvenilir HTML olarak işaretler. Böylece `html/template` onu escape
etmeden yerleştirir. Bu bir kaçış yoludur: yalnızca kendi ürettiğiniz ya da kendiniz
sanitize ettiğiniz markup için kullanın. Kullanıcının yazdığı hiçbir şey için
kullanmayın.

### safeURL

```html
<a href="{{safeURL .ExternalLink}}">Visit</a>
```

Bir string'i güvenilir URL olarak işaretler ve `html/template`'in URL
sanitization'ını devre dışı bırakır. Bu sanitization normalde güvenmediği bir
scheme'i, örneğin `javascript:`'i, `#ZgotmplZ` ile değiştirir. Yalnızca validate
ettiğiniz URL'ler için kullanın.

### dict

```html
{{template "card" dict "Title" .Title "URL" (pageURL "post" "slug" .Slug)}}
```

Sırayla gelen key ve value'lardan bir `map` oluşturur. Bir sub-template'e birden
fazla değer geçirmek için kullanılır. Key'ler string olmalıdır ve
argümanlar çiftler hâlinde gelmelidir. Bu hatalardan herhangi biri template'i
başarısız kılar; mesaj hangisi olduğunu söyler.

### default

```html
<h2>{{default "Untitled" .Subtitle}}</h2>
<h2>{{.Subtitle | default "Untitled"}}</h2>
```

`value`'yu döner. `value` boş string ise `fallback`'i döner. İki argüman da
string'dir. Fallback'in önce gelmesinin nedeni, pipe'lı kullanımın doğal okunmasıdır.
Çünkü bir pipeline, değerini son argüman olarak geçirir.

### upper ve lower

```html
<span class="badge">{{upper .Status}}</span>
<code>{{lower .Code}}</code>
```

`strings.ToUpper` ve `strings.ToLower` ile aynıdır.

### title

```html
<h1>{{title .Name}}</h1>
```

Her kelimenin ilk harfini büyük, geri kalan harflerini küçük yapar. Bir kelime, art
arda gelen harflerden oluşur. Bu yüzden harf olmayan her karakter, yani boşluk, tire
ya da kesme işareti, yeni bir kelime başlatır:

| Girdi | Çıktı |
| --- | --- |
| `hello world` | `Hello World` |
| `iPHONE case` | `Iphone Case` |
| `o'neil-smith` | `O'Neil-Smith` |

### join

```html
<p>Tags: {{join ", " .Tags}}</p>
```

`strings.Join(items, sep)` ile aynıdır, yalnızca separator önce gelir. `items` bir
`[]string` olmalıdır.

### formatTime

```html
<time datetime="{{formatTime .Published "2006-01-02"}}">
  {{formatTime .Published "2 January 2006"}}
</time>
```

[Reference time layout'u](https://pkg.go.dev/time#pkg-constants) ile
`t.Format(layout)` çağırır. `t` bir `time.Time`'dır. Zaman, taşıdığı location'a göre
formatlanır. Başka bir location istiyorsanız zamanı data handler'da dönüştürün.

## Kendi fonksiyonlarınızı eklemek

Fonksiyonlarınızı `New`'dan önce `Config.Template.Funcs` ile ekleyin:

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

Bu fonksiyonlar built-in'lerin üzerine merge edilir. Bu yüzden built-in bir adı
taşıyan entry, o built-in'in yerini alır. Her render'da bağlanan fonksiyonlar bunun
dışındadır.

Fonksiyonların `New`'dan önce tanımlı olması gerekir, çünkü `html/template`
yalnızca template parse edilirken function map'te bulunan bir adı
çağırabilir. Kimsenin register etmediği bir adı çağıran template, ilk request'te
değil, `New`'da hata verir.

Bir plugin de fonksiyonları aynı şekilde, `Configure` aşamasında ekler. Aynı adı
taşıyan bir entry uygulamada da varsa uygulamanınki geçerli olur. Ayrıntılar için
[Plugin yazmak](/docs/writing-plugins#template-functions) sayfasına bakın.

Bir plugin'in `AddRenderFunc` ile eklediği fonksiyon iki yönden de istisnadır:

- Yukarıdaki on bir fonksiyon gibi her render'da bağlanır. Bu yüzden uygulamanın
  aynı adı taşıyan entry'si parse edilir, ama hiçbir zaman çağrılmaz.
- Yirmi built-in addan hiçbirini alamaz. `AddRenderFunc` bunları
  `ErrDuplicateTemplateFunc` ile reddeder.

Request'e ihtiyaç duyan her şey bir fonksiyona değil, bir data handler'a aittir.
Page'in verisi de zaten oradan gelir.


## Örnek: iki dilde bir ürün page'i

[Link'ler ve locale'ler](/docs/links-and-locales#example-a-bilingual-product-catalogue)
sayfasındaki dükkân İngilizce ve Türkçe satış yapar. Ürün page'i, etrafındaki
layout ile birlikte, on bir per-render fonksiyonun dokuzunu çağırır:

| Fonksiyon | Örnekteki işi |
| --- | --- |
| `asset` | dükkânın stylesheet'i ve path'i veriden gelen ürün resmi |
| `hoist` | title'ın ve `product.css`'in head'de yerleşeceği yer |
| `localeURL` | dil seçici |
| `slot` | layout'un içeriği ve ürünün stok kutusu |
| `stylesheet` | yalnızca bu page'in ihtiyaç duyduğu `product.css` |
| `pageURL` | kataloğa geri dönen link |
| `fragmentURL` | stok kutusunu yenileyen bir script için kutunun kendi URL'si |
| `actionURL` | "sepete ekle" form'unun hedefi |
| `csrfToken` | o form'un token'ı |

`catalog` paketi sizin veri katmanınızın yerini tutar ve her ürünün adını iki dilde
saklar. Handler, render'ın locale'ine göre etiketleri seçer ve ürünün adını title
olarak hoist eder:

```go
type labels struct {
	Back string
	Add  string
}

var labelsIn = map[string]labels{
	"en": {Back: "All products", Add: "Add to basket"},
	"tr": {Back: "Tüm ürünler", Add: "Sepete ekle"},
}

type productView struct {
	catalog.Product // ID, Name, Image
	Labels          labels
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (productView, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Find(ctx, id, rc.Locale)
	if errors.Is(err, catalog.ErrNoProduct) {
		return productView{}, nil, fmt.Errorf("catalog: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return productView{}, nil, err
	}
	rc.HoistTitle(p.Name)
	return productView{Product: p, Labels: labelsIn[rc.Locale]}, []string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (int, error) {
	return catalog.Stock(ctx, rc.Param("id"))
}
```

Page'in, stok kutusunun ve sepet action'ının her locale'de bir path'i vardır:

- Katalog page'i, Link'ler ve locale'ler sayfasının register ettiği page'dir.
- `addToBasket`, [Form'lar ve action'lar](/docs/forms-and-actions) sayfasındaki
  gibi action'ın handler'ıdır.
- `static/`, [Static asset'ler](/docs/assets) sayfasındaki gibi `/static/`
  altına mount edilir.
- Ürün `Required`'dır: o olmadan page'in gösterecek bir şeyi yoktur.

```go
layout := collage.NewFragment("layout", "layouts/master.html").Build()

stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

product := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(collage.NewFragment("product", "pages/product.html").
		WithData(collage.LoadTagged(loadProduct)).
		WithSlotFragment("stock", stock).
		Required().
		Build()).
	WithPath("en", "/products/{id}").
	WithPath("tr", "/urunler/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	WithFragmentPath("tr", "/urunler/{id}/stok", stock).
	Build()

if err := app.Register(catalogue, product); err != nil {
	return err
}
return app.RegisterAction(collage.NewAction("basket-add").
	WithPath("en", "/basket/{id}").
	WithPath("tr", "/sepet/{id}").
	WithMethods(http.MethodPost).
	WithHandler(addToBasket).
	Build())
```

Layout dükkânın stylesheet'ine link verir, head'in area'sını işaretler ve dil
seçiciyi taşır:

```html
<!-- templates/layouts/master.html -->
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
  {{hoist "head"}}
</head>
<body>
  <nav>
    {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
    {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
  </nav>
  {{slot "content"}}
</body>
</html>
```

Ürünün template'i link verdiği her page'i, fragment'i ve action'ı adıyla anar, hiçbir
zaman path yazmaz:

```html
<!-- templates/pages/product.html -->
{{stylesheet "/static/product.css"}}
<article class="product">
  <a href="{{pageURL "catalogue"}}">{{.Labels.Back}}</a>
  <h1>{{.Name}}</h1>
  <img src="{{asset .Image}}" alt="">
  <div data-live="{{fragmentURL "product" "product-stock" "id" .ID}}">{{slot "stock"}}</div>
  <form method="post" action="{{actionURL "basket-add" "id" .ID}}">
    {{csrfToken}}
    <button>{{.Labels.Add}}</button>
  </form>
</article>
```

```html
<!-- templates/fragments/stock.html -->
<p class="stock">{{.}}</p>
```

### Okuyucuya ulaşan

`/tr/urunler/42` için gelen request 200 ile cevaplanır. Boş satırlar çıkarılmış ve
token kısaltılmış hâliyle page şöyledir:

```html
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="/static/app.6d6068180a5c710c.css">
  <title>Pirinç güverte lambası</title><link rel="stylesheet" href="/static/product.e38a7cd1dac68eec.css">
</head>
<body>
  <nav>
    <a hreflang="en" href="/products/42">English</a>
    <a hreflang="tr" href="/tr/urunler/42">Türkçe</a>
  </nav>
<article class="product">
  <a href="/tr/urunler">Tüm ürünler</a>
  <h1>Pirinç güverte lambası</h1>
  <img src="/static/products/42.04115f7d8dde41ce.svg" alt="">
  <div data-live="/tr/urunler/42/stok"><p class="stock">12</p>
</div>
  <form method="post" action="/tr/sepet/42">
    <input type="hidden" name="_csrf" value="…">
    <button>Sepete ekle</button>
  </form>
</article>
</body>
</html>
```

Aynı template'ler `/products/42`'yi İngilizce render eder. Ad ve etiketler veriden
gelir; URL'ler render'ın locale'ine göre değişir:

| Template'te | `/products/42`'de | `/tr/urunler/42`'de |
| --- | --- | --- |
| `pageURL "catalogue"` | `/products` | `/tr/urunler` |
| `fragmentURL "product" "product-stock" "id" .ID` | `/products/42/stock` | `/tr/urunler/42/stok` |
| `actionURL "basket-add" "id" .ID` | `/basket/42` | `/tr/sepet/42` |
| `localeURL "en"`, `localeURL "tr"` | `/products/42`, `/tr/urunler/42` | aynı |
| `asset .Image` | `/static/products/42.04115f7d8dde41ce.svg` | aynı |

- **Title ve `product.css` marker'a yerleşir.** İkisi de layout'un marker'ın
  üstüne yazdığı `app.css`'ten sonra gelir. Önce title gelir: handler onu template çalışmadan önce
  tanımlamıştır.
- **`asset`, veriden gelen bir path'i de literal bir path kadar kolay alır.**
  Mount'ta olmayan bir resim `ErrUnknownAsset`'tir. Ürün required olduğu için page,
  kırık bir resim göstermek yerine 500 döner.
- **Token her okuyucunun kendisinindir.** Bu page cache'lenseydi cache'teki byte'lar
  placeholder'ı taşırdı ve her okuyucu onun yerine kendi token'ını alırdı.
