# Postman Test Setup

This folder contains a Postman environment and collection that cover all REST endpoints across the six services in the ecommerce system.

## Files
- `ecommerce-e2e.postman_collection.json` – main collection with health checks, per-service endpoint coverage, and an end-to-end happy path.
- `ecommerce-local.postman_enviroment.json` – local environment variables (base URLs, shared IDs, polling counters).
- `endpoint-inventory.md` – reference table mapping each discovered endpoint to Postman requests.

## Import & Run
1. Import both the collection and environment into Postman (Collection v2.1 compatible).
2. Select the **ecommerce-local** environment.
3. Run the **Health Checks** folder to verify all services are reachable.
4. Run the **E2E Happy Path** folder (collection runner recommended). The flow will:
   - Create a catalog product and capture `productId`.
   - Seed and verify inventory availability for that product.
   - Add an item to the cart, capture `cartId`, and checkout.
   - Poll the order service until an order is created, then fetch the order.
   - Poll inventory until the reserved stock is reflected.
   - Poll payment and shipping services for records related to the order (these may return 404 if the async pipelines are not yet wired up).

## Key Environment Variables
- Base URLs: `cartBaseUrl`, `orderBaseUrl`, `inventoryBaseUrl`, `catalogBaseUrl`, `paymentBaseUrl`, `shippingBaseUrl`, `rabbitmqManagementUrl`.
- Shared IDs: `userId`, `productId`, `cartId`, `orderId`, `paymentId`, `shipmentId`.
- Request data: `quantity`, `price`, `initialStock`, `expectedStockAfter`.
- Polling controls: `pollTry`/`pollMax` (orders), `inventoryPollTry`/`inventoryPollMax`, `paymentPollTry`/`paymentPollMax`, `shippingPollTry`/`shippingPollMax`.

## Known Limitations / TODOs
- The shipping service exposes `GET /actuator/health`; update any placeholder request to use that endpoint.
- Payment and shipping creation are event-driven; if the upstream events are not emitted, the polling steps will continue until the configured retry limit and then fail the assertion.
- If catalog seeding is disabled or you prefer a fixed product ID, set `productId` manually before running the flow.
