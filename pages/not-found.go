package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
)

// NotFoundPage is what an unknown address answers with, and 404.html in an export.
func NotFoundPage() *collage.Page {
	content := collage.NewFragment("not-found-content", "pages/404.html").
		WithDataHandler(collage.Effect(func(_ context.Context, rc *collage.RenderContext) error {
			rc.HoistTitle("Not found — collage")
			return nil
		})).
		Build()

	return collage.NewPage("not-found").
		WithLayout(layouts.Layout()).
		WithContent(content).
		Dynamic().
		Build()
}
