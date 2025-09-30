package operations

import (
	"fmt"
	"strings"
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/report/documents"
	"github.com/outofforest/uepik/types"
)

// UnrecordedSell defines the unrecorded sell operation.
type UnrecordedSell struct {
	Contractor types.Contractor
}

// BankRecords returns bank records for currency diff.
func (us *UnrecordedSell) BankRecords() []*types.BankRecord {
	return nil
}

// BookRecords returns book records for currency diff.
func (us *UnrecordedSell) BookRecords(
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	docs := []types.ReportDocument{}
	for date := period.Start; period.Contains(date); date = date.AddDate(0, 1, 0) {
		unrecordedEntries := coa.EntriesMonth(types.NewAccountID(accounts.SprzedazNieewidencjonowana), date)
		var docIndex uint64
		for len(unrecordedEntries) > 0 {
			docIndex++
			entries := findDayRecords(&unrecordedEntries)
			usDate := entries[0].GetDate()
			usID := fmt.Sprintf("DW/%d/%d/%d", usDate.Year(), usDate.Month(), docIndex)

			source := &UnrecordedSellSource{
				Document: types.Document{
					ID:        types.DocumentID(usID),
					Date:      usDate,
					SheetName: strings.ReplaceAll(usID, "/", "."),
				},
				Contractor: us.Contractor,
			}

			sum := types.BaseZero
			for _, entry := range entries {
				sum = sum.Add(entry.Amount.Credit)
			}
			coa.AddEntry(
				source,
				types.NewEntryRecord(
					types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Operacyjne, accounts.Odplatna),
					types.CreditBalance(sum),
				),
			)

			docs = append(docs, documents.GenerateUnrecordedSellDocument(source.Document, us.Contractor, entries))
		}
	}

	return docs
}

// UnrecordedSellSource is the source of unrecorded sell document.
type UnrecordedSellSource struct {
	Document   types.Document
	Contractor types.Contractor
}

// GetDate returns date of unrecorded sell.
func (uss *UnrecordedSellSource) GetDate() time.Time {
	return uss.Document.Date
}

// GetDocument returns document.
func (uss *UnrecordedSellSource) GetDocument() types.Document {
	return uss.Document
}

// GetContractor returns contractor.
func (uss *UnrecordedSellSource) GetContractor() types.Contractor {
	return uss.Contractor
}

// GetNotes returns notes.
func (uss *UnrecordedSellSource) GetNotes() string {
	return "Sprzedaż nieewidencjonowana"
}

type withDate interface {
	GetDate() time.Time
}

func findDayRecords[T withDate](records *[]T) []T {
	if len(*records) == 0 {
		return nil
	}
	day := (*records)[0].GetDate().Day()
	for i, r := range *records {
		if r.GetDate().Day() != day {
			result := (*records)[:i]
			*records = (*records)[i:]
			return result
		}
	}
	result := *records
	*records = (*records)[:0]
	return result
}
