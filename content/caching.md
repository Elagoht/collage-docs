---
description: How collage caches rendered pages and the data they are made from, and how it knows what to throw away.
reference: CacheConfig, Cached, Once, TaggedCache, SkipCache, Vary, PageBuilder.Static, PageBuilder.Incremental, PageBuilder.Dynamic, FragmentBuilder.Static, StrategyAuto
---

# Caching

collage caches two things. The **page cache** stores what a render produced — the
HTML of one URL — so the next reader of that URL gets it without a render. The
**data cache**, `collage.Cached`, stores what renders are made from, so thirty
pages that show one author fetch that author once.

Both are indexed by the same **dependency tags**, so one call throws away a piece
of content and everything that was built from it.

Nothing is cached until you turn it on:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultTTL: 5 * time.Minute,
		MaxEntries: 10000,
	},
})
```

With `Enabled` false every request renders and nothing is stored, whatever the
pages declare.

## Render strategies

Each page says how its output may be reused, with one call on its builder:

| Builder call | What happens | `Cache-Control` sent |
| --- | --- | --- |
| `Dynamic()` | Rendered on every request, never stored | `no-store` |
| `Static()` | Rendered once, served until something invalidates it | `public, max-age=0, must-revalidate` |
| `Incremental(ttl)` | Served from the cache until `ttl` has passed since the render | `public, max-age=<ttl in seconds>` |

A page carrying a form's `{{csrfToken}}` is the exception to the last column: it is
cached like any other, but each reader is sent their own token, so the response
goes out `private, no-store` whatever the strategy — see
[Forms and actions](/docs/forms-and-actions#pages-with-forms-are-still-cached).

```go
page := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	WithDependency("blog:posts").
	Build()
```

`Static()` has no practical expiry, and `DefaultTTL` does not apply to it. It
renders again only when one of its tags is invalidated, or when the cache evicts
it to make room. That is the right strategy for content that changes when someone
publishes, and only then. Its `Cache-Control` tells browsers and CDNs to keep it
but ask every time, and the [ETag](#etags-and-304) makes asking cheap.

`Incremental(ttl)` is for content that changes on a clock, or that comes from
somewhere that cannot tell you when it changed. A TTL of zero is a registration
error (`collage.ErrMissingTTL`), and a negative one is too
(`collage.ErrInvalidTTL`) — not a page that silently never expires.

[Documents](/docs/documents) — sitemaps, feeds, anything that is not HTML — take
the same three calls and are stored, keyed, invalidated and — since v0.12.0 —
[coalesced](#concurrent-misses-render-once) the same way.

### A page that declares none

A page that calls none of the three is resolved when it is registered. Before
that its strategy is `collage.StrategyAuto`; after, it is one of the other three.

- **Dynamic** if anything it renders fetches per render: a data handler
  (`WithDataHandler`, including `collage.Load` and `collage.Effect`) or a slot
  resolver, in its layout, its content, anything bound into their slots, their
  fallbacks, or a fragment opened with `WithFragmentPath`.
- **Static** otherwise. A page rendering templates and fixed values —
  `WithData(v)`, `WithTitle(s)` — renders the same for every reader, so a site of
  such pages is cached and [exported](/docs/static-export) without a `Static()` on
  each one.

A handler means dynamic because it may read the request, a cookie, the clock, and
nothing outside the function can tell whether it does. When the guess is wrong
it costs a render, never a reader served somebody else's page: a page you forgot
to think about is slow, not stale. When a handler's output is the same for
everyone — a post read from a file — its page says `Static()` or
`Incremental(ttl)` itself.

What differs from page to page is not what makes a page dynamic. The path and its
parameters are part of the cache key, and each URL is its own file in an export:
`/recipes/pancakes` and `/recipes/omelette` are kept apart already. What makes a
page dynamic is output that differs between two requests to *one* URL — a user
from a cookie, the time. A fragment whose handler reads only the path's parameters
and the locale can say so itself, with `Static()` on the fragment, and then it does
not make a page dynamic. Since v0.17.0.

```go
more := collage.NewFragment("more-recipes", "fragments/more-recipes.html").
	WithDataHandler(loadMore). // every recipe but rc.Param("slug")
	Static().
	Build()
