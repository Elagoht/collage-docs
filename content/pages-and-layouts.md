---
description: What a page is, how its layouts and content fit together, who may see it, the paths that reach it, how it is cached, what it shows when it fails, and what registration does to it.
reference: Registrable, ErrNilRegistrable, NewPage, PageBuilder, PageBuilder.WithLayouts, Page, RenderPage, DefaultContentSlot, StrategyAuto, FragmentBuilder.WithGuard, GuardFunc, GuardDecision
---

# Pages and layouts

A **page** is a render configuration with a name. It says which fragments make up
the page — a layout and the content inside it — which URLs reach it, how its
output is cached, and which page to show instead when it cannot be rendered. It
has no template and no data of its own; those belong to its fragments.

```go
page := collage.NewPage("blog-post").
	WithLayouts(layout).
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
builder := collage.NewPage("blog-post").WithLayouts(layout).WithPath("en", "/blog/{slug}")
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
and differ inside. The outside is the **layout**, a fragment whose template calls
a slot named `content`. The inside is the **content fragment**, the one the page
exists to show.

```go
layout := collage.NewFragment("layout", "layouts/default.html").
	WithTitle("My site").
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

The layout declares no slot. The `{{slot "content"}}` in its template is the
declaration, and so is any other slot it calls — see
[Fragments and slots](/docs/fragments-and-slots#slots). `WithTitle` gives every page
that uses it a `<title>` until something inside names a better one.

`WithLayouts(layout)` and `WithContent(post)` name the two, and **registration puts
the content into the layout's `content` slot**
(`collage.DefaultContentSlot`). You do not bind it yourself, and a layout whose
template never calls `{{slot "content"}}` is refused at registration with
`ErrUnknownSlot`, because the content would have nowhere to render.
`WithSlot(collage.DefaultContentSlot, true, false)` on the layout is still
accepted, for a layout that should refuse any fragment bound into `content` besides
the page's own (`ErrSlotOccupied`).

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

The scaffold writes its layout as a function, `layouts.Master()`, which returns a
new fragment each call. That works just as well; sharing one value is simply
allowed.

### Layouts inside layouts

A page can sit in more than one layout: the HTML shell with its `<head>`, and
inside it a narrower frame the sign-in pages share. `WithLayouts` takes the whole
chain, **outermost first** (since v0.28.0, which replaced `WithLayout` with it — a
single layout is `WithLayouts(layout)`):

```go
func Master() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("My site").
		Build()
}

func Auth() *collage.Fragment {
	return collage.NewFragment("auth-layout", "layouts/auth.html").Build()
}

page := collage.NewPage("login").
	WithLayouts(layouts.Master(), layouts.Auth()).
	WithContent(login).
	WithPath("en", "/login").
	Build()
```

Every layout in the chain renders the one inside it with `{{slot "content"}}`, as
a single layout renders the content. Registration does all of the wrapping: the
content goes into the innermost layout's `content` slot, and each layout into the
`content` slot of the one outside it. So a layout with a hole in it is still a
finished `*collage.Fragment`, and a helper that makes one returns it built, not as
a builder the page has to complete.

Each page gets its own copy of every layout's slot table, as it does of a single
layout, so the chain's layouts can be shared across the site. A layout in a chain
must arrive with its `content` slot empty — registration fills it — and is refused
with `ErrSlotOccupied` otherwise.

`WithLayouts` records `ErrMissingLayout` when called with no layouts,
`ErrNilFragment` for a nil one, `ErrFragmentCycle` for the same layout twice, and
`ErrConflictingLayout` when called a second time on one builder.

## Private pages: guards

A section of a site that only signed-in readers may see is a layout that says so.
`WithGuard` (since v0.28.0) gives a fragment a function that is asked, before the
page is served, whether this request may have it:

```go
func requireUser(ctx context.Context, r *http.Request) (*collage.GuardDecision, error) {
	if session.FromContext(ctx).Get("user") != "" {
		return nil, nil // allowed
	}
	return &collage.GuardDecision{
		Status:   http.StatusSeeOther,
		Location: "/login?next=" + url.QueryEscape(r.URL.RequestURI()),
	}, nil
}

func Private() *collage.Fragment {
	return collage.NewFragment("private", "layouts/private.html").
		WithGuard(requireUser).
		Build()
}

page := collage.NewPage("dashboard").
	WithLayouts(layouts.Master(), layouts.Private()).
	WithContent(dashboard).
	WithPath("en", "/dashboard").
	Build()
```

Every page wrapped in `Private()` is private, and a page is public by leaving it
out. There is no list of protected paths to keep in step with the routes: the
guard runs for whatever URL reaches the page, in every locale, with every
parameter.

A guard is asked for a page when it is on the page's **spine**: a layout in the
page's chain, or the page's content fragment. Guards run outermost first, the
content fragment's last, and the first that answers decides. A guard on any other
fragment — one bound into a slot, one a resolver returns, a fallback — is ignored:
who may see a page is a property of the page, not of the parts it is drawn from.

It answers in one of three ways:

| Return | What the reader gets |
| --- | --- |
| `nil, nil` | The page |
| a `3xx` status and a `Location` | A redirect there, with no body. A zero status with a location is `303 See Other`. A request marked with `Collage-Fetch` gets `204` and the destination in `Collage-Location`, as an action's redirect does. |
| a `4xx` or `5xx` status, no location | That status, with no body |

Anything else — a redirect with nowhere to go, a `200` — is
`ErrInvalidGuardDecision`, and the request fails with a 500, as it does when the
guard returns an error. A guard that cannot say what it means is not guessed
about: a redirect with no location would otherwise fall through to the very page
it was meant to keep back.

What the guard covers:

- **The page's renders**, `GET` and `HEAD`. The guard runs after routing and
  **before the page's cache is read**, so a blocked reader never reaches a cached
  render, and a private page may be `Static()`. It also runs before
  `PageResolvedHook`: a blocked request never reached the page, so plugins
  watching pages are not told about it.
- **The actions on the page's own URL.** A form posts to the page it sits on, and
  a page a reader may not see is a page whose form they may not submit. The guard
  runs before the body is read and before the forgery check.
- **The page's fragment paths.** A fragment opened at its own URL with
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) is a
  part of the page, and since v0.34.0 meets the page's guards before the
  fragment's own. Before it, a fragment path ran only its fragment's guard, and
  one on a private page was public unless that fragment carried a guard too.
