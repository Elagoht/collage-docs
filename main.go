// Command collage-docs serves this collage application, or — when invoked with
// -collage-build — renders it to static files instead of serving it.
//
// # The collage CLI contract
//
// collage.Main keeps it: "collage dev" builds this program with the
// collage_dev tag and runs it with COLLAGE_DEV=1; "collage export" runs it with
// -collage-build -out <dir>; "collage inspect", "routes" and "check" run it with
// a word after the flags — collage-inspect, collage-check — which Main hands to
// collage.DispatchCommands. Development mode reads templates, static files,
// catalogs and the pages from disk on every request, so editing those needs no
// restart at all.
//
// "collage build" needs nothing from this file: it compiles the program, which
// is something go build does without being told anything.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/Elagoht/collage/pkg/collage"

	"github.com/Elagoht/collage-docs/content"
	"github.com/Elagoht/collage-docs/site"
)

func main() {
	collage.Main(newApp)
}

// newApp builds the application: its configuration, its routes — registered in
// routes.go — and its static mount.
//
// It is separate from main so that the tests can build the same application
// and drive it through app.Handler(), with no server listening and no port to
// pick, and a cache directory of their own: a disk cache is shared by
// everything that uses the same directory, build and key, so a test would
// otherwise read what the previous test, or the previous run, rendered. What
// they exercise is then the site that actually runs, rather than a second
// wiring that can drift from it.
func newApp(env collage.Env) (*collage.App, error) {
	devMode := env.Dev
	// Plugin configuration, keyed by plugin name. A missing file is not an
	// error: every plugin then runs on its defaults.
	pluginConfig, err := collage.LoadPluginConfig("plugins-config.json")
	if err != nil {
		return nil, fmt.Errorf("plugin configuration: %w", err)
	}

	app, err := collage.New(&collage.Config{
		DevMode: devMode,
		// The pages are Markdown read from content/ in development; a change there
		// reloads the page like a template change does.
		DevWatch: []string{"content"},
		// English at /docs/caching, Turkish at /tr/docs/caching: the URL says the
		// language, and nothing else does.
		Locale: collage.LocaleConfig{
			Default:   site.Original,
			Supported: site.Locales(),
			// English at /en/, like Turkish at /tr/: no language at the bare
			// address. The server redirects "/" to /en/; the export writes the
			// English home at its root too, with /en/ as its canonical address.
			PrefixDefault: true,
		},
		// The export writes each page as <path>/index.html, which GitHub Pages
		// serves at "/docs/caching/" and reaches from "/docs/caching" only by a
		// redirect. With the slash, every link collage builds — canonical,
		// hreflang, sitemap — is the address the host answers.
		TrailingSlash: true,
		// What the sitemap, robots.txt, canonical links and Open Graph tags put in
		// front of a path.
		BaseURL: site.Origin,
		Server: collage.ServerConfig{
			Host: env.Host,
			Port: env.Port,
		},
		Template: collage.TemplateConfig{
			FS:        templatesFS,
			Root:      "templates",
			Extension: ".html",
		},
		Cache: collage.CacheConfig{
			// In development collage never reads from the cache — a cached
			// page would hide the template you just edited — and uses memory
			// instead of disk. In production rendered pages are kept under
			// .cache, namespaced by a hash of this binary.
			Enabled:    true,
			Type:       "disk",
			Dir:        env.CacheDir,
			DefaultTTL: 5 * time.Minute,
		},
		// The plugins are in plugins.go, and their settings in plugins-config.json.
		Plugins:      plugins(devMode),
		PluginConfig: pluginConfig,
		Security: collage.SecurityConfig{
			// Signs the forgery tokens forms carry. Unset, one is generated
			// per process: fine in development, wrong to deploy, because every
			// form submitted before a restart is refused after it. Make one
			// with `openssl rand -hex 32`.
			CSRFKey: []byte(os.Getenv("COLLAGE_CSRF_KEY")),
		},
	})
	if err != nil {
		return nil, err
	}

	docs, err := contentLoader(devMode)
	if err != nil {
		return nil, err
	}

	// Every page, document and action this project has — see routes.go.
	if err := register(app, docs); err != nil {
		return nil, err
	}

	assets, err := staticFiles(devMode)
	if err != nil {
		return nil, err
	}
	if err := app.Mount("/static/", assets); err != nil {
		return nil, fmt.Errorf("mount static files: %w", err)
	}

	return app, nil
}

// contentLoader returns what every page reads the documentation through.
//
// In production the content is embedded and loaded once, here, so a broken page —
// a link to nothing, a page the navigation never reaches — stops the program from
// starting rather than reaching a reader. In development it is read from disk on
// every request, so editing a page shows on the next reload, the way editing a
// template does.
func contentLoader(devMode bool) (func() (*site.Set, error), error) {
	if devMode {
		if _, err := os.Stat("content"); err == nil {
			dir := os.DirFS("content")
			return func() (*site.Set, error) { return loadContent(dir) }, nil
		}
	}
	loaded, err := loadContent(content.FS)
	if err != nil {
		return nil, err
	}
	return func() (*site.Set, error) { return loaded, nil }, nil
}

// loadContent reads the documentation and every translation of it.
func loadContent(fsys fs.FS) (*site.Set, error) {
	return site.LoadSet(fsys, site.Original, site.Translations...)
}

// staticFiles returns the filesystem "/static/" is served from: the embedded
// copy, except in development, where the directory on disk wins so an edited
// stylesheet shows up without a rebuild.
//
// os.OpenRoot rather than os.DirFS: os.DirFS follows a symlink out of the
// directory, and an os.Root does not.
func staticFiles(devMode bool) (fs.FS, error) {
	if devMode {
		if root, err := os.OpenRoot("static"); err == nil {
			return root.FS(), nil
		}
	}

	// fs.Sub, because the embedded tree contains the "static" directory
	// itself: mounting it whole would serve "/static/static/app.css".
	return fs.Sub(staticFS, "static")
}

// locales returns the filesystem the i18n catalogs are read from: the embedded
// copy, except in development, where locales/ on disk wins so an edited catalog
// shows on the next reload.
func locales(devMode bool) fs.FS {
	if devMode {
		if _, err := os.Stat("locales"); err == nil {
			return os.DirFS(".")
		}
	}
	return localesFS
}

// embedded is fsys outside development, and nil in it, for a plugin that reads
// from its FS when one is set and from disk when not: a development build
// embeds nothing (see embed.go).
func embedded(devMode bool, fsys fs.FS) fs.FS {
	if devMode {
		return nil
	}
	return fsys
}
