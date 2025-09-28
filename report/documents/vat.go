package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

var (
	//go:embed vat.tmpl.xml
	vatTmpl     string
	vatTemplate = template.Must(template.New("vat").Parse(vatTmpl))
)

// VATReport is the VAT report.
type VATReport struct {
	CompanyName    string
	CompanyAddress string
	Pages          []VATPage
}

// VATPage represents page in the VAT report.
type VATPage struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []VATRecord
	PreviousPageSummary VATSummary
	CurrentPageSummary  VATSummary
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

// GenerateVATReport generates VAT report.
func GenerateVATReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName, companyAddress string,
) types.ReportDocument {
	report := &VATReport{
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
	}
	entries := coa.Entries(types.NewAccountID(accounts.VAT))
	var index uint64
	previousPage := NewVATSummary()
	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		var added bool
		for !added || (len(entries) > 0 && entries[0].Date.Month() == month.Month()) {
			added = true

			vatCurrentPage := previousPage
			entries := findRecords(&entries, month, 25)
			records := make([]VATRecord, 0, len(entries))
			for _, e := range entries {
				index++
				r := VATRecord{
					Date:       e.Date,
					Index:      index,
					DayOfMonth: uint8(e.Date.Day()),
					Document:   e.Document,
					Contractor: e.Contractor,
					Notes:      e.Notes,
					Income:     e.Amount.Credit,
				}
				records = append(records, r)
				vatCurrentPage = vatCurrentPage.AddRecord(r)
			}

			report.Pages = append(report.Pages, VATPage{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.Pages),
				Records:             records,
				CurrentPageSummary:  vatCurrentPage,
				PreviousPageSummary: previousPage,
			})

			previousPage = vatCurrentPage
		}
	}

	return types.ReportDocument{
		Template: vatTemplate,
		Data:     report,
	}
}
