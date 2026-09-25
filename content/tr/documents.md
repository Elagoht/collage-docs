---
description: Sitemap'ler, feed'ler, robots.txt ve JSON gibi HTML yerine byte dönen route'lar, page'ler gibi cache'lenir ve invalidate edilir.
reference: DocumentBuilder.AtRoot, NewDocument, DocumentBuilder, Document, DocumentResult, DocumentPathProvider, PathInstance
---

# Document'lar: sitemap'ler, feed'ler, robots.txt

Bir sitenin sunduğu her şey page değildir. Crawler `/sitemap.xml` ve `/robots.txt`
ister, feed okuyucu `/feed.xml` ister, load balancer `/healthz` ister. collage'da
bunların her biri bir **document**'tir. Document, handler'ı byte ve bir content type
dönen bir route'tur. Template'i, layout'u ve fragment'i yoktur.

Bunun dışındaki her şeyi document page ile paylaşır. Aynı router'da yaşar. Bu yüzden
bir page ile çakışan bir path, ikisinden ikincisi register edilirken
`collage.ErrDuplicateRoute` ile reddedilir. Document aynı key altında cache'lenir,
aynı content hash tabanlı `ETag`'i ve `304` cevaplarını alır, aynı üç stratejiyi
kullanır. Dependency tag'ler taşır ve aynı `app.InvalidateTags` çağrısıyla cache'ten
düşer.

```go
robots := collage.NewDocument("robots", "text/plain; charset=utf-8").
	WithPath("en", "/robots.txt").
	WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
		return []byte("User-agent: *\nAllow: /\n"), nil, nil
	}).
	Static().
	Build()

if err := app.RegisterDocument(robots); err != nil {
	log.Fatal(err)
}
```

## Neden page değil

Page HTML'dir ve `html/template` üzerinden render edilir. Escape kuralları da HTML'in
kurallarıdır. Bu kurallar XML için de JSON için de yanlıştır. Bu kurallarla üretilen
bir feed, escape edilmeye en çok ihtiyaç duyan karakterlerde sessizce bozulur. Hiçbir
hata almazsınız, elinizde yalnızca crawler'ın reddettiği bir dosya kalır.

Bu yüzden document hiçbir şey render etmez. Body'yi formatı bilen bir encoder'la
(`encoding/xml`, `encoding/json`) üretin. Gerçekten bir template istiyorsanız
`text/template` kullanın. Framework sizin döndüğünüz şeyi byte'ı byte'ına sunar.

## Builder

