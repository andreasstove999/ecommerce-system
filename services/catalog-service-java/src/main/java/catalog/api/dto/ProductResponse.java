package catalog.api.dto;

import java.time.Instant;
import java.util.UUID;

import catalog.domain.Product;

public class ProductResponse {
    private UUID id;
    private String sku;
    private String name;
    private String description;
    private double price;
    private String currency;
    private boolean active;
    private Instant createdAt;
    private Instant updatedAt;

    public static ProductResponse from(Product p) {
        var r = new ProductResponse();
        r.id = p.getId();
        r.sku = p.getSku();
        r.name = p.getName();
        r.description = p.getDescription();
        r.price = p.getPrice();
        r.currency = p.getCurrency();
        r.active = p.isActive();
        r.createdAt = p.getCreatedAt();
        r.updatedAt = p.getUpdatedAt();
        return r;
    }

    public UUID getId() {
        return id;
    }

    public String getSku() {
        return sku;
    }

    public String getName() {
        return name;
    }

    public String getDescription() {
        return description;
    }

    public double getPrice() {
        return price;
    }

    public String getCurrency() {
        return currency;
    }

    public boolean isActive() {
        return active;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }
}
