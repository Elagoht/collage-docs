---
description: The plugin contract, what Host and ConfigHost expose, every hook and what it may change, and a complete plugin with its tests.
reference: Plugin, Host, ConfigHost, ConfigReader, PluginConfig, Configurer, Command, BeforeRenderHook, BeforeActionHook, BeforeActionEvent, AfterRenderHook, CacheInvalidateHook, FragmentRequest, FragmentRender, HoistItem, StreamCloser, DrainHook, ServeHook, PageURL, FragmentReport, PathTag, Finding, FindingLevel, FindingWarning, FindingError, ErrBuildFindings, BuildFinishedHook, BuildFinishedEvent, BuiltFile, BuiltRedirect, RedirectSource, IsCapture, RequestHook, ErrStreamingBody, RouteInfo, Route, ClientIP, IsHTTPS, RouteKindPage, BuiltPage, StageErrorPage, Duration, Key, NewKey, RenderValues
---

# Writing a plugin

A plugin is a Go type with four methods. Everything else — reacting to a render,
rewriting output, adding a template function — is opt-in: you implement the
interface for the hook you want, and the framework finds it by type assertion.

This page is the reference for that surface. [Using plugins](/docs/plugins) is the
other side of it: how an application registers and configures what you write. To have `collage plugin add` install your plugin, meet [the repository contract](/docs/plugin-authors).