| Çağrı | Ne yapar |
| --- | --- |
| `collage.NewDocument(name, contentType)` | Builder'ı başlatır. İki argüman da zorunludur. |
| `AtRoot(pattern)` | Document'a bütün locale'lerin dışından ulaşan URL pattern'i. Locale config'i ne olursa olsun prefix almaz. Sitenin kendi dosyaları içindir: `/robots.txt`, `/llms.txt`. v0.14.1'den beri vardır. |
| `WithPath(locale, pattern)` | Document'a `locale` içinde ulaşan URL pattern'i. `{param}` segment'leri page'lerdeki gibi çalışır. Placeholder bütün bir segment olmalıdır: `/feeds/{category}/rss.xml` yazılabilir, ama `/feeds/{category}.xml` register sırasında `collage.ErrInvalidPattern` ile reddedilir (v0.11.0'dan beri). |
| `WithHandler(fn)` | Body'yi üreten fonksiyon. Zorunludur, çünkü document'ın fallback olarak kullanabileceği bir template'i yoktur. |
| `Dynamic()` | Handler'ı her request'te çalıştırır. **Varsayılan budur.** |
| `Static()` | Handler'ı bir kez çalıştırır, bir tag invalidate edene kadar cache'ten sunar. |
| `Incremental(ttl)` | Cache'ten sunar, `ttl` dolduktan sonra handler'ı yeniden çalıştırır. |
| `WithCacheParams(names...)` | Page'lerde olduğu gibi, hangi query parametrelerinin cache key'ine girdiğini belirler. |
| `WithDependency(tags...)` | Bu document'ın her response'unun taşıdığı tag'ler. |
| `WithRedirect(from, to, status)` / `WithPermanentRedirect(from, to)` | Buraya redirect eden eski path'ler. |
| `Build()` / `BuildErr()` | Document'ı ve zincirin topladığı hataları verir. |

Varsayılana dikkat edin. Strateji vermeyi unuttuğunuz bir page yine de bir page'dir.
Strateji vermeyi unuttuğunuz bir document ise handler'ını her request'te çalıştırır
ve static export'ta atlanır. Sitemap'ler, feed'ler ve `robots.txt` neredeyse her
zaman `Static()` ya da `Incremental(ttl)` ister.

Handler set edilmemişse `Build`, `collage.ErrNoDocumentHandler` hatasını kaydeder.
Builder'ın kaydettiği hatalar document'ın üzerinde kalır. `BuildErr`'i çağırsanız da
çağırmasanız da `RegisterDocument` bu document'ı adını belirterek reddeder. Bu yüzden
hatayı kendiniz kontrol etmeniz isteğe bağlıdır.

### Content type

Content type, document'ı build ettiğinizde sabitlenir. Cache'ten gelen ya da yeni
üretilen her response'a olduğu gibi yazılır. Cache'lenen body ile birlikte saklanmaz,
byte'lardan da tahmin edilmez. Format charset gerektiriyorsa onu da ekleyin:
`text/plain; charset=utf-8`, `application/json`, `application/xml`,
`application/rss+xml`.

## Handler

```go
type DocumentHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (body []byte, tags []string, err error)
```

`rc`, data handler'ın aldığı `*collage.RenderContext`'in aynısıdır. Path parametresi
için `rc.Param("slug")`, çözümlenen locale için `rc.Locale`, request için
`rc.Request` kullanılır. Render edilen bir page olmadığı için `rc.Page` `nil`'dir.
Hoist edilecek bir head de yoktur.

Data handler'daki iki özellik burada da çalışır (v0.10.0'dan beri).
[`collage.Cached`](/docs/caching#caching-data-across-pages) data cache'i page'lerle
paylaşır. Böylece post page'lerinin zaten çektiği post'ları okuyan bir feed, bunların
hiçbirini yeniden çekmez. `collage.Cached`'a verilen tag'ler de document'ın tag'lerine
eklenir. `rc.Asset(path)` ise mount edilmiş bir dosyanın content-addressed URL'sini
verir. Örneğin bir web manifest'indeki ikon için bunu kullanabilirsiniz.

Döndüğünüz tag'ler, document'ın `WithDependency` tag'leriyle birleştirilir. Data
handler'da olduğu gibi, bir hata da dönseniz bu tag'ler okunur. Böylece neye bağlı
olduğunu öğrendikten sonra hata veren bir handler, kendisini neyin stale yapacağını
yine de bildirmiş olur.

Bilmeye değer üç kural var:

- **nil hatayla dönen boş bir body başarısızlık sayılır.** 500 ile cevaplanır ve
  `collage.ErrEmptyDocumentBody` olarak raporlanır, çünkü body'yi doldurmayı unutmuş
  bir handler'dan ayırt edilemez. Gerçekten boş olan bir document tek bir newline
  döner.
- **`collage.ErrNotFound`'u wrap eden bir hata 404 olur.** Diğer her hata 500 olur.
- **Panic recover edilir** ve sıradan bir hataya dönüşür. Böylece bozuk tek bir feed
  bütün process'i çökertmez.

Document'ın kendine ait bir timeout'u yoktur. Handler'ı `Config.Template.Timeout`
altında çalışır (varsayılanı beş saniyedir). Bu, fragment'ler için varsayılan data
handler timeout'u olan ayarın aynısıdır. Sınır size verilen `ctx`'e uygulanır, bu
yüzden onu çağırdığınız her şeye aktarın.

### Hatalar düz metindir

Hata veren bir document hiçbir zaman HTML bir error page ile cevap vermez.
`sitemap.xml` isteyen bir crawler'ın böyle bir sayfayla işi yoktur. Document'ın cevap
vermediği bir method için dönen `405` de buna dahildir (v0.11.0'dan beri düz metin).
Crawler, doğru status ile `text/plain` alır. Production'da tek satırlık genel bir
mesaj döner, development'ta ise route adı ve hata zincirinin tamamı döner.
Response'ta ayrıca `Cache-Control: no-store` bulunur.

Formata özgü bir hata istiyorsanız, örneğin bir JSON `{"error": "..."}`, hatayı
handler'ın içinde yakalayın ve o body'yi kendiniz dönün. Handler ne dönerse o sunulur.

## Caching

[Caching](/docs/caching) sayfasında anlatılan her şey burada da geçerlidir. Cache
key; path, locale, path parametreleri ve query string'den oluşur. Cache'ten yalnızca
`GET` ve `HEAD` request'leri sunulur. `Cache-Control` header'ı stratejiye göre
belirlenir.

Post'larınızdan üretilen bir sitemap, yeni bir post yayımlandığında cache'ten
düşürülmelidir. Tag'ler tam da bunun içindir:

```go
// Declared by the sitemap, the feed, the blog index and every post page.
if err := app.InvalidateTags(ctx, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Plugin'ler de document'ları görür. `collage.DocumentRenderedHook`'u implement eden
bir plugin, document'ın body'sini cache'lenip sunulmadan önce yeniden yazabilir.
Minifier buna bir örnektir. Page render hook'ları (`OnBeforeRender`, `OnAfterRender`)
çalışmaz, çünkü render edilen bir şey yoktur. Bkz.
[Plugin yazmak](/docs/writing-plugins).

## Örnek: sitemap

Sitemap mutlak URL'leri listeler. Bu URL'leri page adlarından `app.URL` ile üretin.
Böylece bir page'in path'i değiştiğinde sitemap de onu takip eder. URL'lerin başına
da sitenin origin'ini ekleyin.

```go
// origin is where the site is published. app.URL returns paths, and a sitemap
// needs absolute URLs.
const origin = "https://example.com"

type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func SitemapDocument(app *collage.App, posts *store.Posts) *collage.Document {
	return collage.NewDocument("sitemap", "application/xml").
		WithPath("en", "/sitemap.xml").
		WithHandler(func(ctx context.Context, _ *collage.RenderContext) ([]byte, []string, error) {
			set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}

			for _, name := range []string{"home", "about", "blog"} {
				path, err := app.URL(name, "", nil)
				if err != nil {
					return nil, nil, err
				}
				set.URLs = append(set.URLs, sitemapURL{Loc: origin + path})
			}

			list, err := posts.List(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("sitemap: list posts: %w", err)
			}
			for _, post := range list {
				path, err := app.URL("blog-post", "", map[string]string{"slug": post.Slug})
				if err != nil {
					return nil, nil, err
				}
				set.URLs = append(set.URLs, sitemapURL{
					Loc:     origin + path,
					LastMod: post.Updated.Format("2006-01-02"),
				})
			}

			body, err := xml.MarshalIndent(set, "", "  ")
			if err != nil {
				return nil, nil, fmt.Errorf("sitemap: marshal: %w", err)
			}
			return append([]byte(xml.Header), body...), []string{"blog:posts"}, nil
		}).
		Incremental(time.Hour).
		Build()
}
```

Handler `app`'i closure ile yakalar ve `app.URL`'i build edildiğinde değil,
çalıştığında çağırır. O ana kadar bütün page'ler register edilmiş olur. `app.URL`
katıdır: var olmayan bir page adı ya da pattern'i doldurmayan parametreler,
sitemap'inizde bozuk bir link olarak değil, hata olarak karşınıza çıkar. Birden çok
locale'i olan bir sitede `app.URL`'i her locale için bir kez çağırın ve locale'i
ikinci argüman olarak verin. Dönen sonuç locale prefix'ini içerir.

Strateji `Incremental(time.Hour)`'dur ve handler `blog:posts` tag'ini döner. Böylece
sitemap en az saatte bir kez yeniden üretilir. Bir şey bu tag'i invalidate ettiğinde
ise hemen yeniden üretilir.

## Örnek: RSS feed'i

```go
type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []item `xml:"item"`
}

