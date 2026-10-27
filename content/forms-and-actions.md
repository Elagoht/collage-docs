---
description: Handling form posts, fetch calls and webhooks with actions, and protecting them from request forgery.
reference: NewAction, ActionBuilder, ActionBuilder.WithGuard, ActionResult, SeeOther, SeeOtherPage, JSON, JSONBytes, Status, RenderPage, RenderFragment, PageBuilder.WithAction, PageBuilder.WithFragmentPath, ActionBuilder.WithStreamingBody, ActionBuilder.WithBodyTimeout, ErrCSRFHeaderRequired, ErrDuplicateRoute, ErrUnknownFragmentPath, FetchHeader, LocationHeader
---

# Forms and actions

A page answers `GET` and `HEAD`. Everything else — a form post, a `DELETE` from a
`fetch()`, a payment provider's webhook — is an **action**: a function that
receives the request and says what to answer with.

A request with a method nothing answers is a `405` with an `Allow` header, not the
page rendered as if nothing had happened, and error hooks see it as
`collage.ErrMethodNotAllowed`. `OPTIONS` is answered from the same list.

`OPTIONS`, `TRACE` and `CONNECT` belong to the server, not to an action. Since
v0.56.0, registering an action that declares one fails with
`collage.ErrInvalidActionMethod`, naming the action and the method, however the
action was built: `NewAction(...).WithMethods`, or a page's `WithAction` or
`WithActionFor`.

The router answers `OPTIONS` itself, and the forgery check counts it as a safe
method like `GET`, so an action claiming it would change state with no token
checked. A CORS preflight is a middleware's job, answered before the router.

A form post passes the same checks, in the same order, every time. Each one can
answer before your handler runs; once the handler has run, what it returns decides
the response:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 404" role="img" aria-labelledby="fsub-title">
<title id="fsub-title">A form post meets the page's guards, then the body limit, which answers 413, then the origin and token check, which answers 403. Then the action's handler runs. An invalid submission is answered with the page and 422. A valid one sent by a script with the Collage-Fetch header can be answered with a fragment and 200. A valid one answered with SeeOther is a 303 with a Location, or, when the request carries Collage-Fetch, a 204 with Collage-Location.</title>
<defs><marker id="fsub-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">form POST</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="52" width="170" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">page guards</text>
<line class="d-flow" x1="200" y1="65" x2="221" y2="65" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="54" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="68" text-anchor="middle">guard's answer</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="98" width="170" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">body limit</text>
<line class="d-flow" x1="200" y1="111" x2="221" y2="111" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="100" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="114" text-anchor="middle">413 · too large</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#fsub-arrow)" />
<rect class="d-frame" x="30" y="144" width="170" height="36" rx="6" /><text x="115" y="160" text-anchor="middle">origin, then token</text><text class="d-small" x="115" y="173" text-anchor="middle">unsafe methods only</text>
<line class="d-flow" x1="200" y1="162" x2="221" y2="162" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="225" y="151" width="93" height="22" rx="11" /><text class="d-small" x="271.5" y="165" text-anchor="middle">403 · refused</text>
<line class="d-flow" x1="115" y1="180" x2="115" y2="196" marker-end="url(#fsub-arrow)" />
<rect class="d-c" x="30" y="200" width="170" height="26" rx="6" /><text class="d-on-c" x="115" y="217" text-anchor="middle">your handler</text>
<path class="d-flow" d="M95 226C95 244 50 240 50 258" marker-end="url(#fsub-arrow)" />
<path class="d-flow" d="M115 226C115 244 150 240 150 258" marker-end="url(#fsub-arrow)" />
<path class="d-flow" d="M135 226C135 244 260 240 260 258" marker-end="url(#fsub-arrow)" />
<rect class="d-a" x="4" y="262" width="92" height="42" rx="5" /><text class="d-on-a" x="50" y="279" text-anchor="middle">invalid</text><text class="d-on-a d-small" x="50" y="295" text-anchor="middle">Page · 422</text>
<rect class="d-b" x="100" y="262" width="100" height="42" rx="5" /><text class="d-on-b" x="150" y="279" text-anchor="middle">valid, fetch</text><text class="d-on-b d-small" x="150" y="295" text-anchor="middle">RenderFragment</text>
<rect class="d-frame" x="204" y="262" width="112" height="42" rx="5" /><text x="260" y="279" text-anchor="middle">valid</text><text class="d-small" x="260" y="295" text-anchor="middle">SeeOther</text>
<line class="d-flow" x1="50" y1="304" x2="50" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="4" y="322" width="92" height="34" rx="6" /><text class="d-small" x="50" y="336" text-anchor="middle">422</text><text class="d-small" x="50" y="349" text-anchor="middle">form + errors</text>
<line class="d-flow" x1="150" y1="304" x2="150" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="100" y="322" width="100" height="34" rx="6" /><text class="d-small" x="150" y="336" text-anchor="middle">200</text><text class="d-small" x="150" y="349" text-anchor="middle">the part alone</text>
<line class="d-flow" x1="240" y1="304" x2="240" y2="318" marker-end="url(#fsub-arrow)" />
<rect class="d-chip" x="204" y="322" width="98" height="34" rx="6" /><text class="d-small" x="253" y="336" text-anchor="middle">no Collage-Fetch</text><text class="d-small" x="253" y="349" text-anchor="middle">303 · Location</text>
<line class="d-flow" x1="309" y1="304" x2="309" y2="358" marker-end="url(#fsub-arrow)" />
<rect class="d-chip-hit" x="204" y="362" width="112" height="34" rx="6" /><text class="d-hit-text d-small" x="260" y="376" text-anchor="middle">Collage-Fetch: 204</text><text class="d-hit-text d-small" x="260" y="389" text-anchor="middle">Collage-Location</text>
</svg>
<figcaption>The guards, the body limit and the forgery check each answer before the handler runs, in that order; the forgery check looks at where the request came from before it looks at the token. After the handler, a refused submission is the page again with 422. A valid one is a redirect, which collage turns into a 204 with Collage-Location for a request carrying Collage-Fetch, or, when the handler chooses to, the changed fragment alone.</figcaption>
</figure>

