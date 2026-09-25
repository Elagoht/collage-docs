---
description: Bir plugin'in neler yapabildiği, bir plugin'in nasıl register edilip yapılandırıldığı ve yayımlanmış üç plugin.
reference: Plugin, LoadPluginConfig, ErrUnknownPluginConfig, ErrAppStarted
---

# Plugin kullanmak

Plugin, sizin oluşturup uygulamanıza verdiğiniz sıradan bir Go değeridir.
Uygulamanın ne yaptığını izleyebilir ve ürettiği şeylerin bir kısmını
değiştirebilir. Ancak router'a, cache'e ya da template kümesine erişemez. Ona dar
bir yetenek kümesi verilir, fazlası verilmez.

Plugin yükleyen bir mekanizma da, plugin yayımlayacağınız bir registry de yoktur.
Plugin, import ettiğiniz bir Go modülüdür ve diğer tüm bağımlılıklar gibi
binary'nizin içine derlenir.

## Bir plugin neler yapabilir

Bir plugin'in yaptığı her şey, ya kendi seçtiği bir hook'tan ya da başlangıçta ona
verilen bir yetenekten geçer. Bu ikisiyle bir plugin şunları yapabilir:

- **Render edilen çıktıyı yeniden yazabilir.** Bir page'in HTML'ini, bir sitemap'in
  ya da bir JSON document'ın gövdesini, sunulmadan ve cache'e yazılmadan önce
  değiştirebilir. Bir minifier bu şekilde çalışır.
- Page render edilmeden önce head'e hoist ederek **page'e katkıda bulunabilir**:
  bir structured data bloğu, bir meta tag ya da bir preload ipucu ekleyebilir.
- **Template fonksiyonları ekleyebilir.** Eklenen fonksiyonları her template
  çağırabilir.
- **Mount edilen her dosya sistemini dönüştürebilir.** Böylece bir mount'un sunduğu
  dosyalar, örneğin, zaten minify edilmiş olur.
- Kendine ait **page'ler, document'lar ve mount'lar register edebilir**. Bir görsel
  optimize edici plugin, link verdiği yeniden boyutlandırılmış görselleri kendi
  mount'undan sunar.
- **Bir cache yazımını ayarlayabilir.** Yazılan kaydın ömrünü ya da tag'lerini
  değiştirebilir veya yazımı tamamen atlayabilir. Invalidation'lardan da haberdar
  olur.
- **Hataları gözlemleyebilir.** Her hatayı, pipeline'ın hangi aşamasında oluştuğu
  bilgisiyle birlikte alır.
