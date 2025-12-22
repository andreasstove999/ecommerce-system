namespace PaymentService.Contracts
{
    public sealed class PaymentSucceeded
    {
        public required Guid OrderId { get; init; }
        public required Guid PaymentId { get; init; }
        public required decimal Amount { get; init; }
        public required string Currency { get; init; }
        public required string Provider { get; init; }
    }
}
