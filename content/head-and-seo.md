---
description: Titles, meta tags, stylesheets and structured data declared by the fragment that knows them, placed by the layout.
reference: RenderContext, FragmentBuilder.WithTitle, Effect, ErrNoPathInLocale
---

# Head and SEO

The fragment that knows a page's title is rarely the one that writes `<head>`. The
post knows its own title and summary; the gallery knows it needs `gallery.css`;
the layout, which wrote `<head>` before either of them rendered, knows neither.

collage solves this by **hoisting**: a fragment *declares* what belongs in the
head, from wherever it sits in the page, and the layout says where declarations
land.

## Where it lands

The layout places a marker in its `<head>`:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
  {{hoist "head"}}
</head>
<body>
  {{slot "content"}}
</body>
</html>
```

`{{hoist "head"}}` writes a marker, not content — nothing below it has rendered
yet.

The marker comes after the site's stylesheet, so a stylesheet a fragment hoists is
linked after it and wins a rule of equal specificity; the layout `collage new`
writes has it there since v0.77.0. A layout of yours that puts the marker first
gives the site's stylesheet the last word instead.

When the whole page is finished, collage replaces the marker with everything
that was declared for the `"head"` area, anywhere in the tree. One pass, and the
layout still decides the position.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 300" role="img" aria-labelledby="hs-hoist-title">
<title id="hs-hoist-title">Hoisting on a blog post page. The layout declares the title The Wire. The post, inside it, declares its own title and a description meta from its data handler. Two gallery fragments inside the post each ask for gallery.css. The declarations are collected one entry per key: the title is the post's, because the innermost declaration wins; the description is the post's; gallery.css was asked for twice and is linked once. Once the whole page has rendered, the entries replace the layout's hoist head marker.</title>
<defs><marker id="hs-hoist-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="4" width="312" height="140" rx="8" />
<text x="14" y="23">layout</text>
<g class="d-dropped"><rect class="d-chip" x="196" y="10" width="110" height="18" rx="9" /><text class="d-small" x="251" y="22" text-anchor="middle">title: The Wire</text></g>
<rect class="d-frame d-frame-a" x="14" y="36" width="292" height="100" rx="6" />
<text class="d-slot-text-a" x="24" y="53">post</text><text class="d-small" x="60" y="53">data handler</text>
<rect class="d-chip-hit" x="24" y="60" width="120" height="18" rx="9" /><text class="d-hit-text d-small" x="84" y="72" text-anchor="middle">title: the post's</text>
<rect class="d-chip-hit" x="152" y="60" width="144" height="18" rx="9" /><text class="d-hit-text d-small" x="224" y="72" text-anchor="middle">meta:description</text>
<rect class="d-c" x="24" y="86" width="130" height="42" rx="5" /><text class="d-on-c" x="89" y="103" text-anchor="middle">gallery</text><text class="d-on-c d-small" x="89" y="119" text-anchor="middle">gallery.css</text>
<rect class="d-c" x="166" y="86" width="130" height="42" rx="5" /><text class="d-on-c" x="231" y="103" text-anchor="middle">gallery</text><text class="d-on-c d-small" x="231" y="119" text-anchor="middle">gallery.css</text>
<line class="d-flow" x1="160" y1="144" x2="160" y2="166" marker-end="url(#hs-hoist-arrow)" />
<text class="d-small" x="168" y="159">one entry per key</text>
<rect class="d-frame" x="4" y="170" width="312" height="70" rx="6" />
<text class="d-small" x="14" y="190">title</text><text class="d-small" x="186" y="190">the post's: inner wins</text>
<text class="d-small" x="14" y="208">meta:description</text><text class="d-small" x="186" y="208">the post's</text>
<text class="d-small" x="14" y="226">stylesheet:/static/gallery.css</text><text class="d-small" x="186" y="226">asked twice, one link</text>
<line class="d-flow" x1="160" y1="240" x2="160" y2="262" marker-end="url(#hs-hoist-arrow)" />
<text class="d-small" x="168" y="255">once the page has rendered</text>
<rect class="d-cmd" x="40" y="266" width="240" height="28" rx="6" /><text class="d-cmd-text" x="160" y="284" text-anchor="middle">hoist "head" marker filled</text>
</svg>
<figcaption>Every fragment declares what it knows, from wherever it sits: the layout a default title, the post its own title and a description, each gallery the stylesheet it needs. The declarations are collected one entry per key. The post's title replaces the layout's because the innermost declaration of a key wins, and two requests for <code>gallery.css</code> become one link. Once the whole page has rendered, the entries replace the layout's <code>{{hoist "head"}}</code> marker, in the order of each key's first declaration.</figcaption>
</figure>

Without the marker, declarations go nowhere. That is the first thing to check when
a title or a plugin's output is missing.

## A title without a handler

A title known when the program starts — the site's name on its layout, the name of
an about page — needs no code. `WithTitle` declares it on the fragment:

```go
layout := collage.NewFragment("layout", "layouts/master.html").
	WithTitle("The Wire").
	Build()
