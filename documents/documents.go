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
		WithPath("en", "/robots.txt").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			return []byte("User-agent: *\nAllow: /\n\nSitemap: " + site.Origin + "/sitemap.xml\n"), nil, nil
		}).
		Static().
		Build()
}

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	NS      string   `xml:"xmlns,attr"`
	URLs    []entry  `xml:"url"`
}

type entry struct {
	Loc string `xml:"loc"`
}

// Sitemap lists the home page and every page of the documentation, built from
// their page names with app.URL — so a page whose path changes is listed where it
// now is.
func Sitemap(app *collage.App, docs func() (*site.Site, error)) *collage.Document {
	return collage.NewDocument("sitemap", "application/xml").
		WithPath("en", "/sitemap.xml").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			loaded, err := docs()
			if err != nil {
				return nil, nil, err
			}
			home, err := app.URL("home", "", nil)
			if err != nil {
				return nil, nil, err
			}
			set := urlset{NS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: []entry{{Loc: site.Origin + home}}}
			for _, page := range loaded.Pages() {
				path, err := app.URL("doc", "", map[string]string{"slug": page.Slug})
				if err != nil {
					return nil, nil, fmt.Errorf("sitemap: %w", err)
				}
				set.URLs = append(set.URLs, entry{Loc: site.Origin + path})
			}
			var out bytes.Buffer
			out.WriteString(xml.Header)
			encoder := xml.NewEncoder(&out)
			encoder.Indent("", "  ")
			if err := encoder.Encode(set); err != nil {
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
func Search(app *collage.App, docs func() (*site.Site, error)) *collage.Document {
	return collage.NewDocument("search", "application/json").
		WithPath("en", "/search.json").
		WithHandler(func(context.Context, *collage.RenderContext) ([]byte, []string, error) {
			loaded, err := docs()
			if err != nil {
				return nil, nil, err
			}
			var entries []searchEntry
			for _, page := range loaded.Pages() {
				path, err := app.URL("doc", "", map[string]string{"slug": page.Slug})
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
		Static().
		Build()
}
