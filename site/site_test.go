package site

import (
	"strings"
	"testing"
	"testing/fstest"
)

func files(pages map[string]string, nav string) fstest.MapFS {
	fsys := fstest.MapFS{"nav.json": {Data: []byte(nav)}}
	for name, body := range pages {
		fsys[name+".md"] = &fstest.MapFile{Data: []byte(body)}
	}
	return fsys
}

func TestLoad(t *testing.T) {
	s, err := Load(files(map[string]string{
		"one": "---\ndescription: First.\n---\n\n# One\n\n## Setup\n\nSee [two](/docs/two#details).\n",
		"two": "# Two\n\n## Details\n\n```go\nfunc main() {}\n```\n",
	}, `[{"title":"Start","pages":["one","two"]}]`))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	one, _ := s.Page("one")
	if one.Title != "One" || one.Description != "First." || one.Section != "Start" {
		t.Errorf("one = %+v", one)
	}
	if strings.Contains(string(one.Body), "<h1") {
		t.Error("the title heading is still in the body; the template writes it")
	}
	if len(one.outline) != 1 || one.outline[0].ID != "setup" {
		t.Errorf("headings = %+v", one.outline)
	}
	if one.Next == nil || one.Next.Slug != "two" || one.Prev != nil {
		t.Errorf("one's neighbours are wrong")
	}
	two, _ := s.Page("two")
	if !strings.Contains(string(two.Body), `<code class="language-go">`) {
		t.Errorf("code does not name its language, which elagoht/highlight needs: %s", two.Body)
	}
}

// A diagram is drawn for the page; the text a reader of Markdown gets is its
// caption, not the SVG.
func TestLoad_DiagramIsItsCaptionInMarkdown(t *testing.T) {
	s, err := Load(files(map[string]string{
		"one": "# One\n\nBefore.\n\n<figure class=\"diagram-figure\" data-search-ignore>\n<svg class=\"diagram\" viewBox=\"0 0 320 40\" role=\"img\" aria-labelledby=\"t\">\n<title id=\"t\">A box.</title>\n<rect class=\"d-a\" x=\"0\" y=\"0\" width=\"10\" height=\"10\" />\n</svg>\n<figcaption>The box holds the <code>slot</code>.</figcaption>\n</figure>\n\nAfter.\n",
	}, `[{"title":"Start","pages":["one"]}]`))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	one, _ := s.Page("one")
	if !strings.Contains(string(one.Body), "<svg") {
		t.Error("the page lost its diagram")
	}
	for _, bad := range []string{"<svg", "<rect", "<figure", "<figcaption"} {
		if strings.Contains(one.Markdown, bad) {
			t.Errorf("the Markdown has %s:\n%s", bad, one.Markdown)
		}
	}
	if !strings.Contains(one.Markdown, "Diagram: The box holds the `slot`.") || !strings.Contains(one.Markdown, "After.") {
		t.Errorf("the Markdown does not keep the caption in the diagram's place:\n%s", one.Markdown)
	}
}

// The failures a documentation site makes silently are errors here.
func TestLoad_Refusals(t *testing.T) {
	for name, c := range map[string]struct {
		pages map[string]string
		nav   string
		want  string
	}{
		"a page the navigation never reaches": {map[string]string{"one": "# One", "stray": "# Stray"}, `[{"title":"S","pages":["one"]}]`, "stray.md"},
		"a navigation entry with no page":     {map[string]string{"one": "# One"}, `[{"title":"S","pages":["one","missing"]}]`, "missing"},
		"a link to no page":                   {map[string]string{"one": "# One\n\n[x](/docs/nope)"}, `[{"title":"S","pages":["one"]}]`, "/docs/nope"},
		"a link to no heading":                {map[string]string{"one": "# One\n\n[x](/docs/one#nowhere)"}, `[{"title":"S","pages":["one"]}]`, "#nowhere"},
		"a page with no title":                {map[string]string{"one": "no heading"}, `[{"title":"S","pages":["one"]}]`, "no title"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(files(c.pages, c.nav)); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Load() = %v, want an error mentioning %q", err, c.want)
			}
		})
	}
}

