package home

import (
	"time"

	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3/island"

	"github.com/deltegui/owl-starter/web/ui/datepicker"
)

// dateDemos are the modes of the date picker, one control each: all of them are
// the same island with another state.
func dateDemos(lang string) (nodes []hx.INode, islands []island.Instance) {
	now := time.Now()
	today := datepicker.Stamp(now) // the wall clock of the user: here, the server's
	month := now.Format("2006-01")

	add := func(label string, state datepicker.State) {
		state.Lang, state.Today = lang, today
		in := island.Mount(datepicker.Picker, state)
		nodes = append(nodes, hx.Div().Class("mb-3").Body(
			hx.El("label").Class("form-label").Text(label), in.Node()))
		islands = append(islands, in)
	}
	add("Date (clearable, only this month)", datepicker.State{
		Name: "date", Clearable: true, Min: month + "-01", Max: now.AddDate(0, 1, -now.Day()).Format("2006-01-02")})
	add("Date and time (minutes every 5)", datepicker.State{
		Name: "moment", Mode: datepicker.ModeDateTime, MinuteStep: 5, Value: now.Format("2006-01-02") + "T09:30"})
	add("Date and time with seconds (weeks start on Sunday)", datepicker.State{
		Name: "exact", Mode: datepicker.ModeDateTime, Seconds: true, SundayFirst: true})
	add("Month", datepicker.State{Name: "month", Mode: datepicker.ModeMonth, Value: month})
	add("Year", datepicker.State{Name: "year", Mode: datepicker.ModeYear, Clearable: true})
	add("Time", datepicker.State{Name: "time", Mode: datepicker.ModeTime, Value: "08:00"})
	add("Range of dates", datepicker.State{Name: "since", NameTo: "until", Range: true, Clearable: true})
	add("Range with time", datepicker.State{Name: "start", NameTo: "end", Range: true, Mode: datepicker.ModeDateTime})
	add("Range of months", datepicker.State{Name: "monthFrom", NameTo: "monthTo", Range: true, Mode: datepicker.ModeMonth})
	return nodes, islands
}
