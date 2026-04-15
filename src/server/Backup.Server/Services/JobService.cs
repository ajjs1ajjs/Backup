using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public interface IJobService
{
    Task<(List<object> Jobs, int Total)> GetJobsAsync(int page, int pageSize);
    Task<Job?> GetJobByIdAsync(string jobId);
    Task<Job> CreateJobAsync(Controllers.JobDto jobDto);
    Task<Job?> UpdateJobAsync(string jobId, Controllers.JobDto jobDto);
    Task<bool> DeleteJobAsync(string jobId);
    Task<(List<object> Runs, int Total)> GetJobRunsAsync(string jobId, int page, int pageSize);
    Task<object?> GetRunByIdAsync(string runId);
}

public class JobService : IJobService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<JobService> _logger;

    public JobService(BackupDbContext db, ILogger<JobService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<(List<object> Jobs, int Total)> GetJobsAsync(int page, int pageSize)
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

                return (object)new
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
        return (jobs, total);
    }

    public async Task<Job?> GetJobByIdAsync(string jobId)
    {
        return await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
    }

    public async Task<Job> CreateJobAsync(Controllers.JobDto jobDto)
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
        return job;
    }

    public async Task<Job?> UpdateJobAsync(string jobId, Controllers.JobDto jobDto)
    {
        var existing = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (existing == null) return null;

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
        return existing;
    }

    public async Task<bool> DeleteJobAsync(string jobId)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId);
        if (job == null) return false;

        _db.Jobs.Remove(job);
        await _db.SaveChangesAsync();
        return true;
    }

    public async Task<(List<object> Runs, int Total)> GetJobRunsAsync(string jobId, int page, int pageSize)
    {
        var query = _db.JobRunHistory
            .Where(r => r.JobId == jobId)
            .OrderByDescending(r => r.StartTime);

        var total = await query.CountAsync();
        var runs = await query
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .Select(r => (object)new
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

        return (runs, total);
    }

    public async Task<object?> GetRunByIdAsync(string runId)
    {
        return await _db.JobRunHistory
            .Where(r => r.RunId == runId)
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
            .FirstOrDefaultAsync();
    }
}
