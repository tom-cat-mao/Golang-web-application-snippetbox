package ui

import "embed"

// the comment directive instructs Go to store the files from ui/static folder
// in an embedded filesystem referenced by the global variable Files.
//
//go:embed "static" "html"
var Files embed.FS
