# collage-docs

The documentation site for [collage](https://github.com/Elagoht/collage), built
with collage and published as static files.

## Writing

Every page is a Markdown file in `content/`, and `content/nav.json` orders them
into sections. A page starts with a front matter block and its title:

```markdown
---
description: One line, shown under the title and in search results.
reference: NewPage, PageBuilder.Static
---

# The page's title

## A section
```

`reference` lists the identifiers of package collage the page is about; they are
linked to their entries on pkg.go.dev at the end of the page, and the tests fail
on one the package does not declare.

Link to another page as `/docs/<slug>`, and to one of its headings as
`/docs/<slug>#<heading-id>`. The site refuses to start — and its tests fail — on a
link to a page or heading that does not exist, a page the navigation never lists,
or a navigation entry with no page.

Search needs nothing from you. `/search.json` is built from the same content, one
entry per section, with code blocks left out; `static/search.js` fetches it the
first time the search box is used and searches it in the browser.

## Running

```
cp .env.example .env.development
collage dev
```

Pages are read from `content/` on every request in development, and the browser
reloads when one changes.

```
go test ./...
collage export     # -> dist/
collage serve      # serves dist/ the way a static host would
```

`go run ./tools/chromacss > static/chroma.css` regenerates the code highlighting
styles.

## Publishing

Every push to `main` runs the tests, exports the site and publishes `dist/` to
GitHub Pages — see `.github/workflows/pages.yml`.
