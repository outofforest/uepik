package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/types"
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
	Document   types.Document
	Contractor types.Contractor
	Pages      []CurrencyDiffPage
	Summary    CurrencyDiffSummary
}

// CurrencyDiffPage is a page in the currency diff document.
type CurrencyDiffPage struct {
	Page    uint64
	Records []CurrencyDiffRecord
	IsLast  bool
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
	contractor types.Contractor,
	entries []*types.Entry,
) types.ReportDocument {
	const perPage = 12

	report := &CurrencyDiffDocument{
		Document:   document,
		Contractor: contractor,
		Summary:    NewCurrencyDiffSummary(),
	}

	var index uint64
	for len(entries) > 0 {
		entriesPage := entries
		if len(entriesPage) > perPage {
			entriesPage = entriesPage[:perPage]
		}
		entries = entries[len(entriesPage):]

		records := make([]CurrencyDiffRecord, 0, len(entriesPage))
		for _, e := range entriesPage {
			data, ok := e.Data.(*types.CurrencyDiff)
			if !ok {
				panic("currency diff data source required")
			}

			index++
			r := CurrencyDiffRecord{
				Date:            data.GetDate(),
				Index:           index,
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
			records = append(records, r)
			report.Summary = report.Summary.AddRecord(r)
		}

		report.Pages = append(report.Pages, CurrencyDiffPage{
			Page:    page(report.Pages),
			Records: records,
		})
	}
	if len(report.Pages) > 0 {
		report.Pages[len(report.Pages)-1].IsLast = true
	}

	return types.ReportDocument{
		Template: currencyDiffTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       document.SheetName,
			LockedRows: 7,
		},
	}
}
