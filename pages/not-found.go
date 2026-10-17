package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	i18n "github.com/Elagoht/collage-i18n"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
)

// NotFoundPage is what an unknown address answers with, and 404.html in an export.
func NotFoundPage(app *collage.App, docs func() (*site.Set, error)) *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").
		WithData(collage.Load(func(_ context.Context, rc *collage.RenderContext) (struct{}, error) {
			rc.HoistTitle(i18n.T(rc, "notFound.title"))
			return struct{}{}, nil
		})).
		Build()

	return collage.NewPage("not-found").
		WithLayouts(layouts.Layout(app, docs)).
		WithContent(content).
		Dynamic().
		Build()
}
