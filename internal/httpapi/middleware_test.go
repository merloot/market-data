package httpapi_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/merloot/market-data/internal/httpapi"
	"github.com/merloot/market-data/internal/observability"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecovery_CatchesPanic(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	handler := httpapi.Recovery(log)(panicky)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestLogging_CallsNext(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := httpapi.Logging(log)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("next not called")
	}
}

func TestMetrics_IncrementsCounter(t *testing.T) {
	t.Parallel()

	method := "GET"
	path := "/test/metrics"
	status := "200"

	before := testutil.ToFloat64(
		observability.HTTPRequests.WithLabelValues(method, path, status),
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := httpapi.Metrics()(next)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	after := testutil.ToFloat64(
		observability.HTTPRequests.WithLabelValues(method, path, status),
	)

	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestMetrics_RecordsErrorStatus(t *testing.T) {
	t.Parallel()

	method := "GET"
	path := "/test/error"
	status := "500"

	before := testutil.ToFloat64(
		observability.HTTPRequests.WithLabelValues(method, path, status),
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := httpapi.Metrics()(next)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	after := testutil.ToFloat64(
		observability.HTTPRequests.WithLabelValues(method, path, status),
	)

	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}

func TestMetrics_RecordsDuration(t *testing.T) {
	t.Parallel()

	path := "/test/duration"

	before := testutil.CollectAndCount(
		observability.HTTPDuration.MustCurryWith(nil),
		"marketdata_http_request_duration_seconds",
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := httpapi.Metrics()(next)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	after := testutil.CollectAndCount(
		observability.HTTPDuration.MustCurryWith(nil),
		"marketdata_http_request_duration_seconds",
	)

	if after < before {
		t.Errorf("histogram count decreased: before=%d after=%d", before, after)
	}
}
