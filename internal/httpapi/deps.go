package httpapi

import "github.com/merloot/market-data/internal/marketservice"

type Deps struct {
	MarketService *marketservice.Service
}
