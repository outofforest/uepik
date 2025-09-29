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
	currencyDiffTemplate = template.Must(template.New("currencyDiff").Parse(currencyDiffTmpl))
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
}

// CurrencyDiffRecord represents currency diff record.
type CurrencyDiffRecord struct {
	Date       time.Time
	Index      uint64
	DayOfMonth uint8
	Document   types.Document
	Contractor types.Contractor
	Notes      string
	Income     types.Denom
}

// NewCurrencyDiffSummary returns new summary of currency diff.
func NewCurrencyDiffSummary() CurrencyDiffSummary {
	return CurrencyDiffSummary{
		Income: types.BaseZero,
	}
}

// CurrencyDiffSummary is the summary of currency diff document.
type CurrencyDiffSummary struct {
	Income types.Denom
}

// AddRecord adds record to the summary.
func (cds CurrencyDiffSummary) AddRecord(r CurrencyDiffRecord) CurrencyDiffSummary {
	cds.Income = cds.Income.Add(r.Income)
	return cds
}

// GenerateCurrencyDiffDocument generates currency diff document.
func GenerateCurrencyDiffDocument(
	document types.Document,
	contractor types.Contractor,
	entries []*types.Entry,
) types.ReportDocument {
	const perPage = 25

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
			index++
			r := CurrencyDiffRecord{
				Date:       e.GetDate(),
				Index:      index,
				DayOfMonth: uint8(e.GetDate().Day()),
				Document:   e.GetDocument(),
				Contractor: e.GetContractor(),
				Notes:      e.GetNotes(),
				Income:     e.Amount.Credit,
			}
			records = append(records, r)
			report.Summary = report.Summary.AddRecord(r)
		}

		report.Pages = append(report.Pages, CurrencyDiffPage{
			Page:    page(report.Pages),
			Records: records,
		})
	}

	return types.ReportDocument{
		Template: currencyDiffTemplate,
		Data:     report,
	}
}
