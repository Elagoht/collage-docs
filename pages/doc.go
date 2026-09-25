package pages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
	"github.com/Elagoht/collage-docs/ui"
)

// docView is what templates/pages/doc.html renders: the page, and the whole
// navigation for the sidebar, in the page's language.
type docView struct {
	T        ui.Text
	Page     *site.Page
	Sections []site.Section
}

// DocPage is every page of the documentation, at /docs/{slug} and, translated,
// /tr/docs/{slug}. A page not translated yet is a 404 in the translation.
//
// Static: a page's content changes when the site is rebuilt and at no other time,
// so it is rendered once per build and, served, once per process.
func DocPage(app *collage.App, docs func() (*site.Set, error)) *collage.Page {
	content := collage.NewFragment("doc-content", "pages/doc.html").
		WithDataHandler(collage.Load(func(_ context.Context, rc *collage.RenderContext) (docView, error) {
			set, err := docs()
			if err != nil {
				return docView{}, err
			}
			loaded := set.Site(rc.Locale)
			if loaded == nil {
				return docView{}, fmt.Errorf("%w: no documentation in %q", collage.ErrNotFound, rc.Locale)
			}
			page, err := loaded.Page(rc.Param("slug"))
			if errors.Is(err, site.ErrNoPage) {
				return docView{}, fmt.Errorf("%w: %w", collage.ErrNotFound, err)
			}
			if err != nil {
				return docView{}, err
			}
			rc.HoistTitle(page.Title + " — collage")
			if page.Description != "" {
				rc.HoistMeta("description", page.Description)
			}
			return docView{T: ui.For(rc.Locale), Page: page, Sections: loaded.Sections}, nil
		})).
		Required().
		Build()

	builder := collage.NewPage("doc").
		WithLayout(layouts.Layout(app, docs)).
		WithContent(content).
		Static()
	for _, locale := range site.Locales() {
		builder = builder.WithPath(locale, "/docs/{slug}")
	}
	return builder.Build()
}
