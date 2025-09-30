package report

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/report/documents"
	"github.com/outofforest/uepik/types"
	"github.com/outofforest/uepik/types/operations"
)

var coaAccounts = []*types.Account{
	types.NewAccount(
		accounts.PiK, types.Liabilities, types.AllValid(),
		types.NewAccount(
			accounts.Przychody, types.Incomes, types.AllValid(),
			types.NewAccount(
				accounts.Finansowe, types.Incomes, types.AllValid(),
				types.NewAccount(accounts.DodatnieRozniceKursowe, types.Incomes,
					types.ValidSources(&operations.CurrencyDiff{})),
			),
			types.NewAccount(
				accounts.Operacyjne, types.Incomes, types.AllValid(),
				types.NewAccount(accounts.ZNieodplatnejDPP, types.Incomes, types.ValidSources(&operations.Donation{})),
				types.NewAccount(accounts.ZOdplatnejDPP, types.Incomes, types.ValidSources(&operations.Sell{})),
			),
		),
		types.NewAccount(
			accounts.Koszty, types.Costs, types.AllValid(),
			types.NewAccount(
				accounts.Podatkowe, types.Costs, types.AllValid(),
				types.NewAccount(
					accounts.Finansowe, types.Costs, types.AllValid(),
					types.NewAccount(accounts.UjemneRozniceKursowe, types.Costs,
						types.ValidSources(&operations.CurrencyDiff{})),
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
		accounts.NiewydatkowanyDochod, types.Liabilities, types.AllValid(),
		types.NewAccount(accounts.WTrakcieRoku, types.Liabilities, types.ValidSources(
			&operations.CurrencyDiff{},
			&operations.Donation{},
			&operations.Purchase{},
			&operations.Sell{},
		)),
		types.NewAccount(accounts.ZLatUbieglych, types.Liabilities, types.ValidSources(&operations.Purchase{})),
	),
	types.NewAccount(accounts.RozniceKursowe, types.Liabilities, types.ValidSources(&types.CurrencyDiff{})),
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
	coa.OpenAccount(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.ZLatUbieglych),
		types.CreditBalance(year.Init.UnspentProfit))

	company := types.Contractor{
		Name:    year.CompanyName,
		Address: year.CompanyAddress,
		TaxID:   year.CompanyTaxID,
	}
	//nolint:lll
	for date := year.Period.Start.AddDate(0, 1, 0).Add(-time.Nanosecond); year.Period.Contains(date); date = date.AddDate(0, 1, 0) {
		id := fmt.Sprintf("RK/%d/%d/1", date.Year(), date.Month())
		year.Operations = append(year.Operations, &operations.CurrencyDiff{
			Document: types.Document{
				ID:        types.DocumentID(id),
				Date:      date,
				SheetName: strings.ReplaceAll(id, "/", "."),
			},
			Contractor: company,
		})
	}

	bankRecords, opBankRecords := year.BankReports(currencyRates, years)
	year.BookRecords(coa, currencyRates, opBankRecords)

	docs := []types.ReportDocument{
		documents.GenerateBookReport(year.Period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateFlowReport(year.Period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateVATReport(year.Period, coa, year.CompanyName, year.CompanyAddress),
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
	for _, op := range year.Operations {
		docs = append(docs, op.Documents(coa)...)
	}

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
