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
- `PaymentSucceeded`
- `PaymentFailed`
- `OrderCompleted`
- `ShippingCreated`

> For architecture diagrams, see: `docs/architecture-overview.md`

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
├── frontend/
├── docker/
├── docs/
└── postman/
```

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
- Frontend (http://localhost:3000 or 5173)
- PostgreSQL databases, one per service

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
| `OrderCreated` | inventory-service-go | Tries to reserve stock |
| `StockReserved` | payment-service-dotnet | Attempts payment |
| `PaymentSucceeded` | shipping-service-java | Issues shipping creation |
| `PaymentFailed` | order-service-go | Cancels order |
| `OrderCompleted` | frontend | Displays success to the user |

---

# 🔷 Testing

A Postman collection is available under `postman/ecommerce-collection.json`.
It includes flows such as:
- Create product
- Add to cart
- Checkout cart
- Observe async events triggering order, inventory, payment, and shipping workflows

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
- Provide strong portfolio material for backend/distributed systems engineering roles

