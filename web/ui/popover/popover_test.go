package popover_test

import (
	"strings"
	"testing"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/ssr"

	"github.com/deltegui/owl-starter/web/ui/popover"
)

func TestServerWritesNothingOfThePanel(t *testing.T) {
	html := ssr.Render(func() hx.INode {
		pop := popover.New("who")
		trigger := hx.Button().Text("open")
		pop.Expanded(trigger)
		built := 0
		panel := pop.Panel(trigger, popover.Options{}, func() []hx.INode {
			built++
			return []hx.INode{hx.P().Text("inside")}
		})
		if built != 0 {
			t.Errorf("the content was built while closed")
		}
		return hx.Div().Body(trigger, panel)
	})
	if !strings.Contains(html, `aria-expanded="false"`) {
		t.Errorf("the control says it is closed: %s", html)
	}
	if strings.Contains(html, "inside") || strings.Contains(html, "who-panel") {
		t.Errorf("a closed popover has no panel: %s", html)
	}
}

func TestPanelID(t *testing.T) {
	ssr.Render(func() hx.INode {
		if got := popover.New("birthday").PanelID(); got != "birthday-panel" {
			t.Errorf("PanelID = %q", got)
		}
		return hx.Noop()
	})
}
