package market_test

import (
	"context"
	"errors"
	"testing"

	marketdomain "github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/market"
)

func TestService_GetMarketDataList_MergesProviders(t *testing.T) {
	t.Parallel()

	cg := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]marketdomain.MarketData{"btc": {CurrentPrice: 50000}},
	}

	cmc := &fakeProvider{
		name:       "coinmarketcap",
		marketData: map[string]marketdomain.MarketData{"eth": {CurrentPrice: 3000}},
	}

	svc := market.NewService(cg, cmc)

	got, err := svc.GetMarketDataList(context.Background(), []marketdomain.Currency{
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
		marketData: map[string]marketdomain.MarketData{"btc": {CurrentPrice: 500}},
	}

	broken := &fakeProvider{
		name: "coinmarketcap",
		err:  errors.New("Api is down"),
	}

	svc := market.NewService(ok, broken)

	got, err := svc.GetMarketDataList(context.Background(), []marketdomain.Currency{
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

	svc := market.NewService()

	_, err := svc.GetMarketDataList(context.Background(), []marketdomain.Currency{
		{Provider: "nope", CoinName: "btc"},
	})
	if !errors.Is(err, marketdomain.ErrNotFound) {
		t.Fatalf("Error = %v, want ErrNotFound", err)
	}
}

func TestService_GetMarketDataListGroupByProvider(t *testing.T) {
	t.Parallel()

	cg := &fakeProvider{
		name:       "coingecko",
		marketData: map[string]marketdomain.MarketData{},
	}

	svc := market.NewService(cg)

	_, _ = svc.GetMarketDataList(context.Background(), []marketdomain.Currency{
		{CoinName: "btc", Provider: "coingecko"},
		{CoinName: "eth", Provider: "coingecko"},
		{CoinName: "sol", Provider: "coingecko"},
	})

	if len(cg.gotCodes) != 3 {
		t.Errorf("Provider got %d codes, want 3 ", len(cg.gotCodes))
	}
}
