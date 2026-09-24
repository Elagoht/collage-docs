---
description: Fragments, the slots they expose, what happens when one fails, and slots filled from content at render time.
---

# Fragments and slots

A **fragment** is the unit a page is made of: a template, an optional data handler
that fetches what the template renders, and named **slots** where other fragments
go. A layout is a fragment. So is the content of a page, a sidebar, an author card,
a comment list. A page is a tree of them, with the layout at the root.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(collage.DataHandler(loadAuthor)).
	WithTimeout(time.Second).
	WithFallback(anonymousAuthor).
	Build()
```

## Building a fragment

`collage.NewFragment(name, templatePath)` starts a builder, and `Build()` returns
the `*collage.Fragment`.

- **The name** is how the fragment is reported: in error messages, in the
  development error panel, in render metadata and traces. Make it say which part of
  the page it is — `"post-comments"`, not `"list"`.
- **The template path** is relative to the template root, extension included:
  `"fragments/author.html"` is `templates/fragments/author.html`. A path the
  template engine did not load is `ErrTemplateNotFound` when the page is
  registered. See [Templates](/docs/templates).

Everything else is optional:

| Method | What it sets |
| --- | --- |
| `WithDataHandler(h)` | The function that fetches the template's data — see [Data handlers](/docs/data-handlers) |
| `WithSlot(name, required, allowMultiple)` | Declares a slot |
| `WithSlotFragment(slot, child)` | Binds a child fragment into a slot |
| `WithSlotResolver(slot, resolve)` | Fills a slot per render instead |
| `Required()` | This fragment's failure fails the page |
| `WithFallback(f)` | What renders when this fragment fails |
| `WithTimeout(d)` | How long its data handler may take |

Like the page builder, the fragment builder records mistakes rather than stopping
the chain, and `BuildErr()` returns them. What it recorded stays on the fragment it
built, so every one of them — a slot declared twice, a child bound into a slot that
was never declared, a negative timeout — stops any page whose tree contains the
fragment at registration, whether or not anyone called `BuildErr()`. Call it when
you want the error at the line that caused it rather than at `RegisterPage`; see
[Pages and layouts](/docs/pages-and-layouts#building-a-page).

A fragment with no data handler renders its template with no data. That is right
for markup that never changes — a footer, a static notice — and for a layout whose
only job is to arrange slots.

## Slots

A slot is a named position in a fragment's template. It is declared on the
fragment and written into the template with `{{slot "name"}}`:

```go
post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	WithSlot("author", true, false).
	WithSlot("related", false, true).
	WithSlotFragment("author", author).
	WithSlotFragment("related", relatedPosts).
	WithSlotFragment("related", popularPosts).
	Build()
```

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
<aside>{{slot "related"}}</aside>
```

`WithSlot(name, required, allowMultiple)` takes two flags.

- **`required`** — the slot must have something in it. A required slot with
  nothing bound is refused at registration (`ErrRequiredSlotUnfilled`), and checked
  again before the template runs, so it is caught even if the template never asks
  for it.
- **`allowMultiple`** — the slot may hold more than one fragment. They render in
  the order they were bound, one after another. Binding a second fragment into a
  slot that holds one records `ErrSlotOccupied`.

`WithSlotFragment` records `ErrUnknownSlot` for a slot that was not declared —
declare slots first — and `ErrNilFragment` for a nil child. Declaring one name
twice records `ErrDuplicateSlot`.

In the template, `{{slot "name"}}` renders what the slot holds as HTML, which is
not escaped again: the children escaped their own values when they rendered. A slot
that holds nothing renders nothing. A name the fragment never declared is an
**error**, not empty output — a typo in a template would otherwise be a section
that is quietly missing.

### Each fragment has its own data

