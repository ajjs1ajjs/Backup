namespace Backup.Server.Services;

public interface INotificationService
{
    Task SendAsync(string recipient, string subject, string body, CancellationToken ct = default);
}

public class NotificationServiceStub : INotificationService
{
    private readonly ILogger<NotificationServiceStub> _logger;

    public NotificationServiceStub(ILogger<NotificationServiceStub> logger)
    {
        _logger = logger;
    }

    public Task SendAsync(string recipient, string subject, string body, CancellationToken ct = default)
    {
        _logger.LogDebug("Notification (stub): To={Recipient}, Subject={Subject}", recipient, subject);
        return Task.CompletedTask;
    }
}
