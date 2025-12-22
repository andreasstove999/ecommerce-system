# Cart Service (Go)

This service manages carts for users and emits `CartCheckedOut` events when a checkout completes.

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

## Running tests

From the service directory:

```bash
go test ./...
```
