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

// Site is the whole of the documentation: every page, and the navigation that
// orders them.
type Site struct {
	Sections []Section
	pages    map[string]*Page
	order    []*Page
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
	Headings    []Heading
	// Parts are the page's text, split at its headings, for search.
	Parts      []Part
	Prev, Next *Page
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

// Heading is one entry of a page's table of contents.
type Heading struct {
	ID    string
	Text  string
	Level int
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

// Load reads the site from fsys, which holds nav.json and one <slug>.md per page.
//
// It is strict, because a documentation site's failure modes are silent ones: a
// page on disk the navigation never reaches, a navigation entry with no page, a
// link to a slug that does not exist. Each is an error here rather than a 404 a
// reader finds.
func Load(fsys fs.FS) (*Site, error) {
	raw, err := fs.ReadFile(fsys, "nav.json")
	if err != nil {
		return nil, fmt.Errorf("site: %w", err)
	}
	var sections nav
	if err := json.Unmarshal(raw, &sections); err != nil {
		return nil, fmt.Errorf("site: nav.json: %w", err)
	}

	site := &Site{pages: make(map[string]*Page)}
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
			page, err := parsePage(md, slug, source)
			if err != nil {
				return nil, err
			}
			page.Section = section.Title
			site.pages[slug] = page
			site.order = append(site.order, page)
			out.Pages = append(out.Pages, page)
		}
		site.Sections = append(site.Sections, out)
	}

	files, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if _, listed := site.pages[strings.TrimSuffix(file, ".md")]; !listed {
			return nil, fmt.Errorf("site: %s is in no section of nav.json, so nothing links to it", file)
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
	if err := site.checkLinks(); err != nil {
		return nil, err
	}
	return site, nil
}

// docLink finds links into the documentation, with an optional fragment.
var docLink = regexp.MustCompile(`href="/docs/([a-z0-9-]+)(#[^"]*)?"`)

// checkLinks reports the first link to a page that does not exist, or to a
// heading the page does not have.
func (s *Site) checkLinks() error {
	for _, page := range s.order {
		for _, match := range docLink.FindAllStringSubmatch(string(page.Body), -1) {
			target, ok := s.pages[match[1]]
			if !ok {
				return fmt.Errorf("site: %s links to /docs/%s, which does not exist", page.Slug, match[1])
			}
			if anchor := strings.TrimPrefix(match[2], "#"); anchor != "" && !target.hasHeading(anchor) {
				return fmt.Errorf("site: %s links to /docs/%s#%s, which has no such heading", page.Slug, match[1], anchor)
			}
		}
	}
	return nil
}

func (p *Page) hasHeading(id string) bool {
	for _, h := range p.Headings {
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
// "---" lines, then Markdown whose first heading is the title.
func parsePage(md goldmark.Markdown, slug string, source []byte) (*Page, error) {
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
			}
		}
		body = after
	}

	doc := md.Parser().Parse(text.NewReader(body))
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
		if heading.Level <= 3 {
			id, _ := heading.AttributeString("id")
			idText, _ := id.([]byte)
			page.Headings = append(page.Headings, Heading{ID: string(idText), Text: label, Level: heading.Level})
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

	var out bytes.Buffer
	if err := md.Renderer().Render(&out, body, doc); err != nil {
		return nil, fmt.Errorf("site: %s.md: %w", slug, err)
	}
	page.Body = template.HTML(out.String())
	return page, nil
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

// URL is the page's address.
func (p *Page) URL() string { return path.Join("/docs", p.Slug) }
