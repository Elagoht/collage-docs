---
description: How a fragment fetches its data — the handler contract, fixed data, dependency tags, 404s, concurrency, the render context, sharing data, timeouts, and how templates are checked against their data.
reference: Data, FragmentBuilder.WithData, FragmentBuilder.WithTitle, FragmentBuilder.WithoutTypeCheck, DataHandler, Load, Value, RenderContext, Key, NewKey, Once, Effect, ErrNotFound, ErrConflictingData, TemplateTypeError, ErrTemplateType, PanicError
---

# Data handlers

A data handler is the function a fragment calls to get what its template renders.
It receives the request's context and the render context, and returns the data in
a Go type of its own, the dependency tags that data came from, and an error.
Everything about a page that talks to the outside world — a database, a CMS, an
API — happens in data handlers, and nowhere else.

```go
content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.DataHandler(func(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
		post, err := store.Post(ctx, rc.Param("slug"))
		if err != nil {
			return nil, nil, err
		}
		return post, []string{"post:" + post.Slug}, nil
	})).
	Required().
	Build()
```

## The contract

A fragment's data is set with `WithData`, which takes a `collage.Data`. Only four
constructors make one, and each says where the data comes from:

| Constructor | What it runs | The template's `.` |
| --- | --- | --- |
| `collage.DataHandler(fn)`, `fn` returning `(T, []string, error)` | `fn`, on every render, reporting its dependency tags | `T` |
| `collage.Load(fn)`, `fn` returning `(T, error)` | `fn`, on every render, reporting no tags | `T` |
| `collage.Value(v)` | Nothing: `v` is handed over on every render | `v`'s type |
| `collage.Effect(fn)`, `fn` returning `error` | `fn`, on every render, for what it declares | Nothing |

A fragment with no `WithData` renders with no data, as one with `collage.Effect`
does.

A handler is an ordinary function that returns its own type. Write it inline, as
above, or name it and hand it over:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	// ...
}

content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.DataHandler(loadPost)).
	Build()
```

`T` comes from the function's signature, and it is the type the template is
checked against when the page is registered — see
[How templates are checked](#how-templates-are-checked) — so the handler and the
template's idea of its data cannot drift apart. A loader written this way is also
easy to call from a test or a sitemap's handler, which get a `*Post` rather than a
value to assert. The constructors are functions of their own rather than forms of
`WithData` because Go methods cannot take type parameters.

Before v0.49.0 a handler was handed over with `WithDataHandler` and returned
`(any, []string, error)`, and fixed data was `WithData(v)`. Both are gone: wrap the
handler in `collage.DataHandler`, returning its real type, and the value in
`collage.Value`.

The three results of a `DataHandler` each have a job.

- **The data** — whatever the template renders, as `.`. A struct written for the
  template, a "view", is usually clearer than handing a template a database row.
  Each fragment gets its own; a child does not see its parent's. On an error the
  data is dropped, so a nil `*Post` returned beside an error never reaches the
  template as a value that looks present.
- **The tags** — the pieces of content this data was built from, such as
  `"post:hello-world"` or `"author:ada"`. They are collected from every fragment on
  the page, together with the page's own `WithDependency` tags, and stored with the
  cached page, so invalidating a tag drops every page that showed it. Return `nil`
  when the data depends on nothing that changes. See [Caching](/docs/caching).
- **The error** — a non-nil error fails the fragment, and the fragment's
  [failure policy](/docs/fragments-and-slots#when-a-fragment-fails) decides what
  that means for the page.

Tags are kept even when the handler returns an error. A handler that worked out
what it depends on and then failed has still said what would invalidate the page.

A handler also decides how a page that declares no strategy is served. Such a page
is dynamic when anything it renders has a data handler, and static otherwise: a
handler may read the request, a cookie, the clock, and nothing outside the function
can tell whether it does. A handler whose output is the same for every reader — a
post read from a file — belongs on a page that says `Static()` or
`Incremental(ttl)` itself. A handler that reads only the path's parameters and the
locale can say it on its fragment instead, with `Static()`, and then it leaves
every page using that fragment static. A handler whose output is the same for
every reader but changes over time — a measurement — says `Shared()` instead:
its render may be shared between readers, and the page stays dynamic. See
[Caching](/docs/caching#a-page-that-declares-none).

### Fixed data: collage.Value

Data that is fixed when the program starts — a list of links, a heading, a site
name — needs no handler. `collage.Value(v)` hands the template `v` on every render:

```go
type homeView struct {
	Links []link
}

