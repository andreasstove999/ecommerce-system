package events

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange             = "ecommerce.events"
	CartCheckedOutRoutingKey   = "cart.checkedout.v1"
	PaymentSucceededRoutingKey = "payment.succeeded.v1"
	PaymentFailedRoutingKey    = "payment.failed.v1"
	StockReservedRoutingKey    = "stock.reserved.v1"
	OrderCreatedRoutingKey     = "order.created.v1"
	OrderCompletedRoutingKey   = "order.completed.v1"
	orderServiceName           = "order-service-go"
)

func serviceQueue(serviceName, routingKey string) string {
	return serviceName + "." + routingKey
}

func orderQueueName(routingKey string) string {
	return serviceQueue(orderServiceName, routingKey)
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
