---
description: Serving a site in several languages, and linking between pages by name so links follow them in every locale.
reference: LocaleConfig, PageBuilder.WithPath, PageBuilder.WithFragmentPath, RenderContext, URL, App, Vary, ErrLocaleUnreachable, ErrNoPathInLocale, ErrUnknownRoute, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Links and locales

A collage site can serve every page in several languages, each at its own URL. It
routes locales; it does not translate. What a word should be in Turkish lives in
your content, wherever that already is — collage's part is knowing which URL
belongs to which language, and building links that know it too. For the words in
your templates, the [elagoht/i18n](/docs/plugins#elagohti18n) plugin keeps a
catalog per locale and translates in the locale the page is rendered in.

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
	WithLayouts(layout).
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda").
	Build()
```

The paths may differ, as here, or be the same pattern in every locale. A page with
a path in only one locale exists only in that one.

The locale in `WithPath` is not a label. It picks which locale's routes the path
joins, and a request only reaches a locale that is `Default` or in `Supported`.
`Default` is `"en"` unless you set it, so a site in Turkish alone sets it:

```go
Locale: collage.LocaleConfig{Default: "tr"},
```

and declares `WithPath("tr", "/hakkinda")`.

Since v0.35.0, a path in a locale no URL reaches is refused at registration with
`collage.ErrLocaleUnreachable`, and the error names the page and the locale
config. Before, the site started without complaint and then answered 404 at every
URL. The same holds for actions and documents.

Since v0.60.0 `WithPath` also refuses two more mistakes, for a page, a document or
an action alike:

- a locale named twice: `collage.ErrDuplicateLocalePath`, keeping the first path;
- an empty locale: `collage.ErrInvalidLocale`.

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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 260" role="img" aria-labelledby="ll-url-title">
<title id="ll-url-title">With Default en, Supported en and tr, and no PrefixDefault, a request path's first segment is checked first. If it names no supported locale, as in /about or /fr/about, the locale is en and the path is matched as it came. If it names tr as spelled, as in /tr/hakkinda, the locale is tr and the prefix is removed. If it is the default locale's own prefix or another case, as in /en/about or /TR/hakkinda, the request is redirected, 301 or 308, to the one spelling, before any page is looked up. The remaining path is matched among that locale's paths only, so /tr/about is a 404.</title>
<defs><marker id="ll-url-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="100" y="8" width="120" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">request path</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#ll-url-arrow)" />
<rect class="d-frame" x="40" y="50" width="240" height="26" rx="6" /><text x="160" y="67" text-anchor="middle">first segment names a locale?</text>
<path class="d-flow" d="M100 76C100 92 54 90 54 106" marker-end="url(#ll-url-arrow)" />
<line class="d-flow" x1="160" y1="76" x2="160" y2="106" marker-end="url(#ll-url-arrow)" />
<path class="d-flow" d="M220 76C220 92 266 90 266 106" marker-end="url(#ll-url-arrow)" />
<text class="d-small" x="58" y="92">no</text>
<text class="d-small" x="166" y="96">tr</text>
<text class="d-small" x="250" y="88">en, TR</text>
<rect class="d-a" x="6" y="110" width="96" height="34" rx="5" /><text class="d-on-a" x="54" y="124" text-anchor="middle">Default: en</text><text class="d-on-a d-small" x="54" y="137" text-anchor="middle">path as it came</text>
<rect class="d-b" x="112" y="110" width="96" height="34" rx="5" /><text class="d-on-b" x="160" y="124" text-anchor="middle">tr</text><text class="d-on-b d-small" x="160" y="137" text-anchor="middle">prefix removed</text>
<rect class="d-c" x="218" y="110" width="96" height="34" rx="5" /><text class="d-on-c" x="266" y="124" text-anchor="middle">redirect</text><text class="d-on-c d-small" x="266" y="137" text-anchor="middle">301 or 308</text>
<text class="d-small" x="54" y="160" text-anchor="middle">/about</text><text class="d-small" x="54" y="172" text-anchor="middle">/fr/about</text>
<text class="d-small" x="160" y="160" text-anchor="middle">/tr/hakkinda</text><text class="d-small" x="160" y="172" text-anchor="middle">/tr/about</text>
<text class="d-small" x="266" y="160" text-anchor="middle">/en/about</text><text class="d-small" x="266" y="172" text-anchor="middle">→ /about</text>
<text class="d-small" x="266" y="188" text-anchor="middle">/TR/hakkinda</text><text class="d-small" x="266" y="200" text-anchor="middle">→ /tr/hakkinda</text>
<line class="d-flow" x1="54" y1="178" x2="54" y2="208" marker-end="url(#ll-url-arrow)" />
<line class="d-flow" x1="160" y1="178" x2="160" y2="208" marker-end="url(#ll-url-arrow)" />
<rect class="d-frame" x="6" y="212" width="202" height="40" rx="6" /><text x="107" y="228" text-anchor="middle">matched in that locale</text><text class="d-small" x="107" y="242" text-anchor="middle">/tr/about: no tr path, a 404</text>
</svg>
<figcaption>With <code>Default: "en"</code> and <code>Supported: []string{"en", "tr"}</code>, and no <code>PrefixDefault</code>, the first segment of the path picks the locale. No supported locale there means the default, with the path as it came. <code>tr</code> means Turkish, with the prefix removed. The default's own prefix, or a locale in another case, is redirected to the one spelling before any page is looked up. What is left is matched among that locale's paths only.</figcaption>
</figure>

Each page has one URL per locale, so the other spellings of a prefix redirect
permanently to it:

- the default locale's own prefix, which its URLs do not carry;
- a supported locale in another case.

The redirect is a `301` for a `GET` or a `HEAD` and a `308` for anything else, so
a form posted to the wrong spelling is posted again rather than turned into a
`GET`, and the query string goes along. Without it, `/en/about` would be a second
copy of `/about` for a search engine and a second entry in the cache.

A site that wants every language under a prefix, the default one included, sets
`PrefixDefault: true` (since v0.14.0). Then `/en/about` is the page beside
`/tr/hakkinda`, and a bare URL is no language at all:

- A page requested without a prefix — `/about`, or the root — redirects
  permanently to the default locale's address, `/en/about` and `/en`.
- Every link built by name carries the prefix.
- Documents are prefixed the same way, `/en/sitemap.xml` beside
  `/tr/sitemap.xml`, except the site's own files, built with
  [`AtRoot`](/docs/documents#example-robotstxt): `/robots.txt` is at the root and
  nowhere else (since v0.14.1).
- An export writes the default locale under `en/`, and at its root a copy of the
  `/en/` home that names `/en/` as its canonical address.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 258" role="img" aria-labelledby="ll-prefix-title">
<title id="ll-prefix-title">With PrefixDefault true, /en/about serves the about page in en and /tr/hakkinda in tr. A request for /about, a page of the default locale without its prefix, is redirected 301 to /en/about, and the root to /en. A static export writes the en pages under en/, the tr pages under tr/, and at its root a copy of the en home page whose canonical address is /en/.</title>
<defs><marker id="ll-prefix-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<text class="d-small" x="6" y="14">a request, PrefixDefault: true</text>
<rect class="d-chip" x="6" y="22" width="104" height="26" rx="13" /><text x="58" y="39" text-anchor="middle">/en/about</text>
<line class="d-flow" x1="110" y1="35" x2="150" y2="35" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-a" x="154" y="22" width="160" height="26" rx="5" /><text class="d-on-a" x="234" y="39" text-anchor="middle">about, in en</text>
<rect class="d-chip" x="6" y="54" width="104" height="26" rx="13" /><text x="58" y="71" text-anchor="middle">/tr/hakkinda</text>
<line class="d-flow" x1="110" y1="67" x2="150" y2="67" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-b" x="154" y="54" width="160" height="26" rx="5" /><text class="d-on-b" x="234" y="71" text-anchor="middle">about, in tr</text>
<rect class="d-chip" x="6" y="86" width="104" height="26" rx="13" /><text x="58" y="103" text-anchor="middle">/about</text>
<line class="d-flow" x1="110" y1="99" x2="150" y2="99" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-c" x="154" y="86" width="160" height="26" rx="5" /><text class="d-on-c" x="234" y="103" text-anchor="middle">301 → /en/about</text>
<rect class="d-chip" x="6" y="118" width="104" height="26" rx="13" /><text x="58" y="135" text-anchor="middle">/</text>
<line class="d-flow" x1="110" y1="131" x2="150" y2="131" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-c" x="154" y="118" width="160" height="26" rx="5" /><text class="d-on-c" x="234" y="135" text-anchor="middle">301 → /en</text>
<line class="d-rule" x1="6" y1="160" x2="314" y2="160" />
<text class="d-small" x="6" y="178">a static export</text>
<rect class="d-a" x="6" y="186" width="96" height="40" rx="5" /><text class="d-on-a" x="54" y="202" text-anchor="middle">en/</text><text class="d-on-a d-small" x="54" y="217" text-anchor="middle">the en pages</text>
<rect class="d-b" x="112" y="186" width="96" height="40" rx="5" /><text class="d-on-b" x="160" y="202" text-anchor="middle">tr/</text><text class="d-on-b d-small" x="160" y="217" text-anchor="middle">the tr pages</text>
<rect class="d-frame" x="218" y="186" width="96" height="40" rx="5" /><text x="266" y="202" text-anchor="middle">the root</text><text class="d-small" x="266" y="217" text-anchor="middle">a copy of /en</text>
<text class="d-small" x="266" y="242" text-anchor="middle">canonical: /en/</text>
</svg>
<figcaption>With <code>PrefixDefault: true</code> every language has a prefix. A default-locale page asked for at its bare address is redirected to the prefixed one with a 301, the root included. A static host has nothing to send the bare domain on to the prefixed home, so the export writes the default locale's pages under <code>en/</code> and puts a copy of the <code>/en/</code> home at its root, naming <code>/en/</code> as its canonical address.</figcaption>
</figure>

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

There is no switch that turns prefixes off: `DisablePathLocale` was removed in
v0.60.0. A site in one language registers its paths in `Default`, which carries no
prefix, and one that [negotiates](#negotiating-a-language-yourself) on its own does
it in middleware.

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
- **Values are strings, integers or `fmt.Stringer`s** (the last two since
  v0.36.0): `{{pageURL "user" "id" .ID}}`.
- **It ends in `/` when [`TrailingSlash`](/docs/configuration#trailingslash) is
  on** — a page's link, not a document's: `/about/`, `/tr/`, but `/feed.xml`.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 200" role="img" aria-labelledby="ll-name-title">
<title id="ll-name-title">One template call, pageURL "about", gives /about when the page renders in en at /about, and /tr/hakkinda when it renders in tr at /tr/hakkinda. A page with no tr path, careers, is linked at its en path, /careers, from a tr render.</title>
<defs><marker id="ll-name-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="70" y="8" width="180" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">pageURL "about"</text>
<path class="d-flow" d="M130 32C130 42 79 38 79 46" marker-end="url(#ll-name-arrow)" />
<path class="d-flow" d="M190 32C190 42 241 38 241 46" marker-end="url(#ll-name-arrow)" />
<rect class="d-a" x="6" y="50" width="146" height="40" rx="5" /><text class="d-on-a" x="79" y="66" text-anchor="middle">on /about</text><text class="d-on-a d-small" x="79" y="81" text-anchor="middle">a render in en</text>
<rect class="d-b" x="168" y="50" width="146" height="40" rx="5" /><text class="d-on-b" x="241" y="66" text-anchor="middle">on /tr/hakkinda</text><text class="d-on-b d-small" x="241" y="81" text-anchor="middle">a render in tr</text>
<line class="d-flow" x1="79" y1="90" x2="79" y2="104" marker-end="url(#ll-name-arrow)" />
<line class="d-flow" x1="241" y1="90" x2="241" y2="104" marker-end="url(#ll-name-arrow)" />
<rect class="d-chip" x="19" y="108" width="120" height="24" rx="12" /><text x="79" y="124" text-anchor="middle">/about</text>
<rect class="d-chip" x="181" y="108" width="120" height="24" rx="12" /><text x="241" y="124" text-anchor="middle">/tr/hakkinda</text>
<rect class="d-frame" x="6" y="148" width="308" height="44" rx="6" /><text x="160" y="165" text-anchor="middle">pageURL "careers", on /tr/hakkinda</text><text class="d-small" x="160" y="181" text-anchor="middle">no tr path: the en one, /careers</text>
</svg>
<figcaption>The same <code>{{pageURL "about"}}</code> in a shared layout is <code>/about</code> on the English page and <code>/tr/hakkinda</code> on the Turkish one. A page with no path in the render's locale, here a <code>careers</code> page that exists only in English, is linked at its default-locale path.</figcaption>
</figure>

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

A form posting to an [action at its own URL](/docs/forms-and-actions#an-action-at-its-own-url)
names the action the same way, with `{{actionURL "logout"}}` (since v0.36.0).

### Links from Go

In a data handler or an action — a redirect to a named page, a canonical URL —
use `rc.URL`, and `rc.ActionURL` for an action (since v0.37.0). They build in the
render's own locale and fall back to the default one, as `pageURL` and
`actionURL` do, so nothing has to hand the `*App` to the code that needs a link:

```go
target, err := rc.URL("blog-post", map[string]string{"slug": post.Slug})
if err != nil {
	return nil, err
}
return collage.SeeOther(target), nil
```

```go
func (rc *RenderContext) URL(name string, params map[string]string) (string, error)
func (rc *RenderContext) ActionURL(name string, params map[string]string) (string, error)
```

A value, even one straight from the request, is one escaped path segment: one
holding `/`, or being `.` or `..`, is `collage.ErrRouteParams`, so it cannot turn
the path into another site's URL. On a `RenderContext` collage did not make — one
a test builds by hand — both return `collage.ErrUnknownRoute`.

In a guard or a middleware, which run in a request but have no render, it is
`collage.URL(ctx, name, params)` (since v0.73.0): the same builder, in the locale
the request resolved to — the default one in a middleware, which runs before
routing — and `ErrUnknownRoute` outside a request collage serves.
`collage.GuardSeeOtherPage(ctx, name, params)` turns it into a guard's redirect
(see [Pages and layouts](/docs/pages-and-layouts#private-pages-guards)).

For a locale other than the render's own, or outside a render — a sitemap
listing every page, the alternates in a page's head — use `app.URL`, which takes
the locale:

```go
target, err := app.URL("blog-post", "tr", map[string]string{"slug": post.Slug})
```

```go
func (a *App) URL(name, locale string, params map[string]string) (string, error)
```

An empty locale means the default one. It is as strict as `pageURLIn`:

| What is wrong | Error |
| --- | --- |
| An unknown name | `collage.ErrUnknownRoute` |
| A locale the route has no path in | `collage.ErrNoPathInLocale` |
| Parameters that do not fill the pattern exactly | `collage.ErrRouteParams` |
| A locale no URL can carry — neither the default nor supported | `collage.ErrLocaleUnreachable` |

The result is a path, prefix included; add your site's origin when you need an
absolute URL.

### A fragment's URL

A fragment a page opened at its own URL with
[`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url) is linked
the same way, by the page's name and the fragment's (since v0.18.0):

