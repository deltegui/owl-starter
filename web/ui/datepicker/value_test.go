package datepicker

import (
	"errors"
	"testing"
	"time"
)

func TestParseAndFormatRoundTrip(t *testing.T) {
	cases := []struct {
		mode    Mode
		seconds bool
		text    string
	}{
		{ModeYear, false, "2026"},
		{ModeMonth, false, "2026-11"},
		{ModeDate, false, "2026-11-01"},
		{ModeDateTime, false, "2026-11-01T14:30"},
		{ModeDateTime, true, "2026-11-01T14:30:45"},
		{ModeTime, false, "09:05"},
		{ModeTime, true, "23:59:59"},
	}
	for _, c := range cases {
		got, err := Parse(c.mode, c.text)
		if err != nil {
			t.Errorf("Parse(%s, %q): %v", c.mode, c.text, err)
			continue
		}
		if back := Format(c.mode, c.seconds, got); back != c.text {
			t.Errorf("%s: %q came back as %q", c.mode, c.text, back)
		}
	}
}

func TestParseRejectsWhatIsNotAValue(t *testing.T) {
	for _, c := range []struct {
		mode Mode
		text string
	}{
		{ModeDate, "2026-02-30"}, {ModeDate, "01/11/2026"}, {ModeDate, "2026-11"}, {ModeMonth, "2026-13"},
		{ModeYear, "26"}, {ModeDateTime, "2026-11-01"}, {ModeDateTime, "2026-11-01T25:00"}, {ModeTime, "24:00"},
	} {
		if _, err := Parse(c.mode, c.text); !errors.Is(err, ErrInvalid) {
			t.Errorf("Parse(%s, %q) = %v, want ErrInvalid", c.mode, c.text, err)
		}
	}
}

func TestEmptyIsNothing(t *testing.T) {
	got, err := Parse(ModeDate, "")
	if err != nil || !got.IsZero() {
		t.Errorf("empty = %v, %v", got, err)
	}
	if Format(ModeDate, false, time.Time{}) != "" {
		t.Error("the zero Time is an empty text")
	}
}

func TestDateTimeAcceptsWithAndWithoutSeconds(t *testing.T) {
	a, _ := Parse(ModeDateTime, "2026-11-01T14:30")
	b, _ := Parse(ModeDateTime, "2026-11-01T14:30:00")
	if !a.Equal(b) || a.IsZero() {
		t.Errorf("%v != %v", a, b)
	}
}

func TestTimeHasAFixedDay(t *testing.T) {
	got, _ := Parse(ModeTime, "14:30")
	if got.Year() != 2000 || got.Hour() != 14 || got.Minute() != 30 {
		t.Errorf("got %v", got)
	}
}

func TestNormalizeDropsWhatTheModeDoesNotChoose(t *testing.T) {
	at := time.Date(2026, 11, 17, 14, 30, 45, 0, time.UTC)
	if got := normalize(ModeYear, at); got != time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("year: %v", got)
	}
	if got := normalize(ModeMonth, at); got != time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("month: %v", got)
	}
	if got := normalize(ModeDate, at); got != time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC) {
		t.Errorf("date: %v", got)
	}
}

func TestWeekGrid(t *testing.T) {
	nov := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC) // a Sunday
	if got := gridStart(nov, 1); got != time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) {
		t.Errorf("weeks that start on Monday: %v", got)
	}
	if got := gridStart(nov, 0); !got.Equal(nov) {
		t.Errorf("weeks that start on Sunday: %v", got)
	}
}

func TestAddMonthsCrossesYears(t *testing.T) {
	got := addMonths(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), 3)
	if got != time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("got %v", got)
	}
	if got := addMonths(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), -1); got != time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("back: %v", got)
	}
}

func TestBounds(t *testing.T) {
	b := bounds{min: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC), max: time.Date(2026, 11, 20, 23, 59, 59, 0, time.UTC)}
	day := func(d int) (time.Time, time.Time) {
		s := time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC)
		return s, endOfDay(s)
	}
	for d, want := range map[int]bool{9: true, 10: false, 20: false, 21: true} {
		if got := b.outside(day(d)); got != want {
			t.Errorf("day %d outside = %v, want %v", d, got, want)
		}
	}
	if got := b.clamp(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)); !got.Equal(b.max) {
		t.Errorf("clamp = %v", got)
	}
}
