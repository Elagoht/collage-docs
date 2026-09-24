---
description: Titles, meta tags, stylesheets and structured data declared by the fragment that knows them, placed by the layout.
reference: RenderContext, Effect, ErrNoPathInLocale
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
  {{hoist "head"}}
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
</head>
<body>
  {{slot "content"}}
</body>
</html>
```

`{{hoist "head"}}` writes a marker, not content — nothing below it has rendered
yet. When the whole page is finished, collage replaces the marker with everything
that was declared for the `"head"` area, anywhere in the tree. One pass, and the
layout still decides the position.

Without the marker, declarations go nowhere. That is the first thing to check when
a title or a plugin's output is missing.

## Declaring from a data handler

The helpers on `RenderContext` cover what nearly every page needs. Call them from
a data handler:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return postView{}, nil, err
	}

	rc.HoistTitle(post.Title + " — The Wire")
	rc.HoistMeta("description", post.Summary)
	rc.HoistProperty("og:title", post.Title)
	rc.HoistProperty("og:image", post.CoverURL)
	rc.HoistLink("canonical", siteOrigin+"/blog/"+post.Slug)
	if err := rc.HoistStylesheet("/static/post.css"); err != nil {
		return postView{}, nil, err
	}

	return postView{Post: post}, []string{"post:" + post.Slug}, nil
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
whatever the post is called. `HoistStylesheet` takes the file's mounted path and
returns an error when no mount has that file — a page linking a stylesheet that
does not exist is a broken page, and it should say so.

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
layout := collage.NewFragment("layout", "layouts/default.html").
	WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
		rc.HoistTitle("The Wire")
		rc.HoistMeta("description", "News about the sea, and the people who live beside it.")
		return nil
	})).
	WithSlot("content", true, false).
	Build()
```

A post nested inside that layout declares its own title and description, and those
are the ones on the page. A page that declares nothing keeps the layout's. Do not
also write a literal `<title>` in the layout's template — it would sit beside the
hoisted one, and a page with two titles has one a browser ignores. The layout
`collage new` scaffolds declares the site's name this way, so a page's
`rc.HoistTitle` replaces it.

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
	return collage.NewFragment("layout", "layouts/default.html").
		WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
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
		WithSlot("content", true, false).
		Build()
}
```

A page whose translation has a different slug — which only its content knows —
declares its own `rc.HoistAlternate` for that language, and, being inner, it
replaces the layout's. The template equivalent, for a layout without a handler, is
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

func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return postView{}, nil, err
	}
	jsonld.Emit(rc, jsonld.BlogPosting{
		Headline:      post.Title,
		Description:   post.Summary,
		DatePublished: post.PublishedAt,
		AuthorName:    post.Author,
	})
	return postView{Post: post}, []string{"post:" + post.Slug}, nil
}
```

`jsonld.Emit` hoists into `"head"`, so it needs the same marker as everything else
on this page, and it follows the same rules: one key per schema.org type, the
innermost winning. Registering the plugin (`jsonld.New()` in `Config.Plugins`)
adds a site-wide `WebSite` node when it is configured with a site name; the
per-page data always comes from your data handlers, because only they know what the
page is about.

## Hoisting and the rest of collage

- **Cached pages keep their head.** The marker is replaced before the page is
  stored, so a cached page is served with everything that was declared.
- **A fragment at its own URL has no head.** A fragment served through
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) has no
  layout, so what it hoists has nowhere to land.
- **Hoist from data handlers, not from `rc.Page`.** `rc.Page` is the registered
  page, shared by every request; writing to it is a data race. What varies per
  request is declared through `rc`.
