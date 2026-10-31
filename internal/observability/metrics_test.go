package observability_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/merloot/market-data/internal/observability"
)

func TestHTTPRequests(t *testing.T) {
	before := testutil.ToFloat64(observability.HTTPRequests.WithLabelValues("GET", "/test", "200"))

	observability.HTTPRequests.WithLabelValues("GET", "/test", "200").Inc()

	after := testutil.ToFloat64(observability.HTTPRequests.WithLabelValues("GET", "/test", "200"))
	if after-before != 1 {
		t.Errorf("counter += %v, want 1", after-before)
	}
}
