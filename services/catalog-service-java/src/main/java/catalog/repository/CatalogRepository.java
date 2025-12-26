package catalog.repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import catalog.domain.Product;

public interface CatalogRepository {
    List<Product> list(int limit, int offset);

    Optional<Product> getById(UUID id);

    Product create(Product product);
}
