//go:build js && wasm

// Command app is the one WASM of the whole app: it carries every island, and each
// page hydrates the ones it has (a page without islands does not load it). To add
// an island, define it in the web package of its feature and list it here.
package main

import (
	"github.com/deltegui/owl/v3/island"

	homeweb "github.com/deltegui/owl-starter/feature/home/web"
)

func main() {
	island.Run(
		homeweb.Counter,
		homeweb.SelectAjax,
	)
}
