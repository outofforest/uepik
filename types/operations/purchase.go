package operations

import (
	"github.com/outofforest/uepik/types"
)

// Purchase defines the cost coming from goods or service purchase.
type Purchase struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment
	CIT        types.CIT
	VAT        types.VAT
}

// BankRecords returns bank records for the purchase.
func (p Purchase) BankRecords(period types.Period) []*types.BankRecord {
	return nil
}

// BookRecords returns book records for the purchase.
func (p Purchase) BookRecords(period types.Period, rates types.CurrencyRates) []types.BookRecord {
	return nil
}
