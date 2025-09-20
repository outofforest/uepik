//nolint:dupword
package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

// R2025 to dane księgowe za rok 2025.
var R2025 = Rok(2025, kursy2025,
	Darowizna(
		Dokument("WB/EUR/2025/01/01", Data(2025, 1, 3)),
		Kontrahent("INVINI sp. z o. o.", "", ""),
		Platnosc(Kwota(500, 0, EUR), Data(2025, 1, 3), 1)),
	rejs2026HR01,
)

var kursy2025 = Kursy(
	Kurs(EUR, Data(2024, 12, 31), 4, 3400),
	Kurs(EUR, Data(2025, 1, 1), 4, 5200),
	Kurs(EUR, Data(2025, 1, 2), 4, 4300),
)
