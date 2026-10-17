---
description: How a fragment fetches its data: the handler contract, dependency tags, 404s, concurrency, the render context, timeouts, and checking templates against data.
reference: Data, FragmentBuilder.WithData, FragmentBuilder.WithTitle, FragmentBuilder.WithoutTypeCheck, LoadTagged, Load, Value, RenderContext, Key, NewKey, Once, Effect, ErrNotFound, ErrConflictingData, TemplateTypeError, ErrTemplateType, PanicError
---

# Data handlers

A data handler is the function a fragment calls to get what its template renders.
It receives the request's context and the render context, and returns the data in
a Go type of its own, the dependency tags that data came from, and an error.

Everything about a page that talks to the outside world — a database, a CMS, an
API — happens in data handlers, and nowhere else.

```go
content := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(func(ctx context.Context, rc *collage.RenderContext) (*Post, []string, error) {
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
| `collage.LoadTagged(fn)`, `fn` returning `(T, []string, error)` | `fn`, on every render, reporting its dependency tags | `T` |
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
	WithData(collage.LoadTagged(loadPost)).
	Build()
```

`T` comes from the function's signature, and it is the type the template is
checked against when the page is registered — see
[How templates are checked](#how-templates-are-checked) — so the handler and the
template's idea of its data cannot drift apart.

A loader written this way is also easy to call from a test or a sitemap's handler,
which get a `*Post` rather than a value to assert. The constructors are functions
of their own rather than forms of `WithData` because Go methods cannot take type
parameters.

Before v0.49.0 a handler was handed over with `WithDataHandler` and returned
`(any, []string, error)`, and fixed data was `WithData(v)`. Both are gone: wrap the
handler in `collage.LoadTagged`, returning its real type, and the value in
`collage.Value`. (`LoadTagged` was named `DataHandler` until v0.60.0.)

The three results of a `LoadTagged` each have a job.

- **The data** — whatever the template renders, as `.`. A struct written for the
  template, a "view", is usually clearer than handing a template a database row.
  Each fragment gets its own; a child does not see its parent's. On an error the
  data is dropped, so a nil `*Post` returned beside an error never reaches the
  template as a value that looks present.
- **The tags** — the pieces of content this data was built from, such as
  `"post:hello-world"` or `"author:ada"`. They are collected from every fragment on
  the page, together with the page's own `WithDependencyTags` tags, and stored with the
  cached page, so invalidating a tag invalidates every page that showed it. Return `nil`
  when the data depends on nothing that changes. See [Caching](/docs/caching).
- **The error** — a non-nil error fails the fragment, and the fragment's
  [failure policy](/docs/fragments-and-slots#when-a-fragment-fails) decides what
  that means for the page.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 244" role="img" aria-labelledby="dh-contract-title">
<title id="dh-contract-title">A data handler receives ctx and rc and returns three things. The data becomes the template's dot, and is dropped when there is an error. The tags join the page's tags, are kept when there is an error, and are stored with a cached page. The error goes to the fragment's failure policy: a required fragment fails the page with a 500, or a 404 when the error wraps ErrNotFound; a fragment with a fallback renders the fallback; any other renders nothing.</title>
<defs><marker id="dh-contract-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">ctx, rc</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#dh-contract-arrow)" />
<rect class="d-a" x="90" y="50" width="140" height="34" rx="5" /><text class="d-on-a" x="160" y="65" text-anchor="middle">data handler</text><text class="d-on-a d-small" x="160" y="78" text-anchor="middle">LoadTagged(fn)</text>
<path class="d-flow" d="M130 84C130 96 56 92 56 104" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="160" y1="84" x2="160" y2="104" marker-end="url(#dh-contract-arrow)" />
<path class="d-flow" d="M190 84C190 96 264 92 264 104" marker-end="url(#dh-contract-arrow)" />
<rect class="d-frame" x="8" y="108" width="96" height="26" rx="5" /><text x="56" y="125" text-anchor="middle">data: T</text>
<rect class="d-frame" x="112" y="108" width="96" height="26" rx="5" /><text x="160" y="125" text-anchor="middle">tags</text>
<rect class="d-frame" x="216" y="108" width="96" height="26" rx="5" /><text x="264" y="125" text-anchor="middle">error</text>
<line class="d-flow" x1="56" y1="134" x2="56" y2="150" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="160" y1="134" x2="160" y2="150" marker-end="url(#dh-contract-arrow)" />
<line class="d-flow" x1="264" y1="134" x2="264" y2="150" marker-end="url(#dh-contract-arrow)" />
<rect class="d-a" x="8" y="154" width="96" height="26" rx="5" /><text class="d-on-a" x="56" y="171" text-anchor="middle">template's .</text>
<rect class="d-b" x="112" y="154" width="96" height="26" rx="5" /><text class="d-on-b" x="160" y="171" text-anchor="middle">page's tags</text>
<rect class="d-c" x="216" y="154" width="96" height="26" rx="5" /><text class="d-on-c" x="264" y="171" text-anchor="middle">failure policy</text>
<text class="d-small" x="56" y="196" text-anchor="middle">dropped on</text><text class="d-small" x="56" y="209" text-anchor="middle">an error</text>
<text class="d-small" x="160" y="196" text-anchor="middle">kept on an error</text><text class="d-small" x="160" y="209" text-anchor="middle">stored with a</text><text class="d-small" x="160" y="222" text-anchor="middle">cached page</text>
<text class="d-small" x="264" y="196" text-anchor="middle">Required(): 500,</text><text class="d-small" x="264" y="209" text-anchor="middle">ErrNotFound: 404</text><text class="d-small" x="264" y="222" text-anchor="middle">WithFallback: f</text><text class="d-small" x="264" y="235" text-anchor="middle">neither: nothing</text>
</svg>
<figcaption>A data handler is handed the request's context and the render context, and each of its three results goes somewhere else. The data becomes the template's <code>.</code>; on an error it is dropped. The tags join the page's own and are stored with the page when it is cached; on an error they are kept. The error goes to the fragment's failure policy: a <code>Required()</code> fragment fails the page with a 500, or a 404 when the error wraps <code>collage.ErrNotFound</code>; a fragment with a fallback renders it; any other renders nothing.</figcaption>
</figure>

Tags are kept even when the handler returns an error. A handler that worked out
what it depends on and then failed has still said what would invalidate the page.

A handler also decides how a page that declares no strategy is served. Such a page
is dynamic when anything it renders has a data handler, and static otherwise: a
handler may read the request, a cookie, the clock, and nothing outside the function
can tell whether it does.

What to say instead depends on the handler's output:

| The handler's output | Say | What it does |
| --- | --- | --- |
| Is the same for every reader — a post read from a file | `Static()` or `Incremental(ttl)` on the page itself | The page keeps the strategy it declares |
| Depends only on the path's parameters and the locale | `Static()` on the handler's fragment | Leaves every page using that fragment static |
| Is the same for every reader but changes over time — a measurement | `Shared()` on the fragment | Its render may be shared between readers, and the page stays dynamic |

See [Caching](/docs/caching#a-page-that-declares-none).

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
the other; before v0.49.0 the second call silently won.

A nil `Data` —
`WithData(nil)`, or a constructor handed a nil function — is no data, and
conflicts with nothing.

### Handlers with no tags

`collage.Load` is `LoadTagged` for a loader that reports no tags, one that
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
| Is fetched, from content whose changes must reach a cached page | `collage.LoadTagged(fn)` |
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
and a write of the wrong type does not compile.

A key is its name *and* its type:
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

Its tags are added to the page's own, and invalidating one marks both the stored
value and every cached page built from it stale. Where nothing is kept across renders —
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
instead, or add them with `WithDependencyTags` on the page.

A title known when the program starts — the site's name, on its layout — needs no
handler at all. `WithTitle(s)` on the fragment declares it, and leaves a page that
declares no strategy static; see [Head and SEO](/docs/head-and-seo).

## Example: a catalogue product page

A shop's product page at `/products/{id}` draws on three services, and each piece
of data is fetched in the way that suits it:

| Data | Comes from | Fetched with |
| --- | --- | --- |
| Name and brand id | The catalogue | The product's handler, on every render |
| Price | A pricing service | The same handler, after the catalogue |
| Brand: name and country | The catalogue, a slow lookup many products share | `collage.Cached`, for a day, tagged `brand:<id>` |
| Reviews | A reviews service | `collage.Once`, from two sibling fragments: the reviews box and the page's description |

The catalog, pricing and reviews packages stand for your own clients. The
product's handler:

```go
var (
	brandKey   = collage.NewKey[catalog.Brand]("brand")
	reviewsKey = collage.NewKey[reviews.Summary]("reviews")
)

type productView struct {
	Name  string
	Price string
	Brand catalog.Brand
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (productView, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Product(ctx, id)
	if errors.Is(err, catalog.ErrNoProduct) {
		return productView{}, nil, fmt.Errorf("shop: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: load product %q: %w", id, err)
	}
	price, err := pricing.Price(ctx, id)
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: price of %q: %w", id, err)
	}
	// Many products share a brand: one fetch serves all of them.
	brand, err := collage.Cached(rc, brandKey.With(p.BrandID), 24*time.Hour, []string{"brand:" + p.BrandID},
		func(ctx context.Context) (catalog.Brand, error) { return catalog.LookupBrand(ctx, p.BrandID) })
	if err != nil {
		return productView{}, nil, fmt.Errorf("shop: brand %q: %w", p.BrandID, err)
	}
	return productView{Name: p.Name, Price: price, Brand: brand}, []string{"product:" + id, "price:" + id}, nil
}
```

- **An unknown id is a 404.** The handler wraps `collage.ErrNotFound`, and its
  fragment is `Required()`, so the reader gets the not-found page. Any other
  error fails the page with a 500.
- **The brand is fetched once for all its products.** The `brand:<id>` tag joins
  the page's tags as if the handler had returned it, so the handler returns only
  `product:<id>` and `price:<id>`.

The reviews are wanted twice: by the reviews box, and by a fragment that puts the
rating in the page's description. Both call one function:

```go
// reviewsOf is called by two sibling fragments; Once makes it one call per render.
func reviewsOf(rc *collage.RenderContext, id string) (reviews.Summary, error) {
	return collage.Once(rc, reviewsKey.With(id), func(ctx context.Context) (reviews.Summary, error) {
		return reviews.For(ctx, id)
	})
}

func loadReviews(ctx context.Context, rc *collage.RenderContext) (reviews.Summary, []string, error) {
	id := rc.Param("id")
	summary, err := reviewsOf(rc, id)
	if err != nil {
		return reviews.Summary{}, nil, err
	}
	return summary, []string{"reviews:" + id}, nil
}

func productSEO(ctx context.Context, rc *collage.RenderContext) error {
	summary, err := reviewsOf(rc, rc.Param("id"))
	if err != nil {
		return err
	}
	rc.HoistMeta("description", fmt.Sprintf("Rated %.1f out of 5 in %d reviews.", summary.Average, summary.Count))
	return nil
}
```

The two fragments are siblings, so their handlers start together. Whichever
reaches `reviewsOf` first calls the reviews service, and the other waits for its
result. A `Get`, a fetch on a miss and a `Set` could have called it twice.

The product binds both into its slots:

```go
reviewsBox := collage.NewFragment("product-reviews", "fragments/reviews.html").
	WithData(collage.LoadTagged(loadReviews)).
	WithTimeout(500 * time.Millisecond).
	Build()

seo := collage.NewFragment("product-seo", "fragments/empty.html").
	WithData(collage.Effect(productSEO)).
	WithTimeout(500 * time.Millisecond).
	Build()

content := collage.NewFragment("product", "pages/product.html").
	WithData(collage.LoadTagged(loadProduct)).
	WithSlotFragment("reviews", reviewsBox).
	WithSlotFragment("seo", seo).
	Required().
	Build()

page := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(content).
	WithPath("en", "/products/{id}").
	Incremental(10 * time.Minute).
	Build()
```

```html
<!-- templates/pages/product.html -->
<article>
  <h1>{{.Name}}</h1>
  <p class="price">{{.Price}}</p>
  <p class="brand">{{.Brand.Name}}, {{.Brand.Country}}</p>
  {{slot "reviews"}}
</article>
{{slot "seo"}}
```

The description lands where the layout's `{{hoist "head"}}` marker is — see
[Head and SEO](/docs/head-and-seo). Both fragments get the same timeout because
the one call runs under the context of whichever fragment asked first.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 200" role="img" aria-labelledby="dh-share-title">
<title id="dh-share-title">Once and Cached side by side. Left, Once: in one render, the sibling fragments reviews and seo both ask for the key reviews:42, and one call serves both; the next render calls again. Right, Cached: the renders of product 42 and product 43 both ask for the key brand:acme, and one call is kept in the data cache, fresh for 24 hours or until brand:acme is invalidated.</title>
<defs><marker id="dh-share-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<line class="d-rule" x1="160" y1="8" x2="160" y2="192" />
<text x="80" y="20" text-anchor="middle">Once</text>
<rect class="d-frame" x="6" y="28" width="146" height="128" rx="6" />
<text class="d-small" x="14" y="42">one render</text>
<rect class="d-b" x="16" y="50" width="60" height="24" rx="4" /><text class="d-on-b d-small" x="46" y="66" text-anchor="middle">reviews</text>
<rect class="d-b" x="84" y="50" width="60" height="24" rx="4" /><text class="d-on-b d-small" x="114" y="66" text-anchor="middle">seo</text>
<path class="d-flow" d="M46 74C46 86 76 82 76 94" marker-end="url(#dh-share-arrow)" />
<path class="d-flow" d="M114 74C114 86 84 82 84 94" marker-end="url(#dh-share-arrow)" />
<rect class="d-chip" x="22" y="98" width="116" height="22" rx="11" /><text class="d-small" x="80" y="113" text-anchor="middle">reviews:42</text>
<text class="d-small" x="80" y="142" text-anchor="middle">one call, shared</text>
<text class="d-small" x="80" y="176" text-anchor="middle">next render:</text>
<text class="d-small" x="80" y="189" text-anchor="middle">called again</text>
<text x="242" y="20" text-anchor="middle">Cached</text>
<rect class="d-frame" x="168" y="28" width="70" height="46" rx="6" /><text class="d-small" x="203" y="48" text-anchor="middle">render of</text><text class="d-small" x="203" y="61" text-anchor="middle">product 42</text>
<rect class="d-frame" x="246" y="28" width="70" height="46" rx="6" /><text class="d-small" x="281" y="48" text-anchor="middle">render of</text><text class="d-small" x="281" y="61" text-anchor="middle">product 43</text>
<path class="d-flow" d="M203 74C203 86 238 82 238 94" marker-end="url(#dh-share-arrow)" />
<path class="d-flow" d="M281 74C281 86 246 82 246 94" marker-end="url(#dh-share-arrow)" />
<rect class="d-chip" x="184" y="98" width="116" height="22" rx="11" /><text class="d-small" x="242" y="113" text-anchor="middle">brand:acme</text>
<rect class="d-soft-a" x="168" y="128" width="148" height="28" rx="6" /><text class="d-small" x="242" y="146" text-anchor="middle">data cache: one call</text>
<text class="d-small" x="242" y="176" text-anchor="middle">fresh 24 h, or until</text>
<text class="d-small" x="242" y="189" text-anchor="middle">brand:acme invalidated</text>
</svg>
<figcaption><code>Once</code> shares a call between the fragments of one render, <code>Cached</code> between renders. On the product page the reviews box and the description share one call to the reviews service per render, and every product of one brand shares one brand lookup. Where nothing is kept across renders (caching off, development, a preview) <code>Cached</code> behaves like <code>Once</code>.</figcaption>
</figure>

What each failure does to this page:

- **There is no product 7.** The required fragment's error wraps
  `collage.ErrNotFound`: the not-found page, with a 404.
- **The catalogue or the pricing service fails.** The product is required, so the
  page's error page is served with a 500.
- **The brand lookup fails.** With no brand stored yet, that is a 500 too. With
  one stored, `Cached` serves it stale instead, and the page, built from stale
  data, is not cached.
- **The reviews service is down, or slower than 500 milliseconds.** Its one call
  fails, and both fragments get that error: the reviews box renders nothing, no
  description is declared, and the page answers 200 without them. A page served
  with a failed fragment is degraded and never cached, so the next request asks
  the reviews service again.

With the cache on, none of this runs while the page is fresh: a cached page is
not rendered, so the next reader of `/products/42` within ten minutes costs no call
at all. See [Caching](/docs/caching#once-cached-and-the-page-cache).

## Timeouts and the context

Every data handler runs under a deadline: the fragment's `WithTimeout(d)`, or
`Config.Template.Timeout` — five seconds by default — for a fragment that sets
none.

The deadline is on the `ctx` the handler receives. **It bounds the context, not
the handler.**

The framework does not abandon a handler at its deadline; it waits for it to
return, because a goroutine cannot be stopped from outside, and abandoning
handlers that never return would leak one goroutine per request, forever. So a
handler that ignores its context can run past its deadline, and the page waits
for it.

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
returns `nil` late has succeeded, and its data renders.

Pair a short timeout with a
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
fails every render of its page, and on a page nobody visits, it never shows.

Since v0.49.0 every fragment's data comes from a typed constructor, so registration
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
- **`range`, `len`, `index` and `slice` on the wrong kind**: ranging over a
  string or a struct, or over an integer with two variables; `len` of a struct;
  `index` into a struct; `slice` of anything but a string, a slice or an array,
  of a string by three indexes, or of an array `text/template` cannot take the
  address of (`{{slice .Days 0}}` on data passed by value: pass the pointer). A
  slice, array or string is indexed and sliced by integers, and a map by a key
  of its own key type: `{{index .ByID "7"}}` on a `map[int]Post` is reported.
- **Number constants**, typed as `text/template` types them: `1` is an `int`,
  `1.5` a `float64`, `1i` a `complex128`, `'a'` an `int`. So `{{(1.5).Nope}}` and
  `{{len 1}}` are reported, as are a constant too large for an `int` and `nil`
  used as a command.
- **An argument its parameter cannot take** (from v0.51.0), judged as
  `text/template` judges it when it calls a template function or a method:
  - a constant by the parameter's kind: an integer parameter takes a number
    written as an integer, an unsigned one a non-negative integer, a float one
    any number but a complex one, a string or bool one a string or a bool. So
    `{{upper 1}}`, `{{printf 1}}`, `{{uint -1}}` and `{{i8 1.5}}` are reported,
    and `nil` goes only where a nil can;
  - any other value must be assignable to the parameter, or be a pointer to
    something that is, or be addressable where the parameter is its pointer. So
    `{{upper .Count}}` with an `int` field and `{{edit .Owner}}` for a
    `func(*User)` on data passed by value are reported, while `{{show .Author}}`
    for a `func(User)` with a `*User` field is fine. A variadic parameter takes
    its element type, and the value piped in, `{{.Count | upper}}`, is judged
    like any other argument;
  - `call` checks what it passes as `text/template`'s `call` does: a value
    assignable to the parameter, or an integer converted to another integer
    type, and nothing else.

What the render can never reach is not reported. An `{{if}}` or `{{with}}` whose
condition is a literal — `true`, `false`, `0`, `1`, `""`, `"x"`, or `not` of one —
never runs the side the literal rules out, even when the condition declares or
assigns a variable: `{{if $x := false}}{{.Nope}}{{end}}` reports nothing. `{{and}}` and `{{or}}` stop at an
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
  with `collage.Effect`, is still walked, but only its function calls are
  checked: their argument counts, and the arguments not read from `.`.
- **Nil pointers inside the data.** `{{.Author.Name}}` with a nil `Author` fails
  when it renders, and whether it is nil is the data's business, not the type's.
- **An argument whose fit the render decides.** A value of an interface type,
  a `reflect.Value`, a parameter of type `any` or `reflect.Value`, and a map
  entry handed to a parameter that can be nil (a key the map does not hold
  arrives as nil, which that parameter takes) are not judged. Nor are the
  arguments of a method of a value read by a map key, which the render never
  calls when the key is missing, nor an index out of range. A plugin's render
  functions take anything as far as the check knows, so their arguments are
  left to the plugin.
- **An integer constant's range.** `text/template` converts a constant to an
  integer parameter without a range check, so `{{i8 300}}` renders 44 rather
  than failing.

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
a test that registers the pages.

Under `collage dev` a template edited while the program runs is reparsed but not
checked again until the next restart, which is the next change to its Go code.
[`collage inspect`](/docs/cli#collage-inspect)
prints each fragment's data type, and the types it reaches, for an editor to check
the same way.

### Turning it off

`WithoutTypeCheck()` leaves one fragment's template out of the check, for the rare
finding that is wrong, while it is fixed. The fragment renders as before, and
`collage inspect` marks it `"typeCheck": false`:

```go
collage.NewFragment("post-body", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithoutTypeCheck().
	Build()
```

A handler declared to return an interface — `collage.Load[any]`, say — leaves the
type unknown, and its template is not walked: that is the way to say on purpose
that a fragment's data has no fixed shape. `collage.Value` is the exception. Its
value is in hand, so a value held in an interface is checked against the type it
holds.
