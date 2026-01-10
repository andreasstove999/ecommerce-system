# Shipping Service (Java)

## Configuration

The service uses `src/main/resources/application.yml` with environment overrides.

| Environment Variable | Description | Default |
| --- | --- | --- |
| `PORT` | HTTP port | `8086` |
| `SPRING_DATASOURCE_URL` | Database URL | `jdbc:postgresql://shipping-db:5432/shipping_db` |
| `SPRING_DATASOURCE_USERNAME` | Database username | `shipping_user` |
| `SPRING_DATASOURCE_PASSWORD` | Database password | `shipping_pass` |
| `RABBITMQ_URL` | RabbitMQ URL (used when `SPRING_RABBITMQ_ADDRESSES` is not set) | `amqp://guest:guest@rabbitmq:5672/` |
| `SPRING_RABBITMQ_ADDRESSES` | RabbitMQ addresses (overrides `RABBITMQ_URL`) | _unset_ |
| `SPRING_RABBITMQ_USERNAME` | RabbitMQ username | `guest` |
| `SPRING_RABBITMQ_PASSWORD` | RabbitMQ password | `guest` |

## Messaging topology

- Exchange: `ecommerce.events`
- Consumes: `order.completed.v1` via queue `shipping-service-java.order.completed.v1`
- Publishes: `shipping.created.v1`
- Dead-lettering:
  - DLX: `shipping-service.dlx`
  - DLQ: `shipping-service.dlq`
  - DLQ routing key: `order.completed.dlq`

## HTTP endpoints
- `GET /api/shipping/{shippingId}`
- `GET /api/shipping/by-order/{orderId}`
- `GET /actuator/health`

## Flyway migration note

The shipping database now uses Flyway for schema management. If you previously ran the service and have the `shipping_db_data` Docker volume with Hibernate-created tables, remove it before starting the service with Flyway migrations:

```bash
docker compose down -v
```

This clears the named volume so Flyway can apply the initial migration cleanly.
