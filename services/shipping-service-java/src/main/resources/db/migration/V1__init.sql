CREATE TABLE shipments (
    shipping_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    line1 VARCHAR(255),
    line2 VARCHAR(255),
    city VARCHAR(255),
    state VARCHAR(255),
    postal_code VARCHAR(255),
    country VARCHAR(255),
    shipping_method VARCHAR(255) NOT NULL,
    carrier VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_shipments_order UNIQUE (order_id)
);

CREATE INDEX idx_shipments_user_id ON shipments (user_id);
CREATE INDEX idx_shipments_created_at ON shipments (created_at);

CREATE TABLE processed_events (
    event_id UUID PRIMARY KEY,
    event_name VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE event_sequences (
    partition_key VARCHAR(200) PRIMARY KEY,
    next_sequence BIGINT NOT NULL
);
