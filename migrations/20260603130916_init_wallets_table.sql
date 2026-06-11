-- +goose Up
CREATE TABLE wallet_transactions (
  id SERIAL PRIMARY KEY,
  wallet_id UUID NOT NULL,
  operation_type VARCHAR(20) NOT NULL,
  amount NUMERIC(15, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallet_id ON wallet_transactions(wallet_id);

-- +goose Down
DROP TABLE wallet_transactions;

