# inventory-service-go

Inventory service that consumes `OrderCreated` events, reserves stock, and emits either `StockReserved` or `StockDepleted`.

## HTTP API

- `GET /health`
- `GET /api/inventory/{productId}`
- `POST /api/inventory/adjust`

## Event contracts

- Consumes `OrderCreated` v1 envelope from `order-service`.
- Emits `StockReserved` / `StockDepleted` using the v1 enveloped contracts in `contracts/events/inventory/`.
- Correlation IDs from the incoming `OrderCreated` are propagated to outgoing events; the incoming event ID is used as `causationId`. A new correlation ID is generated when missing from legacy payloads.
- Partitioning uses `orderId` with a producer-side sequence persisted in the `event_sequence` table.

## Deduplication

- Consumer tracks checkpoints per `(consumer_name, partition_key)` in `event_dedup_checkpoint`.
- Messages with a sequence lower than or equal to the checkpoint are ignored; gaps are logged and still processed.
- Dedup checkpoint updates are part of the same transaction as stock reservation so failed reservations do not advance checkpoints.

## Environment flags

- `CONSUME_ENVELOPED_EVENTS` (default: `true`): When `false`, only the legacy non-enveloped payload is expected.
- `PUBLISH_ENVELOPED_EVENTS` (default: `true`): When `false`, legacy payloads are published instead of the v1 envelopes.

Use these flags for rollout/rollback without redeploying code.

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | HTTP listen address |
| `DATABASE_DSN` | `postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable` | PostgreSQL DSN |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string |
| `RUN_MIGRATIONS` | `true` | Run embedded migrations on startup |
| `CONSUME_ENVELOPED_EVENTS` | `true` | Expect v1 enveloped events (`false` = legacy payloads only) |
| `PUBLISH_ENVELOPED_EVENTS` | `true` | Publish v1 enveloped events (`false` = legacy payloads only) |

### Migrations
- Migrations run from embedded SQL files in `internal/db/migrations` when `RUN_MIGRATIONS=true`.

## HTTP endpoints
- `GET /health`
- `GET /api/inventory/{productId}`
- `POST /api/inventory/adjust`
