package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Elagoht/collage/pkg/collage"
	"github.com/Elagoht/collage/pkg/collagetest"

	"shopping-list/list"
)

// client builds the application main builds, on a fresh list and a disk cache of
// its own, and returns a browser for it.
func client(t *testing.T) *collagetest.Client {
	t.Helper()
	list.Shopping = list.New("Milk", "Bread")
	app, err := newApp(collage.Env{CacheDir: t.TempDir()})
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	// Shut down before the directory is removed: an invalidation renders the
	// pages it reached again in the background, into that directory.
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })
	return collagetest.New(t, app.Handler())
}

// fetchAdd posts name the way static/list.js does: the form's fields, its token
// among them, and the header that asks for the fragment instead of a redirect.
func fetchAdd(c *collagetest.Client, page *collagetest.Response, name string) *collagetest.Response {
	body := url.Values{"_csrf": {page.CSRFToken()}, "name": {name}}
	req := c.Request(http.MethodPost, "/list", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(collage.FetchHeader, "1")
	return c.Do(req)
}

func TestListPage(t *testing.T) {
	page := client(t).Get("/list").WantStatus(http.StatusOK)
	for _, want := range []string{"Milk", "Bread", `data-src="/list/items"`} {
		if !strings.Contains(page.Body, want) {
			t.Errorf("the page has no %s:\n%s", want, page.Body)
		}
	}
}

func TestItemsFragment(t *testing.T) {
	items := client(t).Get("/list/items").WantStatus(http.StatusOK)
	if strings.Contains(items.Body, "<html") || !strings.Contains(items.Body, `id="items"`) {
		t.Errorf("/list/items is not the items fragment alone:\n%s", items.Body)
	}
}

func TestAddWithForm(t *testing.T) {
	c := client(t)
	res := c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Eggs"}}).WantStatus(http.StatusSeeOther)
	if res.Location() != "/list" {
		t.Errorf("Location = %q, want /list", res.Location())
	}
	if page := c.Get("/list"); !strings.Contains(page.Body, "Eggs") {
		t.Errorf("the page does not show the item just added:\n%s", page.Body)
	}
}

func TestAddWithFetch(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "Butter").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<html") || !strings.Contains(res.Body, `id="items"`) || !strings.Contains(res.Body, "Butter") {
		t.Errorf("a fetch was not answered with the items fragment:\n%s", res.Body)
	}
}

func TestAddIgnoresBlank(t *testing.T) {
	c := client(t)
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"   "}}).WantStatus(http.StatusSeeOther)
	res := fetchAdd(c, c.Get("/list"), "").WantStatus(http.StatusOK)
	if n := strings.Count(res.Body, "<li"); n != 2 {
		t.Errorf("the list has %d items after blank names, want 2:\n%s", n, res.Body)
	}
}

func TestItemsFragmentIsFresh(t *testing.T) {
	c := client(t)
	c.Get("/list/items")
	fetchAdd(c, c.Get("/list"), "Tea").WantStatus(http.StatusOK)
	if items := c.Get("/list/items"); !strings.Contains(items.Body, "Tea") {
		t.Errorf("/list/items does not show the item just added:\n%s", items.Body)
	}
}

func TestAddWithoutTokenIsRefused(t *testing.T) {
	c := client(t)
	req := c.Request(http.MethodPost, "/list", strings.NewReader("name=Jam"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Do(req).WantStatus(http.StatusForbidden)
	if items := c.Get("/list/items"); strings.Contains(items.Body, "Jam") {
		t.Errorf("a post without a token added an item:\n%s", items.Body)
	}
}

func TestItemNameIsEscaped(t *testing.T) {
	c := client(t)
	res := fetchAdd(c, c.Get("/list"), "<b>Salt</b>").WantStatus(http.StatusOK)
	if strings.Contains(res.Body, "<b>") || !strings.Contains(res.Body, "&lt;b&gt;Salt") {
		t.Errorf("the item's name was not escaped:\n%s", res.Body)
	}
}

func TestPageIsCachedUntilTheListChanges(t *testing.T) {
	c := client(t)
	c.Get("/list").WantStatus(http.StatusOK)
	if page := c.Get("/list"); page.Header.Get(collage.CacheStatusHeader) != "HIT" {
		t.Errorf("the second request was not served from the cache: %s = %q", collage.CacheStatusHeader, page.Header.Get(collage.CacheStatusHeader))
	}
	c.Submit(c.Get("/list"), "/list", url.Values{"name": {"Honey"}}).WantStatus(http.StatusSeeOther)
	if page := c.Get("/list"); !strings.Contains(page.Body, "Honey") {
		t.Errorf("the cached page still shows the old list:\n%s", page.Body)
	}
}
