package lazy_test

import (
	"context"
	"strings"
	"testing"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/ssr"

	"github.com/deltegui/owl-starter/web/ui/lazy"
)

type data struct{ Name string }

func TestServerShowsSpinnerAndDoesNotFetch(t *testing.T) {
	for name, when := range map[string]hx.Gettable[bool]{
		"closed": hx.Signal(false),
		"open":   hx.Signal(true), // even open, the server cannot ask
		"nil":    nil,
	} {
		fetched := 0
		html := ssr.Render(func() hx.INode {
			return lazy.Load(lazy.Options[data]{
				When:  when,
				Fetch: func(context.Context) (data, error) { fetched++; return data{"x"}, nil },
				View:  func(d data) hx.INode { return hx.P().Text(d.Name) },
			})
		})
		if fetched != 0 {
			t.Errorf("%s: the server fetched %d times", name, fetched)
		}
		if !strings.Contains(html, "spinner-border") || strings.Contains(html, "<p>") {
			t.Errorf("%s: want the spinner only, got %s", name, html)
		}
	}
}

func TestNodeCanBeBuiltManyTimes(t *testing.T) {
	var lz *lazy.Lazy[data]
	ssr.Render(func() hx.INode {
		lz = lazy.New(lazy.Options[data]{
			Fetch: func(context.Context) (data, error) { return data{}, nil },
			View:  func(d data) hx.INode { return hx.P().Text(d.Name) },
		})
		return hx.Noop()
	})
	for i := 0; i < 2; i++ {
		if got := ssr.Render(lz.Node); !strings.Contains(got, "spinner-border") {
			t.Errorf("build %d: %s", i, got)
		}
	}
}

func TestNeedsFetchAndView(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New without Fetch and View must panic")
		}
	}()
	lazy.New(lazy.Options[data]{})
}