- Programınızın çalıştırdığı **komutlar ekleyebilir.** `collage new` ile oluşturulan
  bir projede bunlar `go run . <command>` ile çalışır. Bkz.
  [collage CLI](/docs/cli#plugin-commands).

Bunların her birinin plugin tarafından nasıl göründüğünü
[Plugin yazmak](/docs/writing-plugins) sayfasında bulabilirsiniz.

## Bir plugin'i register etmek

Plugin ayrı bir modüldür. Modülü projeye ekleyin, plugin'i oluşturun ve
`Config.Plugins`'e koyun:

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

İkinci bir yol da vardır: `New`'dan sonra çağrılan `app.RegisterPlugin`.

```go
if err := app.RegisterPlugin(jsonld.New()); err != nil {
	log.Fatal(err)
}
```

**`Config.Plugins`'i tercih edin.** Bazı plugin'lerin, uygulama kurulurken devreye
girmesi gerekir. Örneğin bir template fonksiyonu eklemek ya da mount edilen dosya
sistemlerini sarmalamak için. Bu işler `New` içinde yapılır. Böyle bir plugin,
isteğe bağlı bir `Configure` aşaması implement eder. `RegisterPlugin` bu plugin'i kabul
edip asıl önemli kısmı sessizce atlamaz. Onun yerine plugin'i adıyla belirterek
`collage.ErrConfigurerRegisteredLate` ile reddeder. `Config.Plugins` ise her plugin
için çalışır, bu yüzden varsayılan olarak onu kullanın.

`RegisterPlugin` şu durumlarda da reddeder:

| Hata | Ne zaman |
| --- | --- |
| `ErrAppStarted` | Uygulama zaten başlamıştır: `Handler`, `ListenAndServe`, `Start`, `DispatchCommands` ya da bir render çalışmıştır. Başarısız olan başlatmalar da buna dahildir. Bir plugin'in `Init`'inde başarısız olan başlatma v0.11.0'dan, başka herhangi bir nedenle başarısız olan başlatma v0.12.0'dan beri sayılır. |
| `ErrNilPlugin` | Plugin `nil`'dir. |
| `ErrEmptyPluginName` | Plugin'in `Name()` değeri boştur. |
| `ErrDuplicatePlugin` | Aynı ada sahip başka bir plugin zaten vardır. |

`Config.Plugins` içindeki plugin'ler de aynı şekilde kontrol edilir. Bu durumda
hatayı `New` döndürür.

### Sıra önemlidir

Plugin'ler register edildikleri sırayla çalışır. Önce `Config.Plugins` içindekiler
slice'taki sırayla, ardından `RegisterPlugin` ile eklenenler çağrı sırasıyla
çalışır. Çıktıyı değiştiren hook'larda her plugin, kendinden önceki plugin'in
ürettiği çıktıyı görür. Page'e bir şey ekleyen plugin, genellikle çıktıyı sıkıştıran
plugin'den önce gelmelidir. Böylece eklenen içerik de sıkıştırılır.

## Plugin'leri yapılandırmak

Ayar alan bir plugin, ayarlarını `Config.PluginConfig`'ten okur. Bu alan, plugin'in
adını key olarak kullanan bir `map[string]json.RawMessage`'tır. Plugin adları modül
path'leri gibi yazılır, örneğin `elagoht/minimizer`. Böylece key ile plugin aynı
identifier'ı paylaşır.

En yaygın yöntem, programınızın yanında duran bir JSON dosyasıdır. `collage new` ile
oluşturulan bir projede boş bir `plugins-config.json` dosyası bulunur ve proje bu
dosyayı zaten yükler:

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

Dosya yoksa `LoadPluginConfig` hata vermez, `nil` bir map döndürür. Hiçbir şeyi
yapılandırmayan bir deploy'un, bunu belirtmek için boş bir dosyaya ihtiyacı
olmamalıdır. Dosya varsa ama okunamıyorsa ya da bir JSON object değilse, bu bir
hatadır.

Bu mekanizma JSON dosyalarına bağlı değildir. `LoadPluginConfig` yalnızca bir
kolaylıktır ve framework onu hiçbir zaman kendisi çağırmaz. `PluginConfig`'i
YAML'dan, environment variable'lardan ya da Go sabitlerinden doldurabilirsiniz:

```go
PluginConfig: map[string]json.RawMessage{
	"elagoht/minimizer": json.RawMessage(`{"js": true}`),
},
```

Hangi yolla doldurursanız doldurun, üç kural geçerlidir:

- **Bölüm yoksa varsayılanlar kullanılır.** Bölümü olmayan bir plugin, tam olarak
  constructor'ının kurduğu hâliyle çalışır.
- **Bölüm, varsayılanların üzerine decode edilir.** `{"js": true}` tek bir ayarı
  açar, diğerlerini olduğu gibi bırakır. Belirli bir plugin'in ayarları nasıl
  birleştirdiği, o plugin'in README'sinde yazar.
- **Register edilmiş hiçbir plugin'e karşılık gelmeyen bir key, uygulamanın
  başlamasını engeller** ve `collage.ErrUnknownPluginConfig` döner. Bu kontrol
  olmasaydı, `"elagoht/minimzer"` gibi bir yazım hatası plugin'i varsayılan
  ayarlarında bırakırdı ve siz plugin'in yapılandırıldığından emin olurdunuz.
  Kontrol `New` içinde değil, uygulama başlarken yapılır. Çünkü `RegisterPlugin`
  `New`'dan sonra da plugin ekleyebilir.

Var olan ama decode edilemeyen bir bölüm de hatadır. Örneğin plugin'in boolean
beklediği yerde bir string varsa, plugin bu bölümü okuduğunda hata oluşur.

## Yayımlanmış plugin'ler

Framework ile birlikte üç plugin yayımlanmıştır. Her biri ayrı bir modüldür ve
her birinin tam referans niteliğinde kendi README'si vardır. Aşağıdaki bilgiler
bir plugin'i kurmanız için yeterlidir.

### elagoht/minimizer

[github.com/Elagoht/collage-minimizer](https://github.com/Elagoht/collage-minimizer),
render edilen page'lerden, JSON endpoint'leri gibi document'lardan ve mount'larınızın
sunduğu dosyalardan whitespace'leri ve yorumları temizler.

```go
import minimizer "github.com/Elagoht/collage-minimizer"

Plugins: []collage.Plugin{minimizer.New()},
```

```json
{
  "elagoht/minimizer": { "html": true, "json": true, "css": true, "js": false }
}
```

- `Config.Plugins` içinde olmalıdır. Mount edilen dosya sistemlerini sarmalar ve bu
  işlem uygulama kurulurken yapılır.
- `New()` HTML, JSON ve CSS'i etkinleştirir. JavaScript varsayılan olarak
  kapalıdır, `{"js": true}` ile açabilirsiniz.
  `minimizer.NewWith(minimizer.Config{...})` ile her ayarı kendiniz belirlersiniz ve
  bu varsayılanlar devreye girmez.
- Bir parser değil, bir scanner'dır. Yalnızca anlam taşıyamayacak içeriği kaldırır.
  `<pre>`, `<textarea>`, `<script>` ve `<style>` olduğu gibi korunur, CSS
  string'lerine dokunulmaz, JavaScript'teki her satır sonu korunur ve geçersiz JSON
  olduğu gibi döndürülür.
- Mount edilen dosyalar, response değil dosya sistemi sarmalanarak minify edilir.
  Bu sayede bir mount'a yapılan `Range` request'leri doğru byte'ları döndürmeye
  devam eder.

### elagoht/jsonld

[github.com/Elagoht/collage-jsonld](https://github.com/Elagoht/collage-jsonld),
document'ın head'ine schema.org structured data yazar.

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

Plugin register edildiğinde, `siteName` ayarlanmışsa her page'e site genelinde
geçerli bir `WebSite` node'u ekler. Başka hiçbir şey eklemez, çünkü plugin bir
page'in ne hakkında olduğunu bilemez. Bunu page'in kendisi, makaleyi çeken data
handler'dan bildirir:

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

- `Configure` aşaması yoktur, bu yüzden `RegisterPlugin` de bu plugin'i kabul eder.
- `Emit` mevcut node'lara ekleme yapar ve plugin register edilmiş olsun ya da
  olmasın çalışır. Node'lar schema.org tipine göre key'lenir. Bu yüzden iç içe bir
  fragment'in `Article`'ı, daha dıştaki bir fragment'te tanımlanan `Article`'ın
  yerini alır. Farklı tipteki node'ların ise hepsi çıktıda yer alır.
- Tipli node'lar `Article`, `BlogPosting`, `Blog`, `Person`, `WebSite` ve
  `BreadcrumbList` tiplerini kapsar. Bunların dışındaki her şey için `jsonld.Raw`
  kullanılır. `jsonld.Raw` geçersiz JSON'u reddeder.
- Marshal edilemeyen bir node, page'i başarısız kılmaz, yalnızca atlanır.

**Layout'unuzda `{{hoist "head"}}` bulunması gerekir.** Bkz.
[aşağıdaki bölüm](#plugins-that-write-to-the-head).

### elagoht/opti-image

[github.com/Elagoht/collage-opti-image](https://github.com/Elagoht/collage-opti-image),
piksel cinsinden `width` ve `height` belirten her `<img>`'yi yeniden yazar. Yeni
`<img>`, plugin'in kendi mount'undan sunduğu yeniden boyutlandırılmış bir kopyayı
gösterir.

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

- `Config.Plugins` içinde olmalıdır.
- **`allowedOrigins` boşsa plugin devre dışı kalır.** Listelemediğiniz bir host'tan
  hiçbir zaman görsel çekmez. Scheme de origin'in bir parçasıdır.
- Yalnızca hem `width` hem de `height` değeri piksel sayısı olarak verilmiş
  görseller yeniden yazılır. Hedef boyut olarak güvenilebilecek tek değer, bu
  belirtilen boyuttur.
- Render sırasında hiçbir şey çekilmez. Page, `/_image/8f2a91c0b4e7d3a6.webp` gibi
  içeriğe göre adlandırılmış bir dosyaya link verir. Görsel, bir tarayıcı onu ilk
  kez istediğinde çekilir ve yeniden boyutlandırılır.
- Static export, görselleri de çıktısına yazar. Çünkü görseller bir mount'tan
  sunulur ve tüm mount'lar, page'ler render edildikten sonra kopyalanır.
- `webp` değeri `false` (varsayılan), `true` ya da `"auto"` olabilir. `"auto"`,
  WebP'yi yalnızca görselin aksi hâlde lossless olacağı durumlarda kullanır. Üretilen
  görseller varsayılan olarak bellekte ve `.cache/opti-image` içinde tutulur
  (`cacheDir`). `p.Purge()` ve `p.PurgeSource(url)` bunları temizler.

`optiimage.NewWith(optiimage.Config{...})` Go tarafında bir başlangıç config'i
belirler. JSON bölümü daha sonra bu config'in üzerine key key decode edilir.

## Head'e yazan plugin'ler

Document'ın head'ine structured data, meta tag ya da preload ipucu gibi içerik
ekleyen bir plugin, bunu hoist ederek yapar. Fragment'ler de title'larını ve
stylesheet'lerini aynı mekanizmayla ekler. Hoist edilen içerik, yalnızca layout'un
`{{hoist "head"}}` çağırdığı yere yerleşir:

```html
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
```

**Bu marker olmadan hiçbir şey görünmez.** Plugin register edilir, çalışır ve
bloğunu tanımlar, ama bloğun yerleşeceği bir yer yoktur. Bu bilinçli bir tercihtir.
HTML içinde `</head>`'i arayıp kendini araya ekleyen bir plugin, layout'unuzu
ilgilendiren bir kararı sizin yerinize vermiş olurdu. `collage new` ile oluşturulan
layout'larda marker zaten vardır. Elle yazdığınız bir layout'ta olmayabilir.
Hoist mekanizmasının genel anlatımı için [Head ve SEO](/docs/head-and-seo) sayfasına
bakın.

## Plugin'ler nerede çalışır

Plugin'ler, bir sunucunun render ettiği page'lerden fazlasını görür:

- **Cache'lenen page'ler yalnızca bir kez işlenir.** Plugin'in HTML'de yaptığı
  değişiklikler, page cache'e yazılmadan önce yapılır. Bu yüzden cache hit olduğunda
  işlenmiş byte'lar, plugin yeniden çalıştırılmadan sunulur. Her request'te
  çalışması gereken bir plugin, örneğin ziyaretçiye özel bir değer ekleyen bir
  plugin, cache'lenen bir page ile birlikte kullanılamaz.
- **Error page'leri** de aynı render hook'larından geçer. Böylece 404 page'iniz de
  diğer page'ler gibi minify edilir ve zenginleştirilir.
- **Bir action'ın response olarak döndüğü page** de `OnAfterRender`'dan geçer
  (v0.10.0'dan beri). Bu page, bir form'un validation sonrası yeniden render'ıdır.
  Böylece `GET` ile gelen page gibi minify edilir. Bkz.
  [Form'lar ve action'lar](/docs/forms-and-actions#the-validation-re-render).
- **Document'lar** (sitemap'ler, feed'ler, JSON) kendi hook'larından geçer. Böylece
  bir minifier onları da kapsar. Bkz. [Document'lar](/docs/documents).
- **Static export da sunucudaki render ile aynı render hook'larını çalıştırır.**
  Her plugin önce başlatılır ve yapılandırılır. Böylece export edilen site,
  sunucunun sunduğu siteyle aynı olur. Bkz. [Static export](/docs/static-export).

## Daha ileri

[Plugin yazmak](/docs/writing-plugins) sayfası plugin sözleşmesini, her hook'u ve
her hook'un neyi değiştirebileceğini anlatır. Ayrıca testleriyle birlikte eksiksiz
bir plugin örneği içerir.
