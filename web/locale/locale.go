// Package locale has the translations that belong to the whole application: the
// shared ones and the errors, and the languages it speaks. The ones of a page are
// a localizer.Dict next to its feature.
package locale

import "github.com/deltegui/owl/v3/localizer"

// Languages are the ones the application speaks. The first one is the default.
var Languages = []string{
	localizer.LangEnglish,
	localizer.LangSpanish,
}
