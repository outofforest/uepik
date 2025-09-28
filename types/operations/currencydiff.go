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
	debit := coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Date)
	if debit.NEQ(types.BaseZero) {
		coa.AddEntry(
			types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.Podatkowe, accounts.Finansowe,
				accounts.UjemneRozniceKursowe),
			types.NewEntry(cd.Date, types.Document{}, types.Contractor{}, types.DebitBalance(debit),
				"Różnice kursowe"),
		)
		coa.AddEntry(
			types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
			types.NewEntry(cd.Date, types.Document{}, types.Contractor{}, types.DebitBalance(debit), ""),
		)
	}

	credit := coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Date)
	if credit.NEQ(types.BaseZero) {
		coa.AddEntry(
			types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.Nieoperacyjne, accounts.Finansowe,
				accounts.DodatnieRozniceKursowe),
			types.NewEntry(cd.Date, types.Document{}, types.Contractor{}, types.CreditBalance(credit),
				"Różnice kursowe"),
		)
		coa.AddEntry(
			types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
			types.NewEntry(cd.Date, types.Document{}, types.Contractor{}, types.CreditBalance(credit), ""),
		)
	}
}
