---
description: Test the real application through app.Handler() and net/http/httptest — pages, forms with their forgery token, documents and the static export.
reference: NewBuilder, BuildOptions, App
---

# Testing

A collage application is an `http.Handler`. `app.Handler()` returns it, with no
server listening and no port to choose, so `net/http/httptest` drives the whole site
from an ordinary Go test: routing, data handlers, templates, the cache, forms,
documents and middleware, exactly as they run in production.

The project `collage new` scaffolds comes with a test file built this way. This page
walks through its pattern.

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
func handler(t *testing.T) http.Handler {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v, want nil", err)
	}
	return app.Handler()
}

// get returns the response to a GET of target, and its body.
func get(t *testing.T, h http.Handler, target string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec, rec.Body.String()
}
```

`newApp(false, 0)` builds the production configuration: development mode off, and a
port of `0`, which collage replaces with its default — nothing listens anyway.

A test is then a request and a look at the response:

```go
func TestPagesRender(t *testing.T) {
	h := handler(t)
	for target, want := range map[string]string{
		"/":         "Explore the features",
		"/features": "Four live demos",
		"/hello":    "Hello, stranger!",
	} {
		rec, body := get(t, h, target)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", target, rec.Code)
			continue
		}
		if !strings.Contains(body, want) {
			t.Errorf("GET %s does not contain %q", target, want)
		}
		// Counted, not merely found: a layout writing one and a page hoisting
		// another is two titles, which is a page that looks fine and is not.
		if n := strings.Count(body, "<title>"); n != 1 {
			t.Errorf("GET %s has %d titles, want exactly 1", target, n)
		}
	}
}
```

Build a fresh handler per test, as `handler(t)` does. The first call to
`app.Handler()` starts the application and closes registration, and each test then
starts from an empty cache.

If starting fails — a plugin's `Init` fails, a page names an unregistered error page,
a mount shadows a route — `app.Handler()` answers every request with a 503 and logs
why. Call `app.Start()` first when you want that failure as an error in the test:

```go
if err := app.Start(); err != nil {
	t.Fatalf("start: %v", err)
}
```

## Isolate the disk cache

The scaffold's cache directory is a package variable, and `handler` points it at a
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

A test that is about caching can use this: make two requests to one handler and
check the second is served from the cache, or call `app.InvalidateTags` between
them and check it is not.

## Forms and the forgery token

Every action behind an unsafe method — a form post, a `fetch()` — checks a
request-forgery token, and a test has to send one the way a browser does:

1. `GET` the page with the form. The response sets the `collage_csrf` cookie, and
   `{{csrfToken}}` in the page renders a hidden `_csrf` field carrying the same
   token.
2. Read the token out of the page.
3. `POST` the form with the token in the `_csrf` field — or in the `X-CSRF-Token`
   header, as a `fetch()` does — and the cookie from step 1.

The scaffold's helpers:

```go
// token reads the forgery token out of a rendered page, the way a browser does.
func token(t *testing.T, body string) string {
	t.Helper()
	m := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no forgery token in the page:\n%s", body)
	}
	return m[1]
}

// post sends a POST to target, carrying the cookies from page.
func post(t *testing.T, h http.Handler, target string, form url.Values, header http.Header, page *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for name, values := range header {
		req.Header[name] = values
	}
	if page != nil {
		for _, c := range (&http.Response{Header: page.Header()}).Cookies() {
			req.AddCookie(c)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
```

`(&http.Response{Header: page.Header()}).Cookies()` parses the recorder's
`Set-Cookie` headers with the standard library's own cookie parser, so the test sends
back exactly what a browser would.

A form post, with the token in the field:

```go
func TestHelloGreetsTheSubmittedName(t *testing.T) {
	h := handler(t)
	page, body := get(t, h, "/features")

	rec := post(t, h, "/hello", url.Values{"_csrf": {token(t, body)}, "name": {"Ada"}}, nil, page)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /hello = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello, Ada!") {
		t.Errorf("the response does not greet the name:\n%s", rec.Body.String())
	}
}
```

A JSON endpoint, with the token in the header:

```go
func TestCountAnswersWithTheNewCount(t *testing.T) {
	h := handler(t)
	page, body := get(t, h, "/features")

	rec := post(t, h, "/api/count", nil, http.Header{"X-Csrf-Token": {token(t, body)}}, page)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/count = %d, want 200", rec.Code)
	}
	var answer struct{ Count int64 }
	if err := json.Unmarshal(rec.Body.Bytes(), &answer); err != nil {
		t.Fatalf("body = %q, want JSON: %v", rec.Body.String(), err)
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
	h := handler(t)
	for _, target := range []string{"/api/count", "/hello"} {
		if rec := post(t, h, target, url.Values{"name": {"Ada"}}, nil, nil); rec.Code != http.StatusForbidden {
			t.Errorf("POST %s with no token = %d, want 403", target, rec.Code)
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
	rec, body := get(t, handler(t), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz = %d, want 200", rec.Code)
	}
	var health struct{ Status string }
	if err := json.Unmarshal([]byte(body), &health); err != nil || health.Status != "ok" {
		t.Errorf("body = %q, want a status of ok", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
```

The not-found page, a 405 with its `Allow` header, a redirect's `Location`, a
`Cache-Control` header — each is a field on the recorder:

```go
func TestNotFoundPage(t *testing.T) {
	rec, body := get(t, handler(t), "/there-is-nothing-here")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if !strings.Contains(body, "There is nothing at this address") {
		t.Errorf("body = %q, want this site's own not-found page", body)
	}
}
```

Middleware registered with `app.Use` runs in these tests too, since it is part of
the handler. To test a preview or a `collage.Vary` dimension, set the cookie or
header on the request before `ServeHTTP`.

## Every page renders

A test that visits every page catches the template that fails only on one of them.
collage-docs loads its content and requests each page — here with the `handler` and
`get` helpers above, one application for the whole walk:

```go
// Every page of the documentation renders, with its own title.
func TestEveryDocRenders(t *testing.T) {
	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatalf("site.Load: %v", err)
	}
	h := handler(t)
	for _, page := range loaded.Pages() {
		rec, body := get(t, h, page.URL())
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), rec.Code)
			continue
		}
		if !strings.Contains(body, "<title>"+page.Title+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
	}
}
```

For a site whose pages come from a CMS, the same test can walk the list your
[path provider](/docs/static-export#dynamic-paths-pathprovider) returns.

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

It calls the same `staticBuild` that `collage export` runs, so the test covers the
path provider and the builder options as well as the pages. `staticBuild` returns
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
		t.Errorf("%s has a {param} the path provider does not cover", skip.Page)
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
by `app.Handler()`, which tests what a reader gets; these are for looking at one
render in detail.

## Running them

```sh
go test ./...
```

Run it in CI before building or exporting. A test that uses the real `newApp` fails
there on anything that would stop the site from starting, which is cheaper than
finding out on the server.
