CREATE TABLE IF NOT EXISTS carts (
    id         UUID PRIMARY KEY,
    user_id    TEXT NOT NULL UNIQUE,
    total      NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cart_items (
    id         UUID PRIMARY KEY,
    cart_id    UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    quantity   INT NOT NULL,
    price      NUMERIC(12,2) NOT NULL
);