```

It is for a fragment many pages share — a list of recent posts, navigation built
from content — so that each page using it need not repeat the promise. It is the
same promise `Static()` makes for a page, and a handler that breaks it serves one
reader's render to the next. It covers the fragment's own handler and nothing
else: another fragment's handler still makes the page dynamic; so does a slot
resolver, even on a static fragment, since the fragments it returns are not known
until a render asks for them.

A page's declared strategy is never second-guessed, in either direction: a page
that says `Dynamic()` stays dynamic whatever its fragments say. Until v0.16.0 a
page that declared nothing was dynamic; one that relied on that, with no handler to
make it so, now says `Dynamic()`.

A [document](/docs/documents#a-fixed-body) resolves the same way: dynamic with a
handler, static with a fixed body.

## What is cached, and when

A response enters the page cache only when all of these are true:

- the cache is enabled and the page is static or incremental, declared or
  resolved;
- the request is a `GET`. A `HEAD` can be *served* from the cache, but it never
  fills it — it produced no body to store;
- every fragment rendered successfully.

That last rule matters more than it looks. A fragment that failed — even an
optional one, even one whose fallback covered for it — makes the render
**degraded**, and a degraded render is served but not stored. Storing it would pin
one request's passing failure in front of every reader until the TTL ran out.

Error responses are never cached either. A 404 or a 500 is written `no-store`, and
the tags an error page's own render declared are dropped. An
[action's](/docs/forms-and-actions) response is never cached, whatever the page it
rendered was declared as: it was produced from one submission and belongs to
whoever sent it.

Files served from an [asset mount](/docs/assets) do not go through the page cache
at all. Their freshness is their `Cache-Control` header and nothing else.

## The cache key

A cached page is found by a key made from:

- the request path,
- the resolved locale,
- the captured path parameters,
- the query string (see [below](#query-parameters-in-the-key)),
- any value your middleware declared with `collage.Vary`.

Two requests with the same key are, as far as collage is concerned, the same page.

## Dependency tags

A tag is any string that names a piece of content: `post:hello-world`,
`author:ada`, `blog:posts`. A page's tags come from two places, and they are
merged:

```go
// From the page, for what it always depends on.
WithDependency("blog:posts")

