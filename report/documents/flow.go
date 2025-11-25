package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.SheetSource = &FlowReport{}

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

// GetSheet returns sheet to report.
func (r *FlowReport) GetSheet() types.Sheet {
	return types.Sheet{
		Template: flowTemplate,
		Data:     r,
		Config: types.SheetConfig{
			Name:       "PF",
			LockedRows: 6,
		},
	}
}

// NewFlowReport generates flow report.
func NewFlowReport(
	coa *types.ChartOfAccounts,
	companyName string,
) *FlowReport {
	income := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody))
	costsTaxed := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe))

	return &FlowReport{
		CompanyName:    companyName,
		Income:         income,
		CostsTaxed:     costsTaxed,
		ProfitYear:     income.Sub(costsTaxed),
		ProfitPrevious: coa.OpeningBalance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
		CostsNotTaxed:  coa.Debit(types.NewAccountID(accounts.NiewydatkowanyDochod)),
		Profit:         coa.Balance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
	}
}
