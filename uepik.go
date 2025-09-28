//nolint:gosmopolitan,misspell
package uepik

import (
	"fmt"
	"math"
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
	"github.com/outofforest/uepik/types/operations"
)

var coaAccounts = []*types.Account{
	types.NewAccount(
		accounts.CIT, types.Liabilities,
		types.NewAccount(
			accounts.Przychody, types.Incomes,
			types.NewAccount(
				accounts.Nieoperacyjne, types.Incomes,
				types.NewAccount(
					accounts.Finansowe, types.Incomes,
					types.NewAccount(accounts.DodatnieRozniceKursowe, types.Incomes),
				),
			),
			types.NewAccount(
				accounts.Operacyjne, types.Incomes,
				types.NewAccount(
					accounts.ZNieodplatnejDPP, types.Incomes,
					types.NewAccount(accounts.Darowizny, types.Incomes),
				),
				types.NewAccount(
					accounts.ZOdplatnejDPP, types.Incomes,
					types.NewAccount(accounts.ZeSprzedazy, types.Incomes),
				),
			),
		),
		types.NewAccount(
			accounts.Koszty, types.Costs,
			types.NewAccount(
				accounts.Podatkowe, types.Costs,
				types.NewAccount(
					accounts.Finansowe, types.Costs,
					types.NewAccount(accounts.UjemneRozniceKursowe, types.Costs),
				),
				types.NewAccount(accounts.Operacyjne, types.Costs),
			),
			types.NewAccount(
				accounts.Niepodatkowe, types.Costs,
				types.NewAccount(accounts.Operacyjne, types.Costs),
			),
		),
	),
	types.NewAccount(accounts.VAT, types.Incomes),
	types.NewAccount(
		accounts.NiewydatkowanyDochod, types.Liabilities,
		types.NewAccount(accounts.WTrakcieRoku, types.Liabilities),
		types.NewAccount(accounts.ZLatUbieglych, types.Liabilities),
	),
	types.NewAccount(accounts.RozniceKursowe, types.Liabilities),
}

// Dostępne waluty.
const (
	PLN = types.PLN
	EUR = types.EUR
)

// Typy podatkowe.
const (
	KUP  = types.CostTaxTypeTaxable
	NKUP = types.CostTaxTypeNonTaxable
)

// Data tworzy datę.
func Data(rok, miesiac, dzien uint64) time.Time {
	return time.Date(int(rok), time.Month(miesiac), int(dzien), 0, 0, 0, 0, time.Local)
}

// Kwota tworzy kwotę.
func Kwota(c, u uint64, waluta types.CurrencySymbol) types.Denom {
	currency := types.Currencies.Currency(waluta)
	if u >= uint64(math.Pow10(int(currency.AmountPrecision))) {
		panic("Część ułamkowa jest zbyt duża.")
	}

	return types.Denom{
		Currency: waluta,
		Amount:   types.NewNumber(c, u, currency.AmountPrecision),
	}
}

// Kurs tworzy kurs walutowy.
func Kurs(waluta types.CurrencySymbol, data time.Time, c, u uint64) types.CurrencyRate {
	currency := types.Currencies.Currency(waluta)
	if u >= uint64(math.Pow10(int(currency.RatePrecision))) {
		panic("Część ułamkowa jest zbyt duża.")
	}

	return types.CurrencyRate{
		Key: types.CurrencyRateKey{
			Currency: waluta,
			Date:     data,
		},
		Rate: types.NewNumber(c, u, currency.RatePrecision),
	}
}

// Kursy tworzy słownik kursów walutowych.
func Kursy(kursy ...types.CurrencyRate) types.CurrencyRates {
	rates := make(types.CurrencyRates, len(kursy))
	for _, rate := range kursy {
		rates[rate.Key] = rate.Rate
	}
	return rates
}

