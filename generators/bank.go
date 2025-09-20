package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/outofforest/uepik/types"
)

// Bank generates bank reports.
func Bank(f *excelize.File, year types.FiscalYear) error {
	reports, err := bankReports(year)
	if err != nil {
		return err
	}

	return visualizeBankReports(f, reports)
}

func bankReports(year types.FiscalYear) ([]types.BankReport, error) {
	currencies := map[types.CurrencySymbol][]*types.BankRecord{}
	for _, o := range year.Operations {
		for _, br := range o.BankRecords(year.Period) {
			currencies[br.OriginalAmount.Currency] = append(currencies[br.OriginalAmount.Currency], br)
		}
	}

	var zeroDenom types.Denom
	var zeroRate types.Number

	reports := make([]types.BankReport, 0, len(currencies))

	for currencySymbol, records := range currencies {
		currency := types.Currencies.Currency(currencySymbol)

		sort.Slice(records, func(i, j int) bool {
			return records[i].Date.Before(records[j].Date) || (records[i].Date.Equal(records[j].Date) &&
				records[i].Index < records[j].Index)
		})

		report := types.BankReport{
			OriginalCurrency: currency,
			BaseCurrency:     types.BaseCurrency,
			Records:          make([]types.BankRecord, 0, len(records)),
		}

		originalZero := types.Denom{
			Currency: currencySymbol,
			Amount:   types.NewNumber(0, 0, currency.AmountPrecision),
		}
		originalSum := originalZero
		baseSum := types.Denom{
			Currency: types.PLN,
			Amount:   types.NewNumber(0, 0, types.BaseCurrency.AmountPrecision),
		}
		rate := types.NewNumber(0, 0, currency.RatePrecision)

		for _, br := range records {
			switch {
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate == zeroRate &&
				br.OriginalAmount.GT(originalZero):
				br.BaseAmount, br.Rate = year.CurrencyRates.ToBase(br.OriginalAmount, types.PreviousDay(br.Date))
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate == zeroRate:
				br.Rate = rate
				br.BaseAmount = br.OriginalAmount.ToBase(rate)
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate != zeroRate:
				br.BaseAmount = br.OriginalAmount.ToBase(br.Rate)
			case br.OriginalAmount != zeroDenom && br.BaseAmount != zeroDenom && br.Rate == zeroRate:
				br.Rate = br.BaseAmount.Rate(br.OriginalAmount)
			default:
				return nil, errors.New("invalid data in bank record")
			}

			originalSum = originalSum.Add(br.OriginalAmount)
			baseSum = baseSum.Add(br.BaseAmount)
			rate = baseSum.Rate(originalSum)

			br.OriginalSum = originalSum
			br.BaseSum = baseSum
			br.RateAverage = rate

			report.Records = append(report.Records, *br)
		}

		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return strings.Compare(string(reports[i].OriginalCurrency.Symbol),
			string(reports[j].OriginalCurrency.Symbol)) < 0
	})

	return reports, nil
}

func visualizeBankReports(f *excelize.File, reports []types.BankReport) error {
	for _, r := range reports {
		lo.Must(f.NewSheet(string(r.OriginalCurrency.Symbol)))
		originalStyle := lo.Must(f.NewStyle(excelNumberFormat(r.OriginalCurrency.AmountPrecision)))
		baseStyle := lo.Must(f.NewStyle(excelNumberFormat(r.BaseCurrency.AmountPrecision)))
		rateStyle := lo.Must(f.NewStyle(excelNumberFormat(r.OriginalCurrency.RatePrecision)))

		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "A:A", lo.Must(f.NewStyle(dateStyle))))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "B:B", originalStyle))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "C:C", baseStyle))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "D:D", rateStyle))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "E:E", originalStyle))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "F:F", baseStyle))
		lo.Must0(f.SetColStyle(string(r.OriginalCurrency.Symbol), "G:G", rateStyle))

		lo.Must0(f.SetRowStyle(string(r.OriginalCurrency.Symbol), 1, 1, lo.Must(f.NewStyle(styleHeader))))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "A1", "Data operacji"))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "B1", "Kwota "+string(r.OriginalCurrency.Symbol)))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "C1", "Kwota "+string(r.BaseCurrency.Symbol)))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "D1", "Kurs operacji"))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "E1", "Suma "+string(r.OriginalCurrency.Symbol)))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "F1", "Suma "+string(r.BaseCurrency.Symbol)))
		lo.Must0(f.SetCellStr(string(r.OriginalCurrency.Symbol), "G1", "Kurs średni"))

		for i, br := range r.Records {
			row := i + 2

			lo.Must0(f.SetCellValue(string(r.OriginalCurrency.Symbol), fmt.Sprintf("A%d", row),
				br.Date))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("B%d", row),
				br.OriginalAmount.Amount.ToFloat64(),
				int(r.OriginalCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("C%d", row),
				br.BaseAmount.Amount.ToFloat64(),
				int(r.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("D%d", row),
				br.Rate.ToFloat64(),
				int(r.OriginalCurrency.RatePrecision), 64))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("E%d", row),
				br.OriginalSum.Amount.ToFloat64(),
				int(r.OriginalCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("F%d", row),
				br.BaseSum.Amount.ToFloat64(),
				int(r.BaseCurrency.AmountPrecision), 64))
			lo.Must0(f.SetCellFloat(string(r.OriginalCurrency.Symbol), fmt.Sprintf("G%d", row),
				br.RateAverage.ToFloat64(),
				int(r.OriginalCurrency.RatePrecision), 64))
		}
	}

	return nil
}
