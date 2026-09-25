---
description: A hands-on tutorial — build a recipe page with a layout, a data handler and a template, then add a second fragment in a slot of the layout.
reference: New, NewPage, NewFragment, FragmentBuilder.WithData, FragmentBuilder.WithTitle, DataHandler, Load, ErrNotFound, ErrUnknownSlot
---

# Your first page

This tutorial builds a small recipe site in a fresh project: a page at
`/recipes/{slug}` that loads a recipe and renders it inside the site's layout,
and then a second fragment — a list of other recipes — placed in a slot of the
layout, so that every page shows it. It takes about fifteen minutes, and every
piece of it is one you will use on every page you write.

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
to be while you are building it. Once it works,
`Incremental(10 * time.Minute)` or `Static()` would cache it — see
[Caching](/docs/caching).

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

A page is rarely one piece. Add a list of recipes, as its own fragment with its own
data. Every page should show it, so it belongs to the layout rather than to the
recipe page. Create `fragments/layouts/more.go`, next to the layout:

```go
package layouts

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"cookbook/recipes"
)

// moreView is what the "more recipes" fragment renders.
type moreView struct {
	Recipes []recipes.Recipe
}

// MoreRecipes lists every recipe but the one on the page, if the page shows one.
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
fragment that declared the path. On the home page there is no `{slug}`, it is
empty, and the list is every recipe.

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

Now put the list in a second slot of the layout. In `fragments/layouts/main.go`:

```go
func Layout() *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithTitle("cookbook").
		WithSlotFragment("more", MoreRecipes()).
		Build()
}
```

And render the slot where it belongs, in `templates/layouts/default.html`, after
the content:

```html
<body>
  {{slot "content"}}
  {{slot "more"}}
</body>
```

`content` is the one slot registration fills. Any other slot of the layout you fill
yourself, once, and every page that uses the layout gets it. As with `content`,
nothing declares `more`: the template's `{{slot "more"}}` does. Registration checks
the binding against it, so a typo on either side — `WithSlotFragment("mroe", ...)`
— stops the program with `ErrUnknownSlot`, naming the slot and the ones the
template does call. Save the files. The recipe page now has a list of the other
two recipes, each linking to its own page, and the home page lists all three.

Three things are true of these pages that were not written down anywhere.

- **The list is optional.** A slot is optional unless `WithSlot` says it is
  required, and `MoreRecipes` is not `Required()`. If `loadMore` fails, the page is
  served without the list — and, in development, with a panel saying which fragment
  failed and why. A broken sidebar is a missing sidebar, not a 500.
- **A page's tags are all its fragments' tags.** The recipe page now depends on
  `recipe:pancakes` and on `recipes`, and the home page on `recipes`. Once they are
  cached, invalidating `recipes` drops every page that shows the list — which is
  what adding a recipe should do — and invalidating `recipe:pancakes` drops only
  the pancakes page.
- **The two handlers did not wait on each other.** A child's handler starts once
  its parent's has returned; siblings in the slots of one fragment start together.
  `content` and `more` are both slots of the layout, so `loadRecipe` and `loadMore`
  run at the same time, and the page waits for the slower of them, not for their
  sum. See [Data handlers](/docs/data-handlers#when-handlers-run).

And one thing changed that you did not touch: the home page is dynamic now. It
declares no strategy, and it was static because nothing in it fetched. The layout
is part of every page, and its list has a data handler, so every page that uses it
renders per request until it says `Static()` or `Incremental(ttl)`. See
[Caching](/docs/caching#a-page-that-declares-none).

A fragment that only one page needs belongs in that page's content fragment
instead: bind it into the content fragment, and give its template the `{{slot}}`.
Where a fragment is declared is where it shows up.

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
- [Caching](/docs/caching) — turning this page from a dynamic one into one that is
  rendered once and thrown away when a recipe changes.
