package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/dedup"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

// Routing keys as constants for handler registration
const (
	RoutingCartCheckedOut   = CartCheckedOutRoutingKey
	RoutingPaymentSucceeded = PaymentSucceededRoutingKey
	RoutingPaymentFailed    = PaymentFailedRoutingKey
	RoutingStockReserved    = StockReservedRoutingKey

	consumerNameCartCheckedOut = "order-service.cart-checkedout"
)

// OrderPublisher defines the subset of publisher methods used by handlers.
type OrderPublisher interface {
	PublishOrderCreated(ctx context.Context, o *order.Order, meta EnvelopeMetadata) error
	PublishOrderCompleted(ctx context.Context, orderID, userID string, meta EnvelopeMetadata) error
}

// CartCheckedOutHandler returns a handler for cart.checkedout events.
func CartCheckedOutHandler(
	db *sql.DB,
	repo order.Repository,
	dedupRepo dedup.Repository,
	pub OrderPublisher,
	logger *log.Logger,
	consumeEnveloped bool,
) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		payload, envelope, err := parseCartCheckedOut(body, consumeEnveloped)
		if err != nil {
			return fmt.Errorf("parse CartCheckedOut: %w", err)
		}

		var correlationID string
		var causationID string
		if envelope != nil {
			correlationID = envelope.CorrelationID
			causationID = envelope.EventID
		}
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		if envelope != nil && envelope.Sequence != nil {
			last, found, err := dedupRepo.GetLastSequence(ctx, consumerNameCartCheckedOut, envelope.PartitionKey)
			if err != nil {
				return fmt.Errorf("dedup get last sequence: %w", err)
			}
			if *envelope.Sequence <= last {
				logger.Printf("skipping CartCheckedOut partition=%s seq=%d (last=%d)", envelope.PartitionKey, *envelope.Sequence, last)
				return nil
			}
			if found && *envelope.Sequence > last+1 {
				logger.Printf("warning: possible gap for partition=%s seq=%d last=%d", envelope.PartitionKey, *envelope.Sequence, last)
			}
		}

		o := &order.Order{
			CartID:      payload.CartID,
			UserID:      payload.UserID,
			TotalAmount: payload.TotalAmount,
			CreatedAt:   payload.Timestamp,
		}

		for _, it := range payload.Items {
			o.Items = append(o.Items, order.Item{
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
				Price:     it.Price,
			})
		}

		if envelope != nil && envelope.Sequence != nil {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return fmt.Errorf("begin tx: %w", err)
			}
			defer tx.Rollback()

			if err := repo.CreateWithTx(ctx, tx, o); err != nil {
				return fmt.Errorf("create order: %w", err)
			}
			if err := dedupRepo.UpsertLastSequence(ctx, tx, consumerNameCartCheckedOut, envelope.PartitionKey, *envelope.Sequence); err != nil {
				return fmt.Errorf("update dedup checkpoint: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}
		} else {
			if err := repo.Create(ctx, o); err != nil {
				return fmt.Errorf("create order: %w", err)
			}
		}

		if err := pub.PublishOrderCreated(ctx, o, EnvelopeMetadata{
			CorrelationID: correlationID,
			CausationID:   causationID,
		}); err != nil {
			return fmt.Errorf("publish OrderCreated: %w", err)
		}

		logger.Printf("created order %s for user %s from cart %s", o.ID, o.UserID, o.CartID)
		return nil
	}
}

// TODO: make sure this is correct when events are published
// PaymentSucceededHandler returns a handler for payment.succeeded events.
func PaymentSucceededHandler(repo order.Repository, pub OrderPublisher, logger *log.Logger) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var ev PaymentSucceeded
		if err := json.Unmarshal(body, &ev); err != nil {
			return fmt.Errorf("unmarshal PaymentSucceeded: %w", err)
		}

		state, err := repo.MarkPaymentSucceeded(ctx, ev.OrderID)
		if err != nil {
			return fmt.Errorf("mark payment succeeded: %w", err)
		}

		// If both payment + stock are ready -> complete and publish OrderCompleted
		if state.ReadyToComplete {
			if err := repo.MarkCompleted(ctx, ev.OrderID); err != nil {
				return fmt.Errorf("mark completed: %w", err)
			}
			if err := pub.PublishOrderCompleted(ctx, ev.OrderID, state.UserID, EnvelopeMetadata{}); err != nil {
				return fmt.Errorf("publish OrderCompleted: %w", err)
			}
			logger.Printf("order %s completed (after payment success)", ev.OrderID)
		}

		return nil
	}
}

// TODO: make sure this is correct when events are published
// PaymentFailedHandler returns a handler for payment.failed events.
func PaymentFailedHandler(repo order.Repository, logger *log.Logger) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var ev PaymentFailed
		if err := json.Unmarshal(body, &ev); err != nil {
			return fmt.Errorf("unmarshal PaymentFailed: %w", err)
		}

		if err := repo.MarkPaymentFailed(ctx, ev.OrderID, ev.Reason); err != nil {
			return fmt.Errorf("mark payment failed: %w", err)
		}

		logger.Printf("order %s payment failed: %s", ev.OrderID, ev.Reason)
		return nil
	}
}

// TODO: make sure this is correct when events are published
// StockReservedHandler returns a handler for stock.reserved events.
func StockReservedHandler(repo order.Repository, pub OrderPublisher, logger *log.Logger) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var ev StockReserved
		if err := json.Unmarshal(body, &ev); err != nil {
			return fmt.Errorf("unmarshal StockReserved: %w", err)
		}

		state, err := repo.MarkStockReserved(ctx, ev.OrderID)
		if err != nil {
			return fmt.Errorf("mark stock reserved: %w", err)
		}

		if state.ReadyToComplete {
			if err := repo.MarkCompleted(ctx, ev.OrderID); err != nil {
				return fmt.Errorf("mark completed: %w", err)
			}
			if err := pub.PublishOrderCompleted(ctx, ev.OrderID, state.UserID, EnvelopeMetadata{}); err != nil {
				return fmt.Errorf("publish OrderCompleted: %w", err)
			}
			logger.Printf("order %s completed (after stock reserved)", ev.OrderID)
		}

		return nil
	}
}
