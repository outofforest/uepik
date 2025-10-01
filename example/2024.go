//nolint:dupword
package main

import . "github.com/outofforest/uepik" //nolint:staticcheck

var (
	rejs2026HR01 = Grupa(
		Sprzedaz(
			Data(2025, 1, 2),
			Dokument("FV/01/2024", Data(2025, 1, 1)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Naleznosci(
				Naleznosc(Data(2025, 1, 1), Kwota(1, 23, EUR)),
				Naleznosc(Data(2025, 1, 2), Kwota(4, 23, EUR)),
			),
			Platnosci(
				Platnosc("WB/EUR/2025/01/23", Data(2025, 1, 1), 1, Kwota(1, 23, EUR)),
			),
			Ewidencjonowana,
			"Miejsce na rejsie 2026/01",
		),
		Sprzedaz(
			Data(2025, 1, 2),
			Dokument("U/01/2024", Data(2025, 1, 1)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Naleznosci(
				Naleznosc(Data(2025, 1, 1), Kwota(1, 23, EUR)),
				Naleznosc(Data(2025, 1, 2), Kwota(4, 23, EUR)),
			),
			Platnosci(
				Platnosc("WB/EUR/2025/01/23", Data(2025, 1, 1), 1, Kwota(1, 23, EUR)),
			),
			Nieewidencjonowana,
			"Miejsce na rejsie",
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/01/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc("WB/EUR/2025/01/24", Data(2025, 1, 6), 1, Kwota(10, 11, EUR))),
			KUP,
			Odplatna,
			"Czarter jachtu na rejs",
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/02/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc("WB/EUR/2025/01/25", Data(2024, 1, 6), 1, Kwota(10, 11, EUR))),
			NKUP,
			Nieodplatna,
			"Miejsce na rejsie",
		),
	)
)

// R2024 to dane księgowe za rok 2024.
var R2024 = Rok(
	"NazwaFirmy", "Al. Jerozolimskie 1, 00-199 Warszawa", "1111111111",
	Data(2024, 1, 1), Data(2024, 12, 31),
	BilansOtwarcia(
		Kwota(0, 0, PLN),
		Waluty(
			Waluta(Kwota(100, 0, EUR), Kwota(425, 0, PLN)),
		),
	),
	// ========================================================
	rejs2026HR01,
)
