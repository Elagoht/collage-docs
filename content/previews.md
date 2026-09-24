---
description: Show editors unpublished drafts on the real site, without the cache serving them the published page or anyone else the draft.
reference: SkipCache, Cached, NewAction, ErrVaryTooLate
---

# Previews

An editor who presses "preview" in a CMS wants the draft, on the real site, with the
real layout around it. Caching gets in the way twice. The editor must not be served
the published page the cache already holds, and the draft they are shown must never
be cached — or it becomes the page every other reader gets.

collage provides one primitive for this, `collage.SkipCache`, and leaves the rest to
you: who may preview is a question about your users, and collage has no opinion about
sessions.

## `collage.SkipCache`

```go
func SkipCache(r *http.Request) error
```

Called from middleware, it marks one request as a preview. That request gets:

- **A fresh render.** The page cache is not read, so a cached copy of the published
  page is never served.
- **Nothing stored.** The render is not written to the page cache.
- **`Cache-Control: private, no-store`**, so no browser cache, proxy or CDN keeps it
  either.
- **Fresh data.** Every `collage.Cached` call in that render fetches instead of
  returning a stored value, and stores nothing. An editor shown the draft page
  around a cached copy of the published data has not been shown the draft.

Documents honour it too: a document requested in a preview is neither read from
nor written to the cache. Every other request is untouched: it is served
from the cache as usual, and the cache still holds the published version.

Like `collage.Vary`, it must be called before routing, which means from middleware
registered with `app.Use`. Called after routing — from a data handler, say — it
returns `collage.ErrVaryTooLate`, on every route (since v0.11.0).

## A complete preview flow

The usual arrangement, and the one below:

1. The CMS's preview button opens `/api/preview?secret=…&slug=…` in a new tab.
2. An [action](/docs/forms-and-actions) at that URL checks the secret, sets a signed
   cookie and redirects to the post.
3. Middleware sees the cookie on every later request, calls `collage.SkipCache`, and
   tells data handlers to ask the CMS for drafts.
4. Another action clears the cookie when the editor is done.

Two secrets are involved, both from the environment: `PREVIEW_SECRET`, shared with
the CMS so it can start a preview, and `PREVIEW_KEY`, known only to your server,
which signs the cookie.

### Signing the cookie

The cookie holds when it expires, and a signature over that. Nobody without the key
can make one, and one that has expired is refused even if the browser still sends
it.

```go
package preview

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"
)

// CookieName is the cookie a preview is carried in.
const CookieName = "preview"

// key signs preview cookies. At least 32 random bytes; empty turns previews off.
var key = []byte(os.Getenv("PREVIEW_KEY"))

// sign returns a cookie value that is valid until expires.
func sign(expires time.Time) string {
	stamp := strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(stamp))
	return stamp + "." + hex.EncodeToString(mac.Sum(nil))
}

// Valid reports whether value is a preview cookie this server signed and that has
// not expired.
func Valid(value string) bool {
	if len(key) == 0 {
		return false
	}
	stamp, _, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	unix, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return false
	}
	expires := time.Unix(unix, 0)
	return time.Now().Before(expires) && hmac.Equal([]byte(value), []byte(sign(expires)))
}
```

### The action that starts a preview

```go
// Start is the URL the CMS's preview button opens:
// GET /api/preview?secret=...&slug=...
func Start(app *collage.App) *collage.Action {
	return collage.NewAction("preview-start").
		WithPath("en", "/api/preview").
		WithMethods(http.MethodGet).
		WithHandler(func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			query := rc.Request.URL.Query()
			secret := os.Getenv("PREVIEW_SECRET")
			if len(key) == 0 || secret == "" ||
				!hmac.Equal([]byte(query.Get("secret")), []byte(secret)) {
				return collage.NoContent(http.StatusUnauthorized), nil
			}

			// Built from the page's name, so a slug cannot send the editor anywhere
			// but a blog post.
			target, err := app.URL("blog-post", "", map[string]string{"slug": query.Get("slug")})
			if err != nil {
				return collage.NoContent(http.StatusBadRequest), nil
			}

			expires := time.Now().Add(time.Hour)
			cookie := &http.Cookie{
				Name:     CookieName,
				Value:    sign(expires),
				Path:     "/",
				Expires:  expires,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}

			result := collage.SeeOther(target)
			result.Header = http.Header{}
			result.Header.Add("Set-Cookie", cookie.String())
			return result, nil
		}).
		Build()
}
```

