---
description: Test the real application through app.Handler() with pkg/collagetest — a client that keeps cookies and submits forms the way a browser does — pages, forms with their forgery token, documents and the static export.
reference: NewBuilder, BuildOptions, App
---

# Testing

A collage application is an `http.Handler`. `app.Handler()` returns it, with no
server listening and no port to choose, so an ordinary Go test drives the whole
site: routing, data handlers, templates, the cache, forms, documents, plugins and
middleware, exactly as they run in production. `pkg/collagetest` (since v0.40.0)
is the client that does it the way a browser does: it keeps the cookies a response
sets, and submits a form with every hidden field the page put in it.

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

If starting fails — a plugin's `Init` fails, a page names an unregistered error page,
a mount shadows a route — `app.Handler()` answers every request with a 503 and logs
why. Call `app.Start()` first when you want that failure as an error in the test:

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

## Isolate the disk cache

The scaffold's cache directory is a package variable, and `client` points it at a
fresh directory before building the application:

```go
// cacheDir is where rendered pages are kept between restarts. A variable so the
// tests can point it at a directory of their own.
var cacheDir = ".cache"
```

With development mode off, the application uses the disk cache, and a disk cache is
shared by everything that uses the same directory and the same build. Without
`cacheDir = t.TempDir()` a test could be served a page an earlier test rendered —
or an earlier run, since the cache survives the process — and pass or fail for
reasons that have nothing to do with the code. It would also leave a `.cache`
directory in your package.

Because `cacheDir` is shared by the whole package, these tests do not call
`t.Parallel()`. If you want parallel tests, pass the directory to `newApp` as a
parameter instead.

A test that is about caching can use this: make two requests with one client and
check the second is served from the cache, or call `app.InvalidateTags` between
them and check it is not.

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
the same name. Visible fields are the test's to fill; a trap field a bot would fill
stays empty, so the submission is a reader's.

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
header on a `c.Request` before `c.Do`. The client is a convenience, not a
requirement: `app.Handler()` takes a `net/http/httptest` recorder as well as any
other handler does.

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

To assert on skips and warnings as well, call the builder directly and read the
report:

```go
builder, err := collage.NewBuilder(app, collage.BuildOptions{OutDir: t.TempDir()})
if err != nil {
	t.Fatal(err)
}
report, err := builder.Build(context.Background())
if err != nil {
	t.Fatalf("build: %v", err)
}
for _, skip := range report.Skipped {
	if skip.Page == "home" {
		t.Errorf("home was skipped: %s", skip.Reason)
	}
	if errors.Is(skip.Err, collage.ErrDynamicPathUnresolved) {
		t.Errorf("%s has a {param} and no WithStaticParams", skip.Page)
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

## Running them

```sh
go test ./...
```

Run it in CI before building or exporting. A test that uses the real `newApp` fails
there on anything that would stop the site from starting, which is cheaper than
finding out on the server.
