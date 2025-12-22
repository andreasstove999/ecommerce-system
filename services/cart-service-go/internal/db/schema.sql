CREATE TABLE carts (
    id         UUID PRIMARY KEY,
    user_id    TEXT NOT NULL UNIQUE,
    total      NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    id         UUID PRIMARY KEY,
    cart_id    UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    quantity   INT NOT NULL,
    price      NUMERIC(12,2) NOT NULL
);

CREATE TABLE event_sequences (
    partition_key TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