```

It is the same declaration `rc.HoistTitle` makes, so it follows the same rules
[below](#keys-and-the-innermost-wins): the innermost one wins, and a title the same
fragment's data handler hoists replaces it. Unlike a handler, it leaves a page that
declares no strategy static — see [Caching](/docs/caching#a-page-that-declares-none).

## Declaring from a data handler

Everything else — a title that comes from the content, a description, a
stylesheet — is declared by the helpers on `RenderContext`. Call them from a data
handler:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}

	rc.HoistTitle(post.Title + " — The Wire")
	rc.HoistMeta("description", post.Summary)
	rc.HoistProperty("og:title", post.Title)
	rc.HoistProperty("og:image", post.CoverURL)
	rc.HoistLink("canonical", siteOrigin+"/blog/"+post.Slug)
	if err := rc.HoistStylesheet("/static/post.css"); err != nil {
		return nil, nil, err
	}

	return post, []string{"post:" + post.Slug}, nil
}
```

| Call | Emits | Key |
| --- | --- | --- |
| `rc.HoistTitle(text)` | `<title>text</title>` | `title` |
| `rc.HoistMeta(name, content)` | `<meta name="…" content="…">` | `meta:<name>` |
| `rc.HoistProperty(property, content)` | `<meta property="…" content="…">` (Open Graph) | `property:<property>` |
| `rc.HoistLink(rel, href)` | `<link rel="…" href="…">` | `link:<rel>` |
| `rc.HoistAlternate(hreflang, href)` | `<link rel="alternate" hreflang="…" href="…">` (since v0.10.0) | `alternate:<hreflang>` |
| `rc.HoistStylesheet(path)` | `<link rel="stylesheet" href="…">` at the file's [content-addressed URL](/docs/assets) | `stylesheet:<path>` |

Every helper escapes what it is given, so a title built from a post's name is safe
whatever the post is called.

`HoistStylesheet` takes the file's mounted path and returns an error when no mount
has that file — a page linking a stylesheet that does not exist is a broken page,
and it should say so.

Call them synchronously, from the data handler's own goroutine. A declaration made
from a goroutine the handler started has no position in the page.

### From a template

A fragment's template can ask for a stylesheet itself:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

`{{stylesheet}}` renders nothing where it stands; it does what `HoistStylesheet`
does. The gallery's CSS travels with the gallery, onto whichever page it is on.

## Keys, and the innermost wins

Every declaration has a key, and the key decides what counts as the same thing:

- **Different keys all appear**, each at its earliest declaration in page order.
- **The same key declared twice keeps the innermost declaration.**

That one rule does two jobs. A stylesheet is keyed by its path, so five fragments
asking for `gallery.css` produce one `<link>`, and different stylesheets
accumulate. A title has one key, so a more specific fragment's title replaces a
less specific one's. That makes defaults easy — declare them in the layout, and let
any page override them:

```go
layout := collage.NewFragment("layout", "layouts/master.html").
	WithTitle("The Wire").
	WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistMeta("description", "News about the sea, and the people who live beside it.")
		return nil
	})).
	Build()
```

A post nested inside that layout declares its own title and description, and those
are the ones on the page. A page that declares nothing keeps the layout's.

Do not also write a literal `<title>` in the layout's template — it would sit beside
the hoisted one, and a page with two titles has one a browser ignores. The layout
`collage new` scaffolds names the site with `WithTitle`, so a page's own title
replaces it.

A handler on the layout is a handler on every page, so the description above makes
every page that declares no strategy dynamic. A site that wants them cached says
`Static()` or `Incremental(ttl)` on its pages, or leaves the layout with
`WithTitle` alone.

Three details, so they are not a surprise:

- **Position comes from the first declaration, not the winning one.** The head does
  not reorder itself depending on whether something nested happened to override a
  title.
- **At equal depth, the later-declared fragment wins.** Two sibling fragments
  declaring one key is a real conflict, and it is settled by their order in the
  page — not by whose data handler finished first, which would make the head change
  from request to request.
