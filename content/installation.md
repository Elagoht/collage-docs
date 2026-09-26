---
description: Install Go and the collage CLI, scaffold a project with collage new, and run it under collage dev.
---

# Installation

A collage project is an ordinary Go module. The `collage` command-line tool
scaffolds one, runs it while you work on it, and builds it when you ship it; the
framework itself is a library the project imports. Neither needs anything beyond
Go.

## Install Go

collage needs **Go 1.26 or newer**. Install it from [go.dev/dl](https://go.dev/dl)
or your package manager, and check:

```sh
go version
```

## Install the CLI

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

`go install` puts the binary in `$(go env GOBIN)`, or in `$(go env GOPATH)/bin`
when `GOBIN` is not set; that directory must be on your `PATH`. Check that it is:

```sh
collage version
```

The version it prints is read from the build, so it is whatever `go install`
fetched. `collage help` lists every command, and `collage help new` describes one.

## Create a project

```sh
collage new mysite
```

This writes a runnable project into `./mysite`, with the module path `mysite`, and
prints what to do next:

```text
Scaffolded "mysite" in mysite

Next steps:
  cd mysite
  go mod tidy
  cp .env.example .env.development
  collage dev
```

A minimal project (below) has no `.env.example`, so its steps leave out the `cp`.
`go mod tidy` fetches the framework. The scaffolded `go.mod` names only the module
and the Go version, so the first tidy is what records which collage release the
project is built against.

A few flags change where and how it is written. They take one dash or two, and may
come before or after the name:

| Flag | Effect |
| --- | --- |
| `--template minimal` | One layout around one page and a stylesheet — no demos. The default is `--template demo` |
| `-dir path` | Scaffold into `path` instead of `./<name>` |
| `-module path` | The module path in `go.mod`, such as `github.com/you/mysite`. Defaults to the name |
| `-force` | Scaffold into a directory that is not empty |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite --template minimal
collage new mysite -dir . -force
```

### With the demos, or without

The default template, `demo`, gives you a home page and a `/features` page of live demos: a
button posting to an API action that invalidates a cached page by tag, an HTML
form posting to its own page, a clock fragment opened at its own URL, and a JSON
document at `/healthz`. It is the quickest way to see each of those working, and
the code behind each is short enough to read in one sitting.

`--template minimal` gives you the same `main.go` and directory layout, and in it
only a layout around one page saying hello, with a stylesheet that has a dark
mode — nothing to delete. Use it when you are starting a real site.
[Your first page](/docs/your-first-page) starts from a minimal project.

## What the scaffold contains

A minimal project looks like this:

```text
mysite/
├── main.go                     configuration, the static mount, the CLI contract, plugin commands
├── routes.go                   every page, document and action — a new route goes here
├── go.mod
├── .gitignore
├── README.md
├── pages/
│   └── home.go                 the home page: layout, content, path, and its data
├── fragments/
│   └── layouts/main.go         the layout fragment every page shares, and the site's title
├── templates/
│   ├── layouts/default.html    the layout's HTML, with {{slot "content"}}
│   └── pages/home.html         <h1>Hello from {{.Name}}</h1>
└── static/
    └── app.css                 the background and text colour, light and dark
```

The demo project adds `main_test.go`, whose tests drive `app.Handler()` with no
server; a not-found page; `.env.example`, the variables for `collage dev` to copy;
`plugins-config.json`, plugin settings keyed by plugin name; a favicon; and
`actions/`, `documents/`, `store/` and `fragments/demo/`, with their templates and
scripts.

A few things in it are worth knowing before you change them.

**`main.go` keeps a contract with the CLI.** `collage dev` runs the program with
`COLLAGE_DEV=1`, and the `HOST` and `PORT` to listen on, in its environment, and
`collage export` runs it with `-collage-build -out <dir>`. The scaffolded
`main.go` honours all of it: the variable turns on development mode, the server
listens where it is told, and the flag renders the site to files instead of
serving it. A word after the flags — `go run . <command>` — runs a
[plugin's command](/docs/plugins). If you rewrite `main.go`, keep all of it
working, or those commands stop doing anything useful.
[The CLI reference](/docs/cli) has the details.

**`newApp` is separate from `main`.** It builds the whole application — the
configuration, the routes from `routes.go`, the `/static/` mount — and returns it
without starting a server. The demo project's `main_test.go` calls the same function and drives
`app.Handler()` with `net/http/httptest`, so the tests exercise the site that
actually runs rather than a second wiring of it. See [Testing](/docs/testing).

**Templates and static files are embedded.** `//go:embed all:templates` and
`//go:embed all:static` put them inside the binary, so it runs from any working
directory. In development the directories on disk win whenever they are there, so
an edited template still shows up on the next request.

**In development, static files are mounted with `os.OpenRoot`, not `os.DirFS`.**
An `os.Root` refuses a symlink that leads out of the directory; `os.DirFS` follows
it. In production the embedded copy is served, through `fs.Sub` so that its
`static/` directory is not a second path segment. See [Static assets](/docs/assets).

**Rendered pages are cached on disk in production**, under `.cache/`. In
development the page cache is never read, so an edit is never hidden behind a
stale page. See [Caching](/docs/caching).

## The environment file

The scaffold ships `.env.example`:

```sh
COLLAGE_CSRF_KEY=
PORT=3000
HOST=localhost
```

Copy it to `.env.development`, which is ignored by git:

```sh
cp .env.example .env.development
```

`collage dev` adds the variables in `.env.development` to the program's
environment — or those in `.env`, when there is no `.env.development`. It reads one
file, never both. The rules are short:

- A variable already set in your shell wins, so `PORT=4000 collage dev` works.
  `COLLAGE_DEV=1` is always set, whatever the file says.
- `HOST` and `PORT` are where `collage dev` itself listens, read once when it
  starts. The program is given a `HOST` and `PORT` of its own, on a loopback
  address — see [below](#run-it).
- The file holds `KEY=value` lines, `#` comments and blank lines. An `export `
  prefix and quotes around a value are allowed.
- A malformed line is reported with the file name and line number, rather than
  skipped: a skipped line is a setting you wrote and the program never saw. Until
  you fix it nothing is started or restarted — a build already running keeps
  serving — and `collage dev` keeps watching, so saving the fix carries on.
- No file at all is not an error. When there is one, its name is printed on
  stderr.

**Only `collage dev` reads these files.** `collage build`, `collage export` and the
built binary never do; in production the environment comes from wherever the
binary runs.

`COLLAGE_CSRF_KEY` signs the tokens forms carry. Empty is fine while you develop —
a key is generated for each process. Set one before you deploy anything with a
form, or every form submitted before a restart is refused after it. A cached page
with a form in it is tied to the key too: after a restart with a new one, that page
is rendered afresh rather than served with the old key's token in it — see
[Caching](/docs/caching#the-namespace). Make a key with:

```sh
openssl rand -hex 32
```

## Run it

```sh
collage dev
```

The site is at [http://localhost:3000](http://localhost:3000). That address is
`collage dev` itself: it listens on `HOST` and `PORT` as your program would read
them, and passes each request on to the program, which it runs on a loopback
address of its own. A request made while the program is starting waits for it.
Four things happen while it runs.

### Go changes are rebuilt

`collage dev` builds the project with `go build` and runs the binary. It watches
what the program is made of — `.go` files (tests aside), `go.mod`, `go.sum` and the
environment file — and on a change it rebuilds.

The new build is made first. Only once it compiles is the old process stopped,
gracefully, and the new one started. A change that does not compile leaves the
last good build serving and prints the compiler's error in the terminal, so a
typo never leaves you with nothing at `localhost:3000`.

A burst of saves is one rebuild. Hidden directories, `bin`, `dist`,
`node_modules`, `testdata` and `vendor` are never watched, so nothing the running
program writes can set off a rebuild of itself. And a program that exits by itself
— a panic at startup, a page whose template is missing — is not restarted in a
loop; your next change starts it again.

### Templates and static files are read from disk

In development mode every template is reparsed from disk before each render, and
static files are served from `static/` on disk. Editing either needs no rebuild,
and none happens.

### The browser reloads itself

Every page served in development carries a small script that reloads it when a
template or a static file changes, and when the program comes back from a
rebuild. Save a file and look at the browser; there is nothing to install.
Content your program reads from disk itself — Markdown, JSON — is neither a
template nor a static file, so name its directory in
[`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0) to have its
changes reload the page too. The script is never added to a production page, nor
to the answer to a form submission, which reloading would submit again.

The script keeps a stream open to the server, and a browser allows at most six
connections to one origin over HTTP/1.1, across all its tabs: with a stream in
every open development tab, six pages side by side held all six, and nothing else
loaded — not the seventh tab, not their own fragments. Since v0.20.0 every
development tab listens through one shared worker, `/_collage/reload-worker.js`,
which holds a single stream for all of them, however many pages are open. Where
there is no shared worker, each tab keeps its own stream and, as since v0.18.1,
closes it while hidden, reconnecting when it is seen again; if anything changed
while it was hidden, the page reloads then.

A stream that never closes is a page that never finishes loading, which is what a
screenshot tool or an end-to-end test waits for. A browser driven by Playwright,
Puppeteer or Selenium sets `navigator.webdriver`, and the script does not connect
there (since v0.18.0). For the tools that do not — headless Chrome's `--screenshot`
and `--dump-dom` — add `?collage-reload=0` to the URL and the page is served
without the script.

### Errors show up on the page

A fragment that fails in development does not quietly vanish. The page is served
with a panel over it naming the fragment and its error — for a template, with the
file and line — even when a fallback covered for it. When the whole page fails,
the built-in error page leads with the cause — the template, line and column of
the call that failed, and what it returned — then names the fragment where the
failure started and shows the full error chain, including a panic's stack. An
error page of your own gets the same panel on top, saying what it is standing in
for.

When there is no program to answer — it exited at startup, or the first build did
not compile — the page is a 503 showing what the program or the compiler printed,
instead of a refused connection. A page already open reloads onto it, and reloads
again once a change brings the program back. A program still not listening on the
address it was given 10 seconds after it started is named on that page, with the
address it was given.

None of that exists outside development. A production error page says one generic
sentence, because error messages carry hostnames, file paths and credentials, and
an error page is exactly the response most likely to hand them to a stranger. See
[Errors](/docs/errors).

## Editor support

For VS Code there is the Collage Snippets & Highlighter extension. It adds:

- **Snippets** for Go (`cpage`, `cfragd`, `caction`, `cplugin` …) and for templates
  (`clayout`, `cslot`, `cform` …).
- **Highlighting** of `{{ … }}` inside HTML. Templates stay HTML files, so Emmet,
  tag completion and formatting keep working.
- **Completion and hover** for every template function — collage's, its published
  plugins' and Go's builtins — with its signature and documentation.
- **Your project's names**: pages in `{{pageURL "…"}}` and their parameters,
  fragments in `{{fragmentURL "…" "…"}}`, slots in `{{slot "…"}}` and mounted
  files in `{{asset "…"}}`.
- **Diagnostics and go to definition**: a warning for a page, fragment or file that
  does not exist, and F12 from a name to the Go code declaring it.
- **`plugins-config.json` validation**: completion, descriptions and a warning for
  a misspelt key, for every published plugin and any plugin that ships a
  [`collage.json`](/docs/writing-plugins#editor-support-collagejson).

It learns your project's names by running `go run . collage-inspect` — what
[`collage inspect`](/docs/cli#collage-inspect) runs — in every folder whose `go.mod`
requires collage, and again whenever a Go file is saved. A project on collage
v0.27.0 or later answers it with nothing more to add.

The extension is not on the VS Code Marketplace yet. Download the `.vsix` from its
[latest release](https://github.com/Elagoht/collage-snippets-highlighter/releases/latest)
and install it:

```sh
code --install-extension collage-snippets-highlighter-<version>.vsix
```

| Setting | Default | What it does |
| --- | --- | --- |
| `collage.completions` | `auto` | Where completion and hover are offered: `auto`, only in a folder whose `go.mod` requires collage; `always`; or `never` |
| `collage.inspect` | `true` | Whether to run `go run . collage-inspect` to learn the project. That builds the application without serving it, so a program that connects to a database while it builds does so here too |
| `collage.diagnostics` | `warning` | How a name that does not exist is reported: `off`, `information`, `warning` or `error` |
| `collage.goCommand` | `go` | The go command to run |

Snippets and highlighting are always on. The extension's
[README](https://github.com/Elagoht/collage-snippets-highlighter#readme) lists every
snippet and the scopes to colour.

## Next

[Your first page](/docs/your-first-page) builds a page with data from nothing, in
a minimal project. When you are ready to ship, `collage build` makes a binary and
`collage export` writes static files — see [Deployment](/docs/deployment) and
[Static export](/docs/static-export).
