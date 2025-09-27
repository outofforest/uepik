//nolint:gosmopolitan,misspell
package uepik

import (
	"math"
	"time"

	"github.com/outofforest/uepik/accounts"
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
func Rok(
	nazwaFirmy, adresFirmy string,
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

	return &types.FiscalYear{
		CompanyName:    nazwaFirmy,
		CompanyAddress: adresFirmy,
		ChartOfAccounts: types.NewChartOfAccounts(period,
			types.NewAccount(
				accounts.CIT,
				types.NewAccount(
					accounts.Przychody,
					types.NewAccount(
						accounts.PrzychodyNieoperacyjne,
						types.NewAccount(
							accounts.PrzychodyFinansowe,
							types.NewAccount(accounts.DodatnieRozniceKursowe),
						),
					),
					types.NewAccount(
						accounts.PrzychodyOperacyjne,
						types.NewAccount(
							accounts.PrzychodyZNieodplatnejDPP,
							types.NewAccount(accounts.DarowiznyOtrzymane),
						),
						types.NewAccount(
							accounts.PrzychodyZOdplatnejDPP,
							types.NewAccount(accounts.PrzychodyZeSprzedazy),
						),
					),
				),
				types.NewAccount(
					accounts.Koszty,
					types.NewAccount(
						accounts.KosztyPodatkowe,
						types.NewAccount(
							accounts.KosztyFinansowe,
							types.NewAccount(accounts.UjemneRozniceKursowe),
						),
						types.NewAccount(accounts.PodatkoweKosztyOperacyjne),
					),
					types.NewAccount(
						accounts.KosztyNiepodatkowe,
						types.NewAccount(accounts.NiepodatkoweKosztyOperacyjne),
					),
				),
			),
			types.NewAccount(accounts.VAT),
		),
		Period:     period,
		Init:       bilansOtwarcia,
		Operations: Grupa(operacje...),
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

// Niezaplacono oznacza, że jeszcze nie ma płatnosci.
func Niezaplacono() types.Payment {
	return types.Payment{}
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
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
	cit types.CIT,
	vat types.VAT,
) []types.Operation {
	return []types.Operation{&operations.Sell{
		Document:   dokument,
		Contractor: kontrahent,
		Payment:    platnosc,
		CIT:        cit,
		VAT:        vat,
	}}
}

// Zakup definiuje zakup.
func Zakup(
	dokument types.Document,
	kontrahent types.Contractor,
	platnosc types.Payment,
	typPodatkowy types.CostTaxType,
	cit types.CIT,
	vat types.VAT,
) []types.Operation {
	return []types.Operation{&operations.Purchase{
		Document:    dokument,
		Contractor:  kontrahent,
		Payment:     platnosc,
		CostTaxType: typPodatkowy,
		CIT:         cit,
		VAT:         vat,
	}}
}
