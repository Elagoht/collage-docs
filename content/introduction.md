---
description: Why collage exists, what it is built on, what it leaves to you, and when it is the right tool.
reference: NewPage, NewFragment, PageBuilder.WithFragmentPath, FragmentBuilder.WithGuard, App.InvalidateTags, App.PurgeTags, App.Use, App.Handle
---

# Introduction

collage is a Go framework for websites and web apps rendered on the server. You
write pages as a layout wrapped around small pieces called fragments; each
fragment fetches its own data and renders its own template, and collage assembles
the page, caches it, and knows exactly what to throw away when your content
changes. When one part of a page has to change in the browser, the server renders
that part again and sends it — nothing else.

It is one Go program. The same binary serves your site, answers its forms and API
calls, and — when you ask it to — writes the whole site out as static files. There
is no client framework, no bundler, and no dependency beyond the standard library.

## Why collage

Most websites are not applications. A landing page, a product page, a
documentation site, a site of a few pages with a form on one of them: they are
read far more than they are clicked. Yet the usual way to build one starts with a
JavaScript runtime on the server, a client framework, a bundler and a hydration
step — a runtime to install and keep running, a dependency tree to ship, and a
program sent to every browser so it can build HTML the server could have sent.
Such a runtime is slower than Go and keeps more memory and disk busy on the
server for the same pages.

The browser already does most of this work. HTML renders the page and submits the
forms. A few lines of plain JavaScript handle what is left. And when something on
the server changes, there is no need to turn the whole project into a single-page
app: `fetch` and `XMLHttpRequest`, server-sent events and WebSockets exist to bring
the changed part over and put it in place.

collage is built around that. The server holds the state and renders it as HTML,
one fragment at a time; the browser asks for the fragment it needs.

## Fragments, not components

Most pages are made of parts that do not know about each other. A blog post has
the post, an author card, a list of related posts, a navigation bar. Each part needs
different data, from a different place, and changes for different reasons.

In collage each of those parts is a **fragment**: a template, a function that fetches
what the template needs, and named **slots** where other fragments go. A **page** is
a layout fragment with a content fragment in it, and a URL.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithData(collage.LoadTagged(loadAuthor)).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithSlotFragment("author", author).
	Build()

