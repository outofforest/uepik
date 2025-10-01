package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
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
	CompanyName    string
	CompanyAddress string
	Pages          []BookPage
}

// BookPage is the page in the book report.
type BookPage struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []BookRecord
	CurrentPageSummary  BookSummary
	PreviousPageSummary BookSummary
	MonthSummary        BookSummary
	YearSummary         BookSummary
}

// BookRecord defines the book record.
type BookRecord struct {
	Date            time.Time
	Index           uint64
	DayOfMonth      uint8
	Document        types.Document
	Contractor      types.Contractor
	Notes           string
	IncomeDonations types.Denom
	IncomeTrading   types.Denom
	IncomeOthers    types.Denom
	IncomeSum       types.Denom
	CostTaxed       types.Denom
	CostNotTaxed    types.Denom
}

// NewBookSummary creates new book summary.
func NewBookSummary() BookSummary {
	return BookSummary{
		IncomeDonations: types.BaseZero,
		IncomeTrading:   types.BaseZero,
		IncomeOthers:    types.BaseZero,
		IncomeSum:       types.BaseZero,
		CostTaxed:       types.BaseZero,
		CostNotTaxed:    types.BaseZero,
	}
}

// BookSummary is the page summary of the book report.
type BookSummary struct {
	IncomeDonations types.Denom
	IncomeTrading   types.Denom
	IncomeOthers    types.Denom
	IncomeSum       types.Denom
	CostTaxed       types.Denom
	CostNotTaxed    types.Denom
}

// AddRecord adds record to the summary.
func (bs BookSummary) AddRecord(r BookRecord) BookSummary {
	bs.IncomeDonations = bs.IncomeDonations.Add(r.IncomeDonations)
	bs.IncomeTrading = bs.IncomeTrading.Add(r.IncomeTrading)
	bs.IncomeOthers = bs.IncomeOthers.Add(r.IncomeOthers)
	bs.IncomeSum = bs.IncomeSum.Add(r.IncomeSum)
	bs.CostTaxed = bs.CostTaxed.Add(r.CostTaxed)
	bs.CostNotTaxed = bs.CostNotTaxed.Add(r.CostNotTaxed)
	return bs
}

// AddSummary adds another summary to this one.
func (bs BookSummary) AddSummary(bs2 BookSummary) BookSummary {
	bs.IncomeDonations = bs.IncomeDonations.Add(bs2.IncomeDonations)
	bs.IncomeTrading = bs.IncomeTrading.Add(bs2.IncomeTrading)
	bs.IncomeOthers = bs.IncomeOthers.Add(bs2.IncomeOthers)
	bs.IncomeSum = bs.IncomeSum.Add(bs2.IncomeSum)
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
	const perPage = 9

	report := &BookReport{
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
	}
	entries := coa.Entries(types.NewAccountID(accounts.PiK))
	var index uint64
	summaryYear := NewBookSummary()
	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		summaryPreviousPage := NewBookSummary()
		summaryMonth := NewBookSummary()

		var added bool
		for !added || (len(entries) > 0 && entries[0].GetDate().Month() == month.Month()) {
			added = true

			summaryCurrentPage := NewBookSummary()
			entries := findRecords(&entries, month, perPage)
			records := make([]BookRecord, 0, len(entries))
			for _, e := range entries {
				index++
				r := BookRecord{
					Date:       e.GetDate(),
					Index:      index,
					DayOfMonth: uint8(e.GetDate().Day()),
					Document:   e.GetDocument(),
					Contractor: e.GetContractor(),
					Notes:      e.GetNotes(),
					IncomeDonations: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Przychody,
						accounts.Operacyjne, accounts.Nieodplatna), e.ID).Credit,
					IncomeTrading: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Przychody,
						accounts.Operacyjne, accounts.Odplatna), e.ID).Credit,
					IncomeOthers: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Przychody,
						accounts.Finansowe), e.ID).Credit,
					IncomeSum: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Przychody), e.ID).Credit,
					CostTaxed: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Koszty,
						accounts.Podatkowe), e.ID).Debit,
					CostNotTaxed: coa.Amount(types.NewAccountID(accounts.PiK, accounts.Koszty,
						accounts.Niepodatkowe), e.ID).Debit,
				}
				records = append(records, r)
				summaryCurrentPage = summaryCurrentPage.AddRecord(r)
			}
			summaryMonth = summaryMonth.AddSummary(summaryCurrentPage)
			summaryYear = summaryYear.AddSummary(summaryCurrentPage)

			report.Pages = append(report.Pages, BookPage{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.Pages),
				Records:             records,
				CurrentPageSummary:  summaryCurrentPage,
				PreviousPageSummary: summaryPreviousPage,
				MonthSummary:        summaryMonth,
				YearSummary:         summaryYear,
			})

			summaryPreviousPage = summaryCurrentPage
		}
	}

	return types.ReportDocument{
		Template: bookTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       "PiK",
			LockedRows: 7,
		},
	}
}
