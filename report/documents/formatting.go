package documents

import (
	"sort"
	"time"

	"github.com/outofforest/uepik/types"
)

func notZero(n types.Number) bool {
	return !n.IsZero()
}

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

type withDate interface {
	GetDate() time.Time
}

func findRecords[T withDate](records *[]T, month time.Time, count uint64) []T {
	month = month.AddDate(0, 1, 0)
	i := uint64(sort.Search(len(*records), func(i int) bool {
		return !(*records)[i].GetDate().Before(month)
	}))
	if i > count {
		i = count
	}
	result := (*records)[:i]
	*records = (*records)[i:]
	return result
}

func page[T any](slice []T) uint64 {
	return uint64(len(slice) + 1)
}
