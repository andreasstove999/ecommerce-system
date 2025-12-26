# Architecture Overview

This document describes the high-level architecture of the polyglot, event‑driven e‑commerce system. It includes the system diagram, event flows, service responsibilities, and example event contracts.

---

## System Diagram

```mermaid
flowchart LR
    subgraph Frontend
        FE[React SPA]
    end

    subgraph Gateway
        GW["API Gateway (Go)"]
    end

    subgraph Services
        CART["Cart Service (Go)"]
        ORDER["Order Service (Go)"]
        CATALOG["Catalog Service (Java)"]
        PAYMENT["Payment Service (.NET)"]
        INVENTORY["Inventory Service (Go)"]
        SHIPPING["Shipping Service (Java)"]
    end

    subgraph Infra
        MQ[RabbitMQ]
        DB_CART[(Cart DB)]
        DB_ORDER[(Order DB)]
        DB_CATALOG[(Catalog DB)]
        DB_PAYMENT[(Payment DB)]
        DB_INVENTORY[(Inventory DB)]
        DB_SHIPPING[(Shipping DB)]
    end

    FE --> GW

    GW --> CART
    GW --> CATALOG
    GW --> ORDER
    GW --> PAYMENT
    GW --> INVENTORY
    GW --> SHIPPING

    CART --- DB_CART
    ORDER --- DB_ORDER
    CATALOG --- DB_CATALOG
    PAYMENT --- DB_PAYMENT
    INVENTORY --- DB_INVENTORY
    SHIPPING --- DB_SHIPPING

    CART -->|CartCheckedOut| MQ
    MQ -->|CartCheckedOut| ORDER

    ORDER -->|OrderCreated| MQ
    MQ -->|OrderCreated| PAYMENT
    MQ -->|OrderCreated| INVENTORY

    PAYMENT -->|PaymentSucceeded / PaymentFailed| MQ
    MQ -->|PaymentSucceeded| ORDER

    INVENTORY -->|StockReserved / Depleted| MQ
    MQ -->|StockReserved| ORDER

    ORDER -->|OrderCompleted| MQ
    MQ -->|OrderCompleted| SHIPPING
```

---

## Checkout Flow (Sequence Diagram)

```mermaid
sequenceDiagram
    participant UI as React Frontend
    participant GW as API Gateway (Go)
    participant CART as Cart Service (Go)
    participant MQ as RabbitMQ
    participant ORDER as Order Service (Go)
    participant PAY as Payment Service (.NET)
    participant INV as Inventory Service (Go)
    participant SHIP as Shipping Service (Java)

    UI->>GW: POST /checkout
    GW->>CART: POST /api/cart/checkout
    CART-->>CART: Validate cart, persist checkout
    CART->>MQ: Publish CartCheckedOut

    MQ->>ORDER: CartCheckedOut
    ORDER-->>ORDER: Create order, persist
    ORDER->>MQ: Publish OrderCreated

    MQ->>PAY: OrderCreated
    MQ->>INV: OrderCreated

    PAY-->>PAY: Process payment
    alt Payment success
        PAY->>MQ: PaymentSucceeded
    else Payment failed
        PAY->>MQ: PaymentFailed
    end

    INV-->>INV: Reserve stock
    INV->>MQ: StockReserved

    MQ->>ORDER: PaymentSucceeded / StockReserved
    ORDER-->>ORDER: Mark order as completed
    ORDER->>MQ: OrderCompleted

    MQ->>SHIP: OrderCompleted
    SHIP-->>SHIP: Arrange shipment
    SHIP->>MQ: ShippingCreated
```

---

## Event Contracts & Versioning (Option A)
- Canonical JSON Schemas live in `contracts/`.
- Envelope metadata: `eventName`, `eventVersion`, `eventId`, `correlationId`, `causationId`, `producer`, `partitionKey`, `occurredAt`, optional `sequence` (telemetry only), and `payload`.
- Payloads carry domain data only (no transport metadata).
- Versioning: additive changes stay within the same major version; breaking changes require a new `eventVersion` with new schema files. Older versions remain available.
- Option A (current): correctness via idempotency + domain invariants; ordering is not required. Option B (future): ordering with buffering.

### Idempotency layers
- Transport: `eventId` is globally unique; handlers must be safe on duplicates.
- Domain: uniqueness constraints (one order per cart, one reservation per order, one shipment per order) and sticky state machines prevent duplicate side effects even if events arrive late/out-of-order.

### Correlation example (no ordering reliance)
- `CartCheckedOut` seeds `correlationId`.
- `OrderCreated` copies `correlationId`, sets `causationId` to the cart checkout.
- `PaymentSucceeded` / `PaymentFailed` and `StockReserved` propagate `correlationId`, reference `OrderCreated` via `causationId`.
- `OrderCompleted` fires once when payment succeeded **and** stock reserved, regardless of arrival order.
- `ShippingCreated` keeps the same `correlationId`.

### Using Option A safely
- Do not assume ordered delivery; treat `sequence` as observability only.
- Prefer “set state” over “apply delta.”
- Make final states terminal (e.g., payment succeeded/failed).
- Expect duplicates and late arrivals; guard with `eventId` and domain uniqueness.

### Diagrams

#### Envelope + payload composition
```mermaid
flowchart LR
    E["EventEnvelope"] -->|payload| P["Domain Payload"]
    E --> N["eventName"]
    E --> V["eventVersion"]
    E --> ID["eventId"]
    E --> CO["correlationId/causationId"]
    E --> PK["partitionKey"]
    E --> SEQ["sequence (telemetry)"]
    E --> OA["occurredAt"]
```

#### Ordering-independent dedup/idempotency
```mermaid
sequenceDiagram
    participant MQ as Broker
    participant ORD as Order Service
    MQ->>ORD: PaymentSucceeded (eventId=ps1, sequence=9)
    ORD-->>ORD: handled eventId? if yes -> no-op
    ORD-->>ORD: update payment_status to Succeeded (terminal)
    MQ->>ORD: StockReserved (eventId=sr1, sequence=3)
    ORD-->>ORD: update inventory_status to Reserved
    ORD-->>MQ: emit OrderCompleted once (payment Succeeded AND stock Reserved)
    MQ->>ORD: PaymentSucceeded (eventId=ps1, sequence=2) # late duplicate
    ORD-->>ORD: eventId already handled -> no state change
```

---

## Service Responsibilities

### Cart Service (Go)
- Manages user carts.
- Exposes REST endpoints to add/remove items.
- Publishes **CartCheckedOut** when checkout occurs.

### Order Service (Go)
- Listens to **CartCheckedOut**.
- Creates orders and persists state.
- Publishes **OrderCreated** and **OrderCompleted**.

### Payment Service (.NET)
- Listens to **OrderCreated**.
- Simulates payment processing.
- Publishes **PaymentSucceeded** or **PaymentFailed**.

### Inventory Service (Go)
- Listens to **OrderCreated**.
- Reserves stock.
- Publishes **StockReserved**.

### Shipping Service (Java)
- Listens to **OrderCompleted**.
- Simulates shipment label creation.
- Publishes **ShippingCreated**.

TODO: Catalog Service (Java)    

---

## Event Contracts (Examples)

### CartCheckedOut
```json
{
  "eventType": "CartCheckedOut",
  "cartId": "c123",
  "userId": "u42",
  "items": [
    { "productId": "p1", "quantity": 2, "price": 100.0 }
  ],
  "totalAmount": 200.0,
  "timestamp": "2025-01-01T12:00:00Z"
}
```


