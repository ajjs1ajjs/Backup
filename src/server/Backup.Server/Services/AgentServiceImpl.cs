using Backup.Contracts;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class AgentServiceImpl : AgentService.AgentServiceBase
{
    private readonly ILogger<AgentServiceImpl> _logger;
    private readonly BackupDbContext _db;

    public AgentServiceImpl(ILogger<AgentServiceImpl> logger, BackupDbContext db)
    {
        _logger = logger;
        _db = db;
    }

    public override async Task<AgentRegistrationResponse> Register(
        AgentRegistrationRequest request,
        ServerCallContext context)
    {
        _logger.LogInformation("Agent registration request from {Hostname}", request.Hostname);

        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.AgentId == request.AgentId);
        
        if (agent == null)
        {
            agent = new Agent
            {
                AgentId = request.AgentId,
                Hostname = request.Hostname,
                OsType = request.OsType,
                AgentVersion = request.AgentVersion,
                AgentType = request.AgentType.ToString(),
                Status = "idle",
                Capabilities = System.Text.Json.JsonSerializer.Serialize(
                    request.Capabilities.ToList())
            };
            _db.Agents.Add(agent);
        }
        else
        {
            agent.Hostname = request.Hostname;
            agent.OsType = request.OsType;
            agent.LastHeartbeat = DateTime.UtcNow;
        }

        await _db.SaveChangesAsync();

        _logger.LogInformation("Agent registered: {AgentId}", agent.AgentId);

        return new AgentRegistrationResponse
        {
            Success = true,
            Message = "Agent registered successfully",
            ServerVersion = "1.0.0",
            AssignedAgentId = agent.Id
        };
    }

    public override async Task Heartbeat(
        IAsyncStreamReader<AgentHeartbeat> requestStream,
        IServerStreamWriter<ServerCommand> responseStream,
        ServerCallContext context)
    {
        await foreach (var heartbeat in requestStream.ReadAllAsync())
        {
            _logger.LogDebug("Heartbeat from agent {AgentId}, status: {Status}",
                heartbeat.AgentId, heartbeat.Status);

            var agent = await _db.Agents.FindAsync(heartbeat.AgentId);
            if (agent != null)
            {
                agent.Status = heartbeat.Status.ToString();
                agent.LastHeartbeat = DateTime.UtcNow;
                await _db.SaveChangesAsync();
            }
        }
    }

    public override Task<AgentCapabilitiesResponse> GetCapabilities(
        AgentCapabilitiesRequest request,
        ServerCallContext context)
    {
        return Task.FromResult(new AgentCapabilitiesResponse
        {
            Features = { "backup", "restore", "compression", "deduplication" },
            Options =
            {
                { "max_block_size", "8192" },
                { "supported_compression", "zstd,lz4,gzip" }
            }
        });
    }
}
