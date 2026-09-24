package layouts

import (
	"context"
	"errors"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/site"
	"github.com/Elagoht/collage-docs/ui"
)

// layoutView is what templates/layouts/default.html renders.
type layoutView struct {
	Lang string
	T    ui.Text
	// Languages are the other languages the site is in, each linked to this page
	// in it — or to its home page, where this page is not translated.
	Languages []language
}

type language struct {
	Code string
	Name string
	URL  string
}

// Layout is the frame of every page. Its handler declares, for whichever page is
// being rendered, the canonical link and one alternate link per language the page
// exists in, plus x-default for the original — built with app.URL from the page's
// name, so they follow the page's path.
func Layout(app *collage.App, docs func() (*site.Set, error)) *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithDataHandler(collage.DataHandler(func(_ context.Context, rc *collage.RenderContext) (layoutView, []string, error) {
			view := layoutView{Lang: rc.Locale, T: ui.For(rc.Locale)}
			set, err := docs()
			if err != nil {
				return layoutView{}, nil, err
			}

			// The canonical link first, then one alternate per language, the page's
			// own included, as search engines ask for them.
			hrefs := make(map[string]string, len(set.Locales()))
			for _, locale := range set.Locales() {
				if hrefs[locale], err = pageIn(app, set, rc, locale); err != nil {
					return layoutView{}, nil, err
				}
			}
			if href := hrefs[rc.Locale]; href != "" {
				rc.HoistLink("canonical", site.Origin+href)
			}
			for _, locale := range set.Locales() {
				href := hrefs[locale]
				if href != "" {
					rc.HoistAlternate(locale, site.Origin+href)
					if locale == site.Original {
						rc.HoistAlternate("x-default", site.Origin+href)
					}
				}
				if locale == rc.Locale {
					continue
				}
				if href == "" {
					// Not translated: the language's home page, rather than no
					// way into the language at all.
					if href, err = app.URL("home", locale, nil); err != nil {
						return layoutView{}, nil, err
					}
				}
				view.Languages = append(view.Languages, language{Code: locale, Name: ui.For(locale).Name, URL: href})
			}
			return view, nil, nil
		})).
		WithSlot("content", true, false).
		Build()
}

// pageIn is the address of the page being rendered in locale, or "" where it has
// none: a page with no path, like the not-found page, or a page of the
// documentation not translated into locale.
func pageIn(app *collage.App, set *site.Set, rc *collage.RenderContext, locale string) (string, error) {
	if rc.Page == nil {
		return "", nil
	}
	if rc.Page.Name == "doc" {
		translation := set.Site(locale)
		if translation == nil {
			return "", nil
		}
		if _, err := translation.Page(rc.Param("slug")); err != nil {
			return "", nil
		}
	}
	href, err := app.URL(rc.Page.Name, locale, rc.PathParams)
	if errors.Is(err, collage.ErrNoPathInLocale) || errors.Is(err, collage.ErrUnknownRoute) {
		return "", nil
	}
	return href, err
}
