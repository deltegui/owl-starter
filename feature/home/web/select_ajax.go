package homeweb

import (
	"context"
	"strings"

	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/island/client"
)

type SelectAjaxItem struct {
	Value string
	Label string
}

type SelectAajaxState struct {
	Items          []SelectAjaxItem
	ID             string
	SourceEndpoint string
}

type SelectAjaxSeach struct {
	Search string
}

var SelectAjax = island.Define("select-ajax", func(state SelectAajaxState) hx.INode {
	stateToBsOptions := func(st SelectAajaxState) []bs.Option {
		options := make([]bs.Option, len(st.Items))
		for i, element := range st.Items {
			options[i].Label = element.Label
			options[i].Value = element.Value
		}
		return options
	}

	search := hx.Signal("")
	items := hx.Signal(stateToBsOptions(state))

	doSearch := hx.Action(func(ctx context.Context, search string) (SelectAajaxState, error) {
		return client.Send[SelectAajaxState](ctx, "POST", state.SourceEndpoint, SelectAjaxSeach{search})
	})
	doSearch.OnSuccess(func(result SelectAajaxState, arg string) {
		items.Set(stateToBsOptions(result))
	})
	hx.EffectFunc(func() {
		doSearch.Run(search.Get())
	})

	listText := hx.Computed(func() string {
		it := items.Get()
		strs := make([]string, len(it))
		for index, i := range it {
			strs[index] = i.Label
		}
		return strings.Join(strs, ", ")
	})

	return hx.Div().Body(
		bs.Input("text", state.ID+"-searchbox", search),
		hx.Div().BindText(listText),
		hx.Dynamic(func() hx.INode {
			options := items.Get()
			nodes := make([]hx.INode, len(options))
			for _, opt := range options {
				nodes = append(nodes, hx.Option().Text(opt.Label).Value(opt.Value))
			}
			return hx.Select().BodyList(nodes)
		}),
	)
})
