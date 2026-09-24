---
description: HTML yerine bayt dönen route'lar — sitemap'ler, feed'ler, robots.txt, JSON — sayfalar gibi önbelleğe alınır ve geçersiz kılınır.
---

# Document'ler: sitemap'ler, feed'ler, robots.txt

Bir sitenin sunduğu her şey sayfa değildir. Bir crawler `/sitemap.xml` ve
`/robots.txt` ister, bir feed okuyucu `/feed.xml` ister, bir yük dengeleyici
`/healthz` ister. collage'da bunların her biri bir **document**'tir: handler'ı bayt ve
bir içerik tipi dönen; şablonu, layout'u ve fragment'i olmayan bir route.

Bir document geri kalan her şeyi bir sayfayla paylaşır. Aynı router'da yaşar; bu
yüzden bir sayfayla çakışan bir yol, ikisinden ikincisi kaydedilirken
`collage.ErrDuplicateRoute` ile reddedilir. Aynı anahtar altında önbelleğe alınır,
aynı içerik hash'ine dayalı `ETag`'i ve `304` yanıtlarını alır, aynı üç stratejiyi
kullanır, bağımlılık etiketleri taşır ve aynı `app.InvalidateTags` çağrısıyla düşürülür.

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

## Neden sayfa değil

Bir sayfa HTML'dir ve `html/template` üzerinden render edilir. Kaçış kuralları
HTML'in kurallarıdır: XML için de JSON için de yanlıştır ve bunlarla üretilen bir feed,
tam da kaçışa en çok ihtiyaç duyan karakterlerde sessizce bozulur. Hata yoktur; yalnızca
bir crawler'ın reddettiği bir dosya vardır.

Bu yüzden bir document hiçbir şey render etmez. Gövdeyi formatı bilen bir encoder'la —
`encoding/xml`, `encoding/json` — ya da gerçekten bir şablon istiyorsanız
`text/template` ile üretin. Framework döndüğünüz şeyi bayt bayt sunar.

## Builder

| Çağrı | Ne yapar |
| --- | --- |
| `collage.NewDocument(name, contentType)` | Builder'ı başlatır. İkisi de zorunludur. |
| `WithPath(locale, pattern)` | `locale`'de document'e ulaşan URL pattern'i. `{param}` segmentleri sayfalardaki gibi çalışır. Bir yer tutucu bütün bir segmenttir: `/feeds/{category}/rss.xml` geçerlidir, `/feeds/{category}.xml` ise kayıt sırasında `collage.ErrInvalidPattern` ile reddedilir (v0.11.0'dan itibaren). |
| `WithHandler(fn)` | Gövdeyi üreten fonksiyon. Zorunludur: bir document'in geri düşebileceği bir şablonu yoktur. |
| `Dynamic()` | Handler'ı her istekte çalıştırır. **Varsayılan budur.** |
| `Static()` | Bir kez çalıştırır, bir etiket geçersiz kılana kadar önbellekten sunar. |
| `Incremental(ttl)` | Önbellekten sunar, `ttl` geçtikten sonra yeniden çalıştırır. |
| `WithCacheParams(names...)` | Bir sayfada olduğu gibi, hangi query parametrelerinin önbellek anahtarına katıldığı. |
| `WithDependency(tags...)` | Bu document'ten gelen her yanıtın taşıdığı etiketler. |
| `WithRedirect(from, to, status)` / `WithPermanentRedirect(from, to)` | Buraya yönlendiren eski yollar. |
| `Build()` / `BuildErr()` | Document ve zincirin topladığı hatalar. |

Varsayılana dikkat edin. Strateji vermeyi unuttuğunuz bir sayfa yine bir sayfadır;
strateji vermeyi unuttuğunuz bir document ise handler'ını her istekte çalıştırır ve
statik dışa aktarmada atlanır. Sitemap'ler, feed'ler ve `robots.txt` neredeyse her
zaman `Static()` ya da `Incremental(ttl)` ister.

Hiç handler ayarlanmamışsa `Build`, `collage.ErrNoDocumentHandler`'ı kaydeder.
Builder'ın kaydettiği şey document'in üzerinde kalır ve `BuildErr`'i çağırmış olun ya
da olmayın `RegisterDocument` onu adıyla reddeder; bu yüzden kendiniz denetlemeniz
isteğe bağlıdır.

### İçerik tipi

İçerik tipi document'i build ettiğinizde sabitlenir ve önbellekten ya da taze, her
yanıta olduğu gibi yazılır. Önbelleğe alınmış gövdeyle birlikte saklanmaz ve baytlardan
tahmin edilmez. Format gerektiriyorsa charset'i ekleyin:
`text/plain; charset=utf-8`, `application/json`, `application/xml`,
`application/rss+xml`.

## Handler

```go
type DocumentHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (body []byte, tags []string, err error)
```

`rc`, bir data handler'ın aldığı `*collage.RenderContext`'in aynısıdır: yol
parametresi için `rc.Param("slug")`, çözümlenmiş locale için `rc.Locale`, istek için
`rc.Request`. Render edilen bir sayfa olmadığı için `rc.Page` `nil`'dir ve hoist
edilecek bir head yoktur.

