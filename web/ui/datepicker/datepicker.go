// Package datepicker is a control to choose a year, a month, a date, a date with
// a time, or just a time, and a range of any of them: one island, with the mode
// in its state. It compiles on both sides: the server writes the closed control
// with its hidden inputs, and the browser hydrates it and builds the panel when
// it opens. The date calculations are Go's own, so both sides count the same.
//
//	island.Mount(datepicker.Picker, datepicker.State{
//		Name:  "from",
//		Mode:  datepicker.ModeDateTime,
//		Value: "2026-11-01T14:30",
//		Today: datepicker.Stamp(now),
//		Lang:  ctx.GetCurrentLanguage(),
//	})
//
// Values are wall-clock text with no time zone (see Parse and Format); what zone
// they are in, and the conversion to the UTC that the application stores, are
// the application's to decide at the edges. The control is only used through the
// panel: nothing is typed.
//
// A form that contains it sends the value as hidden inputs called Name (and
// NameTo for the end of a range), and the "change" event of a native control
// reaches the form.
package datepicker

import (
	"slices"
	"strconv"
	"time"

	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/localizer"

	"github.com/deltegui/owl-starter/web/locale"
	"github.com/deltegui/owl-starter/web/ui/popover"
)

// State is what the server gives the island. Only exported fields with a json
// tag travel, and the user can read all of it.
type State struct {
	// Name is the name of the field in the form, and the base of the ids of the
	// control (ID overrides it), so it must be unique in the page. With Range it
	// is the field of the start and NameTo the one of the end.
	Name   string `json:"name"`
	NameTo string `json:"nameTo,omitempty"`
	ID     string `json:"id,omitempty"`

	Mode    Mode `json:"mode,omitempty"`    // ModeDate by default
	Seconds bool `json:"seconds,omitempty"` // with a time: choose seconds as well
	Range   bool `json:"range,omitempty"`   // choose two values: a start and an end

	// Value and ValueTo are what is chosen when the page is written, in the
	// format of the mode (empty is nothing). ValueTo is only for a Range.
	Value   string `json:"value,omitempty"`
	ValueTo string `json:"valueTo,omitempty"`
	// Min and Max, in the same format, are the limits (both included).
	Min string `json:"min,omitempty"`
	Max string `json:"max,omitempty"`
	// MinuteStep is the distance between the minutes that can be chosen (1 by default).
	MinuteStep int `json:"minuteStep,omitempty"`

	// Today is the instant that the panel marks as today, and that the Today and
	// Now buttons choose, in the form that Stamp writes. The page must give it:
	// the control cannot read the clock when it is built, or the server and the
	// browser would not write the same thing. Empty hides the buttons.
	Today string `json:"today,omitempty"`

	Placeholder string `json:"placeholder,omitempty"`
	Clearable   bool   `json:"clearable,omitempty"` // a button in the panel that empties the choice
	Disabled    bool   `json:"disabled,omitempty"`
	Invalid     bool   `json:"invalid,omitempty"`     // paints it as a field with an error
	SundayFirst bool   `json:"sundayFirst,omitempty"` // weeks start on Sunday, not on Monday

	Lang string `json:"l"`
}

// Stamp writes an instant as State.Today wants it. The date and time it has are
// the ones that count: give it the time already in the user's zone.
func Stamp(t time.Time) string { return t.Format("2006-01-02T15:04:05") }

func (s State) id() string {
	if s.ID != "" {
		return s.ID
	}
	return s.Name
}

// Picker is the island. Mount it once for each control of the page.
var Picker = island.Define("datepicker", build)

func build(s State) hx.INode {
	mode := s.Mode.orDefault()
	parse := func(text string) time.Time {
		v, _ := Parse(mode, text) // what does not parse is nothing
		return v
	}
	p := &picker{
		s:      s,
		t:      localizer.Pick(texts, s.Lang, locale.Languages[0]),
		id:     s.id(),
		mode:   mode,
		limits: bounds{min: parse(s.Min), max: parse(s.Max)},
		from:   hx.Signal(parse(s.Value)),
		to:     hx.Signal(parse(s.ValueTo)),

		view:    hx.Signal(""),
		cursor:  hx.Signal(time.Time{}),
		focus:   hx.Signal(time.Time{}),
		picking: hx.Signal(false),
		hover:   hx.Signal(time.Time{}),
	}
	if !s.Range {
		p.to.Set(time.Time{})
	}
	p.today, _ = Parse(ModeDateTime, s.Today)
	p.pop = popover.New(p.id)
	p.root = hx.Div().Class("position-relative")
	p.trigger = p.triggerNode()
	return p.root.Body(p.hiddenInputs(), p.trigger, p.panel())
}

