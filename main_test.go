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
		"/en/robots.txt":   "/robots.txt",
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

// export writes the site the way "collage export" does and returns where.
func export(t *testing.T) string {
	t.Helper()
	cacheDir = t.TempDir()
	app, err := newApp(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := staticBuild(app, out, false); err != nil {
		t.Fatalf("staticBuild: %v", err)
	}
	return out
}

// The export writes the home page, every doc, the 404 page and what the plugins
// serve: the sitemap, robots.txt, the search index and the icons.
func TestExport(t *testing.T) {
	out := export(t)
	want := []string{
		"index.html", "404.html", "robots.txt", "llms.txt", "llms-full.txt",
		"sitemap.xml", "search-index.json", "favicon.ico", "site.webmanifest",
		"en/index.html", "tr/index.html",
	}
	eachPage(t, func(_ string, page *site.Page) {
		want = append(want, filepath.Join(page.URL(), "index.html"))
	})
	for _, file := range want {
		if _, err := os.Stat(filepath.Join(out, file)); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}

// The export's root is the English home page itself, not a page that sends the
// reader on, and it names /en/ as its canonical address.
func TestExportRootIsHome(t *testing.T) {
	root, err := os.ReadFile(filepath.Join(export(t), "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(root)
	if !strings.Contains(body, `class="home"`) {
		t.Errorf("index.html is not the home page:\n%s", body)
	}
	if strings.Contains(body, `http-equiv="refresh"`) {
		t.Errorf("index.html refreshes to another page:\n%s", body)
	}
	if !strings.Contains(body, `rel="canonical"`) || !strings.Contains(body, site.Origin+"/en/") {
		t.Errorf("index.html names no canonical %s/en/", site.Origin)
	}
}

// A page of the documentation has its code coloured by elagoht/highlight, a table
// of contents and a reading time from elagoht/toc, and its tables in a box that
// scrolls.
func TestDocPagePlugins(t *testing.T) {
	body := get(t, "/en/docs/caching/").Body.String()
	for _, want := range []string{`class="chroma"`, `/_highlight/style.`, `class="toc"`, "min read", `class="table-scroll"`} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /en/docs/caching/ has no %s", want)
		}
	}
}

// The sitemap lists every page in every language at its published address, with
// every language it is in as an alternate.
func TestSitemap(t *testing.T) {
	rec := get(t, "/sitemap.xml")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /sitemap.xml = %d", rec.Code)
	}
	body := rec.Body.String()
	eachPage(t, func(locale string, page *site.Page) {
		if loc := "<loc>" + site.Origin + page.URL() + "</loc>"; !strings.Contains(body, loc) {
			t.Errorf("sitemap.xml has no %s", loc)
		}
		if link := `hreflang="` + locale + `" href="` + site.Origin + page.URL() + `"`; !strings.Contains(body, link) {
			t.Errorf("sitemap.xml has no alternate %s", link)
		}
	})
}

// robots.txt is at the root and names the sitemap at an address that answers.
func TestRobots(t *testing.T) {
	body := get(t, "/robots.txt").Body.String()
	if want := "Sitemap: " + site.Origin + "/sitemap.xml\n"; !strings.Contains(body, want) {
		t.Errorf("robots.txt has no %q:\n%s", want, body)
	}
}

// llms.txt lists every English page, absolute, and llms-full.txt carries all of
// them with no link left relative.
func TestLLMs(t *testing.T) {
	index := get(t, "/llms.txt")
	full := get(t, "/llms-full.txt")
	for _, rec := range []*httptest.ResponseRecorder{index, full} {
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/markdown") {
			t.Fatalf("status %d, Content-Type %q", rec.Code, rec.Header().Get("Content-Type"))
		}
	}
	if !strings.HasPrefix(index.Body.String(), "# collage\n\n> ") {
		t.Errorf("llms.txt does not open with a title and a summary:\n%s", index.Body.String()[:200])
	}
	for _, page := range load(t).Site(site.Original).Pages() {
		if link := "](" + site.Origin + page.URL() + "): "; !strings.Contains(index.Body.String(), link) {
			t.Errorf("llms.txt has no %s", link)
		}
		if source := "Source: " + site.Origin + page.URL() + "\n"; !strings.Contains(full.Body.String(), source) {
			t.Errorf("llms-full.txt has no %s", page.Slug)
		}
	}
	if strings.Contains(full.Body.String(), "](/") {
		t.Error("llms-full.txt has a relative link")
	}
}

// The export's search index has an entry for every page in its language, and
// none of them carries the text of a code block. It lists a page by its path
// without the trailing slash, which the host redirects to the page.
func TestSearch(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(export(t), "search-index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []struct {
		Path   string `json:"path"`
		Text   string `json:"text"`
		Locale string `json:"locale"`
	}
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("search-index.json does not decode: %v", err)
	}
	seen := map[string]string{}
	for _, entry := range entries {
		seen[entry.Path] = entry.Locale
		if strings.Contains(entry.Text, `WithSlotFragment("author", author)`) {
			t.Errorf("%s indexes a code block", entry.Path)
		}
	}
	eachPage(t, func(locale string, page *site.Page) {
		if got, ok := seen[strings.TrimSuffix(page.URL(), "/")]; !ok || got != locale {
			t.Errorf("search-index.json has %s as %q, want %q", page.URL(), got, locale)
		}
	})
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
