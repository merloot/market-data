package coinmarketcap

import (
	"context"
	"net/http"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

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

func (p *Provider) GetLogo(_ context.Context, _ market.Currency) (string, error) {
	return "", nil
}

func ParseTimestampForTest(s string) int64 { return parseTimestamp(s) }