// A table scrolls in a box of its own, so it never widens the page.
func TestLoad_TablesScroll(t *testing.T) {
	s, err := Load(files(map[string]string{
		"one": "# One\n\n| a | b |\n|---|---|\n| cell | two |\n",
	}, `[{"title":"S","pages":["one"]}]`))
	if err != nil {
		t.Fatal(err)
	}
	one, _ := s.Page("one")
	if !strings.Contains(string(one.Body), `<div class="table-scroll" tabindex="0"><table>`) ||
		!strings.Contains(string(one.Body), "</table></div>") {
		t.Errorf("the table is not in a scroll box: %s", one.Body)
	}
}

// translated is an original of two pages and a Turkish translation of the first.
func translated(trOne string) fstest.MapFS {
	fsys := files(map[string]string{
		"one": "---\nreference: New\n---\n\n# One\n\n## Setup\n\n### Details\n\nSee [two](/docs/two) and [setup](/docs/one#setup).\n",
		"two": "# Two\n\n## Usage\n",
	}, `[{"title":"Start","pages":["one","two"]}]`)
	fsys["tr/nav.json"] = &fstest.MapFile{Data: []byte(`[{"title":"Başlangıç"}]`)}
	fsys["tr/one.md"] = &fstest.MapFile{Data: []byte(trOne)}
	return fsys
}

func TestLoadSet(t *testing.T) {
	set, err := LoadSet(translated("# Bir\n\n## Kurulum\n\n### Ayrıntılar\n\nBkz. [iki](/docs/two) ve [kurulum](/docs/one#setup).\n"), "en", "tr")
	if err != nil {
		t.Fatalf("LoadSet: %v", err)
	}
	if got := set.Locales(); len(got) != 2 || got[0] != "en" || got[1] != "tr" {
		t.Errorf("Locales() = %v", got)
	}
	tr := set.Site("tr")
	if len(tr.Pages()) != 1 {
		t.Fatalf("the translation has %d pages, want the one translated", len(tr.Pages()))
	}
	one, err := tr.Page("one")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Page("two"); err == nil {
		t.Error("an untranslated page is in the translation")
	}
	if one.URL() != "/tr/docs/one/" || one.Path() != "/docs/one" || one.Section != "Başlangıç" {
		t.Errorf("one: URL %q, Path %q, Section %q", one.URL(), one.Path(), one.Section)
	}
	if got := one.outline; len(got) != 2 || got[0].ID != "setup" || got[1].ID != "details" || got[0].Text != "Kurulum" {
		t.Errorf("headings = %+v, want the original's ids with the translation's text", got)
	}
	body := string(one.Body)
	for _, want := range []string{`id="setup"`, `href="/en/docs/two/"`, `href="/tr/docs/one/#setup"`} {
		if !strings.Contains(body, want) {
			t.Errorf("body has no %s:\n%s", want, body)
		}
	}
	if len(one.References) != 1 || one.References[0].Name != "New" {
		t.Errorf("references = %v, want the original's", one.References)
	}
}

func TestLoadSet_Refusals(t *testing.T) {
	for _, c := range []struct {
		name, trOne, want string
	}{
		{"a missing heading", "# Bir\n\n## Kurulum\n", `the first missing one is "Details"`},
		{"an extra heading", "# Bir\n\n## Kurulum\n\n### Ayrıntılar\n\n## Fazla\n", `"Fazla" has no counterpart`},
		{"a heading at another level", "# Bir\n\n## Kurulum\n\n## Ayrıntılar\n", "is level 2, but the original's"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := LoadSet(translated(c.trOne), "en", "tr"); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("LoadSet() = %v, want an error mentioning %q", err, c.want)
			}
		})
	}

	t.Run("a page the original does not have", func(t *testing.T) {
		fsys := translated("# Bir\n\n## Kurulum\n\n### Ayrıntılar\n")
		fsys["tr/three.md"] = &fstest.MapFile{Data: []byte("# Üç\n")}
		if _, err := LoadSet(fsys, "en", "tr"); err == nil || !strings.Contains(err.Error(), "tr/three.md translates nothing") {
			t.Errorf("LoadSet() = %v", err)
		}
	})
	t.Run("sections that are not the original's", func(t *testing.T) {
		fsys := translated("# Bir\n\n## Kurulum\n\n### Ayrıntılar\n")
		fsys["tr/nav.json"] = &fstest.MapFile{Data: []byte(`[{"title":"A"},{"title":"B"}]`)}
		if _, err := LoadSet(fsys, "en", "tr"); err == nil || !strings.Contains(err.Error(), "has 2 sections, the original has 1") {
			t.Errorf("LoadSet() = %v", err)
		}
	})
}
