// Package homeweb is what the home feature shows that also runs in the browser.
// It compiles to both sides, so it must not import the owl root (that would pull
// net/http and the router into the WASM): only owl/island, localizer and Hix.
package homeweb

import (
	"strconv"

	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/locale"
)

// CounterState is what the server gives the island: where it starts and the
// language, so the browser builds the same words. Only exported fields with a
// json tag travel.
type CounterState struct {
	Count int    `json:"count"`
	Lang  string `json:"l"`
}

var counterTexts = localizer.Dict{
	localizer.LangEnglish: {
		"Title": "Counter",
		"Less":  "Less",
		"More":  "More",
	},
	localizer.LangSpanish: {
		"Title": "Contador",
		"Less":  "Menos",
		"More":  "Más",
	},
}

// Counter is an island: the server renders it from its state and the browser
// hydrates that HTML and makes it interactive. The builder runs on both sides
// with the same state, so it must not read the clock, random numbers or window.
var Counter = island.Define("counter", func(s CounterState) hx.INode {
	t := localizer.Pick(counterTexts, s.Lang, locale.Languages[0])
	count := hx.Signal(s.Count)

	value := hx.Span().Class("fs-4", "mx-3")
	value.Effect(func() {
		value.Text(strconv.Itoa(count.Get()))
	})

	less := bs.Button(t.Get("Less"), bs.Outline(bs.Secondary))
	less.OnClick(func(hx.EventContext) {
		count.Set(count.Get() - 1)
	})

	more := bs.Button(t.Get("More"), bs.Primary)
	more.OnClick(func(hx.EventContext) {
		count.Set(count.Get() + 1)
	})

	return bs.Card(bs.CardBody(
		bs.CardTitle(t.Get("Title")),
		hx.Div().Class("d-flex", "align-items-center").Body(less, value, more)))
})

func CounterIsland(state CounterState) island.Instance {
	return island.Mount(Counter, state)
}
