---
description: What a repository needs for collage plugin add to install it — a name, a manifest, a New with no setup, a release tag — and how collage plugin check proves it.
reference: Plugin, Host, FindPlugin, Configurer, PluginConfig
---

# Writing a plugin: the repository contract

[Writing a plugin](/docs/writing-plugins) is about the Go side: the `Plugin`
interface and its hooks. This page is about the repository around it.

A plugin that meets the rules below can be installed with one command, by anyone,
and removed with another:

```sh
collage plugin add you/stamp      # github.com/you/collage-stamp
collage plugin del you/stamp
```

`collage plugin add` refuses a module that breaks a rule, says which one, and
points here. [`collage plugin check`](#check-it-before-you-tag) runs the same
rules against your repository before anyone else does.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 520" role="img" aria-labelledby="pa-contract-title">
<title id="pa-contract-title">From a plugin's repository to an installed plugin. The repository github.com/you/collage-stamp holds go.mod with its module path, collage.json naming you/stamp, a root package stamp whose func New takes Options, and a release tag v0.1.0 at HEAD. collage plugin check reads collage.json, the root package and the shape of New; with -release it wants a release tag at HEAD, and with -run it calls New and compares Name() with the manifest's name. Then, in a user's project, collage plugin add you/stamp fetches the module with go get, checks the same rules, wants the name you/stamp and a version that is not a pseudo-version, and adds the plugins it requires first. It writes stamp.New(stamp.Options{}) into plugins.go at the place its order gives, an empty you/stamp section into plugins-config.json, and prints the setup note if there is one.</title>
<defs><marker id="pa-contract-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="8" width="312" height="144" rx="8" />
<text x="14" y="25">the plugin's repository</text>
<rect class="d-folder" x="14" y="34" width="292" height="24" rx="4" /><text class="d-small" x="160" y="50" text-anchor="middle">go.mod · module github.com/you/collage-stamp</text>
<rect class="d-folder" x="14" y="62" width="292" height="24" rx="4" /><text class="d-small" x="160" y="78" text-anchor="middle">collage.json · "name": "you/stamp"</text>
<rect class="d-folder" x="14" y="90" width="292" height="24" rx="4" /><text class="d-small" x="160" y="106" text-anchor="middle">stamp.go · package stamp · func New(Options)</text>
<rect class="d-folder" x="14" y="118" width="292" height="24" rx="4" /><text class="d-small" x="160" y="134" text-anchor="middle">git tag v0.1.0 at HEAD, pushed</text>
<line class="d-flow" x1="160" y1="152" x2="160" y2="166" marker-end="url(#pa-contract-arrow)" />
<rect class="d-cmd" x="46" y="170" width="228" height="24" rx="12" /><text class="d-cmd-text" x="160" y="186" text-anchor="middle">collage plugin check -release -run</text>
<line class="d-flow" x1="160" y1="194" x2="160" y2="206" marker-end="url(#pa-contract-arrow)" />
<rect class="d-soft-a" x="14" y="210" width="292" height="58" rx="6" /><text class="d-small" x="160" y="226" text-anchor="middle">collage.json, the root package, New's shape</text><text class="d-small" x="160" y="241" text-anchor="middle">-release: a release tag at HEAD (git)</text><text class="d-small" x="160" y="256" text-anchor="middle">-run: calls New; Name() must equal name</text>
<line class="d-rule" x1="6" y1="284" x2="314" y2="284" />
<text class="d-small" x="160" y="300" text-anchor="middle">then, in a user's project</text>
<rect class="d-cmd" x="60" y="310" width="200" height="24" rx="12" /><text class="d-cmd-text" x="160" y="326" text-anchor="middle">collage plugin add you/stamp</text>
<line class="d-flow" x1="160" y1="334" x2="160" y2="346" marker-end="url(#pa-contract-arrow)" />
<rect class="d-soft-a" x="14" y="350" width="292" height="58" rx="6" /><text class="d-small" x="160" y="366" text-anchor="middle">go get @upgrade, then the same rules</text><text class="d-small" x="160" y="381" text-anchor="middle">name must be you/stamp; no pseudo-version</text><text class="d-small" x="160" y="396" text-anchor="middle">the plugins it requires are added first</text>
<line class="d-flow" x1="160" y1="408" x2="160" y2="420" marker-end="url(#pa-contract-arrow)" />
<rect class="d-folder" x="14" y="424" width="292" height="24" rx="4" /><text class="d-small" x="160" y="440" text-anchor="middle">plugins.go · stamp.New(stamp.Options{}), by order</text>
<rect class="d-folder" x="14" y="452" width="292" height="24" rx="4" /><text class="d-small" x="160" y="468" text-anchor="middle">plugins-config.json · "you/stamp": {}</text>
<line class="d-flow" x1="160" y1="476" x2="160" y2="488" marker-end="url(#pa-contract-arrow)" />
<rect class="d-pill" x="60" y="492" width="200" height="24" rx="12" /><text class="d-small" x="160" y="508" text-anchor="middle">the setup note, if any, printed</text>
</svg>
<figcaption>From a repository to an installed plugin. The repository holds a <code>go.mod</code> whose module path is <code>github.com/you/collage-stamp</code>, a <code>collage.json</code> naming the plugin <code>you/stamp</code>, a root package with a <code>New</code> that needs no configuration, and a release tag. <code>collage plugin check</code> reads the manifest, the package and the shape of <code>New</code>; <code>-release</code> also wants a release tag at <code>HEAD</code>, and <code>-run</code> calls <code>New</code> and compares its <code>Name()</code> with the manifest. A user's <code>collage plugin add you/stamp</code> fetches the module, applies the same rules, wants the manifest's name to be the one typed and the version to be a release, adds the plugins it requires first, then writes the call into <code>plugins.go</code> at the place its <code>order</code> gives and an empty section into <code>plugins-config.json</code>.</figcaption>
</figure>

## The rules

### The repository is named collage-name

A short reference, `you/stamp`, means `github.com/you/collage-stamp`. Name the
repository `collage-<name>` and it can be added that way.

A plugin hosted elsewhere, or named otherwise, is added by its module path
(`collage plugin add gitlab.com/you/stamper`); the other rules still hold.

### The root has a collage.json

The manifest sits at the module root, and its `name` is `owner/name`, each part
lowercase letters, digits and hyphens:

```json
{
  "name": "you/stamp",
  "description": "Adds a build stamp to every page.",
  "repository": "https://github.com/you/collage-stamp"
}
```

That name is also what your plugin's `Name()` returns, and the key of its section
in `plugins-config.json`. For a short reference the manifest's name must equal
the one typed: `collage plugin add you/stamp` needs `"name": "you/stamp"`.

The manifest is read strictly: a key the schema does not know is refused. The
schema ships with [the editor extension](/docs/installation#editor-support),
which completes the keys and checks the file as you type.

### The plugin is the module root

The package that holds `New` is the root package of the module, so its import
path is the module path. It cannot be `package main`, since nothing can import
that. Subpackages are fine for the rest of your code.

### New needs no configuration

`add` writes a call to `New` into the user's `plugins.go` without knowing what
your options are. So `New` has to be a function that can be called with nothing
set up. One of these shapes:

| `New` takes | `add` writes |
| --- | --- |
| nothing | `stamp.New()` |
| only a variadic parameter | `stamp.New()` |
| one struct type of your package | `stamp.New(stamp.Options{})` |
| one pointer, interface, slice, map or func | `stamp.New(nil)` |

It returns exactly one value, a `collage.Plugin`. Refused:

- two or more parameters;
- a basic type such as `string`;
- a struct type from another package, whose zero value the user's file could not
  always write;
- an unexported type, which no other package can name.

The call has to succeed. What the zero value means is up to you: it should be a
working plugin with defaults. A setting with no sensible default is a different
case, covered [below](#a-required-setting-fails-in-init).

### There is a release tag

The version `add` resolves is a semver tag, `v0.1.0`. A repository with no tag
only has a pseudo-version (`v0.0.0-2026…-abcdef`), which is refused: there would
be nothing stable to pin.

## Optional manifest fields

Three fields tell `add` more about where and how your plugin goes.

**`requires`** is a list of plugin names the project must also have. `add` adds
each missing one first, and `del` refuses to remove one while another plugin
requires it. The websocket plugin requires `["elagoht/live"]`.

**`order`** is an integer, 0 when absent, that says where in `plugins.go` the
entry goes. Lower is earlier, and earlier in `Config.Plugins` is the outer
middleware.

`add` puts the new entry before the first existing one whose order is greater,
and never moves an entry that is already there. The published plugins that ask
for a place:

| Plugin | Order | Why |
| --- | --- | --- |
| elagoht/health | −30 | before anything that refuses or answers a request |
| elagoht/fail2ban | −25 | right after health |
| elagoht/compress | −20 | before any plugin that rewrites response bodies |
| elagoht/websocket | −1 | configured before elagoht/live |
| elagoht/devtoolbar | 100 | last, so it sees the other plugins' findings |

Every other plugin is 0. Leave the field out unless position matters.

**`setup`** is a short note printed after `add` succeeds, for what only Go code
can supply. The feed plugin says its feeds are passed to `feed.New(...)` in
`plugins.go`.

```json
{
  "name": "you/stamp",
  "requires": ["elagoht/live"],
  "order": -5,
  "setup": "Pass your stamps to stamp.New(...) in plugins.go."
}
```

## A required setting fails in Init

`add` writes the section `"you/stamp": {}` to `plugins-config.json`: the
defaults. If your plugin cannot work without a setting (an API key, a site
URL), the empty section is missing it.

That is not a contract failure, since `New` succeeded. Refuse in `Init`, in words
the user can act on:

```go
func (p *Plugin) Init(ctx context.Context, host collage.Host) error {
	if p.cfg.APIKey == "" {
		return errors.New(`you/stamp: "apiKey" is required; set it in the "you/stamp" section of plugins-config.json`)
	}
	return nil
}
```

The user adds the plugin, starts the application, reads that message and fills in
the section. Nothing else is needed from you.

## A plugin that needs another

When your plugin works through another one, find it by type instead of asking
the user to pass it in. `collage.FindPlugin[T](host)` looks through the
application's plugins and returns the plugin of type `T` (the first, if there are
several). Call it from `Configure` and keep what it finds: since v0.78.0 the `Host`
given to `Init` lists no plugins, so `FindPlugin` from `Init` reports false.

```go
live, ok := collage.FindPlugin[*live.Plugin](host)
if !ok {
	return errors.New("you/stamp needs elagoht/live in the plugin list")
}
```

It sees every plugin in the list, whether that one has been configured yet or
not. Then name the other plugin in `requires`, so that `add`
brings it along. The websocket plugin does exactly this, which is why
`New(nil)` works for it.

## Check it before you tag

Run it inside your repository:

```sh
collage plugin check              # the manifest and New
collage plugin check -release     # also: a release tag at HEAD (needs git)
collage plugin check -run         # also: compile and run the call
collage plugin check -release -run ./path/to/repo
```

Without a flag it reads the repository and checks the manifest, the name, the
package and the shape of `New`. `-release` wants a release tag at `HEAD`.

`-run` goes one step further:

- It builds a small throwaway program that calls `New` exactly as `add` would
  write it, runs that call, and compares the plugin's `Name()` with the
  manifest's `name`.
- It fails on a build error, a panic in `New`, or a name that differs.
- It does **not** start an application, call `Configure` or `Init`, or read any
  configuration; so a setting an empty section lacks is not something it reports.
- The program's files are removed afterwards, and `go.mod` and `go.sum` put back.

`check` and `add` share the rules but not every step. `check` has no short
reference to compare the name with, and it looks at the tags on your `HEAD`,
where `add` looks at the version it fetched:

| Rule | `collage plugin check` | `collage plugin add` |
| --- | --- | --- |
| `collage.json` and its `name` | yes | yes, and for a short reference the name must be the one typed |
| The root package and the shape of `New` | yes | yes |
| `New` runs and `Name()` matches | with `-run` | no |
| A release | with `-release`: a semver tag at `HEAD` | the version fetched is not a pseudo-version |

The published plugins run `collage plugin check -run` on every release.
Do the same in your CI, and a plugin that passes can be added by anyone with one
line.

## Example: publishing a reading-time plugin

A blog wants "3 min read" under each post's title. Here is that feature as a
plugin, `you/readtime`, from an empty repository to a line in someone's
`plugins.go`.

### The repository

The repository is named `collage-readtime`, so the short reference
`you/readtime` finds it. Everything sits at the module root:

```text
collage-readtime/
├── collage.json
├── go.mod          module github.com/you/collage-readtime
├── go.sum
└── readtime.go     package readtime
```

The manifest names the plugin; the name is what `Name()` returns and the key of
its section:

```json
{
  "name": "you/readtime",
  "description": "Adds a readTime template function: minutes to read a text."
}
```

The package takes one exported struct of options, so `add` will write
`readtime.New(readtime.Options{})`. The zero `Options` is a working plugin: a
reading speed of 0 means the default, 200 words a minute.

```go
// Package readtime adds a readTime template function: the minutes it takes to
// read a text.
package readtime

import (
	"context"
	"strings"

	"github.com/Elagoht/collage/pkg/collage"
)

// Options is the plugin's configuration. The zero value is the defaults.
type Options struct {
	// WordsPerMinute is the reading speed; 200 when zero.
	WordsPerMinute int `json:"wordsPerMinute"`
}

// Plugin is the readtime plugin. Construct it with New.
type Plugin struct{ opts Options }

// New returns the plugin. The application's plugins-config.json section is
// decoded over opts.
func New(opts Options) *Plugin { return &Plugin{opts: opts} }

var (
	_ collage.Plugin     = (*Plugin)(nil)
	_ collage.Configurer = (*Plugin)(nil)
)

func (p *Plugin) Name() string    { return "you/readtime" }
func (p *Plugin) Version() string { return "0.1.0" }

// Configure reads the section and adds the template function.
func (p *Plugin) Configure(_ context.Context, host collage.ConfigHost) error {
	opts, err := collage.PluginConfig(host, p.opts)
	if err != nil {
		return err
	}
	if opts.WordsPerMinute <= 0 {
		opts.WordsPerMinute = 200
	}
	p.opts = opts
	return host.AddTemplateFunc("readTime", func(text string) int {
		words := len(strings.Fields(text))
		return (words + p.opts.WordsPerMinute - 1) / p.opts.WordsPerMinute
	})
}

func (p *Plugin) Init(context.Context, collage.Host) error { return nil }
func (p *Plugin) Shutdown(context.Context) error           { return nil }
```

### Check, tag, check again

With the files committed, check the contract:

```text
$ collage plugin check
you/readtime: meets the contract
```

Had `New` taken the reading speed as an `int`, the check would have refused it,
with the rule and this page's address:

```text
$ collage plugin check
collage: plugin check: not a collage plugin: New takes a int; it may take none, or one parameter whose zero value works (a struct, pointer, interface, slice, map or func); see https://collage.furkanbaytekin.dev/en/docs/plugin-authors/
```

`-release` wants a tag on the commit you are at, so it fails until you tag:

```text
$ collage plugin check -release
you/readtime: meets the contract
collage: plugin check: no release tag at HEAD; add rule 5 with git tag vX.Y.Z
```

Tag the commit, then run every check, `New` included:

```text
$ git tag v0.1.0
$ collage plugin check -release -run
you/readtime: meets the contract
you/readtime: v0.1.0 at HEAD is a release
you/readtime: New runs, and Name() matches collage.json
```

Push the tag with `git push origin v0.1.0`. Until it is pushed, `go get` cannot
see the release, and `add` has nothing to fetch.

### Install it in a blog

In the blog's project, one command fetches the module, checks it again, and
edits both files:

```text
$ collage plugin add you/readtime
added you/readtime v0.1.0
  plugins.go           readtime.New(readtime.Options{})
  plugins-config.json  "you/readtime": {}
```

`plugins.go` gains the import and the entry:

```go
import (
	"github.com/Elagoht/collage/pkg/collage"

	readtime "github.com/you/collage-readtime"
)

func plugins() []collage.Plugin {
	return []collage.Plugin{
		readtime.New(readtime.Options{}),
	}
}
```

The empty section means the defaults. A blog whose readers are slower fills it
in, and the post template calls the function:

```json
{
  "you/readtime": { "wordsPerMinute": 150 }
}
```

```html
<p>{{readTime .Body}} min read</p>
```

Here `.Body` stands for the post's text, whatever the post's data handler calls
it. `collage plugin del you/readtime` removes the entry, the import and the
section again.
