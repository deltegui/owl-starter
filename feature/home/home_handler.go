package home

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/domain/user"
	"github.com/deltegui/owl-starter/web/layout"
)

var localizations = localizer.Dict{
	localizer.LangEnglish: {
		"Title":    "Home",
		"Greeting": "Hello, %s",
		"Body":     "Your app starts here.",
	},
	localizer.LangSpanish: {
		"Title":    "Inicio",
		"Greeting": "Hola, %s",
		"Body":     "Tu aplicación empieza aquí.",
	},
}

func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		loc := ctx.Texts(localizations)
		who, _ := ctx.User()
		return layout.Render(ctx, layout.Page{
			Title: loc.Get("Title"),
			Content: hx.Div().Body(
				hx.H1().Text(loc.GetFormatted("Greeting", who.Claims[user.ClaimName])),
				hx.P().Class("lead").Text(loc.Get("Body"))),
		})
	}
}
