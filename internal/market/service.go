package market

import (
	"context"
	"time"
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
		return nil, ErrProviderNotFound
	}
	return p, nil
}

func (s *Service) GetLogo(ctx context.Context, c Currency) (string, error) {
	p, err := s.impl(c.Provider)
	if err != nil {
		return "", err
	}
	return p.GetLogo(ctx, c.Currency)
}

func (s *Service) CheckCurrencyData(ctx context.Context, currencies []CurrencyToFind) (map[string]CurrencyData, error) {
	grouped := groupToFind(currencies)
	return fanOut(ctx, grouped, func(ctx context.Context, provider string, items []CurrencyToFind) (map[string]CurrencyData, error) {
		p, err := s.impl(provider)

		if err != nil {
			return nil, err
		}
		return p.CheckCurrencyData(ctx, items)
	})
}

func (s *Service) GetMarketDataList(ctx context.Context, currencies []Currency) (map[string]MarketData, error) {
	grouped := groupByCurrencies(currencies)
	return fanOut(ctx, grouped, func(ctx context.Context, provider string, codes []string) (map[string]MarketData, error) {
		p, err := s.impl(provider)
		if err != nil {
			return nil, err
		}
		return p.GetMarketDataList(ctx, codes)
	})
}

func (s *Service) GetMarketDataChart(ctx context.Context, currency Currency, days int) (MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return MarketDataChart{}, err
	}

	return p.GetMarketDataChart(ctx, currency.Currency, days)
}

func (s *Service) GetMarketDataChartRange(ctx context.Context, currency Currency, from, to time.Time) (MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return MarketDataChart{}, err
	}
	return p.GetMarketDataChartRange(ctx, currency.Currency, from, to)
}
