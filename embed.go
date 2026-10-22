//go:build !collage_dev

package main

import "embed"

// Templates, static files and the i18n catalogs are embedded, so this binary
// runs from anywhere: a container with a different WORKDIR, a systemd unit, a
// copy on a server.
//
// Only outside development: "collage dev" builds with the collage_dev tag, which
// leaves this file out for embed_dev.go, and development mode reads the
// directories on disk. Every compile of a package that embeds stores another copy
// of what it embeds in the Go build cache, and in development that is every save.

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS

//go:embed all:locales
var localesFS embed.FS
