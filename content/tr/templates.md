---
description: Template'lerin nerede durduğu, nasıl yüklenip binary'ye embed edildiği, development'ta nasıl yeniden yüklendiği; slot'lar, built-in ve kendi fonksiyonlarınız, escaping.
reference: TemplateConfig, Config, ErrUnknownSlot
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
├── layouts/default.html     →  "layouts/default.html"
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
- **Çalışırken hata veren bir template**, örneğin veride olmayan bir alana erişen ya
  da hata dönen bir fonksiyonu çağıran bir template, kendi fragment'ini başarısız
  kılar.
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
adı `templates/pages/post.html` gibi olurdu. `FS` ile birlikte boş bir `Root`,
filesystem'in kökü anlamına gelir. `all:` prefix'i, adı `.` ya da `_` ile başlayan
dosyaların da embed edilmesini sağlar. Düz bir `//go:embed templates` bu dosyaları
atlar.

`collage new` da projeyi bu şekilde oluşturur.

## Development'ta yeniden yükleme

Development'ta, yani `Config.DevMode` ya da tek başına `Config.Template.DevMode`
açıkken, her render'dan önce bütün template'ler yeniden parse edilir. Bir template'i
kaydettiğinizde bir sonraki request onu kullanır. Framework `COLLAGE_DEV`'i okumaz.
Bu değişkeni `collage dev` set eder, scaffold edilen `main.go` da onu
`Config.DevMode`'a çevirir.

Embed edilmiş bir set bunu yapamaz. Byte'ları binary build edildiği anda sabitlenmiştir
ve onları yeniden parse etmek hiçbir şeyi değiştirmez. Bu yüzden `Root`, çalışma
dizinine göre bir dizin olarak var olduğunda **development'ta diskteki dizin
embed edilmiş kopyanın önüne geçer** ve uygulama bu seçimi log'lar. `collage dev`
proje dizininizde çalıştığı için bu dizin orada her zaman vardır.

Bir değişikliğin hemen görünmesini iki şey daha sağlar:

- Development'ta page cache hiç okunmaz. Böylece değişiklikten önce cache'lenmiş bir
  page, değişikliği gizlemez.
- Development'taki her page, küçük bir script içerir. Bu script bir template ya da
  static bir dosya değiştiğinde tarayıcıyı yeniler. v0.10.0'dan itibaren
  [`Config.DevWatch`](/docs/configuration#devwatch)'ta belirttiğiniz bir dizindeki
  dosyalar da buna dahildir. Bu, handler'larınızın diskten okuduğu Markdown gibi
  içerikler içindir.

Bütün set yeniden parse edildiği için herhangi bir template'teki bir syntax hatası,
düzeltilene kadar bütün render'ları başarısız kılar. Bu development'ta olur. Hatayı orada
hemen, dosya ve satır numarasıyla birlikte görürsünüz. Register sırasında yapılan
kontroller, açılıştan sonra düzenlenen bir template için yeniden yapılmaz. Böyle
bir template'in hatası bunun yerine development error page'inde görünür; önce
template, satır ve neden gelir.

## Bir template ne alır

`.`, tam olarak fragment'in data handler'ının döndürdüğü değerdir ya da
fragment'e `WithData` ile verilen değerdir. İkisi de olmayan ya da handler'ı
`collage.Effect` ile uyarlanmış bir fragment, veri olmadan render edilir. Bkz.
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

Link fonksiyonları katıdır. Bilinmeyen bir page adı ya da eksik bir parametre, bozuk
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
  Render'a bağlı olan sekiz fonksiyon bunun dışındadır: `slot`, `hoist`, `asset`,
  `stylesheet`, `csrfToken`, `pageURL`, `pageURLIn` ve `localeURL`. Render engine
  her render için bunların kendi versiyonlarını bağlar. Bu yüzden bunlardan birini
  override etmek kabul edilir, ama hiçbir etkisi olmaz.
- **Bir fonksiyon request'i göremez.** Fonksiyon, bütün program için bir kez register
  edilir. Request'e, locale'e ya da kullanıcıya bağlı olan her şey data handler'da
  olmalıdır. Veri zaten oradan gelir.

Bir plugin de `Configure` hook'undan fonksiyon ekleyebilir. Ayrıntılar için
[Plugin yazmak](/docs/writing-plugins) sayfasına bakın. Bir plugin ve `Funcs` aynı
adı tanımlarsa `Funcs` kazanır, çünkü ikisini birden görebilen uygulamadır.

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
