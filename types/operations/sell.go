package operations

import (
	"fmt"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Sell defines the income coming from goods or service sell.
type Sell struct {
	Document   types.Document
	Contractor types.Contractor
	Payment    types.Payment
	CIT        types.CIT
	VAT        types.VAT

	paymentBankRecord *types.BankRecord
}

// BankRecords returns bank records for the sell.
func (s *Sell) BankRecords(period types.Period) []*types.BankRecord {
	if !period.Contains(s.Payment.Date) {
		return nil
	}

	s.paymentBankRecord = &types.BankRecord{
		Date:           s.Payment.Date,
		Index:          s.Payment.Index,
		Document:       s.Document,
		Contractor:     s.Contractor,
		OriginalAmount: s.Payment.Amount,
	}

	return []*types.BankRecord{s.paymentBankRecord}
}

// BookRecords returns book records for the sell.
func (s *Sell) BookRecords(coa *types.ChartOfAccounts, rates types.CurrencyRates) {
	incomeBase, incomeRate := rates.ToBase(s.Payment.Amount, types.PreviousDay(s.CIT.Date))
	coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyOperacyjne,
		accounts.PrzychodyZOdplatnejDPP, accounts.PrzychodyZeSprzedazy),
		types.NewEntry(s.CIT.Date, 0, s.Document, s.Contractor, incomeBase,
			fmt.Sprintf("kwota: %s, kurs: %s", s.Payment.Amount, incomeRate)))

	if s.paymentBankRecord != nil && s.paymentBankRecord.BaseAmount.NEQ(incomeBase) {
		if incomeBase.GT(s.paymentBankRecord.BaseAmount) {
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
				accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
				types.NewEntry(s.CIT.Date, 0, types.Document{},
					types.Contractor{}, incomeBase.Sub(s.paymentBankRecord.BaseAmount),
					fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", s.Payment.Amount, incomeRate,
						s.paymentBankRecord.Rate)))
		} else {
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
				accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe), types.NewEntry(s.CIT.Date, 0, types.Document{},
				types.Contractor{}, s.paymentBankRecord.BaseAmount.Sub(incomeBase),
				fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", s.Payment.Amount, incomeRate,
					s.paymentBankRecord.Rate)))
		}
	}

	vatBase, vatRate := rates.ToBase(s.Payment.Amount, types.PreviousDay(s.VAT.Date))
	coa.AddEntry(types.NewAccountID(accounts.VAT), types.NewEntry(s.VAT.Date, 0, s.Document, s.Contractor, vatBase,
		fmt.Sprintf("kwota: %s, kurs: %s", s.Payment.Amount, vatRate)))
}
