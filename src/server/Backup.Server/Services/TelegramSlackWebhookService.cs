using System.Net.Http.Json;
using System.Text.Json;

namespace Backup.Server.Services;

public class TelegramNotificationService : INotificationService
{
    private readonly ILogger<TelegramNotificationService> _logger;
    private readonly HttpClient _httpClient;
    private readonly string _botToken;
    private readonly string _chatId;

    public TelegramNotificationService(
        ILogger<TelegramNotificationService> logger,
        IConfiguration config)
    {
        _logger = logger;
        _httpClient = new HttpClient();
        _botToken = config["Telegram:BotToken"] ?? "";
        _chatId = config["Telegram:ChatId"] ?? "";
    }

    public async Task SendAsync(string recipient, string subject, string body)
    {
        if (string.IsNullOrEmpty(_botToken) || string.IsNullOrEmpty(_chatId))
        {
            _logger.LogWarning("Telegram bot token or chat ID not configured");
            return;
        }

        var message = $"<b>{subject}</b>\n\n{body}";
        
        var url = $"https://api.telegram.org/bot{_botToken}/sendMessage";
        
        try
        {
            var response = await _httpClient.PostAsJsonAsync(url, new
            {
                chat_id = _chatId,
                text = message,
                parse_mode = "HTML"
            });

            if (response.IsSuccessStatusCode)
            {
                _logger.LogInformation("Telegram notification sent");
            }
            else
            {
                var error = await response.Content.ReadAsStringAsync();
                _logger.LogError("Failed to send Telegram notification: {Error}", error);
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send Telegram notification");
        }
    }

    public async Task SendBatchAsync(List<string> recipients, string subject, string body)
    {
        foreach (var recipient in recipients)
        {
            await SendAsync(recipient, subject, body);
        }
    }

    public async Task SendWithInlineButtons(string subject, string body, Dictionary<string, string> buttons)
    {
        if (string.IsNullOrEmpty(_botToken) || string.IsNullOrEmpty(_chatId))
            return;

        var url = $"https://api.telegram.org/bot{_botToken}/sendMessage";

        var inlineButtons = buttons.Select(b => new[]
        {
            new { text = b.Key, url = b.Value }
        }).ToList();

        try
        {
            await _httpClient.PostAsJsonAsync(url, new
            {
                chat_id = _chatId,
                text = $"<b>{subject}</b>\n\n{body}",
                parse_mode = "HTML",
                reply_markup = new
                {
                    inline_keyboard = inlineButtons
                }
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send Telegram notification with buttons");
        }
    }

    public async Task SendPhotoAsync(string caption, byte[] photoData, string fileName)
    {
        using var content = new MultipartFormDataContent();
        
        var imageContent = new ByteArrayContent(photoData);
        imageContent.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("image/jpeg");
        content.Add(imageContent, "photo", fileName);
        
        content.Add(new StringContent(_chatId), "chat_id");
        content.Add(new StringContent(caption), "caption");

        var url = $"https://api.telegram.org/bot{_botToken}/sendPhoto";
        
        try
        {
            await _httpClient.PostAsync(url, content);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send Telegram photo");
        }
    }
}

public class SlackNotificationService : INotificationService
{
    private readonly ILogger<SlackNotificationService> _logger;
    private readonly HttpClient _httpClient;
    private readonly string _webhookUrl;
    private readonly string _channel;

    public SlackNotificationService(
        ILogger<SlackNotificationService> logger,
        IConfiguration config)
    {
        _logger = logger;
        _httpClient = new HttpClient();
        _webhookUrl = config["Slack:WebhookUrl"] ?? "";
        _channel = config["Slack:Channel"] ?? "#backups";
    }

    public async Task SendAsync(string recipient, string subject, string body)
    {
        if (string.IsNullOrEmpty(_webhookUrl))
        {
            _logger.LogWarning("Slack webhook URL not configured");
            return;
        }

        var payload = new
        {
            channel = _channel,
            username = "Backup System",
            icon_emoji = ":backup:",
            attachments = new[]
            {
                new
                {
                    color = subject.Contains("Failed") ? "danger" : "good",
                    title = subject,
                    text = body,
                    footer = "Backup System",
                    ts = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
                }
            }
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync(_webhookUrl, payload);
            if (response.IsSuccessStatusCode)
            {
                _logger.LogInformation("Slack notification sent");
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send Slack notification");
        }
    }

    public async Task SendBatchAsync(List<string> recipients, string subject, string body)
    {
        await SendAsync("", subject, body);
    }

    public async Task SendWithBlocksAsync(string subject, string body, object[] blocks)
    {
        if (string.IsNullOrEmpty(_webhookUrl))
            return;

        var payload = new
        {
            channel = _channel,
            username = "Backup System",
            blocks = blocks
        };

        try
        {
            await _httpClient.PostAsJsonAsync(_webhookUrl, payload);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send Slack notification with blocks");
        }
    }
}

public class WebhookNotificationService : INotificationService
{
    private readonly ILogger<WebhookNotificationService> _logger;
    private readonly HttpClient _httpClient;

    public WebhookNotificationService(ILogger<WebhookNotificationService> logger)
    {
        _logger = logger;
        _httpClient = new HttpClient();
    }

    public async Task SendAsync(string webhookUrl, string subject, string body)
    {
        if (string.IsNullOrEmpty(webhookUrl))
            return;

        var payload = new
        {
            event = subject,
            message = body,
            timestamp = DateTime.UtcNow
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync(webhookUrl, payload);
            _logger.LogInformation("Webhook notification sent to {Url}", webhookUrl);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send webhook notification");
        }
    }

    public async Task SendBatchAsync(List<string> webhookUrls, string subject, string body)
    {
        foreach (var url in webhookUrls)
        {
            await SendAsync(url, subject, body);
        }
    }
}

public class NotificationManager
{
    private readonly EmailNotificationService _emailService;
    private readonly TelegramNotificationService _telegramService;
    private readonly SlackNotificationService _slackService;
    private readonly WebhookNotificationService _webhookService;
    private readonly ILogger<NotificationManager> _logger;

    public NotificationManager(
        EmailNotificationService emailService,
        TelegramNotificationService telegramService,
        SlackNotificationService slackService,
        WebhookNotificationService webhookService,
        ILogger<NotificationManager> logger)
    {
        _emailService = emailService;
        _telegramService = telegramService;
        _slackService = slackService;
        _webhookService = webhookService;
        _logger = logger;
    }

    public async Task NotifyJobCompletedAsync(string jobName, string recipient, Dictionary<string, string> channels)
    {
        var body = NotificationTemplates.JobCompleted(jobName, "1h 23m", "45.2 GB");

        if (channels.ContainsKey("email"))
            await _emailService.SendAsync(recipient, $"Backup Completed: {jobName}", body);
        
        if (channels.ContainsKey("telegram"))
            await _telegramService.SendAsync("", $"Backup Completed: {jobName}", body);
        
        if (channels.ContainsKey("slack"))
            await _slackService.SendAsync("", $"Backup Completed: {jobName}", body);
    }

    public async Task NotifyJobFailedAsync(string jobName, string error, string recipient, Dictionary<string, string> channels)
    {
        var body = NotificationTemplates.JobFailed(jobName, error);

        if (channels.ContainsKey("email"))
            await _emailService.SendAsync(recipient, $"Backup Failed: {jobName}", body);
        
        if (channels.ContainsKey("telegram"))
            await _telegramService.SendAsync("", $"Backup Failed: {jobName}", body);
        
        if (channels.ContainsKey("slack"))
            await _slackService.SendAsync("", $"Backup Failed: {jobName}", body);
    }

    public async Task NotifyStorageAlertAsync(string repositoryName, string usagePercent, string recipient)
    {
        var body = NotificationTemplates.StorageAlert(repositoryName, usagePercent);
        await _emailService.SendAsync(recipient, "Storage Alert", body);
    }
}
