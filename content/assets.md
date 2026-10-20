---
description: Serving stylesheets, scripts, images and downloads from a directory or an embedded file system, with URLs that can be cached for a year.
reference: App.Mount, MountOption, WithCacheControl, WithoutBuildCopy, ErrUnknownAsset, Env.Static
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
system.

What collage does do is clean the path in the URL, with `path.Clean`, before
`Open` is ever called:

- A `..` or `.` that stays inside the mount is normalised and served —
  `/static/css/../app.css` is `/static/app.css`.
- One that climbs out of it, or a path that starts with an empty element
  (`/static//…`), is refused with a `404`.

That stops a traversal written in the URL, not one built out of symlinks on disk.

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
fixed when the binary was built, so editing a file would change nothing.
`Env.Static` (since v0.68.0) makes that choice, and a scaffolded project's
`newApp` mounts what it returns:

```go
return app, app.Mount("/static/", env.Static(files, "static"))
```

With `Env.Dev` on it opens the directory on disk with `os.OpenRoot`; otherwise it
is `fs.Sub(files, "static")`, the embedded copy. Before v0.68.0 the scaffold wrote
the same choice out in a `staticFiles` function of its own.

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

— a year, and a promise that it will never change.

The promise is honest because the name comes from the bytes: when the file changes,
the name changes, pages link the new one, and the old URL is simply never asked
for again. Browsers and CDNs keep the file for a year and never revalidate it, and
nobody is ever served a stylesheet that does not match the page.

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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 264" role="img" aria-labelledby="assets-mount-title">
<title id="assets-mount-title">Your code hands app.Mount a file system: an embed.FS compiled into the binary, or in development the static directory on disk opened with os.OpenRoot. The mount then answers three kinds of name under /static/. The hashed name app.hash.css, which the asset function links, is a 200 cached for a year and immutable. The plain name app.css is a 200 with the mount's own max-age=3600. Anything else, an old or wrong hash or a missing file, is a plain-text 404 with no-store. With DevMode on the hash is recomputed on every request and every 200 is no-store.</title>
<defs><marker id="assets-mount-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="6" y="8" width="150" height="40" rx="6" /><text x="81" y="24" text-anchor="middle">embed.FS</text><text class="d-small" x="81" y="39" text-anchor="middle">in the binary</text>
<rect class="d-frame" x="164" y="8" width="150" height="40" rx="6" /><text x="239" y="24" text-anchor="middle">os.OpenRoot</text><text class="d-small" x="239" y="39" text-anchor="middle">dev: static/ on disk</text>
<path class="d-flow" d="M81 48C81 62 130 60 130 72" marker-end="url(#assets-mount-arrow)" />
<path class="d-flow" d="M239 48C239 62 190 60 190 72" marker-end="url(#assets-mount-arrow)" />
<rect class="d-a" x="50" y="76" width="220" height="26" rx="5" /><text class="d-on-a" x="160" y="93" text-anchor="middle">app.Mount("/static/", fsys)</text>
<path class="d-flow" d="M100 102C100 114 56 110 56 122" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="160" y1="102" x2="160" y2="122" marker-end="url(#assets-mount-arrow)" />
<path class="d-flow" d="M220 102C220 114 264 110 264 122" marker-end="url(#assets-mount-arrow)" />
<rect class="d-frame" x="6" y="126" width="100" height="48" rx="6" /><text x="56" y="141" text-anchor="middle">hashed name</text><text class="d-small" x="56" y="155" text-anchor="middle">app.&lt;hash&gt;.css</text><text class="d-small" x="56" y="167" text-anchor="middle">what asset links</text>
<rect class="d-frame" x="110" y="126" width="100" height="48" rx="6" /><text x="160" y="141" text-anchor="middle">plain name</text><text class="d-small" x="160" y="155" text-anchor="middle">app.css</text><text class="d-small" x="160" y="167" text-anchor="middle">a fixed URL</text>
<rect class="d-frame" x="214" y="126" width="100" height="48" rx="6" /><text x="264" y="141" text-anchor="middle">anything else</text><text class="d-small" x="264" y="155" text-anchor="middle">an old hash,</text><text class="d-small" x="264" y="167" text-anchor="middle">a missing file</text>
<line class="d-flow" x1="56" y1="174" x2="56" y2="190" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="160" y1="174" x2="160" y2="190" marker-end="url(#assets-mount-arrow)" />
<line class="d-flow" x1="264" y1="174" x2="264" y2="190" marker-end="url(#assets-mount-arrow)" />
<rect class="d-b" x="6" y="194" width="100" height="40" rx="5" /><text class="d-on-b" x="56" y="210" text-anchor="middle">200 immutable</text><text class="d-on-b d-small" x="56" y="225" text-anchor="middle">max-age=31536000</text>
<rect class="d-soft-a" x="110" y="194" width="100" height="40" rx="5" /><text x="160" y="210" text-anchor="middle">200</text><text class="d-small" x="160" y="225" text-anchor="middle">max-age=3600</text>
<g class="d-dropped"><rect class="d-frame" x="214" y="194" width="100" height="40" rx="5" /><text x="264" y="210" text-anchor="middle">404 no-store</text><text class="d-small" x="264" y="225" text-anchor="middle">plain text</text></g>
<text class="d-small" x="160" y="254" text-anchor="middle">DevMode: rehashed per request, every 200 no-store</text>
</svg>
<figcaption>Which file system a mount serves is your code's choice: the embedded copy in production, the directory on disk in development. Under the prefix the mount answers three kinds of name. The hashed name that <code>{{asset}}</code> links is served for a year, <code>immutable</code>. The plain name is served with the mount's own <code>Cache-Control</code>, <code>public, max-age=3600</code> unless <code>WithCacheControl</code> says otherwise. A hash that is not the file's current one, or a file that is not there, is a plain-text 404, <code>no-store</code>. With <code>DevMode</code> on, the hash is recomputed on every request and nothing is cacheable.</figcaption>
</figure>

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
| `Content-Type` | From the file extension. A file whose name says nothing is sniffed from its first 512 bytes, but never into a type that runs: one that looks like HTML or XML is `text/plain` (since v0.34.2). Always with `X-Content-Type-Options: nosniff` |
| `ETag` | A strong hash of the file's content, computed on first request and remembered |
| `Range` | Supported, with `If-Range` and `206 Partial Content` |
| Directories | Never listed. A directory, or the bare prefix, is a 404 |
| `index.html` | Never served implicitly |
| Dotfiles | Never served, and never copied by an export: a path with an element starting with `.` — `.env`, `.git/`, an editor's swap file — is a 404. `.well-known` is the exception (since v0.34.0) |
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
[`Config.DevWatch`](/docs/configuration#devwatch).)

Outside development, a file changed on disk keeps its remembered hash until the
process restarts: mounts are for files that are deployed, not edited under a
running server.

## Static export

A [static export](/docs/static-export) copies every mount into its output under its
prefix — `/static/app.css` becomes `dist/static/app.css` — along with a
content-addressed copy of every file a page actually linked with `{{asset}}`. Only
those: a media directory is not doubled for names no page uses.
`collage.WithoutBuildCopy()` leaves a mount out.

## Example: a blog's stylesheet, script and images

The blog from [Templates](/docs/templates#example-a-blog-post) has three kinds of
file to serve:

| File | Linked from | Changes |
| --- | --- | --- |
| `static/app.css` | The layout, on every page | With a redesign |
| `static/app.js` | The layout, on every page | Rarely |
| `static/img/harbour.jpg` | One post's template | Never, once published |

### Mounting

The files are embedded, so the binary runs from any directory, and `collage dev`
serves the directory on disk instead. `env.Static` is the choice from
[Embedding](#embedding):

```go
//go:embed all:static
var staticFS embed.FS

func main() {
	collage.Main(newApp)
}

func newApp(env collage.Env) (*collage.App, error) {
	app, err := collage.New(&collage.Config{
		DevMode:  env.Dev,
		Template: collage.TemplateConfig{Root: "templates"},
	})
	if err != nil {
		return nil, err
	}

	// register pages

	return app, app.Mount("/static/", env.Static(staticFS, "static"))
}
```

### Linking

The layout links the stylesheet and the script, and the post links its image, each
by its plain path:

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<script src="{{asset "/static/app.js"}}" defer></script>
```

```html
<img src="{{asset "/static/img/harbour.jpg"}}" alt="The harbour at dawn">
```

What a reader's browser receives carries each file's hash:

```html
<link rel="stylesheet" href="/static/app.b988912693d9423d.css">
<script src="/static/app.dc18125931ff948b.js" defer></script>
<img src="/static/img/harbour.9e00fefaa8d22370.jpg" alt="The harbour at dawn">
```

The browser then asks for those names, and the mount answers:

| Request | Status | `Cache-Control` |
| --- | --- | --- |
| `/static/app.b988912693d9423d.css` | `200` | `public, max-age=31536000, immutable` |
| `/static/app.css` | `200` | `public, max-age=3600` |
| `/static/app.0000000000000000.css` | `404`, plain text | `no-store` |
| `POST /static/app.css` | `405`, `Allow: GET, HEAD` | `no-store` |

Under `collage dev` the same names are linked and served, every `200` with
`no-store`.

### Deploying a changed stylesheet

A redesign changes `static/app.css`, and a new binary is deployed with the new bytes
embedded. Then:

1. The first page the new process renders hashes the new file, and links
   `/static/app.0a044e6f01a45071.css`.
2. A reader's browser has never seen that name. It fetches the stylesheet once and
   keeps it for a year.
3. The old `app.b988912693d9423d.css` stays in that browser's cache, unused: no page
   links it any more.
4. Asked for the old name, the new process answers `404`, `no-store`. It serves
   neither the old bytes, which it no longer has, nor the new ones under a name
   that does not describe them.

Nothing had to be purged, and no reader was shown the new HTML with the old
stylesheet.

The plain name is where that guarantee stops. A reader who fetched
`/static/app.css` shortly before the deploy may be handed their old copy until an
hour has passed since they fetched it. That is why the pages link `{{asset}}`, and why the plain
name is for URLs outside your control.

### A path no mount serves

A typo in the post's template:

```html
<img src="{{asset "/static/img/hrabour.jpg"}}" alt="The harbour at dawn">
```

There is no such file, so `{{asset}}` fails the post fragment with
`collage.ErrUnknownAsset`. The log names the fragment, the template's line and the
path. What the reader gets is the fragment's
[failure policy](/docs/fragments-and-slots#when-a-fragment-fails):

- **The post is `Required()`**, so the page fails: a `500` and the error page, never
  an article with a broken image.
- **In an optional fragment**, such as a sidebar, the page is served without that
  fragment, answers `200`, goes out `no-store` and is not cached; the failure is
  logged as a warning.

Either way the mistake shows on the first render that reaches it, usually yours in
development, rather than as a `404` in a reader's network panel.

A request is a different matter. A URL under `/static/` that names no file is a
plain-text `404` from the mount. A URL under no mount's prefix never reaches a
mount: the router answers it, with the site's own not-found page.

## What mounts do not do

- **No compression.** No gzip or Brotli, and no pre-compressed sidecar files. Put a
  reverse proxy or CDN in front, wrap `app.Handler()`, or register
  [elagoht/compress](/docs/plugins#elagohtcompress).
- **No bundling, minification or image processing.** A mount serves the bytes it
  is given; [plugins](/docs/plugins) can do more:
  [elagoht/bundle](/docs/plugins#elagohtbundle) bundles scripts and stylesheets with
  esbuild, [elagoht/minimizer](/docs/plugins#elagohtminimizer) minifies what a mount
  serves, [elagoht/opti-image](/docs/plugins#elagohtopti-image) resizes images,
  [elagoht/favicon](/docs/plugins#elagohtfavicon) makes a site's icons from one
  image, and [elagoht/offline](/docs/plugins#elagohtoffline) keeps pages and static
  files for reading without a network.
- **Only `fs.FS` sources.** An object store such as S3 is not one. Serve it
  yourself with [`app.Handle`](/docs/middleware-and-apis), or link to it directly.
