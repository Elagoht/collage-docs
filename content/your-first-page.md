---
description: A hands-on tutorial — build a recipe page with a layout, a data handler and a template, add a second fragment in a slot, and export every recipe to static files.
reference: New, NewPage, NewFragment, FragmentBuilder.WithData, FragmentBuilder.WithTitle, PageBuilder.WithStaticParams, DataHandler, Load, ErrNotFound, ErrUnknownSlot
---

# Your first page

This tutorial builds a small recipe site in a fresh project: a page at
`/recipes/{slug}` that loads a recipe and renders it inside the site's layout,
then a second fragment — a list of the other recipes — in a slot of that page,
and finally a static export with a file for every recipe. It takes about fifteen
minutes, and every piece of it is one you will use on every page you write.

You need Go and the `collage` CLI; see [Installation](/docs/installation).

## Start a project

```sh
collage new cookbook --template minimal
cd cookbook
go mod tidy
collage dev
```

Open [http://localhost:3000](http://localhost:3000) and leave `collage dev`
running. From here on, every Go file you save is rebuilt and the browser reloads
itself; every template you save shows up on the next render without a rebuild.

## Look at the layout

A minimal project already has one layout, and every page uses it. It is two
files. The fragment, in `fragments/layouts/main.go`:

```go
func Layout() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("cookbook").
		Build()
}
```

And its template, `templates/layouts/default.html`:

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
declares nothing: a template calling a slot is all the declaring it needs. You
never fill this slot yourself either: when a page is registered, collage puts the
page's content fragment into it. That is what makes one layout shareable by every
page.

There is no `<title>` in the template. The layout declares one with `WithTitle`,
and `{{hoist "head"}}` is where it lands. A page that declares its own title
replaces the site's rather than adding a second one — the recipe page will, in a
moment. See [Head and SEO](/docs/head-and-seo#keys-and-the-innermost-wins).

## Look at the home page

The home page, in `pages/home.go`, shows the shortest way to give a template data:

```go
// homeView is what templates/pages/home.html renders with, as ".".
type homeView struct {
	Name string
}

func HomePage() *collage.Page {
	content := collage.NewFragment("home-content", "pages/home.html").
		WithData(homeView{Name: "cookbook"}).
		Build()

	// No Static() needed: nothing here fetches per render, so the page is static.
	return collage.NewPage("home").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/").
		Build()
}
```

`WithData` hands the template the same value on every render, and in
`templates/pages/home.html` that value is `.`:

```html
<h1>Hello from {{.Name}}</h1>
```

Add a field to `homeView`, set it in `WithData(...)` and use it in the template,
and the page shows it. That is all a fixed piece of data — a list of links, a
heading — needs.

The page declares no strategy, and it is static all the same, as the comment says:
a page that declares none is static when nothing it renders fetches per render.
Fixed data and a fixed title fetch nothing, so the home page is rendered once and
kept while the cache is on, and `collage export` writes it to a file.

A recipe is not fixed: it depends on the URL, and it comes from somewhere. That
takes a function.

## Where the content comes from

A real site loads recipes from a database or a CMS. Here, a map will do. Create
`recipes/recipes.go`:

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

Two details matter later. A missing recipe is reported by wrapping
`collage.ErrNotFound` with `%w`, which is how the framework tells "this does not
exist" (a 404) from "this broke" (a 500). And both functions take a
`context.Context`, because a data handler's time is bounded through its context.

## The content fragment

A fragment is a template plus, optionally, a function that fetches what the
template renders. Create `pages/recipe.go`:

```go
package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/fragments/layouts"
	"cookbook/recipes"
)

// RecipePage is one recipe, at /recipes/{slug}.
func RecipePage() *collage.Page {
	content := collage.NewFragment("recipe-content", "pages/recipe.html").
		WithDataHandler(loadRecipe).
		Required().
		Build()

	return collage.NewPage("recipe").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/recipes/{slug}").
		Build()
}

// loadRecipe is the content fragment's data handler.
func loadRecipe(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	recipe, err := recipes.Get(ctx, rc.Param("slug"))
	if err != nil {
		return nil, nil, err
	}
	rc.HoistTitle(recipe.Title + " — cookbook")
	return recipe, []string{"recipe:" + recipe.Slug}, nil
}
```

Take it a line at a time.

- **`NewFragment("recipe-content", "pages/recipe.html")`** names the fragment and
  its template. The template path is relative to `templates/`, extension included.
- **`WithDataHandler(loadRecipe)`** gives the fragment its data handler: a function
  of exactly the shape `WithDataHandler` takes, no adapter in between.
- **The handler returns three things**: the data, the dependency tags it was built
  from, and an error. The data is the `recipes.Recipe`, returned as `any`, and it
  is what the template receives as `.`. The tag `recipe:pancakes` says "this page
  shows the pancakes recipe", which is what lets a cached copy be thrown away when
  that recipe changes. A loader you also call from elsewhere — a test, another
  page — can return its own type instead, through `collage.DataHandler`, or
  `collage.Load` when it has no tags; see
  [Data handlers](/docs/data-handlers#loaders-with-a-type-of-their-own).
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

Create `templates/pages/recipe.html`:

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

A page does nothing until the application knows about it. Open `routes.go` and add
`pages.RecipePage()` to the list:

```go
for _, page := range []*collage.Page{pages.HomePage(), pages.RecipePage()} {
	if err := app.RegisterPage(page); err != nil {
		return fmt.Errorf("register page %q: %w", page.Name, err)
	}
}
```

Registration is where mistakes are caught. A template path with a typo, a required
slot with nothing in it, two pages with one name, a malformed path — each one stops
the program at startup with an error naming the page, rather than waiting for the
first visitor to find it.

## See it

Save, and watch the terminal: `collage dev` rebuilds and restarts. Now open
[localhost:3000/recipes/pancakes](http://localhost:3000/recipes/pancakes).

The page is the layout with your fragment in its `content` slot. Edit
`templates/pages/recipe.html` — add a sentence, change a heading — and the browser
reloads with the change; no rebuild happened, because templates are read from disk
on every request in development.

Now try [/recipes/lasagne](http://localhost:3000/recipes/lasagne). `recipes.Get`
wrapped `collage.ErrNotFound`, the fragment is `Required()`, so the response is a
404. The body is collage's own plain not-found page, because the project has not
registered one; `app.RegisterNotFoundPage` gives the site its own — see
[Pages and layouts](/docs/pages-and-layouts#not-found-and-error-pages).

Break something on purpose to see what failure looks like. Change `{{.Title}}` to
`{{.Name}}` in the template, save, and reload: the field does not exist, the
required fragment fails, and the page is a 500. In development the error page
names `recipe-content` as the fragment where the failure started and prints the
whole error chain, down to `pages/recipe.html:2:8` and the field it could not
find. Put it back.

That is also why the handler returns `any` rather than a `recipes.Recipe`. The
template is what reads the data, and it reads it untyped: a concrete return type
would not have caught `{{.Name}}` either.

## Add a second fragment in a slot

A page is rarely one piece. Add a list of the other recipes, as its own fragment
with its own data. Create `pages/more.go`:

```go
package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/recipes"
)

// moreView is what the "more recipes" fragment renders.
type moreView struct {
	Recipes []recipes.Recipe
}

// MoreRecipes lists every recipe but the one on the page.
func MoreRecipes() *collage.Fragment {
	return collage.NewFragment("more-recipes", "fragments/more-recipes.html").
		WithDataHandler(loadMore).
		Build()
}

func loadMore(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	list, err := recipes.List(ctx)
	if err != nil {
		return nil, nil, err
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

Its template, `templates/fragments/more-recipes.html`:

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
`pages/recipe.go`:

```go
content := collage.NewFragment("recipe-content", "pages/recipe.html").
	WithDataHandler(loadRecipe).
	WithSlotFragment("more", MoreRecipes()).
	Required().
	Build()
```

And render the slot where it belongs, in `templates/pages/recipe.html`, before
`</main>`:

```html
  {{slot "more"}}
```

Nothing declares the slot: the template's `{{slot "more"}}` does, as the layout's
`{{slot "content"}}` did. Registration checks the binding against it, so a typo on
either side — `WithSlotFragment("mroe", ...)` — stops the program with
`ErrUnknownSlot`, naming the slot and the ones the template does call. Save the
files. The recipe page now has a list of the other two recipes, each linking to its
own page.

Three things are true of this page that were not written down anywhere.

- **The list is optional.** A slot is optional unless `WithSlot` makes it
  required, and `MoreRecipes` is not `Required()`. If `loadMore` fails, the page is
  served without the list — and, in development, with a panel saying which fragment
  failed and why. A broken sidebar is a missing sidebar, not a 500.
- **The page's tags are both fragments' tags.** The page now depends on
  `recipe:pancakes` and on `recipes`. Once it is cached, invalidating `recipes`
  drops every recipe page — which is what adding a recipe should do — and
  invalidating `recipe:pancakes` drops only this one.
- **The two handlers did not wait on each other more than they had to.** A child's
  handler starts once its parent's has returned; siblings in the slots of one
  fragment start together. Give the recipe fragment a second slot with a slow
  fragment in it and the page waits for the slowest of them, not for their sum.
  See [Data handlers](/docs/data-handlers#when-handlers-run).

## Export it

The recipes do not change between requests: they change when you edit the map and
deploy. So the page does not need to render per request, and the whole site can be
static files. Two lines in `pages/recipe.go` say so:

```go
	return collage.NewPage("recipe").
		WithLayout(layouts.Layout()).
		WithContent(content).
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

A file for each recipe, rendered by the same handlers a request would run, with
the `slug` each one would have carried. The home page is there without a word
from you: it has no handler, so it was static all along. Without
`WithStaticParams` the recipe page would be skipped and named in the report — a
build is not wrong for containing a page it cannot prerender — and without
`Static()` it would be skipped as dynamic. `collage serve` shows `dist/` as a
static host would; see [Static export](/docs/static-export).

A recipe that is not in the list is not written, but a running server still
answers it: `WithStaticParams` is read by the build alone. `/recipes/lasagne`
remains a 404 either way.

## Test it

A collage application is tested without a server: `app.Handler()` is an ordinary
`http.Handler`, and `net/http/httptest` drives it. Create `main_test.go` with a
helper that builds the application through the same `newApp` that `main` uses, and
a test:

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	cacheDir = t.TempDir() // a disk cache of its own, not the last run's
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestRecipePage(t *testing.T) {
	rec := get(t, "/recipes/pancakes")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /recipes/pancakes = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `href="/recipes/omelette"`) {
		t.Errorf("the page does not link to the other recipes:\n%s", rec.Body.String())
	}

	if rec := get(t, "/recipes/lasagne"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /recipes/lasagne = %d, want 404", rec.Code)
	}
}
```

```sh
go test ./...
```

## Where to go next

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
