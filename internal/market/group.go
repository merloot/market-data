package market 

func groupByCurrencies(currencies []Currency) map[string][]string {
	grouped := make(map[string][]string)
	for _, c := range currencies {
		grouped[c.Provider]= append(grouped[c.Provider], c.Currency)
	}
	return grouped
}

func groupToFind(currencies[] CurrencyToFind) map[string][]CurrencyToFind {
	grouped := make(map[string][]CurrencyToFind)
	for _, c := range currencies {
		grouped[c.Provider] = append(grouped[c.Provider], c)
	}
	return grouped
}