- **The order is the page's, not the clock's.** Sibling data handlers run
  concurrently, but a key is placed by where the fragment that first declares it
  sits in the page, then by the order that fragment declared in — so the head is the
  same on every render, stylesheets that override one another included. (Before
  v0.12.0 a key sat where its first declaration happened to arrive, and siblings'
  keys could swap between requests.)

## Anything else: rc.Hoist

The helpers are built on one method:

```go
func (rc *RenderContext) Hoist(area, key string, html template.HTML)
```

It inserts `html` **exactly as written**, under `key`, into `area`. That is what it
is for — markup the helpers do not cover — and it means escaping is your job.
Build the markup from values you control, or escape them:

```go
rc.Hoist("head", "preload:hero", template.HTML(
	`<link rel="preload" as="image" href="`+html.EscapeString(post.CoverURL)+`">`))
```

The same key rules apply, so choose a key that says what the thing is: one per
thing that should appear once, one per value for things that accumulate.

The area is any name, and a layout can have more than one marker — a `{{hoist
"scripts"}}` before `</body>`, say, for scripts a fragment needs:

```go
rc.Hoist("scripts", "script:map", template.HTML(`<script src="`+mapJS+`" defer></script>`))
```

## Canonical and alternate links

A canonical link tells search engines which URL is the real one when several reach
the same content — with and without a tracking query, say. Give it an absolute
URL. collage builds paths, not origins, so keep your site's origin in your own
configuration and join the two:

```go
rc.HoistLink("canonical", siteOrigin+rc.Request.URL.Path)
```

