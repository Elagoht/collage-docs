// Package site reads the documentation's content — Markdown pages and the order
// they are read in — and turns it into what the templates render.
//
// Content is data, not templates: a page is a Markdown file under content/, and
// content/nav.json says which sections exist and in what order their pages come.
// Adding a page is adding a file and a line.
package site

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// ErrNoPage is returned for a slug no content file declares.
var ErrNoPage = errors.New("site: no such page")

// Site is the whole of the documentation in one language: every page, and the
// navigation that orders them.
type Site struct {
	// Locale is the language the site is written in.
	Locale   string
	Sections []Section
	// prefix is what the locale adds in front of a page's path: "/en" for the
	// English original, "/tr" for its translation into Turkish. Every language
	// has one; none is at the bare address.
	prefix string
	pages  map[string]*Page
	order  []*Page
}

// Section is one heading of the navigation and the pages under it.
type Section struct {
	Title string
	Pages []*Page
}

// Page is one document.
type Page struct {
	Slug        string
	Title       string
	Description string
	Section     string
	Body        template.HTML
	// Markdown is the page as written, below its title and front matter, with
	// every link to another page pointing where the site serves it — the text
	// llms-full.txt is made of.
	Markdown string
	Headings []Heading
	// References are the identifiers of package collage the page is about, as
	// "Name" or "Type.Method", linked to their Go reference.
	References []Reference
	// Parts are the page's text, split at its headings, for search.
	Parts      []Part
	Prev, Next *Page

	prefix string
	// outline is every heading below the title, in order, whatever its level:
	// what a translation's headings are held to.
	outline []Heading
}

// Part is one stretch of a page between headings: what search matches and where a
// result links to.
type Part struct {
	// ID is the heading's anchor, empty for the text before the first heading.
	ID string
	// Heading is the heading's text, empty likewise.
	Heading string
	// Text is the prose under it — paragraphs, lists, tables, inline code — with
	// the markup taken away. Code blocks are left out: they are long, and the
	// names in them are in the prose around them too.
	Text string
}

// ReferenceBase is the Go reference of package collage; an identifier's entry is
// its name as the fragment.
const ReferenceBase = "https://pkg.go.dev/github.com/Elagoht/collage/pkg/collage"

// Reference is one identifier a page links to in the Go reference.
type Reference struct {
	Name string
}

// URL is the identifier's entry in the Go reference.
func (r Reference) URL() string { return ReferenceBase + "#" + r.Name }

// Heading is one entry of a page's table of contents.
type Heading struct {
	ID    string
	Text  string
	Level int
}

// Set is the documentation in every language it is written in: the original, and
// its translations.
type Set struct {
	sites   map[string]*Site
	locales []string
}

// Site returns the documentation in locale, or nil for a locale it is not written
// in.
func (s *Set) Site(locale string) *Site { return s.sites[locale] }

// Locales returns every locale, the original's first.
func (s *Set) Locales() []string { return s.locales }

// LoadSet reads the original documentation from fsys, in locale original, and a
// translation of it into each of translations from the directory of that name.
//
// A translation is held to the original: each of its pages translates a page of
// the original and keeps its headings, level for level, so that an anchor is the
// same in every language; its nav.json gives the sections' titles, one per
// section of the original, and nothing else, because the order is the original's.
// A page not translated yet is not in the translation at all, and a link to it
// from a translated page goes to the original.
func LoadSet(fsys fs.FS, original string, translations ...string) (*Set, error) {
	base, err := load(fsys, original, nil)
	if err != nil {
		return nil, err
	}
	set := &Set{sites: map[string]*Site{original: base}, locales: []string{original}}
	for _, locale := range translations {
		sub, err := fs.Sub(fsys, locale)
		if err != nil {
			return nil, err
		}
		translated, err := load(sub, locale, base)
		if err != nil {
			return nil, err
		}
		set.sites[locale] = translated
		set.locales = append(set.locales, locale)
	}
	return set, nil
}

// Page returns the page with slug.
func (s *Site) Page(slug string) (*Page, error) {
	p, ok := s.pages[slug]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNoPage, slug)
	}
	return p, nil
}

// Pages returns every page in reading order.
func (s *Site) Pages() []*Page { return s.order }

// nav is content/nav.json.
type nav []struct {
	Title string   `json:"title"`
	Pages []string `json:"pages"`
}

