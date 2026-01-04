# Shipping Service (Java)

## Flyway migration note

The shipping database now uses Flyway for schema management. If you previously ran the service and have the `shipping_db_data` Docker volume with Hibernate-created tables, remove it before starting the service with Flyway migrations:

```bash
docker compose down -v
```

This clears the named volume so Flyway can apply the initial migration cleanly.
