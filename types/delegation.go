//nolint:misspell
package types

import "time"

// DelegationCountry defines a country person travels to.
type DelegationCountry struct {
	Name               string
	DailyAmount        Denom
	AccommodationLimit Denom
}

// DelegationCost defines cost of delegation.
type DelegationCost interface {
	GetDocument() DocumentID
	GetAmount(start, end time.Time, country DelegationCountry) Denom
	GetNotes() string
}

var (
	_ DelegationCost = DelegationCostAlimentation{}
	_ DelegationCost = DelegationCostAccess{}
	_ DelegationCost = DelegationCostDocumented{}
)

// DelegationCostAlimentation defines delegation cost related to alimentation.
type DelegationCostAlimentation struct{}

// GetDocument returns cost's document ID.
func (c DelegationCostAlimentation) GetDocument() DocumentID {
	return "(ryczałt)"
}

// GetAmount returns cost's amount.
func (c DelegationCostAlimentation) GetAmount(start, end time.Time, country DelegationCountry) Denom {
	days, hours := DelegationDuration(start, end)
	alimentation := country.DailyAmount.MulUint64(days)
	switch {
	case hours == 0.0:
		return alimentation
	case hours <= 8.0:
		return alimentation.Add(country.DailyAmount.MulFloat64(0.3333))
	case hours <= 12.0:
		return alimentation.Add(country.DailyAmount.MulFloat64(0.5))
	default:
		return alimentation.Add(country.DailyAmount)
	}
}

// GetNotes returns cost's description.
func (c DelegationCostAlimentation) GetNotes() string {
	return "Diety"
}

// DelegationCostAccess defines delegation cost related to access.
type DelegationCostAccess struct{}

// GetDocument returns cost's document ID.
func (c DelegationCostAccess) GetDocument() DocumentID {
	return "(ryczałt)"
}

// GetAmount returns cost's amount.
func (c DelegationCostAccess) GetAmount(start, end time.Time, country DelegationCountry) Denom {
	return country.DailyAmount
}

// GetNotes returns cost's description.
func (c DelegationCostAccess) GetNotes() string {
	return "Dojazd z/do dworca/lotniska"
}

// DelegationCostDocumented defines documented delegation cost.
type DelegationCostDocumented struct {
	Document DocumentID
	Amount   Denom
	Notes    string
}

// GetDocument returns cost's document ID.
func (c DelegationCostDocumented) GetDocument() DocumentID {
	return c.Document
}

// GetAmount returns cost's amount.
func (c DelegationCostDocumented) GetAmount(start, end time.Time, country DelegationCountry) Denom {
	return c.Amount
}

// GetNotes returns cost's description.
func (c DelegationCostDocumented) GetNotes() string {
	return c.Notes
}

// DelegationDuration computes delegation duration in days and hours.
func DelegationDuration(start, end time.Time) (uint64, float64) {
	const day = 24 * time.Hour

	d := end.Sub(start)
	return uint64(d / day), float64(d%day) / float64(time.Hour)
}
