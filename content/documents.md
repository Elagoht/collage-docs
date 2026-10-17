---
description: Routes that answer with bytes instead of HTML — sitemaps, feeds, robots.txt, JSON — cached and invalidated like pages.
reference: DocumentBuilder.AtRoot, DocumentBuilder.WithBody, DocumentBuilder.WithTimeout, DocumentBuilder.WithStaticParams, NewDocument, DocumentBuilder, Document, DocumentResult, StaticParamsFunc, ErrConflictingData
---

# Documents: sitemaps, feeds, robots.txt

Not everything a site serves is a page. A crawler wants `/sitemap.xml` and
`/robots.txt`, a feed reader wants `/feed.xml`, a load balancer wants `/healthz`.
In collage each of these is a **document**: a route that answers with bytes and a
content type — produced by a handler, or fixed when the program starts — with no
template, no layout and no fragments.

A document shares everything else with a page:

- **The same router.** A path that collides with a page is refused when the second
  of the two is registered, with `collage.ErrDuplicateRoute`.
- **The same cache.** It is cached under the same key, and gets the same
  content-hash `ETag` and `304` answers.
- **The same three strategies.**
- **The same tags.** It carries dependency tags, and is invalidated by the same
  `app.InvalidateTags` call.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 202" role="img" aria-labelledby="doc-shared-title">
<title id="doc-shared-title">A page and a document compared. A page's body comes from layouts and fragments rendered through html/template; a document's comes from a handler returning bytes, or from a fixed body. What they have in common: one router, where a clash is ErrDuplicateRoute; the same cache key, ETag and 304, and three strategies; dependency tags and app.InvalidateTags. A page answers text/html; a document answers with its own content type.</title>
<text x="80" y="16" text-anchor="middle">page</text>
<text x="240" y="16" text-anchor="middle">document</text>
<rect class="d-frame" x="8" y="24" width="144" height="44" rx="6" /><text x="80" y="42" text-anchor="middle">layouts, fragments</text><text class="d-small" x="80" y="58" text-anchor="middle">html/template</text>
<rect class="d-a" x="168" y="24" width="144" height="44" rx="6" /><text class="d-on-a" x="240" y="42" text-anchor="middle">handler → bytes</text><text class="d-on-a d-small" x="240" y="58" text-anchor="middle">or a fixed body</text>
<line class="d-rule" x1="80" y1="68" x2="80" y2="92" />
<line class="d-rule" x1="240" y1="68" x2="240" y2="92" />
<rect class="d-soft-a" x="8" y="92" width="304" height="56" rx="6" />
<text class="d-small" x="160" y="110" text-anchor="middle">one router: a clash is ErrDuplicateRoute</text>
<text class="d-small" x="160" y="124" text-anchor="middle">cache key · ETag and 304 · three strategies</text>
<text class="d-small" x="160" y="138" text-anchor="middle">dependency tags · app.InvalidateTags</text>
<line class="d-rule" x1="80" y1="148" x2="80" y2="170" />
<line class="d-rule" x1="240" y1="148" x2="240" y2="170" />
<rect class="d-pill" x="20" y="170" width="120" height="24" rx="12" /><text x="80" y="186" text-anchor="middle">text/html</text>
<rect class="d-pill" x="170" y="170" width="140" height="24" rx="12" /><text x="240" y="186" text-anchor="middle">its content type</text>
</svg>
<figcaption>What a page and a document have in common, and where they differ most. A page's body is rendered from layouts and fragments through <code>html/template</code>; a document's is the bytes its handler returns, or a body fixed when the program starts. They share one router, so a clashing path is <code>collage.ErrDuplicateRoute</code>; the same cache key, <code>ETag</code> and <code>304</code> answers and three strategies; and the same dependency tags, invalidated by <code>app.InvalidateTags</code>. A page answers <code>text/html</code>; a document answers with the content type it was built with, written on every response rather than stored with the cached body.</figcaption>
</figure>

