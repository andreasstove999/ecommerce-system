module github.com/andreasstove999/ecommerce-system/cart-service-go

go 1.24.5

require github.com/lib/pq v1.10.9

require github.com/google/uuid v1.6.0

require github.com/rabbitmq/amqp091-go v1.10.0

replace go.uber.org/goleak => ./internal/tools/goleak
