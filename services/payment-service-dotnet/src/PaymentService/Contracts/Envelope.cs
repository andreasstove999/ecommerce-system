namespace PaymentService.Contracts
{
    public sealed class Envelope<TPayload>
    {
        public required string EventName { get; init; }
        public required int EventVersion { get; init; }
        public required Guid EventId { get; init; }
        public Guid? CorrelationId { get; init; }
        public required string Producer { get; init; }
        public string? PartitionKey { get; init; }
        public long? Sequence { get; init; }
        public required DateTimeOffset OccurredAt { get; init; }
        public string? Schema { get; init; }
        public required TPayload Payload { get; init; }
    }

}