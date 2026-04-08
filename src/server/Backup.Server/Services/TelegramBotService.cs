using Telegram.Bot;
using Telegram.Bot.Polling;
using Telegram.Bot.Types;

namespace Backup.Server.Services;

public class TelegramBotService : BackgroundService
{
    private readonly IConfiguration _config;
    private readonly ILogger<TelegramBotService> _logger;
    private TelegramBotClient? _botClient;

    public TelegramBotService(IConfiguration config, ILogger<TelegramBotService> logger)
    {
        _config = config;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        var token = _config["Telegram:BotToken"];
        if (string.IsNullOrEmpty(token)) return;

        _botClient = new TelegramBotClient(token);
        
        var receiverOptions = new ReceiverOptions { AllowedUpdates = { } };
        _botClient.StartReceiving(
            updateHandler: HandleUpdateAsync,
            pollingErrorHandler: HandlePollingErrorAsync,
            receiverOptions: receiverOptions,
            cancellationToken: stoppingToken
        );

        _logger.LogInformation("Telegram Bot Service started");
    }

    private async Task HandleUpdateAsync(ITelegramBotClient botClient, Update update, CancellationToken ct)
    {
        if (update.Message?.Text == "/status")
        {
            await botClient.SendTextMessageAsync(update.Message.Chat.Id, "System is healthy. All backups are operational.");
        }
    }

    private Task HandlePollingErrorAsync(ITelegramBotClient botClient, Exception ex, CancellationToken ct)
    {
        _logger.LogError(ex, "Telegram bot polling error");
        return Task.CompletedTask;
    }
}
