using System.Text;
using System.Text.Json;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;
using PaymentService.Contracts;
using PaymentService.Domain;
using PaymentService.Infrastructure.DB;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;

namespace PaymentService.Infrastructure.Messaging;

public sealed class OrderCreatedConsumer : BackgroundService
{
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly RabbitMqOptions _opt;
    private readonly IConnection _conn;
    private readonly EventPublisher _publisher;

    public OrderCreatedConsumer(
        IServiceScopeFactory scopeFactory,
        IOptions<RabbitMqOptions> options,
        IConnection conn,
        EventPublisher publisher)
    {
        _scopeFactory = scopeFactory;
        _opt = options.Value;
        _conn = conn;
        _publisher = publisher;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        var channel = await _conn.CreateChannelAsync();

        try
        {
            await channel.ExchangeDeclareAsync(_opt.Exchange, ExchangeType.Topic, durable: true);
            await channel.QueueDeclareAsync(_opt.Queue, durable: true, exclusive: false, autoDelete: false);
            await channel.QueueBindAsync(_opt.Queue, _opt.Exchange, _opt.RoutingKeyOrderCreated);

            await channel.BasicQosAsync(prefetchSize: 0, prefetchCount: 10, global: false);

            var consumer = new AsyncEventingBasicConsumer(channel);
            consumer.ReceivedAsync += async (_, ea) =>
            {
                try
                {
                    var json = Encoding.UTF8.GetString(ea.Body.ToArray());
                    var envelope = JsonSerializer.Deserialize<Envelope<OrderCreated>>(json);

                    if (envelope is null)
                    {
                        await channel.BasicAckAsync(ea.DeliveryTag, multiple: false);
                        return;
                    }

                    await Handle(envelope, stoppingToken);
                    await channel.BasicAckAsync(ea.DeliveryTag, multiple: false);
                }
                catch
                {
                    // Requeue=false to avoid poison-message infinite loops
                    await channel.BasicNackAsync(ea.DeliveryTag, multiple: false, requeue: false);
                }
            };

            await channel.BasicConsumeAsync(queue: _opt.Queue, autoAck: false, consumer: consumer);

            // Keep the service running until the host is shutting down.
            await Task.Delay(Timeout.InfiniteTimeSpan, stoppingToken);
        }
        catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
        {
            // expected during shutdown
        }
        finally
        {
            channel.Dispose();
        }
    }

    private async Task Handle(Envelope<OrderCreated> env, CancellationToken ct)
    {
        using var scope = _scopeFactory.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<PaymentDbContext>();

        // Idempotency: ignore duplicates (e.g., redeliveries)
        var existing = await db.Payments.SingleOrDefaultAsync(p => p.OrderId == env.Payload.OrderId, ct);
        if (existing is not null)
        {
            return;
        }

        var payment = new Payment
        {
            OrderId = env.Payload.OrderId,
            UserId = env.Payload.UserId,
            Amount = env.Payload.TotalAmount,
            Currency = env.Payload.Currency,
            Status = PaymentStatus.Pending
        };

        db.Payments.Add(payment);
        await db.SaveChangesAsync(ct);

        var (ok, reason) = Simulate(payment.Amount);

        payment.Status = ok ? PaymentStatus.Succeeded : PaymentStatus.Failed;
        payment.FailureReason = ok ? null : reason;
        await db.SaveChangesAsync(ct);

        if (ok)
        {
            var succeeded = BuildEnvelope(
                name: "PaymentSucceeded",
                version: 1,
                correlationId: env.CorrelationId,
                partitionKey: payment.OrderId.ToString(),
                payload: new PaymentSucceeded
                {
                    OrderId = payment.OrderId,
                    PaymentId = payment.PaymentId,
                    Amount = payment.Amount,
                    Currency = payment.Currency,
                    Provider = payment.Provider
                });

            _publisher.Publish("PaymentSucceeded.v1", succeeded);
        }
        else
        {
            var failed = BuildEnvelope(
                name: "PaymentFailed",
                version: 1,
                correlationId: env.CorrelationId,
                partitionKey: payment.OrderId.ToString(),
                payload: new PaymentFailed
                {
                    OrderId = payment.OrderId,
                    PaymentId = payment.PaymentId,
                    Reason = reason ?? "Unknown"
                });

            _publisher.Publish("PaymentFailed.v1", failed);
        }
    }

    private static (bool ok, string? reason) Simulate(decimal amount)
    {
        if (amount <= 0) return (false, "Amount must be > 0");
        if (amount > 5000) return (false, "Amount exceeds limit");
        return (true, null);
    }

    private static Envelope<T> BuildEnvelope<T>(
        string name,
        int version,
        Guid? correlationId,
        string? partitionKey,
        T payload)
        => new()
        {
            EventName = name,
            EventVersion = version,
            EventId = Guid.NewGuid(),
            CorrelationId = correlationId,
            Producer = "payment-service-dotnet",
            PartitionKey = partitionKey,
            Sequence = null,
            OccurredAt = DateTimeOffset.UtcNow,
            Schema = null,
            Payload = payload
        };
}
