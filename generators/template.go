package generators

import (
	_ "embed"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/types"
)

//go:embed uepik.tmpl.fods
var tmpl string
var tmplParsed = template.Must(template.New("").Parse(tmpl))

// Save saves the report.
func Save(year types.FiscalYear) {
	report := newReport(year)

	f := lo.Must(os.OpenFile("uepik-"+time.Now().Format(time.DateOnly)+".fods", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600))
	defer f.Close()
	lo.Must0(tmplParsed.Execute(f, report))
}

func newReport(year types.FiscalYear) types.Report {
	report := types.Report{
		Book: make([]types.BookReport, 0, 12),
		Flow: make([]types.FlowReport, 0, 12),
		VAT:  make([]types.VATReport, 0, 12),
	}

	bankRecords := Bank(year)
	currencies := lo.Keys(bankRecords)
	sort.Slice(currencies, func(i, j int) bool {
		return strings.Compare(string(currencies[i]), string(currencies[j])) < 0
	})
	for _, c := range currencies {
		report.Bank = append(report.Bank, types.BankCurrency{
			Currency: types.Currencies.Currency(c),
		})
	}

	bookRecords := Book(year)
	vatRecords := VAT(year)
	for month := year.Period.Start; year.Period.Contains(month); month = month.AddDate(0, 1, 0) {
		year := uint64(month.Year())
		monthName := monthName(month.Month())

		report.Book = append(report.Book, types.BookReport{
			Year:    year,
			Month:   monthName,
			Records: findRecords(&bookRecords, month),
		})
		report.Flow = append(report.Flow, types.FlowReport{
			Year:  year,
			Month: monthName,

			MonthIncome:                types.BaseZero,
			MonthCostsTaxed:            types.BaseZero,
			MonthProfit:                types.BaseZero,
			MonthCostsNotTaxedCurrent:  types.BaseZero,
			MonthCostsNotTaxedPrevious: types.BaseZero,

			TotalIncome:                types.BaseZero,
			TotalCostsTaxed:            types.BaseZero,
			TotalProfitYear:            types.BaseZero,
			TotalCostsNotTaxedCurrent:  types.BaseZero,
			TotalProfitPrevious:        types.BaseZero,
			TotalCostsNotTaxedPrevious: types.BaseZero,
			TotalProfit:                types.BaseZero,
		})
		report.VAT = append(report.VAT, types.VATReport{
			Year:    year,
			Month:   monthName,
			Records: findRecords(&vatRecords, month),
		})

		for i, c := range currencies {
			report.Bank[i].Reports = append(report.Bank[i].Reports, types.BankReport{
				Year:    year,
				Month:   monthName,
				Records: findRecords(bankRecords[c], month),
			})
		}
	}

	return report
}

type withDate interface {
	GetDate() time.Time
}

func findRecords[T withDate](records *[]T, month time.Time) []T {
	month = month.AddDate(0, 1, 0)
	i := sort.Search(len(*records), func(i int) bool {
		return !(*records)[i].GetDate().Before(month)
	})
	result := (*records)[:i]
	*records = (*records)[i:]
	return result
}
