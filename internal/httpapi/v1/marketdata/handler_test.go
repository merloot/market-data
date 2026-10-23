package marketdata_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi/v1/marketdata"
)

type fakeService struct {
	currencies []market.Currency
	rows       []market.MarketDataRow
	points     []market.MarketDataPricePoint
	err        error
}

func (f *fakeService) ListCurrencies(context.Context) ([]market.Currency, error) {
	return f.currencies, f.err
}
func (f *fakeService) GetMarketDataInfo(context.Context, []market.Currency, time.Time) ([]market.MarketDataRow, error) {
	return f.rows, f.err
}
func (f *fakeService) GetPriceHistory(context.Context, market.Currency, market.Timeframe, time.Time, time.Time) ([]market.MarketDataPricePoint, error) {
	return f.points, f.err
}

func TestGetCurrencies_HappyPath(t *testing.T) {
	t.Parallel()

	fake := &fakeService{
		currencies: []market.Currency{
			{CoinName: "BTC", CoinGecko: "bitcoin"},
			{CoinName: "ETH", CoinGecko: "ethereum"},
		},
	}

	mux := http.NewServeMux()
	marketdata.Register(mux, fake, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/v1/market-data/currencies", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Errorf("got %d, want 2", len(data))
	}
}

func TestGetCurrencies_Error(t *testing.T) {
	t.Parallel()

	fake := &fakeService{err: errors.New("db down")}
	mux := http.NewServeMux()
	marketdata.Register(mux, fake, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/v1/market-data/currencies", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestGetPriceHistory_MissingCoinName(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	marketdata.Register(mux, &fakeService{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/v1/market-data/price-history-chart", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
