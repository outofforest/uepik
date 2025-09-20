package generators

import (
	"fmt"
	"sort"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/outofforest/uepik/types"
)

const vatSheetName = "VAT"

// VAT generates VAT reports.
func VAT(f *excelize.File, year types.FiscalYear) error {
	records := []types.VATRecord{}
	for _, o := range year.Operations {
		records = append(records, o.VATRecords(year.Period, year.CurrencyRates)...)
	}
	for i := range records {
		records[i].Index = uint64(i)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.Before(records[j].Date) || (records[i].Date.Equal(records[j].Date) &&
			records[i].Index < records[j].Index)
	})

	lo.Must(f.NewSheet(vatSheetName))

	baseStyle := lo.Must(f.NewStyle(excelNumberFormat(types.Currencies.Currency(types.PLN).AmountPrecision)))
	dateStyle := lo.Must(f.NewStyle(dateStyle))
	textStyle := lo.Must(f.NewStyle(textStyle))

	lo.Must0(f.SetPageLayout(vatSheetName, &excelize.PageLayoutOptions{
		Orientation: lo.ToPtr("portrait"),
	}))
	lo.Must0(f.SetPageMargins(vatSheetName, &excelize.PageLayoutMarginsOptions{
		Left:   lo.ToPtr(0.0),
		Right:  lo.ToPtr(0.0),
		Top:    lo.ToPtr(0.0),
		Bottom: lo.ToPtr(0.0),
		Header: lo.ToPtr(0.0),
		Footer: lo.ToPtr(0.0),
	}))

	lo.Must0(f.SetColStyle(vatSheetName, "A:A", dateStyle))
	lo.Must0(f.SetColStyle(vatSheetName, "B:B", textStyle))
	lo.Must0(f.SetColStyle(vatSheetName, "C:C", textStyle))
	lo.Must0(f.SetColStyle(vatSheetName, "D:D", baseStyle))
	lo.Must0(f.SetColStyle(vatSheetName, "E:E", textStyle))

	lo.Must0(f.SetColWidth(vatSheetName, "A", "A", width(2.29)))
	lo.Must0(f.SetColWidth(vatSheetName, "B", "B", width(3.78)))
	lo.Must0(f.SetColWidth(vatSheetName, "C", "C", width(3.54)))
	lo.Must0(f.SetColWidth(vatSheetName, "D", "D", width(1.78)))
	lo.Must0(f.SetColWidth(vatSheetName, "E", "E", width(3.78)))

	lo.Must0(f.SetRowStyle(vatSheetName, 1, 1, lo.Must(f.NewStyle(headerStyle))))
	lo.Must0(f.SetCellStr(vatSheetName, "A1", "Data powstania obowiązku VAT"))
	lo.Must0(f.SetCellStr(vatSheetName, "B1", "Nr dowodu księgowego"))
	lo.Must0(f.SetCellStr(vatSheetName, "C1", "Kontrahent"))
	lo.Must0(f.SetCellStr(vatSheetName, "D1", "Wartość sprzedaży"))
	lo.Must0(f.SetCellStr(vatSheetName, "E1", "Wyjaśnienia"))

	incomeSum := types.BaseZero
	for i, r := range records {
		row := i + 3
		incomeSum = incomeSum.Add(r.Income)

		lo.Must0(f.SetCellValue(vatSheetName, fmt.Sprintf("A%d", row), r.Date))
		lo.Must0(f.SetCellStr(vatSheetName, fmt.Sprintf("B%d", row), r.Document.ID))
		lo.Must0(f.SetCellStr(vatSheetName, fmt.Sprintf("C%d", row), r.Contractor.Name))
		lo.Must0(f.SetCellFloat(vatSheetName, fmt.Sprintf("D%d", row),
			r.Income.Amount.ToFloat64(), int(types.BaseCurrency.AmountPrecision), 64))
		lo.Must0(f.SetCellStr(vatSheetName, fmt.Sprintf("E%d", row), r.Notes))
	}

	lo.Must0(f.SetCellStr(vatSheetName, "A2", "SUMA"))
	lo.Must0(f.SetCellFloat(vatSheetName, "D2", incomeSum.Amount.ToFloat64(),
		int(types.BaseCurrency.AmountPrecision), 64))

	return nil
}
