---
description: Every built-in template function — slot, hoist, asset, stylesheet, csrfToken, the URL functions and the string helpers — with its signature, an example and its edge cases.
reference: TemplateConfig, DefaultContentSlot
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
| [`safeHTML`](#safehtml) | `safeHTML string` | the string, trusted as HTML |
| [`safeURL`](#safeurl) | `safeURL string` | the string, trusted as a URL |
| [`dict`](#dict) | `dict key value [key value]...` | a map built from pairs |
| [`default`](#default) | `default fallback value` | `value`, or `fallback` when it is empty |
| [`upper`](#upper-and-lower) | `upper string` | the string in upper case |
| [`lower`](#upper-and-lower) | `lower string` | the string in lower case |
| [`title`](#title) | `title string` | each word capitalised |
| [`join`](#join) | `join sep items` | the items joined by `sep` |
| [`formatTime`](#formattime) | `formatTime time layout` | the time, formatted |

A function that returns an error fails the template, and the fragment fails with
it — under the fragment's own failure policy: a required fragment fails the page,
an optional one renders its fallback, or nothing when it has none. In development
mode "nothing" is an HTML comment naming the fragment and its error, and the
development panel on top of the page names every failed fragment, fallback or not.
See [Fragments and slots](/docs/fragments-and-slots).

## Bound per render

The first eight functions need the render they are part of — the fragment, the
request, the locale, the application's routes and mounts. They are registered when
templates are parsed as placeholders, so that a template may call them, and the
render engine binds the real implementation on every render. A placeholder that
somehow runs outside a render returns an error rather than guessing.

One consequence: an entry under one of these names in `Config.Template.Funcs`, or
from a plugin, is parsed but never called. The other nine can be replaced.

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

- A slot the fragment never declared with `WithSlot` is an error
  (`ErrUnknownSlot`), naming the slots it does have. Rendering nothing would turn a
  typo into a section that is silently missing.
- A slot declared but empty renders nothing — unless it was declared required. A
  required slot with nothing bound and no resolver is refused when the page is
  registered, with `ErrRequiredSlotUnfilled`, so it never reaches a render. A
  required slot filled by a resolver is checked at render time instead: a resolver
  that returns no fragments fails the render with `ErrRequiredSlotEmpty`.
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
anywhere in the tree — usually from a data handler — and `hoist` decides where it
goes:

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
  too — wrapping `ErrNoMountForAsset` when no mount's prefix covers the path.
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
reader's own token as the response is written, which is why a **cached** page can
carry a form: the cached bytes hold the placeholder, and every reader gets their
own token.

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
- **Values are strings**, and are escaped. A value of `.` or `..`, which a browser
  would resolve as a path step, is refused. Pass a number through `printf`:

```html
<a href="{{pageURL "user" "id" (printf "%d" .ID)}}">{{.Name}}</a>
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
that locale is `ErrNoPathInLocale`, and a locale no URL can reach — one not in
`Locale.Supported`, or any but the default with `DisablePathLocale` set — is
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
- It fails in a render that is not a page's.
- For the `<link rel="alternate" hreflang>` search engines read, declare them from
  Go with `rc.HoistAlternate` — see
  [Head and SEO](/docs/head-and-seo#canonical-and-alternate-links).

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
sub-template. Keys must be strings (`ErrDictKeyNotString`), and the arguments must
come in pairs (`ErrDictOddArgs`); either mistake fails the template.

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
the per-render ones — replaces it. They must be there before `New` because
`html/template` can only call a name that was in its function map when the
template was parsed; a template calling a name nobody registered fails in `New`,
not at the first request.

A plugin adds functions the same way, from its `Configure` phase; an application's
entry under the same name wins. See
[Writing a plugin](/docs/writing-plugins#template-functions).

Anything that needs the request belongs in a data handler rather than a function:
that is where the page's data comes from anyway.