- **Not an action registered at its own URL**, and not the page's not-found and
  error pages — a private error page would otherwise redirect the reader who hit
  the error.

Readers the guard allows share the page's cache, which is the server's own. A CDN
or proxy in front of the server runs no guard, so a guarded page goes out with
`Cache-Control: private, no-cache` whatever its strategy says. The guard's own
answer — a redirect to log in, a refusal — is sent `no-store` since v0.34.0, so a
CDN that keeps a 404 or a 308 by default does not hand one reader's answer to the
next. A page whose
content differs from one reader to the next is a personalisation question, not a
guard question — see [Caching](/docs/caching#render-strategies). A
[static export](/docs/static-export#what-is-skipped) does not write guarded pages
at all.

What a guard checks is not the framework's to know. `requireUser` above reads the
session; a guard can as well read a header, a role, a feature flag — anything a
request carries. The [session plugin](/docs/plugins#elagohtsession) ships this one
ready-made, as `session.RequireUser("/login")`. `collage inspect` lists, for each
page, the fragments that carry a guard.

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
`rc.PathParams["slug"]`. Values arrive percent-decoded, one segment at a time.

A segment holding an encoded slash, `%2F`, is a 404 since v0.34.0. A middleware
reads the decoded path, where `/public%2Fsecret` is two segments; a router that
read it as one — as collage did before — would serve a request the middleware had
judged as something else. A value that is itself a path, such as `guide/intro`,
belongs in a catch-all: `{rest...}`. `{{pageURL}}` and `BuildPath` refuse a `/` in
a single segment's value.

At every level a static segment is tried before a `{param}`, and a `{param}` before
a `{rest...}` — with backtracking, so `/blog/archive` beats `/blog/{slug}` even when
both could match. `/blog` and `/blog/` are the same route.

Mistakes in patterns are errors at registration, not surprises at request time:

- A pattern must start with `/`, have no empty segment and no empty placeholder
  name, and put a catch-all only last — `ErrInvalidPath` or `ErrInvalidPattern`.
- A placeholder is a whole segment. Since v0.11.0 one written inside a segment,
  such as `/feeds/{category}.xml` or `/post-{id}`, is `ErrInvalidPattern`; write
  `/feeds/{category}/rss.xml` instead.
- Two routes at one path in one locale — `ErrDuplicateRoute`. That includes a
  page and a [document](/docs/documents) colliding, since they share one tree.
- Two parameter names at one position, such as `/blog/{slug}` and
  `/blog/{id}/edit` — `ErrAmbiguousParameterName`.

A page answers `GET` and `HEAD`, and `OPTIONS` with a `204` whose `Allow` header
lists what the URL accepts. Any other method is a 405 with that same `Allow`
header, unless the page has an [action](/docs/forms-and-actions) for it — which is
how a form posts to the page it sits on.

Locales, locale prefixes such as `/tr/hakkinda` — and the redirect that sends
`/en/about` to `/about` — and building links from page names are covered in
[Links and locales](/docs/links-and-locales). A page can also
carry redirects from old URLs, with `WithRedirect(from, to, status)` and
`WithPermanentRedirect(from, to)`. Since v0.34.0 a redirect carries the request's
query string to its destination, a value it captured is escaped for where it
lands — a path segment, or a query value after a `?`, so `/login?next={slug}`
cannot be handed a second `next` — and a destination that would leave the site,
`//host` or `/\host`, is `ErrInvalidPath` at registration.

## Render strategies

Every page has one of three strategies, which decide whether its output is cached:

| Method | Strategy | Behaviour |
| --- | --- | --- |
| `Dynamic()` | `StrategyDynamic` | Rendered on every request, never cached |
| `Static()` | `StrategyStatic` | Rendered once, served from cache until its tags are invalidated |
| `Incremental(ttl)` | `StrategyIncremental` | Served from cache, rendered again once `ttl` has passed |

A page that calls none of the three is `StrategyAuto` until it is registered, and
registration resolves it: **dynamic** if anything it renders has a data handler or
a slot resolver, **static** otherwise. A page of templates and fixed values —
`WithData`, `WithTitle` — renders the same for everyone, so it is cached without
saying so. A handler may read the request, a cookie or the clock, and nothing
outside it can tell, so a page with one renders per request until it says
`Static()` or `Incremental(ttl)`. [Caching](/docs/caching#a-page-that-declares-none)
has the whole rule.

`Incremental` needs a positive TTL (`ErrMissingTTL`). Static and incremental
pages are also the ones `collage export` can write to files; a dynamic page is
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
	WithLayouts(layout).
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
		WithLayouts(layouts.Master()).
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
fail at render time fails here, at startup, with the page's name in the message —
for every fragment the page holds, including one it opens at its own URL. A
fragment a [slot resolver](/docs/fragments-and-slots#slots-filled-per-render) returns while a request runs
cannot be seen from here; it is checked when it first renders.
In order, it:

1. refuses a nil page, an empty name, a name another page holds
   (`ErrDuplicatePage`), and any registration after the application has started
   (`ErrAppStarted`);
2. refuses a page whose builder, or the builder of any fragment reachable from it
   or opened with [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url),
   recorded a mistake — what `BuildErr()` would have returned;
3. copies the slot table of every layout in the chain, binds the content fragment
   into the innermost one's `content` slot and each layout into the next outer
   one's;
4. validates the page and its whole fragment tree, fragment paths included:
   paths, strategy and TTL, redirects, required slots with nothing in them, a
   fragment reachable from itself;
5. checks that every fragment's template was loaded (`ErrTemplateNotFound`) —
   including the fragments opened with `WithFragmentPath`, which since v0.11.0 are
   checked like the rest of the tree, and the page's own not-found and error pages
   — and that every slot something is bound into is one its template calls
   (`ErrUnknownSlot`, naming the slot and the calls the template does make);
6. resolves the strategy of a page that declared none, from the whole tree;
7. adds the page's paths, redirects, actions and fragment paths to the router.

Treat any error as fatal. A registration that fails halfway is not rolled back — a
path accepted in one locale stays in the router when the next is refused — because
a failed registration is a program that should not start, not a condition to
recover from.

### Registering several at once

`app.Register` (since v0.40.0) takes pages, documents and actions in one call and
registers each in order, through `RegisterPage`, `RegisterDocument` or
`RegisterAction`. The scaffolded `routes.go` is a single call to it:

```go
func register(app *collage.App) error {
	if err := app.Register(
		landingpages.Home(),
		blogpages.Post(),
		documents.Health(),
		actions.Logout(),
	); err != nil {
		return err
	}
	return app.RegisterNotFoundPage(errorspages.NotFound())
}
```

It stops at the first one refused and returns its error wrapped with its kind and
name — `register page "post": collage: duplicate page ...` — so `errors.Is` still
finds the cause. What was registered before it stays registered, as with the
single methods. A `nil` is `ErrNilRegistrable`. The not-found and error pages are
reached by failing to match, not by a path, so they keep `RegisterNotFoundPage`
and `RegisterErrorPage`. `Register` takes a `collage.Registrable`, which only
`*Page`, `*Document` and `*Action` satisfy. [`collage add`](/docs/cli#collage-add)
adds what it writes to this call.

### Why the registered value matters

Registration changes the page it is given. Step 3 replaces `page.LayoutFragment`
with the page's private, bound copy of its outermost layout, and that copy — with
the rest of the chain and the content inside it — is what renders. A page built
again by calling the same constructor is a different value, whose layouts have
empty `content` slots.

So the value you passed to `RegisterPage` is the page from then on, and everything
that refers to a page by value must use that one:

- **Error pages.** `WithNotFoundPage(p)` must point at the very `*collage.Page` that
  was registered. A page of the same name built separately is refused with
  `ErrUnregisteredErrorPage`, because it would render its layout around nothing.
- **Actions that answer with a page.** `collage.RenderPage(p)` in an action must be
  given the registered value; an unregistered one is refused with
  `ErrUnregisteredPage` rather than rendered empty. An action on the page's own URL
  has it as `rc.Page` (since v0.33.0):

  ```go
  return collage.NewPage("hello").
  	WithLayouts(layouts.Master()).
  	WithContent(content).
  	WithPath("en", "/hello").
  	WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
  		return collage.RenderPage(rc.Page), nil
  	}).
  	Build()
  ```

- **`rc.Page` is that value too**, shared by every request that renders the page
  and every goroutine serving them. Read it freely; never write to it. Anything that
  varies per request belongs in the data a handler returns or in the render's
  shared data — see [Data handlers](/docs/data-handlers#the-render-context).

`app.Page(name)` and `app.Pages()` return copies of registered pages, for looking
at them — a sitemap listing every page's paths, a test checking a strategy. A copy
is not the registered value, so do not pass one where a registered page is
expected.
