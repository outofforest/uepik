package report

import (
	"bytes"
	_ "embed"
	"os"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/report/documents"
	"github.com/outofforest/uepik/types"
)

//go:embed report.tmpl.fods
var tmpl string
var tmplParsed = template.Must(template.New("").Parse(tmpl))

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

	bankRecords, opBankRecords := year.BankReports(currencyRates, years)
	year.BookRecords(currencyRates, opBankRecords)

	docs := []types.ReportDocument{
		documents.GenerateBookReport(period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateFlowReport(period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateVATReport(period, coa, year.CompanyName, year.CompanyAddress),
		documents.GenerateBankReport(period, coa, year.CompanyName, year.CompanyAddress, year.Init.Currencies,
			bankRecords),
	}
	for _, op := range year.Operations {
		docs = append(docs, op.Documents(coa)...)
	}

	report := types.Report{
		Currencies: lo.Values(types.Currencies),
		Documents:  make([]string, 0, len(docs)),
	}

	buf := &bytes.Buffer{}
	for _, doc := range docs {
		buf.Reset()
		lo.Must0(doc.Template.Execute(buf, doc.Data))
		report.Documents = append(report.Documents, buf.String())
	}

	return report
}
