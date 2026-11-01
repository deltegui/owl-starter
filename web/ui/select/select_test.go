package selectui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/ssr"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/locale"
)

func TestTextsHaveEveryLanguage(t *testing.T) {
	for _, problem := range localizer.Check(texts, locale.Languages...) {
		t.Error(problem)
	}
}

func render(s State) string {
	return ssr.Render(island.Mount(Select, s).Node)
}

var people = []Option{{"1", "José"}, {"2", "Sandra"}, {"3", "Álvaro"}}

func TestServerWritesTheClosedControl(t *testing.T) {
	html := render(State{Name: "who", Options: people, Selected: people[:1], Lang: "en"})
	for _, want := range []string{
		`role="combobox"`, `aria-expanded="false"`, `aria-controls="who-list"`,
		`type="hidden"`, `name="who"`, `value="1"`, `>José<`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("lacks %q in %s", want, html)
		}
	}
	for _, not := range []string{"searchbox", `role="listbox"`, "Sandra"} { // the panel is only built when open
		if strings.Contains(html, not) {
			t.Errorf("a closed select must not have %q: %s", not, html)
		}
	}
}

func TestPlaceholderAndEmptyValue(t *testing.T) {
	html := render(State{Name: "who", Options: people, Lang: "es", Placeholder: "Elige a alguien"})
	if !strings.Contains(html, ">Elige a alguien<") {
		t.Errorf("placeholder: %s", html)
	}
	if !strings.Contains(html, `<input name="who" type="hidden"/>`) { // no value attribute is an empty value
		t.Errorf("a simple select with nothing chosen sends an empty value: %s", html)
	}
	html = render(State{Name: "who", Multiple: true, Options: people, Lang: "es"})
	if strings.Contains(html, `type="hidden"`) || !strings.Contains(html, ">Seleccionar...<") {
		t.Errorf("a multiple one sends nothing, and says the default placeholder: %s", html)
	}
}

func TestMultipleShowsChipsWithRemoveButtons(t *testing.T) {
	html := render(State{Name: "who", Multiple: true, Options: people, Selected: people[:2], Lang: "en"})
	if strings.Count(html, `type="hidden"`) != 2 {
		t.Errorf("one hidden input per value: %s", html)
	}
	if strings.Count(html, "<button") != 2 || !strings.Contains(html, `aria-label="Remove Sandra"`) {
		t.Errorf("each chip has a button to remove it: %s", html)
	}
}

func TestDisabledInvalidClearable(t *testing.T) {
	html := render(State{Name: "who", Options: people, Selected: people[:1], Disabled: true, Invalid: true, Clearable: true, Lang: "en"})
	for _, want := range []string{`aria-disabled="true"`, `is-invalid`, `disabled=""`} {
		if !strings.Contains(html, want) {
			t.Errorf("lacks %q in %s", want, html)
		}
	}
	if strings.Contains(html, "<button") {
		t.Errorf("a disabled select cannot be cleared: %s", html)
	}
	html = render(State{Name: "who", Options: people, Selected: people[:1], Clearable: true, Lang: "en"})
	if !strings.Contains(html, `aria-label="Clear"`) {
		t.Errorf("clear button: %s", html)
	}
}

func TestTwoSelectsInOnePage(t *testing.T) {
	a := island.Mount(Select, State{Name: "a", Options: people, Lang: "en"})
	b := island.Mount(Select, State{Name: "b", Multiple: true, Options: people, Lang: "en"})
	html := ssr.Render(func() hx.INode { return hx.Div().Body(a.Node(), b.Node()) })
	scripts := ssr.Render(func() hx.INode { return hx.Div().Body(a.StateScript(), b.StateScript()) })
	if strings.Count(html, `data-island="select"`) != 2 || strings.Count(scripts, `type="application/json"`) != 2 {
		t.Errorf("two islands of the same component: %s %s", html, scripts)
	}
	if !strings.Contains(html, `aria-controls="a-list"`) || !strings.Contains(html, `aria-controls="b-list"`) {
		t.Errorf("ids come from the name: %s", html)
	}
}

func TestFilterIgnoresCaseAndAccents(t *testing.T) {
	cases := map[string][]string{
		"":       {"José", "Sandra", "Álvaro"},
		"jose":   {"José"},
		"JOSÉ":   {"José"},
		"alvaro": {"Álvaro"},
		" san ":  {"Sandra"},
		"zzz":    nil,
	}
	for text, want := range cases {
		var got []string
		for _, o := range Filter(people, text) {
			got = append(got, o.Label)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Filter(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestToggleDoesNotChangeTheOriginal(t *testing.T) {
	cur := []Option{people[0]}
	added := toggled(cur, people[1])
	if len(added) != 2 || len(cur) != 1 {
		t.Fatalf("added %v, original %v", added, cur)
	}
	if got := toggled(added, people[0]); len(got) != 1 || got[0].Value != "2" || len(added) != 2 {
		t.Fatalf("removed %v, before %v", got, added)
	}
}
