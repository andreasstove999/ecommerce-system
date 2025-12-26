# Catalog Service

Catalog Service for Ecommerce System. This service manages product information.

## Prerequisites

- **Java**: 21
- **Maven**: 3.9+
- **Docker** (optional, for containerization)

## Configuration

The application is configured using `src/main/resources/application.yml` and environment variables.

| Environment Variable | Description | Default |
| -------------------- | ----------- | ------- |
| `PORT` | The port the application runs on | `8086` |
| `SPRING_DATASOURCE_URL` | Database URL | `jdbc:postgresql://localhost:5432/catalog_db` |
| `SPRING_DATASOURCE_USERNAME` | Database username | `catalog_user` |
| `SPRING_DATASOURCE_PASSWORD` | Database password | `catalog_pass` |

## Running Locally

To run the application locally using Maven:

```bash
mvn spring-boot:run
```

The application will start on port `8086` (or the port specified by `PORT` env var).

## Running with Docker

### Build the Image

```bash
docker build -t catalog-service .
```

### Run the Container

```bash
docker run -p 8086:8086 -e SPRING_DATASOURCE_URL=jdbc:postgresql://host.docker.internal:5432/catalog_db catalog-service
```

*Note: `host.docker.internal` is used to access the host's Postgres database from within the container. Adjust as necessary for your network setup.*

## API Endpoints

The service exposes the following standard Actuator endpoints:

- `GET /actuator/health`: Health check
- `GET /actuator/info`: Application info