// Load reads an English site from fsys, which holds nav.json and one <slug>.md
// per page.
//
// It is strict, because a documentation site's failure modes are silent ones: a
// page on disk the navigation never reaches, a navigation entry with no page, a
// link to a slug that does not exist. Each is an error here rather than a 404 a
// reader finds.
func Load(fsys fs.FS) (*Site, error) { return load(fsys, "en", nil) }

// load reads the site in locale from fsys: the original when base is nil, a
// translation of base otherwise.
func load(fsys fs.FS, locale string, base *Site) (*Site, error) {
	// Errors name files as they are under content/.
	dir := ""
	site := &Site{Locale: locale, prefix: "/" + locale, pages: make(map[string]*Page)}
	if base != nil {
		dir = locale + "/"
	}

	raw, err := fs.ReadFile(fsys, "nav.json")
	if err != nil {
		return nil, fmt.Errorf("site: %w", err)
	}
	var sections nav
	if err := json.Unmarshal(raw, &sections); err != nil {
		return nil, fmt.Errorf("site: %snav.json: %w", dir, err)
	}

	files, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return nil, err
	}
	onDisk := make(map[string]bool, len(files))
	for _, file := range files {
		onDisk[strings.TrimSuffix(file, ".md")] = true
	}

	if base != nil {
		// The translation's sections are the original's, retitled.
		if len(sections) != len(base.Sections) {
			return nil, fmt.Errorf("site: %snav.json has %d sections, the original has %d; a translation titles the original's sections, in order",
				dir, len(sections), len(base.Sections))
		}
		for slug := range onDisk {
			if _, ok := base.pages[slug]; !ok {
				return nil, fmt.Errorf("site: %s%s.md translates nothing: there is no %s.md", dir, slug, slug)
			}
		}
		translated := make(nav, len(sections))
		for i, section := range base.Sections {
			translated[i].Title = sections[i].Title
			for _, page := range section.Pages {
				if onDisk[page.Slug] {
					translated[i].Pages = append(translated[i].Pages, page.Slug)
				}
			}
		}
		sections = translated
	}

	// Where a link to another page goes: into this language when the page is in
	// it, to the original when it is not translated yet.
	linkTo := func(slug string) string {
		if onDisk[slug] || base == nil {
			return site.prefix + "/docs/" + slug + "/"
		}
		return base.prefix + "/docs/" + slug + "/"
	}

	md := newMarkdown()
	for _, section := range sections {
		out := Section{Title: section.Title}
		for _, slug := range section.Pages {
			if _, dup := site.pages[slug]; dup {
				return nil, fmt.Errorf("site: nav.json lists %q twice", slug)
			}
			source, err := fs.ReadFile(fsys, slug+".md")
			if err != nil {
				return nil, fmt.Errorf("site: nav.json lists %q: %w", slug, err)
			}
			var original *Page
			if base != nil {
				original = base.pages[slug]
			}
			page, err := parsePage(md, dir+slug, source, original, linkTo)
			if err != nil {
				return nil, err
			}
			page.Slug = slug
			page.prefix = site.prefix
			page.Section = section.Title
			site.pages[slug] = page
			site.order = append(site.order, page)
			out.Pages = append(out.Pages, page)
		}
		if len(out.Pages) > 0 {
			site.Sections = append(site.Sections, out)
		}
	}

	for slug := range onDisk {
		if _, listed := site.pages[slug]; !listed {
			return nil, fmt.Errorf("site: %s%s.md is in no section of nav.json, so nothing links to it", dir, slug)
		}
	}

	for i, page := range site.order {
		if i > 0 {
			page.Prev = site.order[i-1]
		}
		if i < len(site.order)-1 {
			page.Next = site.order[i+1]
		}
	}
	if err := site.checkLinks(base); err != nil {
		return nil, err
	}
	return site, nil
}

// docLink finds links into the documentation, in any language, with an optional
// fragment.
var docLink = regexp.MustCompile(`href="(/[a-z]{2})?/docs/([a-z0-9-]+)/?(#[^"]*)?"`)

