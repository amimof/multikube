//go:build !testui_stub

package ui

import "embed"

//go:embed all:web/dist
var WebFS embed.FS
