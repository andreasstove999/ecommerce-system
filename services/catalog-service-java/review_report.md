# Branch Review: `catalog-service`

I have reviewed the changes in the `catalog-service` branch. Overall, the implementation provides a solid foundation for the catalog service with basic CRUD operations.

## Summary

- **New Service**: `catalog-service-java` (Spring Boot 3.2.0, Java 17)
- **Database**: PostgreSQL with Flyway migrations.
- **Components**:
  - `PostgresCatalogRepository`: JDBC-based implementation.
  - `CatalogService`: Business logic.
  - `CatalogController`: (Assumed, based on architecture) API endpoints.

## Findings & Suggestions

### 1. Database & Migrations (`V1__init.sql`)
- **Good**: Uses `TIMESTAMPTZ` for `created_at` and `updated_at`, which is a best practice.
- **Good**: Adds an index on `created_at` for efficient sorting.
- **Note**: The schema defines `price` as `NUMERIC(12,2)`. Ensure that the Java `double` type usage doesn't introduce precision issues for currency calculations. `BigDecimal` is generally preferred for money.

### 2. Repository Layer (`PostgresCatalogRepository.java`)
- **SQL Injection Safety**: Uses `PreparedStatement` (via `JdbcTemplate` arguments) which prevents SQL injection.
- **Row Mapping**: `PRODUCT_ROW_MAPPER` cleanly maps the result set.
- **Observation**: `price` is read as `BigDecimal` but immediately converted to `double` in the mapper.
  ```java
  p.setPrice(rs.getBigDecimal("price").doubleValue());
  ```

### 3. Service Layer (`CatalogService.java`)
- **Logic**: `listProducts` handles pagination limits and offsets defensively.
- **Extension Point**: `createProduct` has a comment `// future: enforce SKU uniqueness, validation, etc.` which is a good reminder for next steps.

### 4. Application Entry (`CatalogServiceApplication.java`)
- **TODOs**: Contains `TODO` comments regarding frontend/gateway testing and seeding data.
  ```java
  // TODO: This is yet to be tested for frontend and gateway.
  // TODO: seed some products for the fontend to show.
  ```

### 5. Build & Deployment
- `pom.xml`: Includes necessary dependencies (Web, JDBC, Flyway, Postgres, Validation).
- `docker-compose.yml`: Correctly defines the service and its database, linking it to RabbitMQ.

## Recommendations
1.  **Money Type**: Consider switching `Product.price` and related DTO fields from `double` to `BigDecimal` to avoid floating-point errors with currency.
2.  **Validation**: Ensure `CatalogController` (if implemented) uses `@Valid` on `CreateProductRequest` to enforce the `@NotBlank` and `@Positive` annotations.
3.  **Testing**: Verify the `TODOs` in `CatalogServiceApplication` by adding integration tests or a seeding script.
