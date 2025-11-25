package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed currencydiff.tmpl.xml
	currencyDiffTmpl     string
	currencyDiffTemplate = template.Must(template.New("currencyDiff").Funcs(template.FuncMap{
		"date": date,
	}).Parse(currencyDiffTmpl))
)

// CurrencyDiffDocument represents currency diff documents.
type CurrencyDiffDocument struct {
	Document types.Document
	Company  types.Contractor
	Records  []CurrencyDiffRecord
	Summary  CurrencyDiffSummary
}

// CurrencyDiffRecord represents currency diff record.
type CurrencyDiffRecord struct {
	Date            time.Time
	Index           uint64
	DayOfMonth      uint8
	Document        types.Document
	PaymentDocument types.DocumentID
	Contractor      types.Contractor
	Amount          types.Denom
	DocumentRate    types.Number
	PaymentRate     types.Number
	Income          types.Denom
	Cost            types.Denom
}

// NewCurrencyDiffSummary returns new summary of currency diff.
func NewCurrencyDiffSummary() CurrencyDiffSummary {
	return CurrencyDiffSummary{
		Income: types.BaseZero,
		Cost:   types.BaseZero,
	}
}

// CurrencyDiffSummary is the summary of currency diff document.
type CurrencyDiffSummary struct {
	Income types.Denom
	Cost   types.Denom
}

// AddRecord adds record to the summary.
func (cds CurrencyDiffSummary) AddRecord(r CurrencyDiffRecord) CurrencyDiffSummary {
	cds.Income = cds.Income.Add(r.Income)
	cds.Cost = cds.Cost.Add(r.Cost)
	return cds
}

// GenerateCurrencyDiffDocument generates currency diff document.
func GenerateCurrencyDiffDocument(
	document types.Document,
	company types.Contractor,
	entries []*types.Entry,
) types.ReportDocument {
	report := &CurrencyDiffDocument{
		Document: document,
		Company:  company,
		Records:  make([]CurrencyDiffRecord, 0, len(entries)),
		Summary:  NewCurrencyDiffSummary(),
	}

	for i, e := range entries {
		data, ok := e.Data.(*types.CurrencyDiff)
		if !ok {
			panic("currency diff data source required")
		}

		r := CurrencyDiffRecord{
			Date:            data.GetDate(),
			Index:           uint64(i + 1),
			DayOfMonth:      uint8(data.GetDate().Day()),
			Document:        data.GetDocument(),
			PaymentDocument: data.BankRecord.Document,
			Contractor:      data.GetContractor(),
			Amount:          data.BankRecord.OriginalAmount.Abs(),
			DocumentRate:    data.DataRate,
			PaymentRate:     data.BankRecord.Rate,
			Income:          e.Amount.Credit,
			Cost:            e.Amount.Debit,
		}
		report.Records = append(report.Records, r)
		report.Summary = report.Summary.AddRecord(r)
	}

	return types.ReportDocument{
		Date:     document.Date,
		Template: currencyDiffTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       document.SheetName,
			LockedRows: 8,
		},
	}
}
