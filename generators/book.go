package generators

import (
	"sort"
	"time"

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

	var (
		previousUnspent = types.BaseZero

		month              time.Month
		monthIncome        = types.BaseZero
		monthCostsTaxed    = types.BaseZero
		monthCostsNotTaxed = types.BaseZero

		yearIncome         = types.BaseZero
		yearCostsTaxed     = types.BaseZero
		yearCostsNotTaxed  = types.BaseZero
		yearCostsNotTaxed2 = types.BaseZero
	)

	flowMonth := func() {
		if month != 0 {
			// monthProfit := monthIncome.Sub(monthCostsTaxed)
			monthCostsNotTaxed2 := previousUnspent.Sub(yearCostsNotTaxed2)
			if monthCostsNotTaxed2.GT(monthCostsNotTaxed) {
				monthCostsNotTaxed2 = monthCostsNotTaxed
			}
			monthCostsNotTaxed = monthCostsNotTaxed.Sub(monthCostsNotTaxed2)

			yearIncome = yearIncome.Add(monthIncome)
			yearCostsTaxed = yearCostsTaxed.Add(monthCostsTaxed)
			// yearProfit := yearIncome.Sub(yearCostsTaxed)
			yearCostsNotTaxed = yearCostsNotTaxed.Add(monthCostsNotTaxed)
			yearCostsNotTaxed2 = yearCostsNotTaxed2.Add(monthCostsNotTaxed2)
			// yearUnspent := yearProfit.Add(previousUnspent).Sub(yearCostsNotTaxed).Sub(yearCostsNotTaxed2)

			monthIncome = types.BaseZero
			monthCostsTaxed = types.BaseZero
			monthCostsNotTaxed = types.BaseZero
		}
	}

	for _, r := range records {
		if r.Date.Month() != month {
			flowMonth()
		}
		month = r.Date.Month()
		monthIncome = monthIncome.Add(r.IncomeDonations).Add(r.IncomeTrading).Add(r.IncomeOthers)
		monthCostsTaxed = monthCostsTaxed.Add(r.CostTaxed)
		monthCostsNotTaxed = monthCostsNotTaxed.Add(r.CostNotTaxed)
	}

	flowMonth()

	return records
}
