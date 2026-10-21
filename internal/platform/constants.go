package platform

// CurrencyTypes holds ISO 4217 currency codes in the uppercase form stored by
// the currencies table.
type CurrencyTypes string

const (
	USD CurrencyTypes = "USD"
	GHS CurrencyTypes = "GHS"
	EUR CurrencyTypes = "EUR"
	JPY CurrencyTypes = "JPY"
)