type item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	GUID    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Summary string `xml:"description"`
}

func FeedDocument(app *collage.App, posts *store.Posts) *collage.Document {
	return collage.NewDocument("feed", "application/rss+xml").
		WithPath("en", "/feed.xml").
		WithHandler(func(ctx context.Context, _ *collage.RenderContext) ([]byte, []string, error) {
			latest, err := posts.Latest(ctx, 20)
			if err != nil {
				return nil, nil, fmt.Errorf("feed: %w", err)
			}

			feed := rss{Version: "2.0", Channel: channel{
				Title:       "Example blog",
				Link:        origin + "/",
				Description: "Posts from the example blog.",
			}}
			tags := []string{"blog:posts"}
			for _, post := range latest {
				path, err := app.URL("blog-post", "", map[string]string{"slug": post.Slug})
				if err != nil {
					return nil, nil, err
				}
				feed.Channel.Items = append(feed.Channel.Items, item{
					Title:   post.Title,
					Link:    origin + path,
					GUID:    origin + path,
					PubDate: post.Published.Format(time.RFC1123Z),
					Summary: post.Summary,
				})
				tags = append(tags, "post:"+post.Slug)
			}

			body, err := xml.MarshalIndent(feed, "", "  ")
			if err != nil {
				return nil, nil, fmt.Errorf("feed: marshal: %w", err)
			}
			return append([]byte(xml.Header), body...), tags, nil
		}).
		Static().
		Build()
}
```

`encoding/xml` başlıkları ve özetleri doğru şekilde escape eder. Page kullanmamanın
bütün nedeni de budur. Feed `Static()`'tir: yalnızca bir post değiştiğinde değişir.
Her post'un tag'i feed'in üzerindedir. Bu yüzden feed'deki herhangi bir post'u
düzenlemek feed'i cache'ten düşürür.

Layout'a bir `<link rel="alternate">` koyarak okuyucuları feed'e yönlendirin. Bkz.
[Head ve SEO](/docs/head-and-seo).

## Örnek: robots.txt

```go
func RobotsDocument(app *collage.App) *collage.Document {
	return collage.NewDocument("robots", "text/plain; charset=utf-8").
		AtRoot("/robots.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			sitemap, err := app.URL("sitemap", "", nil)
			if err != nil {
				return nil, nil, err
			}
			body := "User-agent: *\n" +
				"Disallow: /api/\n" +
				"\n" +
				"Sitemap: " + origin + sitemap + "\n"
			return []byte(body), nil, nil
		}).
		Static().
		Build()
}
```

`robots.txt` bir dile değil siteye aittir. Crawler'lar onu yalnızca kökte arar, başka
hiçbir yerde aramaz. Bu yüzden `AtRoot` ile build edilir. Varsayılan locale'i `/en/`
altında sunulan bir sitede bile tek adresi `/robots.txt`'dir.

`app.URL` page'ler için olduğu gibi document'lar için de çalışır. Bu yüzden
`robots.txt` sitemap'i adıyla bulur. Aynı adı taşıyan bir page ile bir document'a o
adla link verilemez, çünkü `app.URL` hangisini kastettiğinizi tahmin etmeyi reddeder.
Bu yüzden document'lara kendilerine özgü adlar verin.

## Birden çok locale'de document'lar

Document'ın path'leri, page'lerde olduğu gibi locale'e göre tutulur:

```go
collage.NewDocument("feed", "application/rss+xml").
	WithPath("en", "/feed.xml").
	WithPath("tr", "/feed.xml").
	WithHandler(feedHandler).
	Static().
	Build()
