package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

const cgDefaultUrl = "https://api.Provider.com/api/v3"

type Provider struct {
	apiKey  string
	baseUrl string
	client  *http.Client
}

func New(apiKey string) *Provider {
	return NewWithBaseUrl(apiKey, cgDefaultUrl)
}

func NewWithBaseUrl(apiKey string, baseUrl string) *Provider {
	return &Provider{
		apiKey:  apiKey,
		baseUrl: baseUrl,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Provider) Provider() string {
	return market.ProviderCoinGecko
}

func (c *Provider) GetLogo(ctx context.Context, currency string) (string, error) {
	return "", nil
}

func (c *Provider) CheckCurrencyData(ctx context.Context, currencies []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
	return map[string]market.CurrencyData{}, nil
}

func (c *Provider) GetMarketDataList(ctx context.Context, currencies []string) (map[string]market.MarketData, error) {
	ids := make([]string, len(currencies))

	copy(ids, currencies)

	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s", c.baseUrl, strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Cg request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", market.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, market.ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", market.ErrUpstream, resp.StatusCode)
	}

	var raw []struct {
		ID                string    `json:"id"`
		Symbol            string    `json:"symbol"`
		CurrentPrice      float64   `json:"current_price"`
		MarketCap         float64   `json:"market_cap"`
		CirculatingSupply float64   `json:"circulating_supply"`
		TotalSupply       float64   `json:"total_supply"`
		LastUpdated       time.Time `json:"last_updated"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("Cg decode: %w", err)
	}

	result := make(map[string]market.MarketData, len(raw))
	for _, item := range raw {
		result[item.Symbol] = market.MarketData{
			CurrentPrice:      item.CurrentPrice,
			MarketCap:         item.MarketCap,
			CirculatingSupply: item.CirculatingSupply,
			TotalSupply:       item.TotalSupply,
			LastUpdated:       item.LastUpdated,
		}
	}
	return result, nil
}

func (c *Provider) GetMarketDataChart(ctx context.Context, currency string, days int) (market.MarketDataChart, error) {
	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=%d", c.baseUrl, currency, days)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return market.MarketDataChart{}, fmt.Errorf("cg request %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return market.MarketDataChart{}, fmt.Errorf("%w: %v", market.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return market.MarketDataChart{}, fmt.Errorf("%w status: %d", market.ErrUpstream, resp.StatusCode)
	}

	var raw struct {
		Prices       [][2]float64 `json:"prices"`
		MarketCaps   [][2]float64 `json:"market_caps"`
		TotalVolumes [][2]float64 `json:"total_volumes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return market.MarketDataChart{}, fmt.Errorf("Cg decode: %w", err)
	}

	return market.MarketDataChart{
		Prices:      raw.Prices,
		MarketCaps:  raw.MarketCaps,
		TotalVolume: raw.TotalVolumes,
	}, nil
}

func (c *Provider) GetMarketDataChartRange(ctx context.Context, currency string, from, to time.Time) (market.MarketDataChart, error) {
	return market.MarketDataChart{}, nil
}
