package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.SheetSource = &VATReport{}

	//go:embed vat.tmpl.xml
	vatTmpl     string
	vatTemplate = template.Must(template.New("vat").Parse(vatTmpl))
)

// VATReport is the VAT report.
type VATReport struct {
	CompanyName string
	Months      []VATMonth
	Summary     VATSummary
}

// GetSheet returns sheet to report.
func (r *VATReport) GetSheet() types.Sheet {
	return types.Sheet{
		Template: vatTemplate,
		Data:     r,
		Config: types.SheetConfig{
			Name:       "VAT",
			LockedRows: 6,
		},
	}
}

// VATMonth represents month in the VAT report.
type VATMonth struct {
	Year    uint64
	Month   string
	Records []VATRecord
}

// VATRecord represents VAT record.
type VATRecord struct {
	Date       time.Time
	Index      uint64
	DayOfMonth uint8
	Document   types.Document
	Contractor types.Contractor
	Notes      string
	Income     types.Denom
}

// NewVATSummary creates new VAT summary.
func NewVATSummary() VATSummary {
	return VATSummary{
		Income: types.BaseZero,
	}
}

// VATSummary is the page summary of the VAT report.
type VATSummary struct {
	Income types.Denom
}

// AddRecord adds record to the summary.
func (vs VATSummary) AddRecord(r VATRecord) VATSummary {
	vs.Income = vs.Income.Add(r.Income)
	return vs
}

// NewVATReport generates VAT report.
func NewVATReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName string,
) *VATReport {
	report := &VATReport{
		CompanyName: companyName,
		Summary:     NewVATSummary(),
	}

	var index uint64
	for _, month := range period.Months() {
		entries := coa.EntriesMonth(types.NewAccountID(accounts.VAT), month)
		monthReport := VATMonth{
			Year:    uint64(month.Year()),
			Month:   monthName(month.Month()),
			Records: make([]VATRecord, 0, len(entries)),
		}

		for _, e := range entries {
			index++
			r := VATRecord{
				Date:       e.GetDate(),
				Index:      index,
				DayOfMonth: uint8(e.GetDate().Day()),
				Document:   e.GetDocument(),
				Contractor: e.GetContractor(),
				Notes:      e.GetNotes(),
				Income:     e.Amount.Credit,
			}
			monthReport.Records = append(monthReport.Records, r)
			report.Summary = report.Summary.AddRecord(r)
		}

		report.Months = append(report.Months, monthReport)
	}

	return report
}
