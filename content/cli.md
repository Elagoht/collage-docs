---
description: Every command of the collage CLI — new, add, dev, build, export, serve, inspect, check, version and help — with its flags and exactly what it runs.
reference: DispatchCommands, Command, ErrUnknownCommand, InspectCommand, Inspection, App, Registrable
---

# The collage CLI

The `collage` command scaffolds projects and drives the ones you have: running one
in development, compiling the binary you deploy, and exporting it as static files.

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

It never links your application into itself — it cannot, because your application
is your code. `dev`, `build`, `export`, `inspect` and `check` run the `go` tool in the current
directory, exactly as you would by hand, and the rest of this page says precisely
what each one runs.

## Usage

```sh
collage <command> [flags]
```

| Command | What it does |
| --- | --- |
| `new` | Scaffold a new collage project |
| `add` | Write a page, fragment, action or document into the current project (since v0.40.0) |
| `dev` | Run the current directory's project in development mode |
| `build` | Compile the current directory's project into the binary you deploy |
| `export` | Render the current directory's project to static files |
| `serve` | Serve a static export the way a static host would |
| `inspect` | Print what the current directory's project is made of, as JSON |
| `check` | Check the current directory's project's templates without rendering them (since v0.40.0) |
| `version` | Print the collage CLI version |
| `help` | Show help for a command, or list every command |

`collage help <command>` prints that command's own usage, and so does
`collage <command> -h`. Flags use Go's `flag` syntax: `-out dist` and `-out=dist`
are the same, and `--out` works too.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success, including `help`, `collage -h` (since v0.11.0) and `collage <command> -h` |
| `1` | The command parsed correctly and failed to do its work |
| `2` | A usage problem: no command, an unknown command, a bad flag, or an unexpected argument |

## collage new

```sh
collage new <name> [--template minimal|demo] [--dir path] [--module path] [--force]
```

