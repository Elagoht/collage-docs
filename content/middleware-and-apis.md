---
description: Standard net/http middleware with app.Use, cache keys that depend on a header with collage.Vary, and your own http.Handler with app.Handle.
reference: Vary, SkipCache, ErrVaryTooLate, ErrMountShadowsRoute, WithMiddleware, HandleOption
---

# Middleware and your own API

collage renders pages. The rest of what a Go program does over HTTP — an API,
authentication, language negotiation, rate limiting — you write the way you already
would, and plug in at one of two points:

- **Middleware**, which runs around every request but those for your own handler.
- **Your own handler**, which answers every request under a prefix, and brings its
  own middleware.

## Middleware: `app.Use`

```go
if err := app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}); err != nil {
	return err
}
```

It is the standard `func(http.Handler) http.Handler` shape, so any middleware written
for `net/http` works unchanged. The first one registered is the outermost.
`collage add middleware loadUser` writes one and registers it after those already
there (since v0.69.0).

Like every registration it must happen before the application starts; afterwards it
returns `collage.ErrAppStarted`.

Two things about where it runs:

- **Before routing.** It sees every request — pages, documents, actions, mounted
  static files, a plugin's handlers and a handler mounted `WithMiddleware` —
  including the ones that end in a 404. It can answer a request by itself, with a
  401 or a redirect, and nothing behind it runs. Since v0.67.0 a handler mounted
  with plain `app.Handle` is the exception: it is answered without any middleware
  ([below](#no-middleware-unless-you-ask)).
- **Inside the framework's guard.** It runs within collage's span, metrics and panic
  recovery. That is the difference from wrapping `app.Handler()` in middleware of
  your own: a panic in your middleware is an ordinary 500 on the normal error path
  instead of a dropped connection, and a request it answers itself is counted like
  any other.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 376" role="img" aria-labelledby="mw-layers-title">
<title id="mw-layers-title">Inside collage's guard of span, metrics and panic recovery, a request passes the middleware of the plugins in Config.Plugins in list order, then the app.Use middleware in registration order. Routing then checks static mounts and the app.Handle prefixes mounted WithMiddleware first, and hands everything else to the router, which picks a page, one of its fragment paths, a document or an action. A handler mounted with plain app.Handle is matched before the middleware, inside the guard, and runs without it.</title>
<defs><marker id="mw-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="74" y="8" width="100" height="26" rx="13" /><text x="124" y="25" text-anchor="middle">browser</text>
<rect class="d-slot d-slot-a" x="6" y="56" width="308" height="312" rx="8" />
<line class="d-flow" x1="124" y1="34" x2="124" y2="96" marker-end="url(#mw-arrow)" />
<text class="d-slot-text-a d-small" x="18" y="72">collage guard</text><text class="d-small" x="302" y="72" text-anchor="end">span · metrics · panic recovery</text>
<path class="d-flow" d="M124 86H271V96" marker-end="url(#mw-arrow)" />
<rect class="d-frame" x="22" y="100" width="204" height="38" rx="6" /><text x="124" y="116" text-anchor="middle">plugin middleware</text><text class="d-small" x="124" y="130" text-anchor="middle">Config.Plugins, in list order</text>
<line class="d-flow" x1="124" y1="138" x2="124" y2="152" marker-end="url(#mw-arrow)" />
<rect class="d-frame" x="22" y="156" width="204" height="38" rx="6" /><text x="124" y="172" text-anchor="middle">app.Use middleware</text><text class="d-small" x="124" y="186" text-anchor="middle">in registration order</text>
<rect class="d-frame" x="236" y="100" width="70" height="94" rx="6" /><text x="271" y="143" text-anchor="middle">app.Handle</text><text class="d-small" x="271" y="157" text-anchor="middle">bare</text>
<path class="d-flow" d="M124 194C124 204 160 200 160 208" marker-end="url(#mw-arrow)" />
<rect class="d-cmd" x="110" y="212" width="100" height="26" rx="13" /><text class="d-cmd-text" x="160" y="229" text-anchor="middle">routing</text>
<path class="d-flow" d="M135 238C135 250 61 246 61 258" marker-end="url(#mw-arrow)" />
<line class="d-flow" x1="160" y1="238" x2="160" y2="258" marker-end="url(#mw-arrow)" />
<path class="d-flow" d="M185 238C185 250 259 246 259 258" marker-end="url(#mw-arrow)" />
<rect class="d-b" x="15" y="262" width="92" height="30" rx="5" /><text class="d-on-b d-small" x="61" y="281" text-anchor="middle">static mount</text>
<rect class="d-frame" x="114" y="262" width="92" height="30" rx="5" /><text class="d-small" x="160" y="274" text-anchor="middle">app.Handle</text><text class="d-small" x="160" y="286" text-anchor="middle">WithMiddleware</text>
<rect class="d-cmd" x="213" y="262" width="92" height="30" rx="15" /><text class="d-cmd-text" x="259" y="281" text-anchor="middle">router</text>
<text class="d-small" x="110" y="308" text-anchor="middle">matched by prefix first</text>
<path class="d-flow" d="M245 292C245 306 210 304 210 318" marker-end="url(#mw-arrow)" />
<path class="d-flow" d="M273 292C273 306 279 304 279 318" marker-end="url(#mw-arrow)" />
<rect class="d-a" x="176" y="322" width="68" height="34" rx="5" /><text class="d-on-a d-small" x="210" y="336" text-anchor="middle">page or</text><text class="d-on-a d-small" x="210" y="348" text-anchor="middle">document</text>
<rect class="d-c" x="249" y="322" width="60" height="34" rx="5" /><text class="d-on-c" x="279" y="343" text-anchor="middle">action</text>
</svg>
<figcaption>Every request runs inside collage's guard: its span, its metrics, and the panic recovery that turns a panic into a 500. Inside it, the middleware of the plugins in Config.Plugins comes first, in list order, then the app.Use middleware in the order it was registered; a plugin added with app.RegisterPlugin sits where the call was made. Any of them can answer by itself and stop the request there. What gets through is routed: a static mount or the prefix of an app.Handle mounted WithMiddleware is matched first, and the router takes the rest. A handler mounted with plain app.Handle is not in the chain at all: its requests are matched before the middleware, still inside the guard, and go straight to it.</figcaption>
</figure>

### Plugins' middleware and yours

Since v0.38.0 a plugin's middleware counts as registered where the plugin was. The
plugins in `Config.Plugins` are registered in `collage.New`, so their middleware is
outside everything `app.Use` adds.

Yours therefore reads what they put in the request's context: the session, a flash
message. Loading the signed-in user once per request is one middleware:

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := session.FromContext(r.Context()).Get(session.UserKey)
		user, err := users.Active(r.Context(), id) // nil for a disabled account
		if err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
	})
})
```

A plugin that should read what your middleware put there, a rate limit keyed on
that user, is registered after it with `app.RegisterPlugin`. Its middleware goes
where the call is made: inside what `app.Use` added before it, outside what it adds
after.

```go
app.Use(loadUser)
app.RegisterPlugin(ratelimit.New(ratelimit.Options{KeyFunc: userID}))
```

`RegisterPlugin` refuses a plugin that adds template functions; that one has to be
in `Config.Plugins`.

Before v0.38.0 every plugin's middleware ran inside the
application's, and `session.FromContext` was `nil` in `app.Use` middleware.

### Paths are cleaned first

A path with dot segments or doubled slashes — `/a/../b`, `/a//b`, `/a/./b` — is
redirected to its clean spelling before anything reads it, middleware and plugins
included: 301 for `GET` and `HEAD`, 308 for a method that carries a body, the query
kept (since v0.24.0).

