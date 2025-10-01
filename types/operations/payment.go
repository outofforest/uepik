package operations

import (
	"time"

	"github.com/outofforest/uepik/types"
)

// Payment defines payment operation.
type Payment struct {
	Contractor types.Contractor
	Payment    types.Payment
	Notes      string
}

// GetDate returns date of payment.
func (p *Payment) GetDate() time.Time {
	return p.Payment.Date
}

// GetDocument returns document.
func (p *Payment) GetDocument() types.Document {
	return types.Document{
		ID:   p.Payment.DocumentID,
		Date: p.Payment.Date,
	}
}

// GetContractor returns contractor.
func (p *Payment) GetContractor() types.Contractor {
	return p.Contractor
}

// GetNotes returns notes.
func (p *Payment) GetNotes() string {
	return p.Notes
}

// BankRecords returns bank records for the payment.
func (p *Payment) BankRecords() []*types.BankRecord {
	return []*types.BankRecord{{
		Date:           p.Payment.Date,
		Index:          p.Payment.Index,
		Document:       p.Payment.DocumentID,
		PaidDocument:   p.GetDocument(),
		Contractor:     p.Contractor,
		OriginalAmount: p.Payment.Amount,
	}}
}

// BookRecords returns book records for the payment.
func (p *Payment) BookRecords(
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	return nil
}
