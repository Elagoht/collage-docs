//go:build !collage_dev

// Package content holds the documentation's pages, embedded so the site builds
// and serves from anywhere. See package site for what the files contain.
package content

import "embed"

// FS is every page and nav.json, and under tr/ their Turkish translation —
// outside development: a "collage dev" build leaves this file out for
// embed_dev.go, and development mode reads content/ from disk.
//
//go:embed *.md nav.json tr/*.md tr/nav.json
var FS embed.FS
