---
description: What a plugin can do, how to register and configure one, and the five published plugins.
reference: Plugin, LoadPluginConfig, ErrUnknownPluginConfig, ErrAppStarted
---

# Using plugins

A plugin is an ordinary Go value that you construct and hand to your application.
It can watch what the application does and change some of what it produces, but it
cannot reach into the router, the cache or the template set: it gets a narrow set
of capabilities and nothing more.

There is no plugin loader and no registry to publish to. A plugin is a Go module
you import, and it is compiled into your binary like any other dependency.

## What a plugin can do

Everything a plugin does goes through a hook it opts into or a capability it is
handed at startup. Between them, a plugin can:

- **rewrite what was rendered** — the HTML of a page, the body of a sitemap or a
  JSON document — before it is served and cached. A minifier works this way.
- **contribute to the page** before it renders, by hoisting into the head: a
  structured-data block, a meta tag, a preload hint.
- **add template functions**, which every template can then call.
- **transform every mounted filesystem**, so the files a mount serves are, for
  instance, already minified.
- **register pages, documents and mounts** of its own. An image optimiser serves
  the resized images it links to from its own mount.
- **serve a handler** of its own — an event stream, a WebSocket — and render a
  page's [fragment paths](/docs/forms-and-actions#a-fragment-at-its-own-url) to push
  over it (since v0.18.0).
- **adjust a cache write** — change its lifetime or tags, or skip it — and hear
  about invalidations.
- **observe failures**, with the stage of the pipeline they happened in.
- **add commands** that your program runs — `go run . <command>` in a scaffolded
  project; see [The collage CLI](/docs/cli#plugin-commands).

What each of these looks like from the plugin's side is in
[Writing a plugin](/docs/writing-plugins).

## Registering a plugin

A plugin is a separate module. Add it, construct it, and put it in
`Config.Plugins`:

```sh
go get github.com/Elagoht/collage-minimizer
```

```go
import minimizer "github.com/Elagoht/collage-minimizer"

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Plugins:  []collage.Plugin{minimizer.New()},
})
```

There is a second way, `app.RegisterPlugin`, called after `New`:

```go
if err := app.RegisterPlugin(jsonld.New()); err != nil {
	log.Fatal(err)
}
```

**Prefer `Config.Plugins`.** Some plugins have to act while the application is
being built — to add a template function, or to wrap the mounted filesystems —
and `New` is where that happens. Such a plugin implements an optional `Configure`
phase, and `RegisterPlugin` refuses it by name with
`collage.ErrConfigurerRegisteredLate` rather than accepting it and silently
skipping the part that mattered. `Config.Plugins` works for every plugin, so it is
the one to reach for.

`RegisterPlugin` also refuses:

| Error | When |
| --- | --- |
| `ErrAppStarted` | The application has already started — `Handler`, `ListenAndServe`, `Start`, `DispatchCommands` or a render ran. That includes any start that failed — in a plugin's `Init` since v0.11.0, and for any other reason since v0.12.0. |
| `ErrNilPlugin` | The plugin is `nil`. |
| `ErrEmptyPluginName` | Its `Name()` is empty. |
| `ErrDuplicatePlugin` | Another plugin already has the same name. |

A plugin in `Config.Plugins` is checked the same way, and `New` returns the error.

### Order matters

Plugins run in the order they were registered: `Config.Plugins` in slice order,
then any `RegisterPlugin` calls in call order. For hooks that change output, each
plugin sees what the one before it produced. A plugin that adds to the page
should usually come before one that compacts it, so the addition is compacted too.

## Configuring plugins

A plugin that takes settings reads them from `Config.PluginConfig`, a
`map[string]json.RawMessage` keyed by the plugin's name. Names read like module
paths — `elagoht/minimizer` — so the key and the plugin are one identifier.

The common case is a JSON file next to your program. A project scaffolded by
`collage new` has an empty `plugins-config.json` and already loads it:

```json
{
  "elagoht/minimizer": { "js": true },
  "elagoht/jsonld": { "siteName": "The Wire", "siteURL": "https://thewire.example" }
}
```

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	log.Fatal(err)
}

