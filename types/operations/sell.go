package operations

import (
	"fmt"
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
	coa.AddEntry(
		types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.Operacyjne, accounts.ZOdplatnejDPP,
			accounts.ZeSprzedazy),
		types.NewEntry(s.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase),
			fmt.Sprintf("kwota: %s, kurs: %s", s.Amount, incomeRate)))
	coa.AddEntry(
		types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
		types.NewEntry(s.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase), ""),
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

		coa.AddEntry(
			types.NewAccountID(accounts.RozniceKursowe),
			types.NewEntry(types.MaxDate(s.Date, br.Date), s.Document, s.Contractor, amount, ""),
		)

		vatDate := types.MinDate(s.Date, br.Date)
		vatBase, vatRate := rates.ToBase(br.OriginalAmount, types.PreviousDay(vatDate))
		coa.AddEntry(types.NewAccountID(accounts.VAT), types.NewEntry(vatDate, s.Document, s.Contractor,
			types.CreditBalance(vatBase), fmt.Sprintf("kwota: %s, kurs: %s", br.OriginalAmount, vatRate)))
	}
}
