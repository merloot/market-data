package market_test

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type fakeProvider struct {
	name       string
	marketData map[string]market.MarketData
	chart      market.MarketDataChart
	err        error
	gotCodes   []string
}

func (p *fakeProvider) Provider() string {
	return p.name
}

func (p *fakeProvider) GetLogo(context.Context, string) (string, error) {
	return "", nil
}

func (p *fakeProvider) CheckCurrencyData(context.Context, []market.CurrencyToFind) (map[string]market.CurrencyData, error) {
	return nil, nil
}

func (p *fakeProvider) GetMarketDataList(_ context.Context, codes []string) (map[string]market.MarketData, error) {
	p.gotCodes = codes
	return p.marketData, p.err
}

func (p *fakeProvider) GetMarketDataChart(context.Context, string, int) (market.MarketDataChart, error) {
	return p.chart, p.err
}

func (p *fakeProvider) GetMarketDataChartRange(context.Context, string, time.Time, time.Time) (market.MarketDataChart, error) {
	return p.chart, p.err
}
