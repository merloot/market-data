package stream_test

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi/v1/stream"
	"github.com/merloot/market-data/internal/realtime/memory"
)

func setupSSE(t *testing.T) (http.Handler, *memory.Publisher) {
	t.Helper()

	pub, sub := memory.New()
	handler := stream.SSEHandler(sub, slog.Default())
	return handler, pub
}

func readEvent(t *testing.T, r *bufio.Reader, body io.Closer, timeout time.Duration) (string, string) {
	t.Helper()

	type result struct {
		event string
		data  string
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		var event, data string
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				ch <- result{err: err}
				return
			}

			line = strings.TrimSuffix(line, "\n")

			if line == "" {
				ch <- result{event: event, data: data}
				return
			}

			if strings.HasPrefix(line, "event: ") {
				event = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				data = strings.TrimPrefix(line, "data: ")
			}
		}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("read: %v", res.err)
		}
		return res.event, res.data
	case <-time.After(timeout):
		_ = body.Close()
		t.Fatal("timeout reading SSE event")
		return "", ""
	}
}

func TestSSEHandler_ContentType(t *testing.T) {
	t.Parallel()

	handler, _ := setupSSE(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/stream/sse", nil)
	rec := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", got)
	}
}

func TestSSEHandler_PriceUpdated(t *testing.T) {
	t.Parallel()

	handler, pub := setupSSE(t)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	reader := bufio.NewReader(resp.Body)
	event, data := readEvent(t, reader, resp.Body, 2*time.Second)

	if event != "price.updated" {
		t.Errorf("event = %q, want price.updated", event)
	}
	if !strings.Contains(data, `"CoinName":"BTC"`) {
		t.Errorf("data = %q, want BTC", data)
	}
	if !strings.Contains(data, `"Price":50000`) {
		t.Errorf("data = %q, want price 50000", data)
	}
}

func TestSSEHandler_MarketCapUpdated(t *testing.T) {
	t.Parallel()

	handler, pub := setupSSE(t)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:  "ETH",
		MarketCap: 400000000000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	reader := bufio.NewReader(resp.Body)
	event, data := readEvent(t, reader, resp.Body, 2*time.Second)

	if event != "marketcap.updated" {
		t.Errorf("event = %q, want marketcap.updated", event)
	}
	if !strings.Contains(data, `"CoinName":"ETH"`) {
		t.Errorf("data = %q, want ETH", data)
	}
}

func TestSSEHandler_MultipleEvents(t *testing.T) {
	t.Parallel()

	handler, pub := setupSSE(t)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	coins := []string{"BTC", "ETH", "SOL"}
	for _, coin := range coins {
		if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
			CoinName: coin,
			Price:    100,
		}); err != nil {
			t.Fatalf("publish %s: %v", coin, err)
		}
	}

	reader := bufio.NewReader(resp.Body)
	received := make([]string, 0, 3)

	for i := 0; i < 3; i++ {
		_, data := readEvent(t, reader, resp.Body, 2*time.Second)
		received = append(received, data)
	}

	if len(received) != 3 {
		t.Fatalf("got %d events, want 3", len(received))
	}
}

func TestSSEHandler_ContextCancel(t *testing.T) {
	t.Parallel()

	handler, _ := setupSSE(t)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	cancel()

	reader := bufio.NewReader(resp.Body)
	_, err = reader.ReadString('\n')

	if err == nil {
		time.Sleep(200 * time.Millisecond)
	}
}

func TestSSEHandler_NoFlusher(t *testing.T) {
	t.Parallel()

	_, sub := memory.New()
	handler := stream.SSEHandler(sub, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/v1/stream/sse", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Logf("status = %d (expected 200 with Recorder)", rec.Code)
	}
}
