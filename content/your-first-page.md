---
description: A hands-on tutorial: build a recipe page with a layout, a data handler and a template, add a second fragment in a slot, and export every recipe to static files.
reference: New, NewPage, NewFragment, FragmentBuilder.WithData, FragmentBuilder.WithTitle, PageBuilder.WithStaticParams, Registrable, LoadTagged, Load, Value, ErrNotFound, ErrUnknownSlot, ErrTemplateType
---

# Your first page

This tutorial builds a small recipe site in a fresh project. First comes a page at
`/recipes/{slug}` that loads a recipe and renders it inside the site's layout.
Then a second fragment, a list of the other recipes, goes in a slot of that page.
Last, a static export writes a file for every recipe.

It takes about fifteen minutes, and every piece of it is one you will use on every
page you write.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 230" role="img" aria-labelledby="yfp-page-title">
<title id="yfp-page-title">The recipe page at /recipes/pancakes as a tree. The layout's content slot holds the recipe fragment, which is required; its handler loadRecipe reports the tag recipe:pancakes and it renders the title, minutes and ingredients. The recipe template calls a slot named more, which holds the more fragment: its handler loadMore reports the tag recipes, it lists Flatbread and Omelette, and if it fails it is left out.</title>
<rect class="d-frame" x="4" y="4" width="312" height="222" rx="8" />
<text x="14" y="21">layout</text>
<text class="d-small" x="306" y="21" text-anchor="end">/recipes/pancakes</text>
<rect class="d-slot d-slot-a" x="14" y="30" width="292" height="186" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="44">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="52" width="272" height="154" rx="6" />
<text class="d-slot-text-a" x="34" y="69">recipe</text>
<rect class="d-chip-hit" x="196" y="57" width="90" height="18" rx="9" /><text class="d-hit-text d-small" x="241" y="69" text-anchor="middle">Required()</text>
<text class="d-small" x="34" y="88">loadRecipe · tag recipe:pancakes</text>
<text class="d-small" x="34" y="102">title · minutes · ingredients</text>
<rect class="d-slot d-slot-b" x="34" y="112" width="252" height="84" rx="4" />
<text class="d-small" x="44" y="126">slot "more"</text>
<rect class="d-b" x="44" y="134" width="232" height="30" rx="4" /><text class="d-on-b" x="160" y="153" text-anchor="middle">more: Flatbread, Omelette</text>
<text class="d-small" x="44" y="182">loadMore · tag recipes · fails: left out</text>
</svg>
<figcaption>The page this tutorial builds, at <code>/recipes/pancakes</code>. The site's layout puts the page's content fragment, <code>recipe</code>, in its <code>content</code> slot. <code>recipe</code> is required: its handler <code>loadRecipe</code> loads the recipe and reports the tag <code>recipe:pancakes</code>, and if it fails, the page fails. Its own template calls a slot, <code>more</code>, which holds a list of the other recipes, loaded by <code>loadMore</code> with the tag <code>recipes</code>. That fragment has no failure policy, so if it fails, the page is served without it.</figcaption>
</figure>

You need Go and the `collage` CLI; see [Installation](/docs/installation).

## Start a project

```sh
collage new cookbook
cd cookbook
go mod tidy
collage dev
```

`collage new` writes the minimal project: one layout around one page. Its
`go.mod` names the module `cookbook`, which is why the project's own packages are
imported as `cookbook/data/recipes`, `cookbook/fragments/pages` and so on. It
requires nothing yet; `go mod tidy` adds collage itself.

