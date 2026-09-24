---
description: Every field of collage.Config and its sub-structs, with its default and what validation checks.
reference: Config, CacheConfig, ServerConfig, SecurityConfig, TemplateConfig, LocaleConfig, ObservabilityConfig, Config.Validate
---

# Configuration

An application is configured with one `collage.Config` value, handed to
`collage.New`. Every field has a usable zero value, so a configuration only says
what differs from the defaults:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
})
```

`New` does three things with it, in order:

1. **Fills in defaults.** It calls `cfg.ApplyDefaults()`, which sets every
   zero-valued field that has a default. It changes the value you passed, so you
   can read back exactly what the application was built with.
2. **Validates.** It calls `cfg.Validate()` and returns the first problem as an
   error — see [Validation](#validation).
3. **Builds the application.** A template root that does not exist, a template that
   does not parse, a plugin's `Configure` failing — each is reported here, not at
   the first request.

Passing `nil` is `ErrNilConfig`. A nil configuration is a mistake rather than a
request for defaults, because the configuration says where your templates are.

## Config

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `DevMode` | `bool` | `false` | Development mode across the framework. |
| `DevWatch` | `[]string` | none | Extra directories whose changes reload a development page. Since v0.10.0. |
| `Logger` | `*slog.Logger` | chosen at startup | The logger the framework and plugins write through. |
| `Server` | `ServerConfig` | | The HTTP server. |
| `Security` | `SecurityConfig` | | Request-forgery protection. |
| `Template` | `TemplateConfig` | | Template loading and rendering. |
| `Cache` | `CacheConfig` | | The page cache. |
| `Locale` | `LocaleConfig` | | Which locales URLs carry. |
| `Observability` | `ObservabilityConfig` | | Metrics and tracing. |
| `Plugins` | `[]Plugin` | none | Plugins registered while the application is built. |
| `PluginConfig` | `map[string]json.RawMessage` | none | Each plugin's own configuration, keyed by plugin name. |

### DevMode

Turns on development mode: templates reloaded from disk before every render, pages
that reload themselves in the browser, failed fragments shown on the page, the
page cache never read from, a disk cache replaced by memory, directories named in
`DevWatch` watched, and the built-in error page naming the fragment where a
failure started and showing the full error chain. **It must be off in
production** — those diagnostics carry paths, hostnames and whatever else error
messages contain.

The effective value is `cfg.IsDevMode()`, which is `DevMode || Template.DevMode`.
A scaffolded project sets it from `COLLAGE_DEV=1`, which `collage dev` sets for you.

### DevWatch

Directories, besides the templates and the mounts, whose changes reload a
development page — for content your application reads from disk itself, such as
Markdown or JSON. Since v0.10.0.

```go
app, err := collage.New(&collage.Config{
	DevMode:  os.Getenv("COLLAGE_DEV") == "1",
	DevWatch: []string{"content"},
	// ...
})
```

A directory that does not exist is skipped rather than an error, because the same
configuration runs wherever the binary is started. It is ignored outside
development. Watching a directory only reloads the browser; reading the new content
on the next render — rather than a copy loaded at startup — is up to your code.

### Logger

`nil`, the default, lets the framework choose when it builds the application: on a
terminal — and only when nothing has replaced slog's default handler — a compact
handler meant for a person, one line per record with a coloured level marker.
Anywhere else it is `slog.Default()`, unchanged. An application that called
`slog.SetDefault` keeps its handler. Pass a logger to be certain, such as a JSON
handler for logs a machine reads. `ApplyDefaults` leaves this field `nil`.

### Plugins and PluginConfig

`Plugins` are registered in order while `New` runs. A plugin that must act before
templates are parsed — to add a template function or wrap mounts — has to arrive
here rather than through `app.RegisterPlugin`.

`PluginConfig` holds each plugin's section as raw JSON. The framework reads no file:
fill it however you like, or use `collage.LoadPluginConfig("plugins-config.json")`,
which returns `nil` for a missing file. A key that names no registered plugin stops
the application from starting with `ErrUnknownPluginConfig`. See
[Using plugins](/docs/plugins).

## ServerConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Host` | `string` | `"localhost"` | Address the server listens on. |
| `Port` | `int` | `3000` | TCP port, `1`–`65535`. |
| `ReadTimeout` | `time.Duration` | `15s` | How long reading a request may take. |
| `WriteTimeout` | `time.Duration` | `30s` | How long writing a response may take. |
| `IdleTimeout` | `time.Duration` | `60s` | How long a keep-alive connection may sit idle. |
| `ShutdownTimeout` | `time.Duration` | `10s` | How long graceful shutdown waits for in-flight requests. |
| `MaxBodyBytes` | `int64` | 4 MiB | Bound on an action's request body when the action sets none. |

