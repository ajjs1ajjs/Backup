using System.Security.Cryptography;
using System.Text.Json;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class BackupExecutionService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<BackupExecutionService> _logger;

    public BackupExecutionService(BackupDbContext db, ILogger<BackupExecutionService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<BackupQueueResult> QueueJobAsync(string jobId, CancellationToken cancellationToken = default)
    {
        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == jobId, cancellationToken);
        if (job == null)
        {
            return new BackupQueueResult { Success = false, Message = "Job not found" };
        }

        var existingRun = await _db.JobRunHistory
            .FirstOrDefaultAsync(r => r.JobId == jobId && (r.Status == "queued" || r.Status == "running"), cancellationToken);
        if (existingRun != null)
        {
            return new BackupQueueResult
            {
                Success = false,
                RunId = existingRun.RunId,
                Message = "Job is already queued or running"
            };
        }

        var runHistory = new JobRunHistory
        {
            RunId = Guid.NewGuid().ToString(),
            JobId = jobId,
            StartTime = DateTime.UtcNow,
            Status = "queued"
        };

        _db.JobRunHistory.Add(runHistory);
        job.LastRun = DateTime.UtcNow;
        await _db.SaveChangesAsync(cancellationToken);

        return new BackupQueueResult
        {
            Success = true,
            RunId = runHistory.RunId,
            Message = "Backup job queued"
        };
    }

    public async Task<BackupExecutionResult> ExecuteRunAsync(string runId, CancellationToken cancellationToken = default)
    {
        var result = new BackupExecutionResult { Success = false, RunId = runId };

        var run = await _db.JobRunHistory.FirstOrDefaultAsync(r => r.RunId == runId, cancellationToken);
        if (run == null)
        {
            result.Message = "Run history not found";
            return result;
        }

        if (string.Equals(run.Status, "cancelled", StringComparison.OrdinalIgnoreCase))
        {
            result.Message = "Backup run was cancelled before execution";
            return result;
        }

        var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == run.JobId, cancellationToken);
        if (job == null)
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = "Job not found";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        if (job.JobType is JobType.Restore or JobType.Replication)
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = $"Job type {job.JobType} is not supported by the backup execution pipeline";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        var repository = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == job.DestinationId, cancellationToken);
        if (repository == null)
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = "Repository not found";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        var sourcePath = ResolveSourcePath(job);
        if (sourcePath == null)
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = "Backup source path not found";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        if (repository.Type != RepositoryType.Local)
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = $"Repository type {repository.Type} is not supported by the current backup executor";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        if (string.IsNullOrWhiteSpace(repository.Path))
        {
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = "Repository path is not configured";
            await _db.SaveChangesAsync(cancellationToken);
            result.Message = run.ErrorMessage;
            return result;
        }

        run.Status = "running";
        run.ErrorMessage = null;
        run.BytesProcessed = 0;
        run.FilesProcessed = 0;
        run.SpeedMbps = 0;
        await _db.SaveChangesAsync(cancellationToken);

        var backupPoint = new BackupPoint
        {
            BackupId = Guid.NewGuid().ToString(),
            JobId = job.JobId,
            VmId = string.Equals(job.SourceType, "VirtualMachine", StringComparison.OrdinalIgnoreCase) ? job.SourceId : null,
            BackupType = job.JobType,
            RepositoryId = repository.RepositoryId,
            Status = BackupStatus.InProgress,
            Metadata = JsonSerializer.Serialize(new Dictionary<string, string>
            {
                ["sourceId"] = job.SourceId,
                ["sourceType"] = job.SourceType,
                ["runId"] = run.RunId
            }),
            CreatedAt = DateTime.UtcNow
        };

        _db.BackupPoints.Add(backupPoint);
        await _db.SaveChangesAsync(cancellationToken);

        var startedAt = DateTime.UtcNow;

        try
        {
            Directory.CreateDirectory(repository.Path);
            var targetPath = BuildDestinationPath(repository.Path, backupPoint.BackupId, sourcePath);
            await EnsureRunNotCancelledAsync(run.RunId, cancellationToken);

            if (File.Exists(sourcePath))
            {
                backupPoint.OriginalSizeBytes = new FileInfo(sourcePath).Length;
                await _db.SaveChangesAsync(cancellationToken);
                await CopyFileWithProgressAsync(sourcePath, targetPath, run, cancellationToken);
            }
            else
            {
                backupPoint.OriginalSizeBytes = Directory
                    .GetFiles(sourcePath, "*", SearchOption.AllDirectories)
                    .Sum(file => new FileInfo(file).Length);
                await _db.SaveChangesAsync(cancellationToken);
                await CopyDirectoryWithProgressAsync(sourcePath, targetPath, run, cancellationToken);
            }

            backupPoint.FilePath = targetPath;
            backupPoint.SizeBytes = GetPathSize(targetPath);
            backupPoint.Checksum = await ComputeChecksumAsync(targetPath, cancellationToken);
            backupPoint.Status = BackupStatus.Completed;
            backupPoint.CompletedAt = DateTime.UtcNow;

            run.Status = "completed";
            run.EndTime = DateTime.UtcNow;
            run.SpeedMbps = CalculateSpeedMbps(run.BytesProcessed, startedAt, run.EndTime.Value);

            job.LastRun = DateTime.UtcNow;

            var vm = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == job.SourceId, cancellationToken);
            if (vm != null)
            {
                vm.LastBackupAt = DateTime.UtcNow;
                vm.LastBackupId = backupPoint.BackupId;
            }

            repository.LastUsedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync(cancellationToken);

            result.Success = true;
            result.BackupId = backupPoint.BackupId;
            result.Message = "Backup completed successfully";
            return result;
        }
        catch (OperationCanceledException)
        {
            backupPoint.Status = BackupStatus.Failed;
            backupPoint.CompletedAt = DateTime.UtcNow;
            run.Status = "cancelled";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = "Backup cancelled";
            await _db.SaveChangesAsync(CancellationToken.None);

            result.Message = "Backup cancelled";
            return result;
        }
        catch (Exception ex)
        {
            backupPoint.Status = BackupStatus.Failed;
            backupPoint.CompletedAt = DateTime.UtcNow;
            run.Status = "failed";
            run.EndTime = DateTime.UtcNow;
            run.ErrorMessage = ex.Message;
            await _db.SaveChangesAsync(cancellationToken);

            _logger.LogError(ex, "Backup run {RunId} failed", runId);
            result.Message = ex.Message;
            return result;
        }
    }

    private static string? ResolveSourcePath(Job job)
    {
        if (File.Exists(job.SourceId) || Directory.Exists(job.SourceId))
        {
            return job.SourceId;
        }

        return null;
    }

    private static string BuildDestinationPath(string repositoryPath, string backupId, string sourcePath)
    {
        var safeName = Path.GetFileName(sourcePath);
        if (File.Exists(sourcePath))
        {
            var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(sourcePath);
            var extension = Path.GetExtension(sourcePath);
            var targetFileName = string.IsNullOrWhiteSpace(fileNameWithoutExtension)
                ? $"{backupId}{extension}"
                : $"{backupId}_{fileNameWithoutExtension}{extension}";
            return Path.Combine(repositoryPath, targetFileName);
        }

        var directoryName = string.IsNullOrWhiteSpace(safeName) ? backupId : $"{backupId}_{safeName}";
        return Path.Combine(repositoryPath, directoryName);
    }

    private async Task CopyDirectoryWithProgressAsync(string sourcePath, string destinationPath, JobRunHistory run, CancellationToken cancellationToken)
    {
        Directory.CreateDirectory(destinationPath);

        foreach (var directory in Directory.GetDirectories(sourcePath, "*", SearchOption.AllDirectories))
        {
            await EnsureRunNotCancelledAsync(run.RunId, cancellationToken);
            Directory.CreateDirectory(directory.Replace(sourcePath, destinationPath));
        }

        foreach (var file in Directory.GetFiles(sourcePath, "*", SearchOption.AllDirectories))
        {
            await EnsureRunNotCancelledAsync(run.RunId, cancellationToken);
            var destinationFile = file.Replace(sourcePath, destinationPath);
            var destinationDirectory = Path.GetDirectoryName(destinationFile);
            if (!string.IsNullOrWhiteSpace(destinationDirectory))
            {
                Directory.CreateDirectory(destinationDirectory);
            }

            await CopyFileWithProgressAsync(file, destinationFile, run, cancellationToken);
        }
    }

    private async Task CopyFileWithProgressAsync(string sourcePath, string destinationPath, JobRunHistory run, CancellationToken cancellationToken)
    {
        const int bufferSize = 81920;

        await using var source = new FileStream(sourcePath, FileMode.Open, FileAccess.Read, FileShare.Read);
        await using var destination = new FileStream(destinationPath, FileMode.Create, FileAccess.Write, FileShare.None);

        var buffer = new byte[bufferSize];
        int read;
        while ((read = await source.ReadAsync(buffer.AsMemory(0, buffer.Length), cancellationToken)) > 0)
        {
            await EnsureRunNotCancelledAsync(run.RunId, cancellationToken);
            await destination.WriteAsync(buffer.AsMemory(0, read), cancellationToken);
            run.BytesProcessed += read;
            await _db.SaveChangesAsync(cancellationToken);
        }

        run.FilesProcessed += 1;
        await _db.SaveChangesAsync(cancellationToken);
    }

    private static long GetPathSize(string path)
    {
        if (File.Exists(path))
        {
            return new FileInfo(path).Length;
        }

        return Directory.Exists(path)
            ? Directory.GetFiles(path, "*", SearchOption.AllDirectories).Sum(file => new FileInfo(file).Length)
            : 0;
    }

    private static async Task<string?> ComputeChecksumAsync(string path, CancellationToken cancellationToken)
    {
        if (!File.Exists(path))
        {
            return null;
        }

        using var sha256 = SHA256.Create();
        await using var stream = new FileStream(path, FileMode.Open, FileAccess.Read, FileShare.Read);
        var hash = await sha256.ComputeHashAsync(stream, cancellationToken);
        return Convert.ToHexString(hash);
    }

    private static double CalculateSpeedMbps(long bytesProcessed, DateTime startedAt, DateTime endedAt)
    {
        var seconds = Math.Max((endedAt - startedAt).TotalSeconds, 0.001);
        return bytesProcessed * 8d / 1_000_000d / seconds;
    }

    private async Task EnsureRunNotCancelledAsync(string runId, CancellationToken cancellationToken)
    {
        var status = await _db.JobRunHistory
            .Where(run => run.RunId == runId)
            .Select(run => run.Status)
            .FirstOrDefaultAsync(cancellationToken);

        if (string.Equals(status, "cancelled", StringComparison.OrdinalIgnoreCase))
        {
            throw new OperationCanceledException("Backup run was cancelled.");
        }
    }
}

public class BackupQueueResult
{
    public bool Success { get; set; }
    public string RunId { get; set; } = string.Empty;
    public string Message { get; set; } = string.Empty;
}

public class BackupExecutionResult
{
    public bool Success { get; set; }
    public string RunId { get; set; } = string.Empty;
    public string? BackupId { get; set; }
    public string Message { get; set; } = string.Empty;
}
