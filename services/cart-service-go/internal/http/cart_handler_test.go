package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cartpkg "github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/events"
	httphandler "github.com/andreasstove999/ecommerce-system/cart-service-go/internal/http"
	"github.com/google/uuid"
)

// TODO: add RabbitCartEventsPublisher mock
func TestGetCart(t *testing.T) {
	t.Run("missing user id", func(t *testing.T) {
		handler := httphandler.NewCartHandler(&RepositoryMock{}, nil)
		r := httptest.NewRequest(http.MethodGet, "/api/cart/", nil)
		w := httptest.NewRecorder()

		handler.GetCart(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return nil, errors.New("db error")
		}}
		handler := httphandler.NewCartHandler(repo, nil)
		r := httptest.NewRequest(http.MethodGet, "/api/cart/123", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.GetCart(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return nil, nil
		}}
		handler := httphandler.NewCartHandler(repo, nil)
		r := httptest.NewRequest(http.MethodGet, "/api/cart/123", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.GetCart(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		expected := &cartpkg.Cart{ID: "c1", UserID: "123", Items: []cartpkg.Item{{ProductID: "p1", Quantity: 2, Price: 5}}, Total: 10}
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return expected, nil
		}}
		handler := httphandler.NewCartHandler(repo, nil)
		r := httptest.NewRequest(http.MethodGet, "/api/cart/123", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.GetCart(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp cartpkg.Cart
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.ID != expected.ID || resp.UserID != expected.UserID || resp.Total != expected.Total {
			t.Fatalf("unexpected response %+v", resp)
		}
		if len(resp.Items) != 1 || resp.Items[0].ProductID != "p1" {
			t.Fatalf("unexpected items %+v", resp.Items)
		}
	})
}

func TestAddItem(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		handler := httphandler.NewCartHandler(&RepositoryMock{}, nil)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/items", bytes.NewBufferString("{"))
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.AddItem(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("load error", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return nil, errors.New("load error")
		}}
		handler := httphandler.NewCartHandler(repo, nil)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/items", bytes.NewBufferString(`{"productId":"p1","quantity":1,"price":2}`))
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.AddItem(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("create new cart", func(t *testing.T) {
		var saved *cartpkg.Cart
		repo := &RepositoryMock{
			GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return nil, nil },
			UpsertCartFunc: func(ctx context.Context, c *cartpkg.Cart) error {
				saved = c
				return nil
			},
		}
		handler := httphandler.NewCartHandler(repo, nil)
		body := bytes.NewBufferString(`{"productId":"p1","quantity":2,"price":3}`)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/items", body)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.AddItem(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if saved == nil {
			t.Fatalf("expected cart to be saved")
		}
		if saved.UserID != "123" || saved.Total != 6 {
			t.Fatalf("unexpected saved cart %+v", saved)
		}
		if len(saved.Items) != 1 || saved.Items[0].ProductID != "p1" || saved.Items[0].Quantity != 2 {
			t.Fatalf("unexpected items %+v", saved.Items)
		}
	})

	t.Run("update existing item", func(t *testing.T) {
		existing := &cartpkg.Cart{UserID: "123", Items: []cartpkg.Item{{ProductID: "p1", Quantity: 1, Price: 5}}}
		var saved *cartpkg.Cart
		repo := &RepositoryMock{
			GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return existing, nil },
			UpsertCartFunc: func(ctx context.Context, c *cartpkg.Cart) error {
				saved = c
				return nil
			},
		}
		handler := httphandler.NewCartHandler(repo, nil)
		body := bytes.NewBufferString(`{"productId":"p1","quantity":2,"price":5}`)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/items", body)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.AddItem(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if saved == nil || len(saved.Items) != 1 {
			t.Fatalf("expected cart to be saved with items")
		}
		if saved.Items[0].Quantity != 3 {
			t.Fatalf("expected quantity 3, got %d", saved.Items[0].Quantity)
		}
		if saved.Total != 15 {
			t.Fatalf("expected total 15, got %f", saved.Total)
		}
	})

	t.Run("persist error", func(t *testing.T) {
		repo := &RepositoryMock{
			GetCartFunc:    func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return nil, nil },
			UpsertCartFunc: func(ctx context.Context, c *cartpkg.Cart) error { return errors.New("save failed") },
		}
		handler := httphandler.NewCartHandler(repo, nil)
		body := bytes.NewBufferString(`{"productId":"p1","quantity":1,"price":2}`)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/items", body)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.AddItem(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestCheckout(t *testing.T) {
	t.Run("missing user id", func(t *testing.T) {
		handler := httphandler.NewCartHandler(&RepositoryMock{}, &RabbitCartEventsPublisherMock{})
		r := httptest.NewRequest(http.MethodPost, "/api/cart/", nil)
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("load error", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return nil, errors.New("db error")
		}}
		handler := httphandler.NewCartHandler(repo, &RabbitCartEventsPublisherMock{})
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("cart not found", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return nil, nil
		}}
		handler := httphandler.NewCartHandler(repo, &RabbitCartEventsPublisherMock{})
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("publish error", func(t *testing.T) {
		repo := &RepositoryMock{GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) {
			return &cartpkg.Cart{ID: "c1", UserID: userID}, nil
		}}
		publisher := &RabbitCartEventsPublisherMock{PublishCartCheckedOutFunc: func(ctx context.Context, c *cartpkg.Cart, _ events.PublishMetadata) error {
			return errors.New("publish failed")
		}}
		handler := httphandler.NewCartHandler(repo, publisher)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
		if len(publisher.PublishCartCheckedOutCalls()) != 1 {
			t.Fatalf("expected publish to be called once, got %d", len(publisher.PublishCartCheckedOutCalls()))
		}
	})

	t.Run("clear error", func(t *testing.T) {
		cart := &cartpkg.Cart{ID: "c1", UserID: "123"}
		repo := &RepositoryMock{
			GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return cart, nil },
			ClearCartFunc: func(ctx context.Context, userID string) error {
				return errors.New("clear failed")
			},
		}
		publisher := &RabbitCartEventsPublisherMock{PublishCartCheckedOutFunc: func(ctx context.Context, c *cartpkg.Cart, _ events.PublishMetadata) error {
			return nil
		}}
		handler := httphandler.NewCartHandler(repo, publisher)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
		if len(publisher.PublishCartCheckedOutCalls()) != 1 {
			t.Fatalf("expected publish to be called once, got %d", len(publisher.PublishCartCheckedOutCalls()))
		}
	})

	t.Run("propagates correlation and causation headers", func(t *testing.T) {
		cart := &cartpkg.Cart{ID: "c1", UserID: "123"}
		repo := &RepositoryMock{
			GetCartFunc:   func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return cart, nil },
			ClearCartFunc: func(ctx context.Context, userID string) error { return nil },
		}
		var captured events.PublishMetadata
		publisher := &RabbitCartEventsPublisherMock{PublishCartCheckedOutFunc: func(ctx context.Context, c *cartpkg.Cart, metadata events.PublishMetadata) error {
			captured = metadata
			return nil
		}}
		handler := httphandler.NewCartHandler(repo, publisher)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.Header.Set("X-Correlation-Id", "123e4567-e89b-12d3-a456-426614174000")
		r.Header.Set("X-Causation-Id", "223e4567-e89b-12d3-a456-426614174000")
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if captured.CorrelationID != "123e4567-e89b-12d3-a456-426614174000" {
			t.Fatalf("unexpected correlation id %s", captured.CorrelationID)
		}
		if captured.CausationID != "223e4567-e89b-12d3-a456-426614174000" {
			t.Fatalf("unexpected causation id %s", captured.CausationID)
		}
	})

	t.Run("generates correlation id when missing", func(t *testing.T) {
		cart := &cartpkg.Cart{ID: "c1", UserID: "123"}
		repo := &RepositoryMock{
			GetCartFunc:   func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return cart, nil },
			ClearCartFunc: func(ctx context.Context, userID string) error { return nil },
		}
		var captured events.PublishMetadata
		publisher := &RabbitCartEventsPublisherMock{PublishCartCheckedOutFunc: func(ctx context.Context, c *cartpkg.Cart, metadata events.PublishMetadata) error {
			captured = metadata
			return nil
		}}
		handler := httphandler.NewCartHandler(repo, publisher)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if captured.CorrelationID == "" {
			t.Fatalf("expected correlation id to be generated")
		}
		if _, err := uuid.Parse(captured.CorrelationID); err != nil {
			t.Fatalf("expected correlation id to be a valid uuid, got %v", err)
		}
		if captured.CausationID != "" {
			t.Fatalf("did not expect causation id when header missing, got %s", captured.CausationID)
		}
	})

	t.Run("success", func(t *testing.T) {
		cart := &cartpkg.Cart{ID: "c1", UserID: "123"}
		cleared := false
		repo := &RepositoryMock{
			GetCartFunc: func(ctx context.Context, userID string) (*cartpkg.Cart, error) { return cart, nil },
			ClearCartFunc: func(ctx context.Context, userID string) error {
				cleared = true
				return nil
			},
		}
		publisher := &RabbitCartEventsPublisherMock{PublishCartCheckedOutFunc: func(ctx context.Context, c *cartpkg.Cart, _ events.PublishMetadata) error {
			return nil
		}}
		handler := httphandler.NewCartHandler(repo, publisher)
		r := httptest.NewRequest(http.MethodPost, "/api/cart/123/checkout", nil)
		r.SetPathValue("userId", "123")
		w := httptest.NewRecorder()

		handler.Checkout(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if len(publisher.PublishCartCheckedOutCalls()) != 1 {
			t.Fatalf("expected publish to be called once, got %d", len(publisher.PublishCartCheckedOutCalls()))
		}
		if !cleared {
			t.Fatalf("expected cart to be cleared")
		}
	})
}
