package home

import (
	"strings"

	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/domain/user"
	homeweb "github.com/deltegui/owl-starter/feature/home/web"
	"github.com/deltegui/owl-starter/web/layout"
)

func SelectUserAjaxIsland(items []homeweb.SelectAjaxItem) island.Instance {
	return island.Mount(homeweb.SelectAjax, homeweb.SelectAajaxState{
		ID:             "users",
		SourceEndpoint: "/home/selectUsersAjax",
		Items:          items,
	})
}

func SelectGenderAjaxIsland(items []homeweb.SelectAjaxItem) island.Instance {
	return island.Mount(homeweb.SelectAjax, homeweb.SelectAajaxState{
		ID:             "genders",
		SourceEndpoint: "/home/selectGenderAjax",
		Items:          items,
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

var selectUserItems = []homeweb.SelectAjaxItem{
	{
		Label: "Diego",
		Value: "0",
	},
	{
		Label: "Sandra",
		Value: "1",
	},
}

var selectGenderItems = []homeweb.SelectAjaxItem{
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

func handleUsersSelectAjax() owl.Handler {
	return func(ctx owl.Ctx) error {
		var vm homeweb.SelectAjaxSeach
		if err := ctx.ParseJson(&vm); err != nil {
			return ctx.BadRequest()
		}
		if vm.Search == "" {
			return ctx.Json(homeweb.SelectAajaxState{
				Items: selectUserItems,
			})
		}

		filtered := []homeweb.SelectAjaxItem{}
		for _, element := range selectUserItems {
			if strings.Contains(element.Label, vm.Search) {
				filtered = append(filtered, element)
			}
		}
		return ctx.Json(homeweb.SelectAajaxState{
			Items: filtered,
		})
	}
}

func handleGenderSelectAjax() owl.Handler {
	return func(ctx owl.Ctx) error {
		var vm homeweb.SelectAjaxSeach
		if err := ctx.ParseJson(&vm); err != nil {
			return ctx.BadRequest()
		}
		if vm.Search == "" {
			return ctx.Json(homeweb.SelectAajaxState{
				Items: selectGenderItems,
			})
		}

		filtered := []homeweb.SelectAjaxItem{}
		for _, element := range selectGenderItems {
			if strings.Contains(element.Label, vm.Search) {
				filtered = append(filtered, element)
			}
		}
		return ctx.Json(homeweb.SelectAajaxState{
			Items: filtered,
		})
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
		selectAjaxUsers := SelectUserAjaxIsland(selectUserItems)
		selectAjaxGender := SelectGenderAjaxIsland(selectGenderItems)
		return layout.Render(ctx, layout.Page{
			Title: loc.Get("Title"),
			Content: hx.Div().Body(
				hx.H1().Text(loc.GetFormatted("Greeting", who.Claims[user.ClaimName])),
				hx.P().Class("lead").Text(loc.Get("Body")),
				counter.Node(),
				selectAjaxUsers.Node(),
				selectAjaxGender.Node(),
			),
			Islands: []island.Instance{
				counter,
				selectAjaxUsers,
				selectAjaxGender,
			},
		})
	}
}