type picker struct {
	s      State
	t      localizer.Localizer
	id     string
	mode   Mode
	limits bounds
	today  time.Time
	pop    *popover.Popover

	from, to *hx.SignalT[time.Time] // what is chosen; the zero Time is nothing
	view     *hx.SignalT[string]    // "day", "month" or "year": what the panel shows
	cursor   *hx.SignalT[time.Time] // the month (or year) the panel is in
	focus    *hx.SignalT[time.Time] // the day the keyboard is on
	picking  *hx.SignalT[bool]      // a range with its start chosen and its end to choose
	hover    *hx.SignalT[time.Time] // the day under the mouse while choosing the end

	root    hx.INode
	trigger hx.INode
}

// --- what is outside the panel ---------------------------------------------

func (p *picker) hiddenInputs() hx.INode {
	return hx.Dynamic(func() hx.INode {
		field := func(name string, v time.Time) hx.INode {
			in := hx.Input().Type("hidden").Attribute("name", name).Value(Format(p.mode, p.s.Seconds, v))
			if p.s.Disabled {
				in.Attribute("disabled", "")
			}
			return in
		}
		out := []hx.INode{field(p.s.Name, p.from.Get())}
		if p.s.Range {
			out = append(out, field(p.s.NameTo, p.to.Get()))
		}
		return hx.Noop().BodyList(out)
	})
}

// shown is how a value is written in the control, in the language of the user.
func (p *picker) shown(v time.Time) string {
	if v.IsZero() {
		return ""
	}
	date := func() string {
		switch p.mode {
		case ModeYear:
			return v.Format(p.t.Get("FmtYear"))
		case ModeMonth:
			return v.Format(p.t.Get("FmtMonth"))
		}
		return v.Format(p.t.Get("FmtDate"))
	}
	clock := func() string {
		if p.s.Seconds {
			return v.Format(p.t.Get("FmtTimeSec"))
		}
		return v.Format(p.t.Get("FmtTime"))
	}
	switch p.mode {
	case ModeTime:
		return clock()
	case ModeDateTime:
		return date() + " " + clock()
	}
	return date()
}

func (p *picker) placeholder() string {
	if p.s.Placeholder != "" {
		return p.s.Placeholder
	}
	if p.s.Range {
		return p.t.Get("PhRange")
	}
	switch p.mode {
	case ModeYear:
		return p.t.Get("PhYear")
	case ModeMonth:
		return p.t.Get("PhMonth")
	case ModeDateTime:
		return p.t.Get("PhDateTime")
	case ModeTime:
		return p.t.Get("PhTime")
	}
	return p.t.Get("PhDate")
}

func (p *picker) triggerNode() hx.INode {
	t := hx.Button().Type("button").
		Class("form-select", "text-start").
		Attribute("aria-haspopup", "dialog").
		Attribute("aria-controls", p.pop.PanelID()).
		Attribute("aria-label", p.placeholder())
	t.ClassIf("is-invalid", p.s.Invalid)
	if p.s.Disabled {
		t.Attribute("disabled", "")
	}
	p.pop.Expanded(t)
	t.OnClick(func(hx.EventContext) {
		if p.pop.Opened() {
			p.pop.Close()
		} else {
			p.open()
		}
	})
	t.OnKeyDown(func(ctx hx.EventContext) {
		if popover.Key(ctx) == "ArrowDown" {
			popover.PreventDefault(ctx)
			p.open()
		}
	})
	return t.Body(hx.Dynamic(func() hx.INode {
		text := p.shown(p.from.Get())
		if p.s.Range {
			if end := p.shown(p.to.Get()); end != "" {
				text += " – " + end
			}
		}
		if text == "" {
			return hx.Span().Class("text-body-secondary").Text(p.placeholder())
		}
		return hx.Span().Text(text)
	}))
}

