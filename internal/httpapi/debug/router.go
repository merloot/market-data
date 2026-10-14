package debug

import (
	"log/slog"
	"net/http"
)

type Deps struct {
	MarketService MarketService
	Log           *slog.Logger
}

func NewRouter(deps *Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /debug/market-data", getMarketData(deps.MarketService, deps.Log))
	mux.HandleFunc("GET /debug/chart", getChart(deps.MarketService, deps.Log))

	return mux
}
