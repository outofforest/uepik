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
		costBase, fmt.Sprintf("kwota: %s, kurs: %s", p.Payment.Amount, costRate)))

	if p.Payment.IsPaid() && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	if len(bankRecords) > 0 {
		br := bankRecords[0]

		if br.BaseAmount.NEQ(costBase.Neg()) {
			paymentBase := br.BaseAmount.Neg()
			if costBase.GT(paymentBase) {
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
					accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe),
					types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, costBase.Sub(paymentBase),
						fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", p.Payment.Amount, costRate,
							br.Rate)))
			} else {
				coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
					accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
					types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, paymentBase.Sub(costBase),
						fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", p.Payment.Amount, costRate,
							br.Rate)))
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
