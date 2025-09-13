package operations

import (
	"github.com/outofforest/uepik/types"
)

// Service groups operations related to the service.
type Service struct {
	ID         string
	Operations []types.Operation
}

// BankRecords returns bank records for the service.
func (s Service) BankRecords(period types.Period) []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, o := range s.Operations {
		records = append(records, o.BankRecords(period)...)
	}
	return records
}

// BookRecords returns book records for the service.
func (s Service) BookRecords(period types.Period, rates types.CurrencyRates) []types.BookRecord {
	records := []types.BookRecord{}
	for _, o := range s.Operations {
		records = append(records, o.BookRecords(period, rates)...)
	}
	return records
}
