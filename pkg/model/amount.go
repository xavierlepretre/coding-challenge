package model

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

type InvalidNumberError struct {
	Number string
}

func (e InvalidNumberError) Error() string {
	return fmt.Sprintf("invalid number %q", e.Number)
}

type InvalidCurrencyCodeError struct {
	CurrencyCode CurrencyCode
}

func (e InvalidCurrencyCodeError) Error() string {
	return fmt.Sprintf("invalid currency code %q", e.CurrencyCode)
}

type Amount struct {
	Number       apd.Decimal
	CurrencyCode CurrencyCode
}

func NewAmountFromInt64(n int64, currencyCode CurrencyCode) (Amount, error) {
	d, ok := GetDigits(currencyCode)
	if !ok {
		return Amount{}, InvalidCurrencyCodeError{currencyCode}
	}
	number := apd.Decimal{}
	number.SetFinite(n, -int32(d))

	return Amount{number, currencyCode}, nil
}
