---
description: Routes that answer with bytes instead of HTML — sitemaps, feeds, robots.txt, JSON — cached and invalidated like pages.
reference: DocumentBuilder.AtRoot, NewDocument, DocumentBuilder, Document, DocumentResult, DocumentPathProvider, PathInstance
---

# Documents: sitemaps, feeds, robots.txt

Not everything a site serves is a page. A crawler wants `/sitemap.xml` and
`/robots.txt`, a feed reader wants `/feed.xml`, a load balancer wants `/healthz`.
In collage each of these is a **document**: a route whose handler returns bytes and
a content type, with no template, no layout and no fragments.

A document shares everything else with a page. It lives in the same router, so a
path that collides with a page is refused when the second of the two is registered,
with `collage.ErrDuplicateRoute`. It is cached under the same key,
gets the same content-hash `ETag` and `304` answers, uses the same three strategies,
carries dependency tags, and is dropped by the same `app.InvalidateTags` call.

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
| `WithPath(locale, pattern)` | The URL pattern that reaches the document in `locale`. `{param}` segments work as they do for pages. A placeholder is a whole segment: `/feeds/{category}/rss.xml`, not `/feeds/{category}.xml`, which is refused at registration with `collage.ErrInvalidPattern` (since v0.11.0). |
| `WithHandler(fn)` | The function that produces the body. Required: a document has no template to fall back on. |
| `Dynamic()` | Run the handler on every request. **The default.** |
| `Static()` | Run once, serve from cache until a tag invalidates it. |
| `Incremental(ttl)` | Serve from cache, run again once `ttl` has passed. |
| `WithCacheParams(names...)` | Which query parameters take part in the cache key, as for a page. |
| `WithDependency(tags...)` | Tags every response from this document carries. |
| `WithRedirect(from, to, status)` / `WithPermanentRedirect(from, to)` | Old paths that redirect here. |
| `Build()` / `BuildErr()` | The document, and the errors the chain collected. |

Note the default. A page you forget to give a strategy is still a page; a document
you forget to give one runs its handler on every request and is skipped by a static
export. Sitemaps, feeds and `robots.txt` almost always want `Static()` or
`Incremental(ttl)`.

`Build` records `collage.ErrNoDocumentHandler` when no handler was set. What the
builder recorded stays on the document, and `RegisterDocument` refuses it by name
whether or not you called `BuildErr`, so checking it yourself is optional.

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

The tags you return are merged with the document's `WithDependency` tags. As with
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

A document has no timeout of its own. Its handler runs under
`Config.Template.Timeout` (five seconds by default) — the same setting that is the
default data-handler timeout for fragments. The limit applies to the `ctx` you are
given, so pass it on to whatever you call.

### Errors are plain text

A failing document never answers with an HTML error page: a crawler that asked for
`sitemap.xml` has no use for one. That includes a `405` for a method the document
does not answer (plain text since v0.11.0). It gets `text/plain` with the right status, one
generic line in production and the route name with the full error chain in
development, and `Cache-Control: no-store`.

If you want a format-specific error — a JSON `{"error": "..."}` — handle the error
inside the handler and return that body yourself. Whatever the handler returns is
served.

## Caching

Everything in [Caching](/docs/caching) applies. The cache key is the path, the
locale, the path parameters and the query string; only `GET` and `HEAD` are served
from cache; the `Cache-Control` header follows the strategy.

A sitemap built from your posts should be dropped when a post is published, which
is what tags are for:

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
post's tag is on it, so editing any post in the feed drops it.

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

- `Static()` and `Incremental(ttl)` documents are written. `Dynamic()` ones are
  skipped and named in the report, which is why the scaffold's `/healthz` never
  appears in `dist/`.
- An empty body is refused with `collage.ErrEmptyDocumentBody`, and no file is
  written.
- A pattern with a `{param}` needs a `BuildOptions.DocumentPathProvider` to list
  its concrete paths, or it is skipped with `collage.ErrDynamicPathUnresolved`.
- A document with `WithCacheParams` — a paginated feed — is written without a
  query string, since a file cannot have one, and the report warns about it, as it
  does for a page.
- Two paths that resolve to one file — a provider returning a path twice — are
  built once; the rest are skipped with `collage.ErrDuplicateOutputPath`.

`DocumentPathProvider` is `PathProvider`'s counterpart for documents — a separate
interface, so a provider written for pages does not have to change:

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

The document it expands is registered with the placeholder as a whole segment — the
literal `rss.xml` after it is what makes the exported file `feeds/go/rss.xml`:

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

`Params` is what the handler reads through `rc.Param`, the same as a live request
would have captured.

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