app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{minimizer.New(), jsonld.New()},
	PluginConfig: pluginConfig,
})
```

`LoadPluginConfig` returns a `nil` map, and no error, when the file does not
exist: a deployment that configures nothing should not need an empty file to say
so. A file that exists but cannot be read or is not a JSON object is an error.

Nothing about this is tied to JSON files. `LoadPluginConfig` is a convenience the
framework itself never calls; you can fill `PluginConfig` from YAML, the
environment, or Go constants:

```go
PluginConfig: map[string]json.RawMessage{
	"elagoht/minimizer": json.RawMessage(`{"js": true}`),
},
```

Three rules hold whichever way you fill it:

- **No section means defaults.** A plugin with no entry runs exactly as its
  constructor set it up.
- **A section is decoded over the defaults.** `{"js": true}` turns one setting on
  and leaves the rest as they were. How a particular plugin merges is in its
  README.
- **A key that names no registered plugin stops the application from starting**,
  with `collage.ErrUnknownPluginConfig`. A typo like `"elagoht/minimzer"` would
  otherwise leave the plugin on its defaults and you certain it was configured.
  The check runs when the application starts rather than in `New`, because
  `RegisterPlugin` can still add plugins after `New`.

A section that is present but does not decode — a string where the plugin expects
a boolean — is also an error, raised when the plugin reads it.

## The published plugins

Five plugins are published alongside the framework. Each is its own module, with
its own README that is the full reference; what follows is enough to set one up.

### elagoht/minimizer

[github.com/Elagoht/collage-minimizer](https://github.com/Elagoht/collage-minimizer)
strips whitespace and comments from rendered pages, from documents such as JSON
endpoints, and from the files your mounts serve.

```go
import minimizer "github.com/Elagoht/collage-minimizer"

Plugins: []collage.Plugin{minimizer.New()},
```

```json
{
  "elagoht/minimizer": { "html": true, "json": true, "css": true, "js": false }
}
```

- It must go in `Config.Plugins`: it wraps the mounted filesystems, which happens
  while the application is built.
- `New()` enables HTML, JSON and CSS. JavaScript is off by default; turn it on with
  `{"js": true}`. `minimizer.NewWith(minimizer.Config{...})` sets every switch
  yourself and bypasses those defaults.
- It is a scanner, not a parser, and removes only what cannot carry meaning:
  `<pre>`, `<textarea>`, `<script>` and `<style>` are kept verbatim, CSS strings are
  untouched, JavaScript keeps every newline, and invalid JSON is returned as it
  was.
- Mounted files are minified by wrapping the filesystem rather than the response,
  so `Range` requests against a mount keep returning the right bytes.

### elagoht/jsonld

[github.com/Elagoht/collage-jsonld](https://github.com/Elagoht/collage-jsonld)
emits schema.org structured data into the document head.

```go
import "github.com/Elagoht/collage-jsonld"

