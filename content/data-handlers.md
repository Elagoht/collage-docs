---
description: How a fragment fetches its data — the typed handler contract, dependency tags, 404s, concurrency, the render context, sharing data, and timeouts.
reference: DataHandler, Data, Load, RenderContext, Get, Once, Effect, ErrNotFound, PanicError
---

# Data handlers

A data handler is the function a fragment calls to get what its template renders.
It receives the request's context and the render context, and returns the data,
the dependency tags that data came from, and an error. Everything about a page
that talks to the outside world — a database, a CMS, an API — happens in data
handlers, and nowhere else.

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (Post, []string, error) {
	post, err := store.Post(ctx, rc.Param("slug"))
	if err != nil {
		return Post{}, nil, err
	}
	return post, []string{"post:" + post.Slug}, nil
}

content := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	Required().
	Build()
```

## The contract

Write a handler against your own type, and adapt it with `collage.DataHandler`:

```go
func(ctx context.Context, rc *collage.RenderContext) (T, []string, error)
```

`collage.DataHandler` is generic over `T`, so one adapter serves every view type
and the value your handler returns is exactly what the template receives as `.`.
Nothing in your code needs an untyped value or a type assertion. (It is a function
rather than a method on the builder because Go methods cannot take type
parameters.)

The three results each have a job.

- **The data** — whatever the template renders. A struct written for the template,
  a "view", is usually clearer than handing a template a database row. Each
  fragment gets its own; a child does not see its parent's.
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

On an error, `collage.DataHandler` drops the data rather than passing it on. That
matters for pointer types: a nil `*Post` returned alongside an error would
otherwise reach the template as a non-nil value holding a nil pointer.

### Shorter adapters: Data and Load

Not every fragment needs all three results. Two adapters cover the ones that
report no tags.

`collage.Data` takes the value itself, for data that is fixed when the program
starts — a list of links, a heading, a site name. There is no function to write:

```go
type homeView struct {
	Links []link
}

content := collage.NewFragment("home-content", "pages/home.html").
	WithDataHandler(collage.Data(homeView{Links: links})).
	Build()
```

`collage.Load` takes a handler that returns the data and an error, without the
tags:

```go
func loadClock(_ context.Context, rc *collage.RenderContext) (clockView, error) {
	return clockView{Now: time.Now(), Locale: rc.Locale}, nil
}

content := collage.NewFragment("clock", "fragments/clock.html").
	WithDataHandler(collage.Load(loadClock)).
	Build()
