// static.go
package main

import "embed"

//go:embed static/index.html
var content embed.FS
