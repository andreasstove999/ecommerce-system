# Postman – Payment Service

Import both files into Postman:

1) `Payment Service.postman_collection.json`
2) `Payment Service (local docker-compose).postman_environment.json`

## Variables

- `baseUrl` defaults to `http://localhost:8085` (matches the docker-compose mapping)
- `orderId` should be set to the OrderId (GUID) for an order that has emitted an `OrderCreated` event.

## What to expect

- **Health** should return `200` and `{ "status": "ok" }`
- **Get payment by orderId** returns:
  - `200` with a Payment JSON when the consumer has processed the event
  - `404` if payment doesn't exist yet (e.g., order not created, event not processed, or wrong id)
