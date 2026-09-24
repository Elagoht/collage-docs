---
description: What a page is, how its layout and content fit together, the paths that reach it, how it is cached, what it shows when it fails, and what registration does to it.
---

# Pages and layouts

A **page** is a render configuration with a name. It says which fragments make up
the page — a layout and the content inside it — which URLs reach it, how its
output is cached, and which page to show instead when it cannot be rendered. It
has no template and no data of its own; those belong to its fragments.

```go
page := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()

if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

## Building a page

`collage.NewPage(name)` starts a builder, each `WithX` call sets one thing, and
`Build()` returns the `*collage.Page`.

The name is the page's identity. Links are built from it
(`{{pageURL "blog-post" "slug" .Slug}}`), tests and plugins look pages up by it,
and two pages cannot share one. Choose it for what the page is, not where it
lives, so that it survives the URL changing.

The builders never stop a chain to return an error. A call that cannot do what it
was asked records the error and carries on, and `BuildErr()` returns everything
recorded:

```go
builder := collage.NewPage("blog-post").WithLayout(layout).WithPath("en", "/blog/{slug}")
page := builder.Build()
if err := builder.BuildErr(); err != nil {
	return err // collage.ErrMissingContent: there is no WithContent
}
```

Ignoring `BuildErr` does not let a mistake through. What a builder recorded stays
on the value it built, and `RegisterPage` refuses a page carrying any — its own, or
those of any fragment in its tree, such as a slot declared twice — with an error
naming the page, whether or not anyone called `BuildErr`. Check `BuildErr` when a
page is built from input you do not control, where you want the error closer to its
cause.

## Layout and content

Most pages on a site share their outside — the `<head>`, the header, the footer —
and differ inside. The outside is the **layout**, a fragment with a slot named
`content`. The inside is the **content fragment**, the one the page exists to
show.

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithSlot(collage.DefaultContentSlot, true, false). // "content": required, one fragment
	Build()
```

```html
<!-- templates/layouts/default.html -->
<!doctype html>
<html lang="en">
<head>{{hoist "head"}}</head>
<body>
  <header>…</header>
  {{slot "content"}}
  <footer>…</footer>
</body>
</html>
```

`WithLayout(layout)` and `WithContent(post)` name the two, and **registration puts
the content into the layout's `content` slot**. You do not bind it yourself. Leave
the layout's `content` slot empty: registration fills it, and a slot that holds
one fragment refuses a second (`ErrSlotOccupied`).

A page must have content — `ErrMissingContent` otherwise. The layout is optional:
a page without one renders its content fragment as the whole response, which is
what a page whose template is a complete HTML document wants.

### One layout, many pages

A layout is the most reused thing on a site, so it is safe to share one
`*collage.Fragment` value between every page, error pages included. Registration
gives each page a **private copy of the layout's slot table** before binding the
content into it, so page A's content never appears on page B.

Only the slot table is copied. The layout's template, its data handler, its
fallback and the fragments already bound into its other slots stay shared — which
means registration takes a snapshot of those bindings. A fragment bound into the
shared layout after a page was registered does not appear on that page. Build the
layout completely, then register pages with it.

The scaffold writes its layout as a function, `layouts.Layout()`, which returns a
new fragment each call. That works just as well; sharing one value is simply
allowed.

## Paths

`WithPath(locale, pattern)` registers the URL that reaches the page in one locale.
A site with one language uses one locale, and a page may have a different path in
each:

```go
collage.NewPage("about").
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda")
```

A pattern is made of segments:

| Segment | Matches |
| --- | --- |
| `blog` | Exactly that text |
| `{slug}` | Exactly one segment, captured as `slug` |
| `{rest...}` | Everything that is left, captured as `rest`. Only as the last segment |

A data handler reads what was captured with `rc.Param("slug")`, or
`rc.PathParams["slug"]`. Values arrive percent-decoded, one segment at a time, so
an encoded `/` inside a segment is part of the value rather than a new segment.

At every level a static segment is tried before a `{param}`, and a `{param}` before
a `{rest...}` — with backtracking, so `/blog/archive` beats `/blog/{slug}` even when
both could match. `/blog` and `/blog/` are the same route.

Mistakes in patterns are errors at registration, not surprises at request time:

- A pattern must start with `/`, have no empty segment and no empty placeholder
  name, and put a catch-all only last — `ErrInvalidPath` or `ErrInvalidPattern`.
- Two routes at one path in one locale — `ErrDuplicateRoute`. That includes a
  page and a [document](/docs/documents) colliding, since they share one tree.
- Two parameter names at one position, such as `/blog/{slug}` and
  `/blog/{id}/edit` — `ErrAmbiguousParameterName`.

A page answers `GET` and `HEAD`. Any other method is a 405 with an `Allow` header,
unless the page has an [action](/docs/forms-and-actions) for it — which is how a
form posts to the page it sits on.

Locales, locale prefixes such as `/tr/hakkinda` — and the redirect that sends
`/en/about` to `/about` — and building links from page names are covered in
[Links and locales](/docs/links-and-locales). A page can also
carry redirects from old URLs, with `WithRedirect(from, to, status)` and
`WithPermanentRedirect(from, to)`.

## Render strategies

Every page has one of three strategies, which decide whether its output is cached:

| Method | Strategy | Behaviour |
| --- | --- | --- |
| `Dynamic()` | `StrategyDynamic` | Rendered on every request, never cached. **The default** |
| `Static()` | `StrategyStatic` | Rendered once, served from cache until its tags are invalidated |
| `Incremental(ttl)` | `StrategyIncremental` | Served from cache, rendered again once `ttl` has passed |

