package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/dedup"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StockPublisher interface {
	PublishStockReserved(ctx context.Context, meta EventMeta, orderID, userID string, reserved []inventory.Line) error
	PublishStockDepleted(ctx context.Context, meta EventMeta, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error
}

const orderCreatedConsumerName = "inventory-order-created"

// OrderCreatedHandler reserves stock and publishes either StockReserved or StockDepleted.
// Returning an error will NACK the message (and it will be sent to the DLQ by the Consumer).
func OrderCreatedHandler(repo inventory.TransactionalRepository, dedupRepo *dedup.Repository, pub StockPublisher, logger *log.Logger, consumerName string, consumeEnveloped bool) HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		msg, err := parseOrderCreated(body, consumeEnveloped)
		if err != nil {
			return err
		}
		if msg.Payload.OrderID == "" {
			return fmt.Errorf("missing orderId")
		}

		lines := make([]inventory.Line, 0, len(msg.Payload.Items))
		for _, it := range msg.Payload.Items {
			if it.ProductID == "" || it.Quantity <= 0 {
				continue
			}
			lines = append(lines, inventory.Line{ProductID: it.ProductID, Quantity: it.Quantity})
		}

		var partitionKey string
		var incomingSeq int64
		var correlationID, causationID string

		if msg.Envelope != nil {
			partitionKey = msg.Envelope.PartitionKey
			incomingSeq = msg.Envelope.Sequence
			correlationID = msg.Envelope.CorrelationID
			causationID = msg.Envelope.EventID
		}
		if partitionKey == "" {
			partitionKey = msg.Payload.OrderID
		}
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		tx, err := repo.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		localDedup := dedupRepo.WithExecutor(tx)

		if msg.Envelope != nil {
			lastSeq, ok, err := localDedup.GetLastSequence(ctx, consumerName, partitionKey)
			if err != nil {
				return err
			}
			if ok && incomingSeq != 0 {
				if incomingSeq <= lastSeq {
					logger.Printf("skip duplicate orderId=%s partition=%s seq=%d last=%d", msg.Payload.OrderID, partitionKey, incomingSeq, lastSeq)
					return nil
				}
				if incomingSeq > lastSeq+1 {
					logger.Printf("warning: sequence gap for partition=%s seq=%d last=%d", partitionKey, incomingSeq, lastSeq)
				}
			}
		}

		result, err := repo.ReserveWithTx(ctx, tx, msg.Payload.OrderID, lines)
		if err != nil {
			return fmt.Errorf("reserve for order %s: %w", msg.Payload.OrderID, err)
		}

		if msg.Envelope != nil && incomingSeq != 0 {
			if err := localDedup.UpsertLastSequence(ctx, consumerName, partitionKey, incomingSeq); err != nil {
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit reserve: %w", err)
		}

		meta := EventMeta{
			CorrelationID: correlationID,
			CausationID:   causationID,
			PartitionKey:  msg.Payload.OrderID,
		}

		if len(result.Depleted) > 0 {
			logger.Printf("stock depleted for order=%s depleted=%d reserved=%d", msg.Payload.OrderID, len(result.Depleted), len(result.Reserved))
			return pub.PublishStockDepleted(ctx, meta, msg.Payload.OrderID, msg.Payload.UserID, result.Depleted, result.Reserved)
		}

		logger.Printf("stock reserved for order=%s lines=%d", msg.Payload.OrderID, len(result.Reserved))
		return pub.PublishStockReserved(ctx, meta, msg.Payload.OrderID, msg.Payload.UserID, result.Reserved)
	}
}

func consumeEnvelopedEnabled() bool {
	v := os.Getenv(consumeEnvelopedEnv)
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return true
	}
	return b
}

func publishEnvelopedEnabled() bool {
	v := os.Getenv(publishEnvelopedEnv)
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return true
	}
	return b
}
