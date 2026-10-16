---
description: How collage caches rendered pages and the data they are made from, and how it knows what to throw away.
reference: CacheConfig, Cached, App.PurgeTags, CacheStale, Once, Key, TaggedCache, StaleCache, CacheStatusHeader, SkipCache, Vary, PageBuilder.Static, PageBuilder.Incremental, PageBuilder.Dynamic, FragmentBuilder.Static, FragmentBuilder.Shared, StrategyAuto, PathTag
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
		MaxBytes:   256 << 20, // 0 means the type's default; negative, no limit
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

The last column has exceptions. A
[guarded page](/docs/pages-and-layouts#private-pages-guards) is one: its readers
share this cache, because the guard runs before the cache is read, but a CDN or
proxy in front of the server runs no guard. So the response goes out
`private, no-cache` whatever the strategy: the reader's own browser may keep it and
revalidate with its ETag, and a revalidation is a request the guard sees.

A page carrying a form's `{{csrfToken}}` is the other: it is
cached like any other, but each reader is sent their own token, so the response
goes out `private, no-store` whatever the strategy — see
[Forms and actions](/docs/forms-and-actions#pages-with-forms-are-still-cached).

Since v0.34.0 there are two more. A degraded render (see
[below](#what-is-cached-and-when)) goes out `no-store`: it is not kept by this
server, so it must not be kept by a CDN in front of it for the page's TTL either.
So does a stale page served in place of a failed render (since v0.63.0).
And in development every page and document goes out `no-store`, so the browser
does not keep a page from before your edit.

```go
page := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	WithDependencyTags("blog:posts").
	Build()
```

`Static()` has no practical expiry, and `DefaultTTL` does not apply to it. It
renders again only when one of its tags is invalidated, or when the cache evicts
it to make room — and when that render fails, the invalidated copy is what answers
(see [Stale pages](#stale-pages-and-x-collage-cache)). That is the right strategy
for content that changes when someone
publishes, and only then. Its `Cache-Control` tells browsers and CDNs to keep it
but ask every time, and the [ETag](#etags-and-304) makes asking cheap.

`Incremental(ttl)` is for content that changes on a clock, or that comes from
somewhere that cannot tell you when it changed. A TTL of zero is a registration
error (`collage.ErrInvalidTTL`), and a negative one is too — not a page that silently never expires.

[Documents](/docs/documents) — sitemaps, feeds, anything that is not HTML — take
the same three calls and are stored, keyed, invalidated and — since v0.12.0 —
[coalesced](#concurrent-misses-render-once) the same way.

### A page that declares none

A page that calls none of the three is resolved when it is registered. Before
that its strategy is `collage.StrategyAuto`; after, it is one of the other three.

- **Dynamic** if anything it renders fetches per render: a data handler
  (`collage.LoadTagged`, `collage.Load` or `collage.Effect`) or a slot
  resolver, in its layout, its content, anything bound into their slots, their
  fallbacks, or a fragment opened with `WithFragmentPath`.
- **Static** otherwise. A page rendering templates and fixed values —
  `collage.Value(v)`, `WithTitle(s)` — renders the same for every reader, so a
  site of such pages is cached and [exported](/docs/static-export) without a
  `Static()` on each one.

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
	WithData(collage.LoadTagged(loadMore)). // every recipe but rc.Param("slug")
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

`Static()` promises two things at once: the output is the same for every reader,
and it stays the same until something is invalidated. A handler that reads no
cookie but returns a measurement — CPU load, a queue's length, the time — keeps
the first promise and not the second: it is the same for every reader, not the
same over time. Mark it `Shared()` instead. Since v0.19.0.

```go
cpu := collage.NewFragment("cpu", "fragments/cpu.html").
	WithData(collage.Load(cpuUsage)). // the same for everyone, different every second
	Shared().
	Build()
```

`Shared()` leaves the page's strategy alone — the page stays dynamic, and an
export does not write one moment's reading into it — and tells whatever pushes the
fragment to readers that one render may be sent to all of them (see
[Writing plugins](/docs/writing-plugins#pushing-fragments)).
`Static()` implies it. As with `Static()`, a handler that breaks the promise sends
one reader's data to another.

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
Where a stale copy of the page is kept, that copy is served instead (see
[Stale pages](#stale-pages-and-x-collage-cache)).

Error responses are never cached either. A 404 or a 500 is written `no-store`, and
the tags an error page's own render declared are dropped. A 404 also drops the
stale entry the request found. An
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
- any value your middleware declared with `collage.Vary` (read back in a handler
  with `collage.Varied`),
- the request's host (since v0.34.0).

Two requests with the same key are, as far as collage is concerned, the same page.

The host is there because a render may build an absolute URL from `r.Host` — a
canonical link, an `og:url` — and one request naming another host would otherwise
write that URL into the copy every reader is served.

### What a shared render sees

A cached render is served to every reader whose request has its key, so since
v0.34.0 it is handed only what the key holds: the path, the host, the query
parameters in the key, and the headers your middleware declared with
`collage.Vary`. Not the reader's cookies, `Authorization`, address or client
certificate, and not a query parameter `WithCacheParams` left out.

A handler on a static or incremental page that reads a cookie finds none — where
it used to find the first reader's, and write it into the copy everyone after got.
The request's context is stripped the same way: a value your middleware stored in
it — a signed-in user, a tenant, a request id — is hidden from a shared render,
for the same reason the cookie is. Read a value your middleware declared with
`collage.Vary` back with `collage.Varied(ctx, header)`, with the context a data
handler, a page's `StaticParams`, a sitemap's `LastMod` or a document handler
receives (since v0.42.0; before v0.60.0 it took the `RenderContext`, and
`VariedContext` took a context); it is safe there because it
is in the key, so each value has its own cached page. A value that is not a cache
dimension does not belong in a shared render — a page that needs the reader's own
request is `Dynamic()`, and a dynamic page sees all of it. In development, a
hidden context value a handler actually reads is logged, so the mistake is
visible. Documents follow the same rule (since v0.41.1; before it, a cacheable
document's render still saw the first reader's context values).

### Origins

`collage.BaseURL(rc)` is safe in a shared render (since v0.42.0). The host is in the
cache key, so every reader served a copy asked for the same host, and the origin
the render built absolute URLs against is theirs too. A static build has no request
and gets `Config.BaseURL`. A site with no `OriginResolver` plugin sees no
difference; see [Which origin a host has](/docs/writing-plugins#which-origin-a-host-has-originresolver).
When an invalidation reaches entries, `CacheInvalidateEvent.Entries` names each
one's host as well as its path.

## Dependency tags

A tag is any string that names a piece of content: `post:hello-world`,
`author:ada`, `blog:posts`. A page's tags come from two places, and they are
merged:

```go
// From the page, for what it always depends on.
WithDependencyTags("blog:posts")

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

Every cached page carrying any of those tags is marked stale (since v0.63.0; before,
it was dropped). It is no longer served as fresh: the next request for each one
renders it again, and a good render is stored in its place. But the old page is
kept, to answer with when that render fails or comes out degraded — see
[Stale pages](#stale-pages-and-x-collage-cache).
Every value `collage.Cached` stored under the tags is marked stale too: kept, and
served once more while a fetch in the background replaces it (see
[How it behaves](#how-it-behaves)). Nothing else is touched.

That is the built-in stores, memory and disk. A store of your own keeps dropping
entries on `InvalidateTags`, unless it implements
[`collage.StaleCache`](#stale-pages-in-your-own-store).

`App.PurgeTags` takes the same tags and removes outright, with nothing left to
serve stale, from disk too:

```go
if err := app.PurgeTags(ctx, "post:"+slug); err != nil {
	log.Printf("purge: %v", err)
}
```

Use it for data that must not be shown again even once. The cost is that a failed
fetch, or a failed render, then fails the page, as there is nothing to fall back
on.

`InvalidateTagsN` does the same and tells you how many cache keys the tags
reached:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

The count is the number of keys collage's own tag index resolved and invalidated,
not an exact figure for live pages reached, and it errs both ways: a key whose
entry had already expired is still counted, and an entry the store reached through
its own tag index — past the cap [below](#the-tag-index-is-per-process-and-bounded),
or written by another instance — is invalidated without being counted. It is good
for a log line or a metric, not for logic.

If the cache fails to invalidate some keys, the rest are still invalidated and the
failures come back joined in the error. A partial invalidation that reported success is how
stale pages survive a deploy.

The usual place to call it is wherever your content changes: a CMS webhook, an
admin form. An action can do it declaratively, with
[`InvalidateTags` on its result](/docs/forms-and-actions#invalidating-what-an-action-changed),
which runs before the response is written — so a reader redirected to the page
they just changed is not served the old page as though it were fresh: it renders
again. A page that reads `collage.Cached` data still lands on the old page for
that one load, served stale while its data refreshes: see
[the note there](/docs/forms-and-actions#invalidating-what-an-action-changed).

### Invalidating by path

Every cached page and document also depends on a tag naming the URL path it was
rendered for, `collage.PathTag(path)` (since v0.23.0), so an entry can be
invalidated by its path when that is what you know rather than a tag:

```go
app.InvalidateTags(ctx, collage.PathTag("/blog/hello"))
```

It invalidates what is cached for that path, whatever else it depends on.

An invalidation also says what it reached. A plugin hears of every one through
`OnCacheInvalidate`, and the event carries the tags and — in `Entries`, each a host
and a path — the cached pages and documents invalidated: what a CDN has to purge,
and a search engine be told has changed. A page that was never cached is not in it. See
[Writing a plugin](/docs/writing-plugins#cacheinvalidatehook);
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) and
[elagoht/indexnow](/docs/plugins#elagohtindexnow) are built on it.

### The tag index is per process, and bounded

collage's record of which keys belong to which tag lives in memory, in the process
that wrote them. Two things follow.

**Behind a load balancer, each instance knows only its own keys.** With the
built-in caches that is fine, because each instance also has its own cache. With a
shared store of your own — `Cache.Store`, a Redis say — implement
`collage.TaggedCache` too, so the store indexes tags itself; collage then asks it
to invalidate by tag as well, and it can reach what another instance wrote. The
built-in disk cache does this, which is why its tags still work after a restart.

**The record per tag is capped by `Cache.MaxEntries`** (default 10000; before
v0.60.0 a field of its own, `MaxKeysPerTag`). The
query string is part of the key, so a client can mint any number of keys for one
page; without a cap the index would grow for ever. When a tag reaches the cap, the
oldest key recorded under it is forgotten by that index. The built-in memory and
disk caches both implement `collage.TaggedCache`, keeping their own record of tags
per entry, so with them invalidating the tag still reaches every entry — only the
count `InvalidateTagsN` reports falls short. A store of your own that does not
implement `TaggedCache` has only collage's index to go by: a forgotten page stays in
it until it expires, and invalidating the tag no longer reaches it. Set `MaxEntries`
above the number of cached URLs one tag can really cover; a negative `MaxEntries`
keeps this cap at 10000.

**With a built-in cache the index is also capped as a whole** (since v0.40.0), at
`MaxEntries` keys. Every cached path is a tag of its own, so a per-tag cap alone
let the index reach that cap times the number of pages — 10000 keys for each page
of a large site, flooded with invented queries, while the cache itself held
10000 entries in all. No more keys than `MaxEntries` can still be live, so the
oldest written go first; a negative `MaxEntries` leaves this off. A store of your
own keeps the per-tag cap only, since collage cannot know its size.

## Stale pages and `X-Collage-Cache`

Since v0.63.0 a page or document that is stale — marked so by `InvalidateTags`, or
past its TTL — is not thrown away by the built-in stores. A request for it renders
again, through the same shared render as any miss
([below](#concurrent-misses-render-once)), and what happens next depends on that
render:

- **A good render** is stored and served in place of the stale page. The response
  is a `MISS`.
- **A render that fails, or comes out degraded**, is not stored, and the stale
  page is served instead, with `Cache-Control: no-store` so that no browser or
  CDN keeps it, and `X-Collage-Cache: STALE`. The reader sees the old page, not a
  500. Reading a stale `collage.Cached` value makes a render degraded, and that is
  the usual cause right after an invalidation: the page's data is stale until the
  background fetch replaces it.
- **A render that answers 404** drops the entry: the content is gone.

Before v0.63.0 an invalidated page was gone, so an invalidation followed by a
failing render was a 500. `PurgeTags` still leaves nothing to fall back on, so a
purged page that fails to render fails.

An expired page is stale in the same sense, so an `Incremental` page whose upstream
is down keeps answering with its last good render after its TTL. There is no
maximum stale age. Only eviction at `MaxEntries` or `MaxBytes` takes a stale page
away.

A render that overlaps any `InvalidateTags` or `PurgeTags` call — one that began
while a call was under way, or during which one began — is not stored, whatever
its tags: it may have read the data from before the change, and storing it would
bring the old page back as fresh.

**A stale serve is reported, once per failure.** Each stale response is counted as
the `CacheStale` metrics event. The failure it hides is reported once per failed
render, however many requests were waiting on it: logged once at warn
(`collage: serving a stale page, its render failed`, with the path and the error) —
not also as an error-level `collage: request failed`, since no error page was
served — and handed to plugins' `OnError` hook once, with a `Status` of 0, since no
error response was sent. A render degraded only because it read stale
`collage.Cached` data is no outage — its data is refreshing — so its stale serve is
logged at debug (`collage: serving a stale page while its data refreshes`), and
nothing reaches `OnError`. Documents are reported the same way.

**`X-Collage-Cache` says which it was.** `collage.CacheStatusHeader` is
`X-Collage-Cache`, and a cacheable `GET` or `HEAD` that answered with a page or
document carries one of:

| Value | Meaning |
| --- | --- |
| `HIT` | A fresh entry, served without rendering. |
| `MISS` | Rendered — there was no entry, or only a stale one. |
| `STALE` | The stale entry, served because the render replacing it failed or was degraded. |

A `304` answer to a conditional request carries it too. It is not sent in
development, where the cache is not read, on an error response, on a page that is
not cacheable (`Dynamic()`, `collage.SkipCache`), or in a static build, and a store
of your own that is not a `StaleCache` never answers `STALE`. It is there for
whoever debugs from outside; collage itself does not read it.

## Query parameters in the key

By default the whole raw query string is part of the key. That is the only safe
default: a data handler receives the whole request and may read
`rc.Request.URL.Query()`, so `?page=2` could be a different page, and collage
cannot know.

This default is deliberate and will not change. Dropping the query, or keeping only
an allowlist of it, by default would make every page that reads a parameter it did
not declare share one entry across all values of it: `?page=2` served as page one,
one visitor's filter served to the next, and nothing would fail to say so. A page
opts into a smaller key with `WithCacheParams`; the framework never guesses one for
it. What keeps an unbounded number of query variants from becoming unbounded memory
is the cache's bounds, not the key: `Cache.MaxEntries`, `Cache.MaxBytes` and the
tracker's caps (see [the tag index is bounded](#the-tag-index-is-per-process-and-bounded)).
A client minting variants evicts entries, which costs hit rate, not memory.

It is also expensive. A newsletter link with `?utm_source=newsletter` stores a
second copy of the page, and a crawler trying variants fills the cache with copies
nobody asked for, evicting real ones. Say which parameters the page reads:

```go
collage.NewPage("articles").
	WithLayouts(layout).
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
negative for no limit) and at most `MaxBytes` of them (default 256 MiB), and when
either is full drops the oldest by insertion — reading a page does not make it
younger. An expired entry is not served as fresh, but it is kept, as
[stale](#stale-pages-and-x-collage-cache), until eviction needs the room.

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
somewhere nobody looked. Add it to `.gitignore`. The `collage.Cached` values are
kept in `Dir/data/` too: see [Across a restart](#across-a-restart). Since v0.11.0 a directory that
cannot be created — a read-only filesystem, a container with nowhere to write — is
not a reason not to start: collage logs a warning and caches in memory instead.

Since v0.39.0 the disk cache holds at most `MaxEntries` too — the same default of
10000, a negative value for a cache you mean to leave unbounded — and evicts the
oldest by file modification time when it is full. The cache key includes the
request's host and, by default, its whole query, so without the cap an anonymous
caller varying either could fill the disk one never-evicted file at a time.

Since v0.40.0 the eviction scan runs outside the lock writes take, so a flood of
new keys no longer makes every writer wait on a scan of the whole directory; a new
entry that reaches a full cache while a scan is still making room is served and
not stored, so the cap holds.

### Bounded in bytes too

`MaxEntries` counts entries, not what they weigh, and the key holds the query by
default: one large page asked for under thousands of invented queries
(`?utm=1`, `?utm=2`, ...) fills the cache with copies of it. Ten thousand copies of
a 57 KB documentation page is over half a gigabyte of memory; of a 500 KB listing,
five. `MaxBytes` (since v0.40.0) bounds the weight: past it the oldest entries are
evicted, as at `MaxEntries`, and a page larger than the whole cap is served but not
stored.

| `MaxBytes` | Memory | Disk |
| --- | --- | --- |
| `0` | 256 MiB | 1 GiB (its memory tier: 256 MiB) |
| negative | no limit | no limit |

It counts the content a memory cache holds, and the size of the files on disk. A
disk cache's memory tier ([below](#memory-in-front-of-the-disk)) holds 256 MiB,
or `MaxBytes` when that is smaller.
Go's allocator rounds a large value up to whole pages, so a memory cache full of
50 KB pages takes some 10% more heap than `MaxBytes` says, and the process takes
more than its heap: in a container, set `GOMEMLIMIT` too — see
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit). A store of your own
bounds itself and ignores the field. Naming the query parameters a page reads with
[`WithCacheParams`](#query-parameters-in-the-key) is the other half: a page that
ignores the query stores one copy, not thousands.

### Memory in front of the disk

Since v0.63.0 a `"disk"` cache is two tiers: a [memory cache](#memory-or-disk) in
front of the directory. Reading a page from a file is an open, a read and a decode
on every request, so a hit is served from memory instead. A write goes to disk and
then to memory, and a read asks memory first and copies what it finds on disk into
memory as it goes. `MaxEntries` bounds both tiers. `MaxBytes` bounds the disk; the
memory tier holds 256 MiB, or `MaxBytes` when that is smaller. That is memory of
your process on top of the disk, so in a small container set `GOMEMLIMIT` as for a
memory cache, and lower `MaxBytes` or `MaxEntries` if the tier must hold less (see
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit)).

**A restart starts warm.** At startup the memory tier is filled from the current
build's directory, newest entries first, until the next one would make it evict.
The pages written most recently are `HIT`s from memory at once, and the rest are
one file read away. Stale entries are warmed too, so a restart in the middle of an
upstream outage still has them to serve. Files written by v0.62 or earlier carry
no key in their header and are not warmed; they are still served, on the first
request that reads them.

**One `Dir` per instance.** Instances that share a `Dir` do not share a cache.
Each has its own memory tier, and memory's answer is final once it holds a page.
An invalidation in one instance marks the disk and its own memory; every other
instance goes on serving the copy in its memory as a fresh `HIT` until that copy
expires — for a `Static` page, which has no TTL, until the instance restarts or
evicts it — and a CDN that refetches the page meanwhile can be handed the old copy
and keep it. Give each instance its own directory. A fleet that must see one
another's invalidations needs a shared [store of its own](#your-own-store) that
implements `collage.StaleCache` and `collage.TaggedCache`, instead of
`Type: "disk"`.

### The namespace

A disk cache outlives the process that filled it, and that is a hazard as well as
the point. A new binary with a changed template must not serve HTML the old binary
rendered.

So entries live in a subdirectory named for the build. Leave `Cache.Version`
empty and it is a hash of the running executable: it changes exactly when your
code or templates compiled into it change, and two runs of the same build — or
every machine in a fleet running it — derive the same directory. That keeps one
build from reading another's pages; it does not make a fleet share one cache (see
[above](#memory-in-front-of-the-disk)). Set `Version` yourself only
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

Everything that shares a `Dir` and a build shares the files on disk, including
two apps in one test binary. Give each test its own directory with `t.TempDir()` — see
[Testing](/docs/testing#isolate-the-disk-cache).

### Your own store

`Cache.Store` takes any `collage.Cache` implementation, and `Type` is then
ignored. `Enabled` is still the master switch. Do not assign a typed nil pointer to
it — a nil `*myCache` in an interface field is not a nil interface, and collage will
call straight through it.

`InvalidateKey` and `Invalidate` remove. A store that implements only these keeps
the behaviour from before v0.63.0: an invalidated page is gone, and a re-render
that fails fails the request.

### Stale pages in your own store

Implement `collage.StaleCache` as well when your store can keep an entry past its
invalidation or expiry. The framework type-asserts for it, and asks a store that
has it to mark entries stale instead of removing them, which is what lets a failed
re-render be answered with the last page that rendered:

```go
type StaleCache interface {
	Cache
	GetStale(ctx context.Context, key string) (content []byte, etag string, stale, found bool)
	MarkStaleKey(ctx context.Context, key string) error
	MarkStale(ctx context.Context, tags []string) error
}
```

- `GetStale` is `Get` that also returns an entry past its invalidation or its
  expiry, with `stale` true. `found` is false only when the key has no entry at
  all.
- `Get` still reports neither a stale entry nor an expired one: they are misses,
  as before.
- `MarkStaleKey` keeps the entry, with its content, ETag, expiry and tags, and
  marks it. Marking a key with no entry is not an error. A later `Set` or
  `SetTagged` for the key stores a fresh entry, which clears the mark.
- `MarkStale` does the same for every entry carrying any of the tags, which only
  a store that indexes tags can do. A `StaleCache` that is also a `TaggedCache` is
  asked to mark by tag instead of to invalidate.
- Nothing bounds how long an entry stays stale. Evict it yourself, under your own
  limits.

`PurgeTags` still calls `InvalidateKey` and `Invalidate`. A page already marked
stale has been forgotten by collage's own tag index, so only a store that is also
a `TaggedCache` can purge it by tag: a `StaleCache` that is not keeps such a page
until it is evicted or its key is purged.

```go
var _ collage.StaleCache = (*myCache)(nil)
```

## ETags and 304

Every cached page is stored with an ETag, a hash of its content, and every
response served from the cache carries it. A browser or CDN that already has the
page sends it back in `If-None-Match`, and if it still matches, collage answers
`304 Not Modified` with no body.

This is what makes `Static()` pages cheap to revalidate: `must-revalidate` means
the client asks every time, and the answer is usually a few bytes. A `304` carries
[`X-Collage-Cache`](#stale-pages-and-x-collage-cache) like the full answer would.

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
  closes stops waiting. The render itself belongs to none of the requests waiting
  on it, so it does not stop when the first of them does: since v0.18.1 it runs
  detached from that request's cancellation, and since v0.39.0 with its per-reader
  context values stripped as well, bounded by the fragments' own timeouts. One reader pressing stop cannot hand everyone who asked
  at the same moment a page whose parts failed with "context canceled".
- **It is visible.** A request served this way is reported to your metrics twice:
  as a `CacheMiss`, when its lookup found nothing, and then as `CacheCoalesced`,
  when it was served another request's render. So the renders a key cost are its
  misses minus its coalesced events. A coalesced count that climbs steadily means a
  page is expiring faster than it can be rendered, which is what a too-short
  `Incremental` TTL looks like from outside.

## Development never reads the cache

With `Config.DevMode` on — which a scaffolded project
turns on under `collage dev` — cached pages are never served. Templates reload from disk on every request
in development, and a cached page would hide the edit you just made for as long as
its TTL. Pages are still written and tags still tracked, so hooks and metrics
behave as they will in production; what is gone is serving a page rendered before
the edit.

A disk cache is replaced by a memory one in development, and collage logs that it
did. No response carries `X-Collage-Cache` there. `collage.Cached` keeps nothing across renders there either (see below).

## Caching data across pages

The page cache stores whole pages. It does nothing for thirty different blog posts
that each show their author: each is a different URL, each renders on its own, and
each fetches the author. A static export, which renders every one of them, fetches
the author thirty times.

`collage.Cached` stores the author:

```go
var authorKey = collage.NewKey[Author]("author")

func authorCard(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	id := rc.Param("author")
	author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return api.Author(ctx, id)
		})
	return author, nil, err
}
```

```go
func Cached[T any](rc *RenderContext, key Key[T], ttl time.Duration, tags []string,
	fetch func(context.Context) (T, error)) (T, error)
```

The first render that asks for `author:ada` calls `fetch`; every later render, on
any page, gets the stored value. With `Cache.Enabled` on, thirty posts by two
authors now make two author requests — served or exported. (With it off, and in
development, there is no store; see [below](#where-it-keeps-nothing).) A [document's](/docs/documents) handler shares the
same store, so a sitemap or a feed reading those authors fetches none of them
again.

The key is a [`collage.Key`](/docs/data-handlers#collagekey), the same kind the
fragments of one render share values through: a name and a type, declared once,
with `With` adding the part that varies — `authorKey.With("ada")` is named
`author:ada`. A `fetch` returning anything but the key's type does not compile.
Tags stay plain strings — `"author:" + id` here — because they are what
`InvalidateTags` is called with. Before v0.50.0 the key was a string.

### One set of tags for both caches

The `tags` you pass are added to the page's own tags, as if the data handler had
returned them. So a single call —

```go
app.InvalidateTags(ctx, "author:ada")
```

— marks the stored author stale *and* every cached page that showed her, together.

This is the part that is easy to get wrong without it. An application that
memoises authors inside its own API client, and invalidates pages by tag, has two
caches that must be cleared in step. Clear only the pages, and they re-render from
the old author they were meant to replace.

### How it behaves

- **The key is yours.** It names the value across the whole application, so make
  it as specific as the fetch: `authorKey.With(id)`, not `authorKey`. A key is its
  name and its type, so two keys of one name and different types hold two values
  rather than one read as the wrong type (before v0.50.0, that was
  `collage.ErrCachedTypeMismatch`).
- **One fetch per key at a time.** Renders that ask for a key while it is being
  fetched wait for that fetch rather than starting their own.
- **Errors are not stored.** A fetch with no stale value to fall back on hands
  its error to everyone waiting; the next render tries again. (With a stale value,
  a failed fetch is a failed refresh: see below.) Since v0.34.0 two failures are not handed on: a fetch that failed
  because the request that started it went away is fetched again by a waiter
  still being served, and a fetch that panicked frees its key and answers its
  waiters with `collage.ErrCachedFetchPanicked` — before, the key stayed taken
  until a restart.
- **Invalidation keeps the old value, to serve until a new one arrives.**
  `InvalidateTags`, and a value's `ttl` running out, mark it stale rather than
  removing it.
  The next render is served the stale value at once, and one fetch in the
  background replaces it, with a fixed 30-second timeout and detached from the
  request. A request after that fetch has finished sees the change. A fetch that
  was still running when its tags were invalidated hands its result to whoever
  was waiting and is kept, but as stale; one overtaken by `PurgeTags` is not kept
  at all.
- **A failed refresh keeps serving stale.** If the background fetch fails, the
  stale value stays and keeps being served, however long the upstream is down.
  The key cools down: no new fetch starts for it for a second, then two, four,
  and so on up to a minute after each further failure, so a down upstream is not
  asked once per request. A successful fetch ends the cooldown, and so does an
  invalidation of the key: a webhook is news that the upstream is answering.
- **A render built from stale data is not cached.** Plugins see it as degraded
  (`AfterRenderEvent.Degraded`), and it is stored nowhere, so the page is rendered
  again once the value is fresh. Where the page cache holds a stale copy of the
  page — right after an invalidation, say — that old page is served instead,
  marked `STALE` (see [Stale pages](#stale-pages-and-x-collage-cache)); where it
  holds none, the degraded render is served, with `Cache-Control: no-store`.
- **Stale serves are reported.** Each stale value served is reported as the
  `CacheStale` metrics event, cooldown or not. The page cache reports the same
  event, with its own key, for each stale page it serves. Each failed refresh is logged at
  warn with its error, and the first success after failures at info, with how
  many there were.
- **Not found is the exception.** A fetch whose error wraps `collage.ErrNotFound`
  drops the value, so a record deleted upstream disappears from the site and its
  page answers 404.
- **A static export never uses stale data.** `collage.Build` waits for the
  upstream, and fails if the upstream fails. So does `App.RenderFragment`, which
  a plugin calls to push a re-rendered fragment after an invalidation: what it
  pushes is not asked for again, so it must be the new value.
- **`ttl` is independent of the page's.** It bounds how long a value is fresh when
  nothing invalidates it; once it passes, the value is stale, not gone. Zero keeps
  it fresh until something invalidates it. A page that re-renders every minute can
  still reuse an author fetched an hour ago, which is the point.
- **In memory, bounded.** Values are kept in the process, up to
  `Cache.MaxEntries`, the least recently used going first — even when pages are
  cached on disk. Each instance of a multi-instance deployment keeps its own.

### Across a restart

With `Cache.Type` `"disk"`, values are also written to `<Cache.Dir>/data/` as JSON,
by a background writer that `App.Shutdown` waits for, and read back at startup. A
restart comes up warm: it serves the persisted values at once, without waiting on
the upstream, and refreshes each in the background on its first use, since
nothing says what changed while the process was down.

- **Not namespaced by build.** Unlike the pages, the directory is shared across
  builds, so a new deploy with its upstream down renders from the last one's
  data. Only the store's own files in it are ever removed.
- **A type that does not round-trip stays in memory.** A type JSON would not bring
  back as it was — one with an unexported field, a field tagged `json:"-"`, or an
  interface anywhere in it — is kept in memory only, unless that part of it
  implements both `json.Marshaler` and `json.Unmarshaler`, as `time.Time` does. A
  value whose type changed shape is not read back.
- **Not encrypted.** The files are plain JSON, so personal data should not go
  through `Cached` with a disk cache.
- **Instances sharing a `Cache.Dir` share the data directory**, and each prunes it
  to its own `MaxEntries` at startup.

### Where it keeps nothing

With `Cache.Enabled` false, in development, for a request whose middleware called
`collage.SkipCache` (a [preview](/docs/previews)), and inside an action's own
handler, `Cached` stores nothing and behaves like `collage.Once`: fragments of the
same render share one fetch, and the next render fetches again, with nothing
stale to fall back on. A preview therefore sees fresh data as well as a fresh
page.

### Once, Cached and the page cache

| | Shared between | Lives for |
| --- | --- | --- |
| `collage.Once` | the fragments of one render | that render |
| `collage.Cached` | every render in the process | fresh for its `ttl` or until its tags are invalidated, then served stale until replaced |
| page cache | every request for one URL | fresh for the page's strategy or until its tags are invalidated, then served stale only when its re-render fails or is degraded |

Use `Once` for what one page fetches twice, `Cached` for what many pages fetch,
and the page cache for the page itself. They combine: a cached page is not
rendered, so none of its fetches run at all. See
[Data handlers](/docs/data-handlers) for `Once`.

The page cache is kept per host; `collage.Cached` is not. Its store is one per
process, keyed only by the key you give it, so on a site serving several customers
(elagoht/tenant, say) `Cached(rc, postsKey, ...)` fetches acme's posts once and hands
them to globex too. Put the customer in the key, and in the tags, so invalidating
one customer's data leaves the others' alone:

```go
var postsKey = collage.NewKey[[]Post]("posts")

func posts(ctx context.Context, rc *collage.RenderContext) ([]Post, []string, error) {
	id, _ := tenant.ID(rc)
	list, err := collage.Cached(rc, postsKey.With(id), time.Hour, []string{"posts:" + id},
		func(ctx context.Context) ([]Post, error) { return db.Posts(ctx, id) })
	return list, nil, err
}
```

A page refreshed part by part through its
[fragment paths](/docs/forms-and-actions#a-fragment-at-its-own-url) is several
renders, one per fragment path, and `Once` shares nothing between them. Fragments
refreshed separately that read the same data share a fetch only through `Cached`.
