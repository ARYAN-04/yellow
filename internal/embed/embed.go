package embed

import "embed"

// WebDist embeds the static Vite SPA production build files.
//
//go:embed all:dist
var WebDist embed.FS
