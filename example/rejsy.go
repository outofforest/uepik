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
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/01/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc("WB/EUR/2025/01/24", Data(2025, 1, 6), 1, Kwota(10, 11, EUR))),
			KUP,
		),
		Zakup(
			Data(2025, 1, 8),
			Dokument("FV/02/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Kwota(10, 11, EUR),
			Platnosci(Platnosc("WB/EUR/2025/01/25", Data(2024, 1, 6), 1, Kwota(10, 11, EUR))),
			NKUP,
		),
	)
)