```

Both are generic like `collage.DataHandler`, so the template still receives your
own type, and `collage.Load` drops the data on an error the same way. Which one to
reach for:

| The data | Adapter |
| --- | --- |
| Is the same on every render | `collage.Data(v)` |
| Is fetched, and the page is not cached or the data never changes | `collage.Load(fn)` |
| Is fetched, and a cached page must be dropped when it changes | `collage.DataHandler(fn)` |

Moving from one to the next is a change of signature, not a rewrite: when a page
starts being cached, a `Load` handler gains its tags and becomes a `DataHandler`
one.

### Not found is not an error

A missing record and a broken database are different failures, and a reader should
get a different answer for each: a 404 for the first, a 500 for the second. Say
which by wrapping `collage.ErrNotFound`:

```go
func loadPost(ctx context.Context, rc *collage.RenderContext) (Post, []string, error) {
	slug := rc.Param("slug")
	post, err := store.Post(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, nil, fmt.Errorf("blog: no post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return Post{}, nil, fmt.Errorf("blog: load post %q: %w", slug, err)
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
  what its parent put in shared data, and a path parameter its parent resolved.
- **Siblings run at the same time**, each on its own goroutine, in no particular
  order.
- **Templates still render one at a time, in tree order.** Output is identical from
  one request to the next, and anything that depends on order — which title wins in
  the `<head>`, say — is decided by the tree, not by which handler finished first.

That has three consequences for the code you write.

- **Handlers must be safe to run concurrently with their siblings.** Anything they
  share — a map, a counter, a client that is not goroutine-safe — needs the same
  care it would anywhere else in Go.
- **Shared data goes through `rc.Get` and `rc.Set`,** never the `SharedData` map
  directly — see [below](#sharing-data-between-fragments).
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
| `rc.Page` | The page being rendered — **read only** |
| `rc.Get(key)`, `rc.Set(key, value)` | Values shared between the fragments of one render |
| `rc.Context()` | The context the render context carries |
| `rc.HoistTitle`, `rc.HoistMeta`, `rc.HoistProperty`, `rc.HoistLink`, `rc.HoistAlternate`, `rc.HoistStylesheet`, `rc.Hoist` | Declarations for the page's `<head>` — see [Head and SEO](/docs/head-and-seo) |
| `rc.Asset(path)` | A mounted file's content-addressed URL — see [Static assets](/docs/assets) |

Inside a handler, `rc.Context()` is the same context as the `ctx` argument, with
the fragment's timeout on it. Use `ctx`; it is the one already in hand.

Two rules about the render context itself.

- **Do not keep it.** It belongs to one render. Holding it past the handler's
  return — in a goroutine, a cache, a struct — is holding on to a request that has
  finished.
- **Do not write to `rc.Page`.** It is the one registered `*collage.Page`, shared by
  every request rendering that page at the same moment. Writing to its `SEO` map or
  its `DependencyTags` from a handler is a data race on live framework state, which
  `go test -race` reports and production eventually corrupts. Anything that varies
  per request goes in the data you return or in shared data.

## Sharing data between fragments

Fragments on one page often need the same thing. The post page's content, its
`<head>` and its "more by this author" box all want the post.

### rc.Set and collage.Get

`rc.Set(key, value)` stores a value for the rest of the render, and
`collage.Get[T](rc, key)` reads it back as the type it was stored as:

```go
// in the parent's handler
rc.Set("post", post)

// in a child's handler, which starts after the parent's has returned
post, ok := collage.Get[Post](rc, "post")
if !ok {
	return moreView{}, nil, errors.New("more-by-author: no post in shared data")
}
```

`ok` is false when nothing is stored under the key, and also when what is stored is
not a `Post` — to the caller, both mean the value it wanted is not there. Keys are
your application's own namespace; pick names that will not collide.

`rc.Get` and `rc.Set` take the render's lock, which is why they are safe from
concurrent siblings and the bare `rc.SharedData` map is not.

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
	post, err := collage.Once(rc, "post:"+slug, func(ctx context.Context) (Post, error) {
		return store.Post(ctx, slug)
	})
	if err != nil {
		return Author{}, nil, err
	}
	return post.Author, []string{"post:" + slug, "author:" + post.Author.ID}, nil
}
```

The first caller fetches; the rest wait for it and receive what it produced. The
obvious alternative — `rc.Get`, fetch on a miss, `rc.Set` — has a gap between the
check and the write, and two siblings both fall into it and both fetch.

- **An error is a result.** Everyone waiting gets it; the fetch is not retried per
  fragment, which is how one slow failure would become several.
- **It lasts one render.** There is nothing to configure and nothing to evict.
- **A waiter whose own context ends stops waiting** and returns that error.
- **One key, one type.** Asking for a key as two different types is
  `ErrOnceTypeMismatch`, not a silently empty value.

### collage.Cached

`Once` shares work within a page. `collage.Cached` shares it between pages and
between requests: thirty posts by one author fetch the author once.

```go
author, err := collage.Cached(rc, "author:"+id, time.Hour, []string{"author:" + id},
	func(ctx context.Context) (Author, error) { return api.Author(ctx, id) })
```

Its tags are added to the page's own, and invalidating one drops both the stored
value and every cached page built from it. Where nothing is kept across renders —
caching off, development, a preview — it behaves exactly like `Once`. The details
are in [Caching](/docs/caching#caching-data-across-pages).

## Handlers that render nothing

Some fragments exist to declare things for the page rather than to render
markup: a title, a canonical link, structured data. `collage.Effect` adapts a
handler that returns only an error:

```go
seo := collage.NewFragment("post-seo", "fragments/empty.html").
	WithDataHandler(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		post, err := collage.Once(rc, "post:"+rc.Param("slug"), func(ctx context.Context) (Post, error) {
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
content's tags still have to reach the page: return them from a
`collage.DataHandler` instead, or add them with `WithDependency` on the page.

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
func loadWeather(ctx context.Context, rc *collage.RenderContext) (Weather, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherURL(rc.Locale), nil)
	if err != nil {
		return Weather{}, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Weather{}, nil, err // context.DeadlineExceeded when the timeout passed
	}
	defer res.Body.Close()

	var weather Weather
	if err := json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return Weather{}, nil, err
	}
	return weather, nil, nil
}
```

A handler that returns its context's error when time runs out fails like any
other, with an error that still matches `errors.Is(err, context.DeadlineExceeded)`.
The deadline is only an error if the handler says so: one that ignores `ctx` and
returns `nil` late has succeeded, and its data renders. Pair a short timeout with a
fallback on anything that is nice to have, and the page stops waiting on a slow
service at the point you chose.

## Panics

A panic in a data handler does not take the process down. It is recovered into a
`*collage.PanicError`, carrying the panic value and the stack, and the fragment
fails with it like any other error. In development the error page shows the stack;
reach it in your own code with `errors.As`.
