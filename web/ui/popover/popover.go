// Package popover is the panel that opens under a control (a select's list, a
// calendar) and what goes with it: only one open at a time in the page, a layer
// that closes it when clicked outside, a position that follows the control and
// opens upwards when there is no room below, and closing when the window is
// resized or the page scrolls. It compiles on both sides; what only the browser
// can do is in dom_wasm.go.
//
//	pop := popover.New("birthday")           // while the island is built
//	trigger.OnClick(func(hx.EventContext) { pop.Toggle() })
//	pop.Expanded(trigger)                     // keeps aria-expanded
//	panel := pop.Panel(trigger, popover.Options{}, func() []hx.INode { return content })
//
// Key, Focus, EmitChange and the other helpers at the end of the package are the
// small pieces of the browser that the components using it need as well.
package popover

import (
	"strconv"
	"sync"

	"github.com/deltegui/hx"
)

// Which popover is open, for the whole page. All the islands are in the same
// program, so they share it. It is made on first use because a signal needs the
// engine of Hix, which is not ready when the packages start.
var (
	openOnce   sync.Once
	openSignal *hx.SignalT[string]
)

func current() *hx.SignalT[string] {
	openOnce.Do(func() { openSignal = hx.Signal("") })
	return openSignal
}

// Popover is the state of one panel. Make it with New.
type Popover struct {
	id   string
	open *hx.ComputedT[bool]
}

// New makes the popover called id, which must be unique in the page (it is the
// base of the id of the panel). Call it while the island is built: it listens to
// the window until the island is disposed.
func New(id string) *Popover {
	p := &Popover{id: id}
	p.open = hx.Computed(func() bool { return current().Get() == id })
	var stop func()
	hx.OnMount(func() { stop = WatchViewport(p.PanelID(), p.Close) })
	hx.OnCleanup(func() {
		if stop != nil {
			stop()
		}
	})
	return p
}

// PanelID is the id of the element of the panel.
func (p *Popover) PanelID() string { return p.id + "-panel" }

// IsOpen is whether it is open, to read inside effects.
func (p *Popover) IsOpen() hx.Gettable[bool] { return p.open }

// Opened is whether it is open, to ask inside an event handler.
func (p *Popover) Opened() bool { return hx.UntrackGet[bool](p.open) }

// Open opens it, and the one that was open closes.
func (p *Popover) Open() {
	if !p.Opened() {
		current().Set(p.id)
	}
}

// Close closes it if it is open.
func (p *Popover) Close() {
	if p.Opened() {
		current().Set("")
	}
}

// Toggle opens it or closes it.
func (p *Popover) Toggle() {
	if p.Opened() {
		p.Close()
	} else {
		p.Open()
	}
}

// Expanded keeps the aria-expanded of the control that opens the panel.
func (p *Popover) Expanded(trigger hx.INode) hx.INode {
	trigger.Effect(func() { trigger.Attribute("aria-expanded", strconv.FormatBool(p.open.Get())) })
	return trigger
}

// Options of Panel.
type Options struct {
	// MatchWidth makes the panel as wide as the control (a select). Without it the
	// panel is as wide as its content (a calendar).
	MatchWidth bool
	// Track reads the signals whose change makes the panel change height, so it is
	// placed again (the options that arrive, for example). Optional.
	Track func()
	// Dialog, if it is not empty, makes the box a dialog with that name for
	// screen readers (a calendar). A list of options does not need it.
	Dialog string
	// OnShown runs once the panel is in the page and placed (focus a field).
	// Optional.
	OnShown func(box hx.INode)
}

// Panel is what goes next to the control: nothing while the popover is closed
// and, when it is open, the layer that closes it and the box with what build
// returns.
func (p *Popover) Panel(trigger hx.INode, o Options, build func() []hx.INode) hx.INode {
	return hx.Dynamic(func() hx.INode {
		if !p.open.Get() {
			return nil
		}
		var out hx.INode
		hx.Untrack(func() { out = p.view(trigger, o, build) })
		return out
	})
}

// Layers: above Bootstrap's modals and below its tooltips.
func (p *Popover) view(trigger hx.INode, o Options, build func() []hx.INode) hx.INode {
	layer := hx.Div().Style("position", "fixed").Style("inset", "0").Style("z-index", "1070").
		OnClick(func(hx.EventContext) { p.Close() })

	box := hx.Div().Id(p.PanelID()).
		Class("d-flex", "flex-column", "bg-body", "border", "rounded", "shadow", "p-2").
		Style("position", "fixed").Style("z-index", "1071").Style("top", "-9999px") // out of sight until placed
	if o.Dialog != "" {
		box.Attribute("role", "dialog").Attribute("aria-label", o.Dialog)
	}
	box.BodyList(build())
	box.OnRendered(func() {
		Place(trigger, box, o.MatchWidth)
		if o.OnShown != nil {
			o.OnShown(box)
		}
	})
	box.Effect(func() {
		if o.Track != nil {
			o.Track()
		}
		NextFrame(func() { Place(trigger, box, o.MatchWidth) })
	})
	return hx.Noop().Body(layer, box)
}
