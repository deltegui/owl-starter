// Package selectui is a select with a search box and a custom dropdown, one
// island for every case: simple or multiple choice, with the options in the page
// (filtered in the browser) or asked to the server while the user types. It
// compiles on both sides: the server writes the closed control, with its hidden
// inputs, and the browser hydrates it. What only the browser can do (keys,
// measures, focus) is in web/ui/popover, shared with the other components that
// open a panel.
//
// The directory is called select, which is a keyword of Go, so the package is
// selectui.
//
//	island.Mount(selectui.Select, selectui.State{
//		Name:     "country",
//		Endpoint: "/api/countries", // GET ?q=text, answers [{"value":"es","label":"Spain"}]
//		Selected: []selectui.Option{{Value: "es", Label: "Spain"}},
//		Lang:     ctx.GetCurrentLanguage(),
//	})
//
// A form that contains it sends what the user chose as hidden inputs called
// Name (one for each chosen value), and the "change" event that a native select
// has reaches the form.
package selectui

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/deltegui/hx"
	bs "github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/island/client"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/locale"
	"github.com/deltegui/owl-starter/web/ui/popover"
)

// Option is a choice: Value is what the form sends and Label what the user reads.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// State is what the server gives the island. Only exported fields with a json
// tag travel, and the user can read all of it.
type State struct {
	// Name is the name of the field in the form. It is also the base of the ids
	// of the control (ID overrides it), so it must be unique in the page.
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`

	Multiple    bool   `json:"multiple,omitempty"`
	Clearable   bool   `json:"clearable,omitempty"` // a button that empties the choice
	Disabled    bool   `json:"disabled,omitempty"`
	Invalid     bool   `json:"invalid,omitempty"` // paints it as a field with an error
	Placeholder string `json:"placeholder,omitempty"`

	// Selected are the options chosen when the page is written.
	Selected []Option `json:"selected,omitempty"`
	// Options are the ones the user can choose from. With an Endpoint they are
	// what the server answers to the empty search (the first ones, say); with no
	// Options either, the first time the dropdown opens asks for them.
	Options []Option `json:"options,omitempty"`
	// Endpoint, if there is one, is asked with GET Endpoint?q=text and answers a
	// JSON array of options. The server decides how many (the query limits it).
	Endpoint string `json:"endpoint,omitempty"`
	// MinChars is how many letters the user must write before asking. Only with an Endpoint.
	MinChars int `json:"minChars,omitempty"`

	Lang string `json:"l"`
}

func (s State) id() string {
	if s.ID != "" {
		return s.ID
	}
	return s.Name
}

var texts = localizer.Dict{
	localizer.LangEnglish: {
		"Placeholder": "Select...",
		"Search":      "Search...",
		"NoResults":   "No results",
		"Loading":     "Loading...",
		"MinChars":    "Type at least %d characters",
		"Clear":       "Clear",
		"Remove":      "Remove",
		"Failed":      "The options could not be loaded",
		"Retry":       "Retry",
	},
	localizer.LangSpanish: {
		"Placeholder": "Seleccionar...",
		"Search":      "Buscar...",
		"NoResults":   "Sin resultados",
		"Loading":     "Cargando...",
		"MinChars":    "Escribe al menos %d caracteres",
		"Clear":       "Quitar",
		"Remove":      "Quitar",
		"Failed":      "No se pudieron cargar las opciones",
		"Retry":       "Reintentar",
	},
}

// Select is the island. Mount it once for each select of the page.
var Select = island.Define("select", build)

const debounce = 250 * time.Millisecond

func build(s State) hx.INode {
	b := &box{
		s:        s,
		t:        localizer.Pick(texts, s.Lang, locale.Languages[0]),
		id:       s.id(),
		selected: hx.Signal(slices.Clone(s.Selected)),
		query:    hx.Signal(""),
		active:   hx.Signal(0),
	}
	b.listID = b.id + "-list"
	b.pop = popover.New(b.id)
	if s.Endpoint != "" {
		b.data = hx.ResourceForWith(slices.Clone(s.Options), b.query, func(ctx context.Context, q string) ([]Option, error) {
			return fetch(ctx, s, q)
		})
	}
	b.root = hx.Div().Class("position-relative")
	b.trigger = b.triggerNode()

	if b.data != nil && len(s.Options) == 0 && s.MinChars == 0 {
		asked := false
		hx.EffectFunc(func() { // the first time it opens, asks for the first options
			if hx.IsServer() || asked || !b.pop.IsOpen().Get() {
				return
			}
			asked = true
			hx.Untrack(b.data.Refetch)
		})
	}
	return b.root.Body(b.hiddenInputs(), b.trigger, b.panel())
}

// fetch asks the server for the options that match q. It waits a moment first,
// and the request that was waiting is canceled when the user types again.
func fetch(ctx context.Context, s State, q string) ([]Option, error) {
	if len([]rune(strings.TrimSpace(q))) < s.MinChars {
		return nil, nil
	}
	if q != "" {
		select {
		case <-time.After(debounce):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	sep := "?"
	if strings.Contains(s.Endpoint, "?") {
		sep = "&"
	}
	return client.Send[[]Option](ctx, "GET", s.Endpoint+sep+"q="+client.Escape(q), nil)
}

// box is the state of one select.
type box struct {
	s      State
	t      localizer.Localizer
	id     string
	listID string
	pop    *popover.Popover

	selected *hx.SignalT[[]Option]
	query    *hx.SignalT[string]
	active   *hx.SignalT[int]        // index of the highlighted option
	data     *hx.ResourceT[[]Option] // nil when the options are in the page

	root    hx.INode
	trigger hx.INode
}

// results are the options to show now. It reads signals: call it where they are tracked.
func (b *box) results() []Option {
	if b.data != nil {
		return b.data.Value().Get()
	}
	return Filter(b.s.Options, b.query.Get())
}

func (b *box) openPanel() {
	if b.s.Disabled || b.pop.Opened() {
		return
	}
	hx.Batch(func() {
		if hx.UntrackGet[string](b.query) != "" {
			b.query.Set("")
		}
		b.active.Set(0)
		b.pop.Open()
	})
}

func (b *box) choose(o Option) {
	if b.s.Multiple {
		b.selected.Update(func(cur []Option) []Option { return toggled(cur, o) })
	} else {
		b.selected.Set([]Option{o})
		b.pop.Close()
		popover.Focus(b.trigger)
	}
	popover.EmitChange(b.root)
}

func (b *box) clear() {
	b.selected.Set(nil)
	popover.EmitChange(b.root)
}

func (b *box) remove(o Option) {
	b.selected.Update(func(cur []Option) []Option { return without(cur, o) })
	popover.EmitChange(b.root)
}

// hiddenInputs are what the form sends: one input for each chosen value. A
// simple select with nothing chosen sends an empty value, like a native one.
func (b *box) hiddenInputs() hx.INode {
	return hx.Dynamic(func() hx.INode {
		chosen := b.selected.Get()
		inputs := make([]hx.INode, 0, max(len(chosen), 1))
		add := func(value string) {
			in := hx.Input().Type("hidden").Attribute("name", b.s.Name).Value(value)
			if b.s.Disabled {
				in.Attribute("disabled", "")
			}
			inputs = append(inputs, in)
		}
		for _, o := range chosen {
			add(o.Value)
		}
		if len(chosen) == 0 && !b.s.Multiple {
			add("")
		}
		return hx.Noop().BodyList(inputs)
	})
}

// triggerNode is what the user sees closed: the chosen options or the placeholder.
func (b *box) triggerNode() hx.INode {
	t := hx.Div().Class("form-select", "d-flex", "flex-wrap", "align-items-center", "gap-1").
		Style("cursor", "pointer").Style("height", "auto").Style("min-height", "calc(1.5em + .75rem + 2px)").
		Attribute("role", "combobox").
		Attribute("aria-haspopup", "listbox").
		Attribute("aria-controls", b.listID).
		Attribute("aria-label", b.s.placeholderOr(b.t.Get("Placeholder")))
	if b.s.Disabled {
		t.Attribute("aria-disabled", "true").Attribute("tabindex", "-1").Class("disabled")
	} else {
		t.Attribute("tabindex", "0")
	}
	t.ClassIf("is-invalid", b.s.Invalid)
	b.pop.Expanded(t)
	t.OnClick(func(hx.EventContext) {
		if b.pop.Opened() {
			b.pop.Close()
		} else {
			b.openPanel()
		}
	})
	t.OnKeyDown(func(ctx hx.EventContext) {
		switch popover.Key(ctx) {
		case "Enter", " ", "ArrowDown", "ArrowUp":
			popover.PreventDefault(ctx)
			b.openPanel()
		case "Escape":
			b.pop.Close()
		}
	})
	return t.Body(hx.Dynamic(func() hx.INode {
		chosen := b.selected.Get()
		var out hx.INode
		hx.Untrack(func() { out = b.chosenView(chosen) })
		return out
	}))
}

func (s State) placeholderOr(fallback string) string {
	if s.Placeholder != "" {
		return s.Placeholder
	}
	return fallback
}

func (b *box) chosenView(chosen []Option) hx.INode {
	if len(chosen) == 0 {
		return hx.Span().Class("text-body-secondary").Text(b.s.placeholderOr(b.t.Get("Placeholder")))
	}
	var parts []hx.INode
	if b.s.Multiple {
		for _, o := range chosen {
			chip := hx.Span().Class("badge", "text-bg-secondary", "d-inline-flex", "align-items-center", "gap-1")
			children := []hx.INode{hx.Span().Text(o.Label)}
			if !b.s.Disabled {
				children = append(children, hx.Button().Type("button").Class("btn-close", "btn-close-white").
					Style("font-size", ".6em").Attribute("aria-label", b.t.Get("Remove")+" "+o.Label).
					OnClick(func(ctx hx.EventContext) {
						popover.StopClick(ctx)
						b.remove(o)
					}))
			}
			chip.BodyList(children)
			parts = append(parts, chip)
		}
	} else {
		parts = append(parts, hx.Span().Text(chosen[0].Label))
	}
	if b.s.Clearable && !b.s.Disabled {
		parts = append(parts, hx.Button().Type("button").Class("btn-close", "ms-auto").
			Attribute("aria-label", b.t.Get("Clear")).
			OnClick(func(ctx hx.EventContext) {
				popover.StopClick(ctx)
				b.clear()
			}))
	}
	return hx.Noop().BodyList(parts)
}

// panel is the dropdown: nothing while closed and, when open, the box with the
// search and the list.
func (b *box) panel() hx.INode {
	var search hx.INode
	return b.pop.Panel(b.trigger, popover.Options{
		MatchWidth: true,
		// When the options change the panel changes height: place it again.
		Track: func() {
			b.results()
			if b.data != nil {
				b.data.Loading().Get()
				b.data.Error().Get()
			}
		},
		OnShown: func(hx.INode) { popover.Focus(search) },
	}, func() []hx.INode {
		search = hx.Input().Type("text").Class("form-control", "form-control-sm", "mb-2").
			Placeholder(b.t.Get("Search")).
			Attribute("role", "searchbox").
			Attribute("autocomplete", "off").
			Attribute("aria-controls", b.listID).
			BindValue(b.query).BindOnInput(b.query)
		search.OnInput(func(hx.EventContext) { b.active.Set(0) })
		search.Effect(func() { search.Attribute("aria-activedescendant", b.optionID(b.active.Get())) })
		search.OnKeyDown(b.searchKey)

		list := hx.Ul().Id(b.listID).Class("list-unstyled", "mb-0").
			Style("overflow-y", "auto").Style("min-height", "0").
			Attribute("role", "listbox")
		if b.s.Multiple {
			list.Attribute("aria-multiselectable", "true")
		}
		list.Body(b.listView())
		return []hx.INode{search, list}
	})
}

// current is the list of options, read without following the signals.
func (b *box) current() []Option {
	var items []Option
	hx.Untrack(func() { items = b.results() })
	return items
}

func (b *box) searchKey(ctx hx.EventContext) {
	switch popover.Key(ctx) {
	case "ArrowDown":
		popover.PreventDefault(ctx)
		b.move(+1)
	case "ArrowUp":
		popover.PreventDefault(ctx)
		b.move(-1)
	case "Enter":
		popover.PreventDefault(ctx) // choosing must not send the form
		items, at := b.current(), hx.UntrackGet[int](b.active)
		if at >= 0 && at < len(items) {
			b.choose(items[at])
		}
	case "Escape":
		b.pop.Close()
		popover.Focus(b.trigger)
	case "Tab":
		b.pop.Close()
	}
}

// move highlights the next or previous option, and stops at the ends.
func (b *box) move(by int) {
	n := len(b.current())
	if n == 0 {
		return
	}
	next := min(max(hx.UntrackGet[int](b.active)+by, 0), n-1)
	b.active.Set(next)
	popover.ScrollToID(b.optionID(next))
}

func (b *box) optionID(i int) string { return b.id + "-opt-" + strconv.Itoa(i) }

// listView is the options, or what takes their place: a call to write more, the
// spinner, the error, or "no results".
func (b *box) listView() hx.INode {
	return hx.Dynamic(func() hx.INode {
		items := b.results()
		query := b.query.Get()
		loading, failed := false, error(nil)
		if b.data != nil {
			loading, failed = b.data.Loading().Get(), b.data.Error().Get()
		}
		var out hx.INode
		hx.Untrack(func() { out = b.listBody(items, query, loading, failed) })
		return out
	})
}

func (b *box) listBody(items []Option, query string, loading bool, failed error) hx.INode {
	note := func(child hx.INode) hx.INode {
		return hx.Li().Attribute("role", "presentation").Class("px-2", "py-1", "text-body-secondary").Body(child)
	}
	switch {
	case failed != nil:
		return note(hx.Span().Body(
			hx.Span().Text(b.t.Get("Failed")+" "),
			hx.Button().Type("button").Class("btn", "btn-link", "btn-sm", "p-0").Text(b.t.Get("Retry")).
				OnClick(func(hx.EventContext) { b.data.Refetch() })))
	case b.data != nil && len([]rune(strings.TrimSpace(query))) < b.s.MinChars:
		return note(hx.Span().Text(b.t.GetFormatted("MinChars", b.s.MinChars)))
	case loading && len(items) == 0:
		return note(hx.Span().Body(bs.Spinner(bs.Sm), hx.Span().Class("ms-2").Text(b.t.Get("Loading"))))
	case len(items) == 0:
		return note(hx.Span().Text(b.t.Get("NoResults")))
	}
	rows := make([]hx.INode, 0, len(items)+1)
	if loading {
		rows = append(rows, note(bs.Spinner(bs.Sm)))
	}
	for i, o := range items {
		rows = append(rows, b.optionNode(i, o))
	}
	return hx.Noop().BodyList(rows)
}

func (b *box) optionNode(i int, o Option) hx.INode {
	mark := hx.Span().Class("ms-2").Text("\u2713")
	row := hx.Li().Id(b.optionID(i)).Class("dropdown-item", "d-flex", "justify-content-between", "align-items-center").
		Style("cursor", "pointer").
		Attribute("role", "option").
		Body(hx.Span().Text(o.Label), mark)
	row.Effect(func() {
		on := b.isSelected(o.Value)
		row.Attribute("aria-selected", strconv.FormatBool(on))
		mark.ClassIf("invisible", !on)
		row.ClassIf("active", b.active.Get() == i)
	})
	row.OnClick(func(hx.EventContext) { b.choose(o) })
	row.On(hx.EventMouseEnter, func(hx.EventContext) { b.active.Set(i) })
	return row
}

func (b *box) isSelected(value string) bool {
	return slices.ContainsFunc(b.selected.Get(), func(o Option) bool { return o.Value == value })
}

// toggled adds o to cur, or takes it out if it is there.
func toggled(cur []Option, o Option) []Option {
	if slices.ContainsFunc(cur, func(c Option) bool { return c.Value == o.Value }) {
		return without(cur, o)
	}
	return append(slices.Clone(cur), o)
}

func without(cur []Option, o Option) []Option {
	return slices.DeleteFunc(slices.Clone(cur), func(c Option) bool { return c.Value == o.Value })
}

// Filter keeps the options whose label has the text, without telling capital
// letters or accents apart: "jose" finds "José".
func Filter(options []Option, text string) []Option {
	text = fold(strings.TrimSpace(text))
	if text == "" {
		return options
	}
	var out []Option
	for _, o := range options {
		if strings.Contains(fold(o.Label), text) {
			out = append(out, o)
		}
	}
	return out
}

var accents = strings.NewReplacer(
	"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
	"é", "e", "è", "e", "ë", "e", "ê", "e",
	"í", "i", "ì", "i", "ï", "i", "î", "i",
	"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
	"ú", "u", "ù", "u", "ü", "u", "û", "u",
	"ñ", "n", "ç", "c")

func fold(s string) string { return accents.Replace(strings.ToLower(s)) }
