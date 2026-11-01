package homeweb

import (
	"context"
	"strconv"

	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/island/client"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/locale"
	"github.com/deltegui/owl-starter/web/ui/lazy"
)

// ReportRow and Report are what /home/report answers.
type ReportRow struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Report struct {
	Rows []ReportRow `json:"rows"`
}

type ReportTabsState struct {
	Lang string `json:"l"`
}

var reportTabsTexts = localizer.Dict{
	localizer.LangEnglish: {
		"Summary":     "Summary",
		"Report":      "Report",
		"SummaryText": "The report takes a while: it is only asked for when you open its tab.",
	},
	localizer.LangSpanish: {
		"Summary":     "Resumen",
		"Report":      "Informe",
		"SummaryText": "El informe tarda: solo se pide cuando abres su pestaña.",
	},
}

// ReportTabs shows a lazy component in a tab: the server writes the first tab
// and nothing of the report. When the tab opens, the browser asks for the data
// (it can take seconds), shows a spinner meanwhile and then the report. Going
// back to the tab does not ask again.
var ReportTabs = island.Define("report-tabs", func(s ReportTabsState) hx.INode {
	t := localizer.Pick(reportTabsTexts, s.Lang, locale.Languages[0])
	active := hx.Signal("summary")
	reportOpen := hx.Computed(func() bool { return active.Get() == "report" })

	// Made here, not in the panel: a panel is built again each time its tab is chosen.
	report := lazy.New(lazy.Options[Report]{
		When: reportOpen,
		Fetch: func(ctx context.Context) (Report, error) {
			return client.Send[Report](ctx, "GET", "/home/report", nil)
		},
		View: reportView,
	})

	return hx.Div().Body(
		bs.Tabs(active, []bs.Tab{{ID: "summary", Label: t.Get("Summary")}, {ID: "report", Label: t.Get("Report")}}),
		hx.Div().Class("pt-3").Attribute("style", "min-height: 12rem").Body(
			bs.TabPanel(active, "summary", func() hx.INode { return hx.P().Text(t.Get("SummaryText")) }),
			bs.TabPanel(active, "report", report.Node),
		))
})

func reportView(r Report) hx.INode {
	rows := make([]hx.INode, 0, len(r.Rows))
	for _, row := range r.Rows {
		rows = append(rows, hx.Li().Class("list-group-item", "d-flex", "justify-content-between").Body(
			hx.Span().Text(row.Name), bs.Badge(strconv.Itoa(row.Value))))
	}
	return hx.Ul().Class("list-group").BodyList(rows)
}