// Rok tworzy rok obrotowy.
func Rok(
	nazwaFirmy, adresFirmy, nipFirmy string,
	dataRozpoczecia, dataZakonczenia time.Time,
	bilansOtwarcia types.Init,
	operacje ...[]types.Operation,
) *types.FiscalYear {
	end := dataZakonczenia.AddDate(0, 0, 1).Add(-time.Nanosecond)
	now := time.Now().Local()
	if end.After(now) {
		end = now
	}
	period := types.Period{
		Start: dataRozpoczecia,
		End:   end,
	}

	coa := types.NewChartOfAccounts(period, coaAccounts...)
	coa.OpenAccount(types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.ZLatUbieglych),
		types.CreditBalance(bilansOtwarcia.UnspentProfit))

	company := types.Contractor{
		Name:    nazwaFirmy,
		Address: adresFirmy,
		TaxID:   nipFirmy,
	}
	for date := period.Start.AddDate(0, 1, 0).Add(-time.Nanosecond); period.Contains(date); date = date.AddDate(0, 1, 0) {
		operacje = append(operacje, []types.Operation{&operations.CurrencyDiff{
			Document: types.Document{
				ID:   fmt.Sprintf("RK/%d/%d/1", date.Year(), date.Month()),
				Date: date,
			},
			Contractor: company,
		}})
	}

	return &types.FiscalYear{
		CompanyName:     nazwaFirmy,
		CompanyAddress:  adresFirmy,
		CompanyTaxID:    nipFirmy,
		ChartOfAccounts: coa,
		Period:          period,
		Init:            bilansOtwarcia,
		Operations:      Grupa(operacje...),
	}
}

// BilansOtwarcia tworzy bilans otwarcia roku.
func BilansOtwarcia(niewydanyZysk types.Denom, waluty types.InitCurrencies) types.Init {
	if niewydanyZysk.Currency != types.BaseCurrency.Symbol {
		panic("nieprawidłowa waluta dla niewydanego zysku")
	}
	return types.Init{
		UnspentProfit: niewydanyZysk,
		Currencies:    waluty,
	}
}

// Waluty tworzą bilans otwarcia dla banku walut.
func Waluty(waluty ...types.InitCurrency) types.InitCurrencies {
	result := types.InitCurrencies{}

	for _, c := range waluty {
		if _, exists := result[c.OriginalSum.Currency]; exists {
			panic("bilans waluty już istnieje")
		}
		result[c.OriginalSum.Currency] = c
	}

	return result
}

// Waluta tworzy bilans otwarcia dla waluty.
func Waluta(kwota types.Denom, kwotaPLN types.Denom) types.InitCurrency {
	if kwotaPLN.Currency != types.BaseCurrency.Symbol {
		panic("nieprawidłowa waluta dla kwoty PLN")
	}
	return types.InitCurrency{
		OriginalSum: kwota,
		BaseSum:     kwotaPLN,
	}
}

// Grupa grupuje operacje.
func Grupa(operacje ...[]types.Operation) []types.Operation {
	var count int
	for _, ops := range operacje {
		count += len(ops)
	}
	ops := make([]types.Operation, 0, count)
	for _, ops2 := range operacje {
		ops = append(ops, ops2...)
	}
	return ops
}

// Dokument tworzy dokument.
func Dokument(numer string, data time.Time) types.Document {
	return types.Document{
		ID:   numer,
		Date: data,
	}
}

// Kontrahent definiuje kontrahenta.
func Kontrahent(nazwa, adres, nip string) types.Contractor {
	return types.Contractor{
		Name:    nazwa,
		Address: adres,
		TaxID:   nip,
	}
}

// Platnosc definiuje płatność.
func Platnosc(kwota types.Denom, data time.Time, index uint64) types.Payment {
	return types.Payment{
		Amount: kwota,
		Date:   data,
		Index:  index,
	}
}

// Platnosci definiuje płatności.
func Platnosci(platnosci ...types.Payment) []types.Payment {
	return platnosci
}

// Niezaplacono oznacza, że jeszcze nie ma płatnosci.
func Niezaplacono() []types.Payment {
	return nil
}

// Darowizna definiuje darowiznę.
func Darowizna(
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
) []types.Operation {
	return []types.Operation{&operations.Donation{
		Document:   dokument,
		Contractor: kontrahent,
		Payment:    platnosc,
	}}
}

// Sprzedaz definiuje sprzedaż.
func Sprzedaz(
	data time.Time,
	dokument types.Document,
	kontrahent types.Contractor,
	kwota types.Denom,
	platnosci []types.Payment,
) []types.Operation {
	return []types.Operation{&operations.Sell{
		Date:       data,
		Document:   dokument,
		Contractor: kontrahent,
		Amount:     kwota,
		Payments:   platnosci,
	}}
}

// Zakup definiuje zakup.
func Zakup(
	data time.Time,
	dokument types.Document,
	kontrahent types.Contractor,
	kwota types.Denom,
	platnosci []types.Payment,
	typPodatkowy types.CostTaxType,
) []types.Operation {
	return []types.Operation{&operations.Purchase{
		Date:        data,
		Document:    dokument,
		Contractor:  kontrahent,
		Amount:      kwota,
		Payments:    platnosci,
		CostTaxType: typPodatkowy,
	}}
}
