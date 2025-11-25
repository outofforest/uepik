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
	_ types.SheetSource = &DelegationDocument{}

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

// GetSheet returns sheet to report.
func (d *DelegationDocument) GetSheet() types.Sheet {
	return types.Sheet{
		Date:     d.Document.Date,
		Template: delegationTemplate,
		Data:     d,
		Config: types.SheetConfig{
			Name:       d.Document.SheetName,
			LockedRows: 12,
		},
	}
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

// NewDelegationDocument generates delegation document.
func NewDelegationDocument(
	document types.Document,
	company types.Contractor,
	person types.Contractor,
	start, end time.Time,
	country types.DelegationCountry,
	currency types.CurrencySymbol,
	costs []types.DelegationCost,
	notes string,
	rates types.CurrencyRates,
) *DelegationDocument {
	doc := &DelegationDocument{
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
		doc.Costs = append(doc.Costs, r)
		doc.Summary = doc.Summary.AddRecord(r)
	}

	return doc
}

func durationString(start, end time.Time) string {
	days, hours := types.DelegationDuration(start, end)

	return fmt.Sprintf("%d dni, %s godzin", days, strings.ReplaceAll(fmt.Sprintf("%.1f", hours), ".", ","))
}
