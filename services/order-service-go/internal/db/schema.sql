CREATE TABLE orders (
    id           UUID PRIMARY KEY,
    cart_id      TEXT NOT NULL,
    user_id      TEXT NOT NULL,
    total_amount NUMERIC(12,2) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id         UUID PRIMARY KEY,
    order_id   UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    quantity   INT NOT NULL,
    price      NUMERIC(12,2) NOT NULL
);
