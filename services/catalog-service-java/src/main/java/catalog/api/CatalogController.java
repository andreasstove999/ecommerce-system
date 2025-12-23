package catalog.api;

import catalog.domain.Product;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import catalog.api.dto.CreateProductRequest;
import catalog.api.dto.ProductResponse;
import catalog.service.CatalogService;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/catalog")
public class CatalogController {

    // TODO: add CORS support

    // TODO: add post product endpoint

    private final CatalogService service;

    public CatalogController(CatalogService service) {
        this.service = service;
    }

    @GetMapping("/products")
    public List<ProductResponse> listProducts(
            @RequestParam(defaultValue = "50") int limit,
            @RequestParam(defaultValue = "0") int offset) {
        return service.listProducts(limit, offset).stream()
                .map(ProductResponse::from)
                .toList();
    }

    @GetMapping("/products/{id}")
    public ProductResponse getProduct(@PathVariable UUID id) {
        return service.getProduct(id)
                .map(ProductResponse::from)
                .orElseThrow(() -> new NotFoundException("product not found"));
    }

    @PostMapping("/products")
    @ResponseStatus(HttpStatus.CREATED)
    public ProductResponse createProduct(@Valid @RequestBody CreateProductRequest req) {
        var p = new Product();
        p.setSku(req.getSku());
        p.setName(req.getName());
        p.setDescription(req.getDescription() == null ? "" : req.getDescription());
        p.setPrice(req.getPrice());
        p.setCurrency(req.getCurrency());

        return ProductResponse.from(service.createProduct(p));
    }

    @GetMapping("/health")
    public Object health() {
        return java.util.Map.of("status", "ok", "service", "catalog-service");
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    private static class NotFoundException extends RuntimeException {
        public NotFoundException(String message) {
            super(message);
        }
    }
}
