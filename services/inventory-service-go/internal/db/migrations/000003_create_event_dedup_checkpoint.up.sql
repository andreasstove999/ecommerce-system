CREATE TABLE IF NOT EXISTS event_dedup_checkpoint (
  consumer_name TEXT NOT NULL,
  partition_key TEXT NOT NULL,
  last_sequence BIGINT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (consumer_name, partition_key)
);
