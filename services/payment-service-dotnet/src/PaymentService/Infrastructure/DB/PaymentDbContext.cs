using Microsoft.EntityFrameworkCore;
using PaymentService.Domain;

namespace PaymentService.Infrastructure.DB;

public sealed class PaymentDbContext : DbContext
{
    public PaymentDbContext(DbContextOptions<PaymentDbContext> options) : base(options) { }

    public DbSet<Payment> Payments => Set<Payment>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Payment>()
            .HasIndex(p => p.OrderId)
            .IsUnique();

        base.OnModelCreating(modelBuilder);
    }
}
