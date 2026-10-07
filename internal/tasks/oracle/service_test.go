package oracle_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/tasks/oracle"
)

type fakeCurrencyRepository struct {
	list         []market.Currency
	withoutLogo  []market.Currency
	updatedLogos map[string]string
}

func (f *fakeCurrencyRepository) List(context.Context) ([]market.Currency, error) {
	return f.list, nil
}

func (f *fakeCurrencyRepository) ListWithoutLogo(context.Context, string) ([]market.Currency, error) {
	return f.withoutLogo, nil
}

func (f *fakeCurrencyRepository) UpdateLogo(_ context.Context, coinName, logoURL string) error {
	if f.updatedLogos == nil {
		f.updatedLogos = make(map[string]string)
	}
	f.updatedLogos[coinName] = logoURL
	return nil
}

type fakeMarketDataHistoryRepository struct {
	saved []market.MarketDataHistory
	last  time.Time
	rows  []market.MarketDataRow
}

func (f *fakeMarketDataHistoryRepository) Create(
	_ context.Context,
	data []market.MarketDataHistory,
) error {
	f.saved = append(f.saved, data...)
	return nil
}

func (f *fakeMarketDataHistoryRepository) LastUpdated(
	context.Context,
	string,
) (time.Time, error) {
	return f.last, nil
}

func (f *fakeMarketDataHistoryRepository) GetMarketDataList(
	context.Context,
	[]string,
	time.Time,
) ([]market.MarketDataRow, error) {
	return f.rows, nil
}

type fakeProvider struct {
	list  map[string]market.MarketData
	chart market.MarketDataChart
	logo  string
}

func (f *fakeProvider) GetMarketDataList(
	context.Context,
	[]market.Currency,
) (map[string]market.MarketData, error) {
	return f.list, nil
}

func (f *fakeProvider) GetMarketChart(
	context.Context,
	market.Currency,
	int,
) (market.MarketDataChart, error) {
	return f.chart, nil
}

func (f *fakeProvider) GetMarketChartRange(
	context.Context,
	market.Currency,
	time.Time,
	time.Time,
) (market.MarketDataChart, error) {
	return f.chart, nil
}

func (f *fakeProvider) GetLogo(
	context.Context,
	market.Currency,
) (string, error) {
	return f.logo, nil
}

type fakeEvents struct {
	prices []market.PriceUpdated
	caps   []market.MarketCapUpdated
}

func (f *fakeEvents) PublishPriceUpdated(_ context.Context, event market.PriceUpdated) error {
	f.prices = append(f.prices, event)
	return nil
}

func (f *fakeEvents) PublishMarketCapUpdated(_ context.Context, event market.MarketCapUpdated) error {
	f.caps = append(f.caps, event)
	return nil
}

func TestService_Execute_SavesAndPublishes(t *testing.T) {
	currencies := &fakeCurrencyRepository{list: []market.Currency{
		{CoinName: "BTC", Provider: market.ProviderCoinGecko},
	}}
	history := &fakeMarketDataHistoryRepository{}
	provider := &fakeProvider{list: map[string]market.MarketData{
		"BTC": {Currency: "BTC", CurrentPrice: 50000, LastUpdated: time.Now()},
	}}
	events := &fakeEvents{}

	svc := oracle.NewService(slog.Default(), currencies, history, provider, events)

	if err := svc.Execute(context.Background()); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if len(history.saved) != 1 {
		t.Errorf("Saved = %d, want 1", len(history.saved))
	}
	if len(events.prices) != 1 {
		t.Errorf("Publish prices = %d, want 1", len(events.prices))
	}
	if len(events.caps) != 1 {
		t.Errorf("Publish caps = %d, want 1", len(events.caps))
	}
}

func TestService_Execute_NoCurrencies(t *testing.T) {
	svc := oracle.NewService(
		slog.Default(),
		&fakeCurrencyRepository{},
		&fakeMarketDataHistoryRepository{},
		&fakeProvider{},
		&fakeEvents{},
	)

	if err := svc.Execute(context.Background()); err != nil {
		t.Fatalf("Err = %v", err)
	}
}

func TestService_Execute_SkipsZeroPrice(t *testing.T) {
	currencies := &fakeCurrencyRepository{list: []market.Currency{
		{CoinName: "BTC", Provider: market.ProviderCoinGecko},
	}}

	history := &fakeMarketDataHistoryRepository{}
	provider := &fakeProvider{list: map[string]market.MarketData{
		"BTC": {Currency: "BTC", CurrentPrice: 0, LastUpdated: time.Now()},
	}}
	events := &fakeEvents{}

	svc := oracle.NewService(slog.Default(), currencies, history, provider, events)

	if err := svc.Execute(context.Background()); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if len(history.saved) != 0 {
		t.Errorf("Saved = %d want 0 (zero price skipped)", len(history.saved))
	}
}

func TestService_Execute_FetchMissingLogos(t *testing.T) {
	currencies := &fakeCurrencyRepository{
		list: []market.Currency{
			{CoinName: "BTC", Provider: market.ProviderCoinGecko},
		},
		withoutLogo: []market.Currency{
			{CoinName: "BTC", Provider: market.ProviderCoinGecko},
		},
	}

	history := &fakeMarketDataHistoryRepository{}
	provider := &fakeProvider{
		list: map[string]market.MarketData{
			"BTC": {Currency: "BTC", CurrentPrice: 50000, LastUpdated: time.Now()},
		}, logo: "https://logo/btc.png",
	}
	events := &fakeEvents{}

	scv := oracle.NewService(slog.Default(), currencies, history, provider, events)

	if err := scv.Execute(context.Background()); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if currencies.updatedLogos["BTC"] != "https://logo/btc.png" {
		t.Errorf("Logo not updated: %v", currencies.updatedLogos)
	}
}
