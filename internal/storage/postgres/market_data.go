package postgres

import (
	"context"
	"fmt"

	"github.com/merloot/market-data/internal/storage"
)

const marketDataQuery = `
	SELECT
		c.coin_name AS currency,
		c.logo_url AS "logoUrl",
		mdh_cur.price,
		mdh_cur.market_cap AS "marketCap",
		mdh_cur.circulating_supply AS "circulatingSupply",
		mdh_cur.total_supply AS "totalSupply",
		(mdh_cur.price - COALESCE(mdh_prev.price,0)) AS "priceDiff",
		(mdh_cur.market_cap - COALESCE(mdh_prev.market_cap,0)) AS "marketCapDiff",
		(mdh_cur.circulating_supply - COALESCE(mdh_prev.circulating_supply,0)) AS "circulatingSupplyDiff"
	FROM
		currency c
	JOIN LATERAL (
		SELECT
			price,
			market_cap,
			circulating_supply,
			total_supply
		FROM
			market_data_history mdh
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
			circulating_supply,
			total_supply
		FROM
			market_data_history mdh
		WHERE
			mdh.currency = c.coin_name
			AND mdh.timestamp <= $1
		ORDER BY
			mdh.timestamp DESC
		LIMIT
			1
	 )mdh_prev ON TRUE
	WHERE c.coin_name = ANY($2)
`

func (r *Repo) GetMarketDataList(ctx context.Context, filter storage.MarketDataFilter) ([]storage.MarketDataRow, error) {
	prevTimestamp, err := storage.ResolvePrevTimestamp(filter.Period)
	if err != nil {
		return nil, fmt.Errorf("Resolve prev timestamp: %w", err)
	}

	var currencies []string
	if len(filter.Currencies) > 0 {
		currencies = filter.Currencies
	}
	rows, err := r.pool.Query(ctx, marketDataQuery, prevTimestamp, currencies)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var result []storage.MarketDataRow
	for rows.Next() {
		var row storage.MarketDataRow
		if err := rows.Scan(
			&row.Currency,
			&row.LogoURL,
			&row.Price,
			&row.MarketCap,
			&row.CirculatingSupply,
			&row.TotalSupply,
			&row.PriceDiff,
			&row.MarketCapDiff,
			&row.CirculatingSupplyDiff,
		); err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
