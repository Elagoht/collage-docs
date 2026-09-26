---
description: Standard net/http middleware with app.Use, cache keys that depend on a header with collage.Vary, and your own http.Handler with app.Handle.
reference: Vary, SkipCache, ErrVaryTooLate, ErrMountShadowsRoute
---

# Middleware and your own API

collage renders pages. The rest of what a Go program does over HTTP — an API,
authentication, language negotiation, rate limiting — you write the way you already
would, and plug in at one of two points: **middleware**, which runs around every
request, and **your own handler**, which answers every request under a prefix.

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
for `net/http` works unchanged. The first one registered is the outermost. Like every
registration it must happen before the application starts; afterwards it returns
`collage.ErrAppStarted`.

Two things about where it runs:

- **Before routing.** It sees every request — pages, documents, actions, mounted
  static files and your own handlers — including the ones that end in a 404. It can
  answer a request by itself, with a 401 or a redirect, and nothing behind it runs.
- **Inside the framework's guard.** It runs within collage's span, metrics and panic
  recovery. That is the difference from wrapping `app.Handler()` in middleware of
  your own: a panic in your middleware is an ordinary 500 on the normal error path
  instead of a dropped connection, and a request it answers itself is counted like
  any other.

### Paths are cleaned first

A path with dot segments or doubled slashes — `/a/../b`, `/a//b`, `/a/./b` — is
redirected to its clean spelling before anything reads it, middleware and plugins
included: 301 for `GET` and `HEAD`, 308 for a method that carries a body, the query
kept (since v0.24.0). A middleware that skips `/_collage/` or protects `/admin/`
therefore never meets `/_collage/../admin`, and a prefix check is a check on the
path the router will route.

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
func accountData(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	user, ok := ctx.Value(userKey{}).(User)
	if !ok {
		return accountView{}, nil, nil // signed out: the fragment renders its signed-out state
	}
	return accountView{Name: user.Name}, nil, nil
}
```

Mind the cache. A page that renders differently per user must not be cached by URL,
or the first reader's version is everyone's: keep it dynamic — as a page with a
data handler and no declared strategy already is — rather than giving it `Static()`
or `Incremental(ttl)`, or tell the cache what it varies on with `collage.Vary`
below.

A static export renders without a request, so no middleware runs during one. A data
handler reading a context value must cope with its absence — which it has to anyway,
for a reader who is not signed in.

## Content that depends on a header: `collage.Vary`

The page cache is keyed by URL. A cached page whose content depends on a request
header — `Accept-Language`, a device class, an A/B cohort — serves whichever version
was rendered first to everybody. `collage.Vary`, called from middleware, adds a
dimension to the cache key:

```go
type langKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		if err := collage.Vary(r, "Accept-Language", lang); err != nil {
			app.Logger().Error("vary", "error", err)
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), langKey{}, lang)))
	})
})
```

- **The value you resolved goes into the key, not the raw header.** Browsers spell
  their preferences a hundred ways — `tr-TR,tr;q=0.9`, `tr`, `tr-TR` — and all of them
  are one page. Reduce the header to the handful of values your pages actually
  differ by, and the cache holds that many entries.
- **The header name goes into the response's `Vary` header**, so a CDN or proxy
  between you and the reader keeps the versions apart too. It is only set on
  publicly cacheable responses; a `no-store` response has nothing to keep apart.
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
expects otherwise. Any `http.Handler` will do: `http.ServeMux`, chi, a gRPC gateway,
a reverse proxy.

**collage does nothing to it.** There is no request-forgery check, no body-size
limit and no cache. What it accepts, how much it reads and what it caches are yours
to decide. What it does get is what every request gets: the span, the metrics, the
panic guard, the middleware registered with `app.Use`, and draining on graceful
shutdown. A 5xx it answers with is reported to plugins' error hooks as
`collage.ErrHandlerFailed`; its 4xx answers are its own business.

### Prefixes and conflicts

The prefix must begin with `/` and cannot be `/` alone
(`collage.ErrInvalidHandlerPrefix`). Ending in `/`, it claims every path beneath
it. Without the trailing slash it is one exact path, and answers that path alone —
`app.Handle("/metrics", h)` serves `/metrics` and not `/metrics/x` (since v0.24.0;
before, the prefix had to end in `/`).

A handler at `/` would take every request from every page; if that is really what
you want, put `app.Handler()` inside a mux of your own instead.

A prefix may not cover anything collage routes. `app.Handle("/api/", ...)` next to
an action at `/api/count` is refused with `collage.ErrMountShadowsRoute`, and two
handlers, or a handler and a static mount, that overlap are refused with
`collage.ErrMountConflict`. The check runs when the application starts, so it does
not matter which was registered first; `app.ListenAndServe` and `app.Start` return
the error.

The check compares literal paths. It cannot see a dynamic pattern that would match
under the prefix: with a catch-all page at `/{rest...}` and a handler at `/api/`, the URL
`/api/users` goes to the handler, which is the trade you make by mounting a prefix.

### Invalidating from your API

Your handler can drop cached pages like anything else can. When an API call changes
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

A CMS webhook is the usual case: the CMS calls `/api/hooks/cms` with the changed
entry, the handler checks the webhook's signature and invalidates that entry's tag.
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
- An endpoint that drops cached pages when it succeeds, through
  `ActionResult.InvalidateTags`.
- A URL that should exist in every locale the site has: an action's paths are
  keyed by locale, like a page's.

```go
count := collage.NewAction("count").
	WithPath("en", "/api/count").
	WithMethods(http.MethodPost).
	WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
		result, err := collage.JSONOf(http.StatusOK, countResponse{Count: store.Increment()})
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
