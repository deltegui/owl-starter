// Package layout is the page around the features: the top bar, the language
// buttons and the logout button. Written with Hix and rendered by ctx.View.
package layout

import (
	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/island"

	"github.com/deltegui/owl-starter/web/locale"
)

// Name is the name of the app, set from the configuration when it starts. It
// is the brand of the top bar and the end of every title.
var Name = "app"

// Bundle is the one WASM of the app (client/app). A page loads it only if it has islands.
const Bundle = "app"

// Page is what a handler gives to Render.
type Page struct {
	Title   string // already translated
	Content hx.INode
	// Islands are the ones the page has, from island.Mount. Their state travels
	// with the page; put their elements in Content with Instance.Node().
	Islands []island.Instance
}

func Render(ctx owl.Ctx, p Page) error {
	return ctx.View(func(ctx owl.Ctx) hx.INode {
		return build(ctx, p)
	})
}

func build(ctx owl.Ctx, p Page) hx.INode {
	vd := hxui.NewViewData(ctx, locale.Shared)
	opts := hxui.AppOptions{
		Title: p.Title + " · " + Name,
		Head: []hx.INode{
			hxui.Stylesheet(bs.CSS),
			hx.Link().Attribute("rel", "icon").Href("/favicon.svg"),
		},
		Islands: p.Islands,
	}
	if len(p.Islands) > 0 {
		opts.Bundle = Bundle
	}
	return hxui.App(ctx, opts,
		topbar(ctx, vd),
		bs.Container(p.Content).Class("py-4"))
}

func topbar(ctx owl.Ctx, vd hxui.ViewData) hx.INode {
	right := []hx.INode{languages(ctx, vd)}
	if ctx.HaveSession() {
		right = append(right, hx.Form().Attribute("method", "post").Attribute("action", "/logout").Class("ms-3").Body(
			vd.CsrfInput(),
			bs.SubmitButton(vd.T("Shared.Logout"), bs.Outline(bs.Light), bs.Sm)))
	}
	return hx.Nav().Class("navbar", "navbar-dark", "bg-dark", "px-3").Body(
		hx.A().Href("/").Class("navbar-brand").Text(Name),
		hx.Div().Class("d-flex", "align-items-center").BodyList(right))
}

// languages are forms, so they work without JavaScript, and carry the CSRF token.
func languages(ctx owl.Ctx, vd hxui.ViewData) hx.INode {
	buttons := []hx.INode{}
	for _, lang := range locale.Languages {
		current := ctx.GetCurrentLanguage() == lang
		buttons = append(buttons, hx.Form().Attribute("method", "post").Attribute("action", "/language/"+lang).Class("d-inline").Body(
			vd.CsrfInput(),
			hx.Input().Type("hidden").Attribute("name", "Next").Value(ctx.Req.URL.RequestURI()),
			hx.Button().Type("submit").Class("btn", "btn-sm", "me-1").
				ClassIf("btn-light", current).ClassIf("btn-outline-light", !current).
				Text(vd.T("Shared."+lang))))
	}
	return hx.Div().BodyList(buttons)
}
