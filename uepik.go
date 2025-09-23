//nolint:gosmopolitan,misspell
package uepik

import (
	"math"
	"time"

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
func Rok(rok uint64, kursy types.CurrencyRates, operacje ...types.Operation) types.FiscalYear {
	start := time.Date(int(rok), time.January, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	now := time.Now().Local()
	if end.After(now) {
		end = now
	}
	return types.FiscalYear{
		Period: types.Period{
			Start: start,
			End:   end,
		},
		CurrencyRates: kursy,
		Operations:    operacje,
	}
}

// Usluga grupuje operacje dotyczące jednej usługi.
func Usluga(symbol string, operacje ...types.Operation) operations.Service {
	return operations.Service{
		ID:         symbol,
		Operations: operacje,
	}
}

// Dokument tworzy dokument.
func Dokument(numer string, data time.Time) types.Document {
	return types.Document{
		ID:   numer,
		Date: data,
	}
}

// CIT definiuje właściwości dla CIT.
func CIT(data time.Time) types.CIT {
	return types.CIT{
		Date: data,
	}
}

// VAT definiuje właściwości dla VAT.
func VAT(data time.Time) types.VAT {
	return types.VAT{
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

// Darowizna definiuje darowiznę.
func Darowizna(
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
) *operations.Donation {
	return &operations.Donation{
		Document:   dokument,
		Contractor: kontrahent,
		Payment:    platnosc,
	}
}

// Sprzedaz definiuje sprzedaż.
func Sprzedaz(
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
	cit types.CIT,
	vat types.VAT,
) *operations.Sell {
	return &operations.Sell{
		Document:   dokument,
		Contractor: kontrahent,
		Payment:    platnosc,
		CIT:        cit,
		VAT:        vat,
	}
}

// Zakup definiuje zakup.
func Zakup(
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
	typPodatkowy types.CostTaxType,
	cit types.CIT,
	vat types.VAT,
) *operations.Purchase {
	return &operations.Purchase{
		Document:    dokument,
		Contractor:  kontrahent,
		Payment:     platnosc,
		CostTaxType: typPodatkowy,
		CIT:         cit,
		VAT:         vat,
	}
}