// checkLinks reports the first link to a page that does not exist, or to a
// heading the page does not have. A link is into s, or into base, the original,
// when s is a translation; a link with no language prefix is refused, because no
// page is at a bare address.
func (s *Site) checkLinks(base *Site) error {
	for _, page := range s.order {
		for _, match := range docLink.FindAllStringSubmatch(string(page.Body), -1) {
			var into *Site
			switch {
			case match[1] == s.prefix:
				into = s
			case base != nil && match[1] == base.prefix:
				into = base
			default:
				return fmt.Errorf("site: %s links to %s/docs/%s, which is in no language of the site", page.Slug, match[1], match[2])
			}
			target, ok := into.pages[match[2]]
			if !ok {
				return fmt.Errorf("site: %s links to %s/docs/%s, which does not exist", page.Slug, match[1], match[2])
			}
			if anchor := strings.TrimPrefix(match[3], "#"); anchor != "" && !target.hasHeading(anchor) {
				return fmt.Errorf("site: %s links to %s/docs/%s#%s, which has no such heading", page.Slug, match[1], match[2], anchor)
			}
		}
	}
	return nil
}

// hasHeading reports whether the page has a heading with the id, at any level: an
// h4 is as much an anchor as the h2 and h3 the table of contents lists.
func (p *Page) hasHeading(id string) bool {
	for _, h := range p.outline {
		if h.ID == id {
			return true
		}
	}
	return false
}

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

// parsePage reads one page: a front matter block of "key: value" lines between
// "---" lines, then Markdown whose first heading is the title. name is the file's
// name under content/ without ".md", for errors.
//
// original is the page this one translates, or nil for an original; linkTo is
// where a link to /docs/<slug> in it goes.
func parsePage(md goldmark.Markdown, name string, source []byte, original *Page, linkTo func(slug string) string) (*Page, error) {
	slug := name
	page := &Page{Slug: slug}
	body := source
	if rest, ok := bytes.CutPrefix(source, []byte("---\n")); ok {
		front, after, found := bytes.Cut(rest, []byte("\n---\n"))
		if !found {
			return nil, fmt.Errorf("site: %s.md: front matter is not closed", slug)
		}
		for _, line := range strings.Split(string(front), "\n") {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			switch strings.TrimSpace(key) {
			case "title":
				page.Title = strings.TrimSpace(value)
			case "description":
				page.Description = strings.TrimSpace(value)
			case "reference":
				for _, name := range strings.Split(value, ",") {
					if name = strings.TrimSpace(name); name != "" {
						page.References = append(page.References, Reference{Name: name})
					}
				}
			}
		}
		body = after
	}

	doc := md.Parser().Parse(text.NewReader(body))
	if err := page.align(doc, body, name, original); err != nil {
		return nil, err
	}
	retarget(doc, linkTo)
	if original != nil && len(page.References) == 0 {
		page.References = original.References
	}
	// Removed after the walk, not during it: removing a node the walk is on cuts
	// the sibling link it would have followed next, and ends the walk there.
	var titles []ast.Node
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		heading, ok := n.(*ast.Heading)
		if !ok || !entering {
			return ast.WalkContinue, nil
		}
		label := headingText(heading, body)
		if heading.Level == 1 {
			if page.Title == "" {
				page.Title = label
			}
			// The title is the page's own <h1>, written by the template.
			titles = append(titles, heading)
			return ast.WalkSkipChildren, nil
		}
		id, _ := heading.AttributeString("id")
		idText, _ := id.([]byte)
		entry := Heading{ID: string(idText), Text: label, Level: heading.Level}
		page.outline = append(page.outline, entry)
		if heading.Level <= 3 {
			page.Headings = append(page.Headings, entry)
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	for _, title := range titles {
		title.Parent().RemoveChild(title.Parent(), title)
	}
	page.Parts = parts(doc, body)
	if page.Title == "" {
		return nil, fmt.Errorf("site: %s.md has no title", slug)
	}

	page.Markdown = markdownText(body, linkTo)

	var out bytes.Buffer
	if err := md.Renderer().Render(&out, body, doc); err != nil {
		return nil, fmt.Errorf("site: %s.md: %w", slug, err)
	}
	page.Body = template.HTML(out.String())
	return page, nil
}

// align gives a translation's headings the ids of the original's, so that an
// anchor means the same heading in every language and a link can switch
// language without losing its place. It refuses a translation whose headings
// below the title are not the original's, level for level: an id copied onto the
// wrong heading would be worse than none.
func (p *Page) align(doc ast.Node, source []byte, name string, original *Page) error {
	if original == nil {
		return nil
	}
	var headings []*ast.Heading
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering {
			if heading.Level > 1 {
				headings = append(headings, heading)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	for i, heading := range headings {
		if i >= len(original.outline) {
			return fmt.Errorf("site: %s.md has more headings than the original; %q has no counterpart", name, headingText(heading, source))
		}
		want := original.outline[i]
		if heading.Level != want.Level {
			return fmt.Errorf("site: %s.md: heading %d, %q, is level %d, but the original's, %q, is level %d",
				name, i+1, headingText(heading, source), heading.Level, want.Text, want.Level)
		}
		heading.SetAttributeString("id", []byte(want.ID))
	}
	if len(headings) < len(original.outline) {
		return fmt.Errorf("site: %s.md has %d headings below its title, the original has %d; the first missing one is %q",
			name, len(headings), len(original.outline), original.outline[len(headings)].Text)
	}
	return nil
}

// markdownLink is a Markdown link to another page, with an optional fragment.
var markdownLink = regexp.MustCompile(`\]\(/docs/([a-z0-9-]+)(#[^)\s]*)?\)`)

// markdownText is body without its title line, and with the links retarget
// would rewrite in the HTML rewritten the same way in the text.
func markdownText(body []byte, linkTo func(slug string) string) string {
	text := string(body)
	lines := strings.SplitN(strings.TrimLeft(text, "\n"), "\n", 2)
	if strings.HasPrefix(lines[0], "# ") && len(lines) == 2 {
		text = lines[1]
	}
	text = markdownLink.ReplaceAllStringFunc(text, func(link string) string {
		match := markdownLink.FindStringSubmatch(link)
		return "](" + linkTo(match[1]) + match[2] + ")"
	})
	return strings.TrimSpace(text) + "\n"
}

// retarget points every link to /docs/<slug> where linkTo says it goes, which is
// also where a link written without the page's trailing slash gets one.
func retarget(doc ast.Node, linkTo func(slug string) string) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		link, ok := n.(*ast.Link)
		if !ok || !entering {
			return ast.WalkContinue, nil
		}
		rest, ok := strings.CutPrefix(string(link.Destination), "/docs/")
		if !ok {
			return ast.WalkContinue, nil
		}
		slug, fragment, _ := strings.Cut(rest, "#")
		target := linkTo(slug)
		if fragment != "" {
			target += "#" + fragment
		}
		link.Destination = []byte(target)
		return ast.WalkContinue, nil
	})
}

