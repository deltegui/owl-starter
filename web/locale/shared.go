package locale

import "github.com/deltegui/owl/v3/localizer"

// Shared is the translations every page can use.
var Shared = localizer.Dict{
	localizer.LangEnglish: {
		"Shared.Language": "Language",
		"Shared.en":       "English",
		"Shared.es":       "Spanish",
		"Shared.Logout":   "Log out",

		"requiredIdentifier":        "The %s field is required",
		"minStringLengthIdentifier": "The %s field needs at least ${} characters",
		"maxStringLengthIdentifier": "The %s field allows at most ${} characters",
		"containsStringIdentifier":  "The %s field must contain ${}",
	},
	localizer.LangSpanish: {
		"Shared.Language": "Idioma",
		"Shared.en":       "Inglés",
		"Shared.es":       "Español",
		"Shared.Logout":   "Cerrar sesión",

		"requiredIdentifier":        "El campo %s es obligatorio",
		"minStringLengthIdentifier": "El campo %s debe tener al menos ${} caracteres",
		"maxStringLengthIdentifier": "El campo %s debe tener como mucho ${} caracteres",
		"containsStringIdentifier":  "El campo %s debe contener ${}",
	},
}
