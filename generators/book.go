//nolint:dupl
package generators

import (
	"sort"

	"github.com/outofforest/uepik/types"
)

// Book generates book reports.
func Book(year types.FiscalYear) []types.BookRecord {
	records := []types.BookRecord{}
	for _, o := range year.Operations {
		records = append(records, o.BookRecords(year.Period, year.CurrencyRates)...)
	}
	for i := range records {
		records[i].Index = uint64(i)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.Before(records[j].Date) || (records[i].Date.Equal(records[j].Date) &&
			records[i].Index < records[j].Index)
	})

	for i := range records {
		records[i].Index = uint64(i + 1)
		records[i].DayOfMonth = uint8(records[i].Date.Day())
	}

	return records
}
