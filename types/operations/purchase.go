package operations

import (
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

// Purchase defines the cost of purchased goods or service.
type Purchase struct {
	Date             time.Time
	Document         types.Document
	Contractor       types.Contractor
	Amount           types.Denom
	Payments         []types.Payment
	CostTaxType      types.CostTaxType
	CostCategoryType types.CostCategoryType
	Notes            string
}

// GetDate returns date of purchase.
func (p *Purchase) GetDate() time.Time {
	return p.Date
}

// GetDocument returns document.
func (p *Purchase) GetDocument() types.Document {
	return p.Document
}

// GetContractor returns contractor.
func (p *Purchase) GetContractor() types.Contractor {
	return p.Contractor
}

// GetNotes returns notes.
func (p *Purchase) GetNotes() string {
	return p.Notes
}

// BankRecords returns bank records for the purchase.
func (p *Purchase) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range p.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       payment.DocumentID,
			PaidDocument:   p.Document,
			Contractor:     p.Contractor,
			OriginalAmount: payment.Amount.Neg(),
		})
	}
	return records
}

// BookRecords returns book records for the purchase.
func (p *Purchase) BookRecords(
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	if period.End.Before(p.Date) {
		return nil
	}

	costBase, costRate := rates.ToBase(p.Amount, types.PreviousDay(p.Date))

	coa.AddEntry(p,
		types.NewEntryRecord(
			costTaxTypeToAccountID(p.CostTaxType),
			types.DebitBalance(costBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(costCategoryTypeToAccountPart(p.CostCategoryType)),
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

		coa.AddEntry(types.NewCurrencyDiff(p, costRate, br),
			types.NewEntryRecord(
				types.NewAccountID(accounts.RozniceKursowe, costCategoryTypeToAccountPart(p.CostCategoryType)),
				amount,
			),
		)
	}

	return nil
}

func costTaxTypeToAccountID(costTaxType types.CostTaxType) types.AccountID {
	switch costTaxType {
	case types.CostTaxTypeTaxable:
		return types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe,
			accounts.Operacyjne)
	case types.CostTaxTypeNonTaxable:
		return types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Niepodatkowe,
			accounts.Operacyjne)
	default:
		panic("invalid cost tax type")
	}
}

func costCategoryTypeToAccountPart(costCategoryType types.CostCategoryType) types.AccountIDPart {
	switch costCategoryType {
	case types.CostCategoryTypeFreeOfCharge:
		return accounts.Nieodplatna
	case types.CostCategoryTypePaid:
		return accounts.Odplatna
	default:
		panic("invalid cost category type")
	}
}
