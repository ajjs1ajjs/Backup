using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Backup.Server.BackgroundServices;
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
    private readonly Backup.Server.Services.BackupExecutionService _backupExecutionService;
    private readonly IBackupQueue _backupQueue;

    public JobsController(
        BackupDbContext db,
        ILogger<JobsController> logger,
        Backup.Server.Services.BackupExecutionService backupExecutionService,
        IBackupQueue backupQueue)
    {
        _db = db;
        _logger = logger;
        _backupExecutionService = backupExecutionService;
        _backupQueue = backupQueue;
    }

    [HttpGet]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetJobs([FromQuery] int page = 1, [FromQuery] int pageSize = 20)
    {
        var jobsPage = await _db.Jobs
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync();

        var jobIds = jobsPage.Select(job => job.JobId).ToList();
        var latestRuns = await _db.JobRunHistory
            .Where(run => jobIds.Contains(run.JobId))
            .OrderByDescending(run => run.StartTime)
            .ToListAsync();

        var latestRunsByJobId = latestRuns
            .GroupBy(run => run.JobId)
            .ToDictionary(group => group.Key, group => group.First());

        var jobs = jobsPage
            .Select(job =>
            {
                latestRunsByJobId.TryGetValue(job.JobId, out var latestRun);

                return new
                {
                    job.JobId,
                    job.Name,
                    job.JobType,
                    job.SourceId,
                    job.SourceType,
                    job.DestinationId,
                    job.Schedule,
                    job.Options,
                    job.Enabled,
                    job.LastRun,
                    job.NextRun,
                    job.CreatedAt,
                    LatestRun = latestRun == null ? null : new
                    {
                        latestRun.RunId,
                        latestRun.Status,
                        latestRun.StartTime,
                        latestRun.EndTime,
                        latestRun.BytesProcessed,
                        latestRun.FilesProcessed,
                        latestRun.SpeedMbps,
                        latestRun.ErrorMessage
                    }
                };
            })
            .ToList();

        var total = await _db.Jobs.CountAsync();

        return Ok(new { jobs, total, page, pageSize });
    }

    [HttpGet("{jobId}")]
    public async Task<ActionResult> GetJob(string jobId)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (job == null) return NotFound();
        return Ok(job);
    }

    [HttpPost]
    public async Task<ActionResult> CreateJob([FromBody] JobDto jobDto)
    {
        var job = new Job
        {
            JobId = Guid.NewGuid().ToString(),
            Name = jobDto.Name,
            JobType = Enum.Parse<JobType>(jobDto.JobType, true),
            SourceId = jobDto.SourceId,
            SourceType = jobDto.SourceType,
            DestinationId = jobDto.DestinationId,
            Schedule = jobDto.Schedule,
            Options = string.IsNullOrWhiteSpace(jobDto.Options) ? "{}" : jobDto.Options,
            Enabled = jobDto.Enabled,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };
        
        _db.Jobs.Add(job);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created job {JobId}: {Name}", job.JobId, job.Name);
        return CreatedAtAction(nameof(GetJob), new { jobId = job.JobId }, job);
    }

    [HttpPut("{jobId}")]
    public async Task<ActionResult> UpdateJob(string jobId, [FromBody] JobDto jobDto)
    {
        var existing = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (existing == null) return NotFound();
        
        existing.Name = jobDto.Name;
        existing.JobType = Enum.Parse<JobType>(jobDto.JobType, true);
        existing.SourceId = jobDto.SourceId;
        existing.SourceType = jobDto.SourceType;
        existing.DestinationId = jobDto.DestinationId;
        existing.Schedule = jobDto.Schedule;
        existing.Options = string.IsNullOrWhiteSpace(jobDto.Options) ? "{}" : jobDto.Options;
        existing.Enabled = jobDto.Enabled;
        existing.UpdatedAt = DateTime.UtcNow;
        
        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{jobId}")]
    public async Task<ActionResult> DeleteJob(string jobId)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (job == null) return NotFound();
        
        _db.Jobs.Remove(job);
        await _db.SaveChangesAsync();
        
        return NoContent();
    }

    [HttpPost("{jobId}/run")]
    public async Task<ActionResult> RunJob(string jobId)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (job == null) return NotFound();

        var queueResult = await _backupExecutionService.QueueJobAsync(jobId);
        if (!queueResult.Success)
        {
            return BadRequest(new { message = queueResult.Message, runId = queueResult.RunId });
        }

        await _backupQueue.QueueAsync(queueResult.RunId);

        _logger.LogInformation("Queued job {JobId} as run {RunId}", jobId, queueResult.RunId);
        return Ok(new { runId = queueResult.RunId, message = queueResult.Message, status = "queued" });
    }

    [HttpPost("{jobId}/stop")]
    public async Task<ActionResult> StopJob(string jobId)
    {
        var activeRun = await _db.JobRunHistory
            .Where(r => r.JobId == jobId && (r.Status == "running" || r.Status == "queued"))
            .OrderByDescending(r => r.StartTime)
            .FirstOrDefaultAsync();
        
        if (activeRun != null)
        {
            activeRun.Status = "cancelled";
            activeRun.EndTime = DateTime.UtcNow;
            await _db.SaveChangesAsync();
            return Ok(new { message = "Job cancelled", runId = activeRun.RunId });
        }
        
        return Ok(new { message = "No queued or running job was found" });
    }

    [HttpGet("{jobId}/runs")]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetJobRuns(string jobId, [FromQuery] int page = 1, [FromQuery] int pageSize = 20)
    {
        var jobExists = await _db.Jobs.AnyAsync(j => j.JobId == jobId);
        if (!jobExists) return NotFound();

        var query = _db.JobRunHistory
            .Where(r => r.JobId == jobId)
            .OrderByDescending(r => r.StartTime);

        var total = await query.CountAsync();
        var runs = await query
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .Select(r => new
            {
                r.RunId,
                r.JobId,
                r.StartTime,
                r.EndTime,
                r.Status,
                r.BytesProcessed,
                r.FilesProcessed,
                r.SpeedMbps,
                r.ErrorMessage
            })
            .ToListAsync();

        return Ok(new { runs, total, page, pageSize });
    }

    [HttpGet("runs/{runId}")]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetRun(string runId)
    {
        var run = await _db.JobRunHistory
            .Select(r => new
            {
                r.RunId,
                r.JobId,
                r.StartTime,
                r.EndTime,
                r.Status,
                r.BytesProcessed,
                r.FilesProcessed,
                r.SpeedMbps,
                r.ErrorMessage
            })
            .FirstOrDefaultAsync(r => r.RunId == runId);

        if (run == null) return NotFound();
        return Ok(run);
    }
}

