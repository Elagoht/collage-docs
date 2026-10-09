---
description: Render the site to static files with collage export — what is written, what is skipped and why, dynamic paths, and publishing to a static host.
reference: NewBuilder, BuildOptions, BuildReport, PrintBuildReport, StaticParamsFunc, SkipRecord, ErrNotStatic, ErrGuarded, ErrDynamicPathUnresolved, ErrRouteParams, ErrBuildFindings, ErrDuplicateRedirect, ErrRedirectShadowsFile, IsCapture, BuiltFile, BuiltRedirect, RedirectSource
---

# Static export

A site whose pages do not depend on the request does not need a server at all.
`collage export` renders every page that can be a file into `dist/`, copies your
static files beside them, and writes the site's own `404.html`. Put the directory on
any static host.

It is the same program that serves the site, rendering through the same templates,
data handlers and plugins. There is no second build, and nothing to keep in step
with the server. The pages you are reading were produced this way.

```sh
collage export          # -> dist/
collage export -clean   # empty dist/ first
collage serve           # look at dist/ the way a static host would serve it
```

## How it runs

`collage export` does not load your application — it cannot, since your application
is your code. It runs your program in a special mode instead:

```sh
go run . -collage-build -out dist        # plus -clean when you passed it
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-out dir` | `dist` | Where the files are written. |
| `-clean` | off | Remove the directory's existing contents first. |

