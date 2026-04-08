using System.Text.Json;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class FastCloneService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<FastCloneService> _logger;

    public FastCloneService(BackupDbContext db, ILogger<FastCloneService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<FastCloneResult> CreateSyntheticFullBackupAsync(string repositoryId, string vmId)
    {
        var result = new FastCloneResult { Success = false };

        try
        {
            var backups = await _db.BackupPoints
                .Where(b => b.VmId == vmId && b.RepositoryId == repositoryId)
                .OrderByDescending(b => b.CreatedAt)
                .ToListAsync();

            var incrementalBackups = backups
                .Where(b => b.BackupType == JobType.Incremental || b.BackupType == JobType.Differential)
                .ToList();

            if (!incrementalBackups.Any())
            {
                result.Message = "No incremental backups found for synthetic full";
                return result;
            }

            var latestBackup = backups.First();
            var repository = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
            if (repository == null || string.IsNullOrWhiteSpace(repository.Path))
            {
                result.Message = "Repository not found";
                return result;
            }

            var cloneResult = await CreateCloneCopyAsync(repository.Path, latestBackup);
            if (!cloneResult.Success)
            {
                return cloneResult;
            }

            var sizeBytes = GetPathSize(cloneResult.NewBackupPath);
            var syntheticBackup = new BackupPoint
            {
                BackupId = Guid.NewGuid().ToString(),
                JobId = latestBackup.JobId,
                VmId = vmId,
                BackupType = JobType.Full,
                RepositoryId = repositoryId,
                FilePath = cloneResult.NewBackupPath,
                SizeBytes = sizeBytes,
                OriginalSizeBytes = latestBackup.OriginalSizeBytes > 0 ? latestBackup.OriginalSizeBytes : sizeBytes,
                IsSynthetic = true,
                ParentBackupId = incrementalBackups.First().BackupId,
                Status = BackupStatus.Completed,
                Metadata = JsonSerializer.Serialize(new Dictionary<string, string>
                {
                    ["synthetic"] = "true",
                    ["sourceBackupId"] = latestBackup.BackupId
                }),
                CreatedAt = DateTime.UtcNow,
                CompletedAt = DateTime.UtcNow
            };

            _db.BackupPoints.Add(syntheticBackup);
            await _db.SaveChangesAsync();

            cloneResult.Message = "Synthetic full backup created";
            _logger.LogInformation("Created synthetic full backup for VM {VmId}", vmId);
            return cloneResult;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to create synthetic full backup");
            result.Message = ex.Message;
            return result;
        }
    }

    private static async Task<FastCloneResult> CreateCloneCopyAsync(string repositoryPath, BackupPoint sourceBackup)
    {
        var sourcePath = ResolveExistingBackupPath(repositoryPath, sourceBackup);
        if (sourcePath == null)
        {
            return new FastCloneResult
            {
                Success = false,
                Message = "Source backup files were not found"
            };
        }

        var extension = File.Exists(sourcePath) ? Path.GetExtension(sourcePath) : string.Empty;
        var destinationPath = Path.Combine(repositoryPath, $"{sourceBackup.BackupId}_synthetic{extension}");

        if (File.Exists(sourcePath))
        {
            Directory.CreateDirectory(Path.GetDirectoryName(destinationPath)!);
            await CopyFileAsync(sourcePath, destinationPath);
        }
        else
        {
            await CopyDirectoryAsync(sourcePath, destinationPath);
        }

        return new FastCloneResult
        {
            Success = true,
            NewBackupPath = destinationPath
        };
    }

    private static string? ResolveExistingBackupPath(string repositoryPath, BackupPoint backup)
    {
        if (!string.IsNullOrWhiteSpace(backup.FilePath))
        {
            if (File.Exists(backup.FilePath) || Directory.Exists(backup.FilePath))
            {
                return backup.FilePath;
            }

            var combined = Path.Combine(repositoryPath, backup.FilePath);
            if (File.Exists(combined) || Directory.Exists(combined))
            {
                return combined;
            }
        }

        var directFile = Path.Combine(repositoryPath, backup.BackupId);
        if (File.Exists(directFile) || Directory.Exists(directFile))
        {
            return directFile;
        }

        var files = Directory.Exists(repositoryPath)
            ? Directory.GetFileSystemEntries(repositoryPath, $"{backup.BackupId}*", SearchOption.TopDirectoryOnly)
            : Array.Empty<string>();

        return files.FirstOrDefault();
    }

    private static async Task CopyFileAsync(string sourcePath, string destinationPath)
    {
        await using var source = new FileStream(sourcePath, FileMode.Open, FileAccess.Read, FileShare.Read);
        await using var destination = new FileStream(destinationPath, FileMode.Create, FileAccess.Write, FileShare.None);
        await source.CopyToAsync(destination);
    }

    private static async Task CopyDirectoryAsync(string sourcePath, string destinationPath)
    {
        Directory.CreateDirectory(destinationPath);

        foreach (var directory in Directory.GetDirectories(sourcePath, "*", SearchOption.AllDirectories))
        {
            Directory.CreateDirectory(directory.Replace(sourcePath, destinationPath));
        }

        foreach (var file in Directory.GetFiles(sourcePath, "*", SearchOption.AllDirectories))
        {
            var targetFile = file.Replace(sourcePath, destinationPath);
            Directory.CreateDirectory(Path.GetDirectoryName(targetFile)!);
            await CopyFileAsync(file, targetFile);
        }
    }

    private static long GetPathSize(string path)
    {
        if (File.Exists(path))
        {
            return new FileInfo(path).Length;
        }

        if (!Directory.Exists(path))
        {
            return 0;
        }

        return Directory
            .GetFiles(path, "*", SearchOption.AllDirectories)
            .Sum(file => new FileInfo(file).Length);
    }
}

