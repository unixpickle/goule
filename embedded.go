package main

import "embed"

// embeddedFiles contains the static assets and templates used by the control
// panel.
//
//go:embed assets templates
var embeddedFiles embed.FS
