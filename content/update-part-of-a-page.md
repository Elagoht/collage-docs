---
description: A hands-on tutorial with no plugins: a shopping list whose items are a fragment at its own URL, updated in place by a form and a few lines of JavaScript.
reference: PageBuilder.WithFragmentPath, PageBuilder.WithAction, PageBuilder.Incremental, RenderFragment, SeeOther, FetchHeader, ActionResult, LoadTagged, CacheStatusHeader
---

# Update part of a page

This tutorial builds a shopping list. The list is one part of its page; adding an
item replaces that part and nothing else, and a list open in another tab picks the
item up a few seconds later. It uses no plugin and no client framework: a
fragment opened at a URL of its own, an action, and about twenty lines of plain
JavaScript. With JavaScript off, the same page still works as an ordinary form.

The finished project is in the documentation's repository, in
[`examples/shopping-list`](https://github.com/Elagoht/collage-docs/tree/main/examples/shopping-list),
and every file below is that project's file as it is. It assumes you have read
[Your first page](/docs/your-first-page).

## What you will build

- `/list`, a page with the list and a form to add to it, cached until the list
  changes.
- `/list/items`, the list alone — the fragment the page renders, at its own URL.
- A `POST /list` that adds an item. A browser posting the form is redirected back
  to `/list`; a script that asks for it is answered with the list fragment.
- A script that posts the form with `fetch`, swaps the list in, and fetches the
  list again every five seconds.

## Start a project

```sh
collage new shopping-list
cd shopping-list
go mod tidy
collage dev
```

`collage new` writes the layout, a home page and `main.go`; this tutorial adds a
package for the list, two fragments, a page, two templates and a script, and
registers the page in `routes.go`.

## The list

The list stands in for whatever your data really lives in — a database, an API.
The pages read it and the action writes to it, and nothing else in the project
knows how it is stored. Create `list/list.go`:

<!-- file: list/list.go -->

```go
// Package list keeps the shopping list. It stands in for a database: the pages
// read it and the action writes to it, and none of them know how it is stored.
package list

import (
	"strings"
	"sync"
)

// Tag is the dependency tag of everything rendered from the list. Invalidating
// it renders those pages again.
const Tag = "list"

// Item is one thing to buy.
type Item struct {
	Name string
}

// List is a shopping list several readers can add to at once.
type List struct {
	mu    sync.Mutex
	items []Item
}

// Shopping is the list the site shows.
var Shopping = New("Milk", "Bread")

// Returns a list holding names, in order.
func New(names ...string) *List {
	l := &List{}
	for _, name := range names {
		l.Add(name)
	}
	return l
}

// Returns the items, in the order they were added. The slice is a copy, so a
// template can range over it while someone else adds to the list.
func (l *List) Items() []Item {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Item(nil), l.items...)
}

// Adds name to the end of the list, and reports whether it did: a name that is
// blank once trimmed is not added.
func (l *List) Add(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, Item{Name: name})
	return true
}
```

Two things matter beyond keeping items. `Tag` is the dependency tag of everything
rendered from the list: the fragment that reads it says so, and the action that
changes it invalidates it. And the list is safe to add to from several requests at
once, because several readers will.

## The items fragment

A fragment is a template and the function that fetches what it needs. Create
`fragments/pages/list.go` beside the scaffold's `home.go`:

<!-- file: fragments/pages/list.go -->

```go
package fragments

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/list"
)

// Returns the list's items: templates/fragments/items.html. It reads the list on
// every render and tags what it rendered, so invalidating list.Tag renders it
// again.
func Items() *collage.Fragment {
	return collage.NewFragment("items", "fragments/items.html").
		WithData(collage.LoadTagged(func(context.Context, *collage.RenderContext) ([]list.Item, []string, error) {
			return list.Shopping.Items(), []string{list.Tag}, nil
		})).
		Build()
}

// Returns the list page's content: templates/pages/list.html, with items in its
// "items" slot.
func List(items *collage.Fragment) *collage.Fragment {
	return collage.NewFragment("list", "pages/list.html").
		WithTitle("Shopping list").
		WithSlotFragment("items", items).
		Build()
}
```

`Items` returns the items and the tag they came from on every render.
`List` is the page's content, with `Items` in a slot. The fragment is passed in
rather than built inside `List`, because the page will use the same fragment
twice — in the slot, and at a URL of its own — and they must be one fragment.

The list's template is the fragment's whole markup, `<ul>` included:

<!-- file: templates/fragments/items.html -->

```html
<ul id="items" data-src="{{fragmentURL "list" "items"}}">
  {{range .}}<li>{{.Name}}</li>
  {{else}}<li>Nothing to buy.</li>
  {{end}}
</ul>
```

The `id` is what the script replaces, and `data-src` is where it fetches the list
from. `fragmentURL` builds that address from the page's name and the fragment's,
so it follows the page if its path changes. `{{.Name}}` is escaped like everything
a template writes: an item named `<b>Salt</b>` shows those characters, not bold
text.

## The page

The page's template puts the list in its slot, and a form under it:

<!-- file: templates/pages/list.html -->

```html
<main>
  <h1>Shopping list</h1>

  {{slot "items"}}

  <form id="add" method="post" action="{{pageURL "list"}}">
    {{csrfToken}}
    <label>Item <input name="name" required autocomplete="off"></label>
    <button>Add</button>
  </form>

  <script src="{{asset "/static/list.js"}}" defer></script>
</main>
```

`{{csrfToken}}` renders the hidden input that carries the forgery token. A form
without it is refused with a 403; a script that posts the form sends it along
with the other fields. The form names its action with `pageURL`, so it posts to
the page's own URL, and the script comes from `static/`, where `{{asset}}` gives it
an address that changes with its content.

## Open the fragment at its own URL

Now the page itself. Create `pages/list.go`:

<!-- file: pages/list.go -->

```go
package pages

import (
	"context"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/fragments/layouts"
	fragments "shopping-list/fragments/pages"
	"shopping-list/list"
)

// Returns the shopping list page at /list, cached for an hour or until the list
// changes.
//
// Its items fragment is also opened at /list/items, so a script can fetch that
// part alone. Posting the form adds an item: a script that asks for it is
// answered with the items fragment, a browser on its own with a redirect back.
func List() *collage.Page {
	items := fragments.Items()
	return collage.NewPage("list").
		WithLayouts(layouts.Master()).
		WithContent(fragments.List(items)).
		WithPath("en", "/list").
		WithFragmentPath("en", "/list/items", items).
		Incremental(time.Hour).
		WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			list.Shopping.Add(rc.Request.PostFormValue("name"))

			result := collage.SeeOther("/list")
			if rc.Request.Header.Get(collage.FetchHeader) != "" {
				result = collage.RenderFragment(items)
			}
			result.InvalidateTags = []string{list.Tag}
			return result, nil
		}).
		Build()
}
```

`WithFragmentPath("en", "/list/items", items)` is the line this tutorial is about.
It opens the `items` fragment at `/list/items`: a GET there renders that fragment
— its data handler, its template — with no layout around it. Nothing is reachable
this way unless it is declared, and a fragment path is a part of its page: it
meets the page's guards, if the page has any.

The page is `Incremental(time.Hour)`: rendered once, then served from the cache
for an hour, or until something invalidates a tag it was built from. Without it,
the data handler would make the page render on every request — correct, but it
would leave nothing for the tag to do. A fragment path's answer is never cached,
so `/list/items` is always the list as it is now.

Register the page in `routes.go`:

<!-- file: routes.go -->

```go
package main

import (
	"shopping-list/pages"

	"github.com/Elagoht/collage/pkg/collage"
)

// Registers every page, document and action to app. A new route goes here.
func register(app *collage.App) error {
	return app.Register(
		pages.Home(),
		pages.List(),
	)
}
```

Open [http://localhost:6060/list](http://localhost:6060/list) and then
[/list/items](http://localhost:6060/list/items): the same list, once inside the
page and once on its own.

## Add an item

The action is the function given to `WithAction`. It reads the posted name, adds
it, and decides how to answer:

- A browser posting the form gets `SeeOther("/list")`: a 303, so the browser
  loads the page with a GET and a reload does not post the form again.
- A script that sends the `Collage-Fetch` header (`collage.FetchHeader`) gets
  `RenderFragment(items)`: the list, rendered with the item just added, and
  nothing else.

Either way the result carries `InvalidateTags: []string{list.Tag}`. Before the
answer is written, every cached page built from the list is marked stale: it is
rendered again in the background, or by the next GET of `/list` at the latest,
so that GET shows the new item — in this tab and in every other.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 380" role="img" aria-labelledby="add-title">
<title id="add-title">With the script, the browser posts the form with the Collage-Fetch header, collage adds the item, invalidates the list tag and answers with the items fragment, which the script puts in place of the list. Without it, the browser posts the form, is redirected to /list and loads the page again, fresh. Every five seconds the script fetches /list/items.</title>
<defs><marker id="add-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="10" y="8" width="100" height="24" rx="12" /><text x="60" y="24" text-anchor="middle">browser</text>
<rect class="d-cmd" x="210" y="8" width="100" height="24" rx="12" /><text class="d-cmd-text" x="260" y="24" text-anchor="middle">collage</text>
<line class="d-rule" x1="60" y1="32" x2="60" y2="372" />
<line class="d-rule" x1="260" y1="32" x2="260" y2="372" />
<text class="d-small" x="160" y="50" text-anchor="middle">with the script</text>
<line class="d-flow" x1="60" y1="66" x2="254" y2="66" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="61" text-anchor="middle">POST /list · Collage-Fetch</text>
<rect class="d-soft-a" x="200" y="74" width="114" height="40" rx="5" /><text class="d-small" x="257" y="90" text-anchor="middle">add the item</text><text class="d-small" x="257" y="104" text-anchor="middle">invalidate "list"</text>
<line class="d-flow" x1="260" y1="130" x2="66" y2="130" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="125" text-anchor="middle">200 · the items fragment</text>
<rect class="d-b" x="8" y="138" width="118" height="24" rx="5" /><text class="d-on-b d-small" x="67" y="154" text-anchor="middle">replace #items</text>
<line class="d-rule" x1="10" y1="180" x2="310" y2="180" />
<text class="d-small" x="160" y="196" text-anchor="middle">without the script</text>
<line class="d-flow" x1="60" y1="212" x2="254" y2="212" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="207" text-anchor="middle">POST /list</text>
<line class="d-flow" x1="260" y1="236" x2="66" y2="236" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="231" text-anchor="middle">303 · Location: /list</text>
<line class="d-flow" x1="60" y1="260" x2="254" y2="260" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="255" text-anchor="middle">GET /list</text>
<line class="d-flow" x1="260" y1="284" x2="66" y2="284" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="279" text-anchor="middle">200 · the page, rendered again</text>
<line class="d-rule" x1="10" y1="302" x2="310" y2="302" />
<text class="d-small" x="160" y="318" text-anchor="middle">every five seconds</text>
<line class="d-flow" x1="60" y1="334" x2="254" y2="334" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="329" text-anchor="middle">GET /list/items</text>
<line class="d-flow" x1="260" y1="358" x2="66" y2="358" marker-end="url(#add-arrow)" /><text class="d-small" x="160" y="353" text-anchor="middle">200 · the items fragment</text>
</svg>
<figcaption>One action, two answers. With the script, the form is answered with the list alone and the page stays where it is. Without it, the form is answered with a redirect, and the page it leads to is rendered again because the action invalidated its tag. The script also asks for the list on its own every five seconds.</figcaption>
</figure>

The name is trimmed, and a blank one is not added; the answer is the same either
way, so a script never has to tell the two apart. Answering a refused submission
with an error message is the subject of
[Failure renders, success redirects](/docs/forms-and-actions#failure-renders-success-redirects).

## Swap it in with a few lines of JavaScript

Create `static/list.js`:

<!-- file: static/list.js -->

```js
// Posts the form without leaving the page, and puts the list the server answers
// with in place of the old one. Without this script the form posts as any form
// does, and the server redirects back to the page.
const form = document.querySelector("#add");

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    const response = await fetch(form.action, {
      method: "POST",
      headers: { "Collage-Fetch": "1" },
      body: new FormData(form),
    });
    const items = document.querySelector("#items");
    if (!response.ok || !items) {
      throw new Error(`answered ${response.status}`);
    }
    items.outerHTML = await response.text();
    form.reset();
  } catch {
    // No answer, or not the list: post the form the ordinary way, and let the
    // browser show whatever the server says.
    form.submit();
  }
});

// Fetches the list again every five seconds, so an item added in another tab
// shows up here too. A tick the server does not answer is skipped.
setInterval(async () => {
  const items = document.querySelector("#items");
  if (!items) {
    return;
  }
  try {
    const response = await fetch(items.dataset.src);
    if (response.ok) {
      items.outerHTML = await response.text();
    }
  } catch {
    // Try again on the next tick.
  }
}, 5000);
```

The submit handler stops the browser's own post and makes the same one with
`fetch`. `new FormData(form)` sends every field of the form, the forgery token
among them, so there is nothing to add for it. The `Collage-Fetch` header is what
asks the action for the fragment. The answer is the list's whole `<ul>`, and
`outerHTML` puts it in place of the old one, `id` and `data-src` included, so the
next swap finds it the same way. If there is no answer — the network is down —
or it is not the list, the form is submitted the ordinary way and the browser
shows whatever the server says. The second half skips a tick the server does not
answer, and does nothing on a page without the list.

## Keep it current

The second half of the script asks for `/list/items` every five seconds. That is
the whole of "live" for a list like this: a fragment path is rendered afresh on
every request, so an item added in another tab is in the next answer. Open the
page in two tabs, add an item in one, and watch the other.

Polling suits a page that changes now and then. For one that changes often, or
should change the moment the server does, the plugin
[elagoht/live](/docs/plugins#elagohtlive) refreshes fragments marked in the
template over a server-sent event stream, against the same fragment paths you have
just declared.

## Run it

`collage dev` has been serving the project all along; open
[http://localhost:6060/list](http://localhost:6060/list). Add an item, then turn
JavaScript off in the browser and add another: the page reloads, and the list is
right both times.

Development mode does not read the cache. To see the page served from it, stop
`collage dev` and run the project the way it is deployed:

```sh
go run .
```

Open [http://localhost:6060/list](http://localhost:6060/list) with the browser's
network panel open and reload: the page's answer carries
`X-Collage-Cache: HIT`, served without a render. Add an item, and the page you
come back to shows it.

## Test it

The tests drive the application through `collagetest`, the way a browser does:
cookies kept, the form submitted with its hidden token, and a request built by
hand for what the script sends. Create `list_test.go`:

<!-- file: list_test.go -->

```go
package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"shopping-list/list"
)

// client builds the application main builds, on a fresh list and a disk cache of
// its own, and returns a browser for it.
func client(t *testing.T) *collagetest.Client {
	t.Helper()
	list.Shopping = list.New("Milk", "Bread")
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	// Shut down before the directory is removed: an invalidation renders the
	// pages it reached again in the background, into that directory.
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })
	return collagetest.New(t, app.Handler())
}

// fetchAdd posts name the way static/list.js does: the form's fields, its token
// among them, and the header that asks for the fragment instead of a redirect.
func fetchAdd(c *collagetest.Client, page *collagetest.Response, name string) *collagetest.Response {
	body := url.Values{"_csrf": {page.CSRFToken()}, "name": {name}}
	req := c.Request(http.MethodPost, "/list", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(collage.FetchHeader, "1")
	return c.Do(req)
}

func TestListPage(t *testing.T) {
	page := client(t).Get("/list").WantStatus(http.StatusOK)
	for _, want := range []string{"Milk", "Bread", `data-src="/list/items"`} {
		if !strings.Contains(page.Body, want) {
			t.Errorf("the page has no %s:\n%s", want, page.Body)
		}
	}
}

func TestItemsFragment(t *testing.T) {
	items := client(t).Get("/list/items").WantStatus(http.StatusOK)
	if strings.Contains(items.Body, "<html") || !strings.Contains(items.Body, `id="items"`) {
		t.Errorf("/list/items is not the items fragment alone:\n%s", items.Body)
	}
}

func TestAddWithForm(t *testing.T) {
	c := client(t)
	res := c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Eggs"}}).WantStatus(http.StatusSeeOther)
	if res.Location() != "/list" {
		t.Errorf("Location = %q, want /list", res.Location())
	}
	if page := c.Get("/list"); !strings.Contains(page.Body, "Eggs") {
		t.Errorf("the page does not show the item just added:\n%s", page.Body)
	}
}

func TestAddWithFetch(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "Butter").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<html") || !strings.Contains(res.Body, `id="items"`) || !strings.Contains(res.Body, "Butter") {
		t.Errorf("a fetch was not answered with the items fragment:\n%s", res.Body)
	}
}

func TestAddIgnoresBlank(t *testing.T) {
	c := client(t)
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"   "}}).WantStatus(http.StatusSeeOther)
	res := fetchAdd(c, c.Get("/list"), "").WantStatus(http.StatusOK)
	if n := strings.Count(res.Body, "<li"); n != 2 {
		t.Errorf("the list has %d items after blank names, want 2:\n%s", n, res.Body)
	}
}

func TestItemsFragmentIsFresh(t *testing.T) {
	c := client(t)
	c.Get("/list/items")
	fetchAdd(c, c.Get("/list"), "Tea").WantStatus(http.StatusOK)
	if items := c.Get("/list/items"); !strings.Contains(items.Body, "Tea") {
		t.Errorf("/list/items does not show the item just added:\n%s", items.Body)
	}
}

func TestAddWithoutTokenIsRefused(t *testing.T) {
	c := client(t)
	req := c.Request(http.MethodPost, "/list", strings.NewReader("name=Jam"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Do(req).WantStatus(http.StatusForbidden)
	if items := c.Get("/list/items"); strings.Contains(items.Body, "Jam") {
		t.Errorf("a post without a token added an item:\n%s", items.Body)
	}
}

func TestItemNameIsEscaped(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "<b>Salt</b>").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<b>") || !strings.Contains(res.Body, "&lt;b&gt;Salt") {
		t.Errorf("the item's name was not escaped:\n%s", res.Body)
	}
}

func TestPageIsCachedUntilTheListChanges(t *testing.T) {
	c := client(t)
	c.Get("/list").WantStatus(http.StatusOK)
	if page := c.Get("/list"); page.Header.Get(collage.CacheStatusHeader) != "HIT" {
		t.Errorf("the second request was not served from the cache: %s = %q", collage.CacheStatusHeader, page.Header.Get(collage.CacheStatusHeader))
	}
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Honey"}}).WantStatus(http.StatusSeeOther)
	if page := c.Get("/list"); !strings.Contains(page.Body, "Honey") {
		t.Errorf("the cached page still shows the old list:\n%s", page.Body)
	}
}
```

`client` shuts the application down when the test ends, because the invalidation
renders the pages it reached again in the background, into the test's cache
directory. Run them:

```sh
go test ./...
```

## Where to go next

- [Forms and actions](/docs/forms-and-actions) — what an action can answer with,
  forgery protection, and [a fragment at its own URL](/docs/forms-and-actions#a-fragment-at-its-own-url)
  in full.
- [Caching](/docs/caching) — strategies, tags, and what invalidating does.
- [Testing](/docs/testing) — the rest of `collagetest`.
- [Using plugins](/docs/plugins) — when polling is not enough.
