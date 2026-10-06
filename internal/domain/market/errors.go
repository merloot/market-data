package market

import "errors"

var (
	ErrNotFound    = errors.New("Not found")
	ErrUpstream    = errors.New("Upstream errors")
	ErrRateLimited = errors.New("Rate limited")
)