Open [http://localhost:6060](http://localhost:6060) and leave `collage dev`
running. From here on, every Go file you save is rebuilt and the browser reloads
itself; every template you save shows up on the next render without a rebuild.

## Look at the layout

A minimal project already has one layout, and every page uses it. It is two
files. The fragment, in `fragments/layouts/master.go`:

```go
func Master() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/master.html").
		WithTitle("cookbook").
		Build()
}
```

And its template, `templates/layouts/master.html`:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  {{hoist "head"}}
  <link rel="stylesheet" href="{{asset "/static/app.css"}}">
</head>
<body>
  {{slot "content"}}
</body>
</html>
```

`{{slot "content"}}` is a slot named `content`, and where it renders. The Go side
declares nothing: a template calling a slot is all the declaring it needs.

You never fill this slot yourself either: when a page is registered, collage puts
the page's content fragment into it. That is what makes one layout shareable by
every page.

There is no `<title>` in the template. The layout declares one with `WithTitle`,
and `{{hoist "head"}}` is where it lands. A page that declares its own title
replaces the site's rather than adding a second one — the recipe page will, in a
moment. See [Head and SEO](/docs/head-and-seo#keys-and-the-innermost-wins).

## Look at the home page

A page is two files, one beside the other: the page in `pages/`, and its content
in `fragments/pages/`, each in an area's directory below them when the site groups
its pages (`pages/blog/`). The home page has no area. Its content,
`fragments/pages/home.go`, shows the shortest way to give a template data:

```go
// Types data used on this page: what templates/pages/home.html renders as ".".
type homeView struct {
	Name string
}

func Home() *collage.Fragment {
	return collage.NewFragment("home", "pages/home.html").
		WithData(collage.Value(homeView{Name: "cookbook"})).
		Build()
}
```

And the page, `pages/home.go`, puts that content in the layout at a path:

```go
// No Static() needed: nothing here fetches per render, so the page is static.
func Home() *collage.Page {
	return collage.NewPage("home").
		WithLayouts(layouts.Master()).
		WithContent(fragments.Home()).
		WithPath("en", "/").
		Build()
}
```

`fragments` there is the content's package, imported as
`fragments "cookbook/fragments/pages"`: both packages are called after what they
hold, `pages` and `fragments`, and an area's directory below them says which area.

`collage.Value` hands the template the same value on every render, and in
`templates/pages/home.html` that value is `.`:

```html
<h1>Hello from {{.Name}}</h1>
```

Add a field to `homeView`, set it in `collage.Value(...)` and use it in the template,
and the page shows it. That is all a fixed piece of data — a list of links, a
heading — needs.

The page declares no strategy, and it is static all the same, as the comment says:
a page that declares none is static when nothing it renders fetches per render.
Fixed data and a fixed title fetch nothing, so the home page is rendered once and
kept while the cache is on, and `collage export` writes it to a file.

A recipe is not fixed: it depends on the URL, and it comes from somewhere. That
takes a function.

## Where the content comes from

A real site loads recipes from a database or a CMS. Here, a map will do. State
lives in `data/<domain>/`; create `data/recipes/recipes.go`:

```go
// Package recipes is where this site's content comes from.
package recipes

import (
	"context"
	"fmt"
	"sort"

	"github.com/Elagoht/collage/pkg/collage"
)

type Recipe struct {
	Slug        string
	Title       string
	Minutes     int
	Ingredients []string
}

var all = map[string]Recipe{
	"pancakes":  {Slug: "pancakes", Title: "Pancakes", Minutes: 20, Ingredients: []string{"flour", "milk", "eggs", "butter"}},
	"omelette":  {Slug: "omelette", Title: "Omelette", Minutes: 10, Ingredients: []string{"eggs", "butter", "salt"}},
	"flatbread": {Slug: "flatbread", Title: "Flatbread", Minutes: 30, Ingredients: []string{"flour", "water", "salt", "olive oil"}},
}

// Get returns the recipe with slug, or an error wrapping collage.ErrNotFound.
func Get(ctx context.Context, slug string) (Recipe, error) {
	if err := ctx.Err(); err != nil {
		return Recipe{}, err
	}
	recipe, ok := all[slug]
	if !ok {
		return Recipe{}, fmt.Errorf("recipes: no recipe %q: %w", slug, collage.ErrNotFound)
	}
	return recipe, nil
}

// List returns every recipe, by title.
func List(ctx context.Context) ([]Recipe, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	list := make([]Recipe, 0, len(all))
	for _, recipe := range all {
		list = append(list, recipe)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Title < list[j].Title })
	return list, nil
}
```

Both fragments, and the export at the end, read recipes through these two
functions, so where they come from can change without touching a page.

Two details matter later. A missing recipe is reported by wrapping
`collage.ErrNotFound` with `%w`, which is how the framework tells "this does not
exist" (a 404) from "this broke" (a 500). And both functions take a
`context.Context`, because a data handler's time is bounded through its context.

## The content fragment

A fragment is a template plus, optionally, a function that fetches what the
template renders. `collage add` writes a page in the project's layout and registers
it, so start there:

```sh
collage add page recipes/recipe --file --path '/recipes/{slug}'
```

```
wrote  fragments/pages/recipes/recipe.go
wrote  templates/pages/recipes/recipe.html
wrote  pages/recipes/recipe.go
edited routes.go
page "recipe" at /recipes/{slug}
```

The page is `"recipe"`, in the `recipes` area, and `--file` keeps its markup in a
template file rather than inline in Go. What it wrote renders a placeholder title.
Replace the content, `fragments/pages/recipes/recipe.go`, with a recipe:

```go
package fragments

