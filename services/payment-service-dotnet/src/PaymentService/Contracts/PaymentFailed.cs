namespace PaymentService.Contracts
{
    public sealed class PaymentFailed
    {
        public required Guid OrderId { get; init; }
        public required Guid PaymentId { get; init; }
        public required string Reason { get; init; }
    }
}
