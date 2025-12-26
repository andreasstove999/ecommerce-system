package catalog.repository;

import org.springframework.context.annotation.Primary;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.stereotype.Repository;

import catalog.domain.Product;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
@Primary // makes sure this wins over any old InMemory repo
public class PostgresCatalogRepository implements CatalogRepository {

    private final JdbcTemplate jdbc;

    public PostgresCatalogRepository(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    private static final RowMapper<Product> PRODUCT_ROW_MAPPER = new RowMapper<>() {
        @Override
        public Product mapRow(ResultSet rs, int rowNum) throws SQLException {
            var p = new Product();
            p.setId((UUID) rs.getObject("id"));
            p.setSku(rs.getString("sku"));
            p.setName(rs.getString("name"));
            p.setDescription(rs.getString("description"));
            p.setPrice(rs.getBigDecimal("price").doubleValue());
            p.setCurrency(rs.getString("currency"));
            p.setActive(rs.getBoolean("active"));

            // TIMESTAMPTZ -> Instant
            p.setCreatedAt(rs.getTimestamp("created_at").toInstant());
            p.setUpdatedAt(rs.getTimestamp("updated_at").toInstant());
            return p;
        }
    };

    @Override
    public List<Product> list(int limit, int offset) {
        if (limit <= 0)
            limit = 50;
        if (offset < 0)
            offset = 0;

        return jdbc.query("""
                SELECT id, sku, name, description, price, currency, active, created_at, updated_at
                FROM products
                ORDER BY created_at DESC
                LIMIT ? OFFSET ?
                """, PRODUCT_ROW_MAPPER, limit, offset);
    }

    @Override
    public Optional<Product> getById(UUID id) {
        var rows = jdbc.query("""
                SELECT id, sku, name, description, price, currency, active, created_at, updated_at
                FROM products
                WHERE id = ?
                """, PRODUCT_ROW_MAPPER, id);

        return rows.stream().findFirst();
    }

    @Override
    public Product create(Product product) {
        var now = Instant.now();
        if (product.getId() == null)
            product.setId(UUID.randomUUID());
        if (product.getCurrency() == null || product.getCurrency().isBlank())
            product.setCurrency("USD");

        product.setActive(true);
        product.setCreatedAt(now);
        product.setUpdatedAt(now);

        jdbc.update("""
                INSERT INTO products (id, sku, name, description, price, currency, active, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                product.getId(),
                product.getSku(),
                product.getName(),
                product.getDescription() == null ? "" : product.getDescription(),
                // NUMERIC(12,2) prefers BigDecimal, but double works too; JDBC will convert.
                product.getPrice(),
                product.getCurrency(),
                product.isActive(),
                Timestamp.from(product.getCreatedAt()),
                Timestamp.from(product.getUpdatedAt()));

        return product;
    }
}
