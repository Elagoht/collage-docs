---
description: Bir plugin'in neler yapabildiği, bir plugin'in nasıl kaydedilip yapılandırıldığı ve yayımlanmış üç plugin.
---

# Plugin kullanmak

Plugin, sizin oluşturup uygulamanıza verdiğiniz sıradan bir Go değeridir.
Uygulamanın yaptıklarını izleyebilir ve ürettiklerinin bir kısmını değiştirebilir;
ama router'a, önbelleğe ya da şablon kümesine uzanamaz: dar bir yetenek kümesi alır,
fazlasını değil.

Bir plugin yükleyicisi ya da yayımlanacak bir kayıt defteri yoktur. Plugin, import
ettiğiniz bir Go modülüdür ve diğer her bağımlılık gibi binary'nize derlenir.

## Bir plugin neler yapabilir

Bir plugin'in yaptığı her şey, katılmayı seçtiği bir hook'tan ya da başlangıçta
kendisine verilen bir yetenekten geçer. Bunlarla bir plugin şunları yapabilir:

- **render edileni yeniden yazmak** — bir sayfanın HTML'ini, bir sitemap'in ya da
  bir JSON document'ının gövdesini — sunulmadan ve önbelleğe alınmadan önce. Bir
  küçültücü (minifier) bu şekilde çalışır.
- render edilmeden önce head'e hoist ederek **sayfaya katkıda bulunmak**: bir
  yapılandırılmış veri bloğu, bir meta etiketi, bir preload ipucu.
- ardından her şablonun çağırabileceği **şablon fonksiyonları eklemek**.
- **mount edilmiş her dosya sistemini dönüştürmek**; böylece bir mount'un sunduğu
  dosyalar, örneğin, zaten küçültülmüş olur.
- kendine ait **sayfalar, document'lar ve mount'lar kaydetmek**. Bir görsel
  iyileştirici, bağlantısını verdiği yeniden boyutlandırılmış görselleri kendi
  mount'undan sunar.
- **bir önbellek yazımını ayarlamak** — ömrünü ya da etiketlerini değiştirmek veya
  onu atlamak — ve geçersiz kılmalardan haberdar olmak.
- **başarısızlıkları gözlemlemek**, gerçekleştikleri pipeline aşamasıyla birlikte.
- programınızın çalıştırdığı **komutlar eklemek** — iskeleti kurulmuş bir projede
  `go run . <command>`; bkz. [collage CLI](/docs/cli#plugin-commands).

Bunların her birinin plugin tarafından nasıl göründüğü
[Plugin yazmak](/docs/writing-plugins) sayfasındadır.

## Bir plugin'i kaydetmek

Plugin ayrı bir modüldür. Onu ekleyin, oluşturun ve `Config.Plugins`'e koyun:

```sh
go get github.com/Elagoht/collage-minimizer
```

```go
import minimizer "github.com/Elagoht/collage-minimizer"

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Plugins:  []collage.Plugin{minimizer.New()},
})
```

İkinci bir yol daha vardır: `New`'den sonra çağrılan `app.RegisterPlugin`:

```go
if err := app.RegisterPlugin(jsonld.New()); err != nil {
	log.Fatal(err)
}
```

**`Config.Plugins`'i tercih edin.** Bazı plugin'lerin uygulama kurulurken harekete
geçmesi gerekir — bir şablon fonksiyonu eklemek ya da mount edilmiş dosya
sistemlerini sarmalamak için — ve bu `New`'de olur. Böyle bir plugin isteğe bağlı
bir `Configure` aşaması uygular; `RegisterPlugin` de onu kabul edip asıl önemli
kısmı sessizce atlamak yerine adıyla birlikte `collage.ErrConfigurerRegisteredLate`
ile reddeder. `Config.Plugins` her plugin için çalışır; başvurulacak yol odur.

`RegisterPlugin` şu durumlarda da reddeder:

| Hata | Ne zaman |
| --- | --- |
| `ErrAppStarted` | Uygulama zaten başlamıştır — `Handler`, `ListenAndServe`, `Start`, `DispatchCommands` ya da bir render çalışmıştır. Başarısız olan her başlatma da buna dahildir — bir plugin'in `Init`'inde başarısız olan v0.11.0'dan itibaren, başka herhangi bir nedenle başarısız olan v0.12.0'dan itibaren. |
| `ErrNilPlugin` | Plugin `nil`'dir. |
| `ErrEmptyPluginName` | `Name()`'i boştur. |
| `ErrDuplicatePlugin` | Başka bir plugin'in adı zaten aynıdır. |

`Config.Plugins` içindeki bir plugin de aynı şekilde denetlenir ve hatayı `New`
döndürür.

### Sıra önemlidir

Plugin'ler kaydedildikleri sırayla çalışır: önce slice sırasıyla `Config.Plugins`,
ardından çağrı sırasıyla `RegisterPlugin` çağrıları. Çıktıyı değiştiren hook'larda
her plugin, kendinden öncekinin ürettiğini görür. Sayfaya ekleme yapan bir plugin
genellikle sayfayı sıkıştıran birinden önce gelmelidir; böylece eklenen de
sıkıştırılır.

## Plugin'leri yapılandırmak

Ayar alan bir plugin, bunları plugin'in adıyla anahtarlanmış bir
`map[string]json.RawMessage` olan `Config.PluginConfig`'ten okur. Adlar modül
yolları gibi okunur — `elagoht/minimizer` — böylece anahtar ve plugin tek bir
tanımlayıcıdır.

Yaygın durum, programınızın yanında duran bir JSON dosyasıdır. `collage new` ile
iskeleti kurulan bir projede boş bir `plugins-config.json` vardır ve proje onu zaten
yükler:

```json
{
  "elagoht/minimizer": { "js": true },
  "elagoht/jsonld": { "siteName": "The Wire", "siteURL": "https://thewire.example" }
}
```

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	log.Fatal(err)
}

