//go:build collage_dev

package content

import "embed"

// FS is empty in a development build: see embed.go.
var FS embed.FS
