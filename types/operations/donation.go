package operations

import (
	"fmt"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Donation defines the income coming from donation.
type Donation struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment

	paymentBankRecord *types.BankRecord
}

// BankRecords returns bank records for the donation.
func (d *Donation) BankRecords(period types.Period) []*types.BankRecord {
	if !period.Contains(d.Payment.Date) {
		return nil
	}

	d.paymentBankRecord = &types.BankRecord{
		Date:           d.Payment.Date,
		Index:          d.Payment.Index,
		Document:       d.Document,
		Contractor:     d.Contractor,
		OriginalAmount: d.Payment.Amount,
	}

	return []*types.BankRecord{d.paymentBankRecord}
}

// BookRecords returns book records for the donation.
func (d *Donation) BookRecords(coa *types.ChartOfAccounts, rates types.CurrencyRates) {
	incomeBase, incomeRate := rates.ToBase(d.Payment.Amount, types.PreviousDay(d.Payment.Date))

	coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyOperacyjne,
		accounts.PrzychodyZNieodplatnejDPP, accounts.DarowiznyOtrzymane),
		types.NewEntry(d.Payment.Date, 0, d.Document, d.Contractor, incomeBase,
			fmt.Sprintf("kwota: %s, kurs: %s", d.Payment.Amount, incomeRate)))
}
