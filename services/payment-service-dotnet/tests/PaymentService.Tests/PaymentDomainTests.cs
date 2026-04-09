using PaymentService.Domain;

namespace PaymentService.Tests;

public class PaymentDomainTests
{
    [Fact]
    public void Payment_Defaults_AreInitialized()
    {
        var payment = new Payment();

        Assert.NotEqual(Guid.Empty, payment.PaymentId);
        Assert.Equal("DKK", payment.Currency);
        Assert.Equal(PaymentStatus.Pending, payment.Status);
        Assert.Equal("MockProvider", payment.Provider);
        Assert.True(payment.CreatedAt <= DateTimeOffset.UtcNow);
    }

    [Theory]
    [InlineData(PaymentStatus.Pending)]
    [InlineData(PaymentStatus.Succeeded)]
    [InlineData(PaymentStatus.Failed)]
    public void PaymentStatus_EnumValues_AreAvailable(PaymentStatus status)
    {
        Assert.Contains(status, Enum.GetValues<PaymentStatus>());
    }
}
