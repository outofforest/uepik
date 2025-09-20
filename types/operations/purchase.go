package operations

import (
	"fmt"

	"github.com/outofforest/uepik/types"
)

// Purchase defines the cost of purchased goods or service.
type Purchase struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment
	CIT        types.CIT
	VAT        types.VAT

	paymentBankRecord *types.BankRecord
}

// BankRecords returns bank records for the purchase.
func (p *Purchase) BankRecords(period types.Period) []*types.BankRecord {
	if !period.Contains(p.Payment.Date) {
		return nil
	}

	p.paymentBankRecord = &types.BankRecord{
		Date:           p.Payment.Date,
		Index:          p.Payment.Index,
		Document:       p.Document,
		Contractor:     p.Contractor,
		OriginalAmount: p.Payment.Amount.Neg(),
	}

	return []*types.BankRecord{p.paymentBankRecord}
}

// BookRecords returns book records for the purchase.
func (p *Purchase) BookRecords(period types.Period, rates types.CurrencyRates) []types.BookRecord {
	if !period.Contains(p.CIT.Date) {
		return nil
	}

	result := []types.BookRecord{}

	costBase, costRate := rates.ToBase(p.Payment.Amount, types.PreviousDay(p.CIT.Date))

	result = append(result, types.BookRecord{
		Date:            p.CIT.Date,
		Document:        p.Document,
		Contractor:      p.Contractor,
		IncomeDonations: types.BaseZero,
		IncomeTrading:   types.BaseZero,
		IncomeOthers:    types.BaseZero,
		CostTaxed:       costBase,
		CostNotTaxed:    types.BaseZero,
		Notes:           fmt.Sprintf("kwota: %s, kurs: %s", p.Payment.Amount, costRate),
	})

	if p.paymentBankRecord != nil && p.paymentBankRecord.BaseAmount.NEQ(costBase.Neg()) {
		rateDiff := types.BookRecord{
			Date:            p.CIT.Date,
			IncomeDonations: types.BaseZero,
			IncomeTrading:   types.BaseZero,
			IncomeOthers:    types.BaseZero,
			CostTaxed:       types.BaseZero,
			CostNotTaxed:    types.BaseZero,
			Notes: fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s",
				p.Payment.Amount, costRate, p.paymentBankRecord.Rate),
		}
		paymentBase := p.paymentBankRecord.BaseAmount.Neg()
		if costBase.GT(paymentBase) {
			rateDiff.IncomeOthers = costBase.Sub(paymentBase)
		} else {
			rateDiff.CostTaxed = paymentBase.Sub(costBase)
		}

		result = append(result, rateDiff)
	}

	return result
}

// VATRecords returns VAT records for the purchase.
func (p *Purchase) VATRecords(period types.Period, rates types.CurrencyRates) []types.VATRecord {
	return nil
}
