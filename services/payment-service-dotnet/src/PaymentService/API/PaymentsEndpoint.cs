using PaymentService.Infrastructure.DB;
using Microsoft.EntityFrameworkCore;

namespace PaymentService.API
{
    public static class PaymentsEndpoints
    {
        public static void MapPayments(this WebApplication app)
        {
            app.MapGet("/api/payments/by-order/{orderId:guid}", async (Guid orderId, PaymentDbContext db) =>
            {
                var p = await db.Payments.SingleOrDefaultAsync(x => x.OrderId == orderId);
                return p is null ? Results.NotFound() : Results.Ok(p);
            });

            app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
        }
    }
}
