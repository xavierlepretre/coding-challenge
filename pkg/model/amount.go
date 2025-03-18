package model

import (
	"fmt"

	"github.com/JohnCGriffin/overflow"
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
	// This is incomplete. To get the "real" number, you have to shift right by the number of digits of the currency.
	Number       int64
	CurrencyCode CurrencyCode
}

func NewAmountFromInt64(n int64, currencyCode CurrencyCode) (Amount, error) {
	_, ok := GetDigits(currencyCode)
	if !ok {
		return Amount{}, InvalidCurrencyCodeError{currencyCode}
	}
	return Amount{n, currencyCode}, nil
}

func (a Amount) Add(b Amount) (Amount, bool) {
	if a.CurrencyCode != b.CurrencyCode {
		return Amount{}, false
	}
	sum, ok := overflow.Add64(a.Number, b.Number)
	if !ok {
		return Amount{}, false
	}
	return Amount{Number: sum, CurrencyCode: a.CurrencyCode}, ok
}