// open starts the panel where the user is: on the value, or on today.
func (p *picker) open() {
	if p.s.Disabled || p.pop.Opened() {
		return
	}
	hx.Batch(func() {
		anchor := hx.UntrackGet[time.Time](p.from)
		if anchor.IsZero() {
			anchor = p.today
		}
		if anchor.IsZero() {
			anchor = time.Now().UTC() // the panel only exists in the browser, not when it is hydrated
		}
		anchor = p.limits.clamp(anchor)
		p.view.Set(p.firstView())
		p.cursor.Set(firstOfMonth(anchor))
		p.focus.Set(startOfDay(anchor))
		p.picking.Set(false)
		p.hover.Set(time.Time{})
		p.pop.Open()
	})
}

func (p *picker) firstView() string {
	switch p.mode {
	case ModeYear:
		return "year"
	case ModeMonth:
		return "month"
	}
	return "day"
}

// --- choosing --------------------------------------------------------------

// unitStart is the start of the day, the month or the year that t is in, the
// unit that the mode chooses.
func (p *picker) unitStart(t time.Time) time.Time {
	switch p.mode {
	case ModeYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case ModeMonth:
		return firstOfMonth(t)
	}
	return startOfDay(t)
}

func (p *picker) unitEnd(t time.Time) time.Time {
	start := p.unitStart(t)
	switch p.mode {
	case ModeYear:
		return start.AddDate(1, 0, 0).Add(-time.Second)
	case ModeMonth:
		return start.AddDate(0, 1, 0).Add(-time.Second)
	}
	return endOfDay(start)
}

// withClock puts the time of day of like on the day (the start of the day if like
// is nothing, and the end of it if end is true), for the modes that choose a time.
func (p *picker) withClock(day, like time.Time, end bool) time.Time {
	if p.mode != ModeDateTime {
		return day
	}
	h, m, s := like.Hour(), like.Minute(), like.Second()
	if like.IsZero() {
		h, m, s = 0, 0, 0
		if end {
			h, m = 23, 59
			if p.s.Seconds {
				s = 59
			}
		}
	}
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, s, 0, time.UTC)
}

// pick is the user choosing a day, a month or a year, depending on the mode.
func (p *picker) pick(t time.Time) {
	t = p.unitStart(t)
	if p.limits.outside(t, p.unitEnd(t)) {
		return
	}
	if !p.s.Range {
		p.from.Set(p.limits.clamp(p.withClock(t, hx.UntrackGet[time.Time](p.from), false)))
		p.finish(true)
		return
	}
	if !hx.UntrackGet[bool](p.picking) {
		hx.Batch(func() {
			p.from.Set(p.limits.clamp(p.withClock(t, hx.UntrackGet[time.Time](p.from), false)))
			p.to.Set(time.Time{})
			p.picking.Set(true)
			p.hover.Set(time.Time{})
		})
		p.finish(false)
		return
	}
	start := hx.UntrackGet[time.Time](p.from)
	first, last := start, t
	if t.Before(p.unitStart(start)) { // the end was chosen before the start: they swap
		first, last = p.withClock(t, start, false), p.unitStart(start)
	}
	hx.Batch(func() {
		p.from.Set(p.limits.clamp(first))
		end := p.limits.clamp(p.withClock(last, hx.UntrackGet[time.Time](p.to), true))
		if start := hx.UntrackGet[time.Time](p.from); end.Before(start) {
			end = start
		}
		p.to.Set(end)
		p.picking.Set(false)
		p.hover.Set(time.Time{})
	})
	p.finish(true)
}

// finish tells the page that the value changed. A choice that is complete and
// needs no time closes the panel and gives the focus back to the control.
func (p *picker) finish(complete bool) {
	if complete && !p.mode.hasTime() {
		p.pop.Close()
		popover.Focus(p.trigger)
	}
	popover.EmitChange(p.root)
}

func (p *picker) clear() {
	hx.Batch(func() {
		p.from.Set(time.Time{})
		p.to.Set(time.Time{})
		p.picking.Set(false)
	})
	p.pop.Close()
	popover.Focus(p.trigger)
	popover.EmitChange(p.root)
}

func (p *picker) chooseToday() {
	if p.today.IsZero() {
		return
	}
	now := normalize(p.mode, p.today)
	p.from.Set(p.limits.clamp(now))
	p.view.Set(p.firstView())
	p.cursor.Set(firstOfMonth(now))
	p.focus.Set(startOfDay(now))
	p.finish(true)
}