A `GET` action needs no forgery token — only unsafe methods are checked — which is
what lets the CMS open it as a plain link. `ActionResult.Header` is written onto the
response before the redirect, which is where the cookie goes.

`Secure: true` is right for production. Chrome and Firefox accept a secure cookie
on `http://localhost` as well, so `collage dev` works with it unchanged there;
Safari may not, and drops the cookie on plain `http`. If you develop in Safari, set
`Secure` from whether the request arrived over TLS, or preview in another browser.

### The action that ends one

```go
// Exit clears the preview cookie: GET /api/preview/exit
func Exit() *collage.Action {
	return collage.NewAction("preview-exit").
		WithPath("en", "/api/preview/exit").
		WithMethods(http.MethodGet).
		WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
			cookie := &http.Cookie{Name: CookieName, Path: "/", MaxAge: -1}
			result := collage.SeeOther("/")
			result.Header = http.Header{}
			result.Header.Add("Set-Cookie", cookie.String())
			return result, nil
		}).
		Build()
}
```

### The middleware that honours it

```go
type draftsKey struct{}

// Drafts reports whether this request is a preview, and data handlers should
// fetch unpublished content.
func Drafts(ctx context.Context) bool {
	drafts, _ := ctx.Value(draftsKey{}).(bool)
	return drafts
}

// Middleware turns a request carrying a valid preview cookie into a preview.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(CookieName); err == nil && Valid(cookie.Value) {
			if err := collage.SkipCache(r); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), draftsKey{}, true))
			}
		}
		next.ServeHTTP(w, r)
	})
}
```

The context value is set only when `SkipCache` succeeded. A draft rendered into a
request that is still cacheable is the one outcome this whole arrangement exists to
prevent, so the two go together or not at all.

### Wiring it up

```go
if err := app.Use(preview.Middleware); err != nil {
	return err
}
for _, action := range []*collage.Action{preview.Start(app), preview.Exit()} {
	if err := app.RegisterAction(action); err != nil {
		return fmt.Errorf("register action %q: %w", action.Name, err)
	}
}
```

### Data handlers that fetch drafts

A data handler asks for drafts when the request is a preview:

```go
func postData(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	slug := rc.Param("slug")
	tags := []string{"post:" + slug}

	post, err := collage.Cached(rc, "post:"+slug, time.Hour, tags,
		func(ctx context.Context) (cms.Post, error) {
			return cmsClient.Post(ctx, slug, cms.Options{Drafts: preview.Drafts(ctx)})
		})
	if errors.Is(err, cms.ErrNotFound) {
		return postView{}, tags, fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return postView{}, tags, err
	}
	return postView{Post: post, Preview: preview.Drafts(ctx)}, tags, nil
}
```

The `collage.Cached` call needs no special case. For an ordinary reader it returns
the stored published post; in a preview it fetches — with drafts — and stores
nothing, so the draft cannot leak into the value the next reader gets.

`Preview` in the view lets the template show a banner with a link to
`/api/preview/exit`. That is safe precisely because the page is never cached: the
banner cannot end up in front of a reader.

## Things to know

**A CDN can still answer first.** `SkipCache` controls collage's caches and marks
the preview response `no-store`, but a request the CDN answers from its own cache
never reaches your server. `Static()` pages are sent with `max-age=0,
must-revalidate`, so a CDN checks back every time; an `Incremental(ttl)` page may be
served from the CDN for up to its TTL. Configure the CDN to bypass its cache for
requests carrying the preview cookie.

**Previewing inside the CMS.** If the CMS shows the preview in an `<iframe>` on its
own domain, a `SameSite=Lax` cookie is not sent to the framed site. Use
`SameSite: http.SameSiteNoneMode` with `Secure: true` in that case — and know that
it may still not be enough. Inside that frame your cookie is a third-party cookie,
which Safari and Firefox block by default, so a preview in the CMS's iframe may
never receive it. Setting `Partitioned: true` as well (a CHIPS cookie, partitioned
by the CMS's site) gets it through in browsers that support partitioning; the
arrangement that works everywhere is the one above, opening the preview in a new
tab.

**A static export never sees a draft.** An export renders without a request, so no
middleware runs, `preview.Drafts` is false for every page, and only published
content is written. The preview actions are not exported either — an action needs a
server. A site deployed as static files can still offer previews by running the
server somewhere editors can reach it; see [Static export](/docs/static-export).

**Invalidation still matters.** A preview shows the draft; publishing it is what
changes the site. When the CMS publishes, its webhook should invalidate the post's
tag so readers get the new version. See
[Middleware and your own API](/docs/middleware-and-apis#invalidating-from-your-api).
