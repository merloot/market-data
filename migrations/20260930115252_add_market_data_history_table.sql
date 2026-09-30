-- +goose Up
SELECT 'CREATE TABLE market_data_history (
  id BIGSERIAL PRIMARY KEY,
  currency VARCHAR(255) NOT NULL,
  price NUMERIC NOT NULL,
  market_cap NUMERIC NOT NULL,
  circulation_supply NUMERIC NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL,
  total_supply NUMERIC NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mdh_currency_timestamp 
  ON market_data_history (currency, timestamp DESC);

';

-- +goose Down
SELECT 'DROP TABLE IF EXISTS market_data_history;';
