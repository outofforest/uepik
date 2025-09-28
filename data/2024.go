package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

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
