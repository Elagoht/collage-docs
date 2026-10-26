---
description: Every built-in template function — slot, hoist, asset, the URL functions, json, arithmetic, strings and lists — with its signature and an example.
reference: TemplateConfig, DefaultContentSlot, ErrUnknownSlot, FragmentBuilder.WithTitle, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Template functions

Templates are Go's `html/template`, so everything it provides is there: `if`,
`range`, `with`, `define`, `block`, and the standard functions `and`, `or`, `not`,
`len`, `index`, `slice`, `print`, `printf`, `println`, `eq`, `ne`, `lt`, `le`, `gt`,
`ge`, `call`, `html`, `js` and `urlquery`. collage adds the functions on this page.

| Function | Signature | Returns |
| --- | --- | --- |
| [`slot`](#slot) | `slot name` | the markup of the fragments in a slot |
| [`hoist`](#hoist) | `hoist area` | the place hoisted content lands |
| [`asset`](#asset) | `asset path` | a mounted file's content-addressed URL |
| [`stylesheet`](#stylesheet) | `stylesheet path` | nothing; declares a stylesheet for the head |
| [`csrfToken`](#csrftoken) | `csrfToken` | the hidden input a form's token travels in |
| [`pageURL`](#pageurl) | `pageURL name [param value]...` | a route's URL in this render's locale |
| [`pageURLIn`](#pageurlin) | `pageURLIn locale name [param value]...` | a route's URL in exactly that locale |
| [`localeURL`](#localeurl) | `localeURL locale` | this page's URL in another locale |
| [`actionURL`](#actionurl) | `actionURL name [param value]...` | an action's URL in this render's locale |
| [`fragmentURL`](#fragmenturl) | `fragmentURL page fragment [param value]...` | a fragment path's URL in this render's locale |
| [`fragmentURLIn`](#fragmenturlin) | `fragmentURLIn locale page fragment [param value]...` | a fragment path's URL in exactly that locale |
| [`withQuery`](#withquery) | `withQuery name value [name value]...` | this page's URL with query parameters set or removed |
| [`safeHTML`](#safehtml) | `safeHTML string` | the string, trusted as HTML |
| [`safeURL`](#safeurl) | `safeURL string` | the string, trusted as a URL |
| [`dict`](#dict) | `dict key value [key value]...` | a map built from pairs |
| [`default`](#default) | `default fallback value` | `value`, or `fallback` when it is empty |
| [`upper`](#upper-and-lower) | `upper string` | the string in upper case |
| [`lower`](#upper-and-lower) | `lower string` | the string in lower case |
| [`title`](#title) | `title string` | each word capitalised |
| [`join`](#join) | `join sep items` | the items joined by `sep` |
| [`formatTime`](#formattime) | `formatTime time layout` | the time, formatted |
| [`pathEscape`](#pathescape) | `pathEscape string` | the string, escaped for one path segment |
| [`json`](#json) | `json value [spaces] [pretty]` | the value as JSON |
| [`pre`](#pre) | `pre value [spaces] [pretty]` | the value in an escaped `<pre>` block |
| [`add`, `sub`, `mul`, `div`, `mod`, `min`, `max`](#arithmetic) | `add a b` | integer arithmetic |
| [`addf`, `subf`, `mulf`, `divf`](#arithmetic) | `addf a b` | the same for `float64` |
| [`trim`, `replace`, `contains`, `hasPrefix`, `hasSuffix`, `split`](#strings) | `replace old new string` | the `strings` functions, the string last |
| [`truncate`](#strings) | `truncate n string` | the first `n` characters, with `…` when anything was cut |
| [`first`, `last`](#lists-and-maps) | `first list` | a list's first or last element |
| [`seq`](#lists-and-maps) | `seq n` | `1` to `n`, to range over |
| [`keys`, `hasKey`](#lists-and-maps) | `keys map`, `hasKey map key` | a map's keys, sorted; whether it holds one |
| [`dig`](#dig) | `dig value key...` | the value at a path through maps and structs, or nothing |

Everything from `withQuery` down is new in v0.77.0. Every function is typed, so the
[template checker](/docs/data-handlers#how-templates-are-checked) — at
registration and in `collage check` — reports `{{add .Title 1}}` with a string
`Title` before any request does.

A function that returns an error fails the template, and the fragment fails with
it — under the fragment's own failure policy:

- a required fragment fails the page;
- an optional one renders its fallback;
- or nothing when it has none.

In development mode "nothing" is an HTML comment naming the fragment and its
error, and the development panel on top of the page names every failed fragment,
fallback or not. In every mode, since v0.36.0, the failure is also logged as a
warning. See [Fragments and slots](/docs/fragments-and-slots).

## Bound per render

The first twelve functions need the render they are part of — the fragment, the
request, the locale, the application's routes and mounts. Each reads something
different:

| Function | What it reads |
| --- | --- |
| `slot` | this fragment, and the fragments bound into its slots |
| `hoist` | the render's areas, filled once the whole page has rendered |
| `asset` | the application's mounts |
| `stylesheet` | the head this render declares into |
| `csrfToken` | a placeholder each reader's token replaces as the response is written |
| `pageURL`, `actionURL`, `fragmentURL` | the render's locale, with the application's routes |
| `pageURLIn`, `fragmentURLIn` | the application's routes; the locale is an argument |
| `localeURL` | the page being rendered and its path parameters |
| `withQuery` | the page being rendered, its path parameters and the request's query |

They are registered when templates are parsed as placeholders, so that a template
may call them, and the render engine binds the real implementation on every
render. A placeholder that somehow runs outside a render returns an error rather
than guessing.

One consequence: an entry under one of these names in `Config.Template.Funcs` is
parsed but never called, and since v0.77.0 `New` logs a warning naming it. The
other thirty-six can be replaced from `Config.Template.Funcs`. A plugin can replace
none of the forty-eight: see [Adding your own](#adding-your-own).

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 306" role="img" aria-labelledby="tf-bind-title">
<title id="tf-bind-title">Where a template's functions come from. collage.New parses the templates with one function map: collage's forty-eight built-ins, twelve of them placeholders, then the functions plugins add with AddTemplateFunc under names of their own, then Config.Template.Funcs, each merged over the one before. On every render, for each fragment, the twelve per-render names are bound again from that render, and so are the functions plugins add with AddRenderFunc. So a Template.Funcs entry named pageURL is parsed and never runs, while one named upper replaces the built-in.</title>
<defs><marker id="tf-bind-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="40" y="4" width="240" height="22" rx="11" /><text x="160" y="19" text-anchor="middle">collage.New: parse templates</text>
<line class="d-flow" x1="160" y1="26" x2="160" y2="36" marker-end="url(#tf-bind-arrow)" />
<rect class="d-frame" x="4" y="40" width="312" height="96" rx="6" />
<text class="d-small" x="14" y="55">one function map</text>
<rect class="d-b" x="14" y="62" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="76" text-anchor="middle">48 built-ins (12 placeholders)</text>
<rect class="d-b" x="14" y="86" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="100" text-anchor="middle">plugins: AddTemplateFunc</text>
<rect class="d-b" x="14" y="110" width="232" height="20" rx="4" /><text class="d-on-b" x="130" y="124" text-anchor="middle">Config.Template.Funcs</text>
<line class="d-flow" x1="262" y1="66" x2="262" y2="126" marker-end="url(#tf-bind-arrow)" />
<text class="d-small" x="292" y="92" text-anchor="middle">later</text><text class="d-small" x="292" y="104" text-anchor="middle">wins</text>
<line class="d-flow" x1="160" y1="136" x2="160" y2="148" marker-end="url(#tf-bind-arrow)" />
<rect class="d-pill" x="40" y="152" width="240" height="22" rx="11" /><text x="160" y="167" text-anchor="middle">every render, each fragment</text>
<line class="d-flow" x1="140" y1="174" x2="90" y2="188" marker-end="url(#tf-bind-arrow)" />
<line class="d-flow" x1="180" y1="174" x2="230" y2="188" marker-end="url(#tf-bind-arrow)" />
<rect class="d-a" x="12" y="192" width="146" height="40" rx="5" /><text class="d-on-a" x="85" y="208" text-anchor="middle">12 per-render</text><text class="d-on-a d-small" x="85" y="223" text-anchor="middle">bound from this render</text>
<rect class="d-c" x="162" y="192" width="146" height="40" rx="5" /><text class="d-on-c" x="235" y="208" text-anchor="middle">AddRenderFunc</text><text class="d-on-c d-small" x="235" y="223" text-anchor="middle">plugins', per render</text>
<text class="d-small" x="160" y="252" text-anchor="middle">a Template.Funcs entry named…</text>
<g class="d-dropped"><rect class="d-frame" x="12" y="260" width="146" height="40" rx="6" /><text x="85" y="276" text-anchor="middle">"pageURL"</text><text class="d-small" x="85" y="291" text-anchor="middle">parsed, never runs</text></g>
<rect class="d-chip-hit" x="162" y="260" width="146" height="40" rx="8" /><text class="d-hit-text" x="235" y="276" text-anchor="middle">"upper"</text><text class="d-hit-text d-small" x="235" y="291" text-anchor="middle">replaces the built-in</text>
</svg>
<figcaption>Every template is parsed in <code>collage.New</code> with one function map: collage's forty-eight built-ins, then what plugins add with <code>AddTemplateFunc</code> under names of their own, then <code>Config.Template.Funcs</code>, each over the one before. Twelve of the built-ins are placeholders. On every render, for each fragment, those twelve are bound again from that render, and so are the functions plugins add with <code>AddRenderFunc</code>. An entry of yours under one of the other thirty-six names, such as <code>upper</code>, replaces the built-in; one under a per-render name, such as <code>pageURL</code>, is parsed and never runs.</figcaption>
</figure>

### slot

```html
{{slot "name"}}
```

Renders every fragment bound to this fragment's slot `name`, in binding order, and
inserts their markup unescaped. `slot` always means *this* fragment's slot, so two
fragments can each have a slot called `"sidebar"` without meeting.

```html
<article>
  {{slot "content"}}
  <aside>{{slot "related"}}</aside>
</article>
```

- Calling `slot` is what declares the slot: the fragment's Go code needs no
  `WithSlot` for it, and one the template calls with nothing bound renders nothing.
- The check faces the other way: a fragment bound into a slot this template never
  calls fails registration with `ErrUnknownSlot`, naming the slot and the slots the
  template does call. Otherwise a typo on either side of the binding would be a
  section that is silently missing. A slot named by anything but a literal,
  `{{slot .Which}}`, turns the check off for that fragment.
- An empty slot renders nothing — unless `WithSlot` made it required:
  - A required slot with nothing bound and no resolver is refused when the page is
    registered, with `ErrRequiredSlotUnfilled`, so it never reaches a render.
  - A required slot filled by a resolver is checked at render time instead: a
    resolver that returns no fragments fails the render with `ErrRequiredSlotEmpty`.
- A layout's content goes in `{{slot "content"}}`, the name
  `collage.DefaultContentSlot` holds.

The children's data handlers have already started by the time `slot` runs: they
all start together, before the parent renders. See
[Fragments and slots](/docs/fragments-and-slots).

### hoist

```html
{{hoist "head"}}
```

Marks where content hoisted to `area` lands. Fragments declare that content from
anywhere in the tree — usually from a data handler, or with `WithTitle` for a
title fixed when the program starts — and `hoist` decides where it goes:

```html
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
```

```go
rc.HoistTitle(post.Title)
rc.HoistMeta("description", post.Summary)
rc.HoistLink("canonical", canonicalURL)
```

It writes a marker, not content, because nothing below it has rendered yet. Once
the whole tree has rendered, every marker is replaced with what was declared for
its area, so a declaration made deep inside the page still reaches the head. An
area nothing was declared for is replaced with nothing.

`HoistTitle`, `HoistMeta`, `HoistProperty`, `HoistLink`, `HoistAlternate` and
`HoistStylesheet` write to the `"head"` area; `rc.Hoist(area, key, html)` writes
to any area you name. **A plugin that contributes to the head needs this
marker**: without it, its content has nowhere to go. See [Head and SEO](/docs/head-and-seo).

### asset

```html
<link rel="stylesheet" href="{{asset "/static/app.css"}}">
<img src="{{asset "/static/logo.svg"}}" alt="">
```

Returns the content-addressed URL of a file in a mount: the file's path with a
hash of its contents before the extension.

```text
/static/app.css  →  /static/app.41014ebb6c2d9f07.css
```

A name derived from the bytes can only ever mean those bytes, which is what makes
it safe to serve with a one-year, `immutable` lifetime: a changed file is a changed
name, and the old one is simply never requested again.

- The path is the URL the mount serves the file at, including the mount's prefix.
- **A file that does not exist is an error**, not a URL: `ErrUnknownAsset`. The
  alternative is a page that renders fine and links a stylesheet that 404s.
- A path no mount serves, or an application with no mounts, is `ErrUnknownAsset`
  too; the message says when no mount's prefix covers the path at all.
- A static export writes exactly the fingerprinted copies that pages asked for,
  alongside the originals.

From Go, `rc.Asset(path)` returns the same URL. See [Static assets](/docs/assets).

### stylesheet

```html
{{stylesheet "/static/gallery.css"}}
```

Declares that this fragment needs a stylesheet, and renders nothing where it is
called. The page's head gets one
`<link rel="stylesheet" href="…">` with the file's [`asset`](#asset) URL, where
the layout calls `{{hoist "head"}}`.

A fragment can therefore carry its own styles wherever it is used:

```html
{{stylesheet "/static/gallery.css"}}
<div class="gallery">…</div>
```

The declaration is keyed by the path, so a stylesheet several fragments ask for
appears once. It fails on the same terms as `asset`. From Go it is
`rc.HoistStylesheet(path)`.

### csrfToken

```html
<form method="post">
  {{csrfToken}}
  <input name="email" type="email">
  <button>Subscribe</button>
</form>
```

Renders the hidden input a form's request-forgery token travels in:

```html
<input type="hidden" name="_csrf" value="…">
```

A whole input rather than the bare value, because the bare value has to be placed
in a field with exactly the right name, and a form that gets the name wrong is
refused with nothing to explain why.

The value written during the render is a placeholder. It is replaced with each
reader's own token as the response is written.

That is why a **cached** page can carry a form: the cached bytes hold the
placeholder, and every reader gets their own token.

- With `Security.DisableCSRF` set, `csrfToken` fails the render with
  `ErrCSRFDisabled`: a form expecting a token would otherwise be rendered without
  one.
- The field is named `_csrf` unless `Security.CSRFFieldName` says otherwise, and
  it is the name the verifier reads either way, so a renamed field needs no change
  to the template.
- A static export has no server to put a token in, so a page carrying one is not
  exported, and the report says why.

See [Forms and actions](/docs/forms-and-actions).

### pageURL

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

Returns the URL of the page or document registered under `name`, filling its path
pattern from `param value` pairs. Linking by name means a link follows a page when
its path changes.

- **In this render's locale.** On a Turkish page, `pageURL "blog-post"` is the
  Turkish path. A route with no path in the current locale links its
  default-locale path instead, so a Turkish page linking an English-only page
  still renders.
- **Strict.** An unknown name (`ErrUnknownRoute`), an odd number of parameter
  arguments, a missing or empty parameter, or a parameter the pattern has no
  placeholder for (`ErrRouteParams`) fails the render. A link that cannot be built
  is a bug to find in development, not a 404 for a reader.
- **Values are strings, integers or `fmt.Stringer`s** (integers and Stringers
  since v0.36.0), and are escaped.
- **Some values are refused.** A value of `.` or `..`, which a browser would
  resolve as a path step, is refused. So is a float, a bool or nil, with
  `ErrRouteParams` naming its type, since none has one obvious spelling in a URL:

```html
<a href="{{pageURL "user" "id" .ID}}">{{.Name}}</a>
```

- A name registered as both a page and a document is refused rather than guessed.

From Go it is `app.URL(name, locale, params)`. See
[Links and locales](/docs/links-and-locales).

### pageURLIn

```html
<a href="{{pageURLIn "tr" "about"}}">Hakkımızda</a>
<a href="{{pageURLIn "en" "blog-post" "slug" .Slug}}">Read in English</a>
```

`pageURL` in exactly the locale given, with no fallback: a route with no path in
that locale is `ErrNoPathInLocale`, and a locale no URL can reach — neither the default nor in
`Locale.Supported` — is
`ErrLocaleUnreachable`.

### localeURL

```html
{{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
{{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
```

The page being rendered, in another locale, with the same path parameters — what
a language switcher is made of.

- **A page with no path in that locale is the empty string**, not an error, so
  `{{with}}` skips a language the page has not been translated into.
- A locale no URL can reach is still an error (`ErrLocaleUnreachable`): that is a
  mistake in the template, not a missing translation.
- It fails in a render that is not a page's: a document's, or an action's at a URL
  of its own. A fragment served at a URL its page opened for it renders with that
  page (since v0.77.0), so there it links the page.
- For the `<link rel="alternate" hreflang>` search engines read, declare them from
  Go with `rc.HoistAlternate` — see
  [Head and SEO](/docs/head-and-seo#canonical-and-alternate-links).

### actionURL

```html
<form method="post" action="{{actionURL "logout"}}">{{csrfToken}}…</form>
<form method="post" action="{{actionURL "vote" "id" .ID}}">{{csrfToken}}…</form>
```

The URL of the [action](/docs/forms-and-actions#an-action-at-its-own-url)
registered under `name` (since v0.36.0), built as `pageURL` builds a page's: the
path is written once, in Go, and every form posting to it follows it.

- **In this render's locale**, falling back to the default one, as `pageURL` does.
- **As strict as `pageURL`.** An unknown name is `ErrUnknownRoute`, and
  parameters that do not fill the pattern are `ErrRouteParams`.
- **Actions have names of their own.** `actionURL "login"` is the action and
  `pageURL "login"` the page, since a `login` page and a `login` action are the
  usual pair. A page's own action is under the name it was registered with:
  `"contact:POST"` for `WithAction("POST", …)` on a page named `contact`.

From Go it is `app.ActionURL(name, locale, params)`.

### fragmentURL

```html
<div data-live="{{fragmentURL "home" "cpu-usage"}}">{{slot "cpu-usage"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

The path a page opened for one of its fragments with
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url), built from
the page's name and the fragment's (since v0.18.0). The path is written once, in
Go, and every link to it follows it.

- **In this render's locale**, falling back to the default one, as `pageURL` does.
- **As strict as `pageURL`.** An unknown page is `ErrUnknownRoute`, a fragment the
  page did not open is `ErrUnknownFragmentPath`, a fragment opened at two paths in
  one locale is `ErrAmbiguousFragmentPath`, and parameters that do not fill the
  pattern are `ErrRouteParams`. Each fails the render.

From Go it is `app.FragmentURL(page, fragment, locale, params)`. See
[Links and locales](/docs/links-and-locales#a-fragments-url).

### fragmentURLIn

```html
<div data-live="{{fragmentURLIn "tr" "home" "cpu-usage"}}">…</div>
```

`fragmentURL` in exactly the locale given, with no fallback: a fragment with no
path in that locale is `ErrNoPathInLocale`, and a locale no URL can reach is
`ErrLocaleUnreachable`.

### withQuery

```html
<a href="{{withQuery "page" (add .Page 1)}}">Older posts</a>
<a href="{{withQuery "tag" ""}}">Every tag</a>
<a href="{{withQuery "page" 1 "sort" "new"}}">Newest first</a>
```

The page being rendered, with the request's query and those parameters set (since
v0.77.0). A value of `""` removes a parameter, and the others stay as they were.
On `/blog?tag=go&page=2`, the three links above are `/blog?page=3&tag=go`,
`/blog?page=2` and `/blog?page=1&sort=new&tag=go`: the query is written with its
names sorted.

- **The path is the page's own URL**, in the render's locale with its path
  parameters, falling back to the default locale as `pageURL` does. In a fragment
  refreshed from a [URL of its own](/docs/forms-and-actions#a-fragment-at-its-own-url),
  the link still leads to the page, not to the bare fragment. In an action at a URL
  of its own, which has no page, it is the action's URL.
- **The query is the one the render sees.** On a cached page that is only the
  parameters in its [cache key](/docs/caching#what-a-shared-render-sees), never a
  reader's other ones, so a link written into the shared copy cannot carry one
  reader's `utm_source` to the next.
- **Names are strings, and values are strings, integers or `fmt.Stringer`s**, as
  `pageURL`'s are. Another type, or a name without a value, fails the render.

### safeHTML

```html
{{safeHTML .RenderedMarkdown}}
```

Marks a string as trusted HTML, so `html/template` inserts it without escaping. It
is an escape hatch: use it only for markup you generated or sanitised yourself,
never for anything a user typed.

### safeURL

```html
<a href="{{safeURL .ExternalLink}}">Visit</a>
```

Marks a string as a trusted URL, bypassing `html/template`'s URL sanitisation —
which otherwise replaces a scheme it does not trust, such as `javascript:`, with
`#ZgotmplZ`. Only for URLs you have validated.

### dict

```html
{{template "card" dict "Title" .Title "URL" (pageURL "post" "slug" .Slug)}}
```

Builds a `map` from alternating keys and values, for passing several values into a
sub-template. Keys must be strings, and the arguments must come in pairs; either
mistake fails the template, with a message that says which.

### default

```html
<h2>{{default "Untitled" .Subtitle}}</h2>
<h2>{{.Subtitle | default "Untitled"}}</h2>
```

Returns `value`, or `fallback` when `value` is the empty string. Both arguments are
strings — the fallback comes first so that the piped form reads naturally, since a
pipeline passes its value as the last argument.

### upper and lower

```html
<span class="badge">{{upper .Status}}</span>
<code>{{lower .Code}}</code>
```

`strings.ToUpper` and `strings.ToLower`.

### title

```html
<h1>{{title .Name}}</h1>
```

Upper-cases the first letter of every word and lower-cases the rest. A word is a
run of letters, so anything else — a space, a hyphen, an apostrophe — starts a new
one:

| Input | Output |
| --- | --- |
| `hello world` | `Hello World` |
| `iPHONE case` | `Iphone Case` |
| `o'neil-smith` | `O'Neil-Smith` |

### join

```html
<p>Tags: {{join ", " .Tags}}</p>
```

`strings.Join(items, sep)`, with the separator first. `items` must be a
`[]string`.

### formatTime

```html
<time datetime="{{formatTime .Published "2006-01-02"}}">
  {{formatTime .Published "2 January 2006"}}
</time>
```

`t.Format(layout)`, with a [reference-time layout](https://pkg.go.dev/time#pkg-constants).
`t` is a `time.Time`. It is formatted in whatever location it carries; convert it
in the data handler if you want another.

### pathEscape

```html
<a href="https://example.com/wiki/{{pathEscape .Term}}">{{.Term}}</a>
```

`url.PathEscape`: the string escaped to sit in one segment of a path, so a `/` in it
becomes `%2F` rather than a step down (since v0.77.0). A link to one of your own
pages needs none of this: `pageURL` escapes its parameters itself.

### json

```html
<script type="application/json" id="chart-data">{{json .Series 2 false}}</script>
```

The value as JSON (since v0.77.0), indented by two spaces unless told otherwise:
`{{json v 4}}` indents by four, and `{{json v 2 false}}` writes one line. The
spaces are 0 to 16.

It returns `template.JS`. Inside a `<script>` — `application/json` data, or a
script that reads it — it is written as it is, and `<`, `>` and `&` stay `\u003c`-escaped, so
a value cannot close the script. Anywhere else it is escaped like any other text. A
value `encoding/json` cannot encode, such as a channel, fails the render.

### pre

```html
{{pre .}}
```

The value in a `<pre>` block, for looking at what a template was given (since
v0.77.0). A string or a `[]byte` is shown as it is; anything else as `json` writes
it, with the same optional spaces and `pretty`. The content is always escaped.

### Arithmetic

```html
<p>Page {{.Page}} of {{.Pages}}, posts {{add (mul (sub .Page 1) .PerPage) 1}} onwards</p>
<p>{{printf "%.1f" (divf .Rating 2.0)}} / 5</p>
```

`add`, `sub`, `mul`, `div`, `mod`, `min` and `max` take two `int`s, and `addf`,
`subf`, `mulf` and `divf` two `float64`s (since v0.77.0). `div`, `mod` and `divf`
by zero fail the render.

### Strings

```html
<h2>{{.Slug | replace "-" " " | title}}</h2>
{{if hasPrefix "https://" .Link}}<a href="{{.Link}}" rel="external">{{.Link}}</a>{{end}}
<p>{{truncate 140 .Summary}}</p>
```

`trim` is `strings.TrimSpace`. `replace old new s`, `contains sub s`,
`hasPrefix prefix s`, `hasSuffix suffix s` and `split sep s` are the `strings`
functions with the string last, so each reads in a pipeline (since v0.77.0).
`replace` replaces every `old`, and `split` returns a `[]string`.

`truncate n s` cuts `s` to its first `n` characters and puts `…` after them when
anything was cut. It counts characters, not bytes, so a Turkish letter is never
split in half. A negative `n` fails the render.

### Lists and maps

```html
{{with first .Posts}}<a href="{{pageURL "post" "slug" .Slug}}">Latest: {{.Title}}</a>{{end}}
{{range seq 5}}<span class="star">{{.}}</span>{{end}}
{{range keys .Counts}}<li>{{.}}: {{index $.Counts .}}</li>{{end}}
{{if hasKey .Flags "beta"}}<span class="badge">beta</span>{{end}}
```

All since v0.77.0:

- `first list` and `last list` are a slice's or an array's first and last element,
  and nothing for an empty one, so `{{with}}` skips it.
- `seq n` is `1` to `n`, for a loop that counts. It is empty when `n` is below 1,
  and more than 100000 fails the render.
- `keys map` is a string-keyed map's keys, sorted, so a loop over them renders the
  same way every time. `hasKey map "k"` says whether the map holds `k`, even with
  a zero value under it.
- Each fails the render when handed something else: a map to `first`, a list to
  `keys`.

### dig

```html
{{with dig .Settings "theme" "accent"}}<meta name="theme-color" content="{{.}}">{{end}}
```

The value at that path through maps with string keys and the exported fields of
structs, following pointers (since v0.77.0). A path that leads nowhere — a missing
key, an unexported field, a nil pointer on the way — is nothing rather than an
error, so `{{with dig …}}` handles a value that is optional.

## Adding your own

Add functions with `Config.Template.Funcs`, before `New`:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		Root: "templates",
		Funcs: template.FuncMap{
			"money": func(cents int64) string {
				return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
			},
		},
	},
})
```

They are merged over the built-ins, so an entry under a built-in name — other than
the per-render ones — replaces it. An entry under a per-render name is accepted and
never runs; since v0.77.0 `New` logs a warning naming it.

They must be there before `New` because `html/template` can only call a name that
was in its function map when the template was parsed; a template calling a name
nobody registered fails in `New`, not at the first request.

A plugin adds functions the same way, from its `Configure` phase; an application's
entry under the same name wins. A plugin cannot take a built-in's name: since
v0.77.0 `AddTemplateFunc` refuses it with `ErrDuplicateTemplateFunc`, where it used
to replace the built-in without a word. See
[Writing a plugin](/docs/writing-plugins#template-functions).

A plugin's `AddRenderFunc` function is the exception on both sides:

- It is bound per render, like the twelve above, so an application's entry under
  its name is parsed and never called.
- Like `AddTemplateFunc`, it cannot take any of the forty-eight built-in names.
  `AddRenderFunc` refuses them with `ErrDuplicateTemplateFunc`.

Anything that needs the request belongs in a data handler rather than a function:
that is where the page's data comes from anyway.

## Example: a product page in two languages

The shop from [Links and locales](/docs/links-and-locales#example-a-bilingual-product-catalogue)
sells in English and Turkish. Its product page, with a layout around it, calls nine
of the twelve per-render functions:

| Function | In the example |
| --- | --- |
| `asset` | the shop's stylesheet, and the product's picture, whose path comes from the data |
| `hoist` | where the title and `product.css` land in the head |
| `localeURL` | the language switcher |
| `slot` | the layout's content, and the product's stock box |
| `stylesheet` | `product.css`, which only this page needs |
| `pageURL` | the link back to the catalogue |
| `fragmentURL` | the stock box's own URL, for a script that refreshes it |
| `actionURL` | the "add to basket" form's target |
| `csrfToken` | that form's token |

The `catalog` package stands for your own data layer, and keeps each product's
name in both languages. The handler picks the labels for the render's locale and
hoists the product's name as the title:

```go
type labels struct {
	Back string
	Add  string
}

var labelsIn = map[string]labels{
	"en": {Back: "All products", Add: "Add to basket"},
	"tr": {Back: "Tüm ürünler", Add: "Sepete ekle"},
}

type productView struct {
	catalog.Product // ID, Name, Image
	Labels          labels
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (productView, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Find(ctx, id, rc.Locale)
	if errors.Is(err, catalog.ErrNoProduct) {
		return productView{}, nil, fmt.Errorf("catalog: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return productView{}, nil, err
	}
	rc.HoistTitle(p.Name)
	return productView{Product: p, Labels: labelsIn[rc.Locale]}, []string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (int, error) {
	return catalog.Stock(ctx, rc.Param("id"))
}
```

The page and its stock box have a path in each locale, and so does the basket
action:

- The catalogue page is the one Links and locales registers.
- `addToBasket` is the action's handler, as in
  [Forms and actions](/docs/forms-and-actions).
- `static/` is mounted at `/static/`, as in [Static assets](/docs/assets).
- The product is `Required`: a page without it has nothing to show.

```go
layout := collage.NewFragment("layout", "layouts/master.html").Build()

stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

product := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(collage.NewFragment("product", "pages/product.html").
		WithData(collage.LoadTagged(loadProduct)).
		WithSlotFragment("stock", stock).
		Required().
		Build()).
	WithPath("en", "/products/{id}").
	WithPath("tr", "/urunler/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	WithFragmentPath("tr", "/urunler/{id}/stok", stock).
	Build()

if err := app.Register(catalogue, product); err != nil {
	return err
}
return app.RegisterAction(collage.NewAction("basket-add").
	WithPath("en", "/basket/{id}").
	WithPath("tr", "/sepet/{id}").
	WithMethods(http.MethodPost).
	WithHandler(addToBasket).
	Build())
```

The layout links the shop's stylesheet, marks the head's area and carries the
language switcher:

```html
<!-- templates/layouts/master.html -->
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
  {{hoist "head"}}
</head>
<body>
  <nav>
    {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
    {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
  </nav>
  {{slot "content"}}
</body>
</html>
```

The product's template names every page, fragment and action it links, and never
writes a path:

```html
<!-- templates/pages/product.html -->
{{stylesheet "/static/product.css"}}
<article class="product">
  <a href="{{pageURL "catalogue"}}">{{.Labels.Back}}</a>
  <h1>{{.Name}}</h1>
  <img src="{{asset .Image}}" alt="">
  <div data-live="{{fragmentURL "product" "product-stock" "id" .ID}}">{{slot "stock"}}</div>
  <form method="post" action="{{actionURL "basket-add" "id" .ID}}">
    {{csrfToken}}
    <button>{{.Labels.Add}}</button>
  </form>
</article>
```

```html
<!-- templates/fragments/stock.html -->
<p class="stock">{{.}}</p>
```

### What the reader gets

A request for `/tr/urunler/42` answers 200. The page reads as follows, with blank
lines left out and the token shortened:

```html
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="/static/app.6d6068180a5c710c.css">
  <title>Pirinç güverte lambası</title><link rel="stylesheet" href="/static/product.e38a7cd1dac68eec.css">
</head>
<body>
  <nav>
    <a hreflang="en" href="/products/42">English</a>
    <a hreflang="tr" href="/tr/urunler/42">Türkçe</a>
  </nav>
<article class="product">
  <a href="/tr/urunler">Tüm ürünler</a>
  <h1>Pirinç güverte lambası</h1>
  <img src="/static/products/42.04115f7d8dde41ce.svg" alt="">
  <div data-live="/tr/urunler/42/stok"><p class="stock">12</p>
</div>
  <form method="post" action="/tr/sepet/42">
    <input type="hidden" name="_csrf" value="…">
    <button>Sepete ekle</button>
  </form>
</article>
</body>
</html>
```

The same templates render `/products/42` in English. The name and the labels come
from the data; the URLs change with the render's locale:

| In the template | On `/products/42` | On `/tr/urunler/42` |
| --- | --- | --- |
| `pageURL "catalogue"` | `/products` | `/tr/urunler` |
| `fragmentURL "product" "product-stock" "id" .ID` | `/products/42/stock` | `/tr/urunler/42/stok` |
| `actionURL "basket-add" "id" .ID` | `/basket/42` | `/tr/sepet/42` |
| `localeURL "en"`, `localeURL "tr"` | `/products/42`, `/tr/urunler/42` | the same |
| `asset .Image` | `/static/products/42.04115f7d8dde41ce.svg` | the same |

- **The title and `product.css` land at the marker**, after `app.css`, which the
  layout writes above it. The title comes first: the handler declared it before the
  template ran.
- **`asset` takes a path from the data** as readily as a literal one. A picture
  missing from the mount is `ErrUnknownAsset`, and since the product is required,
  the page answers 500 instead of showing a broken image.
- **The token is each reader's own.** Were this page cached, the cached bytes
  would hold the placeholder, and every reader would get their own token in its
  place.