import (
	"context"

	"cookbook/data/recipes"

	"github.com/Elagoht/collage/pkg/collage"
)

// Returns the recipe page's content.
func Recipe() *collage.Fragment {
	return collage.NewFragment("recipe", "pages/recipes/recipe.html").
		WithData(collage.LoadTagged(loadRecipe)).
		Required().
		Build()
}

// loadRecipe is the content fragment's data handler.
func loadRecipe(ctx context.Context, rc *collage.RenderContext) (recipes.Recipe, []string, error) {
	recipe, err := recipes.Get(ctx, rc.Param("slug"))
	if err != nil {
		return recipes.Recipe{}, nil, err
	}
	rc.HoistTitle(recipe.Title + " — cookbook")
	return recipe, []string{"recipe:" + recipe.Slug}, nil
}
```

The page, `pages/recipes/recipe.go`, is as `collage add` wrote it:

```go
func Recipe() *collage.Page {
	return collage.NewPage("recipe").
		WithLayouts(layouts.Master()).
		WithContent(fragments.Recipe()).
		WithPath("en", "/recipes/{slug}").
		Build()
}
```

Take it a line at a time.

- **`NewFragment("recipe", "pages/recipes/recipe.html")`** names the fragment and
  its template. The template path is relative to `templates/`, extension included.
- **`WithData(collage.LoadTagged(loadRecipe))`** gives the fragment its data
  handler. `collage.LoadTagged` takes a function that returns its data in its own
  type, here a `recipes.Recipe`, and that type is what the template is checked
  against when the page is registered.
- **The handler returns three things**: the data, the dependency tags it was built
  from, and an error. The data is the `recipes.Recipe`, and it is what the template
  receives as `.`. The tag `recipe:pancakes` says "this page shows the pancakes
  recipe", which is what lets a cached copy be thrown away when that recipe
  changes. A loader with no tags to report uses `collage.Load` instead, as the one
  `collage add` wrote did; see
  [Data handlers](/docs/data-handlers#handlers-with-no-tags).
- **`rc.Param("slug")`** is the `{slug}` the URL matched.
- **`rc.HoistTitle`** gives the page its own `<title>`. The content fragment sits
  inside the layout, and the innermost declaration wins, so it replaces the
  layout's `cookbook`.
- **`Required()`** says the page cannot exist without this fragment. Its failure
  fails the page, and its `ErrNotFound` makes the page a 404.
- **`NewPage("recipe")`** names the page. The name is how links, tests and plugins
  refer to it, so it should not change when the URL does.
- **`WithPath("en", "/recipes/{slug}")`** is the URL, in the site's default locale.

This page declares no strategy either, but unlike the home page it has a data
handler, so it is dynamic: rendered on every request, never cached. collage cannot
see inside `loadRecipe` to tell whether it reads the request or the clock, so it
does not guess that its output is the same for everyone. That is the right place
to be while you are building it; once it works, the page will say so itself —
see [Export it](#export-it) below.

## The template

Replace `templates/pages/recipes/recipe.html` with:

```html
<main class="recipe">
  <h1>{{.Title}}</h1>
  <p>Ready in {{.Minutes}} minutes.</p>

  <h2>Ingredients</h2>
  <ul>
    {{range .Ingredients}}<li>{{.}}</li>{{end}}
  </ul>
