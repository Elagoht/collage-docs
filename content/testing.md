---
description: Test the real application through app.Handler() with pkg/collagetest, a client that keeps cookies and submits forms like a browser: pages, forms, the export.
reference: Build, BuildOptions, App
---

# Testing

A collage application is an `http.Handler`. `app.Handler()` returns it, with no
server listening and no port to choose, so an ordinary Go test drives the whole
site: routing, data handlers, templates, the cache, forms, documents, plugins and
middleware, exactly as they run in production.

`pkg/collagetest` (since v0.40.0) is the client that does it the way a browser
does: it keeps the cookies a response sets, and submits a form with every hidden
field the page put in it.

```go
import "github.com/Elagoht/collage/pkg/collagetest"

func TestLogin(t *testing.T) {
	c := client(t) // below: a collagetest.Client on the application main builds

	page := c.Get("/login").WantStatus(http.StatusOK)
	res := c.Submit(page, "/login", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct horse"},
	}).WantStatus(http.StatusSeeOther)

	if res.Location() != "/panel" {
		t.Errorf("Location = %q, want /panel", res.Location())
	}
	c.Follow(res).WantStatus(http.StatusOK) // the page the login's cookie opens
}
```

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 300" role="img" aria-labelledby="tst-flow-title">
<title id="tst-flow-title">The login test, request by request, between a collagetest.Client and app.Handler(), with no server and no port. GET /login answers 200 with the form and sets a cookie, which the client's jar keeps. Submit posts the form with its hidden fields and the cookie. The answer, a 303 to /panel, is returned to the test, not followed. Follow sends a GET for /panel with the cookies, and the page answers 200. Each request is one ServeHTTP call.</title>
<defs><marker id="tst-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="8" width="128" height="36" rx="6" /><text x="68" y="23" text-anchor="middle">collagetest.Client</text><text class="d-small" x="68" y="37" text-anchor="middle">cookie jar</text>
<rect class="d-c" x="188" y="8" width="128" height="36" rx="6" /><text class="d-on-c" x="252" y="23" text-anchor="middle">app.Handler()</text><text class="d-on-c d-small" x="252" y="37" text-anchor="middle">no server · no port</text>
<line class="d-rule" x1="68" y1="44" x2="68" y2="272" />
<line class="d-rule" x1="252" y1="44" x2="252" y2="272" />
<text class="d-small" x="160" y="64" text-anchor="middle">GET /login</text>
<line class="d-flow" x1="68" y1="70" x2="248" y2="70" marker-end="url(#tst-arrow)" />
<text class="d-small" x="160" y="92" text-anchor="middle">200 · the form, Set-Cookie</text>
<line class="d-flow" x1="252" y1="98" x2="72" y2="98" marker-end="url(#tst-arrow)" />
<rect class="d-chip" x="4" y="108" width="128" height="22" rx="11" /><text class="d-small" x="68" y="123" text-anchor="middle">the jar keeps it</text>
<text class="d-small" x="160" y="152" text-anchor="middle">Submit · hidden fields + cookie</text>
<line class="d-flow" x1="68" y1="158" x2="248" y2="158" marker-end="url(#tst-arrow)" />
<text class="d-small" x="160" y="180" text-anchor="middle">303 · Location: /panel</text>
<line class="d-flow" x1="252" y1="186" x2="72" y2="186" marker-end="url(#tst-arrow)" />
<rect class="d-chip-hit" x="4" y="196" width="128" height="22" rx="11" /><text class="d-hit-text d-small" x="68" y="211" text-anchor="middle">returned, not followed</text>
<text class="d-small" x="160" y="238" text-anchor="middle">Follow · GET /panel + cookies</text>
<line class="d-flow" x1="68" y1="244" x2="248" y2="244" marker-end="url(#tst-arrow)" />
<text class="d-small" x="160" y="262" text-anchor="middle">200 · the page</text>
<line class="d-flow" x1="252" y1="268" x2="72" y2="268" marker-end="url(#tst-arrow)" />
<text class="d-small" x="160" y="292" text-anchor="middle">each request: one ServeHTTP call</text>
</svg>
<figcaption>The test above, request by request. Nothing listens on a port: each request is one call to the handler <code>app.Handler()</code> returns, with a recorder for its answer. The jar keeps the cookies a response sets and sends them back. <code>Submit</code> posts the form with every hidden field it holds, the forgery token among them. A redirect comes back to the test as it is, and <code>Follow</code> is the separate <code>GET</code> that takes it.</figcaption>
</figure>

