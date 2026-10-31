package metrics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/observability"
	providermetrics "github.com/merloot/market-data/internal/provider/metrics"
)

type fakeProvider struct {
	name string
	err  error

	data  map[string]market.MarketData
	chart market.MarketDataChart
	logo  string
}

func (f *fakeProvider) Provider() string { return f.name }

func (f *fakeProvider) GetMarketDataList(context.Context, []string) (map[string]market.MarketData, error) {
	return f.data, f.err
}

func (f *fakeProvider) GetMarketDataChart(context.Context, string, int) (market.MarketDataChart, error) {
	return f.chart, f.err
}

func (f *fakeProvider) GetMarketDataChartRange(context.Context, string, time.Time, time.Time) (market.MarketDataChart, error) {
	return f.chart, f.err
}

func (f *fakeProvider) GetLogo(context.Context, string) (string, error) {
	return f.logo, f.err
}

func (f *fakeProvider) CheckCurrencyData(context.Context, []string) (map[string]market.CurrencyData, error) {
	return nil, f.err
}

func TestProvider_GetMarketDataList_Ok(t *testing.T) {
	t.Parallel()

	fake := &fakeProvider{
		name: "test-list-ok",
		data: map[string]market.MarketData{"bitcoin": {CurrentPrice: 50000}},
	}
	wrapped := providermetrics.New(fake)

	before := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-list-ok", "GetMarketDataList", "ok",
		),
	)

	if _, err := wrapped.GetMarketDataList(context.Background(), []string{"bitcoin"}); err != nil {
		t.Fatalf("err = %v", err)
	}

	after := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-list-ok", "GetMarketDataList", "ok",
		),
	)
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestProvider_GetMarketDataList_Error(t *testing.T) {
	t.Parallel()

	fake := &fakeProvider{
		name: "test-list-err",
		err:  errors.New("boom"),
	}
	wrapped := providermetrics.New(fake)

	before := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-list-err", "GetMarketDataList", "error",
		),
	)

	if _, err := wrapped.GetMarketDataList(context.Background(), []string{"bitcoin"}); err == nil {
		t.Fatal("expected error")
	}

	after := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-list-err", "GetMarketDataList", "error",
		),
	)
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestProvider_GetMarketDataChart_Ok(t *testing.T) {
	t.Parallel()

	fake := &fakeProvider{
		name:  "test-chart-ok",
		chart: market.MarketDataChart{Prices: [][2]float64{{1, 50000}}},
	}
	wrapped := providermetrics.New(fake)

	before := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-chart-ok", "GetMarketDataChart", "ok",
		),
	)

	if _, err := wrapped.GetMarketDataChart(context.Background(), "bitcoin", 1); err != nil {
		t.Fatalf("err = %v", err)
	}

	after := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-chart-ok", "GetMarketDataChart", "ok",
		),
	)
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestProvider_GetLogo_Ok(t *testing.T) {
	t.Parallel()

	fake := &fakeProvider{name: "test-logo-ok", logo: "https://..."}
	wrapped := providermetrics.New(fake)

	before := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-logo-ok", "GetLogo", "ok",
		),
	)

	if _, err := wrapped.GetLogo(context.Background(), "bitcoin"); err != nil {
		t.Fatalf("err = %v", err)
	}

	after := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-logo-ok", "GetLogo", "ok",
		),
	)
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestProvider_CheckCurrencyData_Ok(t *testing.T) {
	t.Parallel()

	fake := &fakeProvider{name: "test-check-ok"}
	wrapped := providermetrics.New(fake)

	before := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-check-ok", "CheckCurrencyData", "ok",
		),
	)

	if _, err := wrapped.CheckCurrencyData(context.Background(), nil); err != nil {
		t.Fatalf("err = %v", err)
	}

	after := testutil.ToFloat64(
		observability.ProviderRequests.WithLabelValues(
			"test-check-ok", "CheckCurrencyData", "ok",
		),
	)
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestProvider_ImplementsInterface(t *testing.T) {
	t.Parallel()

	var _ marketservice.Provider = providermetrics.New(&fakeProvider{name: "test"})
}
