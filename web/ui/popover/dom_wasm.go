//go:build js && wasm

package popover

import (
	"strconv"
	"syscall/js"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/wasm"
)

// Key is the key of a keyboard event ("Enter", "ArrowDown"...).
func Key(ctx hx.EventContext) string { return wasm.Key(ctx) }

// PreventDefault stops what the browser does by itself with the event.
func PreventDefault(ctx hx.EventContext) { wasm.PreventDefault(ctx) }

// StopClick keeps the event from reaching the elements around the one that has the listener.
func StopClick(ctx hx.EventContext) { wasm.StopPropagation(ctx) }

// Focus gives the keyboard focus to the element of the node, once it is in the page.
func Focus(n hx.INode) {
	if el := wasm.Underlying(n); !el.IsUndefined() {
		el.Call("focus")
	}
}

// TargetValue is the value of the control that fired the event (a select, an input).
func TargetValue(ctx hx.EventContext) string {
	if event := wasm.EventOf(ctx); !event.IsUndefined() {
		if v := event.Get("target").Get("value"); v.Type() == js.TypeString {
			return v.String()
		}
	}
	return ""
}

// FocusID gives the focus to the element with that id once the page has drawn
// what the last change asked for.
func FocusID(id string) {
	NextFrame(func() {
		NextFrame(func() {
			if el := js.Global().Get("document").Call("getElementById", id); !el.IsNull() {
				el.Call("focus")
			}
		})
	})
}

// NextFrame runs fn on the next animation frame, after the renderer of Hix has
// written what was waiting.
func NextFrame(fn func()) {
	var cb js.Func
	cb = js.FuncOf(func(js.Value, []js.Value) any {
		cb.Release()
		fn()
		return nil
	})
	js.Global().Call("requestAnimationFrame", cb)
}

// EmitChange tells the page that the value changed: a "change" event that
// bubbles from the element, like the one of a native control. It waits two
// frames so the hidden inputs already have their new values when a form listens.
func EmitChange(n hx.INode) {
	NextFrame(func() {
		NextFrame(func() {
			el := wasm.Underlying(n)
			if el.IsUndefined() {
				return
			}
			event := js.Global().Get("Event").New("change", map[string]any{"bubbles": true})
			el.Call("dispatchEvent", event)
		})
	})
}

// ScrollToID scrolls the element with that id into view, the least that is needed.
func ScrollToID(id string) {
	NextFrame(func() {
		if el := js.Global().Get("document").Call("getElementById", id); !el.IsNull() {
			el.Call("scrollIntoView", map[string]any{"block": "nearest"})
		}
	})
}

const (
	gap    = 2  // px between the control and the panel
	margin = 8  // px that the panel leaves to the edge of the window
	minFit = 70 // below this much room under the control, the panel opens upwards
)

// Place puts the panel (position: fixed) under the control. With match it is as
// wide as the control; without, as wide as its content, moved to the left if it
// would leave the window. If it does not fit below and there is more room above,
// it opens upwards; its height is limited to the room it has. It is what
// LittleCash does with its shared menu, with viewport coordinates (a fixed
// element needs no scroll offset).
func Place(trigger, panel hx.INode, match bool) {
	t, p := wasm.Underlying(trigger), wasm.Underlying(panel)
	if t.IsUndefined() || p.IsUndefined() {
		return
	}
	rect := t.Call("getBoundingClientRect")
	top, bottom := rect.Get("top").Float(), rect.Get("bottom").Float()
	left, width := rect.Get("left").Float(), rect.Get("width").Float()
	viewport := js.Global().Get("innerHeight").Float()

	style := p.Get("style")
	if match {
		style.Set("width", px(width))
	} else {
		style.Set("width", "max-content")
		over := left + p.Get("offsetWidth").Float() + margin - js.Global().Get("innerWidth").Float()
		if over > 0 {
			left = max(left-over, margin)
		}
	}
	style.Set("left", px(left))
	style.Set("maxHeight", "none") // measure it as tall as it wants to be
	height := p.Get("offsetHeight").Float()

	below := viewport - bottom - gap - margin
	above := top - gap - margin
	if height > below && below < minFit+margin && above > below {
		h := min(height, above)
		style.Set("maxHeight", px(h))
		style.Set("top", px(top-gap-h))
		return
	}
	style.Set("maxHeight", px(min(height, below)))
	style.Set("top", px(bottom+gap))
}

func px(v float64) string { return strconv.FormatFloat(v, 'f', 0, 64) + "px" }

// WatchViewport calls onChange when the window is resized or something outside
// the panel scrolls, so the panel does not stay where it was. It returns what
// stops listening.
func WatchViewport(panelID string, onChange func()) func() {
	win := js.Global()
	resize := js.FuncOf(func(js.Value, []js.Value) any { onChange(); return nil })
	scroll := js.FuncOf(func(_ js.Value, args []js.Value) any {
		panel := win.Get("document").Call("getElementById", panelID)
		if !panel.IsNull() && len(args) > 0 && panel.Call("contains", args[0].Get("target")).Bool() {
			return nil // the panel itself scrolls: not a reason to close
		}
		onChange()
		return nil
	})
	win.Call("addEventListener", "resize", resize)
	win.Call("addEventListener", "scroll", scroll, true)
	return func() {
		win.Call("removeEventListener", "resize", resize)
		win.Call("removeEventListener", "scroll", scroll, true)
		resize.Release()
		scroll.Release()
	}
}
