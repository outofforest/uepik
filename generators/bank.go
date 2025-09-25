package generators

import (
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/outofforest/uepik/types"
)

// Bank generates bank reports.
func Bank(year types.FiscalYear) []types.BankReport {
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

		for i, br := range records {
			br.Index = uint64(i + 1)
			br.DayOfMonth = uint8(br.Date.Day())

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
				panic(errors.New("invalid data in bank record"))
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

	return reports
}
