package marketservice

import "github.com/merloot/market-data/internal/domain/market"

func groupByCurrencies(currencies []market.Currency) map[string][]market.Currency {
	grouped := make(map[string][]market.Currency)
	for _, c := range currencies {
		grouped[c.Provider] = append(grouped[c.Provider], c)
	}
	return grouped
}

func groupToFind(currencies []market.CurrencyToFind) map[string][]market.CurrencyToFind {
	grouped := make(map[string][]market.CurrencyToFind)
	for _, c := range currencies {
		grouped[c.Provider] = append(grouped[c.Provider], c)
	}
	return grouped
}

func idsAndIndex(provider string, currencies []market.Currency) ([]string, map[string]market.Currency) {
	ids := make([]string, 0, len(currencies))
	index := make(map[string]market.Currency, len(currencies))

	for _, c := range currencies {
		id := idForProvider(provider, c)
		if id == "" {
			continue
		}
		ids = append(ids, id)
		index[id] = c
	}
	return ids, index
}

func idForProvider(provider string, c market.Currency) string {
	switch provider {
	case market.ProviderCoinGecko:
		return c.CoinGecko
	case market.ProviderCoinmarketcap:
		return c.CoinMarketCap
	default:
		return c.CoinName
	}
}
