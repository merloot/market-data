package marketdata

import (
	"log/slog"
	"net/http"
)

func Register(mux *http.ServeMux, svc MarketService, log *slog.Logger) {
	mux.HandleFunc("GET /v1/market-data", getCurrenciesInfo(svc, log))
	mux.HandleFunc("GET /v1/market-data/currencies",
		getSupportedCurrencies(svc, log))
	mux.HandleFunc("GET /v1/market-data/price-history-chart",
		getPriceHistoryChart(svc, log))
}
