---
description: What a plugin can do, how to register and configure one, and the forty-five published plugins, grouped by what they are for.
reference: Plugin, LoadPluginConfig, Env.PluginConfig, ErrUnknownPluginConfig, ErrAppStarted
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
- **hear about invalidations** — the tags whose cached pages were just
  invalidated, to purge a CDN's copies or push a fresh fragment.
- **observe failures**, with the stage of the pipeline they happened in.
- **wrap every request** with middleware of its own (since v0.21.0). A plugin in
  `Config.Plugins` is outside everything `app.Use` adds, in list order, so your
  middleware reads what it put in the request's context; one registered with
  `RegisterPlugin` is inside the middleware added before it. Since v0.67.0 a
  handler the application mounts with plain `app.Handle` is the exception: no
  plugin's middleware runs for it. See
  [Middleware and your own API](/docs/middleware-and-apis#plugins-middleware-and-yours).
- **check the output** and report findings — shown over the page in development,
  listed in a static build's report, failing the build at error level (since
  v0.21.0).
- **add commands** that your program runs — `go run . <command>` in a scaffolded
  project; see [The collage CLI](/docs/cli#plugin-commands).

Here is where those hooks sit on the way a page request travels, and the two that
fire outside a request:

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 558" role="img" aria-labelledby="plugin-hooks-title">
<title id="plugin-hooks-title">A request passes plugin middleware, then app.Use, then the router. A page the cache holds goes straight to OnPersonalise. Otherwise OnBeforeRender runs, the page renders, OnAfterRender may replace its HTML, the result is cached by tag, and OnPersonalise rewrites it for each reader. Outside a request, InvalidateTags fires OnCacheInvalidate, and a finished static build fires OnBuildFinished.</title>
<defs><marker id="plugin-hooks-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="plugin-hooks-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="26" rx="13" /><text x="160" y="25" text-anchor="middle">request</text>
<line class="d-flow" x1="160" y1="34" x2="160" y2="50" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-frame" x="50" y="54" width="220" height="28" rx="6" /><text x="160" y="72" text-anchor="middle">plugin middleware</text>
<rect class="d-frame" x="50" y="88" width="220" height="28" rx="6" /><text x="160" y="106" text-anchor="middle">app.Use</text>
<line class="d-flow" x1="160" y1="116" x2="160" y2="132" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-cmd" x="110" y="136" width="100" height="26" rx="13" /><text class="d-cmd-text" x="160" y="153" text-anchor="middle">router</text>
<line class="d-flow" x1="160" y1="162" x2="160" y2="178" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-chip-hit" x="100" y="182" width="120" height="24" rx="12" /><text class="d-hit-text d-small" x="160" y="198" text-anchor="middle">page cache lookup</text>
<path class="d-hitline" d="M220 194H296V410H274" marker-end="url(#plugin-hooks-hit)" />
<text class="d-hit-text d-small" x="258" y="189" text-anchor="middle">hit</text>
<line class="d-flow" x1="160" y1="206" x2="160" y2="222" marker-end="url(#plugin-hooks-arrow)" />
<text class="d-small" x="166" y="218">miss</text>
<rect class="d-a" x="50" y="226" width="220" height="28" rx="5" /><text class="d-on-a" x="160" y="244" text-anchor="middle">OnBeforeRender</text>
<line class="d-flow" x1="160" y1="254" x2="160" y2="264" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-soft-a" x="50" y="268" width="220" height="28" rx="6" /><text x="160" y="286" text-anchor="middle">render: fragments fetch</text>
<line class="d-flow" x1="160" y1="296" x2="160" y2="306" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-a" x="50" y="310" width="220" height="28" rx="5" /><text class="d-on-a" x="160" y="328" text-anchor="middle">OnAfterRender · ev.HTML</text>
<line class="d-flow" x1="160" y1="338" x2="160" y2="348" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-chip-hit" x="90" y="352" width="140" height="24" rx="12" /><text class="d-hit-text d-small" x="160" y="368" text-anchor="middle">cache write · tags</text>
<line class="d-flow" x1="160" y1="376" x2="160" y2="392" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-b" x="50" y="396" width="220" height="28" rx="5" /><text class="d-on-b" x="160" y="414" text-anchor="middle">OnPersonalise · each reader</text>
<line class="d-flow" x1="160" y1="424" x2="160" y2="440" marker-end="url(#plugin-hooks-arrow)" />
<rect class="d-pill" x="110" y="444" width="100" height="26" rx="13" /><text x="160" y="461" text-anchor="middle">response</text>
<line class="d-rule" x1="6" y1="484" x2="314" y2="484" />
<text class="d-small" x="160" y="500" text-anchor="middle">on their own</text>
<rect class="d-c" x="6" y="510" width="150" height="40" rx="5" /><text class="d-on-c d-small" x="81" y="526" text-anchor="middle">Invalidate/PurgeTags</text><text class="d-on-c d-small" x="81" y="540" text-anchor="middle">→ OnCacheInvalidate</text>
<rect class="d-frame" x="164" y="510" width="150" height="40" rx="5" /><text class="d-small" x="239" y="526" text-anchor="middle">static build written</text><text class="d-small" x="239" y="540" text-anchor="middle">→ OnBuildFinished</text>
</svg>
<figcaption>Middleware from the plugins in Config.Plugins wraps the request outside the application's own. A page the cache holds skips the render hooks and goes straight to OnPersonalise, so what OnAfterRender changed is what was cached. A fresh render runs OnBeforeRender, then OnAfterRender, which may replace the HTML before it is cached by tag. A page that is never cached has no lookup and no write: it goes from the router to OnBeforeRender. An action that answers with its page runs OnBeforeRender and OnAfterRender too, uncached. Invalidating or purging tags, wherever it is called from, and finishing a static build fire hooks of their own.</figcaption>
</figure>

What each of these looks like from the plugin's side is in
[Writing a plugin](/docs/writing-plugins). Writing one that `collage plugin add`
can install is on [the repository contract](/docs/plugin-authors).

## Registering a plugin

*Since v0.58.0.* Add a plugin with one command, from the project's directory:

```sh
collage plugin add elagoht/minimizer
```

`elagoht/minimizer` means `github.com/Elagoht/collage-minimizer`. The command
fetches the module, adds the import and the constructor call to `plugins.go`, in
the place the plugin asks for, and adds an empty section `"elagoht/minimizer": {}`
to `plugins-config.json`: the defaults. It adds first any plugin this one needs.
Then it prints what the plugin still needs from you (since v0.68.0): its setup
note, the template functions it adds and the address of its documentation — with
several plugins, together at the end, under "What each plugin needs from you".

A failure leaves every file as it was. Taking one out is the same in reverse, and
the settings its section held are printed so they are not lost:

```sh
collage plugin del elagoht/minimizer
```

Several go in one command, each on its own: one that fails is put back, and the
others stay. A bundle, written with `@`, is a set of them under one name, and
`collage new` takes the same with `--with` (since v0.59.0):

```sh
collage plugin add elagoht/session elagoht/flash elagoht/validate
collage plugin add @prod
collage new mysite --with @blog,@prod
```

The bundles (`@prod`, `@site`, `@blog`, `@forms`, `@live`, `@ops`) and what each
holds are in [the CLI reference](/docs/cli#bundles).

A project made by `collage new` has `plugins.go` already, and `main.go` passes it
to the application:

```go
func plugins() []collage.Plugin {
	return []collage.Plugin{
		minimizer.New(minimizer.Options{}),
	}
}
```

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Plugins:  plugins(),
})
```

The file is yours to edit: reorder it, or give a plugin its options in Go
(`sitemap.New(sitemap.Options{...})`); `add` and `del` only insert and remove
whole entries. An older project gets the file from its first `add`, and the
command prints the `Plugins: plugins(),` line to put in `main.go`.

A plugin by anyone else is added the same way, by its module path:
`collage plugin add gitlab.com/you/stamp`. What a repository must hold to be
added is on [the plugin authors' page](/docs/plugin-authors), and the command is
in [the CLI reference](/docs/cli#collage-plugin).

### Wiring one by hand

The command writes nothing you cannot write yourself, and that stays the way to
do it for a project that builds its plugin list some other way. A plugin is a
separate module. Add it, construct it, and put it in `Config.Plugins`:

```sh
go get github.com/Elagoht/collage-minimizer
```

```go
import minimizer "github.com/Elagoht/collage-minimizer"

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Plugins:  []collage.Plugin{minimizer.New(minimizer.Options{})},
})
```

Add its section to `plugins-config.json` yourself, if it takes settings.

There is a second way, `app.RegisterPlugin`, called after `New`:

```go
if err := app.RegisterPlugin(jsonld.New(jsonld.Options{})); err != nil {
	log.Fatal(err)
}
```

**Prefer `Config.Plugins`.** Some plugins have to act while the application is
being built — to add a template function, or to wrap the mounted filesystems —
and `New` is where that happens. Such a plugin implements an optional `Configure`
phase, and `RegisterPlugin` refuses it by name with
`collage.ErrConfigurerRegisteredLate` rather than accepting it and silently
skipping the part that mattered.

`Config.Plugins` works for every plugin, so it is
the one to reach for.

`RegisterPlugin` also refuses:

| Error | When |
| --- | --- |
| `ErrAppStarted` | The application has already started — `Handler`, `ListenAndServe`, `Start`, `DispatchCommands` or a render ran. That includes any start that failed — in a plugin's `Init` since v0.11.0, and for any other reason since v0.12.0. |
| `ErrDuplicatePlugin` | Another plugin already has the same name. |

A `nil` plugin, or one whose `Name()` is empty, is refused too, with an error that
says so; since v0.60.0 neither has a sentinel of its own. A plugin in
`Config.Plugins` is checked the same way, and `New` returns the error.

### Order matters

Plugins run in the order they were registered: `Config.Plugins` in slice order,
then any `RegisterPlugin` calls in call order. For hooks that change output, each
plugin sees what the one before it produced. A plugin that adds to the page
should usually come before one that compacts it, so the addition is compacted too.

Four published plugins say where they go:
[elagoht/compress](#elagohtcompress) before any plugin that rewrites response
bodies, [elagoht/health](#elagohthealth) before any plugin that can refuse or answer
a request, [elagoht/jobs](#elagohtjobs) after the plugins its work uses, and
[elagoht/devtoolbar](#elagohtdevtoolbar) last.

## Configuring plugins

A plugin that takes settings reads them from `Config.PluginConfig`, a
`map[string]json.RawMessage` keyed by the plugin's name. Names read like module
paths — `elagoht/minimizer` — so the key and the plugin are one identifier.

The common case is a JSON file next to your program. A project scaffolded by
`collage new` has an empty `plugins-config.json` and already loads it, with
`env.PluginConfig(files)` (since v0.68.0): the file in the working directory when
there is one, otherwise the copy the binary embeds, so a binary deployed without
the file keeps its settings. Outside a scaffolded project, `LoadPluginConfig` reads
the file alone:

```json
{
  "elagoht/minimizer": { "js": true },
  "elagoht/jsonld": { "siteName": "The Wire" }
}
```

```go
pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
if err != nil {
	log.Fatal(err)
}

