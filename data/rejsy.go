//nolint:dupword
package data

import . "github.com/outofforest/uepik" //nolint:staticcheck

var (
	rejs2026HR01 = Grupa(
		Sprzedaz(
			Dokument("FV/01/2024", Data(2025, 1, 1)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Platnosc(Kwota(1, 23, EUR), Data(2025, 1, 1), 1),
			CIT(Data(2025, 1, 2)),
			VAT(Data(2025, 1, 1)),
		),
		Zakup(
			Dokument("FV/01/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Platnosc(Kwota(10, 11, EUR), Data(2025, 1, 6), 1),
			KUP,
			CIT(Data(2025, 1, 8)),
			VAT(Data(2025, 1, 5)),
		),
		Zakup(
			Dokument("FV/02/2024", Data(2025, 1, 5)),
			Kontrahent("INVINI sp. z o. o.", "", ""),
			Platnosc(Kwota(10, 11, EUR), Data(2024, 1, 6), 1),
			NKUP,
			CIT(Data(2025, 1, 8)),
			VAT(Data(2025, 1, 5)),
		),
	)
)
