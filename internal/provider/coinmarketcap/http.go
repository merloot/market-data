package coinmarketcap

import (
	"fmt"
	"net/http"

	"github.com/merloot/market-data/internal/domain/market"
)

func checkStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusTooManyRequests:
		return market.ErrRateLimited
	case http.StatusNotFound:
		return market.ErrNotFound
	default:
		return fmt.Errorf("%w: status: %d", market.ErrUpstream, resp.StatusCode)
	}
}
