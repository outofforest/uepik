package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed bank.tmpl.xml
	bankTmpl     string
	bankTemplate = template.Must(template.New("bank").Parse(bankTmpl))
)

// BankReport is the bank report for currency.
type BankReport struct {
	CompanyName     string
	Currency        types.Currency
	Months          []BankMonth
	PreviousSummary BankSummary
	CurrentSummary  BankSummary
}

// BankMonth is the month in the bank report.
type BankMonth struct {
	Year    uint64
	Month   string
	Records []types.BankRecord
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
	companyName string,
	currency types.Currency,
	currencyInit types.InitCurrency,
	records []types.BankRecord,
) types.ReportDocument {
	const perPage = 18

	summary := NewBankSummary(currencyInit)
	report := BankReport{
		CompanyName:     companyName,
		Currency:        currency,
		PreviousSummary: summary,
		CurrentSummary:  summary,
	}
	for _, month := range period.Months() {
		var i int
		for ; i < len(records) && records[i].Date.Month() == month.Month(); i++ {
		}
		monthReport := BankMonth{
			Year:    uint64(month.Year()),
			Month:   monthName(month.Month()),
			Records: records[:i],
		}
		records = records[i:]
		if l := len(monthReport.Records); l > 0 {
			report.CurrentSummary = NewBankSummaryFromRecord(monthReport.Records[l-1])
		}
		report.Months = append(report.Months, monthReport)
	}

	return types.ReportDocument{
		Template: bankTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       "BANK." + string(currency.Symbol),
			LockedRows: 6,
		},
	}
}
