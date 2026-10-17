---
description: What a page is: its layouts and content, who may see it, the paths that reach it, how it is cached, what it shows when it fails, and what registration does.
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
`WithSlot(collage.DefaultContentSlot, collage.SlotRequired, collage.SlotSingle)` on the layout is still
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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 150" role="img" aria-labelledby="pl-chain-title">
<title id="pl-chain-title">The login page's layout chain after registration. The outermost layout, Master, has a content slot holding the auth layout, the second and innermost layout. The auth layout's content slot holds the login content fragment.</title>
<rect class="d-frame" x="4" y="4" width="312" height="142" rx="8" />
<text x="14" y="21">Master()</text>
<text class="d-small" x="306" y="21" text-anchor="end">1st · outermost</text>
<rect class="d-slot d-slot-a" x="14" y="30" width="292" height="106" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="44">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="52" width="272" height="76" rx="6" />
<text class="d-slot-text-a" x="34" y="69">Auth()</text>
<text class="d-small" x="286" y="69" text-anchor="end">2nd · innermost</text>
<rect class="d-slot d-slot-c" x="34" y="78" width="252" height="42" rx="4" />
<text class="d-small" x="40" y="103">slot "content"</text>
<rect class="d-c" x="126" y="84" width="152" height="30" rx="4" />
<text class="d-on-c" x="202" y="103" text-anchor="middle">login · content</text>
</svg>
<figcaption><code>WithLayouts(layouts.Master(), layouts.Auth())</code> lists the chain outermost first. Registration does the nesting: the login content goes into the content slot of <code>Auth()</code>, the innermost layout, and <code>Auth()</code> into the content slot of <code>Master()</code>, each on the page's own copy of the layout's slot table.</figcaption>
</figure>

Every layout in the chain renders the one inside it with `{{slot "content"}}`, as
a single layout renders the content. Registration does all of the wrapping: the
content goes into the innermost layout's `content` slot, and each layout into the
`content` slot of the one outside it.

So a layout with a hole in it is still a finished `*collage.Fragment`, and a helper
that makes one returns it built, not as a builder the page has to complete.

Each page gets its own copy of every layout's slot table, as it does of a single
layout, so the chain's layouts can be shared across the site. A layout in a chain
must arrive with its `content` slot empty — registration fills it — and is refused
with `ErrSlotOccupied` otherwise.

`WithLayouts` records an error, which registration refuses, when it is called:

| Called | Error |
| --- | --- |
| with no layouts | `ErrMissingLayout` |
| with a nil layout | `ErrNilFragment` |
| with the same layout twice | `ErrFragmentCycle` |
| a second time on one builder | `ErrConflictingLayout` |

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
content fragment's last, and the first that answers decides.

