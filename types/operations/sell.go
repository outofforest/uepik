package operations

import (
	"fmt"

	"github.com/outofforest/uepik/types"
)

// Sell defines the income coming from goods or service sell.
type Sell struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment
	CIT        types.CIT
	VAT        types.VAT

	paymentBankRecord *types.BankRecord
}

// BankRecords returns bank records for the sell.
func (s *Sell) BankRecords(period types.Period) []*types.BankRecord {
	if !period.Contains(s.Payment.Date) {
		return nil
	}

	s.paymentBankRecord = &types.BankRecord{
		Date:           s.Payment.Date,
		Index:          s.Payment.Index,
		OriginalAmount: s.Payment.Amount,
	}

	return []*types.BankRecord{s.paymentBankRecord}
}

// BookRecords returns book records for the sell.
func (s *Sell) BookRecords(period types.Period, rates types.CurrencyRates) []types.BookRecord {
	if !period.Contains(s.CIT.Date) {
		return nil
	}

	result := []types.BookRecord{}

	incomeBase, incomeRate := rates.ToBase(s.Payment.Amount, types.PreviousDay(s.CIT.Date))

	result = append(result, types.BookRecord{
		Document:        s.Document,
		Contractor:      s.Contractor,
		IncomeDonations: types.BaseZero,
		IncomeTrading:   incomeBase,
		IncomeOthers:    types.BaseZero,
		CostTaxed:       types.BaseZero,
		CostNotTaxed:    types.BaseZero,
		Notes:           fmt.Sprintf("kwota: %s, kurs: %s", s.Payment.Amount, incomeRate),
	})

	if s.paymentBankRecord != nil && s.paymentBankRecord.BaseAmount.NEQ(incomeBase) {
		rateDiff := types.BookRecord{
			IncomeDonations: types.BaseZero,
			IncomeTrading:   types.BaseZero,
			IncomeOthers:    types.BaseZero,
			CostTaxed:       types.BaseZero,
			CostNotTaxed:    types.BaseZero,
			Notes: fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s",
				s.Payment.Amount, incomeRate, s.paymentBankRecord.Rate),
		}
		if incomeBase.GT(s.paymentBankRecord.BaseAmount) {
			rateDiff.CostTaxed = incomeBase.Sub(s.paymentBankRecord.BaseAmount)
		} else {
			rateDiff.IncomeOthers = s.paymentBankRecord.BaseAmount.Sub(incomeBase)
		}

		result = append(result, rateDiff)
	}

	return result
}
