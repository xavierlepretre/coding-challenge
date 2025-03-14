package model

type currencyInfo struct {
	numericCode CurrencyCode
	digits      uint8
}

var currencies = map[CurrencyCode]currencyInfo{
	"GEL": {"981", 2}, "USD": {"840", 2},
}
