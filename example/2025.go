//nolint:dupword,misspell
package main

import . "github.com/outofforest/uepik/v2" //nolint:staticcheck

// R2025 to dane księgowe za rok 2025.
var R2025 = Rok(
	"NazwaFirmy", "Al. Jerozolimskie 1, 00-199 Warszawa, Chorwacja", "1111111111",
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
)
