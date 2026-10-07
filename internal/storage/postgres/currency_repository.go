package postgres

import (
	"context"
	"fmt"

	"github.com/merloot/market-data/internal/domain/market"
)

type CurrencyRepository struct {
	repo *Repo
}

func NewCurrencyRepository(repo *Repo) *CurrencyRepository {
	return &CurrencyRepository{repo: repo}
}

func (r *CurrencyRepository) List(ctx context.Context) ([]market.Currency, error) {
	const query = `
		SELECT
			coin_name,
			coin_gecko,
			coinmarketcap,
			provider,
			COALESCE(logo_url, '')
		FROM
			currency
		ORDER BY 
			coin_name
	`
	rows, err := r.repo.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var out []market.Currency
	for rows.Next() {
		var c market.Currency
		if err := rows.Scan(
			&c.CoinName,
			&c.CoinGecko,
			&c.CoinMarketCap,
			&c.Provider,
			&c.LogoUrl,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CurrencyRepository) ListWithoutLogo(ctx context.Context, provider string) ([]market.Currency, error) {
	const query = `
		SELECT
			coin_name,
			coin_gecko,
			coinmarketcap,
			provider,
			''
		FROM
			currency
		WHERE 
			provider = $1 
			AND
			logo_url IS NULL
		ORDER BY
			coin_name
	`
	rows, err := r.repo.pool.Query(ctx, query, provider)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var out []market.Currency
	for rows.Next() {
		var c market.Currency
		if err := rows.Scan(
			&c.CoinName,
			&c.CoinGecko,
			&c.CoinMarketCap,
			&c.Provider,
			&c.LogoUrl,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CurrencyRepository) UpdateLogo(ctx context.Context, coinName string, logoURL string) error {
	const query = `
		UPDATE
			currency
		SET 
			logo_url = $1
		WHERE
			coin_name = $2
	`
	if _, err := r.repo.pool.Query(ctx, query, logoURL, coinName); err != nil {
		return fmt.Errorf("Update: %w", err)
	}
	return nil
}