A guard on any other fragment — one bound into a slot, one a resolver returns, a
fallback — is ignored: who may see a page is a property of the page, not of the
parts it is drawn from.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 306" role="img" aria-labelledby="pl-spine-title">
<title id="pl-spine-title">A request for /admin/users meets the guards on the page's spine, outermost first. Master has no guard. Account's guard answers a signed-out reader with 303 to /login. Admin's guard answers a reader who is not an admin with 403 and no body. The content fragment, admin-users, is asked last. Only when no guard answers does the page render; for a cached page, the cache is read only after the guards too. A menu fragment bound into a slot is not on the spine, so a guard on it is not asked.</title>
<defs><marker id="pl-spine-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="40" y="6" width="140" height="24" rx="12" /><text x="110" y="22" text-anchor="middle">GET /admin/users</text>
<line class="d-flow" x1="110" y1="30" x2="110" y2="44" marker-end="url(#pl-spine-arrow)" />
<rect class="d-frame" x="20" y="48" width="180" height="34" rx="6" /><text x="110" y="62" text-anchor="middle">Master</text><text class="d-small" x="110" y="75" text-anchor="middle">layout 1 · no guard</text>
<line class="d-flow" x1="110" y1="82" x2="110" y2="96" marker-end="url(#pl-spine-arrow)" />
<rect class="d-a" x="20" y="100" width="180" height="34" rx="6" /><text class="d-on-a" x="110" y="114" text-anchor="middle">Account</text><text class="d-on-a d-small" x="110" y="127" text-anchor="middle">layout 2 · RequireUser</text>
<line class="d-flow" x1="200" y1="117" x2="212" y2="117" marker-end="url(#pl-spine-arrow)" />
<rect class="d-chip" x="216" y="100" width="100" height="34" rx="6" /><text class="d-small" x="266" y="114" text-anchor="middle">signed out</text><text class="d-small" x="266" y="127" text-anchor="middle">303 · /login</text>
<line class="d-flow" x1="110" y1="134" x2="110" y2="148" marker-end="url(#pl-spine-arrow)" />
<rect class="d-a" x="20" y="152" width="180" height="34" rx="6" /><text class="d-on-a" x="110" y="166" text-anchor="middle">Admin</text><text class="d-on-a d-small" x="110" y="179" text-anchor="middle">layout 3 · requireRole</text>
<line class="d-flow" x1="200" y1="169" x2="212" y2="169" marker-end="url(#pl-spine-arrow)" />
<rect class="d-chip" x="216" y="152" width="100" height="34" rx="6" /><text class="d-small" x="266" y="166" text-anchor="middle">not an admin</text><text class="d-small" x="266" y="179" text-anchor="middle">403 · no body</text>
<line class="d-flow" x1="110" y1="186" x2="110" y2="200" marker-end="url(#pl-spine-arrow)" />
<rect class="d-frame" x="20" y="204" width="180" height="34" rx="6" /><text x="110" y="218" text-anchor="middle">admin-users</text><text class="d-small" x="110" y="231" text-anchor="middle">content · asked last</text>
<g class="d-dropped"><rect class="d-frame" x="216" y="204" width="100" height="34" rx="6" /><text class="d-small" x="266" y="218" text-anchor="middle">menu · in a slot</text><text class="d-small" x="266" y="231" text-anchor="middle">guard not asked</text></g>
<line class="d-flow" x1="110" y1="238" x2="110" y2="252" marker-end="url(#pl-spine-arrow)" />
<rect class="d-c" x="20" y="256" width="180" height="24" rx="6" /><text class="d-on-c" x="110" y="272" text-anchor="middle">page renders</text>
<text class="d-small" x="160" y="298" text-anchor="middle">outermost first · the first answer decides</text>
</svg>
<figcaption>The spine of the admin page in <a href="#example-a-blog-with-a-private-area">Example: a blog with a private area</a>. A request meets the guards of its layouts, outermost first, then the content fragment's. The first guard that answers sends its answer and nothing further runs: a signed-out reader gets a redirect to log in from the account layout, a signed-in reader who is not an admin gets a bare 403 from the admin layout. Only a request no guard answered reaches the render, and on a cached page the cache, which is read only after the guards. A guard on a fragment bound into a slot, such as the admin menu, is not asked.</figcaption>
</figure>

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
  any render hook: a blocked request never reached the page, so plugins
  watching renders are not told about it.
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
`Cache-Control: private, no-cache` whatever its strategy says.

The guard's own answer — a redirect to log in, a refusal — is sent `no-store` since
v0.34.0, so a CDN that keeps a 404 or a 308 by default does not hand one reader's
answer to the next.

