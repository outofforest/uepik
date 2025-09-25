package generators

import (
	_ "embed"
	"os"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/types"
)

//go:embed uepik.tmpl.fods
var tmpl string
var tmplParsed = template.Must(template.New("").Parse(tmpl))

// Save saves the report.
func Save(year types.FiscalYear) {
	data := types.MonthReport{
		Year:  2025,
		Month: "styczeń",
		Bank:  Bank(year),
		Book:  Book(year),
		FlowMonth: types.FlowRecord{
			Income:                types.BaseZero,
			CostsTaxed:            types.BaseZero,
			ProfitCurrent:         types.BaseZero,
			CostsNotTaxedCurrent:  types.BaseZero,
			ProfitPrevious:        types.BaseZero,
			CostsNotTaxedPrevious: types.BaseZero,
			ProfitTotal:           types.BaseZero,
		},
		FlowIncremental: types.FlowRecord{
			Income:                types.BaseZero,
			CostsTaxed:            types.BaseZero,
			ProfitCurrent:         types.BaseZero,
			CostsNotTaxedCurrent:  types.BaseZero,
			ProfitPrevious:        types.BaseZero,
			CostsNotTaxedPrevious: types.BaseZero,
			ProfitTotal:           types.BaseZero,
		},
		VAT: VAT(year),
	}

	f := lo.Must(os.OpenFile("uepik-"+time.Now().Format(time.DateOnly)+".fods", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600))
	defer f.Close()
	lo.Must0(tmplParsed.Execute(f, data))
}