// setClock changes the time of day of the start or the end of the choice.
func (p *picker) setClock(end bool, change func(h, m, s int) (int, int, int)) {
	sig := p.from
	if end {
		sig = p.to
	}
	cur := hx.UntrackGet[time.Time](sig)
	if cur.IsZero() { // no day chosen yet: the one the panel is on
		cur = p.withClock(hx.UntrackGet[time.Time](p.focus), time.Time{}, end)
		if p.mode == ModeTime {
			cur = refDay
		}
	}
	h, m, s := change(cur.Hour(), cur.Minute(), cur.Second())
	next := time.Date(cur.Year(), cur.Month(), cur.Day(), h, m, s, 0, time.UTC)
	sig.Set(p.limits.clamp(next))
	popover.EmitChange(p.root)
}

// --- the panel -------------------------------------------------------------

func (p *picker) panel() hx.INode {
	return p.pop.Panel(p.trigger, popover.Options{
		Dialog: p.t.Get("Calendar"),
		Track: func() { // the panel changes height with the view
			p.view.Get()
			p.cursor.Get()
		},
		OnShown: func(hx.INode) { popover.FocusID(p.initialFocusID()) },
	}, func() []hx.INode {
		content := hx.Div().Style("min-width", "17rem")
		content.OnKeyDown(func(ctx hx.EventContext) {
			if popover.Key(ctx) == "Escape" {
				p.pop.Close()
				popover.Focus(p.trigger)
			}
		})
		parts := []hx.INode{}
		if p.mode.hasDate() {
			parts = append(parts, p.header(), p.grid())
		}
		if p.mode.hasTime() {
			parts = append(parts, p.clocks())
		}
		parts = append(parts, p.footer())
		return []hx.INode{content.BodyList(parts)}
	})
}

func (p *picker) initialFocusID() string {
	if !p.mode.hasDate() {
		return p.clockID("from", "h")
	}
	switch hx.UntrackGet[string](p.view) {
	case "day":
		return p.dayID(hx.UntrackGet[time.Time](p.focus))
	}
	return p.id + "-title"
}

func (p *picker) dayID(d time.Time) string { return p.id + "-d-" + d.Format("20060102") }

func (p *picker) monthName(m time.Month) string { return p.t.Get("M" + strconv.Itoa(int(m))) }

// yearPage is the first of the twelve years that the year view shows with y.
func yearPage(y int) int { return y - ((y%12)+12)%12 }

func (p *picker) header() hx.INode {
	nav := func(label, glyph string, by func(time.Time) time.Time) hx.INode {
		return hx.Button().Type("button").Class("btn", "btn-sm", "btn-light").
			Attribute("aria-label", p.t.Get(label)).Text(glyph).
			OnClick(func(hx.EventContext) {
				p.cursor.Set(by(hx.UntrackGet[time.Time](p.cursor)))
			})
	}
	prev := nav("Prev", "‹", func(c time.Time) time.Time { return p.step(c, -1) })
	next := nav("Next", "›", func(c time.Time) time.Time { return p.step(c, +1) })

	title := hx.Button().Type("button").Id(p.id+"-title").Class("btn", "btn-sm", "btn-light", "flex-grow-1", "fw-semibold")
	title.OnClick(func(hx.EventContext) {
		switch hx.UntrackGet[string](p.view) {
		case "day":
			p.view.Set("month")
		case "month":
			p.view.Set("year")
		}
	})
	title.Effect(func() {
		c := p.cursor.Get()
		switch p.view.Get() {
		case "day":
			title.Text(p.monthName(c.Month()) + " " + strconv.Itoa(c.Year()))
			title.Attribute("aria-label", p.t.Get("ChooseMonth"))
		case "month":
			title.Text(strconv.Itoa(c.Year()))
			title.Attribute("aria-label", p.t.Get("ChooseYear"))
		default:
			first := yearPage(c.Year())
			title.Text(strconv.Itoa(first) + " – " + strconv.Itoa(first+11))
			title.RemoveAttribute("aria-label")
		}
	})
	return hx.Div().Class("d-flex", "gap-1", "mb-2").Body(prev, title, next)
}

