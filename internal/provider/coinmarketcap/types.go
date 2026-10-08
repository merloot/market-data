package coinmarketcap

import "time"

type marketDataListResponse struct {
	Data map[string]marketDataItem `json:"data"`
}

type marketDataItem struct {
	ID                string           `json:"id"`
	CirculatingSupply float64          `json:"circulating_supply"`
	TotalSupply       float64          `json:"total_supply"`
	LastUpdated       time.Time        `json:"last_updated"`
	Quote             map[string]quote `json:"quote"`
}

type quote struct {
	Price       float64   `json:"price"`
	MarketCap   float64   `json:"market_cap"`
	LastUpdated time.Time `json:"last_updated"`
}

type historicalResponse struct {
	Data historicalData `json:"data"`
}

type historicalData struct {
	Quotes []historicalQuote `json:"quotes"`
}

type historicalQuote struct {
	Timestamp string                 `json:"timestamp"`
	Quote     map[string]quoteDetail `json:"quote"`
}

type quoteDetail struct {
	Price             float64 `json:"price"`
	MarketCap         float64 `json:"market_cap"`
	TotalSupply       float64 `json:"total_supply"`
	CirculatingSupply float64 `json:"circulating_supply"`
	Timestamp         string  `json:"timestamp"`
}


