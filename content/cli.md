---
description: Every command of the collage CLI — new, dev, build, export, serve, version and help — with its flags and exactly what it runs.
reference: DispatchCommands, Command, ErrUnknownCommand
---

# The collage CLI

The `collage` command scaffolds projects and drives the ones you have: running one
in development, compiling the binary you deploy, and exporting it as static files.

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

It never links your application into itself — it cannot, because your application
is your code. `dev`, `build` and `export` run the `go` tool in the current
directory, exactly as you would by hand, and the rest of this page says precisely
what each one runs.

## Usage

```sh
collage <command> [flags]
```

| Command | What it does |
| --- | --- |
| `new` | Scaffold a new collage project |
| `dev` | Run the current directory's project in development mode |
| `build` | Compile the current directory's project into the binary you deploy |
| `export` | Render the current directory's project to static files |
| `serve` | Serve a static export the way a static host would |
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
collage new <name> [-minimal] [-dir path] [-module path] [-force]
```

Scaffolds a new, runnable project named `<name>`.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-minimal` | off | Scaffold without the demos |
| `-dir path` | `./<name>` | Directory to scaffold into |
| `-module path` | `<name>` | The module path written into `go.mod` |
| `-force` | off | Scaffold into a non-empty directory anyway |

```sh
collage new myblog                                   # into ./myblog, module "myblog"
collage new myblog -minimal                          # without the demos
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
```

Flags may come before or after the name. Exactly one name is required; none, or
more than one, is a usage error. A target directory that exists and is not empty
is refused unless you pass `-force` — and with `-force`, files the scaffold writes
replace files of the same name.

**Every project** gets `go.mod`, a `main.go` holding the configuration, the static
mount, the CLI contract described [below](#the-contract-with-maingo) and the
dispatch of [plugin commands](#plugin-commands), a `routes.go` registering every
route, a layout that declares the site's title with `rc.HoistTitle` (so a page's
own title replaces it), a home page, a not-found page,
their tests, `plugins-config.json`, `static/`, a `.env.example`, a `.gitignore` and
a README.

**The default project** adds a page of live demos — an action answering JSON, a
form posting to its own page, a fragment with its own URL, a JSON document —
split into `pages/`, `fragments/`, `actions/`, `documents/` and `store/`, with tests
for each.

**`-minimal`** is one layout, an empty home page and a not-found page, with tests
for both: the same `main.go` and the same project shape, with nothing to delete
before you start.

When it is done it prints the next steps:

```sh
cd myblog
go mod tidy
cp .env.example .env.development
collage dev
```

## collage dev

```sh
collage dev
```

Builds the project in the current directory, runs it with `COLLAGE_DEV=1` set,
and rebuilds and restarts it whenever its Go code changes. It takes no flags and
no arguments. Press Ctrl-C to stop; the program receives the interrupt too and
shuts down the way it would in production.

The scaffolded `main.go` turns on development mode when `COLLAGE_DEV=1` is set.
Development mode reads templates and static files from disk on every request, so
editing those needs no rebuild and none happens — the page in your browser reloads
itself instead. Content your program reads from disk itself, such as Markdown,
reloads the page too once its directory is named in
[`Config.DevWatch`](/docs/configuration#devwatch) (since v0.10.0). See
[Templates](/docs/templates#reloading-in-development) for what else development
mode changes.

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
- **A program that exits by itself** — a panic at startup, a port already in use —
  is not restarted in a loop. The next change starts it again.

### Environment files

`collage dev` adds the variables of `.env.development` in the current directory to
the program's environment, or of `.env` when there is no `.env.development`. One
file, never both: `.env.development` replaces `.env` rather than being merged over
it.

```sh
# .env.development
PORT=3000
HOST=localhost
export COLLAGE_CSRF_KEY="0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
```

- The format is `KEY=value` lines, blank lines and lines starting with `#`. An
  `export ` prefix is allowed, and so is a pair of matching single or double quotes
  around a value, which are removed with nothing inside interpreted.
- An unquoted value ends at a `#` that follows whitespace, so `PORT=3000 # dev` is
  `3000`.
- Keys are letters, digits and underscores, not starting with a digit.
- **A malformed line stops the program from starting**: `collage dev` reports it
  with the file and the line (`.env.development:3`) and starts or restarts nothing
  until it is fixed — a program already running keeps running on the values it was
  started with. It keeps watching, so saving the corrected file carries on. A
  skipped line would be a setting you wrote and the program never saw.
- **A variable already set in the shell wins** over the file, so
  `PORT=4000 collage dev` still works.
- `COLLAGE_DEV=1` is always set, whatever the file says.
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
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/<name> .
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o path` | `bin/<name>` | Where to write the binary |
| `-os name` | `linux` | Target operating system (`GOOS`) |
| `-arch name` | `amd64` | Target architecture (`GOARCH`) |
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
- **linux/amd64 by default** rather than the machine you are on: a binary built on
  a Mac does not run in a Linux container, and "exec format error" on a server is
  the wrong place to find that out.
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

What gets exported, what is skipped and why, and the safety checks on the output
directory are in [Static export](/docs/static-export). For a site with forms or
per-request pages, `collage build` is the one you want.

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

The port is 4000 rather than 3000 so it can run beside `collage dev` — which is
exactly when you compare the two.

It behaves like a static host rather than a file server:

- A path with no extension is answered with `<path>/index.html`, which is the
  shape an export writes.
- Directories are never listed.
- A path that resolves to nothing is answered with the export's own `404.html` and
  a 404 status, or a plain 404 when there is none.
- Only `GET` and `HEAD` are answered; anything else is a 405.
- Every response is sent with `Cache-Control: no-store`, so re-exporting and
  reloading shows the new output rather than the old.

A directory that does not exist, or holds no files, is an error that tells you to
run `collage export` first. A path that exists but is not a directory is an error
too, saying just that.

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

`dev` and `export` depend on two things your `main.go` does, and the scaffolded one
does both — along with a third, running plugin commands, that no `collage` command
needs but your plugins do:

| Command | Runs | Your `main.go` must |
| --- | --- | --- |
| `collage dev` | `go build`, then the binary, with `COLLAGE_DEV=1` | turn on development mode when `COLLAGE_DEV` is `1` |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | parse `-collage-build`, `-out` and `-clean`, and on `-collage-build` render to `<dir>` instead of serving |

`collage build` needs nothing from `main.go`: compiling is something `go build`
does without being told anything.

The scaffolded version, trimmed to the part that matters:

```go
func main() {
	buildFlag := flag.Bool("collage-build", false, "render the app to static files instead of serving it")
	outFlag := flag.String("out", "dist", "output directory for -collage-build")
	cleanFlag := flag.Bool("clean", false, "remove -out's existing contents before building")
	portFlag := flag.Int("port", envInt("PORT", 3000), "port to listen on (env PORT)")
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
`collage export` and your plugins' commands stop doing anything useful in your
project.

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
unclaimed word is a usage error rather than a server started by accident. A
program that would rather serve when no command matches can check
`errors.Is(err, collage.ErrUnknownCommand)` and carry on instead of exiting.
`app.Commands()` lists what the plugins registered, if you want to print your own
usage. Dispatching starts the application, which closes registration, and a later
`ListenAndServe` reuses that start. Writing such a command is covered in
[Writing a plugin](/docs/writing-plugins#commands).
