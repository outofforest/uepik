package operations

import (
	"fmt"
	"strings"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/report/documents"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.Operation = &CurrencyDiff{}
)

// CurrencyDiff defines the currency diff.
type CurrencyDiff struct{}

// BankRecords returns bank records for currency diff.
func (cd *CurrencyDiff) BankRecords() []*types.BankRecord {
	return nil
}

// BookRecords returns book records for currency diff.
func (cd *CurrencyDiff) BookRecords(
	company types.Contractor,
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.SheetSource {
	docs := []types.SheetSource{}
	for _, month := range period.Months() {
		cdDate := month.AddDate(0, 1, 0).Add(-time.Nanosecond)
		cdID := fmt.Sprintf("RK/%d/%d/1", cdDate.Year(), cdDate.Month())

		doc := documents.NewCurrencyDiffDocument(coa, types.Document{
			ID:        types.DocumentID(cdID),
			Date:      cdDate,
			SheetName: strings.ReplaceAll(cdID, "/", "."),
		}, company)
		if doc == nil {
			continue
		}

		docs = append(docs, doc)

		debit := coa.DebitMonth(types.NewAccountID(accounts.RozniceKursowe), cdDate)
		credit := coa.CreditMonth(types.NewAccountID(accounts.RozniceKursowe), cdDate)

		coa.AddEntry(
			doc,
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
	}
	return docs
}
