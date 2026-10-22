---
description: Every command of the collage CLI — new, add, dev, build, export, serve, inspect, check, plugin, version and help — with its flags and exactly what it runs.
reference: Main, Env, Env.PluginConfig, Env.Static, DispatchCommands, Command, ErrUnknownCommand, FragmentBuilder.WithoutTypeCheck, App, Registrable
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
| `add` | Write a page, fragment, action, document, layout, guard, middleware or command into the current project (since v0.40.0) |
| `dev` | Run the current directory's project in development mode |
| `build` | Compile the current directory's project into the binary you deploy |
| `export` | Render the current directory's project to static files |
| `serve` | Serve a static export the way a static host would |
| `routes` | List every URL the current directory's project answers, with the guards each meets (since v0.69.0) |
| `inspect` | Print what the current directory's project is made of, as JSON |
| `check` | Check the current directory's project's templates without rendering them (since v0.40.0) |
| `plugin` | Add a plugin to the project, remove one, or check a plugin repository (since v0.58.0) |
| `version` | Print the collage CLI version |
| `help` | Show help for a command, or list every command |

`collage help <command>` prints that command's own usage, and so does
`collage <command> -h`. Flags use Go's `flag` syntax: `-out dist` and `-out=dist`
are the same, and `--out` works too.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 380" role="img" aria-labelledby="cli-life-title">
<title id="cli-life-title">A project's life through the collage commands. collage new scaffolds the project in this directory. While you work on it: collage add writes a page, fragment, action or document; collage dev runs it on localhost:6060 and rebuilds it when its Go code changes; collage check checks its templates' links; collage plugin add and del edit plugins.go and go.mod; collage inspect prints the project as JSON. To ship it, one of two roads. collage build compiles one static binary into bin/, and with -i writes an image file and a service unit beside it; the binary goes to a server, where it reads HOST and PORT. Or collage export renders HTML and static files into dist/; collage serve shows dist/ on localhost:4000, and a static host serves the same files.</title>
<defs><marker id="cli-life-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="100" y="8" width="120" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">collage new</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#cli-life-arrow)" />
<rect class="d-frame" x="6" y="50" width="308" height="122" rx="8" />
<text x="16" y="67">the project in this directory</text>
<rect class="d-cmd" x="14" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="62" y="92" text-anchor="middle">collage add</text>
<rect class="d-cmd" x="112" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="160" y="92" text-anchor="middle">collage dev</text>
<rect class="d-cmd" x="210" y="76" width="96" height="24" rx="12" /><text class="d-cmd-text" x="258" y="92" text-anchor="middle">collage check</text>
<text class="d-small" x="62" y="114" text-anchor="middle">writes Go files</text>
<text class="d-small" x="160" y="114" text-anchor="middle">runs it on :6060</text>
<text class="d-small" x="258" y="114" text-anchor="middle">checks its links</text>
<rect class="d-cmd" x="14" y="124" width="146" height="24" rx="12" /><text class="d-cmd-text" x="87" y="140" text-anchor="middle">collage plugin add/del</text>
<rect class="d-cmd" x="170" y="124" width="136" height="24" rx="12" /><text class="d-cmd-text" x="238" y="140" text-anchor="middle">collage inspect</text>
<text class="d-small" x="87" y="162" text-anchor="middle">edits plugins.go, go.mod</text>
<text class="d-small" x="238" y="162" text-anchor="middle">prints it as JSON</text>
<line class="d-flow" x1="84" y1="172" x2="84" y2="186" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="236" y1="172" x2="236" y2="186" marker-end="url(#cli-life-arrow)" />
<rect class="d-cmd" x="14" y="190" width="140" height="24" rx="12" /><text class="d-cmd-text" x="84" y="206" text-anchor="middle">collage build</text>
<rect class="d-cmd" x="166" y="190" width="140" height="24" rx="12" /><text class="d-cmd-text" x="236" y="206" text-anchor="middle">collage export</text>
<line class="d-flow" x1="84" y1="214" x2="84" y2="228" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="236" y1="214" x2="236" y2="228" marker-end="url(#cli-life-arrow)" />
<rect class="d-folder" x="14" y="232" width="140" height="34" rx="4" /><text x="84" y="246" text-anchor="middle">bin/&lt;name&gt;</text><text class="d-small" x="84" y="259" text-anchor="middle">one static binary</text>
<rect class="d-folder" x="166" y="232" width="140" height="34" rx="4" /><text x="236" y="246" text-anchor="middle">dist/</text><text class="d-small" x="236" y="259" text-anchor="middle">HTML and static files</text>
<line class="d-hitline" x1="66" y1="266" x2="66" y2="284" />
<rect class="d-folder" x="14" y="284" width="104" height="34" rx="4" /><text class="d-small" x="66" y="298" text-anchor="middle">-i: image file</text><text class="d-small" x="66" y="311" text-anchor="middle">and service unit</text>
<line class="d-hitline" x1="218" y1="266" x2="218" y2="284" />
<rect class="d-cmd" x="166" y="284" width="104" height="24" rx="12" /><text class="d-cmd-text" x="218" y="300" text-anchor="middle">collage serve</text>
<text class="d-small" x="218" y="322" text-anchor="middle">localhost:4000</text>
<line class="d-flow" x1="140" y1="266" x2="140" y2="336" marker-end="url(#cli-life-arrow)" />
<line class="d-flow" x1="292" y1="266" x2="292" y2="336" marker-end="url(#cli-life-arrow)" />
<rect class="d-pill" x="14" y="340" width="140" height="34" rx="17" /><text x="84" y="354" text-anchor="middle">a server</text><text class="d-small" x="84" y="367" text-anchor="middle">reads HOST and PORT</text>
<rect class="d-pill" x="166" y="340" width="140" height="34" rx="17" /><text x="236" y="354" text-anchor="middle">a static host</text><text class="d-small" x="236" y="367" text-anchor="middle">serves the same files</text>
</svg>
<figcaption>A project's life through the commands. <code>collage new</code> scaffolds it once. While you work on it, <code>add</code>, <code>dev</code>, <code>check</code>, <code>plugin add</code> and <code>del</code>, and <code>inspect</code> all act on the project in the current directory. To ship it, <code>collage build</code> compiles one binary for a server, which listens on <code>HOST</code> and <code>PORT</code>, and <code>-i</code> writes an image file and a service unit beside it. Or <code>collage export</code> renders <code>dist/</code>, which <code>collage serve</code> shows as a static host would before you upload it. <code>collage plugin check</code> is not in the picture: it runs in a plugin's own repository.</figcaption>
</figure>

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success, including `help`, `collage -h` (since v0.11.0) and `collage <command> -h` |
| `1` | The command parsed correctly and failed to do its work |
| `2` | A usage problem: no command, an unknown command, a bad flag, or an unexpected argument |