public class FastCloneResult
{
    public bool Success { get; set; }
    public string Message { get; set; } = string.Empty;
    public string NewBackupPath { get; set; } = string.Empty;
}

public class RestoreService
{
    private readonly BackupDbContext _db;
    private readonly IAgentManager _agentManager;
    private readonly ICloudStorageService _cloudStorage;
    private readonly IEncryptionService _encryption;
    private readonly INotificationService _notifications;
    private readonly ILogger<RestoreService> _logger;

    public RestoreService(
        BackupDbContext db, 
        IAgentManager agentManager,
        ICloudStorageService cloudStorage,
        IEncryptionService encryption,
        INotificationService notifications,
        ILogger<RestoreService> logger)
    {
        _db = db;
        _agentManager = agentManager;
        _cloudStorage = cloudStorage;
        _encryption = encryption;
        _notifications = notifications;
        _logger = logger;
    }

    public async Task<RestoreQueueResult> QueueRestoreAsync(RestoreQueueRequest request)
    {
        var backup = await _db.BackupPoints.FirstOrDefaultAsync(b => b.BackupId == request.BackupId);
        if (backup == null)
        {
            return new RestoreQueueResult
            {
                Success = false,
                Message = "Backup not found"
            };
        }

        var restoreId = Guid.NewGuid().ToString();
        var restore = new Restore
        {
            RestoreId = restoreId,
            BackupId = request.BackupId,
            RestoreType = string.Equals(request.RestoreType, "instant_restore", StringComparison.OrdinalIgnoreCase)
                ? "instant_restore"
                : "full_vm",
            DestinationPath = request.DestinationPath,
            TargetHost = request.TargetHost,
            Options = JsonSerializer.Serialize(request.Options),
            Status = "queued",
            CreatedAt = DateTime.UtcNow,
            BytesRestored = 0,
            TotalBytes = backup.SizeBytes
        };

        _db.Restores.Add(restore);
        await _db.SaveChangesAsync();

        return new RestoreQueueResult
        {
            Success = true,
            RestoreId = restoreId,
            Message = "Restore queued"
        };
    }

