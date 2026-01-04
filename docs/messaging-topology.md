# Messaging topology

All services exchange domain events through a single durable topic exchange to support fan-out and service-owned queues.

## Exchange
- **Name:** `ecommerce.events`
- **Type:** `topic`
- **Durable:** yes

## Routing keys (versioned)
- `cart.checkedout.v1`
- `order.created.v1`
- `order.completed.v1`
- `stock.reserved.v1`
- `stock.depleted.v1`
- `payment.succeeded.v1`
- `payment.failed.v1`
- `shipping.created.v1`

## Queue ownership rule
- Every consumer declares its **own** durable queue and binds it to `ecommerce.events`.
- Queue name format: `<service>.<routingKey>` (for example `shipping-service-java.order.completed.v1`).
- Publishers only declare the exchange and publish to the routing key; they never declare shared queues.

## Service bindings in this repository

| Service | Consumes (queue → routing key) | Publishes (routing key) |
|---------|--------------------------------|-------------------------|
| cart-service-go | — | `cart.checkedout.v1` |
| order-service-go | `order-service-go.cart.checkedout.v1` → `cart.checkedout.v1`<br>`order-service-go.payment.succeeded.v1` → `payment.succeeded.v1`<br>`order-service-go.payment.failed.v1` → `payment.failed.v1`<br>`order-service-go.stock.reserved.v1` → `stock.reserved.v1` | `order.created.v1`, `order.completed.v1` |
| inventory-service-go | `inventory-service-go.order.created.v1` → `order.created.v1` | `stock.reserved.v1`, `stock.depleted.v1` |
| payment-service-dotnet | `payment-service-dotnet.order.created.v1` → `order.created.v1` | `payment.succeeded.v1`, `payment.failed.v1` |
| shipping-service-java | `shipping-service-java.order.completed.v1` → `order.completed.v1` | `shipping.created.v1` |

Dead-letter queues remain service-specific (for example `order-service.dlq`, `shipping-service.dlq`) and are not shared across services.