```

Pattern locale prefix'ini asla tekrar etmez. `tr` destekleniyorsa `/tr/feed.xml`
request'inin `/tr` kısmı routing'den önce çıkarılır. Request de `tr` kaydının
`/feed.xml` pattern'iyle eşleşir. `WithPath("tr", "/tr/feed.xml")` yazarsanız bu
path'e yalnızca `/tr/tr/feed.xml` ile ulaşılır. Handler `rc.Locale`'i okur. Locale
cache key'inin bir parçası olduğu için iki feed ayrı ayrı cache'lenir. Bkz.
[Link'ler ve locale'ler](/docs/links-and-locales).

Static export, her locale'in document'ını page'lerde yaptığı gibi onu sunan URL'ye
yazar. `en` feed'i `feed.xml` dosyasına, `tr` feed'i `tr/feed.xml` dosyasına yazılır.
Yani iki locale'deki tek bir pattern iki dosya demektir.

[`PrefixDefault`](/docs/links-and-locales#the-url-decides-the-locale) ile varsayılan
locale'in document'ı da prefix alır. Document `/en/feed.xml` adresinde sunulur,
`en/feed.xml` dosyasına yazılır ve `/feed.xml` oraya redirect eder. Her dil için ayrı
sitemap'i olan bir site bunların hepsini `robots.txt`'de listeler. `robots.txt`
istediği sayıda sitemap belirtebilir:

```text
Sitemap: https://example.com/en/sitemap.xml
Sitemap: https://example.com/tr/sitemap.xml
```

`AtRoot` ile build edilen bir document hiçbir locale'e ait değildir. Her config'te
prefix'siz path'inde durur ve her dildeki bir page'den adıyla ona link verilebilir.
`rc.Locale` varsayılan locale'e set edilmiş olarak render edilir.

## Static export'ta document'lar

[Static export](/docs/static-export), document'ı birebir kendi path'ine yazar.
`/sitemap.xml`, `dist/sitemap.xml/index.html` olarak değil `dist/sitemap.xml` olarak
yazılır, çünkü `/sitemap.xml` isteyen bir crawler'a dizin dönmemelidir. Varsayılan
dışındaki bir locale'in document'ı, sunulduğu yere, yani o locale'in prefix'i altına
yazılır: `dist/tr/sitemap.xml`.

- `Static()` ve `Incremental(ttl)` document'ları yazılır. `Dynamic()` olanlar atlanır
  ve raporda adlarıyla listelenir. Scaffold'daki `/healthz`'in `dist/` içinde hiç
  görünmemesinin nedeni budur.
- Boş body `collage.ErrEmptyDocumentBody` ile reddedilir ve hiçbir dosya yazılmaz.
- `{param}` içeren bir pattern'in somut path'lerini listelemek için
  `BuildOptions.DocumentPathProvider` gerekir. Provider yoksa document
  `collage.ErrDynamicPathUnresolved` ile atlanır.
- `WithCacheParams` kullanan bir document (örneğin sayfalanmış bir feed) query string
  olmadan yazılır, çünkü bir dosyanın query string'i olamaz. Rapor, page'lerde olduğu
  gibi bu durum için de uyarı verir.
- Aynı dosyaya çıkan iki path (örneğin aynı path'i iki kez dönen bir provider) bir kez
  build edilir. Diğerleri `collage.ErrDuplicateOutputPath` ile atlanır.

`DocumentPathProvider`, `PathProvider`'ın document'lar için olan karşılığıdır. Ayrı
bir interface olduğu için page'ler için yazılmış bir provider'ı değiştirmeniz
gerekmez:

```go
// categoryFeeds expands "/feeds/{category}/rss.xml" into one path per category.
type categoryFeeds struct{ categories []string }

