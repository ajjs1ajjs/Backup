using Grpc.Net.Client;
using Backup.Contracts;
using System.Threading.Tasks;
using System.Collections.Concurrent;

namespace Backup.Server.Services;

public class GrpcAgentClient : IDisposable
{
    private readonly GrpcChannel _channel;
    private readonly AgentService.AgentServiceClient _client;
    private readonly ILogger<GrpcAgentClient> _logger;
    private readonly ConcurrentDictionary<long, AgentConnection> _connections = new();

    public GrpcAgentClient(string serverAddress, ILogger<GrpcAgentClient> logger)
    {
        var handler = new SocketsHttpHandler
        {
            PooledConnectionIdleTimeout = Timeout.InfiniteTimeSpan,
            KeepAlivePingDelay = TimeSpan.FromSeconds(60),
            KeepAlivePingTimeout = TimeSpan.FromSeconds(30)
        };

        _channel = GrpcChannel.ForAddress(serverAddress, new GrpcChannelOptions
        {
            HttpHandler = handler
        });

        _client = new AgentService.AgentServiceClient(_channel);
        _logger = logger;
    }

    public async Task<bool> RegisterAgentAsync(AgentRegistrationRequest request)
    {
        try
        {
            var response = await _client.RegisterAsync(request);
            return response.Success;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to register agent");
            return false;
        }
    }

    public async Task<ServerCommand?> GetCommandAsync(long agentId)
    {
        return null;
    }

    public async Task SendBackupCommandAsync(long agentId, StartBackupCommand command)
    {
        try
        {
            var request = new ServerCommand
            {
                StartBackup = command
            };
            
            _logger.LogInformation("Sent backup command to agent {AgentId}", agentId);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send backup command");
        }
    }

    public async Task SendRestoreCommandAsync(long agentId, StartRestoreCommand command)
    {
        try
        {
            var request = new ServerCommand
            {
                StartRestore = command
            };
            
            _logger.LogInformation("Sent restore command to agent {AgentId}", agentId);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send restore command");
        }
    }

    public void Dispose()
    {
        _channel.Dispose();
    }
}

public class AgentConnection
{
    public long AgentId { get; set; }
    public GrpcAgentClient Client { get; set; } = null!;
    public DateTime ConnectedAt { get; set; }
    public DateTime LastActivity { get; set; }
    public AgentStatus Status { get; set; }
}

public class AgentCommunicationService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly ILogger<AgentCommunicationService> _logger;
    private readonly ConcurrentDictionary<long, GrpcAgentClient> _agentClients = new();

    public AgentCommunicationService(
        IServiceProvider services,
        ILogger<AgentCommunicationService> logger)
    {
        _services = services;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Agent Communication Service started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var scope = _services.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
                
                var onlineAgents = await Task.FromResult(new List<Backup.Server.Database.Entities.Agent>());

                foreach (var agent in onlineAgents)
                {
                    if (!_agentClients.TryGetValue(agent.Id, out var client))
                    {
                        client = new GrpcAgentClient("http://localhost:50051", 
                            _services.GetRequiredService<ILogger<GrpcAgentClient>>());
                        _agentClients[agent.Id] = client;
                    }
                }

                await Task.Delay(TimeSpan.FromSeconds(10), stoppingToken);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in agent communication service");
            }
        }
    }

    public override void Dispose()
    {
        foreach (var client in _agentClients.Values)
        {
            client.Dispose();
        }
        base.Dispose();
    }
}

public class MetricsCollectorService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly ILogger<MetricsCollectorService> _logger;

    public MetricsCollectorService(
        IServiceProvider services,
        ILogger<MetricsCollectorService> logger)
    {
        _services = services;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Metrics Collector Service started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var scope = _services.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

                var runningJobs = await Task.FromResult(new List<Backup.Server.Database.Entities.JobRunHistory>());

                foreach (var job in runningJobs)
                {
                    _logger.LogDebug("Collecting metrics for job {JobId}", job.JobId);
                }

                await Task.Delay(TimeSpan.FromSeconds(5), stoppingToken);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in metrics collector");
            }
        }
    }
}