The demo project `collage new --template demo` scaffolds comes with a test file
built this way. This page walks through its pattern.

## Test the application `main` builds

The scaffolded `main.go` builds the application in a function of its own:

```go
func newApp(devMode bool, port int) (*collage.App, error)
```

`main` calls it to serve, and the tests call it to test. That is the most important
decision in the test file: what the tests exercise is the site that actually runs,
with its real configuration, routes and mounts, not a second wiring that slowly
drifts from it.

```go
// client returns a browser of its own on a fresh copy of the site.
func client(t *testing.T) *collagetest.Client {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v, want nil", err)
	}
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })
	return collagetest.New(t, app.Handler())
}
```

`newApp(false, 0)` builds the production configuration: development mode off, and a
port of `0`, which collage replaces with its default — nothing listens anyway.

A test is then a request and a look at the response:

```go
func TestPagesRender(t *testing.T) {
	c := client(t)
	for target, want := range map[string]string{
		"/":         "Explore the features",
		"/features": "Four live demos",
		"/hello":    "Hello, stranger!",
	} {
		res := c.Get(target)
		if res.Status != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, res.Status)
			continue
		}
		if !strings.Contains(res.Body, want) {
			t.Errorf("GET %s does not contain %q", target, want)
		}
		// Counted, not merely found: a layout writing one and a page hoisting
		// another is two titles, which is a page that looks fine and is not.
		if n := strings.Count(res.Body, "<title>"); n != 1 {
			t.Errorf("GET %s has %d titles, want exactly 1", target, n)
		}
	}
}
```

Build a fresh client per test, as `client(t)` does. The first call to
`app.Handler()` starts the application and closes registration, and each test then
starts from an empty cache and an empty cookie jar. Two readers are two clients.

Starting can fail:

- a plugin's `Init` fails;
- a page names an unregistered error page;
- a mount shadows a route.

`app.Handler()` then answers every request with a 503 and logs why. Call
`app.Start()` first when you want that failure as an error in the test:

```go
if err := app.Start(); err != nil {
	t.Fatalf("start: %v", err)
}
```

## The client

| | |
| --- | --- |
| `collagetest.New(t, h)` | A client for `h` with an empty cookie jar |
| `c.Get(target)` | A `GET` of a path or an absolute URL |
| `c.Submit(page, action, values)` | Submits the form of `page` whose action is `action` |
| `c.Follow(res)` | A `GET` of the `Location` a redirect names |
| `c.Request(method, target, body)` / `c.Do(req)` | Any other request — a JSON body, a header of its own. `Do` attaches the jar's cookies and keeps the ones the response sets |

A `*Response` carries `Status`, `Header`, `Body`, `URL` and `Method`.
`WantStatus(code)` fails the test with the body when the status differs, and
returns the response, so a request and its check read as one line. `Location()` is
the `Location` header, and `CSRFToken()` the value of the page's first hidden
`_csrf` input.

The jar is a `net/http/cookiejar`, scoped by path and expiry as a browser scopes
cookies. A path alone is addressed to `http://example.com`, the host
`net/http/httptest` uses. A site that marks its cookies `Secure` uses absolute
`https://example.com/...` targets, which arrive as over TLS, so the jar sends those
cookies back.

Redirects are not followed: a test usually wants to see the `303` and where it
points. `Follow` takes it when the page behind it is what the test is about.

## Invalidation and warming

