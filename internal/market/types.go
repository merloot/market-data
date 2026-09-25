package market

import "time"

type Currency struct {
	Currency string
	Provider string
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
	CurrencyPrice     float64
	MarketCap         float64
	CirculatingSupply float64
	TotalSupply       float64
	LastUpdated       time.Time
}


type MarketDataChart struct {
	Price [][2]float64
	MarketCap [][2]float64
	TotalVolume [][2]float64
}