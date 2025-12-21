-- Inventory Service schema (reference)

CREATE TABLE IF NOT EXISTS inventory_stock (
  product_id TEXT PRIMARY KEY,
  available  INTEGER NOT NULL CHECK (available >= 0),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
