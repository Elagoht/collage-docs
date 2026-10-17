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

Search needs nothing from you: `collage export` writes `search-index.json` from
every rendered page, and the box searches it in the browser, keeping to the
page's language. `collage dev` renders pages as they are asked for and has no
index, so there the box says search works on the export. `/llms.txt` and
`/llms-full.txt` are built from the English pages.

The words around the documentation — the header, the home page, the 404 page —
are in `locales/en.json` and `locales/tr.json`, and the search box's and the table
of contents' in `plugins-config.json`.

Translations live in `content/<locale>/`, one file per page of the original, with
the original's headings level for level; `content/<locale>/nav.json` titles the
sections. `go run ./tools/checkcontent` loads everything the way the site does.

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

## Plugins

Everything a plugin does, a plugin does here; `plugins.go` lists them and
`plugins-config.json` configures them.

| Plugin | Does |
| --- | --- |
| elagoht/i18n | the strings in `locales/` |
| elagoht/meta | the canonical link, the alternates of each language, Open Graph |
| elagoht/jsonld | structured data: the site, and each page as an article |
| elagoht/sitemap | `/sitemap.xml` |
| elagoht/robots | `/robots.txt` |
| elagoht/favicon | the icons and the web app manifest, from `static/icon.png` |
| elagoht/toc | each page's table of contents and reading time |
| elagoht/highlight | code colouring, light and dark |
| elagoht/search | the search box and its index |
| elagoht/htmlcheck | checks every page an export writes; an error fails it |
| elagoht/minimizer | minifies the pages, the stylesheet and the scripts |

The Markdown is read by `site/`, not by elagoht/markdown: it holds a translation
to its original heading for heading, so an anchor is the same in every language,
and refuses a broken link or a page the navigation never reaches.

## Publishing

Every push to `main` runs the tests, exports the site and publishes `dist/` to
GitHub Pages — see `.github/workflows/pages.yml`.