```html
<div data-live="{{fragmentURL "home" "cpu-usage"}}">{{slot "cpu-usage"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

`fragmentURL` uses the render's locale and falls back to the default one;
`fragmentURLIn "tr" "home" "cpu-usage"` names the locale. In Go it is `app.FragmentURL`:

```go
func (a *App) FragmentURL(page, fragment, locale string, params map[string]string) (string, error)
```

It is as strict as `app.URL`:

| What is wrong | Error |
| --- | --- |
| An unknown page | `collage.ErrUnknownRoute` |
| A fragment the page did not open | `collage.ErrUnknownFragmentPath` |
| A fragment opened at two paths in one locale | `collage.ErrAmbiguousFragmentPath` |
| Missing parameters | `collage.ErrRouteParams` |

In a template, each fails the render.

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
func pageData(ctx context.Context, rc *collage.RenderContext) (pageView, []string, error) {
	lang, _ := ctx.Value(langKey{}).(string)
	// ...
}
```

`Vary` puts the value you resolved — `"tr"`, not the browser's whole header — into
the page's cache key, and the header name into the response's `Vary` header, so a
CDN keeps the versions apart too.

It must be called from middleware. Since v0.11.0 `Vary` closes when routing
begins, on every route — cached or not — so a call from a data handler, or
anywhere else after routing, changes nothing and returns `collage.ErrVaryTooLate`.

