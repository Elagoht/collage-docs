package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elagoht/collage-docs/content"
	"github.com/Elagoht/collage-docs/site"
)

func get(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatalf("newApp() = %v", err)
	}
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestHomePage(t *testing.T) {
	rec := get(t, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d:\n%s", rec.Code, rec.Body.String())
	}
	if n := strings.Count(rec.Body.String(), "<title>"); n != 1 {
		t.Errorf("home has %d titles, want 1", n)
	}
}

// Every page of the documentation renders, with its own title.
func TestEveryDocRenders(t *testing.T) {
	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatalf("site.Load: %v", err)
	}
	for _, page := range loaded.Pages() {
		rec := get(t, page.URL())
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), rec.Code)
			continue
		}
		if !strings.Contains(rec.Body.String(), "<title>"+page.Title+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
	}
}

func TestUnknownDocIsNotFound(t *testing.T) {
	if rec := get(t, "/docs/no-such-page"); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

// The export writes the home page, every doc and the 404 page.
func TestExport(t *testing.T) {
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := staticBuild(app, out, false); err != nil {
		t.Fatalf("staticBuild: %v", err)
	}
	loaded, _ := site.Load(content.FS)
	want := []string{"index.html", "404.html", "robots.txt", "sitemap.xml"}
	for _, page := range loaded.Pages() {
		want = append(want, filepath.Join("docs", page.Slug, "index.html"))
	}
	for _, file := range want {
		if _, err := os.Stat(filepath.Join(out, file)); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}

// The sitemap lists every page at its published address.
func TestSitemap(t *testing.T) {
	rec := get(t, "/sitemap.xml")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /sitemap.xml = %d", rec.Code)
	}
	loaded, _ := site.Load(content.FS)
	for _, page := range loaded.Pages() {
		if loc := "<loc>" + site.Origin + page.URL() + "</loc>"; !strings.Contains(rec.Body.String(), loc) {
			t.Errorf("sitemap has no %s", loc)
		}
	}
}
