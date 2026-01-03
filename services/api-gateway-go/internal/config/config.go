package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port            string
	UpstreamTimeout time.Duration

	// Upstream base URLs (inside docker network recommended)
	CartURL      string
	OrderURL     string
	InventoryURL string
	CatalogURL   string
	PaymentURL   string
	ShippingURL  string

	// CORS
	CORSAllowOrigins []string
}

func Load() Config {
	port := getenv("PORT", "8080")

	timeout := parseDuration(getenv("UPSTREAM_TIMEOUT", "10s"), 10*time.Second)

	// Defaults match docker-compose service names + internal ports in your repo.
	// Override with env vars if you want to run gateway locally against localhost:ports.
	cfg := Config{
		Port:            port,
		UpstreamTimeout: timeout,

		CartURL:      getenv("CART_URL", "http://cart-service-go:8081"),
		OrderURL:     getenv("ORDER_URL", "http://order-service-go:8082"),
		InventoryURL: getenv("INVENTORY_URL", "http://inventory-service-go:8083"),
		CatalogURL:   getenv("CATALOG_URL", "http://catalog-service-java:8086"),
		PaymentURL:   getenv("PAYMENT_URL", "http://payment-service-dotnet:8080"),
		ShippingURL:  getenv("SHIPPING_URL", "http://shipping-service-java:8086"),

		CORSAllowOrigins: splitCSV(getenv("CORS_ALLOW_ORIGINS", "*")),
	}

	return cfg
}

func getenv(k, def string) string {
	if v := os.Getenv(k); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func parseDuration(v string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
