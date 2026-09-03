//nolint:dupword,misspell
package main

import . "github.com/outofforest/uepik/v2" //nolint:staticcheck

// R2025 to dane księgowe za rok 2025.
var R2025 = Rok(
	Kontrahent("INVINI sp. z o. o.", "", ""),
	Data(2025, 1, 1), Data(2025, 12, 31),
	BilansOtwarcia(
		Kwota(123, 23, PLN),
		Waluty(
			Waluta(Kwota(100, 0, PLN), Kwota(100, 0, PLN)),
			Waluta(Kwota(34, 65, EUR), Kwota(128, 12, PLN)),
		),
	),
	// ========================================================
	Wplata(
		Kontrahent("Wojciech Małota-Wójcik", "Adres", ""),
		Platnosc("WB/PLN/2025/01/01", Data(2025, 2, 3), 1, Kwota(1000, 0, PLN)),
		"Wpłata kapitału założycielskiego",
	),
	Darowizna(
		Kontrahent("INVINI sp. z o. o.", "Felińskiego 2/17", ""),
		Platnosc("WB/EUR/2025/01/01", Data(2025, 5, 3), 1, Kwota(500, 0, EUR)),
	),
	rejs2026HR01,
	Delegacja(
		Dokument("RD/2025/01/01", Data(2025, 2, 5)),
		Kontrahent("Jan Kowalski", "ul. Nowaka 4, 41-001 Poznań", ""),
		Czas(2025, 1, 1, 17, 0), Czas(2025, 1, 12, 18, 30),
		Chorwacja,
		EUR,
		Platnosci(
			Platnosc("WB/01", Data(2025, 2, 6), 1, Kwota(20, 0, EUR)),
			Platnosc("WB/01", Data(2025, 2, 7), 1, Kwota(25, 0, EUR)),
		),
		KUP,
		Odplatna,
		"Delegacja na rejs szkoleniowy",
		Dieta(),
		Dojazd(),
		KosztDelegacji("FV/13/45", Kwota(100, 25, EUR), "Przelot"),
	),
	UmowaODzielo(
		Data(2025, 1, 20),
		Dokument("UD/1/1", Data(2025, 1, 20)),
		Kontrahent("INVINI sp. z o. o.", "", ""),
		Kwota(1000, 0, PLN),
		Platnosci(Platnosc("WB/EUR/2025/01/24", Data(2025, 1, 21), 1, Kwota(940, 0, PLN))),
		KUP,
		Odplatna,
		"Umowa o dzieło",
	),
	Wymiana(
		Dokument("Wymiana01", Data(2025, 1, 22)),
		Kwota(2, 0, EUR),
		Kwota(10, 0, PLN),
		100, 101,
	),
)
