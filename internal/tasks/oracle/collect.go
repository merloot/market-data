package oracle

import (
	"context"
	"fmt"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

const timeWindowMax = 10 * time.Minute

func (s *Service) collectHistory(
	ctx context.Context,
	currencies []market.Currency,
) ([]market.MarketDataHistory, error) {
	var all []market.MarketDataHistory

	for _, c := range currencies {
		last, err := s.marketDataHistoryRepository.LastUpdated(ctx, c.CoinName)
		if err != nil {
			return nil, fmt.Errorf("Last updated %s: %w", c.CoinName, err)
		}
		var data []market.MarketDataHistory
		if last.IsZero() {
			data, err = s.fetchAllTimeRanges(ctx, c)
		} else {
			data, err = s.fetchFromLastUpdated(ctx, c, last)
		}
		if err != nil {
			return nil, err
		}
		all = append(all, data...)
	}
	return all, nil
}

func (s *Service) fetchFromLastUpdated(
	ctx context.Context,
	currency market.Currency,
	last time.Time,
) ([]market.MarketDataHistory, error) {
	if time.Since(last) <= timeWindowMax {
		return nil, nil
	}
	chart, err := s.provider.GetMarketChartRange(ctx, currency, last, time.Now())
	if err != nil {
		return nil, fmt.Errorf("Chart range %s: %w", currency.CoinName, err)
	}
	return formatChart(currency.CoinName, chart), nil
}

func (s *Service) fetchAllTimeRanges(
	ctx context.Context,
	currency market.Currency,
) ([]market.MarketDataHistory, error) {
	var all []market.MarketDataHistory
	for _, days := range []int{1, 90, 365} {
		chart, err := s.provider.GetMarketChart(ctx, currency, days)
		if err != nil {
			return nil, fmt.Errorf("Chart %s/%dd: %w", currency.CoinName, days, err)
		}
		all = append(all, formatChart(currency.CoinName, chart)...)
	}
	return all, nil
}

func (s *Service) collectActual(
	ctx context.Context,
	currencies []market.Currency,
) ([]market.MarketDataHistory, error) {
	list, err := s.provider.GetMarketDataList(ctx, currencies)
	if err != nil {
		return nil, fmt.Errorf("Get market data list: %w", err)
	}

	var history []market.MarketDataHistory
	for _, c := range currencies {
		data, ok := list[c.CoinName]
		if !ok || data.CurrentPrice == 0 {
			continue
		}

		history = append(history, market.MarketDataHistory{
			Currency:          c.CoinGecko,
			Price:             data.CurrentPrice,
			MarketCap:         data.MarketCap,
			CirculatingSupply: data.CirculatingSupply,
			TotalSupply:       data.TotalSupply,
			Timestamp:         data.LastUpdated,
		})
	}
	return history, nil
}
