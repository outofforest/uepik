//nolint:dupword
package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

// R2025 to dane księgowe za rok 2025.
var R2025 = Rok(
	"NazwaFirmy", "Al. Jerozolimskie 1, 00-199 Warszawa", "1111111111",
	Data(2025, 1, 1), Data(2025, 12, 31),
	BilansOtwarcia(
		Kwota(123, 23, PLN),
		Waluty(
			Waluta(Kwota(34, 65, EUR), Kwota(128, 12, PLN)),
		),
	),
	// ========================================================
	Darowizna(
		Kontrahent("INVINI sp. z o. o.", "Felińskiego 2/17", ""),
		Platnosc("WB/EUR/2025/01/01", Data(2025, 5, 3), 1, Kwota(500, 0, EUR)),
	),
	rejs2026HR01,
)
