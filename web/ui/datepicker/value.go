package datepicker

import (
	"errors"
	"time"
)

// Mode says what the picker chooses, and so how its value is written.
type Mode string

const (
	ModeYear     Mode = "year"     // 2026
	ModeMonth    Mode = "month"    // 2026-11
	ModeDate     Mode = "date"     // 2026-11-01 (the default)
	ModeDateTime Mode = "datetime" // 2026-11-01T14:30, or :45 with State.Seconds
	ModeTime     Mode = "time"     // 14:30, or 14:30:45 with State.Seconds
)

func (m Mode) orDefault() Mode {
	switch m {
	case ModeYear, ModeMonth, ModeDate, ModeDateTime, ModeTime:
		return m
	}
	return ModeDate
}

func (m Mode) hasDate() bool { return m != ModeTime }
func (m Mode) hasTime() bool { return m == ModeDateTime || m == ModeTime }

// A value is a wall-clock time with no time zone, in UTC only to hold the numbers
// (the components of the date and the time are what matter). The zero Time means
// "nothing chosen". The conversion between this wall clock and the UTC that the
// application stores belongs to the application, at the edges.

// refDay is the date of a Time that has only a time of day.
var refDay = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func layoutOf(mode Mode, seconds bool) string {
	switch mode {
	case ModeYear:
		return "2006"
	case ModeMonth:
		return "2006-01"
	case ModeDateTime:
		if seconds {
			return "2006-01-02T15:04:05"
		}
		return "2006-01-02T15:04"
	case ModeTime:
		if seconds {
			return "15:04:05"
		}
		return "15:04"
	}
	return "2006-01-02"
}

// ErrInvalid is what Parse answers to a text that is not a value of the mode.
var ErrInvalid = errors.New("datepicker: invalid value")

// Parse reads the text a form sends back: 2026, 2026-11, 2026-11-01,
// 2026-11-01T14:30[:45] or 14:30[:45], depending on the mode. An empty text is
// the zero Time and no error, for a field the user left empty. A mode that has
// no date gives the time on 2000-01-01.
func Parse(mode Mode, text string) (time.Time, error) {
	if text == "" {
		return time.Time{}, nil
	}
	mode = mode.orDefault()
	layouts := []string{layoutOf(mode, false)}
	if mode.hasTime() {
		layouts = []string{layoutOf(mode, true), layoutOf(mode, false)}
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, text); err == nil {
			return normalize(mode, t), nil
		}
	}
	return time.Time{}, ErrInvalid
}

// Format writes a value as the form sends it. The zero Time is an empty text.
func Format(mode Mode, seconds bool, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return normalize(mode.orDefault(), t).Format(layoutOf(mode.orDefault(), seconds))
}

// normalize drops what the mode does not choose, so equal values are equal.
func normalize(mode Mode, t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	switch mode {
	case ModeYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case ModeMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	case ModeDate:
		return startOfDay(t)
	case ModeTime:
		return time.Date(2000, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	}
	return t
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func endOfDay(t time.Time) time.Time { return startOfDay(t).AddDate(0, 0, 1).Add(-time.Second) }

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// addMonths is the first day of the month n months after the one t is in.
func addMonths(t time.Time, n int) time.Time {
	return time.Date(t.Year(), t.Month()+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
}

func sameDay(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func sameMonth(a, b time.Time) bool {
	return !a.IsZero() && !b.IsZero() && a.Year() == b.Year() && a.Month() == b.Month()
}

// weekdayColumn is the column of the day in a week that starts on weekStart
// (0 is Sunday).
func weekdayColumn(t time.Time, weekStart int) int {
	return (int(t.Weekday()) - weekStart + 7) % 7
}

// gridStart is the first day of the six weeks that show the month of first.
func gridStart(first time.Time, weekStart int) time.Time {
	return first.AddDate(0, 0, -weekdayColumn(first, weekStart))
}

// bounds are Min and Max of a picker: the zero Time is no limit.
type bounds struct{ min, max time.Time }

// outside says whether nothing between a and b (inclusive) can be chosen.
func (b bounds) outside(a, z time.Time) bool {
	return (!b.min.IsZero() && z.Before(b.min)) || (!b.max.IsZero() && a.After(b.max))
}

func (b bounds) clamp(t time.Time) time.Time {
	if !b.min.IsZero() && t.Before(b.min) {
		return b.min
	}
	if !b.max.IsZero() && t.After(b.max) {
		return b.max
	}
	return t
}
