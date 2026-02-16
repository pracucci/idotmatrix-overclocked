// Package assets provides embedded assets for the iDotMatrix CLI.
package assets

import "embed"

//go:embed emoji/*.gif
var Emoji embed.FS

//go:embed grot/*.gif
var Grot embed.FS

//go:embed preview/*.gif
var Preview embed.FS
