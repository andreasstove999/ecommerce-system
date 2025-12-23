namespace PaymentService.Infrastructure.Messaging
{
    public sealed class RabbitMqOptions
    {
        public string Url { get; set; } = "amqp://guest:guest@rabbitmq:5672/";
        public string Exchange { get; set; } = string.Empty;
        public string Queue { get; set; } = "order.created";
        public string RoutingKeyOrderCreated { get; set; } = "order.created";
    }
}