app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{minimizer.New(), jsonld.New()},
	PluginConfig: pluginConfig,
})
```

Dosya yoksa `LoadPluginConfig` bir `nil` map döndürür ve hata döndürmez: hiçbir şey
yapılandırmayan bir yayına alma, bunu söylemek için boş bir dosyaya ihtiyaç
duymamalıdır. Var olan ama okunamayan ya da bir JSON nesnesi olmayan bir dosya ise
hatadır.

Bunların hiçbiri JSON dosyalarına bağlı değildir. `LoadPluginConfig`, framework'ün
kendisinin hiç çağırmadığı bir kolaylıktır; `PluginConfig`'i YAML'dan, ortamdan ya
da Go sabitlerinden doldurabilirsiniz:

```go
PluginConfig: map[string]json.RawMessage{
	"elagoht/minimizer": json.RawMessage(`{"js": true}`),
},
```

Onu hangi yolla doldurursanız doldurun üç kural geçerlidir:

- **Bölüm yoksa varsayılanlar geçerlidir.** Girdisi olmayan bir plugin, tam olarak
  constructor'ının kurduğu gibi çalışır.
- **Bir bölüm, varsayılanların üzerine decode edilir.** `{"js": true}` bir ayarı
  açar, geri kalanını olduğu gibi bırakır. Belirli bir plugin'in nasıl birleştirdiği
  kendi README'sinde yazar.
- **Kayıtlı hiçbir plugin'i adlandırmayan bir anahtar, uygulamanın başlamasını
  `collage.ErrUnknownPluginConfig` ile durdurur.** Aksi hâlde `"elagoht/minimzer"`
  gibi bir yazım hatası, plugin'i varsayılanlarında bırakır, sizi de yapılandırıldığından
  emin. Denetim `New`'de değil, uygulama başladığında çalışır, çünkü
  `RegisterPlugin` `New`'den sonra da plugin ekleyebilir.

Var olan ama decode edilemeyen bir bölüm — plugin'in boolean beklediği yerde bir
string — da bir hatadır; plugin onu okuduğunda ortaya çıkar.

## Yayımlanmış plugin'ler

Framework'le birlikte üç plugin yayımlanmıştır. Her biri kendi modülüdür ve tam
başvuru kaynağı olan kendi README'si vardır; aşağıdakiler birini kurmaya yeter.

### elagoht/minimizer

[github.com/Elagoht/collage-minimizer](https://github.com/Elagoht/collage-minimizer),
render edilmiş sayfalardan, JSON uç noktaları gibi document'lardan ve mount'larınızın
sunduğu dosyalardan boşlukları ve yorumları ayıklar.

```go
import minimizer "github.com/Elagoht/collage-minimizer"

