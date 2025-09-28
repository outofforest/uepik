package operations

import (
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// CurrencyDiff defines the currency diff.
type CurrencyDiff struct {
	Date time.Time
}

// BankRecords returns bank records for currency diff.
func (cd *CurrencyDiff) BankRecords() []*types.BankRecord {
	return nil
}

// BookRecords returns book records for currency diff.
func (cd *CurrencyDiff) BookRecords(
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) {
	records := []types.EntryRecord{}
	debit := coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Date)
	if debit.NEQ(types.BaseZero) {
		records = append(records,
			types.NewEntryRecord(
				types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.Podatkowe, accounts.Finansowe,
					accounts.UjemneRozniceKursowe),
				types.DebitBalance(debit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
				types.DebitBalance(debit),
			),
		)
	}

	credit := coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Date)
	if credit.NEQ(types.BaseZero) {
		records = append(records,
			types.NewEntryRecord(
				types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.Nieoperacyjne, accounts.Finansowe,
					accounts.DodatnieRozniceKursowe),
				types.CreditBalance(credit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
				types.CreditBalance(credit),
			),
		)
	}

	coa.AddEntry(cd.Date, types.Document{}, types.Contractor{}, "Różnice kursowe", records...)
}
