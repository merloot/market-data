package httpapi

import (
	"log/slog"
	"net/http"
)

func NewRouter(deps *Deps, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /api/v1/debug/market-data", deps.handleDebugMarketData)
	mux.HandleFunc("GET /api/v1/debug/chart", deps.handleDebugChart)

	var handler http.Handler = mux
	handler = Logging(log)(handler)
	handler = Recovery(log)(handler)
	return handler

}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
