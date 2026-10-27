---
description: Where templates live, how they are loaded and embedded, how they reload in development, slots, built-in and custom functions, and escaping.
reference: TemplateConfig, Config, ErrUnknownSlot, ErrTemplateType
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
├── layouts/master.html      →  "layouts/master.html"
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
- **A fragment bound into a slot its template never calls** fails `RegisterPage`
  with `ErrUnknownSlot` — see [Slots](#slots) below.
- **A template that does not fit its data** — `{{.Titel}}` where the data's type
  has only `Title`, a method called with the wrong number of arguments — fails
  `RegisterPage` with `ErrTemplateType`, naming the file, the line and the column
  (since v0.49.0). See
  [How templates are checked](/docs/data-handlers#how-templates-are-checked).
- **A template that fails while executing** — a field of data whose type was not
  known before the render, a nil pointer, a function returning an error — fails
  its fragment, and the fragment's
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
`templates/pages/post.html`.

An empty `Root` with an `FS` means the root of the filesystem. The `all:` prefix makes the embed include files whose names begin with
`.` or `_`, which a plain `//go:embed templates` would skip.

`collage new` scaffolds the same with one filesystem, `files`, which embeds
`templates`, `static` and `plugins-config.json` together (since v0.68.0) and is
given as `Template.FS`.

## Reloading in development

In development — `Config.DevMode` — every
template is reparsed before each render. Save a template, and the next request uses
it. The framework does not read `COLLAGE_DEV`: `collage dev` sets it, and the
scaffolded `main.go` turns it into `Config.DevMode`.

An embedded set could not do that: its bytes were fixed when the binary was built,
and reparsing them changes nothing.

So **in development the directory on disk wins
over the embedded copy** whenever `Root` exists as a directory relative to the
working directory, and the application logs that it chose it. Under `collage dev`,
which runs in your project directory, it always does.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 254" role="img" aria-labelledby="tpl-src-title">
<title id="tpl-src-title">Where collage.New reads templates from. With Template.FS not set, it reads the Root directory on disk through os.OpenRoot. With FS set and DevMode off, it reads the embedded FS. With FS set and DevMode on, it reads Root on disk when Root is a directory relative to the working directory, and logs that it does; otherwise it reads the embedded FS, where edits do not show. In development every render reparses the set it reads.</title>
<defs><marker id="tpl-src-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="55" y="8" width="120" height="24" rx="12" /><text x="115" y="24" text-anchor="middle">collage.New</text>
<line class="d-flow" x1="115" y1="32" x2="115" y2="48" marker-end="url(#tpl-src-arrow)" />
<rect class="d-frame" x="40" y="52" width="150" height="26" rx="6" /><text x="115" y="69" text-anchor="middle">Template.FS set?</text>
<line class="d-flow" x1="190" y1="65" x2="221" y2="65" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="60" text-anchor="middle">no</text>
<rect class="d-a" x="225" y="50" width="93" height="30" rx="5" /><text class="d-on-a d-small" x="271.5" y="62" text-anchor="middle">Root on disk</text><text class="d-on-a d-small" x="271.5" y="74" text-anchor="middle">os.OpenRoot</text>
<line class="d-flow" x1="115" y1="78" x2="115" y2="94" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="90">yes</text>
<rect class="d-frame" x="40" y="98" width="150" height="26" rx="6" /><text x="115" y="115" text-anchor="middle">Config.DevMode?</text>
<line class="d-flow" x1="190" y1="111" x2="221" y2="111" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="106" text-anchor="middle">no</text>
<rect class="d-b" x="225" y="96" width="93" height="30" rx="5" /><text class="d-on-b d-small" x="271.5" y="108" text-anchor="middle">embedded FS</text><text class="d-on-b d-small" x="271.5" y="120" text-anchor="middle">fixed at build</text>
<line class="d-flow" x1="115" y1="124" x2="115" y2="140" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="136">yes</text>
<rect class="d-frame" x="40" y="144" width="150" height="26" rx="6" /><text x="115" y="161" text-anchor="middle">Root a dir on disk?</text>
<line class="d-flow" x1="190" y1="157" x2="221" y2="157" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="205" y="152" text-anchor="middle">no</text>
<rect class="d-b" x="225" y="142" width="93" height="30" rx="5" /><text class="d-on-b d-small" x="271.5" y="154" text-anchor="middle">embedded FS</text><text class="d-on-b d-small" x="271.5" y="166" text-anchor="middle">edits won't show</text>
<line class="d-flow" x1="115" y1="170" x2="115" y2="186" marker-end="url(#tpl-src-arrow)" /><text class="d-small" x="121" y="182">yes</text>
<rect class="d-a" x="40" y="190" width="150" height="34" rx="5" /><text class="d-on-a" x="115" y="205" text-anchor="middle">Root on disk</text><text class="d-on-a d-small" x="115" y="218" text-anchor="middle">logged · edits show</text>
<text class="d-small" x="160" y="244" text-anchor="middle">In development every render reparses the set.</text>
</svg>
<figcaption>Which copy of the templates an application renders is settled in <code>collage.New</code>. Without <code>Template.FS</code> it is the <code>Root</code> directory on disk. With an <code>FS</code>, it is the embedded copy, except in development when <code>Root</code> is also a directory relative to the working directory: then that directory is read instead, and the application logs it. In development every render reparses whichever copy was chosen, so an edit on disk shows on the next request, and an edit to a directory that is not read shows only once the binary is rebuilt.</figcaption>
</figure>

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

The checks registration makes are not made
again for a template edited after startup — the check against its data's type
included, until the next restart: its failure shows on the development error page
instead, template, line and cause first.

## What a template receives

`.` is exactly what the fragment's data handler returned, or the value
`collage.Value` gave it, in the Go type the handler or the value declares. A
fragment with no data, or whose handler is adapted with `collage.Effect`, renders
with no data. See [Data handlers](/docs/data-handlers).

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
the site's name, is data its fragment is given or a function you register.

## Slots

`{{slot "name"}}` renders what the fragment's slot of that name holds, in order,
and inserts it as HTML. The children's output is not escaped a second time; each
child escaped its own values when it rendered.

```html
<main>{{slot "content"}}</main>
<aside>{{slot "sidebar"}}</aside>
```

- Calling a slot is what declares it; the fragment's Go code needs no `WithSlot`
  for it. `WithSlot` only makes a slot required, or limits it to one fragment.
- A slot that holds nothing renders nothing, whether or not anything was ever
  bound to it.
- A fragment bound into a slot its template never calls fails registration with
  `ErrUnknownSlot`, naming the slot and the slots the template does call: a typo on
  either side of the binding would otherwise be a section silently missing from
  the page. Calls in a template it includes or a block it defines count. A
  template that names a slot by anything but a literal, `{{slot .Which}}`, may
  call any of them, so its fragment is not checked.
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
| `actionURL "name" …` | An action's URL, in this render's locale |
| `fragmentURL "page" "fragment" …` | A fragment path's URL, in this render's locale |
| `fragmentURLIn "locale" "page" "fragment" …` | The same, in exactly the locale given |
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
| `withQuery "page" 2 …` | This page's URL with query parameters set; `""` removes one |
| `pathEscape s` | Escapes a string for one path segment |
| `json v`, `pre v` | A value as JSON, or in an escaped `<pre>` block to see what a template was given |
| `add`, `sub`, `mul`, `div`, `mod`, `min`, `max`; `addf`, `subf`, `mulf`, `divf` | Integer and `float64` arithmetic |
| `trim`, `replace`, `contains`, `hasPrefix`, `hasSuffix`, `split`, `truncate` | String functions, the string last so they read in a pipeline |
| `first`, `last`, `seq`, `keys`, `hasKey`, `dig` | A list's ends, a range to count over, a map's keys, a value deep in the data |

The functions from `withQuery` down are new in v0.77.0. The link functions are
strict: an unknown page name or a missing parameter fails the render instead of
producing a broken link. Each function's arguments and
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
- **An entry under a built-in's name replaces the built-in** — except the twelve
  that depend on the render: `slot`, `hoist`, `asset`, `stylesheet`, `csrfToken`,
  `pageURL`, `pageURLIn`, `localeURL`, `actionURL`, `fragmentURL`,
  `fragmentURLIn` and `withQuery`. The render engine binds its own of each
  for every render, so overriding any of them is accepted and has no effect; since
  v0.77.0 `New` logs a warning naming it.
- **A function cannot see the request.** It is registered once for the whole
  program. Anything that depends on the request, the locale or the user belongs in
  the data handler, which is where the data comes from anyway.

A plugin can contribute functions too, from its `Configure` hook — see
[Writing a plugin](/docs/writing-plugins). When a plugin and `Funcs` both define a
name, `Funcs` wins: the application is the one that can see both — since v0.78.0
for a function the plugin adds per render too. A plugin cannot take a built-in's
name at all (since v0.77.0), nor one of the template language's, such as `html` or
`eq` (since v0.78.0).

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

## Example: a blog post

A blog's post page puts most of this page to work. Each part reaches the page a
different way:

| Part | Declared in | Lands in the page through |
| --- | --- | --- |
| The post | `pages/post.html` | the layout's `{{slot "content"}}` |
| The author card | `fragments/author.html` | the post's `{{slot "author"}}` |
| The post's stylesheet | `{{stylesheet}}` in the post | the layout's `{{hoist "head"}}` |
| `<title>` and description | `rc.HoistTitle`, `rc.HoistMeta` in the handler | the layout's `{{hoist "head"}}` |
| Readers' comments | the post's data | `{{.Text}}`, escaped |

The handlers read the post's slug from the path. The `blog` package stands for your
own data layer:

```go
type comment struct {
	Name string
	Text string
}

type postView struct {
	Title     string
	Published time.Time
	Body      template.HTML // rendered and sanitised by your own code
	Comments  []comment
}

type authorView struct {
	Name string
	Bio  string
}

func loadPost(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	slug := rc.Param("slug")
	p, err := blog.GetPost(ctx, slug)
	if err != nil {
		return postView{}, nil, err
	}
	rc.HoistTitle(p.Title)
	rc.HoistMeta("description", p.Summary)

	view := postView{Title: p.Title, Published: p.Published, Body: p.Body}
	for _, c := range p.Comments {
		view.Comments = append(view.Comments, comment{Name: c.Name, Text: c.Text})
	}
	return view, []string{"post:" + slug}, nil
}

func loadAuthor(ctx context.Context, rc *collage.RenderContext) (authorView, error) {
	a, err := blog.AuthorOf(ctx, rc.Param("slug"))
	if err != nil {
		return authorView{}, err
	}
	return authorView{Name: a.Name, Bio: a.Bio}, nil
}
```

The author card is a fragment of its own, bound into the post's `author` slot:

```go
layout := collage.NewFragment("layout", "layouts/master.html").Build()

author := collage.NewFragment("author", "fragments/author.html").
	WithData(collage.Load(loadAuthor)).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithSlotFragment("author", author).
	Required().
	Build()

page := collage.NewPage("post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/posts/{slug}").
	Build()

if err := app.RegisterPage(page); err != nil {
	return err
}
```

The three templates:

```html
<!-- templates/layouts/master.html -->
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  {{hoist "head"}}
</head>
<body>{{slot "content"}}</body>
</html>
```

```html
<!-- templates/pages/post.html -->
{{stylesheet "/static/post.css"}}
<article>
  <h1>{{.Title}}</h1>
  <p class="meta">{{formatTime .Published "2 January 2006"}}</p>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
  <section class="comments">
    {{range .Comments}}<p><b>{{.Name}}</b> {{.Text}}</p>{{end}}
  </section>
</article>
```

```html
<!-- templates/fragments/author.html -->
<aside class="author">
  <strong>{{.Name}}</strong>
  <p>{{.Bio}}</p>
</aside>
```

`/static/post.css` must be a file in a [mount](/docs/assets): `{{stylesheet}}`
links its content-addressed URL, and a path no mount serves fails the render.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 254" role="img" aria-labelledby="tpl-post-title">
<title id="tpl-post-title">The blog post page as rendered. The layout's hoist head marker is filled, once the whole page has rendered, with the title and the description meta the post's handler hoisted and the link to post.css the post's template declared with stylesheet. The layout's content slot holds the post fragment, and the post's author slot holds the author fragment. Inside the post, the title and the readers' comments are escaped, and the body, a template.HTML, is inserted as is.</title>
<defs><marker id="tpl-post-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="4" width="312" height="246" rx="8" />
<text x="14" y="21">layout</text><text class="d-small" x="60" y="21">layouts/master.html</text>
<rect class="d-slot d-slot-b" x="14" y="30" width="292" height="36" rx="5" />
<text class="d-small" x="24" y="51">hoist "head"</text>
<rect class="d-chip" x="104" y="39" width="60" height="18" rx="9" /><text class="d-small" x="134" y="51" text-anchor="middle">&lt;title&gt;</text>
<rect class="d-chip" x="170" y="39" width="56" height="18" rx="9" /><text class="d-small" x="198" y="51" text-anchor="middle">&lt;meta&gt;</text>
<rect class="d-chip-hit" x="232" y="39" width="74" height="18" rx="9" /><text class="d-hit-text d-small" x="269" y="51" text-anchor="middle">post.css</text>
<rect class="d-slot d-slot-a" x="14" y="74" width="292" height="170" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="88">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="96" width="272" height="140" rx="6" />
<line class="d-flow" x1="134" y1="96" x2="134" y2="61" marker-end="url(#tpl-post-arrow)" />
<line class="d-flow" x1="198" y1="96" x2="198" y2="61" marker-end="url(#tpl-post-arrow)" />
<line class="d-flow" x1="269" y1="101" x2="269" y2="61" marker-end="url(#tpl-post-arrow)" />
<text class="d-slot-text-a" x="34" y="113">post</text><text class="d-small" x="70" y="113">pages/post.html</text>
<rect class="d-chip-hit" x="206" y="101" width="84" height="18" rx="9" /><text class="d-hit-text d-small" x="248" y="113" text-anchor="middle">stylesheet</text>
<text class="d-small" x="34" y="134">h1: escaped title</text>
<rect class="d-slot d-slot-c" x="34" y="142" width="252" height="50" rx="4" />
<text class="d-small" x="44" y="156">slot "author"</text>
<rect class="d-c" x="44" y="162" width="100" height="22" rx="4" /><text class="d-on-c" x="94" y="177" text-anchor="middle">author</text>
<text class="d-small" x="154" y="177">fragments/author.html</text>
<text class="d-small" x="34" y="210">comments: escaped text</text>
<text class="d-small" x="34" y="226">body: template.HTML, as is</text>
</svg>
<figcaption>The post fragment fills the layout's <code>content</code> slot, and the author fragment fills the post's <code>author</code> slot. The title and description the post's handler hoists, and the stylesheet its template declares, go up to the layout's <code>hoist "head"</code> marker, which is filled once the whole page has rendered. Inside the post, the title and the readers' comments are escaped; the body, a <code>template.HTML</code>, is inserted as is.</figcaption>
</figure>

### What the reader gets

Take a post titled `Escaping <script> & friends`, and a comment whose text is
`<script>alert(1)</script>`. A request for `/posts/escaping` answers 200, and the
page reads, from `<head>` to `</body>` and with blank lines left out:

```html
<head>
  <meta charset="utf-8">
  <title>Escaping &lt;script&gt; &amp; friends</title><meta name="description" content="What html/template escapes for you"><link rel="stylesheet" href="/static/post.cba296a6afd593e0.css">
</head>
<body>
<article>
  <h1>Escaping &lt;script&gt; &amp; friends</h1>
  <p class="meta">18 October 2026</p>
<aside class="author">
  <strong>Ada</strong>
  <p>Writes about templates.</p>
</aside>
  <div class="body"><p>Hello <em>world</em></p></div>
  <section class="comments">
    <p><b>mallory</b> &lt;script&gt;alert(1)&lt;/script&gt;</p>
  </section>
</article>
</body>
```

- **The title is escaped in both places it appears.** `rc.HoistTitle` escapes it
  for `<title>`, and `{{.Title}}` escapes it for the `<h1>`.
- **The comment is text.** `{{.Text}}` is a `string`, so the reader's script is
  printed, not run.
- **The body is markup.** Its type is `template.HTML`, so it is inserted as is: it
  is the one value here whose safety is your code's job.
- **The stylesheet is in the head**, linked by its content-addressed URL, though
  the post's template declared it inside `<body>`.

### What registration catches

Write `{{.Titel}}` for the heading in `pages/post.html`, and `{{.Author}}` where
the comment loop means `{{.Name}}`. `RegisterPage` refuses the page with
`ErrTemplateType`, both findings at once, each with its file, line and column:

```text
collage: page "post": fragment "post" (pages/post.html:4:8): {{.Titel}}: type main.postView has no field or method Titel (did you mean Title?)
collage: page "post": fragment "post" (pages/post.html:9:31): {{.Author}}: type main.comment has no field or method Author
```

Inside `{{range .Comments}}`, `.` is a `comment`, so the second finding names that
type, not the post's.

Bind the author card with a typo instead, `WithSlotFragment("autor", author)`, and
`RegisterPage` fails with `ErrUnknownSlot`:

```text
collage: unknown slot: page "post" fragment "post" binds into slot "autor", but template "pages/post.html" never calls it; it calls [author]
```

Neither check judges escaping. `{{.Text | safeHTML}}` fits its types and passes
both, and it would print the reader's script as markup: keep `safeHTML` for content
you generated or sanitised.
