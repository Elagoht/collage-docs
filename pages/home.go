package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
	"github.com/Elagoht/collage-docs/ui"
)

// homeView is what templates/pages/home.html renders.
type homeView struct {
	T        ui.Text
	Sections []site.Section
}

// HomePage is the landing page, at "/" and, in Turkish, "/tr/".
func HomePage(app *collage.App, docs func() (*site.Set, error)) *collage.Page {
	content := collage.NewFragment("home-content", "pages/home.html").
		WithDataHandler(collage.DataHandler(func(_ context.Context, rc *collage.RenderContext) (homeView, []string, error) {
			set, err := docs()
			if err != nil {
				return homeView{}, nil, err
			}
			text := ui.For(rc.Locale)
			rc.HoistTitle(text.HomeTitle)
			rc.HoistMeta("description", text.HomeDescription)
			var sections []site.Section
			if loaded := set.Site(rc.Locale); loaded != nil {
				sections = loaded.Sections
			}
			return homeView{T: text, Sections: sections}, nil, nil
		})).
		Build()

	builder := collage.NewPage("home").
		WithLayout(layouts.Layout(app, docs)).
		WithContent(content).
		Static()
	for _, locale := range site.Locales() {
		builder = builder.WithPath(locale, "/")
	}
	return builder.Build()
}
