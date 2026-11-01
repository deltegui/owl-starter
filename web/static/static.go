// Package static has the files the server gives as they are (favicon, images,
// and the WASM bundles if the app ever gets islands).
package static

import "embed"

//go:embed *
var Files embed.FS
