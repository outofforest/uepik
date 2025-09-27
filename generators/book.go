package generators

import "github.com/outofforest/uepik/types"

// Book generates book reports.
func Book(year types.FiscalYear) {
	for _, o := range year.Operations {
		o.BookRecords(year.ChartOfAccounts, year.CurrencyRates)
	}
}