`Host` defaults to `localhost`, which is unreachable from outside the machine — in
a container, set it to `0.0.0.0`. `MaxBodyBytes` is not filled in by
`ApplyDefaults`: zero means the built-in 4 MiB (`4 << 20` bytes), applied when a
request arrives, and a negative value means unbounded. An unbounded body is memory
an anonymous caller chooses the size of, so choose that deliberately. An action can
set its own bound with `WithMaxBodyBytes`; see
[Forms and actions](/docs/forms-and-actions).

## SecurityConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `CSRFKey` | `[]byte` | generated per process | Key that signs forgery tokens. |
| `CSRFCookieName` | `string` | `"collage_csrf"` | Cookie a token is carried in. |
| `CSRFFieldName` | `string` | `"_csrf"` | Form field a token is submitted in. |
| `CSRFHeaderName` | `string` | `"X-CSRF-Token"` | Header a token may be submitted in instead. |
| `DisableCSRF` | `bool` | `false` | Turns forgery checking off for the whole application. |

**Set `CSRFKey` before deploying anything with a form.** It should be at least 32
random bytes, kept with your other secrets and the same on every instance. Left
empty, a key is generated for each process: fine for a first run, wrong to deploy,
because a token issued before a restart — or by another instance — is refused. The
application logs that it generated one — a warning outside development, at info
level in it — and only when some action answers an unsafe method (`POST`, `PUT`, `PATCH`,
`DELETE`) and is not exempted with `WithoutCSRF`, since only such an action ever
verifies a token (the `WithoutCSRF` exemption since v0.11.0). The key is not part
of the disk cache's namespace, so the cache survives a new key; a cached page with
a form in it, stored under the old one, is rendered again rather than served — see
[Caching](/docs/caching#the-namespace). The scaffolded `main.go` reads it from
`COLLAGE_CSRF_KEY`; `openssl rand -hex 32` makes one.

The name defaults above are applied by the forgery guard, not by `ApplyDefaults`,
so the fields stay empty in your `Config`. `CSRFFieldName` renames the field on
both sides: `{{csrfToken}}` writes the name the check reads, so forms need no
change. Script that reads the token from the form — `input[name="_csrf"]` — has
to use the new name.

`DisableCSRF` is for an application with no browser-submitted forms at all — an API
behind its own authentication. With it set, `{{csrfToken}}` fails the render rather
than rendering a form whose submission would mean nothing. To exempt one action
instead, use `WithoutCSRF` on it.

## TemplateConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `FS` | `fs.FS` | `nil` | Filesystem templates are loaded from; `nil` means disk. |
| `Root` | `string` | `"./templates"` on disk | Directory templates are loaded from, stripped from every template name. |
| `Extension` | `string` | `".html"` | File extension of template files. |
| `Funcs` | `template.FuncMap` | none | Extra template functions, merged over the built-ins. |
| `DevMode` | `bool` | `false` | Reload templates from disk on every render. |
| `Timeout` | `time.Duration` | `5s` | Default data-handler timeout, and the whole budget of every document handler. |

### FS and Root

With `FS` `nil`, `Root` is a path on disk, relative to the working directory, and
defaults to `./templates`. With `FS` set, `Root` is a slash-separated directory
inside it and is **not** defaulted — an empty `Root` means the root of `FS` itself.

Embedding is what lets a binary run from any directory:

```go
//go:embed all:templates
var templates embed.FS

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{FS: templates, Root: "templates"},
})
```

In development, an embedded set cannot change, so collage prefers the directory on
disk when it is there. A `Root` that does not exist is `ErrTemplateRootMissing`
from `New`, and a template under it that resolves outside it — through a symlink —
is `ErrTemplateEscapesRoot`.

### Funcs

Added to the built-in functions when the templates are parsed; an entry under a
built-in name replaces that built-in, and an entry under a name a plugin added
replaces the plugin's. It must be set before `New`, because a template can only
call a function that existed when it was parsed — a template calling an unknown
name fails in `New`. The functions bound per render (`slot`, `hoist`, `asset`,
`stylesheet`, `csrfToken`, `pageURL`, `pageURLIn`, `localeURL`) are rebound on every
render, so overriding them has no effect. See
[Template functions](/docs/template-functions).

### Timeout

The deadline for a data handler whose fragment sets none with `WithTimeout`. It is
also the only bound on a [document](/docs/documents) handler, which has no timeout
of its own — so raising it for one slow fragment raises it for every sitemap and
feed. Like every timeout in collage it bounds the context the handler receives; a
handler that never checks `ctx.Done()` can run past it.

## CacheConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Enabled` | `bool` | `false` | Master switch for the page cache. |
| `Store` | `Cache` | `nil` | Your own cache implementation. |
| `Type` | `string` | `"memory"` when enabled | Built-in cache when `Store` is `nil`: `"memory"` or `"disk"`. |
| `DefaultTTL` | `time.Duration` | `5m` | Entry lifetime when a page sets none. |
| `MaxEntries` | `int` | `10000` | Cap on cache entries; negative means unlimited. |
| `Dir` | `string` | none | Where a disk cache stores entries. Required for `"disk"`. |
| `Version` | `string` | derived from the binary | Identifies the build a disk cache's entries belong to. |
| `MaxKeysPerTag` | `int` | `10000` | Cap on cache keys recorded under one tag; negative means unlimited. |

**Caching is off by default.** With `Enabled` false nothing is cached, even with a
`Store` set. `Type` defaults to `"memory"` only when `Enabled` is true and `Store`
is `nil`.

**A `Store` replaces `Type` entirely**, including in validation. Implement
`collage.Cache` (`Get`, `Set`, `Invalidate`, `InvalidateKey`, `Clear`) to put pages
in Redis or anywhere else; a store that also implements `collage.TaggedCache` is
handed each entry's tags on write. Leave the field unset rather than assigning a
nil pointer of a concrete type, which is a non-nil interface that panics on the
first lookup.

**A disk cache outlives the process.** Its entries live in a subdirectory of `Dir`
named for a hash of `Version`, so a new build reads a different directory and finds
nothing stale. Leave `Version` empty and it is a hash of the running executable,
which changes exactly when the output might; set it — a commit, a release tag —
when something outside the binary decides what pages look like. If the executable
cannot be hashed, an in-memory cache is used instead, with a warning — and so,
since v0.11.0, when the directory cannot be created, a read-only filesystem say:
`collage.New` warns and carries on with memory rather than failing. A write that
fails once the application is running is logged, and the page is served uncached.
In development a disk cache is never used: an in-memory one stands in for it.

**`MaxEntries`** also bounds the values `collage.Cached` keeps across requests.
**`MaxKeysPerTag`** bounds the framework's dependency tracker, the per-process index
that maps a tag back to cache keys. Every distinct query string is a distinct key,
and nothing removes a key from the tracker when the cache evicts or expires its
entry, so without a cap a client could grow that index without limit. When a tag is
at the cap, its oldest key is dropped from the tracker only, not from the cache.

What that costs depends on the store. The built-in memory and disk caches index
tags themselves (they implement `TaggedCache`), so `InvalidateTags` still reaches
every entry they hold; only the count `InvalidateTagsN` reports, which is what the
tracker resolved, can come out lower. A custom `Store` that does not implement
`TaggedCache` relies on the tracker alone, and for it a dropped key is an entry
`InvalidateTags` no longer reaches — served until it expires. With such a store,
set the cap above the number of live entries any one tag can cover.

Zero for either cap means the default; only a negative value means unlimited. See
[Caching](/docs/caching).

## LocaleConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Default` | `string` | `"en"` | Locale of a URL with no locale prefix. |
| `Supported` | `[]string` | `[Default]` | Every locale the application serves. |
| `DisablePathLocale` | `bool` | `false` | Stop resolving the locale from the path; every request is in `Default`. |

The URL is the only thing that selects a locale: `/about` is in `Default`, and
`/tr/hakkinda` is in `"tr"`. collage never picks one from `Accept-Language` or a
cookie, because a URL that means different things to different readers is one
caches, crawlers and shared links all get wrong. `/en/about`, the default
locale's own prefix, redirects permanently to `/about`. To negotiate, do it in
middleware: redirect a Turkish browser to `/tr`, or render one URL per language
and declare it with `collage.Vary`. See
[Links and locales](/docs/links-and-locales#negotiating-a-language-yourself).

## ObservabilityConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Metrics` | `Metrics` | no-op | Receives counters and timings. |
| `Tracer` | `Tracer` | no-op | Starts spans around requests, renders and fragments. |

Both are interfaces you implement to bridge collage into your backend, and `nil`
means a no-op.

```go
type Metrics interface {
	RenderDuration(ctx context.Context, page string, d time.Duration, cacheHit bool)
	FragmentDuration(ctx context.Context, page, fragment string, d time.Duration, err error)
	CacheEvent(ctx context.Context, event collage.CacheEvent, key string)
	HTTPResponse(ctx context.Context, status int, path string, d time.Duration)
	Invalidation(ctx context.Context, tags []string, keys int)
}

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, collage.Span)
}

type Span interface {
	SetAttribute(key, value string)
	RecordError(err error)
	End()
}
```

`CacheEvent` is one of `collage.CacheHit`, `CacheMiss`, `CacheSet`, `CacheEvict`,
`CacheInvalidate` and `CacheCoalesced` — the last meaning a request was served by a
render of the same key that was already running.

## Validation

`Validate` returns the first of these it finds, after `ApplyDefaults` has run:

| Condition | Error |
| --- | --- |
| `Server.Port` outside `1`–`65535` | `ErrInvalidPort` |
| `Template.Root` empty and `Template.FS` nil | `ErrEmptyTemplateRoot` |
| `Cache.Enabled`, no `Store`, and `Type` neither `"memory"` nor `"disk"` | `ErrInvalidCacheType` |
| `Cache.Enabled`, no `Store`, `Type` `"disk"`, and `Cache.Dir` empty | `ErrEmptyCacheDir` |
| `Locale.Default` empty | `ErrEmptyLocaleDefault` |
| `Locale.Default` not in `Locale.Supported` | `ErrLocaleDefaultNotSupported` |
| A negative `Server.ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`, `Template.Timeout` or `Cache.DefaultTTL` | `ErrNegativeDuration` |

The negative-duration error is one sentinel for all six fields; the message names
the one that failed, such as `server.read_timeout`. Because defaults are applied
first, a zero port or an empty default locale never reaches validation from `New`
— only a value you set explicitly can fail. Match these with `errors.Is`; the full
list is in [Errors](/docs/errors#configuration).

You can call both yourself — to check a configuration in a test, say:

```go
cfg := collage.Config{Server: collage.ServerConfig{Port: 70000}}
cfg.ApplyDefaults()
err := cfg.Validate() // wraps collage.ErrInvalidPort
```

`ApplyDefaults` is idempotent: applying it to a configuration that already has
defaults changes nothing.

## A complete configuration

The configuration a scaffolded project starts from, with the parts that vary by
environment read from it:

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	return nil, fmt.Errorf("plugin configuration: %w", err)
}

app, err := collage.New(&collage.Config{
	DevMode: os.Getenv("COLLAGE_DEV") == "1",
	Server: collage.ServerConfig{
		Host: envString("HOST", "localhost"),
		Port: envInt("PORT", 3000),
	},
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "disk",
		Dir:        ".cache",
		DefaultTTL: 5 * time.Minute,
	},
	PluginConfig: pluginConfig,
	Security: collage.SecurityConfig{
		CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	},
})
```

`envString` and `envInt` are two small helpers in the scaffolded `main.go` that
read a variable and fall back to a default.
