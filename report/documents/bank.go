package documents

import (
	_ "embed"
	"sort"
	"strings"
	"text/template"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/types"
)

var (
	//go:embed bank.tmpl.xml
	bankTmpl     string
	bankTemplate = template.Must(template.New("bank").Parse(bankTmpl))
)

// BankReport is the bank report for currency.
type BankReport struct {
	CompanyName    string
	CompanyAddress string
	Currency       types.Currency
	Pages          []BankPage
}

// BankPage is the page in the bank report.
type BankPage struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []types.BankRecord
	PreviousPageSummary BankSummary
	CurrentPageSummary  BankSummary
}

// NewBankSummary creates new bank summary.
func NewBankSummary(currencyInit types.InitCurrency) BankSummary {
	return BankSummary{
		OriginalSum: currencyInit.OriginalSum,
		BaseSum:     currencyInit.BaseSum,
		RateAverage: currencyInit.BaseSum.Rate(currencyInit.OriginalSum),
	}
}

// NewBankSummaryFromRecord creates summary from record.
func NewBankSummaryFromRecord(r types.BankRecord) BankSummary {
	return BankSummary{
		OriginalSum: r.OriginalSum,
		BaseSum:     r.BaseSum,
		RateAverage: r.RateAverage,
	}
}

// BankSummary is the page summary of the bank record.
type BankSummary struct {
	OriginalSum types.Denom
	BaseSum     types.Denom
	RateAverage types.Number
}

// GenerateBankReport generates bank report.
func GenerateBankReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName, companyAddress string,
	currencyInit types.InitCurrencies,
	records map[types.CurrencySymbol]*[]types.BankRecord,
) types.ReportDocument {
	currencies := lo.Keys(records)
	sort.Slice(currencies, func(i, j int) bool {
		return strings.Compare(string(currencies[i]), string(currencies[j])) < 0
	})

	reports := []BankReport{}
	for _, c := range currencies {
		reports = append(reports, BankReport{
			CompanyName:    companyName,
			CompanyAddress: companyAddress,
			Currency:       types.Currencies.Currency(c),
		})
	}

	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		bankAdded := map[types.CurrencySymbol]bool{}
		for i, c := range currencies {
			for !bankAdded[c] || (len(*records[c]) > 0 && (*records[c])[0].Date.Month() == month.Month()) {
				bankAdded[c] = true

				var previous BankSummary
				if len(reports[i].Pages) == 0 {
					currencyInit, exists := currencyInit[c]
					if !exists {
						panic("brak bilansu otwarcia waluty")
					}
					previous = NewBankSummary(currencyInit)
				} else {
					previous = reports[i].Pages[len(reports[i].Pages)-1].CurrentPageSummary
				}

				page := BankPage{
					Year:                yearNumber,
					Month:               monthName,
					Page:                page(reports[i].Pages),
					Records:             findRecords(records[c], month, 26),
					PreviousPageSummary: previous,
					CurrentPageSummary:  previous,
				}
				if len(page.Records) > 0 {
					page.CurrentPageSummary = NewBankSummaryFromRecord(page.Records[len(page.Records)-1])
				}

				reports[i].Pages = append(reports[i].Pages, page)
			}
		}
	}

	return types.ReportDocument{
		Template: bankTemplate,
		Data:     reports,
	}
}