```go
robots := collage.NewDocument("robots", "text/plain; charset=utf-8").
	AtRoot("/robots.txt").
	WithBody([]byte("User-agent: *\nAllow: /\n")).
	Build()

if err := app.RegisterDocument(robots); err != nil {
	log.Fatal(err)
}
```

## Why not a page

A page is HTML, and it renders through `html/template`. Its escaping rules are
HTML's: they are wrong for XML and wrong for JSON, and a feed produced through them
is silently malformed at exactly the characters that most need escaping. There is
no error, just a file a crawler rejects.

So a document renders nothing. Produce the body with an encoder that knows the
format — `encoding/xml`, `encoding/json` — or with `text/template` if you really
want a template. The framework serves what you return, byte for byte.

## The builder

| Call | What it does |
| --- | --- |
| `collage.NewDocument(name, contentType)` | Starts the builder. Both are required. |
| `AtRoot(pattern)` | The URL pattern that reaches the document outside every locale: no prefix, whatever the locale configuration. For the site's own files — `/robots.txt`, `/llms.txt`. Since v0.14.1. |
| `WithPath(locale, pattern)` | The URL pattern that reaches the document in `locale`. Placeholders work as they do for pages, text around them included (since v0.41.0): `/feeds/{category}.xml`, or `/blog/{slug}.md` beside a page at `/blog/{slug}` — see [text around a placeholder](/docs/pages-and-layouts#text-around-a-placeholder). |
| `WithHandler(fn)` | The function that produces the body. |
| `WithBody(b)` | A body fixed when the program starts, in place of a handler. A document has no template to fall back on, so it needs one of the two, and not both. Since v0.16.0. |
| `Dynamic()` | Run the handler on every request. What a document with a handler and no strategy resolves to. |
| `Static()` | Run once, serve from cache until a tag invalidates it. |
| `Incremental(ttl)` | Serve from cache, run again once `ttl` has passed. |
| `WithCacheParams(names...)` | Which query parameters take part in the cache key, as for a page. |
| `WithStaticParams(fn)` | The placeholder values a static build writes a `{param}` pattern for; see [below](#documents-in-a-static-export). Since v0.16.0. |
| `WithDependencyTags(tags...)` | Tags every response from this document carries. |
| `WithTimeout(d)` | Bounds the handler; zero, the default, means `Config.Template.Timeout` (since v0.61.0). |
| `WithRedirect(from, to, status)` / `WithPermanentRedirect(from, to)` | Old paths that redirect here. |
| `Build()` / `BuildErr()` | The document, and the errors the chain collected. |

`Build` records `collage.ErrNoDocumentHandler` when neither a handler nor a body
was set, and `RegisterDocument` refuses a document with both as
`collage.ErrConflictingData`. What the builder recorded stays on the document, and
`RegisterDocument` refuses it by name whether or not you called `BuildErr`, so
checking it yourself is optional.

### A fixed body

A document that declares no strategy is resolved when it is registered, as a
[page](/docs/caching#a-page-that-declares-none) is: static with a fixed body,
dynamic with a handler. A body fixed when the program starts is the same for every
request, so the `robots.txt` above is cached and exported without a `Static()`.

A sitemap or a feed is produced by a handler, so it is dynamic until it says
otherwise — which is why the examples below still say `Static()` or
`Incremental(ttl)`.

The framework cannot see inside a handler: a sitemap built from your posts and a
health check reporting the process's state are, from outside, the same function
returning bytes.

- **Guessing static for both** would cache the health check, and a health check
  served from a cache answers `ok` long after it stopped being true.
- **Guessing dynamic** costs the sitemap a render per request until you say what
  it is — a slower sitemap, not a wrong answer.

A document with a handler that you forget to give a strategy runs on every request
and is skipped by a static export.

### The content type

The content type is fixed when you build the document and written verbatim on every
response, cached or fresh. It is not stored with the cached body, and it is not
guessed from the bytes. Include the charset when the format needs one:
`text/plain; charset=utf-8`, `application/json`, `application/xml`,
`application/rss+xml`.

## The handler

```go
type DocumentHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (body []byte, tags []string, err error)
```

`rc` is the same `*collage.RenderContext` a data handler gets: `rc.Param("slug")`
for a path parameter, `rc.Locale` for the resolved locale, `rc.Request` for the
request. `rc.Page` is `nil`, because no page is being rendered, and there is no
head to hoist into.

Two things a data handler has work here too (since v0.10.0).
[`collage.Cached`](/docs/caching#caching-data-across-pages) shares the data cache
with the pages, so a feed reading the posts every post page already fetched fetches
none of them again, and the tags it was given join the document's.
`rc.Asset(path)` is a mounted file's content-addressed URL — an icon in a web
manifest, say.

The tags you return are merged with the document's `WithDependencyTags` tags. As with
a data handler, they are read even when you also return an error, so a handler that
learned what it depends on and then failed has still said what would make it stale.

Three rules are worth knowing:

- **An empty body with a nil error is a failure.** It is answered with a 500 and
  reported as `collage.ErrEmptyDocumentBody`, because it is indistinguishable from
  a handler that forgot to fill the body in. A document that is genuinely empty
  returns a single newline.
- **An error wrapping `collage.ErrNotFound` is a 404.** Anything else is a 500.
- **A panic is recovered** and becomes an ordinary error, so one broken feed does
  not take the process down.

A document's handler runs under its own timeout when it sets one with
`WithTimeout(d)` (since v0.61.0), and under `Config.Template.Timeout` (five seconds
by default) — the same setting that is the default data-handler timeout for
fragments — when it does not.

A sitemap that walks a large catalogue raises its own budget rather than every
fragment's and every feed's:

```go
collage.NewDocument("sitemap", "application/xml").
	AtRoot("/sitemap.xml").
	WithHandler(sitemap).
	WithTimeout(30 * time.Second).
	Build()
```

The timeout applies on a live request and in a static build alike. Zero means the
default; a negative one records `collage.ErrInvalidTimeout`. The limit applies to
the `ctx` you are given, so pass it on to whatever you call.

### Errors are plain text

A failing document never answers with an HTML error page: a crawler that asked for
`sitemap.xml` has no use for one. That includes a `405` for a method the document
does not answer (plain text since v0.11.0).

It gets `text/plain` with the right status and `Cache-Control: no-store`. The body
depends on the mode:

| Mode | Body |
| --- | --- |
| Production | One generic line |
| Development | The route name with the full error chain |

If you want a format-specific error — a JSON `{"error": "..."}` — handle the error
inside the handler and return that body yourself. Whatever the handler returns is
served.

## Caching

Everything in [Caching](/docs/caching) applies. The cache key is the path, the
locale, the path parameters and the query string; only `GET` and `HEAD` are served
from cache; the `Cache-Control` header follows the strategy.

A sitemap built from your posts should be invalidated when a post is published,
which is what tags are for:

```go
// Declared by the sitemap, the feed, the blog index and every post page.
if err := app.InvalidateTags(ctx, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Plugins see documents too: a plugin implementing `collage.DocumentRenderedHook` can
rewrite a document's body before it is cached and served — a minifier, say. The
page render hooks (`OnBeforeRender`, `OnAfterRender`) do not fire, since nothing is
rendered. See [Writing a plugin](/docs/writing-plugins).

An invalidated document is rendered again in the background right away (since
v0.64.0; see [Caching](/docs/caching#warming-after-an-invalidation)), or on its
next request if that has not happened yet.

Since v0.63.0 it is kept until then, as stale. When that render fails, it is served
the way a stale page is: with `Cache-Control: no-store` and
`X-Collage-Cache: STALE`. A render that answers 404 drops it. See
[Caching](/docs/caching#stale-pages-and-x-collage-cache).

## Example: a sitemap

A sitemap lists absolute URLs. Build them from page names with `app.URL`, so the
sitemap follows a page when its path changes, and put the site's origin in front.

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

The handler closes over `app`, and calls `app.URL` when it runs, not when it is
built — by then every page is registered. `app.URL` is strict: a page name that does
not exist, or parameters that do not fill the pattern, is an error rather than a
broken link in your sitemap. For a site in several locales, call it once per locale
with the locale as its second argument; the result includes the locale prefix.

The strategy is `Incremental(time.Hour)` and the handler returns the `blog:posts`
tag, so the sitemap is rebuilt at least hourly and immediately when something
invalidates that tag.

## Example: an RSS feed

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

`encoding/xml` escapes titles and summaries correctly, which is the whole reason for
not using a page. The feed is `Static()`: it changes only when a post does, and each
post's tag is on it, so editing any post in the feed invalidates it.

Point readers at it from the layout, with a `<link rel="alternate">` — see
[Head and SEO](/docs/head-and-seo).

## Example: robots.txt

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

`robots.txt` is the site's rather than a language's, and crawlers look for it at
the root and nowhere else, so it is built with `AtRoot`: its one address is
`/robots.txt`, even on a site whose default locale is served under `/en/`.

This one needs a handler, not `WithBody`, because it names the sitemap's URL, and
`app.URL` answers only once the sitemap is registered — when the handler runs, not
when the document is built. With a handler it is dynamic unless it says `Static()`.

`app.URL` works for documents as well as pages, so `robots.txt` finds the sitemap by
its name. A page and a document that share a name cannot be linked to by that name
— `app.URL` refuses to guess which one you meant — so give documents names of their
own.

## Documents in several locales

A document's paths are keyed by locale, like a page's:

```go
collage.NewDocument("feed", "application/rss+xml").
	WithPath("en", "/feed.xml").
	WithPath("tr", "/feed.xml").
	WithHandler(feedHandler).
	Static().
	Build()
```

The pattern never repeats the locale prefix. With `tr` supported, the request
`/tr/feed.xml` has its `/tr` stripped before routing and matches the `tr` entry's
`/feed.xml`; writing `WithPath("tr", "/tr/feed.xml")` would only be reached by
`/tr/tr/feed.xml`. The handler reads `rc.Locale`, and the locale is part of the
cache key, so the two feeds are cached separately. See
[Links and locales](/docs/links-and-locales).

A static export writes each locale's document at the URL that serves it, as it does
a page: the `en` feed to `feed.xml` and the `tr` feed to `tr/feed.xml`, so one
pattern in two locales is two files.

With [`PrefixDefault`](/docs/links-and-locales#the-url-decides-the-locale) the
default locale's document is prefixed too — `/en/feed.xml`, written to
`en/feed.xml` — and `/feed.xml` redirects there. A site with one sitemap per
language lists each in `robots.txt`, which may name any number:

```text
Sitemap: https://example.com/en/sitemap.xml
Sitemap: https://example.com/tr/sitemap.xml
```

A document built with `AtRoot` is in no locale: it is at its bare path in every
configuration, reachable by name from a page in any language, and rendered with
`rc.Locale` set to the default.

## Documents in a static export

[Static export](/docs/static-export) writes a document to its literal path:
`/sitemap.xml` becomes `dist/sitemap.xml`, not `dist/sitemap.xml/index.html`, because
a crawler asking for `/sitemap.xml` must not receive a directory. A document in a
locale other than the default is written under that locale's prefix, where it is
served: `dist/tr/sitemap.xml`.

- Static and incremental documents are written: those declared `Static()` or
  `Incremental(ttl)`, and those with a fixed body and no strategy. Dynamic ones —
  declared `Dynamic()`, or with a handler and no strategy — are skipped and named
  in the report, which is why the scaffold's `/healthz` never appears in `dist/`.
- An empty body is refused with `collage.ErrEmptyDocumentBody`, and no file is
  written.
- A pattern with a `{param}` needs `WithStaticParams` to list its values, or it
  is skipped with `collage.ErrDynamicPathUnresolved`.
- A document with `WithCacheParams` — a paginated feed — is written without a
  query string, since a file cannot have one, and the report warns about it, as it
  does for a page.
- Two tasks that resolve to one file — `WithStaticParams` listing one set of values
  twice, or a document at a page's `index.html` — stop the build with
  `collage.ErrOutputPathCollision` before anything is written (since v0.60.0; before,
  the rest were skipped with `ErrDuplicateOutputPath`). Documents' `WithStaticParams`
  run before any page renders, so this is known up front.

`WithStaticParams` works as it does for a
[page](/docs/static-export#dynamic-paths-withstaticparams): one map of placeholder
values per file, per locale the document has a path in. The placeholder is a whole
segment — the literal `rss.xml` after it is what makes the exported file
`feeds/go/rss.xml`:

```go
collage.NewDocument("category-feed", "application/rss+xml").
	WithPath("en", "/feeds/{category}/rss.xml").
	WithHandler(categoryFeedHandler).
	Static().
	WithStaticParams(func(ctx context.Context, locale string) ([]map[string]string, error) {
		params := make([]map[string]string, 0, len(categories))
		for _, category := range categories {
			params = append(params, map[string]string{"category": category})
		}
		return params, nil
	}).
	Build()
```

The build makes each path from the pattern, and the handler reads the values
through `rc.Param`, the same as a live request would have captured. A map that
does not fill the pattern exactly fails that one file with
`collage.ErrRouteParams`. The document has a handler, so it says `Static()` to be
written at all.

## Example: a catalogue's product feed and sitemaps

A shop sells in English and Turkish. Its product page is the `product` page from
the [bilingual catalogue](/docs/links-and-locales#example-a-bilingual-product-catalogue),
with the patterns `/products/{id}` and `/urunler/{id}`. Beside it the shop serves
three documents:

- **A JSON product feed** per language, for partners who list the shop's products.
- **A sitemap** per language.
- **One `robots.txt`**, which belongs to no language and names both sitemaps.

This shop gives English a prefix too, with `PrefixDefault`, and turns the cache on.
Caching is off by default, and with it off every request runs the handler, whatever
the strategy:

```go
cfg := &collage.Config{
	// Server and Template as before.
	Cache: collage.CacheConfig{Enabled: true},
	Locale: collage.LocaleConfig{
		Default:       "en",
		Supported:     []string{"en", "tr"},
		PrefixDefault: true,
	},
}
```

The `catalog` package stands for your own data layer, as on
[Links and locales](/docs/links-and-locales): `catalog.List` returns the products
in a locale, each with an `ID`, a `Name`, a `Price` and an `Updated` time. `origin`
is the constant from the sitemap example above.

The product feed has a path in each locale and reads the render's locale, so each
language gets its own feed, cached separately:

```go
type feedItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
}

func ProductFeedDocument(app *collage.App) *collage.Document {
	return collage.NewDocument("product-feed", "application/json").
		WithPath("en", "/products.json").
		WithPath("tr", "/urunler.json").
		WithHandler(func(ctx context.Context, rc *collage.RenderContext) ([]byte, []string, error) {
			list, err := catalog.List(ctx, rc.Locale)
			if err != nil {
				return nil, nil, fmt.Errorf("product feed: %w", err)
			}

			items := make([]feedItem, 0, len(list))
			tags := []string{"catalog"}
			for _, p := range list {
				path, err := app.URL("product", rc.Locale, map[string]string{"id": p.ID})
				if err != nil {
					return nil, nil, err
				}
				items = append(items, feedItem{ID: p.ID, Name: p.Name, Price: p.Price, URL: origin + path})
				tags = append(tags, "product:"+p.ID)
			}

			body, err := json.Marshal(items)
			if err != nil {
				return nil, nil, fmt.Errorf("product feed: marshal: %w", err)
			}
			return body, tags, nil
		}).
		Static().
		Build()
}
```

`encoding/json` escapes a product name with a quote or a `<` in it, which a page's
HTML escaping would have got wrong. The sitemap is built the same way, with the
`urlSet` and `sitemapURL` types from the sitemap example, and has one pattern in
both locales:

```go
func CatalogueSitemapDocument(app *collage.App) *collage.Document {
	return collage.NewDocument("sitemap", "application/xml").
		WithPath("en", "/sitemap.xml").
		WithPath("tr", "/sitemap.xml").
		WithHandler(func(ctx context.Context, rc *collage.RenderContext) ([]byte, []string, error) {
			list, err := catalog.List(ctx, rc.Locale)
			if err != nil {
				return nil, nil, fmt.Errorf("sitemap: %w", err)
			}

			set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
			tags := []string{"catalog"}
			for _, p := range list {
				path, err := app.URL("product", rc.Locale, map[string]string{"id": p.ID})
				if err != nil {
					return nil, nil, err
				}
				set.URLs = append(set.URLs, sitemapURL{
					Loc:     origin + path,
					LastMod: p.Updated.Format("2006-01-02"),
				})
				tags = append(tags, "product:"+p.ID)
			}

			body, err := xml.MarshalIndent(set, "", "  ")
			if err != nil {
				return nil, nil, fmt.Errorf("sitemap: marshal: %w", err)
			}
			return append([]byte(xml.Header), body...), tags, nil
		}).
		Static().
		Build()
}
```

`robots.txt` is at the root, and asks `app.URL` for the sitemap in each locale. It
returns no tags, since nothing in the catalogue changes it:

```go
func ShopRobotsDocument(app *collage.App) *collage.Document {
	return collage.NewDocument("robots", "text/plain; charset=utf-8").
		AtRoot("/robots.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			body := "User-agent: *\nDisallow: /api/\n\n"
			for _, locale := range []string{"en", "tr"} {
				sitemap, err := app.URL("sitemap", locale, nil)
				if err != nil {
					return nil, nil, err
				}
				body += "Sitemap: " + origin + sitemap + "\n"
			}
			return []byte(body), nil, nil
		}).
		Static().
		Build()
}
```

All three are registered together:

```go
for _, doc := range []*collage.Document{
	ProductFeedDocument(app),
	CatalogueSitemapDocument(app),
	ShopRobotsDocument(app),
} {
	if err := app.RegisterDocument(doc); err != nil {
		return err
	}
}
```

With `PrefixDefault`, the localised documents are under `/en/` and `/tr/`, and
`robots.txt` stays at its one bare address:

| Document | Served at | Written by a static export to |
| --- | --- | --- |
| `product-feed`, `en` | `/en/products.json` | `en/products.json` |
| `product-feed`, `tr` | `/tr/urunler.json` | `tr/urunler.json` |
| `sitemap`, `en` | `/en/sitemap.xml` | `en/sitemap.xml` |
| `sitemap`, `tr` | `/tr/sitemap.xml` | `tr/sitemap.xml` |
| `robots` | `/robots.txt` | `robots.txt` |

The other spellings are handled for you:

- `/products.json` and `/sitemap.xml` redirect, `301`, to their `/en/` addresses.
- `/en/robots.txt` redirects to `/robots.txt`.
- `/tr/robots.txt` is a 404, since a root document has no Turkish path.

`robots.txt` names both sitemaps:

```text
User-agent: *
Disallow: /api/

Sitemap: https://example.com/en/sitemap.xml
Sitemap: https://example.com/tr/sitemap.xml
```

### Changing a product

The feed and the sitemap carry two kinds of tag, for two kinds of change:

| Change | Tag to invalidate | Why |
| --- | --- | --- |
| A listed product is edited: its name, its price | `product:<id>` | Each listed product's tag is on every feed and sitemap that lists it |
| A product is added or removed | `catalog` | A new product's tag is on no cached list yet; `catalog` is on all of them |

```go
// After an edit to product 42.
if err := app.InvalidateTags(ctx, "product:42"); err != nil {
	log.Printf("invalidate: %v", err)
}

// After a product is added or removed.
if err := app.InvalidateTags(ctx, "catalog"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 226" role="img" aria-labelledby="doc-invalidate-title">
<title id="doc-invalidate-title">InvalidateTags with the tag product:42 reaches four cached documents: in en, /en/products.json and /en/sitemap.xml; in tr, /tr/urunler.json and /tr/sitemap.xml. Each is marked stale, and rendered again in the background or on its next request. /robots.txt, at the root and carrying no tags, is not touched.</title>
<defs><marker id="doc-invalidate-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="50" y="8" width="220" height="26" rx="13" /><text class="d-cmd-text" x="160" y="25" text-anchor="middle">InvalidateTags "product:42"</text>
<path class="d-flow" d="M120 34C120 44 82 44 82 52" marker-end="url(#doc-invalidate-arrow)" />
<path class="d-flow" d="M200 34C200 44 238 44 238 52" marker-end="url(#doc-invalidate-arrow)" />
<rect class="d-frame" x="8" y="56" width="148" height="96" rx="6" /><text x="82" y="72" text-anchor="middle">en</text>
<rect class="d-b" x="18" y="80" width="128" height="26" rx="4" /><text class="d-on-b" x="82" y="97" text-anchor="middle">/en/products.json</text>
<rect class="d-b" x="18" y="114" width="128" height="26" rx="4" /><text class="d-on-b" x="82" y="131" text-anchor="middle">/en/sitemap.xml</text>
<rect class="d-frame" x="164" y="56" width="148" height="96" rx="6" /><text x="238" y="72" text-anchor="middle">tr</text>
<rect class="d-b" x="174" y="80" width="128" height="26" rx="4" /><text class="d-on-b" x="238" y="97" text-anchor="middle">/tr/urunler.json</text>
<rect class="d-b" x="174" y="114" width="128" height="26" rx="4" /><text class="d-on-b" x="238" y="131" text-anchor="middle">/tr/sitemap.xml</text>
<text class="d-small" x="160" y="170" text-anchor="middle">stale · warmed, or rendered on next request</text>
<rect class="d-frame" x="8" y="184" width="304" height="34" rx="6" /><text x="160" y="198" text-anchor="middle">/robots.txt</text><text class="d-small" x="160" y="211" text-anchor="middle">AtRoot · no tags · not touched</text>
</svg>
<figcaption>Invalidating <code>product:42</code> reaches every cached document that listed product 42: the English and Turkish product feeds and the English and Turkish sitemaps, each cached on its own because the locale is part of the key. Each is kept as stale, and rendered again in the background or on its next request. <code>robots.txt</code>, at the root and with no tags, stays cached as it is.</figcaption>
</figure>

The documents a visitor has read are warmed, as pages are; see
[Caching](/docs/caching#warming-after-an-invalidation). Each tag
reaches cached pages too, the ones whose handlers returned it: `product:<id>`
also reaches that product's page, and `catalog` also reaches a catalogue page whose
handler returns `catalog`, as the one on
[Links and locales](/docs/links-and-locales#example-a-bilingual-product-catalogue)
does, whenever it is cached. Changing one product does not refresh the
catalogue listing; invalidate `catalog` for that.

## What documents do not do

- **No templates, fragments or slots.** That is the point of them.
- **No `Range` requests.** The body is built in memory and served whole. Audio,
  video and large downloads belong in a mounted file system — see
  [Static assets](/docs/assets).
- **Nothing large.** A cacheable document is kept in the page cache, which is
  bounded by entry count rather than bytes, so one 50 MB document costs as much
  room as thousands of pages.

For anything a document cannot express — a streaming response, a request body,
methods other than `GET` — see [Middleware and your own API](/docs/middleware-and-apis).
