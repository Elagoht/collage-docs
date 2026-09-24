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
