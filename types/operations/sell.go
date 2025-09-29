package operations

import (
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Sell defines the income coming from goods or service sell.
type Sell struct {
	Date       time.Time
	Document   types.Document
	Contractor types.Contractor
	Amount     types.Denom
	Payments   []types.Payment
}

// GetDate returns date of sell.
func (s *Sell) GetDate() time.Time {
	return s.Date
}

// GetDocument returns document.
func (s *Sell) GetDocument() types.Document {
	return s.Document
}

// GetContractor returns contractor.
func (s *Sell) GetContractor() types.Contractor {
	return s.Contractor
}

// GetNotes returns notes.
func (s *Sell) GetNotes() string {
	return "Opis"
}

// BankRecords returns bank records for the sell.
func (s *Sell) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range s.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       s.Document,
			Contractor:     s.Contractor,
			OriginalAmount: payment.Amount,
		})
	}
	return records
}

// BookRecords returns book records for the sell.
func (s *Sell) BookRecords(coa *types.ChartOfAccounts, bankRecords []*types.BankRecord, rates types.CurrencyRates) {
	incomeBase, incomeRate := rates.ToBase(s.Amount, types.PreviousDay(s.Date))

	coa.AddEntry(s,
		types.NewEntryRecord(
			types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.Operacyjne, accounts.ZOdplatnejDPP,
				accounts.ZeSprzedazy),
			types.CreditBalance(incomeBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
			types.CreditBalance(incomeBase),
		),
	)

	if len(s.Payments) > 0 && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	for _, br := range bankRecords {
		if br.Rate.EQ(incomeRate) {
			continue
		}

		var amount types.AccountBalance
		if incomeRate.GT(br.Rate) {
			amount = types.DebitBalance(br.OriginalAmount.ToBase(incomeRate.Sub(br.Rate)))
		} else {
			amount = types.CreditBalance(br.OriginalAmount.ToBase(br.Rate.Sub(incomeRate)))
		}

		coa.AddEntry(types.NewCurrencyDiff(s, br),
			types.NewEntryRecord(
				types.NewAccountID(accounts.RozniceKursowe),
				amount,
			),
		)

		vatDate := types.MinDate(s.Date, br.Date)
		vatBase, _ := rates.ToBase(br.OriginalAmount, types.PreviousDay(vatDate))
		coa.AddEntry(types.NewVAT(vatDate, s),
			types.NewEntryRecord(
				types.NewAccountID(accounts.VAT),
				types.CreditBalance(vatBase),
			),
		)
	}
}

// Documents generate documents for operation.
func (s *Sell) Documents(coa *types.ChartOfAccounts) []types.ReportDocument {
	return nil
}