A middleware that skips `/_collage/` or protects `/admin/` therefore never meets
`/_collage/../admin`, and a prefix check is a check on the path the router will
route.

An encoded slash is not a separator to the router, so since v0.34.0:

- A path holding `%2F` reaches no route: the router answers it 404.
- One that is also dirty — `/blog/%2Fhello` — is answered 404 before your
  middleware runs rather than cleaned into `/blog/hello`.
- Your middleware reads the decoded `r.URL.Path`, where `/public%2Fsecret` is
  `/public/secret`; before, the router read one segment there, and a middleware
  letting `/public/` through let the request through to `/{slug}`.

### The body is bounded before it

Since v0.34.0 a request's body is limited before the first middleware runs, at the
limit of the [action](/docs/forms-and-actions#request-bodies-are-bounded) it routes
to, or `Server.MaxBodyBytes` for anything else.

- A middleware that reads the body — a signature check, a logger, a method override
  parsing the form — reads it bounded.
- An action whose body a middleware already read past its limit answers `413`.
- A handler mounted with `app.Handle` is left alone.

### Passing values to data handlers

What middleware puts in the request's context is what data handlers receive as
`ctx`. That is how a signed-in user, a feature flag or a tenant reaches the
fragments that need it:

```go
type userKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, ok := sessions.User(r); ok {
			r = r.WithContext(context.WithValue(r.Context(), userKey{}, user))
		}
		next.ServeHTTP(w, r)
	})
})
```

```go
func accountData(ctx context.Context, rc *collage.RenderContext) (accountView, []string, error) {
	user, ok := ctx.Value(userKey{}).(User)
	if !ok {
		return accountView{}, nil, nil // signed out: the fragment renders its signed-out state
	}
	return accountView{Name: user.Name}, nil, nil
}
```

Mind the cache. A page that renders differently per user must not be cached by URL,
or the first reader's version is everyone's. Either:

- keep it dynamic — as a page with a data handler and no declared strategy already
  is — rather than giving it `Static()` or `Incremental(ttl)`, or
- tell the cache what it varies on with `collage.Vary` below.

A cached render sees neither the reader's cookies and credentials nor the values
your middleware put in the context ([what a shared render
sees](/docs/caching#what-a-shared-render-sees)): since v0.39.0 a per-user page you
cached by mistake renders signed-out for everyone — the `userKey` value hidden —
rather than showing the first reader's account to all. Keep it dynamic.

A static export renders without a request, so no middleware runs for the files it
writes. A data handler reading a context value must cope with its absence — which it
has to anyway, for a reader who is not signed in.

Since v0.52.0 the build may then
[capture each file's headers](/docs/static-export#headers-and-redirects) through
the handler, and those requests do pass through middleware:

- Middleware that counts or limits traffic skips one when
  `collage.IsCapture(r.Context())` is true.
- Middleware that sets headers handles it as usual.

## Content that depends on a header: `collage.Vary`

The page cache is keyed by URL. A cached page whose content depends on a request
header — `Accept-Language`, a device class, an A/B cohort — serves whichever version
was rendered first to everybody. `collage.Vary`, called from middleware, adds a
dimension to the cache key:

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		if err := collage.Vary(r, "Accept-Language", lang); err != nil {
			app.Logger().Error("vary", "error", err)
		}
		next.ServeHTTP(w, r)
	})
})
```

A fragment's data handler reads the value back with `collage.Varied`:

```go
func languageData(ctx context.Context, rc *collage.RenderContext) (string, []string, error) {
	lang, _ := collage.Varied(ctx, "Accept-Language") // "tr" or "en"
	return greeting(lang), nil, nil
}
```

- **The value you resolved goes into the key, not the raw header.** Browsers spell
  their preferences a hundred ways — `tr-TR,tr;q=0.9`, `tr`, `tr-TR` — and all of them
  are one page. Reduce the header to the handful of values your pages actually
  differ by, and the cache holds that many entries. Never vary on `Cookie` or
  `Authorization` themselves: since v0.64.0 the value is kept in memory and sent
  again with every [warm render](/docs/caching#warming-after-an-invalidation).
- **Read it with `collage.Varied`, not from the context.** On a cacheable page one
  render is served to many readers, so a value your middleware puts in the request
  context with `context.WithValue` is stripped from that render — it would
  otherwise be the first reader's, frozen for the rest. A `Vary` value is safe to
  read because it is in the key, so each value has its own cached page. See
  [what a shared render sees](/docs/caching#what-a-shared-render-sees).
- **The header name goes into the response's `Vary` header**, so a CDN or proxy
  between you and the reader keeps the versions apart too. It is only set on
  publicly cacheable responses; a `no-store` response has nothing to keep apart.
  Since v0.34.0 it is added beside a `Vary` your middleware wrote, where before it
  replaced it.
- **Call it from middleware.** Declarations close when middleware is done and
  routing begins, on every route — a page, cached or not, a document, an action, a
  mount, an `app.Handle` handler. Called after routing, from a data handler say,
  `Vary` returns `collage.ErrVaryTooLate` instead of pretending to work (since
  v0.11.0; before, it did so only on a cached page and silently did nothing
  elsewhere). Called on a request collage is not serving, it returns
  `collage.ErrVaryOutsideRequest`.
- Declaring the same header twice keeps the last value.

This is also how to negotiate a language: collage selects a locale from the URL
only, and leaves reading `Accept-Language` to you. Redirect a browser to `/tr` from
middleware, or render one URL per language and declare it with `Vary`. See
[Links and locales](/docs/links-and-locales).

## Your own handler: `app.Handle`

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", getUser)
api.HandleFunc("POST /api/users", createUser)

if err := app.Handle("/api/", api); err != nil {
	return err
}
```