Plugins: []collage.Plugin{jsonld.New()},
```

```json
{
  "elagoht/jsonld": {
    "siteName": "The Wire",
    "siteURL": "https://thewire.example",
    "searchURL": "https://thewire.example/search?q={query}"
  }
}
```

Registering it emits a site-wide `WebSite` node on every page — once `siteName` is
set — and nothing else, because the plugin cannot know what a page is about. The
page says so from the data handler that fetched the article:

```go
func loadArticle(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	article, err := client.Article(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	jsonld.Emit(rc, jsonld.Article{
		Headline:      article.Title,
		DatePublished: article.PublishedAt,
		AuthorName:    article.Author,
	})
	return article, []string{"article:" + article.Slug}, nil
}
```

- It has no `Configure` phase, so `RegisterPlugin` accepts it too.
- `Emit` appends and works whether or not the plugin is registered. Nodes are keyed
  by schema.org type, so a nested fragment's `Article` replaces one declared
  further out, and nodes of different types all appear.
- Typed nodes cover `Article`, `BlogPosting`, `Blog`, `Person`, `WebSite` and
  `BreadcrumbList`; `jsonld.Raw` is the escape hatch for anything else, and refuses
  invalid JSON.
- A node that cannot be marshalled is skipped rather than failing the page.

**It needs `{{hoist "head"}}` in your layout** — see
[below](#plugins-that-write-to-the-head).

### elagoht/opti-image

[github.com/Elagoht/collage-opti-image](https://github.com/Elagoht/collage-opti-image)
rewrites every `<img>` that declares a pixel `width` and `height` to a resized
copy it serves itself, from its own mount.

```go
import optiimage "github.com/Elagoht/collage-opti-image"

Plugins: []collage.Plugin{optiimage.New()},
```

```json
{
  "elagoht/opti-image": {
    "allowedOrigins": [{ "scheme": "https", "host": "images.example.com" }],
    "webp": "auto"
  }
}
```

- It must go in `Config.Plugins`.
- **An empty `allowedOrigins` disables it.** It never fetches from a host you did
  not list, and the scheme is part of the origin.
- Only images with both `width` and `height` as pixel counts are rewritten; that
  declared size is the only honest target size there is.
- Nothing is fetched during the render. The page links a content-addressed name
  such as `/_image/8f2a91c0b4e7d3a6.webp`, and the image is fetched and resized the
  first time a browser asks for it.
- A static export writes the images into its output, because they are served from
  a mount and every mount is copied after the pages have rendered.
- `webp` is `false` (the default), `true`, or `"auto"`, which uses WebP only where
  the image would otherwise be lossless. Produced images are kept in memory and in
  `.cache/opti-image` by default (`cacheDir`); `p.Purge()` and
  `p.PurgeSource(url)` clear them.

`optiimage.NewWith(optiimage.Config{...})` sets a starting configuration in Go,
which the JSON section is then decoded over key by key.

### elagoht/live

[github.com/Elagoht/collage-live](https://github.com/Elagoht/collage-live) keeps
parts of a page current in the browser. It serves a small client script that
refreshes the fragments a page opened with
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) — on an
interval, or when the server pushes a change over an event stream.

```go
import live "github.com/Elagoht/collage-live"

Plugins: []collage.Plugin{live.New()},
```

```json
{
  "elagoht/live": { "prefix": "/_live/", "noStream": false, "keepAlive": "25s", "maxFragments": 32, "maxStreamAge": "0s" }
}
```

```html
<head>
  {{liveClient}}
</head>

<section data-collage-fragment="{{fragmentURL "home" "cpu"}}" data-collage-interval="2s">
  {{slot "cpu"}}
</section>

<section data-collage-fragment="{{fragmentURL "home" "disks"}}" data-collage-push>
  {{slot "disks"}}
</section>
```

- It must go in `Config.Plugins`: it adds `{{liveClient}}`, which the layout calls
  to include the client. v0.2.1 needs collage v0.20.0 or later; v0.2.0 needed
  v0.19.0, and v0.1.0 v0.18.0.
- The page owns the container and the fragment owns what is inside it.
  `data-collage-interval` fetches on an interval, `data-collage-push` takes the
  fragment from the stream, `data-collage-swap="morph"` patches the DOM in place,
  and `data-collage-target` on a form submits it with `fetch` and puts the answer
  into an element — on a success, or on a `422`, which is how an action says a
  submission did not validate: it answers the form's fragment again, with the
  errors, and status 422 (see
  [Forms and actions](/docs/forms-and-actions#refreshing-it-from-the-browser)). Any
  other failure leaves the target as it was and marks it stale. Since v0.2.1 a
  form whose action redirects navigates in one request, with collage's
  [`Collage-Fetch`](/docs/forms-and-actions#failure-renders-success-redirects).
- The client sends the `ETag` it holds and leaves the DOM alone on a `304`, adds
  what the fragment hoisted to the head once by its key, stops in a hidden tab, and
  backs off when a request fails, marking the element `data-collage-stale`. A
  pushed copy carries the same ETag a poll would get, so one the client already
  holds is not sent again.
- **One connection per browser.** A browser holds at most six connections to one
  origin over HTTP/1.1, across all its tabs, so since v0.2.0 the client opens the
  stream from a shared worker that every tab of the site shares. Where there is no
  shared worker, each tab opens its own and closes it while hidden. Since v0.2.1 a
  tab coming back from the back-forward cache receives pushes again; the worker
  forgot it when it left.
- While the stream is down, pushed elements are marked stale; after three failed
  connections they are polled every five seconds, and the stream is tried again
  every minute.
- **Pushing is tag-based.** When the application invalidates a tag, the plugin
  re-renders every open fragment that depended on it and sends it down the stream;
  invalidating is the whole API.
- **Sampled data** — CPU load, a queue's length — is pushed by invalidating on a
  timer, one tag per fragment. Mark such a fragment
  [`Shared()`](/docs/caching#a-page-that-declares-none), not `Static()`, so that it
  is rendered once per change rather than once per tab while its page stays
  dynamic; several fragments reading one measurement fetch it through
  `collage.Cached` with no tags, so that one fragment's tick does not re-render the
  others. The [README](https://github.com/Elagoht/collage-live#sampled-data-a-system-monitor)
  walks through a system monitor.
- A render the framework reports as `Shared` is made once per URL and sent to every
  reader of it; any other is made for each connection with its own request. A
  connection renders with the cookies it was opened with; with signed, stateless
  cookies, `maxStreamAge` closes it after that long, and the client reopens it at
  once with the cookies it holds then.
- The stream is served at `<prefix>stream/`. A compression middleware of your own
  should leave `text/event-stream` alone, or the stream arrives only when it ends.
  With `noStream` the client only polls.

### elagoht/websocket

[github.com/Elagoht/collage-websocket](https://github.com/Elagoht/collage-websocket)
carries collage-live's pushed fragments over a WebSocket instead of an event stream.

```go
import (
	live "github.com/Elagoht/collage-live"
	"github.com/Elagoht/collage-websocket"
)

lv := live.New()
Plugins: []collage.Plugin{lv, websocket.New(lv)},
```

- Register collage-live as well, before this plugin. v0.2.0 needs collage v0.19.0
  and collage-live v0.2.0 or later.
- Nothing else changes: the layout still includes `{{liveClient}}`, which now tells
  the client to connect here, and elements still say `data-collage-push`.
  collage-live stops serving its event stream. The WebSocket is opened from the
  same shared worker, so every tab still shares one connection, and `maxStreamAge`
  applies to it too.
- **Usually you do not need it.** Pushing fragments is one-way, which is what an
  event stream is for: it needs no dependency and reconnects by itself. This plugin
  is for deployments where streams are what breaks — a proxy that buffers them, a
  platform that limits them. It is the one piece of the live stack with a
  dependency, `github.com/coder/websocket`.
- A connection carries the reader's cookies, so by default only a page from the
  site itself may open one; `websocket.NewWith(lv, websocket.Options{...})` sets
  the `Path` (`/_live/ws/`), the `Ping` interval and the `OriginPatterns` that may
  connect as well.

## Plugins that write to the head

A plugin that contributes to the document head — structured data, meta tags,
preload hints — does it by hoisting, the same mechanism fragments use for their
titles and stylesheets. Hoisted content lands where the layout calls
`{{hoist "head"}}`, and nowhere else:

```html
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
```

**Without that marker, nothing appears.** The plugin registers, runs, declares its
block, and the block has nowhere to go. That is deliberate: a plugin that searched
the HTML for `</head>` and spliced itself in would be deciding a layout question
on your layout's behalf. A layout from `collage new` already has the marker; a
layout you wrote by hand may not. See [Head and SEO](/docs/head-and-seo) for
hoisting in general.

## Where plugins run

Plugins see more than the pages a server renders:

- **Cached pages are processed once.** A plugin's changes to the HTML are made
  before the page is cached, so a cache hit serves the processed bytes without
  running the plugin again. A plugin that must run on every request — to inject a
  per-visitor value — cannot be combined with a cached page.
- **Error pages** go through the same render hooks, so your 404 is minified and
  annotated like any other page.
- **A page an action answers with** — the validation re-render of a form — runs
  `OnAfterRender` too (since v0.10.0), so it is minified like the page a `GET`
  gets. See [Forms and actions](/docs/forms-and-actions#the-validation-re-render).
- **Documents** — sitemaps, feeds, JSON — go through their own hook, so a minifier
  covers them too. See [Documents](/docs/documents).
- **A static export runs the same render hooks** as a served render, with each
  plugin started and configured first, so the exported site is the site the server
  serves. See [Static export](/docs/static-export).

## Going further

[Writing a plugin](/docs/writing-plugins) covers the plugin contract, every hook
and what it may change, and a complete plugin with its tests.
