//nolint:misspell
package generators

import (
	"fmt"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/outofforest/uepik/types"
)

const (
	operationsSheetName = "Przychody i koszty"
	flowSheetName       = "Przepływy finansowe"
)

// Book generates book reports.
func Book(f *excelize.File, year types.FiscalYear) error {
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

	lo.Must(f.NewSheet(operationsSheetName))
	lo.Must(f.NewSheet(flowSheetName))

	baseStyle := lo.Must(f.NewStyle(excelNumberFormat(types.Currencies.Currency(types.PLN).AmountPrecision)))
	intStyle := lo.Must(f.NewStyle(intStyle))
	textStyle := lo.Must(f.NewStyle(textStyle))

	lo.Must0(f.SetColStyle(operationsSheetName, "A:A", intStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "B:B", intStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "C:C", textStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "D:D", textStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "E:E", textStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "F:F", textStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "G:G", baseStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "H:H", baseStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "I:I", baseStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "J:J", baseStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "K:K", baseStyle))
	lo.Must0(f.SetColStyle(operationsSheetName, "L:L", baseStyle))

	lo.Must0(f.SetRowStyle(operationsSheetName, 1, 1, lo.Must(f.NewStyle(styleHeader))))
	lo.Must0(f.SetCellStr(operationsSheetName, "A1", "Lp."))
	lo.Must0(f.SetCellStr(operationsSheetName, "B1", "Data zdarzenia lub operacji"))
	lo.Must0(f.SetCellStr(operationsSheetName, "C1", "Nr dowodu księgowego"))
	lo.Must0(f.SetCellStr(operationsSheetName, "D1", "Nazwa"))
	lo.Must0(f.SetCellStr(operationsSheetName, "E1", "Adres"))
	lo.Must0(f.SetCellStr(operationsSheetName, "F1", "Opis zdarzenia"))
	lo.Must0(f.SetCellStr(operationsSheetName, "G1", "Przychody z działalności nieodpłatnej pożytku publicznego"))
	lo.Must0(f.SetCellStr(operationsSheetName, "H1",
		"Przychody z działalności odpłatnej pożytku publicznego z tytułu sprzedaży towarów i usług"))
	lo.Must0(f.SetCellStr(operationsSheetName, "I1", "Pozostałe przychody"))
	lo.Must0(f.SetCellStr(operationsSheetName, "J1", "Razem przychody"))
	lo.Must0(f.SetCellStr(operationsSheetName, "K1", "Koszty uzyskania przychodów"))
	lo.Must0(f.SetCellStr(operationsSheetName, "L1", "Koszty niestanowiące kosztów uzyskania przychodów"))

	lo.Must0(f.SetColStyle(flowSheetName, "A:A", textStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "B:B", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "C:C", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "D:D", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "E:E", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "F:F", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "G:G", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "H:H", baseStyle))
	lo.Must0(f.SetColStyle(flowSheetName, "I:I", baseStyle))

	lo.Must0(f.SetRowStyle(flowSheetName, 1, 1, lo.Must(f.NewStyle(styleHeader))))
	lo.Must0(f.SetCellStr(flowSheetName, "B1", "Przychody"))
	lo.Must0(f.SetCellStr(flowSheetName, "C1", "Koszty uzyskania przychodów"))
	lo.Must0(f.SetCellStr(flowSheetName, "D1", "Dochód"))
	lo.Must0(f.SetCellStr(flowSheetName, "E1", "Dochód wolny od podatku przeznaczony na cele statutowe"))
	lo.Must0(f.SetCellStr(flowSheetName, "F1",
		"Wydatki na cele statutowe pokryte z dochodu zwolnionego od podatku w roku podatkowym"))
	lo.Must0(f.SetCellStr(flowSheetName, "G1", "Dochód wolny od podatku z lat ubiegłych przeznaczony na cele statutowe"))
	lo.Must0(f.SetCellStr(flowSheetName, "H1",
		"Wydatki na cele statutowe pokryte z dochodu z lat ubiegłych zwolnionego od podatku"))
	lo.Must0(f.SetCellStr(flowSheetName, "I1", "Ogółem dochód wolny od podatku niewydatkowany na cele statutowe"))

	var (
		previousProfit = types.BaseZero

		month               time.Month
		monthIncome         = types.BaseZero
		monthCostsTaxed     = types.BaseZero
		monthCostsNotTaxed  = types.BaseZero
		monthCostsNotTaxed2 = types.BaseZero

		yearIncome         = types.BaseZero
		yearCostsTaxed     = types.BaseZero
		yearCostsNotTaxed  = types.BaseZero
		yearCostsNotTaxed2 = types.BaseZero
	)

	flowMonth := func() {
		if month != 0 {
			row1 := 2*(month-1) + 2
			row2 := row1 + 1

			yearIncome = yearIncome.Add(monthIncome)
			yearCostsTaxed = yearCostsTaxed.Add(monthCostsTaxed)
			yearProfit := yearIncome.Sub(yearCostsTaxed)

			monthProfit := monthIncome.Sub(monthCostsTaxed)
			if monthCostsNotTaxed.GT(yearProfit) {
				monthCostsNotTaxed2 = monthCostsNotTaxed.Sub(yearProfit)
				monthCostsNotTaxed = yearProfit
			}
			monthUnspent := monthProfit.Add(previousProfit).Sub(monthCostsNotTaxed).Sub(monthCostsNotTaxed2)

			yearCostsNotTaxed = yearCostsNotTaxed.Add(monthCostsNotTaxed)
			yearUnspent := yearProfit.Add(previousProfit).Sub(yearCostsNotTaxed).Sub(yearCostsNotTaxed2)

			lo.Must0(f.SetCellStr(flowSheetName, fmt.Sprintf("A%d", row1), monthName(month)))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("B%d", row1),
				monthIncome.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("C%d", row1),
				monthCostsTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("D%d", row1),
				monthProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("E%d", row1),
				monthProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("F%d", row1),
				monthCostsNotTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("G%d", row1),
				previousProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("H%d", row1),
				monthCostsNotTaxed2.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("I%d", row1),
				monthUnspent.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))

			lo.Must0(f.SetCellStr(flowSheetName, fmt.Sprintf("A%d", row2), "od początku roku"))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("B%d", row2),
				yearIncome.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("C%d", row2),
				yearCostsTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("D%d", row2),
				yearProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("E%d", row2),
				yearProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("F%d", row2),
				yearCostsNotTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("G%d", row2),
				previousProfit.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("H%d", row2),
				yearCostsNotTaxed2.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(flowSheetName, fmt.Sprintf("I%d", row2),
				yearUnspent.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))

			monthIncome = types.BaseZero
			monthCostsTaxed = types.BaseZero
			monthCostsNotTaxed = types.BaseZero
			monthCostsNotTaxed2 = types.BaseZero
		}
	}

	for i, r := range records {
		if r.Date.Month() != month {
			flowMonth()
		}
		month = r.Date.Month()
		monthIncome = monthIncome.Add(r.IncomeDonations).Add(r.IncomeTrading).Add(r.IncomeOthers)
		monthCostsTaxed = monthCostsTaxed.Add(r.CostTaxed)
		monthCostsNotTaxed = monthCostsNotTaxed.Add(r.CostNotTaxed)

		row := i + 2

		lo.Must0(f.SetCellInt(operationsSheetName, fmt.Sprintf("A%d", row), int64(i)+1))
		lo.Must0(f.SetCellInt(operationsSheetName, fmt.Sprintf("B%d", row), int64(r.Date.Day())))
		lo.Must0(f.SetCellStr(operationsSheetName, fmt.Sprintf("C%d", row), r.Document.ID))
		lo.Must0(f.SetCellStr(operationsSheetName, fmt.Sprintf("D%d", row), r.Contractor.Name))
		lo.Must0(f.SetCellStr(operationsSheetName, fmt.Sprintf("E%d", row), r.Contractor.Address))
		lo.Must0(f.SetCellStr(operationsSheetName, fmt.Sprintf("F%d", row), r.Notes))

		if r.IncomeDonations.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("G%d", row),
				r.IncomeDonations.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.IncomeTrading.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("H%d", row),
				r.IncomeTrading.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.IncomeOthers.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("I%d", row),
				r.IncomeOthers.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}

		incomeSum := r.IncomeDonations.Add(r.IncomeTrading).Add(r.IncomeOthers)
		if incomeSum.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("J%d", row),
				incomeSum.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}

		if r.CostTaxed.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("K%d", row),
				r.CostTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.CostNotTaxed.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(operationsSheetName, fmt.Sprintf("L%d", row),
				r.CostNotTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
	}

	flowMonth()

	return nil
}
