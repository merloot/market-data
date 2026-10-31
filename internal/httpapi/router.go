package httpapi

import (
	"net/http"

	"github.com/merloot/market-data/internal/httpapi/debug"
	v1 "github.com/merloot/market-data/internal/httpapi/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(deps *Deps) http.Handler {
	root := http.NewServeMux()

	root.HandleFunc("GET /healthz", handleHealthz)
	root.Handle("GET /metrics", promhttp.Handler())
	root.Handle("/v1/", v1.NewRouter(deps.V1))
	root.Handle("/debug/", http.StripPrefix("/debug", debug.NewRouter(deps.Debug)))

	var handler http.Handler = root
	handler = Logging(deps.Log)(handler)
	handler = Recovery(deps.Log)(handler)
	handler = Metrics()(handler)

	return handler
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
