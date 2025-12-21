package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

type StockPublisher interface {
	PublishStockReserved(ctx context.Context, orderID, userID string, reserved []inventory.Line) error
	PublishStockDepleted(ctx context.Context, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error
}

// OrderCreatedHandler reserves stock and publishes either StockReserved or StockDepleted.
// Returning an error will NACK the message (and it will be sent to the DLQ by the Consumer).
func OrderCreatedHandler(repo inventory.Repository, pub StockPublisher, logger *log.Logger) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var ev OrderCreated
		if err := json.Unmarshal(body, &ev); err != nil {
			return fmt.Errorf("unmarshal OrderCreated: %w", err)
		}
		if ev.OrderID == "" {
			return fmt.Errorf("missing orderId")
		}

		lines := make([]inventory.Line, 0, len(ev.Items))
		for _, it := range ev.Items {
			if it.ProductID == "" || it.Quantity <= 0 {
				continue
			}
			lines = append(lines, inventory.Line{ProductID: it.ProductID, Quantity: it.Quantity})
		}

		result, err := repo.Reserve(ctx, ev.OrderID, lines)
		if err != nil {
			return fmt.Errorf("reserve for order %s: %w", ev.OrderID, err)
		}

		if len(result.Depleted) > 0 {
			logger.Printf("stock depleted for order=%s depleted=%d reserved=%d", ev.OrderID, len(result.Depleted), len(result.Reserved))
			return pub.PublishStockDepleted(ctx, ev.OrderID, ev.UserID, result.Depleted, result.Reserved)
		}

		logger.Printf("stock reserved for order=%s lines=%d", ev.OrderID, len(result.Reserved))
		return pub.PublishStockReserved(ctx, ev.OrderID, ev.UserID, result.Reserved)
	}
}
