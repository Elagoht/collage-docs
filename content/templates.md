---
description: Where templates live, how they are loaded and embedded, how they reload in development, slots, built-in and custom functions, and escaping.
reference: TemplateConfig, Config
---

# Templates

Every fragment renders one template. Templates are Go's
[`html/template`](https://pkg.go.dev/html/template) — the same syntax, the same
contextual escaping — with a handful of functions added for the things a
fragment needs: its slots, links to other pages, assets, forms and the page's
`<head>`.

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  <p class="meta">{{formatTime .Published "2 January 2006"}}</p>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
```

## Where templates live

`Config.Template` says where to find them:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		Root:      "templates",
		Extension: ".html",
	},
})
```

Every file under `Root` whose extension is `Extension` is a template. `Root`
defaults to `./templates` and `Extension` to `.html`; files with any other
extension are ignored, so a `README.md` beside the templates is harmless.

A template's name is its path relative to `Root`, extension included, and that is
what a fragment names:

```text
templates/
├── layouts/default.html     →  "layouts/default.html"
├── pages/post.html          →  "pages/post.html"
└── fragments/author.html    →  "fragments/author.html"
```

```go
collage.NewFragment("author", "fragments/author.html")
```

### Loaded once, checked early

`collage.New` parses every template into one set, before anything is served.
Each kind of mistake is caught at the earliest point it can be:

- **A root that does not exist** fails `New` with `ErrTemplateRootMissing` — the most
  common startup failure there is, and worth telling apart from a template that
  does not parse.
- **A template that does not parse**, or that calls a function nobody registered,
  fails `New` with the file named.
- **A fragment naming a template that was not loaded** fails `RegisterPage` with
  `ErrTemplateNotFound`, naming the page, the fragment and the path.
- **A template that fails while executing** — a field the data does not have, a
  function returning an error — fails its fragment, and the fragment's
  [failure policy](/docs/fragments-and-slots#when-a-fragment-fails) applies.
  Output is buffered, so a template that fails halfway writes nothing rather than
  half a fragment.

Because every template is in one set, one template can include another by name:

```html
{{template "partials/byline.html" .}}
```

The names given with `{{define}}` share that one namespace across every file, so
make them distinctive.

With `Root` alone, templates are read from disk through `os.OpenRoot`, so a
symlink leading out of the directory is refused rather than followed: `New` fails
with `ErrTemplateEscapesRoot`, naming the file.

### Embedding templates in the binary

A binary that reads its templates from `./templates` only runs from a directory
that has them. Embed them instead, and it runs from anywhere — a container, a
systemd unit, a copy on a server:

```go
//go:embed all:templates
var templatesFS embed.FS

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		FS:        templatesFS,
		Root:      "templates",
		Extension: ".html",
	},
})
```

With `FS` set, `Root` is a directory inside that filesystem. It is still stripped
from every name, which is why it is needed: `embed.FS` names files by their path in
the source tree, and without `Root` every template would be called
`templates/pages/post.html`. An empty `Root` with an `FS` means the root of the
filesystem. The `all:` prefix makes the embed include files whose names begin with
`.` or `_`, which a plain `//go:embed templates` would skip.

This is what `collage new` scaffolds.

## Reloading in development

In development — `Config.DevMode`, or `Config.Template.DevMode` on its own — every
template is reparsed before each render. Save a template, and the next request uses
it. The framework does not read `COLLAGE_DEV`: `collage dev` sets it, and the
scaffolded `main.go` turns it into `Config.DevMode`.

An embedded set could not do that: its bytes were fixed when the binary was built,
and reparsing them changes nothing. So **in development the directory on disk wins
over the embedded copy** whenever `Root` exists as a directory relative to the
working directory, and the application logs that it chose it. Under `collage dev`,
which runs in your project directory, it always does.

Two other things make an edit show up at once:

- The page cache is never read in development, so a page cached before the edit
  does not hide it.
- Every development page carries a small script that reloads the browser when a
  template or a static file changes — or a file in a directory named in
  [`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0), for content
  such as Markdown that your handlers read from disk.

Since the whole set is reparsed, a syntax error in any one template fails every
render until it is fixed — in development, where you will see it straight away,
with the file and line in the error.

## What a template receives

`.` is exactly what the fragment's data handler returned — with
`collage.DataHandler`, `collage.Load` or `collage.Data`, a value of your own type. A fragment without a data handler,
or one adapted with `collage.Effect`, renders with no data.

```go
type postView struct {
	Title     string
	Published time.Time
	Body      template.HTML // already sanitised, see below
	Tags      []string
}
```

```html
<h1>{{.Title}}</h1>
{{with .Tags}}<p>Tagged {{join ", " .}}</p>{{end}}
```

A template sees its own fragment's data only. A parent's data is not in scope in a
child, and there is no global site object: something every template needs, like
the site's name, is data a handler returns or a function you register.

## Slots

`{{slot "name"}}` renders what the fragment's slot of that name holds, in order,
and inserts it as HTML. The children's output is not escaped a second time; each
child escaped its own values when it rendered.

```html
<main>{{slot "content"}}</main>
<aside>{{slot "sidebar"}}</aside>
```

- A slot that holds nothing renders nothing.
- A name the fragment never declared with `WithSlot` is an error that fails the
  fragment — a typo would otherwise be a section silently missing from the page.
- A slot inside `{{if}}` that the template skips is not rendered, though its
  fragments' data handlers were already started; see
  [Fragments and slots](/docs/fragments-and-slots#a-slot-the-template-skips).

The `slot` function belongs to the fragment executing the template, so the same
`{{slot "content"}}` means a different slot in each fragment that writes it.

## Built-in functions

These are available in every template, alongside `html/template`'s own (`printf`,
`len`, `index`, `eq` and the rest):

| Function | Does |
| --- | --- |
| `slot "name"` | Renders the fragments in one of this fragment's slots |
| `pageURL "name" "param" value …` | The URL of a registered page or document, in this render's locale |
| `pageURLIn "locale" "name" …` | The same, in exactly the locale given |
| `localeURL "locale"` | This page in another locale, or empty when it has no path there |
| `asset "/static/app.css"` | A mounted file's content-addressed URL |
| `stylesheet "/static/app.css"` | Declares a stylesheet for the page's `<head>` |
| `hoist "head"` | Where declarations for an area of the page land |
| `csrfToken` | The hidden input carrying a form's forgery token |
| `safeHTML`, `safeURL` | Mark a string as trusted HTML or URL — escape hatches |
| `dict "key" value …` | Builds a map, to pass several values to `{{template}}` |
| `default fallback value` | `value`, or `fallback` when it is empty. Strings only: another type fails the render |
| `upper`, `lower`, `title` | Case conversion |
| `join sep items` | Joins a `[]string` |
| `formatTime t layout` | Formats a `time.Time` with a Go layout |

The link functions are strict: an unknown page name or a missing parameter fails
the render instead of producing a broken link. Each function's arguments and
behaviour are in [Template functions](/docs/template-functions).

## Adding your own functions

`Config.Template.Funcs` adds functions — an `html/template` `FuncMap` — to every
template:

```go
import "html/template"

app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{
		FS:   templatesFS,
		Root: "templates",
		Funcs: template.FuncMap{
			"money": func(cents int64) string {
				return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
			},
			"readingTime": func(words int) string {
				return fmt.Sprintf("%d min read", max(1, words/200))
			},
		},
	},
})
```

```html
<p>{{money .PriceCents}} · {{readingTime .WordCount}}</p>
```

- **Set it before `New`.** `html/template` can only call a function whose name was
  known when the template was parsed, and `New` is where parsing happens. A template
  calling a name nobody registered fails in `New`, which is the point: it is a typo
  you find at startup.
- **An entry under a built-in's name replaces the built-in** — except the eight
  that depend on the render: `slot`, `hoist`, `asset`, `stylesheet`, `csrfToken`,
  `pageURL`, `pageURLIn` and `localeURL`. The render engine binds its own of each
  for every render, so overriding any of them is accepted and has no effect.
- **A function cannot see the request.** It is registered once for the whole
  program. Anything that depends on the request, the locale or the user belongs in
  the data handler, which is where the data comes from anyway.

A plugin can contribute functions too, from its `Configure` hook — see
[Writing a plugin](/docs/writing-plugins). When a plugin and `Funcs` both define a
name, `Funcs` wins: the application is the one that can see both.

## Escaping

`html/template` escapes every value for the place it appears: HTML text, an
attribute, a URL, inline JavaScript or CSS. A post titled
`<script>alert(1)</script>` is printed as text, and a `javascript:` URL in an
`href` is replaced with a harmless one. You do not escape anything by hand.

Sometimes a value really is HTML — a post body your CMS sanitised, markup your own
code produced. There are two ways to say so:

- Give the field the type `template.HTML` in your view, as `Body` is above. The
  type is the decision, made in Go, where it can be reviewed.
- Use `safeHTML` (or `safeURL` for a URL) in the template.

Either one switches escaping off for that value. **Only use them on content you
generated or have sanitised**, never on anything a user typed; an unescaped user
value is a cross-site scripting hole.

Two built-ins produce HTML that is inserted as is:

- `{{slot}}`, whose children escaped their own values.
- `{{hoist}}`, which writes what fragments declared for the page. The
  `rc.HoistTitle`, `rc.HoistMeta` and similar helpers escape what they are given;
  the underlying `rc.Hoist` inserts markup exactly as written, so escaping there is
  yours. See [Head and SEO](/docs/head-and-seo).

`pageURL` and its siblings escape the parameter values they substitute into a path,
and refuse a value of `.` or `..`, which a browser would resolve as a step up the
path.