## collage new

```sh
collage new <name> [--template minimal|demo] [--with plugins] [--dir path] [--module path] [--force]
```

Scaffolds a new, runnable project named `<name>`.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--template name` | `minimal` | The project to scaffold: `minimal` or `demo` (since v0.14.2; the default is `minimal` since v0.32.0, `demo` before) |
| `--with plugins` | none | Plugins to add once it is scaffolded: [bundles](#bundles) such as `@prod` and plugins such as `elagoht/session`, separated by commas or in more than one `--with` (since v0.59.0) |
| `-dir path` | `./<name>` | Directory to scaffold into |
| `-module path` | `<name>` | The module path written into `go.mod` |
| `-force` | off | Scaffold into a non-empty directory anyway |

```sh
collage new myblog                                   # into ./myblog, module "myblog", one page
collage new myblog --template demo                   # the demos and their tests
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
collage new myblog --with @blog,@prod                # and the plugins of two bundles
```

`--with` is read before anything is written: a bundle that does not exist or a
plugin name that cannot be read is a usage error, and no directory is made. Once
the project is scaffolded, `new` requires the collage the CLI is, then adds each
plugin as [`collage plugin add`](#collage-plugin) would. A plugin that fails is
reported and left out; the project and the other plugins stay, and the command
exits `1`. What each plugin added needs from you — its setup note, template
functions and documentation — comes last, together.

The scaffolded `main.go` reads `BASE_URL` into `Config.BaseURL` (since v0.59.0): the
origin a sitemap, a canonical link or a feed puts in front of a path. Empty is
fine until a plugin that writes one is added.

Flags take one dash or two, and may come before or after the name. Exactly one name is required; none, or
more than one, is a usage error. A target directory that exists and is not empty
is refused unless you pass `-force` — and with `-force`, files the scaffold writes
replace files of the same name, and since v0.69.0 the command names them: `go.mod`
among them is your module path and dependencies. Scaffolding into the directory
you are in no longer tells you to `cd .`.

**Every project** gets:

- `go.mod`;
- a `main.go` of about fifty lines: `newApp`, which builds the configuration and
  the static mount from the `collage.Env` it is given, and a `main` that hands it
  to `collage.Main` (since v0.68.0), which keeps the CLI contract described
  [below](#the-contract-with-maingo) and dispatches
  [plugin commands](#plugin-commands);
- `embed.go` and `embed_dev.go`, which declare `files`: templates, static files
  and `plugins-config.json`, embedded, or the project directory in a development
  build (see [below](#development-builds-embed-nothing));
- a `plugins.go` and a `plugins-config.json`, empty, and a `.env.example` (in
  every project since v0.68.0);
- a `routes.go` registering every route;
- a layout that names the site with `WithTitle` and declares no slots (so a
  page's own title replaces the site's);
- a home page that hands its template the project's name with `collage.Value` —
  static without saying `Static()`, because nothing in it fetches;
- `static/`, a `.gitignore` and a README.

**`--template demo`** adds a page of live demos — an action answering
JSON, a form posting to its own page, a fragment with its own URL, a JSON
document — a not-found page, tests for each written with
[`collagetest`](/docs/testing), and a favicon.

**The layout is by area** (since v0.40.0), the layout of a real application:

```text
pages/<area>/<name>.go            a page: its layouts, content, path, actions
fragments/layouts/master.go       Master(), the layout every page wraps itself in
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
template with `collage.Value` — and a stylesheet that sets the background and text
colour, dark mode included. Nothing else — no tests, and no not-found page:
collage answers an unknown address with its own plain 404 until you register one.

When it is done it prints the next steps:

```sh
cd myblog
go mod tidy
collage dev
```

## collage add

```sh
collage add <kind> <[area/]name> [flags]
```

Writes a page, a fragment, an action or a document in the scaffold's layout, and
registers it in `routes.go` (since v0.40.0) — and, since v0.69.0, a layout, a
guard, a middleware or a command:

