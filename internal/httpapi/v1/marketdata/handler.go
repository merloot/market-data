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
		timeframeStr := r.URL.Query().Get("timeframe")
		if timeframeStr == "" {
			timeframeStr = string(market.TimeframeDay)
		}

		timeframe := market.Timeframe(timeframeStr)

		duration, err := timeframe.Duration()
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		since := time.Now().Add(-duration)
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

		points, err := scv.GetPriceHistory(r.Context(), c, input.Timeframe, input.From, input.To)
		if err != nil {
			log.Error("Price history", "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		dtos := make([]PriceHistoryItemDTO, 0, len(points))
		for _, p := range points {
			dtos = append(dtos, PriceHistoryItemDTO{Time: p.Time, Price: p.Price})
		}

		httputil.WriteJSON(w, http.StatusOK, ToPriceHistoryChartDTO(dtos))
	}
}

type priceHistoryInput struct {
	CoinName  string
	Timeframe market.Timeframe
	From      time.Time
	To        time.Time
}

func parseHistoryInput(r *http.Request) (priceHistoryInput, error) {
	q := r.URL.Query()

	coinName := q.Get("coinName")
	if coinName == "" {
		return priceHistoryInput{}, errors.New("Missing required params: coinName")
	}

	timeframeStr := q.Get("timeframe")
	if timeframeStr == "" {
		timeframeStr = string(market.TimeframeDay)
	}

	timeframe := market.Timeframe(timeframeStr)

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

	if timeframe == market.TimeframeCustom && from.IsZero() {
		return priceHistoryInput{}, errors.New("Custom timeframe requires from parameter")
	}

	if timeframe != market.TimeframeCustom {
		dur, err := timeframe.Duration()
		if err != nil {
			return priceHistoryInput{}, err
		}
		to = time.Now()
		from = to.Add(-dur)
	}

	if to.IsZero() {
		to = time.Now()
	}

	return priceHistoryInput{
		CoinName:  coinName,
		From:      from,
		Timeframe: timeframe,
		To:        to,
	}, nil
}
