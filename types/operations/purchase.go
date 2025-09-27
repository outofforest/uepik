package operations

import (
	"fmt"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Purchase defines the cost of purchased goods or service.
type Purchase struct {
	Document    types.Document
	Contractor  types.Contractor
	Payment     types.Payment
	CostTaxType types.CostTaxType
	CIT         types.CIT
	VAT         types.VAT
}

// BankRecords returns bank records for the purchase.
func (p *Purchase) BankRecords() []*types.BankRecord {
	return []*types.BankRecord{{
		Date:           p.Payment.Date,
		Index:          p.Payment.Index,
		Document:       p.Document,
		Contractor:     p.Contractor,
		OriginalAmount: p.Payment.Amount.Neg(),
	}}
}

// BookRecords returns book records for the purchase.
func (p *Purchase) BookRecords(coa *types.ChartOfAccounts, bankRecords []*types.BankRecord, rates types.CurrencyRates) {
	costBase, costRate := rates.ToBase(p.Payment.Amount, types.PreviousDay(p.CIT.Date))

	coa.AddEntry(costTaxTypeToAccountID(p.CostTaxType), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
		types.DebitBalance(costBase), fmt.Sprintf("kwota: %s, kurs: %s", p.Payment.Amount, costRate)))

	switch p.CostTaxType {
	case types.CostTaxTypeTaxable:
		coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.NiewydatkowanyDochodWTrakcieRoku),
			types.NewEntry(p.CIT.Date, p.Document, p.Contractor, types.DebitBalance(costBase), ""))
	case types.CostTaxTypeNonTaxable:
		cost2 := coa.Balance(types.NewAccountID(accounts.NiewydatkowanyDochod,
			accounts.NiewydatkowanyDochodZLatUbieglych))
		if cost2.LT(types.BaseZero) {
			cost2 = types.BaseZero
		}
		if cost2.GT(costBase) {
			cost2 = costBase
		}
		cost := costBase.Sub(cost2)
		if cost.GT(types.BaseZero) {
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.NiewydatkowanyDochodWTrakcieRoku),
				types.NewEntry(p.CIT.Date, p.Document, p.Contractor, types.DebitBalance(cost), ""))
		}
		if cost2.GT(types.BaseZero) {
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.NiewydatkowanyDochodZLatUbieglych),
				types.NewEntry(p.CIT.Date, p.Document, p.Contractor, types.DebitBalance(cost2), ""))
		}
	default:
		panic("invalid cost tax type")
	}

	if p.Payment.IsPaid() && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	if len(bankRecords) > 0 {
		br := bankRecords[0]

		if br.BaseAmount.NEQ(costBase.Neg()) {
			paymentBase := br.BaseAmount.Neg()
			if costBase.GT(paymentBase) {
				amount := types.CreditBalance(costBase.Sub(paymentBase))
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
					accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe),
					types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, amount,
						fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", p.Payment.Amount,
							costRate, br.Rate)))
				coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
					accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
					amount, ""))
			} else {
				amount := types.DebitBalance(paymentBase.Sub(costBase))
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
					accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
					types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, amount,
						fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", p.Payment.Amount,
							costRate, br.Rate)))
				coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
					accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
					amount, ""))
			}
		}
	}
}

func costTaxTypeToAccountID(costTaxType types.CostTaxType) types.AccountID {
	switch costTaxType {
	case types.CostTaxTypeTaxable:
		return types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
			accounts.PodatkoweKosztyOperacyjne)
	case types.CostTaxTypeNonTaxable:
		return types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyNiepodatkowe,
			accounts.NiepodatkoweKosztyOperacyjne)
	default:
		panic("invalid cost tax type")
	}
}
