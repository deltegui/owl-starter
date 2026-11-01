package account

import (
	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/layout"
)

var localization = localizer.Dict{
	localizer.LangEnglish: {
		"Title":    "Log in",
		"Email":    "Email",
		"Password": "Password",
		"Submit":   "Log in",
		"Invalid":  "Wrong email or password.",
	},
	localizer.LangSpanish: {
		"Title":    "Iniciar sesión",
		"Email":    "Correo",
		"Password": "Contraseña",
		"Submit":   "Entrar",
		"Invalid":  "Correo o contraseña incorrectos.",
	},
}

type loginViewModel struct {
	form      loginForm
	isInvalid bool
}

type loginForm struct {
	Email    string `valtruc:"required, contains=@"`
	Password string `valtruc:"required"`
}

func renderLogin(ctx owl.Ctx, vm loginViewModel) error {
	loc := ctx.Texts(localization)
	return layout.Render(ctx, layout.Page{
		Title: loc.Get("Title"),
		Content: hx.Div().Class("row", "justify-content-center").Body(
			hx.Div().Class("col-md-5").Body(
				hx.H1().Class("h3", "mb-4").Text(loc.Get("Title")),
				loginFormView(ctx, vm))),
	})
}

func loginFormView(ctx owl.Ctx, vm loginViewModel) hx.INode {
	vd := hxui.NewViewData(ctx, localization)
	alert := hx.Noop()
	if vm.isInvalid {
		alert = bs.Alert(hx.Span().Text(vd.T("Invalid")), bs.Danger)
	}
	return hx.Form().Attribute("method", "post").Attribute("action", "/login").Body(
		vd.CsrfInput(),
		alert,
		field(vd, "Email", "email", vm.form.Email),
		field(vd, "Password", "password", ""),
		bs.SubmitButton(vd.T("Submit"), bs.Primary))
}

func field(vd hxui.ViewData, name, kind, value string) hx.INode {
	input := hx.Input().Type(kind).Class("form-control").
		ClassIf("is-invalid", vd.HaveFormError(name)).
		Attribute("id", name).Attribute("name", name)
	if value != "" {
		input.Value(value)
	}
	return hx.Div().Class("mb-3").Body(
		hx.Label().Class("form-label").Attribute("for", name).Text(vd.T(name)),
		input,
		hx.Div().Class("invalid-feedback").Text(vd.FormError(name)))
}
