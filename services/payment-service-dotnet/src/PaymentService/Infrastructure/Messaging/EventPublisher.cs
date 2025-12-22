using RabbitMQ.Client;
using System.Text;
using System.Text.Json;

namespace PaymentService.Infrastructure.Messaging
{
    public sealed class EventPublisher
    {
        private readonly RabbitMqOptions _opt;
        private readonly IConnection _conn;

        public EventPublisher(RabbitMqOptions opt, IConnection conn)
        {
            _opt = opt;
            _conn = conn;
        }

        public void Publish<T>(string routingKey, T message)
        {
            PublishAsync(routingKey, message).GetAwaiter().GetResult();
        }

        private async Task PublishAsync<T>(string routingKey, T message)
        {
            var channel = await _conn.CreateChannelAsync();
            try
            {
                await channel.ExchangeDeclareAsync(_opt.Exchange, ExchangeType.Topic, durable: true);

                var body = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(message));

                // Use the default properties implementation for async client
                var props = new BasicProperties
                {
                    ContentType = "application/json",
                    DeliveryMode = (DeliveryModes)2
                };

                await channel.BasicPublishAsync(
                    exchange: _opt.Exchange,
                    routingKey: routingKey,
                    mandatory: false,
                    basicProperties: props,
                    body: body
                );
            }
            finally
            {
                channel.Dispose();
            }
        }
    }
}
