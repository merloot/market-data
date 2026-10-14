package marketdata

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi/httputil"
)

func Register(mux *http.ServeMux, svc MarketService, log *slog.Logger) {
	mux.HandleFunc("GET /v1/market-data", getCurrenciesInfo(svc, log))
	mux.HandleFunc("GET /v1/market-data/currencies",
		getSupportedCurrencies(svc, log))
	mux.HandleFunc("GET /v1/market-data/price-history-chart",
		getPriceHistoryChart(svc, log))
}

func getSupportedCurrencies(scv MarketService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currencies, err := scv.ListCurrencies(r.Context())
		if err != nil {
			log.Error("List currencies", "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, "Could not list currencies")
			return
		}
		dtos := make([]CurrencyDTO, 0, len(currencies))
		for _, c := range currencies {
			dtos = append(dtos, CurrencyDTO{
				CoinID:   c.CoinGecko,
				CoinName: c.CoinName,
			})
		}
		httputil.WriteJSON(w, http.StatusOK, SupportedCurrenciesDTO{Data: dtos})
	}
}

func getCurrenciesInfo(scv MarketService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeframe := r.URL.Query().Get("timeframe")
		if timeframe == "" {
			timeframe = "24h"
		}
		since := time.Now().Add(-timeframeDuration(timeframe))

		currencies, err := scv.ListCurrencies(r.Context())
		if err != nil {
			log.Error("List currencies", "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, "Could not list currencies")
			return
		}

		info, err := scv.GetMarketDataInfo(r.Context(), currencies, since)
		if err != nil {
			log.Error("MarketDataInfo", "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, "Could not market list info")
			return
		}

		dtos := make([]MarketDataInfoDTO, 0, len(info))

		for _, i := range info {
			dtos = append(dtos, MarketDataInfoDTO{
				CoinName:              i.Currency,
				LogoURL:               i.LogoURL,
				TotalSupply:           i.TotalSupply,
				Price:                 i.Price,
				PriceDiff:             i.PriceDiff,
				MarketCap:             i.MarketCap,
				MarketCapDiff:         i.MarketCapDiff,
				CirculatingSupply:     i.CirculatingSupply,
				CirculatingSupplyDiff: i.CirculatingSupplyDiff,
			})
		}
		httputil.WriteJSON(w, http.StatusOK, MarketDataInfoListDTO{Data: dtos})
	}
}

func getPriceHistoryChart(scv MarketService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := parseHistoryInput(r)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		c := market.Currency{
			CoinName: input.CoinName,
		}

		chart, err := scv.GetPriceHistory(r.Context(), c, input.From, input.To)
		if err != nil {
			log.Error("Price history", "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		httputil.WriteJSON(w, http.StatusOK, PriceHistoryDTO{
			Prices:       chart.Prices,
			MarketCaps:   chart.MarketCaps,
			TotalVolumes: chart.TotalSupply,
		})
	}
}

func timeframeDuration(tf string) time.Duration {
	switch tf {
	case "1h":
		return time.Hour
	case "24h":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	case "30d":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

type priceHistoryInput struct {
	CoinName string
	From     time.Time
	To       time.Time
}

func parseHistoryInput(r *http.Request) (priceHistoryInput, error) {
	q := r.URL.Query()

	coinName := q.Get("coinName")
	if coinName == "" {
		return priceHistoryInput{}, errors.New("Missing required params: coinName")
	}

	var from, to time.Time

	if s := q.Get("from"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return priceHistoryInput{}, fmt.Errorf("Invalid from: %q", s)
		}
		from = t
	}

	if s := q.Get("to"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return priceHistoryInput{}, fmt.Errorf("Invalid to: %q", s)
		}
		to = t
	}
	if from.IsZero() {
		to = time.Now()
		from = to.Add(24 * time.Hour)
	}

	if to.IsZero() {
		to = time.Now()
	}

	return priceHistoryInput{
		CoinName: coinName,
		From:     from,
		To:       to,
	}, nil
}
