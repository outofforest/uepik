package main

import . "github.com/outofforest/uepik/v2" //nolint:staticcheck

// KursyWalutowe przechowuje średnie kursy walutowe NBP.
var KursyWalutowe = Kursy(
	// ====================================
	Kurs(EUR, Data(2024, 12, 31), 4, 3400),
	Kurs(EUR, Data(2025, 1, 1), 4, 5200),
	Kurs(EUR, Data(2025, 1, 2), 4, 4300),
	Kurs(EUR, Data(2025, 1, 7), 4, 4800),
	Kurs(EUR, Data(2025, 5, 2), 4, 4300),
)
