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

`go install` puts the binary in `$(go env GOPATH)/bin`, which must be on your
`PATH`. Check that it is:

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

`go mod tidy` fetches the framework. The scaffolded `go.mod` names only the module
and the Go version, so the first tidy is what records which collage release the
project is built against.

A few flags change where and how it is written. They may come before or after the
name:

| Flag | Effect |
| --- | --- |
| `-minimal` | One layout, an empty home page and a not-found page — no demos |
| `-dir path` | Scaffold into `path` instead of `./<name>` |
| `-module path` | The module path in `go.mod`, such as `github.com/you/mysite`. Defaults to the name |
| `-force` | Scaffold into a directory that is not empty |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite -minimal
collage new mysite -dir . -force
```

### With the demos, or without

Without `-minimal` you get a home page and a `/features` page of live demos: a
button posting to an API action that invalidates a cached page by tag, an HTML
form posting to its own page, a clock fragment opened at its own URL, and a JSON
document at `/healthz`. It is the quickest way to see each of those working, and
the code behind each is short enough to read in one sitting.

With `-minimal` you get the same `main.go`, the same directory layout and the same
kind of tests, with nothing in them to delete. Use it when you are starting a real
site. [Your first page](/docs/your-first-page) starts from a minimal project.

## What the scaffold contains

A minimal project looks like this:

```text
mysite/
├── main.go                     configuration, the static mount, the CLI contract, plugin commands
├── routes.go                   every page, document and action — a new route goes here
├── main_test.go                tests that drive app.Handler() with no server
├── go.mod
├── .env.example                variables for collage dev, to copy
├── .gitignore
├── plugins-config.json         plugin settings, keyed by plugin name
├── README.md
├── pages/
│   ├── home.go                 the home page: layout, content, path
│   └── not-found.go            the site-wide 404 page
├── fragments/
│   └── layouts/main.go         the layout fragment every page shares, and the site's title
├── templates/
│   ├── layouts/default.html    the layout's HTML, with {{slot "content"}}
│   └── pages/
│       ├── home.html
│       └── 404.html
└── static/
    ├── app.css
    └── favicon.svg
```

The demo project adds `actions/`, `documents/`, `store/` and `fragments/demo/`, and
their templates and scripts.

A few things in it are worth knowing before you change them.

**`main.go` keeps a contract with the CLI.** `collage dev` runs the program with
`COLLAGE_DEV=1` in its environment, and `collage export` runs it with
`-collage-build -out <dir>`. The scaffolded `main.go` reads both: the variable
turns on development mode, and the flag renders the site to files instead of
serving it. A word after the flags — `go run . <command>` — runs a
[plugin's command](/docs/plugins). If you rewrite `main.go`, keep all of it
working, or those commands stop doing anything useful.
[The CLI reference](/docs/cli) has the details.

**`newApp` is separate from `main`.** It builds the whole application — the
configuration, the routes from `routes.go`, the `/static/` mount — and returns it
without starting a server. `main_test.go` calls the same function and drives
`app.Handler()` with `net/http/httptest`, so the tests exercise the site that
actually runs rather than a second wiring of it. See [Testing](/docs/testing).

**Templates and static files are embedded.** `//go:embed all:templates` and
`//go:embed all:static` put them inside the binary, so it runs from any working
directory. In development the directories on disk win whenever they are there, so
an edited template still shows up on the next request.

**Static files are mounted with `os.OpenRoot`, not `os.DirFS`.** An `os.Root`
refuses a symlink that leads out of the directory; `os.DirFS` follows it. See
[Static assets](/docs/assets).

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
- The file holds `KEY=value` lines, `#` comments and blank lines. An `export `
  prefix and quotes around a value are allowed.
- A malformed line stops the command with the file name and line number, rather
  than being skipped: a skipped line is a setting you wrote and the program never
  saw.
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

The site is at [http://localhost:3000](http://localhost:3000). Three things happen
while it runs.

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
— a panic at startup, a port already in use — is not restarted in a loop; your
next change starts it again.

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

### Errors show up on the page

A fragment that fails in development does not quietly vanish. The page is served
with a panel over it naming the fragment and its error — for a template, with the
file and line — even when a fallback covered for it. When the whole page fails,
the built-in error page names the fragment where the failure started and shows
the full error chain, including a panic's stack — and an error page of your own
gets the same panel on top, saying what it is standing in for.

None of that exists outside development. A production error page says one generic
sentence, because error messages carry hostnames, file paths and credentials, and
an error page is exactly the response most likely to hand them to a stranger. See
[Errors](/docs/errors).

## Next

[Your first page](/docs/your-first-page) builds a page with data from nothing, in
a minimal project. When you are ready to ship, `collage build` makes a binary and
`collage export` writes static files — see [Deployment](/docs/deployment) and
[Static export](/docs/static-export).
