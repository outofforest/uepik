package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/types"
)

var (
	//go:embed category.tmpl.xml
	categoryTmpl     string
	categoryTemplate = template.Must(template.New("category").Parse(categoryTmpl))
)

// CategoryReport is the category report.
type CategoryReport struct {
	Title          string
	SheetName      string
	CompanyName    string
	CompanyAddress string
	Pages          []CategoryPage
}

// CategoryPage is the page in the category report.
type CategoryPage struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []CategoryRecord
	PreviousPageSummary CategorySummary
	CurrentPageSummary  CategorySummary
}

// NewCategorySummary creates new category summary.
func NewCategorySummary() CategorySummary {
	return CategorySummary{
		Income: types.BaseZero,
		Cost:   types.BaseZero,
	}
}

// CategorySummary is the page summary of the category record.
type CategorySummary struct {
	Income types.Denom
	Cost   types.Denom
}

// AddRecord adds record to the summary.
func (cs CategorySummary) AddRecord(r CategoryRecord) CategorySummary {
	cs.Income = cs.Income.Add(r.Income)
	cs.Cost = cs.Cost.Add(r.Cost)
	return cs
}

// CategoryRecord defines the properties of category record.
type CategoryRecord struct {
	Date       time.Time
	Index      uint64
	DayOfMonth uint8
	Document   types.Document
	Contractor types.Contractor
	Income     types.Denom
	Cost       types.Denom
}

// GenerateCategoryReport generates bank report.
func GenerateCategoryReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName, companyAddress string,
	title, sheetName string,
	accountID types.AccountID,
) types.ReportDocument {
	report := CategoryReport{
		Title:          title,
		SheetName:      sheetName,
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
	}

	entries := coa.Entries(accountID)
	var index uint64
	previous := NewCategorySummary()
	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		var added bool
		for !added || (len(entries) > 0 && entries[0].GetDate().Month() == month.Month()) {
			added = true

			current := previous
			entries := findRecords(&entries, month, 26)
			records := make([]CategoryRecord, 0, len(entries))
			for _, e := range entries {
				index++

				r := CategoryRecord{
					Date:       e.GetDate(),
					Index:      index,
					DayOfMonth: uint8(e.GetDate().Day()),
					Document:   e.GetDocument(),
					Contractor: e.GetContractor(),
					Income:     e.Amount.Credit,
					Cost:       e.Amount.Debit,
				}
				records = append(records, r)
				current = current.AddRecord(r)
			}

			page := CategoryPage{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.Pages),
				Records:             records,
				PreviousPageSummary: previous,
				CurrentPageSummary:  current,
			}

			report.Pages = append(report.Pages, page)
			previous = current
		}
	}

	return types.ReportDocument{
		Template: categoryTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       sheetName,
			LockedRows: 7,
		},
	}
}
