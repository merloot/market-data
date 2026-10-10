package marketservice

import "github.com/merloot/market-data/internal/domain/market"

func groupByCurrencies(currencies []market.Currency) map[string][]string {
	grouped := make(map[string][]string)
	for _, c := range currencies {
		grouped[c.Provider] = append(grouped[c.Provider], c.CoinName)
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