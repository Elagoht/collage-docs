---
description: What collage is, the idea it is built on, and when it is the right tool.
reference: NewPage, NewFragment
---

# Introduction

collage is a Go framework for websites that are rendered on the server. You write
pages as a layout wrapped around small pieces called fragments; each fragment
fetches its own data and renders its own template, and collage assembles the page,
caches it, and knows exactly what to throw away when your content changes.

It is one Go program. The same binary serves your site, answers its forms and API
calls, and — when you ask it to — writes the whole site out as static files. There
is no client framework, no bundler, and no dependency beyond the standard library.

## The idea

Most pages are made of parts that do not know about each other. A blog post has
the post, an author card, a list of related posts, a navigation bar. Each part needs
different data, from a different place, and changes for different reasons.

In collage each of those parts is a **fragment**: a template, a function that fetches
what the template needs, and named **slots** where other fragments go. A **page** is
a layout fragment with a content fragment in it, and a URL.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(loadAuthor).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(loadPost).
	WithSlotFragment("author", author).
	Build()

page := collage.NewPage("post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()
```

Three things follow from building pages this way.

- **Fetching is concurrent, writing is in order.** Before a fragment renders, the
  data handlers of every fragment in its slots start at once. A page of independent
  parts makes its upstream calls together, and the HTML still comes out the same
  every time.
- **A page knows what it is made of.** Every data handler says which pieces of
  content it used — `post:hello-world`, `author:ada` — and the cached page carries
  those tags. When an author changes, one call drops every page that showed them,
  and nothing else.
- **A part can fail without taking the page with it.** A fragment can be required,
  or have a fallback, or simply render nothing. A broken sidebar is a missing
  sidebar, not a 500.

## What you get

- Pages cached per URL, with three strategies: rendered once and kept until you
  invalidate it, re-rendered after a TTL, or never cached. A page that fetches
  nothing is rendered once without being told to.
- Data cached across pages, so with the cache on, thirty posts by one author fetch the author once.
- Forms that post to their own page, with request-forgery protection built in —
  and a cached page can still carry one.
- Links built from page names, so they follow a page when its path changes, in
  every locale the site has.
- Head elements — titles, meta tags, stylesheets — declared by the fragment that
  needs them, wherever it sits in the page.
- Non-HTML routes for sitemaps, feeds and `robots.txt`, and your own `http.Handler`
  for anything else.
- A static export that turns the site into files for any static host. The pages
  you are reading were produced by it.
- A development server that rebuilds on every Go change and reloads the browser on
  every template change. A page that fails to render shows its error in the
  browser, template, line and cause first. Go that fails to compile shows it in the
  terminal while the last good build keeps serving, and a program that cannot
  start at all shows what it printed in the browser rather than a refused
  connection.

## When it is the right tool

collage fits a site whose pages are mostly read: a blog, documentation, a
marketing site, a catalogue, a news site, the public side of an application. It is
at its best when content comes from somewhere else — a CMS, a database, an API —
and pages are assembled from several such sources at once.

It is a poor fit for an application whose screens are mostly interaction: an
editor, a dashboard that updates every second, anything whose state lives in the
browser. Use a client framework for those, and collage — through
[`app.Handle`](/docs/middleware-and-apis) — for the pages around them if you like.

## What it deliberately leaves out

collage does not come with a database layer, a session store, an ORM, a
translation catalogue or a JavaScript toolchain. You are writing Go, and Go already
has good answers to all of those; a framework that picked one for you would be one
more thing to work around. What collage owns is the part that is specific to
rendering pages: composing them, caching them, and knowing when they are stale.

## Where to go next

[Installation](/docs/installation) gets a project running in a minute.
[Your first page](/docs/your-first-page) builds one from nothing, and
[Pages and layouts](/docs/pages-and-layouts) is where the concepts start.
