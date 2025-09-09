package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

// R2025 to dane księgowe za rok 2025.
var R2025 = Rok(2025, kursy2025,
	rejs2026HR01,
)

var kursy2025 = Kursy(
	Kurs(EUR, Data(2024, 12, 31), 4, 3400),
	Kurs(EUR, Data(2025, 1, 1), 4, 5200),
)
