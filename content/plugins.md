---
description: What a plugin can do, how to register and configure one, and the thirty-four published plugins, grouped by what they are for.
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
- **wrap every request** with middleware of its own, after the application's
  (since v0.21.0).
- **check the output** and report findings — shown over the page in development,
  listed in a static build's report, failing the build at error level (since
  v0.21.0).
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
Two published plugins say where they go:
[elagoht/compress](#elagohtcompress) before any plugin that rewrites response
bodies, and [elagoht/devtoolbar](#elagohtdevtoolbar) last.

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

Thirty-four plugins are published alongside the framework, grouped below by what
they are for. Each is its own module, with its own README that is the full
reference; what follows is enough to set one up.

| Group | Plugins |
| --- | --- |
| [SEO and discovery](#seo-and-discovery) | jsonld, meta, sitemap, robots, feed, redirects, indexnow |
| [Content](#content) | markdown, highlight, toc, search, i18n |
| [Forms and state](#forms-and-state) | validate, honeypot, flash, session |
| [Security](#security) | secure, ratelimit, basicauth |
| [Live updates](#live-updates) | live, websocket |
| [Assets and delivery](#assets-and-delivery) | minimizer, opti-image, bundle, favicon, compress, cdnpurge, offline |
| [Operations and development](#operations-and-development) | htmlcheck, devtoolbar, accesslog, prometheus, otel, analytics |

### SEO and discovery

What a search engine, a feed reader or a link preview reads: structured data, the
tags a page is shared by, the sitemap and `robots.txt`, feeds, old addresses and
where they went, and telling search engines what changed.

#### elagoht/jsonld

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

- It needs collage v0.24.0 or later. It has no `Configure` phase, so
  `RegisterPlugin` accepts it too.
- `Emit` appends and works whether or not the plugin is registered. Nodes are keyed
  by schema.org type, so a nested fragment's `Article` replaces one declared
  further out, and nodes of different types all appear.
- Typed nodes cover `Article`, `BlogPosting`, `Blog`, `Person`, `WebSite` and
  `BreadcrumbList`; `jsonld.Raw` is the escape hatch for anything else, and refuses
  invalid JSON.
- A node that cannot be marshalled is skipped rather than failing the page.

**It needs `{{hoist "head"}}` in your layout** — see
[below](#plugins-that-write-to-the-head).

#### elagoht/meta

[github.com/Elagoht/collage-meta](https://github.com/Elagoht/collage-meta) writes
the tags a page is shared and indexed by into its head: Open Graph, Twitter cards,
the canonical URL, the meta description and the page's translations.

```go
import "github.com/Elagoht/collage-meta"

Plugins: []collage.Plugin{meta.New(meta.Options{
	SiteName:     "The blog",
	BaseURL:      "https://example.com",
	DefaultImage: "/static/share.png",
})},
```

```go
meta.Set(rc, meta.Page{
	Title:       article.Title,
	Description: article.Dek,
	Image:       article.Cover,
	Type:        meta.Article,
	Published:   article.PublishedAt,
})
```

```json
{
  "elagoht/meta": {
    "siteName": "The blog",
    "baseURL": "https://example.com",
    "defaultImage": "/static/share.png",
    "twitterSite": "@example",
    "locales": { "en": "en_US", "tr": "tr_TR" },
    "noAlternates": false
  }
}
```

- It needs collage v0.23.0 or later. `baseURL` is required; the application does
  not start without it.
- Registering it gives every page `og:site_name`, `og:type`, the canonical URL —
  built by name, so a query string a reader arrived with is never part of it —
  `og:locale`, an `hreflang` link for every locale the page has a path in, and the
  default image and Twitter card.
- What a page is about it says from its data handler with `meta.Set`. Each tag is
  hoisted under its own key, so a page's declaration replaces the default tag by
  tag, and a deeper fragment's replaces its parent's.
- The `<title>` is not among them: it is `rc.HoistTitle` or `WithTitle`, as without
  the plugin. It needs `{{hoist "head"}}` in the layout.

#### elagoht/sitemap

[github.com/Elagoht/collage-sitemap](https://github.com/Elagoht/collage-sitemap)
serves `/sitemap.xml` from the pages the application registered.

```go
import "github.com/Elagoht/collage-sitemap"

Plugins: []collage.Plugin{sitemap.New(sitemap.Options{
	BaseURL: "https://example.com",
})},
```

```json
{
  "elagoht/sitemap": {
    "baseURL": "https://example.com",
    "path": "/sitemap.xml",
    "exclude": ["thanks"],
    "maxURLs": 50000
  }
}
```

- It needs collage v0.21.0 or later. `baseURL` is required — a sitemap lists
  absolute URLs — and the application does not start without it.
- It lists every page with a path, in every locale, as `App.URL` spells it, with
  its other locales as `hreflang` alternates. A `{param}` pattern is listed once
  for each value its `WithStaticParams` returns, the URLs a static build writes;
  one without it is left out. `Exclude` leaves pages out by name.
- `LastMod`, a Go function, gives a page's `<lastmod>`.
- It is a static document: cached, exported, and made again when `sitemap.Tag` is
  invalidated — invalidate it with the tags of a post you publish.
- Past 50,000 URLs (`maxURLs`) it becomes a sitemap index of numbered files.

#### elagoht/robots

[github.com/Elagoht/collage-robots](https://github.com/Elagoht/collage-robots)
serves `/robots.txt`.

```go
import "github.com/Elagoht/collage-robots"

Plugins: []collage.Plugin{robots.New(robots.Options{
	Rules:    []robots.Rule{{Disallow: []string{"/admin"}}},
	Sitemaps: []string{"https://example.com/sitemap.xml"},
})},
```

```json
{
  "elagoht/robots": {
    "rules": [{ "userAgents": ["*"], "disallow": ["/admin"] }],
    "sitemaps": ["https://example.com/sitemap.xml"],
    "disallowAll": false
  }
}
```

- It needs collage v0.21.0 or later. With no rules it allows every crawler
  everything; a rule with no user agents is for `*`.
- `disallowAll` closes the site to every crawler, whatever the rules say, and
  sends every response with `X-Robots-Tag: noindex, nofollow`. Set it in the
  configuration of a staging deployment, so one binary is open in production and
  closed elsewhere.
- The body is fixed at startup, and a static build writes it to `robots.txt`.

#### elagoht/feed

[github.com/Elagoht/collage-feed](https://github.com/Elagoht/collage-feed) serves
RSS 2.0 and Atom 1.0 feeds from items the application lists, and announces them in
every page's head.

```go
import "github.com/Elagoht/collage-feed"

Plugins: []collage.Plugin{feed.New(feed.Feed{
	Title:   "The blog",
	BaseURL: "https://example.com",
	Link:    "/blog",
	Items:   latestPosts, // func(ctx) ([]feed.Item, error), newest first
	Tags:    []string{"posts"},
})},
```

- It needs collage v0.21.0 or later. It is configured in Go only, since `Items` is
  a function.
- A feed is served at `/feed.xml` as RSS and `/atom.xml` as Atom; `RSS` and `Atom`
  move them, and `"-"` leaves a format out. Several feeds each take a `Name` and
  their own paths. It carries at most `Limit` items, 20 by default.
- Every page gets a `<link rel="alternate">` for every feed, so it needs
  `{{hoist "head"}}` in the layout; `NoDiscovery` keeps a feed out of the heads.
- It is a static document: cached, exported, and made again when one of its
  `Tags` is invalidated.

#### elagoht/redirects

[github.com/Elagoht/collage-redirects](https://github.com/Elagoht/collage-redirects)
serves redirects kept in a file rather than in code — what a site migration leaves
behind — before routing, and writes them for a static host as `_redirects`.

```go
import "github.com/Elagoht/collage-redirects"

//go:embed redirects.txt
var siteFS embed.FS

Plugins: []collage.Plugin{redirects.New(redirects.Options{FS: siteFS})},
```

```
# the old blog
/blog/*          /posts/:splat
/about-us        /about          301
/summer-sale     /sale           302
/old-product     -               410
```

```json
{
  "elagoht/redirects": {
    "file": "redirects.txt",
    "rules": [{ "from": "/careers", "to": "https://jobs.example.com", "status": 302 }],
    "noRedirectsFile": false
  }
}
```

- It needs collage v0.24.0 or later.
- One rule a line: the old path, where it went, and a status — `301` when left out,
  or `302`, `307`, `308`, or `410` with `-` for a page that is gone, answered with
  the site's own not-found page and status `410`. `/blog/*` is a prefix, and
  `:splat` in the target is what the `*` matched. The first rule that matches wins,
  and the reader's query string is carried over.
- A file that is wrong stops the application from starting, and says where: a
  malformed line, a rule no request can reach, rules that send a reader round in a
  circle.
- A static build writes the rules to `_redirects`, the format Netlify and
  Cloudflare Pages read; `noRedirectsFile` leaves it out. Rules can also be given in
  Go or configuration.

#### elagoht/indexnow

[github.com/Elagoht/collage-indexnow](https://github.com/Elagoht/collage-indexnow)
tells search engines which URLs changed, through the IndexNow protocol — Bing,
Yandex, Seznam, Naver and the others that share what one of them is told.

```go
import "github.com/Elagoht/collage-indexnow"

Plugins: []collage.Plugin{indexnow.New(indexnow.Options{
	Key:     "4f1c2e9a7b3d4c8e",
	BaseURL: "https://example.com",
})},
```

```json
{
  "elagoht/indexnow": {
    "key": "4f1c2e9a7b3d4c8e",
    "baseURL": "https://example.com",
    "window": "10s",
    "exclude": ["/api/", "/search"]
  }
}
```

- It needs collage v0.23.0 or later. `key` and `baseURL` are required.
- What it sends is what the cache let go of: the
  [paths an invalidation dropped](/docs/caching#invalidating-by-path). A page that
  was not cached when its tag was invalidated is not sent.
- URLs are batched for `window`, sent in the background and retried on a `429` or
  `5xx`; what is waiting is flushed on shutdown.
- It serves the key at `/<key>.txt`, and a static build writes it there. A
  development server and a static build send nothing.

### Content

Where the words come from and how they are presented: Markdown files as page data,
coloured code, a table of contents, search on a static site, and translations.

#### elagoht/markdown

[github.com/Elagoht/collage-markdown](https://github.com/Elagoht/collage-markdown)
makes a directory of Markdown files page data: YAML front matter,
GitHub-flavoured Markdown with heading ids and footnotes, and one dependency tag
per file.

```go
import "github.com/Elagoht/collage-markdown"

//go:embed content
var content embed.FS

md := markdown.New(markdown.Options{FS: content, Dir: "content/blog"})

Plugins: []collage.Plugin{md},
```

```go
app.RegisterPage(collage.NewPage("post").
	WithContent(collage.NewFragment("post", "pages/post.html").
		WithDataHandler(md.Handler()).
		Required().
		Build()).
	WithPath("en", "/blog/{slug}").
	Static().
	WithStaticParams(md.StaticParams()).
	Build())
```

```json
{
  "elagoht/markdown": {
    "dir": "content/blog",
    "localeDirs": { "tr": "content/blog/tr" },
    "drafts": false,
    "unsafe": false
  }
}
```

- It needs collage v0.24.0 or later, and can go in `Config.Plugins` or
  `RegisterPlugin`: it reads its configuration and files when the application
  starts, which a static build does before it lists the pages to write.
- `md.Handler()` hands the template the `Doc` its `slug` names — `Title`,
  `Description`, `Date`, `Tags`, `HTML`, `Text`, `Headings` — or
  `collage.ErrNotFound`. `md.IndexHandler()`, `md.List` and `md.Get` feed an index
  page, a feed and a sitemap.
- An edit is one invalidation, `md.Tag(locale, slug)`; a file added or removed is
  `md.DirTag(locale)`. In development every request reads the file again.
- A locale can have its own directory, raw HTML is left out unless `unsafe` is on,
  and several sets — a blog and the documentation — are several plugins told apart
  by `Name`.

#### elagoht/highlight

[github.com/Elagoht/collage-highlight](https://github.com/Elagoht/collage-highlight)
colours code with chroma: the code blocks on every rendered page, a
`{{highlight}}` template function, and a light and dark stylesheet.

```go
import "github.com/Elagoht/collage-highlight"

Plugins: []collage.Plugin{highlight.New(highlight.Options{})},
```

```html
{{highlight .Snippet "go"}}
```

```json
{
  "elagoht/highlight": { "light": "github", "dark": "github-dark", "noBackground": true, "auto": true }
}
```

- It needs collage v0.23.0 or later, and must go in `Config.Plugins`: it adds
  `{{highlight}}`.
- With `auto` on, every `<pre><code class="language-go">` a page holds — what
  elagoht/markdown and most Markdown renderers write — is coloured where it
  stands, once per render; a cached page is served as it was coloured.
- The stylesheet is linked from pages with coloured code only, by a
  content-addressed name, each theme under its own `prefers-color-scheme` query. It
  needs `{{hoist "head"}}` in the layout.

#### elagoht/toc

[github.com/Elagoht/collage-toc](https://github.com/Elagoht/collage-toc) gives a
page's headings ids, and puts a table of contents and a reading time where the
templates ask for them.

```go
import "github.com/Elagoht/collage-toc"

Plugins: []collage.Plugin{toc.New(toc.Options{})},
```

```html
<aside>{{toc}}</aside>
<p>{{readingTime}}</p>
```

```json
{
  "elagoht/toc": {
    "minLevel": 2,
    "maxLevel": 4,
    "wpm": 225,
    "locales": { "tr": { "label": "İçindekiler", "readingTime": "{n} dk okuma" } }
  }
}
```

- It needs collage v0.23.0 or later, and must go in `Config.Plugins`: it adds
  template functions.
- Each function writes a placeholder, and the plugin fills it in once the page has
  rendered: the `h2` to `h4` inside `<main>`, as a nested list, and the words
  divided by `wpm`.
- A heading without an `id` gets one from its text, letters of any alphabet kept
  and lowered as the page's language lowers them.
- A fragment served on its own does not run the page hooks, so put `{{toc}}` in the
  page, not in a fragment refreshed on its own.

#### elagoht/search

[github.com/Elagoht/collage-search](https://github.com/Elagoht/collage-search) adds
search to a static site: a static build writes an index of every page, and a small
script searches it in the browser.

```go
import "github.com/Elagoht/collage-search"

Plugins: []collage.Plugin{search.New(search.Options{})},
```

```html
{{searchBox}}
```

```json
{
  "elagoht/search": {
    "maxText": 5000,
    "exclude": ["/admin/", "/tags/"],
    "locales": { "tr": { "label": "Ara", "noResults": "Sonuç yok" } }
  }
}
```

- It needs collage v0.23.0 or later, and must go in `Config.Plugins`: it adds
  `{{searchBox}}`.
- **The index is a static build's.** `collage export` writes `search-index.json`
  from each page's title, description, headings and text; a running server has
  none, and the box says so rather than breaking.
- The script is fetched with the page and loads the index the first time the box
  is focused. It matches every word, ignoring case and accents, ranks titles above
  headings above text, and keeps to the page's language.

#### elagoht/i18n

[github.com/Elagoht/collage-i18n](https://github.com/Elagoht/collage-i18n)
translates: a catalog per locale, `{{t}}` in templates in the locale the page is
rendered in, plurals, and missing translations reported as findings.

```go
import "github.com/Elagoht/collage-i18n"

//go:embed locales
var locales embed.FS

Plugins: []collage.Plugin{i18n.New(i18n.Options{FS: locales})},
```

```html
<a href="{{pageURL "home"}}">{{t "nav.home"}}</a>
<p>{{tn "cart" .Count}}</p>
```

```json
{ "elagoht/i18n": { "dir": "locales" } }
```

- It needs collage v0.22.0 or later, and must go in `Config.Plugins`: it adds
  template functions.
- One JSON file of nested keys per supported locale, `locales/<locale>.json`; the
  application does not start while a supported locale has none.
- `t` translates a key and fills `{name}` from name and value pairs, `tn` picks a
  plural form for a count, `th` allows markup from the catalog. A data handler
  calls `i18n.T(rc, key, pairs...)`.
- A missing key falls back to the default locale, then to the key, and is reported
  as `missing-translation` over the page in development and in a static build's
  report. In development the catalogs are read again on every request.

### Forms and state

What a form needs around an [action](/docs/forms-and-actions): validation, spam
protection, a message shown after the redirect, and a session in a cookie.

#### elagoht/validate

[github.com/Elagoht/collage-validate](https://github.com/Elagoht/collage-validate)
validates a form an action receives: chainable checks per field, a refused
submission rendered again with status 422, and template functions that put each
field's message and what the reader typed back in the form.

```go
import "github.com/Elagoht/collage-validate"

Plugins: []collage.Plugin{validate.New(validate.Options{})},
```

```go
v := validate.Form(rc)
v.Field("email").Required().Email()
v.Field("password").Required().MinLen(8)
if !v.Valid() {
	return validate.Refuse(rc, v, signupPage), nil
}
return collage.SeeOther("/welcome"), nil
```

```html
<input name="email" type="email" value="{{fieldValue "email"}}">
{{with fieldError "email"}}<p class="error">{{.}}</p>{{end}}
```

```json
{
  "elagoht/validate": {
    "messages": { "required": "Please fill this in." },
    "localeMessages": { "tr": { "required": "Bu alan zorunludur." } },
    "noRefill": ["password", "card"]
  }
}
```

- It needs collage v0.23.0 or later, and must go in `Config.Plugins`: it adds
  `{{fieldError}}`, `{{fieldValue}}` and `{{hasErrors}}`.
- It is the first half of collage's rule for a form — see
  [Forms and actions](/docs/forms-and-actions#the-validation-re-render). On a page
  nobody submitted the functions are empty, so one template serves both renders
  and the first stays cacheable.
- The checks are `Required`, `MinLen`, `MaxLen`, `Email`, `URL`, `Int`, `Range`,
  `OneOf`, `Matches`, `Equal` and `Custom`; `v.Fail` reports what only the
  application knows. A field keeps its first message.
- Messages are English by default, and are replaced per check, per locale or for
  every locale. A field whose name contains `password` is never typed back.

#### elagoht/honeypot

[github.com/Elagoht/collage-honeypot](https://github.com/Elagoht/collage-honeypot)
stops form spam without a CAPTCHA: a decoy field people never see and bots fill
in, and a signed timestamp that refuses a form sent back sooner than a person could
have filled it in.

```go
import "github.com/Elagoht/collage-honeypot"

Plugins: []collage.Plugin{honeypot.New(honeypot.Options{Key: key})},
```

```html
<form method="post" action="/contact">
  {{csrfToken}}
  {{honeypot}}
  <textarea name="message"></textarea>
  <button>Send</button>
</form>
```

```json
{
  "elagoht/honeypot": {
    "key": "hex-encoded, 32 bytes or more",
    "minDelay": 2,
    "maxAge": 86400,
    "silent": false,
    "protect": ["/"]
  }
}
```

- It needs collage v0.24.0 or later, and must go in `Config.Plugins`: it adds
  `{{honeypot}}`.
- Every form body posted to a protected path is checked before it reaches the
  action, so every such form must carry `{{honeypot}}`. A JSON body and every `GET`
  pass unchecked.
- The timestamp survives the page cache as collage's forgery token does: the
  cached page carries a placeholder, and the plugin's middleware signs the current
  time into it. Set a key of at least 32 random bytes, the same on every instance.
- A refusal is a `400`; with `silent` it is a `303` back to the form, as an accepted
  form answers. It stops careless bots, not a determined one — pair it with
  elagoht/ratelimit.

#### elagoht/flash

[github.com/Elagoht/collage-flash](https://github.com/Elagoht/collage-flash) adds
flash messages: a message an action sets before it redirects, shown once by the
page it redirects to.

```go
import "github.com/Elagoht/collage-flash"

Plugins: []collage.Plugin{flash.New(flash.Options{Key: key})},
```

```go
flash.Add(rc, flash.Success, "Your changes are saved.")
return collage.SeeOther("/settings"), nil
```

```html
{{range flashes}}
  <p class="flash flash--{{.Kind}}" role="status">{{.Text}}</p>
{{end}}
```

```json
{
  "elagoht/flash": {
    "key": "hex-encoded, 32 bytes or more",
    "cookie": "collage_flash",
    "maxAge": 300
  }
}
```

- It needs collage v0.22.0 or later, and must go in `Config.Plugins`: it adds
  `{{flashes}}`.
- The messages travel in a signed, `HttpOnly` cookie. Set a key of at least 32
  random bytes, the same on every instance; without one a key is made per process
  and a warning logged.
- A request carrying a message is rendered fresh, neither read from the page cache
  nor written to it, and marked `private, no-store`. Every other request is served
  as it would be without the plugin.

#### elagoht/session

[github.com/Elagoht/collage-session](https://github.com/Elagoht/collage-session)
keeps a session in a cookie: a small map of strings the site signs, and can
encrypt, with no database behind it.

```go
import "github.com/Elagoht/collage-session"

Plugins: []collage.Plugin{session.New(session.Options{Key: key})},
```

```go
s := session.Get(rc)
s.Regenerate()
if err := s.Set("user", user.ID); err != nil {
	return nil, err
}
return collage.SeeOther("/account"), nil
```

```json
{
  "elagoht/session": {
    "key": "hex-encoded, 32 bytes or more",
    "encrypt": true,
    "maxAge": 604800,
    "idleTimeout": 0,
    "sameSite": "lax"
  }
}
```

- It needs collage v0.23.0 or later. It adds no template function, so
  `RegisterPlugin` accepts it too.
- `Get`, `Set`, `Delete`, `Clear`, `Regenerate` and `ID`; a handler of your own
  reads it with `session.FromContext(r.Context())`.
- **A request carrying a valid session is rendered fresh**, neither read from the
  page cache nor written to it, and marked `private, no-store`. A reader without
  one is served from the cache as before. Set a session from an action, not from a
  cached page.
- The cookie is `HttpOnly` and `SameSite=Lax`, written only when the session
  changed. Keys rotate through `previousKeys`. A session cannot be revoked: it lives
  in the reader's cookie.

### Security

Headers a site should send, a limit on how fast one client can hit it, and a
password in front of a site that is not public yet.

#### elagoht/secure

[github.com/Elagoht/collage-secure](https://github.com/Elagoht/collage-secure)
sends the security headers a site should, and a Content-Security-Policy whose
nonces survive the page cache.

```go
import "github.com/Elagoht/collage-secure"

Plugins: []collage.Plugin{secure.New(secure.Options{
	CSP: "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
})},
```

```json
{
  "elagoht/secure": {
    "csp": "default-src 'self'; script-src 'self' 'nonce-{nonce}'",
    "cspReportOnly": false,
    "hsts": 63072000,
    "hstsSubdomains": true,
    "frameOptions": "DENY",
    "permissionsPolicy": "camera=(), microphone=(), geolocation=()"
  }
}
```

- It needs collage v0.22.0 or later, and must go in `Config.Plugins`: it adds
  `{{cspNonce}}`.
- By default it sends `X-Content-Type-Options`, `X-Frame-Options`,
  `Referrer-Policy`, `Cross-Origin-Opener-Policy` and, over TLS or behind a proxy
  sending `X-Forwarded-Proto: https`, `Strict-Transport-Security`.
  `Permissions-Policy` and the CSP are sent when set; `"-"` leaves a header out.
- `{nonce}` in the policy and `{{cspNonce}}` on an inline script are one nonce,
  new on every response: the cached page carries a placeholder, and the plugin's
  middleware puts a fresh nonce in its place. A page carrying one is sent with
  `Cache-Control: no-store` and no `ETag`.
- In development the policy is sent report-only, so collage's live-reload script
  keeps working.

#### elagoht/ratelimit

[github.com/Elagoht/collage-ratelimit](https://github.com/Elagoht/collage-ratelimit)
limits how fast one client can hit the site: a token bucket per client, per rule,
and a `429 Too Many Requests` with `Retry-After` once it is empty.

```go
import "github.com/Elagoht/collage-ratelimit"

Plugins: []collage.Plugin{ratelimit.New(ratelimit.Options{
	Rules: []ratelimit.Rule{
		{PathPrefix: "/login", Rate: 0.1, Burst: 5},
		{Rate: 0.5, Burst: 10},
	},
})},
```

```json
{
  "elagoht/ratelimit": {
    "rules": [{ "pathPrefix": "/login", "rate": 0.1, "burst": 5 }],
    "trustProxy": true,
    "trustedProxies": ["10.0.0.0/8"],
    "skip": ["/_collage/", "/healthz"]
  }
}
```

- It needs collage v0.24.0 or later. With no options, every form and action — every
  method but `GET`, `HEAD` and `OPTIONS` — is limited to a burst of ten, then one
  request every two seconds.
- The first rule a request matches counts it, so put narrow rules first; each rule
  has its own buckets. Responses carry `RateLimit-Limit`, `RateLimit-Remaining` and
  `RateLimit-Reset`.
- A client is its IP address, or its `/64` for IPv6. Behind a reverse proxy set
  `trustProxy`, and the address is read from `X-Forwarded-For`, believed only from
  a trusted proxy. `KeyFunc`, in Go, keys by something else.
- Buckets are kept in memory, so limits are per process.

#### elagoht/basicauth

[github.com/Elagoht/collage-basicauth](https://github.com/Elagoht/collage-basicauth)
puts HTTP Basic authentication in front of a site — a staging deployment, a
preview, a site not launched yet.

```go
import "github.com/Elagoht/collage-basicauth"

Plugins: []collage.Plugin{basicauth.New(basicauth.Options{
	Users: map[string]string{"team": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
})},
```

```json
{
  "elagoht/basicauth": {
    "realm": "Staging",
    "protect": ["/"],
    "skip": ["/healthz", "/_collage/"],
    "disabled": false
  }
}
```

- It needs collage v0.24.0 or later. The application does not start with no users.
- A password is written in plain text, as `sha256:` and its hex, or as a bcrypt
  hash. `COLLAGE_BASICAUTH_USERS` adds users from the environment, keeping secrets
  out of files.
- `disabled`, or `COLLAGE_BASICAUTH_DISABLED=true`, opens one deployment of the
  same binary — the production one.
- Every authenticated response has `public` replaced by `private` and carries
  `Vary: Authorization`, so a CDN in front does not hand a page to the next reader
  without asking. Serve it over HTTPS.

### Live updates

Keeping parts of a page current in the browser, without a framework on the client.

#### elagoht/live

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

#### elagoht/websocket

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

- Register collage-live as well, before this plugin. v0.2.1 needs collage v0.24.0
  and collage-live v0.2.1 or later.
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

### Assets and delivery

What the bytes a browser receives look like and how they get there: minified,
resized, bundled, compressed, purged from a CDN, and kept for offline reading.

#### elagoht/minimizer

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

- It needs collage v0.24.0 or later, and must go in `Config.Plugins`: it wraps the
  mounted filesystems, which happens while the application is built.
- `New()` enables HTML, JSON and CSS. JavaScript is off by default; turn it on with
  `{"js": true}`. `minimizer.NewWith(minimizer.Config{...})` sets every switch
  yourself and bypasses those defaults.
- It is a scanner, not a parser, and removes only what cannot carry meaning:
  `<pre>`, `<textarea>`, `<script>` and `<style>` are kept verbatim, CSS strings are
  untouched, JavaScript keeps every newline, and invalid JSON is returned as it
  was.
- Mounted files are minified by wrapping the filesystem rather than the response,
  so `Range` requests against a mount keep returning the right bytes.

#### elagoht/opti-image

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

- It needs collage v0.24.0 or later, and must go in `Config.Plugins`.
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

#### elagoht/bundle

[github.com/Elagoht/collage-bundle](https://github.com/Elagoht/collage-bundle)
bundles an application's JavaScript, TypeScript and CSS with esbuild when it
starts, serves the output under content-hashed names, and rebuilds on an edit in
development.

```go
import "github.com/Elagoht/collage-bundle"

Plugins: []collage.Plugin{bundle.New(bundle.Options{
	Dir:     "assets",
	Entries: []string{"app.ts", "app.css"},
})},
```

```html
<link rel="stylesheet" href="{{bundle "app.css"}}">
<script src="{{bundle "app.js"}}" defer></script>
```

```json
{
  "elagoht/bundle": {
    "dir": "assets",
    "entries": ["app.ts", "app.css"],
    "prefix": "/_bundle/",
    "target": "es2020"
  }
}
```

- It needs collage v0.23.0 or later, and must go in `Config.Plugins`: it adds
  `{{bundle}}`.
- The output is served from a mount at `/_bundle/`, each file under esbuild's hash
  of its content and cached for a year. A name the build did not produce fails the
  render, as `{{asset}}` does for a missing file.
- A production server whose bundle does not build does not start; in development
  the error is shown in the page and the server keeps running.
- The sources are read from disk when the application starts, not from an
  `fs.FS`. A static build copies the mount, so the built site holds the bundle.

#### elagoht/favicon

[github.com/Elagoht/collage-favicon](https://github.com/Elagoht/collage-favicon)
makes a site's icons from one source image — `favicon.ico`, the Apple touch icon,
the web app icons and the manifest that lists them — serves them at the root, and
links them from every page's head.

```go
import "github.com/Elagoht/collage-favicon"

Plugins: []collage.Plugin{favicon.New(favicon.Options{
	Source:     "assets/icon.png",
	SVG:        "assets/icon.svg",
	Name:       "The blog",
	ThemeColor: "#0f172a",
})},
```

```json
{
  "elagoht/favicon": {
    "source": "assets/icon.png",
    "svg": "assets/icon.svg",
    "name": "The blog",
    "themeColor": "#0f172a",
    "backgroundColor": "#ffffff"
  }
}
```

- It needs collage v0.23.0 or later. The source is a square PNG, JPEG or GIF,
  ideally 512 pixels or larger, read from disk or from `FS`.
- The icons are made once at startup and served as static documents, so a static
  build writes them. An SVG is passed through, never rasterised.
- The links carry `?v=` and a digest of the source, so a changed icon is a changed
  URL. It needs `{{hoist "head"}}` in the layout.

#### elagoht/compress

[github.com/Elagoht/collage-compress](https://github.com/Elagoht/collage-compress)
compresses responses with Brotli and gzip, and writes `.br` and `.gz` files beside
a static build.

```go
import "github.com/Elagoht/collage-compress"

Plugins: []collage.Plugin{
	compress.New(compress.Options{}),
	secure.New(secure.Options{CSP: "..."}),
},
```

```json
{
  "elagoht/compress": { "minSize": 512, "gzipLevel": 6, "brotliLevel": 4, "types": ["application/x-yaml"] }
}
```

- It needs collage v0.23.0 or later. **Register it before any plugin that rewrites
  response bodies**, elagoht/secure among them: the first plugin registered is the
  outermost middleware.
- Text types of at least `minSize` bytes are compressed with the best encoding the
  request accepts. `text/event-stream`, a WebSocket and a `Range` request are left
  alone.
- A compressed body is kept per ETag, so a page collage serves from its cache is
  compressed once per encoding, not once per reader. The ETag gains the encoding,
  and a conditional request still gets its `304`.
- A static build writes a `.br` and a `.gz` beside every compressible file, for a
  host that serves precompressed files; `noPrecompress` turns it off.

#### elagoht/cdnpurge

[github.com/Elagoht/collage-cdnpurge](https://github.com/Elagoht/collage-cdnpurge)
purges a CDN's copies of the pages collage invalidates: Cloudflare by zone, or any
service that takes a list of URLs through a webhook.

```go
import "github.com/Elagoht/collage-cdnpurge"

Plugins: []collage.Plugin{cdnpurge.New(cdnpurge.Options{
	BaseURL: "https://example.com",
	Cloudflare: &cdnpurge.Cloudflare{
		ZoneID:   os.Getenv("CF_ZONE_ID"),
		APIToken: os.Getenv("CF_API_TOKEN"),
	},
})},
```

```json
{
  "elagoht/cdnpurge": {
    "baseURL": "https://example.com",
    "webhook": { "url": "https://purge.example.com/", "token": "..." },
    "window": "2s",
    "retries": 4
  }
}
```

- It needs collage v0.23.0 or later. `baseURL` and at least one of `cloudflare` and
  `webhook` are required.
- It purges exactly the [paths an invalidation dropped](/docs/caching#invalidating-by-path),
  under `baseURL`. A page the origin had not cached is not named, so make sure what
  the CDN caches, collage caches too.
- Purges are batched for `window`, sent off the goroutine that invalidated, retried
  on a `429` or `5xx`, and flushed on shutdown. A development server purges nothing
  unless `force` is set.
- Keep the API token out of a file under version control.

#### elagoht/offline

[github.com/Elagoht/collage-offline](https://github.com/Elagoht/collage-offline)
serves a service worker, so the pages a visitor has read open again without a
network.

```go
import "github.com/Elagoht/collage-offline"

Plugins: []collage.Plugin{offline.New(offline.Options{
	Precache: []string{"/"},
	Fallback: "/offline",
})},
```

```html
<head>
  {{offlineScript}}
</head>
```

```json
{
  "elagoht/offline": {
    "precache": ["/", "/about"],
    "fallback": "/offline",
    "assets": ["/static/", "/_bundle/"],
    "maxPages": 100
  }
}
```

- It needs collage v0.24.0 or later, and must go in `Config.Plugins`: it adds
  `{{offlineScript}}`, which installs the worker served at `/sw.js`.
- Pages are fetched network-first and kept; static files under `assets` are served
  stale-while-revalidate; a page neither reachable nor kept gets the `fallback`
  page. A response marked `no-store` is never kept.
- The worker's caches are named after a version that changes with each deployment,
  so a new build replaces what the old one kept. The build is `version` when set,
  and otherwise collage's `Host.BuildID` — `Config.Cache.Version`, or a
  fingerprint of the executable.
- In development `/sw.js` unregisters itself. A static build writes `sw.js`; keep a
  CDN from holding it long.

### Operations and development

Checking what a site renders, seeing a render in development, and watching a
running site: access logs, metrics, traces and analytics.

#### elagoht/htmlcheck

[github.com/Elagoht/collage-htmlcheck](https://github.com/Elagoht/collage-htmlcheck)
checks the HTML a site renders — structure, accessibility, what a search engine
reads, what slows a page down, and the links between pages — and reports what it
finds as [findings](/docs/writing-plugins#checking-the-output-findings).

```go
import "github.com/Elagoht/collage-htmlcheck"

Plugins: []collage.Plugin{htmlcheck.New(htmlcheck.Options{})},
```

```json
{
  "elagoht/htmlcheck": {
    "rules": { "img-dimensions": "off", "heading-order": "error" },
    "titleMax": 60,
    "descriptionMax": 160,
    "pageBudget": 200000,
    "ignoreLinks": ["/api/"]
  }
}
```

- It needs collage v0.22.0 or later.
- In development each page is checked as it renders, and what is found is shown
  over the page. In a static build every page is checked, then the build as a
  whole — titles two pages share, links to pages the build did not write — and an
  error fails the build (see
  [Static export](/docs/static-export#findings)). On a production server nothing
  is checked.
- It has 22 rules — `html-lang`, `title`, `img-alt`, `input-label`,
  `duplicate-id`, `heading-order`, `broken-link` and more — each at `error` or
  `warn` by default. `rules` changes a level or turns one `off`; a rule name it
  does not know stops the application from starting. `htmlcheck.Rules()` lists
  them all.

#### elagoht/devtoolbar

[github.com/Elagoht/collage-devtoolbar](https://github.com/Elagoht/collage-devtoolbar)
shows a small panel at the bottom of every page in development: which page
rendered, in which locale, with what status, how long the render and each of its
fragments took, which fragments failed, the dependency tags the render depended
on, its `Cache-Control` and `ETag`, its size, and how many findings the checking
plugins reported.

```go
import "github.com/Elagoht/collage-devtoolbar"

Plugins: []collage.Plugin{
	htmlcheck.New(htmlcheck.Options{}),
	devtoolbar.New(), // last
},
```

- It needs collage v0.24.0 or later, and has nothing to configure.
- **Register it last**: the findings it counts are those of the plugins that ran
  before it.
- On a server without `DevMode` and in a static build it does nothing at all. The
  panel is added after the page cache, so what collage caches never carries it.

#### elagoht/accesslog

[github.com/Elagoht/collage-accesslog](https://github.com/Elagoht/collage-accesslog)
writes one structured log line per request through `slog`, and gives every request
an id its handlers can log with.

```go
import "github.com/Elagoht/collage-accesslog"

Plugins: []collage.Plugin{accesslog.New(accesslog.Options{})},
```

```json
{
  "elagoht/accesslog": {
    "skip": ["/_collage/", "/healthz", "/static/"],
    "sample": 0.25,
    "trustProxy": true,
    "requestIdHeader": "X-Request-ID"
  }
}
```

- It needs collage v0.24.0 or later.
- The line has the method, path without its query, status, bytes, duration, client
  address, user agent, referer and request id, through the application's logger or
  `Options.Logger`; a `5xx` is logged at `ERROR`.
- An `X-Request-ID` that looks like an id is kept, otherwise a new one is made; it
  is sent back and read with `accesslog.RequestID(ctx)`.
- `sample` logs a fraction of `2xx` responses, never of the rest. `trustProxy`
  reads the client's address from `X-Forwarded-For`, from trusted proxies only.

#### elagoht/prometheus

[github.com/Elagoht/collage-prometheus](https://github.com/Elagoht/collage-prometheus)
exports the framework's metrics to Prometheus — render and fragment timings, cache
events, HTTP responses by route, invalidations — and serves them at `/metrics`.

```go
import "github.com/Elagoht/collage-prometheus"

m := prometheus.NewMetrics(prometheus.Options{Namespace: "collage"})

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Metrics: m},
	Plugins:       []collage.Plugin{m},
})
```

```json
{
  "elagoht/prometheus": { "path": "/metrics", "token": "s3cret", "routes": ["/api/"] }
}
```

- It needs collage v0.24.0 or later. Hand the one value over as both the
  application's `Metrics` and a plugin: without the first nothing is measured,
  without the second nothing is served.
- No label is taken from a request. `route` is the name of the page whose pattern
  the path matched, `/blog/a` and `/blog/b` both `post`, so a crawler inventing
  paths cannot mint time series.
- Set `token` and the scrape must send it as a bearer token; `path: "-"` serves the
  metrics nowhere. The path is exact: one another route already answers, or one
  ending in `/`, stops the application from starting.

#### elagoht/otel

[github.com/Elagoht/collage-otel](https://github.com/Elagoht/collage-otel) turns
collage's own spans into OpenTelemetry spans, and continues the trace a request's
caller started.

```go
import otel "github.com/Elagoht/collage-otel"

t := otel.NewTracer(provider.Tracer("example.com/site"))

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Tracer: t},
	Plugins:       []collage.Plugin{t},
})
```

```json
{ "elagoht/otel": { "skip": ["/healthz"] } }
```

- It needs collage v0.23.0 or later. As the tracer it turns `collage.http`,
  `collage.render` and `collage.fragment` into spans; as a plugin it reads the
  caller's trace context from the headers. Each works alone.
- With both, a request is one server span, a child of the caller's, named for its
  route — `GET /blog/{slug}` — with the render and every fragment nested under it.
- The application owns the SDK: the provider, the exporter, the sampler and the
  propagator. Set a propagator, or every request starts a trace of its own.

#### elagoht/analytics

[github.com/Elagoht/collage-analytics](https://github.com/Elagoht/collage-analytics)
adds an analytics snippet to every page's head — Plausible, Umami, GoatCounter or
Google Analytics 4 — in production and in static builds, never in development.

```go
import "github.com/Elagoht/collage-analytics"

Plugins: []collage.Plugin{analytics.New(analytics.Options{
	Plausible: &analytics.Plausible{Domain: "example.com"},
})},
```

```json
{
  "elagoht/analytics": {
    "plausible": { "domain": "example.com" },
    "exclude": ["admin"],
    "respectDnt": true,
    "requireConsent": true
  }
}
```

- It needs collage v0.23.0 or later, and `{{hoist "head"}}` in the layout.
- Plausible, Umami and GoatCounter count visits without cookies. Google Analytics
  4 sets cookies; use it with `requireConsent`.
- With `respectDnt` or `requireConsent`, a small loader served from the site
  decides in the browser, before any provider's script is fetched: nothing loads
  under Do Not Track or Global Privacy Control, or until the page calls
  `window.collageAnalyticsConsent()`.
- `exclude` names pages that get no snippet.

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
hoisting in general. Among the published plugins, jsonld, meta, feed, favicon and
analytics write to the head this way.

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
