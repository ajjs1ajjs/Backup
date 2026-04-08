using Backup.Contracts;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using System.Collections.Concurrent;

namespace Backup.Server.Services;

public interface IAgentManager
{
    Task<bool> SendCommandAsync(string agentId, ServerCommand command);
    bool IsAgentOnline(string agentId);
}

public class AgentGrpcService : AgentService.AgentServiceBase, IAgentManager
{
    private readonly IServiceProvider _serviceProvider;
    private readonly ILogger<AgentGrpcService> _logger;
    private static readonly ConcurrentDictionary<string, IServerStreamWriter<ServerCommand>> _activeAgents = new();

    public AgentGrpcService(IServiceProvider serviceProvider, ILogger<AgentGrpcService> logger)
    {
        _serviceProvider = serviceProvider;
        _logger = logger;
    }

    public override async Task<AgentRegistrationResponse> Register(AgentRegistrationRequest request, ServerCallContext context)
    {
        using var scope = _serviceProvider.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

        var agent = await db.Agents.FirstOrDefaultAsync(a => a.AgentId == request.AgentId);
        if (agent == null)
        {
            agent = new Agent
            {
                AgentId = request.AgentId,
                Hostname = request.Hostname,
                OsType = request.OsType,
                AgentType = request.AgentType.ToString(),
                AgentVersion = request.AgentVersion,
                CreatedAt = DateTime.UtcNow
            };
            db.Agents.Add(agent);
        }
        else
        {
            agent.Hostname = request.Hostname;
            agent.OsType = request.OsType;
            agent.AgentVersion = request.AgentVersion;
            agent.UpdatedAt = DateTime.UtcNow;
        }

        agent.IpAddress = context.Peer;
        agent.Status = "idle";
        agent.LastHeartbeat = DateTime.UtcNow;
        agent.Capabilities = System.Text.Json.JsonSerializer.Serialize(request.Capabilities);

        await db.SaveChangesAsync();

        return new AgentRegistrationResponse
        {
            Success = true,
            Message = "Registration successful",
            ServerVersion = "1.0.0",
            AssignedAgentId = agent.Id
        };
    }

    public override async Task Heartbeat(IAsyncStreamReader<AgentHeartbeat> requestStream, IServerStreamWriter<ServerCommand> responseStream, ServerCallContext context)
    {
        string? agentId = null;

        try
        {
            while (await requestStream.MoveNext())
            {
                var heartbeat = requestStream.Current;
                
                using var scope = _serviceProvider.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
                var agent = await db.Agents.FirstOrDefaultAsync(a => a.Id == heartbeat.AgentId);

                if (agent != null)
                {
                    agentId = agent.AgentId;
                    _activeAgents[agentId] = responseStream;

                    agent.LastHeartbeat = DateTime.UtcNow;
                    agent.Status = heartbeat.Status.ToString();
                    await db.SaveChangesAsync();
                }

                _logger.LogDebug("Heartbeat received from agent {AgentId}", heartbeat.AgentId);
                await responseStream.WriteAsync(new ServerCommand { Ping = new PingCommand { Timestamp = DateTimeOffset.UtcNow.ToUnixTimeSeconds() } });
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error in heartbeat stream for agent {AgentId}", agentId);
        }
        finally
        {
            if (agentId != null)
            {
                _activeAgents.TryRemove(agentId, out _);
                
                using var scope = _serviceProvider.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
                var agent = await db.Agents.FirstOrDefaultAsync(a => a.AgentId == agentId);
                if (agent != null)
                {
                    agent.Status = "offline";
                    await db.SaveChangesAsync();
                }
            }
        }
    }

    public async Task<bool> SendCommandAsync(string agentId, ServerCommand command)
    {
        if (_activeAgents.TryGetValue(agentId, out var responseStream))
        {
            try 
            {
                await responseStream.WriteAsync(command);
                return true;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to send command to agent {AgentId}", agentId);
                _activeAgents.TryRemove(agentId, out _);
            }
        }
        return false;
    }

    public bool IsAgentOnline(string agentId)
    {
        return _activeAgents.ContainsKey(agentId);
    }
}
