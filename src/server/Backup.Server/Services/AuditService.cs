using Backup.Server.Database;
using Backup.Server.Database.Entities;
using System.Text.Json;

namespace Backup.Server.Services;

public interface IAuditService
{
    Task LogAsync(string? userId, string action, string entityType, string? entityId, object? details = null, string? ipAddress = null);
}

public class AuditService : IAuditService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly ILogger<AuditService> _logger;

    public AuditService(IServiceProvider serviceProvider, ILogger<AuditService> logger)
    {
        _serviceProvider = serviceProvider;
        _logger = logger;
    }

    public async Task LogAsync(string? userId, string action, string entityType, string? entityId, object? details = null, string? ipAddress = null)
    {
        try
        {
            using var scope = _serviceProvider.CreateScope();
            var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

            var log = new AuditLog
            {
                UserId = userId,
                Action = action,
                EntityType = entityType,
                EntityId = entityId,
                Details = details != null ? JsonSerializer.Serialize(details) : "{}",
                IpAddress = ipAddress,
                CreatedAt = DateTime.UtcNow
            };

            db.AuditLogs.Add(log);
            await db.SaveChangesAsync();

            _logger.LogInformation("Audit Log: {Action} on {EntityType} {EntityId} by User {UserId}", action, entityType, entityId, userId);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to save audit log");
        }
    }
}