The scaffolded `main.go` honours that contract: on `-collage-build` it builds the
application as usual and, instead of serving it, hands it to collage's builder and
prints what happened. This is the collage-docs version of that function; which
pages exist is not its business, because each page says that
[itself](#dynamic-paths-withstaticparams):

```go
func staticBuild(app *collage.App, outDir string, clean bool) error {
	builder, err := collage.NewBuilder(app, collage.BuildOptions{
		OutDir: outDir,
		Clean:  clean,
	})
	if err != nil {
		return err
	}

	report, buildErr := builder.Build(context.Background())

	collage.PrintBuildReport(os.Stdout, report, buildErr)

	return buildErr
}
```

`Build` returns a report even when it also returns an error: one failing page does
not stop the others, and every failure is in `report.Errors`. The returned error is
all of them joined, and `main` exits non-zero on it, so a CI job fails when a page
could not be written. Skipped pages and warnings do not fail the build; an
error-level [finding](#reading-the-report) does, with every page still written.

A build does not shut the application down either, and nothing serves after it,
so a program that exits once it has built calls `app.Shutdown` itself; the
scaffolded `main.go` does, from collage v0.55.0 (see
[the CLI](/docs/cli#the-contract-with-maingo)). Otherwise no plugin's `Shutdown`
runs: what a plugin holds is never flushed, and work a queue plugin such as
[elagoht/jobs](/docs/plugins#elagohtjobs) was handed during the build vanishes
without being logged.

If you rewrite `main.go`, keep the `-collage-build`, `-out` and `-clean` flags, or
`collage export` stops doing anything useful.

## What is written

| What | Where |
| --- | --- |
| A static or incremental page: `Static()`, `Incremental(ttl)`, or [no strategy and no data handler](/docs/caching#a-page-that-declares-none) | `dist/<path>/index.html`; `/` is `dist/index.html` |
| The same page in a non-default locale | Under the locale's prefix: `dist/tr/<path>/index.html` |
| A static or incremental document: `Static()`, `Incremental(ttl)`, or no strategy and a [fixed body](/docs/documents#a-fixed-body) | Its literal path: `/sitemap.xml` is `dist/sitemap.xml` |
| The same document in a non-default locale | Under the locale's prefix: `dist/tr/sitemap.xml` |
| A document built with `AtRoot` | Its bare path, in every configuration: `dist/robots.txt` |
| A default-locale page or document, with [`PrefixDefault`](/docs/links-and-locales#the-url-decides-the-locale) | Under its prefix too, `dist/en/<path>/index.html` and `dist/en/sitemap.xml`, and `dist/index.html` sends the reader to `/en/` |
| The not-found page | `dist/404.html`, and `dist/<locale>/404.html` for each other locale |
| Every mounted static file system | Under its prefix: `/static/app.css` is `dist/static/app.css` |

A page becomes a directory with an `index.html`, which is what every static host
looks for when asked for `/about`. A [document](/docs/documents) is written at exactly
its path, because a crawler asking for `/robots.txt` must get a file.

**The not-found page** comes from `app.RegisterNotFoundPage`, and its render strategy
is not consulted — it is almost always `Dynamic()`, because it is never worth
caching, and it still belongs in the export. Hosts differ on whether they look for
a nested `tr/404.html`, so both are written. A site with no not-found page gets
no file, and an unknown URL shows whatever the host decides to.

**Mounted assets** are copied under their original names and under the
content-addressed names `{{asset}}` links to (`app.3a3663df.css`), so both kinds of
link work. A mount that is served from a CDN in production, or is too large to
duplicate, can opt out:

```go
app.Mount("/media/", mediaFS, collage.WithoutBuildCopy())
```

See [Static assets](/docs/assets).

## What is skipped

Some pages cannot be files. The build leaves them out and names each one in the
report, with the reason:

- **Dynamic pages and documents.** They exist to render per request: those
  declared `Dynamic()`, and those that declare no strategy and fetch — a page
  rendering a data handler or a slot resolver, a document with a handler. A page
  whose handler returns the same for every reader says `Static()` to be exported.
- **Pages with a form.** A page whose render contains `{{csrfToken}}` is skipped:
  a form needs a server to post to, and a forgery token belongs to one reader. The
  page may be `Static()` on purpose — cached, and invalidated by the action it posts
  to — and it is served rather than exported. The scaffold's `/features` page is
  one.
- **Guarded pages.** A page with a [guard](/docs/pages-and-layouts#private-pages-guards)
  on its layout chain or content fragment is skipped with `collage.ErrGuarded`,
  whatever its strategy. An export renders without a request, so there is no
  reader to ask the guard about, and a file on a static host is served to anyone.
  The running server keeps serving it, guard and all.
- **A `{param}` pattern with no `WithStaticParams`.** `/blog/{slug}` cannot be
  written until something says which slugs exist, and it is skipped with
  `collage.ErrDynamicPathUnresolved`. See below.
- **Two documents on one file.** A document whose `WithStaticParams` lists one set
  of values twice resolves two tasks to one file; the first is written and the rest are
  skipped with `collage.ErrDuplicateOutputPath`. One pattern in two locales is not
  this — each locale is written under its own prefix. See
  [Documents](/docs/documents#documents-in-a-static-export).

Error pages registered without a path are not listed at all: they are not URLs.

Not exported, and not reported, because they are not pages: [actions](/docs/forms-and-actions),
handlers mounted with `app.Handle`, and middleware. An export renders without a
request, so no middleware runs for the files it writes and `collage.Vary` is never
called — each page is written in the version a request with no preferences would
get. The requests that [capture the headers](#headers-and-redirects) afterwards do
go through middleware, but what they render is not written.

## What is warned about

A page that declared, with `WithCacheParams`, which query parameters it reads
renders differently for each of them. A file has no query string: a static host
answers `/blog?page=2` with the `/blog` file. The page is written without a query,
and the report says so, rather than letting a paginated archive look like it works.
A document with `WithCacheParams` — a paginated feed — is warned about the same way
(since v0.10.0; before, only pages were).

If pagination has to work in an export, put the page number in the path —
`/blog/page/{n}` — and list the pages with `WithStaticParams`. [Pagination](#pagination)
shows the two pages and the library that builds the links.

## What fails

These are errors: the page is not written, the build reports it and exits non-zero.

- **A degraded render.** A page where a fragment failed is refused with
  `collage.ErrDegradedRender`, naming the fragment. A served page with a failed
  fragment is shown but never cached; a file has no TTL to recover through, so it
  would keep that failure until the next export. Set `BuildOptions.AllowDegraded`
  if a page with a missing sidebar is better than no page.
- **An empty render**, `collage.ErrEmptyRender`, and an empty document,
  `collage.ErrEmptyDocumentBody`. A zero-byte `index.html` is never written.
- **A panic** in one page, `collage.ErrBuildPanic`. It is recovered and recorded
  against that page; the rest of the build continues.
- **A not-found page with a form**, `collage.ErrUnresolvedToken`, because a static
  host needs that one as a file.
- **Two pages on one output path**, `collage.ErrOutputPathCollision` — two patterns
  that differ only in a trailing slash, or `WithStaticParams` listing one set of
  values twice. No page is rendered when this is found; the documents, the `404.html`
  pages and the mounted assets are still written, and the build still fails.

- **Two redirects from one path**, `collage.ErrDuplicateRedirect` (since v0.52.0) —
  `/old` and `/old/`, or `/blog/{slug}` and `/blog/{x}`, which the router takes for
  one, whether they come from pages, documents or a plugin. On a static host the
  host's own precedence would pick one. Remove one of the two.
- **A redirect over a written file**, `collage.ErrRedirectShadowsFile` (since
  v0.52.0) — `/about` or `/about/` beside `about/index.html`, `/docs/{rest...}`
  beside `docs/intro/index.html`. A placeholder `From` such as `/{slug}` matches
  every file at its depth, `404.html` included. Narrow the pattern, or drop the page.

Both redirect checks run on every build, with or without a plugin to read the
redirects, and the files are still written.

**An error-level finding** fails the build as well, with `collage.ErrBuildFindings`
(since v0.21.0), but differently: the pages are written either way, and the
findings are in the report. See [Reading the report](#reading-the-report).

## Dynamic paths: `WithStaticParams`

A page at `/blog/{slug}` is one page with many URLs. The page lists them itself,
with `WithStaticParams` — one map of placeholder values per file:

```go
WithStaticParams(fn collage.StaticParamsFunc)

type StaticParamsFunc func(ctx context.Context, locale string) ([]map[string]string, error)
```

It is called once for each locale the page has a path in. These docs are one page,
`doc`, at `/docs/{slug}` in every language, and it lists every page of the
documentation in the original and every page translated so far in a translation:

```go
builder := collage.NewPage("doc").
	WithLayouts(layouts.Layout(app, docs)).
	WithContent(content).
	Static().
	WithStaticParams(func(_ context.Context, locale string) ([]map[string]string, error) {
		set, err := docs()
		if err != nil {
			return nil, err
		}
		loaded := set.Site(locale)
		if loaded == nil {
			return nil, nil
		}
		params := make([]map[string]string, 0, len(loaded.Pages()))
		for _, page := range loaded.Pages() {
			params = append(params, map[string]string{"slug": page.Slug})
		}
		return params, nil
	})
```

- **The build makes the path.** Each map fills the locale's pattern the way a
  link [built by name](/docs/links-and-locales#links-by-name) would, and the file
  is written under the locale's prefix: `dist/docs/caching/index.html`,
  `dist/tr/docs/caching/index.html`. A value that needs escaping in a URL is
  written at its decoded path — `héllo`, not `h%C3%A9llo` — which is where a static
  host looks a request for it up.
- **The values are what the data handlers see.** `rc.Param("slug")` has the value
  a live request to that path would have.
- **The map must fill the pattern exactly.** A name missing, or one the pattern does
  not have, fails that one file with `collage.ErrRouteParams`; the rest are built.
  So does a `/` in a single `{param}`'s value since v0.34.0 — a value that is a
  path belongs in a `{rest...}`. Before, a slug such as `../about`, from a CMS say,
  wrote its page over another one inside the output directory.
- **An error, or a panic, fails that page's locale** and is named in the report —
  a panic as `collage.ErrBuildPanic`. Returning no maps writes nothing, and is not
  an error.
- **Only a build calls it.** A running server answers every value the pattern
  matches, listed or not.
- **The page must still be cacheable.** A page with a data handler is dynamic
  unless it says otherwise, so one that is to be exported says `Static()` or
  `Incremental(ttl)` — as `doc` does above.

Documents take the same `WithStaticParams`: a feed at `/feeds/{category}/rss.xml`
lists its categories. See [Documents](/docs/documents#documents-in-a-static-export).

Every path `WithStaticParams` produces is checked before its own file is written: a path
that would resolve outside the output directory — `/../../etc` — is refused with
`collage.ErrPathEscapesOutDir`, and so is a write through a symlink that leads out
of it. The refusal fails that one path, not the build around it: the other pages
are still rendered and written, and the error is in the report.

## Pagination

Pagination is a library, not a plugin: [elagoht/paginate](/docs/plugins#libraries-that-are-not-plugins)
has nothing to register and nothing to configure. A data handler and a template
use it. Add it with `go get github.com/Elagoht/collage-paginate`; v0.1.0 needs
collage v0.57.0 or later.

`paginate.New(total, perPage, current)` returns a `Pager` with `Current`, `Last`,
`Offset`, `Limit`, `Prev`, `Next` and a `Window` of page numbers (`1 … 4 5 [6] 7 8
… 20`):

```go
p := paginate.New(total, perPage, current)
// p.Current, p.Last, p.Offset, p.Limit, p.Prev, p.Next, p.Window

rows := db.Query("… LIMIT ? OFFSET ?", p.Limit, p.Offset) // fetch only the page
page := paginate.Items(all, p)                             // or slice a list you already have
```

- **`current` is clamped to `[1, Last]`.** `New` never fails, so deciding that a
  page is out of range is the handler's job.
- **An empty listing is one empty page.** `Last` is 1 and `Items` returns an empty,
  non-nil slice.
- **The window** keeps the first and last page and two pages on each side of the
  current one. `p.WithWindow(edges, around)` changes it.
- **`perPage` below 1 panics.** It is a programming error, not input.

### Paging a listing a static export can write

A page has one path pattern per locale, so one page cannot serve both `/blog` and
`/blog/page/{n}`. A paged listing is two pages that share their fragment: one
serves page 1 at the listing's own path, the other serves the rest.

```go
func listing(ctx context.Context, rc *collage.RenderContext) (BlogList, []string, error) {
	total, err := posts.Count(ctx, rc.Locale) // the same count StaticParams uses
	if err != nil {
		return BlogList{}, nil, err
	}
	last := paginate.New(total, perPage, 1).Last
	n := 1
	if rc.Param("n") != "" {
		if n, err = paginate.FromPath(rc, "n", last); err != nil {
			return BlogList{}, nil, err // wraps collage.ErrNotFound: a 404
		}
	}
	p := paginate.New(total, perPage, n)
	links, err := paginate.PathNamed(rc, "blog", "blog-page")
	if err != nil {
		return BlogList{}, nil, err
	}
	items, err := posts.List(ctx, p.Limit, p.Offset)
	return BlogList{Posts: items, Pager: p, Links: links.For(p)}, []string{"posts"}, err
}

list := func(name string) *collage.Fragment {
	return collage.NewFragment(name, "blog/list.html").WithData(collage.DataHandler(listing)).Required().Build()
}
app.RegisterPage(collage.NewPage("blog").
	WithContent(list("blog-list")).
	WithPath("en", "/blog").
	Static(). // a page with a data handler is dynamic unless it says otherwise
	Build())
app.RegisterPage(collage.NewPage("blog-page").
	WithContent(list("blog-page-list")).
	WithPath("en", "/blog/page/{n}").
	WithStaticParams(paginate.StaticParams(posts.Count, perPage)).
	Static().
	Build())
```

- **Both pages are `Static()`** (or `Incremental`). A page with a data handler is
  dynamic by default, and a static export skips dynamic pages.
- **`StaticParams` lists pages 2…Last** for the export and for `PageURLs`, which
  the sitemap reads. Page 1 is not listed, because it is `/blog`. Give it the
  same count the handler uses: if the two disagree, the export writes pages the
  handler answers with a 404.
- **Each page has one spelling.** `FromPath` accepts only a page number spelled the
  canonical way, from 2 to `last`. `/blog/page/1`, `/blog/page/02`, a sign, anything
  other than ASCII digits and any number out of range are not found. The error
  wraps `collage.ErrNotFound`, so a Required fragment renders the page's not-found
  page with a 404.
- **`PathNamed` builds the links through the pages' names,** so locale prefixes
  come out right (`/tr/blog/page/2`). The request's other path parameters go into
  both addresses, so `/tag/{tag}` with `/tag/{tag}/page/{n}` links within its own
  tag. It checks both names up front, so a misspelled one is an error rather than
  an empty link. For a tag listing's export, write the static params yourself:
  one `{"tag": t, "n": k}` per tag and per page 2…`paginate.New(count(t), perPage,
  1).Last`. With no locales, `paginate.Path("/blog", "/blog/page/{n}")` does the
  same with fixed paths.

### Paging by query

A listing only a server can answer, such as a search, pages with the query string:

```go
n := paginate.FromQuery(rc, "page", last)                 // anything invalid is page 1
links := paginate.Query(rc, "page").For(paginate.New(total, perPage, n))
```

- **The request's other parameters stay.** Page 1 drops the key, so its address is
  the one without it.
- **A cached page must name the key** among its cache parameters:
  `WithCacheParams("q", "page")`. Without that, every page shares one cache entry.
- **A static export cannot serve these pages.** A static host answers
  `/search?page=2` with the `/search` file, and the export
  [warns](#what-is-warned-about) about pages that read query parameters.

### In the template

`Links` is a plain struct, so it goes into the fragment's data, and the template
checker sees `.Links.Next` and the rest:

```html
<nav class="pager" aria-label="Pages">
  {{with .Links.Prev}}<a href="{{.}}" rel="prev">Previous</a>{{end}}
  {{range .Links.Window}}
    {{if .Gap}}<span>…</span>
    {{else if .Current}}<span aria-current="page">{{.N}}</span>
    {{else}}<a href="{{.URL}}">{{.N}}</a>{{end}}
  {{end}}
  {{with .Links.Next}}<a href="{{.}}" rel="next">Next</a>{{end}}
</nav>
```

To put `rel="prev"` and `rel="next"` in the head, hoist them from the data
handler. The layout places them where it has `{{hoist "head"}}`:

```go
if l := links.For(p); l.Next != "" {
	rc.Hoist("head", "pager-next", template.HTML(`<link rel="next" href="`+template.HTMLEscapeString(l.Next)+`">`))
}
```

Page totals must be known: cursor-based paging and infinite scroll are out of
scope.

## Build options

| Field | Meaning |
| --- | --- |
| `OutDir` | Where to write. Required. |
| `Clean` | Remove `OutDir`'s contents (not the directory itself) first. |
| `Locales` | Build only these locales. Empty builds every locale a page declares. |
| `Concurrency` | How many pages render and write at once. `0` or `1` is one at a time. The report is in the same order either way. |
| `AllowDegraded` | Write pages whose render had a failed fragment. |

The builder refuses an `OutDir` that resolves to the root of the file system, and
refuses to `Clean` one that is the root of a repository
(`collage.ErrDangerousOutDir`) — `-out .` with `-clean` would otherwise delete your
project.

### Plugins in an export

The export renders in the state the server does. A build starts the application
first, running every plugin's `Init` before pages are enumerated (since v0.24.0), so
a page a plugin registers, or the `WithStaticParams` of data a plugin loads, is
built, and a plugin reads the same configuration; `OnBeforeRender`, `OnAfterRender` and
`OnDocumentRendered` fire for every page and document, so what a minifier or a
structured-data plugin does to a served page it does to the file. `OnPageResolved`
does not fire for those renders, because they are not requests. Since v0.21.0
`OnBuildFinished` runs once every file is written, for a plugin that checks the
build as a whole.

When such a plugin is registered, the build first
[captures every file's headers](#headers-and-redirects) through `App.Handler()`
(since v0.52.0). Those are requests: middleware, `OnRequest` and `OnPageResolved`
see them, and the page renders again with `OnBeforeRender` and `OnAfterRender` —
with `ev.Static` false. A plugin that counts or limits traffic, or writes files
from a render hook, skips them when `collage.IsCapture(ctx)` is true. See
[Using plugins](/docs/plugins).

## Reading the report

`collage.PrintBuildReport` prints what the build did. For the project `collage new`
scaffolds, it looks like this:

```sh
✓ 8 files written
    dist/404.html
    dist/index.html
    dist/static/app.3a3663df973f06fb.js
    dist/static/app.7050c2518057f5c5.css
    dist/static/app.css
    dist/static/app.js
    dist/static/favicon.936907e03c8c09ad.svg
    dist/static/favicon.svg

▲ 3 skipped
    hello  page uses the dynamic render strategy, which cannot be built statically
    features (en)  page carries {{csrfToken}}; a form needs a server to submit to, so it is served rather than exported
    health  document uses the dynamic render strategy, which cannot be built statically

8 written · 3 skipped · 0 failed · 2.5ms
```

Written files are summarised after ten, because a build that wrote three hundred
must not bury the one page it skipped. Skips, warnings and failures are never
truncated. The last line has every count and is coloured by the worst of them.
Skips are counted as routes, pages and documents together — `3 skipped`, not
`3 pages skipped`, since v0.10.0.

To act on the report in code, read `report.Skipped`, `report.Warnings`,
`report.Findings` and `report.Errors` yourself. Each skip is a
`collage.SkipRecord` with the route's name (`Page`), its `Locale`, a `Reason` for
people and, since v0.10.0, an `Err` to match with `errors.Is` — see
[Errors](/docs/errors#static-builds).
[Testing](/docs/testing#testing-the-export) turns that into a test.

Colour and the `✓ ▲ ◆ ✗` marks appear only on a terminal, and not when `NO_COLOR`
is set. In a CI log the marks are plain ASCII (`+ ! * x`).

### Findings

A plugin that checks the output — [elagoht/htmlcheck](/docs/plugins#elagohthtmlcheck)
is one — reports what it finds as findings, and since v0.21.0 the report lists them
under the page they are about, errors first:

```sh
◆ 2 findings
  /about
    error img-alt  <img src="/team.jpg"> has no alt  elagoht/htmlcheck
    warning heading-order  <h4> follows <h2>  elagoht/htmlcheck
```

A warning stops nothing. An error-level finding fails the build with
`collage.ErrBuildFindings`, so `main` exits non-zero and CI fails, but every page
is still written. The findings are in `report.Findings`, each a `collage.Finding`.
How a plugin reports one is in
[Writing a plugin](/docs/writing-plugins#checking-the-output-findings).

The header capture reports what it found here too, as build-wide warnings that
never fail the build: `unstable-header`, `capture-status`, `capture-failed`,
`capture-dev-mode` and `capture-personal`. See
[Headers and redirects](#headers-and-redirects).

## Looking at it: `collage serve`

Opening `dist/index.html` in a browser does not work: a `file://` page has no root,
so every absolute link and stylesheet is broken. `collage serve` serves the export
the way a static host does:

```sh
collage serve                  # http://localhost:4000
collage serve -dir public -port 8000
```

- `/about` is answered with `about/index.html`.
- A directory without an `index.html` is a 404 — no listings.
- An unknown path gets `404.html` with a 404 status.
- Nothing is cached, so exporting again and reloading shows the new output.

Its flags are `-dir` (default `dist`), `-host` (default `localhost`) and `-port`
(default `4000` — not 6060, so it can run next to `collage dev` while you compare
them). It serves files; it does not run your project.

## Hosting

The output is plain files with absolute links, so any static host serves it. Three
things to check on any of them:

- **The site must be at the root of its domain.** Links and asset URLs start at `/`,
  and collage has no base-path setting, so a site published under a sub-path — such
  as a GitHub project page at `user.github.io/project/` — has broken links. Use a
  custom domain, or a host that gives the site its own.
- **`404.html` is at the root.** Most hosts pick it up by that name without any
  configuration.
- **[`TrailingSlash`](/docs/configuration#trailingslash) is on.** A page is written
  as `<path>/index.html`, and a host serves it at `/about/` and redirects `/about`
  there. With the setting on, every link collage builds is already the address the
  host answers, rather than a redirect to it.

### Headers and redirects

A server sends headers with every page and answers redirects itself. A static host
sends what its own configuration files say: Netlify's `_headers` and `_redirects`,
Vercel's `vercel.json`. Since v0.52.0 a build carries what the server would have
said, for a plugin to write those files. Collage writes none of them itself.

**Headers are captured.** When a plugin implementing
[`BuildFinishedHook`](/docs/writing-plugins#buildfinishedhook) is registered, the
build asks the application's own handler for each written file's path, twice — in
process, with no network — and records the answer on the file as
`BuiltFile.Status` and `BuiltFile.Headers`. With no such plugin it asks nothing.
The requests carry the host of `Config.BaseURL`, and come over HTTPS when its
scheme is `https`, so a header sent only over HTTPS, such as
`Strict-Transport-Security`, is captured too. Whatever your middleware and plugins
set — `Cache-Control`, `Content-Security-Policy`, `X-Frame-Options` — comes along
without being declared again. Left out:

- headers about one response rather than the file: `Date`, `ETag`,
  `Last-Modified`, `Content-Length`, `Set-Cookie`, `Vary`, `Content-Encoding`,
  `Transfer-Encoding`, `Connection`, `Age` and `X-Collage-Render-Time`;
- a header whose value differs between the two answers, such as a CSP nonce: a
  file cannot carry a new one per reader;
- the `404.html` pages and the root redirect, which the build makes itself.
  `BuiltFile.Captured` is false for those and true for every file the build asked
  for, also one whose capture failed, which leaves `Status` 0.

What the capture found is reported as warnings in the build's findings:

| Rule | Means |
| --- | --- |
| `unstable-header` | A header was left out because it differs between the two answers — once per header name, with a count of paths. |
| `capture-status` | A file was answered with a status other than 2xx, or with two different statuses. |
| `capture-failed` | A path was not answered — each request has a deadline — or the capture did not run at all. |
| `capture-dev-mode` | The build ran in development mode, whose headers (`Cache-Control: no-store`) are not the ones to deploy. |
| `capture-personal` | Pages answered with `Cache-Control` `private` or `no-store` beside a header that differs between answers. That header is left out, so the file is no longer personal, and the `Cache-Control` only keeps a host from caching it. A `PersonaliseHook` cannot cause it, and a development build does not raise it. |

A `PersonaliseHook` that sets `Personal`, such as a CSP nonce from elagoht/secure, does
not make the capture answer `private, no-store`: the capture runs the hook but sets
`Personal` aside, so an exported nonce page gets the `Cache-Control` its strategy
gives (`public, max-age=…` for an incremental page). A reader's request is unchanged.
A development build does not raise `capture-personal`: its `no-store` is what
`capture-dev-mode` already warns about.

Middleware sees the capture requests. One that counts or limits traffic —
analytics, a rate limiter, a ban list — lets a request through untouched when
`collage.IsCapture(r.Context())` is true. Middleware that sets headers must not
skip it: what it sets is what the deployed file is served with.

**Redirects reach the build hook.** `BuildFinishedEvent.Redirects` holds every
redirect the site declares — every page's `WithRedirect` and
`WithPermanentRedirect`, every document's, and the rules of every plugin
implementing `collage.RedirectSource` — with its status and where it came from. Two
of them from one path, or one over a written file,
[fail the build](#what-fails).

**[elagoht/deploy](/docs/plugins#elagohtdeploy) writes the host's files.** Name the
host, and the plugin turns the captured headers and the redirects into the files
that host reads, warning about whatever the host cannot carry:

```json
{ "elagoht/deploy": { "target": "netlify" } }
```

| `target` | Writes |
| --- | --- |
| `netlify` | `_headers` and `_redirects` |
| `cloudflare` | `_headers` and `_redirects` |
| `vercel` | `vercel.json` |
| `github-pages` | a meta-refresh page per redirect and `.nojekyll`; no headers |

### GitHub Pages

This site is published by a workflow that runs the tests, exports and uploads
`dist/`:

```yaml
name: pages

on:
  push:
    branches: [main]

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - run: go test ./...
      - run: go run . -collage-build -out dist -clean
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

The export step runs the program directly rather than through `collage export`, so
the runner does not need the collage CLI installed. Set the repository's Pages
source to GitHub Actions, and give it a custom domain unless it is your
`user.github.io` repository.

GitHub Pages takes no custom headers. With elagoht/deploy's `github-pages` target
each literal redirect becomes a page that sends the browser on, and the warnings
say how many headers were lost.

### Cloudflare Pages

Export in CI, the same way, and upload the directory with Wrangler:

```sh
go run . -collage-build -out dist -clean
npx wrangler pages deploy dist --project-name mysite
```

Cloudflare Pages serves `404.html` for unknown paths when one is at the root, which
the export always writes when the site has a not-found page. elagoht/deploy's
`cloudflare` target writes its `_headers` and `_redirects`.

### Anything else

Netlify, S3 behind CloudFront, an nginx directory — each needs only the contents
of `dist/` and, if it does not do so already, `404.html` configured as the error
page. elagoht/deploy has a `netlify` target, and a `vercel` one whose `vercel.json`
is read only when the output directory is deployed as the project
(`vercel deploy dist`). When the site needs forms, previews or per-request pages,
it needs a server instead: see [Deployment](/docs/deployment).