A site in several languages should also say where each translation is, with one
`<link rel="alternate" hreflang="…">` per language. `HoistLink` keeps one link per
`rel`, so it cannot say that; `rc.HoistAlternate` keys each link by its language
instead, so every translation appears once. Declare them from the layout, for
whichever page is being rendered, with [`app.URL`](/docs/links-and-locales#links-from-go):

```go
func Layout(app *collage.App) *collage.Fragment {
	return collage.NewFragment("layout", "layouts/master.html").
		WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
			for _, locale := range []string{"en", "tr"} {
				href, err := app.URL(rc.Page.Name, locale, rc.PathParams)
				if errors.Is(err, collage.ErrNoPathInLocale) {
					continue // not translated into this one
				}
				if err != nil {
					return err
				}
				rc.HoistAlternate(locale, siteOrigin+href)
			}
			return nil
		})).
		Build()
}
```

A page whose translation has a different slug — which only its content knows —
declares its own `rc.HoistAlternate` for that language, and, being inner, it
replaces the layout's.

The template equivalent, for a layout without a handler — which leaves pages that
declare no strategy static — is
[`localeURL`](/docs/links-and-locales#a-language-switcher), which is empty for a
language the page does not exist in:

```html
{{with localeURL "tr"}}<link rel="alternate" hreflang="tr" href="https://thewire.example{{.}}">{{end}}
```

## Structured data

Search engines read schema.org structured data from JSON-LD blocks in the head.
The [`elagoht/jsonld` plugin](/docs/plugins#elagohtjsonld) writes them for you, from typed values
rather than hand-built JSON:

```go
import "github.com/Elagoht/collage-jsonld"

func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	jsonld.Emit(rc, jsonld.BlogPosting{
		Headline:      post.Title,
		Description:   post.Summary,
		DatePublished: post.PublishedAt,
		AuthorName:    post.Author,
	})
	return post, []string{"post:" + post.Slug}, nil
}
```

`jsonld.Emit` hoists into `"head"`, so it needs the same marker as everything else
on this page, and it follows the same rules: one key per schema.org type and node
`Key`, the innermost winning.

Registering the plugin (`jsonld.New(jsonld.Options{})` in `Config.Plugins`) adds a
site-wide `WebSite` node when it is configured with a site name. The per-page data
always comes from your data handlers, because only they know what the page is
about.

Six more plugins cover the rest of what crawlers and search engines are told — the
first three what a crawler reads beyond the head:

| Plugin | What it does |
| --- | --- |
| [elagoht/sitemap](/docs/plugins#elagohtsitemap) | Serves `/sitemap.xml` from the pages you registered |
| [elagoht/robots](/docs/plugins#elagohtrobots) | Serves `/robots.txt` |
| [elagoht/feed](/docs/plugins#elagohtfeed) | Serves RSS and Atom feeds, and announces them in every page's head |
| [elagoht/meta](/docs/plugins#elagohtmeta) | Writes the Open Graph and Twitter tags, the canonical URL and the `hreflang` links described on this page, from site-wide defaults and what each page says |
| [elagoht/redirects](/docs/plugins#elagohtredirects) | Serves a migration's old addresses from a file, and hands them to a static build |
| [elagoht/indexnow](/docs/plugins#elagohtindexnow) | Tells search engines which URLs an invalidation changed |

## Hoisting and the rest of collage

- **Cached pages keep their head.** The marker is replaced before the page is
  stored, so a cached page is served with everything that was declared.
- **A fragment at its own URL has no head.** A fragment served through
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) has no
  layout. Since v0.18.0 what it hoists into an area it placed no marker for comes
  ahead of its markup, one inert `<template data-collage-hoist>` per item, for a
  script to add to the page — see
  [What it hoists](/docs/forms-and-actions#what-it-hoists).
- **Hoist from data handlers, not from `rc.Page`.** `rc.Page` is the registered
  page, shared by every request; writing to it is a data race. What varies per
  request is declared through `rc`.

## Example: a catalogue product page

A shop's product page at `/products/{id}` — a smaller version of the one
[Fragments and slots](/docs/fragments-and-slots#example-a-product-page) builds — needs
a head that a search engine and a link preview can read. Each thing in it is
declared by the fragment that knows it:

| In the head | Declared by | Key |
| --- | --- | --- |
| The title, `Brass deck lamp — The Shop` | The product's handler, replacing the layout's `WithTitle` | `title` |
| The description | The product's handler, replacing the layout's | `meta:description` |
| `og:title`, `og:description`, `og:image` | The product's handler | `property:og:…` |
| The canonical link | The product's handler | `link:canonical` |
| `product.css`, which only this page needs | The product's handler, with `HoistStylesheet` | `stylesheet:/static/product.css` |
| `stock.css` | The stock box's template, with `{{stylesheet}}` | `stylesheet:/static/stock.css` |

The layout's template is the one [at the top of this page](#where-it-lands). The
layout names the shop and describes it, for any page that says nothing of its own:

```go
layout := collage.NewFragment("layout", "layouts/master.html").
	WithTitle("The Shop").
	WithData(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistMeta("description", "Lamps, rope and brass fittings for small boats.")
		return nil
	})).
	Build()
```

The product's handler declares the rest as soon as it knows which product it is.
The catalog and inventory packages stand for your own data layer:

```go
const siteOrigin = "https://shop.example"

type product struct {
	ID       string
	Name     string
	Price    string
	Summary  string
	ImageURL string
}

type stockLevel struct {
	Count int
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (product, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Product(ctx, id)
	if errors.Is(err, catalog.ErrNoProduct) {
		return product{}, nil, fmt.Errorf("shop: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return product{}, nil, err
	}

	rc.HoistTitle(p.Name + " — The Shop")
	rc.HoistMeta("description", p.Summary)
	rc.HoistProperty("og:title", p.Name)
	rc.HoistProperty("og:description", p.Summary)
	rc.HoistProperty("og:image", p.ImageURL)
	rc.HoistLink("canonical", siteOrigin+"/products/"+id)
	if err := rc.HoistStylesheet("/static/product.css"); err != nil {
		return product{}, nil, err
	}

	return product{ID: p.ID, Name: p.Name, Price: p.Price, Summary: p.Summary, ImageURL: p.ImageURL},
		[]string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (stockLevel, error) {
	n, err := inventory.Count(ctx, rc.Param("id"))
	if err != nil {
		return stockLevel{}, err
	}
	return stockLevel{Count: n}, nil
}
```

The stock box asks for its own stylesheet, so it has it on whichever page it is
rendered:

```html
<!-- templates/fragments/stock.html -->
{{stylesheet "/static/stock.css"}}
<p class="stock">{{.Count}} in stock</p>
```

The page opens the box at a URL of its own, for a script that keeps the count
fresh:

```go
stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

content := collage.NewFragment("product", "pages/product.html").
	WithData(collage.LoadTagged(loadProduct)).
	WithSlotFragment("stock", stock).
	Required().
	Build()

page := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(content).
	WithPath("en", "/products/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	Build()
```

### What the page's head holds

A request for `/products/42` answers 200. Its head reads as follows, one element per
line here; collage writes the hoisted ones side by side:

```html
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/static/app.2708d73bf31c36cd.css">
  <title>Brass deck lamp — The Shop</title>
  <meta name="description" content="A lamp for a boat&#39;s chart table.">
  <meta property="og:title" content="Brass deck lamp">
  <meta property="og:description" content="A lamp for a boat&#39;s chart table.">
  <meta property="og:image" content="https://shop.example/img/lamp.jpg">
  <link rel="canonical" href="https://shop.example/products/42">
  <link rel="stylesheet" href="/static/product.ab1d57cf7f5ec545.css">
  <link rel="stylesheet" href="/static/stock.1d101be0b97db6e6.css">
</head>
```

- **The title and the description come first among the hoisted items**, where the
  layout declared them, but they carry the product's values: the position is the first declaration's,
  the value the innermost one's.
- **The apostrophe in the summary is escaped**, as every helper escapes what it is
  given.
- **Both stylesheets are linked at their content-addressed URLs**, the product's
  before the stock box's, in the order the page declared them. `app.css` comes
  before them, because the layout's template writes it above the marker, so a
  rule in `product.css` wins over one of equal specificity in `app.css`.

### What the stock box's URL answers

`GET /products/42/stock` renders the stock box alone, with no layout around it and
no marker for `"head"`. What it hoisted comes ahead of its markup, as
[What it hoists](/docs/forms-and-actions#what-it-hoists) describes:

```html
<template data-collage-hoist="head" data-collage-key="stylesheet:/static/stock.css"><link rel="stylesheet" href="/static/stock.1d101be0b97db6e6.css"></template>
<p class="stock">12 in stock</p>
```

Only the box's own declarations travel with it. The product's handler is not part
of this render, so there is no title and no Open Graph tag here — the page that
fetched the box keeps its own.

The stylesheet carries the key the page's head deduplicated it by. This page linked
`stock.css` on its first render already, so a script has nothing to add. On a page
that rendered without the box, this is how the box's stylesheet reaches the client,
inert until a script adds it to the head.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 170" role="img" aria-labelledby="hs-frag-title">
<title id="hs-frag-title">The same stock box answered two ways. Inside the product page, a request for /products/42 puts the title, description and Open Graph tags, product.css and the stock box's stock.css in the layout's head, and the stock box in the body. At its own URL, a request for /products/42/stock answers with a template element carrying stock.css, then the stock box's markup, with no layout around it and nothing the product's handler declared.</title>
<rect class="d-pill" x="4" y="4" width="152" height="22" rx="11" /><text class="d-small" x="80" y="19" text-anchor="middle">GET /products/42</text>
<rect class="d-frame" x="4" y="34" width="152" height="130" rx="6" />
<text class="d-small" x="14" y="50">&lt;head&gt;</text>
<rect class="d-chip" x="14" y="56" width="132" height="16" rx="8" /><text class="d-small" x="80" y="67" text-anchor="middle">title · meta · og:*</text>
<rect class="d-chip" x="14" y="76" width="132" height="16" rx="8" /><text class="d-small" x="80" y="87" text-anchor="middle">product.css</text>
<rect class="d-chip-hit" x="14" y="96" width="132" height="16" rx="8" /><text class="d-hit-text d-small" x="80" y="107" text-anchor="middle">stock.css</text>
<text class="d-small" x="14" y="130">&lt;body&gt;</text>
<rect class="d-c" x="14" y="136" width="132" height="20" rx="4" /><text class="d-on-c d-small" x="80" y="150" text-anchor="middle">stock box</text>
<rect class="d-pill" x="164" y="4" width="152" height="22" rx="11" /><text class="d-small" x="240" y="19" text-anchor="middle">GET /products/42/stock</text>
<rect class="d-frame" x="164" y="34" width="152" height="130" rx="6" />
<text class="d-small" x="174" y="50">response body</text>
<rect class="d-chip-hit" x="174" y="56" width="132" height="34" rx="8" /><text class="d-hit-text d-small" x="240" y="70" text-anchor="middle">&lt;template&gt;</text><text class="d-hit-text d-small" x="240" y="83" text-anchor="middle">stock.css</text>
<rect class="d-c" x="174" y="96" width="132" height="20" rx="4" /><text class="d-on-c d-small" x="240" y="110" text-anchor="middle">stock box</text>
<text class="d-small" x="174" y="138">no layout around it</text>
<text class="d-small" x="174" y="154">no title, no og:*</text>
</svg>
<figcaption>The same stock box, answered two ways. Inside the product page, its <code>stock.css</code> joins the title, the description, the Open Graph tags and <code>product.css</code> in the layout's head. At its own URL there is no layout: the stylesheet comes first, in an inert <code>&lt;template data-collage-hoist&gt;</code> that names its key, then the box's markup, and nothing the product's handler declared is in the answer.</figcaption>
</figure>
