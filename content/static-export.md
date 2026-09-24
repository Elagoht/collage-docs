---
description: Render the site to static files with collage export — what is written, what is skipped and why, dynamic paths, and publishing to a static host.
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
prints what happened. This is the collage-docs version of that function:

```go
func staticBuild(app *collage.App, outDir string, clean bool) error {
	loaded, err := site.Load(content.FS)
	if err != nil {
		return err
	}
	builder, err := collage.NewBuilder(app, collage.BuildOptions{
		OutDir:       outDir,
		Clean:        clean,
		PathProvider: docPaths{loaded},
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
could not be written. Skipped pages and warnings do not fail the build.

If you rewrite `main.go`, keep the `-collage-build`, `-out` and `-clean` flags, or
`collage export` stops doing anything useful.

## What is written

| What | Where |
| --- | --- |
| A `Static()` or `Incremental(ttl)` page | `dist/<path>/index.html`; `/` is `dist/index.html` |
| The same page in a non-default locale | Under the locale's prefix: `dist/tr/<path>/index.html` |
| A `Static()` or `Incremental(ttl)` document | Its literal path: `/sitemap.xml` is `dist/sitemap.xml` |
| The same document in a non-default locale | Under the locale's prefix: `dist/tr/sitemap.xml` |
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

- **`Dynamic()` pages and documents.** They exist to render per request.
- **Pages with a form.** A page whose render contains `{{csrfToken}}` is skipped:
  a form needs a server to post to, and a forgery token belongs to one reader. The
  page may be `Static()` on purpose — cached, and invalidated by the action it posts
  to — and it is served rather than exported. The scaffold's `/features` page is
  one.
- **A `{param}` pattern with no path provider.** `/blog/{slug}` cannot be written
  until something says which slugs exist. See below.
- **Two documents on one file.** A `DocumentPathProvider` that returns one path
  twice resolves two tasks to one file; the first is written and the rest are
  skipped with `collage.ErrDuplicateOutputPath`. One pattern in two locales is not
  this — each locale is written under its own prefix. See
  [Documents](/docs/documents#documents-in-a-static-export).

Error pages registered without a path are not listed at all: they are not URLs.

Not exported, and not reported, because they are not pages: [actions](/docs/forms-and-actions),
handlers mounted with `app.Handle`, and middleware. An export renders without a
request, so no middleware runs and `collage.Vary` is never called — each page is
written in the version a request with no preferences would get.

## What is warned about

A page that declared, with `WithCacheParams`, which query parameters it reads
renders differently for each of them. A file has no query string: a static host
answers `/blog?page=2` with the `/blog` file. The page is written without a query,
and the report says so, rather than letting a paginated archive look like it works.
A document with `WithCacheParams` — a paginated feed — is warned about the same way
(since v0.10.0; before, only pages were).

If pagination has to work in an export, put the page number in the path —
`/blog/page/{n}` — and list the pages with a path provider.

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
  that differ only in a trailing slash, or a path provider that returns a path
  twice. Nothing is rendered when this is found.

## Dynamic paths: `PathProvider`

A page at `/blog/{slug}` is one page with many URLs. The builder asks a
`collage.PathProvider` for them:

```go
type PathProvider interface {
	Paths(ctx context.Context, page *collage.Page, locale string) ([]collage.PathInstance, error)
}
```

It is called once per dynamic page and locale, and returns the concrete paths and
the parameters each one captures. These docs are one page, `doc`, at
`/docs/{slug}`; the provider lists every page of the documentation:

```go
// docPaths tells the static build which /docs/{slug} pages exist: every page of
// the documentation, and nothing else.
type docPaths struct{ site *site.Site }

