package stream

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/merloot/market-data/internal/realtime"
)

func SSEHandler(sub realtime.EventSubscriber, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Error("Flusher not supported")
			http.Error(w, "Flusher not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		priceCh, priceStop, err := sub.SubscribePriceUpdated(r.Context())
		if err != nil {
			log.Error("SubscribePriceUpdated", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer priceStop()

		marketCapCh, marketCapStop, err := sub.SubscribeMarketCapUpdated(r.Context())
		if err != nil {
			log.Error("SubscribeMarketCapUpdated", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer marketCapStop()

		flusher.Flush()

		log.Info("sse client connected", "remote", r.RemoteAddr)
		defer log.Info("sse client disconnected", "remote", r.RemoteAddr)

		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-priceCh:
				writeSSE(w, flusher, "price.updated", event)
			case event := <-marketCapCh:
				writeSSE(w, flusher, "marketcap.updated", event)
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, eventName string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintf(w, "event: %s\n", eventName)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
}
