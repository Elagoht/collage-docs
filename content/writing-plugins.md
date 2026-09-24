---
description: The plugin contract, what Host and ConfigHost expose, every hook and what it may change, and a complete plugin with its tests.
---

# Writing a plugin

A plugin is a Go type with four methods. Everything else — reacting to a render,
rewriting output, adding a template function — is opt-in: you implement the
interface for the hook you want, and the framework finds it by type assertion.

This page is the reference for that surface. [Using plugins](/docs/plugins) is the
other side of it: how an application registers and configures what you write.

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
- **`Shutdown`** releases whatever `Init` acquired. It is called for every
  registered plugin whether or not its `Init` ran or succeeded, and possibly more
  than once, so it must be safe without `Init` and idempotent — see
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
application has registered its pages, so a plugin can read them or add its own.

A plugin that needs both implements both. **A plugin that implements `Configurer`
must be supplied in `Config.Plugins`**: `RegisterPlugin` is called after `New` has
already parsed the templates, so it refuses such a plugin with
`ErrConfigurerRegisteredLate` instead of skipping its `Configure` silently.

### What each phase can reach

| | `ConfigHost` (Configure) | `Host` (Init) |
| --- | --- | --- |
| `DevMode`, `Logger`, `Config` | yes | yes |
| `AddTemplateFunc`, `WrapMount` | yes | — |
| `Pages`, `Page`, `InvalidateTags` | — | yes |
| `RegisterPage`, `RegisterDocument`, `Mount` | — | yes |
| `RegisterCommand` | — | yes |

`ConfigHost` is narrower on purpose. During `Configure` the application has
registered nothing, so pages would be an empty list and invalidation would have no
cache to reach.

### ConfigHost

| Method | What it does |
| --- | --- |
| `DevMode() bool` | Whether the application runs in development mode. |
| `Logger() *slog.Logger` | The application's logger. |
| `Config(v) error` | Decodes this plugin's configuration section into `v` — see [Configuration](#configuration). |
| `AddTemplateFunc(name, fn) error` | Adds a template function. Returns `ErrDuplicateTemplateFunc` when the name was already added — by another plugin, or by this one earlier. |
| `WrapMount(wrap func(fs.FS) fs.FS)` | Registers a transformation applied to every mounted filesystem, in the order wrappers were registered. |

### Host

