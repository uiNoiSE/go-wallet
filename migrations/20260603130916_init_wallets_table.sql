-- +goose Up
CREATE TABLE wallets (
  id UUID PRIMARY KEY,
  balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00
);

-- +goose Down
DROP TABLE wallets;