// step moves the cursor one page forward or back in the view that is shown.
func (p *picker) step(c time.Time, dir int) time.Time {
	switch hx.UntrackGet[string](p.view) {
	case "month":
		return time.Date(c.Year()+dir, c.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "year":
		return time.Date(c.Year()+12*dir, c.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	return addMonths(c, dir)
}

func (p *picker) grid() hx.INode {
	return hx.Dynamic(func() hx.INode {
		view, cursor := p.view.Get(), p.cursor.Get()
		var out hx.INode
		hx.Untrack(func() {
			switch view {
			case "month":
				out = p.monthGrid(cursor)
			case "year":
				out = p.yearGrid(cursor)
			default:
				out = p.dayGrid(cursor)
			}
		})
		return out
	})
}

// weekStart is the day that the weeks of the panel start on (0 is Sunday).
func (p *picker) weekStart() int {
	if p.s.SundayFirst {
		return 0
	}
	return 1
}

// marks is how a cell looks. It reads the choice, so it goes in an effect.
type marks struct{ start, end, between, today, outside bool }

// marksFor says what part of the choice the unit [first,last] is.
func (p *picker) marksFor(first, last time.Time) marks {
	var m marks
	from, to := p.from.Get(), p.to.Get()
	if p.s.Range && to.IsZero() && p.picking.Get() {
		to = p.hover.Get() // while the end is being chosen, the mouse shows it
	}
	fromStart := p.unitStart(from)
	m.start = !from.IsZero() && !first.Before(fromStart) && !first.After(p.unitEnd(from))
	if !to.IsZero() {
		m.end = !first.Before(p.unitStart(to)) && !first.After(p.unitEnd(to))
		if from.IsZero() {
			return m
		}
		lo, hi := fromStart, p.unitStart(to)
		if hi.Before(lo) {
			lo, hi = hi, lo
		}
		m.between = first.After(lo) && first.Before(hi)
	}
	return m
}

func (p *picker) styleCell(b hx.INode, m marks) {
	selected := m.start || m.end
	b.ClassIf("btn-primary", selected)
	b.ClassIf("bg-primary-subtle", m.between && !selected)
	b.ClassIf("fw-bold", m.today && !selected)
	b.ClassIf("text-primary", m.today && !selected)
	b.ClassIf("text-body-secondary", m.outside && !selected && !m.between)
	b.Attribute("aria-pressed", strconv.FormatBool(selected))
}

func (p *picker) dayGrid(cursor time.Time) hx.INode {
	ws := p.weekStart()
	head := make([]hx.INode, 7)
	for i := range head {
		head[i] = hx.Th().Attribute("scope", "col").Class("text-center", "small", "text-body-secondary", "fw-normal").
			Text(p.t.Get("D" + strconv.Itoa((ws+i)%7)))
	}
	rows := make([]hx.INode, 0, 6)
	day := gridStart(firstOfMonth(cursor), ws)
	for w := 0; w < 6; w++ {
		cells := make([]hx.INode, 7)
		for c := range cells {
			cells[c] = hx.Td().Class("p-0").Body(p.dayCell(day, cursor))
			day = day.AddDate(0, 0, 1)
		}
		rows = append(rows, hx.Tr().BodyList(cells))
	}
	table := hx.Table().Class("w-100").Attribute("role", "grid").
		Body(hx.THead().Body(hx.Tr().BodyList(head)), hx.TBody().BodyList(rows))
	table.OnKeyDown(p.dayKey)
	return table
}

func (p *picker) dayCell(day, cursor time.Time) hx.INode {
	disabled := p.limits.outside(day, endOfDay(day))
	b := hx.Button().Type("button").Id(p.dayID(day)).
		Class("btn", "btn-sm", "w-100", "border-0", "rounded-0").
		Attribute("aria-label", strconv.Itoa(day.Day())+" "+p.monthName(day.Month())+" "+strconv.Itoa(day.Year())).
		Text(strconv.Itoa(day.Day()))
	if disabled {
		b.Attribute("disabled", "")
	}
	b.Effect(func() {
		m := p.marksFor(day, endOfDay(day))
		m.today = sameDay(p.today, day)
		m.outside = !sameMonth(day, cursor)
		p.styleCell(b, m)
		if m.today {
			b.Attribute("aria-current", "date")
		} else {
			b.RemoveAttribute("aria-current")
		}
		if sameDay(p.focus.Get(), day) {
			b.Attribute("tabindex", "0")
		} else {
			b.Attribute("tabindex", "-1")
		}
	})
	b.OnClick(func(hx.EventContext) {
		p.focus.Set(day)
		if !sameMonth(day, hx.UntrackGet[time.Time](p.cursor)) {
			p.cursor.Set(firstOfMonth(day))
		}
		p.pick(day)
	})
	b.On(hx.EventMouseEnter, func(hx.EventContext) {
		if hx.UntrackGet[bool](p.picking) {
			p.hover.Set(day)
		}
	})
	return b
}

// dayKey moves the keyboard over the days: arrows by day and week, Page keys by
// month, Home and End to the ends of the week, Enter and Space to choose.
func (p *picker) dayKey(ctx hx.EventContext) {
	cur := hx.UntrackGet[time.Time](p.focus)
	var next time.Time
	switch popover.Key(ctx) {
	case "ArrowLeft":
		next = cur.AddDate(0, 0, -1)
	case "ArrowRight":
		next = cur.AddDate(0, 0, 1)
	case "ArrowUp":
		next = cur.AddDate(0, 0, -7)
	case "ArrowDown":
		next = cur.AddDate(0, 0, 7)
	case "Home":
		next = cur.AddDate(0, 0, -weekdayColumn(cur, p.weekStart()))
	case "End":
		next = cur.AddDate(0, 0, 6-weekdayColumn(cur, p.weekStart()))
	case "PageUp":
		next = sameDayOfMonth(cur, -1)
	case "PageDown":
		next = sameDayOfMonth(cur, +1)
	case "Enter", " ":
		popover.PreventDefault(ctx)
		p.pick(cur)
		return
	default:
		return
	}
	popover.PreventDefault(ctx)
	if p.limits.outside(next, endOfDay(next)) {
		return
	}
	hx.Batch(func() {
		p.focus.Set(next)
		if !sameMonth(next, hx.UntrackGet[time.Time](p.cursor)) {
			p.cursor.Set(firstOfMonth(next))
		}
	})
	popover.FocusID(p.dayID(next))
}

// sameDayOfMonth is the day of the month n months from d, or the last one of that month.
func sameDayOfMonth(d time.Time, n int) time.Time {
	first := addMonths(d, n)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d.Day(), last)-1)
}

func (p *picker) monthGrid(cursor time.Time) hx.INode {
	cells := make([]hx.INode, 12)
	for i := range cells {
		m := time.Date(cursor.Year(), time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
		cells[i] = p.unitCell(m, p.t.Get("S"+strconv.Itoa(i+1)), false, func() {
			hx.Batch(func() {
				p.cursor.Set(m)
				p.focus.Set(m)
				p.view.Set("day")
			})
		})
	}
	return hx.Div().Class("d-grid", "gap-1").Style("grid-template-columns", "repeat(3, 1fr)").BodyList(cells)
}

func (p *picker) yearGrid(cursor time.Time) hx.INode {
	first := yearPage(cursor.Year())
	cells := make([]hx.INode, 12)
	for i := range cells {
		y := time.Date(first+i, 1, 1, 0, 0, 0, 0, time.UTC)
		cells[i] = p.unitCell(y, strconv.Itoa(first+i), true, func() {
			hx.Batch(func() {
				p.cursor.Set(time.Date(y.Year(), cursor.Month(), 1, 0, 0, 0, 0, time.UTC))
				p.view.Set("month")
			})
		})
	}
	return hx.Div().Class("d-grid", "gap-1").Style("grid-template-columns", "repeat(3, 1fr)").BodyList(cells)
}

// unitCell is a month (year false) or a year of its grid. When the mode chooses
// that unit a click chooses it; otherwise the grid is only a way to get to the
// day, and a click moves the view with descend, marking where the choice is.
func (p *picker) unitCell(at time.Time, label string, year bool, descend func()) hx.INode {
	chooses := (year && p.mode == ModeYear) || (!year && p.mode == ModeMonth)
	last := at.AddDate(0, 1, 0).Add(-time.Second)
	if year {
		last = at.AddDate(1, 0, 0).Add(-time.Second)
	}
	in := func(v time.Time) bool { // v is in this month, or this year
		if v.IsZero() {
			return false
		}
		return v.Year() == at.Year() && (year || v.Month() == at.Month())
	}
	b := hx.Button().Type("button").Class("btn", "btn-sm", "py-2").Text(label)
	if chooses && p.limits.outside(at, last) {
		b.Attribute("disabled", "")
	}
	b.Effect(func() {
		var m marks
		if chooses {
			m = p.marksFor(at, last)
		} else {
			m.start = in(p.from.Get())
		}
		m.today = in(p.today)
		p.styleCell(b, m)
	})
	b.OnClick(func(hx.EventContext) {
		if chooses {
			p.pick(at)
			return
		}
		descend()
	})
	b.On(hx.EventMouseEnter, func(hx.EventContext) {
		if chooses && hx.UntrackGet[bool](p.picking) {
			p.hover.Set(at)
		}
	})
	return b
}

// --- time of day -----------------------------------------------------------

func (p *picker) clockID(which, part string) string { return p.id + "-" + part + "-" + which }

func (p *picker) clocks() hx.INode {
	rows := []hx.INode{p.clockRow(false)}
	if p.s.Range {
		rows = append(rows, p.clockRow(true))
	}
	return hx.Div().Class("border-top", "pt-2", "mt-2", "d-grid", "gap-2").BodyList(rows)
}

func (p *picker) clockRow(end bool) hx.INode {
	which, label := "from", p.t.Get("From")
	sig := p.from
	if end {
		which, label, sig = "to", p.t.Get("To"), p.to
	}
	part := func(name, key string, values []int, get func(time.Time) int, set func(h, m, s, v int) (int, int, int)) hx.INode {
		current := hx.Computed(func() string { return strconv.Itoa(get(sig.Get())) })
		sel := hx.Select().Id(p.clockID(which, name)).Class("form-select", "form-select-sm", "w-auto").
			Attribute("aria-label", label+" "+p.t.Get(key)).BindValue(current)
		options := make([]hx.INode, len(values))
		for i, v := range values {
			options[i] = hx.Option().Value(strconv.Itoa(v)).Text(pad(v))
		}
		sel.BodyList(options)
		sel.OnChange(func(ctx hx.EventContext) {
			v, err := strconv.Atoi(popover.TargetValue(ctx))
			if err != nil {
				return
			}
			p.setClock(end, func(h, m, s int) (int, int, int) { return set(h, m, s, v) })
		})
		return sel
	}
	cur := hx.UntrackGet[time.Time](sig)
	name := hx.Span().Class("small", "me-1").Style("min-width", "3rem").Text(label)
	row := []hx.INode{
		name,
		part("h", "Hour", rangeInts(0, 23, 1, -1), func(t time.Time) int { return t.Hour() },
			func(h, m, s, v int) (int, int, int) { return v, m, s }),
		hx.Span().Text(":"),
		part("m", "Minute", rangeInts(0, 59, max(p.s.MinuteStep, 1), cur.Minute()), func(t time.Time) int { return t.Minute() },
			func(h, m, s, v int) (int, int, int) { return h, v, s }),
	}
	if p.s.Seconds {
		row = append(row, hx.Span().Text(":"),
			part("s", "Second", rangeInts(0, 59, 1, -1), func(t time.Time) int { return t.Second() },
				func(h, m, s, v int) (int, int, int) { return h, m, v }))
	}
	return hx.Div().Class("d-flex", "align-items-center", "gap-1").BodyList(row)
}

func pad(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}

// rangeInts are from..to by step, and extra if it is not one of them.
func rangeInts(from, to, step, extra int) []int {
	var out []int
	for v := from; v <= to; v += step {
		out = append(out, v)
	}
	if extra >= from && extra <= to && !slices.Contains(out, extra) {
		out = append(out, extra)
		slices.Sort(out)
	}
	return out
}

// --- the foot of the panel -------------------------------------------------

func (p *picker) footer() hx.INode {
	var buttons []hx.INode
	if !p.today.IsZero() && !p.s.Range {
		key := "Today"
		if p.mode.hasTime() {
			key = "Now"
		}
		buttons = append(buttons, hx.Button().Type("button").Class("btn", "btn-sm", "btn-link").Text(p.t.Get(key)).
			OnClick(func(hx.EventContext) { p.chooseToday() }))
	}
	if p.s.Clearable {
		buttons = append(buttons, hx.Button().Type("button").Class("btn", "btn-sm", "btn-link").Text(p.t.Get("Clear")).
			OnClick(func(hx.EventContext) { p.clear() }))
	}
	if p.mode.hasTime() || p.s.Range {
		buttons = append(buttons, hx.Button().Type("button").Class("btn", "btn-sm", "btn-primary", "ms-auto").Text(p.t.Get("Done")).
			OnClick(func(hx.EventContext) {
				p.pop.Close()
				popover.Focus(p.trigger)
			}))
	}
	if len(buttons) == 0 {
		return hx.Noop()
	}
	return hx.Div().Class("d-flex", "gap-1", "mt-2").BodyList(buttons)
}
