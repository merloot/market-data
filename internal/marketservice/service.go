package marketservice

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

// TODO: implement
func (s *Service) CheckCurrencyData(ctx context.Context, currencies []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
	return nil, nil
}

func (s *Service) GetMarketDataList(
	ctx context.Context,
	currencies []market.Currency,
) (map[string]market.MarketData, error) {
	grouped := groupByCurrencies(currencies)

	return fanOut(ctx, grouped, func(ctx context.Context, provider string, items []market.Currency) (map[string]market.MarketData, error) {
		p, err := s.impl(provider)
		if err != nil {
			return nil, err
		}

		ids, idToCurrency := idsAndIndex(provider, items)
		if len(ids) == 0 {
			return map[string]market.MarketData{}, nil
		}

		byID, err := p.GetMarketDataList(ctx, ids)
		if err != nil {
			return nil, err
		}

		result := make(map[string]market.MarketData, len(byID))
		for id, data := range byID {
			c, ok := idToCurrency[id]
			if !ok {
				continue
			}
			result[c.CoinName] = data
		}
		return result, nil
	})
}

func (s *Service) GetMarketDataChart(ctx context.Context, currency market.Currency, days int) (market.MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return market.MarketDataChart{}, err
	}

	id := idForProvider(currency.Provider, currency)
	return p.GetMarketDataChart(ctx, id, days)
}

func (s *Service) GetMarketDataChartRange(ctx context.Context, currency market.Currency, from, to time.Time) (market.MarketDataChart, error) {
	p, err := s.impl(currency.Provider)
	if err != nil {
		return market.MarketDataChart{}, err
	}

	id := idForProvider(currency.Provider, currency)
	return p.GetMarketDataChartRange(ctx, id, from, to)
}
