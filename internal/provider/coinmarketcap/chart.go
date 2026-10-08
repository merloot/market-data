package coinmarketcap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

func (p *Provider) GetMarketDataChart(
	ctx context.Context,
	c market.Currency,
	days int,
) (market.MarketDataChart, error) {
	to := time.Now()
	from := to.Add(-time.Duration(days) * 24 * time.Hour)
	return p.getHistorical(ctx, c, from, to)
}

func (p *Provider) GetMarketDataChartRange(
	ctx context.Context,
	c market.Currency,
	from, to time.Time,
) (market.MarketDataChart, error) {
	return p.getHistorical(ctx, c, from, to)
}

func (p *Provider) getHistorical(
	ctx context.Context,
	c market.Currency,
	from, to time.Time,
) (market.MarketDataChart, error) {
	if c.CoinMarketCap == "" {
		return market.MarketDataChart{}, fmt.Errorf("%w: no coinmarketcap id for %s",
			market.ErrNotFound, c.CoinName)
	}

	interval := intervalForRange(from, to)
	url := fmt.Sprintf(
		"%s/v2/cryptocurrency/quotes/historical?id=%s&time_start=%d&time_end=%d&interval=%s&convert=USD",
		p.baseURL,
		c.CoinMarketCap,
		from.UnixMilli(),
		to.UnixMilli(),
		interval,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return market.MarketDataChart{}, fmt.Errorf("Request: %w", err)
	}
	req.Header.Set("X-CMC_PRO_API_KEY", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return market.MarketDataChart{}, fmt.Errorf("%w: %v", market.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return market.MarketDataChart{}, err
	}

	var raw historicalResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return market.MarketDataChart{}, fmt.Errorf("Decode: %w",err)
	}

	return convertHistorical(raw),nil
}

func convertHistorical(raw historicalResponse) market.MarketDataChart {
	chart := market.MarketDataChart{
		Prices: make([][2]float64,0, len(raw.Data.Quotes)),
		MarketCaps: make([][2]float64,0, len(raw.Data.Quotes)),
		TotalVolumes: make([][2]float64,0, len(raw.Data.Quotes)),
	}

	for _, q := range raw.Data.Quotes {
		usd, ok := q.Quote["USD"]
		if !ok  {
			continue
		}

		ts := parseTimestamp(usd.Timestamp)
		if ts == 0 {
			continue
		}

		chart.Prices = append(chart.Prices, [2]float64{float64(ts), usd.Price})
		chart.MarketCaps = append(chart.MarketCaps, [2]float64{float64(ts), usd.MarketCap})
		chart.TotalVolumes = append(chart.TotalVolumes, [2]float64{float64(ts), usd.TotalSupply})
	}

	return chart
}

func parseTimestamp(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}
