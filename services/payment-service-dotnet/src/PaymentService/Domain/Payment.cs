using System.ComponentModel.DataAnnotations;

namespace PaymentService.Domain
{
    public sealed class Payment
    {
        [Key]
        public Guid PaymentId { get; set; } = Guid.NewGuid();

        public Guid OrderId { get; set; }
        public Guid UserId { get; set; }

        public decimal Amount { get; set; }
        public string Currency { get; set; } = "DKK";

        public PaymentStatus Status { get; set; } = PaymentStatus.Pending;

        public string Provider { get; set; } = "MockProvider";
        public string? FailureReason { get; set; }

        public DateTimeOffset CreatedAt { get; set; } = DateTimeOffset.UtcNow;
    }
}