A page whose content differs from one reader to the next is a personalisation
question, not a guard question — see [Caching](/docs/caching#render-strategies). A
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
| `{slug}.md`, `post-{id}`, `v{version}.json` | One segment carrying that text before and after, with at least one character between — what is captured (since v0.41.0): `/blog/hello.md` gives `slug` = `hello` |
| `{rest...}` | Everything that is left, captured as `rest`. Only as the last segment, with no text around it |

A data handler reads what was captured with `rc.Param("slug")`, or
`rc.PathParams["slug"]`. Values arrive percent-decoded, one segment at a time.

A segment holding an encoded slash, `%2F`, is a 404 since v0.34.0. By the URL
standard `%2F` is not a separator: `/public%2Fsecret` is one segment whose value is
`public/secret`.

But Go decodes the path into `r.URL.Path` before any middleware reads it, and there
`/public%2Fsecret` has become `/public/secret`, which no longer tells an encoded
slash from a real one. So a middleware sees two segments where the router sees one.
A router that matched the one — as collage did before — would serve a request the
middleware had judged as something else, so collage serves neither reading.

A value that is itself a path, such as `guide/intro`, belongs in a catch-all:
`{rest...}`. `{{pageURL}}` and `BuildPath` refuse a `/` in a single segment's
value.

At every level a static segment is tried first, then a placeholder with text
around it — the most specific first — then a bare `{param}`, then a `{rest...}`,
with backtracking, so `/blog/archive` beats `/blog/{slug}` even when both could
match. `/blog` and `/blog/` are the same route.

### Text around a placeholder

A post and its Markdown, or a feed per category, sit side by side:

```go
collage.NewPage("post").WithPath("en", "/blog/{slug}")            // /blog/hello
collage.NewDocument("post-md", "text/markdown; charset=utf-8").
	WithPath("en", "/blog/{slug}.md")                               // /blog/hello.md
collage.NewDocument("feed", "application/rss+xml").
	WithPath("en", "/feeds/{category}.xml")                         // /feeds/go.xml
```

Which route a segment reaches never depends on the order of registration. Two
placeholders at one position are one of three things:

| | Example | Result |
| --- | --- | --- |
| Disjoint: no segment matches both | `{slug}.md` and `{slug}.json` | Both register; the text decides |
| Nested: one is the other made more specific | `{slug}.min.md` within `{slug}.md`, both within `{slug}` | Both register; the more specific wins |
| Crossing: a segment matches both, neither is more specific | `a{x}` and `{x}b`, both matching `aXb` | `ErrOverlappingPattern`, naming such a segment |

So with the routes above, `/blog/hello.md` is the document, `/blog/hello` the page,
a static `/blog/index.md` beats both, and `/blog/.md` — nothing between the text —
is the page's, with `slug` = `.md`. A placeholder with text around it never
captures `.` or `..`: `/blog/...md` is the page's too, rather than handing the
document the slug `..`. A link escapes only the value:
`{{pageURL "post-md" "slug" "çay"}}` is `/blog/%C3%A7ay.md`, and an export writes
`blog/hello.md` beside `blog/hello/index.html`.

A segment holds one placeholder — `{name}.{ext}` has no single answer for `a.b.c`,
so register `{name}.md` and `{name}.json` as two routes — and there is no regex. A
value is checked by its handler, which returns `collage.ErrNotFound` for one it
does not know; that keeps every collision decidable when the route is registered.

Mistakes in patterns are errors at registration, not surprises at request time:

- A pattern must start with `/`, have no empty segment and no empty placeholder
  name, and put a catch-all only last — `ErrInvalidPath` or `ErrInvalidPattern`.
- A segment holds at most one placeholder, and a catch-all takes no text around
  it: `/{name}.{ext}` and `/files/{path...}.md` are `ErrInvalidPattern`. (From
  v0.11.0 to v0.40.0 any text around a placeholder was refused; since v0.41.0
  `/feeds/{category}.xml` and `/post-{id}` are routes.)
- Two placeholders with text around them that cross, such as `/a{x}` and `/{x}b` —
  `ErrOverlappingPattern`.
- Two routes at one path in one locale — `ErrDuplicateRoute`. That includes a
  page and a [document](/docs/documents) colliding, since they share one tree.
- Two parameter names at one position, such as `/blog/{slug}` and
  `/blog/{id}/edit`, or `{slug}.md` and `{id}.md` — `ErrAmbiguousParameterName`.

A page answers `GET` and `HEAD`, and `OPTIONS` with a `204` whose `Allow` header
lists what the URL accepts. Any other method is a 405 with that same `Allow`
header, unless the page has an [action](/docs/forms-and-actions) for it — which is
how a form posts to the page it sits on.

Locales, locale prefixes such as `/tr/hakkinda` — and the redirect that sends
`/en/about` to `/about` — and building links from page names are covered in
[Links and locales](/docs/links-and-locales).

A page can also carry redirects from old URLs, with
`WithRedirect(from, to, status)` and `WithPermanentRedirect(from, to)`. Since
v0.34.0:

- a redirect carries the request's query string to its destination;
- a value it captured is escaped for where it lands — a path segment, or a query
  value after a `?`, so `/login?next={slug}` cannot be handed a second `next`;
- a destination that would leave the site, `//host` or `/\host`, is
  `ErrInvalidPath` at registration.

Since v0.52.0 a control character in either path, such as a line feed, is
`ErrInvalidRedirect`, and a static export fails on a redirect over a page it wrote;
see [Static export](/docs/static-export#what-fails).

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
`collage.Value`, `WithTitle` — renders the same for everyone, so it is cached without
saying so. A handler may read the request, a cookie or the clock, and nothing
outside it can tell, so a page with one renders per request until it says
`Static()` or `Incremental(ttl)`. [Caching](/docs/caching#a-page-that-declares-none)
has the whole rule.

`Incremental` needs a positive TTL (`ErrInvalidTTL`). Static and incremental
pages are also the ones `collage export` can write to files; a dynamic page is
[skipped](/docs/static-export#what-is-skipped), because it exists to render per
request.

Caching only happens when the application enables it (`Config.Cache.Enabled`, on
in the scaffold), and never in development, where a cached page would hide the
template you just edited.

Two more builder methods shape a cached page. `WithDependencyTags(tags...)` adds
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
		pages.Home(),
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

Registration changes the page it is given. Step 3 replaces the page's outermost
layout with a private, bound copy of it, and that copy — with the rest of the
chain and the content inside it — is what renders. A page built again by calling
the same constructor is a different value, whose layouts have empty `content`
slots. So, since v0.60.0, is a second `Build()` on the same builder: each call
returns an independent copy, and `WithPath` on the builder afterwards reaches
neither. Build once and keep the `*Page`.

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
  varies per request belongs in the data a handler returns or under a
  `collage.Key` — see [Data handlers](/docs/data-handlers#the-render-context).

`app.Page(name)` and `app.Pages()` return copies of registered pages, for looking
at them — a sitemap listing every page's paths, a test checking a strategy. A copy
is not the registered value, so do not pass one where a registered page is
expected.

## Example: a blog with a private area

A multi-user blog has three kinds of page. Posts are public. Each signed-in writer
has an account page listing their drafts. Admins also have a page listing the
users. The three share one shell, and the private ones sit in layouts that say who
may enter:

| Page | Path | Layouts, outermost first | Strategy |
| --- | --- | --- | --- |
| `post` | `/blog/{slug}` | `Master` | `Incremental(10 * time.Minute)` |
| `account` | `/account` | `Master`, `Account` | none declared: dynamic, it has a data handler |
| `admin-users` | `/admin/users` | `Master`, `Account`, `Admin` | none declared: dynamic, it has a data handler |
| `not-found`, `error` | no path | `Master` | `Dynamic()` |

The layouts carry the guards. `Account` uses the
[session plugin](/docs/plugins#elagohtsession)'s `RequireUser`, which sends a
signed-out reader to `/login` with where they were going in `next`. `Admin` checks
a role, under a session key `"role"` that this application's own login action
writes; the plugin defines only the user key.

```go
// Master is the site's shell: <head>, header, footer.
func Master() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("My blog").
		Build()
}

// Account is the signed-in area. Its guard sends everyone else to log in.
func Account() *collage.Fragment {
	return collage.NewFragment("account-layout", "layouts/account.html").
		WithGuard(session.RequireUser("/login")).
		Build()
}

// Admin sits inside Account and lets only admins through.
func Admin() *collage.Fragment {
	menu := collage.NewFragment("admin-menu", "fragments/admin-menu.html").Build()

	return collage.NewFragment("admin-layout", "layouts/admin.html").
		WithGuard(requireRole("admin")).
		WithSlotFragment("menu", menu).
		Build()
}

// requireRole lets through a reader whose session holds role under "role",
// which the login action sets. Anyone else gets a bare 403.
func requireRole(role string) collage.GuardFunc {
	return func(ctx context.Context, _ *http.Request) (*collage.GuardDecision, error) {
		if session.FromContext(ctx).Get("role") == role {
			return nil, nil
		}
		return &collage.GuardDecision{Status: http.StatusForbidden}, nil
	}
}
```

The admin layout calls both of its slots: `menu`, bound above, and `content`, which
registration fills. `layouts/account.html` is the same with only `content`.

```html
<!-- templates/layouts/admin.html -->
<div class="admin">
  <nav>{{slot "menu"}}</nav>
  <main>{{slot "content"}}</main>
</div>
```

A page is public or private by the layouts it lists, and nothing else:

```go
// Post is public and cached; an unknown slug is a 404.
func Post() *collage.Page {
	content := collage.NewFragment("post", "pages/post.html").
		WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) (store.Post, []string, error) {
			slug := rc.Param("slug")
			// PostBySlug wraps collage.ErrNotFound for a slug it does not know.
			post, err := store.PostBySlug(ctx, slug)
			return post, []string{"post:" + slug}, err
		})).
		Required().
		Build()

	return collage.NewPage("post").
		WithLayouts(layouts.Master()).
		WithContent(content).
		WithPath("en", "/blog/{slug}").
		Incremental(10 * time.Minute).
		Build()
}

// Account lists the signed-in reader's drafts.
func Account() *collage.Page {
	content := collage.NewFragment("account", "pages/account.html").
		WithData(collage.Load(func(ctx context.Context, rc *collage.RenderContext) ([]store.Draft, error) {
			return store.Drafts(ctx, session.Get(rc).Get(session.UserKey))
		})).
		Build()

	return collage.NewPage("account").
		WithLayouts(layouts.Master(), layouts.Account()).
		WithContent(content).
		WithPath("en", "/account").
		Build()
}

// AdminUsers is for admins only: it sits in both guarded layouts.
func AdminUsers() *collage.Page {
	content := collage.NewFragment("admin-users", "pages/admin-users.html").
		WithData(collage.Load(func(ctx context.Context, _ *collage.RenderContext) ([]store.User, error) {
			return store.Users(ctx)
		})).
		Build()

	return collage.NewPage("admin-users").
		WithLayouts(layouts.Master(), layouts.Account(), layouts.Admin()).
		WithContent(content).
		WithPath("en", "/admin/users").
		Build()
}

// NotFound and Error wrap the shell alone, so they look like the public site.
func NotFound() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").Build()

	return collage.NewPage("not-found").
		WithLayouts(layouts.Master()).
		WithContent(content).
		Dynamic().
		Build()
}

func Error() *collage.Page {
	content := collage.NewFragment("error-content", "pages/500.html").Build()

	return collage.NewPage("error").
		WithLayouts(layouts.Master()).
		WithContent(content).
		Dynamic().
		Build()
}
```

```go
func register(app *collage.App) error {
	if err := app.Register(
		pages.Post(),
		pages.Account(),
		pages.AdminUsers(),
	); err != nil {
		return err
	}
	if err := app.RegisterNotFoundPage(pages.NotFound()); err != nil {
		return err
	}
	return app.RegisterErrorPage(pages.Error())
}
```

What each reader gets:

| Request | Reader | Answer |
| --- | --- | --- |
| `/blog/hello` | anyone | The post; with caching on, from the cache until ten minutes have passed |
| `/blog/no-such-post` | anyone | `404` and the site's not-found page, in the shell |
| `/account` | signed out | `303` to `/login?next=%2Faccount` |
| `/admin/users` | signed out | `303` to `/login?next=%2Fadmin%2Fusers`: the account guard answers first, and the role is never checked |
| `/admin/users` | signed in, not an admin | `403`, with no body |
| `/admin/users` | an admin | The page |

A few things follow from the rules above:

- **No list of private paths.** A new page under `/admin` is private because it
  lists `layouts.Admin()`, not because of where its path starts. A page that
  forgets it is public, so `collage inspect` is worth a look: it lists, per page,
  the fragments that carry a guard.
- **The admin menu's guard would not be asked.** The menu is bound into a slot, not
  on the spine. Guarding the area is the admin layout's job.
- **Private pages are not shared caches.** The account and admin pages go out with
  `Cache-Control: private, no-cache`, the `303` and `403` with `no-store`, and a
  static export leaves both pages out.
- **The error pages render for anyone.** They wrap `Master()` alone, and the guards
  of the page that failed are not asked for them, so a reader who hit an error is
  never redirected to log in instead of seeing it.