## An action on a page

The ordinary case is a form that posts to the page it sits on. Give the page an
action for that method:

```go
collage.NewPage("contact").
	WithLayouts(layout).
	WithContent(contactForm).
	WithPath("en", "/contact").
	WithAction("POST", sendMessage).
	Build()
```

```html
<form method="post" action="{{pageURL "contact"}}">
  {{csrfToken}}
  <input type="email" name="email">
  <textarea name="message"></textarea>
  <button type="submit">Send</button>
</form>
```

The action inherits the page's paths, in every locale the page declares, so the
form works on `/contact` and on `/tr/iletisim` alike.

An action handler has this shape:

```go
type ActionHandlerFunc func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error)
```

`rc.Request` is the request, with its body already bounded (see
[Request bodies](#request-bodies-are-bounded)); `rc.Locale` and `rc.Param` work as
they do in a data handler.

`WithAction` is a shorthand for one method. For several methods, a body limit of
its own, or no forgery check, build the action with `NewAction` and attach it
with `WithActionFor` — its own paths are replaced by the page's:

```go
collage.NewPage("post").
	WithContent(post).
	WithPath("en", "/posts/{slug}").
	WithActionFor(collage.NewAction("post-edit").
		WithMethods(http.MethodPut, http.MethodDelete).
		WithMaxBodyBytes(64 << 10).
		WithHandler(editPost).
		Build()).
	Build()
```

An action on a page is reached through the page's URL, so it is covered by the
page's [guards](/docs/pages-and-layouts#private-pages-guards): a form on a private
page cannot be submitted by a reader the page refuses. The guard runs before the
body is read and before the forgery check.

An action can carry a guard of its own, too: `WithGuard` takes the same
`GuardFunc` a layout does (since v0.70.0). On a page it is asked after the page's
guards — the page decides who sees the form, the action who may submit it:

```go
WithActionFor(collage.NewAction("delete").
	WithMethods(http.MethodPost).
	WithGuard(guards.Admin). // readers see the page; only admins delete
	WithHandler(deletePost).
	Build())
```

## An action at its own URL

An action that is not about a page — a JSON endpoint, a webhook — is registered on
its own:

```go
err := app.RegisterAction(collage.NewAction("like").
	WithPath("en", "/api/like").
	WithMethods(http.MethodPost).
	WithHandler(like).
	Build())
```

It has no page, so no page guards. To say who may call it, give it a guard of its
own with `WithGuard(guards.SignedIn)` (since v0.70.0): asked before the body is
read and before the forgery check, as a page's are. `collage add action
api/like --path /api/like --guard signedIn` writes it that way, and
[`collage routes`](/docs/cli#collage-routes) shows which actions have one.

A form posting to it names it rather than writing the path, so moving the action
cannot leave a form posting to the old one (since v0.36.0):

```html
<form method="post" action="{{actionURL "like"}}">{{csrfToken}}…</form>
```

`actionURL` takes parameters as `pageURL` does and follows the render's locale.
From Go it is `rc.ActionURL("like", params)` in the render's locale (since
v0.37.0), or `app.ActionURL("like", locale, params)` for another. An action on a page is
named as it was registered — `"contact:POST"` for the `WithAction` above — though
`pageURL` reaches the same URL. See
[`actionURL`](/docs/template-functions#actionurl).

An action may share a path with a page — that is exactly what `WithAction` does —
but not for `GET` or `HEAD`, which the page answers. Since v0.18.0 an action
answering either on the path of a page or a document is refused with
`collage.ErrDuplicateRoute`, in whichever order the two are registered; before, it
was matched first and hid the page without a word. Nor may two actions answer the
same method at the same path.

`RegisterAction` also refuses an action with no name, no path, no method or no
handler, and a second action under a name already taken — each with its own
error, listed in [Errors](/docs/errors#actions).

## What an action answers with

An `ActionResult` decides the response. Build one with a helper:

| Helper | Response |
| --- | --- |
| `collage.SeeOther(url)` | `303 See Other` to `url` |
| `collage.SeeOtherPage(rc, name, params)` | `303 See Other` to a page by its name, its path built by `rc.URL`; returns `(*ActionResult, error)` (since v0.68.0) |
| `collage.RenderPage(page)` | A whole page, rendered with this request's `RenderContext` |
| `collage.RenderFragment(f)` | One fragment's markup, without a layout |
| `collage.JSON(status, v)` | `v` marshalled as `application/json`; returns `(*ActionResult, error)` |
| `collage.JSONBytes(status, body)` | `body`, already encoded, as `application/json` |
| `collage.Status(status)` | A bare status and no body |

Before v0.60.0 `collage.JSON` took the encoded bytes and `JSONOf` took the value;
`JSON` takes the value now, `JSONBytes` the bytes, and `NoContent` is `Status`.

Or fill the struct yourself. It has four ways to produce a body — `Location`,
`Fragment`, `Page` and `Body` — and it sets one of them: since v0.60.0 a result
that sets two or more is a `500`, reported as `collage.ErrAmbiguousActionResult`,
rather than one silently winning. The other fields go with them:

- `Status` overrides the status each of them would choose.
- `Header` is written onto the response (the place for a `Set-Cookie`).
- `ContentType` goes with `Body`. An empty `ContentType` is
  `application/octet-stream`, never guessed from the bytes.

A `nil` result is a `204`.

```go
func like(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	n, err := store.Like(ctx, rc.Request.PostFormValue("post"))
	if err != nil {
		return nil, err
	}
	return collage.JSON(http.StatusOK, likeResponse{Count: n})
}
```

An error from the handler is a `500`, or a `404` if it wraps `collage.ErrNotFound`.

An action's response is sent `Cache-Control: no-store` unless you set the header
yourself, and it never enters the page cache: it was produced from one submission,
for whoever sent it.

## Failure renders, success redirects

Both `RenderPage` and `SeeOther` are good answers to a form. What decides between
them is what the browser does next.

Answering a POST with a page leaves the address bar on the posted URL and the
history entry a POST, so reloading submits the form again. Answering with a `303`
makes the browser's next request a `GET` of somewhere else, and reloading *that*
is harmless. So:

**A refused submission renders the page**, with status `422`. The reader will fix
it and send it again — resubmitting is the point — and rendering is how the reason
and what they typed get back in front of them.

**An accepted submission redirects**, with `303`. The work is done, and a reload
must not do it twice. To say "saved" on the page it redirects to,
[elagoht/flash](/docs/plugins#elagohtflash) carries a one-time message across the
redirect.

A form a script submits with `fetch` would follow the redirect too, downloading the
page it leads to, and then navigate there and have it rendered a second time.

Since v0.20.0 a request carrying the header `Collage-Fetch` (`collage.FetchHeader`)
is answered `204 No Content` with the destination in `Collage-Location`
(`collage.LocationHeader`) instead of a redirect, and the script navigates once.
The action returns `SeeOther` as before. collage-live's forms send the header.

Since v0.71.0 only a path or an `http`/`https` address is handed over that way. A
script navigates with `location.assign`, which runs a `javascript:` URL in the page,
where a browser following a `Location` refuses to; a `javascript:`, `data:` or other
destination is answered as the `303` itself, which `fetch` will not follow.

The header changes only a redirect. A page or a fragment is answered the same way
with it or without it; a handler that wants to answer a script with less than the
page reads the header itself, as [the task list below](#example-a-team-task-list)
does.

### The validation re-render

The handler and the page it answers with share one `RenderContext`. Whatever the
handler puts there with `rc.Set`, the page's data handlers can read with
`collage.Get`. No session, no flash message, nothing in the query string: they are
the same request.

```go
type contactView struct {
	Error string
	Email string
}

func ContactPage() *collage.Page {
	form := collage.NewFragment("contact-form", "pages/contact.html").
		WithData(collage.Load(contactData)).
		Build()

	return collage.NewPage("contact").
		WithLayouts(layouts.Master()).
		WithContent(form).
		WithPath("en", "/contact").
		WithAction("POST", sendContact).
		Build()
}

var contactFormKey = collage.NewKey[contactView]("contact:form")

func sendContact(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	email := strings.TrimSpace(rc.Request.PostFormValue("email"))
	if !strings.Contains(email, "@") {
		contactFormKey.Set(rc, contactView{Error: "That does not look like an email address.", Email: email})
		// rc.Page is the page this action answers on: the contact page.
		return &collage.ActionResult{Status: http.StatusUnprocessableEntity, Page: rc.Page}, nil
	}
	if err := messages.Send(ctx, email, rc.Request.PostFormValue("message")); err != nil {
		return nil, err
	}
	return collage.SeeOther("/contact/thanks"), nil
}

func contactData(_ context.Context, rc *collage.RenderContext) (contactView, error) {
	// Set by the action when this render is its answer; empty on an ordinary GET.
	view, _ := contactFormKey.Get(rc)
	return view, nil
}
```

```html
<form method="post" action="{{pageURL "contact"}}">
  {{csrfToken}}
  {{with .Error}}<p class="error">{{.}}</p>{{end}}
  <input type="email" name="email" value="{{.Email}}">
  <textarea name="message"></textarea>
  <button type="submit">Send</button>
</form>
```

The page is rendered as any page is, render hooks included: a plugin's
`OnAfterRender` — a minifier, an image rewriter — shapes the validation page just
as it shapes the page a `GET` gets. See [Writing a plugin](/docs/writing-plugins#afterrenderhook).

The page you answer with must be **the value you registered** with
`app.RegisterPage`. Registration is what puts a page's content into its layout, so
a page built inside the handler would render as a layout around nothing; collage
refuses it with `collage.ErrUnregisteredPage`, naming the page.

In an action on a page's URL, `rc.Page` is that page, the registered value (since
v0.33.0). So the action above needs nothing handed to it, and there is no
`var page` for a closure to capture — the shape this took before.

An action registered at a URL of its own has no page there and a `nil` `rc.Page`;
it answers with the registered page it wants, built before the action. In dev mode a
form post answered `422` with no body is logged as a warning, because the reader
gets a blank page and a `nil` `rc.Page` is the usual cause.

To redirect to a page by name rather than by path, so a renamed path cannot leave
the redirect pointing at the old one, use `collage.SeeOtherPage` (since v0.68.0).
It builds the path with [`rc.URL`](/docs/links-and-locales#links-from-go), in the
request's locale and with each value kept inside its path segment, and an unknown
name is its error; the action needs no `*App` handed to it for that either:

```go
return collage.SeeOtherPage(rc, "post", map[string]string{"slug": slug})
```

Before v0.68.0, build the path with `rc.URL` (since v0.37.0) and hand it to
`SeeOther`.

A login that sends the reader back to where they came from reads that place from
a `next` query parameter, and the parameter is the reader's input. Pass it through
`collage.SafeRedirect` (since v0.44.0) before it becomes a `Location`:

```go
return &collage.ActionResult{
	Location: collage.SafeRedirect(rc.Request.URL.Query().Get("next"), "/"),
}, nil
```

It returns `next` when that is a path on this site, and the fallback otherwise, so
`next=https://evil.example` ends at `/`. See
[Redirects stay on the site](/docs/security#redirects-stay-on-the-site).

Three plugins cover what a form usually needs around this:

- [elagoht/validate](/docs/plugins#elagohtvalidate) checks the fields and answers
  a refused submission with the page and a 422, the reader's input and each
  field's message in it.
- [elagoht/honeypot](/docs/plugins#elagohthoneypot) refuses spam without a
  CAPTCHA.
- [elagoht/session](/docs/plugins#elagohtsession) keeps who is signed in, in a
  signed cookie.

## Forgery protection

Every request to an action with an unsafe method (anything but `GET`, `HEAD` and
`OPTIONS`) is checked for a forgery token before the handler runs. A
request without a valid one is a `403`, and the handler never sees it. Error hooks
see why:

| Error | Why |
| --- | --- |
| `collage.ErrCSRFMissing` | No token, or no cookie |
| `collage.ErrCSRFMismatch` | A token that is not the cookie's |
| `collage.ErrCSRFInvalid` | One this application did not sign |
| `collage.ErrCSRFCrossOrigin` | A submission the browser says came from another origin (since v0.34.0) |

### In a form

Put `{{csrfToken}}` inside the `<form>`. It renders the whole hidden input, field
name included, so there is nothing to get wrong:

```html
<form method="post">
  {{csrfToken}}
  ...
</form>
```

The scheme is a signed double-submit cookie: the token is a random value and its
signature under `Security.CSRFKey`, sent both in a cookie and in the form, and
checking it needs the key and nothing else. No session store, nothing shared
between instances.

### Where the request came from

A signed token is not a secret from an attacker: any page with a form hands one to
whoever loads it. Someone who can write a cookie for your domain — a sibling
subdomain, a man-in-the-middle on plain HTTP — could plant their own token in a
visitor's browser and submit it. So since v0.34.0 the token is not checked alone.

The browser says where a request came from in `Sec-Fetch-Site`, or in `Origin`
(compared with `Host`) when it is an older one, and a request it marks as coming
from another origin is refused with a `403` whatever it carries. A sibling
subdomain counts as another origin: it is exactly who can plant the cookie. A
request with neither header is not from a browser, and goes on to the token check.

A form on another origin that is meant to post here — an admin panel on its own
subdomain — is named in `Security.CSRFTrustedOrigins`:

```go
Security: collage.SecurityConfig{
	CSRFKey:            []byte(os.Getenv("COLLAGE_CSRF_KEY")),
	CSRFTrustedOrigins: []string{"https://admin.example.com"},
},
```

Each entry is `scheme://host[:port]`. Case and a default port do not matter; a
wildcard is refused when the application starts — name each origin. The token
cookie is `SameSite=Lax`, so a trusted origin on another *site* never sends it:
trusting one is for another origin of the same site.

### Pages with forms are still cached

A token belongs to one reader, and a cached page is shared by everyone. So
`{{csrfToken}}` does not render a token: it renders a marker, and the marker is
what the cache stores. Each response has the marker replaced with that reader's
own token on the way out, and is sent `private, no-store` along with the token's
cookie. The render is shared; the one per-reader string is not.

Without this, a newsletter form in a site's footer would turn caching off for the
whole site. The one thing it does change: a page with a form has no server to
submit to once it is a static file, so [static export](/docs/static-export) skips
such a page and says why.

### Set the key

```go
Security: collage.SecurityConfig{
	CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")), // e.g. openssl rand -hex 32
},
```

Leave it empty and one is generated at startup. collage says so — a warning in
production, a plain log line in development — once the application has an action
that accepts an unsafe method and verifies tokens; an application without one, or
whose only such actions are exempted with `WithoutCSRF`, never checks a token and
is not told about the key.

That is fine for a first run and wrong to deploy: a generated key is different in
every process, so a form loaded before a restart is refused after it, and one
served by one instance is refused by the next. A cached page with a form,
stored under an earlier key, is rendered again rather than served with a token
nothing can verify — see [the disk cache's namespace](/docs/caching#the-namespace).

### From fetch

A `fetch()` with no form sends the token in the `X-CSRF-Token` header. Read it from
any `{{csrfToken}}` input on the page — it holds the reader's own token by the time
the page arrives. (`_csrf` is the field's default name; see
[below](#requests-that-cannot-carry-a-token) for changing it.)

```js
const token = document.querySelector('input[name="_csrf"]')?.value ?? "";
await fetch("/api/like", {
  method: "POST",
  headers: { "X-CSRF-Token": token },
  body: new URLSearchParams({ post: "hello-world" }),
});
```

A `fetch()` that posts a form needs nothing extra. `new FormData(form)` is sent as
`multipart/form-data`, which collage reads like a URL-encoded body, so the hidden
input travels with the rest of the fields:

```js
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const response = await fetch(form.action, { method: "POST", body: new FormData(form) });
  // The action answered with RenderFragment: the form's own fragment, re-rendered.
  form.outerHTML = await response.text();
});
```

A fragment or page an action answers with is rendered like any other, so a form in
it carries a fresh token too: a form that replaces itself keeps working.

An action built with `WithStreamingBody()` is the exception: it reads the token
from the header alone, so even a `fetch()` posting a form sets `X-CSRF-Token`. See
[Streaming bodies](#streaming-bodies).

### Requests that cannot carry a token

A payment provider's webhook, or an API called with a bearer token by something
that is not a browser, cannot carry a token. Turn the check off for that action
alone:

```go
app.RegisterAction(collage.NewAction("stripe-webhook").
	WithPath("en", "/hooks/stripe").
	WithMethods(http.MethodPost).
	WithoutCSRF().
	WithHandler(onStripe).
	Build())
```

Then authenticate it some other way — a webhook's signature header, a bearer
token. On anything a browser submits, `WithoutCSRF()` gives the protection away.
(`Security.DisableCSRF` turns it off for the whole application, which is only right
for one with no browser forms at all.)

The cookie, field and header names can be changed with `Security.CSRFCookieName`,
`CSRFFieldName` and `CSRFHeaderName`. A renamed field is renamed on both sides:
`{{csrfToken}}` renders the name the check reads, so the form needs no change.

## Request bodies are bounded

An action's request body is limited to **4 MiB** by default. Change it for the
application with `Server.MaxBodyBytes`, or for one action with
`WithMaxBodyBytes`; a negative value means no limit.

The limit is applied before your handler runs, not by it — and since v0.34.0
before your [middleware](/docs/middleware-and-apis) runs too, at the limit of the
action the request routes to.

- A middleware that reads or parses the body reads it bounded.
- An action whose body already ran into its limit answers `413` without running.
- A handler mounted with `app.Handle` is left alone.

A limit every handler has to remember is a limit the one that forgot does not
have — and that is the one an anonymous caller will find. A handler that reads past it gets an
`*http.MaxBytesError`, and returning that error answers `413`:

```go
if err := rc.Request.ParseForm(); err != nil {
	return nil, err // a body over the limit becomes a 413
}
```

Since v0.57.0 the answer is `413` whatever error the handler returns once its body
has run into the limit: the read's own, or one of its own such as "upload failed",
which used to be a `500`. The handler's error is kept beside the `413`, in the log
and in what error hooks are handed, and `errors.As` still finds the
`*http.MaxBytesError` on it. It is logged as a warning, not an error.

When the token is checked from the form, the check reads the body first, so an
oversized form is refused there, before your handler runs — still with a `413`,
not a `403`: a body too large to read is not a forgery.

A multipart body larger than the 32 MiB the parser keeps in memory spills its files
to disk. Since v0.34.0 they are removed once the action has answered, whether the
handler ran or the request was refused. Since v0.71.0 that holds whoever parsed the
form — a page's guard, an action's guard, a data handler, a middleware — and an
action's guards meet the body already bounded.

Since v0.71.0 a body whose `Content-Length` is over the limit is refused with `413`
before the forgery check and the handler run, so an action without a forgery check is
never run on a form cut short. A chunked body, which does not say its length, that
runs into the limit is refused after the handler instead of answered as a success.

A size rule of your own, like a photo over 5 MB refused with a message on the form,
only works when the action's limit is above it. Otherwise collage answers the `413`
first, and your message is never shown. A 4.5 MB photo under the default limit is
then refused as too large for the server, even though the form says 5 MB is
allowed. Give such an action room for the file and the rest of the form:
`WithMaxBodyBytes(maxPhoto + 64<<10)`.

## Streaming bodies

An action that takes a large upload reads its body as a stream, and nothing may read
it first. Since v0.57.0 `WithStreamingBody()` makes that a guarantee:

```go
app.RegisterAction(collage.NewAction("upload").
	WithPath("en", "/upload").
	WithMethods(http.MethodPost).
	WithStreamingBody().
	WithMaxBodyBytes(2 << 30).
	WithHandler(upload).
	Build())
```

- **The body arrives unread.** Nothing before the handler parses a form out of it:
  `rc.Request.Body` is the bytes as they were sent, and `rc.Request.Form` and
  `rc.Request.MultipartForm` are nil. The handler reads it — with
  `rc.Request.MultipartReader()`, say, one part at a time. The
  [elagoht/uploads](/docs/plugins#elagohtuploads) plugin does this for you.
- **The token comes in the header, and only there.** Finding it in a form field
  would mean parsing the body, so the field is not looked for. A request without
  `X-CSRF-Token` (or `Security.CSRFHeaderName`) is refused with `403`, and error
  hooks are told why with `collage.ErrCSRFHeaderRequired`, which names the header.
  The reader sees only the 403. The cookie is `HttpOnly`, so read the value from the
  input `{{csrfToken}}` renders (named `_csrf`, or `Security.CSRFFieldName`):

  ```js
  const token = form.querySelector('input[name="_csrf"]').value;
  await fetch(form.action, {
    method: "POST",
    headers: {"X-CSRF-Token": token},
    body: new FormData(form),
  });
  ```

- **A plain HTML form cannot post to it.** A browser submitting a form has no way
  to set a header, so its token is in a field nobody reads, and it is refused.
  Submit with `fetch()` as above. `WithoutCSRF()` still turns the check off
  entirely, and the body still arrives unread.
- **`MaxBodyBytes` still applies, and must be raised.** The 4 MiB default bounds a
  streaming body as it bounds any other, enforced as the handler reads. A handler
  that reads past it and fails is answered with `413`, whatever error it returns.
- **Anything that parses the form gives up the stream.** `rc.Request.FormValue`,
  `ParseMultipartForm`, a helper such as `validate.Form` — each reads the body to
  its end before the handler does. Nothing fails: net/http writes the whole body,
  up to `MaxBodyBytes`, to temp files under `os.TempDir()` (on many Linux hosts a
  `tmpfs`, and so memory), and the upload is simply no longer streamed. A handler
  that reads the body itself afterwards finds it consumed. Validate the fields the
  handler read from the stream instead. A plugin's `BeforeActionEvent.Form()`
  returns `collage.ErrStreamingBody` for such an action without reading a byte —
  see [Writing a plugin](/docs/writing-plugins#beforeactionhook).
- **An upload cut off is the client's failure.** When a read of the body fails
  because the client went away or the connection broke, and the handler then
  returns an error, the action answers `400`; when the read deadline passed, `408`.
  It is logged at debug and no error hook is called: a cancelled upload is the most
  ordinary way an upload ends. As a rule nobody receives that answer — a client
  that left is gone, and under `WithBodyTimeout` the write deadline passed with the
  read one — so the status is for metrics and the access log. This holds for every
  action, streaming or not.

Registration refuses `WithStreamingBody()` on an action answering none of `POST`,
`PUT` or `PATCH`, with `collage.ErrStreamingBodyMethod`.

### Long uploads and timeouts

The server's deadlines are sized for pages. `ReadTimeout` (15s by default) bounds
reading the whole request, body included, and `WriteTimeout` (30s) starts as soon
as the request's headers arrive, so it bounds the upload too, not only the
response. An upload slower than that fails. One that outlives `WriteTimeout` but
not a raised `ReadTimeout` is worse: the handler stores the file, and the client
gets no answer and tries again.

`WithBodyTimeout(d)` (since v0.57.0) gives one action its own deadline. For a
request to that action it replaces both deadlines with now plus `d`, so the rest of
the site keeps its short ones. It does so after the page's guards and before
anything reads the body. On an ordinary action that is before the forgery check,
which reads the body; on a streaming action, whose check reads only a header, it is
after it, so a request without a token is refused under the server's deadlines:

```go
app.RegisterAction(collage.NewAction("upload").
	WithPath("en", "/upload").
	WithMethods(http.MethodPost).
	WithStreamingBody().
	WithMaxBodyBytes(2 << 30).
	WithBodyTimeout(30 * time.Minute).
	WithHandler(upload).
	Build())
```

Choose `d` to cover the whole upload and the answer after it: at least
`MaxBodyBytes` divided by the slowest connection you mean to serve. Zero, the
default, keeps the server's deadlines; a negative value is refused at registration
with `collage.ErrNegativeBodyTimeout`. The deadlines are set through
`http.ResponseController`, which needs every `ResponseWriter` wrapper between the
server and the action to implement `Unwrap`. collage's own and the `elagoht`
plugins' do; when one does not, the request runs under the server's deadlines and a
warning is logged once.

A request that clears those checks holds its connection for up to `d`, and a token
is free to anyone who loads the form. Guard the upload with a page guard or an auth
middleware, which run before all of this; a check inside the handler runs after `d`
has been granted. What a proxy in front needs is in
[Deployment](/docs/deployment#uploads).

## Invalidating what an action changed

An action that changes content usually makes some cached pages wrong. Say which,
on the result:

```go
func publish(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	slug := rc.Param("slug")
	if err := store.Publish(ctx, slug); err != nil {
		return nil, err
	}
	result := collage.SeeOther("/posts/" + slug)
	result.InvalidateTags = []string{"post:" + slug, "blog:posts"}
	return result, nil
}
```

The tags are invalidated **before** the response is written. That order is the
point: the reader follows the redirect to the page they just changed, and it must
not be served from a cache entry the change made stale. A failure to invalidate is
logged rather than failing the response — the change has already happened.

Tags work exactly as in [Caching](/docs/caching#dependency-tags), and reach values
stored with `collage.Cached` too.

The pages are marked stale, and so are the `collage.Cached` values they were
built from: both are kept (pages since v0.63.0). Since v0.64.0 the invalidated
pages are then [warmed](/docs/caching#warming-after-an-invalidation), and the
redirect's own page first.

When `Location` is on the same origin, its variant (its query varied as the cache
key varies it, and every `Vary` header variant of it) is rendered again with fresh
data before the redirect is written, so the redirect lands on the new page, a
fresh `HIT`. The variant the submitter lands on, the one whose `Vary` values their
request declared, goes first, and is warmed even if nobody read it since its last
warm.

Only this action's response waits for that, and for no longer than the shortest
of:

- 5 seconds;
- the request's deadline;
- what is left of the response's write deadline, less half a second to write the
  redirect in. The write deadline is the action's own `WithBodyTimeout` when it
  sets one, and otherwise `Server.WriteTimeout` from when the request arrived.

With no time left the response does not wait at all. The keys not warmed by then
stay in the background run.

Only `ActionResult.InvalidateTags` does this. An action that calls
`app.InvalidateTags` (or `PurgeTags`) itself, inside its handler, has those pages
warmed in the background like any other invalidation's, but its redirect is
written without waiting for its page, so the submitter may land on it `STALE` or
pay for its render.

The page is served stale (`X-Collage-Cache: STALE`, `Cache-Control: no-store`)
only when the render cannot finish in time: the upstream is slow or down, the
budget is spent, or the render failed. The redirect still goes out then, and the
fetch that replaces a `Cached` value runs in the background. That is the price of
a site that keeps answering while its upstream is down.

When an action must show
its own change at once whatever happens, call `app.PurgeTags` from the handler for
those tags instead. Nothing is then kept, so a failed fetch fails the page.

Redirect to a clean URL, or to one whose query holds only parameters the page
names with `WithCacheParams`. A page that declares them drops any other query
parameter from its key, so `/posts?saved=1` is warmed as `/posts`. A page that
declares none keys on the raw query: `/posts?saved=1` is a key of its own that no
recipe holds, and nothing is warmed for it.

## A fragment at its own URL

Refreshing part of a page without a client framework takes two things: a URL that
answers with only that part, and a few lines of JavaScript to swap it in.
`WithFragmentPath` is the first:

```go
collage.NewPage("search").
	WithLayouts(layout).
	WithContent(searchContent).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Dynamic().
	Build()
```

`GET /search/results?q=grid` renders the `results` fragment and nothing else. Its
data handler runs, its own slots are filled and its failure policy applies — it
is the same render, started lower down. There is no layout around it, and its
response is never cached by the framework.

Missing and broken stay apart here, as on the page. A required fragment whose data
handler wraps `collage.ErrNotFound` answers `404`, and any other failure `500`. A
fragment path is answered like an action, so its failure is plain text, never
cached, with nothing of the page in it, which is what a script fetching it can
read. Before v0.39.2 a fragment path answered `500` for both, so a script could not
tell "not yours to see" from "broken".

```js
const input = document.querySelector('input[name="q"]');
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Nothing is reachable unless it is declared. A framework that exposed every
fragment automatically would put every internal part of every page on the public
web.

A fragment path is a part of the page that opened it, and since v0.34.0 it meets
the page's [guards](/docs/pages-and-layouts#private-pages-guards) first —
outermost layout inwards, then the content fragment's — and then the fragment's
own. A private page's fragment path is private too.

Before v0.34.0 it ran only the
fragment's own guard, so a guarded layout whose content was opened at a URL of its
own served that content to anyone. A fragment may still carry a guard of its own,
for a part stricter than the page:

```go
results := collage.NewFragment("results", "fragments/results.html").
	WithData(collage.LoadTagged(search)).
	WithGuard(requireAdmin). // after the page's own guards
	Build()
```

A fragment path is claimed like any other route. One spelled like a page's path, a
document's, another page's fragment path, or a redirect's source is refused at
registration, in whichever order the two arrive: it would otherwise hide the page
without a word.

**Each fragment path is a render of its own.** Inside a page, fragments that need
the same slow value share it with `collage.Once`: one fetch per render. A page
refreshed part by part is several renders, and `Once` shares nothing between them.
For fragments that read the same data, use `collage.Cached`, which keeps the value
across renders for as long as its TTL — or until one of its tags is invalidated:

```go
var statsKey = collage.NewKey[monitor.Stats]("system:stats")

stats, err := collage.Cached(rc, statsKey, time.Second, nil,
	func(ctx context.Context) (monitor.Stats, error) { return monitor.Collect(ctx) })
```

A measurement like this one takes no tags: its TTL keeps it fresh, and each
fragment reading it returns a tag of its own. Tagged, every fragment would depend on
every tag, and invalidating one part would re-render them all.

Combined with an action that answers `RenderFragment`, a form can post and be
answered with only the part that changed:

```go
WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	if err := store.AddComment(ctx, rc.Param("slug"), rc.Request.PostFormValue("body")); err != nil {
		return nil, err
	}
	result := collage.RenderFragment(comments)
	result.InvalidateTags = []string{"comments:" + rc.Param("slug")}
	return result, nil
})
```

The fragment's data handler runs with the action's `RenderContext`, so it sees the
comment just added — and anything the handler put there with `rc.Set`.

### Linking to one

Write the path once, in `WithFragmentPath`, and build every link to it by name, as
`pageURL` does for pages (since v0.18.0):

```html
<div data-live="{{fragmentURL "search" "results"}}">{{slot "results"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

`fragmentURL` uses the render's locale and falls back to the default one;
`fragmentURLIn "tr" "search" "results"` names the locale, and
`app.FragmentURL(page, fragment, locale, params)` is the same from Go. A fragment
the page did not open is `collage.ErrUnknownFragmentPath`; the rest is in
[Links and locales](/docs/links-and-locales#a-fragments-url).

Links inside the fragment lead the other way, to the page. Since v0.77.0 a fragment
answered at its own path renders with the page that opened it as `rc.Page`, so
`localeURL` in it is the page in another locale, and
[`withQuery`](/docs/template-functions#withquery) is the page's URL with the
request's query changed. The search results' pager can link the next page of
results while the script fetched only the results:

```html
<a href="{{withQuery "page" (add .Page 1)}}">More results</a>
```

On `/search/results?q=grid&page=1` that link is `/search?page=2&q=grid`. Before
v0.77.0 the fragment had no page there, and `localeURL` failed.

### What it hoists

A fragment answered on its own has no layout around it. A `{{hoist}}` marker the
fragment writes itself is filled as in a page; what it hoisted into any other area
— a stylesheet asked for with `{{stylesheet}}`, a title — comes ahead of the
markup (since v0.18.0), one inert `<template>` per item:

```html
<template data-collage-hoist="head" data-collage-key="stylesheet:/static/chart.css"><link rel="stylesheet" href="/static/chart.css"></template>
<section>…the fragment…</section>
```

The key is the one the page's own head deduplicated by, so a script can add to
`document.head` what it does not already have. A client that ignores the channel
inserts template elements, which render nothing and run nothing.

### Revalidation

Since v0.18.0 the answer to a GET carries an `ETag`, the hash of the body as sent,
and `Cache-Control: private, no-cache`. A request with a matching `If-None-Match` is
answered `304` with no body. The render still runs; what is saved is the body on
the wire and the client's work, which for a panel refreshed every few seconds is
most of it. A handler that sets `Cache-Control` itself keeps its own, unless a
`PersonaliseHook` made the body personal (since v0.43.0): then it is
`private, no-cache` whatever the handler set.

A fragment holding a form earns a `304` too since v0.78.0: its ETag is hashed
before the forgery token is masked, so it stays the same from one answer to the
next and equals the one `RenderFragment` reports for a pushed copy.

### Refreshing it from the browser

The framework ships no client script. [collage-live](/docs/plugins#elagohtlive) is
a plugin that does: it refreshes elements marked with `data-collage-fragment` on an
interval or when the server pushes a change over an event stream, and applies the
hoist channel and the ETag above. The protocol it speaks is this section, so htmx
or a script of your own works against the same server.

A form it submits with `fetch` follows the [rule above](#failure-renders-success-redirects),
in parts: the answer goes into the form's target on a success, or on a `422`. So an
action says a submission did not validate by answering the form's fragment again,
with the errors, and status 422 (since collage-live v0.2.0):

```go
var problemsKey = collage.NewKey[[]string]("problems")

if problems := validate(rc); len(problems) > 0 {
	problemsKey.Set(rc, problems)
	result := collage.RenderFragment(formFragment)
	result.Status = http.StatusUnprocessableEntity
	return result, nil
}
```

Any other failure — a 500, a refused forgery token — leaves the target as it was
and marks it stale, rather than filling it with an error page. A redirect is
followed as the browser would follow it, in one request (since collage-live
v0.2.1): the form sends [`Collage-Fetch`](#failure-renders-success-redirects), and
collage answers with where the action redirects instead of the redirect, which
`fetch` would otherwise follow and download before the navigation fetched the page
again.

## Example: a team task list

A multi-user web app keeps one task list per team at `/teams/{team}/tasks`. Under
the list sits an "add a task" form. It has to work three ways:

| The submission | The answer |
| --- | --- |
| A title that is empty or too long | `422` and the form again, the error and what was typed in it |
| A valid task, from a plain form post | `303` back to the list, so a reload adds nothing twice |
| A valid task, from a script sending `Collage-Fetch` | `200` and the list fragment alone, swapped in place |

The list is one fragment, used twice: in the page's slot, and at its own URL with
`WithFragmentPath`. Its data handler tags what it read with the team, so adding a
task invalidates that tag and every member of the team sees the new task on their
next load.

```go
// Tag is the dependency tag of one team's task list.
func Tag(team string) string { return "tasks:" + team }

type newTask struct {
	Team  string
	Error string
	Title string
}

var newTaskKey = collage.NewKey[newTask]("tasks:new")

func TasksPage() *collage.Page {
	list := collage.NewFragment("task-list", "fragments/task-list.html").
		WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) ([]store.Task, []string, error) {
			team := rc.Param("team")
			tasks, err := store.Tasks(ctx, team)
			return tasks, []string{Tag(team)}, err
		})).
		Build()
	form := collage.NewFragment("task-form", "fragments/task-form.html").
		WithData(collage.Load(func(_ context.Context, rc *collage.RenderContext) (newTask, error) {
			// Set by the action when this render is its answer; empty on a GET.
			view, _ := newTaskKey.Get(rc)
			view.Team = rc.Param("team")
			return view, nil
		})).
		Build()
	content := collage.NewFragment("tasks", "pages/tasks.html").
		WithSlotFragment("list", list).
		WithSlotFragment("form", form).
		Build()

	return collage.NewPage("tasks").
		WithLayouts(layouts.Master()).
		WithContent(content).
		WithPath("en", "/teams/{team}/tasks").
		WithFragmentPath("en", "/teams/{team}/tasks/list", list).
		Incremental(time.Hour).
		WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			team := rc.Param("team")
			fetched := rc.Request.Header.Get(collage.FetchHeader) != ""

			title := strings.TrimSpace(rc.Request.PostFormValue("title"))
			if title == "" || utf8.RuneCountInString(title) > 200 {
				newTaskKey.Set(rc, newTask{Error: "A title is 1 to 200 characters.", Title: title})
				if fetched {
					// A script swaps the form alone.
					result := collage.RenderFragment(form)
					result.Status = http.StatusUnprocessableEntity
					return result, nil
				}
				// A browser gets the whole page back, its input in the form.
				return &collage.ActionResult{Status: http.StatusUnprocessableEntity, Page: rc.Page}, nil
			}

			if err := store.AddTask(ctx, team, title); err != nil {
				return nil, err
			}

			result := collage.RenderFragment(list)
			if !fetched {
				var err error
				result, err = collage.SeeOtherPage(rc, "tasks", map[string]string{"team": team})
				if err != nil {
					return nil, err
				}
			}
			// The team's cached page renders again, for every member.
			result.InvalidateTags = []string{Tag(team)}
			return result, nil
		}).
		Build()
}
```

The form fragment is the whole `<form>`, so a script can replace it with what a
`422` sends back, and the list fragment is the whole `<ul>`:

```html
<form id="task-form" method="post" action="{{pageURL "tasks" "team" .Team}}">
  {{csrfToken}}
  {{with .Error}}<p class="error">{{.}}</p>{{end}}
  <input name="title" value="{{.Title}}">
  <button type="submit">Add</button>
</form>
```

```html
<ul id="task-list">
  {{range .}}<li>{{.Title}}</li>{{else}}<li>No tasks yet.</li>{{end}}
</ul>
```

A few lines of script post the form with `Collage-Fetch`. The listener is on the
document, so it keeps working after a `422` has replaced the form:

```js
document.addEventListener("submit", async (event) => {
  const form = event.target;
  if (form.id !== "task-form") return;
  event.preventDefault();
  const response = await fetch(form.action, {
    method: "POST",
    headers: { "Collage-Fetch": "1" },
    body: new FormData(form),
  });
  if (response.status === 200) {
    document.querySelector("#task-list").outerHTML = await response.text();
    form.reset();
  } else if (response.status === 422) {
    form.outerHTML = await response.text();
  }
});
```

What the example leans on, each covered above:

- **The forgery token travels with the form.** `new FormData(form)` carries the
  hidden input `{{csrfToken}}` rendered, and a fragment answered after a `422`
  renders a fresh one.
- **The page is cached, the form still works.** `Incremental(time.Hour)` caches
  the page with the token marker in it; each reader gets their own token.
- **The invalidation comes first.** The tag is invalidated before the response is
  written, so the redirect lands on the new list.
- **Who may post is the page's question.** A [guard](/docs/pages-and-layouts#private-pages-guards)
  that lets in only the team's members covers the page, its action and
  `/teams/{team}/tasks/list` alike.

With the script off, the same form posts, redirects and reloads the list. The
tutorial [Update part of a page](/docs/update-part-of-a-page) builds a project of
this shape step by step, with a list that also refreshes itself.
