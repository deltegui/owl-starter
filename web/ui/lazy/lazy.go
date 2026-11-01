// Package lazy is a component that loads what it shows when it is needed: until
// the data arrives it shows a spinner, and then the view of the data. It compiles
// on both sides (only Hix and hxbs): it goes inside an island.
//
//	report := lazy.New(lazy.Options[Report]{
//		When:  reportTabOpen, // loads the first time this is true
//		Fetch: fetchReport,
//		View:  reportView,
//	})
//	bs.TabPanel(active, "report", report.Node)
package lazy

import (
	"context"

	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
)

// Options describe what to load and how to show it.
type Options[T any] struct {
	// When is what makes it load, the first time it is true (an open tab, an
	// opened accordion). Nil loads when the component is in the page.
	When hx.Gettable[bool]
	// Fetch asks for the data. Required. Its context is canceled with the island.
	Fetch func(ctx context.Context) (T, error)
	// View shows the data once it is here. Required.
	View func(T) hx.INode
	// Spinner is shown while there is no data. By default, a Bootstrap spinner.
	Spinner func() hx.INode
	// Failed is shown when Fetch fails; retry asks again. By default, an alert
	// with the error and a button to try again.
	Failed func(err error, retry func()) hx.INode
}

// Lazy is the state of a lazy component. Make it once, when the island is built
// (New), and place Node wherever it goes, as many times as the page builds it: a
// tab panel is built again each time its tab is chosen, and the data stays.
type Lazy[T any] struct {
	opts Options[T]
	data *hx.ResourceT[*T]
}

// New makes the component. It panics without Fetch or View.
func New[T any](o Options[T]) *Lazy[T] {
	if o.Fetch == nil || o.View == nil {
		panic("lazy: Options needs Fetch and View")
	}
	if o.Spinner == nil {
		o.Spinner = defaultSpinner
	}
	if o.Failed == nil {
		o.Failed = defaultFailed
	}
	l := &Lazy[T]{opts: o}
	// ResourceWith does not load until Refetch is called. A pointer says whether
	// there are data: nil is the same on the server and when hydrating.
	l.data = hx.ResourceWith[*T](nil, func(ctx context.Context) (*T, error) {
		v, err := o.Fetch(ctx)
		if err != nil {
			return nil, err
		}
		return &v, nil
	})
	if o.When == nil {
		hx.OnMount(l.Reload)
		return l
	}
	started := false
	hx.EffectFunc(func() {
		if hx.IsServer() || started || !o.When.Get() {
			return // the server only shows the spinner; it cannot ask
		}
		started = true
		hx.Untrack(l.Reload)
	})
	return l
}

// Load is New(o).Node(), for a component that is built once.
func Load[T any](o Options[T]) hx.INode { return New(o).Node() }

// Reload asks for the data again; what there is stays until the new data come.
func (l *Lazy[T]) Reload() { l.data.Refetch() }

// Node is the spinner, the error or the view, whatever the state is now.
func (l *Lazy[T]) Node() hx.INode {
	return hx.Dynamic(func() hx.INode {
		if err := l.data.Error().Get(); err != nil {
			return untracked(func() hx.INode { return l.opts.Failed(err, l.Reload) })
		}
		if v := l.data.Value().Get(); v != nil {
			return untracked(func() hx.INode { return l.opts.View(*v) })
		}
		return untracked(l.opts.Spinner)
	})
}

// untracked builds without following the signals the builder reads: only the
// state of the data makes Dynamic build again.
func untracked(build func() hx.INode) hx.INode {
	var n hx.INode
	hx.Untrack(func() { n = build() })
	return n
}

func defaultSpinner() hx.INode {
	return hx.Div().Class("text-center", "py-5").Body(bs.Spinner())
}

func defaultFailed(err error, retry func()) hx.INode {
	return bs.Alert(hx.Div().Class("d-flex", "justify-content-between", "align-items-center").Body(
		hx.Span().Text(err.Error()),
		bs.Button("Retry", bs.Outline(bs.Danger), bs.Sm).OnClick(func(hx.EventContext) { retry() }),
	), bs.Danger)
}
