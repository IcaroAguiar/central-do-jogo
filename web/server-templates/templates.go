// Package servertemplates embeds the Go SSR HTML templates (PAT-004) consumed
// by internal/platform/render. Keeping the embed directive next to the
// templates lets the render package ship them inside the compiled binary,
// with no runtime file dependency in the distroless deploy image.
package servertemplates

import "embed"

// FS holds the embedded *.tmpl files.
//
//go:embed *.tmpl
var FS embed.FS
