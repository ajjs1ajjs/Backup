using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public interface IAgentManagementService
{
    Task<List<Agent>> GetAgentsAsync();
    Task<Agent?> GetAgentByIdAsync(long id);
    Task<bool> DeleteAgentAsync(long id);
}

public class AgentManagementService : IAgentManagementService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<AgentManagementService> _logger;

    public AgentManagementService(BackupDbContext db, ILogger<AgentManagementService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<List<Agent>> GetAgentsAsync()
    {
        return await _db.Agents.ToListAsync();
    }

    public async Task<Agent?> GetAgentByIdAsync(long id)
    {
        return await _db.Agents.FirstOrDefaultAsync(a => a.Id == id);
    }

    public async Task<bool> DeleteAgentAsync(long id)
    {
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == id);
        if (agent == null) return false;

        _db.Agents.Remove(agent);
        await _db.SaveChangesAsync();
        _logger.LogInformation("Deleted agent {AgentId} (Internal ID: {Id})", agent.AgentId, id);
        return true;
    }
}
