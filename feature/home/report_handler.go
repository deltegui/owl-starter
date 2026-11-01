package home

import (
	"time"

	"github.com/deltegui/owl/v3"

	homeweb "github.com/deltegui/owl-starter/feature/home/web"
)

// handleReport pretends to be a report that takes seconds to build. Leaving the
// page cancels the request, and the wait ends with it.
func handleReport() owl.Handler {
	return func(ctx owl.Ctx) error {
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Context().Done():
			return nil
		}
		return ctx.Json(homeweb.Report{Rows: []homeweb.ReportRow{
			{Name: "Orders", Value: 128},
			{Name: "Customers", Value: 42},
			{Name: "Returns", Value: 7},
		}})
	}
}
