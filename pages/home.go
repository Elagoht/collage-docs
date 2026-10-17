package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	i18n "github.com/Elagoht/collage-i18n"
	meta "github.com/Elagoht/collage-meta"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
)

// homeView is what templates/pages/home.html renders.
type homeView struct {
	Sections []site.Section
}

// HomePage is the landing page, at "/" and, in Turkish, "/tr/".
func HomePage(app *collage.App, docs func() (*site.Set, error)) *collage.Page {
	content := collage.NewFragment("home-content", "pages/home.html").
		WithData(collage.Load(func(_ context.Context, rc *collage.RenderContext) (homeView, error) {
			set, err := docs()
			if err != nil {
				return homeView{}, err
			}
			rc.HoistTitle(i18n.T(rc, "home.title"))
			meta.Set(rc, meta.Page{Description: i18n.T(rc, "home.description")})
			var sections []site.Section
			if loaded := set.Site(rc.Locale); loaded != nil {
				sections = loaded.Sections
			}
			return homeView{Sections: sections}, nil
		})).
		Build()

	builder := collage.NewPage("home").
		WithLayouts(layouts.Layout(app, docs)).
		WithContent(content).
		Static()
	for _, locale := range site.Locales() {
		builder = builder.WithPath(locale, "/")
	}
	return builder.Build()
}
