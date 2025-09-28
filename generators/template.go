package generators

import (
	_ "embed"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

//go:embed uepik.tmpl.fods
var tmpl string
var tmplParsed = template.Must(template.New("").Funcs(template.FuncMap{
	"notZero": notZero,
}).Parse(tmpl))

// Save saves the report.
func Save(year *types.FiscalYear, currencyRates types.CurrencyRates, years ...*types.FiscalYear) {
	report := newReport(year, currencyRates, years)

	f := lo.Must(os.OpenFile("uepik-"+time.Now().Format(time.DateOnly)+".fods", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600))
	defer f.Close()
	lo.Must0(tmplParsed.Execute(f, report))
}

func newReport(year *types.FiscalYear, currencyRates types.CurrencyRates, years []*types.FiscalYear) types.Report {
	coa := year.ChartOfAccounts
	period := year.Period
	unspentProfit := coa.OpeningBalance(types.NewAccountID(accounts.NiewydatkowanyDochod))

	bankRecords, opBankRecords := year.BankReports(currencyRates, years)
	year.BookRecords(currencyRates, opBankRecords)

	report := types.Report{
		CompanyName:    year.CompanyName,
		CompanyAddress: year.CompanyAddress,
		Book:           make([]types.BookReport, 0, 12),
		Flow:           make([]types.FlowReport, 0, 12),
		VAT:            make([]types.VATReport, 0, 12),
	}

	currencies := lo.Keys(bankRecords)
	sort.Slice(currencies, func(i, j int) bool {
		return strings.Compare(string(currencies[i]), string(currencies[j])) < 0
	})
	for _, c := range currencies {
		report.Bank = append(report.Bank, types.BankCurrency{
			Currency: types.Currencies.Currency(c),
		})
	}

	bookEntries := coa.Entries(types.NewAccountID(accounts.CIT))
	vatEntries := coa.Entries(types.NewAccountID(accounts.VAT))

	bookYear := types.NewBookSummary()
	var bookIndex, vatIndex uint64
	vatPreviousPage := types.NewVATSummary()
	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		bookPreviousPage := types.NewBookSummary()
		bookMonth := types.NewBookSummary()

		var bookAdded bool
		for !bookAdded || (len(bookEntries) > 0 && bookEntries[0].Date.Month() == month.Month()) {
			bookAdded = true

			bookCurrentPage := types.NewBookSummary()
			entries := findRecords(&bookEntries, month, 10)
			records := make([]types.BookRecord, 0, len(entries))
			for _, e := range entries {
				bookIndex++
				r := types.BookRecord{
					Date:       e.Date,
					Index:      bookIndex,
					DayOfMonth: uint8(e.Date.Day()),
					Document:   e.Document,
					Contractor: e.Contractor,
					Notes:      e.Notes,
					IncomeDonations: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Przychody,
						accounts.Operacyjne, accounts.ZNieodplatnejDPP), e.ID).Credit,
					IncomeTrading: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Przychody,
						accounts.Operacyjne, accounts.ZOdplatnejDPP), e.ID).Credit,
					IncomeOthers: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Przychody,
						accounts.Nieoperacyjne), e.ID).Credit,
					IncomeSum: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Przychody), e.ID).Credit,
					CostTaxed: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Koszty,
						accounts.Podatkowe), e.ID).Debit,
					CostNotTaxed: coa.Amount(types.NewAccountID(accounts.CIT, accounts.Koszty,
						accounts.Niepodatkowe), e.ID).Debit,
				}
				records = append(records, r)
				bookCurrentPage = bookCurrentPage.AddRecord(r)
			}
			bookMonth = bookMonth.AddSummary(bookCurrentPage)
			bookYear = bookYear.AddSummary(bookCurrentPage)

			report.Book = append(report.Book, types.BookReport{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.Book),
				Records:             records,
				CurrentPageSummary:  bookCurrentPage,
				PreviousPageSummary: bookPreviousPage,
				MonthSummary:        bookMonth,
				YearSummary:         bookYear,
			})

			bookPreviousPage = bookCurrentPage
		}

		monthIncome := coa.BalanceMonth(types.NewAccountID(accounts.CIT, accounts.Przychody), month)
		monthCostsTaxed := coa.BalanceMonth(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.Podatkowe),
			month)
		yearIncome := coa.BalanceIncremental(types.NewAccountID(accounts.CIT, accounts.Przychody), month)
		yearCostsTaxed := coa.BalanceIncremental(types.NewAccountID(accounts.CIT, accounts.Koszty,
			accounts.Podatkowe), month)

		report.Flow = append(report.Flow, types.FlowReport{
			Year:  yearNumber,
			Month: monthName,

			MonthIncome:     monthIncome,
			MonthCostsTaxed: monthCostsTaxed,
			MonthProfit:     monthIncome.Sub(monthCostsTaxed),
			MonthCostsNotTaxedCurrent: coa.DebitMonth(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.WTrakcieRoku), month),
			MonthCostsNotTaxedPrevious: coa.DebitMonth(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.ZLatUbieglych), month),

			TotalIncome:     yearIncome,
			TotalCostsTaxed: yearCostsTaxed,
			TotalProfitYear: yearIncome.Sub(yearCostsTaxed),
			TotalCostsNotTaxedCurrent: coa.DebitIncremental(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.WTrakcieRoku), month),
			TotalProfitPrevious: unspentProfit,
			TotalCostsNotTaxedPrevious: coa.DebitIncremental(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.ZLatUbieglych), month),
			TotalProfit: coa.BalanceIncremental(types.NewAccountID(accounts.NiewydatkowanyDochod), month),
		})

		var vatAdded bool
		for !vatAdded || (len(vatEntries) > 0 && vatEntries[0].Date.Month() == month.Month()) {
			vatAdded = true

			vatCurrentPage := vatPreviousPage
			entries := findRecords(&vatEntries, month, 25)
			records := make([]types.VATRecord, 0, len(entries))
			for _, e := range entries {
				vatIndex++
				r := types.VATRecord{
					Date:       e.Date,
					Index:      vatIndex,
					DayOfMonth: uint8(e.Date.Day()),
					Document:   e.Document,
					Contractor: e.Contractor,
					Notes:      e.Notes,
					Income:     e.Amount.Credit,
				}
				records = append(records, r)
				vatCurrentPage = vatCurrentPage.AddRecord(r)
			}

			report.VAT = append(report.VAT, types.VATReport{
				Year:                yearNumber,
				Month:               monthName,
				Page:                page(report.VAT),
				Records:             records,
				CurrentPageSummary:  vatCurrentPage,
				PreviousPageSummary: vatPreviousPage,
			})

			vatPreviousPage = vatCurrentPage
		}

		bankAdded := map[types.CurrencySymbol]bool{}
		for i, c := range currencies {
			for !bankAdded[c] || (len(*bankRecords[c]) > 0 && (*bankRecords[c])[0].Date.Month() == month.Month()) {
				bankAdded[c] = true

				var previous types.BankSummary
				if len(report.Bank[i].Reports) == 0 {
					currencyInit, exists := year.Init.Currencies[c]
					if !exists {
						panic("brak bilansu otwarcia waluty")
					}
					previous = types.NewBankSummary(currencyInit)
				} else {
					previous = report.Bank[i].Reports[len(report.Bank[i].Reports)-1].CurrentPageSummary
				}

				bankReport := types.BankReport{
					Year:                yearNumber,
					Month:               monthName,
					Page:                page(report.Bank[i].Reports),
					Records:             findRecords(bankRecords[c], month, 26),
					PreviousPageSummary: previous,
					CurrentPageSummary:  previous,
				}
				if len(bankReport.Records) > 0 {
					bankReport.CurrentPageSummary = bankReport.Records[len(bankReport.Records)-1].Summary()
				}

				report.Bank[i].Reports = append(report.Bank[i].Reports, bankReport)
			}
		}
	}

	return report
}

type withDate interface {
	GetDate() time.Time
}

func findRecords[T withDate](records *[]T, month time.Time, count uint64) []T {
	month = month.AddDate(0, 1, 0)
	i := uint64(sort.Search(len(*records), func(i int) bool {
		return !(*records)[i].GetDate().Before(month)
	}))
	if i > count {
		i = count
	}
	result := (*records)[:i]
	*records = (*records)[i:]
	return result
}

func page[T any](slice []T) uint64 {
	return uint64(len(slice) + 1)
}
