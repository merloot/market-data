package metrics

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/observability"
)

type Provider struct {
	next marketservice.Provider
}

func New(next marketservice.Provider) *Provider {
	return &Provider{next: next}
}

func (p *Provider) Provider() string {
	return p.next.Provider()
}

func (p *Provider) GetLogo(ctx context.Context, currency string) (string, error) {
	start := time.Now()
	data, err := p.next.GetLogo(ctx, currency)
	p.record("GetLogo", start, err)
	return data, err
}

func (p *Provider) CheckCurrencyData(ctx context.Context, currencies []string) (map[string]market.CurrencyData, error) {
	start := time.Now()
	data, err := p.next.CheckCurrencyData(ctx, currencies)
	p.record("CheckCurrencyData", start, err)
	return data, err
}

func (p *Provider) GetMarketDataList(ctx context.Context, currencies []string) (map[string]market.MarketData, error) {
	start := time.Now()
	data, err := p.next.GetMarketDataList(ctx, currencies)
	p.record("GetMarketDataList", start, err)
	return data, err
}

func (p *Provider) GetMarketDataChart(ctx context.Context, currency string, days int) (market.MarketDataChart, error) {
	start := time.Now()
	chart, err := p.next.GetMarketDataChart(ctx, currency, days)
	p.record("GetMarketDataChart", start, err)
	return chart, err
}

func (p *Provider) GetMarketDataChartRange(ctx context.Context, currency string, from, to time.Time) (market.MarketDataChart, error) {
	start := time.Now()
	chart, err := p.next.GetMarketDataChartRange(ctx, currency, from, to)
	p.record("GetMarketDataChartRange", start, err)
	return chart, err
}

func (p *Provider) record(method string, start time.Time, err error) {
	status := "ok"
	if err != nil {
		status = "error"
	}
	provider := p.next.Provider()
	observability.ProviderRequests.WithLabelValues(provider, method, status).Inc()
	observability.ProviderDuration.WithLabelValues(provider, method).Observe(time.Since(start).Seconds())
}
