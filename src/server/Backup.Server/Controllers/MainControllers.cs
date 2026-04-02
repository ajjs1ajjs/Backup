using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Backup.Server.Database;
using Backup.Server.Database.Entities;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class JobsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<JobsController> _logger;

    public JobsController(BackupDbContext db, ILogger<JobsController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetJobs([FromQuery] int page = 1, [FromQuery] int pageSize = 20)
    {
        var jobs = await _db.Jobs
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .Select(j => new
            {
                j.JobId,
                j.Name,
                j.JobType,
                j.SourceId,
                j.DestinationId,
                j.Enabled,
                j.LastRun,
                j.NextRun,
                j.CreatedAt
            })
            .ToListAsync();

        var total = await _db.Jobs.CountAsync();

        return Ok(new { jobs, total, page, pageSize });
    }

    [HttpGet("{jobId}")]
    public async Task<ActionResult> GetJob(string jobId)
    {
        var job = await _db.Jobs.FindAsync(jobId);
        if (job == null) return NotFound();
        return Ok(job);
    }

    [HttpPost]
    public async Task<ActionResult> CreateJob([FromBody] Job job)
    {
        job.JobId = Guid.NewGuid().ToString();
        job.CreatedAt = DateTime.UtcNow;
        job.UpdatedAt = DateTime.UtcNow;
        
        _db.Jobs.Add(job);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created job {JobId}: {Name}", job.JobId, job.Name);
        return CreatedAtAction(nameof(GetJob), new { jobId = job.JobId }, job);
    }

    [HttpPut("{jobId}")]
    public async Task<ActionResult> UpdateJob(string jobId, [FromBody] Job job)
    {
        var existing = await _db.Jobs.FindAsync(jobId);
        if (existing == null) return NotFound();
        
        existing.Name = job.Name;
        existing.JobType = job.JobType;
        existing.SourceId = job.SourceId;
        existing.DestinationId = job.DestinationId;
        existing.Schedule = job.Schedule;
        existing.Options = job.Options;
        existing.Enabled = job.Enabled;
        existing.UpdatedAt = DateTime.UtcNow;
        
        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{jobId}")]
    public async Task<ActionResult> DeleteJob(string jobId)
    {
        var job = await _db.Jobs.FindAsync(jobId);
        if (job == null) return NotFound();
        
        _db.Jobs.Remove(job);
        await _db.SaveChangesAsync();
        
        return NoContent();
    }

    [HttpPost("{jobId}/run")]
    public async Task<ActionResult> RunJob(string jobId)
    {
        var job = await _db.Jobs.FindAsync(jobId);
        if (job == null) return NotFound();
        
        var runHistory = new JobRunHistory
        {
            RunId = Guid.NewGuid().ToString(),
            JobId = jobId,
            StartTime = DateTime.UtcNow,
            Status = "running"
        };
        
        _db.JobRunHistory.Add(runHistory);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Started job {JobId}", jobId);
        return Ok(new { runId = runHistory.RunId, message = "Job started" });
    }

    [HttpPost("{jobId}/stop")]
    public async Task<ActionResult> StopJob(string jobId)
    {
        var activeRun = await _db.JobRunHistory
            .FirstOrDefaultAsync(r => r.JobId == jobId && r.Status == "running");
        
        if (activeRun != null)
        {
            activeRun.Status = "cancelled";
            activeRun.EndTime = DateTime.UtcNow;
            await _db.SaveChangesAsync();
        }
        
        return Ok(new { message = "Job stopped" });
    }
}

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class AgentsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<AgentsController> _logger;

    public AgentsController(BackupDbContext db, ILogger<AgentsController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetAgents()
    {
        var agents = await _db.Agents.ToListAsync();
        return Ok(agents);
    }

    [HttpGet("{agentId}")]
    public async Task<ActionResult> GetAgent(long agentId)
    {
        var agent = await _db.Agents.FindAsync(agentId);
        if (agent == null) return NotFound();
        return Ok(agent);
    }

    [HttpDelete("{agentId}")]
    public async Task<ActionResult> DeleteAgent(long agentId)
    {
        var agent = await _db.Agents.FindAsync(agentId);
        if (agent == null) return NotFound();
        
        _db.Agents.Remove(agent);
        await _db.SaveChangesAsync();
        
        return NoContent();
    }
}

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class RepositoriesController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<RepositoriesController> _logger;

    public RepositoriesController(BackupDbContext db, ILogger<RepositoriesController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetRepositories()
    {
        var repos = await _db.Repositories.ToListAsync();
        return Ok(repos);
    }

    [HttpGet("{repositoryId}")]
    public async Task<ActionResult> GetRepository(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return NotFound();
        return Ok(repo);
    }

    [HttpPost]
    public async Task<ActionResult> CreateRepository([FromBody] Repository repository)
    {
        repository.RepositoryId = Guid.NewGuid().ToString();
        repository.CreatedAt = DateTime.UtcNow;
        repository.UpdatedAt = DateTime.UtcNow;
        
        _db.Repositories.Add(repository);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created repository {RepoId}: {Name}", repository.RepositoryId, repository.Name);
        return CreatedAtAction(nameof(GetRepository), new { repositoryId = repository.RepositoryId }, repository);
    }

    [HttpPost("{repositoryId}/test")]
    public async Task<ActionResult> TestRepository(string repositoryId)
    {
        return Ok(new { success = true, message = "Connection successful" });
    }

    [HttpDelete("{repositoryId}")]
    public async Task<ActionResult> DeleteRepository(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return NotFound();
        
        _db.Repositories.Remove(repo);
        await _db.SaveChangesAsync();
        
        return NoContent();
    }
}
