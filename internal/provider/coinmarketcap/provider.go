package coinmarketcap

import (
	"context"
	"net/http"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/marketservice"
)

var _ marketservice.Provider = (*Provider)(nil)

const (
	providerName   = "coinmarketcap"
	defaultBaseURL = "https://pro-api.coinmarketcap.com"
)

type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func New(apiKey string) *Provider {
	return NewWithBaseURL(apiKey, defaultBaseURL)
}

func NewWithBaseURL(apiKey, baseURL string) *Provider {
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Provider) Provider() string {
	return providerName
}

func (c *Provider) GetLogo(ctx context.Context, currency string) (string, error) {
	return "", nil
}

func (c *Provider) CheckCurrencyData(ctx context.Context, currencies []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
	return map[string]market.CurrencyData{}, nil
}

func ParseTimestampForTest(s string) int64 { return parseTimestamp(s) }
