package templates

import "embed"

//go:embed *.html **/*.html
var FS embed.FS
