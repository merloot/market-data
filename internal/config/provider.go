package config

type ProviderConfig struct {
	Coinmarketcap CoinMarketCapConfig
	Coingecko     CoinGeckoConfig
}

type CoinMarketCapConfig struct {
	ApiKey string `env:"COINMARKETCAP_API_KEY"`
}

type CoinGeckoConfig struct {
	ApiKey string `env:"COIN_GECKO_API_KEY"`
}
