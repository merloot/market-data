package stream_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi/v1/stream"
	"github.com/merloot/market-data/internal/realtime/memory"
)

func setupWS(t *testing.T) (http.Handler, *memory.Publisher) {
	t.Helper()

	pub, sub := memory.New()
	handler := stream.WSHandler(sub, slog.Default())
	return handler, pub
}

type wsEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func readWSEvent(t *testing.T, c *websocket.Conn, timeout time.Duration) wsEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var ev wsEvent
	if err := wsjson.Read(ctx, c, &ev); err != nil {
		t.Fatalf("read ws: %v", err)
	}
	return ev
}

func dialWS(t *testing.T, srv *httptest.Server, ctx context.Context) *websocket.Conn {
	t.Helper()

	c, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	ev := readWSEvent(t, c, 2*time.Second)
	if ev.Event != "connected" {
		c.CloseNow()
		t.Fatalf("expected connected event, got %q", ev.Event)
	}

	return c
}

func TestWSHandler_PriceUpdated(t *testing.T) {
	t.Parallel()

	handler, pub := setupWS(t)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := dialWS(t, srv, ctx)
	defer c.CloseNow()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	ev := readWSEvent(t, c, 2*time.Second)

	if ev.Event != "price.updated" {
		t.Errorf("event = %q, want price.updated", ev.Event)
	}
	if !strings.Contains(string(ev.Data), `"CoinName":"BTC"`) {
		t.Errorf("data = %q, want BTC", ev.Data)
	}
	if !strings.Contains(string(ev.Data), `"Price":50000`) {
		t.Errorf("data = %q, want price 50000", ev.Data)
	}
}

func TestWSHandler_MarketCapUpdated(t *testing.T) {
	t.Parallel()

	handler, pub := setupWS(t)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := dialWS(t, srv, ctx)
	defer c.CloseNow()

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:                     "ETH",
		MarketCap:                    400000000000,
		MarketCapChangePercentage24h: 3.2,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	ev := readWSEvent(t, c, 2*time.Second)

	if ev.Event != "marketcap.updated" {
		t.Errorf("event = %q, want marketcap.updated", ev.Event)
	}
	if !strings.Contains(string(ev.Data), `"CoinName":"ETH"`) {
		t.Errorf("data = %q, want ETH", ev.Data)
	}
}

func TestWSHandler_MultipleEvents(t *testing.T) {
	t.Parallel()

	handler, pub := setupWS(t)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := dialWS(t, srv, ctx)
	defer c.CloseNow()

	coins := []string{"BTC", "ETH", "SOL"}
	for _, coin := range coins {
		if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
			CoinName: coin,
			Price:    100,
		}); err != nil {
			t.Fatalf("publish %s: %v", coin, err)
		}
	}

	for i := 0; i < 3; i++ {
		ev := readWSEvent(t, c, 2*time.Second)
		if ev.Event != "price.updated" {
			t.Errorf("event %d = %q, want price.updated", i, ev.Event)
		}
	}
}

func TestWSHandler_ContextCancel(t *testing.T) {
	t.Parallel()

	handler, _ := setupWS(t)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := dialWS(t, srv, ctx)

	c.CloseNow()

	readCtx, readCancel := context.WithTimeout(context.Background(), time.Second)
	defer readCancel()

	var ev wsEvent
	if err := wsjson.Read(readCtx, c, &ev); err == nil {
		t.Error("expected error after close")
	}
}

func TestWSHandler_NoSubscriber(t *testing.T) {
	t.Parallel()

	handler := stream.WSHandler(nil, slog.Default())
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err == nil {
		t.Error("expected dial error with nil subscriber")
	}
}
