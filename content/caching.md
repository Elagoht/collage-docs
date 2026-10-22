---
description: How collage caches rendered pages and the data they are made from, and how it knows what to throw away.
reference: CacheConfig, Cached, App.PurgeTags, CacheStale, CacheWarm, Once, Key, TaggedCache, StaleCache, CacheStatusHeader, SkipCache, Vary, PageBuilder.Static, PageBuilder.Incremental, PageBuilder.Dynamic, FragmentBuilder.Static, FragmentBuilder.Shared, StrategyAuto, PathTag, SharedCache, Reached, WriteOutcome, WriteFresh, WriteStale, WriteSkipped
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
		WarmLimit:  0,         // pages re-rendered after an invalidation; 0 means 1000, negative, off
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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 312" role="img" aria-labelledby="cache-strategy-title">
<title id="cache-strategy-title">When a page is registered, a declared strategy is kept: Dynamic renders on each request and is never stored, Static is kept until a tag is invalidated, Incremental until its ttl has passed. A page that declares none is dynamic when a data handler or slot resolver appears anywhere in what it renders, a Static fragment's own handler aside, and static otherwise.</title>
<defs><marker id="cache-strategy-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="100" y="8" width="120" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">page registered</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-frame" x="70" y="50" width="180" height="26" rx="6" /><text x="160" y="67" text-anchor="middle">strategy declared?</text>
<path class="d-flow" d="M130 76C130 92 42 90 42 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M150 76C150 92 118 90 118 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M170 76C170 92 194 90 194 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M190 76C190 92 270 90 270 106" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-c" x="6" y="110" width="72" height="34" rx="5" /><text class="d-on-c" x="42" y="131" text-anchor="middle">Dynamic()</text>
<rect class="d-a" x="82" y="110" width="72" height="34" rx="5" /><text class="d-on-a" x="118" y="131" text-anchor="middle">Static()</text>
<rect class="d-b" x="158" y="110" width="72" height="34" rx="5" /><text class="d-on-b d-small" x="194" y="124" text-anchor="middle">Incremental</text><text class="d-on-b d-small" x="194" y="137" text-anchor="middle">(ttl)</text>
<rect class="d-frame" x="234" y="110" width="72" height="34" rx="5" /><text x="270" y="124" text-anchor="middle">none</text><text class="d-small" x="270" y="137" text-anchor="middle">StrategyAuto</text>
<text class="d-small" x="42" y="160" text-anchor="middle">each request</text><text class="d-small" x="42" y="172" text-anchor="middle">never stored</text>
<text class="d-small" x="118" y="160" text-anchor="middle">kept until</text><text class="d-small" x="118" y="172" text-anchor="middle">invalidated</text>
<text class="d-small" x="194" y="160" text-anchor="middle">kept until</text><text class="d-small" x="194" y="172" text-anchor="middle">ttl passes</text>
<line class="d-flow" x1="270" y1="144" x2="270" y2="192" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-frame" x="6" y="196" width="300" height="52" rx="6" /><text class="d-small" x="156" y="212" text-anchor="middle">a data handler or a slot resolver</text><text class="d-small" x="156" y="225" text-anchor="middle">anywhere in what the page renders?</text><text class="d-small" x="156" y="238" text-anchor="middle">(a Static() fragment's own handler: no)</text>
<line class="d-flow" x1="90" y1="248" x2="90" y2="272" marker-end="url(#cache-strategy-arrow)" /><text class="d-small" x="98" y="264">yes</text>
<line class="d-flow" x1="222" y1="248" x2="222" y2="272" marker-end="url(#cache-strategy-arrow)" /><text class="d-small" x="230" y="264">no</text>
<rect class="d-c" x="30" y="276" width="120" height="28" rx="5" /><text class="d-on-c" x="90" y="294" text-anchor="middle">dynamic</text>
<rect class="d-a" x="162" y="276" width="120" height="28" rx="5" /><text class="d-on-a" x="222" y="294" text-anchor="middle">static</text>
</svg>
<figcaption>A strategy a page declares is the one it keeps. A page that declares none is resolved when it is registered: anything that fetches per render makes it dynamic, and a page of templates and fixed values is static. The details are <a href="#a-page-that-declares-none">below</a>. With <code>Cache.Enabled</code> false nothing is stored, whatever the strategy.</figcaption>
</figure>

The last column has exceptions. In each of these the response goes out with its
own `Cache-Control`, whatever the strategy:

