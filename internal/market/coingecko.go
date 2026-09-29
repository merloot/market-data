package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const cgDefaultUrl = "https://api.coingecko.com/api/v3"

type CoinGecko struct {
	apiKey  string
	baseUrl string
	client  *http.Client
}

func NewCoinGecko(apiKey string) *CoinGecko {
	return NewCoinGeckoWithBaseUrl(apiKey, cgDefaultUrl)
}

func NewCoinGeckoWithBaseUrl(apiKey string, baseUrl string) *CoinGecko {
	return &CoinGecko{
		apiKey:  apiKey,
		baseUrl: baseUrl,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CoinGecko) Provider() string {
	return "coingecko"
}

func (c *CoinGecko) GetLogo(ctx context.Context, currency string) (string, error) {
	return "", nil
}

func (c *CoinGecko) CheckCurrencyData(ctx context.Context, currencies []CurrencyToFind) (map[string]CurrencyData, error) {
	return map[string]CurrencyData{}, nil
}

func (c *CoinGecko) GetMarketDataList(ctx context.Context, currencies []string) (map[string]MarketData, error) {
	ids := make([]string, len(currencies))

	copy(ids,currencies)

	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s", c.baseUrl, strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Cg request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUpstream, resp.StatusCode)
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

	result := make(map[string]MarketData, len(raw))
	for _, item := range raw {
		result[item.Symbol] = MarketData{
			CurrentPrice:     item.CurrentPrice,
			MarketCap:         item.MarketCap,
			CirculatingSupply: item.CirculatingSupply,
			TotalSupply:       item.TotalSupply,
			LastUpdated:       item.LastUpdated,
		}
	}
	return result, nil
}

func (c *CoinGecko) GetMarketDataChart(ctx context.Context, currency string, days int) (MarketDataChart, error) {
	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=%d", c.baseUrl, currency, days)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) 
	if err != nil {
		return MarketDataChart{}, fmt.Errorf("cg request %w", err)
	}

	resp, err := c.client.Do(req) 
	if err != nil {
		return MarketDataChart{}, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MarketDataChart{}, fmt.Errorf("%w status: %d", ErrUpstream, resp.StatusCode)
	}

	var raw struct {
		Prices [][2] float64 `json:"prices"`
		MarketCaps [][2] float64 `json:"market_caps"`
		TotalVolumes [][2] float64 `json:total_volumes`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return MarketDataChart{}, fmt.Errorf("Cg decode: %w", err)
	}

	return MarketDataChart{
		Prices: raw.Prices,
		MarketCaps: raw.MarketCaps,
		TotalVolume: raw.TotalVolumes,
	}, nil
}

func (c *CoinGecko) GetMarketDataChartRange(ctx context.Context, currency string, from, to time.Time) (MarketDataChart, error) {
	return MarketDataChart{}, nil
}