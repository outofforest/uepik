package types

// Init stores initial values for fiscal year.
type Init struct {
	UnspentProfit Denom
	Currencies    InitCurrencies
}

// InitCurrencies stores initial sums of currencies.
type InitCurrencies map[CurrencySymbol]InitCurrency

// InitCurrency stores initial balance of currency.
type InitCurrency struct {
	OriginalSum Denom
	BaseSum     Denom
}