Every request whose path begins with the prefix goes to your handler, with the path
unchanged — `getUser` sees `/api/users/42`. Wrap it in `http.StripPrefix` if it
expects otherwise.

Any `http.Handler` will do: `http.ServeMux`, chi, a gRPC gateway, a reverse proxy.

**collage does nothing to it.** What it accepts, how much it reads and what it
caches are yours to decide.

| The handler does not get | The handler does get, like every request |
| --- | --- |
| a request-forgery check | the plugins' request hooks |
| a body-size limit | the span and the metrics |
| a cache | the panic guard |
| middleware, yours or a plugin's — unless mounted `WithMiddleware()` | the clean-path redirect and the encoded-slash 404 |
| | the baseline `X-Content-Type-Options` and `X-Frame-Options` headers |
| | draining on graceful shutdown |

A 5xx it answers with is reported to plugins' error hooks as
`collage.ErrHandlerFailed`; its 4xx answers are its own business.

### No middleware, unless you ask

Since v0.67.0 no middleware runs for a handler mounted with `app.Handle`: neither
what `app.Use` added nor any plugin's — session, flash, i18n, consent, secure,
ratelimit, basicauth, fail2ban, accesslog, redirects, tenant, compress, the
development toolbar.

Your handler, your middleware. An API has its own idea of authentication — a bearer
token, not the site's session cookie — its own rate limit and its own error format,
and the site's consent banner, development toolbar or `Vary: Cookie` are nothing it
asked for. So wrap the handler in what it wants, the way you would without collage:

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", getUser)

