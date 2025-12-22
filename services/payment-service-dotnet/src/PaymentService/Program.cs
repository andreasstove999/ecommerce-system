using Microsoft.EntityFrameworkCore;
using PaymentService.API;
using PaymentService.Infrastructure.DB;
using PaymentService.Infrastructure.Messaging;
using RabbitMQ.Client;

var builder = WebApplication.CreateBuilder(args);

builder.Services.Configure<RabbitMqOptions>(builder.Configuration.GetSection("RabbitMQ"));

builder.Services.AddDbContext<PaymentDbContext>(opt =>
{
    var cs = builder.Configuration.GetConnectionString("PaymentDb");
    if (string.IsNullOrWhiteSpace(cs))
        throw new InvalidOperationException("Connection string 'PaymentDb' is missing.");

    opt.UseNpgsql(cs);
});

builder.Services.AddSingleton<IConnection>(_ =>
{
    var opt = builder.Configuration.GetSection("RabbitMQ").Get<RabbitMqOptions>() ?? new RabbitMqOptions();
    var factory = new ConnectionFactory { Uri = new Uri(opt.Url) };

    // RabbitMQ.Client v7 is async-first. Block once during startup to create the singleton connection.
    return factory.CreateConnectionAsync().GetAwaiter().GetResult();
});

builder.Services.AddSingleton<EventPublisher>(sp =>
{
    var opt = builder.Configuration.GetSection("RabbitMQ").Get<RabbitMqOptions>() ?? new RabbitMqOptions();
    var conn = sp.GetRequiredService<IConnection>();
    return new EventPublisher(opt, conn);
});

builder.Services.AddHostedService<OrderCreatedConsumer>();

var app = builder.Build();

// Dev-friendly schema setup: create if missing.
// If you later add EF migrations, replace with: db.Database.Migrate();
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<PaymentDbContext>();
    db.Database.EnsureCreated();
}

app.MapPayments();

app.Run();
