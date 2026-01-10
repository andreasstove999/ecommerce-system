# Cart Service (Go)

This service manages carts for users and emits `CartCheckedOut` events when a checkout completes.

## HTTP API

- `GET /health`
- `GET /api/cart/{userId}`
- `POST /api/cart/{userId}/items`
- `POST /api/cart/{userId}/checkout`

## Event publishing

- Events are emitted using the v1 envelope contract under `contracts/events/cart/CartCheckedOut.v1.enveloped.schema.json`.
- The envelope includes:
  - `eventName=CartCheckedOut`, `eventVersion=1`, `producer=cart-service`, `eventId` as a UUID, and `occurredAt` in RFC3339 UTC.
  - `correlationId` is propagated from the `X-Correlation-Id` request header, or generated when missing.
  - `causationId` is propagated from `X-Causation-Id` when provided.
  - `partitionKey` uses the cart ID (falls back to user ID) and `sequence` is incremented per partition for observability.
  - `schema` points to `contracts/events/cart/CartCheckedOut.v1.enveloped.schema.json`.

### Sequencing

- Sequences are persisted in the `event_sequences` table (`partition_key`, `last_sequence`, `updated_at`) to survive restarts.
- The publisher calls `NextSequence` inside a transaction using `INSERT ... ON CONFLICT ... DO UPDATE SET last_sequence = event_sequences.last_sequence + 1 RETURNING last_sequence` to ensure concurrent safety.

### Dual-publish toggle

- By default, the service publishes the enveloped event to the existing routing key/queue.
- Set `PUBLISH_ENVELOPED_EVENTS=false` to temporarily publish only the legacy (bare payload) shape for rollback scenarios.

## Configuration

| Name | Default | Description |
| ---- | ------- | ----------- |
| `PORT` | `8081` | HTTP port for the service. |
| `CART_DB_DSN` | _required_ | PostgreSQL DSN for the cart database. |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string. |
| `PUBLISH_ENVELOPED_EVENTS` | `true` | Publish v1 enveloped events (set to `false` for legacy payloads). |

## Database setup

The service runs embedded SQL migrations on startup (see `internal/db/migrations`). For Docker Compose, Postgres also loads `internal/db/schema.sql` on first boot; the migrations are idempotent against that schema.

## Running tests

From the service directory:

```bash
go test ./...
```
