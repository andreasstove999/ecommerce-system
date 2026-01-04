package events

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange           = "ecommerce.events"
	CartCheckedOutRoutingKey = "cart.checkedout.v1"
	cartServiceName          = "cart-service-go"
)

func serviceQueue(serviceName, routingKey string) string {
	return serviceName + "." + routingKey
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

func cartQueueName(routingKey string) string {
	return serviceQueue(cartServiceName, routingKey)
}
