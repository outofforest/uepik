package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/v2/types"
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
	Records    []UnrecordedSellRecord
	Summary    UnrecordedSellSummary
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
	report := &UnrecordedSellDocument{
		Document:   document,
		Contractor: contractor,
		Records:    make([]UnrecordedSellRecord, 0, len(entries)),
		Summary:    NewUnrecordedSellSummary(),
	}

	for i, e := range entries {
		r := UnrecordedSellRecord{
			Index:      uint64(i + 1),
			Document:   e.GetDocument(),
			Contractor: e.GetContractor(),
			Income:     e.Amount.Credit,
		}
		report.Records = append(report.Records, r)
		report.Summary = report.Summary.AddRecord(r)
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