content := collage.NewFragment("home-content", "pages/home.html").
	WithData(collage.Value(homeView{Links: links})).
	Build()
```

Nothing in it fetches per render, so unlike a handler it leaves a page that
declares no strategy static.

A fragment has one source of data. Calling `WithData` twice with data is refused at
registration with `collage.ErrConflictingData`, whose message names the fragment:
`fragment "home-content" has its data set twice`. Keep the call you mean and drop
the other; before v0.49.0 the second call silently won. A nil `Data` —
`WithData(nil)`, or a constructor handed a nil function — is no data, and
conflicts with nothing.

### Handlers with no tags

`collage.Load` is `DataHandler` for a loader that reports no tags, one that
returns the data and an error:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithData(collage.Load(loadClock)).
	Build()
```

Both drop the data when the loader returns an error. Which one to reach for:

| The data | Write |
| --- | --- |
| Is the same on every render | `collage.Value(v)` |
| Is fetched, from content whose changes must reach a cached page | `collage.DataHandler(fn)` |
| Is fetched, with no tags to report | `collage.Load(fn)` |
| Is nothing: the fragment only declares things for the page | `collage.Effect(fn)` — see [below](#handlers-that-render-nothing) |

Leave out the tags only where the page is not cached or the data never changes: a
cached page whose data changes wants its tags, so that they invalidate it. And
moving from `collage.Value` to a handler changes more than the fragment — a page
that declares no strategy goes from static to dynamic with it.

### Not found is not an error

A missing record and a broken database are different failures, and a reader should
get a different answer for each: a 404 for the first, a 500 for the second. Say
which by wrapping `collage.ErrNotFound`:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
	slug := rc.Param("slug")
	post, err := store.Post(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("blog: no post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("blog: load post %q: %w", slug, err)
	}
	return post, []string{"post:" + slug}, nil
}
```

On a `Required()` fragment, an error wrapping `collage.ErrNotFound` renders the
page's not-found page with a 404; any other error renders its error page with a
500. See [Pages and layouts](/docs/pages-and-layouts#not-found-and-error-pages).

Two conditions, both easy to miss:

- **Wrap with `%w` all the way up.** The framework checks with `errors.Is`, so a
  `%v` anywhere between your store and the handler's return turns the 404 into a
  500.
- **Only a required fragment turns it into a 404.** On an optional fragment it is an
  ordinary failure — the fragment renders nothing or its fallback — because the
  page is still worth showing without it.

## When handlers run

A page's data handlers do not run one at a time. **Before a fragment renders its
template, the data handlers of every fragment in its slots start at once.** A page
made of a post, a sidebar and a comment list waits for the slowest of the three,
not their sum.

The order that is guaranteed:

- **A parent's handler finishes before its children's start.** A child can read
  what its parent stored under a key, and a path parameter its parent resolved.
- **Siblings run at the same time**, each on its own goroutine, in no particular
  order.
- **Templates still render one at a time, in tree order.** Output is identical from
  one request to the next, and anything that depends on order — which title wins in
  the `<head>`, say — is decided by the tree, not by which handler finished first.

That has three consequences for the code you write.

- **Handlers must be safe to run concurrently with their siblings.** Anything they
  share — a map, a counter, a client that is not goroutine-safe — needs the same
  care it would anywhere else in Go.
- **Shared values go through a `collage.Key`,** whose `Get` and `Set` hold the
  render's lock — see [below](#sharing-data-between-fragments).
- **A fragment in a slot the template does not render still starts.** Its context
  is cancelled as soon as the template has finished, and its failure is discarded.
  A handler that is expensive and usually skipped belongs behind something other
  than a conditional in a template.

## The render context

`rc *collage.RenderContext` is everything a handler knows about the request being
rendered.

| Member | What it gives you |
| --- | --- |
| `rc.Request` | The `*http.Request` being answered. In a static export, a synthetic `GET` for the page's path |
| `rc.Locale` | The locale the URL resolved to |
| `rc.Param(name)`, `rc.PathParams` | What the route's `{name}` placeholders captured |
| `rc.Page` | The page being rendered; in an action, the page whose URL it answers on (since v0.33.0) — **read only** |
| `key.Get(rc)`, `key.Set(rc, value)` | Values shared between the fragments of one render, through a `collage.Key` |
| `rc.Context()` | The context the render context carries |
| `rc.HoistTitle`, `rc.HoistMeta`, `rc.HoistProperty`, `rc.HoistLink`, `rc.HoistAlternate`, `rc.HoistStylesheet`, `rc.Hoist` | Declarations for the page's `<head>` — see [Head and SEO](/docs/head-and-seo) |
| `rc.Asset(path)` | A mounted file's content-addressed URL — see [Static assets](/docs/assets) |
| `rc.URL(name, params)`, `rc.ActionURL(name, params)` | A page's or an action's URL by name, in the render's locale (since v0.37.0) — see [Links from Go](/docs/links-and-locales#links-from-go) |

Inside a handler, `rc.Context()` is the same context as the `ctx` argument, with
the fragment's timeout on it. Use `ctx`; it is the one already in hand.

Two rules about the render context itself.

- **Do not keep it.** It belongs to one render. Holding it past the handler's
  return — in a goroutine, a cache, a struct — is holding on to a request that has
  finished.
- **Do not write to `rc.Page`.** It is the one registered `*collage.Page`, shared by
  every request rendering that page at the same moment. Writing to its `Paths` map
  or its `DependencyTags` from a handler is a data race on live framework state,
  which `go test -race` reports and production eventually corrupts. Anything that
  varies per request goes in the data you return or under a `collage.Key`.

## Sharing data between fragments

Fragments on one page often need the same thing. The post page's content, its
`<head>` and its "more by this author" box all want the post.

### collage.Key

A `collage.Key` is a name and a type, declared once at package level. `Set` stores
a value under it for the rest of the render, and `Get` reads it back as that type:

```go
var postKey = collage.NewKey[*Post]("post")

// in the parent's handler
postKey.Set(rc, post)

// in a child's handler, which starts after the parent's has returned
post, ok := postKey.Get(rc)
if !ok {
	return nil, nil, errors.New("more-by-author: no post stored under postKey")
}
```

`ok` is false when nothing is stored under the key. A read needs no type assertion,
and a write of the wrong type does not compile. A key is its name *and* its type:
`NewKey[*Post]("post")` and `NewKey[*Draft]("post")` are two keys holding two
values, and two `NewKey[*Post]("post")` declared in different files are one. Keys
are your application's own namespace; pick names that will not collide.

`With` derives a key per value from a declared one: `postKey.With(slug)` is named
`post:<slug>` and holds the same type. `NewKey` panics on an empty name, at the
line that declares it, and a zero `Key` — a struct field never made with `NewKey`
— panics where it is used.

`Get` and `Set` take the render's lock, which is why they are safe from concurrent
siblings. Before v0.50.0 values were shared with `rc.Set(key, value)` and
`collage.Get[T](rc, key)`, over a `SharedData` map; all three are gone.

This works from parent to child, because a parent's handler finishes first. It does
not work between siblings: they run at the same time, so one cannot count on the
other having set anything yet.

### collage.Once

For siblings — or any fragments that might each need the same fetch — use
`collage.Once`. It runs a fetch at most once per render for a key, and hands the
result to every fragment that asks:

```go
func loadAuthorCard(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	slug := rc.Param("slug")
	post, err := collage.Once(rc, postKey.With(slug), func(ctx context.Context) (*Post, error) {
		return store.Post(ctx, slug)
	})
	if err != nil {
		return Author{}, nil, err
	}
	return post.Author, []string{"post:" + slug, "author:" + post.Author.ID}, nil
}
```

The first caller fetches; the rest wait for it and receive what it produced. The
obvious alternative — `Get`, fetch on a miss, `Set` — has a gap between the check
and the write, and two siblings both fall into it and both fetch.

- **An error is a result.** Everyone waiting gets it; the fetch is not retried per
  fragment, which is how one slow failure would become several.
- **It lasts one render.** There is nothing to configure and nothing to evict.
- **A waiter whose own context ends stops waiting** and returns that error.
- **Its values are its own.** What `Once` fetched is not readable with `Get`, and
  what `Set` stored is not handed to `Once`. Two keys of one name and different
  types are two fetches; before v0.50.0 that was `ErrOnceTypeMismatch`.

### collage.Cached

`Once` shares work within a page. `collage.Cached` shares it between pages and
between requests: thirty posts by one author fetch the author once.

```go
var authorKey = collage.NewKey[Author]("author")

author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
	func(ctx context.Context) (Author, error) { return api.Author(ctx, id) })
