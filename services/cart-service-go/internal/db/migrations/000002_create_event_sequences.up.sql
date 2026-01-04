CREATE TABLE IF NOT EXISTS event_sequences (
    partition_key TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
