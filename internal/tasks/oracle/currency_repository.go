package oracle

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
)

type CurrencyRepository interface {
	List(ctx context.Context) ([]market.Currency, error)

	ListWithoutLogo(ctx context.Context, provider string) ([]market.Currency, error)

	UpdateLogo(ctx context.Context, coinName, logoUrl string) error
}
