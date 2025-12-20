package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

// Queue names as constants for handler registration
const (
	QueueCartCheckedOut   = "cart.checkedout"
	QueuePaymentSucceeded = "payment.succeeded"
	QueuePaymentFailed    = "payment.failed"
	QueueStockReserved    = "stock.reserved"
)

// CartCheckedOutHandler returns a handler for cart.checkedout events.
func CartCheckedOutHandler(repo order.Repository, pub *Publisher, logger *log.Logger) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var ev CartCheckedOut
		if err := json.Unmarshal(body, &ev); err != nil {
			return fmt.Errorf("unmarshal CartCheckedOut: %w", err)
		}

		o := &order.Order{
			CartID:      ev.CartID,
			UserID:      ev.UserID,
			TotalAmount: ev.TotalAmount,
			CreatedAt:   ev.Timestamp,
		}

		for _, it := range ev.Items {
			o.Items = append(o.Items, order.Item{
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
				Price:     it.Price,
			})
		}

		if err := repo.Create(ctx, o); err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		if err := pub.PublishOrderCreated(ctx, o); err != nil {
			return fmt.Errorf("publish OrderCreated: %w", err)
		}

		logger.Printf("created order %s for user %s from cart %s", o.ID, o.UserID, o.CartID)
		return nil
	}
}

// TODO: make sure this is correct when events are published
// PaymentSucceededHandler returns a handler for payment.succeeded events.
func PaymentSucceededHandler(repo order.Repository, pub *Publisher, logger *log.Logger) HandlerFunc {
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
			if err := pub.PublishOrderCompleted(ctx, ev.OrderID, state.UserID); err != nil {
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
func StockReservedHandler(repo order.Repository, pub *Publisher, logger *log.Logger) HandlerFunc {
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
			if err := pub.PublishOrderCompleted(ctx, ev.OrderID, state.UserID); err != nil {
				return fmt.Errorf("publish OrderCompleted: %w", err)
			}
			logger.Printf("order %s completed (after stock reserved)", ev.OrderID)
		}

		return nil
	}
}
