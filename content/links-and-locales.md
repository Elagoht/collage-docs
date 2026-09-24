---
description: Serving a site in several languages, and linking between pages by name so links follow them in every locale.
reference: LocaleConfig, PageBuilder.WithPath, Vary, ErrNoPathInLocale, ErrUnknownRoute
---

# Links and locales

A collage site can serve every page in several languages, each at its own URL. It
routes locales; it does not translate. What a word should be in Turkish lives in
your content, wherever that already is — collage's part is knowing which URL
belongs to which language, and building links that know it too.

## Configuring locales

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Locale: collage.LocaleConfig{
		Default:   "en",
		Supported: []string{"en", "tr"},
	},
})
```

`Default` is the locale of a URL with no locale prefix, and defaults to `"en"`.
`Supported` lists every locale the site serves, and must include `Default`
(`collage.ErrLocaleDefaultNotSupported` otherwise); left empty, it is just
`Default`.

A page then declares its path in each locale it exists in:

```go
collage.NewPage("about").
	WithLayout(layout).
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda").
	Build()
```

The paths may differ, as here, or be the same pattern in every locale. A page with
a path in only one locale exists only in that one.

## The URL decides the locale

A path with no locale prefix is in `Default`. A path whose first segment names a
supported locale is in that locale, and the prefix is removed before the path is
matched:

| URL | Locale | Matched against |
| --- | --- | --- |
| `/about` | `en` | `/about` in the `en` paths |
| `/tr/hakkinda` | `tr` | `/hakkinda` in the `tr` paths |
| `/tr/about` | `tr` | `/about` in the `tr` paths — not found, a 404 |
| `/fr/about` | `en` | `/fr/about` — `fr` is not supported, so it is an ordinary segment |
| `/en/about` | — | Redirected to `/about` |
| `/TR/hakkinda` | — | Redirected to `/tr/hakkinda` |

Each page has one URL per locale, so the other spellings of a prefix redirect
permanently to it: the default locale's own prefix, which its URLs do not carry,
and a supported locale in another case. The redirect is a `301` for a `GET` or a
`HEAD` and a `308` for anything else, so a form posted to the wrong spelling is
posted again rather than turned into a `GET`, and the query string goes along.
Without it, `/en/about` would be a second copy of `/about` for a search engine and
a second entry in the cache.

A fragment reads the result as `rc.Locale`, and uses it to fetch the right
content:

```go
func aboutData(ctx context.Context, rc *collage.RenderContext) (aboutView, []string, error) {
	text, err := cms.Page(ctx, "about", rc.Locale)
	if err != nil {
		return aboutView{}, nil, err
	}
	return aboutView{Body: text}, []string{"page:about:" + rc.Locale}, nil
}
```

`DisablePathLocale: true` turns prefixes off entirely, leaving every request in
`Default` — for a site that is in one language, or one that
[negotiates](#negotiating-a-language-yourself) on its own.

### Why only the URL

collage never picks a locale from the `Accept-Language` header or a cookie.

A URL whose content depends on who is asking is one URL with several contents, and
everything that stores URLs gets it wrong: a cache serves the first reader's
language to everyone, a search engine indexes one version and never sees the
other, and a link someone shares opens in a language they did not send. It also
breaks in a less obvious way — a Turkish browser following a link to `/about`
would be looked up among the Turkish paths, where `/about` does not exist, and be
handed a 404.

With the locale in the URL, each URL means one thing, for every reader and every
cache. If you do want to use the reader's preferences, you still can, deliberately
— see [below](#negotiating-a-language-yourself).

## Links by name

A link written as a path breaks silently when the path changes, and cannot know
which locale it is in. Link by the name the page was registered under:

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

`pageURL` takes the page's name and then its path parameters as name–value pairs.
It works for [documents](/docs/documents) as well as pages.

- **It follows the render's locale.** On `/tr/hakkinda`, `{{pageURL "about"}}` is
  `/tr/hakkinda`; on `/about` it is `/about`. The prefix is added for you.
- **It falls back to the default locale.** A page with no path in the current
  locale links its default-locale path, so a Turkish page linking a page that
  exists only in English still renders.
- **It is strict.** An unknown name, a missing or empty parameter, or a parameter
  the pattern has no placeholder for fails the render. A link that cannot be built
  is a bug to find in development, not a 404 for a reader.
- **Values are escaped**, and a value of `.` or `..` is refused.
- **Values are strings.** Pass a number through `printf`:
  `{{pageURL "user" "id" (printf "%d" .ID)}}`.

### A link in a specific locale

`pageURLIn` takes the locale first, and links exactly that locale — no fallback:

```html
<a href="{{pageURLIn "tr" "about"}}" hreflang="tr">Hakkımızda</a>
```

A page with no path in that locale fails the render, as does a locale the site
does not support.

### A language switcher

`localeURL` is the page being rendered, in another locale, with the same path
parameters. It is empty when the page has no path in that locale, so `with` skips
a language the page has not been translated into:

```html
<nav class="languages">
  {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
  {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
</nav>
```

Put it in the layout and every page gets a switcher that only offers what exists.
A locale that is not supported at all is still an error — that is a typo in the
template, not an untranslated page.

Path parameters are carried over as they are. A page at `/blog/{slug}` in both
locales switches from `/blog/hello` to `/tr/blog/hello`; if your Turkish posts have
Turkish slugs, the switcher cannot know that, and you build that link yourself
from your content.

### Telling search engines about translations

A switcher is for readers. Search engines learn a page's translations from
`<link rel="alternate" hreflang="…">` in its head, which `rc.HoistAlternate(hreflang, href)`
(since v0.10.0) declares, one per language. Pair it with `app.URL` for each
locale the page exists in — [Head and SEO](/docs/head-and-seo#canonical-and-alternate-links)
has a layout that does it for every page.

### Links from Go

In Go — an action redirecting to a named page, a sitemap listing every page, the
alternates in a page's head — use `app.URL`:

```go
target, err := app.URL("blog-post", "tr", map[string]string{"slug": post.Slug})
if err != nil {
	return nil, err
}
return collage.SeeOther(target), nil
```

```go
func (a *App) URL(name, locale string, params map[string]string) (string, error)
```

An empty locale means the default one. It is as strict as `pageURLIn`: an unknown
name is `collage.ErrUnknownRoute`, a locale the route has no path in is
`collage.ErrNoPathInLocale`, and parameters that do not fill the pattern exactly
are `collage.ErrRouteParams`, and a locale no URL can carry — unsupported, or
anything but the default with `DisablePathLocale` on — is
`collage.ErrLocaleUnreachable`. The result is a path, prefix included; add your
site's origin when you need an absolute URL.

## Negotiating a language yourself

Choosing a language from the reader's browser is a decision about your site, so it
is yours to make, in [middleware](/docs/middleware-and-apis). There are two honest
ways to do it.

**Redirect** a reader who arrives at the default-locale home page with a Turkish
browser, or a language cookie of your own, to `/tr`. Every URL still means one
thing; you have only chosen where a new reader starts.

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			http.Redirect(w, r, "/tr", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
})
```

**Render one URL in the reader's language.** Decide in middleware, pass the
decision to data handlers through the request context, and tell collage with
`collage.Vary`, so the cache keeps one copy per language:

```go
type langKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		collage.Vary(r, "Accept-Language", lang)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), langKey{}, lang)))
	})
})
```

```go
func pageData(ctx context.Context, rc *collage.RenderContext) (view, []string, error) {
	lang, _ := ctx.Value(langKey{}).(string)
	// ...
}
```

`Vary` puts the value you resolved — `"tr"`, not the browser's whole header — into
the page's cache key, and the header name into the response's `Vary` header, so a
CDN keeps the versions apart too. It must be called from middleware. Since
v0.11.0 `Vary` closes when routing begins, on every route — cached or not — so a
call from a data handler, or anywhere else after routing, changes nothing and
returns `collage.ErrVaryTooLate`. This is the shape that trades the benefits of
[one URL, one content](#why-only-the-url) away, so choose it knowingly — usually
alongside `DisablePathLocale`.

A [static export](/docs/static-export) renders without requests, so no middleware
runs for it: an exported site can only have its locales in its URLs.
