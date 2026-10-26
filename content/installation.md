---
description: Install Go and the collage CLI, scaffold a project with collage new, and run it under collage dev.
---

# Installation

A collage project is an ordinary Go module. The `collage` command-line tool
scaffolds one, runs it while you work on it, and builds it when you ship it; the
framework itself is a library the project imports. Neither needs anything beyond
Go.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 364" role="img" aria-labelledby="inst-flow-title">
<title id="inst-flow-title">From nothing to a running project. go install puts the collage CLI on your PATH. collage new mysite writes the minimal project by default, one layout around one page and a stylesheet, or with --template demo the demos, their tests and a .env.example. go mod tidy makes go.mod require collage and writes go.sum. collage dev then serves localhost:6060 and passes each request to the program on a loopback address of its own; a saved Go file is rebuilt and restarted, and a saved template or static file reloads the page.</title>
<defs><marker id="inst-flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="50" y="8" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">go install …/collage@latest</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="48" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="64" text-anchor="middle">collage new mysite</text>
<line class="d-flow" x1="160" y1="72" x2="83" y2="88" marker-end="url(#inst-flow-arrow)" />
<line class="d-flow" x1="160" y1="72" x2="237" y2="88" marker-end="url(#inst-flow-arrow)" />
<rect class="d-a" x="10" y="90" width="145" height="26" rx="4" /><text class="d-on-a" x="82" y="107" text-anchor="middle">minimal</text>
<rect class="d-b" x="165" y="90" width="145" height="26" rx="4" /><text class="d-on-b" x="237" y="107" text-anchor="middle">--template demo</text>
<text class="d-small" x="82" y="130" text-anchor="middle">the default: a layout,</text>
<text class="d-small" x="82" y="142" text-anchor="middle">a page, a stylesheet</text>
<text class="d-small" x="237" y="130" text-anchor="middle">demos and their tests,</text>
<text class="d-small" x="237" y="142" text-anchor="middle">a .env.example</text>
<line class="d-flow" x1="82" y1="148" x2="140" y2="164" marker-end="url(#inst-flow-arrow)" />
<line class="d-flow" x1="237" y1="148" x2="180" y2="164" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="166" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="182" text-anchor="middle">go mod tidy</text>
<text class="d-small" x="160" y="204" text-anchor="middle">go.mod requires collage, go.sum</text>
<line class="d-flow" x1="160" y1="208" x2="160" y2="222" marker-end="url(#inst-flow-arrow)" />
<rect class="d-cmd" x="50" y="224" width="220" height="24" rx="12" /><text class="d-cmd-text" x="160" y="240" text-anchor="middle">collage dev</text>
<line class="d-flow" x1="160" y1="248" x2="160" y2="262" marker-end="url(#inst-flow-arrow)" />
<rect class="d-frame" x="10" y="264" width="300" height="92" rx="8" />
<rect class="d-pill" x="80" y="272" width="160" height="22" rx="11" /><text x="160" y="287" text-anchor="middle">localhost:6060</text>
<text class="d-small" x="22" y="312">Go file saved: rebuild, restart</text>
<text class="d-small" x="22" y="328">template or static file: reload</text>
<text class="d-small" x="22" y="344">program: a loopback address of its own</text>
</svg>
<figcaption>From nothing to a running project. <code>go install</code> puts the CLI on your <code>PATH</code>. <code>collage new</code> writes the minimal project unless you ask for <code>--template demo</code>, which adds the demos, their tests and a <code>.env.example</code>. <code>go mod tidy</code> fetches the framework and records its release in <code>go.mod</code> and <code>go.sum</code>; <code>--with</code> does that for you. <code>collage dev</code> listens on port 6060 by default and passes each request on to your program, which it runs on a loopback address of its own: a saved Go file is rebuilt and restarted, and a saved template or static file only reloads the page.</figcaption>
</figure>

## Install Go

