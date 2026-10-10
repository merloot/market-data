package oracle

import (
	"context"
	"fmt"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

func (s *Service) publishUpdates(
	ctx context.Context,
	actual []market.MarketDataHistory,
) error {
	if len(actual) == 0 {
		return nil
	}

	names := make([]string, len(actual))
	for i, a := range actual {
		names[i] = a.Currency
	}

	since := time.Now().Add(-24 * time.Hour)
	rows, err := s.marketDataHistoryRepository.GetMarketDataList(ctx, names, since)
	if err != nil {
		return fmt.Errorf("Get market data list: %w", err)
	}

	rowsByName := make(map[string]market.MarketDataRow, len(rows))
	for _, r := range rows {
		rowsByName[r.Currency] = r
	}

	for _, a := range actual {
		row := rowsByName[a.Currency]

		priceChange := calcChange(row.Price, row.PriceDiff)
		marketChange := calcChange(row.MarketCap, row.MarketCapDiff)

		if err := s.events.PublishPriceUpdated(ctx, market.PriceUpdated{
			CoinName:                 row.Currency,
			Price:                    row.Price,
			PriceChangePercentage24h: priceChange,
		}); err != nil {
			s.log.Warn("Publish price", "currency", a.Currency, "err", err)
		}

		if err := s.events.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
			CoinName:                     row.Currency,
			MarketCap:                    row.MarketCap,
			MarketCapChangePercentage24h: marketChange,
		}); err != nil {
			s.log.Warn("Publish market cap", "currency", a.Currency, "err", err)
		}
	}
	return nil
}
