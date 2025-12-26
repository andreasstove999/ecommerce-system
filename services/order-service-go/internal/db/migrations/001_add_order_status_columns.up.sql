-- Migration: 001_add_order_status_columns
-- Description: Add status tracking columns and idempotency index for orders

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending',
  ADD COLUMN IF NOT EXISTS payment_ok BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS stock_ok   BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS payment_error TEXT NULL;

-- Idempotency: don't create multiple orders for same cart
CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_cart_id ON orders(cart_id);