This is the shape that trades the benefits of
[one URL, one content](#why-only-the-url) away, so choose it knowingly — usually on
a site whose pages all have their paths in `Default`, so no URL names a locale.

A [static export](/docs/static-export) renders without requests, so no middleware
runs for the files it writes: an exported site can only have its locales in its
URLs.

## Example: a bilingual product catalogue

A shop sells in English and Turkish. Its two pages have a path in each locale, and
the product page opens its stock box at a URL of its own, so a script can refresh
it:

| Page | `en` | `tr` |
| --- | --- | --- |
| `catalogue` | `/products` | `/tr/urunler` |
| `product` | `/products/{id}` | `/tr/urunler/{id}` |
| `product`'s stock box | `/products/{id}/stock` | `/tr/urunler/{id}/stok` |

The locale configuration is the one [above](#configuring-locales): `Default: "en"`,
`Supported: []string{"en", "tr"}`. The `catalog` package stands for your own data
layer; it keeps each product's name in both languages, and the handlers ask it for
the render's locale:

```go
func loadCatalogue(ctx context.Context, rc *collage.RenderContext) ([]catalog.Product, []string, error) {
	list, err := catalog.List(ctx, rc.Locale)
	if err != nil {
		return nil, nil, err
	}
	return list, []string{"catalog"}, nil
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (catalog.Product, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Find(ctx, id, rc.Locale)
	if errors.Is(err, catalog.ErrNoProduct) {
		return catalog.Product{}, nil, fmt.Errorf("catalog: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return catalog.Product{}, nil, err
	}
	return p, []string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (int, error) {
	return catalog.Stock(ctx, rc.Param("id"))
}
```

Each page declares a path per locale, and the product page opens the stock
fragment in each:

```go
layout := collage.NewFragment("layout", "layouts/master.html").Build()

stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

catalogue := collage.NewPage("catalogue").
	WithLayouts(layout).
	WithContent(collage.NewFragment("catalogue", "pages/catalogue.html").
		WithData(collage.LoadTagged(loadCatalogue)).
		Build()).
	WithPath("en", "/products").
	WithPath("tr", "/urunler").
	Build()

product := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(collage.NewFragment("product", "pages/product.html").
		WithData(collage.LoadTagged(loadProduct)).
		WithSlotFragment("stock", stock).
		Build()).
	WithPath("en", "/products/{id}").
	WithPath("tr", "/urunler/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	WithFragmentPath("tr", "/urunler/{id}/stok", stock).
	Build()

if err := app.Register(catalogue, product); err != nil {
	return err
}
```

The templates name pages, never paths. The layout carries the language switcher,
the catalogue links each product by name with its `id`, and the product page
names its stock box's URL:

```html
<!-- templates/layouts/master.html -->
<nav class="languages">
  {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
  {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
</nav>
{{slot "content"}}
```

```html
<!-- templates/pages/catalogue.html -->
<ul>
  {{range .}}
  <li><a href="{{pageURL "product" "id" .ID}}">{{.Name}}</a></li>
  {{end}}
</ul>
```

```html
<!-- templates/pages/product.html -->
<article>
  <h1>{{.Name}}</h1>
  <div data-live="{{fragmentURL "product" "product-stock" "id" .ID}}">{{slot "stock"}}</div>
</article>
```

```html
<!-- templates/fragments/stock.html -->
<p class="stock">{{.}}</p>
```

On `/tr/urunler/42`, the links come out as:

| In the template | Becomes |
| --- | --- |
| `localeURL "en"` | `/products/42` |
| `localeURL "tr"` | `/tr/urunler/42` |
| `pageURL "product" "id" .ID`, in the catalogue | `/tr/urunler/42` |
| `fragmentURL "product" "product-stock" "id" .ID` | `/tr/urunler/42/stok` |

On `/products/42` the same templates give `/products/42` and
`/products/42/stock`. A few things make this work:

- **Both patterns name the parameter `{id}`.** `localeURL` carries the render's
  parameters over by name, and a pattern they do not fill fails the render. The
  id is the same number in both languages, so the switcher needs nothing else; a
  slug translated per language would need the link built from your content, as
  [above](#a-language-switcher).
- **`fragmentURL` names the fragment, not the slot**: `product-stock`, the name
  given to `NewFragment`.
- **The stock box has a path in each locale.** With only the English one, a
  Turkish page would link `/products/42/stock`, its
  [default-locale fallback](#a-fragments-url).
- **The wrong spellings are handled for you.** `/en/products/42` redirects to
  `/products/42`, and `/tr/products/42` is a 404, since `/products/{id}` is not
  among the Turkish paths.
