package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

var (
	//go:embed cit8.tmpl.xml
	cit8Tmpl     string
	cit8Template = template.Must(template.New("cit8").Parse(cit8Tmpl))
)

// CIT8Report is the CIT-8 report.
type CIT8Report struct {
	IncomesFinancial          types.Denom
	IncomesOthers             types.Denom
	CostsFinancial            types.Denom
	CostsOthers               types.Denom
	NonTaxableProfitFinancial types.Denom
	NonTaxableProfitOthers    types.Denom
	UnspentProfit             types.Denom
	ReceivedDonations         types.Denom
}

// GenerateCIT8Report generates CIT-8 report.
func GenerateCIT8Report(coa *types.ChartOfAccounts) types.ReportDocument {
	incomesFinancial := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Finansowe))
	incomesOthers := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Operacyjne))
	costsFinancial := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe,
		accounts.Finansowe))
	costsOthers := coa.Balance(types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe,
		accounts.Operacyjne))
	nonTaxableProfitFinancial := incomesFinancial.Sub(costsFinancial)
	if nonTaxableProfitFinancial.LT(types.BaseZero) {
		nonTaxableProfitFinancial = types.BaseZero
	}
	nonTaxableProfitOthers := incomesOthers.Sub(costsOthers)
	if nonTaxableProfitOthers.LT(types.BaseZero) {
		nonTaxableProfitOthers = types.BaseZero
	}
	return types.ReportDocument{
		Data: &CIT8Report{
			IncomesFinancial:          incomesFinancial,
			IncomesOthers:             incomesOthers,
			CostsFinancial:            costsFinancial,
			CostsOthers:               costsOthers,
			NonTaxableProfitFinancial: nonTaxableProfitFinancial,
			NonTaxableProfitOthers:    nonTaxableProfitOthers,
			UnspentProfit:             coa.Balance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
			ReceivedDonations: coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody,
				accounts.Operacyjne, accounts.ZNieodplatnejDPP)),
		},
		Template: cit8Template,
		Config: types.SheetConfig{
			Name:       "CIT-8",
			LockedRows: 0,
		},
	}
}
