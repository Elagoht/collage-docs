---
description: Template'lerin nerede durduğu, nasıl yüklenip binary'ye embed edildiği, development'ta nasıl yeniden yüklendiği; slot'lar, fonksiyonlar ve escaping.
reference: TemplateConfig, Config, ErrUnknownSlot, ErrTemplateType
---

# Template'ler

Her fragment bir template render eder. Template'ler Go'nun
[`html/template`](https://pkg.go.dev/html/template) paketiyle yazılır: sözdizimi
de, bağlama göre yapılan escaping de aynıdır. Buna, bir fragment'in ihtiyaç duyduğu
şeyler için birkaç fonksiyon eklenmiştir. Bu fonksiyonlar fragment'in slot'larını,
diğer page'lere link'leri, asset'leri, form'ları ve page'in `<head>`'ini kapsar.

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  <p class="meta">{{formatTime .Published "2 January 2006"}}</p>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
```

## Template'ler nerede durur

Template'lerin nerede bulunacağını `Config.Template` belirler:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		Root:      "templates",
		Extension: ".html",
	},
})
```

`Root` altındaki, uzantısı `Extension` olan her dosya bir template'tir. `Root`'un
varsayılan değeri `./templates`, `Extension`'ınki ise `.html`'dir. Başka uzantılı
dosyalar yok sayılır. Bu yüzden template'lerin yanında duran bir `README.md` sorun
çıkarmaz.

Bir template'in adı, `Root`'a göre göreli yoludur ve uzantıyı da içerir. Fragment'te
yazdığınız ad da budur:

```text
templates/
├── layouts/master.html      →  "layouts/master.html"
├── pages/post.html          →  "pages/post.html"
└── fragments/author.html    →  "fragments/author.html"
```

```go
collage.NewFragment("author", "fragments/author.html")
```

### Bir kez yüklenir, erkenden kontrol edilir

`collage.New`, daha hiçbir şey serve edilmeden bütün template'leri parse edip tek bir
set'e toplar. Her hata türü, yakalanabileceği en erken noktada yakalanır:

- **Var olmayan bir root** olduğunda `New`, `ErrTemplateRootMissing` ile hata döner.
  Bu, açılışta en sık karşılaşılan hatadır. Parse edilemeyen bir template'ten ayırt
  edebilmeniz için ayrı bir hatadır.
- **Parse edilemeyen bir template** ya da register edilmemiş bir fonksiyonu çağıran bir
  template olduğunda `New` hata döner ve hatada dosyanın adı yer alır.
- **Yüklenmemiş bir template'i gösteren bir fragment** olduğunda `RegisterPage`,
  `ErrTemplateNotFound` ile hata döner. Hata page'i, fragment'i ve yolu belirtir.
- **Template'inin hiç çağırmadığı bir slot'a bağlanmış bir fragment** olduğunda
  `RegisterPage`, `ErrUnknownSlot` ile hata döner. Ayrıntılar aşağıda,
  [Slot'lar](#slots) bölümündedir.
- **Verisine uymayan bir template** olduğunda (verinin tipinde yalnızca `Title`
  varken `{{.Titel}}` ya da yanlış sayıda argümanla çağrılan bir method gibi)
  `RegisterPage`, `ErrTemplateType` ile hata döner. Hata dosyayı, satırı ve
  sütunu belirtir (v0.49.0'dan beri). Bkz.
  [Template'ler nasıl kontrol edilir](/docs/data-handlers#how-templates-are-checked).
- **Çalışırken hata veren bir template**, örneğin tipi render'dan önce bilinmeyen
  bir verinin alanına erişen, nil bir pointer'a takılan ya da hata dönen bir
  fonksiyonu çağıran bir template, kendi fragment'ini başarısız kılar.
  Bu durumda fragment'in
  [failure policy'si](/docs/fragments-and-slots#when-a-fragment-fails) devreye
  girer. Çıktı buffer'lanır. Bu yüzden yarıda hata veren bir template yarım bir
  fragment yazmaz, hiçbir şey yazmaz.

Bütün template'ler tek bir set'te olduğu için bir template başka bir template'i
adıyla include edebilir:

```html
{{template "partials/byline.html" .}}
```

`{{define}}` ile verdiğiniz adlar, bütün dosyalarda bu tek namespace'i paylaşır. Bu
yüzden ayırt edici adlar seçin.

Yalnızca `Root` verildiğinde template'ler diskten `os.OpenRoot` üzerinden okunur.
Dolayısıyla dizinin dışına çıkan bir symlink takip edilmez, reddedilir: `New`,
`ErrTemplateEscapesRoot` ile hata döner ve hatada dosyanın adı yer alır.

### Template'leri binary'ye embed etmek

Template'lerini `./templates`'ten okuyan bir binary, yalnızca bu dizinin bulunduğu
bir yerden çalışır. Template'leri embed ederseniz binary her yerden çalışır: bir
container'da, bir systemd unit'inde ya da bir sunucuya kopyalanmış hâliyle.

```go
//go:embed all:templates
var templatesFS embed.FS

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
})
```

`FS` verildiğinde `Root`, o filesystem içindeki bir dizini gösterir. `Root` yine her
template adının başından çıkarılır ve gerekli olmasının sebebi de budur. `embed.FS`
dosyaları source tree'deki yollarıyla adlandırır. `Root` olmasaydı her template'in
adı `templates/pages/post.html` gibi olurdu.

`FS` ile birlikte boş bir `Root`, filesystem'in kökü anlamına gelir. `all:`
prefix'i, adı `.` ya da `_` ile başlayan dosyaların da embed edilmesini sağlar. Düz
bir `//go:embed templates` bu dosyaları atlar.

`collage new` aynısını tek bir filesystem olan `files` ile oluşturur. Bu filesystem
`templates`, `static` ve `plugins-config.json`'ı birlikte embed eder (v0.68.0'dan
beri) ve `Template.FS` olarak verilir.

## Development'ta yeniden yükleme

Development'ta, yani `Config.DevMode` açıkken, her render'dan önce bütün
template'ler yeniden parse edilir. Bir template'i kaydettiğinizde bir sonraki
request onu kullanır. Framework `COLLAGE_DEV`'i okumaz.
Bu değişkeni `collage dev` set eder, scaffold edilen `main.go` da onu
`Config.DevMode`'a çevirir.

Embed edilmiş bir set bunu yapamaz. Byte'ları binary build edildiği anda
sabitlenmiştir ve onları yeniden parse etmek hiçbir şeyi değiştirmez.

Bu yüzden `Root`, çalışma dizinine göre bir dizin olarak var olduğunda
**development'ta diskteki dizin embed edilmiş kopyanın önüne geçer** ve uygulama
bu seçimi log'lar. `collage dev` proje dizininizde çalıştığı için bu dizin orada
her zaman vardır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 254" role="img" aria-labelledby="tpl-src-title">
<title id="tpl-src-title">collage.New template'leri nereden okur. Template.FS verilmemişse Root dizinini diskten os.OpenRoot üzerinden okur. FS verilmiş ve DevMode kapalıysa embed edilmiş FS'i okur. FS verilmiş ve DevMode açıksa, Root çalışma dizinine göre bir dizinse diskteki Root'u okur ve bunu log'lar; değilse embed edilmiş FS'i okur ve değişiklikler görünmez. Development'ta her render, okunan set'i yeniden parse eder.</title>
<defs><marker id="tpl-src-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">collage.New</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#tpl-src-arrow)" />
<rect class="d-frame" x="40" y="52" width="150" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">Template.FS var mı?</text>
<line class="d-flow" x1="190" y1="65" x2="221" y2="65" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="60" text-anchor="middle">hayır</text>
<rect class="d-a" x="225" y="50" width="93" height="30" rx="5" /><text class="d-on-a d-small" x="271.5" y="62" text-anchor="middle">diskteki Root</text><text class="d-on-a d-small" x="271.5" y="74" text-anchor="middle">os.OpenRoot</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="90">evet</text>
<rect class="d-frame" x="40" y="98" width="150" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">DevMode açık mı?</text>
<line class="d-flow" x1="190" y1="111" x2="221" y2="111" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="106" text-anchor="middle">hayır</text>
<rect class="d-b" x="225" y="96" width="93" height="30" rx="5" /><text class="d-on-b d-small" x="271.5" y="108" text-anchor="middle">embed FS</text><text class="d-on-b d-small" x="271.5" y="120" text-anchor="middle">build'de sabit</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="136">evet</text>
<rect class="d-frame" x="40" y="144" width="150" height="26" rx="6" /><text x="115" y="161" text-anchor="middle">Root diskte dizin mi?</text>
<line class="d-flow" x1="190" y1="157" x2="221" y2="157" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="152" text-anchor="middle">hayır</text>
<rect class="d-b" x="225" y="142" width="93" height="30" rx="5" /><text class="d-on-b d-small" x="271.5" y="154" text-anchor="middle">embed FS</text><text class="d-on-b d-small" x="271.5" y="166" text-anchor="middle">eski kalır</text>
<line class="d-flow" x1="115" y1="170" x2="115" y2="186" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="182">evet</text>
<rect class="d-a" x="40" y="190" width="150" height="34" rx="5" /><text class="d-on-a" x="115" y="205" text-anchor="middle">diskteki Root</text><text class="d-on-a d-small" x="115" y="218" text-anchor="middle">log'lanır · güncel</text>
<text class="d-small" x="160" y="244" text-anchor="middle">Development'ta her render set'i yeniden parse eder.</text>
</svg>
<figcaption>Bir uygulamanın template'lerin hangi kopyasını render edeceği <code>collage.New</code>'da belirlenir. <code>Template.FS</code> yoksa bu, diskteki <code>Root</code> dizinidir. Bir <code>FS</code> varsa embed edilmiş kopyadır. Tek istisna development'tır: <code>Root</code> çalışma dizinine göre bir dizin olarak da varsa onun yerine o dizin okunur ve uygulama bunu log'lar. Development'ta her render, seçilen kopyayı yeniden parse eder. Böylece diskte yapılan bir değişiklik bir sonraki request'te görünür, okunmayan bir dizinde yapılan değişiklik ise ancak binary yeniden build edildiğinde görünür.</figcaption>
</figure>

Bir değişikliğin hemen görünmesini iki şey daha sağlar:

- Development'ta page cache hiç okunmaz. Böylece değişiklikten önce cache'lenmiş bir
  page, değişikliği gizlemez.
- Development'taki her page, küçük bir script içerir. Bu script bir template ya da
  static bir dosya değiştiğinde tarayıcıyı yeniler. v0.10.0'dan itibaren
  [`Config.DevWatch`](/docs/configuration#devwatch)'ta belirttiğiniz bir dizindeki
  dosyalar da buna dahildir. Bu, handler'larınızın diskten okuduğu Markdown gibi
  içerikler içindir.

Bütün set yeniden parse edildiği için herhangi bir template'teki bir syntax hatası,
düzeltilene kadar bütün render'ları başarısız kılar. Bu development'ta olur.
Hatayı orada hemen, dosya ve satır numarasıyla birlikte görürsünüz.

Register sırasında yapılan kontroller, verinin tipine göre yapılan kontrol de
dahil, açılıştan sonra düzenlenen bir template için bir sonraki yeniden başlatmaya
kadar yeniden yapılmaz. Böyle bir template'in hatası bunun yerine development
error page'inde görünür; önce template, satır ve neden gelir.

## Bir template ne alır

`.`, tam olarak fragment'in data handler'ının döndürdüğü değerdir ya da
`collage.Value`'nun verdiği değerdir. Tipi de handler'ın ya da değerin belirttiği
Go tipidir. Verisi olmayan ya da handler'ı `collage.Effect` ile uyarlanmış bir
fragment, veri olmadan render edilir. Bkz.
[Data handler'lar](/docs/data-handlers).

```go
type postView struct {
	Title     string
	Published time.Time
	Body      template.HTML // already sanitised, see below
	Tags      []string
}
```

```html
<h1>{{.Title}}</h1>
{{with .Tags}}<p>Tagged {{join ", " .}}</p>{{end}}
```

Bir template yalnızca kendi fragment'inin verisini görür. Parent'ın verisine
child'dan erişilemez ve global bir site nesnesi yoktur. Sitenin adı gibi her
template'in ihtiyaç duyduğu bir şey, ya fragment'ine verilen veridir ya da sizin
register ettiğiniz bir fonksiyondur.

## Slot'lar

`{{slot "name"}}`, fragment'in o adı taşıyan slot'unda ne varsa sırasıyla render eder
ve HTML olarak ekler. Child'ların çıktısı ikinci kez escape edilmez. Her child,
render edilirken kendi değerlerini zaten escape etmiştir.

```html
<main>{{slot "content"}}</main>
<aside>{{slot "sidebar"}}</aside>
```

- Bir slot'u tanımlayan, onun çağrılmasıdır. Fragment'in Go kodunda bunun için
  `WithSlot` gerekmez. `WithSlot` yalnızca bir slot'u zorunlu kılar ya da onu tek
  bir fragment'le sınırlar.
- Boş bir slot hiçbir şey render etmez. Ona daha önce bir şey bağlanmış olup
  olmaması bunu değiştirmez.
- Template'inin hiç çağırmadığı bir slot'a bağlanmış bir fragment, register
  sırasında `ErrUnknownSlot` ile başarısız olur. Hata, slot'u ve template'in
  gerçekten çağırdığı slot'ları söyler. Aksi hâlde bağlamanın iki tarafından
  birindeki bir yazım hatası, page'den sessizce eksilen bir bölüm olurdu.
  Template'in include ettiği bir template'teki ya da tanımladığı bir block'taki
  çağrılar da sayılır. Bir slot'u literal dışında bir şeyle (`{{slot .Which}}`)
  adlandıran bir template bunlardan herhangi birini çağırabilir. Bu yüzden onun
  fragment'i kontrol edilmez.
- Template'in atladığı bir `{{if}}` içindeki slot render edilmez. Ancak o slot'taki
  fragment'lerin data handler'ları zaten başlatılmıştır. Ayrıntılar için
  [Fragment'ler ve slot'lar](/docs/fragments-and-slots#a-slot-the-template-skips)
  sayfasına bakın.

`slot` fonksiyonu, template'i çalıştıran fragment'e aittir. Bu yüzden aynı
`{{slot "content"}}`, onu yazan her fragment'te farklı bir slot anlamına gelir.

## Built-in fonksiyonlar

Aşağıdaki fonksiyonlar, `html/template`'in kendi fonksiyonlarıyla (`printf`, `len`,
`index`, `eq` ve diğerleri) birlikte her template'te kullanılabilir:

| Fonksiyon | Ne yapar |
| --- | --- |
| `slot "name"` | Bu fragment'in slot'larından birindeki fragment'leri render eder |
| `pageURL "name" "param" value …` | Register edilmiş bir page'in ya da document'ın, bu render'ın locale'indeki URL'sini verir |
| `pageURLIn "locale" "name" …` | Aynısını, tam olarak verilen locale'de verir |
| `localeURL "locale"` | Bu page'in başka bir locale'deki URL'sini verir. Page'in o locale'de bir path'i yoksa boş döner |
| `actionURL "name" …` | Bir action'ın, bu render'ın locale'indeki URL'sini verir |
| `fragmentURL "page" "fragment" …` | Bir fragment path'inin, bu render'ın locale'indeki URL'sini verir |
| `fragmentURLIn "locale" "page" "fragment" …` | Aynısını, tam olarak verilen locale'de verir |
| `asset "/static/app.css"` | Mount edilmiş bir dosyanın content-addressed URL'sini verir |
| `stylesheet "/static/app.css"` | Page'in `<head>`'i için bir stylesheet tanımlar |
| `hoist "head"` | Page'in bir alanı için yapılan tanımların yerleşeceği yeri belirler |
| `csrfToken` | Form'un forgery token'ını taşıyan hidden input'u üretir |
| `safeHTML`, `safeURL` | Bir string'i güvenilir HTML ya da URL olarak işaretler. Escaping'i atlamanın yoludur |
| `dict "key" value …` | `{{template}}`'e birden fazla değer geçirmek için bir map oluşturur |
| `default fallback value` | `value`'yu, boşsa `fallback`'i döner. Yalnızca string'lerle çalışır, başka bir tip render'ı başarısız kılar |
| `upper`, `lower`, `title` | Büyük/küçük harf dönüşümü yapar |
| `join sep items` | Bir `[]string`'i birleştirir |
| `formatTime t layout` | Bir `time.Time`'ı bir Go layout'uyla formatlar |
| `withQuery "page" 2 …` | Query parametreleri set edilmiş hâliyle bu page'in URL'si; `""` bir parametreyi kaldırır |
| `pathEscape s` | Bir string'i tek bir path segment'i için escape eder |
| `json v`, `pre v` | Bir değerin JSON hâli, ya da bir template'e ne verildiğini görmek için escape edilmiş bir `<pre>` bloğu |
| `add`, `sub`, `mul`, `div`, `mod`, `min`, `max`; `addf`, `subf`, `mulf`, `divf` | Tam sayı ve `float64` aritmetiği |
| `trim`, `replace`, `contains`, `hasPrefix`, `hasSuffix`, `split`, `truncate` | String'i en sonda alan, bu yüzden pipeline içinde doğal okunan string fonksiyonları |
| `first`, `last`, `seq`, `keys`, `hasKey`, `dig` | Bir listenin uçları, üzerinde sayılacak bir aralık, bir map'in key'leri, verinin derinindeki bir değer |

`withQuery`'den aşağıdaki fonksiyonlar v0.77.0'da geldi. Link fonksiyonları
katıdır. Bilinmeyen bir page adı ya da eksik bir parametre, bozuk
bir link üretmek yerine render'ı başarısız kılar. Her fonksiyonun argümanlarını ve
davranışını [Template fonksiyonları](/docs/template-functions) sayfasında
bulabilirsiniz.

## Kendi fonksiyonlarınızı eklemek

`Config.Template.Funcs`, bütün template'lere fonksiyon ekler. Bu alan bir
`html/template` `FuncMap`'idir:

```go
import "html/template"

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		FS:   templatesFS,
		Root: "templates",
		Funcs: template.FuncMap{
			"money": func(cents int64) string {
				return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
			},
			"readingTime": func(words int) string {
				return fmt.Sprintf("%d min read", max(1, words/200))
			},
		},
	},
})
```

```html
<p>{{money .PriceCents}} · {{readingTime .WordCount}}</p>
```

- **`New`'dan önce set edin.** `html/template` yalnızca adı template parse edilirken
  bilinen bir fonksiyonu çağırabilir ve parse işlemi `New`'da yapılır. Register
  edilmemiş bir adı çağıran template `New`'da hata verir. Amaç da tam olarak budur:
  yazım hatasını uygulama açılırken yakalarsınız.
- **Built-in bir fonksiyonun adıyla eklenen bir entry, o built-in'in yerini alır.**
  Render'a bağlı olan on iki fonksiyon bunun dışındadır: `slot`, `hoist`, `asset`,
  `stylesheet`, `csrfToken`, `pageURL`, `pageURLIn`, `localeURL`, `actionURL`,
  `fragmentURL`, `fragmentURLIn` ve `withQuery`. Render engine her render için
  bunların kendi versiyonlarını bağlar. Bu yüzden bunlardan birini override etmek
  kabul edilir, ama hiçbir etkisi olmaz; v0.77.0'dan beri `New` onu adıyla anan bir
  uyarı log'lar.
- **Bir fonksiyon request'i göremez.** Fonksiyon, bütün program için bir kez register
  edilir. Request'e, locale'e ya da kullanıcıya bağlı olan her şey data handler'da
  olmalıdır. Veri zaten oradan gelir.

Bir plugin de `Configure` hook'undan fonksiyon ekleyebilir. Ayrıntılar için
[Plugin yazmak](/docs/writing-plugins) sayfasına bakın. Bir plugin ve `Funcs` aynı
adı tanımlarsa `Funcs` kazanır, çünkü ikisini birden görebilen uygulamadır. Bir
plugin built-in bir adı ise hiç alamaz (v0.77.0'dan beri).

## Escaping

`html/template` her değeri, göründüğü yere göre escape eder. Bu yer HTML metni, bir
attribute, bir URL, inline JavaScript ya da CSS olabilir. Başlığı
`<script>alert(1)</script>` olan bir yazı düz metin olarak basılır. Bir `href`
içindeki `javascript:` URL'si de zararsız bir URL ile değiştirilir. Hiçbir şeyi elle
escape etmezsiniz.

Bazen bir değer gerçekten HTML'dir. Örneğin CMS'inizin sanitize ettiği bir yazı
gövdesi ya da kendi kodunuzun ürettiği markup böyledir. Bunu belirtmenin iki yolu
vardır:

- View'ınızda alana `template.HTML` tipini verin, yukarıdaki `Body` gibi. Bu durumda
  kararı tip taşır. Karar Go'da verilir ve orada review edilebilir.
- Template'te `safeHTML` (URL için `safeURL`) kullanın.

İkisi de o değer için escaping'i kapatır. **Bunları yalnızca kendi ürettiğiniz ya da
sanitize ettiğiniz içerikte kullanın.** Kullanıcının yazdığı hiçbir şeyde
kullanmayın. Escape edilmemiş bir kullanıcı değeri bir cross-site scripting (XSS)
açığıdır.

İki built-in, olduğu gibi eklenen HTML üretir:

- `{{slot}}`. Child'ları kendi değerlerini zaten escape etmiştir.
- `{{hoist}}`. Fragment'lerin page için tanımladıklarını yazar. `rc.HoistTitle`,
  `rc.HoistMeta` ve benzeri helper'lar kendilerine verilen değeri escape eder. Bunların
  altındaki `rc.Hoist` ise markup'ı tam yazıldığı gibi ekler, yani orada escaping
  sizin sorumluluğunuzdadır. Ayrıntılar için [Head ve SEO](/docs/head-and-seo)
  sayfasına bakın.

`pageURL` ve benzerleri, path'e yerleştirdikleri parametre değerlerini escape eder.
Ayrıca `.` ya da `..` değerini reddeder, çünkü tarayıcı bunları path'te bir üst
seviyeye çıkmak olarak yorumlar.

## Örnek: bir blog yazısı

Bir blogun yazı page'i, bu sayfadakilerin çoğunu bir arada kullanır. Her kısım page'e
farklı bir yoldan ulaşır:

| Kısım | Nerede tanımlanır | Page'e nereden girer |
| --- | --- | --- |
| Yazı | `pages/post.html` | layout'un `{{slot "content"}}` çağrısı |
| Yazar kartı | `fragments/author.html` | yazının `{{slot "author"}}` çağrısı |
| Yazının stylesheet'i | yazıdaki `{{stylesheet}}` | layout'un `{{hoist "head"}}` çağrısı |
| `<title>` ve description | handler'daki `rc.HoistTitle`, `rc.HoistMeta` | layout'un `{{hoist "head"}}` çağrısı |
| Okurların yorumları | yazının verisi | escape edilen `{{.Text}}` |

Handler'lar yazının slug'ını path'ten okur. `blog` paketi, kendi veri katmanınızın
yerini tutar:

```go
type comment struct {
	Name string
	Text string
}

type postView struct {
	Title     string
	Published time.Time
	Body      template.HTML // rendered and sanitised by your own code
	Comments  []comment
}

type authorView struct {
	Name string
	Bio  string
}

func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	slug := rc.Param("slug")
	p, err := blog.GetPost(ctx, slug)
	if err != nil {
		return postView{}, nil, err
	}
	rc.HoistTitle(p.Title)
	rc.HoistMeta("description", p.Summary)

	view := postView{Title: p.Title, Published: p.Published, Body: p.Body}
	for _, c := range p.Comments {
		view.Comments = append(view.Comments, comment{Name: c.Name, Text: c.Text})
	}
	return view, []string{"post:" + slug}, nil
}

func loadAuthor(ctx context.Context, rc *collage.RenderContext) (authorView, error) {
	a, err := blog.AuthorOf(ctx, rc.Param("slug"))
	if err != nil {
		return authorView{}, err
	}
	return authorView{Name: a.Name, Bio: a.Bio}, nil
}
```

Yazar kartı kendi başına bir fragment'tir ve yazının `author` slot'una bağlanır:

```go
layout := collage.NewFragment("layout", "layouts/master.html").Build()

author := collage.NewFragment("author", "fragments/author.html").
	WithData(collage.Load(loadAuthor)).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithSlotFragment("author", author).
	Required().
	Build()

page := collage.NewPage("post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/posts/{slug}").
	Build()

if err := app.RegisterPage(page); err != nil {
	return err
}
```

Üç template:

```html
<!-- templates/layouts/master.html -->
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
<body>{{slot "content"}}</body>
</html>
```

```html
<!-- templates/pages/post.html -->
{{stylesheet "/static/post.css"}}
<article>
  <h1>{{.Title}}</h1>
  <p class="meta">{{formatTime .Published "2 January 2006"}}</p>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
  <section class="comments">
    {{range .Comments}}<p><b>{{.Name}}</b> {{.Text}}</p>{{end}}
  </section>
</article>
```

```html
<!-- templates/fragments/author.html -->
<aside class="author">
  <strong>{{.Name}}</strong>
  <p>{{.Bio}}</p>
</aside>
```

`/static/post.css` bir [mount](/docs/assets) içindeki bir dosya olmalıdır.
`{{stylesheet}}` onun content-addressed URL'sine link verir. Hiçbir mount'un serve
etmediği bir path render'ı başarısız kılar.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 254" role="img" aria-labelledby="tpl-post-title">
<title id="tpl-post-title">Render edilmiş blog yazısı page'i. Layout'un hoist head işareti, page'in tamamı render edildikten sonra doldurulur: yazının handler'ının hoist ettiği title ve description meta'sı ile yazının template'inin stylesheet ile tanımladığı post.css link'i oraya gider. Layout'un content slot'u post fragment'ini, yazının author slot'u da author fragment'ini tutar. Yazının içinde başlık ve okurların yorumları escape edilir; template.HTML tipindeki gövde olduğu gibi eklenir.</title>
<defs><marker id="tpl-post-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="4" width="312" height="246" rx="8" />
<text x="14" y="21">layout</text><text class="d-small" x="60" y="21">layouts/master.html</text>
<rect class="d-slot d-slot-b" x="14" y="30" width="292" height="36" rx="5" />
<text class="d-small" x="24" y="51">hoist "head"</text>
<rect class="d-chip" x="104" y="39" width="60" height="18" rx="9" /><text class="d-small" x="134" y="51" text-anchor="middle">&lt;title&gt;</text>
<rect class="d-chip" x="170" y="39" width="56" height="18" rx="9" /><text class="d-small" x="198" y="51" text-anchor="middle">&lt;meta&gt;</text>
<rect class="d-chip-hit" x="232" y="39" width="74" height="18" rx="9" /><text class="d-hit-text d-small" x="269" y="51" text-anchor="middle">post.css</text>
<rect class="d-slot d-slot-a" x="14" y="74" width="292" height="170" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="88">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="96" width="272" height="140" rx="6" />
<line class="d-flow" x1="134" y1="96" x2="134" y2="61" marker-end="url(#tpl-post-arrow)" />
<line class="d-flow" x1="198" y1="96" x2="198" y2="61" marker-end="url(#tpl-post-arrow)" />
<line class="d-flow" x1="269" y1="101" x2="269" y2="61" marker-end="url(#tpl-post-arrow)" />
<text class="d-slot-text-a" x="34" y="113">post</text><text class="d-small" x="70" y="113">pages/post.html</text>
<rect class="d-chip-hit" x="206" y="101" width="84" height="18" rx="9" /><text class="d-hit-text d-small" x="248" y="113" text-anchor="middle">stylesheet</text>
<text class="d-small" x="34" y="134">h1: escape edilir</text>
<rect class="d-slot d-slot-c" x="34" y="142" width="252" height="50" rx="4" />
<text class="d-small" x="44" y="156">slot "author"</text>
<rect class="d-c" x="44" y="162" width="100" height="22" rx="4" /><text class="d-on-c" x="94" y="177" text-anchor="middle">author</text>
<text class="d-small" x="154" y="177">fragments/author.html</text>
<text class="d-small" x="34" y="210">yorumlar: escape edilir</text>
<text class="d-small" x="34" y="226">gövde: template.HTML, aynen</text>
</svg>
<figcaption>Post fragment'i layout'un <code>content</code> slot'unu, author fragment'i de yazının <code>author</code> slot'unu doldurur. Yazının handler'ının hoist ettiği title ve description ile template'inin tanımladığı stylesheet, layout'un <code>hoist "head"</code> işaretine çıkar. Bu işaret page'in tamamı render edildikten sonra doldurulur. Yazının içinde başlık ve okurların yorumları escape edilir; <code>template.HTML</code> tipindeki gövde olduğu gibi eklenir.</figcaption>
</figure>

### Okurun gördüğü

Başlığı `Escaping <script> & friends` olan bir yazı ve metni
`<script>alert(1)</script>` olan bir yorum düşünün. `/posts/escaping` için gelen
bir request 200 ile cevaplanır. Page, `<head>`'den `</body>`'ye kadar ve boş satırlar
çıkarılmış hâliyle şöyledir:

```html
<head>
  <meta charset="utf-8">
  <title>Escaping &lt;script&gt; &amp; friends</title><meta name="description" content="What html/template escapes for you"><link rel="stylesheet" href="/static/post.cba296a6afd593e0.css">
</head>
<body>
<article>
  <h1>Escaping &lt;script&gt; &amp; friends</h1>
  <p class="meta">18 October 2026</p>
<aside class="author">
  <strong>Ada</strong>
  <p>Writes about templates.</p>
</aside>
  <div class="body"><p>Hello <em>world</em></p></div>
  <section class="comments">
    <p><b>mallory</b> &lt;script&gt;alert(1)&lt;/script&gt;</p>
  </section>
</article>
</body>
```

- **Başlık, göründüğü iki yerde de escape edilir.** `rc.HoistTitle` onu `<title>`
  için, `{{.Title}}` ise `<h1>` için escape eder.
- **Yorum düz metindir.** `{{.Text}}` bir `string`'dir. Bu yüzden okurun script'i
  çalışmaz, metin olarak basılır.
- **Gövde markup'tır.** Tipi `template.HTML` olduğu için olduğu gibi eklenir.
  Güvenliği sizin kodunuza kalan tek değer budur.
- **Stylesheet `<head>`'dedir.** Yazının template'i onu `<body>` içinde tanımlasa
  da content-addressed URL'siyle head'e link'lenir.

### Register'ın yakaladıkları

`pages/post.html`'de başlık için `{{.Titel}}`, yorum döngüsünde de `{{.Name}}`
yerine `{{.Author}}` yazın. `RegisterPage`, page'i `ErrTemplateType` ile reddeder.
İki bulguyu birden, her birini dosyası, satırı ve sütunuyla verir:

```text
collage: page "post": fragment "post" (pages/post.html:4:8): {{.Titel}}: type main.postView has no field or method Titel (did you mean Title?)
collage: page "post": fragment "post" (pages/post.html:9:31): {{.Author}}: type main.comment has no field or method Author
```

`{{range .Comments}}` içinde `.` bir `comment`'tir. Bu yüzden ikinci bulgu yazının
değil, o tipin adını verir.

Bunun yerine yazar kartını bir yazım hatasıyla bağlayın:
`WithSlotFragment("autor", author)`. Bu durumda `RegisterPage`, `ErrUnknownSlot`
ile başarısız olur:

```text
collage: unknown slot: page "post" fragment "post" binds into slot "autor", but template "pages/post.html" never calls it; it calls [author]
```

Bu kontrollerin ikisi de escaping'e bakmaz. `{{.Text | safeHTML}}` tiplere uyar ve
iki kontrolden de geçer, ama okurun script'ini markup olarak basar. `safeHTML`'i
yalnızca kendi ürettiğiniz ya da sanitize ettiğiniz içerik için kullanın.
