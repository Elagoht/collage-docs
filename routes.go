package main

import (
	"fmt"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/documents"
	"github.com/Elagoht/collage-docs/pages"
	"github.com/Elagoht/collage-docs/site"
)

// register adds every page, document and action to app. A new route goes here.
func register(app *collage.App, docs func() (*site.Site, error)) error {
	for _, page := range []*collage.Page{pages.HomePage(docs), pages.DocPage(docs)} {
		if err := app.RegisterPage(page); err != nil {
			return fmt.Errorf("register page %q: %w", page.Name, err)
		}
	}
	for _, document := range []*collage.Document{documents.Robots(), documents.Sitemap(app, docs), documents.Search(app, docs)} {
		if err := app.RegisterDocument(document); err != nil {
			return fmt.Errorf("register document %q: %w", document.Name, err)
		}
	}
	// Registered rather than given a path: it is reached by failing to match.
	// "collage export" writes it as 404.html.
	if err := app.RegisterNotFoundPage(pages.NotFoundPage()); err != nil {
		return fmt.Errorf("register not-found page: %w", err)
	}
	return nil
}
