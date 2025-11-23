package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed category.tmpl.xml
	categoryTmpl     string
	categoryTemplate = template.Must(template.New("category").Parse(categoryTmpl))
)

// CategoryReport is the category report.
type CategoryReport struct {
	Title       string
	SheetName   string
	CompanyName string
	Months      []CategoryMonth
	Summary     CategorySummary
}

// CategoryMonth is the month in the category report.
type CategoryMonth struct {
	Year    uint64
	Month   string
	Records []CategoryRecord
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
	companyName string,
	title, sheetName string,
	accountID types.AccountID,
) types.ReportDocument {
	report := CategoryReport{
		Title:       title,
		SheetName:   sheetName,
		CompanyName: companyName,
		Summary:     NewCategorySummary(),
	}

	var index uint64
	for _, month := range period.Months() {
		entries := coa.EntriesMonth(accountID, month)
		monthReport := CategoryMonth{
			Year:    uint64(month.Year()),
			Month:   monthName(month.Month()),
			Records: make([]CategoryRecord, 0, len(entries)),
		}

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
			monthReport.Records = append(monthReport.Records, r)
			report.Summary = report.Summary.AddRecord(r)
		}

		report.Months = append(report.Months, monthReport)
	}

	return types.ReportDocument{
		Template: categoryTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       sheetName,
			LockedRows: 6,
		},
	}
}
