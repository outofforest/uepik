package operations

import (
	"github.com/outofforest/uepik/v2/types"
)

var _ types.Operation = &Exchange{}

// Exchange defines the currency exchange.
type Exchange struct {
	Document             types.Document
	SrcAmount, DstAmount types.Denom
	SrcIndex, DstIndex   uint64
}

// BankRecords returns bank records for the exchange.
func (e *Exchange) BankRecords() []*types.BankRecord {
	return []*types.BankRecord{
		{
			Date:           e.Document.Date,
			Index:          e.SrcIndex,
			Document:       e.Document.ID,
			OriginalAmount: e.SrcAmount.Neg(),
		},
		{
			Date:           e.Document.Date,
			Index:          e.DstIndex,
			Document:       e.Document.ID,
			OriginalAmount: e.DstAmount,
		},
	}
}

// BookRecords returns book records for the exchange.
func (e *Exchange) BookRecords(
	company types.Contractor,
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.SheetSource {
	return nil
}
