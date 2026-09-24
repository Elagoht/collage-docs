// Package documents holds the site's non-HTML routes: the files a static host and
// a search engine read rather than a reader.
package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/site"
)

// Robots allows everything and points at the sitemap.
func Robots() *collage.Document {
	return collage.NewDocument("robots", "text/plain; charset=utf-8").
		WithPath(site.Original, "/robots.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			return []byte("User-agent: *\nAllow: /\n\nSitemap: " + site.Origin + "/sitemap.xml\n"), nil, nil
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

// Sitemap lists the home page and every page of the documentation, in every
// language each is in, with the other languages as alternates. Addresses are
// built from page names with app.URL, so a page whose path changes is listed where
// it now is.
func Sitemap(app *collage.App, docs func() (*site.Set, error)) *collage.Document {
	return collage.NewDocument("sitemap", "application/xml").
		WithPath(site.Original, "/sitemap.xml").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			set, err := docs()
			if err != nil {
				return nil, nil, err
			}
			urls := urlset{NS: "http://www.sitemaps.org/schemas/sitemap/0.9", XHTML: "http://www.w3.org/1999/xhtml"}

			// add lists one page: an entry per language it is in, each carrying
			// all of them.
			add := func(name string, params map[string]string, in func(locale string) bool) error {
				var links []alternate
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
				}
				for _, link := range links {
					if link.Hreflang != "x-default" {
						urls.URLs = append(urls.URLs, entry{Loc: link.Href, Alternates: links})
					}
				}
				return nil
			}

			if err := add("home", nil, func(string) bool { return true }); err != nil {
				return nil, nil, err
			}
			for _, page := range set.Site(site.Original).Pages() {
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
		Static().
		Build()
}

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
