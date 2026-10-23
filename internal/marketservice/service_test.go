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
		name:       market.ProviderCoinGecko,
		marketData: map[string]market.MarketData{"bitcoin": {CurrentPrice: 50000, Currency: "bitcoin"}},
	}

	cmc := &fakeProvider{
		name:       market.ProviderCoinmarketcap,
		marketData: map[string]market.MarketData{"1": {CurrentPrice: 3000, Currency: "1"}},
	}

	svc := marketservice.NewService(cg, cmc)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{CoinName: "btc", CoinGecko: "bitcoin", Provider: market.ProviderCoinGecko},
		{CoinName: "eth", CoinMarketCap: "1", Provider: market.ProviderCoinmarketcap},
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
		name:       market.ProviderCoinGecko,
		marketData: map[string]market.MarketData{"bitcoin": {CurrentPrice: 500, Currency: "bitcoin",},},
	}

	broken := &fakeProvider{
		name: market.ProviderCoinmarketcap,
		err:  errors.New("Api is down"),
	}

	svc := marketservice.NewService(ok, broken)

	got, err := svc.GetMarketDataList(context.Background(), []market.Currency{
		{CoinName: "BTC", CoinGecko: "bitcoin", Provider: market.ProviderCoinGecko},
		{CoinName: "ETH", CoinMarketCap: "1", Provider: market.ProviderCoinmarketcap},
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
		name:       market.ProviderCoinGecko,
		marketData: map[string]market.MarketData{},
	}

	svc := marketservice.NewService(cg)

	_, _ = svc.GetMarketDataList(context.Background(), []market.Currency{
		{CoinName: "btc", Provider: market.ProviderCoinGecko, CoinGecko: "btc"},
		{CoinName: "eth", Provider: market.ProviderCoinGecko, CoinGecko: "eth"},
		{CoinName: "sol", Provider: market.ProviderCoinGecko, CoinGecko: "sol"},
	})

	if len(cg.gotCodes) != 3 {
		t.Errorf("Provider got %d codes, want 3 ", len(cg.gotCodes))
	}
}
