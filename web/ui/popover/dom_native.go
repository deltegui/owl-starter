//go:build !(js && wasm)

package popover

import "github.com/deltegui/hx"

// What only exists in the browser does nothing on the server: it never builds
// the panel (nothing is open) and no event runs.

func Key(hx.EventContext) string                { return "" }
func PreventDefault(hx.EventContext)            {}
func StopClick(hx.EventContext)                 {}
func TargetValue(hx.EventContext) string        { return "" }
func FocusID(string)                            {}
func Focus(hx.INode)                            {}
func NextFrame(func())                          {}
func Place(trigger, panel hx.INode, match bool) {}
func EmitChange(hx.INode)                       {}
func ScrollToID(string)                         {}
func WatchViewport(string, func()) func()       { return func() {} }