// parts splits a page's top-level blocks at its headings, keeping the text of
// everything but code and raw HTML.
func parts(doc ast.Node, source []byte) []Part {
	current := Part{}
	var out []Part
	var text strings.Builder
	flush := func() {
		current.Text = strings.Join(strings.Fields(text.String()), " ")
		if current.Text != "" || current.Heading != "" {
			out = append(out, current)
		}
		text.Reset()
	}
	for block := doc.FirstChild(); block != nil; block = block.NextSibling() {
		switch node := block.(type) {
		case *ast.Heading:
			flush()
			id, _ := node.AttributeString("id")
			idText, _ := id.([]byte)
			current = Part{ID: string(idText), Heading: headingText(node, source)}
		case *ast.FencedCodeBlock, *ast.CodeBlock, *ast.HTMLBlock:
		default:
			text.WriteString(headingText(block, source))
			text.WriteByte(' ')
		}
	}
	flush()
	return out
}

// headingText is a heading's text as a reader sees it: its inline content with
// the markup taken away.
func headingText(heading ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(heading, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			// Table cells and list items are blocks inside a block; without a
			// space between them their words run together.
			if n.Type() == ast.TypeBlock {
				b.WriteByte(' ')
			}
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Text:
			b.Write(node.Segment.Value(source))
			if node.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(node.Value)
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

// URL is the page's address, with its locale's prefix and the trailing slash the
// site's pages are answered at: "/en/docs/caching/", "/tr/docs/caching/".
func (p *Page) URL() string { return p.prefix + p.Path() + "/" }

// Path is the page's path within its locale — the address without the locale's
// prefix, which is what a route matches.
func (p *Page) Path() string { return path.Join("/docs", p.Slug) }
