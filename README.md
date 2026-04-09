# Polyglot E-commerce System (Event-Driven, Microservices)

This repository contains a **polyglot, event-driven e-commerce system** designed for learning and portfolio demonstration.
It showcases how real-world distributed systems are built using microservices, asynchronous messaging, and language diversity.

- **Microservices:** Go, Java, .NET
- **Frontend:** React
- **Messaging:** RabbitMQ
- **Databases:** One PostgreSQL instance per service
- **Runtime:** Docker + Docker Compose

> **Goal:** Demonstrate realistic backend architecture similar to what is used in banking, logistics, and large-scale marketplace systems.

---

# 🔷 High-Level Architecture

- **Frontend (React)** communicates with an **API Gateway** via REST.
- The API Gateway routes requests to backend services:
  - `cart-service` (Go)
  - `order-service` (Go)
  - `catalog-service` (Java)
  - `payment-service` (.NET)
  - `inventory-service` (Go)
  - `shipping-service` (Java)

All services communicate asynchronously through **RabbitMQ**, exchanging domain events:

- `CartCheckedOut`
- `OrderCreated`
- `StockReserved`
- `StockDepleted`
- `PaymentSucceeded`
- `PaymentFailed`
- `OrderCompleted`
- `ShippingCreated`

> For architecture diagrams, see: `docs/architecture_overview.md`

---

# 🔷 Folder Structure
```
/
├── services/
│   ├── cart-service-go/
│   ├── order-service-go/
│   ├── inventory-service-go/
│   ├── catalog-service-java/
│   ├── payment-service-dotnet/
│   └── shipping-service-java/
├── docker/
├── docs/
├── contracts/
└── Postman-test-so-far/
```

## Contracts
- `contracts/events` — event schemas
- `contracts/http` — API gateway HTTP contract (OpenAPI)

---

# 🔷 Tech Stack

## Backend
- **Go:** cart, order, inventory, API gateway
- **Java (Spring Boot):** catalog, shipping
- **.NET (ASP.NET Core):** payment
- **RabbitMQ:** async communication
- **PostgreSQL:** database-per-service pattern

## Frontend
- **React (Vite or CRA)**
- Communicates with API Gateway via REST
- Frontend source is not included in this repository

---

# 🔷 Running the System

## Requirements
- Docker & Docker Compose
- (Optional) Go / Java / .NET SDKs for local dev outside Docker

## Start everything
```bash
cd docker
docker compose up --build
```

This starts:
- RabbitMQ (UI at http://localhost:15672 — guest/guest)
- All backend microservices
- API Gateway (http://localhost:8080)
- Swagger UI (http://localhost:8090)
- PostgreSQL databases, one per service
> Note: A frontend container is not included in the current compose stack.

---

# 🔷 Services Overview

| Service | Language | Purpose |
|---------|----------|----------|
| cart-service-go | Go | Cart operations, checkout triggers |
| order-service-go | Go | Creates and manages orders |
| inventory-service-go | Go | Reserves and releases stock |
| catalog-service-java | Java | Product catalog and pricing |
| shipping-service-java | Java | Shipping request creation |
| payment-service-dotnet | .NET | Payment workflow (success/failure) |

---

# 🔷 Event Consumers

| Event | Consumed By | Description |
|--------|-------------|-------------|
| `CartCheckedOut` | order-service-go | Creates an order based on cart data |
| `OrderCreated` | inventory-service-go, payment-service-dotnet | Reserves stock and attempts payment |
| `StockReserved` | order-service-go | Marks inventory as reserved |
| `StockDepleted` | — | Emitted on insufficient stock (no consumer yet) |
| `PaymentSucceeded` | order-service-go | Marks payment as succeeded |
| `PaymentFailed` | order-service-go | Marks payment as failed |
| `OrderCompleted` | shipping-service-java | Creates shipment |
| `ShippingCreated` | — | Emitted after shipment creation (no consumer yet) |

---

# 🔷 Messaging topology

All services publish domain events to the durable topic exchange `ecommerce.events` using versioned routing keys (for example `order.created.v1`).
Each consumer owns its own queue following the `<service>.<routingKey>` convention to avoid cross-service conflicts and enable fan-out delivery.
See [docs/messaging-topology.md](docs/messaging-topology.md) for full bindings per service.

---

# 🔷 Testing

A Postman collection is available under `Postman-test-so-far/ecommerce-e2e.postman_collection.json`
with the environment in `Postman-test-so-far/ecommerce-local.postman_enviroment.json`.
It includes flows such as:
- Create product
- Add to cart
- Checkout cart
- Observe async events triggering order, inventory, payment, and shipping workflows


## Coverage gates (local + CI)

Each service now includes a coverage gate script that can be run locally and in CI.

```bash
# Go services
( cd services/api-gateway-go && ./scripts/check-coverage.sh )
( cd services/cart-service-go && ./scripts/check-coverage.sh )
( cd services/inventory-service-go && ./scripts/check-coverage.sh )
( cd services/order-service-go && ./scripts/check-coverage.sh )

# Java services
( cd services/catalog-service-java && ./scripts/check-coverage.sh )
( cd services/shipping-service-java && ./scripts/check-coverage.sh )

# .NET service
( cd services/payment-service-dotnet && ./scripts/check-coverage.sh )
```

Set `COVERAGE_THRESHOLD` to override the default line-coverage gate (`100` for Go/.NET, `1.0` for Java ratio).

---

# 🔷 Possible Extensions
- Add metrics (Prometheus + Grafana)
- OpenTelemetry tracing
- Kubernetes manifests for deployment
- Authentication (JWT)
- Saga orchestration alternative

---

# 🔷 Project Goals
- Demonstrate correct microservice boundaries
- Show event-driven collaboration between services
- Implement polyglot distributed system patterns
- Full local reproducibility using Docker

---

# 🔷 Self-hosted CI (GitHub Actions)

This repository includes an optional self-hosted CI workflow at:
- `.github/workflows/tests-self-hosted.yml`

It is designed for running tests on your own Docker-based GitHub Actions runner (GitHub Free friendly), documented in:
- `.github/runner/README.md`

Important behavior:
- Jobs target `runs-on: [self-hosted, linux, x64, ecommerce]`.
- If your self-hosted runner is offline, workflow jobs will queue until it comes back online.
- PR merges are blocked only if you mark this workflow as a **required** check in GitHub branch protection.
- To keep merges possible while the runner is offline, do **not** make `tests-self-hosted` a required status check.
