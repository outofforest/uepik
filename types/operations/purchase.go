package operations

import (
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.Operation       = &Purchase{}
	_ types.EntryDataSource = &Purchase{}
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
func (d *Purchase) GetDate() time.Time {
	return d.Date
}

// GetDocument returns document.
func (d *Purchase) GetDocument() types.Document {
	return d.Document
}

// GetContractor returns contractor.
func (d *Purchase) GetContractor() types.Contractor {
	return d.Contractor
}

// GetNotes returns notes.
func (d *Purchase) GetNotes() string {
	return d.Notes
}

// BankRecords returns bank records for the purchase.
func (d *Purchase) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range d.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       payment.DocumentID,
			PaidDocument:   d.Document,
			Contractor:     d.Contractor,
			OriginalAmount: payment.Amount.Neg(),
		})
	}
	return records
}

// BookRecords returns book records for the purchase.
func (d *Purchase) BookRecords(
	company types.Contractor,
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	if period.End.Before(d.Date) {
		return nil
	}

	costBase, costRate := rates.ToBase(d.Amount, types.PreviousDay(d.Date))

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

func costCategoryTypeToAccountID(costCategoryType types.CostCategoryType) types.AccountIDPart {
	switch costCategoryType {
	case types.CostCategoryTypeFreeOfCharge:
		return accounts.Nieodplatna
	case types.CostCategoryTypePaid:
		return accounts.Odplatna
	default:
		panic("invalid cost category type")
	}
}
