CREATE TABLE shipments (
    shipping_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    line1 TEXT,
    line2 TEXT,
    city TEXT,
    state TEXT,
    postal_code TEXT,
    country TEXT,
    shipping_method TEXT NOT NULL,
    carrier TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_shipments_order UNIQUE (order_id)
);

CREATE INDEX idx_shipments_user_id ON shipments (user_id);
CREATE INDEX idx_shipments_created_at ON shipments (created_at);

CREATE TABLE processed_events (
    event_id UUID PRIMARY KEY,
    event_name TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE event_sequences (
    partition_key VARCHAR(200) PRIMARY KEY,
    next_sequence BIGINT NOT NULL
);
