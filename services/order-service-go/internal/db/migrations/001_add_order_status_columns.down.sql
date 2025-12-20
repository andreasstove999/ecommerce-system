-- Rollback: 001_add_order_status_columns
-- Description: Remove status tracking columns and idempotency index

DROP INDEX IF EXISTS ux_orders_cart_id;

ALTER TABLE orders
DROP COLUMN IF EXISTS payment_error,
DROP COLUMN IF EXISTS stock_ok,
DROP COLUMN IF EXISTS payment_ok,
DROP COLUMN IF EXISTS status;
