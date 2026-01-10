# Order Service (Go)

This service consumes cart checkout events to create orders and publishes order lifecycle events via RabbitMQ.

## Environment variables

| Name | Default | Description |
| ---- | ------- | ----------- |
| `PORT` | `8082` | HTTP port for the order service |
| `ORDER_DB_DSN` | _required_ | PostgreSQL DSN for the order database. |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string. |
| `CONSUME_ENVELOPED_EVENTS` | `true` | When `true`, the consumer expects v1 enveloped events (with payload fallback). Set to `false` to process only the legacy bare payload format. |
| `PUBLISH_ENVELOPED_EVENTS` | `true` | When `true`, `OrderCreated`/`OrderCompleted` are published using the v1 envelope. Set to `false` to publish the legacy bare payloads (rollback switch). |

### Rollback / compatibility

- To rollback envelope consumption, set `CONSUME_ENVELOPED_EVENTS=false` so only legacy payloads are processed.
- To rollback envelope publishing, set `PUBLISH_ENVELOPED_EVENTS=false` so the service emits the legacy payloads while leaving the database state intact.

### Migrations
- Migrations run automatically on startup using embedded SQL files in `internal/db/migrations`.

## HTTP endpoints
- `GET /health`
- `GET /api/orders/{orderId}`
- `GET /api/users/{userId}/orders`

## Running tests

```bash
go test ./...
```

Integration tests use Testcontainers; ensure Docker is available when running them.
