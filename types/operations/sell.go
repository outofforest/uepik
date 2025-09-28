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
	Amount     types.Denom
	Payments   []types.Payment
	CIT        types.CIT
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
	incomeBase, incomeRate := rates.ToBase(s.Amount, types.PreviousDay(s.CIT.Date))
	coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyOperacyjne,
		accounts.PrzychodyZOdplatnejDPP, accounts.PrzychodyZeSprzedazy),
		types.NewEntry(s.CIT.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase),
			fmt.Sprintf("kwota: %s, kurs: %s", s.Amount, incomeRate)))
	coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.NiewydatkowanyDochodWTrakcieRoku),
		types.NewEntry(s.CIT.Date, s.Document, s.Contractor, types.CreditBalance(incomeBase), ""))

	if len(s.Payments) > 0 && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	for _, br := range bankRecords {
		if br.Rate.EQ(incomeRate) {
			continue
		}
		if incomeRate.GT(br.Rate) {
			amount := types.DebitBalance(br.OriginalAmount.ToBase(incomeRate.Sub(br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
				accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
				types.NewEntry(s.CIT.Date, types.Document{},
					types.Contractor{}, amount,
					fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", br.OriginalAmount,
						incomeRate, br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(s.CIT.Date, s.Document, s.Contractor,
				amount, ""))
		} else {
			amount := types.CreditBalance(br.OriginalAmount.ToBase(br.Rate.Sub(incomeRate)))
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
				accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe), types.NewEntry(s.CIT.Date,
				types.Document{}, types.Contractor{}, amount,
				fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", br.OriginalAmount,
					incomeRate, br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(s.CIT.Date, s.Document, s.Contractor,
				amount, ""))
		}

		vatDate := types.MinDate(s.CIT.Date, br.Date)
		vatBase, vatRate := rates.ToBase(br.OriginalAmount, types.PreviousDay(vatDate))
		coa.AddEntry(types.NewAccountID(accounts.VAT), types.NewEntry(vatDate, s.Document, s.Contractor,
			types.CreditBalance(vatBase), fmt.Sprintf("kwota: %s, kurs: %s", br.OriginalAmount, vatRate)))
	}
}
