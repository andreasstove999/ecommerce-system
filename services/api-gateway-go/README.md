# API Gateway (Go) — BFF

This service is the **single HTTP entry-point** for the frontend and Postman flows.

It exposes **BFF-style endpoints** (product/user-centric) and forwards requests to internal services using **HTTP clients** (no reverse proxy). The gateway is **stateless** and contains **no domain logic**.

## Frontend contract

- OpenAPI source of truth: `contracts/http/bff/v1/openapi.yaml`
- Gateway DTOs (mirror the OpenAPI): `services/api-gateway-go/internal/http/dto`

OpenAPI is the canonical contract; DTOs should reflect it.

## Responsibilities

- Provide a stable public API for clients (React / Postman)
- Route requests to internal services
- Enforce cross-cutting concerns:
  - CORS
  - correlation IDs
  - consistent error responses
- Temporary “auth” mechanism for local development via `X-User-Id` for `/me/*` routes  
  (later replace with JWT and derive user identity from token)

## Public API (Gateway Endpoints)

### Health
- `GET /health` — gateway health
- `GET /health/upstreams` — calls upstream services’ health endpoints
  - Probes: cart (`/health`), order (`/health`), inventory (`/health`), catalog (`/api/catalog/health`), payment (`/health`)
  - Shipping is not included in `/health/upstreams` yet

### Cart (current user)
> Requires header `X-User-Id`

- `GET /me/cart`
- `POST /me/cart/items`
- `POST /me/cart/checkout`

### Products (Catalog)
- `GET /products`
- `GET /products/{id}`
- `POST /products` *(admin later; currently open)*

### Orders
> `/me/orders` requires `X-User-Id`

- `GET /me/orders`
- `GET /orders/{orderId}`

### Inventory
- `GET /products/{productId}/availability`
- `POST /inventory/adjust` *(admin-ish; keep protected later)*

### Payment / Shipping (by order)
- `GET /orders/{orderId}/payment`
- `GET /orders/{orderId}/shipping`

## Required Headers

### `X-User-Id` (required for `/me/*`)
All `/me/*` endpoints require:

- `X-User-Id: u42`

The gateway uses this value to call the underlying services that still use `{userId}` path params.

### `X-Correlation-Id` (optional)
If provided, the gateway will:
- echo it back in the response headers
- propagate it to upstream services

If not provided, the gateway generates one.

## Configuration (Environment Variables)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port gateway listens on |
| `UPSTREAM_TIMEOUT` | `10s` | Timeout for upstream HTTP calls |
| `CART_URL` | `http://cart-service-go:8081` | Cart service base URL |
| `ORDER_URL` | `http://order-service-go:8082` | Order service base URL |
| `INVENTORY_URL` | `http://inventory-service-go:8083` | Inventory service base URL |
| `CATALOG_URL` | `http://catalog-service-java:8086` | Catalog service base URL |
| `PAYMENT_URL` | `http://payment-service-dotnet:8080` | Payment service base URL |
| `SHIPPING_URL` | `http://shipping-service-java:8086` | Shipping service base URL |
| `CORS_ALLOW_ORIGINS` | `*` | Comma-separated list of allowed origins |

> Note: Defaults are meant for docker-compose network service names.
> If you run locally, set these to `http://localhost:<port>` based on your compose port mappings.

## Run with Docker Compose

From the repo’s docker folder:

```bash
docker compose up --build
```

## Tests

- Unit tests: `go test ./...`
- Integration tests (requires stack running): `go test -tags=integration ./...`
- Start integration stack: `docker compose -f ../docker/docker-compose.yml up --build -d`
  (same stack used for local development)
