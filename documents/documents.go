// Package documents holds the site's non-HTML routes that no plugin serves:
// llms.txt and llms-full.txt, read by a language model rather than a reader.
// robots.txt, the sitemap and the search index are elagoht/robots',
// elagoht/sitemap's and elagoht/search's.
package documents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/site"
)

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
