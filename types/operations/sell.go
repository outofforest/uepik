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
}

// BankRecords returns bank records for the sell.
func (s *Sell) BankRecords() []*types.BankRecord {
	return []*types.BankRecord{{
		Date:           s.Payment.Date,
		Index:          s.Payment.Index,
		Document:       s.Document,
		Contractor:     s.Contractor,
		OriginalAmount: s.Payment.Amount,
	}}
}

// BookRecords returns book records for the sell.
func (s *Sell) BookRecords(coa *types.ChartOfAccounts, bankRecords []*types.BankRecord, rates types.CurrencyRates) {
	incomeBase, incomeRate := rates.ToBase(s.Payment.Amount, types.PreviousDay(s.CIT.Date))
	coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyOperacyjne,
		accounts.PrzychodyZOdplatnejDPP, accounts.PrzychodyZeSprzedazy),
		types.NewEntry(s.CIT.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase),
			fmt.Sprintf("kwota: %s, kurs: %s", s.Payment.Amount, incomeRate)))
	coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.NiewydatkowanyDochodWTrakcieRoku),
		types.NewEntry(s.CIT.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase), ""))

	if s.Payment.IsPaid() && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	if len(bankRecords) > 0 {
		br := bankRecords[0]

		if br.BaseAmount.NEQ(incomeBase) {
			if incomeBase.GT(br.BaseAmount) {
				amount := types.DebitBalance(incomeBase.Sub(br.BaseAmount))
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
					accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
					types.NewEntry(s.CIT.Date, types.Document{},
						types.Contractor{}, amount,
						fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", s.Payment.Amount,
							incomeRate, br.Rate)))
				coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
					accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(s.CIT.Date, s.Document, s.Contractor,
					amount, ""))
			} else {
				amount := types.CreditBalance(br.BaseAmount.Sub(incomeBase))
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
					accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe), types.NewEntry(s.CIT.Date,
					types.Document{}, types.Contractor{}, amount,
					fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", s.Payment.Amount,
						incomeRate, br.Rate)))
				coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
					accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(s.CIT.Date, s.Document, s.Contractor,
					amount, ""))
			}
		}
	}

	vatBase, vatRate := rates.ToBase(s.Payment.Amount, types.PreviousDay(s.VAT.Date))
	coa.AddEntry(types.NewAccountID(accounts.VAT), types.NewEntry(s.VAT.Date, s.Document, s.Contractor,
		types.CreditBalance(vatBase), fmt.Sprintf("kwota: %s, kurs: %s", s.Payment.Amount, vatRate)))
}
