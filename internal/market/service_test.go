package market_test

import (
	"context"
	"errors"
	"testing"

	"github.com/merloot/market-data/internal/market"
)

func TestService_GetMarketDataList_MergesProviders(t *testing.T) {
	t.Parallel()

	cg := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]market.MarketData{"btc": {CurrentPrice: 50000}},
	}

	cmc := &fakeProvider{
		name:       "coinmarketcap",
		marketData: map[string]market.MarketData{"eth": {CurrentPrice: 3000}},
	}

	svc := market.NewService(cg, cmc)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "coingecko", Currency: "btc"},
		{Provider: "coinmarketcap", Currency: "eth"},
	})

	if err != nil {
		t.Fatalf("error =%v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	if got["btc"].CurrentPrice != 50000 {
		t.Fatalf("btc price = %v, want 50000", got["btc"].CurrentPrice)
	}
}

func TestService_GetMarketDataList_PartialFailure(t *testing.T) {
	t.Parallel()

	ok := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]market.MarketData{"btc": {CurrentPrice: 500}},
	}

	broken := &fakeProvider{
		name: "coinmarketcap",
		err:  errors.New("Api is down"),
	}

	svc := market.NewService(ok, broken)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "coingecko", Currency: "btc"},
		{Provider: "coinmarketcap", Currency: "eth"},
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if len(got) != 1 {
		t.Errorf("got %d items want 1 (partial)", len(got))
	}
}

func TestService_GetMarketDataList_UnknownProvider(t *testing.T) {
	t.Parallel()

	svc := market.NewService()

	_, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "nope", Currency: "btc"},
	})
	if !errors.Is(err, market.ErrProviderNotFound) {
		t.Fatalf("Error = %v, want ErrProviderNotFound", err)
	}
}

func TestService_GetMarketDataListGroupByProvider(t *testing.T) {
	t.Parallel()

	cg := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]market.MarketData{},
	}

	svc := market.NewService(cg)

	_, _ = svc.GetMarketDataList(context.Background(), []market.Currency{
		{Currency: "btc", Provider: "coingecko"},
		{Currency: "eth", Provider: "coingecko"},
		{Currency: "sol", Provider: "coingecko"},
	})

	if len(cg.gotCodes) != 3 {
		t.Errorf("Provider got %d codes, want 3 ", len(cg.gotCodes))
	}
}
