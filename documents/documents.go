// Package documents holds the site's non-HTML routes: the files a static host and
// a search engine read rather than a reader.
package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/site"
)

// Robots allows everything and names every language's sitemap. It is the site's
// rather than a language's, and crawlers read it at the root alone, so it is at
// the root whatever the locale prefixes: AtRoot.
func Robots(app *collage.App) *collage.Document {
	return collage.NewDocument("robots", "text/plain; charset=utf-8").
		AtRoot("/robots.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			var body strings.Builder
			body.WriteString("User-agent: *\nAllow: /\n\n")
			for _, locale := range site.Locales() {
				sitemap, err := app.URL("sitemap", locale, nil)
				if err != nil {
					return nil, nil, fmt.Errorf("robots: %w", err)
				}
				body.WriteString("Sitemap: " + site.Origin + sitemap + "\n")
			}
			return []byte(body.String()), nil, nil
		}).
		Static().
		Build()
}

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	NS      string   `xml:"xmlns,attr"`
	XHTML   string   `xml:"xmlns:xhtml,attr"`
	URLs    []entry  `xml:"url"`
}

type entry struct {
	Loc string `xml:"loc"`
	// Alternates are the page in every language it is in, itself included, as
	// search engines ask for them.
	Alternates []alternate `xml:"xhtml:link"`
}

type alternate struct {
	Rel      string `xml:"rel,attr"`
	Hreflang string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
}

// Sitemap lists the pages of one language — /en/sitemap.xml the English ones,
// /tr/sitemap.xml the Turkish — each with every language it is in as an
// alternate. Addresses are built from page names with app.URL, so a page whose
// path changes is listed where it now is.
func Sitemap(app *collage.App, docs func() (*site.Set, error)) *collage.Document {
	builder := collage.NewDocument("sitemap", "application/xml").
		WithHandler(func(_ context.Context, rc *collage.RenderContext) ([]byte, []string, error) {
			set, err := docs()
			if err != nil {
				return nil, nil, err
			}
			loaded := set.Site(rc.Locale)
			if loaded == nil {
				return nil, nil, fmt.Errorf("%w: no documentation in %q", collage.ErrNotFound, rc.Locale)
			}
			urls := urlset{NS: "http://www.sitemaps.org/schemas/sitemap/0.9", XHTML: "http://www.w3.org/1999/xhtml"}

			// add lists one page in this sitemap's language, carrying it in every
			// language it is in.
			add := func(name string, params map[string]string, in func(locale string) bool) error {
				var links []alternate
				var loc string
				for _, locale := range set.Locales() {
					if !in(locale) {
						continue
					}
					path, err := app.URL(name, locale, params)
					if err != nil {
						return fmt.Errorf("sitemap: %w", err)
					}
					links = append(links, alternate{Rel: "alternate", Hreflang: locale, Href: site.Origin + path})
					if locale == site.Original {
						links = append(links, alternate{Rel: "alternate", Hreflang: "x-default", Href: site.Origin + path})
					}
					if locale == rc.Locale {
						loc = site.Origin + path
					}
				}
				urls.URLs = append(urls.URLs, entry{Loc: loc, Alternates: links})
				return nil
			}

			if err := add("home", nil, func(string) bool { return true }); err != nil {
				return nil, nil, err
			}
			for _, page := range loaded.Pages() {
				translated := func(locale string) bool {
					_, err := set.Site(locale).Page(page.Slug)
					return err == nil
				}
				if err := add("doc", map[string]string{"slug": page.Slug}, translated); err != nil {
					return nil, nil, err
				}
			}

			var out bytes.Buffer
			out.WriteString(xml.Header)
			encoder := xml.NewEncoder(&out)
			encoder.Indent("", "  ")
			if err := encoder.Encode(urls); err != nil {
				return nil, nil, err
			}
			return []byte(strings.TrimSpace(out.String()) + "\n"), nil, nil
		}).
		Static()
	for _, locale := range site.Locales() {
		builder = builder.WithPath(locale, "/sitemap.xml")
	}
	return builder.Build()
}

// summary is what llms.txt says collage is, before it lists the pages.
const summary = `> collage is a Go framework for server-rendered websites. A page is a layout
> around fragments; each fragment fetches its own data concurrently and renders its
> own html/template, and the page is cached and invalidated by the dependency tags
> its data carried. One Go binary serves the site, its forms and APIs, or exports it
> as static files. No client framework, no JavaScript build step, no dependencies
> beyond the standard library.

- Module: ` + "`github.com/Elagoht/collage`" + `; the package applications import is ` + "`github.com/Elagoht/collage/pkg/collage`" + `.
- CLI: ` + "`go install github.com/Elagoht/collage/cmd/collage@latest`" + `, then ` + "`collage new mysite`" + ` and ` + "`collage dev`" + `.
- License: MIT. Source: https://github.com/Elagoht/collage
`

