package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/types"
)

var (
	//go:embed unrecordedsell.tmpl.xml
	unrecordedSellTmpl     string
	unrecordedSellTemplate = template.Must(template.New("unrecordedSell").Funcs(template.FuncMap{
		"date": date,
	}).Parse(unrecordedSellTmpl))
)

// UnrecordedSellDocument represents unrecorded sell documents.
type UnrecordedSellDocument struct {
	Document   types.Document
	Contractor types.Contractor
	Pages      []UnrecordedSellPage
	Summary    UnrecordedSellSummary
}

// UnrecordedSellPage is a page in the unrecorded sell document.
type UnrecordedSellPage struct {
	Page    uint64
	Records []UnrecordedSellRecord
	IsLast  bool
}

// UnrecordedSellRecord represents unrecorded sell record.
type UnrecordedSellRecord struct {
	Index      uint64
	Document   types.Document
	Contractor types.Contractor
	Income     types.Denom
}

// NewUnrecordedSellSummary returns new summary of unrecorded sell.
func NewUnrecordedSellSummary() UnrecordedSellSummary {
	return UnrecordedSellSummary{
		Income: types.BaseZero,
	}
}

// UnrecordedSellSummary is the summary of unrecorded sell document.
type UnrecordedSellSummary struct {
	Income types.Denom
}

// AddRecord adds record to the summary.
func (uss UnrecordedSellSummary) AddRecord(r UnrecordedSellRecord) UnrecordedSellSummary {
	uss.Income = uss.Income.Add(r.Income)
	return uss
}

// GenerateUnrecordedSellDocument generates unrecorded sell document.
func GenerateUnrecordedSellDocument(
	document types.Document,
	contractor types.Contractor,
	entries []*types.Entry,
) types.ReportDocument {
	const perPage = 9

	report := &UnrecordedSellDocument{
		Document:   document,
		Contractor: contractor,
		Summary:    NewUnrecordedSellSummary(),
	}

	var index uint64
	for len(entries) > 0 {
		entriesPage := entries
		if len(entriesPage) > perPage {
			entriesPage = entriesPage[:perPage]
		}
		entries = entries[len(entriesPage):]

		records := make([]UnrecordedSellRecord, 0, len(entriesPage))
		for _, e := range entriesPage {
			index++
			r := UnrecordedSellRecord{
				Index:      index,
				Document:   e.GetDocument(),
				Contractor: e.GetContractor(),
				Income:     e.Amount.Credit,
			}
			records = append(records, r)
			report.Summary = report.Summary.AddRecord(r)
		}

		report.Pages = append(report.Pages, UnrecordedSellPage{
			Page:    page(report.Pages),
			Records: records,
		})
	}
	if len(report.Pages) > 0 {
		report.Pages[len(report.Pages)-1].IsLast = true
	}

	return types.ReportDocument{
		Date:     document.Date,
		Template: unrecordedSellTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       document.SheetName,
			LockedRows: 8,
		},
	}
}
