package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed flow.tmpl.xml
	flowTmpl     string
	flowTemplate = template.Must(template.New("flow").Parse(flowTmpl))
)

// FlowReport is the financial flow report.
type FlowReport struct {
	CompanyName    string
	Income         types.Denom
	CostsTaxed     types.Denom
	ProfitYear     types.Denom
	ProfitPrevious types.Denom
	CostsNotTaxed  types.Denom
	Profit         types.Denom
}

// GenerateFlowReport generates flow report.
func GenerateFlowReport(
	period types.Period,
	coa *types.ChartOfAccounts,
	companyName string,
) types.ReportDocument {
	income := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody))
	costsTaxed := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe))

	return types.ReportDocument{
		Template: flowTemplate,
		Data: &FlowReport{
			CompanyName:    companyName,
			Income:         income,
			CostsTaxed:     costsTaxed,
			ProfitYear:     income.Sub(costsTaxed),
			ProfitPrevious: coa.OpeningBalance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
			CostsNotTaxed:  coa.Debit(types.NewAccountID(accounts.NiewydatkowanyDochod)),
			Profit:         coa.Balance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
		},
		Config: types.SheetConfig{
			Name:       "PF",
			LockedRows: 6,
		},
	}
}
