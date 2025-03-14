package model

import "fmt"

type IncompatibleCurrencyCodesError struct {
	ExpectedCurrencyCode CurrencyCode
	ReceivedCurrencyCode CurrencyCode
}

func (e IncompatibleCurrencyCodesError) Error() string {
	return fmt.Sprintf("incompatible currency codes, expected %q, received %q", e.ExpectedCurrencyCode, e.ReceivedCurrencyCode)
}

type CurrencyCode string

func CheckCurrencyCodeCompatible(expected CurrencyCode, received CurrencyCode) error {
	if expected != received {
		return IncompatibleCurrencyCodesError{expected, received}
	}
	return nil
}

// An empty currencyCode is considered valid.
func IsValid(currencyCode CurrencyCode) bool {
	if currencyCode == "" {
		return true
	}
	_, ok := currencies[currencyCode]

	return ok
}

func GetDigits(currencyCode CurrencyCode) (digits uint8, ok bool) {
	if currencyCode == "" || !IsValid(currencyCode) {
		return 0, false
	}
	return currencies[currencyCode].digits, true
}
