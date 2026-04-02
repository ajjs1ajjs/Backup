using Backup.Contracts;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class JobServiceImpl : JobService.JobServiceBase
{
    private readonly ILogger<JobServiceImpl> _logger;
    private readonly BackupDbContext _db;

    public JobServiceImpl(ILogger<JobServiceImpl> logger, BackupDbContext db)
    {
        _logger = logger;
        _db = db;
    }

    public override async Task<JobResponse> CreateJob(JobRequest request, ServerCallContext context)
    {
        var jobId = Guid.NewGuid().ToString();
        var job = new Job
        {
            JobId = jobId,
            Name = request.Name,
            JobType = request.JobType.ToString(),
            SourceId = request.SourceId,
            SourceType = "vm",
            DestinationId = request.DestinationId,
            Schedule = request.Schedule?.ToString(),
            Options = "{}",
            Enabled = request.Enabled,
            CreatedAt = DateTime.UtcNow
        };
        
        _db.Jobs.Add(job);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created job {JobId}: {Name}", jobId, request.Name);
        
        return new JobResponse
        {
            JobId = jobId,
            Success = true,
            Message = "Job created successfully"
        };
    }

    public override async Task<JobResponse> UpdateJob(JobRequest request, ServerCallContext context)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == request.JobId);
        if (job != null)
        {
            job.Name = request.Name;
            job.DestinationId = request.DestinationId;
            await _db.SaveChangesAsync();
            return new JobResponse { Success = true, Message = "Job updated" };
        }
        return new JobResponse { Success = false, Message = "Job not found" };
    }

    public override async Task<JobResponse> DeleteJob(DeleteJobRequest request, ServerCallContext context)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == request.JobId);
        if (job != null)
        {
            _db.Jobs.Remove(job);
            await _db.SaveChangesAsync();
            _logger.LogInformation("Deleted job {JobId}", request.JobId);
            return new JobResponse { Success = true, Message = "Job deleted" };
        }
        return new JobResponse { Success = false, Message = "Job not found" };
    }

    public override async Task<JobDetails> GetJob(GetJobRequest request, ServerCallContext context)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == request.JobId);
        if (job == null)
            throw new RpcException(new Status(StatusCode.NotFound, "Job not found"));

        return new JobDetails
        {
            JobId = job.JobId,
            JobType = Enum.Parse<Backup.Contracts.JobType>(job.JobType),
            Status = Backup.Contracts.JobStatus.JobStatusPending,
            Name = job.Name,
            CreatedAt = new DateTimeOffset(job.CreatedAt).ToUnixTimeSeconds()
        };
    }

    public override async Task<JobListResponse> ListJobs(JobListRequest request, ServerCallContext context)
    {
        var query = _db.Jobs.AsQueryable();
        
        if (request.FilterStatus != Backup.Contracts.JobStatus.JobStatusUnspecified)
            query = query.Where(j => j.Enabled);
            
        var jobs = await query.Skip((request.Page - 1) * request.PageSize)
            .Take(request.PageSize)
            .Select(j => new JobDetails
            {
                JobId = j.JobId,
                Name = j.Name,
                JobType = Enum.Parse<Backup.Contracts.JobType>(j.JobType),
                Status = j.Enabled ? Backup.Contracts.JobStatus.JobStatusPending : Backup.Contracts.JobStatus.JobStatusCancelled,
                CreatedAt = new DateTimeOffset(j.CreatedAt).ToUnixTimeSeconds()
            })
            .ToListAsync();

        return new JobListResponse
        {
            Jobs = { jobs },
            TotalCount = await query.CountAsync(),
            Page = request.Page,
            PageSize = request.PageSize
        };
    }

    public override async Task<JobResponse> RunJob(RunJobRequest request, ServerCallContext context)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == request.JobId);
        if (job != null)
        {
            job.LastRun = DateTime.UtcNow;
            await _db.SaveChangesAsync();
            _logger.LogInformation("Started job {JobId}", request.JobId);
            return new JobResponse { Success = true, Message = "Job started" };
        }
        return new JobResponse { Success = false, Message = "Job not found" };
    }

    public override async Task<JobResponse> StopJob(StopJobRequest request, ServerCallContext context)
    {
        _logger.LogInformation("Stopped job {JobId}", request.JobId);
        return new JobResponse { Success = true, Message = "Job stopped" };
    }
}
