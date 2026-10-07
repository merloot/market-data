package oracle

import (
	"context"
	"fmt"

	"github.com/merloot/market-data/internal/domain/market"
)

func (s *Service) fetchMissingLogos(ctx context.Context) error {
	currencies, err := s.currencyRepository.ListWithoutLogo(ctx, market.ProviderCoinGecko)
	if err != nil {
		return fmt.Errorf("List without logo: %w", err)
	}

	for _, c := range currencies {
		logo, err := s.provider.GetLogo(ctx, c)
		if err != nil {
			s.log.Warn("Get logo", "currency", c.CoinName, "err", err)
		}
		if logo == "" {
			continue
		}
		if err := s.currencyRepository.UpdateLogo(ctx, c.CoinName, logo); err != nil {
			s.log.Warn("Update logo", "currency", c.CoinName, "err", err)
		}
	}
	return nil
}
