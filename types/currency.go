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
	return d.Amount.EQ(denom.Amount)
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
	return d.Amount.GT(denom.Amount)
}

// LT checks if denom is less than the other one.
func (d Denom) LT(denom Denom) bool {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return d.Amount.LT(denom.Amount)
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
		Amount:   d.Amount.Add(denom.Amount),
	}
}

// Sub subtracts two denoms.
func (d Denom) Sub(denom Denom) Denom {
	if d.Currency != denom.Currency {
		panic("Currency mismatch.")
	}
	return Denom{
		Currency: d.Currency,
		Amount:   d.Amount.Sub(denom.Amount),
	}
}

// Neg negates denom.
func (d Denom) Neg() Denom {
	return Denom{
		Currency: d.Currency,
		Amount:   d.Amount.Neg(),
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
	if d.Amount.IsZero() && denom.Amount.IsZero() {
		return NewNumber(0, 0, currency.RatePrecision)
	}
	if d.Amount.IsZero() || denom.Amount.IsZero() {
		panic("one of denoms is zero")
	}

	return newNumberFromDecimal(d.Amount.decimal.DivRound(denom.Amount.decimal, int32(currency.RatePrecision)),
		currency.RatePrecision)
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

// IsZero checks if stored value is zero.
func (n Number) IsZero() bool {
	return n.decimal.IsZero()
}

// EQ checks if two numbers are equal.
func (n Number) EQ(n2 Number) bool {
	return n.decimal.Equal(n2.decimal)
}

// NEQ checks if two numbers are not equal.
func (n Number) NEQ(n2 Number) bool {
	return !n.EQ(n2)
}

// GT checks if number is greater than the other one.
func (n Number) GT(n2 Number) bool {
	return n.decimal.GreaterThan(n2.decimal)
}

// LT checks if number is less than the other one.
func (n Number) LT(n2 Number) bool {
	return n.decimal.LessThan(n2.decimal)
}

// GTE checks if number is greater than or equal to the other one.
func (n Number) GTE(n2 Number) bool {
	return !n.LT(n2)
}

// LTE checks if number is less than or equal to the other one.
func (n Number) LTE(n2 Number) bool {
	return !n.GT(n2)
}

// Add adds numbers.
func (n Number) Add(n2 Number) Number {
	return newNumberFromDecimal(n.decimal.Add(n2.decimal), n.precision)
}

// Sub subtracts two numbers.
func (n Number) Sub(n2 Number) Number {
	return newNumberFromDecimal(n.decimal.Sub(n2.decimal), n.precision)
}

// Neg negates number.
func (n Number) Neg() Number {
	return newNumberFromDecimal(n.decimal.Neg(), n.precision)
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
