---
description: Şablonlar nerede durur, nasıl yüklenir ve gömülür, geliştirmede nasıl yeniden yüklenir; slot'lar, yerleşik ve özel fonksiyonlar ve kaçışlama.
---

# Şablonlar

Her fragment bir şablon render eder. Şablonlar Go'nun
[`html/template`](https://pkg.go.dev/html/template) paketidir — aynı sözdizimi,
aynı bağlama duyarlı kaçışlama — ve bir fragment'in ihtiyaç duyduğu şeyler için
eklenmiş birkaç fonksiyon: slot'ları, diğer sayfalara bağlantılar, statik dosyalar,
formlar ve sayfanın `<head>`'i.

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  <p class="meta">{{formatTime .Published "2 January 2006"}}</p>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
```

## Şablonlar nerede durur

Onları nerede bulacağını `Config.Template` söyler:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		Root:      "templates",
		Extension: ".html",
	},
})
```

`Root` altında, uzantısı `Extension` olan her dosya bir şablondur. `Root`'un
varsayılanı `./templates`, `Extension`'ınki `.html`'dir; başka uzantılı dosyalar yok
sayılır, dolayısıyla şablonların yanındaki bir `README.md` zararsızdır.

Bir şablonun adı, uzantı dahil, `Root`'a göre yoludur ve bir fragment'in belirttiği
de budur:

```text
templates/
├── layouts/default.html     →  "layouts/default.html"
├── pages/post.html          →  "pages/post.html"
└── fragments/author.html    →  "fragments/author.html"
```

```go
collage.NewFragment("author", "fragments/author.html")
```

### Bir kez yüklenir, erken denetlenir

`collage.New`, herhangi bir şey sunulmadan önce her şablonu tek bir kümeye
ayrıştırır. Her hata türü, yakalanabileceği en erken noktada yakalanır:

- **Var olmayan bir kök dizin**, `New`'u `ErrTemplateRootMissing` ile başarısız
  kılar — bu, en sık karşılaşılan başlangıç hatasıdır ve ayrıştırılamayan bir şablondan
  ayırt edilmeye değer.
- **Ayrıştırılamayan** ya da kimsenin kaydetmediği bir fonksiyonu çağıran **bir
  şablon**, `New`'u dosyayı adlandırarak başarısız kılar.
- **Yüklenmemiş bir şablonu belirten bir fragment**, `RegisterPage`'i sayfayı,
  fragment'i ve yolu adlandıran bir `ErrTemplateNotFound` ile başarısız kılar.
- **Çalışırken hata veren bir şablon** — verinin sahip olmadığı bir alan, hata
  döndüren bir fonksiyon — kendi fragment'ini başarısız kılar ve fragment'in
  [hata politikası](/docs/fragments-and-slots#when-a-fragment-fails) uygulanır.
  Çıktı tamponlanır; bu yüzden yarı yolda hata veren bir şablon yarım bir fragment
  değil, hiçbir şey yazmaz.

Her şablon tek bir kümede olduğundan bir şablon, bir başkasını adıyla içerebilir:

```html
{{template "partials/byline.html" .}}
```

`{{define}}` ile verilen adlar, her dosya genelinde bu tek ad alanını paylaşır; bu
yüzden onları ayırt edici seçin.

Yalnızca `Root` ile şablonlar diskten `os.OpenRoot` üzerinden okunur; böylece
dizinin dışına çıkan bir sembolik bağlantı izlenmez, reddedilir: `New`, dosyayı
adlandıran bir `ErrTemplateEscapesRoot` ile başarısız olur.

### Şablonları binary'ye gömmek

Şablonlarını `./templates`'ten okuyan bir binary yalnızca onların bulunduğu bir
dizinden çalışır. Bunun yerine onları gömün; binary her yerden çalışır — bir
container'dan, bir systemd biriminden, bir sunucudaki bir kopyadan:

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

`FS` ayarlıyken `Root`, o dosya sistemi içindeki bir dizindir. Yine de her addan
çıkarılır ve gerekli olmasının nedeni budur: `embed.FS` dosyaları kaynak
ağacındaki yollarıyla adlandırır ve `Root` olmadan her şablonun adı
`templates/pages/post.html` olurdu. `FS` ile birlikte boş bir `Root`, dosya
sisteminin kökü anlamına gelir. `all:` öneki, gömmenin adları `.` ya da `_` ile
başlayan dosyaları da içermesini sağlar; düz bir `//go:embed templates` bunları
atlardı.

`collage new`'un iskeletini oluşturduğu da budur.

## Geliştirmede yeniden yükleme

Geliştirmede — `Config.DevMode` ya da tek başına `Config.Template.DevMode` — her
şablon, her render'dan önce yeniden ayrıştırılır. Bir şablonu kaydedin; bir sonraki
istek onu kullanır. Framework `COLLAGE_DEV`'i okumaz: onu `collage dev` ayarlar,
iskeletteki `main.go` da onu `Config.DevMode`'a dönüştürür.

Gömülü bir küme bunu yapamazdı: baytları binary derlendiğinde sabitlenmiştir ve
onları yeniden ayrıştırmak hiçbir şeyi değiştirmez. Bu yüzden `Root` çalışma
dizinine göre bir dizin olarak var olduğu her durumda **geliştirmede diskteki dizin
gömülü kopyaya göre önceliklidir** ve uygulama bu seçimi log'a yazar. Proje
dizininizde çalışan `collage dev` altında bu her zaman böyledir.

Bir düzenlemenin hemen görünmesini sağlayan iki şey daha var:

- Geliştirmede sayfa önbelleği hiç okunmaz; böylece düzenlemeden önce önbelleğe
  alınmış bir sayfa onu gizlemez.
- Her geliştirme sayfası, bir şablon ya da statik dosya değiştiğinde — ya da
  handler'larınızın diskten okuduğu Markdown gibi içerikler için
  [`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtilen bir dizindeki
  bir dosya değiştiğinde (v0.10.0'dan itibaren) — tarayıcıyı yenileyen küçük bir
  script taşır.

Kümenin tamamı yeniden ayrıştırıldığından herhangi bir şablondaki bir sözdizimi
hatası, düzeltilene kadar her render'ı başarısız kılar — geliştirmede, yani hatada
dosya ve satırla birlikte onu hemen göreceğiniz yerde.

## Bir şablon ne alır

`.`, tam olarak fragment'in data handler'ının döndürdüğü şeydir —
`collage.DataHandler` ile kendi tipinizden bir değer. Data handler'ı olmayan ya da
`collage.Effect` ile uyarlanmış bir fragment veri olmadan render edilir.

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

Bir şablon yalnızca kendi fragment'inin verisini görür. Bir ebeveynin verisi
çocukta erişilebilir değildir ve global bir site nesnesi yoktur: sitenin adı gibi
her şablonun ihtiyaç duyduğu bir şey, bir handler'ın döndürdüğü veri ya da sizin
kaydettiğiniz bir fonksiyondur.

## Slot'lar

`{{slot "name"}}`, fragment'in bu adlı slot'unun tuttuklarını sırasıyla render eder
ve HTML olarak ekler. Çocukların çıktısı ikinci kez kaçışlanmaz; her çocuk, render
edilirken kendi değerlerini kaçışlamıştır.

```html
<main>{{slot "content"}}</main>
<aside>{{slot "sidebar"}}</aside>
```

- Hiçbir şey tutmayan bir slot hiçbir şey render etmez.
- Fragment'in `WithSlot` ile hiç bildirmediği bir ad, fragment'i başarısız kılan bir
  hatadır — aksi hâlde bir yazım hatası, sayfadan sessizce eksilen bir bölüm olurdu.
- Şablonun atladığı bir `{{if}}` içindeki slot render edilmez, ancak fragment'lerinin
  data handler'ları zaten başlatılmıştır; bkz.
  [Fragment'ler ve slot'lar](/docs/fragments-and-slots#a-slot-the-template-skips).

`slot` fonksiyonu şablonu çalıştıran fragment'e aittir; bu yüzden aynı
`{{slot "content"}}`, onu yazan her fragment'te farklı bir slot anlamına gelir.

## Yerleşik fonksiyonlar

Bunlar her şablonda, `html/template`'in kendi fonksiyonlarının (`printf`, `len`,
`index`, `eq` ve diğerleri) yanında kullanılabilir:

| Fonksiyon | Ne yapar |
| --- | --- |
| `slot "name"` | Bu fragment'in slot'larından birindeki fragment'leri render eder |
| `pageURL "name" "param" value …` | Kayıtlı bir sayfanın ya da document'ın, bu render'ın locale'indeki URL'si |
| `pageURLIn "locale" "name" …` | Aynısı, tam olarak verilen locale'de |
| `localeURL "locale"` | Bu sayfanın başka bir locale'deki hâli; orada yolu yoksa boş |
| `asset "/static/app.css"` | Mount edilmiş bir dosyanın içerik adresli URL'si |
| `stylesheet "/static/app.css"` | Sayfanın `<head>`'i için bir stil dosyası bildirir |
| `hoist "head"` | Sayfanın bir alanı için yapılan bildirimlerin yerleştiği yer |
| `csrfToken` | Bir formun sahtecilik token'ını taşıyan gizli input |
| `safeHTML`, `safeURL` | Bir string'i güvenilir HTML ya da URL olarak işaretler — kaçış kapıları |
| `dict "key" value …` | `{{template}}`'e birkaç değer geçirmek için bir map kurar |
| `default fallback value` | `value`; boşsa `fallback`. Yalnızca string'ler: başka bir tip render'ı başarısız kılar |
| `upper`, `lower`, `title` | Büyük/küçük harf dönüşümü |
| `join sep items` | Bir `[]string`'i birleştirir |
| `formatTime t layout` | Bir `time.Time`'ı bir Go layout'uyla biçimlendirir |

Bağlantı fonksiyonları katıdır: bilinmeyen bir sayfa adı ya da eksik bir parametre,
bozuk bir bağlantı üretmek yerine render'ı başarısız kılar. Her fonksiyonun
argümanları ve davranışı [Şablon fonksiyonları](/docs/template-functions)
sayfasındadır.

## Kendi fonksiyonlarınızı eklemek

`Config.Template.Funcs`, her şablona fonksiyonlar — bir `html/template` `FuncMap`'i
— ekler:

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

- **`New`'dan önce ayarlayın.** `html/template` yalnızca adı şablon ayrıştırılırken
  bilinen bir fonksiyonu çağırabilir ve ayrıştırma `New`'da olur. Kimsenin
  kaydetmediği bir adı çağıran şablon `New`'da başarısız olur; amaç da budur:
  başlangıçta bulduğunuz bir yazım hatasıdır.
- **Yerleşik bir fonksiyonun adıyla eklenen bir girdi, yerleşik fonksiyonun yerini
  alır** — render'a bağlı sekiz tanesi hariç: `slot`, `hoist`, `asset`,
  `stylesheet`, `csrfToken`, `pageURL`, `pageURLIn` ve `localeURL`. Render motoru
  her render için bunların her birinin kendi sürümünü bağlar; dolayısıyla bunlardan
  herhangi birini ezmek kabul edilir ama hiçbir etkisi olmaz.
- **Bir fonksiyon isteği göremez.** Programın tamamı için bir kez kaydedilir.
  İsteğe, locale'e ya da kullanıcıya bağlı her şey data handler'a aittir; veri
  zaten oradan gelir.

Bir plugin de `Configure` hook'undan fonksiyon ekleyebilir — bkz.
[Plugin yazmak](/docs/writing-plugins). Bir plugin ve `Funcs` aynı adı
tanımladığında `Funcs` kazanır: ikisini de görebilen uygulamadır.

## Kaçışlama

`html/template` her değeri göründüğü yere göre kaçışlar: HTML metni, bir öznitelik,
bir URL, satır içi JavaScript ya da CSS. `<script>alert(1)</script>` başlıklı bir
yazı metin olarak yazdırılır ve bir `href` içindeki `javascript:` URL'si zararsız
bir URL ile değiştirilir. Hiçbir şeyi elle kaçışlamazsınız.

Bazen bir değer gerçekten HTML'dir — CMS'inizin temizlediği bir yazı gövdesi, kendi
kodunuzun ürettiği bir işaretleme. Bunu söylemenin iki yolu var:

- Yukarıdaki `Body` gibi, view'ınızda alana `template.HTML` tipini verin. Karar
  tiptir; Go'da, gözden geçirilebileceği yerde verilir.
- Şablonda `safeHTML`'i (URL için `safeURL`'i) kullanın.

İkisi de o değer için kaçışlamayı kapatır. **Bunları yalnızca sizin ürettiğiniz ya
da temizlediğiniz içerikte kullanın**, bir kullanıcının yazdığı hiçbir şeyde asla;
kaçışlanmamış bir kullanıcı değeri bir siteler arası betik çalıştırma (XSS)
açığıdır.

İki yerleşik fonksiyon, olduğu gibi eklenen HTML üretir:

- Çocukları kendi değerlerini kaçışlamış olan `{{slot}}`.
- Fragment'lerin sayfa için bildirdiklerini yazan `{{hoist}}`. `rc.HoistTitle`,
  `rc.HoistMeta` ve benzeri yardımcılar kendilerine verileni kaçışlar; alttaki
  `rc.Hoist` ise işaretlemeyi tam olarak yazıldığı gibi ekler, dolayısıyla orada
  kaçışlama sizin sorumluluğunuzdadır. Bkz. [Head ve SEO](/docs/head-and-seo).

`pageURL` ve kardeşleri bir yola yerleştirdikleri parametre değerlerini kaçışlar ve
bir tarayıcının yolda bir üst adım olarak çözümleyeceği `.` ya da `..` değerini
reddeder.
