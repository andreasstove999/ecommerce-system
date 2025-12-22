-- Rollback: 002_add_event_sequence_tables
-- Description: Drop producer sequence and consumer dedup checkpoints

DROP TABLE IF EXISTS event_dedup_checkpoint;
DROP TABLE IF EXISTS event_sequence;
