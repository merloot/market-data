package market

import "errors"

var (
	ErrProviderNotFound = errors.New("Provider not found")
	ErrCurrencyNotFound = errors.New("Currency not found")
	ErrUpstream = errors.New("Upstream errors")
	ErrRateLimited = errors.New("Rate limited")
)