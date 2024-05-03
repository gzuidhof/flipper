// Package static provides the static assets for the web views.
package static

import (
	"embed"
	"io/fs"
)

//go:embed assets/*
var assetsFS embed.FS

// EmbeddedFS returns the embedded filesystem for the static assets.
func EmbeddedFS() fs.FS {
	// Return embedded fs without assets prefix
	fs, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic(err)
	}

	return fs
}
