package market

import "time"

type Currency struct {
	CoinName      string
	CoinGecko     string
	CoinMarketCap string
	LogoUrl       string
	Provider      string
}

type MarketDataHistory struct {
	Currency          string
	Price             float64
	MarketCap         float64
	CirculatingSupply float64
	TotalSupply       float64
	Timestamp         time.Time
}

type CurrencyToFind struct {
	Provider string
	Symbol   string
	Slug     string
}

type CurrencyData struct {
	ID      string
	LogoURL string
}

type MarketData struct {
	Currency          string
	CurrentPrice      float64
	MarketCap         float64
	CirculatingSupply float64
	TotalSupply       float64
	LastUpdated       time.Time
}

type MarketDataChart struct {
	Prices      [][2]float64
	MarketCaps  [][2]float64
	TotalVolume [][2]float64
}

type MarketDataRow struct {
	Currency      string
	Price         float64
	MarketCap     float64
	PriceDiff     float64
	MarketCapDiff float64
}

const (
	ProviderCoinGecko = "coin-gecko"
)
