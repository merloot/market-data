package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merloot/market-data/internal/domain/market"
)

type MarketDataHistoryRepository struct {
	repo *Repo
}

func NewMarketDataHistoryRepository(repo *Repo) *MarketDataHistoryRepository {
	return &MarketDataHistoryRepository{repo: repo}
}

// TODO add bulk insert
func (r *MarketDataHistoryRepository) Create(ctx context.Context, data []market.MarketDataHistory) error {
	if len(data) == 0 {
		return nil
	}
	tx, err := r.repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Begin tx: %w", err)
	}
	defer func() { tx.Rollback(ctx) }()

	const query = `
		INSERT INTO market_data_history
			(currency,price,market_cap, circulating_supply,total_supply, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (currency, date_trunc('minute',timestamp)) DO NOTHING
	`

	for _, d := range data {
		if _, err := tx.Exec(ctx, query,
			d.Currency,
			d.Price,
			d.MarketCap,
			d.CirculatingSupply,
			d.TotalSupply,
			d.Timestamp,
		); err != nil {
			return fmt.Errorf("Insert %s: %w", d.Currency, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

func (r *MarketDataHistoryRepository) LastUpdated(ctx context.Context, currency string) (time.Time, error) {
	const query = `
		SELECT
			timestamp
		FROM
			market_data_history
		WHERE 
			currency = $1
		ORDER BY 
			timestamp DESC
		LIMIT
			1
	`
	var ts time.Time
	err := r.repo.pool.QueryRow(ctx, query, currency).Scan(&ts)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("Query: %w", err)
	}
	return ts, nil
}

func (r *MarketDataHistoryRepository) GetMarketDataList(
	ctx context.Context,
	currencies []string,
	since time.Time,
) ([]market.MarketDataRow, error) {
	const query = `
		SELECT
			c.coin_name AS currency,
			c.logo_url AS "logoUrl",
			mdh_cur.price AS price,
			mdh_cur.market_cap AS "marketCap",
			mdh_cur.circulating_supply AS "circulatingSupply",
			mdh_cur.total_supply AS "totalSupply",
			(mdh_cur.price - COALESCE(mdh_prev.price,0)) AS "priceDiff",
			(mdh_cur.market_cap - COALESCE(mdh_prev.market_cap,0)) AS "marketCapDiff",
			(mdh_cur.circulating_supply - COALESCE(mdh_prev.circulating_supply,0)) "circulatingSupplyDiff" 
		FROM
			currency c
		LEFT JOIN LATERAL (
			SELECT
				price,
				market_cap,
				circulating_supply,
				total_supply
			FROM
				market_data_history
			WHERE
				mdh.currency = c.coin_name
			ORDER BY
				mdh.timestamp DESC
			LIMIT
				1
		) mdh_cur ON TRUE
		LEFT JOIN LATERAL (
			SELECT
				price,
				market_cap,
				circulating_supply
			FROM
				market_data_history mdh
			WHERE mdh.currency = c.coin_name
			AND timestamp >= $1
			ORDER BY
				mdh.timestamp DESC
			LIMIT
				1
		) mdh_prev ON TRUE
		 WHERE c.coin_name = ANY($2)
	`
	rows, err := r.repo.pool.Query(ctx, query, since, currencies)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var out []market.MarketDataRow
	for rows.Next() {
		var row market.MarketDataRow
		if err := rows.Scan(
			&row.Currency,
			&row.Price,
			&row.MarketCap,
			&row.PriceDiff,
			&row.MarketCapDiff,
			&row.TotalSupply,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
