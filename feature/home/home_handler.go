package home

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/domain/user"
	homeweb "github.com/deltegui/owl-starter/feature/home/web"
	"github.com/deltegui/owl-starter/web/layout"
	selectui "github.com/deltegui/owl-starter/web/ui/select"
)

func SelectUserAjaxIsland(lang string, items []selectui.Option) island.Instance {
	return island.Mount(selectui.Select, selectui.State{
		Name:      "users",
		Endpoint:  "/home/selectUsersAjax",
		Options:   items,
		Clearable: true,
		Lang:      lang,
	})
}

func SelectGenderAjaxIsland(lang string, items []selectui.Option) island.Instance {
	return island.Mount(selectui.Select, selectui.State{
		Name:      "genders",
		Endpoint:  "/home/selectGenderAjax",
		Options:   items,
		Clearable: true,
		Lang:      lang,
	})
}

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

var selectUserItems = []selectui.Option{
	{
		Label: "Diego",
		Value: "0",
	},
	{
		Label: "Sandra",
		Value: "1",
	},
}

var selectGenderItems = []selectui.Option{
	{
		Label: "Masculino",
		Value: "0",
	},
	{
		Label: "Fememnino",
		Value: "1",
	},
	{
		Label: "Hetero-curioso",
		Value: "2",
	},
	{
		Label: "Menos-culino",
		Value: "3",
	},
	{
		Label: "Fluido",
		Value: "4",
	},
	{
		Label: "Helicoptero",
		Value: "5",
	},
}

func handleUsersSelectAjax() owl.Handler { return handleSelectAjax(selectUserItems) }

func handleGenderSelectAjax() owl.Handler { return handleSelectAjax(selectGenderItems) }

// handleSelectAjax answers GET ?q=text with the items that match it, as a JSON
// array of options. Filter does not tell capital letters or accents apart.
func handleSelectAjax(items []selectui.Option) owl.Handler {
	return func(ctx owl.Ctx) error {
		found := selectui.Filter(items, ctx.Req.URL.Query().Get("q"))
		if found == nil {
			found = []selectui.Option{}
		}
		return ctx.Json(found)
	}
}

func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		loc := ctx.Texts(localizations)
		who, _ := ctx.User()
		counter := homeweb.CounterIsland(homeweb.CounterState{
			Count: 20,
			Lang:  ctx.GetCurrentLanguage(),
		})
		selectAjaxUsers := SelectUserAjaxIsland(ctx.GetCurrentLanguage(), selectUserItems)
		selectAjaxGender := SelectGenderAjaxIsland(ctx.GetCurrentLanguage(), selectGenderItems)
		demoNodes, demoIslands := selectDemos(ctx.GetCurrentLanguage())
		dateNodes, dateIslands := dateDemos(ctx.GetCurrentLanguage())
		reportTabs := island.Mount(homeweb.ReportTabs, homeweb.ReportTabsState{Lang: ctx.GetCurrentLanguage()})
		return layout.Render(ctx, layout.Page{
			Title: loc.Get("Title"),
			Content: hx.Div().Body(
				hx.H1().Text(loc.GetFormatted("Greeting", who.Claims[user.ClaimName])),
				hx.P().Class("lead").Text(loc.Get("Body")),
				counter.Node(),
				selectAjaxUsers.Node(),
				selectAjaxGender.Node(),
				reportTabs.Node(),
				hx.Div().Class("mt-4").Body(demoNodes...),
				hx.Div().Class("mt-4").Body(dateNodes...),
			),
			Islands: append([]island.Instance{
				counter,
				selectAjaxUsers,
				selectAjaxGender,
				reportTabs,
			}, append(demoIslands, dateIslands...)...),
		})
	}
}
