package datepicker

import "github.com/deltegui/owl/v3/localizer"

// texts are the months, the days of the week and the words of the panel. The
// Fmt keys are the layouts (Go's reference date) of what the control shows; what a
// form sends does not depend on the language.
var texts = localizer.Dict{
	localizer.LangEnglish: {
		"M1": "January", "M2": "February", "M3": "March", "M4": "April", "M5": "May", "M6": "June",
		"M7": "July", "M8": "August", "M9": "September", "M10": "October", "M11": "November", "M12": "December",
		"S1": "Jan", "S2": "Feb", "S3": "Mar", "S4": "Apr", "S5": "May", "S6": "Jun",
		"S7": "Jul", "S8": "Aug", "S9": "Sep", "S10": "Oct", "S11": "Nov", "S12": "Dec",
		"D0": "Su", "D1": "Mo", "D2": "Tu", "D3": "We", "D4": "Th", "D5": "Fr", "D6": "Sa",

		"FmtDate": "01/02/2006", "FmtMonth": "01/2006", "FmtYear": "2006", "FmtTime": "15:04", "FmtTimeSec": "15:04:05",

		"PhDate": "Select a date", "PhMonth": "Select a month", "PhYear": "Select a year",
		"PhDateTime": "Select date and time", "PhTime": "Select a time", "PhRange": "Select a range",

		"Calendar": "Calendar", "Today": "Today", "Now": "Now", "Clear": "Clear", "Done": "Done",
		"Prev": "Previous", "Next": "Next", "ChooseMonth": "Choose month", "ChooseYear": "Choose year",
		"Hour": "Hour", "Minute": "Minute", "Second": "Second", "From": "From", "To": "To",
	},
	localizer.LangSpanish: {
		"M1": "enero", "M2": "febrero", "M3": "marzo", "M4": "abril", "M5": "mayo", "M6": "junio",
		"M7": "julio", "M8": "agosto", "M9": "septiembre", "M10": "octubre", "M11": "noviembre", "M12": "diciembre",
		"S1": "ene", "S2": "feb", "S3": "mar", "S4": "abr", "S5": "may", "S6": "jun",
		"S7": "jul", "S8": "ago", "S9": "sep", "S10": "oct", "S11": "nov", "S12": "dic",
		"D0": "Do", "D1": "Lu", "D2": "Ma", "D3": "Mi", "D4": "Ju", "D5": "Vi", "D6": "Sá",

		"FmtDate": "02/01/2006", "FmtMonth": "01/2006", "FmtYear": "2006", "FmtTime": "15:04", "FmtTimeSec": "15:04:05",

		"PhDate": "Elige una fecha", "PhMonth": "Elige un mes", "PhYear": "Elige un año",
		"PhDateTime": "Elige fecha y hora", "PhTime": "Elige una hora", "PhRange": "Elige un rango",

		"Calendar": "Calendario", "Today": "Hoy", "Now": "Ahora", "Clear": "Quitar", "Done": "Hecho",
		"Prev": "Anterior", "Next": "Siguiente", "ChooseMonth": "Elegir mes", "ChooseYear": "Elegir año",
		"Hour": "Hora", "Minute": "Minuto", "Second": "Segundo", "From": "Desde", "To": "Hasta",
	},
}
