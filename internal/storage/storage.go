package storage

import (
	"context"
	"errors"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

var ErrNotFound = errors.New("Not found")

type Currency struct {
	CoinName      string
	CoinSlug      string
	CoinGecko     string
	Coinmarketcap string
	Provider      string
	LogoUrl       string
}

type MarketDataHistory struct {
	Currency          string
	Price             float64
	MarketCap         float64
	CirculatingSupply float64
	Timestamp         time.Time
	TotalSupply       float64
}

type CurrencyRepository interface {
	UpsertCurrency(ctx context.Context, c market.Currency) error
	GetCurrency(ctx context.Context, slug string) (Currency, error)
	GetCurrencies(ctx context.Context) ([]Currency, error)
}

type MarketDataHistoryRepository interface {
	UpsertMarketDataHistory(ctx context.Context, mdh MarketDataHistory) error
}