```sh
collage add page blog/post          # pages/blog/post.go + fragments/pages/blog/post.go
collage add page blog/post --file   # its template in templates/pages/blog/post.html
collage add fragment blog/sidebar   # fragments/blog/sidebar.go, for a slot
collage add fragment layouts/partials/siteFooter --file
                                    # fragments/layouts/partials/siteFooter.go
                                    # + templates/layouts/partials/siteFooter.html
collage add action blog/comment     # actions/blog.go + actions/funcs/blog.go
collage add action blog/ping --path /api/ping
collage add document feed --path /feed.xml --type application/xml
collage add action blog/like --page blog/post   # attached to pages/blog/post.go
collage add layout panel            # fragments/layouts/panel.go, nested
collage add guard admin             # guards/admin.go + fragments/layouts/admin.go
collage add page admin/users --layout panel --guard admin
collage add middleware loadUser     # middleware/loadUser.go + app.Use in routes.go
collage add command seed            # commands/seed.go + app.RegisterCommand
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--file` | off | Keep the template in a file under the template root, not inline |
| `--path pattern` | `/<area>/<name>` | The URL; an action without one answers at the page it is attached to |
| `--name name` | the last segment | The page's, action's, document's or command's name |
| `--locale code` | `Locale.Default`, or `en` | The locale its path is in |
| `--type type` | `text/plain; charset=utf-8` | A document's content type |
| `--layout name` | none | A page's layout: the one `add layout <name>` wrote, inside the site's (since v0.69.0) |
| `--guard name` | none | A page's guard: the layout `add guard <name>` wrote, inside the site's and any `--layout` (since v0.69.0). An action's: `WithGuard(guards.<Name>)` (since v0.70.0) |
| `--page target` | none | An action's page, `[area/]name`: attaches the action to it (since v0.69.0) |
| `--dir path` | `.` | The project |

- **A page** is `pages/<area>/<name>.go`, wrapped in `layouts.Master()` — or the
  first layout in `fragments/layouts` that takes no arguments — with its content in
  `fragments/pages/<area>/<name>.go`: a template, a view struct, and a typed
  `collage.Load` handler that declares the page's title.
- **An action** gets a builder appended to `actions/<area>.go` and a handler to
  `actions/funcs/<area>.go`, either created when it is not there. With `--path` it
  answers at a URL of its own and is registered; without one it answers at the
  page it is attached to, and the command prints the `.WithActionFor(...)` to add.
- **`--page <[area/]name>`** (since v0.69.0) attaches a new action to that page
  instead of printing how: it adds `.WithActionFor(actions.Like())` before the
  `Build()` that ends the page's `NewPage` chain. A page file with more than one
  chain is left alone, and the line to add is printed. It cannot go with
  `--path`.
- **A document** is `documents/<name>.go`.
- **A layout** (since v0.69.0) is `fragments/layouts/<name>.go`, a layout that
  nests inside the site's own around a page's content,
  `<div class="<name>">{{slot "content"}}</div>` — inline, or in a template file
  with `--file`. `--layout <name>` wraps a new page in it. A layout `add layout`
  or `add guard` wrote is marked `//collage:nested`, and one that carries a guard
  counts as nested too: neither is ever taken for the site's layout a new page
  goes into.
