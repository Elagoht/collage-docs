package layouts

import (
	"context"
	"errors"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/site"
)

// layoutView is what templates/layouts/default.html renders. Its words come from
// the i18n catalogs, through {{t}}.
type layoutView struct {
	Lang string
	// Section is the part of the site the page is in, for the navigation to mark:
	// "docs" on a page of the documentation, "" elsewhere.
	Section string
	// Languages are the other languages the site is in, each linked to this page
	// in it — or to its home page, where this page is not translated.
	Languages []language
}

type language struct {
	Code string
	Name string
	URL  string
}

// Layout is the frame of every page. The canonical link and the alternates of
// each language are elagoht/meta's; the layout links the reader to the page in
// the site's other languages.
func Layout(app *collage.App, docs func() (*site.Set, error)) *collage.Fragment {
	return collage.NewFragment("layout", "layouts/default.html").
		WithData(collage.Load(func(_ context.Context, rc *collage.RenderContext) (layoutView, error) {
			view := layoutView{Lang: rc.Locale}
			if rc.Page != nil && rc.Page.Name == "doc" {
				view.Section = "docs"
			}
			set, err := docs()
			if err != nil {
				return layoutView{}, err
			}
			for _, locale := range set.Locales() {
				if locale == rc.Locale {
					continue
				}
				href, err := pageIn(app, set, rc, locale)
				if err != nil {
					return layoutView{}, err
				}
				if href == "" {
					// Not translated: the language's home page, rather than no
					// way into the language at all.
					if href, err = app.URL("home", locale, nil); err != nil {
						return layoutView{}, err
					}
				}
				view.Languages = append(view.Languages, language{Code: locale, Name: site.LanguageNames[locale], URL: href})
			}
			return view, nil
		})).
		WithSlot("content", collage.SlotRequired, collage.SlotSingle).
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