// From a data handler, for what this particular render used.
return view, []string{"post:" + post.Slug, "author:" + post.AuthorID}, nil
```

The page's own tags are known before anything renders. The data handler's are
the useful ones, because only the handler knows which post and which author this
URL turned out to show. Tags are collected even when a handler then returns an
error: it still said what the page depends on.

When the page is stored, its tags are stored with it, and collage remembers which
cache keys were built from each tag.

### Invalidating

When content changes, name it:

```go
if err := app.InvalidateTags(ctx, "post:"+slug, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Every cached page carrying any of those tags is dropped, and so is every value
`collage.Cached` stored under them. Nothing else is touched. The next reader of
each dropped page gets a fresh render.

`InvalidateTagsN` does the same and tells you how many cache keys the tags
reached:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

The count is the number of keys collage's own tag index resolved and dropped, not
an exact figure for live pages removed, and it errs both ways: a key whose entry
had already expired is still counted, and an entry the store reached through its
own tag index — past the cap [below](#the-tag-index-is-per-process-and-bounded), or
written by another instance — is removed without being counted. It is good for a
log line or a metric, not for logic.

If the cache fails to drop some keys, the rest are still dropped and the failures
come back joined in the error. A partial invalidation that reported success is how
stale pages survive a deploy.

The usual place to call it is wherever your content changes: a CMS webhook, an
admin form. An action can do it declaratively, with
[`InvalidateTags` on its result](/docs/forms-and-actions#invalidating-what-an-action-changed),
which runs before the response is written — so a reader redirected to the page
they just changed never sees the old version.

### The tag index is per process, and bounded

collage's record of which keys belong to which tag lives in memory, in the process
that wrote them. Two things follow.

**Behind a load balancer, each instance knows only its own keys.** With the
built-in caches that is fine, because each instance also has its own cache. With a
shared store of your own — `Cache.Store`, a Redis say — implement
`collage.TaggedCache` too, so the store indexes tags itself; collage then asks it
to invalidate by tag as well, and it can reach what another instance wrote. The
built-in disk cache does this, which is why its tags still work after a restart.

**The record per tag is capped by `Cache.MaxKeysPerTag`** (default 10000). The
query string is part of the key, so a client can mint any number of keys for one
page; without a cap the index would grow for ever. When a tag reaches the cap, the
oldest key recorded under it is forgotten by that index. The built-in memory and
disk caches both implement `collage.TaggedCache`, keeping their own record of tags
per entry, so with them invalidating the tag still reaches every entry — only the
count `InvalidateTagsN` reports falls short. A store of your own that does not
implement `TaggedCache` has only collage's index to go by: a forgotten page stays in
it until it expires, and invalidating the tag no longer reaches it. Set the cap
above the number of cached URLs one tag can really cover, or negative for no cap.

## Query parameters in the key

By default the whole raw query string is part of the key. That is the only safe
default: a data handler receives the whole request and may read
`rc.Request.URL.Query()`, so `?page=2` could be a different page, and collage
cannot know.

It is also expensive. A newsletter link with `?utm_source=newsletter` stores a
second copy of the page, and a crawler trying variants fills the cache with copies
nobody asked for, evicting real ones. Say which parameters the page reads:

```go
collage.NewPage("articles").
	WithLayout(layout).
	WithContent(list).
	WithPath("en", "/articles").
	WithCacheParams("page", "sort").
	Incremental(time.Minute).
	Build()
```

Now only `page` and `sort` are in the key, and they are put in a fixed order, so
`?page=2&sort=new` and `?sort=new&page=2` share one entry. Everything else in the
query is ignored for caching — the handler can still read it, but it must not
change what the page shows.

`WithCacheParams()` with no names drops the query from the key entirely: the page
renders the same whatever the query says.

Naming a parameter the page does not read costs nothing. Failing to name one it
does read is a bug: two different pages share one entry, and one reader is served
another's. And some pages should not be cached at all — a search page's key would
be the search term, which is whatever the reader typed. That is a page that wants
`Dynamic()`.

Documents have the same `WithCacheParams`.

## Memory or disk

`Type: "memory"`, the default, keeps pages in the process. It is fast and it is
empty after every restart. It holds at most `MaxEntries` pages (default 10000;
negative for no limit) and, when full, drops the oldest by insertion — reading a
page does not make it younger. An expired entry is dropped when it is next looked
up.

`Type: "disk"` keeps pages as files, so a restart does not render everything
again:

```go
Cache: collage.CacheConfig{
	Enabled: true,
	Type:    "disk",
	Dir:     ".cache",
},
```

`Dir` has no default: a framework that picks where to write files writes them
somewhere nobody looked. Add it to `.gitignore`. Since v0.11.0 a directory that
cannot be created — a read-only filesystem, a container with nowhere to write — is
not a reason not to start: collage logs a warning and caches in memory instead.

### The namespace

A disk cache outlives the process that filled it, and that is a hazard as well as
the point. A new binary with a changed template must not serve HTML the old binary
rendered.

So entries live in a subdirectory named for the build. Leave `Cache.Version`
empty and it is a hash of the running executable: it changes exactly when your
code or templates compiled into it change, and two runs of the same build — or
every machine in a fleet running it — share one cache. Set `Version` yourself only
when something outside the binary decides what the output looks like, such as a
content revision.

The forgery-protection key is not part of the namespace, so a disk cache survives
a restart whether or not `Security.CSRFKey` is set. The key still matters to one
kind of page: a cached page with a form in it carries a marker derived from the key
where each reader's token goes (see
[Forms and actions](/docs/forms-and-actions#pages-with-forms-are-still-cached)),
and a page stored under one key cannot be served under another. So a stored page
carrying another key's marker — rendered before the key changed, or by a process
that generated its own — is treated as a miss: it is dropped and rendered again.
Pages without a form are unaffected by the key entirely.

Everything that shares a `Dir` and a build shares entries, including two apps in
one test binary. Give each test its own directory with `t.TempDir()` — see
[Testing](/docs/testing#isolate-the-disk-cache).

### Your own store

`Cache.Store` takes any `collage.Cache` implementation, and `Type` is then
ignored. `Enabled` is still the master switch. Do not assign a typed nil pointer to
it — a nil `*myCache` in an interface field is not a nil interface, and collage will
call straight through it.

## ETags and 304

Every cached page is stored with an ETag, a hash of its content, and every
response served from the cache carries it. A browser or CDN that already has the
page sends it back in `If-None-Match`, and if it still matches, collage answers
`304 Not Modified` with no body.

This is what makes `Static()` pages cheap to revalidate: `must-revalidate` means
the client asks every time, and the answer is usually a few bytes.

A page with a form is the exception again. What a reader is sent carries their own
forgery token, so its ETag names that reader's copy rather than the stored one, and
the response is `private, no-store`: a browser that does send it back gets a `304`
only for the copy it was given, and nothing shared keeps it.

## Concurrent misses render once

When a popular page expires, every request that arrives before the first
re-render finishes is a miss. Left alone, each would render: the same page, the
same upstream calls, at the same moment, in numbers that grow with traffic.

collage does not let that happen. The first request for a key renders, and the
others that arrive meanwhile wait for it and are served the same bytes. There is
nothing to configure.

- **Only cached routes coalesce.** A dynamic page has no cache key, so two
  requests are two renders, as the page asked. A cached [document](/docs/documents)
  coalesces like a page since v0.12.0: an expiring feed that many clients poll runs
  its handler once.
- **One reader giving up does not fail the others.** A request whose connection
  closes stops waiting. If the rendering request itself is cancelled, the ones
  waiting behind it try again instead of receiving its error.
- **It is visible.** A request served this way is reported to your metrics twice:
  as a `CacheMiss`, when its lookup found nothing, and then as `CacheCoalesced`,
  when it was served another request's render. So the renders a key cost are its
  misses minus its coalesced events. A coalesced count that climbs steadily means a
  page is expiring faster than it can be rendered, which is what a too-short
  `Incremental` TTL looks like from outside.

## Development never reads the cache

With `Config.DevMode` (or `Template.DevMode`) on — which a scaffolded project
turns on under `collage dev` — cached pages are never served. Templates reload from disk on every request
in development, and a cached page would hide the edit you just made for as long as
its TTL. Pages are still written and tags still tracked, so hooks and metrics
behave as they will in production; what is gone is serving a page rendered before
the edit.

A disk cache is replaced by a memory one in development, and collage logs that it
did. `collage.Cached` keeps nothing across renders there either (see below).

## Caching data across pages

The page cache stores whole pages. It does nothing for thirty different blog posts
that each show their author: each is a different URL, each renders on its own, and
each fetches the author. A static export, which renders every one of them, fetches
the author thirty times.

`collage.Cached` stores the author:

```go
func authorCard(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	id := rc.Param("author")
	author, err := collage.Cached(rc, "author:"+id, time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return api.Author(ctx, id)
		})
	return author, nil, err
}
```

```go
func Cached[T any](rc *RenderContext, key string, ttl time.Duration, tags []string,
	fetch func(context.Context) (T, error)) (T, error)
```

The first render that asks for `author:ada` calls `fetch`; every later render, on
any page, gets the stored value. With `Cache.Enabled` on, thirty posts by two
authors now make two author requests — served or exported. (With it off, and in
development, there is no store; see [below](#where-it-keeps-nothing).) A [document's](/docs/documents) handler shares the
same store, so a sitemap or a feed reading those authors fetches none of them
again.

### One set of tags for both caches

The `tags` you pass are added to the page's own tags, as if the data handler had
returned them. So a single call —

```go
app.InvalidateTags(ctx, "author:ada")
```

— drops the stored author *and* every cached page that showed her, together.

This is the part that is easy to get wrong without it. An application that
memoises authors inside its own API client, and invalidates pages by tag, has two
caches that must be cleared in step. Clear only the pages, and they re-render from
the stale author they were meant to replace.

### How it behaves

- **The key is yours.** It names the value across the whole application, so make
  it as specific as the fetch: `author:ada`, not `author`. Asking for one key as
  two different types is `collage.ErrCachedTypeMismatch`.
- **One fetch per key at a time.** Renders that ask for a key while it is being
  fetched wait for that fetch rather than starting their own.
- **Errors are not stored.** Everyone waiting gets the error; the next render
  tries again.
- **An invalidated fetch is not stored.** If a key's tags are invalidated while its
  fetch is still running, the result goes to whoever was waiting but is not kept —
  it is exactly what the invalidation was meant to replace.
- **`ttl` is independent of the page's.** It bounds how long the value is kept when
  nothing invalidates it; zero keeps it until something does. A page that
  re-renders every minute can still reuse an author fetched an hour ago, which is
  the point.
- **In memory, bounded.** Values are kept in the process, up to
  `Cache.MaxEntries`, the least recently used going first — even when pages are
  cached on disk. Each instance of a multi-instance deployment keeps its own.

### Where it keeps nothing

With `Cache.Enabled` false, in development, for a request whose middleware called
`collage.SkipCache` (a [preview](/docs/previews)), and inside an action's own
handler, `Cached` stores nothing and behaves like `collage.Once`: fragments of the
same render share one fetch, and the next render fetches again. A preview
therefore sees fresh data as well as a fresh page.

### Once, Cached and the page cache

| | Shared between | Lives for |
| --- | --- | --- |
| `collage.Once` | the fragments of one render | that render |
| `collage.Cached` | every render in the process | its `ttl`, or until its tags are invalidated |
| page cache | every request for one URL | the page's strategy, or until its tags are invalidated |

Use `Once` for what one page fetches twice, `Cached` for what many pages fetch,
and the page cache for the page itself. They combine: a cached page is not
rendered, so none of its fetches run at all. See
[Data handlers](/docs/data-handlers) for `Once`.

A page refreshed part by part through its
[fragment paths](/docs/forms-and-actions#a-fragment-at-its-own-url) is several
renders, one per fragment path, and `Once` shares nothing between them. Fragments
refreshed separately that read the same data share a fetch only through `Cached`.
