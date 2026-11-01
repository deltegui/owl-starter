package home

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/island"

	selectui "github.com/deltegui/owl-starter/web/ui/select"
)

var cities = []selectui.Option{
	{Value: "mad", Label: "Madrid"}, {Value: "bcn", Label: "Barcelona"}, {Value: "val", Label: "Valencia"},
	{Value: "sev", Label: "Sevilla"}, {Value: "bil", Label: "Bilbao"}, {Value: "mal", Label: "Málaga"},
	{Value: "zar", Label: "Zaragoza"}, {Value: "cor", Label: "Córdoba"},
}

// people stands for something that is too big to send: the server looks in it
// for what the user typed and answers a few.
var people = []selectui.Option{
	{Value: "1", Label: "Ana García"}, {Value: "2", Label: "Álvaro Pérez"}, {Value: "3", Label: "Beatriz Ruiz"},
	{Value: "4", Label: "Carlos Gómez"}, {Value: "5", Label: "Clara Sanz"}, {Value: "6", Label: "Diego López"},
	{Value: "7", Label: "Elena Martín"}, {Value: "8", Label: "Félix Núñez"}, {Value: "9", Label: "Gloria Díaz"},
	{Value: "10", Label: "Héctor Mora"}, {Value: "11", Label: "Inés Vega"}, {Value: "12", Label: "José Romero"},
	{Value: "13", Label: "Julia Castro"}, {Value: "14", Label: "Laura Ortiz"}, {Value: "15", Label: "Luis Marín"},
	{Value: "16", Label: "Marta Gil"}, {Value: "17", Label: "Nuria Ramos"}, {Value: "18", Label: "Óscar Blanco"},
	{Value: "19", Label: "Pablo Herrero"}, {Value: "20", Label: "Raquel Soto"}, {Value: "21", Label: "Sandra Prieto"},
	{Value: "22", Label: "Tomás Cano"}, {Value: "23", Label: "Úrsula Vidal"}, {Value: "24", Label: "Víctor Peña"},
}

const peopleLimit = 10

// handlePeople answers GET /home/people?q=text with the people that match, as
// many as peopleLimit: the limit is the query's, not the component's.
func handlePeople() owl.Handler {
	return func(ctx owl.Ctx) error {
		found := selectui.Filter(people, ctx.Req.URL.Query().Get("q"))
		if len(found) > peopleLimit {
			found = found[:peopleLimit]
		}
		if found == nil {
			found = []selectui.Option{}
		}
		return ctx.Json(found)
	}
}

// selectDemos are four selects of the same island: what is in the page, a choice
// of several, and the two that ask the server.
func selectDemos(lang string) (nodes []hx.INode, islands []island.Instance) {
	add := func(label string, state selectui.State) {
		state.Lang = lang
		in := island.Mount(selectui.Select, state)
		nodes = append(nodes, hx.Div().Class("mb-3").Body(
			hx.El("label").Class("form-label").Text(label), in.Node()))
		islands = append(islands, in)
	}
	add("Select (options in the page)", selectui.State{
		Name: "city", Options: cities, Selected: cities[:1], Clearable: true})
	add("Multiple (options in the page)", selectui.State{
		Name: "cities", Multiple: true, Options: cities, Selected: cities[1:3]})
	add("Ask the server (the first ones come with the page)", selectui.State{
		Name: "person", Endpoint: "/home/people", Options: people[:peopleLimit], Clearable: true})
	add("Multiple, ask the server (from 2 letters)", selectui.State{
		Name: "team", Multiple: true, Endpoint: "/home/people", MinChars: 2})
	return nodes, islands
}
