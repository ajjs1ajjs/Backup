namespace Backup.Server.Services;

public class NotificationServiceStub : INotificationService
{
    private readonly ILogger<NotificationServiceStub> _logger;

    public NotificationServiceStub(ILogger<NotificationServiceStub> logger)
    {
        _logger = logger;
    }

    public Task SendAsync(string recipient, string subject, string body)
    {
        _logger.LogDebug("Notification (stub): To={Recipient}, Subject={Subject}", recipient, subject);
        return Task.CompletedTask;
    }

    public Task SendBatchAsync(List<string> recipients, string subject, string body)
    {
        _logger.LogDebug("Notification batch (stub): Count={Count}, Subject={Subject}", recipients.Count, subject);
        return Task.CompletedTask;
    }
}
