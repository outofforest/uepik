package operations

import (
	"fmt"
	"strings"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/report/documents"
	"github.com/outofforest/uepik/v2/types"
)

// CurrencyDiff defines the currency diff.
type CurrencyDiff struct {
	Contractor types.Contractor
}

// BankRecords returns bank records for currency diff.
func (cd *CurrencyDiff) BankRecords() []*types.BankRecord {
	return nil
}

// BookRecords returns book records for currency diff.
func (cd *CurrencyDiff) BookRecords(
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	docs := []types.ReportDocument{}
	for _, month := range period.Months() {
		cdDate := month.AddDate(0, 1, 0).Add(-time.Nanosecond)
		cdID := fmt.Sprintf("RK/%d/%d/1", cdDate.Year(), cdDate.Month())
		source := &CurrencyDiffSource{
			Document: types.Document{
				ID:        types.DocumentID(cdID),
				Date:      cdDate,
				SheetName: strings.ReplaceAll(cdID, "/", "."),
			},
			Contractor: cd.Contractor,
		}

		debit := coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe), cdDate)
		credit := coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe), cdDate)

		coa.AddEntry(
			source,
			types.NewEntryRecord(
				types.NewAccountID(accounts.PiK, accounts.Koszty, accounts.Podatkowe, accounts.Finansowe,
					accounts.UjemneRozniceKursowe),
				types.DebitBalance(debit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod),
				types.DebitBalance(debit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.Nieodplatna),
				types.DebitBalance(coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe, accounts.Nieodplatna),
					cdDate)),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.Odplatna),
				types.DebitBalance(coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe, accounts.Odplatna),
					cdDate)),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Finansowe,
					accounts.DodatnieRozniceKursowe),
				types.CreditBalance(credit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.NiewydatkowanyDochod),
				types.CreditBalance(credit),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.Nieodplatna),
				types.CreditBalance(coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe, accounts.Nieodplatna),
					cdDate)),
			),
			types.NewEntryRecord(
				types.NewAccountID(accounts.Odplatna),
				types.CreditBalance(coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe, accounts.Odplatna),
					cdDate)),
			),
		)

		entries := coa.EntriesMonth(types.NewAccountID(accounts.RozniceKursowe), cdDate)
		if len(entries) > 0 {
			docs = append(docs, documents.GenerateCurrencyDiffDocument(source.Document, cd.Contractor, entries))
		}
	}
	return docs
}

// CurrencyDiffSource is the source of currency diff document.
type CurrencyDiffSource struct {
	Document   types.Document
	Contractor types.Contractor
}

// GetDate returns date of currency diff.
func (cds *CurrencyDiffSource) GetDate() time.Time {
	return cds.Document.Date
}

// GetDocument returns document.
func (cds *CurrencyDiffSource) GetDocument() types.Document {
	return cds.Document
}

// GetContractor returns contractor.
func (cds *CurrencyDiffSource) GetContractor() types.Contractor {
	return cds.Contractor
}

// GetNotes returns notes.
func (cds *CurrencyDiffSource) GetNotes() string {
	return "Różnice kursowe"
}
