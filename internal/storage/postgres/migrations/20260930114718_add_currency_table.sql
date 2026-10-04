-- +goose Up
CREATE TABLE currency (
  coin_name VARCHAR(255) PRIMARY KEY,
  coin_gecko VARCHAR(255) NOT NULL,
  coinmarketcap VARCHAR(255) NOT NULL,
  provider VARCHAR(255) NOT NULL,
  logo_url VARCHAR(255)
);

-- +goose Down
DROP TABLE IF EXISTS currency;
