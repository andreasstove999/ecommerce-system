package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/config"
	httpapi "github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/http"
)

func main() {
	cfg := config.Load()

	logger := log.New(os.Stdout, "[api-gateway] ", log.LstdFlags|log.Lmicroseconds)

	// Base HTTP client (shared)
	sharedHTTP := &http.Client{
		Timeout: cfg.UpstreamTimeout,
	}

	// Upstream clients
	cartBase := clients.NewClient("cart-service", cfg.CartURL, sharedHTTP)
	orderBase := clients.NewClient("order-service", cfg.OrderURL, sharedHTTP)
	inventoryBase := clients.NewClient("inventory-service", cfg.InventoryURL, sharedHTTP)
	catalogBase := clients.NewClient("catalog-service", cfg.CatalogURL, sharedHTTP)
	paymentBase := clients.NewClient("payment-service", cfg.PaymentURL, sharedHTTP)
	shippingBase := clients.NewClient("shipping-service", cfg.ShippingURL, sharedHTTP)

	// Typed clients
	cart := clients.NewCartClient(cartBase)
	order := clients.NewOrderClient(orderBase)
	inventory := clients.NewInventoryClient(inventoryBase)
	catalog := clients.NewCatalogClient(catalogBase)
	payment := clients.NewPaymentClient(paymentBase)
	shipping := clients.NewShippingClient(shippingBase)

	// Health probes (note: different services use different health paths)
	healthProbes := []clients.HealthProbe{
		{Name: "cart-service", Client: cartBase, Path: "/health"},
		{Name: "order-service", Client: orderBase, Path: "/health"},
		{Name: "inventory-service", Client: inventoryBase, Path: "/health"},
		{Name: "catalog-service", Client: catalogBase, Path: "/api/catalog/health"},
		{Name: "payment-service", Client: paymentBase, Path: "/health"},
		// shipping-service health endpoint is not implemented
	}

	router := httpapi.NewRouter(httpapi.Deps{
		Logger:       logger,
		Cfg:          cfg,
		Cart:         cart,
		Order:        order,
		Inventory:    inventory,
		Catalog:      catalog,
		Payment:      payment,
		Shipping:     shipping,
		HealthProbes: healthProbes,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Printf("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
	logger.Printf("shutdown complete")
}
