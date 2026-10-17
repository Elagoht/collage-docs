package pages

import (
	"context"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/fragments/layouts"
	fragments "shopping-list/fragments/pages"
	"shopping-list/list"
)

// Returns the shopping list page at /list, cached for an hour or until the list
// changes.
//
// Its items fragment is also opened at /list/items, so a script can fetch that
// part alone. Posting the form adds an item: a script that asks for it is
// answered with the items fragment, a browser on its own with a redirect back.
func List() *collage.Page {
	items := fragments.Items()
	return collage.NewPage("list").
		WithLayouts(layouts.Master()).
		WithContent(fragments.List(items)).
		WithPath("en", "/list").
		WithFragmentPath("en", "/list/items", items).
		Incremental(time.Hour).
		WithAction("POST", func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			list.Shopping.Add(rc.Request.PostFormValue("name"))

			result := collage.SeeOther("/list")
			if rc.Request.Header.Get(collage.FetchHeader) != "" {
				result = collage.RenderFragment(items)
			}
			result.InvalidateTags = []string{list.Tag}
			return result, nil
		}).
		Build()
}
