package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
	"github.com/Elagoht/collage-docs/ui"
)

// NotFoundPage is what an unknown address answers with, and 404.html in an export.
func NotFoundPage(app *collage.App, docs func() (*site.Set, error)) *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").
		WithDataHandler(collage.Load(func(_ context.Context, rc *collage.RenderContext) (ui.Text, error) {
			text := ui.For(rc.Locale)
			rc.HoistTitle(text.NotFoundTitle)
			return text, nil
		})).
		Build()

	return collage.NewPage("not-found").
		WithLayout(layouts.Layout(app, docs)).
		WithContent(content).
		Dynamic().
		Build()
}
