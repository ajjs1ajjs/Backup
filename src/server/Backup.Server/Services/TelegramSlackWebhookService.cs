using System.Net.Http.Json;

namespace Backup.Server.Services;

public interface INotificationService
{
    Task SendAsync(string recipient, string subject, string body);
    Task SendBatchAsync(List<string> recipients, string subject, string body);
}

public class EmailNotificationService : INotificationService
{
    private readonly ILogger<EmailNotificationService> _logger;
    public EmailNotificationService(ILogger<EmailNotificationService> logger) => _logger = logger;
    public Task SendAsync(string recipient, string subject, string body) => Task.CompletedTask;
    public Task SendBatchAsync(List<string> recipients, string subject, string body) => Task.CompletedTask;
}

public class TelegramNotificationService : INotificationService
{
    private readonly ILogger<TelegramNotificationService> _logger;
    private readonly HttpClient _httpClient = new();
    private readonly string _botToken;
    private readonly string _chatId;

    public TelegramNotificationService(ILogger<TelegramNotificationService> logger, IConfiguration config)
    {
        _logger = logger;
        _botToken = config["Telegram:BotToken"] ?? "";
        _chatId = config["Telegram:ChatId"] ?? "";
    }

    public async Task SendAsync(string recipient, string subject, string body)
    {
        if (string.IsNullOrEmpty(_botToken) || string.IsNullOrEmpty(_chatId)) return;
        var url = $"https://api.telegram.org/bot{_botToken}/sendMessage";
        try
        {
            await _httpClient.PostAsJsonAsync(url, new { chat_id = _chatId, text = $"<b>{subject}</b>\n\n{body}", parse_mode = "HTML" });
            _logger.LogInformation("Telegram notification sent");
        }
        catch (Exception ex) { _logger.LogError(ex, "Failed to send Telegram notification"); }
    }
    public Task SendBatchAsync(List<string> recipients, string subject, string body) => SendAsync("", subject, body);
}

public class SlackNotificationService : INotificationService
{
    private readonly ILogger<SlackNotificationService> _logger;
    private readonly HttpClient _httpClient = new();
    private readonly string _webhookUrl;

    public SlackNotificationService(ILogger<SlackNotificationService> logger, IConfiguration config)
    {
        _logger = logger;
        _webhookUrl = config["Slack:WebhookUrl"] ?? "";
    }

    public async Task SendAsync(string recipient, string subject, string body)
    {
        if (string.IsNullOrEmpty(_webhookUrl)) return;
        try
        {
            await _httpClient.PostAsJsonAsync(_webhookUrl, new { channel = "#backups", username = "Backup System", text = $"*{subject}*\n\n{body}" });
            _logger.LogInformation("Slack notification sent");
        }
        catch (Exception ex) { _logger.LogError(ex, "Failed to send Slack notification"); }
    }
    public Task SendBatchAsync(List<string> recipients, string subject, string body) => SendAsync("", subject, body);
}

public class WebhookNotificationService : INotificationService
{
    private readonly ILogger<WebhookNotificationService> _logger;
    private readonly HttpClient _httpClient = new();

    public WebhookNotificationService(ILogger<WebhookNotificationService> logger) => _logger = logger;

    public async Task SendAsync(string webhookUrl, string subject, string body)
    {
        if (string.IsNullOrEmpty(webhookUrl)) return;
        try
        {
            await _httpClient.PostAsJsonAsync(webhookUrl, new { event = subject, message = body, timestamp = DateTime.UtcNow });
            _logger.LogInformation("Webhook notification sent to {Url}", webhookUrl);
        }
        catch (Exception ex) { _logger.LogError(ex, "Failed to send webhook notification"); }
    }
    public Task SendBatchAsync(List<string> webhookUrls, string subject, string body) => Task.CompletedTask;
}
