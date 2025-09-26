package accounts

import "github.com/outofforest/uepik/types"

// Accounts.
const (
	CIT types.AccountIDPart = iota
	Przychody
	PrzychodyNieoperacyjne
	PrzychodyFinansowe
	DodatnieRozniceKursowe
	PrzychodyOperacyjne
	PrzychodyZNieodplatnejDPP
	DarowiznyOtrzymane
	PrzychodyZOdplatnejDPP
	PrzychodyZeSprzedazy
	Koszty
	KosztyPodatkowe
	KosztyFinansowe
	UjemneRozniceKursowe
	PodatkoweKosztyOperacyjne
	KosztyNiepodatkowe
	NiepodatkoweKosztyOperacyjne
	VAT
)
