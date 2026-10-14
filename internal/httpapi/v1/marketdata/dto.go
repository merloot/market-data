package marketdata

type CurrencyDTO struct {
	CoinID   string `json:"coinId"`
	CoinName string `json:"coinName"`
}

type SupportedCurrenciesDTO struct {
	Data []CurrencyDTO `json:"data"`
}

type MarketDataInfoDTO struct {
	CoinName              string  `json:"coinName"`
	LogoURL               *string `json:"LogoUrl"`
	Price                 float64 `json:"price"`
	MarketCap             float64 `json:"marketCap"`
	TotalSupply           float64 `json:"total_supply"`
	CirculatingSupply     float64 `json:"circulatingSupply"`
	CirculatingSupplyDiff float64 `json:"circulatingSupplyDiff"`
	PriceDiff             float64 `json:"priceDiff"`
	MarketCapDiff         float64 `json:"MarketCapDiff"`
}

type MarketDataInfoListDTO struct {
	Data []MarketDataInfoDTO `json:"data"`
}

type PriceHistoryDTO struct {
	Prices       [][2]float64 `json:"prices"`
	MarketCaps   [][2]float64 `json:"marketCaps"`
	TotalVolumes [][2]float64 `json:"totalVolumes"`
}
