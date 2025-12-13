package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

const cartCheckedOutQueue = "cart.checkedout"

func MustDialRabbit() *amqp.Connection {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("connect to RabbitMQ: %v", err)
	}
	return conn
}

func StartCartCheckedOutConsumer(ctx context.Context, conn *amqp.Connection, repo order.Repository, logger *log.Logger) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		cartCheckedOutQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	msgs, err := ch.Consume(
		cartCheckedOutQueue,
		"order-service", // consumer tag
		false,           // autoAck
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Println("stopping cart.checkedout consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					logger.Println("messages channel closed")
					return
				}

				if err := handleCartCheckedOut(ctx, repo, msg.Body, logger); err != nil {
					logger.Printf("handle message error: %v", err)
					_ = msg.Nack(false, false) // drop for now
					continue
				}
				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

func handleCartCheckedOut(ctx context.Context, repo order.Repository, body []byte, logger *log.Logger) error {
	var ev CartCheckedOut
	if err := json.Unmarshal(body, &ev); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	o := &order.Order{
		CartID:      ev.CartID,
		UserID:      ev.UserID,
		TotalAmount: ev.TotalAmount,
		CreatedAt:   time.Now().UTC(),
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

	logger.Printf("created order %s for user %s from cart %s", o.ID, o.UserID, o.CartID)
	return nil
}
