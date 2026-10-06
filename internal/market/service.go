package market

import (
	"context"
	"fmt"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type Service struct {
	impls map[string]Provider
}

func NewService(impls ...Provider) *Service {
	m := make(map[string]Provider, len(impls))
	for _, impl := range impls {
		m[impl.Provider()] = impl
	}
	return &Service{impls: m}
}

func (s *Service) impl(provider string) (Provider, error) {
	p, ok := s.impls[provider]
	if !ok {
		return nil, fmt.Errorf("%w:, %q", market.ErrNotFound, provider)
	}
	return p, nil
}

func (s *Service) GetLogo(ctx context.Context, c market.Currency) (string, error) {
	p, err := s.impl(c.Provider)
	if err != nil {
		return "", err
	}
	return p.GetLogo(ctx, c.CoinName)
}

func (s *Service) CheckCurrencyData(ctx context.Context, currencies []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
	grouped := groupToFind(currencies)
	return fanOut(ctx, grouped, func(ctx context.Context, provider string, items []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
		p, err := s.impl(provider)

		if err != nil {
			return nil, err
		}
		return p.CheckCurrencyData(ctx, items)
	})
}

func (s *Service) GetMarketDataList(ctx context.Context, currencies []market.Currency) (map[string]market.MarketData, error) {
	grouped := groupByCurrencies(currencies)
	return fanOut(ctx, grouped, func(ctx context.Context, provider string, codes []string) (map[string]market.MarketData, error) {
		p, err := s.impl(provider)
		if err != nil {
			return nil, err
		}
		return p.GetMarketDataList(ctx, codes)
	})
}

func (s *Service) GetMarketDataChart(ctx context.Context, currency market.Currency, days int) (market.MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return market.MarketDataChart{}, err
	}

	return p.GetMarketDataChart(ctx, currency.CoinName, days)
}

func (s *Service) GetMarketDataChartRange(ctx context.Context, currency market.Currency, from, to time.Time) (market.MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return market.MarketDataChart{}, err
	}
	return p.GetMarketDataChartRange(ctx, currency.CoinName, from, to)
}
