// Package list keeps the shopping list. It stands in for a database: the pages
// read it and the action writes to it, and none of them know how it is stored.
package list

import (
	"strings"
	"sync"
)

// Tag is the dependency tag of everything rendered from the list. Invalidating
// it renders those pages again.
const Tag = "list"

// Item is one thing to buy.
type Item struct {
	Name string
}

// List is a shopping list several readers can add to at once.
type List struct {
	mu    sync.Mutex
	items []Item
}

// Shopping is the list the site shows.
var Shopping = New("Milk", "Bread")

// Returns a list holding names, in order.
func New(names ...string) *List {
	l := &List{}
	for _, name := range names {
		l.Add(name)
	}
	return l
}

// Returns the items, in the order they were added. The slice is a copy, so a
// template can range over it while someone else adds to the list.
func (l *List) Items() []Item {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Item(nil), l.items...)
}

// Adds name to the end of the list, and reports whether it did: a name that is
// blank once trimmed is not added.
func (l *List) Add(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, Item{Name: name})
	return true
}
