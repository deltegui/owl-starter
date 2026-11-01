package datepicker

import (
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

func render(s State) string { return ssr.Render(island.Mount(Picker, s).Node) }

func TestServerWritesTheClosedControl(t *testing.T) {
	html := render(State{Name: "when", Value: "2026-11-01", Lang: "es"})
	for _, want := range []string{
		`type="hidden"`, `name="when"`, `value="2026-11-01"`, `>01/11/2026<`,
		`aria-haspopup="dialog"`, `aria-expanded="false"`, `aria-controls="when-panel"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("lacks %q in %s", want, html)
		}
	}
	if strings.Contains(html, "role=\"dialog\"") || strings.Contains(html, "<table") {
		t.Errorf("a closed control has no panel: %s", html)
	}
}

func TestEveryModeWritesItsValueAndItsText(t *testing.T) {
	cases := []struct {
		name  string
		state State
		value string
		text  string
	}{
		{"year", State{Mode: ModeYear, Value: "2026"}, `value="2026"`, ">2026<"},
		{"month", State{Mode: ModeMonth, Value: "2026-11"}, `value="2026-11"`, ">11/2026<"},
		{"date", State{Mode: ModeDate, Value: "2026-11-01"}, `value="2026-11-01"`, ">01/11/2026<"},
		{"datetime", State{Mode: ModeDateTime, Value: "2026-11-01T14:30"}, `value="2026-11-01T14:30"`, ">01/11/2026 14:30<"},
		{"datetime with seconds", State{Mode: ModeDateTime, Seconds: true, Value: "2026-11-01T14:30:45"}, `value="2026-11-01T14:30:45"`, ">01/11/2026 14:30:45<"},
		{"time", State{Mode: ModeTime, Value: "09:05"}, `value="09:05"`, ">09:05<"},
	}
	for _, c := range cases {
		c.state.Name, c.state.Lang = "x", "es"
		html := render(c.state)
		if !strings.Contains(html, c.value) || !strings.Contains(html, c.text) {
			t.Errorf("%s: want %s and %s in %s", c.name, c.value, c.text, html)
		}
	}
}

func TestEnglishWritesTheMonthFirst(t *testing.T) {
	html := render(State{Name: "x", Value: "2026-11-01", Lang: "en"})
	if !strings.Contains(html, ">11/01/2026<") {
		t.Errorf("english date: %s", html)
	}
}

func TestPlaceholderAndEmptyValue(t *testing.T) {
	html := render(State{Name: "x", Lang: "es"})
	if !strings.Contains(html, ">Elige una fecha<") || !strings.Contains(html, `<input name="x" type="hidden"/>`) {
		t.Errorf("empty: %s", html)
	}
	for mode, want := range map[Mode]string{ModeYear: "Elige un año", ModeMonth: "Elige un mes", ModeDateTime: "Elige fecha y hora", ModeTime: "Elige una hora"} {
		if html := render(State{Name: "x", Mode: mode, Lang: "es"}); !strings.Contains(html, ">"+want+"<") {
			t.Errorf("%s placeholder: %s", mode, html)
		}
	}
	if html := render(State{Name: "x", Lang: "es", Placeholder: "Nacimiento"}); !strings.Contains(html, ">Nacimiento<") {
		t.Errorf("own placeholder: %s", html)
	}
}

func TestRangeHasTwoFieldsAndShowsBothEnds(t *testing.T) {
	html := render(State{Name: "from", NameTo: "to", Range: true, Value: "2026-11-01", ValueTo: "2026-11-15", Lang: "es"})
	for _, want := range []string{`name="from"`, `value="2026-11-01"`, `name="to"`, `value="2026-11-15"`, ">01/11/2026 – 15/11/2026<"} {
		if !strings.Contains(html, want) {
			t.Errorf("lacks %q in %s", want, html)
		}
	}
	if html := render(State{Name: "from", NameTo: "to", Range: true, Lang: "es"}); !strings.Contains(html, ">Elige un rango<") {
		t.Errorf("range placeholder: %s", html)
	}
}

func TestWhatDoesNotParseIsNothing(t *testing.T) {
	html := render(State{Name: "x", Value: "mañana", Lang: "es"})
	if !strings.Contains(html, ">Elige una fecha<") || strings.Contains(html, "mañana") {
		t.Errorf("bad value: %s", html)
	}
}

func TestDisabledInvalid(t *testing.T) {
	html := render(State{Name: "x", Value: "2026-11-01", Disabled: true, Invalid: true, Lang: "es"})
	for _, want := range []string{"is-invalid", `disabled=""`} {
		if !strings.Contains(html, want) {
			t.Errorf("lacks %q in %s", want, html)
		}
	}
}

func TestTwoPickersInOnePage(t *testing.T) {
	a := island.Mount(Picker, State{Name: "a", Lang: "es"})
	b := island.Mount(Picker, State{Name: "b", Mode: ModeYear, Lang: "es"})
	html := ssr.Render(func() hx.INode { return hx.Div().Body(a.Node(), b.Node()) })
	if strings.Count(html, `data-island="datepicker"`) != 2 || !strings.Contains(html, "a-panel") || !strings.Contains(html, "b-panel") {
		t.Errorf("two islands of the same component: %s", html)
	}
}

func TestStampIsWhatTodayWants(t *testing.T) {
	now, _ := Parse(ModeDateTime, "2026-11-01T14:30:45")
	if got := Stamp(now); got != "2026-11-01T14:30:45" {
		t.Errorf("Stamp = %q", got)
	}
}

func TestYearPage(t *testing.T) {
	for y, want := range map[int]int{2026: 2016, 2016: 2016, 2027: 2016, 2028: 2028, 12: 12, 11: 0} {
		if got := yearPage(y); got != want {
			t.Errorf("yearPage(%d) = %d, want %d", y, got, want)
		}
	}
}

func TestSameDayOfMonthClampsToTheEnd(t *testing.T) {
	jan31, _ := Parse(ModeDate, "2026-01-31")
	if got := Format(ModeDate, false, sameDayOfMonth(jan31, 1)); got != "2026-02-28" {
		t.Errorf("Jan 31 + 1 month = %s", got)
	}
	if got := Format(ModeDate, false, sameDayOfMonth(jan31, -1)); got != "2025-12-31" {
		t.Errorf("Jan 31 - 1 month = %s", got)
	}
}

func TestRangeInts(t *testing.T) {
	got := rangeInts(0, 59, 15, 7)
	want := []int{0, 7, 15, 30, 45}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