func (p categoryFeeds) Paths(_ context.Context, doc *collage.Document, locale string) ([]collage.PathInstance, error) {
	if doc.Name != "category-feed" {
		return nil, nil
	}
	var paths []collage.PathInstance
	for _, category := range p.categories {
		paths = append(paths, collage.PathInstance{
			Path:   "/feeds/" + category + "/rss.xml",
			Params: map[string]string{"category": category},
		})
	}
	return paths, nil
}
```

Genişlettiği document, placeholder bütün bir segment olacak şekilde register edilir.
Export edilen dosyayı `feeds/go/rss.xml` yapan, placeholder'dan sonra gelen sabit
`rss.xml` kısmıdır:

```go
collage.NewDocument("category-feed", "application/rss+xml").
	WithPath("en", "/feeds/{category}/rss.xml").
	WithHandler(categoryFeedHandler).
	Static().
	Build()
```

```go
builder, err := collage.NewBuilder(app, collage.BuildOptions{
	OutDir:               outDir,
	PathProvider:         postPaths{store},
	DocumentPathProvider: categoryFeeds{categories},
})
```

`Params`, handler'ın `rc.Param` ile okuduğu değerlerdir. Canlı bir request'in
yakalayacağı değerlerle aynıdır.

## Document'ların yapmadıkları

- **Template, fragment ya da slot yok.** Document'ların varlık nedeni zaten budur.
- **`Range` request'leri yok.** Body bellekte oluşturulur ve bütün olarak sunulur.
  Ses, video ve büyük indirme dosyaları mount edilmiş bir dosya sisteminde durmalıdır.
  Bkz. [Static asset'ler](/docs/assets).
- **Büyük hiçbir şey yok.** Cache'lenebilen bir document page cache'te tutulur. Page
  cache byte ile değil entry sayısıyla sınırlıdır. Bu yüzden 50 MB'lık tek bir
  document, binlerce page kadar yer kaplar.

Document'ın ifade edemediği her şey için (streaming bir response, request body'si,
`GET` dışındaki method'lar) [Middleware ve kendi API'niz](/docs/middleware-and-apis)
sayfasına bakın.
