package coinmarketcap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/merloot/market-data/internal/domain/market"
)

func (p *Provider) GetMarketDataList(
	ctx context.Context,
	currencies []market.Currency,
) (map[string]market.MarketData, error) {
	ids := make([]string, 0, len(currencies))
	for _, c := range currencies {
		if c.CoinMarketCap != "" {
			ids = append(ids, c.CoinMarketCap)
		}
	}
	if len(ids) == 0 {
		return map[string]market.MarketData{}, nil
	}

	url := fmt.Sprintf("%s/v2/cryptocurrency/quotes/latest?id=%s&convert=USD",
		p.baseURL, strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", market.ErrUpstream, err)
	}
	req.Header.Set("X-CMC_PRO_API_KEY", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", market.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var raw marketDataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("Decode: %w", err)
	}

	return convertMarketDataList(raw), nil
}

func convertMarketDataList(raw marketDataListResponse) map[string]market.MarketData {
	result := make(map[string]market.MarketData, len(raw.Data))
	for coinName, item := range raw.Data {
		usd, ok := item.Quote["USD"]
		if !ok {
			continue
		}
		result[coinName] = market.MarketData{
			Currency:          coinName,
			CurrentPrice:      usd.Price,
			MarketCap:         usd.MarketCap,
			CirculatingSupply: item.CirculatingSupply,
			TotalSupply:       item.TotalSupply,
			LastUpdated:       item.LastUpdated,
		}
	}
	return result
}
