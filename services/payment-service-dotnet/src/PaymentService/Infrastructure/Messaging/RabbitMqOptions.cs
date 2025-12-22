namespace PaymentService.Infrastructure.Messaging
{
    public sealed class RabbitMqOptions
    {
        public string Url { get; set; } = "amqp://guest:guest@rabbitmq:5672/";
        public string Exchange { get; set; } = "domain-events";
        public string Queue { get; set; } = "payment-service";
        public string RoutingKeyOrderCreated { get; set; } = "OrderCreated.v1";
    }
}