app, err := collage.New(&collage.Config{
	Template:     collage.TemplateConfig{Root: "templates"},
	Plugins:      []collage.Plugin{minimizer.New(minimizer.Options{}), jsonld.New(jsonld.Options{})},
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

Five rules hold whichever way you fill it:

- **No section means defaults.** A plugin with no entry runs exactly as its
  constructor set it up.
- **A section is decoded over the defaults.** `{"js": true}` turns one setting on
  and leaves the rest as they were. A list in the section replaces the default
  list whole, rather than being merged into it element by element (since
  v0.61.2), and a map adds its entries to the default's. How a particular plugin
  merges is in its README.
- **A key the plugin does not know stops the application from starting** (since
  v0.61.0): `collage: elagoht/robots: unknown key "dissallow" in
  plugins-config.json`. A misspelt key, or one a plugin has renamed, used to be
  ignored and the option left at its default. Keys still match
  case-insensitively. Remove the key the error names, or rename it to the one the
  plugin's README documents.
- **A duration is a string**, `"10s"` or `"1h30m"`, as `time.ParseDuration`
  reads it. A bare number is refused with an error naming its key —
  `"banTime": write a duration as a string such as "10s", not 3600` — where it
  would otherwise have been read as nanoseconds.
- **A key that names no registered plugin stops the application from starting**,
  with `collage.ErrUnknownPluginConfig`. A typo like `"elagoht/minimzer"` would
  otherwise leave the plugin on its defaults and you certain it was configured.
  The check runs when the application starts rather than in `New`, because
  `RegisterPlugin` can still add plugins after `New`.

A section that is present but does not decode — a string where the plugin expects
a boolean — is also an error, raised when the plugin reads it, and the message
names the key whose value failed.

## The published plugins

Forty-five plugins are published alongside the framework, grouped below by what
they are for. Each is its own module, with its own README that is the full
reference; what follows is enough to set one up.

| Group | Plugins |
| --- | --- |
| [SEO and discovery](#seo-and-discovery) | jsonld, meta, ogimage, sitemap, robots, feed, redirects, indexnow |
| [Content](#content) | markdown, highlight, toc, search, i18n |
| [Forms and state](#forms-and-state) | validate, honeypot, flash, session, uploads |
| [Security](#security) | secure, ratelimit, basicauth, oauth, fail2ban, consent |
| [Live updates](#live-updates) | live, websocket, client |
| [Assets and delivery](#assets-and-delivery) | minimizer, opti-image, bundle, favicon, compress, cdnpurge, offline, deploy |
| [Operations and development](#operations-and-development) | htmlcheck, devtoolbar, accesslog, prometheus, otel, analytics, tenant, errortrack, health, jobs |

[Libraries that are not plugins](#libraries-that-are-not-plugins) are listed after the groups.

Most sites start from what they are, not from a group. Three common starting
points, each plugin linked to its section below:

| You are building | Start with | What they do for it |
| --- | --- | --- |
| A blog | [markdown](#elagohtmarkdown), [highlight](#elagohthighlight), [toc](#elagohttoc), [feed](#elagohtfeed), [sitemap](#elagohtsitemap), [robots](#elagohtrobots), [meta](#elagohtmeta), [jsonld](#elagohtjsonld), [ogimage](#elagohtogimage), [search](#elagohtsearch) | Posts written in Markdown with highlighted code and a table of contents, a feed for readers, what search engines and link previews read, and search on a statically built site. `@blog` adds toc, feed, ogimage and the search-engine ones in one command; add markdown, highlight and search yourself. |
| A product catalogue | [jsonld](#elagohtjsonld), [meta](#elagohtmeta), [sitemap](#elagohtsitemap), [opti-image](#elagohtopti-image), [i18n](#elagohti18n), [cdnpurge](#elagohtcdnpurge), [compress](#elagohtcompress), [minimizer](#elagohtminimizer) | `Product` structured data, resized product images served from the plugin's own mount, translations per locale, and a CDN's copies purged when a product's tags are invalidated. |
| A multi-user web app with roles | [session](#elagohtsession), [oauth](#elagohtoauth), [validate](#elagohtvalidate), [flash](#elagohtflash), [honeypot](#elagohthoneypot), [uploads](#elagohtuploads), [secure](#elagohtsecure), [ratelimit](#elagohtratelimit), [fail2ban](#elagohtfail2ban), [live](#elagohtlive) | Sign-in through an identity provider, the signed-in user kept in a session, forms checked and answered, uploads checked by their content, abuse limited or banned, and parts of a page kept current. The roles are yours: a middleware added with `app.Use` reads the session the plugin put in the request's context, as [Plugins' middleware and yours](/docs/middleware-and-apis#plugins-middleware-and-yours) shows. |

`@prod` adds what any of them wants in production; the bundles are in
[the CLI reference](/docs/cli#bundles).

### SEO and discovery

What a search engine, a feed reader or a link preview reads: structured data, the
tags a page is shared by, the sitemap and `robots.txt`, feeds, old addresses and
where they went, and telling search engines what changed.

#### elagoht/jsonld

[github.com/Elagoht/collage-jsonld](https://github.com/Elagoht/collage-jsonld)
emits schema.org structured data into the document head.

```go
import "github.com/Elagoht/collage-jsonld"

Plugins: []collage.Plugin{jsonld.New(jsonld.Options{})},
```

```json
{
  "elagoht/jsonld": {
    "siteName": "The Wire",
    "searchURL": "https://thewire.example/search?q={query}"
  }
}
```

Registering it emits a site-wide `WebSite` node on every page — once `siteName` is
set, with the application's origin (`collage.BaseURL`) as its `url` — and nothing else, because the plugin cannot know what a page is about. The
page says so from the data handler that fetched the article:

```go
func loadArticle(ctx context.Context, rc *collage.RenderContext) (*Article, []string, error) {
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

- v0.3.0 needs collage v0.61.0 or later (v0.2.10 was the last for v0.60.0). It has no `Configure` phase, so
  `RegisterPlugin` accepts it too. v0.3.0 removed `siteURL`: set the application's
  `BaseURL` instead.
- `Emit` appends and works whether or not the plugin is registered. Nodes are keyed
  by schema.org type and their `Key` field, so a nested fragment's `Article`
  replaces one declared further out, and nodes of different types all appear.
- Several nodes of one type need a `Key` each (v0.4.0), never written to the page:
  `jsonld.Emit(rc, jsonld.Person{Key: p.Slug, Name: p.Name})` in a loop keeps
  every person. Without a `Key`, a later `Person` replaces the earlier one. `ID` is
  separate: it is written as `"@id"`, so set it only to a real IRI.
- Typed nodes cover every feature in Google Search's structured data gallery
  (v0.6.0): pages (`WebSite`, `WebPage` and its kinds, `ProfilePage`,
  `BreadcrumbList`, `ItemList`), writing (`Article`, `NewsArticle`,
  `BlogPosting`, `Blog`, `DiscussionForumPosting`, `QAPage`, `FAQPage`), people
  and organizations (`Person`, `Organization`, `LocalBusiness`,
  `EmployerAggregateRating`), `Event`, commerce (`Product` with shipping and
  returns, `SoftwareApplication`, `VacationRental`, `Review`,
  `AggregateRating`), media (`VideoObject`, `ImageObject`, `Movie`, `Recipe`,
  `HowTo`) and learning and data (`Course`, `Quiz`, `MathSolver`, `Dataset`,
  `JobPosting`). Shared parts — `Agent`, `Offer`, `Money`, `Rating`,
  `PostalAddress`, `GeoCoordinates` — mean the same in every node. `jsonld.Raw`
  is the escape hatch for anything else, and refuses invalid JSON.
- An `Event`'s location is a `Place`, an `OnlineURL`, or both for a mixed event.
  Its times are written with their own offset: read a wall-clock time with
  `time.ParseInLocation` in the event's zone, not `time.Parse`, which means UTC.
- A node that cannot be marshalled is skipped rather than failing the page, with a
  warning in the log naming its type.

**It needs `{{hoist "head"}}` in your layout** — see
[below](#plugins-that-write-to-the-head).

#### elagoht/meta

[github.com/Elagoht/collage-meta](https://github.com/Elagoht/collage-meta) writes
the tags a page is shared and indexed by into its head: Open Graph, Twitter cards,
the canonical URL, the meta description and the page's translations.

```go
import "github.com/Elagoht/collage-meta"

BaseURL: "https://example.com",
Plugins: []collage.Plugin{meta.New(meta.Options{
	SiteName:     "The blog",
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
    "defaultImage": "/static/share.png",
    "twitterSite": "@example",
    "locales": { "en": "en_US", "tr": "tr_TR" },
    "noAlternates": false
  }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.4 was the last for v0.60.0). The
  canonical and `og:` URLs follow the site's origin: `Config.BaseURL` or, per host,
  what an `OriginResolver` plugin such as elagoht/tenant gives. With neither, the
  application does not start. v0.3.0 removed the plugin's own `baseURL`.
- Registering it gives every page `og:site_name`, `og:type`, the canonical URL —
  built by name, so a query string a reader arrived with is never part of it —
  `og:locale`, an `hreflang` link for every locale the page has a path in, and the
  default image and Twitter card.
- What a page is about it says from its data handler with `meta.Set`. Each tag is
  hoisted under its own key, so a page's declaration replaces the default tag by
  tag, and a deeper fragment's replaces its parent's.
- The `<title>` is not among them: it is `rc.HoistTitle` or `WithTitle`, as without
  the plugin. It needs `{{hoist "head"}}` in the layout.

#### elagoht/ogimage

[github.com/Elagoht/collage-ogimage](https://github.com/Elagoht/collage-ogimage)
draws each page's share card — the 1200×630 image a link is shown with — from an
HTML template, in pure Go, and serves it at a URL made from its content.

```go
import "github.com/Elagoht/collage-ogimage"

Plugins: []collage.Plugin{
	meta.New(meta.Options{SiteName: "The blog"}),
	ogimage.New(ogimage.Options{
		Templates: templatesFS,
		Root:      "templates",
		Default:   "og/default.html",
		Dir:       ".cache/ogimage",
	}),
},
```

```html
<!-- templates/og/post.html -->
<div style="display:flex; flex-direction:column; justify-content:space-between;
            width:1200px; height:630px; padding:72px; background:#0f172a; color:#fff">
  <span style="font-size:28px">{{.Site.Name}}</span>
  <h1 data-fit style="font-size:72px">{{.Title}}</h1>
  <span style="font-size:28px">{{.Label}}</span>
</div>
```

```go
if err := ogimage.Set(rc, "og/post.html", ogimage.Card{Title: post.Title, Label: post.Category}); err != nil {
	return view{}, nil, err
}
```

- v0.4.0 needs collage v0.61.4 or later (v0.2.4 was the last for v0.60.0), and the application's `Config.BaseURL`;
  a card's URL follows the request's origin on a site served on several hosts, so
  `Config.BaseURL` is not required when an `OriginResolver` plugin is registered.
  Register it after elagoht/meta, and leave meta's `DefaultImage` unset.
- v0.3.0 has one constructor, `ogimage.New(ogimage.Options{…})`. `renderTimeout`
  is a duration string (`"5s"`), a negative `maxEntries` or `maxBytes` is an error
  rather than unlimited, and the build finding is `ogimage-draw`.
- v0.4.0 takes the two strings of `ogimage.Draw` as one value:
  `ogimage.Draw(ctx, opts, ogimage.Input{HTML: html, Locale: locale})`, so swapping
  them no longer compiles; an empty `Locale` still draws as `"en"`. A static build
  now draws every card, even when there are more than `maxEntries` or `maxBytes`
  keep: it used to fail with `ogimage-draw` "file does not exist" for the cards
  evicted before they were drawn.
- A card is an `html/template` file under `og/` in a subset of HTML and CSS: an
  element holding boxes says `display:flex`, one holding text is a paragraph of
  runs. What it cannot draw stops the application at startup, with the template,
  line and column.
- A page that sets no card gets `Default`, drawn from its title and description.
- The URL is a hash of what is drawn, served immutable; nothing is invalidated, a
  changed page has a new card. A card is drawn on its first request, only if a
  render recorded it; `collage export` writes them. `Dir` keeps them across
  restarts, which a disk page cache needs.
- In development `/_og-preview/` lists them; `go run . ogimage og/post.html
  card.json > card.png` draws one without the site, in the default locale or the
  one `-locale tr` names.

#### elagoht/sitemap

[github.com/Elagoht/collage-sitemap](https://github.com/Elagoht/collage-sitemap)
serves `/sitemap.xml` from the pages the application registered.

```go
import "github.com/Elagoht/collage-sitemap"

BaseURL: "https://example.com",
Plugins: []collage.Plugin{sitemap.New(sitemap.Options{})},
```

```json
{
  "elagoht/sitemap": {
    "path": "/sitemap.xml",
    "exclude": ["thanks"],
    "maxURLs": 50000
  }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.3 was the last for v0.60.0). A sitemap lists absolute URLs, so it needs an
  origin: `Config.BaseURL`, or the request's own origin on a site served on several
  hosts, as an `OriginResolver` plugin names it. With neither, the application does
  not start. v0.3.0 removed the plugin's own `baseURL`.
- It lists every page with a path, in every locale, as `App.URL` spells it, with
  its other locales as `hreflang` alternates. A `{param}` pattern is listed once
  for each value its `WithStaticParams` returns, the URLs a static build writes;
  one without it is left out. `Exclude` leaves pages out by name.
- `LastMod`, a Go function, gives a page's `<lastmod>`.
- It is a static document: cached, exported, and made again when `sitemap.Tag` is
  invalidated — invalidate it with the tags of a post you publish.
- Past 50,000 URLs (`maxURLs`; 0 is 50,000, a negative value stops the
  application) it becomes a sitemap index of numbered files.

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

- v0.3.0 needs collage v0.61.0 or later (v0.2.4 was the last for v0.60.0). With no rules it allows every crawler
  everything; a rule with no user agents is for `*`. A negative `crawlDelay` stops
  the application. v0.3.0 removed `(*Plugin).Render`: request `/robots.txt` from
  the application instead.
- `disallowAll` closes the site to every crawler, whatever the rules say, and
  sends every response with `X-Robots-Tag: noindex, nofollow`. Set it in the
  configuration of a staging deployment, so one binary is open in production and
  closed elsewhere.
- A sitemap may be given as a path (`"/sitemap.xml"`), which is made absolute per
  request against the request's origin: `Config.BaseURL` or, on a multi-host site,
  an `OriginResolver` plugin's. A scheme-relative
  `//host/path` is refused, and with no origin to resolve a path against the
  application does not start.
- The body is fixed at startup, and a static build writes it to `robots.txt`.

#### elagoht/feed

[github.com/Elagoht/collage-feed](https://github.com/Elagoht/collage-feed) serves
RSS 2.0 and Atom 1.0 feeds from items the application lists, and announces them in
every page's head.

```go
import "github.com/Elagoht/collage-feed"

BaseURL: "https://example.com",
Plugins: []collage.Plugin{feed.New(feed.Feed{
	Title:   "The blog",
	Link:    "/blog",
	Items:   latestPosts, // func(ctx) ([]feed.Item, error), newest first
	Tags:    []string{"posts"},
})},
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.4 was the last for v0.60.0). It is configured in Go only, since `Items` is
  a function; `collage plugin add` writes `feed.New()`, and with no feed given the
  plugin logs a warning and serves nothing.
- A feed is served at `/feed.xml` as RSS and `/atom.xml` as Atom; `RSS` and `Atom`
  move them, and `"-"` leaves a format out. Several feeds each take a `Name` and
  their own paths. It carries at most `Limit` items, 20 by default.
- Every page gets a `<link rel="alternate">` for every feed, so it needs
  `{{hoist "head"}}` in the layout; `NoDiscovery` keeps a feed out of the heads.
- It is a static document: cached, exported, and made again when one of its
  `Tags` is invalidated.
- **Links are written only as `http` or `https`** (since v0.2.1). An item link, a
  URL-shaped `ID`, or the channel link with any other scheme (`javascript:`, `data:`,
  `file:`, however it is spelled) is dropped and logged once at `WARN`, with the scheme
  and not the link. A feed's own `Link`, `RSS` or `Atom` with such a scheme stops the
  application from starting with `ErrUnsafeLink`. An absolute `http` or `https` link
  is written as given (v0.2.2; v0.2.1 re-encoded a non-ASCII path).
- A feed's links follow the request's origin, `Config.BaseURL` or an
  `OriginResolver` plugin's, so one feed serves each host its own links. v0.3.0
  removed `Feed.BaseURL`; with no origin the application does not start.

#### elagoht/redirects

[github.com/Elagoht/collage-redirects](https://github.com/Elagoht/collage-redirects)
serves redirects kept in a file rather than in code — what a site migration leaves
behind — before routing, and hands them to a static build for the host's own
configuration.

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
/moved           https://new.example/moved 308
/old-product     -               410
```

```json
{
  "elagoht/redirects": {
    "file": "redirects.txt",
    "rules": [{ "from": "/careers", "to": "https://jobs.example.com", "status": 302 }]
  }
}
```

- v0.2.3 needs collage v0.61.4 or later (v0.2.2 was the last for v0.61.0); it changes no behaviour.
- One rule a line: the old path, where it went, and a status — `301` when left out,
  or `302`, `307`, `308`, or `410` with `-` for a page that is gone, answered with
  the site's own not-found page and status `410`. `/blog/*` is a prefix, and
  `:splat` in the target is what the `*` matched. The first rule that matches wins,
  and the reader's query string is carried over.
- A file that is wrong stops the application from starting, and says where: a
  malformed line, a rule no request can reach, rules that send a reader round in a
  circle. Since v0.2.0 so does a control character in a rule, or a target
  beginning `/\`, which a browser reads as another host.
- Since v0.2.0 it no longer serves or writes `_redirects` (breaking):
  `Options.NoRedirectsFile` and `RedirectsFile()` are gone, and since collage
  v0.61.0 a `noRedirectsFile` key in configuration stops the application as an
  unknown key. It is a [`RedirectSource`](/docs/writing-plugins#redirectsource): a
  static build takes its rules, and [elagoht/deploy](#elagohtdeploy) writes them
  in the host's format. Without a deploy plugin nothing is written.
- A rule collage's pattern syntax cannot say is still served, but left out of the
  build with a `redirects-not-exported` warning. A rule covering a page's or a
  document's own redirect is a `redirects-overlap` warning: the server follows the
  rule, a static host the page's redirect.
- A static build fails on a rule over a file it wrote
  (`collage.ErrRedirectShadowsFile`) or from the same path as another redirect
  (`collage.ErrDuplicateRedirect`), so a site whose export worked with v0.1.x can
  fail until the rule or the page goes. Rules can also be given in Go or
  configuration.

#### elagoht/indexnow

[github.com/Elagoht/collage-indexnow](https://github.com/Elagoht/collage-indexnow)
tells search engines which URLs changed, through the IndexNow protocol — Bing,
Yandex, Seznam, Naver and the others that share what one of them is told.

```go
import "github.com/Elagoht/collage-indexnow"

BaseURL: "https://example.com",
Plugins: []collage.Plugin{indexnow.New(indexnow.Options{
	Key: "4f1c2e9a7b3d4c8e",
})},
```

```json
{
  "elagoht/indexnow": {
    "key": "4f1c2e9a7b3d4c8e",
    "window": "10s",
    "skip": ["/api", "/search"]
  }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.2 was the last for v0.60.0). `key` is required. URLs follow
  `Config.BaseURL`; on a site served on several hosts one submission is sent per
  origin, and an origin two hosts share is sent once.
- v0.3.0 removed `baseURL` and renamed `exclude` to `skip` — request path
  prefixes never sent, `"/api"` covering `/api` and what is under it. `window` and
  `backoff` are duration strings, and a negative `retries` or `batchSize` stops the
  application.
- What it sends is what the cache let go of: the
  [paths an invalidation reached](/docs/caching#invalidating-by-path). A page that
  was not cached when its tag was invalidated is not sent.
- URLs are batched for `window`, sent in the background and retried on a `429` or
  `5xx`; what is waiting is flushed on shutdown.
- It serves the key at `/<key>.txt`, and a static build writes it there. A
  development server and a static build send nothing.

### Content

Where the words come from and how they are presented: Markdown rendered as HTML,
coloured code, a table of contents, search on a static site, and translations.

#### elagoht/markdown

[github.com/Elagoht/collage-markdown](https://github.com/Elagoht/collage-markdown)
renders Markdown strings as HTML: a `{{markdown}}` template function, `Render`
from Go, and `Parse`, which also reads a document's front matter into a type of
the application's own. It reads no files and builds no pages: where the Markdown
comes from, and the page, its data and its SEO, are the application's.

```go
import "github.com/Elagoht/collage-markdown"

md := markdown.New(markdown.Options{})

Plugins: []collage.Plugin{md},
```

```html
<article>{{markdown .Body}}</article>
```

```go
type Post struct {
	Title string    `md:"title"`
	Date  time.Time `md:"date"`
	Tags  []string  `md:"tags"`
}

doc, err := markdown.Parse[Post](md, rc.Locale, src) // doc.Meta, doc.HTML, doc.Text, doc.Headings
```

```json
{
  "elagoht/markdown": {
    "extensions": ["table", "strikethrough", "linkify", "taskList", "footnote", "mark", "headingIDs", "headingAttributes"],
    "unsafe": false
  }
}
```

- v0.4.0 needs collage v0.77.0 or later and must be in `Config.Plugins`: it adds a
  template function.
- `extensions` picks from `table`, `strikethrough`, `linkify`, `taskList`,
  `footnote`, `mark` (`==text==` is `<mark>text</mark>`), `headingIDs`,
  `headingAttributes`, `definitionList` and `typographer`. Left out, it is all but
  the last two; a list replaces that whole, `[]` turns them all off, and an unknown
  name stops the application.
- A front matter key is read into the field tagged `md:"key"`, or the untagged field
  of that name; a key no field takes is ignored, a value that does not convert is an
  error. `markdown.FrontMatter[Post](src)` reads it without rendering.
- Heading ids keep letters of any alphabet and follow the page's locale
  (`## IŞIK` is `ışık` in Turkish). Code blocks are `<pre><code class="language-go">`,
  which [elagoht/highlight](#elagohthighlight) colours when it is registered.
- Raw HTML, and a heading's `{style="…"}` and `on…` attributes, are left out unless
  `unsafe` is on; `id` and `class` stay.
- Changed in v0.4.0: the plugin no longer reads a directory. `FS`, `Dir`,
  `LocaleDirs`, `Drafts`, `Name`, `Handler`, `IndexHandler`, `StaticParams`, `List`,
  `Get`, `Tag` and `DirTag` are gone; the application reads its own files and builds
  its pages with `markdown.Parse`. The plugin's README shows a blog built that way.

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
  "elagoht/highlight": { "light": "github", "dark": "github-dark", "noBackground": true }
}
```

- v0.3.0 needs collage v0.61.0 or later. It must go in
  `Config.Plugins`: it adds `{{highlight}}`.
- Unless `noAuto` is set, every `<pre><code class="language-go">` a page holds — what
  elagoht/markdown and most Markdown renderers write — is coloured where it
  stands, once per render; a cached page is served as it was coloured.
- The stylesheet is linked from pages with coloured code only, by a
  content-addressed name, each theme under its own `prefers-color-scheme` query.
  `noAuto: true` leaves those blocks alone, so only `{{highlight}}` colours code;
  since v0.3.0 it replaces `"auto": false`, and `"auto"` is an unknown key.
  Since v0.2.0 the automatic pass hoists it with `AfterRenderEvent.Hoist`, so it lands
  where the layout put `{{hoist "head"}}`, or before `</head>` without one;
  `{{highlight}}` needs `{{hoist "head"}}` in the layout.

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
    "text": { "label": "Table of contents", "readingTime": "{n} min read" },
    "locales": { "tr": { "label": "İçindekiler", "readingTime": "{n} dk okuma" } }
  }
}
```

- v0.2.0 needs collage v0.61.0 or later, and must go in `Config.Plugins`: it adds
  template functions.
- Since v0.2.0 the default locale's words are under `text` (`Options.Text`, a named
  field rather than an embedded one), not at the top level, and `locales` holds
  the others; `"ids": false` is `"noIDs": true`. An old top-level `label` or `ids`
  stops the application as an unknown key.
- Each function writes a placeholder, and the plugin fills it in once the page has
  rendered: the `h2` to `h4` inside `<main>`, as a nested list, and the words
  divided by `wpm`.
- A heading without an `id` gets one from its text, letters of any alphabet kept
  and lowered as the page's language lowers them; `noIDs` leaves a page without
  placeholders alone.
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
    "exclude": ["admin", "tags"],
    "text": { "label": "Search" },
    "locales": { "tr": { "label": "Ara", "noResults": "Sonuç yok" } }
  }
}
```

- v0.2.0 needs collage v0.61.0 or later, and must go in `Config.Plugins`: it adds
  `{{searchBox}}`.
- Since v0.2.0 `exclude` takes page names, left out in every locale, not path
  prefixes: an old `"/admin/"`, or a name no page has, stops the application. What
  the box says is under `text` (`label`, `placeholder`, `results`, `oneResult`,
  `noResults`, `unavailable`), and `locales` holds the same per locale.
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
{ "elagoht/i18n": { "dir": "locales", "strict": true } }
```

- v0.4.0 needs collage v0.61.4 or later (v0.3.0 needed v0.61.0), and must go in
  `Config.Plugins`: it adds template functions. The catalogs come from
  `Options.FS`, so it is constructed in Go.
- One JSON file of nested keys per supported locale, `locales/<locale>.json`; the
  application does not start while a supported locale has none.
- `t` translates a key and fills `{name}` from name and value pairs, `tn` picks a
  plural form for a count, `th` allows markup from the catalog. A data handler
  calls `i18n.T(rc, key, pairs...)`; since v0.3.0 the pairs are strings
  (`i18n.T(rc, "visits", "n", strconv.Itoa(n))`), while the template functions
  still print any value.
- **`th` escapes each value for where it lands** (since v0.4.0), as html/template
  escapes `{{.Value}}` there: the catalog's markup is trusted, the values filling it
  are not. In text a value is escaped; in a URL attribute (`href`, `src`, `action`…)
  one that is not `http:`, `https:`, `mailto:` or relative becomes `#ZgotmplZ`; in
  an unquoted attribute its spaces and `=` are escaped, so it cannot add an
  attribute. Up to v0.3.0 a value was HTML-escaped wherever it was, so a
  `javascript:` value in `<a href="{url}">` made a working link. An entry that ends
  inside a tag or an attribute now fails the render with an `i18n:` error: close
  the tag in the catalog. A safe URL or plain text renders as before, except that a
  URL's spaces and non-ASCII letters are now percent-encoded.
- A missing key falls back to the default locale, then to the key, and is reported
  as `i18n-missing` over the page in development and in a static build's
  report. In development the catalogs are read again on every request.
- **Outside a render** (since v0.2.0), for an email or a background job, keep the
  plugin's value: `tr.In(user.Locale)` returns a `Translator` whose `T`, `TN` and
  `TH` follow the same catalogs, fallback and plurals, and whose `Funcs()` gives an
  application's own `html/template` `t`, `tn` and `th`. A locale the application
  does not support is the default one. A key it cannot find is logged, once per key
  and locale. The catalogs are read when the application starts, not in
  `collage.New`; a translation asked for before then is the key itself.
- A key one catalog has and another lacks is logged at startup and listed as
  `i18n-untranslated` in a build, in either direction since v0.2.0 (the codes were
  `missing-translation` and `untranslated` before v0.3.0). A plural form a
  locale's rule never picks is not counted; `zero` is. With `strict` the
  application does not start while the catalogs differ.
- A key written both dotted and nested, `{"nav.home": …}` beside
  `{"nav": {"home": …}}`, refuses to start since v0.2.0; which text it showed
  used to depend on map order.

### Forms and state

What a form needs around an [action](/docs/forms-and-actions): validation, spam
protection, a message shown after the redirect, a session in a cookie, and file
uploads.

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

Each field gives back what was checked, as its type (since v0.3.0):

```go
servings := v.Field("servings").Required("How many servings?").Int().Range(1, 50)
minutes := v.Field("minutes").Int()                // optional
steps := v.Field("steps").Required().Lines()       // a <textarea>, one per line
tags := v.Field("tags").All().OneOf(tagNames...)   // a group of checkboxes
if !v.Valid() {
	return validate.Refuse(rc, v, rc.Page), nil
}
create(servings.Value(), minutes.Ptr(), steps.Value(), tags.Value()) // int, *int, []string, []string
```

```html
<input name="email" type="email" value="{{fieldValue "email"}}">
{{with fieldError "email"}}<p class="error">{{.}}</p>{{end}}

<!-- a form that starts from a record: the stored name, or what was typed -->
<input name="fullname" value="{{fieldValue "fullname" .User.FullName}}">
```

```json
{
  "elagoht/validate": {
    "messages": { "required": "Please fill this in." },
    "localeMessages": { "tr": { "required": "Bu alan zorunludur." } },
    "noRefill": ["pin", "card"]
  }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.1.7 was the last for v0.60.0). It must go in
  `Config.Plugins`: it adds `{{fieldError}}`, `{{fieldValue}}` and `{{hasErrors}}`.
- `{{fieldValue "name" fallback}}` shows what was submitted when a submission was
  refused, and the fallback when the form is shown for the first time — the value
  a profile or an edit form starts with.
- Since v0.1.2 `validate.Form` reads nothing until a check or `v.Value` needs the
  form. An action refusing a body before reading it — a photo whose
  `Content-Length` is over its size — builds a validator for `v.Fail` and
  `validate.Refuse`, and the body is never read. Since v0.1.3 such a refusal shows
  each field's fallback, since there is nothing typed to show. The action's
  `WithMaxBodyBytes` must be above the size it checks, or collage's `413` answers
  first.
- It is the first half of collage's rule for a form — see
  [Forms and actions](/docs/forms-and-actions#the-validation-re-render). On a page
  nobody submitted the functions are empty, so one template serves both renders
  and the first stays cacheable.
- The checks are `Required`, `MinLen`, `MaxLen`, `Email`, `URL`, `Int`, `Range`,
  `OneOf`, `Matches`, `Equal` and `Custom`; `v.Fail` reports what only the
  application knows. A field keeps its first message. Every check but `OneOf`
  takes an optional message of its own, `Required("…")`, as `.Message("…")`
  after it would (since v0.3.0).
- **Typed values (since v0.3.0).** `Field.Value()` is the string, trimmed.
  `Int()`, `Float()`, `Bool()` (a checkbox), `Date()` (`2026-10-25`), `All()` (every
  value under the name) and `Lines()` (a textarea's lines) turn the field into a
  typed one whose `Value()` is an `int`, `float64`, `bool`, `time.Time` or
  `[]string`, and whose checks work on it: `Min`, `Max`, `Range`, `MinItems`,
  `MaxItems`, `Custom(func(int) string)`. `Ptr()` is nil for a field left empty —
  an optional field's "not given". A field that failed gives the zero value.
  **`Int()` returns an `*IntField`**: a string check after it no longer compiles;
  `.Int().Message(…)` and `.Int().Range(…)` still work.
- Messages are English by default, and are replaced per check, per locale or for
  every locale. A field whose name contains `password` is never typed back;
  `noRefill` adds names of your own to it (since v0.2.0 it no longer replaces
  `password`).
- `validate.Values(rc, values)` (since v0.2.0) checks values something else
  already read — the fields `uploads.Receive` hands back after streaming a
  body — and `Refuse` types them back the same way.
- Field names are one space per render: two forms on one page sharing a field
  name both show that field's message and value. Give each form's fields names of
  their own.

#### elagoht/honeypot

[github.com/Elagoht/collage-honeypot](https://github.com/Elagoht/collage-honeypot)
stops form spam without a CAPTCHA: a decoy field people never see and bots fill
in, and a signed timestamp that refuses a form that was never served, was served
too long ago, or — when the form asks — came back too soon.

```go
import "github.com/Elagoht/collage-honeypot"

Plugins: []collage.Plugin{honeypot.New(honeypot.Options{})},
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
    "minDelay": "0s",
    "maxAge": "24h",
    "silent": false,
    "protect": []
  }
}
```

- v0.5.0 needs collage v0.61.0 or later (v0.4.4 was the last for v0.60.0), and must go in
  `Config.Plugins`: it adds `{{honeypot}}`.
- **Since v0.6.0** the key is read from `HONEYPOT_KEY` with nothing configured:
  put `HONEYPOT_KEY=` (`openssl rand -hex 32`) in `.env.development` and where
  production runs. `keyEnv` names another variable and wins, then `Options.Key`,
  then `HONEYPOT_KEY`; a named variable must be set, `HONEYPOT_KEY` may be empty,
  and either one not hex stops the application.
- **Changed in v0.5.0:** the hex `key` is gone — `keyEnv` names the environment
  variable holding it, and `Options.Key` takes the bytes from Go. `maxAge` and
  `minDelay` are duration strings (`"24h"`, `"300ms"`); `{{honeypot 0.3}}` still
  takes seconds, and `"0s"`, the default, is no delay (`-1` is now an error). An
  old key or a bare number stops the application at start, naming the key.
- It has no say in how large a submission may be. It checks the form in
  `BeforeActionHook`, through the action's own body limit, so a form that uploads
  large files needs only the action's `WithMaxBodyBytes`. Only collage actions are
  checked, not handlers mounted with `app.Handle`.
- The forms say which paths are protected: once a page with a `{{honeypot}}` form
  has been served, a form body posted to that form's `action` is checked before it
  reaches the action. A form without `{{honeypot}}` is never checked, and nothing
  in the configuration has to name the forms. A JSON body and every `GET` pass
  unchecked.
- Since v0.4.3 an action built with `WithStreamingBody()` passes unchecked too: its
  form is never parsed, so there is nothing to check. Before, every streaming upload
  on a protected path was refused with `400`, or with `silent` dropped — upgrade it
  together with collage v0.57.0.
- A process learns the paths from the pages it serves. After a restart or a new
  release, until the form's page has gone out again, a bot posting straight to
  the path is not checked. `protect` lists prefixes checked from the first request;
  `["/"]` checks every form the site accepts. Before v0.2.0 that was the default.
- No delay by default. `{{honeypot 0.3}}` asks a form to be open at least that
  many seconds before it is sent, signed into the timestamp so a bot cannot
  shorten it; `minDelay` (`"300ms"`) sets one for every form that does not choose
  its own.
- The timestamp survives the page cache as collage's forgery token does: the
  cached page carries a placeholder, and the plugin signs the current time into it
  through `PersonaliseHook`, after the page cache and before compression. Its place
  in `Config.Plugins` does not matter. (Before v0.4.0, listed before
  elagoht/compress it left the placeholder in the gzip body and every real
  submission was refused.) A stamped page is sent with `Cache-Control: private,
  no-store` (set by collage) and a new `ETag` on every response. Set a key of at
  least 32 random bytes, the same on every instance, hex-encoded in the variable
  `keyEnv` names; it never goes in `plugins-config.json`.
- Only what collage renders is stamped: a page, a fragment, an action's HTML
  answer, an error page. A placeholder a hand-written `app.Handle` handler writes
  itself goes out as it is, and its form is refused. (Before v0.4.0 such a
  response was stamped, its type sniffed from the first bytes.)
- A refusal is a `400`; with `silent` it is a `303` back to the form, as an accepted
  form answers. It stops careless bots, not a determined one — pair it with
  elagoht/ratelimit.

#### elagoht/flash

[github.com/Elagoht/collage-flash](https://github.com/Elagoht/collage-flash) adds
flash messages: a message an action sets before it redirects, shown once by the
page it redirects to.

```go
import "github.com/Elagoht/collage-flash"

Plugins: []collage.Plugin{flash.New(flash.Options{})},
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
    "cookie": "collage_flash",
    "maxAge": "5m"
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.4 was the last for v0.50.0), and must go in `Config.Plugins`: it adds
  `{{flashes}}`.
- **Since v0.3.0** the key is read from `FLASH_KEY` with nothing configured:
  put `FLASH_KEY=` (`openssl rand -hex 32`) in `.env.development` and where
  production runs. `keyEnv` names another variable and wins, then `Options.Key`,
  then `FLASH_KEY`; a named variable must be set, `FLASH_KEY` may be empty, and
  either one not hex stops the application.
- **Changed in v0.2.0:** the hex `key` became `keyEnv`, the name of the
  environment variable holding it (`Options.Key` from Go is unchanged), and
  `maxAge` is a duration string, `"5m"` by default, where it was seconds. An old
  `key` stops the application with `unknown key "key"`.
- The messages travel in a signed, `HttpOnly` cookie, `Secure` when
  `collage.IsHTTPS` says the request came over HTTPS. Set a key of at least 32
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

Plugins: []collage.Plugin{session.New(session.Options{})},
```

```go
s := session.Get(rc)
s.Regenerate()
if err := s.Set(session.UserKey, user.ID); err != nil {
	return nil, err
}
return collage.SeeOther("/account"), nil
```

A layout carrying `session.RequireUserPage` keeps signed-out readers off every page
it wraps — a [guard](/docs/pages-and-layouts#private-pages-guards) that sends them
to the page named `login`:

```go
private := collage.NewFragment("private", "layouts/private.html").
	WithGuard(session.RequireUserPage("login")).
	Build()
```

```json
{
  "elagoht/session": {
    "encrypt": true,
    "maxAge": "168h",
    "idleTimeout": "0s",
    "sameSite": "lax"
  }
}
```

- **Since v0.5.0** the key is read from `SESSION_KEY`, and the previous keys
  from `SESSION_PREVIOUS_KEYS`, with nothing configured: put `SESSION_KEY=`
  (`openssl rand -hex 32`) in `.env.development` and where production runs.
  `keyEnv` and `previousKeysEnv` name other variables and win, then the options
  set from Go, then these two; a named variable must be set, these may be empty,
  and a key that is not hex stops the application.
- v0.4.0 needs collage v0.73.0 or later (v0.3.1 was the last for v0.61.4); v0.3.0 needed v0.61.0 (v0.2.4 was the last for v0.60.0). It adds no
  template function, so `RegisterPlugin` accepts it too.
- **Changed in v0.3.0:** keys never go in `plugins-config.json`. `keyEnv` names
  the environment variable holding the hex key and `previousKeysEnv` the one
  holding previous keys, comma-separated; from Go, `Options.Key` and
  `Options.PreviousKeys` take the bytes. `maxAge` and `idleTimeout` are duration
  strings (`"168h"`, `"30m"`) where they were seconds. An old `key` or
  `previousKeys` stops the application with `unknown key`.
- **`RequireUser(loginPath)`** (since v0.2.0) lets through a reader whose session
  holds a value under `session.UserKey` (`"user"`), and sends everyone else `303` to
  `loginPath` with where they were going in `next`: `/login?next=%2Fdashboard`. A
  login path with a query of its own keeps it. `Require(key, loginPath)` is the
  same for any key — `Require("admin", "/login")`. Without the plugin in the
  application the guard fails the request with `ErrNoSession` rather than sending
  every reader to a login that could never let them in. Pass `next` through
  `collage.SafeRedirect` (collage v0.44.0) before the login action redirects to
  it: `collage.SafeRedirect(rc.Request.URL.Query().Get("next"), "/")`.
- **`RequireUserPage(loginPage)`** (since v0.4.0) is `RequireUser` with the login
  page named rather than its path written: the path is built with `collage.URL`
  for each request, in its locale — `/tr/giris?next=…` from a Turkish page — so a
  renamed or translated path cannot leave the guard pointing at the old one.
  `RequirePage(key, loginPage)` is its form for any key. A name the application
  has not registered fails the request rather than letting anyone through.
- **`RequireGuest(homePath)`** and **`RequireGuestPage(homePage)`** (since
  v0.6.0) guard the pages only a signed-out reader needs — sign in, sign up, a
  forgotten password. A signed-out reader goes through; a signed-in one gets
  `303` to the request's `next` when `collage.SafeRedirect` finds it a path on
  this site, or to the home path, or the named page in the request's locale.
  `WithGuard(session.RequireGuestPage("account"))` on the auth pages' layout.
- `Get`, `Set`, `Delete`, `Clear`, `Regenerate` and `ID`; a handler of your own
  reads it with `session.FromContext(r.Context())`.
- **A request carrying a valid session is rendered fresh**, neither read from the
  page cache nor written to it, and marked `private, no-store`. A reader without
  one is served from the cache as before. Set a session from an action, not from a
  cached page.
- The cookie is `HttpOnly` and `SameSite=Lax`, written only when the session
  changed, and `Secure` when `collage.IsHTTPS` says the request came over HTTPS.
  Keys rotate through the previous keys: a cookie under one is accepted and
  rewritten under the current key. A session cannot be revoked: it lives in the
  reader's cookie.
- **A session is bound to the host it was issued on** (since v0.7.0): the host,
  lower-cased and without its port, is inside the signed payload, and a cookie
  presented on another host is no session there — ignored, not cleared. On one
  process serving several sites with one key, elagoht/tenant's tenants, a cookie
  copied from `acme.test` to `globex.test` is anonymous there. With `domain` on a
  one-site application the session is bound to the domain instead, so its
  subdomains share it; with tenants (any plugin resolving hosts to origins) each
  host stays its own, so do not set `domain` there. The host is the request's
  `Host`: behind a proxy that rewrites it, forward the original, or every session
  is bound to the proxy's.
- **Cookies from before v0.7.0** carry no host. A one-site application accepts one
  and rewrites it bound to the host it arrived on, so the upgrade signs nobody out;
  a multi-site one cannot tell which tenant issued it, so it is refused and
  cleared, and every reader signs in again once. Nothing in the configuration
  changes.

#### elagoht/uploads

[github.com/Elagoht/collage-uploads](https://github.com/Elagoht/collage-uploads)
receives file uploads: it checks each file by its content, stores it under a random
key, and serves it back without letting it act as your site. It reads a form that
was already parsed, or streams the body part by part, with the same result.

```go
import "github.com/Elagoht/collage-uploads"

u := uploads.New(uploads.Options{Dir: "var/uploads"})

Plugins: []collage.Plugin{u},
```

```go
got, err := u.Receive(rc,
	uploads.Field("avatar", uploads.Rules{
		MaxSize:  5 << 20,
		MaxFiles: 1,
		Types:    []string{"image/png", "image/jpeg"},
	}),
)
var rule *uploads.Error
if errors.As(err, &rule) {
	// Show rule.Field's error beside the form; got.Values holds what was typed.
	return showFormAgain(rc, got.Values, rule)
}
if err != nil {
	return nil, err
}
avatar := got.Files["avatar"][0] // Key, Name, Type, SHA256, Size
```

```go
app.Handle("/uploads/", u.Handler())
```

```json
{
  "elagoht/uploads": {
    "dir": "var/uploads"
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.1 was the last for v0.58.0). It adds no
  template function, so `RegisterPlugin` accepts it too. Keep it in a variable:
  the action calls `u.Receive` and the application mounts `u.Handler()`. Since
  collage v0.67.0 that mount runs with no middleware; add
  `collage.WithMiddleware()` to the `app.Handle` call if a basicauth, a session
  check or elagoht/compress should cover the files.
- **Changed in v0.2.0:** one constructor, `uploads.New(uploads.Options{…})`:
  `New(store)` is `New(uploads.Options{Store: store})` and `NewWith(uploads.Config{Dir: d})`
  is `New(uploads.Options{Dir: d})`. `Handler` takes no argument; its
  `Cache-Control` is `cacheControl` in configuration or `Options.CacheControl`.
  The configuration is read even with a store given in Go, so an unknown key
  stops the application there too.
- **Two paths.** On an ordinary action, a plain `<form enctype="multipart/form-data">`
  with `{{csrfToken}}` works with no JavaScript, and `Receive` reads the parsed
  form — fine for files up to tens of MB. On an action built with
  `WithStreamingBody()`, it walks the body with `MultipartReader` and streams each
  file straight into the store; such an action takes its token only from the
  `X-CSRF-Token` header, so post to it with `fetch()`. Anything that parses the form
  first, `validate.Form` included, gives up the stream; read the other fields from
  `got.Values`. See [Streaming bodies](/docs/forms-and-actions#streaming-bodies).
- **Typed by content.** A file's type is sniffed from its first 512 bytes; its
  name, extension and `Content-Type` are never consulted. A broken rule is an
  `*uploads.Error` with the field and one of `ErrTooLarge`, `ErrType`, `ErrTooMany`
  or `ErrUnexpectedField`.
- **All or nothing.** If anything fails — a broken rule, a client that went away, a
  body over its limit, the store — every file that call stored is deleted before
  `Receive` returns. Once it succeeds the files are yours: delete them with
  `u.Store().Delete` if the action fails after that.
- **`Dir`** keys each file by 128 random bits plus an extension for its sniffed
  type, so nothing the client sent goes into a path. `Name` is the client's file
  name, cleaned, for display only. It writes 0600 files in a 0700 directory, never
  overwrites, and reads only regular files. Metadata — name, owner, upload time —
  is yours to keep in your database next to the key.
- **`Handler`** serves a file by key with `X-Content-Type-Options: nosniff`,
  `Content-Security-Policy: sandbox`, its type sniffed again, and
  `Content-Disposition: attachment` unless it is a PNG, JPEG, GIF or WebP image. It
  does no access control: anyone with a key can fetch the file, so serve private
  files from an action of your own. Never serve the directory directly.
- **Abuse and disk.** Guard the upload action (a token is not authentication), keep
  `MaxBodyBytes` and `MaxSize` tight, rate-limit the path, keep the store on a volume
  of its own, and give the action `WithBodyTimeout`; see
  [Deployment](/docs/deployment#uploads) for proxies. A crash between `Receive` and
  your database write leaves files no row points to; sweep the directory for old
  keys your database does not know, skipping `.upload-*` temp files younger than an
  hour, which may be uploads in progress.

### Security

Headers a site should send, a limit on how fast one client can hit it, a
password in front of a site that is not public yet, sign-in with an identity
provider, and the visitor's consent to cookies.

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

- v0.3.1 needs collage v0.61.4 or later (v0.2.5 was the last for v0.60.0), and must go in
  `Config.Plugins`: it adds `{{cspNonce}}`. Its place in the list does not matter:
  the nonce goes in after the cache and before compression, through
  `PersonaliseHook`, so it works whether it is listed before or after
  elagoht/compress. (Before v0.2.0, listed before compress it left the marker in
  the gzip body and the CSP blocked every inline script.)
- By default it sends `X-Content-Type-Options`, `X-Frame-Options`,
  `Referrer-Policy`, `Cross-Origin-Opener-Policy` and, over TLS or behind a proxy
  sending `X-Forwarded-Proto: https`, `Strict-Transport-Security`. Since v0.3.0
  that header is believed only from a proxy in `Server.TrustedProxies` when any
  are listed (`collage.IsHTTPS`). `Permissions-Policy` and the CSP are sent when
  set; `"-"` leaves a header out, `X-Content-Type-Options` included
  (`"contentTypeOptions": "-"`, which replaced `"noSniff": false` in v0.3.0).
- `{nonce}` in the policy and `{{cspNonce}}` on an inline script are one nonce,
  new on every response: the cached page carries a placeholder, and the plugin's
  hook puts a fresh nonce in its place. A page carrying one is sent with
  `Cache-Control: private, no-store` (set by collage) and a new `ETag` on every
  response. Every other response answers conditional requests as usual: a feed,
  a document or a mounted file still gets its `304` (since v0.1.5; before, with a
  policy set, nothing behind the plugin could).
- Set a key: at least 32 random bytes, the same on every instance and across
  restarts, in hex (`openssl rand -hex 32`) in `SECURE_KEY`, which is read with
  nothing configured (since v0.4.0). A secret never goes in `plugins-config.json`:
  `keyEnv` names another variable and wins, or `Options.Key` takes the bytes from
  Go, and wins over `SECURE_KEY` (v0.3.0; `key` and `Options.KeyHex` are gone). With it the
  `{{cspNonce}}` placeholder is the same in every process, so a page from a disk
  cache that survived a deploy, or from a cache shared by several instances, still
  gets its nonce. Without a key the placeholder changes per process, and such a
  page carries another process's placeholder, matching no header until the entry
  expires: set a key when you use a disk cache or several instances. A key that is
  short, a named variable that is empty or not hex, or a `SECURE_KEY` that is not
  hex, stops the application from starting; keep it secret.
- With no policy, or one without `{nonce}`, there is no nonce to name: the `nonce`
  attribute is removed, as in a static build, and the page stays cacheable with a
  stable `ETag`.
- A static export writes pages this middleware never serves, and a file cannot
  carry a per-response nonce. So in a static build the `nonce` attribute is
  removed and the script kept, with one warning; allow those inline scripts in
  the host's own policy (since v0.1.2 — before, the placeholder was left in the
  exported HTML).
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
    "skip": ["/_collage/", "/healthz", "/readyz", "/api/webhooks"]
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.8 was the last for v0.60.0). With no options, every form and action — every
  method but `GET`, `HEAD` and `OPTIONS` — is limited to a burst of ten, then one
  request every two seconds.
- The first rule a request matches counts it, so put narrow rules first; each rule
  has its own buckets. `skip` (by default `/_collage/`, `/healthz` and `/readyz`)
  and a rule's `pathPrefix` match whole segments, with or without a trailing slash:
  `/login` is not `/login-help`. Responses carry `RateLimit-Limit`,
  `RateLimit-Remaining` and `RateLimit-Reset`.
- A client is its IP address, or its `/64` for IPv6. Behind a reverse proxy list
  the proxies in collage's `Server.TrustedProxies`, and the client is
  `collage.ClientIP`. The plugin's own `trustProxy` and `trustedProxies` are gone
  since v0.2.0, and an old configuration naming them stops at start. `KeyFunc`, in
  Go, keys by something else.
- Buckets are kept in memory, so limits are per process.
- Since v0.1.7 a static build's header capture is not limited and gets no
  `RateLimit` headers.

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
    "skip": ["/healthz", "/readyz", "/_collage/"],
    "disabled": false
  }
}
```

- v0.1.9 needs collage v0.61.0 or later (v0.1.8 was the last for v0.60.0). The application does not start with no users,
  unless `disabled` is set. `skip` defaults to `/healthz`, `/readyz` and `/_collage/`.
- A password is written in plain text, as `sha256:` and its hex, or as a bcrypt
  hash. `COLLAGE_BASICAUTH_USERS` adds users from the environment, keeping secrets
  out of files.
- `disabled`, or `COLLAGE_BASICAUTH_DISABLED=true`, opens one deployment of the
  same binary — the production one.
- Every authenticated response has `public` replaced by `private` and carries
  `Vary: Authorization`, so a CDN in front does not hand a page to the next reader
  without asking. Serve it over HTTPS.
- A static build's header capture is let through, so a deployed page does not carry
  a `401`'s headers. A build that writes files under a protected path warns
  `basicauth-exported`: a static host serves them without a password, so protect
  them at the host or leave them out of the build.

#### elagoht/oauth

[github.com/Elagoht/collage-oauth](https://github.com/Elagoht/collage-oauth)
signs a reader in with an OpenID Connect provider (Google, Microsoft, GitLab or any
other with a discovery document) and, when you ask, keeps the provider's API tokens
so the application can call the provider for the reader. The application keeps its
own users: you answer with your user ID, and the plugin signs the reader in through
elagoht/session, so `session.RequireUser` guards work unchanged.

```go
import "github.com/Elagoht/collage-oauth"

Plugins: []collage.Plugin{
	session.New(session.Options{}),
	oauth.New(oauth.Options{
		Providers: []oauth.Provider{{
			Name:            "google",
			Preset:          "google",
			ClientID:        "1234-abcd.apps.googleusercontent.com",
			ClientSecretEnv: "GOOGLE_CLIENT_SECRET",
		}},
		OnLogin: func(ctx context.Context, id oauth.Identity) (string, error) {
			return users.FindOrCreate(ctx, id.Provider, id.Subject, id.Name)
		},
	}),
},
```

```html
<a href="{{oauthLogin "google"}}">Sign in with Google</a>
```

```json
{
  "elagoht/oauth": {
    "providers": [
      { "name": "google", "preset": "google", "clientID": "1234-abcd.apps.googleusercontent.com",
        "clientSecretEnv": "GOOGLE_CLIENT_SECRET", "offline": true }
    ],
    "afterLogin": "/account",
    "errorPath": "/sign-in-failed"
  }
}
```

- v0.2.0 needs collage v0.61.1 or later (v0.1.3 was the last for v0.60.0) and elagoht/session
  v0.3.0 or later, registered too: without it the login routes answer `500`.
  `collage plugin add elagoht/oauth` adds elagoht/session first. `OnLogin` and
  `Store` are code, so set them in Go. The secret and the key stay in the
  environment: the file names the variables, with `clientSecretEnv`, `keyEnv` and
  `previousKeysEnv` (since v0.2.0; `key`, `previousKeys` and `Options.KeyHex` are
  gone, and an old configuration stops at start).
- **Since v0.3.0** the key is read from `OAUTH_KEY`, and the previous keys from
  `OAUTH_PREVIOUS_KEYS`, when neither `keyEnv`/`previousKeysEnv` nor the Go
  options set them: put `OAUTH_KEY=` (`openssl rand -hex 32`) in
  `.env.development` and where production runs when `Store` keeps tokens. These
  may be empty; a key in them that is not hex stops the application.
- The routes are `{prefix}/{name}/login` and `{prefix}/{name}/callback` (prefix
  `/auth`), `GET` only. Register `{origin}/auth/google/callback` with the provider:
  the origin is `Config.BaseURL`, or what an origin resolver such as elagoht/tenant
  says for the request's host, so a multi-tenant site registers each host's.
- An `issuer` must be `https` (plain `http` only to `localhost` or a loopback IP),
  or the site does not start; a discovery document naming an endpoint that is not
  ends the sign-in with `unavailable`. Token and revocation calls never follow a
  redirect. Keep elagoht/session's `sameSite` at `"lax"`, its default: with
  `"strict"` the cookie misses the provider's return and every sign-in fails.
- The `microsoft` preset alone accepts any Entra ID tenant and any personal
  Microsoft account. Set `issuer` to one tenant's
  (`https://login.microsoftonline.com/<tenant-id>/v2.0`) to admit only it: a
  configured issuer is matched exactly. Microsoft sends no `email_verified`, so
  `EmailVerified` is false for it. `offline` asks a provider without a preset for
  the `offline_access` scope.
- A sign-in uses `state`, a `nonce` and PKCE, checks the id_token's claims, gives
  the session a new ID, and sends the reader to `next`, checked with
  `collage.SafeRedirect` and written as checked. A failure leaves no one signed in: `errorPath` gets
  `?error=<code>` (`state`, `expired`, `denied`, `rejected`, `exchange`, `token`,
  `unavailable`), or the site's status page is shown.
- **Key accounts on `Provider` + `Subject`**, and trust the e-mail only when
  `EmailVerified` is true: matching an account by an unverified address lets one
  reader take over another's.
- With a `Store` and a key (at least 32 bytes: `Options.Key` from Go, or hex in the
  variable `keyEnv` names) the tokens are kept sealed with
  AES-256-GCM, and `plug.Client(ctx, userID, "google")` returns an `*http.Client`
  that sends the user's token and refreshes it. `errors.Is(err, oauth.ErrNotLinked)`
  means send the reader through sign-in again; `Revoke` unlinks. To rotate the key,
  move the old one to `PreviousKeys` (or the variable `previousKeysEnv` names,
  comma-separated hex). The token goes only
  to the host it was first sent to, or its subdomains.
- Limits: no GitHub (not OpenID Connect), no account linking, and the id_token's
  signature is not checked (it comes from the token endpoint over TLS, which OpenID
  Connect Core allows). Refresh coalescing is per process. A guard's or an action's
  redirect stays unchecked: only `next` is.

#### elagoht/fail2ban

[github.com/Elagoht/collage-fail2ban](https://github.com/Elagoht/collage-fail2ban)
bans clients that probe or brute-force a site. It counts strikes per client in named
jails, and a client that reaches a jail's limit gets a plain `403` for the length of
the ban.

```go
import "github.com/Elagoht/collage-fail2ban"

bans := make(chan fail2ban.Ban, 100) // read by a goroutine of yours that batches alerts

f2b := fail2ban.New(fail2ban.Options{
	OnBan: func(b fail2ban.Ban) {
		select {
		case bans <- b:
		default: // full: drop it rather than block the request
		}
	},
})

Plugins: []collage.Plugin{f2b /* , the rest */},
```

```json
{
  "elagoht/fail2ban": {
    "jails": {
      "probe": { "banTime": "6h" },
      "notfound": { "disabled": true }
    },
    "allow": ["203.0.113.7", "10.0.0.0/8"]
  }
}
```

- v0.2.1 needs collage v0.61.4 or later (v0.1.8 was the last for v0.60.0), for `collage.Duration`, `Server.TrustedProxies`, `collage.ClientIP`
  and a `RequestHook` that sees requests collage rejects before routing. List it
  first in `Config.Plugins`, after `elagoht/health` when that is used, so its ban
  check runs before the other plugins' middleware and health's probes are not
  refused by a ban.
- When the client cannot be told from the forwarded chain (an entry that is not an
  address) and the connection's own address, `RemoteAddr`, is used instead, the
  plugin logs it once, at Warn (since v0.2.1). Limiting is unchanged.
- **Behind a reverse proxy or a CDN, set `Server.TrustedProxies`.** Without it the
  proxy is the client: one scanner gets your proxy banned, and with it every
  visitor. List every hop, a CDN's published ranges included. If you forget, a
  request from a loopback or private address that carries `X-Forwarded-For` but
  whose `ClientIP` is still that address is not counted, and one Warn says
  `TrustedProxies` is probably missing. A client is an IPv4 address, or an IPv6
  `/64`.
- The `probe` jail (3 strikes in `10m`, banned for `1h`) strikes on a request for a
  probe path (`/.env`, `/.git/`, `/wp-login.php` and the like, and anywhere a
  segment starting with `.env` or a `.git` segment; add your own with `probePaths`) or one collage rejects
  early, with an encoded slash or a `.` / `..` segment. A probe path your site
  really serves as a page, document or action is not counted, unless it answers
  `404` or another error. The `notfound` jail
  (50 in `1m`, `10m`) strikes on any other `404`, except a mount's. A browser's
  subresource request (`Sec-Fetch-Dest` such as `image` or `script`) never strikes,
  nor does a request marked `Sec-Fetch-Site: cross-site` (a no-cors fetch, a hidden
  iframe, a link from another site). That keeps a page on another site from getting
  your readers banned, in browsers that send these headers; a page on a subdomain of
  yours, or user content on your own origin, using an iframe or a fetch still
  counts. A scanner forging these headers goes undetected. A jail of your
  own strikes only on `Report`; override any by name, or turn it off with
  `"disabled": true` (`"off"` until v0.2.0). Durations are strings — `"10m"`,
  `"1h30m"` — and a bare number such as `"banTime": 3600` stops the application,
  naming the key.
- A repeat ban doubles, up to `maxBanTime` (`24h`). A banned client gets a plain
  `403 Forbidden`, never the site's error page. Collage's own `404` or redirect for
  an encoded-slash or dirty-path request still reaches it.
- **The ban is middleware.** Since collage v0.67.0 a handler mounted with plain
  `app.Handle` runs none, so a banned client keeps full access to it, `/api/` login
  brute force included, though its probes there still count. Mount it
  `collage.WithMiddleware()` to cover it; see
  [No middleware, unless you ask](/docs/middleware-and-apis#no-middleware-unless-you-ask).
- Call `Report(r, "login")` on a failed login. `Forgive(r, "login")` clears the
  strikes, but is only safe when the success is for the account whose password was
  being guessed: forgiving any success lets an attacker with their own account guess
  without limit. Failures age out after `findTime` anyway. `allow` lists addresses
  and ranges never counted or banned. `Ban`, `Unban` and `Bans` are the manual API
  (`Ban` never shortens a ban); `OnBan` (Go only) is called after each ban, maybe
  thousands of times in a distributed scan, so never notify per ban synchronously.
- Bans are per process and in memory: instances do not share them and a restart
  clears them. In dev mode nothing happens unless `inDevelopment` is set. It reacts
  after strikes; it does not replace `elagoht/ratelimit`.

#### elagoht/consent

[github.com/Elagoht/collage-consent](https://github.com/Elagoht/collage-consent)
adds a cookie-consent banner, holds scripts and iframes back until their category
is granted, and gives the page a small JS API to read the visitor's choice. On the
paths you list, the server can read the choice too.

```go
import "github.com/Elagoht/collage-consent"

Plugins: []collage.Plugin{consent.New(consent.Options{})}, // configured from plugins-config.json
```

```json
{
  "elagoht/consent": {
    "version": 1,
    "categories": [
      { "name": "necessary", "required": true },
      { "name": "analytics" },
      { "name": "media" }
    ],
    "text": {
      "title": "Cookies", "body": "We use cookies to measure visits and to show embedded media.",
      "accept": "Accept all", "reject": "Reject all", "save": "Save choices", "settings": "Choose",
      "placeholder": "This content loads from {host}.", "allow": "Allow {category}",
      "categories": { "necessary": "Necessary", "analytics": "Analytics", "media": "Embedded media" }
    },
    "locales": {
      "tr": { "title": "Çerezler", "accept": "Tümünü kabul et", "reject": "Tümünü reddet" }
    },
    "policyURL": "/privacy",
    "maxAgeDays": 180,
    "serverPaths": []
  }
}
```

- v0.2.1 needs collage v0.61.4 or later (v0.2.0 needed v0.61.0; v0.1.3 was the last for v0.60.0) and `{{hoist "head"}}` in the layout.
  Without that marker nothing appears and nothing gated ever runs.
- Consent is decided in the browser. Every visitor gets the same HTML, so the page
  cache works as before. `version` asks every visitor again when you bump it. A
  `required` category is always granted and never written to the cookie. `text` is
  the default locale's wording and must label every category; `locales` holds the
  others, which fall back to `text` key by key (since v0.2.0; the old
  `text: {"en": {…}}` shape stops the application, naming the locale key).
- Mark up what needs consent so it is inert as written. A granted category turns a
  gated script into a real one, in document order, and gives a gated iframe its
  `src`. Until then the iframe has a placeholder button, and markup added later by
  a fragment swap or a router is gated too:

  ```html
  <script type="text/plain" data-consent="analytics" src="https://example.com/a.js"></script>
  <iframe data-consent="media" data-src="https://www.youtube-nocookie.com/embed/…" title="…"></iframe>
  ```

  A gated script cannot be a module. Taking a category back after its content has
  run reloads the page.
- The JS API is `collageConsent.get()` (the granted categories, sorted),
  `collageConsent.set({ analytics: true })`, `collageConsent.open()`, and a
  `collage:consent` event on `document` after every save. An element with
  `data-consent-open` reopens the banner. The banner is a `<dialog>` themed with
  `--consent-*` custom properties.
- Server side, list the paths in `serverPaths` (whole segments: `/shop` covers
  `/shop/cart`, not `/shopping`), then call `consent.Granted(rc, "analytics")` in a
  data handler there. The page cache keeps one entry per choice, not per cookie.
  Outside `serverPaths`, in a static export and in a build's capture, `Granted`
  answers false (a required category: true). Every response on those paths carries
  `Vary: Cookie`, which most CDNs will not cache, so list only pages whose HTML must
  differ. Never list a stream path: collage-live's `/_live/` and `/_collage/` are
  never varied.
- The choice is kept in the first-party cookie `collage_consent`, readable by the
  script, so never set it from the server or as `HttpOnly`.
- GPC (`navigator.globalPrivacyControl`) leaves every optional box unticked, and
  Accept all still grants everything. Do Not Track is not read.
- Content-Security-Policy: the script tag has no nonce, so `script-src` needs
  `'self'`; a policy that is nonce-only or relies on `'strict-dynamic'` blocks it. The
  banner's stylesheet is built by the script, so `style-src` must allow its hash,
  which the plugin's README keeps current (`'unsafe-inline'` also works). A gated
  inline script needs its own `nonce` or a `'sha256-…'` of its text in `script-src`.
- With collage-live, a pushed fragment passes the middleware for its own path
  since collage v0.78.0: on a path `serverPaths` covers, it reads the reader's own
  choice and is rendered for that reader, never shared (v0.2.1 tests it; nothing
  to change). A static export writes the script beside the pages and decides in the browser, so
  exported pages always render as "not granted" on the server.
- `Granted` reads the configuration of the App it renders for (since v0.2.0; before,
  the last App initialised), and is false without a render context, even for a
  required category. For analytics, set `consentCategory` on
  [elagoht/analytics](#elagohtanalytics).

### Live updates

Keeping parts of a page current in the browser, without a framework on the client.

#### elagoht/live

[github.com/Elagoht/collage-live](https://github.com/Elagoht/collage-live) keeps
parts of a page current in the browser. It serves a small client script that
refreshes the fragments a page opened with
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) — on an
interval, or when the server pushes a change over an event stream.

```go
import (
	client "github.com/Elagoht/collage-client"
	live "github.com/Elagoht/collage-live"
)

Plugins: []collage.Plugin{client.New(), live.New(live.Options{})},
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
  to include the client. v0.7.0 needs collage v0.71.0, and v0.78.0 or later for
  each push to meet the middleware; v0.6.0 needed collage v0.67.0. Both need
  [elagoht/client](#elagohtclient) v0.2.0: every answer goes into the page through
  collage-client's `collage.put`. `collage plugin add elagoht/live` adds client
  too; without it, `collage.New` fails. `{{liveClient}}` writes client's script
  ahead of its own, so the layout does not change, and one that also calls
  `{{clientjs}}` loads it once.
- A transport of another plugin, such as elagoht/websocket, calls `UseTransport`
  from its `Configure` as well as its `Init`; the last call wins (v0.5.2).
- `keepAlive` and `maxStreamAge` are duration strings (`collage.Duration` in
  `live.Options`): a number, a negative duration or a negative `maxFragments`
  stops the application, and `null` leaves the default. Since v0.5.0 `live.New()`
  and `live.NewWith(live.Config{…})` are one `live.New(live.Options{…})`.
- The page owns the container and the fragment owns what is inside it.
  `data-collage-interval` fetches on an interval, `data-collage-push` takes the
  fragment from the stream, `data-collage-swap="morph"` patches the DOM in place
  with collage-client's morph,
  and `data-collage-target` on a form submits it with `fetch` and puts the answer
  into an element — on a success, or on a `422`, which is how an action says a
  submission did not validate: it answers the form's fragment again, with the
  errors, and status 422 (see
  [Forms and actions](/docs/forms-and-actions#refreshing-it-from-the-browser)). Any
  other failure leaves the target as it was and marks it stale. Since v0.2.1 a
  form whose action redirects navigates in one request, with collage's
  [`Collage-Fetch`](/docs/forms-and-actions#failure-renders-success-redirects).
- **Since v0.6.0 the morph is idiomorph's**, which matches elements by id. The
  old one matched them by position, so an item added at the top of a list rebuilt
  every item after it, and a field being typed into lost its text and focus. A
  focused field still keeps its value, but a focused `<select>`'s options now take
  the server's. `collage:swap` comes from collage-client, with
  `detail: { url: null, mode, inner: true }`.
- The client sends the `ETag` it holds and leaves the DOM alone on a `304`, adds
  what the fragment hoisted to the head once by its key, stops in a hidden tab, and
  backs off when a request fails, marking the element `data-collage-stale`. A
  pushed copy carries the same ETag a poll would get, so one the client already
  holds is not sent again.
- **Holding an element still** (since v0.3.0). While the reader is in the middle
  of something, a drag, a menu held open, `collageLive.pause(el)` holds the
  element: polls, pushes and form answers keep arriving, but only the latest is
  kept, and `collageLive.resume(el)` puts it in. An element inside a fragment names
  that fragment, so the list a drag started from holds the board around it; pauses
  nest, and a held element carries `data-collage-paused`. `collageLive.put(el, html)`
  puts in an answer the page fetched itself, an action's fragment, the way a form's
  answer is put in; patching the DOM directly would leave the client believing it
  still shows what it last put in. The
  [README](https://github.com/Elagoht/collage-live#holding-an-element-still) has a
  drag-and-drop example.
- **Animating a swap** (since v0.4.0). A view transition has to begin before the
  DOM changes, and `collage:swap` fires after. `data-collage-transition` puts every
  answer in inside `document.startViewTransition` — pushes, polls, form answers and
  `put` alike — so cards that keep their `view-transition-name` from one answer to
  the next slide to their new places; without the API, or under
  `prefers-reduced-motion: reduce`, the answer goes in at once. For more control,
  `collage:before-swap` fires on the element just before it changes: a listener
  that calls `preventDefault` runs `e.detail.swap()` itself, inside its own
  transition, naming the elements that should move before and after. A late
  `swap()` after a newer answer does nothing, and one called while the element is
  held waits for the resume. The
  [README](https://github.com/Elagoht/collage-live#animating-a-swap) has the
  example.
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
  [`Shared()`](/docs/caching#a-page-that-declares-none), not `Static()`, so that
  collage renders it once per change rather than once per tab while its page stays
  dynamic; several fragments reading one measurement fetch it through
  `collage.Cached` with no tags, so that one fragment's tick does not re-render the
  others. A fragment that is not `Shared` is rendered for each connection, and
  `Cached`, with the data cache on, keeps what it fetches fetched once. The [README](https://github.com/Elagoht/collage-live#sampled-data-a-system-monitor)
  walks through a system monitor.
- **Every push is made for one connection** (since v0.7.0), with that
  connection's own request, and sent to it alone: under collage v0.78.0 it is a
  request to the fragment's URL, which the application's and every plugin's
  middleware answer, then the guards. A reader one of them now refuses — a session
  revoked on the server, changed credentials, a ban — is sent a stale message
  instead of the fragment. Before, a `Shared` render was made once and sent to
  every reader of its URL, so such a reader kept receiving one no guard covered.
  The render itself is still made once per change, which collage shares between
  the readers it admits; a rate limit counts each push as a request.
- **A form in a pushed fragment is accepted** (since v0.7.0). Every render after a
  connection's first used to carry a token its forgery cookie did not verify, so
  submitting a pushed form was refused until the connection reopened.
- A connection renders with the cookies it was opened with; with signed, stateless
  cookies, `maxStreamAge` closes it after that long, and the client reopens it at
  once with the cookies it holds then.
- The stream is served at `<prefix>stream/`. A compression middleware of your own
  should leave `text/event-stream` alone, or the stream arrives only when it ends.
  With `noStream` the client only polls, and nothing is pushed over any other
  transport either.

#### elagoht/websocket

[github.com/Elagoht/collage-websocket](https://github.com/Elagoht/collage-websocket)
carries collage-live's pushed fragments over a WebSocket instead of an event stream.

```go
import (
	client "github.com/Elagoht/collage-client"
	live "github.com/Elagoht/collage-live"
	"github.com/Elagoht/collage-websocket"
)

Plugins: []collage.Plugin{client.New(), websocket.New(websocket.Options{}), live.New(live.Options{})},
```

- `collage plugin add elagoht/websocket` adds collage-live and collage-client too,
  and puts websocket before live. The plugin finds the application's live plugin
  by itself, wherever it is listed; without one, `collage.New` fails. v0.4.2 needs
  collage v0.67.0, collage-live v0.6.0 and collage-client v0.2.0 or later; `websocket.New(lv)`, `New(nil)` and `NewWith` are
  gone.
- Nothing else changes: the layout still includes `{{liveClient}}`, which now tells
  the client to connect here, and elements still say `data-collage-push`.
  collage-live stops serving its event stream. The WebSocket is opened from the
  same shared worker, so every tab still shares one connection.
- **The push settings are live's.** The WebSocket is served at live's `prefix`
  plus `ws/` (`/_live/ws/` by default), pinged every live `keepAlive`, and closed
  after live's `maxStreamAge`; with live's `noStream` a connection is refused with
  the site's own not-found page (404; a plain-text 404 before v0.4.1). `Path` and `Ping` are gone, and `path` or `ping` in its configuration is an
  unknown key.
- **Usually you do not need it.** Pushing fragments is one-way, which is what an
  event stream is for: it needs no dependency and reconnects by itself. This plugin
  is for deployments where streams are what breaks — a proxy that buffers them, a
  platform that limits them. It is the one piece of the live stack with a
  dependency, `github.com/coder/websocket`.
- A connection carries the reader's cookies, so by default only a page from the
  site itself may open one; `originPatterns` names other hosts whose pages may:

```json
{
  "elagoht/live": { "prefix": "/push/", "keepAlive": "20s", "maxStreamAge": "1h" },
  "elagoht/websocket": { "originPatterns": ["admin.example.com"] }
}
```

#### elagoht/client

[github.com/Elagoht/collage-client](https://github.com/Elagoht/collage-client)
serves a tiny client script for the swaps a page does itself: element getters, and
a fragment put in place, replacing the element or morphing it — fetched from its
own URL, or handed over as markup. Everything else is left to plain JavaScript.
elagoht/live puts its answers in with it too.

```go
import client "github.com/Elagoht/collage-client"

Plugins: []collage.Plugin{client.New()},
```

```html
<head>
  {{clientjs}}
</head>

<!-- the fragment's own template: its root is what is swapped -->
<ul id="items" data-src="{{fragmentURL "list" "items"}}">
  {{range .}}<li>{{.Name}}</li>{{end}}
</ul>
```

```js
const names = collage.$$("li", collage.$("#items")).map((li) => li.textContent);

collage.replace("#items").catch(() => {}); // fetch data-src, put it in as outerHTML would
collage.morph("#items").catch(() => {});   // the same fetch, the DOM patched in place
collage.put("#items", html, { morph: true }); // markup in hand, no fetch

document.addEventListener("collage:swap", (event) => {
  console.log(event.target.id, event.detail.mode, event.detail.url);
});
```

- v0.2.0 needs collage v0.67.0 or later, and must go in `Config.Plugins`: it adds
  `{{clientjs}}`, which the master layout calls once, in its `<head>`. There is
  nothing to configure.
- **The script defines `window.collage` and nothing else.** `collage.$` is
  `querySelector` and `collage.$$` is `querySelectorAll` returned as an Array, both
  on `document` unless a root is given as the second argument.
- **`collage.replace(target, url?)`** fetches a fragment and puts it in place of
  `target`, a selector or an Element, as `outerHTML` would. `url` defaults to the
  element's `data-src`, where a page opened with
  [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) puts
  `{{fragmentURL}}`. The answer takes the place of the whole element, so the
  element is the fragment's own root, its template's outermost tag. It resolves to
  the new element.
- **`collage.morph(target, url?)`** is the same fetch, with the DOM patched in place
  by [idiomorph](https://github.com/bigskysoftware/idiomorph): elements that match
  are kept, so the field being typed into keeps its focus, its caret and what is
  typed in it, and a kept element keeps its scroll position. Another field of the
  morphed element takes the server's value, so morph a part of the page that does
  not hold a half-filled form, or `replace` it.
- **`collage.put(target, content, { morph?, inner? })`** (since v0.2.0) is the same
  swap without the fetch: markup, an Element or a DocumentFragment already in hand,
  such as what a form's `fetch` answered. It changes the page before it returns,
  returns the element now in the page, and throws where the others reject. Without
  `inner` the content must be one element; with it, the element stays and its
  children become the content, replaced or morphed. It checks no origin, as
  `innerHTML` does not, so never pass it what a visitor wrote. It aborts a
  `replace` or `morph` in flight on the element.
- After any of them, a **`collage:swap`** event bubbles from the element now in the
  page, with `detail: { url, mode, inner }`: `mode` `"replace"` or `"morph"`, and
  `url` `null` for `put`.
- **The request** is a `GET` with `Collage-Fetch: 1` (`collage.FetchHeader`) and the
  page's cookies, as fragment paths and actions expect. The answer is inserted as
  trusted markup, so `url` must be on the page's own origin: another origin is
  refused before the fetch, and so is an answer a redirect brought from one. Never
  build `url` or `data-src` from visitor input.
- **The newest call on an element wins.** A `replace` or `morph` on an element that
  already has one in flight aborts the older call, which rejects as superseded, so
  a slow answer never lands on top of a newer one.
- **A failure leaves the page as it was.** A target not in the page, no `url` and
  no `data-src`, another origin, a network error, an answer that is not 2xx or not
  a single element (a plain-text error body, a whole document, two elements side by
  side), an element taken out of the page while it was fetched, or a newer call:
  each fires no event and rejects with an `Error` that says which, such as
  `collage: morph /list/items: answered 500`. A polling tick that may fail quietly
  ends in `.catch(() => {})`. Scripts inside a fetched fragment do not run, as with
  `outerHTML`.
- **With a Content-Security-Policy nonce** from elagoht/secure, pass it:
  `{{clientjs cspNonce}}`. A policy that allows `'self'` loads the script without
  it. Under a nonce-only or `'strict-dynamic'` policy, a layout that forgot it still
  works under `collage dev`, where the policy only reports, and stops loading the
  script in production.
- The script is served at `/_client/client.<hash>.js`, cached for a year as
  `immutable`, and loaded with `defer`. A page that includes it twice, as
  `{{clientjs}}` and `{{liveClient}}` together do, runs it once. It is about 7 KB
  gzipped, idiomorph 0.8.0
  (0BSD) included, and depends on nothing.
- It neither polls, nor posts forms, nor takes pushes: for those, use
  [elagoht/live](#elagohtlive). A form's post stays your own `fetch`, its answer put in with `put`, as
  [Update part of a page](/docs/update-part-of-a-page#swap-it-in-with-a-few-lines-of-javascript)
  shows.

### Assets and delivery

What the bytes a browser receives look like and how they get there: minified,
resized, bundled, compressed, purged from a CDN, kept for offline reading, and
deployed to a static host with their headers and redirects.

#### elagoht/minimizer

[github.com/Elagoht/collage-minimizer](https://github.com/Elagoht/collage-minimizer)
strips whitespace and comments from rendered pages, from documents such as JSON
endpoints, and from the files your mounts serve.

```go
import minimizer "github.com/Elagoht/collage-minimizer"

Plugins: []collage.Plugin{minimizer.New(minimizer.Options{})},
```

```json
{
  "elagoht/minimizer": { "noCSS": false, "js": true }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.11 was the last for v0.60.0), and must go in `Config.Plugins`: it wraps the
  mounted filesystems, which happens while the application is built.
- The zero `Options` minifies HTML, JSON and CSS; `noHTML`, `noJSON` and `noCSS`
  turn one off. JavaScript is off by default; turn it on with `{"js": true}`.
  `minimizer.New(minimizer.Options{...})` sets a starting point that the JSON
  section is decoded over.
- **Changed in v0.2.0:** `New()` and `NewWith(Config{…})` are `New(Options{…})`;
  `"html": false`, `"json": false` and `"css": false` are `"noHTML": true`,
  `"noJSON": true` and `"noCSS": true`, and `"html": true` and the like are simply
  left out. An old key stops the application at start.
- Each language is covered wherever it appears (since v0.1.5): JavaScript a mounted
  script and an inline `<script>` of JavaScript or a module, CSS a stylesheet and
  a `<style>`, JSON a JSON document and a `<script>` holding JSON —
  `application/ld+json`, `importmap`. A script of a type it does not know, such as
  `text/template`, stays verbatim.
- It is a scanner, not a parser, and removes only what cannot carry meaning:
  `<pre>` and `<textarea>` are kept verbatim, CSS strings are untouched,
  JavaScript keeps a line break wherever a semicolon may stand for it, and
  invalid JSON is returned as it was. A JavaScript line is joined only after a
  token no statement can end with — `{ ( [ , ;` or an operator such as `=`, `&&`
  or `=>` — and kept after a name, a closing bracket, a string or `++` (since
  v0.1.6; before, every newline was kept).
- Mounted files are minified by wrapping the filesystem rather than the response,
  so `Range` requests against a mount keep returning the right bytes.

#### elagoht/opti-image

[github.com/Elagoht/collage-opti-image](https://github.com/Elagoht/collage-opti-image)
rewrites every `<img>` that declares a pixel `width` and `height` to a resized
copy it serves itself, from its own mount.

```go
import optiimage "github.com/Elagoht/collage-opti-image"

Plugins: []collage.Plugin{optiimage.New(optiimage.Options{})},
```

```json
{
  "elagoht/opti-image": {
    "allowedOrigins": ["https://images.example.com"],
    "webp": "auto"
  }
}
```

- v0.4.1 needs collage v0.61.4 or later (v0.3.7 was the last for v0.60.0), and must go in `Config.Plugins`.
- An allowed origin that spells its scheme's default port, `https://cms.example.com:443`,
  matches the same host's images whether or not the image's URL writes the port (v0.4.1);
  it was accepted at start but matched nothing.
- **Changed in v0.4.0:** `New()` and `NewWith(Config{…})` are `New(Options{…})`;
  `allowedOrigins` is a list of origin strings (`"https://cms.example.com"`), not
  `{scheme, host}` objects, each checked as `collage.ParseOrigin` checks one;
  `fetchTimeout` and `originMaxAge` are strings such as `"10s"`, and a bare number
  stops the application; a negative size, pixel count or quality is an error.
- **An empty `allowedOrigins`, with no `Files`, disables it.** It never fetches
  from a host you did not list, and the scheme is part of the origin.
- **The site's own images come from its files** (since v0.3.0):
  `optiimage.New(optiimage.Options{Files: map[string]fs.FS{"/static/": staticFS}})`
  makes `<img src="/static/avatar.png" width="256" height="256">` a resized copy
  read from `staticFS`, with no request made — so the site does not fetch from its
  own public address, and a test or an export needs no network. The file's content
  is part of the image's name, so a changed file is a new name. `Files` has no
  JSON form.
- **An image from an origin is not `immutable`** (since v0.3.5). Its name is made from
  the recipe (URL, size, format), not from the content, so the origin can change the
  file under the same name. It is served with `public, max-age=<originMaxAge>`;
  `originMaxAge` is a duration, `"24h"` by default, at least one second. An image of
  the site's own `Files`, whose name carries its content's digest, keeps
  `public, max-age=31536000, immutable`.
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

`optiimage.New(optiimage.Options{...})` sets a starting configuration in Go,
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

- v0.2.0 needs collage v0.61.0 or later (v0.1.6 was the last for v0.60.0), and must go in `Config.Plugins`: it adds
  `{{bundle}}`.
- `minify` and `sourcemap` are `""` (left out: on outside development only for
  `minify`), `"always"` or `"never"`. **Changed in v0.2.0:** they were booleans, and
  `true` or `false` now stops the application at start; `prefix` no longer needs its
  trailing `/`.
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

```sh
collage plugin add elagoht/favicon
```

Since v0.3.0 that is all it needs. The icons are made from the site's own
`static/icon.png` (and `static/icon.svg`, if there is one), and until there is one
they are collage's. The settings are for the manifest and the colours:

```json
{
  "elagoht/favicon": {
    "name": "The blog",
    "themeColor": "#0f172a",
    "backgroundColor": "#ffffff"
  }
}
```

- v0.3.0 needs collage v0.68.0 or later (v0.2.0 was the last for v0.61.0–v0.67.x,
  v0.1.7 for v0.60.0). The source is a square PNG, JPEG or GIF, ideally 512
  pixels or larger.
- **The source defaults to `/static/icon.png`** (since v0.3.0), and the SVG to
  `/static/icon.svg` when there is one: URLs of the site's own, read through the
  mount that serves them with
  [`Host.OpenMounted`](/docs/writing-plugins#contributing-documents-and-mounts) — from disk in development, from the
  copy the binary embeds once built. `source` and `svg` name another file: one
  starting with `/` is a URL, read the same way; any other path is read from disk,
  or from `FS`. `SourceBytes` and `SVGBytes` give the bytes in Go. The start page
  key is `startURL`.
- **A missing source never stops the application** (since v0.3.0). With no file at
  the default URL it serves collage's icons and logs where to put one; a source
  that was named but is missing, cannot be decoded or is not square is a warning,
  with the same fallback. Only a colour that is not one stops it.
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

- v0.2.0 needs collage v0.61.0 or later (v0.1.10 was the last for v0.60.0). **Register it before any plugin that rewrites
  response bodies**: the first plugin registered is the outermost middleware.
  elagoht/secure and elagoht/honeypot no longer need this (secure v0.2.0, honeypot
  v0.4.0, collage v0.43.0): they rewrite through `PersonaliseHook`, inside every
  middleware. `elagoht/health` may be listed before or after compress.
- Text types of at least `minSize` bytes are compressed with the best encoding the
  request accepts. `text/event-stream`, a WebSocket and a `Range` request are left
  alone.
- A compressed body is kept per ETag, so a page collage serves from its cache is
  compressed once per encoding, not once per reader. The ETag gains the encoding,
  and a conditional request still gets its `304`. A response whose `Cache-Control`
  has `private` or `no-store`, such as a page carrying elagoht/secure's nonce, is
  still compressed but never kept (v0.1.3): it changes every time, so what was
  kept would never be reused.
- **`skipPrivate`** (since v0.1.8, `false` by default) sends a response whose
  `Cache-Control` carries `private` or `no-store` uncompressed, against BREACH: when a
  response holds a secret, such as a forgery token, next to input the page reflects,
  its compressed size can leak the secret a character at a time. Nothing changes unless
  it is turned on. The README's BREACH section lists the other mitigations.
- A static build writes a `.br` and a `.gz` beside every compressible file, for a
  host that serves precompressed files; `noPrecompress` turns it off.
- `noCache` (since v0.2.0) keeps no compressed bodies; it replaces a negative
  `cacheBytes`, which now stops the application.

#### elagoht/cdnpurge

[github.com/Elagoht/collage-cdnpurge](https://github.com/Elagoht/collage-cdnpurge)
purges a CDN's copies of the pages collage invalidates: Cloudflare by zone, or any
service that takes a list of URLs through a webhook.

```go
import "github.com/Elagoht/collage-cdnpurge"

Plugins: []collage.Plugin{cdnpurge.New(cdnpurge.Options{
	Cloudflare: &cdnpurge.Cloudflare{
		ZoneID:      os.Getenv("CF_ZONE_ID"),
		APITokenEnv: "CF_API_TOKEN",
	},
})},
```

```json
{
  "elagoht/cdnpurge": {
    "webhook": { "url": "https://purge.example.com/", "tokenEnv": "PURGE_TOKEN" },
    "window": "2s",
    "retries": 4
  }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.5 was the last for v0.60.0). At least one of `cloudflare` and `webhook` is
  required, and an origin: `Config.BaseURL` or an origin resolver.
- **Changed in v0.3.0:** `baseURL` is gone (set `Config.BaseURL`); `force` is
  `inDevelopment`; `cloudflare.apiToken` and `webhook.token` are `apiTokenEnv` and
  `tokenEnv`, naming the environment variable that holds the token (`APIToken` and
  `Token` take it from Go); durations are strings, and a negative `retries` is an
  error. An old key stops the application at start.
- It purges exactly the [paths an invalidation reached](/docs/caching#invalidating-by-path),
  under `Config.BaseURL`, and purges each entry under its own host's origin, as an
  `OriginResolver` plugin names it; two hosts of one origin purge the URL once. A page the origin had not cached is not named, so make sure what
  the CDN caches, collage caches too.
- Purges are batched for `window`, sent off the goroutine that invalidated, retried
  on a `429` or `5xx`, and flushed on shutdown. A development server purges nothing
  unless `inDevelopment` is set.
- The token never goes in `plugins-config.json`, only its variable's name.

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

- v0.2.1 needs collage v0.61.0 or later (v0.1.6 was the last for v0.60.0), and must go in `Config.Plugins`: it adds
  `{{offlineScript}}`, which installs the worker served at `/sw.js`.
- **Changed in v0.2.0:** `version` is `buildID`, `inDev` is `inDevelopment`, and a
  negative `maxPages` or `maxAssets` (keep everything) is an error — write a large
  number. An `assets` prefix covers whole path segments, with or without its
  trailing `/`.
- Pages are fetched network-first and kept; static files under `assets` are served
  stale-while-revalidate; a page neither reachable nor kept gets the `fallback`
  page. A response marked `no-store` or `private` is never kept, nor precached.
- **Private pages are not kept offline** (since v0.2.1). A guarded page
  (`WithGuard`) goes out `private, no-cache`, and basicauth marks its pages
  `private`: a browser may keep such a page because it asks the server again
  before showing it, but the worker answers offline without asking anybody. Up to
  v0.2.0 it kept them and showed them offline to whoever used the device, with no
  guard asked. The cache version changes with the plugin's, so the first visit
  after the upgrade drops what the old worker kept. A page of your own that holds
  one reader's data says so with `Cache-Control: private` or `no-store`.
- The worker's caches are named after a version that changes with each deployment,
  so a new build replaces what the old one kept. The build is `buildID` when set,
  and otherwise collage's `Host.BuildID` — `Config.Cache.Version`, or a
  fingerprint of the executable.
- In development `/sw.js` unregisters itself, unless `inDevelopment` is set. A static build writes `sw.js`; keep a
  CDN from holding it long.

#### elagoht/deploy

[github.com/Elagoht/collage-deploy](https://github.com/Elagoht/collage-deploy)
writes the configuration files a static host reads from a static build's captured
headers and redirects, so a site's redirects, caching and security headers survive
the export.

```go
import "github.com/Elagoht/collage-deploy"

Plugins: []collage.Plugin{deploy.New(deploy.Options{Target: "netlify"})},
```

```json
{
  "elagoht/deploy": {
    "target": "netlify"
  }
}
```

- v0.2.0 needs collage v0.61.2 or later (v0.1.2 was the last for v0.60.0); its
  constructor is `deploy.New(deploy.Options{…})`, and the JSON is unchanged. `target` is `netlify`, `cloudflare`,
  `vercel` or `github-pages`; another value stops the application from starting.
  Left empty, it writes nothing, removes what the last export wrote, and warns.
- `netlify` and `cloudflare` get `_headers` and `_redirects`, `vercel` a
  `vercel.json`, and `github-pages` one meta-refresh page per literal redirect and
  a `.nojekyll`, but no headers.
- What a host cannot carry is a warning, never a rule that would redirect
  differently from the application: a 410 on every host; 307 and 308 written as 302
  and 301 on Netlify; Cloudflare's rule limits; patterned redirects and every
  header on GitHub Pages. Vercel reads `vercel.json` only when the output directory
  is deployed as the project (`vercel deploy dist`).
- Headers are compacted into the fewest rules: what every file shares goes under
  `/*`, what a directory shares under `/dir/*`, the rest at each file's path.
  `Content-Type` never goes in a wildcard rule, and a `Content-Type` the file's
  extension implies is left to the host.
- It records what it wrote in `.collage-deploy.json` and replaces those files on
  the next export into the same directory. It never overwrites a `_headers`,
  `_redirects` or `vercel.json` it did not write: that is an error, and nothing is
  written. So is a control character in a redirect or a header.
- Its redirects come from pages, documents and any
  [`RedirectSource`](/docs/writing-plugins#redirectsource) plugin, such as
  [elagoht/redirects](#elagohtredirects); its headers from the
  [capture](/docs/static-export#headers-and-redirects), so build in production
  mode with the real `BaseURL`.

### Operations and development

Checking what a site renders, seeing a render in development, and watching a
running site: access logs, metrics, traces, analytics, error reports and health
probes, and the scheduled jobs and background queues it runs.

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

- v0.2.0 needs collage v0.61.0 or later (v0.1.6 was the last for v0.60.0). It
  has no `inProduction` key any more: findings on a production server went
  nowhere, so read them in development or in a static build.
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
- `ignoreLinks` prefixes cover whole segments, with or without the trailing `/`,
  and must begin with `/`. A negative `titleMax`, `descriptionMax` or
  `pageBudget` stops the application; 0 is the default (60, 160, no budget).

#### elagoht/devtoolbar

[github.com/Elagoht/collage-devtoolbar](https://github.com/Elagoht/collage-devtoolbar)
shows a small panel on every page in development: which page
rendered, in which locale, with what status, how long the render and each of its
fragments took, which fragments failed, the dependency tags the render depended
on, its `Cache-Control` and `ETag`, its size, and how many findings the checking
plugins reported.

```go
import "github.com/Elagoht/collage-devtoolbar"

Plugins: []collage.Plugin{
	htmlcheck.New(htmlcheck.Options{}),
	devtoolbar.New(devtoolbar.Options{}), // last
},
```

- v0.4.0 needs collage v0.61.0 or later (v0.3.2 was the last for v0.60.0).
- **Local clients only.** The panel is drawn only for a client on the developer's
  machine or a private network (loopback, RFC 1918, IPv6 unique-local,
  link-local), as `collage.ClientIP` names it, so behind a proxy listed in
  `Server.TrustedProxies` it is the forwarded client. Any other client gets the
  response untouched, and the first one logs a warning. Since v0.4.2 a local peer
  that forwards a non-local client — a proxy missing from `Server.TrustedProxies` —
  gets no panel either, and the first one logs a warning naming that setting.
- A tailnet or carrier-grade NAT address (`100.64.0.0/10`) is not a private
  network, so a phone reaching the machine over Tailscale gets no panel. To show
  it to remote clients, turn the check off with
  `devtoolbar.New(devtoolbar.Options{AllowRemote: true})`, or in the
  plugin configuration:

```json
{ "elagoht/devtoolbar": { "allowRemote": true } }
```

- **Position.** The panel sits in the bottom-left corner. Since v0.5.0 `position`
  puts it in another corner or in the middle of the top or bottom edge:
  `top-left`, `top`, `top-right`, `bottom-right`, `bottom` or `bottom-left`. In Go
  it is `devtoolbar.Options{Position: devtoolbar.PositionTopRight}`; any other
  value stops the application at start-up.

```json
{ "elagoht/devtoolbar": { "position": "top-right" } }
```

- **Register it last**: the findings it counts are those of the plugins that ran
  before it.
- On a server without `DevMode` and in a static build it does nothing at all. The
  panel is added after the page cache, so what collage caches never carries it.
- Since v0.2.3 a page written without a `Content-Type` gets the panel too: the
  type is sniffed from the first bytes, as net/http does.

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
    "skip": ["/_collage/", "/healthz", "/readyz", "/static/"],
    "sample": 0.25,
    "requestIDHeader": "X-Request-ID"
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.13 was the last for v0.60.0).
  `/_collage/`, and `/healthz` and `/readyz`, the paths of `elagoht/health`, are
  skipped by default. A `skip` prefix covers whole segments, with or without the
  trailing `/`: `/healthz` does not cover `/healthz-report`.
- The line has the method, path without its query, status, bytes, duration, client
  address, user agent, referer and request id, through the application's logger or
  `Options.Logger`; a `5xx` is logged at `ERROR`.
- An `X-Request-ID` that looks like an id is kept, otherwise a new one is made; it
  is sent back and read with `accesslog.RequestID(ctx)`.
- `sample` logs a fraction of `2xx` responses, never of the rest. The client's
  address is `collage.ClientIP`, so behind a proxy list it in
  `Server.TrustedProxies`; the `trustProxy` and `trustedProxies` keys are gone, and
  `X-Real-IP` is never read.
- **The referer is logged without its query** (since v0.1.11): its scheme, host and path
  only, with no query, fragment or user info, where a reset token or a session id can
  travel. One that does not parse is logged as `""`. `fullReferer: true` logs the header
  exactly as before.

#### elagoht/prometheus

[github.com/Elagoht/collage-prometheus](https://github.com/Elagoht/collage-prometheus)
exports the framework's metrics to Prometheus — render and fragment timings, cache
events, HTTP responses by route, invalidations — and serves them at `/metrics`.

```go
import "github.com/Elagoht/collage-prometheus"

m := prometheus.New(prometheus.Options{Namespace: "collage"})

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Metrics: m},
	Plugins:       []collage.Plugin{m},
})
```

```json
{
  "elagoht/prometheus": { "path": "/metrics", "tokenEnv": "METRICS_TOKEN" }
}
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.8 was the last for v0.60.0), and
  `NewMetrics` is now `New`. Hand the one value
  over as both the application's `Metrics` and a plugin: without the first nothing
  is measured, without the second nothing is served.
- No label is taken from a request. Since v0.2.0 `route` is what
  `collage.RouteInfo` says the request resolved to, as kind and name — `/blog/a` and
  `/blog/b` both `page:post`, `/robots.txt` `document:robots`, a mount
  `mount:/static/`, a 404 `other` — so a crawler inventing paths cannot mint time
  series. The `routes` option is gone: mounts and handlers are labelled by their
  prefix without it.
- With a token, the scrape must send it as a bearer token. The token is a secret,
  so it never goes in `plugins-config.json`: `tokenEnv` names the environment
  variable that holds it, or `Options.Token` sets it from Go; a `token` key stops
  the application. `path: "-"` serves the metrics nowhere. The path is exact: one another route already answers, or one
  ending in `/`, stops the application from starting.

#### elagoht/otel

[github.com/Elagoht/collage-otel](https://github.com/Elagoht/collage-otel) turns
collage's own spans into OpenTelemetry spans, and continues the trace a request's
caller started.

```go
import otel "github.com/Elagoht/collage-otel"

t := otel.New(otel.Options{Tracer: provider.Tracer("example.com/site")})

app, err := collage.New(&collage.Config{
	Observability: collage.ObservabilityConfig{Tracer: t},
	Plugins:       []collage.Plugin{t},
})
```

```json
{ "elagoht/otel": { "skip": ["/healthz"] } }
```

- v0.3.0 needs collage v0.61.0 or later (v0.2.10 was the last for v0.60.0), and
  `NewTracer(tracer)` is now `New(otel.Options{Tracer: tracer})`. As the tracer it
  turns `collage.http`, `collage.render` and `collage.fragment` into spans; as a
  plugin it reads the caller's trace context from the headers, in a `RequestHook`
  that runs before collage's own span. Each works alone.
- With both, a request is one trace: a server span, a child of the caller's, with
  `collage.http` under it and the render and every fragment under that. The server
  span is named for the route `collage.RouteInfo` reports — the pattern a page, a
  document or an action was registered with (`GET /blog/{slug}`, `GET /feed.xml`),
  the prefix of a mount or a handler.
- The application owns the SDK: the provider, the exporter, the sampler and the
  propagator. Set a propagator, or every request starts a trace of its own.
- `skip` prefixes match at a segment boundary, with or without the trailing `/`.
  `url.scheme` follows `collage.IsHTTPS`, so it is `https` behind a trusted proxy
  that says so.

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
    "respectDNT": true,
    "requireConsent": true
  }
}
```

- v0.2.3 needs collage v0.61.0 or later (v0.2.2 was the last for v0.60.0), and `{{hoist "head"}}` in the layout.
  Its keys spell initialisms as Go does: `respectDNT`, `goatCounter`, Umami's
  `websiteID`, Google Analytics' `measurementID`.
- Plausible, Umami and GoatCounter count visits without cookies. Google Analytics
  4 sets cookies; use it with `requireConsent`.
- With `respectDNT` or `requireConsent`, a small loader served from the site
  decides in the browser, before any provider's script is fetched: nothing loads
  under Do Not Track or Global Privacy Control, or until the page calls
  `window.collageAnalyticsConsent()`.
- `exclude` names pages that get no snippet.
- With [elagoht/consent](#elagohtconsent) on the site, set `consentCategory` (say
  `"analytics"`) instead of `requireConsent`. Every tag the plugin writes becomes a
  `type="text/plain" data-consent="analytics"` tag that the consent plugin runs once
  the visitor agrees to that category. The category must be one the consent plugin
  configures, or the tags never run. Google Analytics becomes two gated tags, its
  script and the inline call that configures it, with no loader.
- `consentCategory` cannot be combined with `requireConsent`: the application does
  not start with both, nor with a name that is not lowercase letters, digits and
  hyphens. `respectDNT` still applies and also honours Global Privacy Control, so a
  browser sending GPC loads nothing even after "Accept all". Under a strict
  Content-Security-Policy, Google Analytics' inline tag needs a `'sha256-…'` in
  `script-src` (see the consent plugin's CSP notes above).

#### elagoht/tenant

[github.com/Elagoht/collage-tenant](https://github.com/Elagoht/collage-tenant)
serves one site to many customers, each on a host of its own: `acme.app.com`,
`globex.app.com` or a customer's own domain. A cached page is kept per tenant, and
every absolute URL the site writes follows the host.

```go
import "github.com/Elagoht/collage-tenant"

Plugins: []collage.Plugin{tenant.New(tenant.Options{
	Tenants: []tenant.Static{
		{ID: "acme", Origin: "https://acme.app.com", Hosts: []string{"acme.app.com", "acme.localhost"}},
	},
	Resolve: lookupCustomDomain, // or a database, for hosts added at run time
})},
```

```json
{
  "elagoht/tenant": {
    "tenants": [
      { "id": "acme", "origin": "https://acme.app.com", "hosts": ["acme.app.com"] }
    ],
    "bypass": ["app.com"],
    "ttl": "1m"
  }
}
```

- v0.2.0 needs collage v0.61.2 or later (v0.1.4 was the last for v0.60.0), and `tenants` or a `Resolve` function (Go only).
- A host that is no tenant's is answered 404 with the site's own 404 page. A
  resolver that fails or panics is answered 503 with a `Retry-After`, and that is
  not cached. A "no tenant" answer is, for `ttl` (`"1m"`, a duration string; a
  number is an error); `maxHosts` (10000) bounds how many hosts are remembered —
  since v0.2.1 tenants and "no tenant" answers each up to that many, so a flood of
  made-up hosts cannot push a real tenant out, and concurrent misses for one host
  share one resolver call — and a negative one stops the application. `bypass` hosts are served without a tenant.
- A handler reads the tenant with `tenant.ID(rc)`, and code that holds only a
  context, such as a page's `StaticParams`, with `tenant.IDFromContext(ctx)`. Both
  are safe in a cached page. `collage.BaseURL(rc)` is the tenant's origin.
- `elagoht/sitemap`, `feed`, `meta`, `ogimage`, `indexnow`, `cdnpurge` and `robots`
  follow the host: the URLs they write are the tenant's origin.
- `acme.localhost` works through `collage dev`. A static build has no host, so it
  renders without a tenant and warns `tenant-no-host`. A static export is the site of
  `Config.BaseURL`'s host, rendered without a tenant; exporting per tenant is not
  supported. Its header capture is answered as that host, so it must be a tenant's
  or listed in `bypass`.

#### elagoht/errortrack

[github.com/Elagoht/collage-errortrack](https://github.com/Elagoht/collage-errortrack)
reports server errors and panics to Sentry, or to any service that speaks its
protocol (GlitchTip, self-hosted Sentry), with the standard library alone. It hears
every `5xx` and every recovered panic through `ErrorHook` and sends from a
background goroutine, so a slow or down Sentry never slows a request.

```go
import "github.com/Elagoht/collage-errortrack"

Plugins: []collage.Plugin{errortrack.New(errortrack.Options{
	User: func(r *http.Request) errortrack.User {
		return errortrack.User{ID: userID(r)}
	},
})},
```

```json
{
  "elagoht/errortrack": {
    "dsnEnv": "SENTRY_DSN",
    "environment": "production",
    "sampleRate": 1
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.7 was the last for v0.60.0), whose `ErrorEvent` carries the status and the
  request, and a `DSN` or `DSNEnv`: startup fails without one, except in dev mode
  without `inDevelopment`, where nothing is sent and no DSN is needed. The DSN is a
  credential; keep it in the environment, not in `plugins-config.json`. `timeout`
  (`"5s"`) is a duration string; a number stops the application.
- What leaves the server is limited: the error chain, a panic's stack, the route
  pattern rather than the raw path (only `scheme://host` when no route resolved, as
  for a middleware's panic; the raw path only with `sendPath`), query keys with
  `[filtered]` values, and six request headers, the `Referer` cut to its origin
  unless `sendPath` is set; the URL's scheme follows `collage.IsHTTPS`. Cookies, `Authorization`, the body and the client address are
  never sent (the address only with `sendIP`, as `collage.ClientIP` names it). A user is sent only as far as your
  `User` callback returns one.
- A message is sent as the failing code wrote it, and may carry a secret. Scrub it
  with `BeforeSend`, which may edit the event or return `false` to drop it.
- `minStatus` (500) lowers or raises what is reported; panics always are. `sampleRate`,
  `perMinute` (60) and `queueSize` (100) bound the volume: what exceeds them is
  dropped and counted, never retried, and a `429` pauses sending.
- `Capture(ctx, err)` on the plugin you registered reports an error from outside a
  request, such as a background job. Nothing is sent in dev mode unless
  `inDevelopment` is set.

#### elagoht/health

[github.com/Elagoht/collage-health](https://github.com/Elagoht/collage-health)
answers a platform's liveness and readiness probes, and turns readiness false the
moment collage starts [draining](/docs/deployment#graceful-shutdown-and-draining)
for a shutdown, so a load balancer stops sending requests before the port closes.
The endpoints are answered by its middleware, not by pages or documents, so a
static export never contains them.

```go
import "github.com/Elagoht/collage-health"

h := health.New(health.Options{})
h.Check("db", func(ctx context.Context) error { return db.PingContext(ctx) })

app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{DrainDelay: 5 * time.Second, ShutdownTimeout: 10 * time.Second},
	Plugins: []collage.Plugin{
		h, // before anything that can refuse or answer a request
		basicauth.New(basicauth.Options{}),
		ratelimit.New(ratelimit.Options{}),
	},
})
```

```json
{
  "elagoht/health": {
    "livePath": "/healthz",
    "readyPath": "/readyz",
    "checkTimeout": "2s",
    "cacheFor": "1s",
    "details": false,
    "maxInFlight": 0
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.2 was the last for v0.60.0). The
  durations are strings: `"checkTimeout": 2` stops the application, naming the
  key; write `"2s"`.
- `/healthz` (`livePath`) answers `200 ok` whenever the process serves a request,
  and runs no checks. `/readyz` (`readyPath`) answers `200 ok`, `503 unavailable`
  while a check fails, and `503 draining` once a drain has started. Both answer
  `GET` and `HEAD` with `Cache-Control: no-store`. A page registered at either
  path stops the application from starting; a document there is shadowed — the
  demo scaffold's `/healthz` document is never reached, so delete it or change
  `livePath`.
- `h.Check(name, fn)` adds a readiness check, before the application starts. A
  probe runs every check at once, each under a context that ends after
  `checkTimeout` (2s); one that errors, panics or has not returned by then fails.
  A result is reused for `cacheFor` (1s), and probes arriving during a run share
  it. An error is logged, never sent; with `details` readiness answers JSON naming
  each check `ok` or `failing` — leave it off where the names say too much.
- Once a drain starts, readiness is `503 draining` and the checks are not run. Set
  `DrainDelay` a little longer than the readiness probe's period times its failure
  threshold.
- `maxInFlight` (0, no cap) sheds load: a request past the cap is answered
  `503 overloaded` with `Retry-After: 1` at once, never queued. The cap is global,
  not per client — that is ratelimit's job — and the probes are never counted or
  refused.
- **List it before any plugin that can refuse or answer a request**, so a probe
  always reaches it: tenant answers a probe's pod-IP host with `404`, basicauth asks
  it for a password, ratelimit and fail2ban refuse it. compress may sit before it.
  accesslog listed after it logs neither probes nor shed requests; listed before,
  it logs both.

#### elagoht/jobs

[github.com/Elagoht/collage-jobs](https://github.com/Elagoht/collage-jobs) runs
scheduled jobs, on an interval or a cron expression, and typed background queues,
only while the application serves. Jobs and queue workers start when
`ListenAndServe` has bound its port, through
[`ServeHook`](/docs/writing-plugins#streams-and-shutdown), and jobs stop being
triggered when the [drain](/docs/deployment#graceful-shutdown-and-draining) begins.
Nothing runs in `collage build`, in a plugin command, or in a test that only calls
`app.Handler()` or `app.Start()`.

```go
import "github.com/Elagoht/collage-jobs"

j := jobs.New(jobs.Options{})
j.Every("refresh-feed", 15*time.Minute, refreshFeed).RunOnStart()
j.Cron("cleanup", "0 3 * * *", cleanup).Timeout(10 * time.Minute)

type Email struct{ To, Subject string }

mail := jobs.NewQueue(j, "email", func(ctx context.Context, m Email) error {
	return send(ctx, m)
}, jobs.QueueOptions[Email]{Workers: 2, MaxAttempts: 3})

app, err := collage.New(&collage.Config{
	Plugins: []collage.Plugin{j},
})

// In an action:
if err := mail.Enqueue(r.Context(), Email{To: to}); err != nil {
	// jobs.ErrQueueFull, or jobs.ErrDraining once Shutdown has begun:
	// tell the user to try again
}
```

```json
{
  "elagoht/jobs": {
    "timezone": "Europe/Istanbul",
    "disabled": ["cleanup"]
  }
}
```

- v0.2.0 needs collage v0.61.0 or later (v0.1.1 was the last for v0.60.0).
  Keep the plugin in a variable, as `j` above: `collage plugin add` writes
  `jobs.New(jobs.Options{})` straight into the list, where nothing can call
  `Every` on it.
- **Every instance runs its own scheduled jobs.** With N replicas behind a load
  balancer, a job runs N times per trigger; there is no leader election. Make such
  a job safe to run concurrently, or run it on one instance only, deciding that
  from the environment with `jobs.New(jobs.Options{Disabled: …})` rather than
  with `disabled` in a configuration every replica shares.
- A job is a `func(ctx context.Context) error`, defined before the application
  starts. `Every(name, d, fn)` runs it every `d`, measured in elapsed time, so
  neither DST nor a wall clock set back changes it. `Cron(name, expr, fn)` takes a
  five-field expression — minute, hour, day of month, month, day of week — with
  `*`, values, ranges, steps and comma lists; no names, macros or seconds. When day
  of month and day of week are both restricted, a day matches if either one does,
  as in Vixie cron: `0 0 1 * 1` runs on the 1st and on every Monday. An expression
  that is malformed or can never match panics at `Cron`.
- `RunOnStart()` also runs the job as soon as the plugin starts; `Timeout(d)`
  bounds each run's ctx, which the job must watch. A run in progress makes a
  trigger skip, so runs never overlap, and a failed or panicking run is logged at
  Error and not retried.
- Cron expressions are read in `timezone` (UTC). Across a DST change, a time in the
  skipped hour runs once, at the first instant after the gap; a time in the
  repeated hour runs once, on its first occurrence, for a schedule with a fixed
  hour; and a schedule whose hour field starts with `*` fires through both copies
  of the repeated hour, as Vixie cron does. `disabled` names jobs that do not run;
  an unknown name, like an invalid zone, fails startup.
- The zone is read from the system's zoneinfo, which an image `FROM scratch` does
  not have. Embed it, about 450 KB, with `import _ "time/tzdata"` in `main`. The
  scaffold's distroless image ships zoneinfo already.
- `NewQueue[T]` hands items of type `T` to a function, with no type assertion.
  `Workers` (1) is how many items are handled at the same time, `Capacity` (1000)
  bounds how many wait, and `MaxAttempts` (1, no retry) is how many times an item
  is tried; each retry waits 1s, 2s, 4s… capped at a minute. `OnFailure` is called
  once with the last error. `Enqueue` never blocks: it returns
  `ErrQueueFull` on a full queue, and `ErrDraining` only once the plugin's
  `Shutdown` has begun. The drain refuses nothing, since the server still serves
  then. Items live in memory and are lost if the process ends.
- **Shutdown finishes queued work.** It waits for runs in progress and keeps taking
  queued items until every queue is empty. If its ctx ends first, it cancels the
  runs and waits one more second for them. Items left are logged at Warn as
  `dropped`, and runs still going after the give-up as `abandoned`. With it, a
  stop can take `DrainDelay + ShutdownTimeout + 1s`, which your platform's grace
  period must cover; see [Deployment](/docs/deployment#graceful-shutdown-and-draining).
- **Register it after the plugins its work uses**, so it shuts down first, while
  they still work. They share its shutdown deadline, so size `ShutdownTimeout` for
  the queued work plus what they need.
- Do not enqueue from data handlers: a build or a header capture runs them, never
  serves, and the item is never handled. A project scaffolded by collage v0.55.0
  or later at least logs it as dropped, because its `main.go` shuts the
  application down after a build. An application on its own `http.Server` starts
  the plugin with `j.Start(ctx)`; see
  [Deployment](/docs/deployment#serving-with-your-own-server).

## Libraries that are not plugins

The plugins above are all registered in `Config.Plugins`. A library is a module
your own code uses — from a page, or in the application's configuration — so there
is nothing to register, no `collage plugin add` and no `plugins-config.json`
section, and it is not counted among the forty-five plugins.

### elagoht/paginate

The arithmetic of a pager, a numbered window (`1 … 4 5 [6] 7 8 … 20`), and links
for a paged listing: path-paged ones a static export writes (`/blog`,
`/blog/page/2`) and query-paged ones only a server answers (`/search?page=3`). It
is used from a data handler and a template; see [Pagination](/docs/static-export#pagination).

```sh
go get github.com/Elagoht/collage-paginate
```

- v0.2.0 needs collage v0.61.0 or later. `FromPath(rc, last)` reads the path
  parameter `n`, the one `PathNamed` and `StaticParams` write, where v0.1 took the
  parameter's name; `paginate.Version` is gone.

### elagoht/redis

A Redis-backed [`collage.SharedCache`](/docs/caching#a-cache-a-fleet-shares): every
instance of an application runs against one page cache and one `collage.Cached`
data store, so an invalidation on any instance is seen by all of them. It is handed
to `Cache.Store`:

```sh
go get github.com/Elagoht/collage-redis
```

```go
import "github.com/Elagoht/collage-redis" // package redis

store, err := redis.New(redis.Config{URL: os.Getenv("REDIS_URL")})
if err != nil {
	log.Fatal(err)
}
defer store.Close()

app, err := collage.New(&collage.Config{
	Cache: collage.CacheConfig{Enabled: true, Store: store},
})
```

- v0.1.0 needs collage v0.66.0 or later and Redis 6 or later. A single node and
  Sentinel (through `Config.Client`) are supported; Redis Cluster is not.
- `New` never dials Redis, so an application starts, and serves uncached, while
  Redis is down. A Redis error is a miss, never a failed request.
- By default every deploy of a new binary starts with an empty namespace, and `MaxAge` (seven
  days) bounds every key, the stale fallback included. The configuration, and how
  to run Redis for it, are in [Caching](/docs/caching#a-cache-a-fleet-shares) and
  the library's README.

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
on your layout's behalf.

A layout from `collage new` already has the marker; a
layout you wrote by hand may not. See [Head and SEO](/docs/head-and-seo) for
hoisting in general. Among the published plugins, jsonld, meta, ogimage, feed,
favicon and analytics write to the head this way.

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
- **A static build's header capture** (since v0.52.0) asks the handler for every
  written file when a `BuildFinishedHook` plugin is registered, so middleware and
  the request hooks see a build's paths. A plugin that counts or limits traffic
  skips them with `collage.IsCapture(r.Context())`. See
  [Static export](/docs/static-export#headers-and-redirects).
- **Not around a plain `app.Handle` mount.** Since v0.67.0 a handler the
  application mounts with `app.Handle` runs no plugin's middleware — no session,
  basicauth, rate limit, ban, access log, redirect or tenant — unless it is mounted
  `collage.WithMiddleware()`. The request hooks still run for it, and a plugin's own
  `Host.Handle` endpoints always run inside the middleware. See
  [No middleware, unless you ask](/docs/middleware-and-apis#no-middleware-unless-you-ask).
- **Work that runs only while serving** — a scheduler, a queue worker — starts in
  `OnServe` (since v0.55.0), which only `ListenAndServe` calls: a static export, a
  plugin command and `app.Handler()` never start it. See
  [Writing a plugin](/docs/writing-plugins#streams-and-shutdown).

## Going further

[Writing a plugin](/docs/writing-plugins) covers the plugin contract, every hook
and what it may change, and a complete plugin with its tests.
