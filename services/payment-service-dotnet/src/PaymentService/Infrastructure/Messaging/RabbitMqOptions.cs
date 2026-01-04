namespace PaymentService.Infrastructure.Messaging
{
    public sealed class RabbitMqOptions
    {
        public string Url { get; set; } = "amqp://guest:guest@rabbitmq:5672/";
        public string Exchange { get; set; } = "ecommerce.events";
        public string Queue { get; set; } = "payment-service-dotnet.order.created.v1";
        public string RoutingKeyOrderCreated { get; set; } = "order.created.v1";
    }
}
