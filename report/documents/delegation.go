package documents

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/types"
)

var (
	//go:embed delegation.tmpl.xml
	delegationTmpl     string
	delegationTemplate = template.Must(template.New("delegation").Funcs(template.FuncMap{
		"date":        date,
		"dateAndTime": dateAndTime,
		"duration":    durationString,
	}).Parse(delegationTmpl))
)

// DelegationDocument represents delegation documents.
type DelegationDocument struct {
	Document   types.Document
	Company    types.Contractor
	Person     types.Contractor
	Start, End time.Time
	Country    types.DelegationCountry
	Currency   types.CurrencySymbol
	Costs      []DelegationCost
	Notes      string
	Summary    DelegationSummary
}

// DelegationCost represents delegation cost.
type DelegationCost struct {
	Document       types.DocumentID
	OriginalAmount types.Denom
	Amount         types.Denom
	Notes          string
}

// NewDelegationSummary returns new summary of delegation.
func NewDelegationSummary(currency types.CurrencySymbol) DelegationSummary {
	return DelegationSummary{
		Amount: types.NewDenom(currency),
	}
}

// DelegationSummary is the summary of delegation document.
type DelegationSummary struct {
	Amount types.Denom
}

// AddRecord adds record to the summary.
func (ds DelegationSummary) AddRecord(r DelegationCost) DelegationSummary {
	ds.Amount = ds.Amount.Add(r.Amount)
	return ds
}

// GenerateDelegationDocument generates delegation document.
func GenerateDelegationDocument(
	document types.Document,
	company types.Contractor,
	person types.Contractor,
	start, end time.Time,
	country types.DelegationCountry,
	currency types.CurrencySymbol,
	costs []types.DelegationCost,
	notes string,
	rates types.CurrencyRates,
) (types.ReportDocument, types.Denom) {
	report := &DelegationDocument{
		Document: document,
		Company:  company,
		Person:   person,
		Start:    start,
		End:      end,
		Country:  country,
		Currency: currency,
		Costs:    make([]DelegationCost, 0, len(costs)),
		Notes:    notes,
		Summary:  NewDelegationSummary(currency),
	}

	for _, c := range costs {
		originalAmount := c.GetAmount(start, end, country)
		amount := originalAmount
		if amount.Currency != currency {
			amountBase, _ := rates.ToBase(amount, types.PreviousDay(document.Date))
			amount, _ = rates.FromBase(amountBase, currency, types.PreviousDay(document.Date))
		}

		r := DelegationCost{
			Document:       c.GetDocument(),
			OriginalAmount: originalAmount,
			Amount:         amount,
			Notes:          c.GetNotes(),
		}
		report.Costs = append(report.Costs, r)
		report.Summary = report.Summary.AddRecord(r)
	}

	return types.ReportDocument{
		Date:     document.Date,
		Template: delegationTemplate,
		Data:     report,
		Config: types.SheetConfig{
			Name:       document.SheetName,
			LockedRows: 12,
		},
	}, report.Summary.Amount
}

func durationString(start, end time.Time) string {
	days, hours := types.DelegationDuration(start, end)

	return fmt.Sprintf("%d dni, %s godzin", days, strings.ReplaceAll(fmt.Sprintf("%.1f", hours), ".", ","))
}