`Incremental` needs a positive TTL (`ErrMissingTTL`). `Static` and `Incremental`
pages are also the ones `collage export` can write to files; a `Dynamic` page is
[skipped](/docs/static-export#what-is-skipped), because it exists to render per
request.

Caching only happens when the application enables it (`Config.Cache.Enabled`, on
in the scaffold), and never in development, where a cached page would hide the
template you just edited.

Two more builder methods shape a cached page. `WithDependency(tags...)` adds
dependency tags of the page's own to those its data handlers report, and
`WithCacheParams(names...)` says which query parameters take part in the cache
key. How all of this fits together — tags, invalidation, query parameters, caching
data rather than pages — is in [Caching](/docs/caching).

## Not-found and error pages

When a page cannot be shown, another page is shown in its place. There are two
situations and two levels.

- **Not found — 404.** No route matched the URL, or a `Required()` fragment's data
  handler returned an error wrapping `collage.ErrNotFound`: the content the URL
  names does not exist.
- **Error — 500.** A `Required()` fragment failed any other way, or something in
  the framework's own path did.

A page can name its own replacement pages, and the application can name
site-wide ones:

```go
post := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(postContent).
	WithPath("en", "/blog/{slug}").
	WithNotFoundPage(postNotFound). // "no such post", with a search box
	WithErrorPage(postError).
	Build()

app.RegisterNotFoundPage(siteNotFound)
app.RegisterErrorPage(siteError)
```

For a failure, the page's own `NotFoundPage` or `ErrorPage` is used first, then the
site-wide one, then the framework's built-in page. A URL that matched no route has
no page to ask, so it goes straight to the site-wide not-found page.

An error page is a page like any other — a layout, a content fragment, data
handlers if it wants them — with no path. It does not need one; it is reached by
something failing.

```go
func NotFoundPage() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").Build()

	return collage.NewPage("not-found").
		WithLayout(layouts.Layout()).
		WithContent(content).
		Dynamic().
		Build()
}
```

A few rules keep error pages from failing at the worst moment:

- **Every page named in `WithNotFoundPage` or `WithErrorPage` must be registered**,
  with `RegisterPage`, `RegisterNotFoundPage` or `RegisterErrorPage`. Otherwise the
  application refuses to start with `ErrUnregisteredErrorPage` — see
  [below](#why-the-registered-value-matters) for why.
- Their templates are checked when the page that names them is registered, so a
  typo in a 500 page is a startup error, not something discovered while the site is
  already failing.
- A page cannot be its own error page (`ErrSelfErrorPage`).
- An error page's render is never cached, and if it fails or renders nothing, the
  built-in page is served instead, and the failure is logged and reported to
  plugins. A 500 page that can 500 into itself is an outage.

The built-in page is self-contained HTML with no external files, so it renders
even when the assets are what broke. In development it names the fragment where
the failure started and shows the error chain; in production one generic
sentence, because error text leaks hostnames, file paths and credentials.
[Errors](/docs/errors) lists every error the framework reports.

`collage export` writes the page registered with `RegisterNotFoundPage` as
`404.html`, which is the file static hosts serve for a missing URL.

## Registration

```go
if err := app.RegisterPage(page); err != nil {
	log.Fatal(err)
}
```

`RegisterPage` is where a page is checked and put together, so that what would
fail at render time fails here, at startup, with the page's name in the message.
In order, it:

1. refuses a nil page, an empty name, a name another page holds
   (`ErrDuplicatePage`), and any registration after the application has started
   (`ErrAppStarted`);
2. refuses a page whose builder, or the builder of any fragment reachable from it,
   recorded a mistake — what `BuildErr()` would have returned;
3. copies the layout's slot table and binds the content fragment into its
   `content` slot;
4. validates the page and its whole fragment tree: paths, strategy and TTL,
   redirects, required slots with nothing in them, a fragment reachable from
   itself;
5. checks that every fragment's template was loaded (`ErrTemplateNotFound`) —
   including in the page's own not-found and error pages;
6. adds the page's paths, redirects, actions and fragment paths to the router.

Treat any error as fatal. A registration that fails halfway is not rolled back — a
path accepted in one locale stays in the router when the next is refused — because
a failed registration is a program that should not start, not a condition to
recover from.

### Why the registered value matters

Registration changes the page it is given. Step 3 replaces `page.LayoutFragment`
with the page's private, bound copy of the layout, and that copy — with the
content in its slot — is what renders. A page built again by calling the same
constructor is a different value, whose layout has an empty `content` slot.

So the value you passed to `RegisterPage` is the page from then on, and everything
that refers to a page by value must use that one:

- **Error pages.** `WithNotFoundPage(p)` must point at the very `*collage.Page` that
  was registered. A page of the same name built separately is refused with
  `ErrUnregisteredErrorPage`, because it would render its layout around nothing.
- **Actions that answer with a page.** `collage.RenderPage(p)` in an action must be
  given the registered value; an unregistered one is refused with
  `ErrUnregisteredPage` rather than rendered empty. The usual shape is to build the
  page once and let the action's closure capture it:

  ```go
  var page *collage.Page
  page = collage.NewPage("hello").
  	WithLayout(layouts.Layout()).
  	WithContent(content).
  	WithPath("en", "/hello").
  	WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
  		return collage.RenderPage(page), nil
  	}).
  	Build()
  return page
  ```

- **`rc.Page` is that value too**, shared by every request that renders the page
  and every goroutine serving them. Read it freely; never write to it. Anything that
  varies per request belongs in the data a handler returns or in the render's
  shared data — see [Data handlers](/docs/data-handlers#the-render-context).

`app.Page(name)` and `app.Pages()` return copies of registered pages, for looking
at them — a sitemap listing every page's paths, a test checking a strategy. A copy
is not the registered value, so do not pass one where a registered page is
expected.
