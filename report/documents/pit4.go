package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.SheetSource = &CIT8Report{}

	//go:embed pit4.tmpl.xml
	pit4Tmpl     string
	pit4Template = template.Must(template.New("pit4").Funcs(template.FuncMap{
		"date": date,
	}).Parse(pit4Tmpl))
)

// PIT4Report is the PIT-4 report.
type PIT4Report struct {
	Months []PIT4Month
}

// GetSheet returns sheet to report.
func (r *PIT4Report) GetSheet() types.Sheet {
	return types.Sheet{
		Data:     r,
		Template: pit4Template,
		Config: types.SheetConfig{
			Name:       "PIT-4",
			LockedRows: 0,
		},
	}
}

// NewPIT4Report generates PIT-4 report.
func NewPIT4Report(
	period types.Period,
	coa *types.ChartOfAccounts,
) *PIT4Report {
	report := &PIT4Report{}
	for _, month := range period.Months() {
		report.Months = append(report.Months, PIT4Month{
			Date:   month,
			Amount: coa.BalanceMonth(types.NewAccountID(accounts.ZaliczkiPIT), month),
		})
	}
	return report
}

// PIT4Month stores month data for PIT-4 report.
type PIT4Month struct {
	Date   time.Time
	Amount types.Denom
}
