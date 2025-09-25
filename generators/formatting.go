package generators

import "time"

func monthName(month time.Month) string {
	switch month {
	case time.January:
		return "styczeń"
	case time.February:
		return "luty"
	case time.March:
		return "marzec"
	case time.April:
		return "kwiecień"
	case time.May:
		return "maj"
	case time.June:
		return "czerwiec"
	case time.July:
		return "lipiec"
	case time.August:
		return "sierpień"
	case time.September:
		return "wrzesień"
	case time.October:
		return "październik"
	case time.November:
		return "listopad"
	case time.December:
		return "grudzień"
	default:
		panic("invalid month")
	}
}
