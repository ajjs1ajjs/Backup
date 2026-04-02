using Backup.Server.Database;
using Backup.Server.Services;
using Backup.Server.BackgroundServices;
using Microsoft.EntityFrameworkCore;
using Serilog;
using Swashbuckle.AspNetCore.SwaggerGen;
using AspNetCoreRateLimit;

Log.Logger = new LoggerConfiguration()
    .WriteTo.Console()
    .CreateLogger();

try
{
    Log.Information("Starting Backup Server...");

    var builder = WebApplication.CreateBuilder(args);

    builder.Host.UseSerilog();

    var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
        ?? "Host=localhost;Database=backup;Username=postgres;Password=postgres";

    builder.Services.AddDbContext<BackupDbContext>(options =>
        options.UseNpgsql(connectionString));

    builder.Services.AddControllers();
    builder.Services.AddEndpointsApiExplorer();
    builder.Services.AddSwaggerGen(c =>
    {
        c.SwaggerDoc("v1", new() { Title = "Backup API", Version = "v1" });
    });

    builder.Services.AddGrpc();
    builder.Services.AddCors(o => o.AddDefaultPolicy(p => p
        .AllowAnyOrigin()
        .AllowAnyMethod()
        .AllowAnyHeader()));

    builder.Services.AddScoped<IAgentRegistry, AgentRegistry>();

    builder.Services.AddHostedService<JobSchedulerService>();
    builder.Services.AddHostedService<AgentHealthCheckService>();
    builder.Services.AddHostedService<RetentionPolicyService>();

    builder.Services.Configure<IpRateLimitOptions>(options =>
    {
        options.GeneralRules = new List<RateLimitRule>
        {
            new() { Endpoint = "*", Period = "1m", Limit = 100 },
            new() { Endpoint = "*/api/*", Period = "1s", Limit = 10 }
        };
    });
    builder.Services.AddSingleton<IRateLimitCounterStore, MemoryCacheRateLimitCounterStore>();
    builder.Services.AddSingleton<IIpPolicyStore, MemoryCacheIpPolicyStore>();
    builder.Services.AddSingleton<IRateLimitConfiguration, RateLimitConfiguration>();
    builder.Services.AddSingleton<IProcessingStrategy, AsyncKeyLockProcessingStrategy>();
    builder.Services.AddInMemoryRateLimiting();

    var app = builder.Build();

    app.UseSwagger();
    app.UseSwaggerUI(c =>
    {
        c.SwaggerEndpoint("/swagger/v1/swagger.json", "Backup API v1");
        c.RoutePrefix = "swagger";
    });

    using (var scope = app.Services.CreateScope())
    {
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
        db.Database.EnsureCreated();
        Log.Information("Database initialized");
    }

    app.UseSerilogRequestLogging();
    app.UseCors();
    app.UseIpRateLimiting();

    app.MapGrpcService<AgentServiceImpl>();
    app.MapGrpcService<JobServiceImpl>();
    app.MapGrpcService<BackupServiceImpl>();
    app.MapGrpcService<RestoreServiceImpl>();
    app.MapGrpcService<RepositoryServiceImpl>();
    app.MapGrpcService<FileTransferServiceImpl>();
    app.MapGrpcService<LogServiceImpl>();
    app.MapGrpcService<DashboardServiceImpl>();

    app.MapControllers();

    app.MapGet("/", () => "Backup Server v1.0.0");
    app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));

    app.Run();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Server terminated unexpectedly");
    return 1;
}
finally
{
    Log.CloseAndFlush();
}

return 0;
