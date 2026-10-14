package v1

import (
	"log/slog"
	"net/http"

	"github.com/merloot/market-data/internal/httpapi/v1/marketdata"
)

type Deps struct {
	MarketService marketdata.MarketService
	Log           *slog.Logger
}

func NewRouter(deps *Deps) http.Handler {
	mux := http.NewServeMux()

	marketdata.Register(mux, deps.MarketService, deps.Log)
	return mux
}
