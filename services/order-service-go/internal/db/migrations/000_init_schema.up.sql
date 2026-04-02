CREATE TABLE orders (
    id           UUID PRIMARY KEY,
    cart_id      TEXT NOT NULL,
    user_id      TEXT NOT NULL,
    total_amount NUMERIC(12,2) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    status       TEXT NOT NULL DEFAULT 'pending',
    payment_ok   BOOLEAN NOT NULL DEFAULT false,
    stock_ok     BOOLEAN NOT NULL DEFAULT false,
    payment_error TEXT NULL
);

CREATE UNIQUE INDEX ux_orders_cart_id ON orders(cart_id);

CREATE TABLE order_items (
    id         UUID PRIMARY KEY,
    order_id   UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    quantity   INT NOT NULL,
    price      NUMERIC(12,2) NOT NULL
);

CREATE TABLE event_sequence (
    partition_key TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE event_dedup_checkpoint (
    consumer_name TEXT NOT NULL,
    partition_key TEXT NOT NULL,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (consumer_name, partition_key)
);
