package marketdata

import "time"

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

type PriceHistoryChartDTO struct {
	Prices [][2]any `json:"prices"`
}

type PriceHistoryItemDTO struct {
	Time  time.Time
	Price float64
}

func ToPriceHistoryChartDTO(points []PriceHistoryItemDTO) PriceHistoryChartDTO {
	prices := make([][2]any, 0, len(points))
	for _, p := range points {
		prices = append(prices, [2]any{
			p.Time.UnixMilli(),
			p.Price,
		})
	}
	return PriceHistoryChartDTO{Prices: prices}
}