| Method | What it does |
| --- | --- |
| `DevMode() bool` | Whether the application runs in development mode. |
| `Logger() *slog.Logger` | The application's logger. |
| `Config(v) error` | Decodes this plugin's configuration section into `v`. |
| `Pages() []*collage.Page` | Every registered page, each a defensive copy. |
| `Page(name) (*collage.Page, bool)` | One page by name, a defensive copy. |
| `InvalidateTags(ctx, tags...) error` | Drops every cached entry built from any of the tags. |
| `RegisterPage(page) error` | Registers a page the plugin contributes. |
| `RegisterDocument(doc) error` | Registers a document the plugin contributes. |
| `Mount(prefix, fsys, opts...) error` | Serves a filesystem under a URL prefix. |
| `RegisterCommand(cmd) error` | Contributes a command — see [Commands](#commands). |

What `Init` receives is not the `*App`. It is a narrow value that forwards these
methods and nothing else, so a plugin cannot assert its way to `ListenAndServe`,
`Shutdown`, the router, the cache or the template set.

**`Host` limits what a plugin can reach, not what it can change.** `Pages` and
`Page` return copies of the page struct and of its `Paths`, `Redirects`, `SEO` and
`DependencyTags` containers, so editing those does not touch the application's own
page. The fragment pointers inside a copy are still shared, and the events below
carry the *live* page, not a copy — copying a page and its fragment tree on every
request would cost the hot path. Writing through an event's `Page` changes the page
every concurrent request is reading: that is a data race, and `go test -race` will
say so. Treat pages as read-only. Plugins are trusted code, not a sandbox.

Pages, documents and mounts a plugin registers are held to the same rules as the
application's own: a name or a path that is already taken is a startup error, not
a race decided by registration order.

## The hooks

| Interface | Method | Event | Fires | May change |
| --- | --- | --- | --- | --- |
| `PageResolvedHook` | `OnPageResolved` | `PageResolvedEvent` | Once per page request, right after routing — cache hits included | nothing |
| `BeforeRenderHook` | `OnBeforeRender` | `BeforeRenderEvent` | Before a fresh page render | nothing on the event; may hoist through `ev.Context` |
| `AfterRenderHook` | `OnAfterRender` | `AfterRenderEvent` | After a page render succeeded | `ev.HTML` |
| `DocumentRenderedHook` | `OnDocumentRendered` | `DocumentRenderedEvent` | After a document handler produced its body | `ev.Body` |
| `CacheWriteHook` | `OnCacheWrite` | `CacheWriteEvent` | Before a page or document is written to the cache | `ev.Skip`, `ev.TTL`, `ev.Tags` |
| `CacheInvalidateHook` | `OnCacheInvalidate` | `CacheInvalidateEvent` | After entries were invalidated by tag | nothing |
| `ErrorHook` | `OnError` | `ErrorEvent` | On a failure while serving a request | nothing |

Every hook method has the shape `func(ctx context.Context, ev *Event) error`.

### PageResolvedHook

```go
type PageResolvedEvent struct {
	Page   *collage.Page // live — do not write through it
	Locale string
	Path   string
}
```

Fires once per request that routed to a page, before the cache is consulted, so
it sees cache hits as well as fresh renders. It never fires for a document, and
it does not fire during a static build — a build is not a request, and a plugin
counting requests would count renders nobody asked for. An error fails the request
with a 500, under the stage `"page_resolved"`.

### BeforeRenderHook

```go
type BeforeRenderEvent struct {
	Context *collage.RenderContext // the render about to run
	Page    *collage.Page
	Locale  string
	Path    string
}
```

Fires immediately before a fresh render and **not** on a cache hit — that is the
difference from `PageResolvedHook`. It fires for pages, for error pages, for a page
an action answers with through `RenderPage`, and for every page a static build
renders.

It is the only hook that gets the render context, and the reason is hoisting. A
plugin that contributes to the page has to declare before the tree renders:

```go
func (p *Plugin) OnBeforeRender(_ context.Context, ev *collage.BeforeRenderEvent) error {
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}
```

A declaration made here sits at depth zero, so any fragment that declares the same
key replaces it: the plugin provides the default, the page provides the specific
thing. It lands only where the layout calls `{{hoist "head"}}`. An error fails the
request with a 500, under `"before_render"`.

### AfterRenderHook

```go
type AfterRenderEvent struct {
	Page     *collage.Page
	Locale   string
	Degraded bool   // some fragment failed, fallback or not
	HTML     []byte // replace it to post-process the page
	// Data: the render's shared data, the map behind rc.Set and rc.Get
}
```

Fires after a page render succeeded, for pages, error pages, a page an action
answers with through `RenderPage`, and static builds.
Replace `ev.HTML` to post-process; what you leave there is what is served, and —
unless a cache-write hook skips it — what is cached. Later plugins see what earlier
ones produced.

`ev.Data` is the render's shared data — the same map fragments read and write with
`rc.Set` and `rc.Get` — so it is what the page was built *from*, for a plugin that
wants the article rather than markup to parse back. What is in it is entirely the
application's convention; the framework puts nothing there. It is the live map:
reading it is fine, keeping it past the hook is holding request state.

Two consequences of where it sits:

- **It does not run again on a cache hit.** Its output is what was cached. A hook
  that must run per request cannot be combined with a cached page.
- **An empty result is a failure.** If `ev.HTML` is empty after dispatch, the
  request fails with a 500 rather than serving a blank page.

A page an action answers with through `RenderPage` runs it too (since v0.10.0;
before, it ran `BeforeRender` only), so a validation page is minified like any
other. An error fails the request with a 500, under `"after_render"`.

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

A document does not dispatch `OnPageResolved`, `OnBeforeRender` or `OnAfterRender`:
it has no page and renders no templates.

### CacheWriteHook

```go
type CacheWriteEvent struct {
	Key  string        // the cache key; changing it changes nothing
	Page *collage.Page // nil for a document
	TTL  time.Duration // may be adjusted
	Tags []string      // may be adjusted
	Skip bool          // set true to suppress the write
}
```

Fires before a rendered page or document is stored. **`Page` is `nil` for a
document**, and a hook that dereferences it unguarded panics on every document
request — the panic is contained, but the write it was dispatched for is abandoned,
so the document is never cached:

```go
func (p *Plugin) OnCacheWrite(_ context.Context, ev *collage.CacheWriteEvent) error {
	if ev.Page == nil {
		return nil // a document; Key, TTL and Tags are still valid
	}
	if ev.Page.Name == "home" {
		ev.TTL = time.Minute
	}
	return nil
}
```

An error, or `Skip`, suppresses the write and the request still succeeds: the page
has already rendered, and serving it uncached is better than turning a cache
problem into a 500. The error is reported to error hooks under `"cache_write"`.

### CacheInvalidateHook

```go
type CacheInvalidateEvent struct {
	Tags []string
}
```

Fires after `InvalidateTags` dropped the entries for some tags — whether the
application, an action or a plugin called it. It is dispatched from that call, not
from a request, and an error is joined into what `InvalidateTags` returns. To
trigger an invalidation yourself, call `Host.InvalidateTags`.

### ErrorHook

```go
type ErrorEvent struct {
	Err   error
	Page  *collage.Page // nil unless the failure was a page's own; see below
	Path  string
	Stage string
}
```

Fires on a failure while serving a request: a page, a document, an action, a mount
or a handler registered with `App.Handle`. `Stage` names where it happened.

`Page` is set only for a failure of a page that routing resolved. It is `nil` when
no page was resolved, and also for a document, an action — including a page the
action answers with through `RenderPage` — a mount and an `App.Handle` handler.
Read `Path` to tell those apart, and guard every use of `Page`. The
stages the framework uses are `"route"`, `"not_found"`, `"page_resolved"`,
`"before_render"`, `"render"`, `"after_render"`, `"cache_write"`, `"error_page"`,
`"asset"`, `"handler"` and `"panic"` — the set is not a closed enum.

`"error_page"` is the one worth alerting on: it means the page that reports
failures itself failed, and the client still received a plausible built-in page,
so nobody would otherwise find out.

Classify `Err` with `errors.Is` — `collage.ErrNoRoute` for a URL that matched
nothing, `collage.ErrNotFound` for content that does not exist,
`collage.ErrMethodNotAllowed` for a 405, `collage.ErrCSRFMissing` and its siblings
for a refused submission, `collage.ErrAssetFailed` for a mount answering 4xx or
5xx, `collage.ErrPanic` for a recovered panic, and the rest listed in
[Errors](/docs/errors#reported-to-error-hooks).

An error returned from `OnError` is logged and swallowed, and the remaining
plugins still receive the event: an error handler that fails must not start another
round of error handling.

### Dispatch rules

- Hooks run in **registration order**.
- Every call is **panic-guarded**. A panicking hook fails like one that returned an
  error; it does not take the process down.
- For `OnPageResolved`, `OnBeforeRender`, `OnAfterRender` and
  `OnDocumentRendered`, the **first error stops dispatch** and fails the request.
- For `OnCacheWrite`, the first error stops dispatch and suppresses the write.
- For `OnCacheInvalidate`, the first error stops dispatch and is returned from
  `InvalidateTags`.
- For `OnError`, errors are logged and dispatch continues.

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
  and decide.
- A plugin function under a built-in name replaces the built-in — except the
  functions bound per render (`slot`, `hoist`, `asset`, `stylesheet`, `csrfToken`,
  `pageURL`, `pageURLIn`, `localeURL`), which the render engine rebinds every time.
  See [Template functions](/docs/template-functions).
- It must be called from `Configure`. There is no way to add one later, because a
  function added after parsing is one no template can call.

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

## Contributing pages, documents and mounts

From `Init`, a plugin can add routes of its own through `Host.RegisterPage`,
`Host.RegisterDocument` and `Host.Mount`, built with the same builders an
application uses.

A plugin that *produces files* — resized images, generated icons — should serve
them from a mount rather than a route. A static build copies every mount into its
output after every page has rendered, so a filesystem that records what the pages
asked for hands the builder exactly the right set, and the exported site needs
nothing running behind it. A document at a dynamic path cannot be enumerated that
way.

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

`RegisterCommand` rejects an empty name (`ErrEmptyCommandName`) and a name another
command already has (`ErrDuplicateCommand`). It is never closed by `ErrAppStarted`:
it works during `Init`, as the other `Host` registration calls do, and after
startup too — though a command registered after `DispatchCommands` has run is one
nobody dispatches.

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
the commands. The exit codes are `0` for success; `1` for a startup failure, a
command that ran and failed, or a command with no `Run`; and `2` for a nil app, no
arguments, or an unclaimed name (`ErrUnknownCommand`). Say in your plugin's README that its commands
run through the application — a project scaffolded before v0.10.0 has to add that
block itself. See [The collage CLI](/docs/cli#plugin-commands).

## Configuration

A plugin reads its own section of `Config.PluginConfig` into a typed struct with
`host.Config`, available in both phases. Set your defaults first; `Config` decodes
the application's section over them:

```go
type Config struct {
	Generator string `json:"generator"`
	Disabled  bool   `json:"disabled"`
}

func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	p.cfg = Config{Generator: "collage"} // defaults
	return host.Config(&p.cfg)           // overlaid by the application's section, if any
}
```

- **An absent section leaves `v` untouched**, so "not configured" and "configured to
  the zero value" stay different statements.
- **A present but malformed section is an error.** The operator wrote something,
  and running on defaults instead would be the silent failure this refuses.
- The section is decoded with `json.Unmarshal` over your defaults, and follows its
  rules. A scalar or a slice in the JSON replaces your default — a slice is not
  merged. A JSON object decoded into a map adds its entries to the map you set,
  keeping the others. A JSON object decoded into a nested struct sets only the
  fields it names, leaving the rest at your defaults.
- The application sees a key that names no registered plugin as a startup error
  (`ErrUnknownPluginConfig`). Your `Name` is therefore the whole of your
  configuration's address; changing it is a breaking change.

Document every key, its type and its default in your README. Offering a
`NewWith(Config)` constructor alongside `New()` lets an application configure you
in Go as well.

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

// Config is the plugin's configuration.
type Config struct {
	// Generator is what the generator meta tag says.
	Generator string `json:"generator"`
}

// Plugin is the stamp plugin. Construct it with New.
type Plugin struct {
	cfg    Config
	logger *slog.Logger
}

// New returns the plugin with its defaults.
func New() *Plugin { return &Plugin{} }

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
	p.cfg = Config{Generator: "collage"} // the defaults
	if err := host.Config(&p.cfg); err != nil {
		return err // a section that is present but malformed
	}
	return host.AddTemplateFunc("generator", func() string { return p.cfg.Generator })
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
	ev.Context.HoistMeta("generator", p.cfg.Generator)
	return nil
}

// OnError reports the one failure nobody would otherwise notice: the error page
// itself failing.
func (p *Plugin) OnError(_ context.Context, ev *collage.ErrorEvent) error {
	if ev.Stage == "error_page" {
		p.logger.Error("stamp: the error page failed", "path", ev.Path, "error", ev.Err)
	}
	return nil
}
```

An application uses it like any other plugin:

```go
app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{stamp.New()},
	PluginConfig: pluginConfig, // {"acme/stamp": {"generator": "The Wire"}}
})
```

## Lifecycle

1. **Registration.** `Config.Plugins` inside `New`, or `RegisterPlugin` before the
   application starts. After that, `RegisterPlugin` returns `ErrAppStarted` — also
   after a start that failed in a plugin's `Init` (since v0.11.0).
2. **Configure**, inside `New`, for plugins that implement it — in registration
   order, stopping at the first error.
3. **Init**, when the application starts, in registration order. If one fails,
   startup is aborted and every plugin already initialised is shut down in reverse
   order. The failing plugin is not, since it never finished initialising.
4. **Shutdown**, from `App.Shutdown` — which `ListenAndServe` calls on `SIGINT` or
   `SIGTERM` — in reverse registration order. It calls **every registered
   plugin's** `Shutdown`, whether or not that plugin's `Init` ran or succeeded: an
   application that never started, one whose start failed, and the plugins the
   failed start already rolled back all get the call. So `Shutdown` must be safe
   to call without `Init` and more than once. Every plugin gets its turn even if one
   fails, and the errors are joined.

   With `ListenAndServe`, plugins are shut down after the server has drained its
   requests, or once `Server.ShutdownTimeout` has passed if it has not — past the
   deadline a request may still be running. With a server you own, the `App` knows
   of none: stop your server first, then call `App.Shutdown`, or a plugin can be
   torn out from under an in-flight request.

## Testing a plugin

Test a plugin the way an application would use it: build a real `App` with the
plugin in `Config.Plugins`, register a page whose template exercises it, and drive
the application through `app.Handler()` with `httptest`. No server listens and no
port is chosen.

```go
package stamp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"

	"example.com/stamp"
)

