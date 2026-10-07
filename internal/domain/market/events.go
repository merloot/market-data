package market

type PriceUpdated struct {
	CoinName                 string
	Price                    float64
	PriceChangePercentage24h float64
}

type MarketCapUpdated struct {
	CoinName                     string
	MarketCap                    float64
	MarketCapChangePercentage24h float64
}
