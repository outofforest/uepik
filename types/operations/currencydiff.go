package operations

import (
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/report/documents"
	"github.com/outofforest/uepik/types"
)

// CurrencyDiff defines the currency diff.
type CurrencyDiff struct {
	Document   types.Document
	Contractor types.Contractor
}

// GetDate returns date of currency diff.
func (cd *CurrencyDiff) GetDate() time.Time {
	return cd.Document.Date
}

// GetDocument returns document.
func (cd *CurrencyDiff) GetDocument() types.Document {
	return cd.Document
}

// GetContractor returns contractor.
func (cd *CurrencyDiff) GetContractor() types.Contractor {
	return cd.Contractor
}

// GetNotes returns notes.
func (cd *CurrencyDiff) GetNotes() string {
	return "Różnice kursowe"
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
	debit := coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Document.Date)
	if debit.NEQ(types.BaseZero) {
		records = append(records,
			types.NewEntryRecord(
				types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe, accounts.Finansowe,
					accounts.UjemneRozniceKursowe),
				types.DebitBalance(debit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
				types.DebitBalance(debit),
			),
		)
	}

	credit := coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Document.Date)
	if credit.NEQ(types.BaseZero) {
		records = append(records,
			types.NewEntryRecord(
				types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Finansowe,
					accounts.DodatnieRozniceKursowe),
				types.CreditBalance(credit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
				types.CreditBalance(credit),
			),
		)
	}

	coa.AddEntry(cd, records...)
}

// Documents generate documents for operation.
func (cd *CurrencyDiff) Documents(coa *types.ChartOfAccounts) []types.ReportDocument {
	entries := coa.EntriesMonth(types.NewAccountID(accounts.RozniceKursowe), cd.Document.Date)
	if len(entries) == 0 {
		return nil
	}

	return []types.ReportDocument{documents.GenerateCurrencyDiffDocument(cd.Document, cd.Contractor, entries)}
}
