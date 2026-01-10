# Shipping Service (Java)

## HTTP API

- `GET /api/shipping/{shippingId}`
- `GET /api/shipping/by-order/{orderId}`
- `GET /actuator/health`

## Messaging

Consumes (from exchange `ecommerce.events`):
- `order.completed.v1` (queue: `shipping-service-java.order.completed.v1`)

Publishes (to exchange `ecommerce.events`):
- `shipping.created.v1`

## Configuration

| Environment Variable | Default | Description |
| -------------------- | ------- | ----------- |
| `PORT` | `8086` | HTTP port for the service. |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string (alternative to `SPRING_RABBITMQ_ADDRESSES`). |
| `SPRING_RABBITMQ_ADDRESSES` | _none_ | Overrides RabbitMQ addresses via Spring Boot. |
| `SPRING_RABBITMQ_USERNAME` | _none_ | Optional RabbitMQ username override. |
| `SPRING_RABBITMQ_PASSWORD` | _none_ | Optional RabbitMQ password override. |
| `SPRING_DATASOURCE_URL` | `jdbc:postgresql://shipping-db:5432/shipping_db` | Database URL. |
| `SPRING_DATASOURCE_USERNAME` | `shipping_user` | Database username. |
| `SPRING_DATASOURCE_PASSWORD` | `shipping_pass` | Database password. |

## Flyway migration note

The shipping database now uses Flyway for schema management. If you previously ran the service and have the `shipping_db_data` Docker volume with Hibernate-created tables, remove it before starting the service with Flyway migrations:

```bash
docker compose down -v
```

This clears the named volume so Flyway can apply the initial migration cleanly.