A child does not see its parent's data. Inside `fragments/author.html`, `.` is what
`loadAuthor` returned, not the post. A child that needs something its parent
fetched reads it from the render's shared data, or asks for it itself with
`collage.Once` or `collage.Cached` so it is fetched only once — see
[Data handlers](/docs/data-handlers#sharing-data-between-fragments).

### Reusing fragments

A `*collage.Fragment` can be bound into many slots on many pages; `author` above
could appear on the post page and on a search result page. A fragment bound twice
renders twice, running its data handler once per binding.

A fragment must not contain itself. Binding one, directly or through a chain of
children, into its own slot is refused at registration with `ErrFragmentCycle`.

## When a fragment fails

A fragment fails when its data handler returns an error, when its template fails
to execute, when a required slot is empty, or when either panics — a panic in a
data handler or a template function is recovered into a `*collage.PanicError` and
treated as a failure, not a crashed process.

What happens next is the fragment's **failure policy**, and there are three.

```go
postContent := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(collage.DataHandler(loadPost)).
	Required().
	Build()

comments := collage.NewFragment("comments", "fragments/comments.html").
	WithDataHandler(collage.DataHandler(loadComments)).
	WithFallback(collage.NewFragment("comments-unavailable", "fragments/comments-unavailable.html").Build()).
	Build()

related := collage.NewFragment("related", "fragments/related.html").
	WithDataHandler(collage.DataHandler(loadRelated)).
	Build()
```

- **`Required()`** — the fragment's failure fails the page. Use it for what the page
  exists to show. The page's error page is served with a 500 — or, when the error
  wraps `collage.ErrNotFound`, its not-found page with a 404. The failure travels
  through any optional fragments above it: a required fragment inside an optional
  sidebar still fails the page.
- **`WithFallback(f)`** — `f` renders in its place. If the fallback fails too, the
  fragment renders nothing and the page still succeeds: a fallback exists to contain
  a failure, so its own failure is contained as well, even if something inside it
  is marked required.
- **Neither** — the fragment renders nothing, and the page is served without it.

`collage.ErrNotFound` only means "404" on a required fragment. On an optional one it
is a failure like any other, because the page is still renderable without it.

A page served with any failed fragment, fallback or not, is **degraded**. It is
sent to the reader but never cached, so the next request tries again rather than
serving the failure for a TTL. A static export does not write a degraded page
unless it is told to.

In development a failure is never quiet. A failed fragment that renders nothing
leaves an HTML comment with its name and error where its output would have been,
and the page carries a panel naming every fragment that failed, with the error —
for a template, the file and line — even when a fallback covered for it. When a
required fragment fails the whole page, the development error page names that
fragment — where the failure started, not the layout it travelled up through.
Outside development there is neither.

### A slot the template skips

A fragment's children start fetching before its template runs, so that they fetch
at the same time rather than one after another. That means a fragment in a slot
the template decides not to render — `{{if .ShowComments}}{{slot "comments"}}{{end}}`
— still had its handler started. Its context is cancelled as soon as the template
is done, and its failure is discarded: a fragment that was never rendered cannot
fail the page, even if it is `Required()`.

## Timeouts

`WithTimeout(d)` bounds the fragment's data handler. A fragment that sets none gets
`Config.Template.Timeout`, which defaults to five seconds; a negative duration
records `ErrInvalidTimeout`.

```go
recommendations := collage.NewFragment("recommendations", "fragments/recommendations.html").
	WithDataHandler(collage.DataHandler(loadRecommendations)).
	WithTimeout(300 * time.Millisecond).
	WithFallback(nothingToRecommend).
	Build()
```

A timeout pairs naturally with a fallback. A slow recommendation service costs
the page 300 milliseconds and a fallback, rather than the reader's patience.

The timeout bounds the **context** the handler receives, not the handler: a
handler that ignores `ctx` runs as long as it likes. Pass `ctx` to every call that
can wait — see [Data handlers](/docs/data-handlers#timeouts-and-the-context).

## Slots filled per render

Everything above binds children when the program starts. Some pages are arranged
by their content instead: a landing page whose sections an editor picks and orders
in a CMS, a dashboard of widgets a user chose. Binding those at startup would mean
a restart for every change, and code that has to agree with the data about what
goes where.

`WithSlotResolver` fills a slot per render. The resolver receives the render
context and returns the fragments the slot holds this time:

```go
type block struct {
	Kind    string
	Heading string
	Text    string
}

landing := collage.NewFragment("landing", "pages/landing.html").
	WithDataHandler(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		blocks, err := cms.Blocks(ctx, "landing")
		if err != nil {
			return err
		}
		rc.Set("blocks", blocks)
		return nil
	})).
	WithSlot("blocks", true, true).
	WithSlotResolver("blocks", func(rc *collage.RenderContext) ([]*collage.Fragment, error) {
		blocks, _ := collage.Get[[]block](rc, "blocks")
		fragments := make([]*collage.Fragment, 0, len(blocks))
		for i, b := range blocks {
			f, err := blockFragment(i, b)
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, f)
		}
		return fragments, nil
	}).
	Required().
	Build()
```

A resolver may return fragments that exist for the whole program, or build them on
the spot. Building them is how each block gets its own data:

```go
func blockFragment(i int, b block) (*collage.Fragment, error) {
	switch b.Kind {
	case "hero", "text":
		return collage.NewFragment(fmt.Sprintf("block-%d-%s", i, b.Kind), "blocks/"+b.Kind+".html").
			WithDataHandler(collage.DataHandler(func(context.Context, *collage.RenderContext) (block, []string, error) {
				return b, nil, nil
			})).
			Build(), nil
	}
	return nil, fmt.Errorf("landing: unknown block kind %q", b.Kind)
}
```

The rules:

- **The resolver runs after its own fragment's data handler**, so it can read what
  that handler fetched, and **before the fragments it returns start theirs** — which
  then run concurrently, like any other children.
- **What it returns is held to the slot's rules**: at most one fragment unless the
  slot allows multiple, at least one if it is required (`ErrRequiredSlotEmpty`). An
  error from the resolver, or a panic, fails its fragment, and that fragment's
  failure policy applies.
- **Declare the slot first**, with `WithSlot`. A slot is filled by a resolver or by
  `WithSlotFragment`, never both — mixing them records `ErrSlotResolved`.
- **Its fragments are only checked when they render.** Registration cannot see
  them, so a returned fragment whose template does not exist fails that render
  rather than startup — and a mistake its builder recorded is not refused by
  anything, so check `BuildErr()` on a fragment you build in a resolver.
  Templates are all loaded at startup, so the set of block kinds a resolver can
  use is still fixed by the program.

A page whose sections come from content should also report that content's tags —
here the landing handler could return `"landing"` as a tag with
`collage.DataHandler` instead of `collage.Effect` — so a cached page is dropped when
an editor reorders it. See [Caching](/docs/caching).

## The nesting limit

A fragment tree may be at most **32 levels** deep, counting the root. A page that
goes deeper fails to render with `ErrMaxDepthExceeded`, whatever its fragments'
failure policies, and the error names the chain of fragments that reached the
limit.

Real pages come nowhere near it. The limit exists for the one case registration
cannot rule out: a resolver that returns a fragment which, somewhere below,
resolves to itself again. A bound cycle is already refused at registration as
`ErrFragmentCycle`.

## Fragments with their own URL

A fragment can also be fetched without the page around it — search results
refreshed by a `fetch()`, a panel swapped in after a form posts. That is declared
on the page, with `WithFragmentPath(locale, pattern, fragment)`, and nothing is
reachable that way unless it is declared. See
[Forms and actions](/docs/forms-and-actions).
