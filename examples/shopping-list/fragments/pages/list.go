package fragments

import (
	"context"

	"github.com/Elagoht/collage/pkg/collage"

	"shopping-list/list"
)

// Returns the list's items: templates/fragments/items.html. It reads the list on
// every render and tags what it rendered, so invalidating list.Tag renders it
// again.
func Items() *collage.Fragment {
	return collage.NewFragment("items", "fragments/items.html").
		WithData(collage.LoadTagged(func(context.Context, *collage.RenderContext) ([]list.Item, []string, error) {
			return list.Shopping.Items(), []string{list.Tag}, nil
		})).
		Build()
}

// Returns the list page's content: templates/pages/list.html, with items in its
// "items" slot.
func List(items *collage.Fragment) *collage.Fragment {
	return collage.NewFragment("list", "pages/list.html").
		WithTitle("Shopping list").
		WithSlotFragment("items", items).
		Build()
}
