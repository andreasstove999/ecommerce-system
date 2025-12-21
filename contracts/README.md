# Event Contracts

This directory defines canonical, versioned event contracts shared across all services. Contracts are expressed as JSON Schema (draft 2020-12) so producers and consumers can validate messages consistently and evolve independently.

## Purpose
- Provide a single source of truth for event envelopes and payloads.
- Enable validation in CI and locally using JSON Schema.
- Guide additive, versioned evolution of events without breaking existing consumers.

## Structure

```
contracts/
  README.md
  events/
    envelope/                  # Shared event envelope
    cart/                      # Cart domain events
    order/                     # Order domain events
    payment/                   # Payment domain events
    inventory/                 # Inventory domain events
    shipping/                  # Shipping domain events
  examples/                    # Complete enveloped event examples
```

Each event has:
- A **payload schema** describing the event-specific body.
- An **enveloped schema** that references the shared envelope, constrains envelope metadata (name, version, producer, partitionKey semantics), and references the payload schema.
- A corresponding **example** under `contracts/examples/<service>/<EventName>.v1.json` showing a complete envelope.

## Envelope fields (Option A: no ordering guarantees)

The shared envelope lives at `events/envelope/EventEnvelope.v1.schema.json` and applies to every event:
- `eventName` (string) — logical event name, e.g., `OrderCreated`.
- `eventVersion` (integer) — contract version for the event.
- `eventId` (UUID string) — globally unique per emission (primary idempotency key).
- `correlationId` (UUID string, optional) — links a request/flow across services.
- `causationId` (UUID string, optional) — identifies the triggering event.
- `producer` (string) — service emitting the event.
- `partitionKey` (string) — key for correlating events (e.g., `orderId`); **not** relied upon for ordering.
- `sequence` (integer, optional) — publisher-provided for observability only; **consumers must not drop events because of sequence gaps or ordering**.
- `occurredAt` (RFC3339 timestamp) — when the event happened.
- `schema` (string) — path to the payload schema.
- `payload` (object) — event-specific body (domain data only; transport metadata stays in the envelope).

### Envelope example
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
    "cartId": "7d8e9f10-1112-1314-1516-171819202122",
    "userId": "1a2b3c4d-5e6f-7081-920a-bc0d1e2f3a4b",
    "items": [
      { "productId": "9a8b7c6d-5e4f-3a2b-1c0d-9e8f7a6b5c4d", "quantity": 2, "price": 49.99 }
    ],
    "totalAmount": 99.98,
    "timestamp": "2024-05-01T12:35:10Z"
  }
}
```

## Option A: Deduplication without ordering dependence
- Transport idempotency: `eventId` is unique; handling must be safe to run multiple times.
- Domain idempotency: services apply state machines and uniqueness constraints (one order per cart, one reservation per order, one shipment per order, etc.).
- `sequence` is **non-authoritative** and used only for debugging/telemetry. Late or out-of-order events must still be evaluated against the domain state machine.
- Consumers should persist processed `eventId`s per partition key for replay safety, but must re-evaluate against sticky state to prevent double side effects.

## Versioning rules

- Schemas are immutable once published.
- Prefer **additive** changes (adding optional fields) to avoid breaking consumers.
- Breaking changes require bumping `eventVersion` and creating new files (e.g., `OrderCreated.v2.payload.schema.json` and `OrderCreated.v2.enveloped.schema.json`).
- Keep older versions alongside new ones so consumers upgrade on their own timeline.

### Versioning example
- v1: `OrderCreated.v1` includes `orderId`, `userId`, `items`, `totalAmount`.
- v2 (additive): add optional `discountCode`.
- v3 (breaking): change a field type or remove a field — requires `eventVersion: 3` and new schema files.

## Validating locally

Use any JSON Schema validator. Example with Python:
```bash
pip install jsonschema
python - <<'PY'
import json, pathlib
from jsonschema import validate

base = pathlib.Path("contracts")
envelope = json.loads((base / "events/envelope/EventEnvelope.v1.schema.json").read_text())
example = json.loads((base / "examples/order/OrderCreated.v1.json").read_text())

validate(instance=example, schema=envelope)
print("Envelope is valid")
PY
```

## Event catalog

| Domain | Event | Enveloped schema | Payload schema |
| ------ | ----- | ---------------- | -------------- |
| envelope | EventEnvelope.v1 | `events/envelope/EventEnvelope.v1.schema.json` | — |
| cart | CartCheckedOut.v1 | `events/cart/CartCheckedOut.v1.enveloped.schema.json` | `events/cart/CartCheckedOut.v1.payload.schema.json` |
| order | OrderCreated.v1 | `events/order/OrderCreated.v1.enveloped.schema.json` | `events/order/OrderCreated.v1.payload.schema.json` |
| order | OrderCompleted.v1 | `events/order/OrderCompleted.v1.enveloped.schema.json` | `events/order/OrderCompleted.v1.payload.schema.json` |
| payment | PaymentSucceeded.v1 | `events/payment/PaymentSucceeded.v1.enveloped.schema.json` | `events/payment/PaymentSucceeded.v1.payload.schema.json` |
| payment | PaymentFailed.v1 | `events/payment/PaymentFailed.v1.enveloped.schema.json` | `events/payment/PaymentFailed.v1.payload.schema.json` |
| inventory | StockReserved.v1 | `events/inventory/StockReserved.v1.enveloped.schema.json` | `events/inventory/StockReserved.v1.payload.schema.json` |
| inventory | StockDepleted.v1 | `events/inventory/StockDepleted.v1.enveloped.schema.json` | `events/inventory/StockDepleted.v1.payload.schema.json` |
| shipping | ShippingCreated.v1 | `events/shipping/ShippingCreated.v1.enveloped.schema.json` | `events/shipping/ShippingCreated.v1.payload.schema.json` |

Refer to the examples directory for end-to-end message samples.

## Using Option A safely (no ordering guarantees)
- Never assume events arrive in order; late arrivals are expected.
- Never apply arithmetic deltas blindly; prefer setting target state.
- Make final states sticky (e.g., payment succeeded/failed is terminal).
- Expect duplicate publishes; guard with `eventId` and domain uniqueness rules.
- Expect retries/redelivery; handlers must be idempotent.
