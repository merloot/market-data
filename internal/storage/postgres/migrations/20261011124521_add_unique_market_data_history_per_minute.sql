-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uniq_market_data_history_minute
ON market_data_history (currency, date_trunc('minute', timestamp));


-- +goose Down
DROP INDEX IF EXISTS uniq_market_data_history_per_minute;
