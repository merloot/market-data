package v1

import (
	"log/slog"
	"net/http"

	"github.com/merloot/market-data/internal/httpapi/v1/marketdata"
	"github.com/merloot/market-data/internal/httpapi/v1/stream"
	"github.com/merloot/market-data/internal/realtime"
)

type Deps struct {
	MarketService marketdata.MarketService
	Subscriber    realtime.EventSubscriber
	Log           *slog.Logger
}

func NewRouter(deps *Deps) http.Handler {
	mux := http.NewServeMux()

	marketdata.Register(mux, deps.MarketService, deps.Log)

	if deps.Subscriber != nil {
		stream.Register(mux, deps.Subscriber, deps.Log)
	}
	return mux
}
