# Payment Service (.NET 10)

This service handles **payments** for orders.

It:
- Consumes `OrderCreated.v1` events from RabbitMQ
- Creates a `Payment` record in Postgres (**idempotent by `OrderId`**)
- Simulates payment processing (deterministic)
- Publishes either `PaymentSucceeded.v1` or `PaymentFailed.v1`
- Exposes a small HTTP API to query payment status

## Message topology

RabbitMQ:
- Exchange: `domain-events` (type: `topic`)
- Queue: `payment-service`
- Binding key: `OrderCreated.v1`

Publishes:
- `PaymentSucceeded.v1`
- `PaymentFailed.v1`

## Event format

All events are JSON with a small envelope wrapper.

Envelope fields (see `src/PaymentService/Contracts/Envelope.cs`):
- `eventName`, `eventVersion`, `eventId`
- `correlationId`
- `producer`
- `partitionKey`, `sequence`
- `occurredAt`
- `schema`
- `payload`

Example (trimmed):

```json
{
  "eventName": "OrderCreated",
  "eventVersion": 1,
  "eventId": "f79c138b-4250-4ce3-82ab-bd9f4bc1f7de",
  "correlationId": "b5c1d8dd-5b86-4fde-9c8e-4f7a9c5c0bba",
  "producer": "order-service",
  "partitionKey": "6b5ab234-5cbe-4d1c-8d0d-8c3671f0d4f2",
  "sequence": 4,
  "occurredAt": "2024-05-01T12:35:10Z",
  "schema": "contracts/events/order/OrderCreated.v1.payload.schema.json",
  "payload": {
    "orderId": "6b5ab234-5cbe-4d1c-8d0d-8c3671f0d4f2",
    "userId": "0b39a1c6-76d4-4c30-9f90-2b7a5a7a8d9a",
    "totalAmount": 199.95,
    "currency": "DKK"
  }
}
```

## Payment simulation rules

The current implementation is deterministic to keep local dev + tests stable:
- `amount <= 0` → **fail** (`Amount must be > 0`)
- `amount > 5000` → **fail** (`Amount exceeds limit`)
- otherwise → **succeed**

## HTTP API

- `GET /api/payments/by-order/{orderId}` → returns a payment record (or 404)
- `GET /health` → `{ "status": "ok" }`

When running via Docker Compose the service is mapped to:
- `http://localhost:8085`

## Configuration

Defaults are in `src/PaymentService/appsettings.json`.

Common overrides via environment variables:
- `ConnectionStrings__PaymentDb`
- `RabbitMQ__Url`
- `RabbitMQ__Exchange` (empty string = default exchange)
- `RabbitMQ__Queue`
- `RabbitMQ__RoutingKeyOrderCreated`

## Run with Docker Compose

From the repo root:

```bash
docker compose -f docker/docker-compose.yml up --build rabbitmq payment-db payment-service-dotnet
```

RabbitMQ UI:
- `http://localhost:15672` (guest/guest)

## Run locally (dotnet)

1) Start dependencies:

```bash
docker compose -f docker/docker-compose.yml up -d rabbitmq payment-db
```

2) Run the service:

```bash
cd services/payment-service-dotnet/src/PaymentService
dotnet restore
dotnet run
```

If you run the service **outside Docker**, the DB host in the connection string should be `localhost` and the port should be `5436` (compose maps `5436 -> 5432`).

## Persistence notes

At startup the service uses `db.Database.EnsureCreated()` for dev convenience (no migrations required).

Idempotency is ensured by a **unique index on `OrderId`** and a pre-check before inserting a new payment.

If you want real migrations later:
- add an initial migration:
  - `dotnet ef migrations add InitialCreate`
- update database:
  - `dotnet ef database update`
- change `EnsureCreated()` to `Migrate()` in `Program.cs`

## Troubleshooting

- **No events consumed**
  - Verify the order service publishes to queue `order.created` (default exchange).
  - If you override the exchange/routing key, ensure the queue binding matches in RabbitMQ UI.

- **DB connection errors**
  - In Docker: `Host=payment-db;Port=5432`.
  - Locally: `Host=localhost;Port=5436`.

- **Duplicates / redeliveries**
  - This service ACKs messages only after it has persisted and published the outcome.
  - Duplicate `OrderCreated` messages are ignored due to the `OrderId` uniqueness.