Since v0.64.0, after an `InvalidateTags` or `PurgeTags` the application
re-renders the pages it reached in the background (see
[Caching](/docs/caching#warming-after-an-invalidation)). A test that asserts
`X-Collage-Cache: STALE` (or a `MISS`) right after an invalidation races that
warmer; set `WarmLimit: -1` in `CacheConfig` for such a test. Warming is already
off in development.

A test that asserts the opposite, a `HIT` once the warmer has re-rendered the
page, waits for it with `site.WaitWarmed()` on a
[`collagetest.Site`](#test-a-plugin-with-a-site-of-its-own), which returns once
nothing is queued or rendering, and at once when warming is off:

```go
if err := site.App.InvalidateTags(ctx, "posts"); err != nil {
	t.Fatal(err)
}
site.WaitWarmed()
res := site.Get("/posts").WantStatus(http.StatusOK)
// res.Header.Get(collage.CacheStatusHeader) == "HIT"
```

## Isolate the disk cache

The scaffold's cache directory is a package variable, and `client` points it at a
fresh directory before building the application:

```go
// cacheDir is where rendered pages are kept between restarts. A variable so the
// tests can point it at a directory of their own.
var cacheDir = ".cache"
```

With development mode off, the application uses the disk cache, and a disk cache is
shared by everything that uses the same directory and the same build.

Without `cacheDir = t.TempDir()` a test could be served a page an earlier test
rendered — or an earlier run, since the cache survives the process — and pass or
fail for reasons that have nothing to do with the code. It would also leave a
`.cache` directory in your package.

Shut the application down in a `t.Cleanup` registered after it, as `client` does:
the data cache writes to that directory in the background, and `App.Shutdown`
waits for those writes, so the directory is not removed while one is still being
made. `collagetest.NewSite` does this itself.

Because `cacheDir` is shared by the whole package, these tests do not call
`t.Parallel()`. If you want parallel tests, pass the directory to `newApp` as a
parameter instead.

A test that is about caching can use this: make two requests with one client and
check the second is served from the cache, or call `app.InvalidateTags` between
them and check it is not.

## A fleet on one shared cache

An application that runs as several instances on a
[shared cache](/docs/caching#a-cache-a-fleet-shares) is tested as two Apps on one
store. `collagetest.NewSharedCache()` (since v0.66.0) is an in-memory
`collage.SharedCache` standing in for Redis: a page one App renders is served by
the other, and an invalidation one sends is seen by both.

```go
store := collagetest.NewSharedCache()
shared := collagetest.WithConfig(func(c *collage.Config) {
	c.Cache = collage.CacheConfig{Enabled: true, Store: store}
})
a := collagetest.NewSite(t, shared)
b := collagetest.NewSite(t, shared)

a.Get("/").WantStatus(http.StatusOK) // X-Collage-Cache: MISS, rendered by a
b.Get("/").WantStatus(http.StatusOK) // X-Collage-Cache: HIT, from the store
```

A store of your own is checked with `collagetest.CacheSuite`, for the `Cache`,
`TaggedCache` and `StaleCache` contracts, and `collagetest.SharedCacheSuite`, for
`SharedCache`: see [Caching](/docs/caching#writing-your-own-sharedcache).

## Forms and the forgery token

Every action behind an unsafe method — a form post, a `fetch()` — checks a
request-forgery token, and a test sends one the way a browser does: it `GET`s the
page with the form, which sets the `collage_csrf` cookie and renders
`{{csrfToken}}` as a hidden `_csrf` field, and posts the form with that field and
that cookie. `Submit` does all of it:

```go
func TestHelloGreetsTheSubmittedName(t *testing.T) {
	c := client(t)

	res := c.Submit(c.Get("/features"), "/hello", url.Values{"name": {"Ada"}}).WantStatus(http.StatusOK)
	if !strings.Contains(res.Body, "Hello, Ada!") {
		t.Errorf("the response does not greet the name:\n%s", res.Body)
	}
}
```

`Submit` finds the form, carries across **every hidden input** it holds — the
forgery token, and whatever a plugin stamps into a form, such as a honeypot's signed
timestamp — and puts `values` on top: a name in `values` replaces a hidden input of
the same name.

Visible fields are the test's to fill. A trap field a bot would fill stays empty,
so the submission is a reader's.

- `action` is resolved against the page's URL and compared decoded, as is the
  form's: `"/login"`, `"login"` from a page beside it, an absolute URL, and
  `"/giriş"` for a form whose action is `"/giri%c5%9f"` all name the same form. A
  form with no `action` submits to its page. An empty `action` means the page's
  only form. No match, or several, fails the test and lists the actions the page's
  forms have.
- The form's `method` and `enctype` are honoured: a `GET` form sends its values in
  the query, `multipart/form-data` is sent as multipart, anything else as
  `application/x-www-form-urlencoded`.
- Forms are found by a scanner, not an HTML parser: comments and the bodies of
  `<script>`, `<style>`, `<template>` and `<textarea>` are skipped, so markup
  written out as text there is not taken for a form. A disabled hidden input is
  not sent, as a browser does not send it.

A JSON endpoint takes the token in the `X-CSRF-Token` header, as a `fetch()` sends
it:

```go
func TestCountAnswersWithTheNewCount(t *testing.T) {
	c := client(t)
	page := c.Get("/features")

	req := c.Request(http.MethodPost, "/api/count", nil)
	req.Header.Set("X-CSRF-Token", page.CSRFToken())
	res := c.Do(req).WantStatus(http.StatusOK)

	var answer struct{ Count int64 }
	if err := json.Unmarshal([]byte(res.Body), &answer); err != nil {
		t.Fatalf("body = %q, want JSON: %v", res.Body, err)
	}
	if answer.Count < 1 {
		t.Errorf("count = %d, want at least 1 after a click", answer.Count)
	}
}
```

And the test that keeps the protection on:

```go
// Without a token a submission never reaches its handler. This is the test that
// fails if the protection is ever turned off by accident.
func TestASubmissionWithNoTokenIsRefused(t *testing.T) {
	c := client(t)
	for _, target := range []string{"/api/count", "/hello"} {
		req := c.Request(http.MethodPost, target, strings.NewReader(url.Values{"name": {"Ada"}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if res := c.Do(req); res.Status != http.StatusForbidden {
			t.Errorf("POST %s with no token = %d, want 403", target, res.Status)
		}
	}
}
```

The token is signed with `Security.CSRFKey`, or with a key generated per
application when none is set. Either way the page and the post in one test go to the
same application, so the tests need no key of their own. `_csrf` and `collage_csrf`
are the default names; an application that renames them with
`Security.CSRFFieldName` or `CSRFCookieName` renames them in its tests too.

## Documents, redirects and status codes

A document is tested like a page. Check the content type as well as the body, since
the content type is half of what a document is:

```go
func TestHealthCheck(t *testing.T) {
	res := client(t).Get("/healthz").WantStatus(http.StatusOK)

	var health struct{ Status string }
	if err := json.Unmarshal([]byte(res.Body), &health); err != nil || health.Status != "ok" {
		t.Errorf("body = %q, want a status of ok", res.Body)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
```

The not-found page, a 405 with its `Allow` header, a redirect's `Location`, a
`Cache-Control` header — each is a field on the response:

```go
func TestNotFoundPage(t *testing.T) {
	res := client(t).Get("/there-is-nothing-here").WantStatus(http.StatusNotFound)

	if !strings.Contains(res.Body, "There is nothing at this address") {
		t.Errorf("body = %q, want this site's own not-found page", res.Body)
	}
}
```

Middleware registered with `app.Use` runs in these tests too, since it is part of
the handler. To test a preview or a `collage.Vary` dimension, set the cookie or
header on a `c.Request` before `c.Do`.

The client is a convenience, not a requirement: `app.Handler()` takes a
`net/http/httptest` recorder as well as any other handler does.

## Every page renders

A test that visits every page catches the template that fails only on one of them.
collage-docs loads its content and requests each page — here with one client, so
one application, for the whole walk:

```go
// Every page of the documentation renders, with its own title.
func TestEveryDocRenders(t *testing.T) {
	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatalf("site.Load: %v", err)
	}
	c := client(t)
	for _, page := range loaded.Pages() {
		res := c.Get(page.URL())
		if res.Status != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), res.Status)
			continue
		}
		if !strings.Contains(res.Body, "<title>"+page.Title+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
	}
}
```

For a site whose pages come from a CMS, the same test can walk the values your
pages' [`WithStaticParams`](/docs/static-export#dynamic-paths-withstaticparams)
functions list.

## Testing the export

The static export is code too, and the one place where a skipped page is silent in
production. Run it into `t.TempDir()` and check the files you expect are there:

```go
// The export writes the home page, every doc and the 404 page.
func TestExport(t *testing.T) {
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := staticBuild(app, out, false); err != nil {
		t.Fatalf("staticBuild: %v", err)
	}
	loaded, _ := site.Load(content.FS)
	want := []string{"index.html", "404.html"}
	for _, page := range loaded.Pages() {
		want = append(want, filepath.Join("docs", page.Slug, "index.html"))
	}
	for _, file := range want {
		if _, err := os.Stat(filepath.Join(out, file)); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}
```

It calls the same `staticBuild` that `collage export` runs, so the test covers
`WithStaticParams` and the builder options as well as the pages. `staticBuild` returns
the build's error, which fails the test for a degraded page, an empty render or a
panic.

To assert on skips and warnings as well, call `collage.Build` directly and read the
report:

```go
report, err := collage.Build(context.Background(), app, collage.BuildOptions{OutDir: t.TempDir()})
if err != nil {
	t.Fatalf("build: %v", err)
}
for _, skip := range report.Skipped {
	if skip.Name == "home" {
		t.Errorf("home was skipped: %s", skip.Reason)
	}
	if errors.Is(skip.Err, collage.ErrDynamicPathUnresolved) {
		t.Errorf("%s has a {param} and no WithStaticParams", skip.Name)
	}
}
if len(report.Warnings) > 0 {
	t.Errorf("warnings: %v", report.Warnings)
}
```

`skip.Err` is the sentinel behind each skip, since v0.10.0; see
[Errors](/docs/errors#static-builds).

## Rendering without HTTP

`app.RenderPath` renders one page the way the export does — no request, no page
cache, no middleware — and returns the HTML and what the render did. The render
hooks of your plugins still run, and so does the data cache: a value
[`collage.Cached`](/docs/caching#caching-data-across-pages) stored during one
render, or one served request, is what the next `RenderPath` on the same
application is given. Build a fresh application when a test needs fresh data:

```go
result, err := app.RenderPath(context.Background(), "/blog/hello-world", "", nil)
if err != nil {
	t.Fatalf("render: %v", err)
}
if result.Degraded() {
	t.Errorf("a fragment failed: %+v", result.Metadata)
}
```

`app.RenderDocumentPath` does the same for a document. Most tests are better served
by `app.Handler()` and `collagetest`, which test what a reader gets; these are for looking at one
render in detail.

## Every link resolves

A link built by name — `{{pageURL "post" "slug" .Slug}}`, `{{actionURL "logout"}}` —
fails only when the template renders, and only on the page that reaches it.
`app.Check()` (since v0.40.0) checks every template's links at once, with nothing
rendered, and is what [`collage check`](/docs/cli#collage-check) runs:

```go
func TestLinks(t *testing.T) {
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if findings := app.Check(); len(findings) > 0 {
		t.Errorf("broken links: %v", findings)
	}
}
```

## Test a plugin with a site of its own

A plugin has no application of its own to build. `collagetest.NewSite(t, ...)`
(since v0.61.0) builds a real one around it — a layout, a home page at `/`, a
captured debug log — starts it, shuts it down when the test ends, and gives back
the app, a `Client` on it and its `Logs`:

```go
site := collagetest.NewSite(t,
	collagetest.WithPlugins(robots.New(robots.Options{})),
	collagetest.WithPluginConfig("elagoht/robots", `{"disallow":["/admin"]}`),
)
res := site.Get("/robots.txt").WantStatus(http.StatusOK)
```

The options, `TrySite` for a start that is meant to fail, and `site.Shutdown` are
in [Writing a plugin](/docs/writing-plugins#testing-a-plugin).

## Example: a team task list

The team task list from
[Forms and actions](/docs/forms-and-actions#example-a-team-task-list) is a page at
`/teams/{team}/tasks` with an "add a task" form, the list as a fragment of its
own at `/teams/{team}/tasks/list`, and an action that answers in three ways. Put
it behind a layout whose guard sends a signed-out reader to `/login?next=…`, like
`requireUser` in [Pages and layouts](/docs/pages-and-layouts#private-pages-guards),
and give the site a `/login` form whose action sets the session cookie and
redirects to `next`. The session is the application's own `sid` cookie, read by a
middleware as in
[Middleware and your own API](/docs/middleware-and-apis#example-a-teams-task-list).
With [elagoht/session](/docs/plugins#elagohtsession) instead, a signed-in reader's
page is rendered fresh on every request, and the cache test below would see a
`MISS` where it asks for a `HIT`.

Its tests, one per thing a reader can do:

| Test | What it catches |
| --- | --- |
| `TestAMemberAddsATask` | The form, its token or the `303` back to the list breaking |
| `TestAnEmptyTitleIsRefused` | A bad title answered with anything but `422` and the form with its error |
| `TestAFetchGetsTheListAlone` | A script's post answered with a whole page, or a stale list |
| `TestASignedOutReaderSignsInFirst` | The page, its fragment path or its action left open |
| `TestTheCachedListChangesWithATask` | A missing `InvalidateTags`: the cached page keeps the old list |

They use `client(t)` from [above](#test-the-application-main-builds), and one
helper that signs a client in. The cookie the sign-in sets stays in that client's
jar, so every later request of the test is the member's:

```go
// signIn submits the sign-in form; the session cookie it sets stays in c's jar.
func signIn(t *testing.T, c *collagetest.Client, email string) {
	t.Helper()
	c.Submit(c.Get("/login"), "/login", url.Values{
		"email":    {email},
		"password": {"correct horse"},
	}).WantStatus(http.StatusSeeOther)
}
```

A member adds a task with the form, as a browser posts it: `Submit` carries the
hidden `_csrf` field, the answer is a `303` to the list, and `Follow` loads it:

```go
func TestAMemberAddsATask(t *testing.T) {
	c := client(t)
	signIn(t, c, "ada@example.com")

	page := c.Get("/teams/blue/tasks").WantStatus(http.StatusOK)
	res := c.Submit(page, "/teams/blue/tasks", url.Values{"title": {"Write the release notes"}}).
		WantStatus(http.StatusSeeOther)
	if res.Location() != "/teams/blue/tasks" {
		t.Errorf("Location = %q, want the list", res.Location())
	}
	if list := c.Follow(res); !strings.Contains(list.Body, "Write the release notes") {
		t.Errorf("the list does not show the new task:\n%s", list.Body)
	}
}
```

An empty title is refused. A browser's post gets the whole page back with `422`,
the form in it with the error:

```go
func TestAnEmptyTitleIsRefused(t *testing.T) {
	c := client(t)
	signIn(t, c, "ada@example.com")

	page := c.Get("/teams/blue/tasks").WantStatus(http.StatusOK)
	res := c.Submit(page, "/teams/blue/tasks", url.Values{"title": {"   "}}).
		WantStatus(http.StatusUnprocessableEntity)
	for _, want := range []string{"<html", `id="task-form"`, "A title is 1 to 200 characters."} {
		if !strings.Contains(res.Body, want) {
			t.Errorf("the 422 has no %s:\n%s", want, res.Body)
		}
	}
}
```

A script posts the same form with `Collage-Fetch` and takes the token from the
page, as the
[script on that page](/docs/forms-and-actions#example-a-team-task-list) does with
`new FormData(form)`. The answer is the list fragment alone, and so is
`/teams/{team}/tasks/list`:

```go
func TestAFetchGetsTheListAlone(t *testing.T) {
	c := client(t)
	signIn(t, c, "ada@example.com")
	page := c.Get("/teams/blue/tasks").WantStatus(http.StatusOK)

	body := url.Values{"_csrf": {page.CSRFToken()}, "title": {"Book the venue"}}
	req := c.Request(http.MethodPost, "/teams/blue/tasks", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(collage.FetchHeader, "1")
	res := c.Do(req).WantStatus(http.StatusOK)

	if strings.Contains(res.Body, "<html") || !strings.Contains(res.Body, `id="task-list"`) {
		t.Errorf("a fetch was not answered with the list fragment alone:\n%s", res.Body)
	}
	if !strings.Contains(res.Body, "Book the venue") {
		t.Errorf("the fragment does not show the new task:\n%s", res.Body)
	}
	list := c.Get("/teams/blue/tasks/list").WantStatus(http.StatusOK)
	if strings.Contains(list.Body, "<html") || !strings.Contains(list.Body, "Book the venue") {
		t.Errorf("/teams/blue/tasks/list is not the current list alone:\n%s", list.Body)
	}
}
```

A signed-out reader is redirected from the page and from its fragment path, which
meets the page's guards. A post meets the guard before the forgery check, so it is
the same redirect, not a `403`. Then the whole way in, through the hidden `next`
field of the sign-in form:

```go
func TestASignedOutReaderSignsInFirst(t *testing.T) {
	c := client(t)
	for _, target := range []string{"/teams/blue/tasks", "/teams/blue/tasks/list"} {
		res := c.Get(target).WantStatus(http.StatusSeeOther)
		if !strings.HasPrefix(res.Location(), "/login?next=") {
			t.Errorf("GET %s: Location = %q, want the sign-in page", target, res.Location())
		}
	}
	// The guard answers before the forgery check: a redirect, not a 403.
	req := c.Request(http.MethodPost, "/teams/blue/tasks", strings.NewReader("title=Sneak+in"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Do(req).WantStatus(http.StatusSeeOther)

	login := c.Follow(c.Get("/teams/blue/tasks")).WantStatus(http.StatusOK)
	res := c.Submit(login, "/login", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct horse"},
	}).WantStatus(http.StatusSeeOther)
	if res.Location() != "/teams/blue/tasks" {
		t.Errorf("after signing in, Location = %q, want the page asked for", res.Location())
	}
	c.Follow(res).WantStatus(http.StatusOK)
}
```

The page is `Incremental(time.Hour)`, so its second load is a `HIT`. Adding a task
invalidates the team's tag before the `303` is written, and the next load shows
the task:

```go
func TestTheCachedListChangesWithATask(t *testing.T) {
	c := client(t)
	signIn(t, c, "ada@example.com")
	cache := func(res *collagetest.Response) string { return res.Header.Get(collage.CacheStatusHeader) }

	c.Get("/teams/green/tasks").WantStatus(http.StatusOK)
	page := c.Get("/teams/green/tasks").WantStatus(http.StatusOK)
	if cache(page) != "HIT" {
		t.Fatalf("the second load is %q, want HIT", cache(page))
	}

	res := c.Submit(page, "/teams/green/tasks", url.Values{"title": {"Order the banners"}}).
		WantStatus(http.StatusSeeOther)
	after := c.Follow(res).WantStatus(http.StatusOK)
	if !strings.Contains(after.Body, "Order the banners") {
		t.Errorf("the cached page still shows the old list:\n%s", after.Body)
	}
	if cache(after) != "MISS" {
		t.Errorf("the load after the change is %q, want MISS", cache(after))
	}
	if again := c.Get("/teams/green/tasks"); cache(again) != "HIT" {
		t.Errorf("the next load is %q, want HIT", cache(again))
	}
}
```

The test asserts the `MISS` without `WaitWarmed` or `WarmLimit: -1`, because the
page is guarded and a guarded page is
[not warmed](/docs/caching#warming-after-an-invalidation): its first reader after
the change renders it. On a page without a guard the warmer may have rendered it
first, so assert the `HIT` after `WaitWarmed()` on a
[`collagetest.Site`](#test-a-plugin-with-a-site-of-its-own), or set
`WarmLimit: -1` to assert the `MISS`, as
[Invalidation and warming](#invalidation-and-warming) describes.

## Running them

```sh
go test ./...
```

Run it in CI before building or exporting. A test that uses the real `newApp` fails
there on anything that would stop the site from starting, which is cheaper than
finding out on the server.