// requireToken is yours, below; perClientLimit stands for any net/http rate
// limiter — golang.org/x/time/rate behind a map, httprate, tollbooth.
if err := app.Handle("/api/", requireToken(perClientLimit(api))); err != nil {
	return err
}

func requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validToken(r.Header.Get("Authorization")) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

A handler that does rely on the site's middleware — a basicauth plugin you
registered to guard it, the reader's session, a rate limit you already configured,
a value your `app.Use` middleware puts in the context — asks for it back with
`collage.WithMiddleware()`. It then runs inside the chain exactly as a page does, in
the same order, as every mount did before v0.67.0:

```go
if err := app.Handle("/admin/api/", adminAPI, collage.WithMiddleware()); err != nil {
	return err
}
```

**No plugin can answer for a plain mount.** A ban, an IP allowlist, a maintenance
page or an access log over the whole site does not cover it; mount it
`WithMiddleware()` for that, or put your own middleware around your mux. In
particular:

- elagoht/fail2ban still counts a plain mount's probes — its request hook runs for
  every request — but the 403 it answers a banned client with is middleware, so
  that client keeps full access to `/api/`, which is where login brute force
  usually happens.
- elagoht/basicauth and elagoht/ratelimit do not guard it, and elagoht/session's
  session is not in its requests.
