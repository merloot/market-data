package postgres

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

//go:embed seeddata/*.json
var seedFS embed.FS

const seedDataPath = "seeddata/currencies.json"

type SeedCurrency struct {
	CoinName      string  `json:"coin_name"`
	CoinGecko     string  `json:"coin_gecko"`
	CoinMarketCap string  `json:"coinmarketcap"`
	Provider      string  `json:"provider"`
	LogoURL       *string `json:"logo_url"`
}

type seedFile struct {
	Currencies []SeedCurrency `json:"currencies"`
}

func loadSeedData() ([]SeedCurrency, error) {
	raw, err := fs.ReadFile(seedFS, seedDataPath)
	if err != nil {
		return nil, fmt.Errorf("Read seed file: %w", err)
	}

	var file seedFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("Parse seed json: %w", err)
	}

	if len(file.Currencies) == 0 {
		return nil, fmt.Errorf("Seed file is empty")
	}

	for i, c := range file.Currencies {
		if c.CoinName == "" {
			return nil, fmt.Errorf("Currency[%d]: coin_name is required", i)
		}
		if c.Provider == "" {
			return nil, fmt.Errorf("Currency[%q]: provider is required", c.CoinName)
		}
	}

	return file.Currencies, nil
}

func (r *Repo) SeedCurrencies(ctx context.Context) error {
	currencies, err := loadSeedData()
	if err != nil {
		return fmt.Errorf("Load seed data: %w", err)
	}
	r.log.Info("Seeding currencies", "count", len(currencies))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	inserted := 0
	skipped := 0

	for _, c := range currencies {
		tag, err := tx.Exec(ctx, `
		INSERT INTO currency (coin_name, coin_gecko, coinmarketcap, provider, logo_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (coin_name) DO NOTHING
		`, c.CoinName, c.CoinGecko, c.CoinMarketCap, c.Provider, c.LogoURL)
		if err != nil {
			return fmt.Errorf("Insert currency %s: %w", c.CoinName, err)
		}

		if tag.RowsAffected() == 0 {
			skipped++
		}
		inserted++
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	r.log.Info(`Seeding done`, "inserted", inserted, "skipped", skipped)
	return nil
}