</main>
```

`.` is the `recipes.Recipe` the handler returned. This is Go's `html/template`, so
every value is escaped for where it appears; a recipe titled
`<script>` would be printed, not run. See [Templates](/docs/templates).

## Register the page

A page does nothing until the application knows about it. `collage add` already
told it: `routes.go` now lists the recipe page beside the home page.

```go
func register(app *collage.App) error {
	return app.Register(
		pages.Home(),
		recipespages.Recipe(),
	)
}
```

`app.Register` takes pages, documents and actions, and registers each in order. A
page written by hand goes here the same way: one more line in the call, and its
area's package in the imports under an alias like `recipespages`.

Registration is where mistakes are caught. A template path with a typo, a required
slot with nothing in it, two pages with one name, a malformed path — each one stops
the program at startup with an error naming the page, rather than waiting for the
first visitor to find it.

## See it

Save, and watch the terminal: `collage dev` rebuilds and restarts. Now open
[localhost:6060/recipes/pancakes](http://localhost:6060/recipes/pancakes).

The page is the layout with your fragment in its `content` slot. Edit
`templates/pages/recipes/recipe.html` — add a sentence, change a heading — and the browser
reloads with the change; no rebuild happened, because templates are read from disk
on every request in development.

Now try [/recipes/lasagne](http://localhost:6060/recipes/lasagne). `recipes.Get`
wrapped `collage.ErrNotFound`, the fragment is `Required()`, so the response is a
404. The body is collage's own plain not-found page, because the project has not
registered one; `app.RegisterNotFoundPage` gives the site its own — see
[Pages and layouts](/docs/pages-and-layouts#not-found-and-error-pages).

Break something on purpose to see what failure looks like. Change `{{.Title}}` to
`{{.Name}}` in the template, save, and reload: the field does not exist, the
required fragment fails, and the page is a 500.

In development the error page
names `recipe` as the fragment where the failure started and prints the whole
error chain, down to `pages/recipes/recipe.html:2:8` and the field it could not
find. Put it back.

That is the development server reparsing a template it has already checked. The
next time the program starts — the next change to its Go code, `collage check`, or
a test that registers the page — it does not get that far: the handler returns a
`recipes.Recipe`, which has no `Name`, so `RegisterPage` refuses the page with a
`collage.ErrTemplateType` naming the fragment, the file, the line and the column.
See [How templates are checked](/docs/data-handlers#how-templates-are-checked).

## Add a second fragment in a slot

A page is rarely one piece. Add a list of the other recipes, as its own fragment
with its own data. `collage add fragment` writes a fragment without a page of its
own, where its path says; `pages/recipes/more` puts it beside the page's content,
in the same package:

```sh
collage add fragment pages/recipes/more --file
```

Replace `fragments/pages/recipes/more.go` with:

```go
package fragments

import (
	"context"

	"cookbook/data/recipes"

	"github.com/Elagoht/collage/pkg/collage"
)

// moreView is what the "more" fragment renders.
type moreView struct {
	Recipes []recipes.Recipe
}

// Returns the list of every recipe but the one on the page.
func More() *collage.Fragment {
	return collage.NewFragment("more", "pages/recipes/more.html").
		WithData(collage.LoadTagged(loadMore)).
		Build()
}

func loadMore(ctx context.Context, rc *collage.RenderContext) (moreView, []string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return moreView{}, nil, err
	}
	var view moreView
	for _, recipe := range list {
		if recipe.Slug != rc.Param("slug") {
			view.Recipes = append(view.Recipes, recipe)
		}
	}
	return view, []string{"recipes"}, nil
}
```

`rc.Param("slug")` works here too: parameters belong to the request, not to the
fragment whose page declared the path.

Its template, `templates/pages/recipes/more.html`:

```html
<aside class="more-recipes">
  <h2>More recipes</h2>
  <ul>
    {{range .Recipes}}
      <li><a href="{{pageURL "recipe" "slug" .Slug}}">{{.Title}}</a></li>
    {{end}}
  </ul>
</aside>
```

`pageURL "recipe" "slug" .Slug` builds the link from the page's name rather than
from its path, so it follows the page if `/recipes/{slug}` ever becomes
`/r/{slug}`. See [Links and locales](/docs/links-and-locales).

The list is about the recipe on the page, so it belongs to the recipe page, not to
the layout every page shares. Put it in a slot of the recipe fragment. In
`fragments/pages/recipes/recipe.go`, where both live in one package:

```go
return collage.NewFragment("recipe", "pages/recipes/recipe.html").
	WithData(collage.LoadTagged(loadRecipe)).
	WithSlotFragment("more", More()).
	Required().
	Build()
```

And render the slot where it belongs, in `templates/pages/recipes/recipe.html`, before
`</main>`:

```html
  {{slot "more"}}