collage needs **Go 1.26.10 or newer** (since v0.72.0; with the default `GOTOOLCHAIN=auto`, an older Go fetches it by itself). Install it from [go.dev/dl](https://go.dev/dl)
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

### Shell completion and the man page

Since v0.77.0 the CLI completes its commands, flags and their values when TAB is
pressed. Load the script from your shell's startup file:

```sh
eval "$(collage completion bash)"      # ~/.bashrc
source <(collage completion zsh)       # ~/.zshrc, after compinit
collage completion fish | source       # ~/.config/fish/config.fish
```

`collage man --install` puts collage(1) where `man` looks, so `man collage` works.
Both are in [the CLI reference](/docs/cli#collage-completion).

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
  collage dev
```

A demo project (below) also has a `.env.example`, so its steps add
`cp .env.example .env.development` before `collage dev`.

`go mod tidy` fetches the framework. The scaffolded `go.mod` names only the module
and the Go version, so the first tidy is what records which collage release the
project is built against.

A few flags change where and how it is written. They take one dash or two, and may
come before or after the name:

| Flag | Effect |
| --- | --- |
| `--template demo` | The demos, their tests and a `.env.example`. The default, since v0.32.0, is `--template minimal`: one layout around one page and a stylesheet |
| `-dir path` | Scaffold into `path` instead of `./<name>` |
| `-module path` | The module path in `go.mod`, such as `github.com/you/mysite`. Defaults to the name |
| `-force` | Scaffold into a directory that is not empty |
| `--ai` | Write `AGENTS.md`, the rules a coding agent needs in a collage project, and a `CLAUDE.md` importing it (since v0.77.0); `collage ai init` does the same later — see [the CLI](/docs/cli#collage-ai) |
| `--with plugins` | Add plugins once it is written: bundles such as `@prod` and plugins such as `elagoht/session`, comma-separated (since v0.59.0). The plugins are fetched, so `go mod tidy` is no longer a next step |

```sh
collage new mysite -module github.com/you/mysite
collage new mysite --template demo
collage new mysite -dir . -force
collage new mysite --with @site,@prod
```

### With the demos, or without

`--template demo` gives you a home page and a `/features` page of live demos:

- a button posting to an API action that invalidates a cached page by tag;
- an HTML form posting to its own page;
- a clock fragment opened at its own URL;
- a JSON document at `/healthz`.

It is the quickest way to see each of those working, and the code behind each is
short enough to read in one sitting.

The default, `minimal`, gives you the same `main.go` and directory layout, and in it
only a layout around one page saying hello, with a stylesheet that has a dark
mode — nothing to delete. Use it when you are starting a real site.
[Your first page](/docs/your-first-page) starts from a minimal project.

## What the scaffold contains

A minimal project looks like this:

```text
mysite/
├── main.go                     newApp, the configuration; main hands it to collage.Main
├── embed.go                    files: templates/, static/ and plugins-config.json, embedded
├── embed_dev.go                files as the project directory, for collage dev's builds
├── routes.go                   every page, document and action — a new route goes here
├── plugins.go                  the plugin list; collage plugin add and del edit it
├── plugins-config.json         plugin settings, keyed by plugin name
├── .env.example                the environment variables, to copy to .env
├── go.mod
├── .gitignore
├── README.md
├── pages/
│   └── home.go                 the home page: its layout, content and path
├── fragments/
│   ├── layouts/master.go       Master(), the layout every page shares, and the site's title
│   └── pages/home.go           the home page's content, its data and its markup:
│                               <h1>Hello from {{.Name}}</h1>
├── templates/
│   └── layouts/master.html     the layout's HTML, with {{slot "content"}}
└── static/
    └── app.css                 the background and text colour, light and dark
```

Code is laid out by area. A page is `pages/<area>/<name>.go`, and its content
`fragments/pages/<area>/<name>.go` beside it, mirroring `pages/`. As a project
grows, the rest goes where the demo project puts it:

```text
pages/<area>/<name>.go            a page: its layouts, content, path, actions
fragments/layouts/master.go       Master(), the layout every page wraps itself in
fragments/pages/<area>/<name>.go  each page's content, mirroring pages/
actions/<area>.go                 action builders, one file per area
actions/funcs/<area>.go           their handlers
documents/<name>.go               routes that are not HTML
data/<domain>/                    state, by domain
templates/layouts/                the layouts' HTML
```

Since v0.77.0 a page's and a fragment's markup stays in its Go file, as a
`collage.InlineHTML` constant beside its data handler, and only layouts are
template files. Editing a page's markup is a Go change, which `collage dev`
rebuilds for; `collage add page --file` still gives a page a template file of its
own when you want one.

`collage add page blog/post` writes a page there and registers it in `routes.go`,
and `collage add` does the same for a fragment, an action or a document; see
[the CLI](/docs/cli#collage-add).

The demo project adds:

- `main_test.go`, whose tests drive `app.Handler()` with
  [`collagetest`](/docs/testing) and no server;
- a not-found page;
- a favicon;
- pages in the `demo` and `errors` areas, with their actions, a document, a
  counter in `data/count/`, and their script.

A few things in it are worth knowing before you change them.

**`main.go` keeps a contract with the CLI.** `collage dev` runs the program with
`COLLAGE_DEV=1`, and the `HOST` and `PORT` to listen on, in its environment, and
`collage export` runs it with `-collage-build -out <dir>`.

The scaffolded `main.go` is about fifty lines, because `main` only calls
`collage.Main(newApp)` (since v0.68.0), and `collage.Main` honours all of it. It
hands `newApp` a `collage.Env` — `Dev` set by the variable, the `Host` and `Port`
to listen on, and `CacheDir` — serves where it is told, and renders the site to
files when the flag is passed instead of serving it.

A word after the flags — `go run . <command>` — runs a
[plugin's command](/docs/plugins), and `collage.Main` dispatches that too. If you
replace it with a `main` of your own, keep all of it working, or those commands
stop doing anything useful. [The CLI reference](/docs/cli#the-contract-with-maingo) has the details.

**`plugins.go` is the plugin list.** It is a function, `plugins()`, that returns the
plugins in order, and `main.go` passes it once as `Plugins: plugins()`. It starts
empty. `collage plugin add elagoht/sitemap` writes an entry into it, and
`collage plugin del` removes one; see [Using plugins](/docs/plugins#registering-a-plugin).

**`newApp` is separate from `main`.** It builds the whole application — the
configuration, the routes from `routes.go`, the `/static/` mount — and returns it
without starting a server. The demo project's `main_test.go` calls the same function and drives
`app.Handler()` with `collagetest`, so the tests exercise the site that
actually runs rather than a second wiring of it. See [Testing](/docs/testing).

**Templates, static files and plugin settings are embedded.** One
`//go:embed all:templates all:static plugins-config.json` line, in `embed.go`,
puts them inside the binary as one filesystem, `files`, so it runs from any
working directory. `collage dev` builds with `-tags collage_dev`, which swaps
`embed.go` for `embed_dev.go` and embeds nothing: there `files` is the project
directory on disk, so an edited template still shows up on the next request, a
plugin given `files` reads the copy you are editing, and the Go build cache does
not gain a copy of them with every save (since v0.46.0; see
[collage dev](/docs/cli#development-builds-embed-nothing)).

`newApp` reads two things from `files` through the `Env`. `env.PluginConfig(files)`
is `plugins-config.json`: the file in the working directory when there is one, so
a deployment can change a setting without a rebuild, and otherwise the copy in the
binary. `env.Static(files, "static")` is what `/static/` serves.

**In development, static files are mounted with `os.OpenRoot`, not `os.DirFS`.**
`env.Static` does it: an `os.Root` refuses a symlink that leads out of the
directory; `os.DirFS` follows it. In production the embedded copy is served,
through `fs.Sub` so that its `static/` directory is not a second path segment. See
[Static assets](/docs/assets).

**Rendered pages are cached on disk in production**, under `.cache/`. In
development the page cache is never read, so an edit is never hidden behind a
stale page. See [Caching](/docs/caching).

## The environment file

Every project ships `.env.example` (a minimal one since v0.68.0). Without its
comments it is:

```sh
COLLAGE_CSRF_KEY=
PORT=6060
HOST=localhost
BASE_URL=http://localhost:6060
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
- `BASE_URL`, when neither your shell nor the file sets it, is `collage dev`'s own
  address, `http://localhost:6060` (since v0.68.0). A plugin that writes absolute
  URLs — a sitemap, a canonical link — refuses to start without an origin, and in
  development that is the right one.

**The program reads them too** (since v0.68.0). `collage.Main`, which the
scaffolded `main` calls, loads `.env` from the working directory before `newApp`
runs — `.env.development` first under `COLLAGE_DEV=1` — with the same rules. So
`os.Getenv("BASE_URL")` sees the file in `go run .`, `collage export` and the
built binary, with no `source .env`. On a server that sets its variables itself,
those win, and a missing file is no error.

`BASE_URL` is the address readers reach the site at, which a sitemap, a canonical
link or a feed writes in front of a path. Change it with `PORT`; in production it
is the site's own, `https://example.com`.

`COLLAGE_CSRF_KEY` signs the tokens forms carry. Empty is fine while you develop —
a key is generated for each process. Set one before you deploy anything with a
form, or every form submitted before a restart is refused after it.

A cached page
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

The site is at [http://localhost:6060](http://localhost:6060) (port 3000 before v0.32.0). That address is
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
typo never leaves you with nothing at `localhost:6060`.

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
changes reload the page too.

The script is never added to a production page, nor
to the answer to a form submission, which reloading would submit again.

The script keeps a stream open to the server, and a browser allows at most six
connections to one origin over HTTP/1.1, across all its tabs: with a stream in
every open development tab, six pages side by side held all six, and nothing else
loaded — not the seventh tab, not their own fragments.

Since v0.20.0 every
development tab listens through one shared worker, `/_collage/reload-worker.js`,
which holds a single stream for all of them, however many pages are open. Where
there is no shared worker, each tab keeps its own stream and, as since v0.18.1,
closes it while hidden, reconnecting when it is seen again; if anything changed
while it was hidden, the page reloads then.

A stream that never closes is a page that never finishes loading, which is what a
screenshot tool or an end-to-end test waits for.

A browser driven by Playwright,
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

## Example: setting up a project for a teammate

You start a project that someone else will work on too: the team's task list app. You scaffold it once,
commit it, and each of you runs it on your own machine with settings of your own.

Once, on your machine:

```sh
collage new team --template demo -module github.com/you/team
cd team
go mod tidy
```

The scaffolded `.gitignore` already draws the line between what is shared and what
is not. Its rules, without their comments:

```text
/bin/team
/bin/team.exe
/dist/
/.cache/
*.exe
/.env
/.env.*
!/.env.example
```

| Path | Committed | Why |
| --- | --- | --- |
| `go.mod`, `go.sum` | yes | They pin the collage release everyone builds against |
| `.env.example` | yes | The variables to copy, with no secret in them |
| The rest of what `collage new` wrote | yes | The project itself |
| `.env.development`, `.env` | no | Each person's own variables |
| `/bin/team` | no | What `collage build` writes |
| `/dist/` | no | What `collage export` writes |
| `/.cache/` | no | Rendered pages, kept between restarts |

Your teammate needs Go and the CLI, as above. After cloning the repository, they
run the same steps the demo's `README.md` lists:

```sh
cd team
go mod tidy
cp .env.example .env.development
collage dev
```

`go.mod` already requires a collage release, so this `go mod tidy` fetches that
one rather than picking another. A minimal project has no `.env.example`: write
one as in [the environment file](#the-environment-file) and commit it — the
`!/.env.example` line keeps it out of the ignored files.

**Each person's `.env.development` is their own.** `COLLAGE_CSRF_KEY` may stay
empty there, as above, or hold a key of that person's own. The key a deployed site
uses goes into the production environment's secrets, never into `.env.example`,
which everyone can read.

**Changing the port.** Someone whose port 6060 is taken sets another in their own
`.env.development`. In a demo project, move `BASE_URL` with it: it is the address
`collage dev` serves at, and what a sitemap, a canonical link or a feed puts in
front of a path.

```sh
PORT=6161
HOST=localhost
BASE_URL=http://localhost:6161
```

`collage dev` names the address and the file it read when it starts:

```text
01:09:37 - collage dev: serving http://localhost:6161
01:09:37 - collage dev: loaded .env.development
```

That address is settled once, at start. Saving a new `PORT` in the file rebuilds
the program, but `collage dev` keeps listening where it was until you stop it and
start it again. For one run, `PORT=4000 collage dev` does the same without
touching the file.

## Editor support

For VS Code there is the Collage Snippets & Highlighter extension. It adds:

- **Snippets** for Go (`cpage`, `cfragd`, `caction`, `cplugin` …) and for templates
  (`clayout`, `cslot`, `cform`, `cformv` …). Many Go ones — `cpagef`, `cactionf`,
  `cvalid`, `credirect`, `cdataf`, `cmeta` … — also add the file's `package` clause
  when it has none (the package of the files beside it, `main` beside `go.mod`, or
  the directory's name) and the imports they use, collage-validate's and
  collage-flash's among them.
- **Highlighting** of `{{ … }}` inside HTML. Templates stay HTML files, so Emmet,
  tag completion and formatting keep working. In Go, the template of an
  [inline fragment](/docs/fragments-and-slots#inline-templates) is coloured as HTML
  too, its arguments on one line or several, and so is a constant declared as
  `collage.InlineHTML`.
- **HTML inside Go edited as HTML** (since 0.6.0): in an inline template, tag and
  attribute completion, your HTML snippets, Emmet, hover, folding, the matching
  tag and tags closed as you type — and all of the below inside its `{{ … }}`.
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

It learns your project's names by building the project as `collage dev` does and
running it with `collage-inspect` — what
[`collage inspect`](/docs/cli#collage-inspect) does — in every folder whose `go.mod`
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
| `collage.inspect` | `true` | Whether to run the project with `collage-inspect` to learn it. That builds the application without serving it, so a program that connects to a database while it builds does so here too |
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
