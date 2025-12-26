package catalog.service;

import org.springframework.stereotype.Service;

import catalog.domain.Product;
import catalog.repository.CatalogRepository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class CatalogService {

    private final CatalogRepository repo;

    public CatalogService(CatalogRepository repo) {
        this.repo = repo;
    }

    public List<Product> listProducts(int limit, int offset) {
        if (limit <= 0)
            limit = 50;
        if (offset < 0)
            offset = 0;
        return repo.list(limit, offset);
    }

    public Optional<Product> getProduct(UUID id) {
        return repo.getById(id);
    }

    public Product createProduct(Product p) {
        // future: enforce SKU uniqueness, validation, etc.
        return repo.create(p);
    }
}
