package operations

import (
	"fmt"

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
func (d *Donation) BookRecords(period types.Period, rates types.CurrencyRates) []types.BookRecord {
	if !period.Contains(d.Payment.Date) {
		return nil
	}

	result := []types.BookRecord{}

	incomeBase, incomeRate := rates.ToBase(d.Payment.Amount, types.PreviousDay(d.Payment.Date))

	result = append(result, types.BookRecord{
		Date:            d.Payment.Date,
		Document:        d.Document,
		Contractor:      d.Contractor,
		IncomeDonations: incomeBase,
		IncomeTrading:   types.BaseZero,
		IncomeOthers:    types.BaseZero,
		CostTaxed:       types.BaseZero,
		CostNotTaxed:    types.BaseZero,
		Notes:           fmt.Sprintf("kwota: %s, kurs: %s", d.Payment.Amount, incomeRate),
	})

	return result
}

// VATRecords returns VAT records for the donation.
func (d *Donation) VATRecords(period types.Period, rates types.CurrencyRates) []types.VATRecord {
	return nil
}