Plugins: []collage.Plugin{minimizer.New()},
```

```json
{
  "elagoht/minimizer": { "html": true, "json": true, "css": true, "js": false }
}
```

- `Config.Plugins`'e konmalıdır: mount edilmiş dosya sistemlerini sarmalar ve bu,
  uygulama kurulurken olur.
- `New()` HTML, JSON ve CSS'i etkinleştirir. JavaScript varsayılan olarak kapalıdır;
  `{"js": true}` ile açın. `minimizer.NewWith(minimizer.Config{...})` her anahtarı
  kendiniz ayarlamanızı sağlar ve bu varsayılanları devre dışı bırakır.
- Bir ayrıştırıcı (parser) değil, bir tarayıcıdır (scanner) ve yalnızca anlam
  taşıyamayacak olanı kaldırır: `<pre>`, `<textarea>`, `<script>` ve `<style>`
  olduğu gibi korunur, CSS string'lerine dokunulmaz, JavaScript her satır sonunu
  korur ve geçersiz JSON olduğu gibi döndürülür.
- Mount edilmiş dosyalar yanıt değil, dosya sistemi sarmalanarak küçültülür; böylece
  bir mount'a yapılan `Range` istekleri doğru baytları döndürmeye devam eder.

### elagoht/jsonld

[github.com/Elagoht/collage-jsonld](https://github.com/Elagoht/collage-jsonld),
document'ın head'ine schema.org yapılandırılmış verisi yazar.

```go
import "github.com/Elagoht/collage-jsonld"

Plugins: []collage.Plugin{jsonld.New()},
```

```json
{
  "elagoht/jsonld": {
    "siteName": "The Wire",
    "siteURL": "https://thewire.example",
    "searchURL": "https://thewire.example/search?q={query}"
  }
}
```

Kaydedildiğinde her sayfaya site geneli bir `WebSite` düğümü yazar — `siteName`
ayarlanmışsa — ve başka hiçbir şey yazmaz, çünkü plugin bir sayfanın ne hakkında
olduğunu bilemez. Bunu sayfa, makaleyi getiren data handler'dan söyler:

```go
type articleView struct {
	Article Article
}

func loadArticle(ctx context.Context, rc *collage.RenderContext) (articleView, []string, error) {
	article, err := client.Article(ctx, rc.Param("slug"))
	if err != nil {
		return articleView{}, nil, err
	}
	jsonld.Emit(rc, jsonld.Article{
		Headline:      article.Title,
		DatePublished: article.PublishedAt,
		AuthorName:    article.Author,
	})
	return articleView{Article: article}, []string{"article:" + article.Slug}, nil
}
```

- `Configure` aşaması yoktur; bu yüzden `RegisterPlugin` de onu kabul eder.
- `Emit` ekleme yapar ve plugin kayıtlı olsun ya da olmasın çalışır. Düğümler
  schema.org tipine göre anahtarlanır; böylece iç içe bir fragment'in `Article`'ı,
  daha dışarıda bildirilmiş olanın yerini alır ve farklı tiplerdeki düğümlerin hepsi
  görünür.
- Tipli düğümler `Article`, `BlogPosting`, `Blog`, `Person`, `WebSite` ve
  `BreadcrumbList`'i kapsar; geri kalan her şey için kaçış kapısı `jsonld.Raw`'dır
  ve geçersiz JSON'u reddeder.
- Marshal edilemeyen bir düğüm, sayfayı başarısız kılmak yerine atlanır.

**Layout'unuzda `{{hoist "head"}}` olması gerekir** — bkz.
[aşağısı](#plugins-that-write-to-the-head).

### elagoht/opti-image

[github.com/Elagoht/collage-opti-image](https://github.com/Elagoht/collage-opti-image),
piksel cinsinden bir `width` ve `height` bildiren her `<img>`'yi, kendi mount'undan
kendisinin sunduğu yeniden boyutlandırılmış bir kopyaya yönlendirecek şekilde yeniden
yazar.

```go
import optiimage "github.com/Elagoht/collage-opti-image"

