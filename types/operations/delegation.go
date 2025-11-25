package operations

import (
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/report/documents"
	"github.com/outofforest/uepik/v2/types"
)

// Delegation represents delegation.
type Delegation struct {
	Document         types.Document
	Person           types.Contractor
	Start            time.Time
	End              time.Time
	Country          types.DelegationCountry
	Currency         types.CurrencySymbol
	Payments         []types.Payment
	CostTaxType      types.CostTaxType
	CostCategoryType types.CostCategoryType
	Notes            string
	Costs            []types.DelegationCost
}

// GetDate returns date of delegation.
func (d *Delegation) GetDate() time.Time {
	return d.Document.Date
}

// GetDocument returns document.
func (d *Delegation) GetDocument() types.Document {
	return d.Document
}

// GetContractor returns contractor.
func (d *Delegation) GetContractor() types.Contractor {
	return d.Person
}

// GetNotes returns notes.
func (d *Delegation) GetNotes() string {
	return d.Notes
}

// BankRecords returns bank records for the delegation.
func (d *Delegation) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range d.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       payment.DocumentID,
			PaidDocument:   d.Document,
			Contractor:     d.Person,
			OriginalAmount: payment.Amount.Neg(),
		})
	}
	return records
}

// BookRecords returns book records for the delegation.
func (d *Delegation) BookRecords(
	company types.Contractor,
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	if period.End.Before(d.Document.Date) {
		return nil
	}

	doc, amount := documents.GenerateDelegationDocument(d.Document, company, d.Person, d.Start, d.End, d.Country,
		d.Currency, d.Costs, d.Notes, rates)

	costBase, costRate := rates.ToBase(amount, types.PreviousDay(d.Document.Date))

	coa.AddEntry(d,
		types.NewEntryRecord(
			costTaxTypeToAccountID(d.CostTaxType),
			types.DebitBalance(costBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(costCategoryTypeToAccountID(d.CostCategoryType)),
			types.DebitBalance(costBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(accounts.NiewydatkowanyDochod),
			types.DebitBalance(costBase),
		),
	)

	for _, br := range bankRecords {
		if br.Rate.EQ(costRate) {
			continue
		}

		paymentOriginal := br.OriginalAmount.Neg()
		var amount types.AccountBalance
		if costRate.GT(br.Rate) {
			amount = types.CreditBalance(paymentOriginal.ToBase(costRate.Sub(br.Rate)))
		} else {
			amount = types.DebitBalance(paymentOriginal.ToBase(br.Rate.Sub(costRate)))
		}

		coa.AddEntry(types.NewCurrencyDiff(d, costRate, br),
			types.NewEntryRecord(
				types.NewAccountID(accounts.RozniceKursowe, costCategoryTypeToAccountID(d.CostCategoryType)),
				amount,
			),
		)
	}

	return []types.ReportDocument{doc}
}
