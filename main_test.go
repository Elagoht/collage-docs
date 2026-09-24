package main

import (
	"encoding/json"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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
	want := []string{"index.html", "404.html", "robots.txt", "sitemap.xml", "search.json"}
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

// The search index has an entry for every page, each linking to a page that
// exists, and none of them carries the text of a code block.
func TestSearch(t *testing.T) {
	rec := get(t, "/search.json")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /search.json = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	var entries []struct {
		Title string `json:"t"`
		URL   string `json:"u"`
		Text  string `json:"x"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("search.json does not decode: %v", err)
	}
	loaded, _ := site.Load(content.FS)
	seen := map[string]bool{}
	for _, entry := range entries {
		path, _, _ := strings.Cut(entry.URL, "#")
		seen[path] = true
		if strings.Contains(entry.Text, `WithSlotFragment("author", author)`) {
			t.Errorf("%s indexes a code block", entry.URL)
		}
	}
	for _, page := range loaded.Pages() {
		if !seen[page.URL()] {
			t.Errorf("search.json has nothing for %s", page.URL())
		}
	}
}

// Every identifier a page links to in the Go reference exists in the version of
// collage this site is built against, so no link lands on the top of the page
// instead of its entry.
func TestReferencesExist(t *testing.T) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/Elagoht/collage").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	dir := filepath.Join(strings.TrimSpace(string(out)), "pkg", "collage")
	fset := token.NewFileSet()
	var files []*ast.File
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}
	pkg, err := doc.NewFromFiles(fset, files, "github.com/Elagoht/collage/pkg/collage")
	if err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{}
	values := func(list []*doc.Value) {
		for _, value := range list {
			for _, name := range value.Names {
				known[name] = true
			}
		}
	}
	values(pkg.Consts)
	values(pkg.Vars)
	for _, fn := range pkg.Funcs {
		known[fn.Name] = true
	}
	for _, typ := range pkg.Types {
		known[typ.Name] = true
		values(typ.Consts)
		values(typ.Vars)
		for _, fn := range typ.Funcs {
			known[fn.Name] = true
		}
		for _, method := range typ.Methods {
			known[typ.Name+"."+method.Name] = true
		}
	}

	loaded, err := site.Load(content.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range loaded.Pages() {
		for _, ref := range page.References {
			if !known[ref.Name] {
				t.Errorf("%s.md refers to collage.%s, which package collage does not declare", page.Slug, ref.Name)
			}
		}
	}
}
