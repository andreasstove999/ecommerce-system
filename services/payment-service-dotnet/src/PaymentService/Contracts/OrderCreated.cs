namespace PaymentService.Contracts
{
    public sealed class OrderCreated
    {
        public required Guid OrderId { get; init; }
        public required Guid UserId { get; init; }
        public required decimal TotalAmount { get; init; }
        public required string Currency { get; init; } // "DKK", "EUR", etc.
    }

}
