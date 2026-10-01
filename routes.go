package main

import (
	"fmt"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/documents"
	"github.com/Elagoht/collage-docs/pages"
	"github.com/Elagoht/collage-docs/site"
)

// register adds every page, document and action to app. A new route goes here.
func register(app *collage.App, docs func() (*site.Set, error)) error {
	if err := app.Register(
		pages.HomePage(app, docs), pages.DocPage(app, docs),
		documents.Robots(app), documents.Sitemap(app, docs), documents.Search(app, docs),
		documents.LLMs(app, docs), documents.LLMsFull(app, docs),
	); err != nil {
		return err
	}
	// Registered rather than given a path: it is reached by failing to match.
	// "collage export" writes it as 404.html.
	if err := app.RegisterNotFoundPage(pages.NotFoundPage(app, docs)); err != nil {
		return fmt.Errorf("register not-found page: %w", err)
	}
	return nil
}