```

Nothing declares the slot: the template's `{{slot "more"}}` does, as the layout's
`{{slot "content"}}` did. Registration checks the binding against it, so a typo on
either side — `WithSlotFragment("mroe", ...)` — stops the program with
`ErrUnknownSlot`, naming the slot and the ones the template does call.

Save the files. The recipe page now has a list of the other two recipes, each
linking to its own page: the page in the diagram at the top.

Three things are true of this page that were not written down anywhere.

- **The list is optional.** A slot is optional unless `WithSlot` makes it
  required, and `More` is not `Required()`. If `loadMore` fails, the page is
  served without the list — and, in development, with a panel saying which fragment
  failed and why. A broken sidebar is a missing sidebar, not a 500.
- **The page's tags are both fragments' tags.** The page now depends on
  `recipe:pancakes` and on `recipes`. Once it is cached, invalidating `recipes`
  re-renders every recipe page — which is what adding a recipe should do — and
  invalidating `recipe:pancakes` only this one.
- **The two handlers did not wait on each other more than they had to.** A child's
  handler starts once its parent's has returned; siblings in the slots of one
  fragment start together. Give the recipe fragment a second slot with a slow
  fragment in it and the page waits for the slowest of them, not for their sum.
  See [Data handlers](/docs/data-handlers#when-handlers-run).

## Export it

The recipes do not change between requests: they change when you edit the map and
deploy. So the page does not need to render per request, and the whole site can be
static files. Two lines in `pages/recipes/recipe.go` say so, with `recipes`
imported from `cookbook/data/recipes` and `context` from the standard library:

```go
	return collage.NewPage("recipe").
		WithLayouts(layouts.Master()).
		WithContent(fragments.Recipe()).
		WithPath("en", "/recipes/{slug}").
		Static().
		WithStaticParams(recipeParams).
		Build()
}

// recipeParams lists the recipes a static build writes a page for.
func recipeParams(ctx context.Context, locale string) ([]map[string]string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return nil, err
	}
	params := make([]map[string]string, 0, len(list))
	for _, recipe := range list {
		params = append(params, map[string]string{"slug": recipe.Slug})
	}
	return params, nil
}
```

- **`Static()`** is what collage could not decide on its own. The page has data
  handlers, and only you know that what they return is the same for every reader.
  Served, the page now renders once and is kept until its tags are invalidated.
- **`WithStaticParams`** answers the question a build cannot answer from the path:
  which recipes are there? `/recipes/{slug}` is one page with many URLs, and the
  function returns one map of placeholder values per URL. It is called once per
  locale the page has a path in; this site has one.

Export:

```sh
collage export
```

```
+ 6 files written
    dist/index.html
    dist/recipes/flatbread/index.html
    dist/recipes/omelette/index.html
    dist/recipes/pancakes/index.html
    dist/static/app.css
    dist/static/app.f106e88ebb47f51d.css