- **A [guarded page](/docs/pages-and-layouts#private-pages-guards)** goes out
  `private, no-cache`. Its readers share this cache, because the guard runs before
  the cache is read, but a CDN or proxy in front of the server runs no guard. The
  reader's own browser may keep it and revalidate with its ETag, and a
  revalidation is a request the guard sees.
- **A page carrying a form's `{{csrfToken}}`** goes out `private, no-store`. It is
  cached like any other, but each reader is sent their own token — see
  [Forms and actions](/docs/forms-and-actions#pages-with-forms-are-still-cached).
- **A degraded render** (see [below](#what-is-cached-and-when)) goes out
  `no-store`, since v0.34.0. It is not kept by this server, so it must not be kept
  by a CDN in front of it for the page's TTL either. So does a stale page served
  in place of a failed render (since v0.63.0).
- **In development** every page and document goes out `no-store`, since v0.34.0,
  so the browser does not keep a page from before your edit.

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
(see [Stale pages](#stale-pages-and-x-collage-cache)).

That is the right strategy for content that changes when someone publishes, and
only then. Its `Cache-Control` tells browsers and CDNs to keep it but ask every
time, and the [ETag](#etags-and-304) makes asking cheap.

`Incremental(ttl)` is for content that changes on a clock, or that comes from
somewhere that cannot tell you when it changed. A TTL of zero is a registration
error (`collage.ErrInvalidTTL`), and a negative one is too — not a page that
silently never expires.

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
to think about is slow, not stale.

When a handler's output is the same for everyone — a post read from a file — its
page says `Static()` or `Incremental(ttl)` itself.

What differs from page to page is not what makes a page dynamic. The path and its
parameters are part of the cache key, and each URL is its own file in an export:
`/recipes/pancakes` and `/recipes/omelette` are kept apart already. What makes a
page dynamic is output that differs between two requests to *one* URL — a user
from a cookie, the time.

A fragment whose handler reads only the path's parameters and the locale can say
so itself, with `Static()` on the fragment, and then it does not make a page
dynamic. Since v0.17.0.

```go
more := collage.NewFragment("more-recipes", "fragments/more-recipes.html").
	WithData(collage.LoadTagged(loadMore)). // every recipe but rc.Param("slug")
	Static().
	Build()
```

It is for a fragment many pages share — a list of recent posts, navigation built
from content — so that each page using it need not repeat the promise. It is the
same promise `Static()` makes for a page, and a handler that breaks it serves one
reader's render to the next.

It covers the fragment's own handler and nothing else:

- another fragment's handler still makes the page dynamic;
- so does a slot resolver, even on a static fragment, since the fragments it
  returns are not known until a render asks for them.

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
stale entry the request found.

An [action's](/docs/forms-and-actions) response is never cached, whatever the page
it rendered was declared as: it was produced from one submission and belongs to
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
for the same reason the cookie is.

Read a value your middleware declared with `collage.Vary` back with
`collage.Varied(ctx, header)`, with the context a data handler, a page's
`StaticParams`, a sitemap's `LastMod` or a document handler receives (since
v0.42.0; before v0.60.0 it took the `RenderContext`, and `VariedContext` took a
context); it is safe there because it is in the key, so each value has its own
cached page.

A value that is not a cache dimension does not belong in a shared render — a page
that needs the reader's own request is `Dynamic()`, and a dynamic page sees all of
it. In development, a hidden context value a handler actually reads is logged, so
the mistake is visible. Documents follow the same rule (since v0.41.1; before it, a
cacheable document's render still saw the first reader's context values).

### Origins

`collage.BaseURL(rc)` is safe in a shared render (since v0.42.0). The host is in the
cache key, so every reader served a copy asked for the same host, and the origin
the render built absolute URLs against is theirs too.

A static build has no request and gets `Config.BaseURL`. A site with no
`OriginResolver` plugin sees no difference; see
[Which origin a host has](/docs/writing-plugins#which-origin-a-host-has-originresolver).
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
it was dropped). It is no longer served as fresh:

- it is rendered again in the background right away (since v0.64.0; see
  [below](#warming-after-an-invalidation)), or by the next request if that has not
  happened yet, and a good render is stored in its place;
- but the old page is kept, to answer with when that render fails or comes out
  degraded — see [Stale pages](#stale-pages-and-x-collage-cache).

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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 332" role="img" aria-labelledby="cache-tags-title">
<title id="cache-tags-title">One tag, author:ada, reaches both the page cache and the data cache. InvalidateTags marks what it reaches stale and keeps it; PurgeTags removes it outright, from disk too. Either way the pages reached are warmed in the background. When a render then fails, an invalidated page is served stale, and a purged one fails.</title>
<defs><marker id="cache-tags-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="cache-tags-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-chip-hit" x="110" y="8" width="100" height="24" rx="12" /><text class="d-hit-text" x="160" y="24" text-anchor="middle">author:ada</text>
<path class="d-hitline" d="M140 32 84 58" marker-end="url(#cache-tags-hit)" />
<path class="d-hitline" d="M180 32 236 58" marker-end="url(#cache-tags-hit)" />
<rect class="d-a" x="10" y="62" width="140" height="40" rx="5" /><text class="d-on-a" x="80" y="78" text-anchor="middle">page cache</text><text class="d-on-a d-small" x="80" y="93" text-anchor="middle">/authors/ada, posts</text>
<rect class="d-b" x="170" y="62" width="140" height="40" rx="5" /><text class="d-on-b" x="240" y="78" text-anchor="middle">data cache</text><text class="d-on-b d-small" x="240" y="93" text-anchor="middle">Cached author:ada</text>
<text class="d-small" x="160" y="122" text-anchor="middle">one call reaches both</text>
<rect class="d-cmd" x="10" y="132" width="140" height="24" rx="12" /><text class="d-cmd-text" x="80" y="148" text-anchor="middle">InvalidateTags</text>
<rect class="d-cmd" x="170" y="132" width="140" height="24" rx="12" /><text class="d-cmd-text" x="240" y="148" text-anchor="middle">PurgeTags</text>
<line class="d-flow" x1="80" y1="156" x2="80" y2="172" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="156" x2="240" y2="172" marker-end="url(#cache-tags-arrow)" />
<rect class="d-frame" x="10" y="176" width="140" height="40" rx="6" /><text class="d-small" x="80" y="193" text-anchor="middle">marked stale, kept</text><text class="d-small" x="80" y="206" text-anchor="middle">as a fallback</text>
<g class="d-dropped"><rect class="d-frame" x="170" y="176" width="140" height="40" rx="6" /><text class="d-small" x="240" y="193" text-anchor="middle">removed outright</text><text class="d-small" x="240" y="206" text-anchor="middle">from disk too</text></g>
<line class="d-flow" x1="80" y1="216" x2="80" y2="232" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="216" x2="240" y2="232" marker-end="url(#cache-tags-arrow)" />
<rect class="d-soft-a" x="10" y="236" width="140" height="28" rx="6" /><text class="d-small" x="80" y="254" text-anchor="middle">warmed in background</text>
<rect class="d-soft-a" x="170" y="236" width="140" height="28" rx="6" /><text class="d-small" x="240" y="254" text-anchor="middle">warmed in background</text>
<line class="d-flow" x1="80" y1="264" x2="80" y2="280" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="264" x2="240" y2="280" marker-end="url(#cache-tags-arrow)" />
<rect class="d-frame" x="10" y="284" width="140" height="40" rx="6" /><text class="d-small" x="80" y="301" text-anchor="middle">a render fails:</text><text class="d-small" x="80" y="314" text-anchor="middle">old copy, STALE</text>
<rect class="d-frame" x="170" y="284" width="140" height="40" rx="6" /><text class="d-small" x="240" y="301" text-anchor="middle">a render fails:</text><text class="d-small" x="240" y="314" text-anchor="middle">the page fails</text>
</svg>
<figcaption>Pages and <code>collage.Cached</code> values carry the same tags, so one call reaches both. The difference between the two calls is what is left behind: <code>InvalidateTags</code> keeps the old page and value to fall back on, <code>PurgeTags</code> keeps nothing. Both warm the pages they reached. A store of your own that is not a <code>collage.StaleCache</code> drops its entries on <code>InvalidateTags</code> too.</figcaption>
</figure>

`InvalidateTagsN` does the same and tells you how many cache keys the tags
reached:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

The count is the number of keys collage's own tag index resolved and invalidated,
not an exact figure for live pages reached, and it errs both ways:

- a key whose entry had already expired is still counted;
- an entry the store reached through its own tag index — past the cap
  [below](#the-tag-index-is-per-process-and-bounded), or written by another
  instance — is invalidated without being counted. A
  [`collage.SharedCache`](#a-cache-a-fleet-shares) is the exception: the pages its
  own index reached, whichever instance wrote them, are counted (since v0.66.0).

It is good for a log line or a metric, not for logic.

If the cache fails to invalidate some keys, the rest are still invalidated and the
failures come back joined in the error. A partial invalidation that reported success is how
stale pages survive a deploy.

The usual place to call it is wherever your content changes: a CMS webhook, an
admin form (a webhook, end to end, is [below](#example-author-pages-on-a-blog)).

An action can do it declaratively, with
[`InvalidateTags` on its result](/docs/forms-and-actions#invalidating-what-an-action-changed),
which runs before the response is written — so a reader redirected to the page
they just changed is not served the old page as though it were fresh: since
v0.64.0 the redirect's own page is rendered again before the redirect is written
(see [Warming](#warming-after-an-invalidation)), and the old one is served only if
that render fails or is slow: see
[the note there](/docs/forms-and-actions#invalidating-what-an-action-changed).

Call `app.InvalidateTags` yourself from the handler and the invalidation's pages
are warmed in the background only: the redirect's page is not rendered before the
redirect is written.

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
and a search engine be told has changed. A page that was never cached is not in it.

See [Writing a plugin](/docs/writing-plugins#cacheinvalidatehook);
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) and
[elagoht/indexnow](/docs/plugins#elagohtindexnow) are built on it.

### The tag index is per process, and bounded

collage's record of which keys belong to which tag lives in memory, in the process
that wrote them. Two things follow.

**Behind a load balancer, each instance knows only its own keys.** With the
built-in caches that is fine, because each instance also has its own cache.

With a shared store of your own — `Cache.Store`, a Redis say — implement
`collage.TaggedCache` too, so the store indexes tags itself; collage then asks it
to invalidate by tag as well, and it can reach what another instance wrote. The
built-in disk cache does this, which is why its tags still work after a restart.

A store that is a [`collage.SharedCache`](#a-cache-a-fleet-shares) (since v0.66.0)
closes the gap completely: its own tag index names every page and data value any
instance wrote, so the count, `CacheInvalidateEvent.Entries` and warming cover the
fleet's pages, not only this process's.

**The record per tag is capped by `Cache.MaxEntries`** (default 10000; before
v0.60.0 a field of its own, `MaxKeysPerTag`). The query string is part of the key,
so a client can mint any number of keys for one page; without a cap the index would
grow for ever. When a tag reaches the cap, the oldest key recorded under it is
forgotten by that index. What that costs depends on the store:

- The built-in memory and disk caches both implement `collage.TaggedCache`,
  keeping their own record of tags per entry, so with them invalidating the tag
  still reaches every entry — only the count `InvalidateTagsN` reports falls short.
- A store of your own that does not implement `TaggedCache` has only collage's
  index to go by: a forgotten page stays in it until it expires, and invalidating
  the tag no longer reaches it.

Set `MaxEntries` above the number of cached URLs one tag can really cover; a
negative `MaxEntries` keeps this cap at 10000.

**With a built-in cache the index is also capped as a whole** (since v0.40.0), at
`MaxEntries` keys. Every cached path is a tag of its own, so a per-tag cap alone
let the index reach that cap times the number of pages — 10000 keys for each page
of a large site, flooded with invented queries, while the cache itself held
10000 entries in all.

No more keys than `MaxEntries` can still be live, so the oldest written go first; a
negative `MaxEntries` leaves this off. A store of your own keeps the per-tag cap
only, since collage cannot know its size.

## Warming after an invalidation

An invalidated page is stale until something renders it again, and the first
visitor would pay for that render. So since v0.64.0, after every `InvalidateTags`
and `PurgeTags` — yours, a plugin's, an action's — the pages and documents the
invalidation reached are rendered again in the background, with fresh data, and
stored: the next visitor is handed a fresh `HIT`.

Four workers do it, and the first key alone goes first, as a probe; the workers
start only after the first successful warm, so skips keep the run one key at a time
until one succeeds.

- **Only what was served is warmed.** An entry written to the cache records how
  to ask for it again — its path, query, locale and `Vary` header values. Every
  variant a visitor was actually served is warmed; one nobody asked for is not.
  The recipes live in memory only: after a restart there are none until pages
  are written again, and the first invalidation then warms nothing.
- **Only what is still read is warmed again.** A page is warmed when a visitor
  has had it since its last warm: served it fresh or stale, or rendered it. One
  warmed and then not read is passed over by the next invalidation, and renders
  on its next request as it would without warming; that visitor makes it worth
  warming again. A variant somebody asked for once is not kept warm for good.
- **What the cache let go of stays gone.** A page an `InvalidateTags` reached
  whose entry the cache no longer holds, not even stale — evicted at
  `MaxEntries` or `MaxBytes`, or dropped — is not brought back. `PurgeTags`
  removes its entries itself, and warms them all the same. (A store of your own
  that is not a `StaleCache` drops every entry it invalidates, so this check does
  not apply to it.)
- **`Cache.WarmLimit` caps one invalidation.** Zero means 1000, the most
  recently written pages first; a negative value turns warming off. Pages past
  the cap render on their first request, as before.
- **A run stops at the first real render failure**, with one warning
  (`collage: warming stopped, a page failed to render`), so a down upstream is
  not asked again for every remaining page; those are left stale, and
  [stale serving](#stale-pages-and-x-collage-cache) applies to them as usual.
  Skips do not stop it: a page that answers 404 (its recipe is dropped), a
  guarded page (its recipe is dropped, and its first reader renders it), one
  already fresh, and the two above.
- **A warm another invalidation overlapped is warmed again.** A render that any
  other `InvalidateTags` or `PurgeTags` overlapped is not stored, as no render is.
  The warmer queues that page again once no invalidation is under way, at most
  once for each invalidation that overlapped it, so a storm of invalidations
  cannot keep one page rendering.
- **A warm render is not a request.** It skips `App.Use` and plugin middleware,
  request hooks, personalisation, error hooks and per-request logs.
  `BeforeRender`, `AfterRender` and `DocumentRendered` still run, since they
  shape the stored body. It runs in its own flight, so no visitor ever waits on a
  warm render, and a page may render twice while it is being warmed.
- **Off** in development, in a static build and with the cache off. `Shutdown`
  stops warming; renders already under way finish under the shutdown context.

What follows from those rules:

- **Never declare a credential header as a `Vary` header.** A warm request
  carries no visitor's `Cookie` or `Authorization`; it carries the values of the
  headers the cache key varies on. Declare `Cookie` or `Authorization` there and
  its value is kept in the recipe, in memory, and sent again with every warm
  render of that page. Vary on what a middleware resolved from it (a plan, a
  role) instead, or make the page `Dynamic`.
- **Middleware does not gate a warm render.** A middleware that refuses a
  request — a tenant plugin answering 404 for a host whose tenant was deleted, an
  auth check — never sees a warm. After a purge, warming can therefore write an
  entry no visitor reaches any more. It is harmless, as nobody is served it, and
  it is evicted like any other; purge such a host's pages after the middleware
  stops answering for it, or let them age out.
- **A page left out of a run is untracked until written again.** An invalidation
  forgets the keys it reached. A key a run did not warm — over `WarmLimit`, left
  by a run that stopped on a failure or at `Shutdown`, or passed over — is
  reached by no later invalidation until a visitor's request renders and stores
  it again. Its stale entry is served (or rendered again) as usual meanwhile.
- **Junk query variants are kept out by `WithCacheParams`.** On a page that
  declares none, every distinct query is a key, and every key a visitor was
  served is warmed while it is still read. A page whose query can carry anything
  (tracking parameters, cache-busters) should
  [name the parameters it reads](#query-parameters-in-the-key), so a junk query is
  the same key as the clean URL.

The warmed keys are tracked again, so a later invalidation's `OnCacheInvalidate`
names them again. A run logs one line at info (`collage: warmed pages`, with the
counts warmed, skipped and failed), and each warmed key is a `CacheWarm` metrics
event.

A warm request still counts in `CacheHit`, `CacheMiss` and `CacheSet`, and its
render in `RenderDuration` and the fragment metrics, so expect a burst of misses,
sets and renders after an invalidation. It also costs CPU right after one: a site
invalidating thousands of pages at once re-renders them, four at a time. Lower
`WarmLimit` if that matters.

A test that asserts `STALE` right after an invalidation races the warmer. Set
`WarmLimit: -1` in such a test. One that asserts the `HIT` after it waits for
the warmer first, with `collagetest`'s `site.WaitWarmed()` (see
[Testing](/docs/testing#invalidation-and-warming)).

An action that invalidates and redirects warms the redirect's page first, held
to 5 seconds and to its response's write deadline — see
[Forms and actions](/docs/forms-and-actions#invalidating-what-an-action-changed).

## Stale pages and `X-Collage-Cache`

Since v0.63.0 a page or document that is stale — marked so by `InvalidateTags`, or
past its TTL — is not thrown away by the built-in stores. (After an invalidation
it is normally [warmed](#warming-after-an-invalidation) before anyone asks, so
what follows is what a visitor meets when the warm render failed, was skipped or
has not run yet.)

A request for it renders again, through the same shared render as any miss
([below](#concurrent-misses-render-once)), and what happens next depends on that
render:

- **A good render** is stored and served in place of the stale page. The response
  is a `MISS`.
- **A render that fails, or comes out degraded**, is not stored, and the stale
  page is served instead, with `Cache-Control: no-store` so that no browser or
  CDN keeps it, and `X-Collage-Cache: STALE`. The reader sees the old page, not a
  500. Reading a stale `collage.Cached` value makes a render degraded, and,
  without warming, that is the usual cause right after an invalidation: the
  page's data is stale until the background fetch replaces it.
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
render, however many requests were waiting on it:

- logged once at warn (`collage: serving a stale page, its render failed`, with
  the path and the error) — not also as an error-level `collage: request failed`,
  since no error page was served;
- handed to plugins' `OnError` hook once, with a `Status` of 0, since no error
  response was sent.

A render degraded only because it read stale `collage.Cached` data is no outage —
its data is refreshing — so its stale serve is logged at debug
(`collage: serving a stale page while its data refreshes`), and nothing reaches
`OnError`. Documents are reported the same way.

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
it.

What keeps an unbounded number of query variants from becoming unbounded memory
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
kept in `Dir/data/` too: see [Across a restart](#across-a-restart).

Since v0.11.0 a directory that cannot be created — a read-only filesystem, a
container with nowhere to write — is not a reason not to start: collage logs a
warning and caches in memory instead.

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
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit). A store of
your own bounds itself and ignores the field.

Naming the query parameters a page reads with
[`WithCacheParams`](#query-parameters-in-the-key) is the other half: a page that
ignores the query stores one copy, not thousands.

### Memory in front of the disk

Since v0.63.0 a `"disk"` cache is two tiers: a [memory cache](#memory-or-disk) in
front of the directory. Reading a page from a file is an open, a read and a decode
on every request, so a hit is served from memory instead.

- A write goes to disk and then to memory.
- A read asks memory first and copies what it finds on disk into memory as it goes.
- `MaxEntries` bounds both tiers. `MaxBytes` bounds the disk; the memory tier holds
  256 MiB, or `MaxBytes` when that is smaller.

That is memory of your process on top of the disk, so in a small container set
`GOMEMLIMIT` as for a memory cache, and lower `MaxBytes` or `MaxEntries` if the tier
must hold less (see
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
and keep it.

Give each instance its own directory. A fleet that must see one another's
invalidations needs a shared store in `Cache.Store` instead of `Type: "disk"`: one
that implements [`collage.SharedCache`](#a-cache-a-fleet-shares), such as
collage-redis, keeps pages and data consistent across the fleet.

### The namespace

A disk cache outlives the process that filled it, and that is a hazard as well as
the point. A new binary with a changed template must not serve HTML the old binary
rendered.

So entries live in a subdirectory named for the build. Leave `Cache.Version`
empty and it is a hash of the running executable: it changes exactly when your
code or templates compiled into it change, and two runs of the same build — or
every machine in a fleet running it — derive the same directory.

That keeps one build from reading another's pages; it does not make a fleet share
one cache (see [above](#memory-in-front-of-the-disk)). Set `Version` yourself only
when something outside the binary decides what the output looks like, such as a
content revision.

The forgery-protection key is not part of the namespace, so a disk cache survives
a restart whether or not `Security.CSRFKey` is set. The key still matters to one
kind of page: a cached page with a form in it carries a marker derived from the key
where each reader's token goes (see
[Forms and actions](/docs/forms-and-actions#pages-with-forms-are-still-cached)),
and a page stored under one key cannot be served under another.

So a stored page carrying another key's marker — rendered before the key changed,
or by a process that generated its own — is treated as a miss: it is dropped and
rendered again. Pages without a form are unaffected by the key entirely.

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
  as before. The framework looks a page up in a `StaleCache` with `GetStale`
  alone, so one lookup is one hit or one miss in your counters.
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

## A cache a fleet shares

Behind a load balancer every instance has its own page cache, its own
`collage.Cached` values and its own [tag index](#the-tag-index-is-per-process-and-bounded).
A CMS webhook reaches one of them, and its `InvalidateTags` reaches only that one:
the others go on serving the old page as a fresh `HIT`, and rendering from the old
value, until it expires — for a `Static` page, until the instance restarts or evicts
it. A `"disk"` cache whose `Dir` they share does not change that (see
[above](#memory-in-front-of-the-disk)).

Since v0.66.0 a store that implements `collage.SharedCache` is shared by the whole
fleet. Pages and `collage.Cached` data both live in it, and nothing is kept per
instance in front of it, so an `InvalidateTags` or `PurgeTags` on any instance
reaches the next read on every instance.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 276" role="img" aria-labelledby="cache-fleet-title">
<title id="cache-fleet-title">Three instances behind a load balancer share one store. A webhook reaches instance A, which makes one DropTags call on the shared store; the store holds the pages, the Cached data, the tag index and the epoch. Instances B and C read and write the same store, so their next read sees the invalidation. A warms the pages the invalidation reached, within its own WarmLimit.</title>
<defs><marker id="cache-fleet-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="cache-fleet-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-chip-hit" x="10" y="8" width="90" height="24" rx="12" /><text class="d-hit-text" x="55" y="24" text-anchor="middle">webhook</text>
<line class="d-hitline" x1="55" y1="32" x2="55" y2="50" marker-end="url(#cache-fleet-hit)" />
<rect class="d-a" x="10" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="55" y="75" text-anchor="middle">instance A</text>
<rect class="d-a" x="115" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="160" y="75" text-anchor="middle">instance B</text>
<rect class="d-a" x="220" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="265" y="75" text-anchor="middle">instance C</text>
<line class="d-flow" x1="55" y1="88" x2="55" y2="100" marker-end="url(#cache-fleet-arrow)" />
<rect class="d-cmd" x="10" y="104" width="90" height="24" rx="12" /><text class="d-cmd-text" x="55" y="120" text-anchor="middle">DropTags</text>
<line class="d-flow" x1="55" y1="128" x2="55" y2="150" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="160" y1="88" x2="160" y2="150" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="265" y1="88" x2="265" y2="150" marker-end="url(#cache-fleet-arrow)" />
<text class="d-small" x="212" y="122" text-anchor="middle">reads, writes</text>
<rect class="d-b" x="10" y="154" width="300" height="48" rx="5" /><text class="d-on-b" x="160" y="173" text-anchor="middle">shared store (collage-redis)</text><text class="d-on-b d-small" x="160" y="189" text-anchor="middle">pages, Cached data, tag index, epoch</text>
<line class="d-flow" x1="82" y1="202" x2="82" y2="222" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="237" y1="202" x2="237" y2="222" marker-end="url(#cache-fleet-arrow)" />
<rect class="d-soft-a" x="10" y="226" width="145" height="42" rx="6" /><text class="d-small" x="82" y="243" text-anchor="middle">A warms the pages</text><text class="d-small" x="82" y="256" text-anchor="middle">within its WarmLimit</text>
<rect class="d-frame" x="165" y="226" width="145" height="42" rx="6" /><text class="d-small" x="237" y="243" text-anchor="middle">B, C: the next read</text><text class="d-small" x="237" y="256" text-anchor="middle">sees the invalidation</text>
</svg>
<figcaption>One store for the fleet. The instance a webhook reaches makes one <code>DropTags</code> call, which marks stale (or, for a purge, removes) every page and value under the tags, whichever instance wrote them, and names the pages it reached for the count, the plugin event and warming. Every other instance reads the same store, so its next read sees the change. Only the instance that invalidated warms, within its own <code>WarmLimit</code>.</figcaption>
</figure>

### Setting one up with collage-redis

[collage-redis](/docs/plugins#elagohtredis) is a `SharedCache` on Redis. It is a
library, not a plugin: there is nothing to register, and the store goes in
`Cache.Store`, where `Type` is then ignored.

```sh
go get github.com/Elagoht/collage-redis
```

```go
import "github.com/Elagoht/collage-redis" // package redis

store, err := redis.New(redis.Config{URL: os.Getenv("REDIS_URL")})
if err != nil {
	log.Fatal(err)
}
defer store.Close()

app, err := collage.New(&collage.Config{
	Cache: collage.CacheConfig{Enabled: true, Store: store},
})
```

`New` never dials Redis. It fails only for a bad config or URL, so an application
starts, and serves uncached, while its Redis is down. Use `rediss://` for TLS; the
password and the database number come from the URL too. v0.1.0 needs collage
v0.66.0 or later and Redis 6 or later.

Only `URL` (or `Client`) is needed:

| Field | Default | Meaning |
| --- | --- | --- |
| `URL` | none | A `redis://` or `rediss://` URL. Required unless `Client` is set; not both. |
| `Client` | none | An existing go-redis client to reuse. `Close` leaves it open: its owner closes it. |
| `Prefix` | `collage:` | Starts every key. |
| `Version` | a hash of the running executable | Namespaces the keys after `Prefix`. Blank, whitespace only or containing `:` is rejected. |
| `MaxAge` | 7 days | The longest any key lives, and the longest an entry is fresh. |
| `Timeout` | 250ms | Bounds each Redis command. |
| `Logger` | `slog.Default()` | Receives the store's warnings. |

A client built from `URL` does not retry a failed command, because a failure is
only a miss and retrying would slow every request while Redis is down. With a
`Client` of your own, set `ContextTimeoutEnabled` on it, so `Timeout` bounds a
command already waiting on the network, and `MaxRetries` to -1. A `Client` is also
how to reach what a URL cannot describe, such as Sentinel.

Running Redis for it:

- **`maxmemory-policy noeviction`**, with memory for about `MaxAge` of writes.
  Every key expires on its own; eviction would only remove entries the store still
  counts on. A full Redis fails every write as a whole — a miss, the page served
  and not stored — while reads and invalidations still work.
- **Persistence (RDB or AOF)**, or accept a rare stale-as-fresh page: a
  `FLUSHALL`, or a restart without persistence, between an invalidation and the
  write of a render it overlapped erases what lets that write be told apart.
- **A single node or Sentinel.** Redis Cluster is not supported in v0.1.0.
- **Clocks in sync** (NTP, say): freshness is judged with each instance's clock.

### What is shared, and what is not

Shared by every instance:

- rendered pages and their tags;
- the [stale fallback](#stale-pages-and-x-collage-cache): the last good copy,
  served when a re-render fails;
- `collage.Cached` values. Every `Cached` call reads the store — one round trip
  and one JSON decode per load. A value JSON cannot bring back as it was (see
  [Across a restart](#across-a-restart)) is not cached across renders at all, and
  behaves as with `Once`, since keeping it in memory would bring back the
  per-instance copy;
- invalidations, and their reach: the count `InvalidateTagsN` returns, the
  `CacheInvalidateEvent.Entries` a CDN purge acts on, and warming cover the pages
  any instance wrote.

Kept per instance:

- the single fetch per `Cached` key, the background refresh and the failure
  cooldown: N instances missing one key at once make up to N fetches;
- warming: only the instance that ran the invalidation re-renders.

### What it guarantees

- **A render an invalidation overlapped lands stale, on every instance.** The
  store owns an epoch, a counter that only grows. A render or a data fetch reads it
  before it starts and hands it back when it writes. A write carrying a tag
  invalidated since is stored stale, so a page B rendered from data A was
  invalidating is kept as the fallback, never as a fresh page that would stay — for
  a `Static` page, for good. One carrying a tag purged since is not stored at all.
- **A late overlapped write does not replace a newer page.** When the warm after
  an invalidation has stored a fresh page, a render that began before the
  invalidation and finishes after it does not overwrite it with its stale one.
  The same holds for data values.
- **An invalidation is one call.** `InvalidateTags` and `PurgeTags` make one
  `DropTags` call each, whatever the number of tags or entries, within five
  seconds. It marks stale (or removes) every page and value under the tags and
  returns the pages it reached, as one step; when it fails, the error is
  returned and nothing is counted.
- **It reaches every instance's pages.** The pages `DropTags` returns, which any
  instance wrote, are counted, named in `CacheInvalidateEvent.Entries` — so
  [elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) purges what B served — and
  warmed, within the invalidating instance's `WarmLimit`. Past it, a page is
  rendered by whichever instance gets its next visitor, a `MISS` once; the plugin
  event still gets every entry.

### When the store fails

A store that errors, or takes longer than 500 milliseconds, is a miss: the page is
rendered and served without being stored, the value fetched and returned. It is
never a failed request, so it is not reported as one — no `"cache_write"` error log
and no `OnError` event, which a full Redis would otherwise send on every cacheable
miss.

The failures of pages and data together go into one health per App, logged at warn
at most once every 30 seconds with how many went unlogged
(`collage: shared page and data cache unavailable`), and at info once when a call
succeeds again (`… available again`). collage-redis logs the same outage on its
own side, at most once every 10 seconds, and once at info when Redis answers again.

### Things to know

- **A deploy starts cold.** collage-redis's `Version` is a hash of the executable,
  so a new build gets an empty namespace and a rolling deploy's old and new builds
  never read each other's pages. The old build's keys are not deleted; they expire
  within `MaxAge`. Set `Version` yourself to share entries across binaries.
- **`MaxAge` bounds the stale fallback.** A page or value not re-rendered within
  `MaxAge` (seven days by default) is gone, even during an upstream outage. There
  is no such bound in the built-in stores.
- **One fetch per key is per instance**, as above: N instances missing one key at
  once make up to N upstream fetches.
- **A page invalidated and not yet re-rendered stays in the store's tag index.** A
  second invalidation of its tags before that re-render counts it again, names it
  again in the plugin event, and offers it for warming again.

A store that is not a `SharedCache`, the memory cache and `Type: "disk"` behave as
before.

### Writing your own SharedCache

Any store can be one. The framework type-asserts for it, as for `TaggedCache` and
`StaleCache`, and it includes both:

```go
type SharedCache interface {
	TaggedCache
	StaleCache
	Epoch(ctx context.Context) (uint64, error)
	SetSince(ctx context.Context, key string, content []byte, ttl time.Duration,
		tags []string, meta []byte, since uint64) (etag string, outcome WriteOutcome, err error)
	DropTags(ctx context.Context, tags []string, purge bool) ([]Reached, error)
	GetData(ctx context.Context, key string) (value []byte, stale, found bool, err error)
	SetData(ctx context.Context, key string, value []byte, ttl time.Duration,
		tags []string, since uint64) (WriteOutcome, error)
	DeleteData(ctx context.Context, key string) error
}
```

- `Epoch` returns the store's counter. Each tag has a generation, the epoch at its
  last invalidation, and a purge generation likewise.
- `SetSince` and `SetData` take the epoch the render or fetch began at, `since`,
  and say what they did: `collage.WriteSkipped` when a tag was purged after
  `since`, `collage.WriteStale` when one was invalidated after it, and
  `collage.WriteFresh` otherwise.
- `DropTags` advances the epoch, records the generations, marks stale (or, with
  `purge`, removes) every page and value under the tags, and returns each page it
  reached as a `collage.Reached`: its key and the meta it was written with, which
  is how collage knows its host and path.

`collagetest.SharedCacheSuite` runs the whole contract against a store, through two
handles on one backend, as two instances hold two connections to one Redis:

```go
func TestMySharedStore(t *testing.T) {
	collagetest.SharedCacheSuite(t, func(t *testing.T) (a, b collage.SharedCache) {
		// Two handles on one empty store.
		return newHandle(t, backend), newHandle(t, backend)
	})
}
```

It runs `collagetest.CacheSuite` on the first handle — the `Cache`, `TaggedCache`
and `StaleCache` contracts, which is also what to call for a store that is not
shared — and then writes through one handle and checks through the other: the
epoch ordering, overlapped writes landing stale or skipped, what `DropTags`
reaches, data values, and concurrent use, which means something under `-race`.
collage-redis passes it. To test an application that runs as a fleet, see
[Testing](/docs/testing#a-fleet-on-one-shared-cache).

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
  context values stripped as well, bounded by the fragments' own timeouts.
  One reader pressing stop cannot hand everyone who asked at the same moment a page
  whose parts failed with "context canceled".
- **It is visible.** A request served this way is reported to your metrics twice:
  as a `CacheMiss`, when its lookup found nothing, and then as `CacheCoalesced`,
  when it was served another request's render. So the renders a key cost are its
  misses minus its coalesced events. A coalesced count that climbs steadily means a
  page is expiring faster than it can be rendered, which is what a too-short
  `Incremental` TTL looks like from outside.

## Development never reads the cache

With `Config.DevMode` on — which a scaffolded project turns on under
`collage dev` — cached pages are never served. Templates reload from disk on every
request in development, and a cached page would hide the edit you just made for as
long as its TTL.

Pages are still written and tags still tracked, so hooks and metrics behave as they
will in production, warming aside; what is gone is serving a page rendered before
the edit.

A disk cache is replaced by a memory one in development, and collage logs that it
did. No response carries `X-Collage-Cache` there. `collage.Cached` keeps nothing
across renders there either (see below).

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

Build the varying part from what the application chose — an id found in its own
database, a page number it capped — not from raw request input. The store is
bounded by `MaxEntries`, not by `MaxBytes`: it holds Go values whose size is the
application's, so a key taken from a query parameter is one entry per value a
visitor invents, and on a `"disk"` cache they outlive the process.
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
  a failed fetch is a failed refresh: see below.)
  Since v0.34.0 two failures are not handed on: a fetch that failed because the
  request that started it went away is fetched again by a waiter still being
  served, and a fetch that panicked frees its key and answers its waiters with
  `collage.ErrCachedFetchPanicked` — before, the key stayed taken until a restart.
- **Invalidation keeps the old value, to serve until a new one arrives.**
  `InvalidateTags`, and a value's `ttl` running out, mark it stale rather than
  removing it.
  When no warm render reached the page (warming off, over `WarmLimit`, the run
  stopped on a failure, or a visitor arrived before the warmer got there; see
  [Warming](#warming-after-an-invalidation)), the next render is served the stale
  value at once, and one fetch in the background replaces it, with a fixed
  30-second timeout and detached from the request. A request after that fetch has
  finished sees the change.
  A fetch that was still running when its tags were invalidated hands its result
  to whoever was waiting and is kept, but as stale; one overtaken by `PurgeTags` is
  not kept at all. A render that began after the invalidation and joined such a
  fetch gets its value stale too (since v0.66.0, for every store; before, it got
  it as fresh, and its page could be stored fresh from the data the invalidation
  replaced), so that render is served `no-store` and not stored. A warm render
  instead waits for fresh data and stores a fresh page; one that joins such a
  fetch fetches again.
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
  cached on disk. Each instance of a multi-instance deployment keeps its own,
  unless `Cache.Store` is a [`collage.SharedCache`](#a-cache-a-fleet-shares): then
  the values live in that store, beside the pages, and no instance keeps a copy.

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
| `collage.Cached` | every render in the process (in the fleet, with a [`SharedCache`](#a-cache-a-fleet-shares)) | fresh for its `ttl` or until its tags are invalidated, then served stale until replaced |
| page cache | every request for one URL | fresh for the page's strategy or until its tags are invalidated, then served stale only when its re-render fails or is degraded |

Use `Once` for what one page fetches twice, `Cached` for what many pages fetch,
and the page cache for the page itself. They combine: a cached page is not
rendered, so none of its fetches run at all. See
[Data handlers](/docs/data-handlers) for `Once`.

The page cache is kept per host; `collage.Cached` is not. Its store is one per
process (or the fleet's shared store), keyed only by the key you give it, so on a site serving several customers
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

## Example: author pages on a blog

A blog keeps its authors in a CMS. Each author has a page at `/authors/{id}`, and
every post shows a card with its author's name and bio. The site has to follow two
things that happen in the CMS:

- an author edits their profile;
- an author deletes their account.

The profile is fetched through `collage.Cached`, once for every page that shows it,
and tagged `author:<id>`. A post's author card fetches through the same key and tag.

```go
var authorKey = collage.NewKey[Author]("author")

func loadAuthor(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	id := rc.Param("id")
	author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return cms.Author(ctx, id) // wraps collage.ErrNotFound for a deleted author
		})
	return author, nil, err
}

authorPage := collage.NewPage("author").
	WithLayouts(layout).
	WithContent(collage.NewFragment("author", "pages/author.html").
		WithData(collage.LoadTagged(loadAuthor)).
		Required(). // a deleted author is a 404, not a page without its content
		Build()).
	WithPath("en", "/authors/{id}").
	Static(). // the same for every reader, until the CMS says otherwise
	Build()
```

The handler reads only the path parameter, and the profile changes only when the
CMS says so. So the page says `Static()` itself — with no call, its handler would
make it dynamic — and the webhook below is what renders it again.

The CMS calls a webhook on every change. It is an ordinary `net/http` handler,
mounted with [`app.Handle`](/docs/middleware-and-apis):

```go
type cmsEvent struct {
	Type     string `json:"type"`      // "author.updated" or "author.deleted"
	AuthorID string `json:"author_id"` // the CMS's id for the author
}

func cmsWebhook(app *collage.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// app.Handle adds no forgery check and no body limit: check the CMS's
		// signature here, and bound what you read.
		var event cmsEvent
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&event); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		tag := "author:" + event.AuthorID
		var err error
		switch event.Type {
		case "author.updated":
			err = app.InvalidateTags(r.Context(), tag)
		case "author.deleted":
			// The CMS has already deleted the author, so the author's page
			// warms as a 404, and posts lose the card, instead of showing
			// the old profile.
			err = app.PurgeTags(r.Context(), tag)
		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			log.Printf("cms webhook %s %s: %v", event.Type, event.AuthorID, err)
			http.Error(w, "cache not updated", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// After collage.New:
if err := app.Handle("/hooks/cms", cmsWebhook(app)); err != nil {
	log.Fatal(err)
}
```

**A profile edit calls `InvalidateTags`.** The tag reaches the stored author and
every cached page that showed them — the author page and each post with the card —
because `Cached` adds its tags to the page's ([above](#one-set-of-tags-for-both-caches)).

- They are marked stale, not removed, and the pages are
  [warmed](#warming-after-an-invalidation) in the background with fresh data.
- A reader who arrives before the warm is served the old page, marked `STALE`,
  while the value refreshes.
- If the CMS is down, readers keep the last good page instead of a 500.

**An account deletion calls `PurgeTags`.** Here the old profile is what must not be
shown again, even once. `InvalidateTags` would keep it: a reader who arrives before
the warmer could still be served it, from the stale page or the stale value.
`PurgeTags` keeps nothing, from disk too.

Delete the author in the CMS before the purge, not after. A purge warms the pages
it reached, as an invalidation does. With the author gone, the warm render of
`/authors/{id}` meets `collage.ErrNotFound` on a required fragment and answers 404,
so nothing is stored. With the author still there, the warm would store the
profile again.

The example assumes a memory cache. A profile is personal data, and with
`Type: "disk"` `Cached` writes its values to `Dir/data/` as plain JSON (see
[Across a restart](#across-a-restart)).