page := collage.NewPage("post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()
```

Three things follow from building pages this way.

- **Fetching is concurrent, writing is in order.** Before a fragment renders, the
  data handlers of every fragment in its slots start at once. A page of independent
  parts makes its upstream calls together, and the HTML still comes out the same
  every time.
- **A page knows what it is made of.** Every data handler says which pieces of
  content it used — `post:hello-world`, `author:ada` — and the cached page carries
  those tags. When an author changes, one call invalidates every page that showed
  them, and nothing else.
- **A part can fail without taking the page with it.** A fragment can be required,
  or have a fallback, or simply render nothing. A broken sidebar is a missing
  sidebar, not a 500.

A fragment can also be opened at a URL of its own. That is how part of a page is
refreshed without a client framework: the server renders the fragment, and three
lines of JavaScript swap it in.

```go
collage.NewPage("search").
	WithLayouts(layout).
	WithContent(searchPage).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Build()
```

```js
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Only what is declared is reachable, and a fragment path meets the same guards as
its page. An action can answer a form with the one fragment that changed, too. The
details are in [A fragment at its own URL](/docs/forms-and-actions#a-fragment-at-its-own-url);
for fragments that refresh on an interval or when the server pushes a change,
there is [elagoht/live](/docs/plugins#elagohtlive).

## A page knows what it is made of

Every cached page, and every value cached with `collage.Cached`, carries the tags
of the content it was built from. When that content changes, name it:

```go
// An author edited their profile: every page that shows them renders again.
app.InvalidateTags(ctx, "author:ada")

// An account was deleted: nothing built from it may be served again.
app.PurgeTags(ctx, "user:42")
```

`InvalidateTags` marks the pages that used a tag stale and renders them again in
the background, keeping the old page to answer with if the new render fails.
`PurgeTags` removes them outright, from disk too, for content that must not be
shown even once more. Pages that did not use the tag are not touched. The whole
story — render strategies, the data cache, warming and stale pages — is in
[Caching](/docs/caching#invalidating).

## Middleware and plugins, as many as you need

A framework that gives you a single middleware file has you chain whatever a
request needs — authentication, redirects, headers, logging — inside it by
hand. In collage `app.Use` takes any standard
`func(http.Handler) http.Handler`, as many times as you like, and the first
registered is the outermost:

```go
app.Use(requestID)
app.Use(auditLog)
app.Use(loadUser)
```

Plugins bring their own middleware, along with template functions, pages, mounts
and hooks into rendering and caching, and run in the order you list them. The core
stays small on purpose; what a particular site needs is a plugin:

- **Forms and state:** validation, a honeypot against spam, flash messages,
  sessions, uploads.
- **Security:** security headers, rate limits, basic auth, OAuth, fail2ban, a
  consent banner.
- **Live updates:** fragments refreshed over server-sent events or a WebSocket, and
  a small script that swaps one in.
- **SEO and content:** meta tags, structured data, sitemaps, feeds, Open Graph
  images, Markdown, syntax highlighting, search, translations.
- **Delivery and operations:** minifying, compression, image optimisation, deploy
  targets, metrics, tracing, health checks, scheduled jobs and queues.

Each is a Go module you import; there is no loader and no registry. A plugin of
your own is a Go type that implements the hooks it needs. See
[Using plugins](/docs/plugins), [Writing plugins](/docs/writing-plugins) and
[Middleware and your own API](/docs/middleware-and-apis).

## Still Go

There is no second language and no second process. Your data layer is your Go
code — the database, the queue, the external API — called from data handlers and
actions like any other function. An API or a service you already have mounts on
collage's own mux, beside the pages:

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", users.Get)
api.HandleFunc("POST /api/users", users.Create)

if err := app.Handle("/api/", api); err != nil {
	return err
}
```

Any `http.Handler` will do: a `ServeMux`, a router you already use, a reverse proxy. It
can invalidate cached pages when it changes something, and it ships in the same
binary as the site. It brings its own middleware: the site's runs for it only when
it is mounted with `collage.WithMiddleware()`. See
[Your own handler](/docs/middleware-and-apis#your-own-handler-apphandle).

## Sites and web apps

collage is not only for pages that are read. What a server-rendered front end
needs is all here, so a multi-user, role-based web app is as much at home as a
landing page: [sessions](/docs/plugins#elagohtsession), forms that cannot be
forged, a [honeypot](/docs/plugins#elagohthoneypot) against spam, parts of the
page that update live, and guards that decide who may see a page before it is
rendered.

```go
func requireRole(role string) collage.GuardFunc {
	return func(ctx context.Context, r *http.Request) (*collage.GuardDecision, error) {
		if users.HasRole(ctx, role) {
			return nil, nil // allowed
		}
		return &collage.GuardDecision{Status: http.StatusForbidden}, nil
	}
}

admin := collage.NewFragment("admin", "layouts/admin.html").
	WithGuard(requireRole("admin")).
	Build()
```

Every page wrapped in the `admin` layout is for administrators, in every locale
and at every URL that reaches it, its fragment paths included. See
[Private pages](/docs/pages-and-layouts#private-pages-guards) and
[Forms and actions](/docs/forms-and-actions).

### What it leaves out

collage is not batteries-included. It has no ORM, no migrations and no database
layer, and sessions, translations and a JavaScript toolchain are not in the core.
You are writing Go, and Go already has good answers to all of those; a framework
that picked one for you would be one more thing to work around. What collage owns
is the part that is specific to the web front end: composing pages, caching them,
knowing when they are stale, and answering the forms on them. Sessions and
translations are plugins.

### When it is not the right tool

collage is a poor fit for screens whose state lives in the browser: a rich text
editor, a drawing canvas, an app that has to work offline. Use a client framework
for those, and collage — through
[`app.Handle`](/docs/middleware-and-apis#your-own-handler-apphandle) — for the
pages around them if you like.

## How a request flows

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 400" role="img" aria-labelledby="flow-title">
<title id="flow-title">A request passes through plugin middleware and app.Use, and routing sends it to a page, a fragment path, an action or your own handler. Only a page is cached, by tag; a fragment path and an action render fresh each time. All of them read from your data layer. A request for your own handler skips the middleware unless the handler is mounted WithMiddleware.</title>
<defs><marker id="flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="26" rx="13" /><text x="160" y="25" text-anchor="middle">browser</text>
<line class="d-flow" x1="160" y1="34" x2="160" y2="54" marker-end="url(#flow-arrow)" />
<rect class="d-frame" x="50" y="58" width="220" height="28" rx="6" /><text x="160" y="76" text-anchor="middle">plugin middleware</text>
<rect class="d-frame" x="50" y="92" width="220" height="28" rx="6" /><text x="160" y="110" text-anchor="middle">app.Use</text>
<line class="d-flow" x1="160" y1="120" x2="160" y2="140" marker-end="url(#flow-arrow)" />
<rect class="d-cmd" x="110" y="144" width="100" height="26" rx="13" /><text class="d-cmd-text" x="160" y="161" text-anchor="middle">routing</text>
<path class="d-flow" d="M130 170C130 185 42 182 42 198" marker-end="url(#flow-arrow)" />
<path class="d-flow" d="M150 170C150 185 118 182 118 198" marker-end="url(#flow-arrow)" />
<path class="d-flow" d="M170 170C170 185 194 182 194 198" marker-end="url(#flow-arrow)" />
<path class="d-flow" d="M190 170C190 185 270 182 270 198" marker-end="url(#flow-arrow)" />
<rect class="d-a" x="6" y="202" width="72" height="34" rx="5" /><text class="d-on-a" x="42" y="223" text-anchor="middle">page</text>
<rect class="d-b" x="82" y="202" width="72" height="34" rx="5" /><text class="d-on-b d-small" x="118" y="216" text-anchor="middle">fragment</text><text class="d-on-b d-small" x="118" y="228" text-anchor="middle">path</text>
<rect class="d-c" x="158" y="202" width="72" height="34" rx="5" /><text class="d-on-c" x="194" y="223" text-anchor="middle">action</text>
<rect class="d-frame" x="234" y="202" width="72" height="34" rx="5" /><text class="d-small" x="270" y="223" text-anchor="middle">app.Handle</text>
<line class="d-flow" x1="42" y1="236" x2="42" y2="250" marker-end="url(#flow-arrow)" />
<rect class="d-chip-hit" x="6" y="254" width="72" height="24" rx="12" /><text class="d-hit-text d-small" x="42" y="270" text-anchor="middle">cache · tags</text>
<line class="d-flow" x1="42" y1="278" x2="42" y2="294" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="118" y1="236" x2="118" y2="294" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="194" y1="236" x2="194" y2="294" marker-end="url(#flow-arrow)" />
<rect class="d-soft-a" x="6" y="298" width="224" height="30" rx="6" /><text x="118" y="317" text-anchor="middle">render: fragments fetch at once</text>
<line class="d-flow" x1="118" y1="328" x2="118" y2="346" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="270" y1="236" x2="270" y2="346" marker-end="url(#flow-arrow)" />
<rect class="d-frame" x="6" y="350" width="300" height="44" rx="6" /><text x="156" y="369" text-anchor="middle">your data layer, in Go</text><text class="d-small" x="156" y="384" text-anchor="middle">database · API · CMS · queue</text>
</svg>
<figcaption>Every request passes the middleware first, in the order it was registered; one for a handler of your own skips it unless that handler is mounted WithMiddleware. Routing then hands it to a page, one of a page's fragment paths, an action, or a handler of your own. A page is cached by the tags of what it was built from; a fragment path and an action render fresh each time. All of them read from your data layer, and your own handler talks to it directly.</figcaption>
</figure>

## What you get

- Pages cached per URL, with three strategies: rendered once and kept until you
  invalidate it, re-rendered after a TTL, or never cached. A page that fetches
  nothing is rendered once without being told to.
- Data cached across pages, so with the cache on, thirty posts by one author fetch
  the author once.
- Invalidation by tag, which renders the affected pages again in the background,
  and purging, which removes them outright.
- Fragments at their own URLs, for refreshing part of a page with a few lines of
  JavaScript.
- Forms that post to their own page, with request-forgery protection built in —
  and a cached page can still carry one.
- Guards that decide who may see a page, before anything is rendered.
- Links built from page names, so they follow a page when its path changes, in
  every locale the site has.
- Head elements — titles, meta tags, stylesheets — declared by the fragment that
  needs them, wherever it sits in the page.
- Non-HTML routes for sitemaps, feeds and `robots.txt`, middleware with `app.Use`,
  and your own `http.Handler` with `app.Handle` for anything else.
- A static export that turns the site into files for any static host. The pages
  you are reading were produced by it.
- A development server that rebuilds on every Go change and reloads the browser on
  every template change. A page that fails to render shows its error in the
  browser, template, line and cause first. Go that fails to compile shows it in the
  terminal while the last good build keeps serving, and a program that cannot
  start at all shows what it printed in the browser rather than a refused
  connection.

## Where to go next

[Installation](/docs/installation) gets a project running in a minute.
[Your first page](/docs/your-first-page) builds one from nothing, and
[Pages and layouts](/docs/pages-and-layouts) is where the concepts start.
