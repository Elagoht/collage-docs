//go:build collage_dev

package main

import "embed"

// Empty in a development build: see embed.go.
var templatesFS, staticFS, localesFS embed.FS
