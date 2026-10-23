package stream

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/merloot/market-data/internal/realtime"
)

func WSHandler(sub realtime.EventSubscriber, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sub == nil {
			http.Error(w, "subscriber not configured", http.StatusInternalServerError)
			return
		}

		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Error("ws accept", "err", err)
			return
		}
		defer c.CloseNow()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		priceCh, priceStop, err := sub.SubscribePriceUpdated(ctx)
		if err != nil {
			_ = c.Close(websocket.StatusInternalError, err.Error())
			return
		}
		defer priceStop()

		capCh, capStop, err := sub.SubscribeMarketCapUpdated(ctx)
		if err != nil {
			_ = c.Close(websocket.StatusInternalError, err.Error())
			return
		}
		defer capStop()

		var mu sync.Mutex

		if err := writeWS(ctx, &mu, c, "connected", map[string]string{"status": "ok"}); err != nil {
			return
		}

		log.Info("ws client connected", "remote", r.RemoteAddr)
		defer log.Info("ws client disconnected", "remote", r.RemoteAddr)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-priceCh:
				if err := writeWS(ctx, &mu, c, "price.updated", event); err != nil {
					return
				}
			case event := <-capCh:
				if err := writeWS(ctx, &mu, c, "marketcap.updated", event); err != nil {
					return
				}
			}
		}
	}
}

func writeWS(ctx context.Context, mu *sync.Mutex, c *websocket.Conn, eventName string, data any) error {
	mu.Lock()
	defer mu.Unlock()

	return wsjson.Write(ctx, c, map[string]any{
		"event": eventName,
		"data":  data,
	})
}
