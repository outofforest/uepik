package operations

import (
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.Operation       = &ContractResult{}
	_ types.EntryDataSource = &ContractResult{}
)

const (
	taxFreePercent = 0.5
	taxPercent     = 0.12
)

// ContractResult represents contract of result.
type ContractResult struct {
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
func (d *ContractResult) GetDate() time.Time {
	return d.Date
}

// GetDocument returns document.
func (d *ContractResult) GetDocument() types.Document {
	return d.Document
}

// GetContractor returns contractor.
func (d *ContractResult) GetContractor() types.Contractor {
	return d.Contractor
}

// GetNotes returns notes.
func (d *ContractResult) GetNotes() string {
	return d.Notes
}

// BankRecords returns bank records for the purchase.
func (d *ContractResult) BankRecords() []*types.BankRecord {
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
func (d *ContractResult) BookRecords(
	company types.Contractor,
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.SheetSource {
	if period.End.Before(d.Date) {
		return nil
	}

	costBase, costRate := rates.ToBase(d.Amount, types.PreviousDay(d.Date))
	taxAmount, _ := rates.ToBase(contractResultTax(d.Amount), types.PreviousDay(d.Date))

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
		types.NewEntryRecord(
			types.NewAccountID(accounts.ZaliczkiPIT),
			types.DebitBalance(taxAmount),
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

func contractResultTax(amount types.Denom) types.Denom {
	return amount.Sub(amount.MulFloat64(taxFreePercent)).MulFloat64(taxPercent)
}
