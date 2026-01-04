package events

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange          = "ecommerce.events"
	OrderCreatedRoutingKey  = "order.created.v1"
	StockReservedRoutingKey = "stock.reserved.v1"
	StockDepletedRoutingKey = "stock.depleted.v1"
	inventoryServiceName    = "inventory-service-go"
)

func serviceQueue(serviceName, routingKey string) string {
	return serviceName + "." + routingKey
}

func inventoryQueueName(routingKey string) string {
	return serviceQueue(inventoryServiceName, routingKey)
}

func declareEventsExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		EventsExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}
