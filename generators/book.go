//nolint:misspell
package generators

import (
	"fmt"
	"sort"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/outofforest/uepik/types"
)

const bookSheetName = "Przychody i koszty"

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

	lo.Must(f.NewSheet(bookSheetName))

	baseStyle := lo.Must(f.NewStyle(excelNumberFormat(types.Currencies.Currency(types.PLN).AmountPrecision)))
	intStyle := lo.Must(f.NewStyle(intStyle))
	textStyle := lo.Must(f.NewStyle(textStyle))

	lo.Must0(f.SetColStyle(bookSheetName, "A:A", intStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "B:B", intStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "C:C", textStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "D:D", textStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "E:E", textStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "F:F", textStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "G:G", baseStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "H:H", baseStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "I:I", baseStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "J:J", baseStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "K:K", baseStyle))
	lo.Must0(f.SetColStyle(bookSheetName, "L:L", baseStyle))

	lo.Must0(f.SetRowStyle(bookSheetName, 1, 1, lo.Must(f.NewStyle(styleHeader))))
	lo.Must0(f.SetCellStr(bookSheetName, "A1", "Lp."))
	lo.Must0(f.SetCellStr(bookSheetName, "B1", "Data zdarzenia lub operacji"))
	lo.Must0(f.SetCellStr(bookSheetName, "C1", "Nr dowodu księgowego"))
	lo.Must0(f.SetCellStr(bookSheetName, "D1", "Nazwa"))
	lo.Must0(f.SetCellStr(bookSheetName, "E1", "Adres"))
	lo.Must0(f.SetCellStr(bookSheetName, "F1", "Opis zdarzenia"))
	lo.Must0(f.SetCellStr(bookSheetName, "G1", "Przychody z działalności nieodpłatnej pożytku publicznego"))
	lo.Must0(f.SetCellStr(bookSheetName, "H1",
		"Przychody z działalności odpłatnej pożytku publicznego z tytułu sprzedaży towarów i usług"))
	lo.Must0(f.SetCellStr(bookSheetName, "I1", "Pozostałe przychody"))
	lo.Must0(f.SetCellStr(bookSheetName, "J1", "Razem przychody"))
	lo.Must0(f.SetCellStr(bookSheetName, "K1", "Koszty uzyskania przychodów"))
	lo.Must0(f.SetCellStr(bookSheetName, "L1", "Koszty niestanowiące kosztów uzyskania przychodów"))

	for i, r := range records {
		lo.Must0(f.SetCellInt(bookSheetName, fmt.Sprintf("A%d", i+rowOffset), int64(i)+1))
		lo.Must0(f.SetCellInt(bookSheetName, fmt.Sprintf("B%d", i+rowOffset), int64(r.Date.Day())))
		lo.Must0(f.SetCellStr(bookSheetName, fmt.Sprintf("C%d", i+rowOffset), r.Document.ID))
		lo.Must0(f.SetCellStr(bookSheetName, fmt.Sprintf("D%d", i+rowOffset), r.Contractor.Name))
		lo.Must0(f.SetCellStr(bookSheetName, fmt.Sprintf("E%d", i+rowOffset), r.Contractor.Address))
		lo.Must0(f.SetCellStr(bookSheetName, fmt.Sprintf("F%d", i+rowOffset), r.Notes))

		if r.IncomeDonations.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(bookSheetName, fmt.Sprintf("G%d", i+rowOffset),
				r.IncomeDonations.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.IncomeTrading.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(bookSheetName, fmt.Sprintf("H%d", i+rowOffset),
				r.IncomeTrading.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.IncomeOthers.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(bookSheetName, fmt.Sprintf("H%d", i+rowOffset),
				r.IncomeOthers.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.CostTaxed.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(bookSheetName, fmt.Sprintf("K%d", i+rowOffset),
				r.CostTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
		if r.CostNotTaxed.NEQ(types.BaseZero) {
			lo.Must0(f.SetCellFloat(bookSheetName, fmt.Sprintf("L%d", i+rowOffset),
				r.CostNotTaxed.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		}
	}

	return nil
}