func (d docPaths) Paths(_ context.Context, page *collage.Page, _ string) ([]collage.PathInstance, error) {
	if page.Name != "doc" {
		return nil, nil
	}
	var paths []collage.PathInstance
	for _, p := range d.site.Pages() {
		paths = append(paths, collage.PathInstance{Path: p.URL(), Params: map[string]string{"slug": p.Slug}})
	}
	return paths, nil
}
```

- **Check the page.** One provider answers for every dynamic page. Returning `nil`
  for a page you do not know writes nothing for it; that is not an error.
- **`Params` is what the data handlers see.** It is laid over what the router
  captured from `Path`, so `rc.Param("slug")` has the value a live request would have.
- **`Path` is without the locale prefix.** Return the pattern's path, `/blog/hello`;
  the builder writes a non-default locale under its own directory.
- **Returning an error** records it against that page and locale and carries on.

Documents with a `{param}` have their own interface, `BuildOptions.DocumentPathProvider`,
described in [Documents](/docs/documents).

Every path a provider returns is checked before anything is written: a path that
would resolve outside the output directory — `/../../etc` — is refused with
`collage.ErrPathEscapesOutDir`, and so is a write through a symlink that leads out
of it.

## Build options

| Field | Meaning |
| --- | --- |
| `OutDir` | Where to write. Required. |
| `Clean` | Remove `OutDir`'s contents (not the directory itself) first. |
| `Locales` | Build only these locales. Empty builds every locale a page declares. |
| `Concurrency` | How many pages render and write at once. `0` or `1` is one at a time. The report is in the same order either way. |
| `PathProvider` | Concrete paths for pages with a `{param}`. |
| `DocumentPathProvider` | Concrete paths for documents with a `{param}`. |
| `AllowDegraded` | Write pages whose render had a failed fragment. |

The builder refuses an `OutDir` that resolves to the root of the file system, and
refuses to `Clean` one that is the root of a repository
(`collage.ErrDangerousOutDir`) — `-out .` with `-clean` would otherwise delete your
project.

### Plugins in an export

The export renders in the state the server does. Every plugin's `Init` runs first,
so a plugin reads the same configuration; `OnBeforeRender`, `OnAfterRender` and
`OnDocumentRendered` fire for every page and document, so what a minifier or a
structured-data plugin does to a served page it does to the file. `OnPageResolved`
does not fire, because an export is not a request. See [Using plugins](/docs/plugins).

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

To act on the report in code, read `report.Skipped`, `report.Warnings` and
`report.Errors` yourself. Each skip is a `collage.SkipRecord` with the route's name
(`Page`), its `Locale`, a `Reason` for people and, since v0.10.0, an `Err` to match
with `errors.Is` — see [Errors](/docs/errors#static-builds).
[Testing](/docs/testing#testing-the-export) turns that into a test.

Colour and the `✓ ▲ ✗` marks appear only on a terminal, and not when `NO_COLOR` is
set. In a CI log the marks are plain ASCII (`+ ! x`).

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
(default `4000` — not 3000, so it can run next to `collage dev` while you compare
them). It serves files; it does not run your project.

## Hosting

The output is plain files with absolute links, so any static host serves it. Two
things to check on any of them:

- **The site must be at the root of its domain.** Links and asset URLs start at `/`,
  and collage has no base-path setting, so a site published under a sub-path — such
  as a GitHub project page at `user.github.io/project/` — has broken links. Use a
  custom domain, or a host that gives the site its own.
- **`404.html` is at the root.** Most hosts pick it up by that name without any
  configuration.

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

### Cloudflare Pages

Export in CI, the same way, and upload the directory with Wrangler:

```sh
go run . -collage-build -out dist -clean
npx wrangler pages deploy dist --project-name mysite
```

Cloudflare Pages serves `404.html` for unknown paths when one is at the root, which
the export always writes when the site has a not-found page.

### Anything else

Netlify, S3 behind CloudFront, an nginx directory — each needs only the contents
of `dist/` and, if it does not do so already, `404.html` configured as the error
page. When the site needs forms, previews or per-request pages, it needs a server
instead: see [Deployment](/docs/deployment).