6 written · 0 skipped · 0 failed
```

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 268" role="img" aria-labelledby="yfp-export-title">
<title id="yfp-export-title">collage export writes the home page, which has no handler, to index.html. The recipe page, with Static() and WithStaticParams, returns three slugs and is written to recipes/flatbread/index.html, recipes/omelette/index.html and recipes/pancakes/index.html. The static mount is copied to static/app.css and to static/app.hash.css.</title>
<defs><marker id="yfp-export-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="100" y="8" width="120" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">collage export</text>
<text class="d-small" x="60" y="54" text-anchor="middle">from</text>
<text class="d-small" x="225" y="54" text-anchor="middle">written to dist/</text>
<rect class="d-pill" x="10" y="64" width="100" height="26" rx="13" /><text x="60" y="81" text-anchor="middle">home</text>
<text class="d-small" x="60" y="104" text-anchor="middle">no handler</text>
<line class="d-flow" x1="110" y1="77" x2="136" y2="77" marker-end="url(#yfp-export-arrow)" />
<rect class="d-folder" x="140" y="64" width="170" height="26" rx="4" /><text class="d-small" x="225" y="81" text-anchor="middle">index.html</text>
<rect class="d-a" x="10" y="118" width="100" height="26" rx="4" /><text class="d-on-a" x="60" y="135" text-anchor="middle">recipe</text>
<text class="d-small" x="60" y="160" text-anchor="middle">Static()</text>
<text class="d-small" x="60" y="173" text-anchor="middle">WithStaticParams</text>
<text class="d-small" x="60" y="186" text-anchor="middle">3 slugs</text>
<line class="d-flow" x1="110" y1="131" x2="136" y2="130" marker-end="url(#yfp-export-arrow)" />
<line class="d-flow" x1="110" y1="131" x2="136" y2="162" marker-end="url(#yfp-export-arrow)" />
<line class="d-flow" x1="110" y1="131" x2="136" y2="194" marker-end="url(#yfp-export-arrow)" />
<rect class="d-folder" x="140" y="118" width="170" height="24" rx="4" /><text class="d-small" x="225" y="134" text-anchor="middle">recipes/flatbread/index.html</text>
<rect class="d-folder" x="140" y="150" width="170" height="24" rx="4" /><text class="d-small" x="225" y="166" text-anchor="middle">recipes/omelette/index.html</text>
<rect class="d-folder" x="140" y="182" width="170" height="24" rx="4" /><text class="d-small" x="225" y="198" text-anchor="middle">recipes/pancakes/index.html</text>
<rect class="d-pill" x="10" y="224" width="100" height="26" rx="13" /><text x="60" y="241" text-anchor="middle">static/</text>
<line class="d-flow" x1="110" y1="237" x2="136" y2="225" marker-end="url(#yfp-export-arrow)" />
<line class="d-flow" x1="110" y1="237" x2="136" y2="251" marker-end="url(#yfp-export-arrow)" />
<rect class="d-folder" x="140" y="214" width="170" height="22" rx="4" /><text class="d-small" x="225" y="229" text-anchor="middle">static/app.css</text>
<rect class="d-folder" x="140" y="240" width="170" height="22" rx="4" /><text class="d-small" x="225" y="255" text-anchor="middle">static/app.&lt;hash&gt;.css</text>
</svg>
<figcaption>What <code>collage export</code> writes for the cookbook. The home page has no data handler, so it is static and becomes <code>dist/index.html</code>. The recipe page is one page with many URLs: <code>Static()</code> lets it be prerendered, and <code>WithStaticParams</code> returns one <code>slug</code> per recipe, so each recipe becomes <code>dist/recipes/&lt;slug&gt;/index.html</code>, the file a static host serves for that path. The mounted <code>static/</code> files are copied under their own names and under the hashed names <code>{{asset}}</code> links to.</figcaption>
</figure>

A file for each recipe, rendered by the same handlers a request would run, with
the `slug` each one would have carried. The home page is there without a word
from you: it has no handler, so it was static all along.

Without `WithStaticParams` the recipe page would be skipped and named in the
report — a build is not wrong for containing a page it cannot prerender — and
without `Static()` it would be skipped as dynamic.

`collage serve` shows `dist/` as a static host would; see
[Static export](/docs/static-export).

A recipe that is not in the list is not written, but a running server still
answers it: `WithStaticParams` is read by the build alone. `/recipes/lasagne`
remains a 404 either way.

## Test it

A collage application is tested without a server: `app.Handler()` is an ordinary
`http.Handler`, and `pkg/collagetest` drives it the way a browser does — cookies
kept, forms submitted with their hidden fields, redirects there to follow.

A test is what keeps the page working after you stop looking at it. Create
`main_test.go` with a helper that builds the application through the same
`newApp` that `main` uses, and a test:

```go
package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"
)

func client(t *testing.T) *collagetest.Client {
	t.Helper()
	app, err := newApp(collage.Env{CacheDir: t.TempDir()}) // a disk cache of its own, not the last run's
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	return collagetest.New(t, app.Handler())
}

func TestRecipePage(t *testing.T) {
	c := client(t)
	page := c.Get("/recipes/pancakes").WantStatus(http.StatusOK)
	if !strings.Contains(page.Body, `href="/recipes/omelette"`) {
		t.Errorf("the page does not link to the other recipes:\n%s", page.Body)
	}
	c.Get("/recipes/lasagne").WantStatus(http.StatusNotFound)
}
```

```sh
go test ./...
```

`WantStatus` fails the test with the body when the status is not the one asked
for. One more check is worth having from the start: every `pageURL` in the
templates names a page that exists. `collage check` finds the ones that do not,
without rendering anything:

```sh
collage check
```

```
collage: check: nothing found
```

See [Testing](/docs/testing) for forms, cookies and the rest of the client, and
[the CLI](/docs/cli#collage-check) for what `check` reports.

## Where to go next

- [Update part of a page](/docs/update-part-of-a-page) — the next tutorial: a
  fragment at its own URL, a form, and a few lines of JavaScript.
- [Pages and layouts](/docs/pages-and-layouts) — paths, strategies, error pages,
  and what registration does to a page.
- [Fragments and slots](/docs/fragments-and-slots) — slots, failure policy, and
  slots filled from content.
- [Data handlers](/docs/data-handlers) — the handler contract, sharing data between
  fragments, and timeouts.
- [Caching](/docs/caching) — how a static page is thrown away when a recipe changes,
  and when a page should render per request instead.
- [Static export](/docs/static-export) — everything `collage export` writes, skips
  and warns about, and hosting the result.
