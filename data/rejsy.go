//nolint:dupword
package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

var (
	rejs2026HR01 = Grupa(
		Sprzedaz(
			Data(2025, 1, 2),
			Dokument("FV/01/2024", Data(2025, 1, 1)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(1, 23, EUR),
			Platnosci(Platnosc(Kwota(1, 23, EUR), Data(2025, 1, 1), 1)),
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/01/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc(Kwota(10, 11, EUR), Data(2025, 1, 6), 1)),
			KUP,
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/02/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc(Kwota(10, 11, EUR), Data(2024, 1, 6), 1)),
			NKUP,
		),
	)
)
