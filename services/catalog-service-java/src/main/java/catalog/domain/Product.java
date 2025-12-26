package catalog.domain;

import java.time.Instant;
import java.util.UUID;

public class Product {
  private UUID id;
  private String sku;
  private String name;
  private String description;
  private double price;
  private String currency;
  private boolean active;
  private Instant createdAt;
  private Instant updatedAt;

  public Product() {
  }

  public Product(UUID id, String sku, String name, String description, double price, String currency,
      boolean active, Instant createdAt, Instant updatedAt) {
    this.id = id;
    this.sku = sku;
    this.name = name;
    this.description = description;
    this.price = price;
    this.currency = currency;
    this.active = active;
    this.createdAt = createdAt;
    this.updatedAt = updatedAt;
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

  public void setId(UUID id) {
    this.id = id;
  }

  public void setSku(String sku) {
    this.sku = sku;
  }

  public void setName(String name) {
    this.name = name;
  }

  public void setDescription(String description) {
    this.description = description;
  }

  public void setPrice(double price) {
    this.price = price;
  }

  public void setCurrency(String currency) {
    this.currency = currency;
  }

  public void setActive(boolean active) {
    this.active = active;
  }

  public void setCreatedAt(Instant createdAt) {
    this.createdAt = createdAt;
  }

  public void setUpdatedAt(Instant updatedAt) {
    this.updatedAt = updatedAt;
  }
}
