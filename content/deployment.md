---
description: Build the binary, run it in a container or under systemd, set what production needs, and put it behind TLS.
reference: ServerConfig, CacheConfig, LoadPluginConfig
---

# Deployment

A collage site is a Go program, so what you deploy is a compiled binary. Templates
and static files are embedded in it, so it runs from any directory with nothing
copied beside it — except `plugins-config.json`, if you configure plugins; see
[below](#plugin-configuration). This page is about running that binary on a server; for a site
with no server at all, see [Static export](/docs/static-export).

## Building: `collage build`

```sh
collage build
./bin/mysite
```

`collage build` runs the `go build` you would otherwise have to remember:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/mysite .
```

- **CGO off**, because collage and the standard library need no C, and a static
  binary can go into an image that contains nothing else.
- **`-trimpath`**, so the binary does not carry the paths of the machine that built
  it.
- **`-s -w`** drops the debug tables, which is most of the size.
- **linux/amd64 by default, not this machine.** A binary built for a Mac does not run
  in a Linux container, and `exec format error` on a server is the wrong place to
  find that out.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o path` | `bin/<module name>` | Where the binary is written. |
| `-os name` | `linux` | Target operating system. |
| `-arch name` | `amd64` | Target architecture — `arm64` for Graviton or Ampere machines. |
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
$ collage build -i
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

The scaffolded `main.go` reads four variables. Nothing reads `.env` files in
production — those are for `collage dev` — so set them wherever the binary runs.

| Variable | Default | Set it to |
| --- | --- | --- |
| `HOST` | `localhost` | `0.0.0.0` in a container. `localhost` accepts nothing from outside the machine, which is right behind a local reverse proxy and wrong everywhere else. |
| `PORT` | `3000` | Whatever your platform assigns. The binary's `-port` flag overrides it. |
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

## Graceful shutdown

`app.ListenAndServe` traps `SIGINT` and `SIGTERM`. On either it stops accepting
connections, waits up to `Server.ShutdownTimeout` (10 seconds by default) for
requests in flight to finish, runs every plugin's `Shutdown`, and returns `nil`. A
container runtime that sends `SIGTERM` and waits gets a clean drain with nothing
added. The generated systemd unit sets `TimeoutStopSec=30`, comfortably longer than
the shutdown timeout, so systemd does not kill a process that is still draining.

If the drain runs out of time — a request still open when the timeout passes — the
plugins are shut down anyway, and `ListenAndServe` returns an error saying so
(`collage: server shutdown: context deadline exceeded`), as it does for a plugin
whose `Shutdown` failed. The scaffold's `main.go` passes that to `log.Fatalf`, so
the process exits with status 1 rather than 0; a platform that treats a non-zero
exit on stop as a crash will say so.

If you raise `ShutdownTimeout`, raise your platform's grace period with it.

## Health checks

The project `collage new` scaffolds answers `/healthz` with a small JSON body whose
`status` is `ok`. It is a [document](/docs/documents), not a page, so it involves no
template and cannot start failing because one did. It is dynamic — a document
produced by a handler is, unless it says otherwise, and the scaffold says
`Dynamic()` explicitly — so every check really reaches the process. A health check
served from a cache would answer `ok` long after it stopped being true.

Point your platform's liveness check at it. It tells you the process is up and
serving. A readiness check that should also fail when your database is unreachable
is a document of your own, written the same way; return an error and it answers 500.
A project made with `collage new --template minimal` has no `/healthz`; copy the
demo scaffold's `documents/health.go` if you want one.

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
  instance.

### Invalidation with several instances

`app.InvalidateTags` drops entries from the cache of the process it runs in. With
one instance, a CMS webhook that invalidates a post's tag updates the site. With
several, it updates only the instance the webhook reached; the others keep serving
the old page until its TTL runs out.

Choose one of three answers: give pages that change `Incremental(ttl)` so a stale
copy expires on its own, send the webhook to every instance, or configure a shared
store through `Cache.Store` that implements `collage.TaggedCache`, which lets one
invalidation reach entries any instance wrote. See [Caching](/docs/caching).

A shared store holds pages, not data. Values kept with
[`collage.Cached`](/docs/caching#caching-data-across-pages) live in each process's
memory, whatever `Cache.Store` is, and an invalidation reaches only the instance it
ran in. After a webhook reaches instance A, instance B renders a fresh page — the
shared store dropped it — from the old value it still holds. Give those values a
TTL short enough to live with, or send the invalidation to every instance.

## TLS, behind a proxy

collage serves plain HTTP, and has no `ListenAndServeTLS`. Terminating TLS well means
certificates, renewal, HTTP/2, HSTS and a redirect from port 80, and every platform
this runs on — a load balancer, a reverse proxy, Cloudflare, Fly, Render — already
does it better than a framework flag could.

Put the binary behind one, and make sure the proxy sends `X-Forwarded-Proto: https`.
collage uses it to mark the forgery-token cookie `Secure`, so a site served over
HTTPS never sends that cookie over plain HTTP. A minimal Caddy configuration, which
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
did for you, and you call `app.Shutdown(ctx)` yourself so plugins shut down.

## Timeouts

`ListenAndServe` applies the timeouts in `Config.Server`:

| Field | Default | Bounds |
| --- | --- | --- |
| `ReadTimeout` | 15s | Reading a request, headers included — a client that sends headers slowly cannot hold a connection open. |
| `WriteTimeout` | 30s | Writing the response. A page whose data takes longer than this is cut off. |
| `IdleTimeout` | 60s | A keep-alive connection waiting for its next request. |
| `ShutdownTimeout` | 10s | The graceful drain on `SIGTERM`. |
| `MaxBodyBytes` | 4 MiB | An action's request body, unless the action sets its own. Negative is unbounded. |

```go
Server: collage.ServerConfig{
	Host:         envString("HOST", "localhost"),
	Port:         port,
	WriteTimeout: time.Minute,
},
```

The time a data handler may take is a different setting: each fragment's
`WithTimeout`, or `Template.Timeout` (5 seconds) for fragments and documents that
set none. Keep it well under `WriteTimeout`, so a slow upstream turns into a
fragment's fallback rather than a connection closed halfway through the page. See
[Configuration](/docs/configuration).

## Logs

With no `Config.Logger`, collage logs to `slog`'s default handler — or, on a
terminal, to one formatted for a person. On a server, where logs are read by a
machine, pass a JSON handler:

```go
Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
```

## A checklist

- `COLLAGE_CSRF_KEY` set, the same on every instance.
- `HOST=0.0.0.0` in a container; `PORT` from the platform.
- A writable working directory for `.cache`, and old `.cache/<hash>` directories
  removed on deploy.
- `plugins-config.json` in the working directory, or embedded, if you configure
  plugins.
- `COLLAGE_DEV` unset.
- TLS at the proxy, with `X-Forwarded-Proto` passed on.
- The platform's stop grace period longer than `Server.ShutdownTimeout`.
- The liveness check on `/healthz`.
- `go test ./...` in CI before the build — see [Testing](/docs/testing).