- **A guard** (since v0.69.0) is two files. `guards/<name>.go` is the
  [guard](/docs/pages-and-layouts#private-pages-guards), a `GuardFunc` that refuses
  everyone with `403` until you write its check — return `nil, nil` to let a
  request through, or a redirect. `fragments/layouts/<name>.go` is a layout that
  renders nothing of its own and carries the guard, since a guard acts where a
  page's layout chain carries it. `--guard <name>` wraps a new page in it:
  `WithLayouts(layouts.Master(), layouts.Admin())`. The page's renders, its
  attached actions and its fragment paths meet the guard first. An action with
  its own `--path` has no page; since v0.70.0, `add action ... --guard admin`
  gives it the guard itself, `.WithGuard(guards.Admin)`.
- **A middleware** (since v0.69.0) is `middleware/<name>.go`, a
  `func(http.Handler) http.Handler` that passes every request on, registered with
  [`app.Use`](/docs/middleware-and-apis#middleware-appuse) after the middleware already there —
  the first registered is the outermost. It runs before routing and guards run
  after it, so a guard reads what it puts in the request's context: load the
  signed-in user there, decide in the guard.
- **A command** (since v0.69.0) is `commands/<name>.go`, a `collage.Command`
  registered with `app.RegisterCommand`: `go run . <name>`, or the built binary
  followed by it, runs it instead of the server. `--name` gives it a name that is
  no Go name, `user-add`; a name another `collage.Command` in the project has is
  refused. A middleware or a command that needs a service takes it as an argument
  you add. Both go into the register function as statements of their own,
  `if err := app.Use(...); err != nil { return err }`, which needs one that takes
  the `*collage.App` and returns an `error`; otherwise the line to add is printed.
- **A fragment** is a page's content without the page, for a slot of a page or a
  layout; the command prints the `.WithSlotFragment(...)` that binds it. Since
  v0.68.0 it goes where its path says, `fragments/<area>/<name>.go`, with its
  template at `<area>/<name>` under the template root — before, it went under
  `fragments/pages/`. Its package is the one already in that directory, or one
  named after it: `fragments/layouts/partials` is `partials`. A fragment that
  belongs beside a page's content is `collage add fragment pages/blog/related`.

The last segment is a Go name, kept as written (since v0.68.0):
`blog/storyDetail`, `layouts/partials/siteFooter`. What the command writes is
named after it — the file, the default URL and the name, so `blog/storyDetail` is
the page `"storyDetail"` at `/blog/storyDetail` — unless `--name` gives another,
and its function is that name with the first letter capitalised, `StoryDetail()`.
A dashed name is refused, as it is no Go name. The directories before it are only
a path and may have dashes, `user-settings/profile`; their package and import
alias drop them, `usersettings`. The locale of its path is `Locale.Default` as
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
— `localhost` or a name under it, an IP address, a reserved name such as
`example.com` or `*.test` (since v0.56.0) — or the `HOST` it was started with, or a
name in `COLLAGE_DEV_HOST`, and refuses anything else with a `403`.

A page on another site can make its
own name resolve to `127.0.0.1` (DNS rebinding) and would then be same-origin with
`collage dev`: it could read your pages, the development error pages with their
stacks, and the program's output. The `Host` it sends is its own name, which it
cannot change.

To reach `collage dev` under another name, start it with `HOST` set
to it, or list the name in `COLLAGE_DEV_HOST`, a comma-separated list such as
`app,mybox.lan` for a docker-compose service and a LAN name. `collage dev` reads it
from the shell or the environment file, applies it to its proxy, and starts the
program with the list plus its own `HOST`, since the program listens on loopback but
sees the browser's `Host`; the program applies the same rule in development mode.

Since v0.56.0 `collage dev` also logs a Warn when its own proxy is bound anywhere
but loopback, since the development pages are then reachable from other machines.
See [Deployment](/docs/deployment#development-mode-stays-on-this-machine).

What the program prints still reaches the terminal, as the program printed it —
with one change: its own address is replaced by `collage dev`'s, so its
`collage: listening` line names the address to open (since v0.32.0).

On a colour
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
and listens on the `HOST` and `PORT` it is given: `collage.Main` puts both in the
`Env` it hands `newApp`.
Development mode reads templates and static files from disk on every request, so
editing those needs no rebuild and none happens — the page in your browser reloads
itself instead. Content your program reads from disk itself, such as Markdown,
reloads the page too once its directory is named in
[`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0). See
[Templates](/docs/templates#reloading-in-development) for what else development
mode changes.

### Development builds embed nothing

Since v0.46.0 `collage dev` builds with `go build -tags collage_dev`, and the
scaffold keeps its `//go:embed` line in `embed.go`, which starts with
`//go:build !collage_dev`. Beside it, `embed_dev.go` (`//go:build collage_dev`)
declares the same variable. Development mode reads `templates/` and `static/` from
disk anyway, so a development build has nothing to embed, and `collage build`,
`collage export` and a plain `go build` pass no tag and embed as before.

Since v0.68.0 the two files declare one filesystem, `files`. In `embed.go` it is
`//go:embed all:templates all:static plugins-config.json`; in `embed_dev.go` it is
the project directory itself, `os.DirFS(".")`, so a plugin given `files` — a card
template, a favicon source, Markdown posts — reads the copy you are editing
rather than an empty filesystem. (Before v0.68.0 they declared `templatesFS` and
`staticFS`, and the development build's were empty.)

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

To fix it, move the `//go:embed` variables out of `main.go` (and `"embed"` out
of its imports) into two files of their own. The ones a new project gets are:

```go
// embed.go
//go:build !collage_dev

package main

import "embed"

//go:embed all:templates all:static plugins-config.json
var files embed.FS
```

```go
// embed_dev.go
//go:build collage_dev

package main

import (
	"io/fs"
	"os"
)

var files fs.FS = os.DirFS(".")
```

A `main.go` that still reads `templatesFS` and `staticFS` keeps the two
variables, each with its own `//go:embed` line, and declares them empty in
`embed_dev.go`; to move to `files`, copy `main.go`, `embed.go` and `embed_dev.go`
from a new project.

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
  `collage dev` itself listens, read once when it starts. `COLLAGE_DEV_HOST` is
  read from the file or the shell too, and the program is given it with
  `collage dev`'s own `HOST` added.
- `BASE_URL`, when neither the shell nor the file sets it, is `collage dev`'s own
  address, `http://localhost:6060` (since v0.68.0). A plugin that writes absolute
  URLs — a sitemap, a canonical link — refuses to start without an origin, and in
  development that is the right one. Deployed, set it yourself.
- No file is not an error. When a file is read, its name is printed on stderr.
- The file is read again on every restart, and it is watched, so editing it
  restarts the program with the new values.

A program whose `main` calls `collage.Main`, as a scaffolded one does, reads them
too, with the same rules (since v0.68.0): `go run .`, `collage export` and the
built binary load `.env` from the working directory — `.env.development` first
under `COLLAGE_DEV=1` — before `newApp` runs, so `os.Getenv("BASE_URL")` sees it
with no `source .env`. Outside development the file read is named on stderr. On a
server that sets its variables itself, the variables win and a missing file is no
error; a malformed line stops the program with the file and the line. A `main` of
your own that does not call `collage.Main` reads no file.

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

## collage routes

```sh
collage routes
```

Lists every URL the project answers through collage, a row per method and path,
sorted by path (since v0.69.0):

```text
METHOD  PATH          KIND    NAME    NOTES
GET     /             page    home    static
GET     /admin/users  page    users   dynamic, guard adminGuard
POST    /admin/users  action  rename  on page users, guard adminGuard
POST    /api/fire     action  fire    own URL, own guard
POST    /api/ping     action  ping    own URL, no guard
GET     /static/*     mount           12 files

Not listed: handlers mounted with app.Handle, and plugins' own endpoints.
```

A page says how it is cached and which guards it meets; an action says whether it
answers on a page, and so meets the page's guards, or at a URL of its own, and
whether it has a guard of its own (since v0.70.0); a page's fragment URLs are listed with the guards they meet too. A `LOCALE`
column appears when the site has more than one locale. It reads what
[`collage inspect`](#collage-inspect) prints, so the project is built and `newApp`
runs. What is not registered through collage — `app.Handle`, a plugin's own
endpoints such as `/healthz` — is not in the list. It takes no flags.

## collage inspect

```sh
collage inspect
```

Prints what the current directory's project is made of, as JSON (since v0.27.0):
every page with its patterns, parameters, `layouts` (outermost first) and `guards`,
every fragment with its template and slots — an inline fragment with `"inline":
true` and no template path — the documents, the actions, the template functions,
the plugins, the locales, and the files the mounts serve. It takes no flags; positional arguments are a usage
error. It builds the project as `collage dev` does and runs the program with
`COLLAGE_DEV=1` (since v0.70.1; before, `go run . collage-inspect`):

```sh
go build -tags collage_dev -o <temporary file> .
COLLAGE_DEV=1 <temporary file> collage-inspect
```

Nothing is embedded, so the program reads `templates/` and `static/` from disk,
and the build goes to a file of its own: `go run` stores the linked program in
the Go build cache, and without the tag the compiled `main` package holds every
file the scaffold embeds, so each run after a change to `main` stored another
copy of them there, kept for days. `collage routes` and `collage check` run the
project the same way.

`collage-inspect` is a command `DispatchCommands` answers itself rather than handing
it to a plugin: it prints the application's description as indented JSON. That JSON
is the contract; the Go types behind it (`App.Inspect`, `collage.Inspection`,
`collage.InspectCommand`) left the public API in v0.60.0, and the output is
byte-for-byte what it was.

The name is prefixed so that no plugin's
command is taken. The scaffolded `main.go` hands the word after its flags to
`DispatchCommands` (see [Plugin commands](#plugin-commands)), so a scaffolded
project on collage v0.27.0 or later needs nothing more. A program that does not
dispatch its arguments that way has nothing to answer with.

It is what an editor's completion reads: the Collage Snippets & Highlighter
extension for VS Code ([Editor support](/docs/installation#editor-support)) offers
page names in `{{pageURL "…"}}`, slots in `{{slot "…"}}` and files in
`{{asset "…"}}` from it.

Since v0.49.0 each fragment also carries the Go type its template sees as `.`, and
the root carries `types`, a table of the types those reach. An editor can then
complete `{{.` and check a field name the way registration does — see
[How templates are checked](/docs/data-handlers#how-templates-are-checked):

```json
"fragments": [
  {"name": "post-body", "template": "pages/post.html", "handler": true, "dataType": "*blog.Post"},
  {"name": "layout", "template": "layouts/master.html", "dataType": "nil"},
  {"name": "legacy", "template": "pages/legacy.html", "handler": true, "dataType": null, "typeCheck": false}
],
"types": {
  "blog.Post": {
    "kind": "struct",
    "fields": [
      {"name": "Base", "type": "blog.Base", "embedded": true},
      {"name": "ID", "type": "int"},
      {"name": "Title", "type": "string"},
      {"name": "Author", "type": "*blog.User"},
      {"name": "Comments", "type": "blog.Comments"},
      {"name": "Related", "type": "blog.Tags"}
    ],
    "methods": [{"name": "URL", "args": 0, "returns": "string"}]
  },
  "blog.Base": {"kind": "struct", "fields": [{"name": "ID", "type": "int"}]},
  "blog.Comments": {"kind": "slice", "elem": "blog.Comment"},
  "blog.Tags": {"kind": "map", "key": "string", "elem": "[]blog.Post"},
  "blog.User": {"kind": "struct", "fields": [{"name": "Name", "type": "string"}]}
}
```

- **`dataType`** is always written: the Go type, `"nil"` for a fragment with no
  data or with `collage.Effect`, and `null` when the type is an interface and so
  unknown until the page renders — a handler declared to return `any`.
- **`typeCheck: false`** appears only on a fragment built `WithoutTypeCheck()`.
- **`types`** holds every named type reachable from some fragment's data type —
  through fields, element and key types, and method results — keyed by its Go
  name, with its exported fields (promoted ones included) and the exported methods
  of it and its pointer that return something.
  - An exported embedded field is a
    field too — a template reaches it as `{{.Base}}` — and carries
    `"embedded": true`; the fields it promotes are listed beside it as usual (since
    v0.51.1).
  - A named pointer, slice, array, map or chan carries `elem`, the type
    it holds, and a map also `key` (since v0.51.1).
  - A type is listed once and named
    everywhere else, so a recursive type ends. An unnamed struct is listed under its
    Go spelling; other unnamed composites are written inline in a type string
    (`[]blog.Comment`).
  - Standard library types such as `time.Time` and
    `template.HTML` are named, not described.
  - `types` is left out when no fragment
    has a known data type.
  - A key two different types share (`models.User` from two
    packages named `models`) is written as `{"kind": "struct", "ambiguous": true}`
    with no fields or methods, and an editor treats it as unknown (since v0.51.2).

These are additions: the output's `version` is still `1`.

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
It runs the program as `collage inspect` does:

```sh
go build -tags collage_dev -o <temporary file> .
COLLAGE_DEV=1 <temporary file> collage-check
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

## collage plugin

*Since v0.58.0.* Adds plugins to the project, takes them out, or checks a plugin
repository:

```sh
collage plugin add <plugin>... [-dir dir]
collage plugin del <plugin>... [-dir dir]
collage plugin check [-release] [-run] [dir]
```

```sh
collage plugin add elagoht/consent            # github.com/elagoht/collage-consent
collage plugin add elagoht/sitemap@v0.3.1     # a version go get accepts
collage plugin add gitlab.com/kisi/plugin     # a module path, used as written
collage plugin del elagoht/consent
collage plugin add elagoht/session elagoht/flash   # several, in order (since v0.59.0)
collage plugin add @prod                      # a bundle
collage plugin del @live elagoht/session
collage plugin check -release -run            # in a plugin's own repository
```

`owner/name` means `github.com/<owner>/collage-<name>`; a path whose first
segment has a dot is a module path and is used as written. `add` and `del` work
on the project in the current directory, or the one `-dir` names: it must hold
`go.mod` and a `package main`.

**`add`** fetches the module with `go get`. Without `@version` it asks for
`@upgrade`: the latest release for a module `go.mod` does not have yet, and never
a move down for one it has. It then checks that the module is a plugin; the
contract is on [the plugin authors' page](/docs/plugin-authors):

- a `collage.json` at the module root whose `name` is `owner/name`, the one you
  typed when you gave a short name;
- the plugin's package at the module root, not `package main`;
- a `New` that can be called with nothing configured: no parameters, a variadic
  one, or one whose zero value can be written, an exported struct of the
  package's own as `pkg.Options{}`, or `nil` for a pointer, interface, slice,
  map or func;
- a release tag, not a pseudo-version.

It adds the plugins the manifest `requires` first, then writes the import and the
call into `plugins.go`, before the first entry whose manifest has a greater
`order`, and an empty section to `plugins-config.json`. It ends with
`go mod tidy`:

```text
$ collage plugin add elagoht/consent
added elagoht/consent v0.2.0
  plugins.go           consent.New(consent.Options{})
  plugins-config.json  "elagoht/consent": {}

elagoht/consent
  The banner needs its categories and wording: set `"categories"` (at least one) and `"text"` (the default locale's wording, labelling every category; other locales under `"locales"`) in this plugin's section of `plugins-config.json`. Without categories the application does not start.
  Docs: https://github.com/Elagoht/collage-consent
```

It then prints what the plugin still needs from you, from its manifest (since
v0.68.0): its `setup` note, the template functions it adds, each with the first
sentence of its description, and its documentation's address. With several
plugins — a bundle, `new --with` — these come together at the end, under "What
each plugin needs from you". An empty section means the defaults, and a setting
the plugin requires is reported by the plugin when the application starts.

- **The import is named after the package.** When that name is already an import
  of `plugins.go`, or a name declared anywhere in the package, it becomes
  `<owner><Name>`, and then gets a number.
- **A plugin `plugins.go` already has is not added twice.** Only its version
  moves, and the command says so.
- **An existing section is kept.** Both files are edited in place, so your
  arguments, comments, order and formatting stay.
- **Nothing is left half-done.** Any failure, Ctrl-C included, puts `go.mod`,
  `go.sum`, `plugins.go` and `plugins-config.json` back as they were. A file that
  was absent is removed again.

A project made before `plugins.go` existed gets the file from its first `add`,
and the command prints the `Plugins: plugins(),` line to give `collage.New`, once.
New projects already have both.

**`del`** finds the plugin among `plugins.go`'s entries, by the name in their
manifests or by the module path, in any case. It removes its entries, its import
and its section, prints the section when it held settings, and runs
`go mod tidy`. It refuses:

- while another plugin's manifest `requires` it;
- when another entry's `collage.json` cannot be read (the module cannot be
  fetched, or the file is malformed), since it cannot then tell whether that
  plugin requires this one; the message names the module and the reason;
- when it was given a module path and the plugin's own `collage.json` cannot be
  read, since the name of its section is then unknown.

Another entry's `collage.json` is read for its `name`, `requires` and `order`
only, so a key from a newer manifest schema stops nothing. An entry wired by hand
that is not a plugin module's root package (a package of your own module, or a
subpackage of a dependency), or a module with no `collage.json`, requires nothing
and stops nothing. A module the module cache lacks, as after a fresh clone, is
fetched to read its manifest. Like `add`, `del` puts every file back when a step
fails.

### Several at once

*Since v0.59.0.* `add` and `del` take several plugins, and work through them in
the order given. Each is a step of its own: one that fails is reported and put
back, and the ones before and after it stay. The command then exits `1`. A name
that cannot be read stops the command before anything changes, as does one
plugin given with two versions; a plugin given twice is done once.

With more than one, a summary follows, and `add` prints the plugins' setup notes
together at the end rather than after each:

```text
$ collage plugin add elagoht/feed elagoht/nope elagoht/toc
added elagoht/feed v0.3.0
  ...
collage: plugin add: elagoht/nope: ...

added: elagoht/feed, elagoht/toc
failed: elagoht/nope

To finish setting up:

elagoht/feed
  Feeds are Go values: give them to `feed.New(...)` in `plugins.go`, ...
```

`del` takes the order off your hands: a plugin another one on the list requires
is removed after that one, so `del elagoht/live elagoht/websocket` works.

### Bundles

*Since v0.59.0.* A bundle is a set of plugins under one name, written with `@`:
`collage plugin add @prod`, `collage new site --with @blog`. Its plugins are
added in the order below, and each goes to the place in `plugins.go` its manifest
asks for. `del @name` removes those of a bundle's plugins the project has and
passes over the rest. `collage help plugin` lists them too.

| Bundle | For | Plugins |
| --- | --- | --- |
| `@prod` | what a site in production wants | health, secure, compress, minimizer, accesslog, ratelimit |
| `@site` | a public site's search and share basics | robots, sitemap, meta, jsonld, favicon |
| `@blog` | `@site`, and posts written in Markdown | the plugins of `@site`, markdown, feed, toc, highlight, ogimage |
| `@forms` | forms, and state kept in cookies | session, flash, validate, honeypot |
| `@live` | fragments the server keeps current | live, websocket |
| `@ops` | metrics and error reports | prometheus, errortrack |

Every plugin is `elagoht/<name>`. Some need a setting before the application
starts: sitemap and meta the site's origin (`BASE_URL` in a scaffolded project),
favicon a source image, markdown and ogimage a filesystem given in `plugins.go`,
errortrack a DSN outside development. Their setup notes say what to do, and the
application refuses to start, naming the setting, until it is done.

**`check`** checks a plugin repository (`dir`, default `.`) against the contract
`add` enforces. The check itself reads only the repository:

- `-release` also wants a release tag at `HEAD`, which needs git;
- `-run` builds a throwaway program in the repository's `collage-plugin-check/`
  directory and runs it, which fetches what the module cache lacks. The
  directory has one name, so the Go build cache keeps one copy of the program
  (since v0.70.1); one already there — another check running, or one
  interrupted — is refused and left alone. It calls `New` exactly as `add` would
  write it, as a `collage.Plugin`, and checks that `Name()` is the manifest's
  `name`. It fails on a build error, a panic in `New` or a different name. It
  does not configure or start the plugin, so a setting an empty section lacks is
  not its concern. `go.mod`, `go.sum` and `go.work.sum` are put back and the
  program is removed afterwards.

These are commands of the `collage` binary. A plugin's own subcommands are
another thing: see [Plugin commands](#plugin-commands).

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

## Example: from a new project to a first deploy

The commands above, in the order a site meets them: a project called `myblog`,
written on a laptop and deployed to a server of another platform. Times, sizes, hashes
and cache directories in the output differ on your machine.

**1. Scaffold it.**

```sh
collage new myblog
cd myblog
go mod tidy
```

`new` writes the [minimal project](#collage-new) and prints the next steps:

```text
Scaffolded "myblog" in myblog

Next steps:
  cd myblog
  go mod tidy
  collage dev
```

The scaffolded `go.mod` names only the module and the Go version; `go mod tidy`
adds collage to it.

**2. Run it.**

```sh
collage dev
```

```text
01:13:17 • collage dev: serving http://localhost:6060
01:13:18 • collage: dev mode is rendering templates from disk, not the embedded copy  root=templates
01:13:18 • collage: dev mode, using an in-memory cache instead of the configured disk cache
01:13:18 • collage: listening  addr=localhost:6060
```

`http://localhost:6060` shows `Hello from myblog`. Leave it running: the next
step adds Go files, and it rebuilds when they appear.

**3. Add a page, and check the links.** In a second terminal:

```sh
collage add page about
```

```text
wrote  fragments/pages/about.go
wrote  pages/about.go
edited routes.go
locale en, from the default; main.go sets no Locale.Default literal, --locale changes it
page "about" at /about
```

`collage dev` rebuilds, and `/about` answers. Then:

```sh
collage check
```

```text
01:14:05 • collage: caching to disk  dir=.cache/6ba57471adaf48bf  version=19be6c1c06fea3dcff022e9dab873f1f
collage: check: nothing found
```

The first line is the program's own, since `check` starts it. The second is the
verdict, and the exit code is `0`; in CI, this is the step before `collage build`.

**4. Build for the server.** The laptop is not the server, so name the server's
platform, and ask for the two extra files with `-i`:

```sh
collage build -i -os linux -arch amd64
```

```text
Building myblog for linux/amd64.

  Write a Dockerfile? [y/N]: y
  Write a systemd unit? [y/N]: y


✓ bin/myblog
    linux/amd64 · 10.3 MB
    wrote bin/Dockerfile    docker build -f bin/Dockerfile .
    wrote bin/myblog.service

10.3 MB · linux/amd64 · 6.7s
```

The scaffold's `.gitignore` ignores `bin/myblog` by name and not the two files
beside it: those are yours to edit and commit. Run the command again and
both questions are skipped, since the files are there.

**5. Put it on the server.** The two files are two ways to run it, and only one
of them uses `bin/myblog`:

- **The service unit** runs the binary you built. Copy `bin/myblog` to
  `/usr/local/bin/myblog` on the server, adjust the unit's `User` and
  `WorkingDirectory` — the file cannot know them — and install it where its first
  comment says. It listens on `127.0.0.1:8080`, for a reverse proxy in front.
- **The image file** builds its own binary: its first stage compiles the project
  from source with the same flags, and its second stage holds the result. Build
  it from the project root, as the report says. It listens on `0.0.0.0:8080`.

The `8080` is those files' own setting, written as `PORT=8080`. The binary run
with no `HOST` or `PORT` listens on `localhost:6060`, the scaffold's default; see
[Deployment](/docs/deployment#environment). Both files also carry a commented-out
`COLLAGE_CSRF_KEY`: set it to at least 32 random bytes, or a key is generated per
process and every form submitted before a restart is refused after it.

Add plugins before `-i`. The image file copies `plugins-config.json` only when the
project has one as the file is written, and `-i` never rewrites the file; a
plugin added afterwards means adding that `COPY` line by hand.

**Or export it.** For a host that serves files, `collage export -clean` renders
`dist/` instead, and its report leaves the new page out:

```text
✓ 3 files written
    dist/index.html
    dist/static/app.css
    dist/static/app.f106e88ebb47f51d.css

▲ 1 skipped
    about  page uses the dynamic render strategy, which cannot be built statically

3 written · 1 skipped · 0 failed · 1.5ms
```

`collage add` gave `about` a `collage.Load` handler and no strategy, so it is
dynamic. Add `.Static()` to its builder in `pages/about.go` and the next export
writes `dist/about/index.html` too. `collage serve` then shows `dist/` on
`localhost:4000`, as a static host would serve it.

## The contract with main.go

`dev`, `export` and `inspect` depend on three things your `main.go` does, and
`collage.Main`, which the scaffolded one calls (since v0.68.0), does all three. The
third, handing the words after its flags to `collage.DispatchCommands`, is also
what runs your plugins' commands:

| Command | Runs | Your `main.go` must |
| --- | --- | --- |
| `collage dev` | `go build -tags collage_dev`, then the binary, with `COLLAGE_DEV=1`, the `HOST` and `PORT` to listen on, and `COLLAGE_DEV_HOST` | turn on development mode when `COLLAGE_DEV` is `1`, and listen on `HOST` and `PORT` |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | parse `-collage-build`, `-out` and `-clean`, and on `-collage-build` render to `<dir>` instead of serving |
| `collage inspect` | `go build -tags collage_dev`, then the binary with `collage-inspect` and `COLLAGE_DEV=1`; `routes` and `check` too | turn on development mode when `COLLAGE_DEV` is `1`, and pass the words after its flags to `collage.DispatchCommands` |

`collage build` needs nothing from `main.go`: compiling is something `go build`
does without being told anything.

The scaffolded `main.go` leaves all of it to `collage.Main`:

```go
func main() {
	collage.Main(newApp)
}

func newApp(env collage.Env) (*collage.App, error) {
	pluginConfig, err := env.PluginConfig(files)
	if err != nil {
		return nil, err
	}

	app, err := collage.New(&collage.Config{
		DevMode: env.Dev,
		BaseURL: os.Getenv("BASE_URL"),
		Server:  collage.ServerConfig{Host: env.Host, Port: env.Port},
		Template: collage.TemplateConfig{
			FS:        files,
			Root:      "templates",
			Extension: ".html",
		},
		Cache: collage.CacheConfig{
			Enabled:    true,
			Type:       "disk",
			Dir:        env.CacheDir,
			DefaultTTL: 5 * time.Minute,
		},
		PluginConfig: pluginConfig,
		Plugins:      plugins(),
		Security: collage.SecurityConfig{
			CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
		},
	})
	if err != nil {
		return nil, err
	}

	if err := register(app); err != nil {
		return nil, err
	}

	return app, app.Mount("/static/", env.Static(files, "static"))
}
```

`collage.Main` loads the [environment file](#environment-files), parses
`-collage-build`, `-out`, `-clean` and `-port`, and calls `newApp` with a
`collage.Env`:

| Field | Value |
| --- | --- |
| `Dev` | `COLLAGE_DEV=1`, which `collage dev` sets |
| `Host` | `HOST`, or `localhost` |
| `Port` | `-port`, `PORT`, or `6060` |
| `CacheDir` | `.cache`; a test passes a directory of its own |

`env.PluginConfig(files)` reads `plugins-config.json` from the working directory
when it is there, so a deployment can change a setting without a rebuild, and
otherwise the copy the binary embeds, so a binary deployed without the file keeps
its settings. `env.Static(files, "static")` serves the static directory from disk
in development and embedded otherwise. Then `collage.Main` dispatches a plugin's
command, renders to `-out` on `-collage-build`, or serves, and shuts the
application down after a build or a command.

A project scaffolded before v0.68.0 keeps working as it is. To move, copy the new
`main.go`, `embed.go` and `embed_dev.go` from `collage new`, and call
`newApp(collage.Env{CacheDir: t.TempDir()})` in tests. A `main` written without
`collage.Main` — the shape a project scaffolded before v0.68.0 has, trimmed to the
part that matters — keeps the contract this way:

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
		if err := shutdown(app); err != nil {
			fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
			code = max(code, 1)
		}
		os.Exit(code)
	}

	if *buildFlag {
		buildErr := staticBuild(app, *outFlag, *cleanFlag)
		if err := shutdown(app); err != nil {
			log.Printf("shutdown: %v", err)
			buildErr = errors.Join(buildErr, err)
		}
		if buildErr != nil {
			log.Fatalf("static build: %v", buildErr)
		}
		return
	}

	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// shutdown runs every plugin's Shutdown once a build or a command is done.
// Neither serves, so nothing else would. Only after a successful start: a start
// that failed has already shut down the plugins it started.
func shutdown(app *collage.App) error {
	if app.Start() != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return app.Shutdown(ctx)
}
```

A build and a command serve nothing, so `ListenAndServe` never shuts the
application down for them. `collage.Main` calls `app.Shutdown` itself once either
is done, and a `main` of your own does it as above, so every plugin's `Shutdown`
runs: what a plugin holds is flushed, and work handed to a queue plugin such as
[elagoht/jobs](/docs/plugins#elagohtjobs) is at least logged when it is dropped.

If you write `main` without `collage.Main`, keep all of it working, or `collage dev`,
`collage export`, `collage inspect` and your plugins' commands stop doing anything
useful in your project.

## Plugin commands

Adding and removing plugins is [`collage plugin`](#collage-plugin). This section is about a plugin's own subcommands. A plugin can contribute one through `Host.RegisterCommand`. **The `collage`
CLI does not run it.** The CLI never loads your application, so a command that
only exists once your plugins have started is out of its reach, and
`collage <plugin-command>` is an unknown command.

Your own program dispatches them, with `collage.DispatchCommands`.
`collage.Main`, which the scaffolded `main.go` calls, already does. A `main` of
your own does it this way (the block [above](#the-contract-with-maingo)): after
the flags are parsed and the application is built, a word left over is a command:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := shutdown(app); err != nil { // see above
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
		code = max(code, 1)
	}
	os.Exit(code) // runs no defers, so the shutdown comes first
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

The exit code follows the CLI's:

| Code | When |
| --- | --- |
| `0` | success |
| `1` | a startup failure, a command that ran and failed, or a command with no `Run` |
| `2` | a nil app, no arguments, or a word no plugin registered (`ErrUnknownCommand`) |

An
unclaimed word is a usage error rather than a server started by accident. One word
no plugin registers is answered all the same: `collage-inspect`, which
[`collage inspect`](#collage-inspect) runs. A
program that would rather serve when no command matches can check
`errors.Is(err, collage.ErrUnknownCommand)` and carry on instead of exiting.
`app.Commands()` lists what the plugins registered, if you want to print your own
usage.

Dispatching starts the application, which closes registration, and a later
`ListenAndServe` reuses that start. `DispatchCommands` does not shut the
application down, since a program may go on to serve it: once a command has run,
call `app.Shutdown` yourself, as above, so that every plugin's `Shutdown` runs;
`collage.Main` does.
Writing such a command is covered in
[Writing a plugin](/docs/writing-plugins#commands).
