package documents

import (
	_ "embed"
	"text/template"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.SheetSource = &CIT8Report{}

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

// GetSheet returns sheet to report.
func (r *CIT8Report) GetSheet() types.Sheet {
	return types.Sheet{
		Data:     r,
		Template: cit8Template,
		Config: types.SheetConfig{
			Name:       "CIT-8",
			LockedRows: 0,
		},
	}
}

// NewCIT8Report generates CIT-8 report.
func NewCIT8Report(coa *types.ChartOfAccounts) *CIT8Report {
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
	return &CIT8Report{
		IncomesFinancial:          incomesFinancial,
		IncomesOthers:             incomesOthers,
		CostsFinancial:            costsFinancial,
		CostsOthers:               costsOthers,
		NonTaxableProfitFinancial: nonTaxableProfitFinancial,
		NonTaxableProfitOthers:    nonTaxableProfitOthers,
		UnspentProfit:             coa.Balance(types.NewAccountID(accounts.NiewydatkowanyDochod)),
		ReceivedDonations: coa.Balance(types.NewAccountID(accounts.PiK, accounts.Przychody,
			accounts.Operacyjne, accounts.Nieodplatna)),
	}
}
