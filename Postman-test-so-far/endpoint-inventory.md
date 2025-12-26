# Endpoint Inventory

Endpoints discovered directly from the service code. Each row lists the Postman request that exercises the route in `ecommerce-e2e.postman_collection.json`.

| Service | Method | Path | Notes | Postman Request |
| --- | --- | --- | --- | --- |
| Cart (Go) | GET | `/health` | JSON status | Health Checks / Cart Health - GET /health |
| Cart (Go) | POST | `/api/cart/{userId}/items` | Adds or updates item in cart | Cart Service / POST /api/cart/{userId}/items; E2E Happy Path / Cart AddItem - POST /api/cart/{userId}/items |
| Cart (Go) | GET | `/api/cart/{userId}` | Fetch cart by user | Cart Service / GET /api/cart/{userId}; E2E Happy Path / Cart Get - GET /api/cart/{userId} |
| Cart (Go) | POST | `/api/cart/{userId}/checkout` | Publishes checkout event, clears cart | Cart Service / POST /api/cart/{userId}/checkout; E2E Happy Path / Cart Checkout - POST /api/cart/{userId}/checkout |
| Order (Go) | GET | `/health` | JSON status | Health Checks / Order Health - GET /health |
| Order (Go) | GET | `/api/orders/{orderId}` | Get single order | Order Service / GET /api/orders/{orderId}; E2E Happy Path / Get Order - GET /api/orders/{orderId} |
| Order (Go) | GET | `/api/users/{userId}/orders` | List orders for user | Order Service / GET /api/users/{userId}/orders; E2E Happy Path / Poll Orders Until Created |
| Inventory (Go) | GET | `/health` | Plain `ok` | Health Checks / Inventory Health - GET /health |
| Inventory (Go) | GET | `/api/inventory/{productId}` | Get availability for product | Inventory Service / GET /api/inventory/{productId}; E2E Happy Path / Inventory Verify Seed; E2E Happy Path / Poll Inventory Until Reserved |
| Inventory (Go) | POST | `/api/inventory/adjust` | Set availability for product | Inventory Service / POST /api/inventory/adjust; E2E Happy Path / Inventory Seed Stock - POST /api/inventory/adjust |
| Catalog (Java) | GET | `/api/catalog/health` | JSON status | Health Checks / Catalog Health - GET /api/catalog/health |
| Catalog (Java) | GET | `/api/catalog/products` | List products (limit/offset) | Catalog Service / GET /api/catalog/products; E2E Happy Path / Catalog Create Product (pre-run uses generated product) |
| Catalog (Java) | GET | `/api/catalog/products/{id}` | Get product by UUID | Catalog Service / GET /api/catalog/products/{id} |
| Catalog (Java) | POST | `/api/catalog/products` | Create product | Catalog Service / POST /api/catalog/products; E2E Happy Path / Catalog Create Product |
| Payment (.NET) | GET | `/health` | JSON status | Health Checks / Payment Health - GET /health |
| Payment (.NET) | GET | `/api/payments/by-order/{orderId}` | Lookup payment by order | Payment Service / GET /api/payments/by-order/{orderId}; E2E Happy Path / Poll Payment Until Recorded |
| Shipping (Java) | GET | `/api/shipping/{shippingId}` | Get shipment by ID | Shipping Service / GET /api/shipping/{shippingId}; E2E Happy Path / Poll Shipping Until Ready |
| Shipping (Java) | GET | `/api/shipping/by-order/{orderId}` | Get shipment by order ID | Shipping Service / GET /api/shipping/by-order/{orderId}; E2E Happy Path / Poll Shipping Until Ready |
| Shipping (Java) | GET | `/health` (TODO) | Not implemented in code | Health Checks / Shipping Health (TODO) - no health endpoint |
