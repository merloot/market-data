package storage

import "github.com/merloot/market-data/internal/domain/market"

type MarketDataFilter struct {
	Currencies []string
	market.Period
}

type MarketDataRow struct {
	Currency              string
	LogoURL               *string
	Price                 string
	MarketCap             string
	CirculatingSupply     string
	TotalSupply           string
	PriceDiff             string
	MarketCapDiff         string
	CirculatingSupplyDiff string
}
