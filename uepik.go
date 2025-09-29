//nolint:misspell
package uepik

import (
	"math"
	"time"

	"github.com/samber/lo"

	"github.com/outofforest/uepik/report"
	"github.com/outofforest/uepik/types"
	"github.com/outofforest/uepik/types/operations"
)

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

var timeLocation = lo.Must(time.LoadLocation("Europe/Warsaw"))

// Data tworzy datę.
func Data(rok, miesiac, dzien uint64) time.Time {
	return time.Date(int(rok), time.Month(miesiac), int(dzien), 0, 0, 0, 0, timeLocation)
}

// Teraz zwraca bieżący czas.
func Teraz() time.Time {
	return time.Now().In(timeLocation)
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
	period := types.Period{
		Start: dataRozpoczecia,
		End:   dataZakonczenia.AddDate(0, 0, 1).Add(-time.Nanosecond),
	}

	return &types.FiscalYear{
		CompanyName:    nazwaFirmy,
		CompanyAddress: adresFirmy,
		CompanyTaxID:   nipFirmy,
		Period:         period,
		Init:           bilansOtwarcia,
		Operations:     Grupa(operacje...),
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
func Dokument(numer types.DocumentID, data time.Time) types.Document {
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
func Platnosc(dokument types.DocumentID, data time.Time, index uint64, kwota types.Denom) types.Payment {
	return types.Payment{
		DocumentID: dokument,
		Date:       data,
		Index:      index,
		Amount:     kwota,
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
	kontrahent types.Contractor,
	platnosc types.Payment,
) []types.Operation {
	return []types.Operation{&operations.Donation{
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

// Raport generuje raport.
func Raport(
	naDzien time.Time,
	biezacyRok *types.FiscalYear,
	kursyWalutowe types.CurrencyRates,
	lata ...*types.FiscalYear,
) {
	report.Save(naDzien, biezacyRok, kursyWalutowe, lata)
}
