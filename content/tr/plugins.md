---
description: Bir plugin'in neler yapabildiği, bir plugin'in nasıl register edilip yapılandırıldığı ve ne işe yaradıklarına göre gruplanmış, yayımlanmış otuz dört plugin.
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
- **Her request'i sarmalayabilir.** Bunu, uygulamanınkinden sonra gelen kendi
  middleware'iyle yapar (v0.21.0'dan beri).
- **Çıktıyı denetleyip finding raporlayabilir.** Finding'ler development'ta page'in
  üzerinde gösterilir, static build'in raporunda listelenir ve error seviyesinde
  build'i başarısız kılar (v0.21.0'dan beri).
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
Yayımlanmış iki plugin nerede duracağını söyler:
[elagoht/compress](#elagohtcompress) response body'lerini yeniden yazan her
plugin'den önce, [elagoht/devtoolbar](#elagohtdevtoolbar) ise en sona gelir.

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

Framework ile birlikte otuz dört plugin yayımlanmıştır. Aşağıda ne işe
yaradıklarına göre gruplanmışlardır. Her biri ayrı bir modüldür ve her birinin tam
referans niteliğinde kendi README'si vardır. Aşağıdaki bilgiler bir plugin'i
kurmanız için yeterlidir.

| Grup | Plugin'ler |
| --- | --- |
| [SEO ve keşfedilebilirlik](#seo-and-discovery) | jsonld, meta, sitemap, robots, feed, redirects, indexnow |
| [İçerik](#content) | markdown, highlight, toc, search, i18n |
| [Form'lar ve state](#forms-and-state) | validate, honeypot, flash, session |
| [Güvenlik](#security) | secure, ratelimit, basicauth |
| [Canlı güncellemeler](#live-updates) | live, websocket |
| [Asset'ler ve teslimat](#assets-and-delivery) | minimizer, opti-image, bundle, favicon, compress, cdnpurge, offline |
| [Operasyon ve development](#operations-and-development) | htmlcheck, devtoolbar, accesslog, prometheus, otel, analytics |

### SEO ve keşfedilebilirlik

Bir arama motorunun, bir feed okuyucunun ya da bir link önizlemesinin okudukları:
structured data, bir page'in paylaşılırken okunan tag'leri, sitemap ve
`robots.txt`, feed'ler, eski adresler ve nereye gittikleri, arama motorlarına neyin
değiştiğini bildirmek.

#### elagoht/jsonld

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

- collage v0.24.0 ya da sonrasını gerektirir. `Configure` aşaması yoktur, bu
  yüzden `RegisterPlugin` de bu plugin'i kabul eder.
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

#### elagoht/meta

[github.com/Elagoht/collage-meta](https://github.com/Elagoht/collage-meta), bir
page'in paylaşılırken ve index'lenirken okunan tag'lerini head'ine yazar: Open
Graph, Twitter card'ları, canonical URL, meta description ve page'in çevirileri.

```go
import "github.com/Elagoht/collage-meta"

Plugins: []collage.Plugin{meta.New(meta.Options{
	SiteName:     "The blog",
	BaseURL:      "https://example.com",
	DefaultImage: "/static/share.png",
})},
```

```go
meta.Set(rc, meta.Page{
	Title:       article.Title,
	Description: article.Dek,
	Image:       article.Cover,
	Type:        meta.Article,
	Published:   article.PublishedAt,
})
```

```json
{
  "elagoht/meta": {
    "siteName": "The blog",
    "baseURL": "https://example.com",
    "defaultImage": "/static/share.png",
    "twitterSite": "@example",
    "locales": { "en": "en_US", "tr": "tr_TR" },
    "noAlternates": false
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. `baseURL` zorunludur; bu ayar olmadan
  uygulama başlamaz.
- Register etmek her page'e şunları verir: `og:site_name`, `og:type`, canonical URL,
  `og:locale`, page'in path'i olan her locale için bir `hreflang` link'i, varsayılan
  görsel ve Twitter card'ı. Canonical URL ada göre oluşturulur, bu yüzden okuyucunun
  geldiği query string asla onun parçası olmaz.
- Page neyle ilgili olduğunu data handler'ından `meta.Set` ile söyler. Her tag kendi
  key'i altında hoist edilir. Bu yüzden page'in tanımı varsayılanın yerini tag tag
  alır, daha derindeki bir fragment'inki de üst fragment'inkinin yerini alır.
- `<title>` bunlardan biri değildir: plugin yokken olduğu gibi `rc.HoistTitle` ya da
  `WithTitle`'dır. Layout'ta `{{hoist "head"}}` gerektirir.

#### elagoht/sitemap

[github.com/Elagoht/collage-sitemap](https://github.com/Elagoht/collage-sitemap),
uygulamanın register ettiği page'lerden `/sitemap.xml`'i sunar.

```go
import "github.com/Elagoht/collage-sitemap"

Plugins: []collage.Plugin{sitemap.New(sitemap.Options{
	BaseURL: "https://example.com",
})},
```

```json
{
  "elagoht/sitemap": {
    "baseURL": "https://example.com",
    "path": "/sitemap.xml",
    "exclude": ["thanks"],
    "maxURLs": 50000
  }
}
```

- collage v0.21.0 ya da sonrasını gerektirir. `baseURL` zorunludur, çünkü bir
  sitemap mutlak URL'ler listeler. Bu ayar olmadan uygulama başlamaz.
- Path'i olan her page'i, her locale'de, `App.URL`'in yazdığı biçimde listeler.
  Page'in diğer locale'leri `hreflang` alternate'leri olarak eklenir. Bir
  `{param}` pattern'i, `WithStaticParams`'ının döndürdüğü her değer için bir kez
  listelenir; bunlar static build'in yazdığı URL'lerdir. `WithStaticParams`'ı
  olmayan bir pattern dışarıda kalır. `Exclude`, page'leri adlarıyla dışarıda
  bırakır.
- Bir page'in `<lastmod>`'unu bir Go fonksiyonu olan `LastMod` verir.
- Static bir document'tır: cache'lenir, export edilir ve `sitemap.Tag` invalidate
  edildiğinde yeniden üretilir. Yayımladığınız bir yazının tag'leriyle birlikte
  onu da invalidate edin.
- 50.000 URL'yi (`maxURLs`) aşınca numaralı dosyalardan oluşan bir sitemap
  index'ine dönüşür.

#### elagoht/robots

[github.com/Elagoht/collage-robots](https://github.com/Elagoht/collage-robots),
`/robots.txt`'yi sunar.

```go
import "github.com/Elagoht/collage-robots"

Plugins: []collage.Plugin{robots.New(robots.Options{
	Rules:    []robots.Rule{{Disallow: []string{"/admin"}}},
	Sitemaps: []string{"https://example.com/sitemap.xml"},
})},
```

```json
{
  "elagoht/robots": {
    "rules": [{ "userAgents": ["*"], "disallow": ["/admin"] }],
    "sitemaps": ["https://example.com/sitemap.xml"],
    "disallowAll": false
  }
}
```

- collage v0.21.0 ya da sonrasını gerektirir. Hiç kural yoksa her crawler'a her
  şeye izin verir. User agent belirtmeyen bir kural `*` içindir.
- `disallowAll`, kurallar ne derse desin siteyi bütün crawler'lara kapatır ve her
  response'u `X-Robots-Tag: noindex, nofollow` ile gönderir. Bunu bir staging
  deploy'unun config'inde açın. Böylece aynı binary production'da açık, başka
  yerlerde kapalı olur.
- Body uygulama başlarken sabitlenir ve static build onu `robots.txt` olarak
  yazar.

#### elagoht/feed

[github.com/Elagoht/collage-feed](https://github.com/Elagoht/collage-feed),
uygulamanın listelediği öğelerden RSS 2.0 ve Atom 1.0 feed'leri sunar ve bunları
her page'in head'inde duyurur.

```go
import "github.com/Elagoht/collage-feed"

Plugins: []collage.Plugin{feed.New(feed.Feed{
	Title:   "The blog",
	BaseURL: "https://example.com",
	Link:    "/blog",
	Items:   latestPosts, // func(ctx) ([]feed.Item, error), newest first
	Tags:    []string{"posts"},
})},
```

- collage v0.21.0 ya da sonrasını gerektirir. `Items` bir fonksiyon olduğu için
  yalnızca Go'da yapılandırılır.
- Bir feed RSS olarak `/feed.xml`'de, Atom olarak `/atom.xml`'de sunulur. `RSS` ve
  `Atom` bu path'leri değiştirir, `"-"` ise bir formatı dışarıda bırakır. Birden
  fazla feed'in her biri bir `Name` ve kendi path'lerini alır. Bir feed en fazla
  `Limit` kadar öğe taşır; varsayılan 20'dir.
- Her page, her feed için bir `<link rel="alternate">` alır. Bu yüzden layout'ta
  `{{hoist "head"}}` bulunması gerekir. `NoDiscovery`, bir feed'i head'lerin
  dışında tutar.
- Static bir document'tır: cache'lenir, export edilir ve `Tags`'inden biri
  invalidate edildiğinde yeniden üretilir.

#### elagoht/redirects

[github.com/Elagoht/collage-redirects](https://github.com/Elagoht/collage-redirects),
kodda değil bir dosyada tutulan redirect'leri routing'den önce sunar. Bunlar bir
site taşımasının geride bıraktıklarıdır. Plugin, redirect'leri static bir host için
`_redirects` olarak da yazar.

```go
import "github.com/Elagoht/collage-redirects"

//go:embed redirects.txt
var siteFS embed.FS

Plugins: []collage.Plugin{redirects.New(redirects.Options{FS: siteFS})},
```

```
# the old blog
/blog/*          /posts/:splat
/about-us        /about          301
/summer-sale     /sale           302
/old-product     -               410
```

```json
{
  "elagoht/redirects": {
    "file": "redirects.txt",
    "rules": [{ "from": "/careers", "to": "https://jobs.example.com", "status": 302 }],
    "noRedirectsFile": false
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir.
- Her satırda bir kural vardır: eski path, gittiği yer ve bir status. Status
  yazılmazsa `301`'dir; `302`, `307`, `308` ya da artık olmayan bir page için `-`
  ile birlikte `410` olabilir. Bu durumda sitenin kendi not-found page'i `410`
  status'uyla sunulur. `/blog/*` bir prefix'tir ve hedefteki `:splat`, `*`'ın
  eşleştiği kısımdır. Eşleşen ilk kural kazanır, okuyucunun query string'i de
  taşınır.
- Hatalı bir dosya uygulamanın başlamasını engeller ve hatanın yerini söyler: bozuk
  bir satır, hiçbir request'in ulaşamayacağı bir kural, okuyucuyu bir döngüde
  dolaştıran kurallar.
- Static build, kuralları Netlify ve Cloudflare Pages'in okuduğu biçimde
  `_redirects`'e yazar; `noRedirectsFile` bunu dışarıda bırakır. Kurallar Go'da ya da
  config'te de verilebilir.

#### elagoht/indexnow

[github.com/Elagoht/collage-indexnow](https://github.com/Elagoht/collage-indexnow),
hangi URL'lerin değiştiğini IndexNow protokolüyle arama motorlarına bildirir. Bing,
Yandex, Seznam, Naver ve birine söyleneni paylaşan diğerleri bu protokolü kullanır.

```go
import "github.com/Elagoht/collage-indexnow"

Plugins: []collage.Plugin{indexnow.New(indexnow.Options{
	Key:     "4f1c2e9a7b3d4c8e",
	BaseURL: "https://example.com",
})},
```

```json
{
  "elagoht/indexnow": {
    "key": "4f1c2e9a7b3d4c8e",
    "baseURL": "https://example.com",
    "window": "10s",
    "exclude": ["/api/", "/search"]
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. `key` ve `baseURL` zorunludur.
- Gönderdiği, cache'in bıraktıklarıdır: yani
  [bir invalidation'ın düşürdüğü path'ler](/docs/caching#invalidating-by-path). Tag'i
  invalidate edildiğinde cache'te olmayan bir page gönderilmez.
- URL'ler `window` boyunca toplanır, arka planda gönderilir ve bir `429` ya da `5xx`
  gelirse yeniden denenir. Bekleyenler shutdown'da gönderilir.
- Key'i `/<key>.txt`'de sunar, static build da onu oraya yazar. Bir development
  sunucusu ve bir static build hiçbir şey göndermez.

### İçerik

Metnin nereden geldiği ve nasıl sunulduğu: page verisi olarak Markdown dosyaları,
renklendirilmiş kod, içindekiler tablosu, static bir sitede arama ve çeviriler.

#### elagoht/markdown

[github.com/Elagoht/collage-markdown](https://github.com/Elagoht/collage-markdown),
bir Markdown dosyaları dizinini page verisine dönüştürür: YAML front matter,
heading id'leri ve dipnotlarıyla GitHub-flavoured Markdown, dosya başına bir
dependency tag.

```go
import "github.com/Elagoht/collage-markdown"

//go:embed content
var content embed.FS

md := markdown.New(markdown.Options{FS: content, Dir: "content/blog"})

Plugins: []collage.Plugin{md},
```

```go
app.RegisterPage(collage.NewPage("post").
	WithContent(collage.NewFragment("post", "pages/post.html").
		WithDataHandler(md.Handler()).
		Required().
		Build()).
	WithPath("en", "/blog/{slug}").
	Static().
	WithStaticParams(md.StaticParams()).
	Build())
```

```json
{
  "elagoht/markdown": {
    "dir": "content/blog",
    "localeDirs": { "tr": "content/blog/tr" },
    "drafts": false,
    "unsafe": false
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir. `Config.Plugins` içinde ya da
  `RegisterPlugin` ile register edilebilir: config'ini ve dosyalarını uygulama
  başlarken okur, static build de bunu yazacağı page'leri listelemeden önce yapar.
- `md.Handler()`, template'e `slug`'ın adlandırdığı `Doc`'u verir (`Title`,
  `Description`, `Date`, `Tags`, `HTML`, `Text`, `Headings`) ya da
  `collage.ErrNotFound` döner. `md.IndexHandler()`, `md.List` ve `md.Get` bir index
  page'ini, bir feed'i ve bir sitemap'i besler.
- Bir düzenleme tek bir invalidation'dır: `md.Tag(locale, slug)`. Eklenen ya da
  silinen bir dosya ise `md.DirTag(locale)`'dır. Development'ta her request dosyayı
  yeniden okur.
- Bir locale'in kendi dizini olabilir. Ham HTML, `unsafe` açık değilse dışarıda
  bırakılır. Birden fazla küme, örneğin bir blog ve dokümantasyon, `Name` ile ayırt
  edilen birden fazla plugin'dir.

#### elagoht/highlight

[github.com/Elagoht/collage-highlight](https://github.com/Elagoht/collage-highlight),
kodu chroma ile renklendirir: render edilen her page'deki code block'ları, bir
`{{highlight}}` template fonksiyonu ve açık ile koyu bir stylesheet.

```go
import "github.com/Elagoht/collage-highlight"

Plugins: []collage.Plugin{highlight.New(highlight.Options{})},
```

```html
{{highlight .Snippet "go"}}
```

```json
{
  "elagoht/highlight": { "light": "github", "dark": "github-dark", "noBackground": true, "auto": true }
}
```

- v0.2.1, collage v0.26.0 ya da sonrasını gerektirir; v0.2.0 v0.25.0'ı, v0.1.0
  v0.23.0'ı gerektiriyordu. `Config.Plugins` içinde olmalıdır: `{{highlight}}`'ı ekler.
- `auto` açıkken bir page'in içerdiği her `<pre><code class="language-go">`
  bulunduğu yerde, render başına bir kez renklendirilir. elagoht/markdown ve çoğu
  Markdown renderer'ı bu biçimi yazar. Cache'lenen bir page renklendirildiği hâliyle
  sunulur.
- Stylesheet yalnızca renklendirilmiş kod içeren page'lerden, içeriğe göre
  adlandırılmış bir adla link'lenir. Her tema kendi `prefers-color-scheme` sorgusu
  altındadır. v0.2.0'dan beri `auto` geçişi stylesheet'i `AfterRenderEvent.Hoist`
  ile hoist eder. Böylece stylesheet, layout'un `{{hoist "head"}}` koyduğu yere,
  o yoksa `</head>`'in önüne yerleşir. `{{highlight}}` ise layout'ta
  `{{hoist "head"}}` gerektirir.

#### elagoht/toc

[github.com/Elagoht/collage-toc](https://github.com/Elagoht/collage-toc), bir
page'in heading'lerine id verir ve template'lerin istediği yere bir içindekiler
tablosu ve okuma süresi koyar.

```go
import "github.com/Elagoht/collage-toc"

Plugins: []collage.Plugin{toc.New(toc.Options{})},
```

```html
<aside>{{toc}}</aside>
<p>{{readingTime}}</p>
```

```json
{
  "elagoht/toc": {
    "minLevel": 2,
    "maxLevel": 4,
    "wpm": 225,
    "locales": { "tr": { "label": "İçindekiler", "readingTime": "{n} dk okuma" } }
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  template fonksiyonları ekler.
- Her fonksiyon bir placeholder yazar. Plugin, page render edildikten sonra onu
  doldurur: `<main>` içindeki `h2`–`h4` heading'lerini iç içe bir liste olarak ve
  kelime sayısını `wpm`'e bölerek.
- `id`'si olmayan bir heading, metninden bir id alır. Her alfabenin harfleri korunur
  ve page'in dilinin küçülttüğü gibi küçük harfe çevrilir.
- Kendi başına sunulan bir fragment page hook'larını çalıştırmaz. Bu yüzden
  `{{toc}}`'u kendi başına yenilenen bir fragment'e değil, page'e koyun.

#### elagoht/search

[github.com/Elagoht/collage-search](https://github.com/Elagoht/collage-search),
static bir siteye arama ekler: static build her page'in bir index'ini yazar, küçük
bir script de bu index'te tarayıcıda arama yapar.

```go
import "github.com/Elagoht/collage-search"

Plugins: []collage.Plugin{search.New(search.Options{})},
```

```html
{{searchBox}}
```

```json
{
  "elagoht/search": {
    "maxText": 5000,
    "exclude": ["/admin/", "/tags/"],
    "locales": { "tr": { "label": "Ara", "noResults": "Sonuç yok" } }
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{searchBox}}`'ı ekler.
- **Index bir static build'e aittir.** `collage export`, her page'in title'ından,
  description'ından, heading'lerinden ve metninden `search-index.json`'ı yazar.
  Çalışan bir sunucuda index yoktur; arama kutusu bozulmak yerine bunu söyler.
- Script page ile birlikte çekilir ve index'i kutuya ilk odaklanıldığında yükler.
  Her kelimeyi büyük/küçük harf ve aksan farkı gözetmeden eşleştirir, title'ları
  heading'lerin, heading'leri metnin üstünde sıralar ve page'in diliyle sınırlı
  kalır.

#### elagoht/i18n

[github.com/Elagoht/collage-i18n](https://github.com/Elagoht/collage-i18n),
çeviri yapar: her locale için bir katalog, template'lerde page'in render edildiği
locale'de `{{t}}`, çoğul biçimler ve finding olarak raporlanan eksik çeviriler.

```go
import "github.com/Elagoht/collage-i18n"

//go:embed locales
var locales embed.FS

Plugins: []collage.Plugin{i18n.New(i18n.Options{FS: locales})},
```

```html
<a href="{{pageURL "home"}}">{{t "nav.home"}}</a>
<p>{{tn "cart" .Count}}</p>
```

```json
{ "elagoht/i18n": { "dir": "locales" } }
```

- collage v0.22.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  template fonksiyonları ekler.
- Desteklenen her locale için iç içe key'lerden oluşan bir JSON dosyası olur:
  `locales/<locale>.json`. Desteklenen bir locale'in dosyası yoksa uygulama
  başlamaz.
- `t` bir key'i çevirir ve `{name}`'i ad-değer çiftlerinden doldurur. `tn` bir
  sayı için çoğul biçimi seçer, `th` ise katalogdan gelen markup'a izin verir. Bir
  data handler `i18n.T(rc, key, pairs...)` çağırır.
- Eksik bir key önce varsayılan locale'e, sonra key'in kendisine düşer.
  Development'ta page'in üzerinde, static build'de ise build raporunda
  `missing-translation` olarak raporlanır. Development'ta kataloglar her
  request'te yeniden okunur.

### Form'lar ve state

Bir form'un bir [action](/docs/forms-and-actions) etrafında ihtiyaç duydukları:
validation, spam koruması, redirect'ten sonra gösterilen bir mesaj ve bir cookie'de
tutulan session.

#### elagoht/validate

[github.com/Elagoht/collage-validate](https://github.com/Elagoht/collage-validate),
bir action'ın aldığı form'u doğrular: alan başına zincirlenebilen kontroller, status
422 ile yeniden render edilen reddedilmiş bir gönderim ve her alanın mesajını ve
okuyucunun yazdığını form'a geri koyan template fonksiyonları.

```go
import "github.com/Elagoht/collage-validate"

Plugins: []collage.Plugin{validate.New(validate.Options{})},
```

```go
v := validate.Form(rc)
v.Field("email").Required().Email()
v.Field("password").Required().MinLen(8)
if !v.Valid() {
	return validate.Refuse(rc, v, signupPage), nil
}
return collage.SeeOther("/welcome"), nil
```

```html
<input name="email" type="email" value="{{fieldValue "email"}}">
{{with fieldError "email"}}<p class="error">{{.}}</p>{{end}}
```

```json
{
  "elagoht/validate": {
    "messages": { "required": "Please fill this in." },
    "localeMessages": { "tr": { "required": "Bu alan zorunludur." } },
    "noRefill": ["password", "card"]
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{fieldError}}`, `{{fieldValue}}` ve `{{hasErrors}}`'ı ekler.
- collage'ın form kuralının ilk yarısıdır. Bkz.
  [Form'lar ve action'lar](/docs/forms-and-actions#the-validation-re-render). Kimsenin
  göndermediği bir page'de fonksiyonlar boştur. Bu yüzden tek bir template iki
  render'a da hizmet eder ve ilk render cache'lenebilir kalır.
- Kontroller `Required`, `MinLen`, `MaxLen`, `Email`, `URL`, `Int`, `Range`, `OneOf`,
  `Matches`, `Equal` ve `Custom`'dır. `v.Fail`, yalnızca uygulamanın bildiği bir
  durumu raporlar. Bir alan ilk mesajını korur.
- Mesajlar varsayılan olarak İngilizcedir. Kontrol başına, locale başına ya da her
  locale için değiştirilebilir. Adında `password` geçen bir alan asla geri
  doldurulmaz.

#### elagoht/honeypot

[github.com/Elagoht/collage-honeypot](https://github.com/Elagoht/collage-honeypot),
form spam'ini CAPTCHA olmadan durdurur. İnsanların hiç görmediği ama bot'ların
doldurduğu bir tuzak alan kullanır. İmzalı bir zaman damgası da bir insanın
doldurabileceğinden daha çabuk geri gönderilen form'u reddeder.

```go
import "github.com/Elagoht/collage-honeypot"

Plugins: []collage.Plugin{honeypot.New(honeypot.Options{Key: key})},
```

```html
<form method="post" action="/contact">
  {{csrfToken}}
  {{honeypot}}
  <textarea name="message"></textarea>
  <button>Send</button>
</form>
```

```json
{
  "elagoht/honeypot": {
    "key": "hex-encoded, 32 bytes or more",
    "minDelay": 2,
    "maxAge": 86400,
    "silent": false,
    "protect": ["/"]
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{honeypot}}`'ı ekler.
- Korunan bir path'e gönderilen her form body'si action'a ulaşmadan önce kontrol
  edilir. Bu yüzden böyle her form `{{honeypot}}` taşımalıdır. JSON body'ler ve her
  `GET` kontrol edilmeden geçer.
- Zaman damgası, collage'ın forgery token'ı gibi page cache'ten sağ çıkar: cache'lenen
  page bir placeholder taşır, plugin'in middleware'i de o anki zamanı imzalayıp onun
  yerine koyar. En az 32 rastgele byte'lık, her instance'ta aynı olan bir key
  ayarlayın.
- Ret bir `400`'dür. `silent` ile ise kabul edilmiş bir form'un cevabı gibi, form'a
  geri dönen bir `303`'tür. Dikkatsiz bot'ları durdurur, kararlı birini durdurmaz;
  onu elagoht/ratelimit ile birlikte kullanın.

#### elagoht/flash

[github.com/Elagoht/collage-flash](https://github.com/Elagoht/collage-flash),
flash mesajları ekler: bir action'ın redirect etmeden önce ayarladığı ve redirect
ettiği page'in bir kez gösterdiği mesaj.

```go
import "github.com/Elagoht/collage-flash"

Plugins: []collage.Plugin{flash.New(flash.Options{Key: key})},
```

```go
flash.Add(rc, flash.Success, "Your changes are saved.")
return collage.SeeOther("/settings"), nil
```

```html
{{range flashes}}
  <p class="flash flash--{{.Kind}}" role="status">{{.Text}}</p>
{{end}}
```

```json
{
  "elagoht/flash": {
    "key": "hex-encoded, 32 bytes or more",
    "cookie": "collage_flash",
    "maxAge": 300
  }
}
```

- collage v0.22.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{flashes}}`'ı ekler.
- Mesajlar imzalı, `HttpOnly` bir cookie içinde taşınır. En az 32 rastgele
  byte'lık, her instance'ta aynı olan bir key ayarlayın. Key yoksa her process için
  bir key üretilir ve bir uyarı log'lanır.
- Mesaj taşıyan bir request yeniden render edilir. Page cache'ten okunmaz, ona
  yazılmaz da ve `private, no-store` olarak işaretlenir. Diğer her request,
  plugin yokmuş gibi sunulur.

#### elagoht/session

[github.com/Elagoht/collage-session](https://github.com/Elagoht/collage-session),
session'ı bir cookie'de tutar: sitenin imzaladığı ve şifreleyebildiği küçük bir
string map'i. Arkasında bir veritabanı yoktur.

```go
import "github.com/Elagoht/collage-session"

Plugins: []collage.Plugin{session.New(session.Options{Key: key})},
```

```go
s := session.Get(rc)
s.Regenerate()
if err := s.Set("user", user.ID); err != nil {
	return nil, err
}
return collage.SeeOther("/account"), nil
```

```json
{
  "elagoht/session": {
    "key": "hex-encoded, 32 bytes or more",
    "encrypt": true,
    "maxAge": 604800,
    "idleTimeout": 0,
    "sameSite": "lax"
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. Template fonksiyonu eklemediği için
  `RegisterPlugin` de onu kabul eder.
- Metotları `Get`, `Set`, `Delete`, `Clear`, `Regenerate` ve `ID`'dir. Kendi
  handler'ınız session'ı `session.FromContext(r.Context())` ile okur.
- **Geçerli bir session taşıyan request yeniden render edilir.** Page cache'ten
  okunmaz, ona yazılmaz da ve `private, no-store` olarak işaretlenir. Session'ı
  olmayan bir okuyucu eskisi gibi cache'ten sunulur. Session'ı cache'lenen bir
  page'den değil, bir action'dan ayarlayın.
- Cookie `HttpOnly` ve `SameSite=Lax`'tir ve yalnızca session değiştiğinde yazılır.
  Key'ler `previousKeys` ile döndürülür. Bir session iptal edilemez, çünkü
  okuyucunun cookie'sinde yaşar.

### Güvenlik

Bir sitenin göndermesi gereken header'lar, tek bir istemcinin siteye ne kadar hızlı
istek atabileceğine bir sınır ve henüz herkese açık olmayan bir sitenin önünde bir
parola.

#### elagoht/secure

[github.com/Elagoht/collage-secure](https://github.com/Elagoht/collage-secure),
bir sitenin göndermesi gereken güvenlik header'larını ve nonce'ları page cache'ten
sağ çıkan bir Content-Security-Policy'yi gönderir.

```go
import "github.com/Elagoht/collage-secure"

Plugins: []collage.Plugin{secure.New(secure.Options{
	CSP: "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
})},
```

```json
{
  "elagoht/secure": {
    "csp": "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
    "cspReportOnly": false,
    "hsts": 63072000,
    "hstsSubdomains": true,
    "frameOptions": "DENY",
    "permissionsPolicy": "camera=(), microphone=(), geolocation=()"
  }
}
```

- collage v0.22.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{cspNonce}}`'ı ekler.
- Varsayılan olarak `X-Content-Type-Options`, `X-Frame-Options`,
  `Referrer-Policy` ve `Cross-Origin-Opener-Policy` gönderir.
  `Strict-Transport-Security`'yi ise TLS üzerinden ya da `X-Forwarded-Proto: https`
  gönderen bir proxy'nin arkasında gönderir. `Permissions-Policy` ve CSP,
  ayarlandıklarında gönderilir. `"-"` bir header'ı dışarıda bırakır.
- Policy'deki `{nonce}` ile inline bir script'teki `{{cspNonce}}` aynı nonce'tur ve
  her response'ta yenidir. Cache'lenen page bir placeholder taşır, plugin'in
  middleware'i de onun yerine taze bir nonce koyar. Nonce taşıyan bir page
  `Cache-Control: no-store` ile ve `ETag` olmadan gönderilir.
- Development'ta policy report-only olarak gönderilir. Böylece collage'ın
  live-reload script'i çalışmaya devam eder.

#### elagoht/ratelimit

[github.com/Elagoht/collage-ratelimit](https://github.com/Elagoht/collage-ratelimit),
tek bir istemcinin siteye ne kadar hızlı istek atabileceğini sınırlar: istemci ve
kural başına bir token bucket, bucket boşalınca da `Retry-After` ile bir
`429 Too Many Requests`.

```go
import "github.com/Elagoht/collage-ratelimit"

Plugins: []collage.Plugin{ratelimit.New(ratelimit.Options{
	Rules: []ratelimit.Rule{
		{PathPrefix: "/login", Rate: 0.1, Burst: 5},
		{Rate: 0.5, Burst: 10},
	},
})},
```

```json
{
  "elagoht/ratelimit": {
    "rules": [{ "pathPrefix": "/login", "rate": 0.1, "burst": 5 }],
    "trustProxy": true,
    "trustedProxies": ["10.0.0.0/8"],
    "skip": ["/_collage/", "/healthz"]
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir. Hiç seçenek verilmezse her form ve
  action, yani `GET`, `HEAD` ve `OPTIONS` dışındaki her metot, önce ondan oluşan bir
  burst'le, ardından iki saniyede bir request'le sınırlanır.
- Bir request'i, eşleştiği ilk kural sayar. Bu yüzden dar kuralları başa koyun; her
  kuralın kendi bucket'ları vardır. Response'lar `RateLimit-Limit`,
  `RateLimit-Remaining` ve `RateLimit-Reset` taşır.
- Bir istemci IP adresidir, IPv6'da ise `/64`'üdür. Bir reverse proxy'nin arkasında
  `trustProxy`'yi ayarlayın. Adres o zaman `X-Forwarded-For`'dan okunur ve yalnızca
  güvenilen bir proxy'den geldiğinde inanılır. Go'daki `KeyFunc` başka bir şeye göre
  key üretir.
- Bucket'lar memory'de tutulur, bu yüzden sınırlar process başınadır.

#### elagoht/basicauth

[github.com/Elagoht/collage-basicauth](https://github.com/Elagoht/collage-basicauth),
bir sitenin önüne HTTP Basic authentication koyar: bir staging deploy'u, bir
preview ya da henüz yayına girmemiş bir site için.

```go
import "github.com/Elagoht/collage-basicauth"

Plugins: []collage.Plugin{basicauth.New(basicauth.Options{
	Users: map[string]string{"team": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
})},
```

```json
{
  "elagoht/basicauth": {
    "realm": "Staging",
    "protect": ["/"],
    "skip": ["/healthz", "/_collage/"],
    "disabled": false
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir. Hiç kullanıcı yoksa uygulama başlamaz.
- Parola düz metin olarak, `sha256:` ve hex'i olarak ya da bir bcrypt hash'i olarak
  yazılır. `COLLAGE_BASICAUTH_USERS`, kullanıcıları environment'tan ekler ve
  secret'ları dosyaların dışında tutar.
- `disabled` ya da `COLLAGE_BASICAUTH_DISABLED=true`, aynı binary'nin bir deploy'unu,
  production'dakini, açık bırakır.
- Kimliği doğrulanmış her response'ta `public`, `private` ile değiştirilir ve
  response `Vary: Authorization` taşır. Böylece öndeki bir CDN, bir page'i sormadan
  sonraki okuyucuya vermez. Siteyi HTTPS üzerinden sunun.

### Canlı güncellemeler

Bir page'in parçalarını, istemcide bir framework olmadan tarayıcıda güncel tutmak.

#### elagoht/live

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

#### elagoht/websocket

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

- collage-live'ı da bu plugin'den önce register edin. v0.2.1, collage v0.24.0 ve
  collage-live v0.2.1 ya da sonrasını gerektirir.
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

### Asset'ler ve teslimat

Tarayıcının aldığı byte'ların neye benzediği ve oraya nasıl ulaştığı: minify
edilmiş, yeniden boyutlandırılmış, bundle edilmiş, sıkıştırılmış, bir CDN'den purge
edilmiş ve çevrimdışı okuma için saklanmış.

#### elagoht/minimizer

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

- collage v0.24.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır.
  Mount edilen dosya sistemlerini sarmalar ve bu işlem uygulama kurulurken yapılır.
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

#### elagoht/opti-image

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

- collage v0.24.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır.
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

#### elagoht/bundle

[github.com/Elagoht/collage-bundle](https://github.com/Elagoht/collage-bundle),
uygulamanın JavaScript, TypeScript ve CSS'ini uygulama başlarken esbuild ile bundle
eder, çıktıyı içeriğe göre hash'lenmiş adlarla sunar ve development'ta bir
düzenlemede yeniden build eder.

```go
import "github.com/Elagoht/collage-bundle"

Plugins: []collage.Plugin{bundle.New(bundle.Options{
	Dir:     "assets",
	Entries: []string{"app.ts", "app.css"},
})},
```

```html
<link rel="stylesheet" href="{{bundle "app.css"}}">
<script src="{{bundle "app.js"}}" defer></script>
```

```json
{
  "elagoht/bundle": {
    "dir": "assets",
    "entries": ["app.ts", "app.css"],
    "prefix": "/_bundle/",
    "target": "es2020"
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `{{bundle}}`'ı ekler.
- Çıktı `/_bundle/`'daki bir mount'tan sunulur. Her dosya, esbuild'in içeriğinden
  ürettiği hash'i taşıyan bir adla sunulur ve bir yıl cache'lenir. Build'in
  üretmediği bir ad, `{{asset}}`'in eksik bir dosyada yaptığı gibi render'ı
  başarısız kılar.
- Bundle'ı build edilemeyen bir production sunucusu başlamaz. Development'ta hata
  page'de gösterilir ve sunucu çalışmaya devam eder.
- Kaynaklar uygulama başlarken bir `fs.FS`'ten değil, diskten okunur. Static build
  mount'u kopyalar, bu yüzden build edilen site bundle'ı içerir.

#### elagoht/favicon

[github.com/Elagoht/collage-favicon](https://github.com/Elagoht/collage-favicon),
bir sitenin ikonlarını tek bir kaynak görselden üretir: `favicon.ico`, Apple touch
icon, web app ikonları ve onları listeleyen manifest. İkonları kökte sunar ve her
page'in head'inden link'ler.

```go
import "github.com/Elagoht/collage-favicon"

Plugins: []collage.Plugin{favicon.New(favicon.Options{
	Source:     "assets/icon.png",
	SVG:        "assets/icon.svg",
	Name:       "The blog",
	ThemeColor: "#0f172a",
})},
```

```json
{
  "elagoht/favicon": {
    "source": "assets/icon.png",
    "svg": "assets/icon.svg",
    "name": "The blog",
    "themeColor": "#0f172a",
    "backgroundColor": "#ffffff"
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. Kaynak kare bir PNG, JPEG ya da
  GIF'tir; ideal olarak 512 piksel ya da daha büyüktür. Diskten ya da `FS`'ten
  okunur.
- İkonlar başlangıçta bir kez üretilir ve static document olarak sunulur. Bu yüzden
  static build onları yazar. SVG olduğu gibi geçirilir, asla rasterize edilmez.
- Link'ler `?v=` ve kaynağın bir özetini taşır. Böylece değişen bir ikon, değişen bir
  URL demektir. Layout'ta `{{hoist "head"}}` gerektirir.

#### elagoht/compress

[github.com/Elagoht/collage-compress](https://github.com/Elagoht/collage-compress),
response'ları Brotli ve gzip ile sıkıştırır ve bir static build'in yanına `.br` ve
`.gz` dosyaları yazar.

```go
import "github.com/Elagoht/collage-compress"

Plugins: []collage.Plugin{
	compress.New(compress.Options{}),
	secure.New(secure.Options{CSP: "..."}),
},
```

```json
{
  "elagoht/compress": { "minSize": 512, "gzipLevel": 6, "brotliLevel": 4, "types": ["application/x-yaml"] }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. **Onu response body'lerini yeniden
  yazan her plugin'den önce register edin.** elagoht/secure da bunlardan biridir.
  İlk register edilen plugin en dıştaki middleware'dir.
- En az `minSize` byte'lık metin türleri, request'in kabul ettiği en iyi encoding ile
  sıkıştırılır. `text/event-stream`, bir WebSocket ve bir `Range` request'ine
  dokunulmaz.
- Sıkıştırılmış bir body ETag başına saklanır. Böylece collage'ın cache'ten sunduğu
  bir page, okuyucu başına değil, encoding başına bir kez sıkıştırılır. ETag
  encoding'i de içerir ve conditional bir request yine `304`'ünü alır.
- Static build, önceden sıkıştırılmış dosyaları sunan bir host için sıkıştırılabilen
  her dosyanın yanına bir `.br` ve bir `.gz` yazar; `noPrecompress` bunu kapatır.

#### elagoht/cdnpurge

[github.com/Elagoht/collage-cdnpurge](https://github.com/Elagoht/collage-cdnpurge),
collage'ın invalidate ettiği page'lerin CDN'deki kopyalarını purge eder: zone'a göre
Cloudflare'de ya da bir webhook üzerinden URL listesi alan herhangi bir serviste.

```go
import "github.com/Elagoht/collage-cdnpurge"

Plugins: []collage.Plugin{cdnpurge.New(cdnpurge.Options{
	BaseURL: "https://example.com",
	Cloudflare: &cdnpurge.Cloudflare{
		ZoneID:   os.Getenv("CF_ZONE_ID"),
		APIToken: os.Getenv("CF_API_TOKEN"),
	},
})},
```

```json
{
  "elagoht/cdnpurge": {
    "baseURL": "https://example.com",
    "webhook": { "url": "https://purge.example.com/", "token": "..." },
    "window": "2s",
    "retries": 4
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir. `baseURL` ve `cloudflare` ile
  `webhook`'tan en az biri zorunludur.
- Tam olarak [bir invalidation'ın düşürdüğü path'leri](/docs/caching#invalidating-by-path)
  `baseURL` altında purge eder. Origin'in cache'lemediği bir page adlandırılmaz. Bu
  yüzden CDN'in cache'lediği her şeyi collage'ın da cache'lediğinden emin olun.
- Purge'ler `window` boyunca toplanır, invalidate eden goroutine'in dışında
  gönderilir, bir `429` ya da `5xx` gelirse yeniden denenir ve shutdown'da gönderilir.
  Bir development sunucusu, `force` ayarlanmadıkça hiçbir şeyi purge etmez.
- API token'ını sürüm kontrolündeki bir dosyanın dışında tutun.

#### elagoht/offline

[github.com/Elagoht/collage-offline](https://github.com/Elagoht/collage-offline),
bir service worker sunar. Böylece bir ziyaretçinin okuduğu page'ler ağ olmadan da
yeniden açılır.

```go
import "github.com/Elagoht/collage-offline"

Plugins: []collage.Plugin{offline.New(offline.Options{
	Precache: []string{"/"},
	Fallback: "/offline",
})},
```

```html
<head>
  {{offlineScript}}
</head>
```

```json
{
  "elagoht/offline": {
    "precache": ["/", "/about"],
    "fallback": "/offline",
    "assets": ["/static/", "/_bundle/"],
    "maxPages": 100
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir ve `Config.Plugins` içinde olmalıdır:
  `/sw.js`'de sunulan worker'ı kuran `{{offlineScript}}`'i ekler.
- Page'ler önce ağdan çekilir ve saklanır. `assets` altındaki static dosyalar
  stale-while-revalidate ile sunulur. Ne ulaşılabilen ne de saklanan bir page için
  `fallback` page'i gösterilir. `no-store` olarak işaretlenmiş bir response asla
  saklanmaz.
- Worker'ın cache'leri her deploy'la değişen bir sürümle adlandırılır. Böylece yeni
  bir build, eskisinin sakladıklarının yerini alır. Build, ayarlanmışsa
  `version`'dır; değilse collage'ın `Host.BuildID`'sidir, yani
  `Config.Cache.Version` ya da çalıştırılabilir dosyanın parmak izi.
- Development'ta `/sw.js` kendi kaydını siler. Static build `sw.js`'i yazar; bir
  CDN'in onu uzun süre tutmasını engelleyin.

### Operasyon ve development

Bir sitenin render ettiğini denetlemek, development'ta bir render'ı görmek ve
çalışan bir siteyi izlemek: access log'ları, metric'ler, trace'ler ve analytics.

#### elagoht/htmlcheck

[github.com/Elagoht/collage-htmlcheck](https://github.com/Elagoht/collage-htmlcheck),
bir sitenin render ettiği HTML'i denetler: yapıyı, erişilebilirliği, bir arama
motorunun okuduklarını, bir page'i yavaşlatanları ve page'ler arasındaki link'leri.
Bulduklarını [finding](/docs/writing-plugins#checking-the-output-findings) olarak
raporlar.

```go
import "github.com/Elagoht/collage-htmlcheck"

Plugins: []collage.Plugin{htmlcheck.New(htmlcheck.Options{})},
```

```json
{
  "elagoht/htmlcheck": {
    "rules": { "img-dimensions": "off", "heading-order": "error" },
    "titleMax": 60,
    "descriptionMax": 160,
    "pageBudget": 200000,
    "ignoreLinks": ["/api/"]
  }
}
```

- collage v0.22.0 ya da sonrasını gerektirir.
- Development'ta her page render edilirken denetlenir ve bulunanlar page'in
  üzerinde gösterilir. Static build'de önce her page, sonra build'in bütünü
  denetlenir: iki page'in paylaştığı title'lar, build'in yazmadığı page'lere
  giden link'ler. Bir error build'i başarısız kılar (bkz.
  [Static export](/docs/static-export#findings)). Production sunucusunda hiçbir
  şey denetlenmez.
- 22 kuralı vardır: `html-lang`, `title`, `img-alt`, `input-label`,
  `duplicate-id`, `heading-order`, `broken-link` ve diğerleri. Her biri
  varsayılan olarak `error` ya da `warn` seviyesindedir. `rules` bir kuralın
  seviyesini değiştirir ya da onu `off` ile kapatır. Plugin'in tanımadığı bir
  kural adı uygulamanın başlamasını engeller. `htmlcheck.Rules()` hepsini
  listeler.

#### elagoht/devtoolbar

[github.com/Elagoht/collage-devtoolbar](https://github.com/Elagoht/collage-devtoolbar),
development'ta her page'in altında küçük bir panel gösterir: hangi page'in, hangi
locale'de, hangi status'la render edildiği, render'ın ve her fragment'inin ne kadar
sürdüğü, hangi fragment'lerin başarısız olduğu, render'ın bağlı olduğu dependency
tag'ler, `Cache-Control`'ü ve `ETag`'i, boyutu ve denetleyen plugin'lerin kaç
finding raporladığı.

```go
import "github.com/Elagoht/collage-devtoolbar"

Plugins: []collage.Plugin{
	htmlcheck.New(htmlcheck.Options{}),
	devtoolbar.New(), // last
},
```

- collage v0.24.0 ya da sonrasını gerektirir ve yapılandırılacak bir şeyi yoktur.
- **Onu en son register edin:** saydığı finding'ler, ondan önce çalışan
  plugin'lerinkidir.
- `DevMode` olmadan başlatılan bir sunucuda ve bir static build'de hiçbir şey
  yapmaz. Panel page cache'ten sonra eklenir, bu yüzden collage'ın cache'lediği
  içerik onu asla taşımaz.

#### elagoht/accesslog

[github.com/Elagoht/collage-accesslog](https://github.com/Elagoht/collage-accesslog),
her request için `slog` üzerinden yapılandırılmış bir log satırı yazar ve her
request'e handler'larının log'larken kullanabileceği bir id verir.

```go
import "github.com/Elagoht/collage-accesslog"

Plugins: []collage.Plugin{accesslog.New(accesslog.Options{})},
```

```json
{
  "elagoht/accesslog": {
    "skip": ["/_collage/", "/healthz", "/static/"],
    "sample": 0.25,
    "trustProxy": true,
    "requestIdHeader": "X-Request-ID"
  }
}
```

- collage v0.24.0 ya da sonrasını gerektirir.
- Satırda metot, query'siz path, status, byte sayısı, süre, istemci adresi, user
  agent, referer ve request id bulunur. Satır uygulamanın logger'ıyla ya da
  `Options.Logger` ile yazılır; bir `5xx`, `ERROR` seviyesinde log'lanır.
- Bir id'ye benzeyen `X-Request-ID` korunur, değilse yenisi üretilir. Id response'ta
  geri gönderilir ve `accesslog.RequestID(ctx)` ile okunur.
- `sample`, `2xx` response'ların bir kısmını log'lar, diğerlerinin hiçbirini
  atlamaz. `trustProxy`, istemcinin adresini yalnızca güvenilen proxy'lerden gelen
  `X-Forwarded-For`'dan okur.

#### elagoht/prometheus

[github.com/Elagoht/collage-prometheus](https://github.com/Elagoht/collage-prometheus),
framework'ün metric'lerini Prometheus'a aktarır ve `/metrics`'te sunar: render ve
fragment süreleri, cache olayları, route'a göre HTTP response'ları ve
invalidation'lar.

```go
import "github.com/Elagoht/collage-prometheus"

m := prometheus.NewMetrics(prometheus.Options{Namespace: "collage"})

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Metrics: m},
	Plugins:       []collage.Plugin{m},
})
```

```json
{
  "elagoht/prometheus": { "path": "/metrics", "token": "s3cret" }
}
```

- v0.2.0, collage v0.25.0 ya da sonrasını gerektirir; v0.1.1 v0.24.0'ı
  gerektiriyordu. Tek değeri hem uygulamanın `Metrics`'i hem de bir plugin olarak
  verin. Birincisi olmadan hiçbir şey ölçülmez, ikincisi
  olmadan hiçbir şey sunulmaz.
- Hiçbir label bir request'ten alınmaz. v0.2.0'dan beri `route`, `collage.RouteOf`'un
  request'in neye resolve edildiğini söylediği tür ve addır: `/blog/a` ve `/blog/b`
  ikisi de `page:post`, `/robots.txt` `document:robots`, bir mount
  `mount:/static/`, bir 404 ise `other` olur. Böylece path uyduran bir crawler yeni
  time series üretemez. `routes` seçeneği kaldırıldı: mount'lar ve handler'lar o
  olmadan da prefix'leriyle etiketlenir.
- `token` ayarlanırsa scrape onu bir bearer token olarak göndermelidir.
  `path: "-"` metric'leri hiçbir yerde sunmaz. Path birebir eşleşir: başka bir
  route'un zaten yanıtladığı ya da `/` ile biten bir path uygulamanın başlamasını
  engeller.

#### elagoht/otel

[github.com/Elagoht/collage-otel](https://github.com/Elagoht/collage-otel),
collage'ın kendi span'lerini OpenTelemetry span'lerine dönüştürür ve bir request'in
çağıranın başlattığı trace'i sürdürür.

```go
import otel "github.com/Elagoht/collage-otel"

t := otel.NewTracer(provider.Tracer("example.com/site"))

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Tracer: t},
	Plugins:       []collage.Plugin{t},
})
```

```json
{ "elagoht/otel": { "skip": ["/healthz"] } }
```

- v0.2.1, collage v0.26.0 ya da sonrasını gerektirir; v0.2.0 v0.25.0'ı, v0.1.0
  v0.23.0'ı gerektiriyordu. Tracer olarak `collage.http`, `collage.render` ve
  `collage.fragment`'i span'lere dönüştürür. Plugin olarak da çağıranın trace
  context'ini header'lardan okur; bunu collage'ın kendi span'inden önce çalışan bir
  `RequestHook` içinde yapar. İkisi de tek başına çalışır.
- İkisi birlikteyken bir request tek bir trace'tir. Çağıranınkinin child'ı olan bir
  server span vardır; `collage.http` onun, render ve her fragment de
  `collage.http`'nin altında yer alır. Server span, `collage.RouteInfo`'nun bildirdiği
  route'a göre adlandırılır: page, document ya da action için register edildiği
  pattern (`GET /blog/{slug}`, `GET /feed.xml`), mount ya da handler için prefix'i.
- SDK uygulamaya aittir: provider, exporter, sampler ve propagator. Bir propagator
  ayarlayın; yoksa her request kendi trace'ini başlatır.

#### elagoht/analytics

[github.com/Elagoht/collage-analytics](https://github.com/Elagoht/collage-analytics),
her page'in head'ine bir analytics snippet'i ekler: Plausible, Umami, GoatCounter ya
da Google Analytics 4. Bunu production'da ve static build'lerde yapar, development'ta
asla yapmaz.

```go
import "github.com/Elagoht/collage-analytics"

Plugins: []collage.Plugin{analytics.New(analytics.Options{
	Plausible: &analytics.Plausible{Domain: "example.com"},
})},
```

```json
{
  "elagoht/analytics": {
    "plausible": { "domain": "example.com" },
    "exclude": ["admin"],
    "respectDnt": true,
    "requireConsent": true
  }
}
```

- collage v0.23.0 ya da sonrasını gerektirir ve layout'ta `{{hoist "head"}}` ister.
- Plausible, Umami ve GoatCounter ziyaretleri cookie olmadan sayar. Google Analytics
  4 cookie ayarlar; onu `requireConsent` ile kullanın.
- `respectDnt` ya da `requireConsent` ile siteden sunulan küçük bir loader, herhangi
  bir sağlayıcının script'i çekilmeden önce tarayıcıda karar verir. Do Not Track ya
  da Global Privacy Control altında hiçbir şey yüklenmez. `requireConsent` ile de
  page `window.collageAnalyticsConsent()` çağırana kadar hiçbir şey yüklenmez.
- `exclude`, snippet almayan page'leri adlandırır.

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
bakın. Yayımlanmış plugin'lerden jsonld, meta, feed, favicon ve analytics head'e bu
yolla yazar.

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
