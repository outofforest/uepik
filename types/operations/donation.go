package operations

import (
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Donation defines the income coming from donation.
type Donation struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment
}

// GetDate returns date of donation.
func (d *Donation) GetDate() time.Time {
	return d.Payment.Date
}

// GetDocument returns document.
func (d *Donation) GetDocument() types.Document {
	return d.Document
}

// GetContractor returns contractor.
func (d *Donation) GetContractor() types.Contractor {
	return d.Contractor
}

// GetNotes returns notes.
func (d *Donation) GetNotes() string {
	return "Darowizna"
}

// BankRecords returns bank records for the donation.
func (d *Donation) BankRecords() []*types.BankRecord {
	return []*types.BankRecord{{
		Date:           d.Payment.Date,
		Index:          d.Payment.Index,
		Document:       d.Document,
		Contractor:     d.Contractor,
		OriginalAmount: d.Payment.Amount,
	}}
}

// BookRecords returns book records for the donation.
func (d *Donation) BookRecords(coa *types.ChartOfAccounts, bankRecords []*types.BankRecord, rates types.CurrencyRates) {
	incomeBase, _ := rates.ToBase(d.Payment.Amount, types.PreviousDay(d.Payment.Date))

	coa.AddEntry(d,
		types.NewEntryRecord(
			types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.Operacyjne, accounts.ZNieodplatnejDPP,
				accounts.Darowizny),
			types.CreditBalance(incomeBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
			types.CreditBalance(incomeBase),
		),
	)
}

// Documents generate documents for operation.
func (d *Donation) Documents(coa *types.ChartOfAccounts) []types.ReportDocument {
	return nil
}