Bir data handler'ın sahip olduğu iki şey burada da çalışır (v0.10.0'dan itibaren).
[`collage.Cached`](/docs/caching#caching-data-across-pages) veri önbelleğini
sayfalarla paylaşır; böylece her yazı sayfasının zaten çektiği yazıları okuyan bir
feed, hiçbirini yeniden çekmez ve ona verilen etiketler document'inkilere katılır.
`rc.Asset(path)`, mount edilmiş bir dosyanın içerik adresli URL'sidir — örneğin bir
web manifest'indeki bir ikon.

Döndüğünüz etiketler document'in `WithDependency` etiketleriyle birleştirilir. Bir
data handler'da olduğu gibi, bir hata da dönseniz okunurlar; böylece neye bağlı
olduğunu öğrenip ardından başarısız olan bir handler, onu neyin bayatlatacağını yine
de söylemiş olur.

Bilmeye değer üç kural var:

- **nil hatayla boş bir gövde bir başarısızlıktır.** 500 ile yanıtlanır ve
  `collage.ErrEmptyDocumentBody` olarak bildirilir; çünkü gövdeyi doldurmayı unutmuş
  bir handler'dan ayırt edilemez. Gerçekten boş olan bir document tek bir satır sonu
  döner.
- **`collage.ErrNotFound`'u sarmalayan bir hata 404'tür.** Diğer her şey 500'dür.
- **Bir panic yakalanır** ve sıradan bir hataya dönüşür; böylece bozuk tek bir feed
  süreci çökertmez.

Bir document'in kendi zaman aşımı yoktur. Handler'ı `Config.Template.Timeout`
(varsayılan olarak beş saniye) altında çalışır — fragment'ler için varsayılan data
handler zaman aşımı olan ayarın aynısı. Sınır size verilen `ctx`'e uygulanır; bu
yüzden onu çağırdığınız her şeye aktarın.

### Hatalar düz metindir

Başarısız olan bir document hiçbir zaman bir HTML hata sayfasıyla yanıt vermez:
`sitemap.xml` isteyen bir crawler'ın onunla işi yoktur. Buna document'in yanıt
vermediği bir metot için dönen `405` de dahildir (v0.11.0'dan itibaren düz metin).
Doğru status ile `text/plain` alır; production'da tek satırlık genel bir mesaj,
geliştirmede route adı ve hata zincirinin tamamı; ayrıca `Cache-Control: no-store`.

Formata özgü bir hata istiyorsanız — bir JSON `{"error": "..."}` — hatayı handler'ın
içinde ele alın ve o gövdeyi kendiniz dönün. Handler ne dönerse o sunulur.

## Önbellekleme

[Önbellekleme](/docs/caching) sayfasındaki her şey geçerlidir. Önbellek anahtarı yol,
locale, yol parametreleri ve query string'dir; önbellekten yalnızca `GET` ve `HEAD`
sunulur; `Cache-Control` header'ı stratejiyi izler.

Yazılarınızdan kurulan bir sitemap, bir yazı yayımlandığında düşürülmelidir; etiketler
tam da bunun içindir:

```go
// Declared by the sitemap, the feed, the blog index and every post page.
if err := app.InvalidateTags(ctx, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Plugin'ler de document'leri görür: `collage.DocumentRenderedHook`'u gerçekleyen bir
plugin, bir document'in gövdesini önbelleğe alınıp sunulmadan önce yeniden yazabilir —
örneğin bir minifier. Hiçbir şey render edilmediği için sayfa render hook'ları
(`OnBeforeRender`, `OnAfterRender`) tetiklenmez. Bkz.
[Plugin yazmak](/docs/writing-plugins).

## Örnek: bir sitemap

Bir sitemap mutlak URL'leri listeler. Bunları sayfa adlarından `app.URL` ile kurun;
böylece bir sayfanın yolu değiştiğinde sitemap onu izler. Önlerine de sitenin
origin'ini koyun.

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

Handler `app`'i closure içine alır ve `app.URL`'i build edildiğinde değil, çalıştığında
çağırır — o sırada her sayfa kaydedilmiştir. `app.URL` katıdır: var olmayan bir sayfa
adı ya da pattern'i doldurmayan parametreler, sitemap'inizde bozuk bir bağlantı değil,
bir hatadır. Birden çok locale'i olan bir site için onu her locale için bir kez, locale'i
ikinci argüman olarak vererek çağırın; sonuç locale önekini içerir.

Strateji `Incremental(time.Hour)`'dur ve handler `blog:posts` etiketini döner; böylece
sitemap en az saatte bir ve bir şey o etiketi geçersiz kıldığında hemen yeniden
kurulur.

## Örnek: bir RSS feed'i

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

`encoding/xml` başlıkları ve özetleri doğru şekilde kaçışlar; sayfa kullanmamanın
bütün nedeni de budur. Feed `Static()`'tir: yalnızca bir yazı değiştiğinde değişir ve
her yazının etiketi üzerindedir; bu yüzden feed'deki herhangi bir yazıyı düzenlemek
onu düşürür.

Okuyucuları layout'tan bir `<link rel="alternate">` ile ona yönlendirin — bkz.
[Head ve SEO](/docs/head-and-seo).

## Örnek: robots.txt

```go
func RobotsDocument(app *collage.App) *collage.Document {
	return collage.NewDocument("robots", "text/plain; charset=utf-8").
		WithPath("en", "/robots.txt").
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

`app.URL` sayfalar kadar document'ler için de çalışır; bu yüzden `robots.txt`
sitemap'i adıyla bulur. Aynı adı paylaşan bir sayfaya ve bir document'e o adla
bağlantı verilemez — `app.URL` hangisini kastettiğinizi tahmin etmeyi reddeder — bu
yüzden document'lere kendilerine ait adlar verin.

## Birden çok locale'de document'ler

Bir document'in yolları, bir sayfanınkiler gibi locale'e göre anahtarlanır:

```go
collage.NewDocument("feed", "application/rss+xml").
	WithPath("en", "/feed.xml").
	WithPath("tr", "/feed.xml").
	WithHandler(feedHandler).
	Static().
	Build()
```

Pattern locale önekini asla tekrarlamaz. `tr` destekliyken `/tr/feed.xml` isteğinin
`/tr`'si routing'den önce çıkarılır ve istek `tr` kaydının `/feed.xml`'iyle eşleşir;
`WithPath("tr", "/tr/feed.xml")` yazarsanız ona yalnızca `/tr/tr/feed.xml` ile
ulaşılır. Handler `rc.Locale`'i okur ve locale önbellek anahtarının bir parçasıdır; bu
yüzden iki feed ayrı ayrı önbelleğe alınır. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales).

Statik dışa aktarma, her locale'in document'ini tıpkı bir sayfada olduğu gibi onu
sunan URL'ye yazar: `en` feed'ini `feed.xml`'e, `tr` feed'ini `tr/feed.xml`'e; yani
iki locale'deki tek bir pattern iki dosyadır.

## Statik dışa aktarmada document'ler

[Statik dışa aktarma](/docs/static-export) bir document'i kendi birebir yoluna yazar:
`/sitemap.xml`, `dist/sitemap.xml/index.html` değil `dist/sitemap.xml` olur; çünkü
`/sitemap.xml` isteyen bir crawler'a bir dizin gitmemelidir. Varsayılan dışında bir
locale'deki document, sunulduğu yer olan o locale'in öneki altına yazılır:
`dist/tr/sitemap.xml`.

- `Static()` ve `Incremental(ttl)` document'leri yazılır. `Dynamic()` olanlar atlanır
  ve raporda adlarıyla belirtilir; iskeletteki `/healthz`'in `dist/`'te hiç
  görünmemesinin nedeni budur.
- Boş bir gövde `collage.ErrEmptyDocumentBody` ile reddedilir ve hiçbir dosya
  yazılmaz.
- `{param}` içeren bir pattern'in somut yollarını listelemek için bir
  `BuildOptions.DocumentPathProvider` gerekir; yoksa `collage.ErrDynamicPathUnresolved`
  ile atlanır.
- `WithCacheParams` kullanan bir document — sayfalanmış bir feed — bir dosyanın query
  string'i olamayacağı için query string olmadan yazılır ve rapor, bir sayfada olduğu
  gibi, bununla ilgili uyarır.
- Tek bir dosyaya çözümlenen iki yol — bir yolu iki kez dönen bir provider — bir kez
  build edilir; geri kalanlar `collage.ErrDuplicateOutputPath` ile atlanır.

`DocumentPathProvider`, `PathProvider`'ın document'lerdeki karşılığıdır — ayrı bir
interface'tir, böylece sayfalar için yazılmış bir provider'ın değişmesi gerekmez:

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

Genişlettiği document, yer tutucu bütün bir segment olacak şekilde kaydedilir —
ardındaki birebir `rss.xml`, dışa aktarılan dosyayı `feeds/go/rss.xml` yapan şeydir:

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

`Params`, handler'ın `rc.Param` üzerinden okuduğu şeydir; canlı bir isteğin yakalayacağı
değerlerin aynısı.

## Document'lerin yapmadıkları

- **Şablon, fragment ya da slot yok.** Varlık nedenleri de bu.
- **`Range` istekleri yok.** Gövde bellekte kurulur ve bütün olarak sunulur. Ses,
  video ve büyük indirmeler mount edilmiş bir dosya sistemine aittir — bkz.
  [Statik dosyalar](/docs/assets).
- **Büyük hiçbir şey yok.** Önbelleğe alınabilen bir document sayfa önbelleğinde
  tutulur; bu önbellek bayt sayısıyla değil kayıt sayısıyla sınırlanır, bu yüzden 50 MB'lık
  tek bir document binlerce sayfa kadar yer kaplar.

Bir document'in ifade edemediği her şey için — akış hâlinde bir yanıt, bir istek
gövdesi, `GET` dışındaki metotlar — bkz.
[Middleware ve kendi API'niz](/docs/middleware-and-apis).
