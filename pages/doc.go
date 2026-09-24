package pages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
)

// docView is what templates/pages/doc.html renders: the page, and the whole
// navigation for the sidebar.
type docView struct {
	Page     *site.Page
	Sections []site.Section
}

// DocPage is every page of the documentation, at /docs/{slug}.
//
// Static: a page's content changes when the site is rebuilt and at no other time,
// so it is rendered once per build and, served, once per process.
func DocPage(docs func() (*site.Site, error)) *collage.Page {
	content := collage.NewFragment("doc-content", "pages/doc.html").
		WithDataHandler(collage.DataHandler(func(_ context.Context, rc *collage.RenderContext) (docView, []string, error) {
			loaded, err := docs()
			if err != nil {
				return docView{}, nil, err
			}
			page, err := loaded.Page(rc.Param("slug"))
			if errors.Is(err, site.ErrNoPage) {
				return docView{}, nil, fmt.Errorf("%w: %w", collage.ErrNotFound, err)
			}
			if err != nil {
				return docView{}, nil, err
			}
			rc.HoistTitle(page.Title + " — collage")
			if page.Description != "" {
				rc.HoistMeta("description", page.Description)
			}
			rc.HoistLink("canonical", site.Origin+page.URL())
			return docView{Page: page, Sections: loaded.Sections}, nil, nil
		})).
		Required().
		Build()

	return collage.NewPage("doc").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/docs/{slug}").
		Static().
		Build()
}
