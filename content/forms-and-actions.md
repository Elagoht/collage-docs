---
description: Handling form posts, fetch calls and webhooks with actions, and protecting them from request forgery.
reference: NewAction, ActionBuilder, ActionResult, SeeOther, JSONOf, RenderPage, RenderFragment, PageBuilder.WithAction, PageBuilder.WithFragmentPath, ErrDuplicateRoute, ErrUnknownFragmentPath, FetchHeader, LocationHeader
---

# Forms and actions

A page answers `GET` and `HEAD`. Everything else — a form post, a `DELETE` from a
`fetch()`, a payment provider's webhook — is an **action**: a function that
receives the request and says what to answer with.

A request with a method nothing answers is a `405` with an `Allow` header, not the
page rendered as if nothing had happened, and error hooks see it as
`collage.ErrMethodNotAllowed`. `OPTIONS` is answered from the same list.

## An action on a page

The ordinary case is a form that posts to the page it sits on. Give the page an
action for that method:

```go
collage.NewPage("contact").
	WithLayout(layout).
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

An action may share a path with a page — that is exactly what `WithAction` does —
but not for `GET` or `HEAD`, which the page answers. Since v0.18.0 an action
answering either on the path of a page or a document is refused with
`collage.ErrDuplicateRoute`, in whichever order the two are registered; before, it
was matched first and hid the page without a word. Nor may two actions answer the
same method at the same path. An action with no
name, no path, no method or no handler is refused by `RegisterAction`, and so is a
second action under a name already taken — each with its own error, listed in
[Errors](/docs/errors#actions).

## What an action answers with

An `ActionResult` decides the response. Build one with a helper:

| Helper | Response |
| --- | --- |
| `collage.SeeOther(url)` | `303 See Other` to `url` |
| `collage.RenderPage(page)` | A whole page, rendered with this request's `RenderContext` |
| `collage.RenderFragment(f)` | One fragment's markup, without a layout |
| `collage.JSON(status, body)` | `body` as `application/json` |
| `collage.JSONOf(status, v)` | `v` marshalled as `application/json`; returns `(*ActionResult, error)` |
| `collage.NoContent(status)` | A bare status and no body |

Or fill the struct yourself. It has four ways to produce a body — `Location`,
`Fragment`, `Page` and `Body` — and the first one set wins, in that order. `Status`
overrides the status each of them would choose, `Header` is written onto the
response (the place for a `Set-Cookie`), and `ContentType` goes with `Body`. An
empty `ContentType` is `application/octet-stream`, never guessed from the bytes. A
`nil` result is a `204`.

```go
func like(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
	n, err := store.Like(ctx, rc.Request.PostFormValue("post"))
	if err != nil {
		return nil, err
	}
	return collage.JSONOf(http.StatusOK, likeResponse{Count: n})
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
must not do it twice.

A form a script submits with `fetch` would follow the redirect too, downloading the
page it leads to, and then navigate there and have it rendered a second time. Since
v0.20.0 a request carrying the header `Collage-Fetch` (`collage.FetchHeader`) is
answered `204 No Content` with the destination in `Collage-Location`
(`collage.LocationHeader`) instead of a redirect, and the script navigates once.
The action returns `SeeOther` as before. collage-live's forms send the header.

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
		WithDataHandler(contactData).
		Build()

	var page *collage.Page
	page = collage.NewPage("contact").
		WithLayout(layouts.Layout()).
		WithContent(form).
		WithPath("en", "/contact").
		WithAction("POST", func(ctx context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			email := strings.TrimSpace(rc.Request.PostFormValue("email"))
			if !strings.Contains(email, "@") {
				rc.Set("contact:form", contactView{Error: "That does not look like an email address.", Email: email})
				result := collage.RenderPage(page)
				result.Status = http.StatusUnprocessableEntity
				return result, nil
			}
			if err := messages.Send(ctx, email, rc.Request.PostFormValue("message")); err != nil {
				return nil, err
			}
			return collage.SeeOther("/contact/thanks"), nil
		}).
		Build()
	return page
}

func contactData(_ context.Context, rc *collage.RenderContext) (any, []string, error) {
	// Set by the action when this render is its answer; empty on an ordinary GET.
	view, _ := collage.Get[contactView](rc, "contact:form")
	return view, nil, nil
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
refuses it with `collage.ErrUnregisteredPage`, naming the page. The closure over
`page` above is the simplest way to hand the action its own page.

To redirect to a page by name rather than by path, use
[`app.URL`](/docs/links-and-locales#links-from-go).

## Forgery protection

Every request to an action with an unsafe method (anything but `GET`, `HEAD` and
`OPTIONS`) is checked for a forgery token before the handler runs. A
request without a valid one is a `403`, and the handler never sees it. Error hooks
see why: `collage.ErrCSRFMissing` for no token or no cookie,
`collage.ErrCSRFMismatch` for a token that is not the cookie's, and
`collage.ErrCSRFInvalid` for one this application did not sign.

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
is not told about the key. That is fine for a first run and wrong to deploy: a generated key is
different in every process, so a form loaded before a restart is refused after it,
and one served by one instance is refused by the next. A cached page with a form,
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

The limit is applied before your handler runs, not by it. A limit every handler has
to remember is a limit the one that forgot does not have — and that is the one an
anonymous caller will find. A handler that reads past it gets an
`*http.MaxBytesError`, and returning that error answers `413`:

```go
if err := rc.Request.ParseForm(); err != nil {
	return nil, err // a body over the limit becomes a 413
}
```

When the token is checked from the form, the check reads the body first, so an
oversized form is refused there, before your handler runs — still with a `413`,
not a `403`: a body too large to read is not a forgery.

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

## A fragment at its own URL

Refreshing part of a page without a client framework takes two things: a URL that
answers with only that part, and a few lines of JavaScript to swap it in.
`WithFragmentPath` is the first:

```go
collage.NewPage("search").
	WithLayout(layout).
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
stats, err := collage.Cached(rc, "system:stats", time.Second, nil,
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
most of it. A handler that sets `Cache-Control` itself keeps its own.

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
if problems := validate(rc); len(problems) > 0 {
	rc.Set("problems", problems)
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