    public async Task<RestoreExecutionResult> ExecuteRestoreAsync(string restoreId, CancellationToken cancellationToken = default)
    {
        var result = new RestoreExecutionResult { Success = false, RestoreId = restoreId };

        try
        {
            var restore = await _db.Restores.FirstOrDefaultAsync(r => r.RestoreId == restoreId, cancellationToken);
            if (restore == null)
            {
                result.Message = "Restore not found";
                return result;
            }

            if (string.Equals(restore.Status, "cancelled", StringComparison.OrdinalIgnoreCase))
            {
                result.Message = "Restore was cancelled before execution";
                return result;
            }

            var backup = await _db.BackupPoints.FirstOrDefaultAsync(b => b.BackupId == restore.BackupId, cancellationToken);
            if (backup == null)
            {
                restore.Status = "failed";
                restore.ErrorMessage = "Backup not found";
                restore.CompletedAt = DateTime.UtcNow;
                await _db.SaveChangesAsync(cancellationToken);
                result.Message = restore.ErrorMessage;
                return result;
            }

            var repository = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == backup.RepositoryId, cancellationToken);
            if (repository == null)
            {
                restore.Status = "failed";
                restore.ErrorMessage = "Repository not found";
                restore.CompletedAt = DateTime.UtcNow;
                await _db.SaveChangesAsync(cancellationToken);
                result.Message = restore.ErrorMessage;
                return result;
            }

            restore.Status = "running";
            restore.StartedAt = DateTime.UtcNow;
            restore.ErrorMessage = null;
            restore.BytesRestored = 0;
            restore.TotalBytes = backup.SizeBytes;
            await _db.SaveChangesAsync(cancellationToken);

            // --- gRPC RESTORE FOR AGENTS ---
            var job = await _db.Jobs.FirstOrDefaultAsync(j => j.JobId == backup.JobId, cancellationToken);
            if (job != null && !string.IsNullOrEmpty(job.AgentId) && _agentManager.IsAgentOnline(job.AgentId))
            {
                var command = new Backup.Contracts.ServerCommand
                {
                    StartRestore = new Backup.Contracts.StartRestoreCommand
                    {
                        JobId = job.JobId,
                        BackupPath = backup.FilePath,
                        DestinationPath = restore.DestinationPath,
                        FullRestore = true
                    }
                };

                if (await _agentManager.SendCommandAsync(job.AgentId, command))
                {
                    _logger.LogInformation("Sent restore command to agent {AgentId} for restore {RestoreId}", job.AgentId, restoreId);
                    result.Success = true;
                    result.Message = "Restore command sent to agent";
                    return result;
                }
            }
            // ------------------------------

            var targetPath = ResolveTargetPath(restore);
            restore.DestinationPath = targetPath;
            await _db.SaveChangesAsync(cancellationToken);
            
            result = await RestoreBackupContentAsync(backup, repository, restore, targetPath, cancellationToken);

            restore.Status = result.Success ? "completed" : (string.Equals(restore.Status, "cancelled", StringComparison.OrdinalIgnoreCase) ? "cancelled" : "failed");
            restore.CompletedAt = DateTime.UtcNow;
            restore.ErrorMessage = result.Success ? null : result.Message;
            await _db.SaveChangesAsync(cancellationToken);

            _logger.LogInformation("Restore {RestoreId} completed with status {Status}", restoreId, restore.Status);

            await _notifications.SendAsync(
                "admin@backupsystem.com",
                $"Restore Completed: {restore.RestoreId}",
                NotificationTemplates.RestoreCompleted(restore.RestoreId, backup.VmId ?? "Unknown VM")
            );

            return result;
        }
        catch (OperationCanceledException)
        {
            var restore = await _db.Restores.FirstOrDefaultAsync(r => r.RestoreId == restoreId, CancellationToken.None);
            if (restore != null && !string.Equals(restore.Status, "cancelled", StringComparison.OrdinalIgnoreCase))
            {
                restore.Status = "cancelled";
                restore.CompletedAt = DateTime.UtcNow;
                restore.ErrorMessage = "Restore cancelled";
                await _db.SaveChangesAsync(CancellationToken.None);
            }

            await _notifications.SendAsync("admin@backupsystem.com", $"Restore CANCELLED: {restoreId}", NotificationTemplates.JobFailed("Restore", "Restore was cancelled"));

            result.Message = "Restore execution cancelled";
            return result;
        }
        catch (Exception ex)
        {
            var restore = await _db.Restores.FirstOrDefaultAsync(r => r.RestoreId == restoreId, CancellationToken.None);
            if (restore != null)
            {
                restore.Status = "failed";
                restore.ErrorMessage = ex.Message;
                restore.CompletedAt = DateTime.UtcNow;
                await _db.SaveChangesAsync(CancellationToken.None);
            }

            _logger.LogError(ex, "Restore execution failed for {RestoreId}", restoreId);

            await _notifications.SendAsync("admin@backupsystem.com", $"Restore FAILED: {restoreId}", NotificationTemplates.JobFailed("Restore", ex.Message));

            result.Message = ex.Message;
            return result;
        }
    }

    private async Task<RestoreExecutionResult> RestoreBackupContentAsync(BackupPoint backup, Repository repository, Restore restore, string targetPath, CancellationToken cancellationToken)
    {
        try
        {
            if (repository.Type == RepositoryType.Local)
            {
                if (string.IsNullOrEmpty(backup.FilePath)) throw new InvalidOperationException("Backup file path is empty");
                var restoredPath = await CopyBackupToDestinationAsync(backup.FilePath, targetPath, restore, cancellationToken);
                return new RestoreExecutionResult { Success = true, RestoreId = restore.RestoreId, Message = "Local restore completed", NewVMId = Path.GetFileName(restoredPath) };
            }
            else if (repository.Type == RepositoryType.S3)
            {
                var bucket = repository.Path;
                var key = backup.FilePath?.Replace($"s3://{bucket}/", string.Empty);
                if (string.IsNullOrEmpty(key)) throw new InvalidOperationException("Invalid S3 key");

                var decryptedCreds = string.IsNullOrEmpty(repository.Credentials) ? null : _encryption.Decrypt(repository.Credentials);
                
                using var cloudStream = await _cloudStorage.DownloadFromS3Async(bucket, key, decryptedCreds, cancellationToken);
                var destinationFile = Path.Combine(targetPath, Path.GetFileName(key));
                Directory.CreateDirectory(targetPath);

                await using var fileStream = new FileStream(destinationFile, FileMode.Create, FileAccess.Write, FileShare.None);
                await cloudStream.CopyToAsync(fileStream, cancellationToken);

                restore.BytesRestored = backup.SizeBytes;
                return new RestoreExecutionResult { Success = true, RestoreId = restore.RestoreId, Message = "S3 restore completed" };
            }
            else if (repository.Type == RepositoryType.AzureBlob)
            {
                var container = repository.Path;
                var blobName = backup.FilePath?.Replace($"azure://{container}/", string.Empty);
                if (string.IsNullOrEmpty(blobName)) throw new InvalidOperationException("Invalid Azure blob name");

                var connectionString = string.IsNullOrEmpty(repository.Credentials) ? null : _encryption.Decrypt(repository.Credentials);
                
                using var cloudStream = await _cloudStorage.DownloadFromAzureBlobAsync(container, blobName, connectionString, cancellationToken);
                var destinationFile = Path.Combine(targetPath, Path.GetFileName(blobName));
                Directory.CreateDirectory(targetPath);

                await using var fileStream = new FileStream(destinationFile, FileMode.Create, FileAccess.Write, FileShare.None);
                await cloudStream.CopyToAsync(fileStream, cancellationToken);

                restore.BytesRestored = backup.SizeBytes;
                return new RestoreExecutionResult { Success = true, RestoreId = restore.RestoreId, Message = "Azure restore completed" };
            }

            return new RestoreExecutionResult { Success = false, Message = $"Repository type {repository.Type} not supported for restore" };
        }
        catch (Exception ex)
        {
            return new RestoreExecutionResult { Success = false, RestoreId = restore.RestoreId, Message = ex.Message };
        }
    }

    private string? ResolveBackupPath(BackupPoint backup)
    {
        if (!string.IsNullOrWhiteSpace(backup.FilePath) && (File.Exists(backup.FilePath) || Directory.Exists(backup.FilePath)))
        {
            return backup.FilePath;
        }

        return null;
    }

    private async Task<string> CopyBackupToDestinationAsync(string sourcePath, string targetPath, Restore restore, CancellationToken cancellationToken)
    {
        if (File.Exists(sourcePath))
        {
            var destination = Directory.Exists(targetPath)
                ? Path.Combine(targetPath, Path.GetFileName(sourcePath))
                : targetPath;

            var destinationDirectory = Path.GetDirectoryName(destination);
            if (!string.IsNullOrWhiteSpace(destinationDirectory))
            {
                Directory.CreateDirectory(destinationDirectory);
            }
            restore.TotalBytes = new FileInfo(sourcePath).Length;
            await _db.SaveChangesAsync(cancellationToken);

            await CopyFileWithProgressAsync(sourcePath, destination, restore, cancellationToken);
            return destination;
        }

        if (!Directory.Exists(sourcePath))
        {
            throw new DirectoryNotFoundException("Restore source directory was not found.");
        }

        var destinationRoot = Path.Combine(targetPath, Path.GetFileName(sourcePath));
        Directory.CreateDirectory(destinationRoot);

        var files = Directory.GetFiles(sourcePath, "*", SearchOption.AllDirectories);
        restore.TotalBytes = files.Sum(file => new FileInfo(file).Length);
        await _db.SaveChangesAsync(cancellationToken);

        foreach (var directory in Directory.GetDirectories(sourcePath, "*", SearchOption.AllDirectories))
        {
            await EnsureRestoreNotCancelledAsync(restore.RestoreId, cancellationToken);
            Directory.CreateDirectory(directory.Replace(sourcePath, destinationRoot));
        }

        foreach (var file in files)
        {
            await EnsureRestoreNotCancelledAsync(restore.RestoreId, cancellationToken);
            var destinationFile = file.Replace(sourcePath, destinationRoot);
            var destinationFileDirectory = Path.GetDirectoryName(destinationFile);
            if (!string.IsNullOrWhiteSpace(destinationFileDirectory))
            {
                Directory.CreateDirectory(destinationFileDirectory);
            }
            await CopyFileWithProgressAsync(file, destinationFile, restore, cancellationToken);
        }

        return destinationRoot;
    }

    private async Task CopyFileWithProgressAsync(string sourcePath, string destinationPath, Restore restore, CancellationToken cancellationToken)
    {
        const int bufferSize = 81920;

        await using var source = new FileStream(sourcePath, FileMode.Open, FileAccess.Read, FileShare.Read);
        await using var destination = new FileStream(destinationPath, FileMode.Create, FileAccess.Write, FileShare.None);

        var buffer = new byte[bufferSize];
        int read;
        while ((read = await source.ReadAsync(buffer.AsMemory(0, buffer.Length), cancellationToken)) > 0)
        {
            await EnsureRestoreNotCancelledAsync(restore.RestoreId, cancellationToken);
            await destination.WriteAsync(buffer.AsMemory(0, read), cancellationToken);
            restore.BytesRestored += read;
            await _db.SaveChangesAsync(cancellationToken);
        }
    }

    private async Task EnsureRestoreNotCancelledAsync(string restoreId, CancellationToken cancellationToken)
    {
        var status = await _db.Restores
            .Where(r => r.RestoreId == restoreId)
            .Select(r => r.Status)
            .FirstOrDefaultAsync(cancellationToken);

        if (string.Equals(status, "cancelled", StringComparison.OrdinalIgnoreCase))
        {
            throw new OperationCanceledException("Restore cancelled.");
        }
    }
}

public class RestoreQueueRequest
{
    public string BackupId { get; set; } = string.Empty;
    public string RestoreType { get; set; } = "full_vm";
    public string TargetHost { get; set; } = string.Empty;
    public string DestinationPath { get; set; } = string.Empty;
    public Dictionary<string, string> Options { get; set; } = new();
}

public class RestoreQueueResult
{
    public bool Success { get; set; }
    public string RestoreId { get; set; } = string.Empty;
    public string Message { get; set; } = string.Empty;
}

public class RestoreExecutionResult
{
    public bool Success { get; set; }
    public string RestoreId { get; set; } = string.Empty;
    public string NewVMId { get; set; } = string.Empty;
    public string Message { get; set; } = string.Empty;
}

public class RestoreFileEntry
{
    public string Name { get; set; } = string.Empty;
    public string Path { get; set; } = string.Empty;
    public bool IsDirectory { get; set; }
    public long SizeBytes { get; set; }
    public DateTime LastModifiedAt { get; set; }
}
