---
description: Build the binary, run it in a container or under systemd, set what production needs, and put it behind TLS.
reference: ServerConfig, CacheConfig, LoadPluginConfig, IsHTTPS
---

# Deployment

A collage site is a Go program, so what you deploy is a compiled binary. Templates
and static files are embedded in it, so it runs from any directory with nothing
copied beside it — except `plugins-config.json`, if you configure plugins; see
[below](#plugin-configuration).

This page is about running that binary on a server; for a site with no server at
all, see [Static export](/docs/static-export).

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 290" role="img" aria-labelledby="deploy-server-title">
<title id="deploy-server-title">One collage binary on a server. A reader's browser connects over TLS to a reverse proxy or load balancer, where TLS ends. The proxy passes the request on over plain HTTP with the browser's Host, X-Forwarded-For and X-Forwarded-Proto; Server.TrustedProxies decides whether the binary believes those headers. The binary listens on HOST:PORT. From outside itself it reads its environment: PORT, 6060 when unset; HOST, localhost when unset; COLLAGE_CSRF_KEY, which must be set; COLLAGE_DEV, which must stay unset. In its working directory it keeps the page cache under .cache/hash and reads plugins-config.json.</title>
<defs><marker id="deploy-server-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="90" y="8" width="140" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">reader's browser</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="50" marker-end="url(#deploy-server-arrow)" />
<text class="d-small" x="168" y="45">TLS</text>
<rect class="d-b" x="30" y="54" width="260" height="38" rx="5" /><text class="d-on-b" x="160" y="70" text-anchor="middle">reverse proxy or load balancer</text><text class="d-on-b d-small" x="160" y="84" text-anchor="middle">TLS ends here</text>
<line class="d-flow" x1="160" y1="92" x2="160" y2="122" marker-end="url(#deploy-server-arrow)" />
<text class="d-small" x="168" y="104">Host, X-Forwarded-For,</text>
<text class="d-small" x="168" y="116">X-Forwarded-Proto</text>
<text class="d-small" x="152" y="110" text-anchor="end">plain HTTP</text>
<rect class="d-a" x="6" y="126" width="308" height="52" rx="6" /><text class="d-on-a" x="160" y="143" text-anchor="middle">bin/mysite on HOST:PORT</text><text class="d-on-a d-small" x="160" y="158" text-anchor="middle">Server.TrustedProxies: whose headers count</text><text class="d-on-a d-small" x="160" y="171" text-anchor="middle">templates and static files inside</text>
<line class="d-flow" x1="80" y1="178" x2="80" y2="194" marker-end="url(#deploy-server-arrow)" />
<line class="d-flow" x1="240" y1="178" x2="240" y2="194" marker-end="url(#deploy-server-arrow)" />
<rect class="d-frame" x="6" y="198" width="150" height="86" rx="5" /><text x="81" y="214" text-anchor="middle">environment</text><text class="d-small" x="14" y="232">PORT: 6060 if unset</text><text class="d-small" x="14" y="246">HOST: localhost if unset</text><text class="d-small" x="14" y="260">COLLAGE_CSRF_KEY: set</text><text class="d-small" x="14" y="274">COLLAGE_DEV: unset</text>
<rect class="d-folder" x="164" y="198" width="150" height="86" rx="4" /><text x="239" y="214" text-anchor="middle">working directory</text><text class="d-small" x="239" y="234" text-anchor="middle">.cache/&lt;hash&gt;/</text><text class="d-small" x="239" y="247" text-anchor="middle">the page cache</text><text class="d-small" x="239" y="266" text-anchor="middle">plugins-config.json</text>
</svg>
<figcaption>One binary on a server. The reverse proxy or load balancer in front ends TLS and passes each request on over plain HTTP, with the browser's <code>Host</code> and the <code>X-Forwarded-For</code> and <code>X-Forwarded-Proto</code> headers; <code>Server.TrustedProxies</code> decides whether the binary believes them. The binary listens on <code>HOST:PORT</code>, which is <code>localhost:6060</code> when neither is set; the image file that <code>collage build -i</code> writes sets <code>0.0.0.0:8080</code>, and its service unit <code>127.0.0.1:8080</code>. Everything else it needs is inside it, except its environment and, relative to its working directory, the page cache and <code>plugins-config.json</code>.</figcaption>
</figure>

## Building: `collage build`

```sh
collage build
./bin/mysite
```

`collage build` runs the `go build` you would otherwise have to remember:

```sh
CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build -trimpath -ldflags="-s -w" -o bin/mysite .
```

- **CGO off**, because collage and the standard library need no C, and a static
  binary can go into an image that contains nothing else.
- **`-trimpath`**, so the binary does not carry the paths of the machine that built
  it.
- **`-s -w`** drops the debug tables, which is most of the size.
- **This machine by default** (since v0.32.0; linux/amd64 before). A binary built
  for a Mac does not run in a Linux container, so for a server name the target:
  `collage build -os linux -arch amd64`.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o path` | `bin/<module name>` | Where the binary is written. |
| `-os name` | this machine's | Target operating system — `linux` for most servers. |
| `-arch name` | this machine's | Target architecture — `amd64`, or `arm64` for Graviton or Ampere machines. |
| `-i` | off | Offer to write a Dockerfile and a systemd unit beside the binary. |

The binary goes in `bin/`, not `dist/`: `dist/` belongs to `collage export`, whose
`-clean` would delete a binary sitting in it.

There is no `collage start`. Running a compiled binary needs nothing remembered, and
a start command could only run `go run .` — which puts the Go toolchain in your
production image and compiles on every boot.

### `collage build -i`: a Dockerfile and a systemd unit

With `-i`, `collage build` asks whether to write a `Dockerfile` and a systemd unit
next to the binary:

```sh
$ collage build -i -os linux -arch amd64
Building mysite for linux/amd64.

  Write a Dockerfile? [y/N]: y
  Write a systemd unit? [y/N]: y

✓ bin/mysite
    linux/amd64 · 9.3 MB
    wrote bin/Dockerfile    docker build -f bin/Dockerfile .
    wrote bin/mysite.service

9.3 MB · linux/amd64 · 4.1s
```

They go in `bin/` because they are generated, and the project root is for what you
wrote. An existing file is never overwritten — these are files you are expected to
edit — and a scaffolded project's `.gitignore` ignores only the binary, so you commit
them like any other file.

The Dockerfile has two stages: it builds in the Go image and ships the binary alone
on a distroless base:

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mysite .

FROM gcr.io/distroless/static-debian12
WORKDIR /srv
COPY --from=build /mysite /usr/local/bin/mysite
ENV HOST=0.0.0.0 PORT=8080
# ENV COLLAGE_CSRF_KEY=
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mysite"]
```

Build it from the project root with `docker build -f bin/Dockerfile .`. There is no
`COPY` of templates or static files, because they are inside the binary. The final
stage copies the binary — and, since v0.11.1, `plugins-config.json` when the project
has one when the Dockerfile is written; see [Plugin configuration](#plugin-configuration). The `WORKDIR`
matters for the page cache and for that file; see below. Set `COLLAGE_CSRF_KEY` from
your platform's secrets rather than in the file.

The systemd unit runs the binary from `/usr/local/bin`, in `/srv/<name>`, as a user
of the same name, listening on `127.0.0.1:8080` for a reverse proxy in front. Adjust
the user, directory and path before installing it at
`/etc/systemd/system/<name>.service`:

```sh
systemctl daemon-reload && systemctl enable --now mysite
```

## Environment

The scaffolded `main.go` reads five variables. Nothing reads `.env` files in
production — those are for `collage dev` — so set them wherever the binary runs.

One more is read by the framework itself, not by `main.go`: `COLLAGE_DEV_HOST`, which
names extra hosts a development server answers (since v0.56.0). Leave it unset in
production; see [Development mode stays on this machine](#development-mode-stays-on-this-machine).

| Variable | Default | Set it to |
| --- | --- | --- |
| `HOST` | `localhost` | `0.0.0.0` in a container. `localhost` accepts nothing from outside the machine, which is right behind a local reverse proxy and wrong everywhere else. |
| `PORT` | `6060` | Whatever your platform assigns. The binary's `-port` flag overrides it. |
| `BASE_URL` | unset | The site's public origin, `https://example.com`: what a canonical link, an `og:url` or a sitemap puts in front of a path. A bare origin, with no path; `collage.New` refuses anything else with `ErrInvalidOrigin`. |
| `COLLAGE_CSRF_KEY` | generated | At least 32 random bytes. **Set this.** |
| `COLLAGE_DEV` | unset | **Nothing — leave it unset.** `collage dev` sets it to `1`, which turns on development mode: templates and static files read from disk, an in-memory cache that is never read, full error chains on error pages. A server running with it is a development server. |

Make a key with:

```sh
openssl rand -hex 32
```

Without one, collage generates a key per process and warns at startup. That is fine
on your machine and wrong on a server, for two reasons:

- **Forms break across restarts and instances.** A form rendered before a deploy is
  refused when submitted after it, and one instance refuses what another issued.
  Keep the key the same on every instance and across restarts.
- **Cached pages with forms are rendered again after every start.** The
  forgery-token placeholder in a cached page is derived from the key, so a page
  stored under the previous process's key is treated as a miss and rendered afresh.
  Pages without a form are unaffected; the rest of the disk cache survives the
  restart either way. See [Caching](/docs/caching#the-namespace).

Rotating the key is safe but not free: forms rendered before the change are refused,
and cached pages with a form are rendered again once each.

## Embedded templates and static files

A scaffolded project embeds both:

```go
//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS
```

Templates are given to collage as `Template.FS`, and static files are mounted from
`fs.Sub(staticFS, "static")`. In development mode the directories on disk win, so
editing a template shows on the next request; in production only the embedded copy
is read, so what you tested is what runs, wherever the process starts.

Content your pages read at runtime — Markdown files, a JSON catalogue — is yours to
embed the same way. collage-docs embeds its `content/` directory for exactly that
reason, and reads it from disk in development. Name such a directory in
[`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0) and editing a file
in it reloads the browser too.

## Plugin configuration

The scaffolded `main.go` reads `plugins-config.json` with
`collage.LoadPluginConfig("plugins-config.json")` — a path relative to the working
directory, not to the binary, and not embedded. A missing file is not an error: every
plugin silently runs on its defaults. So a server started somewhere the file is not
ignores your plugin settings without a word.

The Dockerfile `collage build -i` writes copies it when the project has one at the
time the Dockerfile is written (since v0.11.1; earlier ones copied only the binary).
Add the line yourself if you create the file later — `collage build -i` never
overwrites a Dockerfile — beside the `WORKDIR` the process starts in:

```dockerfile
WORKDIR /srv
COPY --from=build /mysite /usr/local/bin/mysite
COPY --from=build /src/plugins-config.json /srv/plugins-config.json
```

or embed it, so it travels in the binary like the templates:

```go
//go:embed plugins-config.json
var pluginConfigJSON []byte

var pluginConfig map[string]json.RawMessage
if err := json.Unmarshal(pluginConfigJSON, &pluginConfig); err != nil {
	return nil, fmt.Errorf("plugin configuration: %w", err)
}
```

With systemd, keep the file in the unit's `WorkingDirectory`.

## Graceful shutdown and draining

`app.ListenAndServe` traps `SIGINT` and `SIGTERM` and shuts down gracefully on
either; `app.Shutdown(ctx)` does the same when you call it yourself. The order is
fixed:

1. **Drain.** The ctx `ListenAndServe` gave
   [`ServeHook`](/docs/writing-plugins#streams-and-shutdown) plugins in `OnServe` is
   cancelled, so a scheduler starts no new work (since v0.55.0). Every plugin
   implementing [`DrainHook`](/docs/writing-plugins#streams-and-shutdown) is told,
   once — a health plugin turns its readiness check false here. Keep-alives are
   turned off: idle kept-alive connections close at once, busy ones after their
   current response, and their clients reconnect through the load balancer. The
   port stays open and requests are served as normal for `Server.DrainDelay`
   (since v0.53.0).
2. **Streams.** Development reload streams and plugin streams are closed; they
   never end on their own.
3. **Server.** The port closes, and requests in flight get up to
   `Server.ShutdownTimeout` (10 seconds by default) to finish.
4. **Plugins.** Every plugin's `Shutdown` runs, and `ListenAndServe` returns `nil`.
   Before that, the warm renders already running after an invalidation (queued
   ones are dropped) and the data cache's refreshes and disk writes are waited for,
   under the same deadline as step 3.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 334" role="img" aria-labelledby="deploy-stop-title">
<title id="deploy-stop-title">How a collage server stops on SIGTERM or SIGINT. First the drain: the ServeHook ctx is cancelled, DrainHook plugins are told, keep-alives are turned off, and the port keeps serving for DrainDelay, 0 by default. Then, under ShutdownTimeout, 10 seconds by default and counted from the end of the drain: streams are closed; the port closes and requests in flight finish; warm renders and the data cache's writes are waited for, every plugin's Shutdown runs, and ListenAndServe returns.</title>
<defs><marker id="deploy-stop-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="85" y="8" width="150" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">SIGTERM or SIGINT</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#deploy-stop-arrow)" />
<rect class="d-a" x="6" y="50" width="308" height="66" rx="6" /><text class="d-on-a" x="160" y="66" text-anchor="middle">1 · drain</text><text class="d-on-a d-small" x="160" y="82" text-anchor="middle">ServeHook ctx cancelled, DrainHook told,</text><text class="d-on-a d-small" x="160" y="95" text-anchor="middle">keep-alives off; the port still serves</text><text class="d-on-a d-small" x="160" y="108" text-anchor="middle">for DrainDelay (0: no wait)</text>
<line class="d-flow" x1="160" y1="116" x2="160" y2="132" marker-end="url(#deploy-stop-arrow)" />
<rect class="d-frame-a" x="6" y="136" width="308" height="190" rx="6" /><text class="d-small" x="160" y="152" text-anchor="middle">under ShutdownTimeout (10s), counted from here</text>
<rect class="d-frame" x="18" y="160" width="284" height="38" rx="5" /><text x="160" y="176" text-anchor="middle">2 · streams closed</text><text class="d-small" x="160" y="190" text-anchor="middle">reload streams, plugin streams</text>
<line class="d-flow" x1="160" y1="198" x2="160" y2="210" marker-end="url(#deploy-stop-arrow)" />
<rect class="d-frame" x="18" y="214" width="284" height="38" rx="5" /><text x="160" y="230" text-anchor="middle">3 · port closes</text><text class="d-small" x="160" y="244" text-anchor="middle">requests in flight finish</text>
<line class="d-flow" x1="160" y1="252" x2="160" y2="264" marker-end="url(#deploy-stop-arrow)" />
<rect class="d-frame" x="18" y="268" width="284" height="50" rx="5" /><text x="160" y="284" text-anchor="middle">4 · plugins stop</text><text class="d-small" x="160" y="298" text-anchor="middle">after warm renders and data-cache writes;</text><text class="d-small" x="160" y="311" text-anchor="middle">every Shutdown, then ListenAndServe returns</text>
</svg>
<figcaption>What a signal sets off. The drain comes first, and the port goes on serving through it for <code>DrainDelay</code>, which is 0 by default and ignored in development mode. <code>ShutdownTimeout</code> starts only when the drain ends, and bounds everything after it, so a stop can take <code>DrainDelay + ShutdownTimeout</code>. Calling <code>app.Shutdown(ctx)</code> yourself runs the same steps, with your ctx bounding all of them, the drain included.</figcaption>
</figure>

`DrainDelay` is 0 by default: no wait, and a single instance behind nginx, Caddy or
Cloudflare stops as fast as it did before v0.53.0. Set it when a load balancer has
to notice the instance is leaving before the port closes — a few seconds longer
than its readiness check takes to fail:

```go
app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{
		DrainDelay:      10 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	},
})
```

**The two add up.** On a signal, `ShutdownTimeout` starts when the drain ends, so a
stop can take `DrainDelay + ShutdownTimeout`. When you call `app.Shutdown(ctx)`
yourself, the one ctx bounds both the drain and the wait for requests in flight,
and `ShutdownTimeout` is not used: give it a deadline of at least `DrainDelay` plus
the time your requests need, or the drain uses up the time they would have had.

A plugin's `Shutdown` may run a little past that deadline:
[elagoht/jobs](/docs/plugins#elagohtjobs) waits up to one more second for a job
that ignores its ctx, so with it a stop can take
`DrainDelay + ShutdownTimeout + 1s`.

The wait can end early. A second `SIGINT` or `SIGTERM` — Ctrl-C pressed twice —
ends it at once and goes straight on to stopping the server; so does a `Shutdown`
ctx that is done.

In development mode `DrainDelay` is ignored, so restarts stay instant, though
`OnDrain` still fires. A `Shutdown` before anything serves does not wait, since
there is no traffic to drain, and tells no plugin, since none has started (since
v0.61.4).

Whatever stops the process has to wait longer than that sum before it kills it:

- **systemd:** `TimeoutStopSec` greater than `DrainDelay + ShutdownTimeout`, plus
  one second with elagoht/jobs. The
  unit `collage build -i` writes sets `TimeoutStopSec=30`, which covers the
  defaults; raise it with `DrainDelay`.
- **Kubernetes:** `terminationGracePeriodSeconds` greater than the sum. Point the
  readiness probe at `/readyz` and the liveness probe at `/healthz`, served by
  [elagoht/health](/docs/plugins#elagohthealth), so readiness fails as soon as the
  drain starts and the pod is taken out of the Service before its port closes:

```yaml
spec:
  terminationGracePeriodSeconds: 30   # > DrainDelay + ShutdownTimeout (+ 1s with elagoht/jobs)
  containers:
    - name: app
      readinessProbe:
        httpGet: { path: /readyz, port: 8080 }
        periodSeconds: 2
        failureThreshold: 1
      livenessProbe:
        httpGet: { path: /healthz, port: 8080 }
        periodSeconds: 10
        failureThreshold: 3
```

  With that readiness probe, a `DrainDelay` of 5 seconds covers its two-second
  period with room to spare.

If the server's wait runs out of time — a request still open when
`ShutdownTimeout` passes — the plugins are shut down anyway, and `ListenAndServe`
returns an error saying so (`collage: server shutdown: context deadline exceeded`),
as it does for a plugin whose `Shutdown` failed. The scaffold's `main.go` passes
that to `log.Fatalf`, so the process exits with status 1 rather than 0; a platform
that treats a non-zero exit on stop as a crash will say so.

If you raise `DrainDelay` or `ShutdownTimeout`, raise your platform's grace period
with it.

### Serving with your own server

`OnServe` is called only by `ListenAndServe`. An application that runs its own
`http.Server` over `app.Handler()` gets no `OnServe`, so it must start such
plugins itself and cancel their ctx when it begins to stop — for
[elagoht/jobs](/docs/plugins#elagohtjobs), call its `Start(ctx)` once the server is
listening. It must also call `app.Shutdown(ctx)` after its server has stopped:
that is what runs the plugins' `Shutdown`, and without it a queue's waiting items
vanish without even being logged.

```go
// Init first, and its error: app.Handler() would only log it and answer 503,
// and the jobs plugin's Start panics before its Init has run.
if err := app.Start(); err != nil {
	return err
}
ln, err := net.Listen("tcp", ":8080")
if err != nil {
	return err
}
srv := &http.Server{Handler: app.Handler()}
go srv.Serve(ln)
ctx, stopJobs := context.WithCancel(context.Background())
j.Start(ctx) // the port is bound: start the jobs

// On shutdown:
stopJobs()                // the drain: no more triggers; Enqueue still accepts
srv.Shutdown(shutdownCtx) // requests in flight finish
app.Shutdown(shutdownCtx) // the plugins stop; jobs finishes its queues and refuses more
```

`app.Shutdown` tells the `DrainHook` plugins that are running, but does not wait out `DrainDelay`: the
`App` knows of no server of yours to keep serving. Stop your server before calling
it.

## Health checks

Liveness and readiness come from the [elagoht/health](/docs/plugins#elagohthealth)
plugin: `/healthz` answers `200` while the process serves requests, and `/readyz`
answers `200` until a check of yours fails or a drain starts, when it turns `503`.
Point every platform at those two paths — the Kubernetes probes above, or a load
balancer's health check at `/readyz` in front of a systemd unit — so readiness
fails as soon as the drain starts. List the plugin before any plugin that can
refuse or answer a request; see its entry for why.

The project `collage new --template demo` scaffolds ships a `/healthz`
[document](/docs/documents), `documents/health.go`. Once `elagoht/health` is added,
the plugin's middleware answers `/healthz` first and the document is never reached;
startup does not catch this, because the plugin checks pages, not documents.
Delete the document, or move the plugin's liveness endpoint with `livePath`.

A project without the plugin can still answer its probes with documents of its own
— bytes and a content type, no template, so a check cannot start failing because a
template did. Keep them dynamic: a health check served from a cache would answer
`ok` long after it stopped being true. The demo's `documents/health.go` is one to
copy for liveness. Such a readiness document does not know about the drain,
though, so it keeps answering `200` until the port closes. A minimal project —
`collage new` without `--template demo` — has no `/healthz` at all.

## The page cache in production

The scaffold configures a disk cache:

```go
Cache: collage.CacheConfig{
	Enabled:    true,
	Type:       "disk",
	Dir:        cacheDir, // ".cache"
	DefaultTTL: 5 * time.Minute,
},
```

It holds every static and incremental page, and a page that declares no strategy
is static when nothing it renders has a data handler — see
[Caching](/docs/caching#a-page-that-declares-none). Rendered pages survive a
restart, so a redeploy of the same build does not re-render the site into a cold
cache. What a new build finds depends on what changed:

- **The cache is namespaced by a hash of the binary.** A new build reads a different
  directory, `.cache/<hash>`, so it never serves pages the previous build rendered.
  Nothing has to be cleared for correctness — but nothing clears it for space
  either: the directories of earlier builds are never removed, so a server that is
  redeployed in place, rather than as a fresh container, accumulates one per build.
  Delete the old ones from your deploy script. Set `Cache.Version` — a commit, a
  release tag — if something outside the binary decides what pages look like.
- **The directory is relative to the working directory.** It is the one thing about
  a scaffolded project that depends on where it was started. Set `WORKDIR` in the
  container or `WorkingDirectory` in the unit, or give `Dir` an absolute path, and
  make sure the process can write there.
- **A cache it cannot write to is not an error.** At startup, a directory that
  cannot be created — a read-only filesystem, a working directory the process may
  not write to — makes collage fall back to an in-memory cache with a warning,
  rather than fail to start (since v0.11.0). Once running, a write that fails is
  logged, the page is served uncached, and the next request renders it again: slow,
  not broken. Look for the warning, and check that the directory fills up after a
  deploy.
- **Each instance has its own.** In a container the directory is inside the
  container, so each instance fills its own cache, at one render per page per
  instance. Keep it that way: instances that share a directory do not share a
  cache. An invalidation on one marks the disk and its own memory, and every other
  instance keeps serving its in-memory copy as a fresh `HIT` until it expires —
  for a `Static` page, until a restart or an eviction — where a CDN refetching it
  can pick it up again.
- **It has memory in front of it, and a restart starts warm** (since v0.63.0). A
  hit is served from a memory tier, filled from the directory at startup, newest
  pages first, so a restarted instance serves its recent pages from memory at once.
  That tier is up to 256 MiB of the process, or `MaxBytes` when that is smaller:
  see [below](#a-memory-cache-and-the-containers-limit).
- **An invalidated page is kept** to be served, marked `STALE`, if its re-render
  fails or is degraded — see [Caching](/docs/caching#stale-pages-and-x-collage-cache).
  Since v0.64.0 it is also re-rendered in the background right after the
  invalidation, which costs CPU for a moment; see
  [warming](/docs/caching#warming-after-an-invalidation). Keep `Server.WriteTimeout`
  above 5 seconds, so an action's redirect can wait for its page to warm.

A CDN in front keeps copies of its own:
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) purges them when collage
invalidates a page, and [elagoht/compress](/docs/plugins#elagohtcompress) compresses
responses with Brotli and gzip, once per cached page rather than once per reader.

### A memory cache and the container's limit

A `"memory"` cache is bounded by `MaxBytes` — 256 MiB of stored pages unless you
set it — and the bound holds: the live heap stops growing once the cache is full.
The process does not stop there. Go's collector lets the heap grow to about twice
what was live after the last collection before it runs again, so a full 256 MiB
cache can mean a process well past 512 MiB. Measured on a page of about 14 KB asked
for under 30,000 distinct URLs, the process settled at about 720 MB — enough for a
512 MiB container to be killed with every setting left at its default.

Tell the runtime what it has with `GOMEMLIMIT`, a little under the container's
limit, and it collects harder as it approaches it instead of being killed:

```dockerfile
# In a 512 MiB container.
ENV GOMEMLIMIT=400MiB
```

The same run with `GOMEMLIMIT=320MiB` stayed at about 430 MB. The limit is soft: it
is what the runtime aims for, not a cap it enforces, so leave `MaxBytes` well below
it — what the cache holds is live and cannot be collected however hard the runtime
tries. A cache sized at a third to a half of the limit leaves room for the renders
themselves; in a 512 MiB container, lower `MaxBytes` from its default to
128–192 MiB. The scaffold's disk cache keeps its pages on disk, but since v0.63.0
it has a memory tier in front of them that this applies to as well: up to 256 MiB,
or `MaxBytes` when that is smaller. Set `GOMEMLIMIT`, and lower `MaxBytes` or
`MaxEntries` if the tier must hold less.

### Invalidation with several instances

`app.InvalidateTags` invalidates entries in the cache of the process it runs in. With
one instance, a CMS webhook that invalidates a post's tag updates the site. With
several, it updates only the instance the webhook reached; the others keep serving
the old page until its TTL runs out.

Choose one of three answers: give pages that change `Incremental(ttl)` so a stale
copy expires on its own, send the webhook to every instance, or configure a shared
store through `Cache.Store` that implements `collage.TaggedCache`, which lets one
invalidation reach entries any instance wrote, and `collage.StaleCache` too, so
that it keeps an invalidated page to fall back on. A `"disk"` cache whose `Dir` the
instances share is not such a store: each instance's memory tier keeps serving
what it holds. See [Caching](/docs/caching#memory-in-front-of-the-disk).

A shared store holds pages, not data. Values kept with
[`collage.Cached`](/docs/caching#caching-data-across-pages) live in each process's
memory, whatever `Cache.Store` is, and an invalidation reaches only the instance it
ran in. After a webhook reaches instance A, instance B renders a fresh page — the
shared store invalidated it — from the old value it still holds.

Give those values a
TTL short enough to live with, or send the invalidation to every instance.

## TLS, behind a proxy

collage serves plain HTTP, and has no `ListenAndServeTLS`. Terminating TLS well means
certificates, renewal, HTTP/2, HSTS and a redirect from port 80, and every platform
this runs on — a load balancer, a reverse proxy, Cloudflare, Fly, Render — already
does it better than a framework flag could.

Put the binary behind one, and make sure the proxy sends `X-Forwarded-Proto: https`.
collage uses it to mark the forgery-token cookie `Secure`, so a site served over
HTTPS never sends that cookie over plain HTTP.

A proxy that appends to the header
rather than replacing it sends a list, `https, http`; since v0.34.0 its first
entry, the reader's own connection, is what counts.

Pass the `Host` the browser sent on as well: the forgery check compares an older
browser's `Origin` with it, and it is part of the page cache's key. A minimal Caddy
configuration, which
also obtains the certificate:

```
example.com {
	reverse_proxy 127.0.0.1:8080
}
```

With nginx, `proxy_set_header X-Forwarded-Proto $scheme;` in the `location` block
does the same. The security headers themselves — HSTS, a Content-Security-Policy
and the rest — can come from the binary, with the
[elagoht/secure](/docs/plugins#elagohtsecure) plugin.
[elagoht/ratelimit](/docs/plugins#elagohtratelimit) limits how fast one client can
submit forms, and [elagoht/basicauth](/docs/plugins#elagohtbasicauth) puts a
password in front of a staging deployment.

If you want the binary to terminate TLS itself, `app.Handler()` is an ordinary
`http.Handler`:

```go
srv := &http.Server{
	Addr:              ":443",
	Handler:           app.Handler(),
	ReadHeaderTimeout: 15 * time.Second,
}
log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
```

Doing that means you own the timeouts and the signal handling that `ListenAndServe`
did for you: you call `app.Start()` first, start `ServeHook` plugins yourself, and
call `app.Shutdown(ctx)` once your server has stopped, so plugins shut down. See
[Serving with your own server](#serving-with-your-own-server).


## Behind a proxy: `TrustedProxies`

Behind a proxy, every request's `RemoteAddr` is the proxy's. The proxy names the real
client in `X-Forwarded-For`, but so can anyone else: a client sending its own
`X-Forwarded-For` is not to be believed. `Server.TrustedProxies` (since v0.47.0) says
whose to believe:

```go
Server: collage.ServerConfig{
	TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1"},
},
```

Each entry is an address or a CIDR range; one that is neither makes `collage.New`
fail.

`collage.ClientIP(r)` then answers who the client is: `RemoteAddr`'s host,
unless that is a trusted proxy, in which case the header is read from the right,
skipping trusted addresses, and the first untrusted one is the client (the leftmost,
when every one is trusted). Empty, the default, trusts no header, and `ClientIP` is
always `RemoteAddr`.

A header entry may carry a port (`9.9.9.9:4567`, `[2001:db8::1]:443`) or brackets
(`[2001:db8::1]`), as some proxies write it; it is read as the address. An entry that
is still not an address, such as the `unknown` some proxies send, ends the walk. If no
untrusted address was found before it, the client is unknown and `ClientIP` is the
zero `netip.Addr`: never the proxy, which would make every visitor the same client. A
plugin keying on the client skips such a request.

List every hop between the client and the server: your own proxies and, behind a CDN,
the CDN's published ranges too. A hop left out is taken for the client, and every
visitor coming through it becomes one.

List only proxies you run, or your platform's
documented ranges: trusting a range a client can send from lets that client name any
address it likes. An entry with zero bits (`0.0.0.0/0`, `::/0`) trusts everyone;
`collage.New` accepts it but logs a Warn saying so.

The same list decides whose `X-Forwarded-Proto` is believed (since v0.56.0). The
forgery cookie is `Secure` when the request arrived over TLS, and behind a proxy that
terminates TLS the proxy says so in that header.

With `TrustedProxies` set, a request
whose `RemoteAddr` is not a listed proxy cannot set it: a client reaching the port
directly over plain http gets a cookie it can send back over plain http. Empty, the
header is believed from anyone, as it was before the list existed, and that is right
only when nothing but your proxy can reach the server.

Since v0.61.0 that decision is one function, `collage.IsHTTPS(r)`, which the
forgery guard's cookie and plugins setting `Secure` cookies or HSTS all ask —
elagoht/session, flash, secure and the rest. So the list protects their cookies and
headers as well: list your proxy, and a client cannot claim the TLS connection it
does not have.

## Timeouts

`ListenAndServe` applies the timeouts in `Config.Server`:

| Field | Default | Bounds |
| --- | --- | --- |
| `ReadTimeout` | 15s | Reading a request, headers included — a client that sends headers slowly cannot hold a connection open. |
| `ReadHeaderTimeout` | 0 | Reading the request headers alone, so a client sending them a byte at a time is dropped early while a large body still gets the whole `ReadTimeout`. Zero uses `ReadTimeout` (since v0.56.0). |
| `WriteTimeout` | 30s | Writing the response. It starts when the request's headers have been read, so it bounds an upload's body too. A page whose data takes longer than this is cut off. |
| `IdleTimeout` | 60s | A keep-alive connection waiting for its next request. |
| `DrainDelay` | 0 | Serving on after a `SIGTERM`, keep-alives off, so a load balancer can stop sending traffic (since v0.53.0). |
| `ShutdownTimeout` | 10s | Requests in flight finishing once the port has closed, after the drain. |
| `MaxBodyBytes` | 4 MiB | An action's request body, unless the action sets its own. Negative is unbounded. |

```go
Server: collage.ServerConfig{
	Host:         envString("HOST", "localhost"),
	Port:         port,
	WriteTimeout: time.Minute,
},
```

`collage dev`'s proxy keeps its own fixed 10s header timeout, in development only. A
negative value in any of these is `collage.ErrNegativeDuration`.

Raising only `ReadTimeout` for uploads is a trap. An upload that takes longer than
`WriteTimeout` is read, and the handler stores it, but the answer is cut off: the
client sees the connection close, and tries again.

Raising both for the whole server
lets any slow client hold any endpoint that long. Give the upload action its own
deadline instead, with `WithBodyTimeout` — see [Uploads](#uploads).

The time a data handler may take is a different setting: each fragment's
`WithTimeout`, or `Template.Timeout` (5 seconds) for fragments and documents that
set none. Keep it well under `WriteTimeout`, so a slow upstream turns into a
fragment's fallback rather than a connection closed halfway through the page. See
[Configuration](/docs/configuration).

## Uploads

An upload (see [Streaming bodies](/docs/forms-and-actions#streaming-bodies)) meets
every limit between the browser and the handler, and the defaults of most of them
are sized for forms.

**The server's deadlines.** The defaults above end any upload after 15s of reading.
`WithBodyTimeout(d)` (since v0.57.0) on the upload action replaces both deadlines
with now plus `d` for that action alone, after its page's guards (and, for a
streaming action, its forgery check). Make `d` at least `MaxBodyBytes` divided by
the slowest connection you serve, plus the time to answer. An upload whose client
went away, or ran out of time, is recorded as `400` or `408` and logged at debug,
not as a server error. A client that left receives nothing, and under
`WithBodyTimeout` neither does one that ran out of time: its write deadline passed
with the read one.

**nginx** answers anything over `client_max_body_size` — 1 MiB by default — with
its own `413`, before collage sees it. By default it also buffers the whole request
body to its own disk before passing any of it on. The stream, the early refusal of
a bad file and the upload's progress are all lost, and nginx's temp directory is
what fills up. For the upload location:

```nginx
location /upload {
    client_max_body_size     2g;   # MaxBodyBytes, or a little more
    proxy_request_buffering  off;  # stream the body through
    proxy_send_timeout       30m;  # between two writes to collage
    proxy_read_timeout       30m;  # waiting for collage's answer
    proxy_pass               http://127.0.0.1:8080;
}
```

**Caddy** streams request bodies; its limit, when set, is
`request_body { max_size 2GB }`.

**Cloudflare and other CDNs** cap the request body by plan — 100 MB on the Free and
Pro plans when this was written; check Cloudflare's documentation for yours — and
answer a larger one themselves. Load balancers have caps and idle timeouts of their
own. Larger files need a hostname that bypasses the CDN, or resumable uploads in
chunks, which collage does not do.

**[elagoht/health](/docs/plugins#elagohthealth)'s `maxInFlight`** counts every
request being served, an upload included, for as long as it lasts. Size it for the
uploads you expect at once, or slow uploads will hold the slots and the rest of the
site will be answered `503`.

## Logs

With no `Config.Logger`, collage logs to `slog`'s default handler — or, on a
terminal, to one formatted for a person. On a server, where logs are read by a
machine, pass a JSON handler:

```go
Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
```

collage logs what goes wrong, not every request.
[elagoht/accesslog](/docs/plugins#elagohtaccesslog) writes one line per request, with
a request id, through the same logger;
[elagoht/prometheus](/docs/plugins#elagohtprometheus) serves the framework's metrics
at `/metrics`; and [elagoht/otel](/docs/plugins#elagohtotel) turns its spans into
OpenTelemetry traces.

## Development mode stays on this machine

`DevMode` serves things production must not: error pages with stacks and source, the
reload stream, a development toolbar. Two guards keep them here (since v0.56.0):

- **A foreign `Host` is refused.** In development the server answers `403` to a
  request whose `Host` is not localhost (or a name under `.localhost`), an IP
  address, a reserved name, `Server.Host`, or a name in `COLLAGE_DEV_HOST`. The
  reserved names are `example.com`, `example.net`, `example.org` and names under
  them, and names under `.example`, `.test` and `.invalid`. This holds when the
  application is run directly, not only under `collage dev`. A page on another
  site can make its own name resolve to `127.0.0.1` (DNS rebinding) and would then
  be same-origin with the development server, but its `Host` is still its own name.
  Reserved names are allowed because nobody can register one to point at you; they
  are what tests use (httptest's requests are for `example.com`). A static build's
  captures are not checked.
- **A reachable address is warned about.** When `ListenAndServe` binds anything but
  loopback in development (`0.0.0.0`, `::`, a LAN address), it logs a Warn saying
  the development pages expose the application's internals. `collage dev` logs the
  same when its own proxy is bound that way.

**Upgrading:** a `DevMode` test or request that uses a custom `Host` outside the
reserved names, a tenant domain for example, now gets a `403`. Set the name as
`Server.Host`, list it in `COLLAGE_DEV_HOST`, or use a reserved name such as
`tenant.test`.

### Reaching development by another name

`COLLAGE_DEV_HOST` is a comma-separated list of extra names a development server
answers. Set it for:

- a docker-compose service name, `http://app:6060` from another container;
- a name in `/etc/hosts` outside the reserved ones;
- several tenant hosts in one development server (or use `*.localhost` names, which
  need no setting);
- reaching a server bound to `0.0.0.0` from a phone by the machine's LAN name. By IP
  address it needs nothing.

```
COLLAGE_DEV_HOST=app,mybox.lan
```

An application in development mode reads it from its own process environment. Run
directly, set it in the shell or wherever the process gets its environment.
`collage dev` reads it from the shell or `.env.development`, applies it to its
proxy, and hands the program the list plus its own `HOST`, since the program
listens on loopback but sees the browser's `Host`. A refused request's `403` quotes
the `Host` and names the setting. A tunnel such as ngrok is refused unless you list
its name, on purpose: it puts the development pages on the internet.

Neither guard applies in production, where `DevMode` is off and the `Host` is
whatever your DNS sends.

## Example: deploying a team's task list

The [team's task list](/docs/middleware-and-apis#example-a-teams-task-list) — every
member signs in, an admin may change any task of the team, a member only their own —
goes to production on two servers behind a load balancer. The load balancer ends
TLS and sends each request to either server.

Two, because restarting an instance closes its port: with one, every restart is an
outage; with two, one serves while the other restarts. Sessions are the
application's own `sid` cookie, looked up through its own data layer, so either
server can read one. For the messages it shows after a change, this deployment adds
[elagoht/flash](/docs/plugins#elagohtflash), which signs its cookie with a key of
its own.

| It needs | From |
| --- | --- |
| A forgery key, the same on both servers | `COLLAGE_CSRF_KEY`, in a secrets file the unit reads |
| The flash key, the same on both servers | `FLASH_KEY`, the variable `KeyEnv` names |
| A page cache it can write | `.cache/<hash>` in the unit's `WorkingDirectory`, one per server |
| The reader's address and scheme | `Server.TrustedProxies` listing the load balancer |
| A drain the load balancer notices | elagoht/health first in the list, and `Server.DrainDelay` |

### The configuration

The scaffold's `newApp`, with the load balancer, the drain and the two plugins added:

```go
h := health.New(health.Options{})
h.Check("db", func(ctx context.Context) error { return db.PingContext(ctx) })

app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{
		Host: envString("HOST", "localhost"),
		Port: port,
		// The load balancer's private address: only its X-Forwarded-For
		// and X-Forwarded-Proto are believed.
		TrustedProxies: []string{"10.0.1.10"},
		// The load balancer checks /readyz every 2 seconds and takes a
		// server out after one failure: keep serving until it has.
		DrainDelay: 5 * time.Second,
	},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "disk",
		Dir:        cacheDir, // ".cache", under the unit's WorkingDirectory
		DefaultTTL: 5 * time.Minute,
	},
	Plugins: []collage.Plugin{
		h, // first, so a probe always reaches it
		flash.New(flash.Options{KeyEnv: "FLASH_KEY"}),
	},
	Security: collage.SecurityConfig{
		CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	},
	// Template and the rest as scaffolded.
})
```

With the load balancer listed, `collage.ClientIP` is the reader's address rather
than the load balancer's, and a request that reaches a server's port some other way
cannot claim HTTPS with its own `X-Forwarded-Proto`. Point the load balancer's
health check at `/readyz`: it answers `503` while the database check fails, and
from the moment a drain starts.

### The unit and its secrets

`collage build -i -os linux -arch amd64` writes `bin/tasks.service`. On each server,
its `[Service]` section changes in two places:

```text
WorkingDirectory=/srv/tasks
ExecStart=/usr/local/bin/tasks
Environment=HOST=10.0.1.11
Environment=PORT=8080
EnvironmentFile=/etc/tasks/secrets.env
```

- **`HOST` is the server's private address.** The unit's `127.0.0.1` would accept
  nothing from a load balancer on another machine. `PORT=8080` is the value the
  unit sets; without it collage would listen on 6060.
- **The secrets are in a file of their own**, kept out of the repository and the
  unit:

```sh
# 32 random bytes, hex
COLLAGE_CSRF_KEY=…
# 32 random bytes, hex, another one
FLASH_KEY=…
```

Both servers get the same file. A form one server rendered is then accepted by the
other, and a flash cookie one set is read by the other; with keys of their own,
neither would accept what the other issued.

The page cache lands in `/srv/tasks/.cache/<hash>`, so the unit's user, `tasks`,
must be able to write to `/srv/tasks`. Each server fills its own.

### What one server's invalidation misses

The task list's data handler keeps the team's list with `collage.Cached` and a `ttl`
of zero: fresh until its tag is invalidated. The API invalidates that tag on the
server its request reached. The other server never hears of it, and keeps serving
the list it holds.

Give the list a TTL instead:

```go
list, err := collage.Cached(rc, teamTasksKey.With(user.TeamID), 30*time.Second,
	[]string{"team:" + user.TeamID + ":tasks"},
	func(ctx context.Context) ([]Task, error) { return tasks.ByTeam(ctx, user.TeamID) })
```

Both servers now catch up the same way. On the server the API call reached, the
invalidation marks the list stale, as before; on the other, the 30 seconds running
out does. Either way the next render is served the old list while one fetch in the
background replaces it, and a request after that sees the change. Sending the invalidation to both
servers is the other answer; see
[Invalidation with several instances](#invalidation-with-several-instances).

### A restart with no downtime

A new build goes out one server at a time:

1. **Put the new binary in place** at `/usr/local/bin/tasks`, and run
   restart the service.
2. **The drain.** The service manager sends `SIGTERM`. `/readyz` answers `503 draining`,
   keep-alives are turned off, and the server keeps serving for `DrainDelay`. Within
   2 seconds the load balancer's check fails, and it sends everything to the other
   server.
3. **The stop.** The port closes, and requests still in flight get up to
   `ShutdownTimeout`, 10 seconds. The unit's `TimeoutStopSec=30` is longer than the
   15 seconds this can take.
4. **The new binary.** The service manager starts it. It reads a new `.cache/<hash>` directory,
   so it never serves a page the previous build rendered.
5. **Wait for `/readyz` to answer `200`** and the load balancer to take the server
   back, then do the same on the other.

Readers notice nothing: a form rendered before the restart is accepted after it,
because the key did not change. Once both servers run the new build, remove the
directories earlier builds left in `/srv/tasks/.cache`.

## A checklist

- `COLLAGE_CSRF_KEY` set, the same on every instance.
- `HOST=0.0.0.0` in a container; `PORT` from the platform.
- A writable working directory for `.cache`, and old `.cache/<hash>` directories
  removed on deploy.
- `plugins-config.json` in the working directory, or embedded, if you configure
  plugins.
- `COLLAGE_DEV` unset.
- TLS at the proxy, with `X-Forwarded-Proto` and the browser's `Host` passed on, and
  `Server.TrustedProxies` listing the proxy.
- Any other origin whose forms post here — an admin subdomain — named in
  `Security.CSRFTrustedOrigins`.
- An upload action with its own `WithMaxBodyBytes` and `WithBodyTimeout`, and the
  proxy in front allowing that size and streaming it through — see
  [Uploads](#uploads).
- The platform's stop grace period longer than `Server.DrainDelay +
  Server.ShutdownTimeout`, plus one second with elagoht/jobs.
- The liveness check on `/healthz` and the readiness check on `/readyz`, with
  [elagoht/health](/docs/plugins#elagohthealth).
- `go test ./...` in CI before the build — see [Testing](/docs/testing).