```

Its tags are added to the page's own, and invalidating one drops both the stored
value and every cached page built from it. Where nothing is kept across renders —
caching off, development, a preview — it behaves exactly like `Once`. The details
are in [Caching](/docs/caching#caching-data-across-pages).

It is also the one to reach for when fragments are refreshed separately. Each
[fragment path](/docs/forms-and-actions#a-fragment-at-its-own-url) is a render of
its own, so `Once` shares nothing between two of them; `Cached` does.

## Handlers that render nothing

Some fragments exist to declare things for the page rather than to render
markup: a title, a canonical link, structured data. `collage.Effect` adapts a
handler that returns only an error:

```go
seo := collage.NewFragment("post-seo", "fragments/empty.html").
	WithData(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		post, err := collage.Once(rc, postKey.With(rc.Param("slug")), func(ctx context.Context) (*Post, error) {
			return store.Post(ctx, rc.Param("slug"))
		})
		if err != nil {
			return err
		}
		rc.HoistTitle(post.Title)
		rc.HoistMeta("description", post.Summary)
		return nil
	})).
	Build()
```

The fragment's template receives no data, and the handler reports no tags. When
what it declares comes from content that changes and the page is cached, that
content's tags still have to reach the page: return them from an ordinary handler
instead, or add them with `WithDependency` on the page.

A title known when the program starts — the site's name, on its layout — needs no
handler at all. `WithTitle(s)` on the fragment declares it, and leaves a page that
declares no strategy static; see [Head and SEO](/docs/head-and-seo).

## Timeouts and the context

Every data handler runs under a deadline: the fragment's `WithTimeout(d)`, or
`Config.Template.Timeout` — five seconds by default — for a fragment that sets
none.

The deadline is on the `ctx` the handler receives. **It bounds the context, not
the handler.** The framework does not abandon a handler at its deadline; it waits
for it to return, because a goroutine cannot be stopped from outside, and
abandoning handlers that never return would leak one goroutine per request,
forever. So a handler that ignores its context can run past its deadline, and the
page waits for it.

Honouring it is one habit: pass `ctx` to everything that can wait.

```go
func loadWeather(ctx context.Context, rc *collage.RenderContext) (*Weather, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherURL(rc.Locale), nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err // context.DeadlineExceeded when the timeout passed
	}
	defer res.Body.Close()

	var weather Weather
	if err := json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return nil, nil, err
	}
	return &weather, nil, nil
}
```

A handler that returns its context's error when time runs out fails like any
other, with an error that still matches `errors.Is(err, context.DeadlineExceeded)`.
The deadline is only an error if the handler says so: one that ignores `ctx` and
returns `nil` late has succeeded, and its data renders. Pair a short timeout with a
fallback on anything that is nice to have, and the page stops waiting on a slow
service at the point you chose.

The error says which deadline it was. Only the fragment's own timeout is reported
as one — `collage: execution exceeded 5s` — and a context that ended above it, a
request cancelled or a shorter deadline, is reported as `collage: execution
stopped by context`, with its cause when it has one. Before v0.18.1 both read as
a timeout, and a cancelled request sent you looking for a handler that was never
slow.

## Panics

A panic in a data handler does not take the process down. It is recovered into a
`*collage.PanicError`, carrying the panic value and the stack, and the fragment
fails with it like any other error. In development the error page shows the stack;
reach it in your own code with `errors.As`.

## How templates are checked

A template reads its data by name — `{{.Title}}`, `{{.Author.Name}}` — and
`html/template` resolves those names only when the template runs. A `{{.Titel}}`
fails every render of its page, and on a page nobody visits, it never shows. Since
v0.49.0 every fragment's data comes from a typed constructor, so registration
knows the Go type each template will run with and walks the template against it.
`RegisterPage` refuses a page whose templates do not fit their data, naming the
page, the fragment, the file, the line and the column:

```text
collage: page "post": fragment "post-body" (post.html:1:6): {{.Titel}}: type blog.Post has no field or method Titel (did you mean Title?)
```

It reports only what is certain to fail: each finding is an expression
`text/template` would fail on when the render reached it. An application that now
fails at startup had a template that would already have failed when rendered; the
check reports it earlier, and adds no rule of its own.

### What is walked

Every fragment a page reaches: its layouts, its content, everything bound into
their slots, fallbacks, inline fragments, fragments opened with
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url), and the
pages named as its [not-found and error pages](/docs/pages-and-layouts#not-found-and-error-pages).
A partial included with `{{template "partials/author.html" .Author}}` is walked
with the type it is handed, and once for each type it is handed across the
application.

A fragment a [slot resolver](/docs/fragments-and-slots#slots-filled-per-render)
returns is built while the page renders, so registration never sees it, and its
template is not checked.

`.` starts as the fragment's data type and follows the template: inside
`{{range}}` it is the element, inside `{{with}}` the value, and a `$x := …`
variable keeps the type it was given for its scope.

### What is reported

- **A field or method the type does not have**, or a field that is unexported.
  Pointers are followed, and fields promoted from embedded structs count. The
  closest exported name is suggested when one is near.
- **A method with a pointer receiver, reached through a value that is not
  addressable.** `text/template` calls such a method through the value's
  address, and data handed to a template by value has none — nor have its
  fields, nor a map's elements. What a pointer points to and a slice's elements
  are addressable. If `URL` is declared on `*Post`, `{{.URL}}` fails on a `Post`
  and works on a `*Post`; the fix is to hand the template the pointer.
- **A map keyed by a named string type** (`map[Slug]Post`): `.key` looks a key up
  as a plain `string`, which such a map does not accept. A map keyed by `string`
  is fine, and a key it does not hold is not an error either: it silently ends
  the chain, so `{{.Meta.absent.Name}}` renders nothing. The check, knowing only
  the map's element type, still reports a name that element type lacks.
- **Calls of the wrong shape**: a method, a template function or a built-in
  given the wrong number of arguments; a field or map key given arguments; a
  method or function returning more than a value and an error; `call` on
  something that is not a function.
- **`range`, `len` and `index` on the wrong kind**: ranging over a string or a
  struct, or over an integer with two variables; `len` of a struct; `index` into
  a struct.

What the render can never reach is not reported. An `{{if}}` or `{{with}}` whose
condition is a literal — `true`, `false`, `0`, `1`, `""`, `"x"`, or `not` of one —
never runs the side the literal rules out. `{{and}}` and `{{or}}` stop at an
operand only when it is a literal that decides them: `false`, `0`, `""` or `nil`
for `and`; `true`, a non-zero number or a non-empty string for `or`. So
`{{and 0 .Nope}}` reports nothing, while `{{and 1 .Nope}}` and
`{{and .Title .Nope}}` report `.Nope`, since the render may reach it.

### What is not

- **What cannot be known before the render.** An interface — `any`, a
  `map[string]any`'s entries, an `error` — could hold anything, so nothing read
  from one is reported. Nor is anything read from a `reflect.Value` that a method
  or function returns, which `text/template` unwraps into whatever it holds.
- **A fragment with no data.** Reading a field of nil data is not an error in
  `html/template`: `{{.Title}}` renders nothing. A fragment without `WithData`, or
  with `collage.Effect`, is still walked, but only its function calls and their
  argument counts are checked.
- **Nil pointers inside the data.** `{{.Author.Name}}` with a nil `Author` fails
  when it renders, and whether it is nil is the data's business, not the type's.
- **Argument types.** Only the number of arguments is checked: `text/template`
  converts some arguments itself, and second-guessing it is where false alarms
  would come from.

### Reading the error

`RegisterPage` returns every finding in the page's own fragments at once, joined
with `errors.Join` and ordered by file, line and column, so ten mistakes are one
restart rather than ten. The page's not-found page is checked once the page itself
passes, and its error page once both have; their findings arrive wrapped in a
`collage: page "post" not-found page: …` (or `error page: …`) error, on the start
after the page's own are fixed.

Each finding is a `*collage.TemplateTypeError`, and each matches
`collage.ErrTemplateType`. A finding may sit under a wrapping error as well as a
join, so walk the whole tree to list them:

```go
// typeErrors collects every *collage.TemplateTypeError in err's tree.
func typeErrors(err error) []*collage.TemplateTypeError {
	switch e := err.(type) {
	case nil:
		return nil
	case *collage.TemplateTypeError:
		return []*collage.TemplateTypeError{e}
	case interface{ Unwrap() []error }:
		var found []*collage.TemplateTypeError
		for _, inner := range e.Unwrap() {
			found = append(found, typeErrors(inner)...)
		}
		return found
	case interface{ Unwrap() error }:
		return typeErrors(e.Unwrap())
	}
	return nil
}
```

and, where the page is registered:

```go
if err := app.RegisterPage(page); errors.Is(err, collage.ErrTemplateType) {
	for _, typeErr := range typeErrors(err) {
		fmt.Printf("%s:%d:%d %s\n", typeErr.Template, typeErr.Line, typeErr.Col, typeErr.Reason)
	}
}
```

`errors.As` on the returned error finds the first finding. Its fields:

| Field | What it holds |
| --- | --- |
| `Page`, `Fragment` | The page and the fragment the template belongs to |
| `Template` | The file the expression is in — the fragment's own, or a partial it includes — or `inline template of fragment "x"` |
| `Line`, `Col` | Where `text/template` itself would name the error, which may lie inside `Expr` rather than at its start: at the last argument of a call, say |
| `Expr` | The expression as written, `{{.Titel}}` |
| `Reason` | What is wrong |
| `Suggestion` | The closest name the type has, or empty |

The check runs where the other registration checks run: at startup, and so in
[`collage check`](/docs/cli#collage-check), which starts the application, and in
a test that registers the pages. Under `collage dev` a template edited while the
program runs is reparsed but not checked again until the next restart, which is
the next change to its Go code. [`collage inspect`](/docs/cli#collage-inspect)
prints each fragment's data type, and the types it reaches, for an editor to check
the same way.

### Turning it off

`WithoutTypeCheck()` leaves one fragment's template out of the check, for the rare
finding that is wrong, while it is fixed. The fragment renders as before, and
`collage inspect` marks it `"typeCheck": false`:

```go
collage.NewFragment("post-body", "pages/post.html").
	WithData(collage.DataHandler(loadPost)).
	WithoutTypeCheck().
	Build()
```

A handler declared to return an interface — `collage.Load[any]`, say — leaves the
type unknown, and its template is not walked: that is the way to say on purpose
that a fragment's data has no fixed shape. `collage.Value` is the exception. Its
value is in hand, so a value held in an interface is checked against the type it
holds.
