package types

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Currency symbols.
const (
	PLN CurrencySymbol = "PLN"
	EUR CurrencySymbol = "EUR"
)

// BaseCurrency is the currency used to generate reports.
var BaseCurrency = Currency{Symbol: PLN, AmountPrecision: 2, RatePrecision: 0}

// BaseZero is the zero value of base currency.
var BaseZero = Denom{
	Currency: BaseCurrency.Symbol,
	Amount:   NewNumber(0, 0, BaseCurrency.AmountPrecision),
}

// Currencies is the dictionary of defined currencies.
var Currencies = CurrencyMap{
	PLN: BaseCurrency,
	EUR: {Symbol: EUR, AmountPrecision: 2, RatePrecision: 4},
}

// CurrencyMap is used to define currencies.
type CurrencyMap map[CurrencySymbol]Currency

// Currency returns currency by symbol.
func (cm CurrencyMap) Currency(symbol CurrencySymbol) Currency {
	c, exists := cm[symbol]
	if !exists {
		panic(errors.Errorf("unknown currency '%s'", symbol))
	}
	return c
}

// Period defines date range for fiscal year.
type Period struct {
	Start time.Time
	End   time.Time
}

// Contains checks if provided date fits into the range.
func (p Period) Contains(date time.Time) bool {
	return !date.Before(p.Start) && !date.After(p.End)
}

// CurrencySymbol defines type for symbol of the currency.
type CurrencySymbol string

// Currency defines properties of currency.
type Currency struct {
	Symbol          CurrencySymbol
	AmountPrecision uint64
	RatePrecision   uint64
}

// Denom is the amount of currency.
type Denom struct {
	Currency CurrencySymbol
	Amount   Number
}

// String converts denom to string representation.
func (d Denom) String() string {
	return fmt.Sprintf("%s %s", d.Amount, d.Currency)
}

// EQ checks if two denoms are equal.
func (d Denom) EQ(denom Denom) bool {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return d.Amount.decimal.Equal(denom.Amount.decimal)
}

// NEQ checks if two denoms are not equal.
func (d Denom) NEQ(denom Denom) bool {
	return !d.EQ(denom)
}

// GT checks if denom is greater than the other one.
func (d Denom) GT(denom Denom) bool {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return d.Amount.decimal.GreaterThan(denom.Amount.decimal)
}

// LT checks if denom is less than the other one.
func (d Denom) LT(denom Denom) bool {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return d.Amount.decimal.LessThan(denom.Amount.decimal)
}

// GTE checks if denom is greater than or equal to the other one.
func (d Denom) GTE(denom Denom) bool {
	return !d.LT(denom)
}

// LTE checks if denom is less than or equal to the other one.
func (d Denom) LTE(denom Denom) bool {
	return !d.GT(denom)
}

// Add adds denoms.
func (d Denom) Add(denom Denom) Denom {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return Denom{
		Currency: d.Currency,
		Amount:   newNumberFromDecimal(d.Amount.decimal.Add(denom.Amount.decimal), d.Amount.precision),
	}
}

// ToBase converts denom to the base currency.
func (d Denom) ToBase(rate Number) Denom {
	currency := Currencies.Currency(d.Currency)
	if rate.precision != currency.RatePrecision {
		panic("Currency mismatch.")
	}

	return Denom{
		Currency: PLN,
		Amount: newNumberFromDecimal(
			d.Amount.decimal.Mul(rate.decimal).Round(int32(BaseCurrency.AmountPrecision)),
			BaseCurrency.AmountPrecision,
		),
	}
}

// Rate calculates the rate between two denoms.
func (d Denom) Rate(denom Denom) Number {
	currency := Currencies.Currency(denom.Currency)
	return newNumberFromDecimal(d.Amount.decimal.DivRound(denom.Amount.decimal, int32(currency.RatePrecision)),
		currency.RatePrecision)
}

