package report

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/report/documents"
	"github.com/outofforest/uepik/v2/types"
	"github.com/outofforest/uepik/v2/types/operations"
)

var coaAccounts = []*types.Account{
	types.NewAccount(
		accounts.PiK, types.Liabilities, types.AllValid(),
		types.NewAccount(
			accounts.Przychody, types.Incomes, types.AllValid(),
			types.NewAccount(
				accounts.Finansowe, types.Incomes, types.AllValid(),
				types.NewAccount(accounts.DodatnieRozniceKursowe, types.Incomes,
					types.ValidSources(&operations.CurrencyDiffSource{})),
			),
			types.NewAccount(
				accounts.Operacyjne, types.Incomes, types.AllValid(),
				types.NewAccount(accounts.Nieodplatna, types.Incomes, types.ValidSources(&operations.Donation{})),
				types.NewAccount(accounts.Odplatna, types.Incomes, types.ValidSources(
					&operations.Sell{},
					&operations.UnrecordedSellSource{},
				)),
			),
		),
		types.NewAccount(
			accounts.Koszty, types.Costs, types.AllValid(),
			types.NewAccount(
				accounts.Podatkowe, types.Costs, types.AllValid(),
				types.NewAccount(
					accounts.Finansowe, types.Costs, types.AllValid(),
					types.NewAccount(accounts.UjemneRozniceKursowe, types.Costs,
						types.ValidSources(&operations.CurrencyDiffSource{})),
				),
				types.NewAccount(accounts.Operacyjne, types.Costs, types.ValidSources(&operations.Purchase{})),
			),
			types.NewAccount(
				accounts.Niepodatkowe, types.Costs, types.AllValid(),
				types.NewAccount(accounts.Operacyjne, types.Costs, types.ValidSources(&operations.Purchase{})),
			),
		),
	),
	types.NewAccount(accounts.VAT, types.Incomes, types.ValidSources(&types.VAT{})),
	types.NewAccount(
		accounts.NiewydatkowanyDochod, types.Liabilities, types.ValidSources(
			&operations.CurrencyDiffSource{},
			&operations.Donation{},
			&operations.Purchase{},
			&operations.Sell{},
		),
	),
	types.NewAccount(
		accounts.RozniceKursowe, types.Liabilities, types.AllValid(),
		types.NewAccount(accounts.Nieodplatna, types.Liabilities, types.ValidSources(&types.CurrencyDiff{})),
		types.NewAccount(accounts.Odplatna, types.Liabilities, types.ValidSources(&types.CurrencyDiff{})),
	),
	types.NewAccount(accounts.SprzedazNieewidencjonowana, types.Incomes, types.ValidSources(&operations.Sell{})),
	types.NewAccount(accounts.Nieodplatna, types.Liabilities, types.ValidSources(
		&operations.CurrencyDiffSource{},
		&operations.Donation{},
		&operations.Purchase{},
	)),
	types.NewAccount(accounts.Odplatna, types.Liabilities, types.ValidSources(
		&operations.CurrencyDiffSource{},
		&operations.Sell{},
		&operations.Purchase{},
	)),
}

//go:embed report.tmpl.xml
var tmpl string
var tmplParsed = template.Must(template.New("report").Parse(tmpl))

//go:embed config.tmpl.xml
var configTmpl string
var configTmplParsed = template.Must(template.New("config").Parse(configTmpl))

// Save saves the report.
func Save(
	viewDate time.Time,
	currentYear *types.FiscalYear,
	currencyRates types.CurrencyRates,
	years []*types.FiscalYear,
) {
	report := newReport(viewDate, currentYear, currencyRates, years)

	file := filepath.Join("reports", "uepik-"+viewDate.Format(time.DateOnly)+".fods")
	lo.Must0(os.MkdirAll(filepath.Dir(file), 0o700))
	f := lo.Must(os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600))
	defer f.Close()
	lo.Must0(tmplParsed.Execute(f, report))
}

func newReport(
	viewDate time.Time,
	year *types.FiscalYear,
	currencyRates types.CurrencyRates,
	years []*types.FiscalYear,
) types.Report {
	if year.Period.End.After(viewDate) {
		year.Period.End = viewDate
	}

	coa := types.NewChartOfAccounts(year.Period, coaAccounts...)
	coa.OpenAccount(types.NewAccountID(accounts.NiewydatkowanyDochod), types.CreditBalance(year.Init.UnspentProfit))

	company := types.Contractor{
		Name:    year.CompanyName,
		Address: year.CompanyAddress,
		TaxID:   year.CompanyTaxID,
	}
	year.Operations = append(
		year.Operations,
		&operations.UnrecordedSell{
			Contractor: company,
		},
		&operations.CurrencyDiff{
			Contractor: company,
		},
	)

	bankRecords, opBankRecords := year.BankReports(currencyRates, years)

	opDocs := year.BookRecords(coa, currencyRates, opBankRecords)

	docs := []types.ReportDocument{
		documents.GenerateBookReport(year.Period, coa, year.CompanyName),
		documents.GenerateFlowReport(year.Period, coa, year.CompanyName),
		documents.GenerateVATReport(year.Period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateCategoryReport(year.Period, coa, year.CompanyName, year.CompanyAddress,
			"ZESTAWIENIE DZIAŁALNOŚCI NIEODPŁATNEJ",
			"Nieodpłatna",
			types.NewAccountID(accounts.Nieodplatna)),
		documents.GenerateCategoryReport(year.Period, coa, year.CompanyName, year.CompanyAddress,
			"ZESTAWIENIE DZIAŁALNOŚCI ODPŁATNEJ",
			"Odpłatna",
			types.NewAccountID(accounts.Odplatna)),
		documents.GenerateCIT8Report(coa),
	}
	currencies := lo.Keys(bankRecords)
	sort.Slice(currencies, func(i, j int) bool {
		return strings.Compare(string(currencies[i]), string(currencies[j])) < 0
	})
	for _, c := range currencies {
		ci, exists := year.Init.Currencies[c]
		if !exists {
			panic("currency not initialized")
		}
		docs = append(docs, documents.GenerateBankReport(year.Period, year.CompanyName, year.CompanyAddress,
			types.Currencies.Currency(c), ci, bankRecords[c]))
	}
	docs = append(docs, documents.GenerateOverDueReport(year.Period, year.Operations))
	docs = append(docs, opDocs...)

	report := types.Report{
		Currencies: lo.Values(types.Currencies),
		Configs:    make([]string, 0, len(docs)),
		Documents:  make([]string, 0, len(docs)),
	}

	buf := &bytes.Buffer{}
	for _, doc := range docs {
		buf.Reset()
		lo.Must0(doc.Template.Execute(buf, doc.Data))
		report.Documents = append(report.Documents, buf.String())

		buf.Reset()
		lo.Must0(configTmplParsed.Execute(buf, doc.Config))
		report.Configs = append(report.Configs, buf.String())
	}

	return report
}
