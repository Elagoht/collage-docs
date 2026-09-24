---
description: Serving stylesheets, scripts, images and downloads from a directory or an embedded file system, with URLs that can be cached for a year.
---

# Static assets

Stylesheets, scripts, images, fonts and downloads are served by an **asset
mount**: a file system served under a URL prefix, as files.

```go
root, err := os.OpenRoot("static")
if err != nil {
	log.Fatal(err)
}
if err := app.Mount("/static/", root.FS()); err != nil {
	log.Fatal(err)
}
```

`app.Mount` takes any `fs.FS`, which covers both sources that matter: a directory
on disk and an `embed.FS` compiled into the binary. `static/app.css` is now
`/static/app.css`.

Mounts are not pages. They never enter the [page cache](/docs/caching), are not
tagged and are not reached by `InvalidateTags`. A 50 MB video in a cache sized for
pages would push out thousands of them; so files are served as files.

## Use os.OpenRoot, not os.DirFS

`os.DirFS` looks like the obvious way to serve a directory, and it is the wrong
one for the internet. Its own documentation says it does not prevent symlink
traversal: a symlink inside the directory that points outside it is followed, and
whatever it points at is served.

`os.OpenRoot` is enforced by the operating system. Every path it opens is resolved
inside the directory it holds, and a symlink that leaves the directory fails to
open at all. It has been in Go since 1.24.

collage cannot close this gap for you. It is handed an `fs.FS` and calls `Open` on
it; whether the file system stays inside its directory is a property of the file
system. What collage does do is refuse `..`, absolute paths and empty path
elements in the URL before `Open` is ever called — which stops a traversal written
in the URL, not one built out of symlinks on disk.

Keep the `*os.Root` open for the life of the process: its file system serves every
request.

### Embedding

An `embed.FS` has no symlinks and is safe by construction. It also lets the binary
run from any working directory. Use `fs.Sub` to strip the directory name, or the
stylesheet answers at `/static/static/app.css`:

```go
//go:embed all:static
var staticFS embed.FS

func mountAssets(app *collage.App) error {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	return app.Mount("/static/", static)
}
```

In development you want the directory on disk instead — the embedded copy was
fixed when the binary was built, so editing a file would change nothing. A
scaffolded project does this:

```go
func staticFiles(devMode bool) (fs.FS, error) {
	if devMode {
		if root, err := os.OpenRoot("static"); err == nil {
			return root.FS(), nil
		}
	}
	return fs.Sub(staticFS, "static")
}
```

collage makes this choice for templates itself, because `Template.Root` tells it
where they are on disk. A mount is handed only an `fs.FS`, so where its files came
from is something only your code knows.

## Content-addressed URLs

