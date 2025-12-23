-- Migration: 002_add_event_sequence_tables
-- Description: Add producer sequence tracking and consumer dedup checkpoints

CREATE TABLE IF NOT EXISTS event_sequence (
    partition_key TEXT PRIMARY KEY,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS event_dedup_checkpoint (
    consumer_name TEXT NOT NULL,
    partition_key TEXT NOT NULL,
    last_sequence BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (consumer_name, partition_key)
);
