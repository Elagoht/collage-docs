---
description: Every field of collage.Config and its sub-structs, with its default and what validation checks.
reference: Config, CacheConfig, ServerConfig, SecurityConfig, TemplateConfig, LocaleConfig, ObservabilityConfig, CacheType, ParseOrigin
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

1. **Fills in defaults.** It sets every zero-valued field that has a default. It
   changes the value you passed, so you can read back exactly what the application
   was built with.
2. **Validates.** It checks the result and returns the first problem as an
   error — see [Validation](#validation).
3. **Builds the application.** A template root that does not exist, a template that
   does not parse, a plugin's `Configure` failing — each is reported here, not at
   the first request.

Passing `nil` is an error. A nil configuration is a mistake rather than a
request for defaults, because the configuration says where your templates are.

A scaffolded project builds its configuration in `newApp`, in `main.go`. What
changes between your machine and a server comes from the environment; the rest is
written there:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 456" role="img" aria-labelledby="config-shape-title">
<title id="config-shape-title">Where the scaffolded main.go gets each field of collage.Config. From the environment: DevMode from COLLAGE_DEV=1, BaseURL from BASE_URL, Server.Host from HOST, Server.Port from the -port flag, whose default is PORT or 6060, and Security.CSRFKey from COLLAGE_CSRF_KEY. Set in main.go and the files beside it: PluginConfig from plugins-config.json, Plugins from plugins() in plugins.go, Template from the embedded templates directory, and Cache as a disk cache under .cache with a five-minute TTL. Not set, keeping their defaults: Locale, en only, Logger, DevWatch, TrailingSlash and Observability. The whole value goes to collage.New, which fills in defaults, validates and builds the application.</title>
<defs><marker id="config-shape-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<text class="d-small" x="78" y="14" text-anchor="middle">in the scaffold's main.go</text>
<text class="d-small" x="242" y="14" text-anchor="middle">collage.Config field</text>
<text class="d-small" x="160" y="34" text-anchor="middle">read from the environment</text>
<rect class="d-pill" x="6" y="40" width="144" height="24" rx="12" /><text class="d-small" x="78" y="56" text-anchor="middle">COLLAGE_DEV=1</text>
<line class="d-flow" x1="150" y1="52" x2="166" y2="52" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="40" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="56" text-anchor="middle">DevMode</text>
<rect class="d-pill" x="6" y="70" width="144" height="24" rx="12" /><text class="d-small" x="78" y="86" text-anchor="middle">BASE_URL</text>
<line class="d-flow" x1="150" y1="82" x2="166" y2="82" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="70" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="86" text-anchor="middle">BaseURL</text>
<rect class="d-pill" x="6" y="100" width="144" height="24" rx="12" /><text class="d-small" x="78" y="116" text-anchor="middle">HOST</text>
<line class="d-flow" x1="150" y1="112" x2="166" y2="112" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="100" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="116" text-anchor="middle">Server.Host</text>
<rect class="d-pill" x="6" y="130" width="144" height="24" rx="12" /><text class="d-small" x="78" y="146" text-anchor="middle">-port, PORT or 6060</text>
<line class="d-flow" x1="150" y1="142" x2="166" y2="142" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="130" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="146" text-anchor="middle">Server.Port</text>
<rect class="d-pill" x="6" y="160" width="144" height="24" rx="12" /><text class="d-small" x="78" y="176" text-anchor="middle">COLLAGE_CSRF_KEY</text>
<line class="d-flow" x1="150" y1="172" x2="166" y2="172" marker-end="url(#config-shape-arrow)" />
<rect class="d-a" x="170" y="160" width="144" height="24" rx="5" /><text class="d-on-a" x="242" y="176" text-anchor="middle">Security.CSRFKey</text>
<text class="d-small" x="160" y="202" text-anchor="middle">set in main.go and the files beside it</text>
<rect class="d-folder" x="6" y="208" width="144" height="24" rx="4" /><text class="d-small" x="78" y="224" text-anchor="middle">plugins-config.json</text>
<line class="d-flow" x1="150" y1="220" x2="166" y2="220" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="208" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="224" text-anchor="middle">PluginConfig</text>
<rect class="d-folder" x="6" y="238" width="144" height="24" rx="4" /><text class="d-small" x="78" y="254" text-anchor="middle">plugins() in plugins.go</text>
<line class="d-flow" x1="150" y1="250" x2="166" y2="250" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="238" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="254" text-anchor="middle">Plugins</text>
<rect class="d-folder" x="6" y="268" width="144" height="24" rx="4" /><text class="d-small" x="78" y="284" text-anchor="middle">embedded templates/</text>
<line class="d-flow" x1="150" y1="280" x2="166" y2="280" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="268" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="284" text-anchor="middle">Template</text>
<rect class="d-folder" x="6" y="298" width="144" height="24" rx="4" /><text class="d-small" x="78" y="314" text-anchor="middle">disk, .cache, 5m</text>
<line class="d-flow" x1="150" y1="310" x2="166" y2="310" marker-end="url(#config-shape-arrow)" />
<rect class="d-b" x="170" y="298" width="144" height="24" rx="5" /><text class="d-on-b" x="242" y="314" text-anchor="middle">Cache</text>
<text class="d-small" x="160" y="340" text-anchor="middle">not set: each keeps its default</text>
<rect class="d-frame" x="6" y="346" width="308" height="40" rx="6" /><text class="d-small" x="160" y="362" text-anchor="middle">Locale (en only), Logger, DevWatch,</text><text class="d-small" x="160" y="376" text-anchor="middle">TrailingSlash, Observability</text>
<line class="d-flow" x1="160" y1="386" x2="160" y2="404" marker-end="url(#config-shape-arrow)" />
<rect class="d-cmd" x="90" y="408" width="140" height="24" rx="12" /><text class="d-cmd-text" x="160" y="424" text-anchor="middle">collage.New</text>
<text class="d-small" x="160" y="448" text-anchor="middle">fills in defaults, validates, builds</text>
</svg>
<figcaption>Where the scaffolded <code>main.go</code> gets each field of <code>collage.Config</code>. Development mode, the public origin, the address the server listens on and the forgery key come from the environment, so one binary runs on your machine and on a server; the port's <code>-port</code> flag defaults to <code>PORT</code>, or 6060. The plugins, their configuration, the embedded templates and the disk cache are written in <code>main.go</code> and the files beside it. <code>Locale</code>, <code>Logger</code>, <code>DevWatch</code>, <code>TrailingSlash</code> and <code>Observability</code> are not set, so each keeps its default. <code>collage.New</code> receives the whole value.</figcaption>
</figure>

The fields and their defaults are in the tables below.
[A complete configuration](#a-complete-configuration) shows the scaffold's code,
and [Example: one configuration for development and production](#example-one-configuration-for-development-and-production)
what a production application adds to it.

## Config

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `DevMode` | `bool` | `false` | Development mode across the framework. |
| `DevWatch` | `[]string` | none | Extra directories whose changes reload a development page. Since v0.10.0. |
| `Logger` | `*slog.Logger` | chosen at startup | The logger the framework and plugins write through. |
| `Server` | `ServerConfig` | | The HTTP server. |
| `BaseURL` | `string` | none | The site's public origin, `https://example.com`, for the absolute URLs a render cannot know otherwise. |
| `Security` | `SecurityConfig` | | Request-forgery protection. |
| `Template` | `TemplateConfig` | | Template loading and rendering. |
| `Cache` | `CacheConfig` | | The page cache. |
| `Locale` | `LocaleConfig` | | Which locales URLs carry. |
| `TrailingSlash` | `bool` | `false` | End every page's URL in `/`. Since v0.13.0. |
| `Observability` | `ObservabilityConfig` | | Metrics and tracing. |
| `Plugins` | `[]Plugin` | none | Plugins registered while the application is built. |
| `PluginConfig` | `map[string]json.RawMessage` | none | Each plugin's own configuration, keyed by plugin name. |

### DevMode

Turns on development mode:

- templates reloaded from disk before every render;
- pages that reload themselves in the browser;
- failed fragments shown on the page;
- the page cache never read from, and a disk cache replaced by memory;
- directories named in `DevWatch` watched;
- the built-in error page naming the fragment where a failure started, leading
  with the cause — for a template, the file, line and column of the call that
  failed (since v0.15.0) — and showing the full error chain.

**It must be off in production** — those diagnostics carry paths, hostnames and
whatever else error messages contain.

`DevMode` is the one switch: `Template.DevMode`, which did the same before v0.60.0,
is gone. A plugin reads it with `host.DevMode()`. A scaffolded project sets it from
`COLLAGE_DEV=1`, which `collage dev` sets for you.

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

`nil`, the default, lets the framework choose when it builds the application:

- **On a terminal** — and only when nothing has replaced slog's default handler — a
  compact handler meant for a person, one line per record with a coloured level
  marker.
- **Anywhere else** it is `slog.Default()`, unchanged — except where `FORCE_COLOR`
  is set, as [`collage dev`](/docs/cli#collage-dev) does for the program it runs on
  a colour terminal, and `NO_COLOR` is not (since v0.32.0).

An application that called `slog.SetDefault` keeps its handler. Pass a logger to be
certain, such as a JSON handler for logs a machine reads. `New` leaves this field
`nil`.

### Plugins and PluginConfig

`Plugins` are registered in order while `New` runs. A plugin that must act before
templates are parsed — to add a template function or wrap mounts — has to arrive
here rather than through `app.RegisterPlugin`.

`PluginConfig` holds each plugin's section as raw JSON. The framework reads no file:
fill it however you like, or use `collage.LoadPluginConfig("plugins-config.json")`,
which returns `nil` for a missing file. A key that names no registered plugin stops
the application from starting with `ErrUnknownPluginConfig`, and since v0.61.0 so
does a key inside a section that the plugin does not know. See
[Configuring plugins](/docs/plugins#configuring-plugins).

### TrailingSlash

Every page has one address, and `TrailingSlash` says whether it ends in `/`. On,
`/blog/hello/` is the page and `/blog/hello` redirects to it with a `301`; off,
which is the default, it is the other way round. Links built by name —
`pageURL`, `pageURLIn`, `localeURL`, `app.URL` — come out in the chosen spelling,
a locale's home included: `/tr/` on, `/tr` off.

Turn it on for a site you [export](/docs/static-export#hosting) to a static host.
The export writes a page as `<path>/index.html`, which a host serves at
`/blog/hello/` and reaches from `/blog/hello` only through a redirect — so with it
off, every canonical link, sitemap entry and internal link points at a redirect.

It applies to pages. A [document](/docs/documents) is a file and keeps its path as
written, `/sitemap.xml` either way, and an action is answered at whichever
spelling it was posted to.

## ServerConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Host` | `string` | `"localhost"` | Address the server listens on. |
| `Port` | `int` | `6060` (`3000` before v0.32.0) | TCP port, `1`–`65535`. |
| `ReadTimeout` | `time.Duration` | `15s` | How long reading a request may take. |
| `ReadHeaderTimeout` | `time.Duration` | `0` | How long reading the request headers alone may take. Zero uses `ReadTimeout` (since v0.56.0). |
| `WriteTimeout` | `time.Duration` | `30s` | How long writing a response may take. |
| `IdleTimeout` | `time.Duration` | `60s` | How long a keep-alive connection may sit idle. |
| `ShutdownTimeout` | `time.Duration` | `10s` | How long graceful shutdown waits for in-flight requests. |
| `DrainDelay` | `time.Duration` | `0` | How long the server keeps serving, keep-alives off, once a shutdown starts and before the port closes; ignored in development (since v0.53.0). See [Deployment](/docs/deployment#graceful-shutdown-and-draining). |
| `MaxBodyBytes` | `int64` | 4 MiB | Bound on a request body when the action it routes to sets none — applied before middleware since v0.34.0. |
| `TrustedProxies` | `[]string` | none | Addresses and CIDR ranges of the proxies in front of the server, whose `X-Forwarded-For` is believed (since v0.47.0), and, when the list is set, whose `X-Forwarded-Proto` is (since v0.56.0). |

`Host` defaults to `localhost`, which is unreachable from outside the machine — in
a container, set it to `0.0.0.0`.

A scaffolded project fills `Host` and `Port` from the `HOST` and `PORT` environment
variables. Under `collage dev` those are a loopback address `collage dev` chose,
and it listens on yours itself, so keep reading them from the environment; see
[CLI](/docs/cli#collage-dev).

`MaxBodyBytes` is not filled in by `New`: zero means the built-in 4 MiB
(`4 << 20` bytes), applied when a request arrives, and a negative value means
unbounded. An unbounded body is memory an anonymous caller chooses the size of, so
choose that deliberately. An action can set its own bound with `WithMaxBodyBytes`;
see [Forms and actions](/docs/forms-and-actions).

`TrustedProxies` is what makes the client's address knowable behind a reverse proxy,
where every request's `RemoteAddr` is the proxy's. Each entry is an address or a CIDR
range (`"10.0.0.0/8"`, `"127.0.0.1"`); one that is neither makes `collage.New` fail,
naming it. Empty, the default, `X-Forwarded-For` is never read, so a client cannot
forge it, and `collage.ClientIP(r)` is `RemoteAddr`'s host:

```go
Server: collage.ServerConfig{
	TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1"},
},
```

A request from a trusted proxy names its client in `X-Forwarded-For`, read from the
right: the first address that is not a trusted proxy is the client, or the leftmost
when every one is.

- An entry may carry a port (`9.9.9.9:4567`, `[2001:db8::1]:443`) or brackets
  (`[2001:db8::1]`).
- An entry that is not an address (`unknown`), met before any untrusted one, leaves
  the client unknown: `ClientIP` is the zero `netip.Addr`, never the proxy.

List every hop, a CDN's published ranges included, and only proxies you run or your
platform's documented ranges. Trusting a range a client can send from lets that
client name any address it likes; a range with zero bits (`0.0.0.0/0`, `::/0`)
trusts everyone, and `collage.New` logs a Warn for it.

Since v0.56.0 the same list decides whose `X-Forwarded-Proto` marks the forgery cookie
`Secure`; empty, that header is believed from anyone, as before.
See [Deployment](/docs/deployment#behind-a-proxy-trustedproxies).

## SecurityConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `CSRFKey` | `[]byte` | generated per process | Key that signs forgery tokens. |
| `CSRFCookieName` | `string` | `"collage_csrf"` | Cookie a token is carried in. |
| `CSRFFieldName` | `string` | `"_csrf"` | Form field a token is submitted in. |
| `CSRFHeaderName` | `string` | `"X-CSRF-Token"` | Header a token may be submitted in instead. |
| `CSRFTrustedOrigins` | `[]string` | none | Other origins whose forms may post here, each `scheme://host[:port]` (since v0.34.0). |
| `DisableCSRF` | `bool` | `false` | Turns forgery checking off for the whole application. |
| `CSRFTokenTTL` | `time.Duration` | `12h` | How long a forgery token stays valid after it is issued; negative disables expiry (since v0.39.0). |
| `FrameOptions` | `string` | `"SAMEORIGIN"` | `X-Frame-Options` sent on every response; `"-"` sends none (since v0.39.0). |
| `DisableNoSniff` | `bool` | `false` | Stops the `X-Content-Type-Options: nosniff` header sent on every response (the header since v0.39.0; before v0.60.0 this field was `NoSniff *bool`). |

**Set `CSRFKey` before deploying anything with a form.** It should be at least 32
random bytes, kept with your other secrets and the same on every instance.

Left empty, a key is generated for each process: fine for a first run, wrong to
deploy, because a token issued before a restart — or by another instance — is
refused. The application logs that it generated one — a warning outside
development, at info level in it — and only when some action answers an unsafe
method (`POST`, `PUT`, `PATCH`, `DELETE`) and is not exempted with `WithoutCSRF`,
since only such an action ever verifies a token (the `WithoutCSRF` exemption since
v0.11.0).

The key is not part of the disk cache's namespace, so the cache survives a new key;
a cached page with a form in it, stored under the old one, is rendered again rather
than served — see [Caching](/docs/caching#the-namespace). The scaffolded `main.go`
reads it from `COLLAGE_CSRF_KEY`; `openssl rand -hex 32` makes one.

`CSRFTrustedOrigins` names the origins, other than this one, whose forms may post
here; every other origin is refused even with a valid token. See
[Forms and actions](/docs/forms-and-actions#where-the-request-came-from).

The name defaults above are applied by the forgery guard, not by `New`,
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
| `Type` | `CacheType` | `CacheMemory` when enabled | Built-in cache when `Store` is `nil`: `collage.CacheMemory` or `collage.CacheDisk`. |
| `DefaultTTL` | `time.Duration` | `5m` | Entry lifetime when a page sets none. |
| `MaxEntries` | `int` | `10000` | Cap on cache entries, memory and disk alike, and on the keys tracked under one tag; negative means unlimited entries (disk since v0.39.0). |
| `MaxBytes` | `int64` | 256 MiB memory, 1 GiB disk | Cap on the total size a built-in cache stores; negative means unlimited (since v0.40.0). A disk cache's memory tier holds 256 MiB, or `MaxBytes` when smaller (since v0.63.0). |
| `Dir` | `string` | none | Where a disk cache stores entries, and, in `Dir/data/`, the `collage.Cached` values. Required for `CacheDisk`. One per instance: instances sharing it do not share a cache. |
| `Version` | `string` | derived from the binary | Identifies the build a disk cache's entries belong to. |
| `WarmLimit` | `int` | `1000` | Cap on the pages one invalidation re-renders in the background, newest first; negative turns warming off (since v0.64.0). |

**Caching is off by default.** With `Enabled` false nothing is cached, even with a
`Store` set. `Type` defaults to `collage.CacheMemory` only when `Enabled` is true
and `Store` is `nil`. An untyped literal, `Type: "disk"`, still compiles; a string
read from a variable needs converting: `Type: collage.CacheType(v)`.

**A `Store` replaces `Type` entirely**, including in validation. Implement
`collage.Cache` (`Get`, `Set`, `InvalidateKey`, `Clear`) to put pages in Redis or
anywhere else; a store that also implements `collage.TaggedCache` (`SetTagged`,
`Invalidate`) is handed each entry's tags on write and asked to invalidate by tag.
Before v0.60.0 `Invalidate` was a method of `Cache` itself. A store that also
implements `collage.StaleCache` (`GetStale`, `MarkStaleKey`, `MarkStale`, since
v0.63.0) is asked to mark invalidated entries stale instead of removing them, and
can then answer a failed re-render with the old page: see
[Caching](/docs/caching#stale-pages-in-your-own-store). Leave the field unset rather than assigning a
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

**A disk cache has memory in front of it** (since v0.63.0). A hit is served from a
memory tier, bounded by `MaxEntries` and by 256 MiB, or `MaxBytes` when that is
smaller, and filled from the current build's directory at startup, newest first,
so a restart serves its recent pages from memory at once. Each instance has its
own tier, so give each instance its own `Dir`: with a shared one, an invalidation
in one instance leaves the others serving their in-memory copy as fresh. See
[Caching](/docs/caching#memory-in-front-of-the-disk).

**`Dir/data/` holds the `collage.Cached` values** of a disk cache, so a restart
comes up warm. Unlike the entries, they are not in a build's subdirectory, and
they are not encrypted: see [Caching](/docs/caching#across-a-restart).

**`MaxEntries`** also bounds the values `collage.Cached` keeps across requests, and,
with a built-in cache, the number of keys the dependency tracker holds in all
(since v0.40.0).

**`MaxBytes`** bounds what a built-in cache weighs: zero picks the type's default,
and a page larger than the whole cap is served but not stored. A custom `Store`
ignores it. See [Caching](/docs/caching#bounded-in-bytes-too).

**`MaxEntries` also bounds the framework's dependency tracker per tag** — the
per-process index that maps a tag back to cache keys. Before v0.60.0 that was a
field of its own, `MaxKeysPerTag`.

Every distinct query string is a distinct key, and nothing removes a key from the
tracker when the cache evicts or expires its entry, so without a cap a client could
grow that index without limit. When a tag is at the cap, its oldest key is dropped
from the tracker only, not from the cache. A negative `MaxEntries` keeps this
per-tag cap at 10000.

What that costs depends on the store. The built-in memory and disk caches index
tags themselves (they implement `TaggedCache`), so `InvalidateTags` still reaches
every entry they hold; only the count `InvalidateTagsN` reports, which is what the
tracker resolved, can come out lower. A custom `Store` that does not implement
`TaggedCache` relies on the tracker alone, and for it a dropped key is an entry
`InvalidateTags` no longer reaches — served until it expires. With such a store,
set `MaxEntries` above the number of live entries any one tag can cover.

Zero for `MaxEntries` or `MaxBytes` means the default; only a negative value means
unlimited. See
[Caching](/docs/caching).

## LocaleConfig

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `Default` | `string` | `"en"` | Locale of a URL with no locale prefix. |
| `Supported` | `[]string` | `[Default]` | Every locale the application serves. |
| `PrefixDefault` | `bool` | `false` | Give `Default`'s pages a prefix too: `/en/about`. Since v0.14.0. |

The URL is the only thing that selects a locale: `/about` is in `Default`, and
`/tr/hakkinda` is in `"tr"`. collage never picks one from `Accept-Language` or a
cookie, because a URL that means different things to different readers is one
caches, crawlers and shared links all get wrong.

`/en/about`, the default locale's own prefix, redirects permanently to `/about` —
unless `PrefixDefault` is set, which makes `/en/about` the address and redirects
`/about` to it; see
[Links and locales](/docs/links-and-locales#the-url-decides-the-locale).

To negotiate, do it in middleware: redirect a Turkish browser to `/tr`, or render
one URL per language and declare it with `collage.Vary`. See
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
`CacheInvalidate`, `CacheCoalesced`, `CacheStale` and `CacheWarm`. `CacheCoalesced`
means a request was served by a render of the same key that was already running;
`CacheStale`, that a `collage.Cached` value or a cached page was served stale,
counted once per value or response served; `CacheWarm` (since v0.64.0), that a
page or document an invalidation reached was rendered again in the background and
stored, once per key.

## Validation

`New` returns the first of these it finds, after the defaults are filled in:

| Condition | Error |
| --- | --- |
| `Server.Port` outside `1`–`65535` | `ErrInvalidPort` |
| `Cache.Enabled`, no `Store`, and `Type` neither `CacheMemory` nor `CacheDisk` | `ErrInvalidCacheType` |
| `Cache.Enabled`, no `Store`, `Type` `CacheDisk`, and `Cache.Dir` empty | `ErrEmptyCacheDir` |
| `Locale.Default` not in `Locale.Supported` | `ErrLocaleDefaultNotSupported` |
| A negative `Server.ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`, `DrainDelay`, `Template.Timeout` or `Cache.DefaultTTL` | `ErrNegativeDuration` |
| `BaseURL` set and not a bare `scheme://host[:port]` | `ErrInvalidOrigin` |

The negative-duration error is one sentinel for all eight fields; the message names
the one that failed, such as `server.read_timeout`. Because defaults are applied
first, a zero port or an empty default locale never reaches validation from `New`
— only a value you set explicitly can fail. Match these with `errors.Is`; the full
list is in [Errors](/docs/errors#configuration).

Filling in defaults and validating are not exported on their own since v0.60.0:
`New` is the one place they run, so a test checks a configuration by calling it:

```go
_, err := collage.New(&collage.Config{Server: collage.ServerConfig{Port: 70000}})
// errors.Is(err, collage.ErrInvalidPort)
```

Filling in defaults is idempotent: a configuration that already has them is left
as it is.

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
		Port: envInt("PORT", 6060),
	},
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       collage.CacheDisk,
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
read a variable and fall back to a default. The scaffold also reads `BaseURL` from
`BASE_URL`, and sets `Plugins` to `plugins()`, from the `plugins.go` that
`collage plugin add` and `del` manage.

## Example: one configuration for development and production

The [team's task list](/docs/middleware-and-apis#example-a-teams-task-list) — every
member signs in, an admin may change any task of the team, a member only their own —
runs under `collage dev` on a developer's machine and, in production, on
[two servers behind a load balancer](/docs/deployment#example-deploying-a-teams-task-list).
Its pages are in English and Turkish, and the team's admin panel, on a subdomain of
its own, posts forms to it.

One binary serves both. Its `newApp` is the scaffold's, with the production
settings added; what differs between the two is read from the environment or
decided by `devMode`:

```go
func newApp(devMode bool, port int) (*collage.App, error) {
	pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
	if err != nil {
		return nil, fmt.Errorf("plugin configuration: %w", err)
	}

	cfg := &collage.Config{
		DevMode: devMode,
		BaseURL: os.Getenv("BASE_URL"), // empty in development
		Server: collage.ServerConfig{
			Host:              envString("HOST", "localhost"),
			Port:              port,
			ReadHeaderTimeout: 5 * time.Second,
			DrainDelay:        5 * time.Second, // ignored in development
		},
		Template: collage.TemplateConfig{
			FS:        templatesFS,
			Root:      "templates",
			Extension: ".html",
		},
		Cache: collage.CacheConfig{
			Enabled:    true,
			Type:       collage.CacheDisk,
			Dir:        cacheDir,
			DefaultTTL: 5 * time.Minute,
		},
		Locale: collage.LocaleConfig{
			Default:   "en",
			Supported: []string{"en", "tr"},
		},
		PluginConfig: pluginConfig,
		Plugins:      plugins(),
		Security: collage.SecurityConfig{
			// Empty in development: a key is generated per process.
			CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
		},
	}

	if !devMode {
		// The load balancer in front of both servers, and the admin
		// panel whose forms post here.
		cfg.Server.TrustedProxies = []string{"10.0.1.10"}
		cfg.Security.CSRFTrustedOrigins = []string{"https://admin.tasks.example.com"}
	}

	app, err := collage.New(cfg)
	if err != nil {
		return nil, err
	}
	// Routes and the static mount as scaffolded.
	return app, nil
}
```

What each setting is in each place, and why:

| Setting | Development | Production | Why |
| --- | --- | --- | --- |
| `DevMode` | on: `collage dev` sets `COLLAGE_DEV=1` | off: the variable is unset | Development's error pages carry paths and error chains. A development server also answers only requests whose `Host` names this machine; see [Deployment](/docs/deployment#development-mode-stays-on-this-machine). |
| `Server.Host`, `Server.Port` | the loopback address `collage dev` chose | `HOST` and `PORT`, or `-port` | `collage dev` listens on your address itself and passes requests on; see [CLI](/docs/cli#collage-dev). |
| `BaseURL` | empty | `https://tasks.example.com` from `BASE_URL` | Empty is not an error: a site that builds no absolute URL needs none. Set, it must be a bare origin, or `New` fails with `ErrInvalidOrigin`. |
| `Security.CSRFKey` | empty: a key per process, logged at info level | at least 32 random bytes, the same on both servers | A generated key refuses every token issued before a restart, or by the other server. |
| `Security.CSRFTrustedOrigins` | none | the admin panel's origin | Every other origin is refused, even with a valid token. |
| `Server.TrustedProxies` | none | the load balancer | Without it, every client's address is the load balancer's. |
| `Server.DrainDelay` | 5s, ignored | 5s | The load balancer stops sending traffic before the port closes. |
| `Server.ReadHeaderTimeout` | 5s | 5s | A client sending headers a byte at a time is dropped early, while a body still has the whole `ReadTimeout`. |
| `Cache` | the same lines: memory stands in, never read | a disk cache under `.cache`, one per server | Development shows the template you just edited. A disk cache comes up warm after a restart. |
| `Locale` | `en`, `tr` | `en`, `tr` | The URL selects the locale, so nothing about it depends on where the binary runs. |
| `WriteTimeout`, `Template.Timeout` | defaults, 30s and 5s | defaults, 30s and 5s | Keeping `Template.Timeout` well under `WriteTimeout` turns a slow upstream into a fragment's fallback rather than a connection closed halfway through the page; see [Deployment](/docs/deployment#timeouts). |

The development side needs nothing written down: `collage dev` sets `COLLAGE_DEV`,
`HOST` and `PORT`. To keep a form open across the restarts `collage dev` makes when
Go code changes, give development a key of its own in `.env.development` — never
the production one; see [CLI](/docs/cli#environment-files).

The production side is four variables — `HOST`, `PORT`, `BASE_URL` and
`COLLAGE_CSRF_KEY` — set where the binary runs. Nothing reads `.env` files there;
see [Deployment](/docs/deployment#environment).
