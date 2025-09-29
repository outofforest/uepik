package documents

import (
	_ "embed"
	"text/template"

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
	companyName, companyAddress string,
	currency types.Currency,
	currencyInit types.InitCurrency,
	records []types.BankRecord,
) types.ReportDocument {
	report := BankReport{
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
		Currency:       currency,
	}

	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		var added bool
		for !added || (len(records) > 0 && records[0].Date.Month() == month.Month()) {
			added = true

			var previous BankSummary
			if len(report.Pages) == 0 {
				previous = NewBankSummary(currencyInit)
			} else {
				previous = report.Pages[len(report.Pages)-1].CurrentPageSummary
			}

			page := BankPage{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.Pages),
				Records:             findRecords(&records, month, 26),
				PreviousPageSummary: previous,
				CurrentPageSummary:  previous,
			}
			if len(page.Records) > 0 {
				page.CurrentPageSummary = NewBankSummaryFromRecord(page.Records[len(page.Records)-1])
			}

			report.Pages = append(report.Pages, page)
		}
	}

	return types.ReportDocument{
		Template: bankTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       "BANK." + string(currency.Symbol),
			LockedRows: 7,
		},
	}
}