Scaffolds a new, runnable project named `<name>`.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--template name` | `minimal` | The project to scaffold: `minimal` or `demo` (since v0.14.2; the default is `minimal` since v0.32.0, `demo` before) |
| `-dir path` | `./<name>` | Directory to scaffold into |
| `-module path` | `<name>` | The module path written into `go.mod` |
| `-force` | off | Scaffold into a non-empty directory anyway |

```sh
collage new myblog                                   # into ./myblog, module "myblog", one page
collage new myblog --template demo                   # the demos, their tests, a .env.example
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
```

Flags take one dash or two, and may come before or after the name. Exactly one name is required; none, or
more than one, is a usage error. A target directory that exists and is not empty
is refused unless you pass `-force` — and with `-force`, files the scaffold writes
replace files of the same name.

**Every project** gets `go.mod`, a `main.go` holding the configuration, the static
mount, the CLI contract described [below](#the-contract-with-maingo) and the
dispatch of [plugin commands](#plugin-commands), a `routes.go` registering every
route, a layout that names the site with `WithTitle` and declares no slots (so a
page's own title replaces the site's), a home page that hands its template the
project's name with `WithData` — static without saying `Static()`, because nothing
in it fetches — `static/`, a `.gitignore` and a README.

**`--template demo`** adds a page of live demos — an action answering
JSON, a form posting to its own page, a fragment with its own URL, a JSON
document — a not-found page, tests for each written with
[`collagetest`](/docs/testing), `plugins-config.json`, a favicon and a
`.env.example`.

**The layout is by area** (since v0.40.0), the layout of a real application:

```text
pages/<area>/<name>.go            a page: its layouts, content, path, actions
fragments/layouts/main.go         Master(), the layout every page wraps itself in
fragments/pages/<area>/<name>.go  each page's content, mirroring pages/
actions/<area>.go                 action builders, one file per area
actions/funcs/<area>.go           their handlers
documents/<name>.go               routes that are not HTML
data/<domain>/                    state, by domain
templates/                        the HTML kept in files
```

`pages/` and `fragments/pages/` mirror each other by area, and every package in
them is called `pages` or `fragments`, so a page file imports its content as
`fragments "<module>/fragments/pages/<area>"` and `routes.go` imports each area's
pages under an alias, `demopages`. Small fragments keep their markup inline, as a
`collage.InlineHTML` const beside their data handler; the layout and the larger
pages keep theirs in `templates/`. An action a page attaches with `WithActionFor`
shares the page's name and has no path of its own. `routes.go` registers
everything with one [`app.Register`](/docs/pages-and-layouts#registration) call.

**The minimal project**, the default since v0.32.0, is the least a project can be: the layout around one
page, `<h1>Hello from {{.Name}}</h1>` — the project's name, handed to the
template with `WithData` — and a stylesheet that sets the background and text
colour, dark mode included. Nothing else — no tests, and no not-found page:
collage answers an unknown address with its own plain 404 until you register one.

When it is done it prints the next steps. A demo project, the one with a
`.env.example`, also gets `cp .env.example .env.development` before `collage dev`:

```sh
cd myblog
go mod tidy
collage dev
```

## collage add

```sh
collage add <page|fragment|action|document> <[area/]name> [flags]
```

Writes a page, a fragment, an action or a document in the scaffold's layout, and
registers it in `routes.go` (since v0.40.0):

```sh
collage add page blog/post          # pages/blog/post.go + fragments/pages/blog/post.go
collage add page blog/post --file   # its template in templates/pages/blog/post.html
collage add fragment blog/sidebar   # fragments/pages/blog/sidebar.go, for a slot
collage add action blog/comment     # actions/blog.go + actions/funcs/blog.go
collage add action blog/ping --path /api/ping
collage add document feed --path /feed.xml --type application/xml
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--file` | off | Keep the template in a file under the template root, not inline |
| `--path pattern` | `/<area>/<name>` | The URL; an action without one answers at the page it is attached to |
| `--name name` | the last segment | The page's, action's or document's name |
| `--locale code` | `Locale.Default`, or `en` | The locale its path is in |
| `--type type` | `text/plain; charset=utf-8` | A document's content type |
| `--dir path` | `.` | The project |

- **A page** is `pages/<area>/<name>.go`, wrapped in `layouts.Master()` — or the
  first layout in `fragments/layouts` that takes no arguments — with its content in
  `fragments/pages/<area>/<name>.go`: a template, a view struct, and a typed
  `collage.Load` handler that declares the page's title.
- **An action** gets a builder appended to `actions/<area>.go` and a handler to
  `actions/funcs/<area>.go`, either created when it is not there. With `--path` it
  answers at a URL of its own and is registered; without one it answers at the
  page it is attached to, and the command prints the `.WithActionFor(...)` to add.
- **A document** is `documents/<name>.go`.
- **A fragment** is a page's content without the page, for a slot; the command
  prints the `.WithSlotFragment(...)` that binds it.

A name is lowercase letters, digits and dashes. What the command writes is named
after the last segment — `blog/post` is the page `"post"`, built by `Post()` —
unless `--name` gives another. The locale of its path is `Locale.Default` as
`main.go` writes it, `en` when it is not a literal there, or `--locale`; a template
file goes under `Template.Root` with its `Extension`, read the same way. The
command says which locale it used.

It registers into the list `routes.go` already keeps: an `app.Register(...)` call,
where the new item goes after the last one of its kind, or a
`[]*collage.Page{...}` literal ranged over with `RegisterPage` — and the same for
`Document` and `Action`. It edits the file in place, so its comments stay where
they were, and imports the area's package under the same kind of alias the
scaffold uses. With no such list it changes nothing there and prints the line to
add. Constructors take no arguments: a page that needs a service gets it added by
hand.

**Nothing is overwritten.** Everything is worked out before anything is written,
and a file that exists, a page, action or document name the project already
declares, or an identifier already in the package a file goes into stops the
command with nothing written.

## collage dev

```sh
collage dev
```

Builds the project in the current directory, runs it with `COLLAGE_DEV=1` set,
and rebuilds and restarts it whenever its Go code changes. It takes no flags and
no arguments. Press Ctrl-C to stop; the program receives the interrupt too and
shuts down the way it would in production.

The browser talks to `collage dev`, not to the program: `collage dev` listens on
`HOST` and `PORT` as your program would read them (`localhost:6060` by default since v0.32.0, `localhost:3000` before),
and passes each request on to the program, which it starts with `HOST` and `PORT`
set to a loopback address of its own. That is how it can
[show errors in the browser](#errors-in-the-browser) when there is no program to
answer.

Since v0.34.0 `collage dev` answers only a request whose `Host` names this machine
— `localhost` or a name under it, an IP address — or the `HOST` it was started
with, and refuses anything else with a `403`. A page on another site can make its
own name resolve to `127.0.0.1` (DNS rebinding) and would then be same-origin with
`collage dev`: it could read your pages, the development error pages with their
stacks, and the program's output. The `Host` it sends is its own name, which it
cannot change. To reach `collage dev` under another name, start it with `HOST` set
to it.

What the program prints still reaches the terminal, as the program printed it —
with one change: its own address is replaced by `collage dev`'s, so its
`collage: listening` line names the address to open (since v0.32.0). On a colour
terminal the program is started with `FORCE_COLOR=1`, which the
[default logger](/docs/configuration#logger) honours, so its lines keep their time,
coloured markers and dimmed attributes although its output reaches `collage dev`
through a pipe; `collage dev`'s own lines take the same shape. `NO_COLOR` turns
both off.

```text
16:10:23 • collage dev: serving http://localhost:6060
16:10:24 • collage: listening  addr=localhost:6060
16:10:29 • collage dev: change detected, rebuilding
16:10:29 ✗ collage dev: build failed; the last good build is still serving
```

The scaffolded `main.go` turns on development mode when `COLLAGE_DEV=1` is set,
and listens on the `HOST` and `PORT` it is given.
Development mode reads templates and static files from disk on every request, so
editing those needs no rebuild and none happens — the page in your browser reloads
itself instead. Content your program reads from disk itself, such as Markdown,
reloads the page too once its directory is named in
[`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0). See
[Templates](/docs/templates#reloading-in-development) for what else development
mode changes.

### Development builds embed nothing

Since v0.46.0 `collage dev` builds with `go build -tags collage_dev`, and the
scaffold keeps its `//go:embed` lines in `embed.go`, which starts with
`//go:build !collage_dev`. Beside it, `embed_dev.go` (`//go:build collage_dev`)
declares the same two variables, empty. Development mode reads `templates/` and
`static/` from disk anyway, so a development build has nothing to embed, and
`collage build`, `collage export` and a plain `go build` pass no tag and embed as
before.

The reason is the Go build cache. A package that embeds files carries them in its
compiled form, and a change anywhere in your project recompiles `main`, so every
build that embeds them stores another copy of `templates/` and `static/` there —
one per save, kept for five days. A 30 MB `static/` grew the cache by about 30 MB
per save; a few hundred saves are gigabytes.

A project scaffolded before v0.46.0 still embeds in `main.go`, and `collage dev`
says so when it starts:

```text
16:10:23 ! collage dev: main.go embeds files into every development build, and each build stores another copy of them in the Go build cache; move the //go:embed lines into a file constrained with //go:build !collage_dev (see collage help dev)  patterns=all:static all:templates
```

To fix it, move the two `//go:embed` variables out of `main.go` (and `"embed"` out
of its imports) into two files of their own:

```go
// embed.go
//go:build !collage_dev

package main

import "embed"

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS
```

```go
// embed_dev.go
//go:build collage_dev

package main

import "embed"

var (
	templatesFS embed.FS
	staticFS    embed.FS
)
```

`go clean -cache` reclaims what is already in the cache. Your editor's `gopls`,
which builds without the tag, reports `embed_dev.go` as excluded by build
constraints when you open it; that is expected.

### What triggers a rebuild

| Watched | Not watched |
| --- | --- |
| `.go` files anywhere in the project | `_test.go` files |
| `go.mod` and `go.sum` at the project root | templates, static files and `DevWatch` directories (reloaded without a rebuild) |
| the environment file below | hidden directories (`.git`, `.cache`, …) |
| | `bin`, `dist`, `node_modules`, `testdata`, `vendor` |

The skipped directories are what keeps the loop from feeding itself: nothing the
running program writes — its cache, an export — can trigger a rebuild.

How a rebuild goes:

- **It polls** every 300 ms, comparing modification times and sizes. No
  file-system notification library, and the same behaviour on every platform.
- **A burst of writes is one rebuild.** After a change it waits until the files
  have been quiet for one interval, so a formatter rewriting several files costs
  one build.
- **The new build is made first.** `go build` writes a binary to a temporary
  directory outside the project; only once it compiles is the old process
  interrupted — it drains its requests as on Ctrl-C, and is killed only if it has
  not exited within 10 seconds — and the new one started.
- **A change that does not compile leaves the last good build serving**, with the
  compiler's error on screen.
- **A program that exits by itself** — a panic at startup, a page whose template
  is missing — is not restarted in a loop. The next change starts it again.

### Errors in the browser

Since v0.15.0, a program that is not running is a page, not a refused connection:

- **When the program exits** — a template not found at registration, a panic at
  startup — **or the first build fails**, every page is a 503 showing what the
  program or the compiler printed. A page already open reloads onto it, and
  reloads again once a change brings the program back.
- **A request made while the program starts waits for it**, rather than failing
  because the program is not listening yet.
- **A program that never listens where it was told** — a `main.go` that ignores
  `HOST` and `PORT` — is named on the page after 10 seconds, with the address it
  was given.

A render that fails once the program is running is the program's own development
error page, which leads with the cause — see [Errors](/docs/errors#rendering).

### Environment files

`collage dev` adds the variables of `.env.development` in the current directory to
the program's environment, or of `.env` when there is no `.env.development`. One
file, never both: `.env.development` replaces `.env` rather than being merged over
it.

```sh
# .env.development
PORT=6060
HOST=localhost
export COLLAGE_CSRF_KEY="0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
```

- The format is `KEY=value` lines, blank lines and lines starting with `#`. An
  `export ` prefix is allowed, and so is a pair of matching single or double quotes
  around a value, which are removed with nothing inside interpreted.
- An unquoted value ends at a `#` that follows whitespace, so `PORT=6060 # dev` is
  `6060`.
- Keys are letters, digits and underscores, not starting with a digit.
- **A malformed line stops the program from starting**: `collage dev` reports it
  with the file and the line (`.env.development:3`) and starts or restarts nothing
  until it is fixed — a program already running keeps running on the values it was
  started with. It keeps watching, so saving the corrected file carries on. A
  skipped line would be a setting you wrote and the program never saw.
- **A variable already set in the shell wins** over the file, so
  `PORT=4000 collage dev` still works.
- `COLLAGE_DEV=1` is always set, whatever the file says, and so are the `HOST` and
  `PORT` the program is to listen on. The file's `HOST` and `PORT` are where
  `collage dev` itself listens, read once when it starts.
- No file is not an error. When a file is read, its name is printed on stderr.
- The file is read again on every restart, and it is watched, so editing it
  restarts the program with the new values.

Only `collage dev` reads these files. `collage build`, `collage export` and the
built binary take their environment from wherever they run.

## collage build

```sh
collage build [-o path] [-os name] [-arch name] [-i]
```

Compiles the current directory's project into the binary you deploy. It runs:

```sh
CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build -trimpath -ldflags="-s -w" -o bin/<name> .
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o path` | `bin/<name>` | Where to write the binary |
| `-os name` | this machine's | Target operating system (`GOOS`) |
| `-arch name` | this machine's | Target architecture (`GOARCH`) |
| `-i` | off | Ask which extra files to write beside the binary |

`<name>` is the last element of the module path in `go.mod`; a directory with no
`go.mod`, or one that declares no module, is an error. For `-os windows`, `.exe` is
appended to the path when it is not already there. Positional arguments are a
usage error.

Why those settings:

- **CGO off**, because collage and the standard library need no C, and a static
  binary can go into an image with nothing else in it.
- **`-trimpath`**, so the binary does not carry the paths of the machine that built
  it.
- **`-s -w`** drops the debug tables, which is most of the size.
- **This machine by default** (since v0.32.0; linux/amd64 before), so the binary
  runs where it was built. A server is often something else — a binary built on a
  Mac does not run in a Linux container — so name it there:
  `collage build -os linux -arch amd64`.
- **`bin/`, not `dist/`**: `dist/` is where `collage export` writes, and
  `export -clean` empties it.

When it finishes it prints the binary's path, platform, size and build time.

### Extra files with -i

With `-i` it asks two questions before building, reading answers from standard
input — `y` or `yes` means yes, anything else no:

- **Write a Dockerfile?** A two-stage image: a `golang` build stage pinned to the
  major and minor version of the Go that built the CLI, and a
  `gcr.io/distroless/static-debian12` stage holding only the binary, with
  `HOST=0.0.0.0`, `PORT=8080`, port 8080 exposed and a commented-out
  `COLLAGE_CSRF_KEY`.
- **Write a systemd unit?** A `<name>.service` running `/usr/local/bin/<name>` from
  `/srv/<name>` with `HOST=127.0.0.1`, `PORT=8080`, `Restart=on-failure` and
  `TimeoutStopSec=30`, to adjust before installing.

Both are written beside the binary — `bin/Dockerfile`, `bin/<name>.service` —
because they are generated, and the project root is for what a person wrote. The
report prints the one command that placement costs:

```sh
docker build -f bin/Dockerfile .
```

**An existing file is never overwritten.** If one is already there the question is
not asked. These are files a project edits, and a build command that replaced one
with a default would quietly undo somebody's work.

Without `-i`, only the binary is written. See [Deployment](/docs/deployment).

## collage export

```sh
collage export [-out dir] [-clean]
```

Renders the current directory's project to static files — HTML for every page that
can be one, plus every mounted asset — for a static host. It runs:

```sh
go run . -collage-build -out <dir>          # plus -clean when you passed it
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-out dir` | `dist` | Directory the project renders into |
| `-clean` | off | Remove the directory's existing contents before building |

The program's output — the build report — is streamed straight through, not
reformatted. Positional arguments are a usage error.

A page that declares no strategy is exported when nothing it renders has a data
handler; one with a handler is exported when it says `Static()` or
`Incremental(ttl)`. A page whose path has a `{param}` is exported once per value
its `WithStaticParams` lists, and skipped without one. What gets exported, what is
skipped and why, and the safety checks on the output directory are in
[Static export](/docs/static-export). For a site with forms or per-request pages,
`collage build` is the one you want.

## collage serve

```sh
collage serve [-dir dir] [-host name] [-port n]
```

Serves a static export the way a static host would, so what you see is what you
will get after deploying it. It serves files; it does not run your project.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-dir dir` | `dist` | Directory to serve |
| `-host name` | `localhost` | Interface to listen on |
| `-port n` | `4000` | Port to listen on |

The port is 4000 rather than 6060 so it can run beside `collage dev` — which is
exactly when you compare the two.

It behaves like a static host rather than a file server:

- A path with no extension is answered with `<path>/index.html`, which is the
  shape an export writes.
- Directories are never listed.
- A path that resolves to nothing is answered with the export's own `404.html` and
  a 404 status, or a plain 404 when there is none.
- Only `GET` and `HEAD` are answered; anything else is a 405.
- A dotfile — `.env`, `.git/config` — is never served (since v0.34.0). A file
  whose extension names no type is typed from its content as a mount types it,
  never into HTML or XML, and every file carries `nosniff` (since v0.34.2).
- Every response is sent with `Cache-Control: no-store`, so re-exporting and
  reloading shows the new output rather than the old.

A directory that does not exist, or holds no files, is an error that tells you to
run `collage export` first. A path that exists but is not a directory is an error
too, saying just that.

## collage inspect

```sh
collage inspect
```

Prints what the current directory's project is made of, as JSON (since v0.27.0):
every page with its patterns, parameters, `layouts` (outermost first) and `guards`,
every fragment with its template and slots — an inline fragment with `"inline":
true` and no template path — the documents, the actions, the template functions,
the plugins, the locales, and the files the mounts serve. It takes no flags; positional arguments are a usage
error. It runs:

```sh
go run . collage-inspect
```

`collage-inspect` is `collage.InspectCommand`, a command `DispatchCommands` answers
itself rather than handing it to a plugin: it prints `App.Inspect()` — a
`collage.Inspection` — as indented JSON. The name is prefixed so that no plugin's
command is taken. The scaffolded `main.go` hands the word after its flags to
`DispatchCommands` (see [Plugin commands](#plugin-commands)), so a scaffolded
project on collage v0.27.0 or later needs nothing more. A program that does not
dispatch its arguments that way has nothing to answer with.

It is what an editor's completion reads: the Collage Snippets & Highlighter
extension for VS Code ([Editor support](/docs/installation#editor-support)) offers
page names in `{{pageURL "…"}}`, slots in `{{slot "…"}}` and files in
`{{asset "…"}}` from it.

## collage check

```sh
collage check [-json]
```

Checks every template's links without rendering anything (since v0.40.0). A link
built by name — `{{pageURL "post" "slug" .Slug}}`, `{{pageURLIn "en" "about"}}`,
`{{actionURL "logout"}}`, `{{fragmentURL "home" "clock"}}`, `{{localeURL "en"}}` —
fails when the template renders, and only on the page that reaches it. `check`
finds them all at once:

```text
$ collage check
error [unknown-route] inline template of fragment "hello":4:16: {{pageURL "featurs"}}: collage: no page or document by that name: "featurs"; did you mean "features"?
```

| Rule | What it means |
| --- | --- |
| `unknown-route` | No page, document, action or fragment path by that name; the closest registered name is suggested |
| `route-params` | The parameters do not fill the route's pattern — one missing, one it has no placeholder for, or not in name and value pairs |
| `unreachable-locale` | A locale no URL can carry: not `Locale.Default`, nor in `Locale.Supported` |
| `no-path-in-locale` | The route has no path in the locale `pageURLIn` or `fragmentURLIn` names |

It checks with the framework's own URL builders, so what it reports is exactly
what a render would fail on. Only names written as string literals are checked: a
name from a field, `{{pageURL .Name}}`, is known only when the template renders,
and a parameter whose name is not a literal leaves the parameters unchecked. A
link with no locale of its own passes when its route can be built in some locale,
since a render falls back to the default one.

It exits `1` when it finds anything, so it can stand in CI before `collage build`;
`-json` prints the findings as an array of `{level, rule, message}` for an editor.
It runs:

```sh
go run . collage-check
```

which `collage.DispatchCommands` answers by starting the application and calling
`App.Check` — so a program that opens a database on start opens it here too. That
needs the project's collage at v0.40.0 or later; an earlier one answers
`unknown command: "collage-check"`. A test can call the same function:

```go
if findings := app.Check(); len(findings) > 0 {
	t.Errorf("broken links: %v", findings)
}
```

## collage version

```sh
collage version
```

Prints `collage version <version>`. The version is read from the binary's build
information, so `go install ...@v0.9.0` reports `0.9.0`, and a binary built in a git
checkout reports a pseudo-version — `0.11.1-0.<timestamp>-<commit>` for a commit
after the v0.11.0 tag.
Only a binary with no version information at all — built outside a repository, or
with `-buildvcs=false` — reports `devel`.

## collage help

```sh
collage help            # every command
collage help export     # one command's usage
```

With no argument it lists every command and exits `0`. With a command name it
prints that command's usage; an unknown name prints
`collage: unknown command: "nope"` and exits `2`.

## The contract with main.go

`dev`, `export` and `inspect` depend on three things your `main.go` does, and the
scaffolded one does all three. The third, handing the words after its flags to
`collage.DispatchCommands`, is also what runs your plugins' commands:

| Command | Runs | Your `main.go` must |
| --- | --- | --- |
| `collage dev` | `go build -tags collage_dev`, then the binary, with `COLLAGE_DEV=1` and the `HOST` and `PORT` to listen on | turn on development mode when `COLLAGE_DEV` is `1`, and listen on `HOST` and `PORT` |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | parse `-collage-build`, `-out` and `-clean`, and on `-collage-build` render to `<dir>` instead of serving |
| `collage inspect` | `go run . collage-inspect` | pass the words after its flags to `collage.DispatchCommands` |

`collage build` needs nothing from `main.go`: compiling is something `go build`
does without being told anything.

The scaffolded version, trimmed to the part that matters:

```go
func main() {
	buildFlag := flag.Bool("collage-build", false, "render the app to static files instead of serving it")
	outFlag := flag.String("out", "dist", "output directory for -collage-build")
	cleanFlag := flag.Bool("clean", false, "remove -out's existing contents before building")
	portFlag := flag.Int("port", envInt("PORT", 6060), "port to listen on (env PORT)")
	flag.Parse()

	devMode := os.Getenv("COLLAGE_DEV") == "1"

	app, err := newApp(devMode, *portFlag)
	if err != nil {
		log.Fatal(err)
	}

	// A word after the flags is a plugin's command: go run . <command>
	if args := flag.Args(); len(args) > 0 {
		code, err := collage.DispatchCommands(context.Background(), app, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(code)
	}

	if *buildFlag {
		if err := staticBuild(app, *outFlag, *cleanFlag); err != nil {
			log.Fatalf("static build: %v", err)
		}
		return
	}

	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

If you rewrite `main.go`, keep all of it working, or `collage dev`,
`collage export`, `collage inspect` and your plugins' commands stop doing anything
useful in your project.

## Plugin commands

A plugin can contribute a command through `Host.RegisterCommand`. **The `collage`
CLI does not run it.** The CLI never loads your application, so a command that
only exists once your plugins have started is out of its reach, and
`collage <plugin-command>` is an unknown command.

Your own program dispatches them, with `collage.DispatchCommands`, and the
scaffolded `main.go` does (since v0.10.0 — a project scaffolded earlier can copy
the block [above](#the-contract-with-maingo)). After the flags are parsed and the
application is built, a word left over is a command:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
```

`DispatchCommands` starts the application — running every plugin's `Init`, which
is what registers their commands — and runs the command named by the first
argument. So a plugin's command runs as your program's subcommand, after any
flags:

```sh
go run . pages
./bin/myblog -port 4000 pages
```

The exit code follows the CLI's: `0` for success; `1` for a startup failure, a
command that ran and failed, or a command with no `Run`; `2` for a nil app, no
arguments, or a word no plugin registered (`ErrUnknownCommand`). An
unclaimed word is a usage error rather than a server started by accident. One word
no plugin registers is answered all the same: `collage-inspect`, which
[`collage inspect`](#collage-inspect) runs. A
program that would rather serve when no command matches can check
`errors.Is(err, collage.ErrUnknownCommand)` and carry on instead of exiting.
`app.Commands()` lists what the plugins registered, if you want to print your own
usage. Dispatching starts the application, which closes registration, and a later
`ListenAndServe` reuses that start. Writing such a command is covered in
[Writing a plugin](/docs/writing-plugins#commands).