Plugins: []collage.Plugin{optiimage.New()},
```

```json
{
  "elagoht/opti-image": {
    "allowedOrigins": [{ "scheme": "https", "host": "images.example.com" }],
    "webp": "auto"
  }
}
```

- `Config.Plugins`'e konmalıdır.
- **Boş bir `allowedOrigins` onu devre dışı bırakır.** Listelemediğiniz bir
  host'tan asla veri çekmez ve şema (scheme) origin'in bir parçasıdır.
- Yalnızca hem `width` hem de `height` değeri piksel sayısı olan görseller yeniden
  yazılır; bildirilmiş bu boyut, var olan tek dürüst hedef boyuttur.
- Render sırasında hiçbir şey çekilmez. Sayfa `/_image/8f2a91c0b4e7d3a6.webp` gibi
  içerikle adreslenen bir ada bağlantı verir; görsel, bir tarayıcı onu ilk kez
  istediğinde çekilir ve yeniden boyutlandırılır.
- Statik dışa aktarma görselleri çıktısına yazar, çünkü bunlar bir mount'tan sunulur
  ve her mount, sayfalar render edildikten sonra kopyalanır.
- `webp`; `false` (varsayılan), `true` ya da yalnızca görselin aksi hâlde kayıpsız
  olacağı yerlerde WebP kullanan `"auto"` olabilir. Üretilen görseller varsayılan
  olarak bellekte ve `.cache/opti-image` içinde tutulur (`cacheDir`); `p.Purge()` ve
  `p.PurgeSource(url)` bunları temizler.

`optiimage.NewWith(optiimage.Config{...})`, Go'da bir başlangıç yapılandırması
ayarlar; JSON bölümü ardından bunun üzerine anahtar anahtar decode edilir.

## Head'e yazan plugin'ler

Document head'ine katkıda bulunan bir plugin — yapılandırılmış veri, meta
etiketleri, preload ipuçları — bunu, fragment'lerin başlıkları ve stil dosyaları için
kullandığı mekanizmanın aynısıyla, hoist ederek yapar. Hoist edilen içerik, layout'un
`{{hoist "head"}}`'i çağırdığı yere düşer, başka hiçbir yere değil:

```html
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
```

**Bu işaretçi olmadan hiçbir şey görünmez.** Plugin kaydolur, çalışır, bloğunu
bildirir ve bloğun gidecek bir yeri olmaz. Bu bilinçlidir: HTML'de `</head>`'i arayıp
kendini araya ekleyen bir plugin, bir layout sorusunu layout'unuz adına karara
bağlamış olurdu. `collage new`'den gelen bir layout'ta işaretçi zaten vardır; elle
yazdığınız bir layout'ta olmayabilir. Genel olarak hoist etmek için bkz.
[Head ve SEO](/docs/head-and-seo).

## Plugin'ler nerede çalışır

Plugin'ler, bir sunucunun render ettiği sayfalardan fazlasını görür:

- **Önbellekteki sayfalar bir kez işlenir.** Bir plugin'in HTML'de yaptığı
  değişiklikler sayfa önbelleğe alınmadan önce yapılır; böylece bir önbellek isabeti
  (hit), plugin'i yeniden çalıştırmadan işlenmiş baytları sunar. Her istekte
  çalışması gereken bir plugin — ziyaretçiye özgü bir değer eklemek için — önbelleğe
  alınan bir sayfayla birleştirilemez.
- **Hata sayfaları** aynı render hook'larından geçer; böylece 404'ünüz de diğer her
  sayfa gibi küçültülür ve zenginleştirilir.
- **Bir action'ın yanıt olarak verdiği sayfa** — bir formun doğrulama sonrası yeniden
  render'ı — da `OnAfterRender`'ı çalıştırır (v0.10.0'dan itibaren); böylece bir
  `GET`'in aldığı sayfa gibi küçültülür. Bkz.
  [Formlar ve action'lar](/docs/forms-and-actions#the-validation-re-render).
- **Document'lar** — sitemap'ler, feed'ler, JSON — kendi hook'larından geçer; böylece
  bir küçültücü onları da kapsar. Bkz. [Document'lar](/docs/documents).
- **Statik dışa aktarma, sunulan bir render'la aynı render hook'larını çalıştırır**
  ve her plugin önce başlatılıp yapılandırılır; böylece dışa aktarılan site, sunucunun
  sunduğu sitedir. Bkz. [Statik dışa aktarma](/docs/static-export).

## Daha ileri

[Plugin yazmak](/docs/writing-plugins); plugin sözleşmesini, her hook'u ve neyi
değiştirebileceğini ve testleriyle birlikte eksiksiz bir plugin'i anlatır.