A plugin's life runs from the value an application constructs to its `Shutdown`.
Each stage below has a section of its own, and [Lifecycle](#lifecycle) lists them
in order.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 506" role="img" aria-labelledby="wp-life-title">
<title id="wp-life-title">The application constructs the plugin with New(Options) and puts it in Config.Plugins. collage.New runs Configure with a ConfigHost, in registration order: configuration, logger, template functions, mount wrappers. The application registers its pages, then the first Handler, Start or build runs Init with a Host, in registration order: it reads configuration and pages and adds documents, mounts, handlers, middleware and commands. While serving, ListenAndServe calls OnServe, every request passes the plugin's middleware and hooks, and invalidation and static builds fire OnCacheInvalidate and OnBuildFinished. Stopping calls OnDrain, then CloseStreams, then Shutdown in reverse order, only for plugins whose Init succeeded.</title>
<defs><marker id="wp-life-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="95" y="8" width="130" height="26" rx="13" /><text x="160" y="25" text-anchor="middle">New(Options{})</text>
<line class="d-flow" x1="160" y1="34" x2="160" y2="50" marker-end="url(#wp-life-arrow)" />
<text class="d-small" x="168" y="46">Config.Plugins</text>
<rect class="d-slot d-slot-a" x="6" y="54" width="308" height="74" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="70">inside collage.New</text><text class="d-small" x="302" y="70" text-anchor="end">registration order</text>
<rect class="d-a" x="18" y="78" width="284" height="40" rx="5" /><text class="d-on-a" x="160" y="94" text-anchor="middle">Configure(ctx, ConfigHost)</text><text class="d-on-a d-small" x="160" y="109" text-anchor="middle">config · logger · template funcs · WrapMount</text>
<line class="d-flow" x1="160" y1="128" x2="160" y2="144" marker-end="url(#wp-life-arrow)" />
<text class="d-small" x="166" y="140">app registers its pages</text>
<rect class="d-slot d-slot-b" x="6" y="148" width="308" height="88" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="164">first Handler, Start, build</text><text class="d-small" x="302" y="164" text-anchor="end">registration order</text>
<rect class="d-a" x="18" y="172" width="284" height="54" rx="5" /><text class="d-on-a" x="160" y="189" text-anchor="middle">Init(ctx, Host)</text><text class="d-on-a d-small" x="160" y="204" text-anchor="middle">config · logger · reads pages and URLs</text><text class="d-on-a d-small" x="160" y="217" text-anchor="middle">adds documents · mounts · Handle · Use · commands</text>
<line class="d-flow" x1="160" y1="236" x2="160" y2="252" marker-end="url(#wp-life-arrow)" />
<rect class="d-slot d-slot-c" x="6" y="256" width="308" height="130" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="272">serving</text><text class="d-small" x="302" y="272" text-anchor="end">hooks in registration order</text>
<rect class="d-frame" x="18" y="280" width="284" height="24" rx="5" /><text x="160" y="296" text-anchor="middle">OnServe · ListenAndServe only</text>
<rect class="d-b" x="18" y="312" width="138" height="64" rx="5" /><text class="d-on-b" x="87" y="329" text-anchor="middle">each request</text><text class="d-on-b d-small" x="87" y="344" text-anchor="middle">OnRequest · middleware</text><text class="d-on-b d-small" x="87" y="357" text-anchor="middle">render · action hooks</text><text class="d-on-b d-small" x="87" y="370" text-anchor="middle">OnPersonalise · OnError</text>
<rect class="d-c" x="164" y="312" width="138" height="64" rx="5" /><text class="d-on-c" x="233" y="329" text-anchor="middle">outside requests</text><text class="d-on-c d-small" x="233" y="348" text-anchor="middle">OnCacheInvalidate</text><text class="d-on-c d-small" x="233" y="364" text-anchor="middle">OnBuildFinished</text>
<line class="d-flow" x1="160" y1="386" x2="160" y2="402" marker-end="url(#wp-life-arrow)" />
<rect class="d-slot d-slot-a" x="6" y="406" width="308" height="92" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="422">stopping</text><text class="d-small" x="302" y="422" text-anchor="end">App.Shutdown</text>
<rect class="d-frame" x="17" y="432" width="86" height="28" rx="5" /><text x="60" y="450" text-anchor="middle">OnDrain</text>
<line class="d-flow" x1="103" y1="446" x2="115" y2="446" marker-end="url(#wp-life-arrow)" />
<rect class="d-frame" x="117" y="432" width="86" height="28" rx="5" /><text x="160" y="450" text-anchor="middle">CloseStreams</text>
<line class="d-flow" x1="203" y1="446" x2="215" y2="446" marker-end="url(#wp-life-arrow)" />
<rect class="d-a" x="217" y="432" width="86" height="28" rx="5" /><text class="d-on-a" x="260" y="450" text-anchor="middle">Shutdown</text>
<text class="d-small" x="160" y="476" text-anchor="middle">only plugins whose Init succeeded;</text>
<text class="d-small" x="160" y="489" text-anchor="middle">Shutdown once the server stops, in reverse</text>
</svg>
<figcaption>Your code constructs the plugin with New(Options) and hands it to Config.Plugins. collage.New runs its Configure before the templates are parsed, the one moment a template function can be added. The application then registers its pages, and the first call that starts it — Handler, ListenAndServe, Start, RenderPath, a static build — runs Init, which reads the configuration and the pages and adds documents, mounts, handlers, middleware and commands. An Init that fails stops startup, and the plugins initialised before it are shut down in reverse order. While serving, ListenAndServe calls OnServe once, every request passes the plugin's middleware and hooks, and invalidating tags or finishing a static build fires a hook of its own. Stopping calls OnDrain, then CloseStreams, waits for the server's requests, and calls Shutdown in reverse registration order, each only for plugins whose Init succeeded.</figcaption>
</figure>

## The contract

```go
type Plugin interface {
	Name() string
	Version() string
	Init(ctx context.Context, host collage.Host) error
	Shutdown(ctx context.Context) error
}
```

- **`Name`** identifies the plugin. It must be non-empty and unique within the
  application, and it is the key of the plugin's configuration section — so make
  it read like a module path: `acme/stamp`.
- **`Version`** is your plugin's own version, for diagnostics.
- **`Init`** runs once when the application starts, after the application has
  registered its own pages and before the first request.
- **`Shutdown`** releases whatever `Init` acquired. It is called once, and only for
  a plugin whose `Init` succeeded (since v0.61.0); keep it idempotent anyway — see
  [Lifecycle](#lifecycle).

Because hooks are discovered by type assertion, a misspelled hook method is not a
compile error — it is a hook that never runs. Assert every interface you mean to
implement:

```go
var (
	_ collage.Plugin          = (*Plugin)(nil)
	_ collage.AfterRenderHook = (*Plugin)(nil)
)
```

## Two phases: Configure and Init

Some work has to happen before templates are parsed. `html/template` can only call
a function that was in its function map when the template was parsed, and parsing
happens inside `collage.New`. So there is an optional earlier phase:

```go
type Configurer interface {
	Configure(ctx context.Context, host collage.ConfigHost) error
}
```

`Configure` runs inside `New`, once per plugin, in registration order, before the
templates are parsed. A returned error aborts `New`. Nothing has been acquired
yet, so there is no rollback.

`Init` runs later, when the application starts — the first call to `Handler`,
`ListenAndServe`, `Start`, `RenderPath`, `RenderDocumentPath` or `DispatchCommands`,
which includes a static build, since the builder renders through `RenderPath`. By then the
application has registered its pages, so a plugin can read them and add documents
and mounts of its own.

A plugin that needs both implements both. **A plugin that implements `Configurer`
must be supplied in `Config.Plugins`**: `RegisterPlugin` is called after `New` has
already parsed the templates, so it refuses such a plugin with
`ErrConfigurerRegisteredLate` instead of skipping its `Configure` silently.

### What each phase can reach

| | `ConfigHost` (Configure) | `Host` (Init) |
| --- | --- | --- |
| `DevMode`, `Logger`, `collage.PluginConfig(host, …)` | yes | yes |
| `AddTemplateFunc`, `AddRenderFunc`, `WrapMount` | yes | — |
| `collage.FindPlugin[T](host)` | yes | — |
| `Pages`, `Page` | — | yes |
| `URL`, `Locales`, `PageURLs`, `OriginFor`, `MultiOrigin` | — | yes |
| `BuildID`, `ServeStatus` | — | yes |
| `RegisterDocument`, `Mount`, `OpenMounted`, `Handle`, `Use` | — | yes |
| `RenderFragment` | — | yes |
| `RegisterCommand` | — | yes |

`ConfigHost` is narrower on purpose. During `Configure` the application has
registered nothing, so pages would be an empty list, and a site's origins are not
known until every plugin, an `OriginResolver` among them, is.

v0.78.0 narrowed both before v1, taking off what no plugin used:

- `Host.InvalidateTags` is gone. Invalidating is the application's —
  `App.InvalidateTags`, `ActionResult.InvalidateTags` — and a plugin hears of it
  with [`OnCacheInvalidate`](#cacheinvalidatehook).
- `ConfigHost.BaseURL`, `OriginFor` and `MultiOrigin` are gone. `Host` keeps all
  three: read them in `Init`.
- The `Host` given to `Init` lists no plugins, so `collage.FindPlugin` from `Init`
  reports false. Call it from `Configure` and keep what it finds.
- `ErrorEvent.Page` and `BeforeActionEvent.Page` are gone: a hook was handed the
  framework's live page, which nothing read.

### ConfigHost

| Method | What it does |
| --- | --- |
| `DevMode() bool` | Whether the application runs in development mode. |
| `Logger() *slog.Logger` | The application's logger. |
| `ConfigReader` (embedded) | What `collage.PluginConfig(host, defaults)` reads this plugin's configuration section through — see [Configuration](#configuration). |
| `AddTemplateFunc(name, fn) error` | Adds a template function. Returns `ErrDuplicateTemplateFunc` when the name was already added — by another plugin, or by this one earlier — or is a built-in's (since v0.77.0) or the template language's (since v0.78.0). |
| `WrapMount(wrap func(fs.FS) fs.FS)` | Registers a transformation applied to every mounted filesystem, in the order wrappers were registered. |
| `AddRenderFunc(name, factory) error` | Adds a template function made anew for each render from its `*RenderContext` — see [Template functions](#template-functions) (since v0.21.0). |

A `ConfigHost` closes when the plugin's `Configure` returns (since v0.78.0). A
plugin that kept it gets `collage.ErrConfigureClosed` from a later
`AddTemplateFunc` or `AddRenderFunc`, a later `WrapMount` wraps nothing and is
logged, and `collage.FindPlugin` through it finds nothing: templates are parsed
once construction finishes, and requests render with what was registered by then.
Before, a late render function was written into the map every render reads, which
the Go runtime aborts the process on.

### Host

| Method | What it does |
| --- | --- |
| `DevMode() bool` | Whether the application runs in development mode. |
| `Logger() *slog.Logger` | The application's logger. |
| `ConfigReader` (embedded) | What `collage.PluginConfig(host, defaults)` reads this plugin's configuration section through. |
| `Pages() []*collage.Page` | Every registered page, each a defensive copy. |
| `Page(name) (*collage.Page, bool)` | One page by name, a defensive copy. |
| `RegisterDocument(doc) error` | Registers a document the plugin contributes. |
| `Mount(prefix, fsys, opts...) error` | Serves a filesystem under a URL prefix. |
| `OpenMounted(urlPath) (fs.File, error)` | Opens one of the site's own files by the URL it is served at — `/static/icon.png` — through the mount whose prefix it is under: from disk in development, embedded in a built binary. A missing file is `fs.ErrNotExist` (since v0.68.0). |
| `Handle(prefix, handler) error` | Serves an `http.Handler` under a URL prefix ending in `/`, or at one exact path without it (`/metrics`, since v0.24.0), as `App.Handle` does — an event stream, a WebSocket (since v0.18.0). Unlike a plain `App.Handle` mount since v0.67.0, it runs inside the middleware, and since v0.78.0 its request body is bounded at `Server.MaxBodyBytes` before the middleware, as a page's is. |
| `RenderFragment(r, req) (*collage.FragmentRender, error)` | Renders a fragment a page opened with `WithFragmentPath`, in parts — see [Pushing fragments](#pushing-fragments) (since v0.18.0). |
| `RegisterCommand(cmd) error` | Contributes a command — see [Commands](#commands). |
| `Use(middleware) error` | Wraps every request (since v0.21.0), where the plugin was registered: for one in `Config.Plugins`, outside the application's own middleware (since v0.38.0; inside it before). Since v0.67.0 not a request for a handler the application mounted with plain `App.Handle`. |
| `URL(name, locale, params) (string, error)` | The path of a page or document, as `App.URL` builds it (since v0.21.0). |
| `BaseURL() string` | `Config.BaseURL`, without a trailing slash; empty when none is set. |
| `OriginFor(ctx, host) string` | The public origin of a host: an `OriginResolver`'s, else `Config.BaseURL` — see [Which origin a host has](#which-origin-a-host-has-originresolver) (since v0.60.0). |
| `MultiOrigin() bool` | Whether an `OriginResolver` is registered, so origins vary by host (since v0.60.0). |
| `Locales() (default, supported)` | The default locale and every supported one, the default included (since v0.21.0). |
| `PageURLs(ctx, name) ([]collage.PageURL, error)` | Every URL a page answers, in every locale, a pattern expanded through its `WithStaticParams` — a sitemap's contents (since v0.21.0). |
| `BuildID() string` | The build serving — `Config.Cache.Version`, or a fingerprint of the executable — for versioning what a browser keeps across deploys: a service worker's caches, an asset's query string (since v0.24.0). |
| `ServeStatus(w, r, status)` | Answers the request with the status and the site's own page for it — the not-found page for 404 and 410, the error page otherwise — for a plugin answering a request itself that should look like the site (since v0.24.0). |

v0.60.0 took `RegisterPage` and `FragmentURL` off `Host` — the application keeps
both, as `App.RegisterPage` and `App.FragmentURL` — and made the `collage.Origins`
capability, which a plugin reached by type assertion, the two methods `OriginFor`
and `MultiOrigin`.

What `Init` receives is not the `*App`. It is a narrow value that forwards these
methods and nothing else, so a plugin cannot assert its way to `ListenAndServe`,
`Shutdown`, the router, the cache or the template set.

`Handle` and
`RenderFragment` were added in v0.18.0, and `Use`, `URL`, `Locales`
and `PageURLs` in v0.21.0, with `AddRenderFunc` on `ConfigHost`, and `BuildID` and
`ServeStatus` in v0.24.0. **Each of these is a breaking change for a test double:**
one implementing `Host` or `ConfigHost` needs the new methods too — since v0.24.0,
`BuildID` and `ServeStatus`.

Since v0.50.0 both include `collage.ConfigReader`, whose
one method is unexported, so only collage's own hosts satisfy it; test a plugin
through a real application instead — see [Testing a plugin](#testing-a-plugin).

**`Host` limits what a plugin can reach, not what it can change.** `Pages` and
`Page` return copies of the page struct, of every slice and map it holds —
`Paths`, `CacheParams`, `FragmentPaths` and its per-locale maps, `DependencyTags`,
`LayoutChain` — and of its `Redirect` and `Action` values, so editing those does
not touch the application's own page (since v0.78.0; before, `CacheParams`,
`FragmentPaths` and the actions were shared, so a plugin writing its copy changed
the cache key and the fragment URLs). The fragments inside a copy are still
shared, and the events below
carry the *live* page, not a copy — copying a page and its fragment tree on every
request would cost the hot path.

Writing through an event's `Page` changes the page
every concurrent request is reading: that is a data race, and `go test -race` will
say so. Treat pages as read-only. Plugins are trusted code, not a sandbox.

Documents and mounts a plugin registers are held to the same rules as the
application's own: a name or a path that is already taken is a startup error, not
a race decided by registration order.

### Pushing fragments

`RenderFragment` is for a plugin that sends fragments over a connection it owns — an
event stream or a WebSocket — instead of waiting for the browser to ask for them. It
renders exactly what a request to the fragment's path renders, and returns the parts
instead of a response. One difference: values read through `collage.Cached` are
fresh. A request may be served a stale value while it is refreshed, but a pushed
fragment is not asked for again, so `RenderFragment` waits for the upstream, and
returns an error when it fails:

| Field | |
| --- | --- |
| `HTML` | The markup, with the reader's forgery token in any form it holds |
| `Head` | What the fragment hoisted into an area it placed no marker for, as `HoistItem`s with their area and key |
| `DependencyTags` | The tags the render depended on — match them against `CacheInvalidateEvent.Tags` to know what to push |
| `Shared` | The render is the same for every reader on the same host at this moment: the page is cached for everyone, or every handler in the subtree is declared `Static()` or `Shared()` and there is no slot resolver — and there is no form token, and (since v0.78.0) nothing a `PersonaliseHook` made personal. (since v0.19.0; before, a handler counted as reading the request unless it was `Static()`) Since v0.71.0 such a render is made from a request with the reader's cookies, headers and context values stripped, as the page cache's is |
| `ETag` | The ETag a request to the fragment's path would get for the same body, so a client can tell a pushed copy from one it already shows, or recognise the two as the same (since v0.19.0) |
| `Cookie` | The forgery cookie the forms in `HTML` need, when the request carried none |

A `FragmentRequest` names the fragment either by `Page`, `Fragment`, `Locale` and
`Params`, or by `Path` — the URL a page linked with `{{fragmentURL}}`, query
included — which is resolved as a request to it would be. A client subscribing to
the elements it shows only knows their URLs, so `Path` is usually what a stream
has.

```go
out, err := host.RenderFragment(r, collage.FragmentRequest{Path: "/live/cpu"})
```

Only fragments the page opened are rendered: a stream reaches exactly what HTTP
reaches. Since v0.78.0 each call is served as a request to the fragment's URL
would be: a `GET` to it carrying `r`'s host, client address, headers and cookies —
not its path values, body or method — passes the application's and every
plugin's middleware for the fragment's path, in the order a request does, then the
fragment's own [guard](/docs/pages-and-layouts#private-pages-guards) and the
page's (since v0.28.0), which read the fragment URL's parameters with
`r.PathValue`, never the caller's own.

- **A refusal is `collage.ErrGuardRefused`** instead of a render: a guard that
  refuses, or a middleware that answers instead of calling next — a basicauth
  401, a ratelimit 429, a login redirect, a timeout's 503. Send that connection
  nothing. What the middleware wrote, and the headers and cookies it set, are
  dropped; the context values it sets reach the guard and the render. A
  panicking middleware is an error wrapping `ErrPanic`.
- **The caller's context values do not carry over.** What the stream's own
  middleware put there was about the stream's path. The render ends when `r`
  does, and a trace span `r` carries is not its parent.
- **Every `PersonaliseHook` runs** on the markup and on each hoisted item, so a CSP
  nonce or a honeypot marker is filled in as on a page; a render a hook made
  personal is not `Shared`.
- **Every call is a request to the middleware**: a rate limit counts each push.
- Called from a plugin's `Init`, while the application is starting, it returns an
  error rather than render without the middleware.

By `Page` and `Fragment`, the URL is built from the page's pattern in the
requested locale, as `FragmentURL` builds it (since v0.78.0): a parameter no
request could carry — an undeclared name, a value with `/` outside a catch-all,
`.`, `..`, empty — a locale the fragment has no path in, and a path a request
would be redirected to its clean spelling from are `ErrUnknownFragmentPath`.

To ask each reader's middleware and guards about a change, call it once per
reader, with that reader's request, and send each connection the render made for
it. A render that is not `Shared` may hold one reader's data. A `Shared` one costs
one render per change however many readers it reaches (since v0.78.0): once a
reader's middleware and guards have passed, a `Shared` render already made for the
same host, URL, locale and cached query is handed over as a copy, and readers
arriving during the render wait for it. That reader's forgery token and
`PersonaliseHook` pass go into the copy, so a copy that took a token, or that a
hook made personal, is reported not `Shared`. It is made again once a tag it depended on — its URL's
`PathTag` and its host's among them — is invalidated or purged, after the page's
cache lifetime, and, for a page rendered per request, with the cache off or with a
store shared between instances, a second after it was made. At most 4096 renders
or 32 MiB (`Cache.MaxBytes`, if lower) are kept, none in development, and none for
a reader whose request skips the cache: a middleware's `SkipCache` makes the
render that reader's own, as over HTTP.
A fragment whose handler returns the same for everyone at one moment — a
measurement — is marked [`Shared()`](/docs/caching#a-page-that-declares-none) so
that its render counts, without making its page static.
`App.RenderFragment` is the same method, for an application's own code.

### Streams and shutdown

A plugin's `Shutdown` runs after the server has stopped, and the server stops by
waiting for every open request to end — which an event stream or a WebSocket never
does by itself. A plugin serving one implements `StreamCloser` (since v0.18.0):

```go
var _ collage.StreamCloser = (*Plugin)(nil)

func (p *Plugin) CloseStreams() { p.hub.close() }
```

`CloseStreams` runs when shutdown begins, before the server waits, and must end the
streams without waiting for them. A handler served through `Handle` can push its
write deadline forward with `http.NewResponseController(w).SetWriteDeadline`, and
take the connection over with `Hijack`, as it could on a bare `net/http` server.

A plugin that must know the server is about to stop taking traffic — readiness
turning false, new work refused — implements `DrainHook` (since v0.53.0):

```go
var _ collage.DrainHook = (*Plugin)(nil)

func (p *Plugin) OnDrain() { p.draining.Store(true) }
```

`OnDrain` is called once, when the drain starts: before
[`Server.DrainDelay`](/docs/deployment#graceful-shutdown-and-draining) is waited out
and before `CloseStreams`, and even when `DrainDelay` is 0. It must not block — the
server goes on serving while the drain lasts, and a slow hook delays every step
after it. A panic in it is contained and logged at Warn with the plugin's name.
Like `Shutdown` and `CloseStreams`, it reaches only plugins whose `Init`
succeeded (since v0.61.4).

A `Shutdown` before `Start`, or after a `Start` that
failed, tells none of them: before v0.61.4 it called `OnDrain` and `CloseStreams`
on plugins that never ran, so a plugin reused for a second start — as
`collagetest.TrySite` followed by `NewSite` does with the same plugin value — began
already draining.

A plugin that does work only while the application is serving — a scheduler, a
queue worker — implements `ServeHook` (since v0.55.0):

```go
var _ collage.ServeHook = (*Plugin)(nil)

func (p *Plugin) OnServe(ctx context.Context) { go p.run(ctx) }
```

`OnServe` is called once, in registration order, by `ListenAndServe`: after the
port is bound — when `collage: listening` is logged — and before it serves. It is
not called by `Start()` or `Handler()`, so not in a static build, not in a plugin
command, not by an application that runs its own `http.Server` over
`app.Handler()`, and not when `Shutdown` ran before `ListenAndServe`. `collage dev`
serves through `ListenAndServe`, so it is called in development too, again on every
restart.

`ctx` derives from `context.Background()` and is cancelled when the drain starts,
at the same moment `OnDrain` runs, whether the stop came from a signal or from
`App.Shutdown(ctx)`.

It means "start no new work"; finishing the work already
running belongs in the plugin's `Shutdown`, which runs after the server stops and
is bounded by the shutdown ctx. The ctx may already be done when `OnServe` is
called — a drain that started a moment earlier — so check it before starting work.

`OnServe` must not block: start goroutines and return. A panic in it is contained
and logged at Warn with the plugin's name, and the server still serves.

## The hooks

| Interface | Method | Event | Fires | May change |
| --- | --- | --- | --- | --- |
| `RequestHook` | `OnRequest` | the `*http.Request` | First on every request, before collage's request span, middleware and routing (since v0.25.0) | the request's context |
| `BeforeRenderHook` | `OnBeforeRender` | `BeforeRenderEvent` | Before a fresh page render | nothing on the event; may hoist through `ev.Context` |
| `BeforeActionHook` | `OnBeforeAction` | `BeforeActionEvent` | Before an action's handler, after its guards, body limit and forgery check (since v0.31.0) | `ev.Result`, which answers in the handler's place |
| `AfterRenderHook` | `OnAfterRender` | `AfterRenderEvent` | After a page render succeeded, with `ev.Values` (read with a key's `In`) | `ev.HTML`; reports with `ev.Warn`, `ev.Error` |
| `PersonaliseHook` | `OnPersonalise` | `PersonaliseEvent` | For every HTML response on its way to one reader, after the cache (since v0.43.0) | `ev.Body`, `ev.Header`, `ev.Personal` |
| `DocumentRenderedHook` | `OnDocumentRendered` | `DocumentRenderedEvent` | After a document handler produced its body | `ev.Body` |
| `CacheInvalidateHook` | `OnCacheInvalidate` | `CacheInvalidateEvent` | After entries were invalidated by tag | nothing |
| `OriginResolver` | `Origin` | the host, as a string | When `collage.BaseURL` or `Host.OriginFor` names a host's origin (since v0.42.0) | the origin returned |
| `ErrorHook` | `OnError` | `ErrorEvent` | On a failure while serving a request | nothing |
| `BuildFinishedHook` | `OnBuildFinished` | `BuildFinishedEvent` | Once, when a static build has written every file and captured its headers (since v0.21.0) | reports with `ev.Warn`, `ev.Error` |

On a request, the hooks run in this order:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 552" role="img" aria-labelledby="wp-req-title">
<title id="wp-req-title">A request runs OnRequest, then the plugins' middleware, then app.Use, then the router. A page looks in the cache: a hit goes straight to OnPersonalise; a miss runs OnBeforeRender, the render and OnAfterRender, and a GET is written to the cache before OnPersonalise. An action passes its guards and forgery check, runs OnBeforeAction and its handler, and an answer with a page runs the render hooks without the cache before OnPersonalise. A failure reaches OnError, and a document runs OnDocumentRendered.</title>
<defs><marker id="wp-req-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="wp-req-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">request</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="50" y="50" width="220" height="26" rx="5" /><text class="d-on-a" x="160" y="67" text-anchor="middle">OnRequest</text>
<line class="d-flow" x1="160" y1="76" x2="160" y2="88" marker-end="url(#wp-req-arrow)" />
<rect class="d-frame" x="50" y="92" width="220" height="26" rx="6" /><text x="160" y="109" text-anchor="middle">plugin middleware</text>
<rect class="d-frame" x="50" y="122" width="220" height="26" rx="6" /><text x="160" y="139" text-anchor="middle">app.Use</text>
<line class="d-flow" x1="160" y1="148" x2="160" y2="162" marker-end="url(#wp-req-arrow)" />
<rect class="d-cmd" x="110" y="166" width="100" height="24" rx="12" /><text class="d-cmd-text" x="160" y="182" text-anchor="middle">router</text>
<path class="d-flow" d="M135 190C135 206 84 206 84 222" marker-end="url(#wp-req-arrow)" />
<path class="d-flow" d="M185 190C185 206 238 206 238 222" marker-end="url(#wp-req-arrow)" />
<text class="d-small" x="88" y="204" text-anchor="end">page</text>
<text class="d-small" x="232" y="204">action</text>
<rect class="d-chip-hit" x="34" y="226" width="100" height="22" rx="11" /><text class="d-hit-text d-small" x="84" y="241" text-anchor="middle">cache lookup</text>
<path class="d-hitline" d="M34 237H12V445H22" marker-end="url(#wp-req-hit)" />
<text class="d-hit-text d-small" x="14" y="231">hit</text>
<line class="d-flow" x1="84" y1="248" x2="84" y2="262" marker-end="url(#wp-req-arrow)" />
<text class="d-small" x="90" y="259">miss</text>
<rect class="d-a" x="24" y="266" width="120" height="26" rx="5" /><text class="d-on-a" x="84" y="283" text-anchor="middle">OnBeforeRender</text>
<line class="d-flow" x1="84" y1="292" x2="84" y2="304" marker-end="url(#wp-req-arrow)" />
<rect class="d-soft-a" x="24" y="308" width="120" height="26" rx="6" /><text x="84" y="325" text-anchor="middle">render</text>
<line class="d-flow" x1="84" y1="334" x2="84" y2="346" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="24" y="350" width="120" height="26" rx="5" /><text class="d-on-a" x="84" y="367" text-anchor="middle">OnAfterRender</text>
<line class="d-flow" x1="84" y1="376" x2="84" y2="388" marker-end="url(#wp-req-arrow)" />
<rect class="d-chip-hit" x="29" y="392" width="110" height="22" rx="11" /><text class="d-hit-text d-small" x="84" y="407" text-anchor="middle">cache write · GET</text>
<line class="d-flow" x1="84" y1="414" x2="84" y2="428" marker-end="url(#wp-req-arrow)" />
<rect class="d-b" x="24" y="432" width="120" height="26" rx="5" /><text class="d-on-b" x="84" y="449" text-anchor="middle">OnPersonalise</text>
<rect class="d-frame" x="170" y="226" width="136" height="22" rx="5" /><text class="d-small" x="238" y="241" text-anchor="middle">guards · forgery check</text>
<line class="d-flow" x1="238" y1="248" x2="238" y2="262" marker-end="url(#wp-req-arrow)" />
<rect class="d-a" x="170" y="266" width="136" height="26" rx="5" /><text class="d-on-a" x="238" y="283" text-anchor="middle">OnBeforeAction</text>
<line class="d-flow" x1="238" y1="292" x2="238" y2="304" marker-end="url(#wp-req-arrow)" />
<rect class="d-soft-a" x="170" y="308" width="136" height="26" rx="6" /><text x="238" y="325" text-anchor="middle">action handler</text>
<line class="d-flow" x1="238" y1="334" x2="238" y2="346" marker-end="url(#wp-req-arrow)" />
<rect class="d-frame" x="170" y="350" width="136" height="46" rx="5" /><text class="d-small" x="238" y="364" text-anchor="middle">answers with a page:</text><text class="d-small" x="238" y="376" text-anchor="middle">render hooks</text><text class="d-small" x="238" y="388" text-anchor="middle">without the cache</text>
<line class="d-flow" x1="238" y1="396" x2="238" y2="428" marker-end="url(#wp-req-arrow)" />
<rect class="d-b" x="170" y="432" width="136" height="26" rx="5" /><text class="d-on-b" x="238" y="449" text-anchor="middle">OnPersonalise</text>
<path class="d-flow" d="M84 458C84 476 140 474 140 486" marker-end="url(#wp-req-arrow)" />
<path class="d-flow" d="M238 458C238 476 180 474 180 486" marker-end="url(#wp-req-arrow)" />
<rect class="d-pill" x="110" y="490" width="100" height="24" rx="12" /><text x="160" y="506" text-anchor="middle">response</text>
<line class="d-rule" x1="6" y1="524" x2="314" y2="524" />
<text class="d-small" x="160" y="542" text-anchor="middle">a failure: OnError · a document: OnDocumentRendered</text>
</svg>
<figcaption>OnRequest runs first, before collage's request span, then the middleware of the plugins in Config.Plugins, outside the application's app.Use. A page the cache holds, for GET or HEAD, skips the render hooks and goes straight to OnPersonalise. A fresh render runs OnBeforeRender and OnAfterRender, and is written to the cache by tag only for a GET whose render was not degraded; a page that is never cached has neither lookup nor write. An action runs OnBeforeAction after its guards, body limit and forgery check, and a plugin that sets ev.Result answers in the handler's place. An action that answers with its page runs the render hooks without the cache, and every HTML answer passes OnPersonalise; one that answers with a redirect reaches neither. A document runs OnDocumentRendered instead of the render hooks, and a failure while serving reaches OnError.</figcaption>
</figure>

Every hook method has the shape `func(ctx context.Context, ev *Event) error`, except
`OnRequest`. v0.60.0 removed `PageResolvedHook` and `CacheWriteHook`, which nothing
used: a request a plugin wants to see is `OnRequest`'s, and a render about to run
is `OnBeforeRender`'s.

### RequestHook

```go
type RequestHook interface {
	OnRequest(r *http.Request) (context.Context, func(status int))
}
```

Runs before anything else on a request (since v0.25.0): before collage starts its
request span, before middleware, before routing. It returns the context the request
is served under — derived from `r.Context()` — and a function collage calls with
the status when the response is written: once, unless serving the request panics.
The function may be `nil`, and it must not write to the response.

It is what a tracing plugin needs. A trace carried in from the caller has to be the
parent of collage's own `collage.http` span, and middleware added with `Host.Use`
runs too late for that, inside the span collage has already started:

```go
func (p *Plugin) OnRequest(r *http.Request) (context.Context, func(status int)) {
	ctx := p.propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := p.tracer.Start(ctx, r.Method, trace.WithSpanKind(trace.SpanKindServer))
	return ctx, func(status int) {
		if route := collage.RouteInfo(ctx); route.Kind != "" {
			span.SetName(r.Method + " " + route.Kind + ":" + route.Name)
		}
		span.SetAttributes(attribute.Int("http.response.status_code", status))
		span.End()
	}
}
```

`collage.RouteInfo(ctx)` reports what the request resolved to, as a `collage.Route`:
its `Kind` — `collage.RouteKindPage`, `RouteKindDocument`, `RouteKindAction`,
`RouteKindMount` or `RouteKindHandler` (since v0.61.0; the strings `"page"`,
`"document"`, `"action"`, `"mount"`, `"handler"`, and the set may grow), or
`RouteKindRejected` (`"rejected"`, since v0.79.0, below) — and its
`Name`, the registered name or prefix: the page's, the document's or the action's
name, the mount's or the handler's prefix. It is the zero `Route` while the request
is unresolved, and for a request that resolved to nothing, such as a 404 for a path
no route has. (Before v0.60.0, `collage.RouteOf(ctx)` returned the kind and the
name alone.)

It reads the context collage serves the
request under, which the finish function's context shares and
`Metrics.HTTPResponse` receives, so a span or a metric can be labelled with the
route rather than the raw path — a bounded set rather than one per URL.

`Route` carries two more fields: `Pattern`, the path
pattern the route was registered with (`/blog/{slug}`, without a locale prefix, or a
mount's or handler's prefix), and `Locale`. It is what to label with when the name
is not enough — a trace's `http.route` for a document as well as a page.

Hooks run in registration order, and each is handed the request under the context
the one before it made; the finish functions run in the reverse order. One that
panics is skipped, and the request is served under the context before it.

Since v0.47.0 the hooks also see a request collage answers before routing: a path
with an encoded slash (`/a%2f..%2fb`), answered with a 404, and a dirty path
(`/a/../b`), redirected with 301 (308 for a method other than `GET` and `HEAD`).
These are a scanner's probes more often than not.

For such a request the finish
function is called with the status written (once, unless serving the request
panics), and `r.URL` is the raw path. `collage.ClientIP(r)` works inside
`OnRequest`, so a plugin that bans or counts clients can see them.

Since v0.79.0, once the response is written, `collage.RouteInfo` reports such a
request with `Kind` `RouteKindRejected` (`"rejected"`) and `Name`
`RouteNameNormalise` (`"normalise"`), as it does the router's 404 for a path no
route can ever match (an encoded slash `/a%2fb`, a NUL, malformed
percent-encoding). A label built from `RouteInfo` stays bounded; one built from the
path, a scanner's, does not. `Metrics.HTTPResponse` and collage's own request span
see these requests too, with an empty path and no `http.path`. Before v0.79.0
`RouteInfo` was the zero `Route` for them, and the metrics and the span skipped
them.

Collage's own `collage.http` span records `http.method` bucketed since v0.79.0: the
standard methods as themselves, anything else a client sends as `OTHER`.

The hooks also see a static build's header capture (since v0.52.0): when a
`BuildFinishedHook` plugin is registered, the build asks the handler for each file
it wrote, and `collage.IsCapture(r.Context())` is true for those requests. A plugin
that counts or bans clients skips them. See
[BuildFinishedHook](#buildfinishedhook).

Use `collage.ClientIP(r)` rather than `r.RemoteAddr` for the client's address. It is
a `netip.Addr`: `RemoteAddr`'s host, or, behind the proxies the application lists in
`Server.TrustedProxies` (see [Deployment](/docs/deployment#behind-a-proxy-trustedproxies)),
the client the proxy reports. It is unmapped and has no zone, and is the zero
`netip.Addr` when there is no address to give: in a shared page render, whose
request has no `RemoteAddr`, and behind a trusted proxy whose `X-Forwarded-For`
names no usable client (`unknown`). Never count that one.

`collage.IsHTTPS(r)` (since v0.61.0) is whether the request arrived over TLS: on
its own connection, or at a proxy that says so in `X-Forwarded-Proto`. With
`Server.TrustedProxies` listed, the header is believed only from one of them, so a
client cannot claim a TLS connection it does not have; with none listed it is
believed from anyone. It is the decision the forgery guard makes for its own
cookie's `Secure` flag. A plugin setting `Secure` on a cookie, or sending
`Strict-Transport-Security`, uses it rather than reading the header itself.

### BeforeRenderHook

```go
type BeforeRenderEvent struct {
	Context *collage.RenderContext // the render about to run
	Page    *collage.Page
	Locale  string
	Path    string
	Static  bool // rendered for a static build, not for a request
}
```

Fires immediately before a fresh render and **not** on a cache hit. It fires after
the page's [guards](/docs/pages-and-layouts#private-pages-guards), and not at all for
a request one of them blocked. It fires for pages, for error pages, for a page
an action answers with through `RenderPage`, and for every page a static build
renders.

It is the only hook that gets the render context, and the reason is hoisting. A
plugin that contributes to the page declares before the tree renders, so that a
fragment can still override it:

```go
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.opts.Generator)
	return nil
}
```

A declaration made here sits at depth zero, so any fragment that declares the same
key replaces it: the plugin provides the default, the page provides the specific
thing. It lands only where the layout calls `{{hoist "head"}}`. An error fails the
request with a 500, under `"before_render"`.

`Static` (since v0.22.0) says the page is being rendered for a static build —
through `App.RenderPath` — rather than for a request. It is on `AfterRenderEvent`
too. A render for the build's header capture is a request, so `Static` is false
for it; `collage.IsCapture(ctx)` tells it apart.

### BeforeActionHook

```go
type BeforeActionEvent struct {
	Action  *collage.Action
	Locale  string
	Request *http.Request
	Result  *collage.ActionResult // set it to answer in the handler's place
}

func (ev *BeforeActionEvent) Form() (url.Values, error)
```

`Page` is gone since v0.78.0: nothing read it.

Fires once per action request, immediately before the handler: after the page's
guards, after the action's body limit is in place, and after the forgery check
(since v0.31.0). It is where a plugin checks a submission, a spam filter or a
quota, because the body there is bounded by the action's own limit, and a plugin
that read it earlier, in middleware, would need a limit of its own.

`ev.Form()` parses the form, URL-encoded or multipart, and returns its fields.
What it parsed stays parsed, so the handler reads the same form without the body
being read twice. Use it rather than `ParseMultipartForm` alone, which for a
URL-encoded body discards the error of reading it, a body past the limit among
them.

For an action built with `WithStreamingBody()` (since v0.57.0), `ev.Form()` reads
nothing and returns `collage.ErrStreamingBody`, so a plugin cannot consume the
stream the handler reads. A plugin that inspects forms must treat that error as
"this action has no form to check" and let the request through; one that returns
it fails every upload. A plugin that calls `ev.Request.ParseForm()` itself still
reads the body, and takes the stream from the handler.

```go
func (p *Plugin) OnBeforeAction(_ context.Context, ev *collage.BeforeActionEvent) error {
	form, err := ev.Form()
	if errors.Is(err, collage.ErrStreamingBody) {
		return nil // a streaming action: no form to check
	}
	if err != nil {
		return err // past the action's limit: 413
	}
	if form.Get("website") != "" {
		ev.Result = &collage.ActionResult{Status: http.StatusBadRequest}
	}
	return nil
}
```

Setting `ev.Result` answers the request with it, exactly as if the handler had
returned it, and stops dispatch: no later plugin and no handler runs. Since v0.79.0
that holds for all of it: its `InvalidateTags` are invalidated before the response
and its `Location` is warmed, as a handler's are, where before the tags were
dropped and nothing was warmed. One setting more than one of `Location`,
`Fragment`, `Page` and `Body` is refused with a 500, and `Location` and `Status` go
out as given. An error
fails the request under `"before_action"`, with `413` when it wraps
`*http.MaxBytesError` and `500` otherwise — or, since v0.57.0, `400` (`408` for a
read deadline) when the body was cut off while the plugin read it, logged at debug
and handed to no error hook.

### AfterRenderHook

```go
type AfterRenderEvent struct {
	Page     *collage.Page
	Locale   string
	Degraded bool   // some fragment failed, fallback or not, or stale cached data was used
	Static   bool   // rendered for a static build, not for a request
	HTML     []byte // replace it to post-process the page
	// Fragments: each fragment's time and failure, as collage.FragmentReport
	// DependencyTags: the tags the render depended on
	// Values: the render's shared values, read with a collage.Key's In
	// Findings: what ev.Warn and ev.Error reported so far
}
```

Fires after a page render succeeded, for pages, error pages, a page an action
answers with through `RenderPage`, and static builds.
Replace `ev.HTML` to post-process; what you leave there is what is served, and
what is cached. Later plugins see what earlier
ones produced.

`ev.Values` (a `*collage.RenderValues`) are the values the page's fragments
exchanged through keys, so they are what the page was built *from*, for a plugin
that wants the article rather than markup to parse back. A hook has no
`RenderContext`, so it reads them with the key's `In`:

```go
var ArticleKey = collage.NewKey[Article]("jsonld:article")

func (p *Plugin) OnAfterRender(_ context.Context, ev *collage.AfterRenderEvent) error {
	article, ok := ArticleKey.In(ev.Values)
	if !ok {
		return nil // this page stored no article
	}
	ev.Hoist("head", "jsonld", p.script(article))
	return nil
}
```

The values are opaque: a plugin reads only what it holds a key for, and there is no
listing them. A plugin that wants the application to hand it something exports the
key, and the application stores under it; a plugin that carries state from its own
render functions to its hook keeps the key unexported. The framework puts nothing
there. They are the render's own, not a copy: keeping them past the hook is holding
request state.

Begin a key's name with the plugin's own name — `"jsonld:article"`,
`"validate:errors"`. A key is only its name and its type, so two plugins that each
declare `NewKey[string]("errors")` hold one value between them, and whichever sets
it last wins. Before v0.50.0 this was `ev.Data`, the render's shared map.

`ev.Fragments` and `ev.DependencyTags` (since v0.24.0) report how the render went,
for a development tool to show beside the page. Each `collage.FragmentReport` has
the fragment's `Name`; its `Duration`, the fragments in its slots included;
`Failed`, even when a fallback stood in; `UsedFallback`; and `Err`, what it failed
with. `DependencyTags` are the tags the render depended on. Both are the event's
own copies.

Two consequences of where it sits:

- **It does not run again on a cache hit.** Its output is what was cached. A hook
  that must run per request cannot be combined with a cached page.
- **An empty result is a failure.** If `ev.HTML` is empty after dispatch, the
  request fails with a 500 rather than serving a blank page.

A page an action answers with through `RenderPage` runs it too (since v0.10.0;
before, it ran `BeforeRender` only), so a validation page is minified like any
other. An error fails the request with a 500, under `"after_render"`.

### Adding to the head after the render

A plugin that learns what a page needs only from its finished markup — the code
blocks it highlighted need a stylesheet — adds it with `ev.Hoist(area, key, html)`
in `OnAfterRender` (since v0.25.0):

```go
if highlighted {
	ev.Hoist("head", "highlight:css", `<link rel="stylesheet" href="/_highlight/style.css">`)
}
```

It lands where the layout put `{{hoist "head"}}`, after what the render declared
there. A key the render already declared in that area is left alone — the page is
more specific than a plugin afterwards — and so is a key hoisted twice. `ev.Hoist`
reports whether it added the HTML.

It finds the layout's place only in the HTML the render produced. If an earlier
plugin replaced `ev.HTML` — or you did, in the same hook — `"head"` lands before
`</head>` instead, and any other area reports false. Hoist first, then rewrite.

### Checking the output: findings

A plugin that checks what a page renders — a heading level skipped, an image
without `alt`, a form field without a label — reports what it finds rather than
failing the render (since v0.21.0):

```go
func (p *Plugin) OnAfterRender(_ context.Context, ev *collage.AfterRenderEvent) error {
	if !bytes.Contains(ev.HTML, []byte("<h1")) {
		ev.Error("one-h1", "the page has no <h1>")
	}
	return nil
}
```

`ev.Warn(rule, message)` reports a `collage.FindingWarning`, worth fixing and
stopping nothing; `ev.Error` a `collage.FindingError`. A finding
(`collage.Finding`) has its `Level`, the `Rule` that found it, a `Message`, the
`Plugin` and the page's `Path` — the last two filled in by the framework. Where it
goes depends on where the page was rendered:

- **In development** it is shown over the page, in the panel a failed fragment
  uses, and the page is served as it is.
- **In a static build** it is listed in the report under the page it is about
  (`BuildReport.Findings`). An error-level finding fails the build with
  `collage.ErrBuildFindings`; the pages are written either way. See
  [Static export](/docs/static-export#reading-the-report).
- **In production** nothing is done with it. A check re-run on every render would
  spend a server's time on what the build already knew, so a checking plugin turns
  itself off there: `ev.Static` says a render is a static build's, and
  `host.DevMode()` says the server is a development one.

```go
if !ev.Static && !p.dev { // p.dev from host.DevMode() in Init
	return nil
}
```

What no single render can tell — two pages with one title, a link to a page the
build did not write — belongs in [`OnBuildFinished`](#buildfinishedhook).
[elagoht/htmlcheck](/docs/plugins#elagohthtmlcheck) is a checking plugin built on
both.

### Rewriting each response: `PersonaliseHook`

`OnAfterRender` runs once, before the cache, so it cannot give each reader
something of their own. A plugin that must, such as a Content-Security-Policy nonce
or a per-visitor token, implements `PersonaliseHook` (since v0.43.0), called for
every HTML response on its way to one reader:

```go
func (p *Plugin) OnPersonalise(_ context.Context, ev *collage.PersonaliseEvent) error {
	if !bytes.Contains(ev.Body, []byte(p.marker)) {
		return nil
	}
	nonce, err := newNonce()
	if err != nil {
		return err
	}
	ev.Body = bytes.ReplaceAll(ev.Body, []byte(p.marker), []byte(nonce))
	ev.Header.Set("Content-Security-Policy", "script-src 'nonce-"+nonce+"'")
	ev.Personal = true
	return nil
}
```

The event carries:

- `ev.Request`: the reader's own request, not a shared render's stripped one.
- `ev.Header`: the response header; set here what must match the body.
- `ev.Body`: what will be written, the reader's forgery token already in it. A hook
  replaces it (assigns a new slice); it must never write into it: the slice may be
  the cache's, shared with concurrent readers.
- `ev.Personal`: set it when the body became particular to this reader.

It runs after the forgery token goes into the shared body, inside every
middleware, so before compression, and before the dev reload script. On a page it
runs before the dev overlay; on an error page, or a page an action answers with,
the overlay is already in the body.
Plugins run in registration order and each sees the previous body. It covers a
page (from the cache or fresh), a fragment path or fragment read, an action's HTML
answer, an error page and, since v0.78.0, what `RenderFragment` returns — the
markup and each hoisted item — never a document. Once a hook sets `Personal`, a
later one cannot clear it (since v0.78.0).

`Personal` treats the response as a forgery token does: its ETag is recomputed
from the body sent, a page answers `private, no-store`, and neither a page from
the cache nor a fragment read is answered `304`, not even to `If-None-Match: *`.

A fragment read answers `private, no-cache`
with the recomputed ETag and an action's HTML `private, no-store`, overriding even
a `Cache-Control` the handler set. A compressing middleware that keeps compressed
bodies per ETag, such as elagoht/compress, gains nothing from a personal
response: its ETag is new every time. elagoht/compress compresses it for each
reader and keeps nothing (since v0.1.3).

A body changed without `Personal`
still gets an ETag naming the bytes sent. A hook that returns an error or panics is
logged and the page, fragment or action answers 500; on an error page the built-in
page for that status is sent instead. A site without such a plugin sends the same
responses as before.

### DocumentRenderedHook

```go
type DocumentRenderedEvent struct {
	Document    *collage.Document
	ContentType string
	Locale      string
	Path        string
	Body        []byte // replace it to transform the document
}
```

`AfterRenderHook`'s counterpart for [documents](/docs/documents) — sitemaps, feeds,
JSON. Without it, a plugin that post-processes output would cover pages and
silently skip everything else. It fires before the ETag is computed and before
the body is cached, so what you produce is what is stored and what the ETag
describes. Check `ContentType` to decide whether a body is yours to touch. An empty
body after dispatch, or an error, fails the request with a 500.

A document does not dispatch `OnBeforeRender` or `OnAfterRender`:
it has no page and renders no templates.

### CacheInvalidateHook

```go
type CacheInvalidateEvent struct {
	Tags    []string
	Entries []InvalidatedEntry // the cached entries invalidated, each with its host, sorted
}

type InvalidatedEntry struct {
	Host string
	Path string
}
```

Fires after `InvalidateTags` or `PurgeTags` reached the entries for some tags —
whether the application or an action called it. It is dispatched from that call,
not from a request. Every plugin implementing it is called, even after one before
it failed or panicked (since v0.78.0; one failing plugin used to keep a CDN purge
or a live update from running), and each failure is logged and joined into what
`InvalidateTags` returns. A plugin does not invalidate: `Host.InvalidateTags` is
gone since v0.78.0.

`Entries` names the cached pages and documents the invalidation reached — what a
CDN has to purge, and a search engine be told has changed. On the built-in stores
an `InvalidateTags` entry was marked stale, not removed (since v0.63.0): it is no
longer served as fresh, and is served again only, marked `STALE`, when its
re-render fails or is degraded. A page that was not cached is never in it: nothing of it was
reached. Nor is a page an earlier invalidation already marked stale, until it is
rendered and stored again — except with a store a fleet shares, a
[`collage.SharedCache`](/docs/caching#a-cache-a-fleet-shares) (since v0.66.0): its
own tag index names every page any instance wrote, and keeps a page invalidated
and not yet re-rendered, so a second invalidation names it again.

Every cached entry also depends on the tag
`collage.PathTag(path)`, so an application that knows a path rather than a tag
invalidates what is cached there with
`app.InvalidateTags(ctx, collage.PathTag("/blog"))`. See
[Caching](/docs/caching#invalidating-by-path).

Each entry is a host and a URL path, sorted by host and then path. A path cached
under two hosts is two entries, and the host keeps its port if the request had one.
A plugin that purges a CDN or pings a search engine reads it to build each URL
against the right origin. (Paths came in v0.23.0 as `Paths`, hosts in v0.42.0 as
`Entries`; v0.60.0 kept `Entries` alone.) The host is recorded as a
`collage:host:<host>` dependency tag beside `collage:path:`; host tags are exempt
from the tracker's per-tag cap, and `MaxEntries` still bounds memory.

### Which origin a host has: `OriginResolver`

A plugin that serves several sites from one process, a host per customer, says
which public origin absolute URLs for a host are built against by implementing
`OriginResolver` (since v0.42.0):

```go
func (p *Plugin) Origin(ctx context.Context, host string) (string, bool) {
	origin, ok := p.origins[host]
	return origin, ok
}
```

The host arrives lower-cased with its port stripped, and an IPv6 literal without its
brackets (`::1`). The first plugin in registration order that returns `ok` wins. An
origin must be a bare `scheme://host[:port]` with scheme `http` or `https`, which
`collage.ParseOrigin` checks and normalizes, and must name a host (`https://:8080`
does not); an invalid one counts as not known, and dev mode logs it once per plugin
and origin.

A host no plugin knows gets `Config.BaseURL`, and a resolver that panics
is logged and skipped. Application code reads the result with `collage.BaseURL(rc)`,
which is safe in a cached render because the host is in the cache key, and a static
build, which has no request, gets `Config.BaseURL`. A site without a resolver sees no
change.

A plugin that builds absolute URLs itself asks the `Host` it was given in `Init`
(since v0.60.0; before, through the `collage.Origins` capability, by type
assertion). `ConfigHost` has no `BaseURL`, `OriginFor` or `MultiOrigin` since
v0.78.0: during `Configure` not every resolver is known.

```go
origin := host.OriginFor(ctx, "acme.test") // a resolver's, else Config.BaseURL
varies := host.MultiOrigin()               // is a resolver registered?
```

A plugin that takes origins in its own configuration validates them with
`collage.ParseOrigin` rather than a copy of the rule; what it refuses is
`collage.ErrInvalidOrigin` (`ErrInvalidBaseURL` before v0.60.0). Code that holds
only a context reads a header declared with `collage.Vary` through
`collage.Varied(ctx, header)`.

### ErrorHook

```go
type ErrorEvent struct {
	Err     error
	Path    string
	Stage   string
	Status  int           // the HTTP status the failure is answered with; 0 for none
	Request *http.Request // the reader's request, read-only; nil outside a request
}
```

Fires on a failure while serving a request: a page, a document, an action, a mount
or a handler registered with `App.Handle`. `Stage` names where it happened.

Since v0.78.0 the event carries no `Page`; read `Path` to tell what failed.

The stages the framework uses each have a constant (since v0.61.0), to compare
against instead of the literal:

| Stage | Constant |
| --- | --- |
| `"route"` | `collage.StageRoute` |
| `"not_found"` | `collage.StageNotFound` |
| `"guard"` (since v0.28.0: a page's guard failed or answered with a decision it cannot write) | `collage.StageGuard` |
| `"before_render"` | `collage.StageBeforeRender` |
| `"before_action"` | `collage.StageBeforeAction` |
| `"render"` | `collage.StageRender` |
| `"after_render"` | `collage.StageAfterRender` |
| `"cache_write"` | `collage.StageCacheWrite` |
| `"error_page"` | `collage.StageErrorPage` |
| `"asset"` | `collage.StageAsset` |
| `"plugin"` | `collage.StagePlugin` |
| `"handler"` | `collage.StageHandler` |
| `"panic"` | `collage.StagePanic` |

The set is not a closed
enum and may grow, so name the stages you handle rather than assuming you have seen
them all.

`"error_page"` is the one worth alerting on: it means the page that reports
failures itself failed, and the client still received a plausible built-in page,
so nobody would otherwise find out.

Classify `Err` with `errors.Is` — `collage.ErrNoRoute` for a URL that matched
nothing, `collage.ErrNotFound` for content that does not exist,
`collage.ErrMethodNotAllowed` for a 405, `collage.ErrCSRFMissing` and its siblings
for a refused submission, `collage.ErrAssetFailed` for a mount answering 4xx or
5xx, `collage.ErrPanic` for a recovered panic, and the rest listed in
[Errors](/docs/errors#reported-to-error-hooks).

A panic recovered while serving a request (stage `"panic"`) is a
`*collage.PanicError` (since v0.45.0). It still matches `errors.Is(err, collage.ErrPanic)`;
`errors.As` reaches the panic value and the stack it was raised on, and `errors.Is`
reaches the value itself when the code panicked with an error. The stack is no longer
in the error's message, only in a separate `stack` attribute of the framework's
`collage: request failed` log line, so code that matched the stack in the message
reads `PanicError.Stack` instead.

`Status` (since v0.45.0) is the HTTP status the failure is answered with: 404 for a
missing asset, 500 for a failed render or a panic, and 0 when no response is written
for it, as for a cache write that failed after the page was served, or a failed
render a stale page was served in place of (since v0.63.0). That one keeps the
stage it failed in, `"render"` say, and reaches `OnError` once per failed render,
however many readers were waiting on it and were served the stale page; see
[Caching](/docs/caching#stale-pages-and-x-collage-cache). An
`"error_page"` event always says 500.

`Request` is the reader's live request, for
reading only: its URL, method and headers, with a body that may already have been
read. It is `nil` when the failure did not come from a request. Read what you need
during `OnError`; do not keep it or hand it to another goroutine. Together they let
a plugin tell a 5xx from a 4xx and name the URL without a middleware of its own.

A page's failed cache write (`"cache_write"`) happens inside the render shared by
every reader waiting on the same cache key, so its `Request` is that shared
render's stripped request, not a particular reader's: no cookies, only the headers
middleware declared with `collage.Vary`, no `RemoteAddr`. A document's cache write
carries the request of the reader whose miss started the shared render.

An error returned from `OnError` is logged and swallowed, and the remaining
plugins still receive the event: an error handler that fails must not start another
round of error handling.

### BuildFinishedHook

```go
type BuildFinishedEvent struct {
	OutDir    string                  // the directory the build wrote into
	Files     []collage.BuiltFile     // every file it wrote, in no particular order
	Redirects []collage.BuiltRedirect // every redirect the site declares
	// Findings: what ev.Warn and ev.Error reported so far
}

type BuiltFile struct {
	Kind     string // collage.BuiltPage, BuiltDocument or BuiltAsset: "page", "document", "asset"
	Name     string // the page's or document's name; empty for an asset
	Locale   string
	Path     string // the URL path the file answers
	File     string // its absolute path on disk
	Captured bool        // the build asked the application for this path
	Status   int         // what it answered with; 0 when not captured or failed
	Headers  http.Header // the headers a static host can carry
}

type BuiltRedirect struct {
	From   string // "/old/{slug}", as registered
	To     string // "/new/{slug}", an absolute URL, or empty for a 410
	Status int    // 301, 302, 307, 308 or 410
	Source string // "page:<name>", "document:<name>" or a plugin's name
}
```

Fires once, when a static build has written every page, document and asset (since
v0.21.0), for checks across pages. Read a file with `os.ReadFile` when its content
is needed. `ev.Warn(path, rule, message)` and `ev.Error` report a finding against
the page at `path`, or against the build as a whole with an empty one.

An error
returned from the hook fails the build too, beside the files already written. It
never fires on a server, and a build whose context ended is not handed to it.

**Headers.** Since v0.52.0, before the hook runs, the build asks the application's
handler for each file's path twice and puts the answer on the file: `Status` and
`Headers`, without the headers about one response (`Date`, `ETag`,
`Set-Cookie`, …) and without any that differ between the two answers, such as a
CSP nonce. `Captured` is true for every file the build asked for, also one whose
capture failed (`Status` 0), and false for the `404.html` pages, which the build
makes itself. The root copy of the home page, under `PrefixDefault`, carries the
home page's capture.

The capture runs only when a
`BuildFinishedHook` plugin is registered. What it finds is reported as warnings;
see [Static export](/docs/static-export#headers-and-redirects).

**Redirects.** `ev.Redirects` holds every redirect the site declares: every page's
in registration order (`Source` `"page:<name>"`), every document's
(`"document:<name>"`), then every plugin's. `Status` is the redirect's
status, `302` for one that set none. A page registered only with `RegisterNotFoundPage` or
`RegisterErrorPage` is never matched, so its redirects are left out.

Together they are what a deploy adapter needs:
[elagoht/deploy](/docs/plugins#elagohtdeploy) writes them as a host's `_headers`,
`_redirects` or `vercel.json`.

**The capture is a request.** Middleware and the request hooks see it, and the page
renders again, with `OnBeforeRender`, `OnAfterRender` and `OnDocumentRendered`. A
plugin that counts or limits traffic, or writes files or keeps a tally from a render
hook, checks `collage.IsCapture(ctx)` and skips it. A capture request never writes
the response cache.

### RedirectSource

```go
type RedirectSource interface {
	Redirects() []collage.BuiltRedirect
}
```

A plugin that serves redirects of its own — from a file, a database — implements
`RedirectSource` (since v0.52.0), so a static export carries its rules to the host.
The build asks it once, after every file is written, and adds its rules to
`ev.Redirects` with `Source` set to the plugin's name. Each is checked as a
registered redirect is: `From` starts with one `/` and parses as a route pattern,
and every placeholder in `To` is captured by it. Unlike a page's redirect, `To` may
be an absolute `http` or `https` URL, and a `410` (a path that is gone) has no `To`;
301, 302, 307 and 308 need one.

Any other `Status` fails the build with
`collage.ErrInvalidRedirectStatus`, and a malformed rule, or one with a control
character, with `collage.ErrInvalidRedirect`, naming the rule and the plugin.
Since v0.78.0 a `From` or `To` holding a space, which a host's redirect file reads
as a field separator, is malformed, and each bad rule is reported, every plugin's,
where one used to drop every plugin's rules. A
plugin rule from the same path as another redirect, or over a written file, fails
the build like any other ([What fails](/docs/static-export#what-fails)).
[elagoht/redirects](/docs/plugins#elagohtredirects) is one.

### Dispatch rules

- Hooks run in **registration order**.
- Every call is **panic-guarded**. A panicking hook fails like one that returned an
  error; it does not take the process down.
- For `OnBeforeRender`, `OnAfterRender` and
  `OnDocumentRendered`, the **first error stops dispatch** and fails the request.
- For `OnPersonalise`, the first error stops dispatch; a page, fragment or action
  answers 500 and an error page falls back to the built-in page for its status.
- For `OnBeforeAction`, the first error stops dispatch and fails the request, and
  so does the first plugin to set `ev.Result`, which answers it.
- For `OnCacheInvalidate`, every plugin is called, even after one failed or
  panicked; the errors are logged and returned joined from `InvalidateTags` (since
  v0.78.0; before, the first error stopped dispatch).
- For `OriginResolver`, the first plugin that returns `ok` with a valid origin wins;
  a resolver that panics is logged and skipped.
- For `OnError`, errors are logged and dispatch continues.
- `OnRequest` returns no error. A panicking one is logged and skipped, and the
  request goes on under the context before it.
- For `OnBuildFinished`, the first error stops dispatch and fails the build.
- A finding is not an error: it never stops dispatch, and every later plugin still
  runs.

## Template functions

A plugin adds a template function from `Configure`:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	return host.AddTemplateFunc("readingTime", func(words int) string {
		return fmt.Sprintf("%d min read", max(1, words/200))
	})
}
```

Every template can then call `{{readingTime .Words}}`. The function is any value
`html/template` accepts in a function map.

- A name added twice — by two plugins, or twice by the same one — is
  `ErrDuplicateTemplateFunc`, returned from the second `AddTemplateFunc`. `New`
  fails if your `Configure` returns it, as it does when you pass the error on. The
  application cannot resolve it with `Template.Funcs`: the conflict is between
  plugins, and one of them has to give way.
- **The application wins** otherwise. An entry in `Config.Template.Funcs` under a
  name a plugin added replaces the plugin's function: the application can see both
  and decide, and `New` logs each such name once. Since v0.78.0 that holds for
  `AddRenderFunc` too, whose function used to be bound again on every render, after
  the application's, and so was the one that ran.
- **A built-in's name is not a plugin's to take.** Since v0.77.0 `AddTemplateFunc`
  refuses the name of any [built-in function](/docs/template-functions), `json` or
  `truncate` as much as `pageURL`, with `ErrDuplicateTemplateFunc`. Before, the
  plugin's function replaced the built-in without a word, under every template
  written against it, or never ran when the built-in is bound per render. Pick a
  name of your own; `Config.Template.Funcs` is where an application replaces a
  built-in.
- **Nor is a name of the template language.** Since v0.78.0 `AddTemplateFunc` and
  `AddRenderFunc` refuse Go's builtins (`eq`, `len`, `index`, `print`…) and
  escapers (`html`, `js`, `urlquery`, html/template's own) with
  `ErrDuplicateTemplateFunc`: a plugin's `html` was all the escaping a value piped
  through it got.
- It must be called from `Configure`. There is no way to add one later, because a
  function added after parsing is one no template can call; a call through a kept
  `ConfigHost` after `Configure` returned is `ErrConfigureClosed` (since v0.78.0).

`AddTemplateFunc`'s function is one value for the life of the application. Since
v0.21.0 `AddRenderFunc` takes a factory instead, called for each render with its
`*RenderContext`, so the function it returns can read what that render holds — a
nonce a `BeforeRender` hook set, the render's locale:

```go
var nonceKey = collage.NewKey[string]("csp:nonce")

host.AddRenderFunc("nonce", func(rc *collage.RenderContext) any {
	nonce, _ := nonceKey.Get(rc)
	return func() string { return nonce }
})
```

It follows `AddTemplateFunc`'s rules: it is called from `Configure`, and a name
another plugin added, or a built-in's, is `ErrDuplicateTemplateFunc`: a render
function cannot replace one of the framework's own, those listed on
[Template functions](/docs/template-functions).

## Wrapping mounts

`WrapMount` registers a function applied to every mounted filesystem:

```go
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	host.WrapMount(func(fsys fs.FS) fs.FS {
		return minifyingFS{inner: fsys} // your own fs.FS
	})
	return nil
}
```

It wraps the filesystem rather than the response because mounts serve through
`http.ServeContent`, which supports `Range`, `If-Range` and partial responses.
Changing bytes per response shifts every offset, and a range request would return
the wrong slice of a file whose advertised length no longer matches. A wrapper
whose files *are* the transformed ones keeps that arithmetic true.
Wrappers run in the order they were registered, and a `nil` wrapper is ignored.
It is called from `Configure`; since v0.78.0 a later call wraps nothing and is
logged.

## Contributing documents and mounts

From `Init`, a plugin can add routes of its own through `Host.RegisterDocument`
and `Host.Mount`, built with the same builders an application uses. Since v0.60.0
a plugin registers no pages: `Host.RegisterPage` is gone, and pages are the
application's. `Host.Handle` serves a plain `http.Handler` under a prefix, for
what is not a page — an event stream, a WebSocket. A prefix ending in `/` claims
every path beneath it; since v0.24.0 one without, `/metrics`, is a single exact
path. It always runs inside the middleware, the application's and every plugin's:
it is part of the site, and may read the session or rely on a basicauth in front of
it. Since v0.78.0 its request body is bounded at `Server.MaxBodyBytes` before the
middleware, as a page's is, so no middleware reads an unbounded body there;
`App.Handle` is unchanged.

`Host.Use` (since v0.21.0) wraps
every request, as `App.Use` does, where the plugin was registered. Since v0.38.0 a
plugin in `Config.Plugins` is outside the application's own middleware, so what it
puts in the request's context, a session, is there for the application's
middleware to read, and its headers and cookies wrap what that middleware
produced. One registered with `App.RegisterPlugin` sits among the application's
middleware where the call was made. Since v0.67.0 it does not run for a handler the
application mounted with plain `App.Handle`, which brings its own middleware: a
plugin that bans, limits, guards or logs cannot answer for such a mount. See
[No middleware, unless you ask](/docs/middleware-and-apis#no-middleware-unless-you-ask).

A plugin that links to pages asks where they live rather than guessing:
`Host.URL` builds a path as `App.URL` does, `Host.Locales` returns the default locale and every supported one, and
`Host.PageURLs(ctx, name)` lists every URL a page answers — one `collage.PageURL`
(`Locale`, `Path`, `Params`) per locale, and per parameter set a pattern's
`WithStaticParams` lists. A pattern without it has none a plugin could know. It is
what a sitemap is made of. `App.Locales` and `App.PageURLs` are the same, for an
application's own code.

A plugin that *reads* one of the site's own files — a source image, a stylesheet —
asks for it by the URL the site serves it at, with `Host.OpenMounted` (since
v0.68.0), rather than by a path on disk:

```go
f, err := host.OpenMounted("/static/icon.png")
if errors.Is(err, fs.ErrNotExist) {
	// no icon yet: fall back, and say where to put one
}
```

It goes through the application's mount, so it reads `static/icon.png` from disk in
development and the embedded copy in a built binary, and nothing has to be copied
beside the binary. Only what a request could reach opens: a dotfile, or a path
outside every mount, is `fs.ErrNotExist`. elagoht/favicon reads its source this
way.

A plugin that *produces files* — resized images, generated icons — should serve
them from a mount rather than a route. A static build copies every mount into its
output after every page has rendered, so a filesystem that records what the pages
asked for hands the builder exactly the right set, and the exported site needs
nothing running behind it. A document at a dynamic path cannot be enumerated that
way.

A plugin that answers a request itself — a page that is gone, a request it
refuses — can answer with the site's own page rather than a line of text:
`host.ServeStatus(w, r, http.StatusGone)` serves the not-found page for 404 and
410, and the error page for any other status (since v0.24.0).

A plugin versioning
what a browser keeps across deploys — a service worker's caches, an asset's query
string — reads `host.BuildID()`: `Config.Cache.Version` when it is set, or a
fingerprint of the executable. `App.BuildID` and `App.ServeStatus` are the same,
for an application's own code.

## Commands

A plugin contributes a command from `Init`:

```go
type Command struct {
	Name  string // as typed on the command line
	Usage string // for your program's own help; the framework never prints it
	Short string // one line
	Run   func(ctx context.Context, args []string) error
}
```

`RegisterCommand` refuses a name it could never run (since v0.78.0):

- an empty name;
- one that is not a single shell word — letters and digits of any script, `-` `_`
  `.` `:`, a letter or digit first; no space, no shell syntax — is
  `ErrInvalidCommandName`;
- a `collage-` name, which `DispatchCommands` answers itself (`collage-check`,
  `collage-inspect`), and a name another command already holds are
  `ErrDuplicateCommand`;
- so is, for a plugin's command, one of the `collage` binary's own names (`new`,
  `dev`, `build`, `export`, `serve`, `version`, `help`…): `go run . build` and
  `collage build` doing different things is a trap. An application's own command
  may still take one.

It stays open while every plugin's `Init` runs, and is closed with `ErrAppStarted`
once the application has started (since v0.78.0; before, it stayed open, and a
command registered then was never dispatched).

`Usage` and `Short` are data. Neither the framework nor the `collage` binary prints
them; a program that wants a help listing builds it from `app.Commands()`.

The `collage` CLI does not run plugin commands: it never loads your application.
The application's own `main` dispatches them with `collage.DispatchCommands`, and a
scaffolded `main.go` does so for any word left after the flags, so a user of your
plugin runs `go run . <command>`:

```go
flag.Parse()

// ... build app and register everything ...

if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
log.Fatal(app.ListenAndServe())
```

`DispatchCommands` starts the application first, because `Init` is what registers
the commands. The exit codes are:

- `0` for success;
- `1` for a startup failure, a command that ran and failed, or a command with no
  `Run`;
- `2` for a nil app, no arguments, or an unclaimed name (`ErrUnknownCommand`).

Say in your plugin's README that its commands
run through the application — a project scaffolded before v0.10.0 has to add that
block itself. See [The collage CLI](/docs/cli#plugin-commands).

## Configuration

A plugin reads its own section of `Config.PluginConfig` into a typed struct with
`collage.PluginConfig(host, defaults)`, which takes `Host` or `ConfigHost` alike and
returns the type of the defaults. It decodes the application's section over a copy
of them:

```go
type Options struct {
	Generator string           `json:"generator"`
	Disabled  bool             `json:"disabled"`
	Timeout   collage.Duration `json:"timeout"` // "5s", "1m30s"
}

func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts) // what New was given, overlaid by the application's section
	p.opts = opts
	return err
}
```

- **An absent, empty or `null` section returns the defaults unchanged**, so "not
  configured" and "configured to the zero value" stay different statements.
- **A present but malformed section is an error** naming the plugin and, when one
  member's value is what fails, its key — `collage: acme/stamp: "timeout": …` —
  returned with the defaults. The operator wrote something, and running on defaults
  silently would hide it. The section is decoded into a copy, so a failure leaves
  the defaults as they were.
- **A key your options do not have is an error** (since v0.61.0), at any depth,
  and so is one named after a field tagged `json:"-"`: `collage: acme/stamp:
  unknown key "generater" in plugins-config.json; remove it, or rename it to a key
  the plugin documents`. It stops the application at `Configure` or `Init`. Keys
  match fields case-insensitively, as `encoding/json` matches them. Decode the
  full `Options` type in one `PluginConfig` call: the check is made against the
  type it decodes into, and a type with its own `UnmarshalJSON` checks its own keys,
  or none. A renamed or removed key therefore fails loudly by itself.
- A scalar in the JSON replaces your default. **A top-level list replaces the
  default list whole** (since v0.61.2): its elements are new, so a field an element
  leaves out is that field's zero value, not the default element's. A top-level map
  adds its entries to a copy of yours. Neither writes into your slices or maps.
  Below the top level, a JSON object is decoded over the default's, setting only the
  fields it names, and a nested map is still filled in place: write the defaults as
  a literal rather than sharing one value between calls.
- **A duration is a `collage.Duration`** (since v0.61.0), not a `time.Duration`. It
  reads a string `time.ParseDuration` understands — `"10s"`, `"1h30m"` — and writes
  one back. A bare number is refused with an error naming its key, where
  `time.Duration` would have read `3600` as 3.6µs without a word, and so is a
  negative duration; `null` leaves the default. Convert it where you use it:
  `time.Duration(p.opts.Timeout)`.
- The application sees a key that names no registered plugin as a startup error
  (`ErrUnknownPluginConfig`). Your `Name` is therefore the whole of your
  configuration's address; changing it is a breaking change.

Before v0.50.0 a plugin set its defaults on a field and called `host.Config(&cfg)`.

Document every key, its type and its default in your README. Take the same
options in Go through one constructor, `func New(opts Options) *Plugin`, and decode
the section over them: an application then configures you in Go, in JSON, or both,
and `collage plugin add` writes `New(Options{})`, so the zero `Options` must be a
working plugin (or fail in `Init` with a message that says what is missing).

Call
the type `Options`, not `Config`, which reads as `collage.Config` at the call site.
Booleans default to their zero value: an opt-out is a `noX` key rather than a key
that defaults to true.

A secret never goes in `plugins-config.json`: a `<x>Env`
key names the environment variable that holds it, and a `json:"-"` field takes the
value from Go.

## Editor support: collage.json

Since v0.27.0 a plugin can describe itself to editors in a `collage.json` at its
module's root. The file ships in the module like any other, and an editor extension
finds it through the project's module graph — `go list -m -json all` — so a plugin
the project depends on is described whether it is published or not. The extension
offers what the file lists: the template functions with their signatures and
documentation, the attributes, the snippets, and the schema of the plugin's section
of `plugins-config.json`.

```json
{
  "$schema": "https://raw.githubusercontent.com/Elagoht/collage-snippets-highlighter/main/schemas/collage-plugin-manifest.schema.json",
  "name": "you/greeting",
  "description": "Greets the reader by name.",
  "templateFunctions": [
    {
      "name": "greet",
      "signature": "greet name",
      "insert": "greet ${1:.Name}",
      "doc": "\"Hello, name\" in the render's locale."
    }
  ],
  "attributes": [],
  "snippets": {
    "greeting": { "language": "html", "prefix": "cgreet", "body": ["<p>{{greet .Name}}</p>"], "description": "A greeting" }
  },
  "config": {
    "type": "object",
    "properties": { "formal": { "type": "boolean", "description": "Use the formal greeting." } }
  }
}
```

| Field | What it is for |
| --- | --- |
| `$schema` | The manifest's [JSON Schema](https://raw.githubusercontent.com/Elagoht/collage-snippets-highlighter/main/schemas/collage-plugin-manifest.schema.json), so an editor checks the file as you write it |
| `name` | The plugin's `Name()`, the key its configuration is found under in `plugins-config.json` |
| `description` | What the plugin does, in a sentence |
| `repository` | Where the plugin's documentation lives |
| `templateFunctions` | The functions the plugin adds with `AddTemplateFunc` or `AddRenderFunc`: each one's `name`, its `signature` as a template calls it, what completion inserts (`insert`, the name by default) and its `doc` in Markdown |
| `attributes` | HTML attributes the plugin reads, offered in tags: each one's `name`, the `value` completion inserts (`null` for a boolean attribute) and its `doc` |
| `snippets` | Snippets by name, each with its `language` (`html` or `go`), `prefix`, `body` and `description` |
| `config` | The JSON Schema of the plugin's section of `plugins-config.json` |

Every field but `name` is optional. `insert`, `value` and snippet bodies use VS
Code's snippet syntax. The editor learns which of those functions the application
actually has, and every page, fragment, slot and mounted file by name, from
[`collage inspect`](/docs/cli#collage-inspect).

## A complete plugin

`acme/stamp` names the generator in every page's head, offers the same name to
templates, adds a command that lists the pages, and reports a failing error page.
It uses both phases, one render hook, one error hook and a command.

```go
// Package stamp names the generator in every page's head, offers the same name
// to templates, and adds a command that lists the application's pages.
package stamp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key of its section in plugins-config.json.
const Name = "acme/stamp"

// Options is the plugin's configuration. The zero value is the defaults.
type Options struct {
	// Generator is what the generator meta tag says; "collage" when empty.
	Generator string `json:"generator"`
}

// Plugin is the stamp plugin. Construct it with New.
type Plugin struct {
	opts   Options
	logger *slog.Logger
}

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

var (
	_ collage.Plugin           = (*Plugin)(nil)
	_ collage.Configurer       = (*Plugin)(nil)
	_ collage.BeforeRenderHook = (*Plugin)(nil)
	_ collage.ErrorHook        = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure runs inside collage.New, before templates are parsed: the one
// moment a template function can still be added.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts) // the options New was given
	if err != nil {
		return err // a section that is malformed or names an unknown key
	}
	if opts.Generator == "" {
		opts.Generator = "collage"
	}
	p.opts = opts
	return host.AddTemplateFunc("generator", func() string { return p.opts.Generator })
}

// Init runs when the application starts, after it has registered its pages.
func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	p.logger = host.Logger()
	return host.RegisterCommand(collage.Command{
		Name:  "pages",
		Usage: "pages",
		Short: "List every registered page",
		Run: func(_ context.Context, _ []string) error {
			for _, page := range host.Pages() {
				fmt.Println(page.Name)
			}
			return nil
		},
	})
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBeforeRender declares the meta tag before the page renders, at depth zero,
// so a fragment that declares its own generator replaces this one.
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.opts.Generator)
	return nil
}

// OnError reports the one failure nobody would otherwise notice: the error page
// itself failing.
func (p *Plugin) OnError(_ context.Context, ev *collage.ErrorEvent) error {
	if ev.Stage == collage.StageErrorPage {
		p.logger.Error("stamp: the error page failed", "path", ev.Path, "error", ev.Err)
	}
	return nil
}
```

An application uses it like any other plugin:

```go
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{stamp.New(stamp.Options{})},
	PluginConfig: pluginConfig, // {"acme/stamp": {"generator": "The Wire"}}
})
```

## Lifecycle

1. **Registration.** `Config.Plugins` inside `New`, or `RegisterPlugin` before the
   application starts. After that, `RegisterPlugin` returns `ErrAppStarted` — also
   after any start that failed (since v0.12.0).
2. **Configure**, inside `New`, for plugins that implement it — in registration
   order, stopping at the first error.
3. **Init**, when the application starts, in registration order. If one fails,
   startup is aborted and every plugin already initialised is shut down in reverse
   order. The failing plugin is not, since it never finished initialising.
4. **OnServe**, for a `ServeHook`, once `ListenAndServe` has bound its port and
   before it serves, in registration order (since v0.55.0). Only `ListenAndServe`
   calls it; its ctx is cancelled when the drain starts — see
   [Streams and shutdown](#streams-and-shutdown).
5. **Shutdown**, from `App.Shutdown` — which `ListenAndServe` calls on `SIGINT` or
   `SIGTERM` — in reverse registration order.

   - Since v0.61.0 it calls `Shutdown`
     **once, and only for a plugin whose `Init` succeeded**: a plugin whose `Init`
     never ran or failed is not called, and one already shut down is not called
     again.
   - A start that fails *after* every `Init` succeeded — a mount that swallows
     a page, an error page never registered — shuts the plugins down there and then,
     as a failing `Init` rolls back the ones before it.
   - A `Shutdown` before `Start`
     initialises no plugin at all (since v0.61.1): `Start` then fails, and
     `ListenAndServe` returns nil. A `Start` after `Shutdown` fails whether or not
     plugins are registered (since v0.61.3).
   - Every plugin gets its turn even if one fails, and
     the errors are joined.
   - A plugin implementing `StreamCloser` has its
     `CloseStreams` called earlier, when shutdown begins, and a plugin implementing
     `DrainHook` has its `OnDrain` called before that, when the drain starts — both
     only for plugins whose `Init` succeeded (since v0.61.4); a `Shutdown` before
     `Start` tells none — see
     [Streams and shutdown](#streams-and-shutdown).

   With `ListenAndServe`, plugins are shut down after the drain's
   `Server.DrainDelay` and after the server has finished its requests, or once
   `Server.ShutdownTimeout` has passed if it has not — past the
   deadline a request may still be running.

   With a server you own, the `App` knows
   of none: start `ServeHook` plugins yourself, stop your server first, then call
   `App.Shutdown`, or a plugin can be torn out from under an in-flight request (see
   [Deployment](/docs/deployment#serving-with-your-own-server)).

   A static build or
   a plugin command serves nothing, so nothing calls `App.Shutdown` for it either:
   a scaffolded `main.go` calls it once the build or the command is done (since
   v0.55.0).

## Testing a plugin

Test a plugin the way an application would use it: in a real application, driven
through its handler. `collagetest.NewSite` (since v0.61.0) builds one with the
least configuration a test needs — a layout with `{{hoist "head"}}` and
`{{slot "content"}}`, a `home` page at `/`, a debug-level logger whose records it
keeps — puts your plugin in it, and starts it, so `Configure` and `Init` run at once
and their errors fail the test.

The application shuts down when the test ends. No
server listens and no port is chosen.

```go
package stamp_test

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"example.com/stamp"
)

// newSite builds a one-page application with the plugin in it.
func newSite(t *testing.T, options ...collagetest.SiteOption) *collagetest.Site {
	t.Helper()
	return collagetest.NewSite(t, append([]collagetest.SiteOption{
		collagetest.WithPlugins(stamp.New(stamp.Options{})),
		collagetest.WithTemplates(map[string]string{"home.html": `<p>{{generator}}</p>`}),
	}, options...)...)
}

func TestStamp_UsesItsDefaults(t *testing.T) {
	site := newSite(t)
	body := site.Get("/").WantStatus(http.StatusOK).Body
	if !strings.Contains(body, `<meta name="generator" content="collage">`) {
		t.Errorf("no default meta tag in %s", body)
	}
	if site.Logs.Has(slog.LevelError) {
		t.Errorf("the plugin logged an error:\n%s", site.Logs)
	}
}

func TestStamp_ReadsItsConfiguration(t *testing.T) {
	site := newSite(t, collagetest.WithPluginConfig(stamp.Name, `{"generator": "my site"}`))
	body := site.Get("/").WantStatus(http.StatusOK).Body
	if !strings.Contains(body, `<meta name="generator" content="my site">`) {
		t.Errorf("no configured meta tag in %s", body)
	}
	if !strings.Contains(body, "<p>my site</p>") {
		t.Errorf("template function not applied in %s", body)
	}
}

func TestStamp_RefusesAnUnknownKey(t *testing.T) {
	_, err := collagetest.TrySite(t,
		collagetest.WithPlugins(stamp.New(stamp.Options{})),
		collagetest.WithPluginConfig(stamp.Name, `{"generater": "typo"}`),
	)
	if err == nil || !strings.Contains(err.Error(), `unknown key "generater"`) {
		t.Errorf("err = %v, want the unknown key named", err)
	}
}

func TestStamp_RegistersItsCommand(t *testing.T) {
	commands := newSite(t).App.Commands()
	if len(commands) != 1 || commands[0].Name != "pages" {
		t.Errorf("commands = %v, want one named pages", commands)
	}
}

func TestStamp_RefusesLateRegistration(t *testing.T) {
	app, err := collage.New(&collage.Config{Template: collage.TemplateConfig{Root: t.TempDir()}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := app.RegisterPlugin(stamp.New(stamp.Options{})); !errors.Is(err, collage.ErrConfigurerRegisteredLate) {
		t.Errorf("RegisterPlugin: err = %v, want ErrConfigurerRegisteredLate", err)
	}
}
```

A `Site` is the started `App`, a `collagetest.Client` on its handler — so `Get`,
`Submit`, `Follow` and every other [client](/docs/testing#the-client) method work on it — and
`Logs`, every record the application wrote at debug level and above.

`Logs.Has(level, parts...)` reports whether a record at that level or above
contains every part, and it is safe to call while the application serves requests
from other goroutines. A test adds what its plugin needs:

| Option | Effect |
| --- | --- |
| `WithPlugins(plugins...)` | Adds plugins, in order |
| `WithPluginConfig(name, json)` | Sets one plugin's section, as `plugins-config.json` would. Invalid JSON fails the test |
| `WithTemplates(files)` | Adds or replaces templates by name under the root: `"layout.html"`, `"home.html"`, or one a registered page renders |
| `WithRegister(items...)` | Registers pages, documents and actions after the home page |
| `WithoutHome()` | Leaves the home page out, so `/` is the test's |
| `WithConfig(func(*collage.Config))` | Edits the configuration before the application is built: locales, a cache, `BaseURL`, development mode, anything |
| `WithSetup(func(*collage.App) error)` | Registers on the application itself before it starts — `Handle`, `Mount`, `Use`, `RegisterNotFoundPage`, `RegisterErrorPage` (since v0.61.1). Runs after the default registrations; several run in order; an error fails the site |

`collagetest.TrySite` is `NewSite` for a test that expects building or starting to
fail — a plugin refusing a bad configuration — and returns the error instead of
failing the test. When building the site fails — a `WithRegister` item, a
`WithSetup` or `Start` — it shuts the application down first (since v0.61.3), so a
plugin value can be used again for the next site.

A `collagetest.Site` can be
copied (`s := *site`); `go vet`'s copylocks check no longer objects.

A test that shuts the application down itself — to check what your `Shutdown`
returns — calls `site.Shutdown(ctx)` (since v0.61.1), not `site.App.Shutdown(ctx)`.
`App.Shutdown` runs once and returns its first error on every later call, so the
cleanup would report the error the test expected; `site.Shutdown` closes the site
and the cleanup does nothing more.

A few things worth testing that are easy to forget:

- **The unconfigured case.** Most applications will never write a section for you.
- **Documents**, if you implement `OnDocumentRendered`: register a document and
  request it, since a plugin tested only against pages never sees one. If you
  implement `OnError`, request a document that fails too.
- **Run with `-race`.** A hook that writes through an event's live `Page` is a data
  race the detector reports and nothing else will.
- **A static export**, if you produce files: `collage.Build(ctx, app, ...)` into a
  `t.TempDir()` shows whether the exported site has what the served one had. See
  [Testing](/docs/testing) and [Static export](/docs/static-export).

## Example: a catalogue's referrer policy, from scratch

A product catalogue links out to its suppliers, a payment provider and review
sites. Its URLs say what a shopper was looking at —
`/products/{slug}?colour=red&campaign=autumn` — and the team wants every page to
tell the browser, in a `Referrer-Policy` header, how much of that it may pass on.
The policy must be the same on the server and in the exported copy of the site,
and the footer names the address to report a vulnerability to.

`acme/policy` is a plugin written for that, small but complete:

- **Options**, set in Go or in `plugins-config.json`: the policy and the contact.
- **Middleware** added in `Init` with `Host.Use`, which sets the header on every
  response.
- **A template function**, `{{securityContact}}`, added in `Configure`.
- **A build finding** from `OnBuildFinished`: a page the static build captured
  without the configured header.

A site that wants the whole set of security headers uses
[elagoht/secure](/docs/plugins#elagohtsecure), which sends `Referrer-Policy` among
them. This one stays small so that each part of a plugin shows.

### The plugin

```go
// Package policy sends a Referrer-Policy header on every response, offers the
// site's security contact to templates, and checks a static build for pages
// that went out without the header.
package policy

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/Elagoht/collage/pkg/collage"
)

// Name is the plugin's name, and the key of its section in plugins-config.json.
const Name = "acme/policy"

// DefaultReferrerPolicy is what is sent when Options.ReferrerPolicy is empty.
const DefaultReferrerPolicy = "strict-origin-when-cross-origin"

// policies are the values a Referrer-Policy header may hold.
var policies = []string{
	"no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin",
	"same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url",
}

// Options is the plugin's configuration. The zero value is a working plugin.
type Options struct {
	// ReferrerPolicy is the header's value; DefaultReferrerPolicy when empty.
	ReferrerPolicy string `json:"referrerPolicy"`
	// Contact is where to report a vulnerability, offered as {{securityContact}}.
	Contact string `json:"contact"`
}

// Plugin is the policy plugin. Construct it with New.
type Plugin struct {
	opts Options
}

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

var (
	_ collage.Plugin            = (*Plugin)(nil)
	_ collage.Configurer        = (*Plugin)(nil)
	_ collage.BuildFinishedHook = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return Name }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure reads the options and adds the template function: it runs inside
// collage.New, before the templates are parsed.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts)
	if err != nil {
		return err
	}
	if opts.ReferrerPolicy == "" {
		opts.ReferrerPolicy = DefaultReferrerPolicy
	}
	if !slices.Contains(policies, opts.ReferrerPolicy) {
		return fmt.Errorf("%s: %q is not a referrer policy", Name, opts.ReferrerPolicy)
	}
	p.opts = opts
	return host.AddTemplateFunc("securityContact", func() string { return p.opts.Contact })
}

// Init adds the middleware. A plugin in Config.Plugins wraps the application's
// own middleware, so a header the application sets wins over this one.
func (p *Plugin) Init(_ context.Context, host collage.Host) error {
	return host.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Referrer-Policy", p.opts.ReferrerPolicy)
			next.ServeHTTP(w, r)
		})
	})
}

func (p *Plugin) Shutdown(context.Context) error { return nil }

// OnBuildFinished checks what a static host will be told to send: the build
// asked the application for every file it wrote, and kept the headers.
func (p *Plugin) OnBuildFinished(_ context.Context, ev *collage.BuildFinishedEvent) error {
	if p.opts.Contact == "" {
		ev.Warn("", "no-contact", "no security contact is configured; set contact in plugins-config.json")
	}
	for _, f := range ev.Files {
		if f.Kind != collage.BuiltPage || f.Status < 200 || f.Status > 299 {
			continue // not a page, or not captured with a 2xx
		}
		if got := f.Headers.Get("Referrer-Policy"); got != p.opts.ReferrerPolicy {
			ev.Error(f.Path, "referrer-policy", fmt.Sprintf("Referrer-Policy is %q, want %q", got, p.opts.ReferrerPolicy))
		}
	}
	return nil
}
```

Where each part sits follows from the sections above:

- **The options are checked in `Configure`.** A policy the header cannot hold
  fails `collage.New`, as an unknown key does, rather than going out on every
  response.
- **The middleware sets the header before calling `next`.** A plugin in
  `Config.Plugins` wraps the application's own middleware, so a header the
  application's middleware or a handler sets afterwards replaces this one: the
  application has the last word.
- **The check reads `f.Headers`**, which a build fills in only because a
  `BuildFinishedHook` plugin is registered, so it costs a server nothing. A file
  the build did not capture, or that was not answered with a 2xx, is skipped.
- **A missing contact is a warning, a wrong header an error.** The warning is
  listed in the report; the error fails the build with `collage.ErrBuildFindings`,
  and the pages are written either way.

### Registered in the application

The catalogue's `plugins-config.json`:

```json
{
  "acme/policy": {
    "referrerPolicy": "same-origin",
    "contact": "security@catalogue.example"
  }
}
```

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	log.Fatal(err)
}
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{policy.New(policy.Options{})},
	PluginConfig: pluginConfig,
})
if err != nil {
	log.Fatal(err)
}
```

The footer template then shows the contact:

```html
<footer>Found a vulnerability? Write to {{securityContact}}.</footer>
```

### Its tests

```go
package policy_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"example.com/policy"
)

// newSite builds a one-page application with the plugin in it.
func newSite(t *testing.T, options ...collagetest.SiteOption) *collagetest.Site {
	t.Helper()
	return collagetest.NewSite(t, append([]collagetest.SiteOption{
		collagetest.WithPlugins(policy.New(policy.Options{})),
		collagetest.WithTemplates(map[string]string{"home.html": `<footer>{{securityContact}}</footer>`}),
	}, options...)...)
}

func TestPolicy_SendsTheDefault(t *testing.T) {
	res := newSite(t).Get("/").WantStatus(http.StatusOK)
	if got := res.Header.Get("Referrer-Policy"); got != policy.DefaultReferrerPolicy {
		t.Errorf("Referrer-Policy = %q, want %q", got, policy.DefaultReferrerPolicy)
	}
}

func TestPolicy_ReadsItsConfiguration(t *testing.T) {
	site := newSite(t, collagetest.WithPluginConfig(policy.Name,
		`{"referrerPolicy": "no-referrer", "contact": "security@example.com"}`))
	res := site.Get("/").WantStatus(http.StatusOK)
	if got := res.Header.Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("Referrer-Policy = %q, want no-referrer", got)
	}
	if !strings.Contains(res.Body, "<footer>security@example.com</footer>") {
		t.Errorf("no contact in %s", res.Body)
	}
}

func TestPolicy_RefusesAnUnknownPolicy(t *testing.T) {
	_, err := collagetest.TrySite(t,
		collagetest.WithPlugins(policy.New(policy.Options{})),
		collagetest.WithPluginConfig(policy.Name, `{"referrerPolicy": "never"}`),
	)
	if err == nil || !strings.Contains(err.Error(), "not a referrer policy") {
		t.Errorf("err = %v, want the policy refused", err)
	}
}

func TestPolicy_BuildFindsAnOverriddenHeader(t *testing.T) {
	site := newSite(t,
		collagetest.WithPluginConfig(policy.Name, `{"contact": "security@example.com"}`),
		// The application's own middleware runs inside the plugin's, so its
		// header replaces the plugin's.
		collagetest.WithSetup(func(app *collage.App) error {
			return app.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Referrer-Policy", "unsafe-url")
					next.ServeHTTP(w, r)
				})
			})
		}),
	)
	report, err := collage.Build(context.Background(), site.App, collage.BuildOptions{OutDir: t.TempDir()})
	if !errors.Is(err, collage.ErrBuildFindings) {
		t.Fatalf("Build: err = %v, want ErrBuildFindings", err)
	}
	var found bool
	for _, f := range report.Findings {
		if f.Rule == "referrer-policy" && f.Path == "/" && f.Level == collage.FindingError {
			found = true
		}
	}
	if !found {
		t.Errorf("findings = %+v, want referrer-policy on /", report.Findings)
	}
}

func TestPolicy_BuildWarnsWithoutAContact(t *testing.T) {
	report, err := collage.Build(context.Background(), newSite(t).App, collage.BuildOptions{OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Build: %v", err) // a warning fails nothing
	}
	if len(report.Findings) != 1 || report.Findings[0].Rule != "no-contact" || report.Findings[0].Level != collage.FindingWarning {
		t.Errorf("findings = %+v, want one no-contact warning", report.Findings)
	}
}
```

Each test covers one promise of the plugin:

| Test | What it shows |
| --- | --- |
| `SendsTheDefault` | The zero `Options` is a working plugin |
| `ReadsItsConfiguration` | The section reaches both the header and the template function |
| `RefusesAnUnknownPolicy` | A bad value fails `collage.New`, and `TrySite` returns the error |
| `BuildFindsAnOverriddenHeader` | The application's own middleware, registered with `WithSetup`, runs inside the plugin's and wins; the build reports the page and fails with `ErrBuildFindings` |
| `BuildWarnsWithoutAContact` | A warning about the whole build has an empty `Path` and fails nothing |
