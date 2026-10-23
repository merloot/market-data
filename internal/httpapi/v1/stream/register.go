package stream

import (
	"log/slog"
	"net/http"

	"github.com/merloot/market-data/internal/realtime"
)

func Register(mux *http.ServeMux, sub realtime.EventSubscriber, log *slog.Logger) {
	mux.HandleFunc("GET /v1/stream/sse", SSEHandler(sub, log))
	mux.HandleFunc("GET /v1/stream/ws", WSHandler(sub, log))
}
