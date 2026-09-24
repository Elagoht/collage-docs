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
	if len(one.Headings) != 1 || one.Headings[0].ID != "setup" {
		t.Errorf("headings = %+v", one.Headings)
	}
	if one.Next == nil || one.Next.Slug != "two" || one.Prev != nil {
		t.Errorf("one's neighbours are wrong")
	}
	two, _ := s.Page("two")
	if !strings.Contains(string(two.Body), `class="chroma"`) {
		t.Errorf("code was not highlighted: %s", two.Body)
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

// A page's text is split at its headings for search, code blocks left out.
func TestLoad_Parts(t *testing.T) {
	s, err := Load(files(map[string]string{
		"one": "# One\n\nIntro with `inline()` code.\n\n## Setup\n\nFirst **step**.\n\n```go\nfunc hidden() {}\n```\n\n- a list item\n\n| a | b |\n|---|---|\n| cell | two |\n",
	}, `[{"title":"S","pages":["one"]}]`))
	if err != nil {
		t.Fatal(err)
	}
	one, _ := s.Page("one")
	if len(one.Parts) != 2 {
		t.Fatalf("parts = %+v, want the intro and Setup", one.Parts)
	}
	if one.Parts[0].Heading != "" || one.Parts[0].Text != "Intro with inline() code." {
		t.Errorf("intro = %+v", one.Parts[0])
	}
	setup := one.Parts[1]
	if setup.ID != "setup" || setup.Heading != "Setup" {
		t.Errorf("setup = %+v", setup)
	}
	for _, want := range []string{"First step.", "a list item", "cell", "two"} {
		if !strings.Contains(setup.Text, want) {
			t.Errorf("setup text %q lacks %q", setup.Text, want)
		}
	}
	if strings.Contains(setup.Text, "hidden") {
		t.Errorf("setup text %q includes a code block", setup.Text)
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
	if got := one.Headings; len(got) != 2 || got[0].ID != "setup" || got[1].ID != "details" || got[0].Text != "Kurulum" {
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
