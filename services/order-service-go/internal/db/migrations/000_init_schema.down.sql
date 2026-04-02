-- Rollback: 000_init_schema
-- Description: Drop all initial tables

DROP TABLE IF EXISTS event_dedup_checkpoint CASCADE;
DROP TABLE IF EXISTS event_sequence CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
