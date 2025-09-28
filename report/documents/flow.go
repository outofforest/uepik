package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

var (
	//go:embed flow.tmpl.xml
	flowTmpl     string
	flowTemplate = template.Must(template.New("flow").Parse(flowTmpl))
)

// FlowReport is the financial flow report.
type FlowReport struct {
	CompanyName    string
	CompanyAddress string
	Records        []FlowRecord
}

// FlowRecord represents month in the flow report.
type FlowRecord struct {
	Year  uint64
	Month string

	MonthIncome                types.Denom
	MonthCostsTaxed            types.Denom
	MonthProfit                types.Denom
	MonthCostsNotTaxedCurrent  types.Denom
	MonthCostsNotTaxedPrevious types.Denom

	TotalIncome                types.Denom
	TotalCostsTaxed            types.Denom
	TotalProfitYear            types.Denom
	TotalCostsNotTaxedCurrent  types.Denom
	TotalProfitPrevious        types.Denom
	TotalCostsNotTaxedPrevious types.Denom
	TotalProfit                types.Denom
}

// GenerateFlowReport generates flow report.
func GenerateFlowReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName, companyAddress string,
) types.ReportDocument {
	report := &FlowReport{
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
	}

	unspentProfit := coa.OpeningBalance(types.NewAccountID(accounts.NiewydatkowanyDochod))

	for month := period.Start; period.Contains(month); month = month.AddDate(0, 1, 0) {
		yearNumber := uint64(month.Year())
		monthName := monthName(month.Month())

		monthIncome := coa.BalanceMonth(types.NewAccountID(accounts.CIT, accounts.Przychody), month)
		monthCostsTaxed := coa.BalanceMonth(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.Podatkowe),
			month)
		yearIncome := coa.BalanceIncremental(types.NewAccountID(accounts.CIT, accounts.Przychody), month)
		yearCostsTaxed := coa.BalanceIncremental(types.NewAccountID(accounts.CIT, accounts.Koszty,
			accounts.Podatkowe), month)

		report.Records = append(report.Records, FlowRecord{
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
	}

	return types.ReportDocument{
		Template: flowTemplate,
		Data:     report,
	}
}
