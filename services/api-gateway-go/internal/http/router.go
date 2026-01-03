package http

import (
	"log"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/config"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/http/handlers"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
)

type Deps struct {
	Logger *log.Logger
	Cfg    config.Config

	Cart      *clients.CartClient
	Order     *clients.OrderClient
	Inventory *clients.InventoryClient
	Catalog   *clients.CatalogClient
	Payment   *clients.PaymentClient
	Shipping  *clients.ShippingClient

	HealthProbes []clients.HealthProbe
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	// Health
	health := &handlers.HealthHandler{Probes: d.HealthProbes}
	mux.HandleFunc("GET /health", health.Gateway)
	mux.HandleFunc("GET /health/upstreams", health.Upstreams)

	// BFF: Cart (me)
	cart := handlers.NewCartHandler(d.Cart)
	mux.HandleFunc("GET /me/cart", cart.GetCartMe)
	mux.HandleFunc("POST /me/cart/items", cart.AddItemMe)
	mux.HandleFunc("POST /me/cart/checkout", cart.CheckoutMe)

	// BFF: Products (catalog)
	cat := handlers.NewCatalogHandler(d.Catalog)
	mux.HandleFunc("GET /products", cat.ListProducts)
	mux.HandleFunc("GET /products/{id}", cat.GetProduct)
	mux.HandleFunc("POST /products", cat.CreateProduct)

	// BFF: Orders
	order := handlers.NewOrderHandler(d.Order)
	mux.HandleFunc("GET /me/orders", order.ListOrdersMe)
	mux.HandleFunc("GET /orders/{orderId}", order.GetOrder)

	// BFF: Inventory
	inv := handlers.NewInventoryHandler(d.Inventory)
	mux.HandleFunc("GET /products/{productId}/availability", inv.Availability)
	mux.HandleFunc("POST /inventory/adjust", inv.Adjust) // optional/admin-y

	// BFF: Payment + Shipping (by order)
	pay := handlers.NewPaymentHandler(d.Payment)
	mux.HandleFunc("GET /orders/{orderId}/payment", pay.ByOrder)

	ship := handlers.NewShippingHandler(d.Shipping)
	mux.HandleFunc("GET /orders/{orderId}/shipping", ship.ByOrder)

	// Middlewares (outer -> inner)
	var h http.Handler = mux
	h = middleware.Recover(d.Logger)(h)
	h = middleware.CORS(d.Cfg.CORSAllowOrigins)(h)
	h = middleware.CorrelationID(h)
	h = middleware.RequireUserIDForMeRoutes(h) // <-- new
	h = middleware.AuthJWT(h)                  // still placeholder
	h = middleware.Logging(d.Logger)(h)

	return h
}
