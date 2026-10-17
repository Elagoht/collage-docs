---
description: What a repository needs for collage plugin add to install it — a name, a manifest, a New with no setup, a release tag — and how collage plugin check proves it.
reference: Plugin, Host, FindPlugin, Configurer, PluginConfig
---

# Writing a plugin: the repository contract

[Writing a plugin](/docs/writing-plugins) is about the Go side: the `Plugin`
interface and its hooks. This page is about the repository around it. A plugin
that meets the rules below can be installed with one command, by anyone, and
removed with another:

```sh
collage plugin add you/stamp      # github.com/you/collage-stamp
collage plugin del you/stamp
```

`collage plugin add` refuses a module that breaks a rule, says which one, and
points here. [`collage plugin check`](#check-it-before-you-tag) runs the same
rules against your repository before anyone else does.

## The rules

### The repository is named collage-name

A short reference, `you/stamp`, means `github.com/you/collage-stamp`. Name the
repository `collage-<name>` and it can be added that way. A plugin hosted
elsewhere, or named otherwise, is added by its module path
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

It returns exactly one value, a `collage.Plugin`. Refused: two or more
parameters, a basic type such as `string`, a struct type from another package
(whose zero value the user's file could not always write), and an unexported
type, which no other package can name.

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
middleware. `add` puts the new entry before the first existing one whose order
is greater, and never moves an entry that is already there. The published
plugins that ask for a place:

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
URL), the empty section is missing it. That is not a contract failure, since
`New` succeeded. Refuse in `Init`, in words the user can act on:

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
application's plugins and returns the plugin of type `T` (the first, if there are several), from `Configure` or
from `Init`:

```go
live, ok := collage.FindPlugin[*live.Plugin](host)
if !ok {
	return errors.New("you/stamp needs elagoht/live in the plugin list")
}
```

It sees every plugin in the list, whether that one has been configured or
initialised yet or not. Then name the other plugin in `requires`, so that `add`
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

`-run` goes one step further. It builds a small throwaway program that calls
`New` exactly as `add` would write it, runs that call, and compares the plugin's
`Name()` with the manifest's `name`. It fails on a build error, a panic in
`New`, or a name that differs. It does **not** start an application, call
`Configure` or `Init`, or read any configuration; so a setting an empty section
lacks is not something it reports. The program's files are removed afterwards,
and `go.mod` and `go.sum` put back.

The published plugins run `collage plugin check -run` on every release.
Do the same in your CI, and a plugin that passes can be added by anyone with one
line.