// newApp builds a one-page application with the plugin in it.
func newApp(t *testing.T, config string) *collage.App {
	t.Helper()

	root := t.TempDir()
	page := `<html><head>{{hoist "head"}}</head><body>{{generator}}</body></html>`
	if err := os.MkdirAll(filepath.Join(root, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pages", "home.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &collage.Config{
		Template: collage.TemplateConfig{Root: root},
		Plugins:  []collage.Plugin{stamp.New()},
	}
	if config != "" {
		cfg.PluginConfig = map[string]json.RawMessage{stamp.Name: json.RawMessage(config)}
	}
	app, err := collage.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	home := collage.NewPage("home").
		WithContent(collage.NewFragment("home", "pages/home.html").Build()).
		WithPath("en", "/").
		Build()
	if err := app.RegisterPage(home); err != nil {
		t.Fatalf("RegisterPage: %v", err)
	}
	return app
}

func get(t *testing.T, app *collage.App, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d: %s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestStamp_UsesItsDefaults(t *testing.T) {
	body := get(t, newApp(t, ""), "/")
	if !strings.Contains(body, `<meta name="generator" content="collage">`) {
		t.Errorf("no default meta tag in %s", body)
	}
}

func TestStamp_ReadsItsConfiguration(t *testing.T) {
	body := get(t, newApp(t, `{"generator": "my site"}`), "/")
	if !strings.Contains(body, `<meta name="generator" content="my site">`) {
		t.Errorf("no configured meta tag in %s", body)
	}
	if !strings.Contains(body, "<body>my site</body>") {
		t.Errorf("template function not applied in %s", body)
	}
}

func TestStamp_RegistersItsCommand(t *testing.T) {
	app := newApp(t, "")
	if err := app.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	commands := app.Commands()
	if len(commands) != 1 || commands[0].Name != "pages" {
		t.Errorf("commands = %v, want one named pages", commands)
	}
}

func TestStamp_RefusesLateRegistration(t *testing.T) {
	app, err := collage.New(&collage.Config{Template: collage.TemplateConfig{Root: t.TempDir()}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := app.RegisterPlugin(stamp.New()); !errors.Is(err, collage.ErrConfigurerRegisteredLate) {
		t.Errorf("RegisterPlugin: err = %v, want ErrConfigurerRegisteredLate", err)
	}
}
```

A few things worth testing that are easy to forget:

- **The unconfigured case.** Most applications will never write a section for you.
- **Documents**, if you implement `OnCacheWrite` or `OnDocumentRendered`: register a
  document and request it, so a `nil` `Page` is caught in a test rather than as a
  document that is never cached.
- **Run with `-race`.** A hook that writes through an event's live `Page` is a data
  race the detector reports and nothing else will.
- **A static export**, if you produce files: `collage.NewBuilder(app, ...)` into a
  `t.TempDir()` shows whether the exported site has what the served one had. See
  [Testing](/docs/testing) and [Static export](/docs/static-export).