public class JobDto
{
    public string Name { get; set; } = string.Empty;
    public string JobType { get; set; } = "Full";
    public string SourceId { get; set; } = string.Empty;
    public string SourceType { get; set; } = "VirtualMachine";
    public string DestinationId { get; set; } = string.Empty;
    public string? Schedule { get; set; }
    public string Options { get; set; } = "{}";
    public bool Enabled { get; set; } = true;
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
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == agentId);
        if (agent == null) return NotFound();
        return Ok(agent);
    }

    [HttpDelete("{agentId}")]
    public async Task<ActionResult> DeleteAgent(long agentId)
    {
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == agentId);
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
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null) return NotFound();
        return Ok(repo);
    }

    [HttpPost]
    public async Task<ActionResult> CreateRepository([FromBody] RepositoryDto repositoryDto)
    {
        var repository = new Repository
        {
            RepositoryId = Guid.NewGuid().ToString(),
            Name = repositoryDto.Name,
            Type = Enum.Parse<RepositoryType>(repositoryDto.Type, true),
            Path = repositoryDto.Path,
            Status = string.IsNullOrWhiteSpace(repositoryDto.Status) ? "online" : repositoryDto.Status,
            CapacityBytes = repositoryDto.CapacityBytes,
            Credentials = repositoryDto.Credentials,
            Options = string.IsNullOrWhiteSpace(repositoryDto.Options) ? "{}" : repositoryDto.Options,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };
        
        _db.Repositories.Add(repository);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created repository {RepoId}: {Name}", repository.RepositoryId, repository.Name);
        return CreatedAtAction(nameof(GetRepository), new { repositoryId = repository.RepositoryId }, repository);
    }

    [HttpPost("{repositoryId}/test")]
    public async Task<ActionResult> TestRepository(string repositoryId, [FromServices] Services.RepositoryService repoService)
    {
        var success = await repoService.TestConnectionAsync(repositoryId);
        if (success)
            return Ok(new { success = true, message = "Connection successful" });
        else
            return BadRequest(new { success = false, message = "Connection failed" });
    }

    [HttpDelete("{repositoryId}")]
    public async Task<ActionResult> DeleteRepository(string repositoryId)
    {
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null) return NotFound();
        
        _db.Repositories.Remove(repo);
        await _db.SaveChangesAsync();
        
        return NoContent();
    }
}

public class RepositoryDto
{
    public string Name { get; set; } = string.Empty;
    public string Type { get; set; } = "Local";
    public string Path { get; set; } = string.Empty;
    public string? Status { get; set; }
    public long? CapacityBytes { get; set; }
    public string? Credentials { get; set; }
    public string? Options { get; set; }
}
