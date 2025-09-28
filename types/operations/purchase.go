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
	Amount      types.Denom
	Payments    []types.Payment
	CostTaxType types.CostTaxType
	CIT         types.CIT
}

// BankRecords returns bank records for the purchase.
func (p *Purchase) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range p.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       p.Document,
			Contractor:     p.Contractor,
			OriginalAmount: payment.Amount.Neg(),
		})
	}
	return records
}

// BookRecords returns book records for the purchase.
func (p *Purchase) BookRecords(coa *types.ChartOfAccounts, bankRecords []*types.BankRecord, rates types.CurrencyRates) {
	costBase, costRate := rates.ToBase(p.Amount, types.PreviousDay(p.CIT.Date))

	coa.AddEntry(costTaxTypeToAccountID(p.CostTaxType), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
		types.DebitBalance(costBase), fmt.Sprintf("kwota: %s, kurs: %s", p.Amount, costRate)))

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

	if len(p.Payments) > 0 && len(bankRecords) == 0 {
		panic("brak rekordu walutowego dla płatności")
	}

	for _, br := range bankRecords {
		if br.Rate.EQ(costRate) {
			continue
		}
		paymentOriginal := br.OriginalAmount.Neg()
		if costRate.GT(br.Rate) {
			amount := types.CreditBalance(paymentOriginal.ToBase(costRate.Sub(br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Przychody, accounts.PrzychodyNieoperacyjne,
				accounts.PrzychodyFinansowe, accounts.DodatnieRozniceKursowe),
				types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, amount,
					fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", br.OriginalAmount,
						costRate, br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
				amount, ""))
		} else {
			amount := types.DebitBalance(paymentOriginal.ToBase(br.Rate.Sub(costRate)))
			coa.AddEntry(types.NewAccountID(accounts.CIT, accounts.Koszty, accounts.KosztyPodatkowe,
				accounts.KosztyFinansowe, accounts.UjemneRozniceKursowe),
				types.NewEntry(p.CIT.Date, types.Document{}, types.Contractor{}, amount,
					fmt.Sprintf("Różnice kursowe. Kwota: %s, kurs CIT: %s, kurs wpłaty: %s", br.OriginalAmount,
						costRate, br.Rate)))
			coa.AddEntry(types.NewAccountID(accounts.NiewydatkowanyDochod,
				accounts.NiewydatkowanyDochodWTrakcieRoku), types.NewEntry(p.CIT.Date, p.Document, p.Contractor,
				amount, ""))
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
