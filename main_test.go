package main

import (
	"encoding/json"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"html/template"
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
	rec := get(t, "/en/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /en/ = %d:\n%s", rec.Code, rec.Body.String())
	}
	if n := strings.Count(rec.Body.String(), "<title>"); n != 1 {
		t.Errorf("home has %d titles, want 1", n)
	}
}

// load is the documentation in every language, as the site loads it.
func load(t *testing.T) *site.Set {
	t.Helper()
	set, err := loadContent(content.FS)
	if err != nil {
		t.Fatalf("loadContent: %v", err)
	}
	return set
}

// eachPage calls fn for every page in every language.
func eachPage(t *testing.T, fn func(locale string, page *site.Page)) {
	t.Helper()
	set := load(t)
	for _, locale := range set.Locales() {
		for _, page := range set.Site(locale).Pages() {
			fn(locale, page)
		}
	}
}

// Every page of the documentation renders in every language, with its own title,
// in a document that says which language it is in.
func TestEveryDocRenders(t *testing.T) {
	eachPage(t, func(locale string, page *site.Page) {
		rec := get(t, page.URL())
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d", page.URL(), rec.Code)
			return
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<title>"+template.HTMLEscapeString(page.Title)+" — collage</title>") {
			t.Errorf("GET %s has no title %q", page.URL(), page.Title)
		}
		if !strings.Contains(body, `<html lang="`+locale+`">`) {
			t.Errorf("GET %s is not marked as %q", page.URL(), locale)
		}
	})
}

// The whole documentation is translated. A page added in English and not yet in
// Turkish fails this test rather than a reader, who would find a Turkish site
// with a page missing.
func TestEveryPageIsTranslated(t *testing.T) {
	set := load(t)
	for _, locale := range site.Translations {
		for _, page := range set.Site(site.Original).Pages() {
			if _, err := set.Site(locale).Page(page.Slug); err != nil {
				t.Errorf("%s.md has no %s translation: content/%s/%s.md", page.Slug, locale, locale, page.Slug)
			}
		}
	}
}

// A page names its canonical address and every language it is in, each at the
// address the host answers — with the trailing slash, not a redirect to it.
func TestCanonicalAndAlternates(t *testing.T) {
	for _, c := range []struct {
		path, canonical string
	}{
		{"/en/", "/en/"},
		{"/tr/", "/tr/"},
		{"/en/docs/caching/", "/en/docs/caching/"},
		{"/tr/docs/caching/", "/tr/docs/caching/"},
	} {
		body := get(t, c.path).Body.String()
		want := []string{`<link rel="canonical" href="` + site.Origin + c.canonical + `">`}
		rest := c.canonical[len("/en"):]
		for _, alt := range []struct{ lang, href string }{
			{"en", "/en" + rest}, {"x-default", "/en" + rest}, {"tr", "/tr" + rest},
		} {
			want = append(want, `<link rel="alternate" hreflang="`+alt.lang+`" href="`+site.Origin+alt.href+`">`)
		}
		for _, w := range want {
			if !strings.Contains(body, w) {
				t.Errorf("GET %s has no %s", c.path, w)
			}
		}
		if strings.Index(body, `rel="canonical"`) > strings.Index(body, `rel="alternate" hreflang`) {
			t.Errorf("GET %s declares its alternates before its canonical link", c.path)
		}
	}
}

// Every other spelling of a page redirects, in one hop, to the one the site
// links to: with its language's prefix and its trailing slash. The bare root is
// English.
func TestOneAddressPerPage(t *testing.T) {
	for from, to := range map[string]string{
		"/":                "/en/",
		"/docs/caching":    "/en/docs/caching/",
		"/docs/caching/":   "/en/docs/caching/",
		"/en":              "/en/",
		"/en/docs/caching": "/en/docs/caching/",
		"/tr":              "/tr/",
		"/tr/docs/caching": "/tr/docs/caching/",
		"/en/sitemap.xml":  "/sitemap.xml",
	} {
		rec := get(t, from)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != to {
			t.Errorf("GET %s = %d to %q, want 301 to %q", from, rec.Code, rec.Header().Get("Location"), to)
		}
	}
}

func TestUnknownDocIsNotFound(t *testing.T) {
	for _, path := range []string{"/en/docs/no-such-page/", "/tr/docs/no-such-page/"} {
		if rec := get(t, path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
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
	want := []string{"index.html", "en/index.html", "404.html", "robots.txt", "sitemap.xml", "search.json", "tr/index.html", "tr/search.json"}
	eachPage(t, func(_ string, page *site.Page) {
		want = append(want, filepath.Join(page.URL(), "index.html"))
	})
	for _, file := range want {
		if _, err := os.Stat(filepath.Join(out, file)); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}

// The sitemap lists every page in every language at its published address, with
// its translations as alternates.
func TestSitemap(t *testing.T) {
	rec := get(t, "/sitemap.xml")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /sitemap.xml = %d", rec.Code)
	}
	body := rec.Body.String()
	eachPage(t, func(locale string, page *site.Page) {
		if loc := "<loc>" + site.Origin + page.URL() + "</loc>"; !strings.Contains(body, loc) {
			t.Errorf("sitemap has no %s", loc)
		}
		if link := `hreflang="` + locale + `" href="` + site.Origin + page.URL() + `"`; !strings.Contains(body, link) {
			t.Errorf("sitemap has no alternate %s", link)
		}
	})
}

// The search index has an entry for every page, each linking to a page that
// exists, and none of them carries the text of a code block.
func TestSearch(t *testing.T) {
	set := load(t)
	for _, locale := range set.Locales() {
		index := "/search.json"
		if locale != site.Original {
			index = "/" + locale + index
		}
		searchIndex(t, index, set.Site(locale))
	}
}

func searchIndex(t *testing.T, index string, loaded *site.Site) {
	rec := get(t, index)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d", index, rec.Code)
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
			t.Errorf("%s has nothing for %s", index, page.URL())
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

	eachPage(t, func(locale string, page *site.Page) {
		for _, ref := range page.References {
			if !known[ref.Name] {
				t.Errorf("%s %s.md refers to collage.%s, which package collage does not declare", locale, page.Slug, ref.Name)
			}
		}
	})
}