// LLMs is /llms.txt, the index a language model reads a site by
// (https://llmstxt.org): what collage is, then every page of the English
// documentation by section, each with its description, and where to find the
// rest. At the root, as the convention has it.
func LLMs(app *collage.App, docs func() (*site.Set, error)) *collage.Document {
	return collage.NewDocument("llms", "text/markdown; charset=utf-8").
		AtRoot("/llms.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			set, err := docs()
			if err != nil {
				return nil, nil, err
			}
			var body strings.Builder
			body.WriteString("# collage\n\n" + summary)
			for _, section := range set.Site(site.Original).Sections {
				body.WriteString("\n## " + section.Title + "\n\n")
				for _, page := range section.Pages {
					path, err := app.URL("doc", site.Original, map[string]string{"slug": page.Slug})
					if err != nil {
						return nil, nil, fmt.Errorf("llms.txt: %w", err)
					}
					fmt.Fprintf(&body, "- [%s](%s): %s\n", page.Title, site.Origin+path, page.Description)
				}
			}
			full, err := app.URL("llms-full", "", nil)
			if err != nil {
				return nil, nil, fmt.Errorf("llms.txt: %w", err)
			}
			turkish, err := app.URL("home", "tr", nil)
			if err != nil {
				return nil, nil, fmt.Errorf("llms.txt: %w", err)
			}
			body.WriteString("\n## Optional\n\n")
			fmt.Fprintf(&body, "- [The whole documentation in one file](%s): every page above, in reading order, as Markdown\n", site.Origin+full)
			fmt.Fprintf(&body, "- [Go reference](%s): every exported identifier of package collage, generated from its doc comments\n", site.ReferenceBase)
			fmt.Fprintf(&body, "- [Türkçe dokümantasyon](%s): the same documentation in Turkish\n", site.Origin+turkish)
			return []byte(body.String()), nil, nil
		}).
		Static().
		Build()
}

// LLMsFull is /llms-full.txt: the English documentation whole, one page after
// another in reading order, as the Markdown it is written in, with every link to
// another page absolute. For a model that would rather read the lot than follow
// links.
func LLMsFull(app *collage.App, docs func() (*site.Set, error)) *collage.Document {
	return collage.NewDocument("llms-full", "text/markdown; charset=utf-8").
		AtRoot("/llms-full.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			set, err := docs()
			if err != nil {
				return nil, nil, err
			}
			var body strings.Builder
			body.WriteString("# collage documentation\n\n" + summary)
			for _, page := range set.Site(site.Original).Pages() {
				path, err := app.URL("doc", site.Original, map[string]string{"slug": page.Slug})
				if err != nil {
					return nil, nil, fmt.Errorf("llms-full.txt: %w", err)
				}
				fmt.Fprintf(&body, "\n---\n\n# %s\n\nSource: %s\n\n", page.Title, site.Origin+path)
				if page.Description != "" {
					body.WriteString("> " + page.Description + "\n\n")
				}
				body.WriteString(absolute.ReplaceAllString(page.Markdown, "]("+site.Origin+"/"))
			}
			return []byte(body.String()), nil, nil
		}).
		Static().
		Build()
}

// absolute finds a Markdown link to a path on this site.
var absolute = regexp.MustCompile(`\]\(/`)

// searchEntry is one searchable stretch of a page. The keys are one letter
// because the file holds every word of the documentation and is fetched whole.
type searchEntry struct {
	Title   string `json:"t"`
	Section string `json:"s"`
	Heading string `json:"h,omitempty"`
	URL     string `json:"u"`
	Text    string `json:"x"`
}

// Search is the index the site's search box reads: every page, split at its
// headings, so a result links to the heading it matched under. It is a Static
// document, so an export writes it as a file and no server is involved in
// searching at all — static/search.js fetches it the first time someone searches.
//
// Each language has its own, /search.json and /tr/search.json, holding the pages
// in that language: a Turkish reader searches in Turkish.
func Search(app *collage.App, docs func() (*site.Set, error)) *collage.Document {
	builder := collage.NewDocument("search", "application/json").
		WithHandler(func(_ context.Context, rc *collage.RenderContext) ([]byte, []string, error) {
			set, err := docs()
			if err != nil {
				return nil, nil, err
			}
			loaded := set.Site(rc.Locale)
			if loaded == nil {
				return nil, nil, fmt.Errorf("%w: no documentation in %q", collage.ErrNotFound, rc.Locale)
			}
			entries := []searchEntry{}
			for _, page := range loaded.Pages() {
				path, err := app.URL("doc", rc.Locale, map[string]string{"slug": page.Slug})
				if err != nil {
					return nil, nil, fmt.Errorf("search: %w", err)
				}
				for _, part := range page.Parts {
					url := path
					if part.ID != "" {
						url += "#" + part.ID
					}
					text := part.Text
					if part.Heading == "" && page.Description != "" {
						text = page.Description + " " + text
					}
					entries = append(entries, searchEntry{
						Title: page.Title, Section: page.Section, Heading: part.Heading, URL: url, Text: text,
					})
				}
			}
			body, err := json.Marshal(entries)
			return body, nil, err
		}).
		Static()
	for _, locale := range site.Locales() {
		builder = builder.WithPath(locale, "/search.json")
	}
	return builder.Build()
}
