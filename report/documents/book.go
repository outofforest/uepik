package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed book.tmpl.xml
	bookTmpl     string
	bookTemplate = template.Must(template.New("book").Funcs(template.FuncMap{
		"notZero": notZero,
	}).Parse(bookTmpl))
)

// BookReport is the book report.
type BookReport struct {
	CompanyName string
	Months      []BookMonth
}

// BookMonth is the book month.
type BookMonth struct {
	Year                       uint64
	Month                      string
	Records                    []BookRecord
	AccumulatedPreviousSummary BookSummary
	MonthSummary               BookSummary
	AccumulatedCurrentSummary  BookSummary
}

// BookRecord defines the book record.
type BookRecord struct {
	Date         time.Time
	Index        uint64
	DayOfMonth   uint8
	Document     types.Document
	Contractor   types.Contractor
	Notes        string
	Income       types.Denom
	CostTaxed    types.Denom
	CostNotTaxed types.Denom
}

// NewBookSummary creates new book summary.
func NewBookSummary() BookSummary {
	return BookSummary{
		Income:       types.BaseZero,
		CostTaxed:    types.BaseZero,
		CostNotTaxed: types.BaseZero,
	}
}

// BookSummary is the page summary of the book report.
type BookSummary struct {
	Show         bool
	Income       types.Denom
	CostTaxed    types.Denom
	CostNotTaxed types.Denom
}

// AddRecord adds record to the summary.
func (bs BookSummary) AddRecord(r BookRecord) BookSummary {
	bs.Income = bs.Income.Add(r.Income)
	bs.CostTaxed = bs.CostTaxed.Add(r.CostTaxed)
	bs.CostNotTaxed = bs.CostNotTaxed.Add(r.CostNotTaxed)
	return bs
}

// AddSummary adds another summary to this one.
func (bs BookSummary) AddSummary(bs2 BookSummary) BookSummary {
	bs.Income = bs.Income.Add(bs2.Income)
	bs.CostTaxed = bs.CostTaxed.Add(bs2.CostTaxed)
	bs.CostNotTaxed = bs.CostNotTaxed.Add(bs2.CostNotTaxed)
	return bs
}

// GenerateBookReport generates book report.
func GenerateBookReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName, companyAddress string,
) types.ReportDocument {
	report := &BookReport{
		CompanyName: companyName,
	}
	var index uint64
	summaryAccumulatedCurrent := NewBookSummary()
	for _, month := range period.Months() {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		entries := coa.EntriesMonth(types.NewAccountID(accounts.PiK), month)
		monthReport := BookMonth{
			Year:                       yearNumber,
			Month:                      monthName,
			Records:                    make([]BookRecord, 0, len(entries)),
			AccumulatedPreviousSummary: summaryAccumulatedCurrent,
			MonthSummary:               NewBookSummary(),
		}

		for _, e := range entries {
			index++
			r := BookRecord{
				Date:       e.GetDate(),
				Index:      index,
				DayOfMonth: uint8(e.GetDate().Day()),
				Document:   e.GetDocument(),
				Contractor: e.GetContractor(),
				Notes:      e.GetNotes(),
				Income:     coa.Amount(types.NewAccountID(accounts.PiK, accounts.Przychody), e.ID).Credit,
				CostTaxed: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Koszty,
					accounts.Podatkowe), e.ID).Debit,
				CostNotTaxed: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Koszty,
					accounts.Niepodatkowe), e.ID).Debit,
			}
			monthReport.Records = append(monthReport.Records, r)
			monthReport.MonthSummary = monthReport.MonthSummary.AddRecord(r)
		}

		summaryAccumulatedCurrent = summaryAccumulatedCurrent.AddSummary(monthReport.MonthSummary)
		monthReport.AccumulatedCurrentSummary = summaryAccumulatedCurrent
		report.Months = append(report.Months, monthReport)
	}

	return types.ReportDocument{
		Template: bookTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       "PiK",
			LockedRows: 6,
		},
	}
}
