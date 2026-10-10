package marketservice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/marketservice"
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

	svc := marketservice.NewService(cg, cmc)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "coingecko", CoinName: "btc"},
		{Provider: "coinmarketcap", CoinName: "eth"},
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

	svc := marketservice.NewService(ok, broken)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "coingecko", CoinName: "btc"},
		{Provider: "coinmarketcap", CoinName: "eth"},
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

	svc := marketservice.NewService()

	_, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{Provider: "nope", CoinName: "btc"},
	})
	if !errors.Is(err, market.ErrNotFound) {
		t.Fatalf("Error = %v, want ErrNotFound", err)
	}
}

func TestService_GetMarketDataListGroupByProvider(t *testing.T) {
	t.Parallel()

	cg := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]market.MarketData{},
	}

	svc := marketservice.NewService(cg)

	_, _ = svc.GetMarketDataList(context.Background(), []market.Currency{
		{CoinName: "btc", Provider: "coingecko"},
		{CoinName: "eth", Provider: "coingecko"},
		{CoinName: "sol", Provider: "coingecko"},
	})

	if len(cg.gotCodes) != 3 {
		t.Errorf("Provider got %d codes, want 3 ", len(cg.gotCodes))
	}
}