// Sub subtracts two denoms.
func (d Denom) Sub(denom Denom) Denom {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return Denom{
		Currency: d.Currency,
		Amount:   newNumberFromDecimal(d.Amount.decimal.Sub(denom.Amount.decimal), d.Amount.precision),
	}
}

// NewNumber creates new number.
func NewNumber(i, d, precision uint64) Number {
	dec := decimal.New(int64(d), int32(-precision))
	if dec.GreaterThanOrEqual(decimal.New(1, 0)) {
		panic("decimal overflow")
	}
	return Number{
		precision: precision,
		decimal:   decimal.New(int64(i), 0).Add(dec),
	}
}

func newNumberFromDecimal(dec decimal.Decimal, precision uint64) Number {
	rounded := dec.Round(int32(precision))
	if !dec.Equal(rounded) {
		panic("precision mismatch")
	}
	return Number{
		precision: precision,
		decimal:   rounded,
	}
}

// Number represents decimal number.
type Number struct {
	precision uint64
	decimal   decimal.Decimal
}

// String returns string representation of the number.
func (n Number) String() string {
	return n.decimal.StringFixed(int32(n.precision))
}

// ToFloat64 converts decimal number to float64.
func (n Number) ToFloat64() float64 {
	f, _ := n.decimal.Float64()
	return f
}

// CurrencyRateKey is used a key in the currency rate map.
type CurrencyRateKey struct {
	Currency CurrencySymbol
	Date     time.Time
}

// CurrencyRate defines the currency rate for currency and date.
type CurrencyRate struct {
	Key  CurrencyRateKey
	Rate Number
}

// CurrencyRates defines dictionary of currency rates.
type CurrencyRates map[CurrencyRateKey]Number

// ToBase converts denom to the base currency.
func (cr CurrencyRates) ToBase(denom Denom, date time.Time) (Denom, Number) {
	rate := cr.rate(denom.Currency, date)
	return denom.ToBase(rate), rate
}

func (cr CurrencyRates) rate(currency CurrencySymbol, date time.Time) Number {
	if currency == PLN {
		return Number{
			decimal: decimal.New(1, 0),
		}
	}
	rate := cr[CurrencyRateKey{
		Currency: currency,
		Date:     date,
	}]

	var zeroRate Number
	if rate == zeroRate {
		panic(errors.Errorf("invalid rate for %s@%s", currency, date))
	}

	return rate
}

// Document defines the document.
type Document struct {
	ID   string
	Date time.Time
}

// CIT defines CIT properties.
type CIT struct {
	Date time.Time
}

// VAT defines VAT properties.
type VAT struct {
	Date time.Time
}

// Contractor defines contractor.
type Contractor struct {
	Name    string
	Address string
	TaxID   string
}

// Payment defines payment.
type Payment struct {
	Amount Denom
	Date   time.Time
	Index  uint64
}

// Operation defines operation which might bee accounted.
type Operation interface {
	BankRecords(period Period) []*BankRecord
	BookRecords(period Period, rates CurrencyRates) []BookRecord
}

// BookRecord defines the book record.
type BookRecord struct {
	Date            time.Time
	Index           uint64
	Document        Document
	Contractor      Contractor
	IncomeDonations Denom
	IncomeTrading   Denom
	IncomeOthers    Denom
	CostTaxed       Denom
	CostNotTaxed    Denom
	Notes           string
}

// BankRecord defines the properties of bank record.
type BankRecord struct {
	Date           time.Time
	Index          uint64
	OriginalAmount Denom
	BaseAmount     Denom
	Rate           Number
	OriginalSum    Denom
	BaseSum        Denom
	RateAverage    Number
}

// PreviousDay computes the date of the previous day.
func PreviousDay(date time.Time) time.Time {
	return date.AddDate(0, 0, -1)
}

// BankReport is the bank report.
type BankReport struct {
	OriginalCurrency Currency
	BaseCurrency     Currency
	Records          []BankRecord
}

// FiscalYear defines fiscal year.
type FiscalYear struct {
	Period        Period
	CurrencyRates CurrencyRates
	Operations    []Operation
}