Link a mounted file with `asset`, not by writing its path:

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<script src="{{asset "/static/app.js"}}" defer></script>
```

What goes into the page is the file's name with a hash of its content in it:

```html
<link rel="stylesheet" href="/static/app.0d5f2b53aebf6c72.css">
```

and that URL is served with

```
Cache-Control: public, max-age=31536000, immutable
```

— a year, and a promise that it will never change. The promise is honest because
the name comes from the bytes: when the file changes, the name changes, pages link
the new one, and the old URL is simply never asked for again. Browsers and CDNs
keep the file for a year and never revalidate it, and nobody is ever served a
stylesheet that does not match the page.

A plain name cannot be cached that way at any lifetime. Sooner or later it hands
someone yesterday's stylesheet with today's HTML.

Three details:

- **A file that does not exist is an error, not a URL.** `{{asset
  "/static/typo.css"}}` fails the fragment with `collage.ErrUnknownAsset`. The
  alternative is a page that renders "successfully" while linking a stylesheet
  that 404s.
- **The hash in a request is checked.** A URL whose hash is not the file's
  current one is a 404, not the file. Serving it would pin a wrong answer in front
  of everyone behind a shared cache, for a year.
- **Plain names still work.** `/static/app.css` is served too, with the mount's
  own, shorter `Cache-Control`. Use it where a URL must stay the same — a favicon
  referenced from outside the site, a file linked from an email.

### From Go

A data handler that needs the URL — for an Open Graph image, a preload hint —
asks the render context:

```go
cover, err := rc.Asset("/static/covers/" + post.Slug + ".jpg")
if err != nil {
	return postView{}, nil, err
}
rc.HoistProperty("og:image", siteOrigin+cover)
```

`rc.Asset` returns exactly what `{{asset}}` renders, and `collage.ErrUnknownAsset`
for a path no mount serves. It works in a [document's](/docs/documents#the-handler)
handler too, since v0.10.0 — an icon URL in a web manifest. For stylesheets,
`rc.HoistStylesheet` and `{{stylesheet}}` do the lookup and the `<link>` in one
step — see [Head and SEO](/docs/head-and-seo).

## Options

```go
err := app.Mount("/media/", media.FS(),
	collage.WithCacheControl("public, max-age=86400"),
	collage.WithoutBuildCopy(),
)
```

| Option | Effect |
| --- | --- |
| `collage.WithCacheControl(value)` | The `Cache-Control` for files requested by their plain name. Default `public, max-age=3600` |
| `collage.WithoutBuildCopy()` | [Static export](/docs/static-export) does not copy this mount into its output |

`WithCacheControl` is about plain names only; content-addressed URLs are always
served with a year and `immutable`, because that is what their names earn. How
long a plain name may be kept depends on how often the file changes under it, which
only you know — so it is yours to set.

`WithoutBuildCopy` is for a mount that production serves from somewhere else, such
as a CDN, or one too large to duplicate into every build.

## What a file response contains

A mount is a thin layer over `fs.Open` and Go's `http.ServeContent`, not
`http.FileServer`:

| | |
| --- | --- |
| Methods | `GET` and `HEAD`. Anything else is a `405` with `Allow: GET, HEAD` |
| `Content-Type` | From the file extension, falling back to sniffing the first 512 bytes |
| `ETag` | A strong hash of the file's content, computed on first request and remembered |
| `Range` | Supported, with `If-Range` and `206 Partial Content` |
| Directories | Never listed. A directory, or the bare prefix, is a 404 |
| `index.html` | Never served implicitly |
| A missing file | A plain-text `404`, `no-store` — never your HTML not-found page |

**`Range` requests** are what let a browser seek in an audio or video file, or
resume a download, without fetching it from the start. They come from
`http.ServeContent`, and they are the reason files are a separate mechanism from
[documents](/docs/documents), whose bodies are produced whole in memory.

**The ETag is a content hash** because the most common source of assets in a Go
program, `embed.FS`, reports a zero modification time for every file — so
`Last-Modified` and size-and-date validation do not work for it at all. Only the
hash is remembered, never the file, so the memory it costs grows with the number of
files, not their size.

A missing stylesheet gets a plain-text 404 rather than your site's not-found page
on purpose: an HTML error page handed to a browser that asked for CSS is a
mistake in either direction.

## Prefixes and routes

A prefix must begin and end with `/`, and cannot be `/` alone — a mount at the root
would swallow every page (`collage.ErrInvalidPrefix`). A nil file system is
`collage.ErrNilFS`.

A request under a mount's prefix is answered by the mount and never reaches the
router. That is safe because startup refuses a prefix that would hide a URL a page,
document or redirect already answers — including locale-prefixed ones, so a mount
at `/tr/` fails on a site with Turkish pages (`collage.ErrMountShadowsRoute`). Two
mounts whose prefixes overlap are refused too (`collage.ErrMountConflict`). Both
checks run when registration closes, so the order you call `Mount` and
`RegisterPage` in does not matter.

The check compares prefixes with registered patterns, so it cannot see a catch-all
above the prefix: a page at `/{rest...}` would have matched `/static/…`, and the
mount takes those URLs. That is the trade of claiming a prefix.

## In development

A content-addressed URL and a file you are editing are in conflict. Outside
development the hash is computed once and remembered, and the URL is promised for a
year; edit the file under a running process and the page would go on linking the
old name, which the browser was told can never change.

So with `DevMode` on, every mount recomputes a file's hash on each request and
serves everything `no-store`. An edited file gets a new URL, the page links it, and
the browser fetches it — and the development reload script refreshes the page when
a mounted file changes. (A directory that is not mounted, such as Markdown your
handlers read, reloads the page once it is named in
[`Config.DevWatch`](/docs/configuration#devwatch).) Outside development, a file
changed on disk keeps its remembered hash until the process restarts: mounts are
for files that are deployed, not edited under a running server.

## Static export

A [static export](/docs/static-export) copies every mount into its output under its
prefix — `/static/app.css` becomes `dist/static/app.css` — along with a
content-addressed copy of every file a page actually linked with `{{asset}}`. Only
those: a media directory is not doubled for names no page uses.
`collage.WithoutBuildCopy()` leaves a mount out.

## What mounts do not do

- **No compression.** No gzip or Brotli, and no pre-compressed sidecar files. Put a
  reverse proxy or CDN in front, or wrap `app.Handler()`.
- **No bundling, minification or image processing.** A mount serves the bytes it
  is given; [plugins](/docs/plugins) can do more.
- **Only `fs.FS` sources.** An object store such as S3 is not one. Serve it
  yourself with [`app.Handle`](/docs/middleware-and-apis), or link to it directly.