- elagoht/accesslog does not log its traffic.
- An elagoht/redirects rule under its prefix does not fire.
- elagoht/tenant's tenant is not in its context.

An application that relied on one of these for a mount before v0.67.0 adds
`collage.WithMiddleware()` to that call to keep it. Without it, a mount a basicauth
or a rate limit used to guard is now open.

Which requests skip the middleware is decided on the path once it is
[cleaned](#paths-are-cleaned-first), by the same match routing uses, so a request
that skips it can only reach that handler. Two consequences for middleware that
rewrites paths:

- One that rewrites a request's path *into* a plain mount's prefix no longer
  reaches it: the rewritten request is routed like everything else the chain hands
  on, to a 404. Mount the target `WithMiddleware()`.
- One that rewrites a path *out of* a plain mount's prefix — `/api/legacy` to a page
  — no longer runs for it: the mount claims the request first. Mount it
  `WithMiddleware()` if a middleware should see its requests.

A plugin's own endpoints — a live-reload stream, an OAuth callback, a Prometheus
`/metrics`, an Open Graph image — are mounted with `Host.Handle`, and those always
run inside the middleware: they are part of the site, and may read the session or
rely on a basicauth in front of them.

### Prefixes and conflicts

The prefix must begin with `/` and cannot be `/` alone
(`collage.ErrInvalidPrefix`).

- **Ending in `/`**, it claims every path beneath it.
- **Without the trailing slash** it is one exact path, and answers that path alone —
  `app.Handle("/metrics", h)` serves `/metrics` and not `/metrics/x` (since
  v0.24.0; before, the prefix had to end in `/`).

A handler at `/` would take every request from every page; if that is really what
you want, put `app.Handler()` inside a mux of your own instead.

A prefix may not cover anything collage routes:

- `app.Handle("/api/", ...)` next to an action at `/api/count` is refused with
  `collage.ErrMountShadowsRoute`.
- Two handlers, or a handler and a static mount, that overlap are refused with
  `collage.ErrMountConflict`.

The check runs when the application starts, so it does not matter which was
registered first; `app.ListenAndServe` and `app.Start` return the error.

The check compares literal paths. It cannot see a dynamic pattern that would match
under the prefix: with a catch-all page at `/{rest...}` and a handler at `/api/`, the URL
`/api/users` goes to the handler, which is the trade you make by mounting a prefix.

### Invalidating from your API

Your handler can invalidate cached pages like anything else can. When an API call changes
content, invalidate the tags the pages built from it declared:

```go
func updatePost(app *collage.App, posts *store.Posts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if err := posts.Update(r.Context(), slug, r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.InvalidateTags(r.Context(), "post:"+slug, "blog:posts"); err != nil {
			app.Logger().Error("invalidate", "slug", slug, "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
```

`app.InvalidateTagsN` does the same and also returns how many cache entries the tags
reached, which is useful in a webhook's response or a log line.

A CMS webhook is the usual case:

1. The CMS calls `/api/hooks/cms` with the changed entry.
2. The handler checks the webhook's signature and invalidates that entry's tag.

Only that entry's pages, and whatever `collage.Cached` values carried the same tag,
are rendered again. See [Caching](/docs/caching).

## When to use an action instead

`app.Handle` is for code that is not about your pages — an API with its own
authentication, a service you proxy to, a router you already have. For an endpoint
that belongs to the site, use an [action](/docs/forms-and-actions):

- A form post, or a `fetch()` from one of your pages. An action checks the
  request-forgery token; a handler does not, so a browser-facing `POST` through
  `app.Handle` is one any other site can make on your reader's behalf.
- Anything that should be bounded. An action's body is limited by
  `Server.MaxBodyBytes` (4 MiB by default) or its own `WithMaxBodyBytes`.
- An endpoint that answers with a fragment or a page — the changed part of a form,
  a validation failure — since an action can return `collage.RenderFragment` or
  `collage.RenderPage`.
- An endpoint that invalidates cached pages when it succeeds, through
  `ActionResult.InvalidateTags`.
- A URL that should exist in every locale the site has: an action's paths are
  keyed by locale, like a page's.

```go
count := collage.NewAction("count").
	WithPath("en", "/api/count").
	WithMethods(http.MethodPost).
	WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
		result, err := collage.JSON(http.StatusOK, countResponse{Count: store.Increment()})
		if err != nil {
			return nil, err
		}
		result.InvalidateTags = []string{store.CountTag}
		return result, nil
	}).
	Build()
```

That is the counter from the project `collage new` scaffolds: a JSON endpoint that
needs the forgery token and invalidates the page showing the count. For a webhook
from a service that cannot send a token, an action with `WithoutCSRF()` still gets
the body limit.

## Example: a team's task list

A team's task list, where every member signs in and has a role: an admin may change
any task of the team, a member only the tasks assigned to them. Pages show the
team's tasks, and a JSON API that the team's command-line tool already talks to is
kept as it is. Three pieces of this page meet here:

- one middleware that finds out who is asking, once per request;
- the existing API, mounted with `app.Handle` inside the middleware;
- the API invalidating the tag of the list it changed.

### Who is asking, once per request

The middleware reads either an API token or the session cookie, and puts the user
in the request's context under a key of its own type:

```go
type User struct {
	ID      string
	TeamID  string
	Role    string // "admin" or "member"
	ByToken bool   // signed in with an API token, not the session cookie
}

// Users is your own data layer.
type Users interface {
	BySession(ctx context.Context, cookie string) (User, bool, error)
	ByToken(ctx context.Context, token string) (User, bool, error)
}

type userKey struct{}

// UserFrom returns the signed-in user; false for a reader who is not.
func UserFrom(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey{}).(User)
	return user, ok
}

// loadUser resolves who is asking once per request. Pages, actions and the
// API behind it read the result with UserFrom.
func loadUser(users Users) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				user User
				ok   bool
				err  error
			)
			if token, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); bearer {
				user, ok, err = users.ByToken(r.Context(), token)
				user.ByToken = true
			} else if cookie, cerr := r.Cookie("sid"); cerr == nil {
				user, ok, err = users.BySession(r.Context(), cookie.Value)
			}
			if err != nil {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			if ok {
				r = r.WithContext(context.WithValue(r.Context(), userKey{}, user))
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

Because middleware runs before routing, the same lookup serves a page, an action
and the mounted API alike; nothing behind it asks the database who the reader is
again. The API gets it only because it is mounted `WithMiddleware()`, below: a plain
`app.Handle` mount runs no middleware, and `UserFrom` would find no one there.

### The API, mounted as it is

```go
type Task struct {
	ID         string
	TeamID     string
	AssigneeID string
	Title      string
	Done       bool
}

// Tasks is your own data layer.
type Tasks interface {
	Get(ctx context.Context, id string) (Task, bool, error)
	SetDone(ctx context.Context, id string, done bool) error
	ByTeam(ctx context.Context, teamID string) ([]Task, error)
}

// taskAPI is the JSON API the team's command-line tool already talks to.
func taskAPI(app *collage.App, tasks Tasks) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("PATCH /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok || !user.ByToken {
			http.Error(w, "token required", http.StatusUnauthorized)
			return
		}
		task, found, err := tasks.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if !found || task.TeamID != user.TeamID {
			http.NotFound(w, r)
			return
		}
		if user.Role != "admin" && task.AssigneeID != user.ID {
			http.Error(w, "not your task", http.StatusForbidden)
			return
		}
		var body struct {
			Done bool `json:"done"`
		}
		// collage does not bound a mounted handler's body; this is ours to do.
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := tasks.SetDone(r.Context(), task.ID, body.Done); err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := app.InvalidateTags(r.Context(), "team:"+task.TeamID+":tasks"); err != nil {
			app.Logger().Error("invalidate", "task", task.ID, "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return api
}
```

Each check in it follows from what `app.Handle` does and does not do:

- **A token, not the cookie.** A handler has no request-forgery check, so a write
  it accepts on the strength of the session cookie is unprotected, and a
  browser-facing `POST` through it is one any other site can make on the reader's
  behalf. The API accepts only requests signed in with a token; the pages change
  tasks through an [action](/docs/forms-and-actions), which checks the forgery
  token.
- **The body is bounded by hand.** collage leaves a mounted handler's body alone, so
  the handler wraps it in `http.MaxBytesReader` itself.
- **The role is checked where the data is.** The middleware only says who is
  asking; what that user may change is the API's decision.
- **The actions live outside `/api/`.** A prefix may not cover a route, so an
  action at `/api/tasks/new` next to this mount would fail the start with
  `collage.ErrMountShadowsRoute`.

### The list the API invalidates

The task list fragment reads the user from the context and keeps the team's list in
the [data cache](/docs/caching#caching-data-across-pages), tagged with the team:

```go
var teamTasksKey = collage.NewKey[[]Task]("team-tasks")

// taskListData feeds the task list fragment. The page has no declared
// strategy, so it is dynamic and sees the user; the list itself is shared
// by everyone on the team and fresh until the API changes it.
func taskListData(tasks Tasks) func(context.Context, *collage.RenderContext) ([]Task, []string, error) {
	return func(ctx context.Context, rc *collage.RenderContext) ([]Task, []string, error) {
		user, ok := UserFrom(ctx)
		if !ok {
			return nil, nil, nil // signed out: the fragment links to the sign-in page
		}
		list, err := collage.Cached(rc, teamTasksKey.With(user.TeamID), 0,
			[]string{"team:" + user.TeamID + ":tasks"},
			func(ctx context.Context) ([]Task, error) { return tasks.ByTeam(ctx, user.TeamID) })
		return list, nil, err
	}
}

func taskList(tasks Tasks) *collage.Fragment {
	return collage.NewFragment("task-list", "fragments/task-list.html").
		WithData(collage.LoadTagged(taskListData(tasks))).
		Build()
}
```

The page stays dynamic, because it renders per user. The list behind it is the same
for the whole team, so it is fetched once and shared: a `ttl` of zero keeps it fresh
until something invalidates its tag.

When the command-line tool marks a task done, the API's `InvalidateTags` call marks
that team's list stale. The next render is served the old list at once while one
fetch in the background replaces it, and a request after that fetch has finished
sees the change ([how it behaves](/docs/caching#how-it-behaves)). Other teams'
lists are not touched.

Wiring it up is two calls:

```go
if err := app.Use(loadUser(users)); err != nil {
	return err
}
if err := app.Handle("/api/", taskAPI(app, tasks), collage.WithMiddleware()); err != nil {
	return err
}
```

`collage.WithMiddleware()` puts the API inside the chain, so `loadUser` runs for it,
and so does every plugin's middleware: the rate limit
[Security](/docs/security#example-a-teams-task-list-before-it-goes-live) adds counts
its writes too. Without it every call would be answered `401 token required`, since
no one would have put the user in the context.
