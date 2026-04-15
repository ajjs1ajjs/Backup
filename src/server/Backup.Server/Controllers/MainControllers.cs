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
    private readonly IJobService _jobService;
    private readonly ILogger<JobsController> _logger;
    private readonly Backup.Server.Services.BackupExecutionService _backupExecutionService;
    private readonly IBackupQueue _backupQueue;
    private readonly BackupDbContext _db; // Still needed for StopJob until moved to service

    public JobsController(
        IJobService jobService,
        ILogger<JobsController> logger,
        Backup.Server.Services.BackupExecutionService backupExecutionService,
        IBackupQueue backupQueue,
        BackupDbContext db)
    {
        _jobService = jobService;
        _logger = logger;
        _backupExecutionService = backupExecutionService;
        _backupQueue = backupQueue;
        _db = db;
    }

    [HttpGet]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetJobs([FromQuery] int page = 1, [FromQuery] int pageSize = 20)
    {
        var (jobs, total) = await _jobService.GetJobsAsync(page, pageSize);
        return Ok(new { jobs, total, page, pageSize });
    }

    [HttpGet("{jobId}")]
    public async Task<ActionResult> GetJob(string jobId)
    {
        var job = await _jobService.GetJobByIdAsync(jobId);
        if (job == null) return NotFound();
        return Ok(job);
    }

    [HttpPost]
    public async Task<ActionResult> CreateJob([FromBody] JobDto jobDto)
    {
        var job = await _jobService.CreateJobAsync(jobDto);
        return CreatedAtAction(nameof(GetJob), new { jobId = job.JobId }, job);
    }

    [HttpPut("{jobId}")]
    public async Task<ActionResult> UpdateJob(string jobId, [FromBody] JobDto jobDto)
    {
        var result = await _jobService.UpdateJobAsync(jobId, jobDto);
        if (result == null) return NotFound();
        return Ok(result);
    }

    [HttpDelete("{jobId}")]
    public async Task<ActionResult> DeleteJob(string jobId)
    {
        var success = await _jobService.DeleteJobAsync(jobId);
        if (!success) return NotFound();
        return NoContent();
    }

    [HttpPost("{jobId}/run")]
    public async Task<ActionResult> RunJob(string jobId)
    {
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
        // Keeping this logic here for now as it involves complex state management better handled in ExecutionService later
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
        var (runs, total) = await _jobService.GetJobRunsAsync(jobId, page, pageSize);
        return Ok(new { runs, total, page, pageSize });
    }

    [HttpGet("runs/{runId}")]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetRun(string runId)
    {
        var run = await _jobService.GetRunByIdAsync(runId);
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
    private readonly IAgentService _agentService;
    private readonly ILogger<AgentsController> _logger;

    public AgentsController(IAgentService agentService, ILogger<AgentsController> logger)
    {
        _agentService = agentService;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetAgents()
    {
        var agents = await _agentService.GetAgentsAsync();
        return Ok(agents);
    }

    [HttpGet("{agentId}")]
    public async Task<ActionResult> GetAgent(long agentId)
    {
        var agent = await _agentService.GetAgentByIdAsync(agentId);
        if (agent == null) return NotFound();
        return Ok(agent);
    }

    [HttpDelete("{agentId}")]
    public async Task<ActionResult> DeleteAgent(long agentId)
    {
        var success = await _agentService.DeleteAgentAsync(agentId);
        if (!success) return NotFound();
        return NoContent();
    }
}

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class RepositoriesController : ControllerBase
{
    private readonly IRepositoryService _repoService;
    private readonly ILogger<RepositoriesController> _logger;

    public RepositoriesController(IRepositoryService repoService, ILogger<RepositoriesController> logger)
    {
        _repoService = repoService;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetRepositories()
    {
        var repos = await _repoService.GetRepositoriesAsync();
        return Ok(repos);
    }

    [HttpGet("{repositoryId}")]
    public async Task<ActionResult> GetRepository(string repositoryId)
    {
        var repo = await _repoService.GetRepositoryByIdAsync(repositoryId);
        if (repo == null) return NotFound();
        return Ok(repo);
    }

    [HttpPost]
    public async Task<ActionResult> CreateRepository([FromBody] RepositoryDto repositoryDto)
    {
        var repository = await _repoService.CreateRepositoryAsync(repositoryDto);
        return CreatedAtAction(nameof(GetRepository), new { repositoryId = repository.RepositoryId }, repository);
    }

    [HttpPost("{repositoryId}/test")]
    public async Task<ActionResult> TestRepository(string repositoryId)
    {
        var success = await _repoService.TestConnectionAsync(repositoryId);
        if (success)
            return Ok(new { success = true, message = "Connection successful" });
        else
            return BadRequest(new { success = false, message = "Connection failed" });
    }

    [HttpDelete("{repositoryId}")]
    public async Task<ActionResult> DeleteRepository(string repositoryId)
    {
        var success = await _repoService.DeleteRepositoryAsync(repositoryId);
        if (!success) return NotFound();
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
