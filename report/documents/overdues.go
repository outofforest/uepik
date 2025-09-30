package documents

import (
	_ "embed"
	"sort"
	"text/template"
	"time"

	"github.com/outofforest/uepik/types"
)

var (
	//go:embed overdues.tmpl.xml
	overDuesTmpl     string
	overDuesTemplate = template.Must(template.New("overDues").Parse(overDuesTmpl))
)

// OverDueRecord represents VAT over due.
type OverDueRecord struct {
	DueDate    time.Time
	Document   types.Document
	Contractor types.Contractor
	Amount     types.Denom
}

type overDueSource interface {
	GetDocument() types.Document
	GetContractor() types.Contractor
	GetDues() []types.Due
	GetPayments() []types.Payment
}

// GenerateOverDueReport generates over due report.
func GenerateOverDueReport(
	period types.Period,
	operations []types.Operation,
) types.ReportDocument {
	report := []OverDueRecord{}

	for _, op := range operations {
		overDueSource, ok := op.(overDueSource)
		if !ok {
			continue
		}

		paid := map[types.CurrencySymbol]types.Denom{}
		for _, p := range overDueSource.GetPayments() {
			paidCurrency, exists := paid[p.Amount.Currency]
			if !exists {
				paidCurrency = types.NewDenom(p.Amount.Currency)
			}
			paid[p.Amount.Currency] = paidCurrency.Add(p.Amount)
		}

		dues := overDueSource.GetDues()
		sort.Slice(dues, func(i, j int) bool {
			return dues[i].Date.Before(dues[j].Date)
		})

		for _, d := range dues {
			if !period.Contains(d.Date) {
				break
			}

			p, exists := paid[d.Amount.Currency]
			switch {
			case !exists:
			case p.GT(d.Amount):
				paid[d.Amount.Currency] = p.Sub(d.Amount)
				continue
			case p.EQ(d.Amount):
				delete(paid, d.Amount.Currency)
				continue
			default:
				d.Amount = d.Amount.Sub(p)
				delete(paid, d.Amount.Currency)
			}

			report = append(report, OverDueRecord{
				DueDate:    d.Date,
				Document:   overDueSource.GetDocument(),
				Contractor: overDueSource.GetContractor(),
				Amount:     d.Amount,
			})
		}
	}

	return types.ReportDocument{
		Template: overDuesTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       "Zaległości",
			LockedRows: 1,
		},
	}
}
