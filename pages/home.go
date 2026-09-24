package pages

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/fragments/layouts"
	"github.com/Elagoht/collage-docs/site"
)

// homeView is what templates/pages/home.html renders.
type homeView struct {
	Sections []site.Section
}

// HomePage is the landing page, at "/".
func HomePage(docs func() (*site.Site, error)) *collage.Page {
	content := collage.NewFragment("home-content", "pages/home.html").
		WithDataHandler(collage.DataHandler(func(_ context.Context, rc *collage.RenderContext) (homeView, []string, error) {
			loaded, err := docs()
			if err != nil {
				return homeView{}, nil, err
			}
			rc.HoistTitle("collage — server-rendered pages from cached fragments, in Go")
			rc.HoistMeta("description", "A Go framework for server-rendered pages composed from cached fragments. No dependencies.")
			rc.HoistLink("canonical", site.Origin+"/")
			return homeView{Sections: loaded.Sections}, nil, nil
		})).
		Build()

	return collage.NewPage("home").
		WithLayout(layouts.Layout()).
		WithContent(content).
		WithPath("en", "/").
		Static().
		Build()
}
