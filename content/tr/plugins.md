---
description: Bir plugin'in neler yapabildiği, bir plugin'in nasıl register edilip yapılandırıldığı ve yayımlanmış beş plugin.
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
- Kendine ait **bir handler sunabilir.** Bu bir event stream ya da bir WebSocket
  olabilir. Plugin, bir page'in
  [fragment path'lerini](/docs/forms-and-actions#a-fragment-at-its-own-url) render
  edip bu bağlantı üzerinden gönderebilir (v0.18.0'dan beri).
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

Framework ile birlikte beş plugin yayımlanmıştır. Her biri ayrı bir modüldür ve
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
func loadArticle(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	article, err := client.Article(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	jsonld.Emit(rc, jsonld.Article{
		Headline:      article.Title,
		DatePublished: article.PublishedAt,
		AuthorName:    article.Author,
	})
	return article, []string{"article:" + article.Slug}, nil
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

### elagoht/live

[github.com/Elagoht/collage-live](https://github.com/Elagoht/collage-live), bir
page'in parçalarını tarayıcıda güncel tutar. Bir page'in
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) ile açtığı
fragment'leri yenileyen küçük bir client script'i sunar. Yenileme belli aralıklarla
ya da sunucu bir event stream üzerinden bir değişiklik gönderdiğinde yapılır.

```go
import live "github.com/Elagoht/collage-live"

Plugins: []collage.Plugin{live.New()},
```

```json
{
  "elagoht/live": { "prefix": "/_live/", "noStream": false, "keepAlive": "25s", "maxFragments": 32, "maxStreamAge": "0s" }
}
```

```html
<head>
  {{liveClient}}
</head>

<section data-collage-fragment="{{fragmentURL "home" "cpu"}}" data-collage-interval="2s">
  {{slot "cpu"}}
</section>

<section data-collage-fragment="{{fragmentURL "home" "disks"}}" data-collage-push>
  {{slot "disks"}}
</section>
```

- `Config.Plugins` içinde olmalıdır, çünkü layout'un client'ı eklemek için çağırdığı
  `{{liveClient}}`'ı ekler. v0.2.1, collage v0.20.0 ya da sonrasını gerektirir;
  v0.2.0 v0.19.0'ı, v0.1.0 ise v0.18.0'ı gerektiriyordu.
- Container'ın sahibi page'dir, içindekinin sahibi fragment'tir.
  `data-collage-interval` belli aralıklarla fetch eder, `data-collage-push`
  fragment'i stream'den alır, `data-collage-swap="morph"` DOM'u yerinde patch eder.
  Bir form üzerindeki `data-collage-target` ise form'u `fetch` ile gönderir ve
  cevabı bir element'in içine koyar. Bunu başarı durumunda ya da bir `422`
  geldiğinde yapar. Bir action, bir gönderimin validation'dan geçmediğini böyle
  söyler: form'un fragment'ini hatalarla birlikte ve 422 status'uyla yeniden
  döndürür (bkz.
  [Form'lar ve action'lar](/docs/forms-and-actions#refreshing-it-from-the-browser)).
  Başka her hata hedefi olduğu gibi bırakır ve onu stale olarak işaretler.
  v0.2.1'den beri action'ı redirect eden bir form, collage'ın
  [`Collage-Fetch`](/docs/forms-and-actions#failure-renders-success-redirects)
  header'ı sayesinde tek request'te yönlendirilir.
- Client elindeki `ETag`'i gönderir ve `304` gelirse DOM'a dokunmaz. Fragment'in
  hoist ettiklerini key'lerine göre head'e bir kez ekler. Gizli bir sekmede durur.
  Bir request başarısız olduğunda element'i `data-collage-stale` ile işaretler ve
  beklemeyi artırır. Gönderilen bir kopya, polling'in alacağı ETag'in aynısını
  taşır. Bu yüzden client'ın zaten elinde olan bir kopya yeniden gönderilmez.
- **Tarayıcı başına tek bağlantı.** Bir tarayıcı HTTP/1.1 üzerinden bir origin'e,
  tüm sekmeleri toplamında en fazla altı bağlantı açar. Bu yüzden v0.2.0'dan beri
  client stream'i, sitenin tüm sekmelerinin paylaştığı bir shared worker'dan açar.
  Shared worker olmayan yerlerde her sekme kendi stream'ini açar ve gizliyken
  kapatır. v0.2.1'den beri back-forward cache'ten dönen bir sekme yeniden push
  alır. Önceden worker, sekme ayrıldığında onu unutuyordu.
- Stream kapalıyken gönderilen element'ler stale olarak işaretlenir. Üç başarısız
  bağlantıdan sonra beş saniyede bir polling ile yenilenir ve stream dakikada bir
  yeniden denenir.
- **Gönderme tag'lere dayanır.** Uygulama bir tag'i invalidate ettiğinde plugin, o
  tag'e bağlı her açık fragment'i yeniden render eder ve stream'den gönderir.
  API'nin tamamı invalidate etmektir.
- **Örneklenen veri** (CPU yükü, bir kuyruğun uzunluğu gibi), fragment başına bir
  tag ile bir timer'la invalidate edilerek gönderilir. Böyle bir fragment'i
  `Static()` ile değil, [`Shared()`](/docs/caching#a-page-that-declares-none) ile
  işaretleyin. Böylece page dynamic kalırken fragment sekme başına değil, her
  değişiklikte bir kez render edilir. Aynı ölçümü okuyan birkaç fragment onu tag'siz
  bir `collage.Cached` ile çeker. Böylece bir fragment'in tick'i diğerlerini yeniden
  render ettirmez. Bir sistem monitörü örneği
  [README](https://github.com/Elagoht/collage-live#sampled-data-a-system-monitor)'de
  adım adım anlatılır.
- Framework'ün `Shared` olarak bildirdiği bir render URL başına bir kez yapılır ve o
  URL'nin her okuyucusuna gönderilir. Diğer render'lar her bağlantı için o
  bağlantının kendi request'iyle yapılır. Bir bağlantı, açıldığı andaki
  cookie'lerle render eder. İmzalı, stateless cookie'ler kullanıyorsanız
  `maxStreamAge` bağlantıyı bu süre sonunda kapatır ve client onu o anda elinde
  olan cookie'lerle hemen yeniden açar.
- Stream `<prefix>stream/` adresinde sunulur. Kendi compression middleware'iniz
  `text/event-stream`'e dokunmamalıdır. Aksi hâlde stream ancak bittiğinde ulaşır.
  `noStream` açıkken client yalnızca polling yapar.

### elagoht/websocket

[github.com/Elagoht/collage-websocket](https://github.com/Elagoht/collage-websocket),
collage-live'ın gönderdiği fragment'leri bir event stream yerine bir WebSocket
üzerinden taşır.

```go
import (
	live "github.com/Elagoht/collage-live"
	"github.com/Elagoht/collage-websocket"
)

lv := live.New()
Plugins: []collage.Plugin{lv, websocket.New(lv)},
```

- collage-live'ı da bu plugin'den önce register edin. v0.2.0, collage v0.19.0 ve
  collage-live v0.2.0 ya da sonrasını gerektirir.
- Başka hiçbir şey değişmez. Layout yine `{{liveClient}}`'ı içerir, bu artık
  client'a buraya bağlanmasını söyler. Element'ler de yine `data-collage-push`
  taşır. collage-live kendi event stream'ini sunmayı bırakır. WebSocket da aynı
  shared worker'dan açılır. Bu yüzden tüm sekmeler yine tek bir bağlantıyı
  paylaşır ve `maxStreamAge` burada da geçerlidir.
- **Genellikle buna ihtiyacınız olmaz.** Fragment göndermek tek yönlüdür ve event
  stream tam da bunun içindir: bağımlılık gerektirmez ve kendiliğinden yeniden
  bağlanır. Bu plugin, stream'lerin bozulduğu deploy'lar içindir. Örneğin
  stream'leri buffer'layan bir proxy ya da onları sınırlayan bir platform. Live
  ailesinin bağımlılığı olan tek parçasıdır: `github.com/coder/websocket`.
- Bir bağlantı okuyucunun cookie'lerini taşır. Bu yüzden varsayılan olarak yalnızca
  sitenin kendi page'leri bağlantı açabilir.
  `websocket.NewWith(lv, websocket.Options{...})`; `Path`'i (`/_live/ws/`), `Ping`
  aralığını ve ayrıca bağlanabilecek `OriginPatterns`'ı belirler.

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
