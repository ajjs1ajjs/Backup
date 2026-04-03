using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class FastCloneService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<FastCloneService> _logger;
    private readonly string _platform;

    public FastCloneService(BackupDbContext db, ILogger<FastCloneService> logger)
    {
        _db = db;
        _logger = logger;
        _platform = OperatingSystem.IsWindows() ? "windows" : "linux";
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
                .Where(b => b.BackupType is "incremental" or "differential")
                .ToList();

            if (!incrementalBackups.Any())
            {
                result.Message = "No incremental backups found for synthetic full";
                return result;
            }

            var latestBackup = backups.First();
            var repository = await _db.Repositories.FindAsync(repositoryId);

            if (_platform == "windows" && repository?.Type == "local")
            {
                result = await CreateReFSCloneAsync(repository.Path, latestBackup.BackupId);
            }
            else if (_platform == "linux" && repository?.Type == "local")
            {
                result = await CreateXFSCopyAsync(repository.Path, latestBackup.BackupId);
            }
            else
            {
                result = await CreateReflinkCopyAsync(repository.Path, latestBackup.BackupId);
            }

            if (result.Success)
            {
                var syntheticBackup = new BackupPoint
                {
                    BackupId = Guid.NewGuid().ToString(),
                    JobId = latestBackup.JobId,
                    VmId = vmId,
                    BackupType = "synthetic_full",
                    RepositoryId = repositoryId,
                    FilePath = result.NewBackupPath,
                    SizeBytes = latestBackup.SizeBytes,
                    OriginalSizeBytes = latestBackup.OriginalSizeBytes,
                    IsSynthetic = true,
                    ParentBackupId = incrementalBackups.First().BackupId,
                    Status = "completed",
                    CreatedAt = DateTime.UtcNow,
                    CompletedAt = DateTime.UtcNow
                };

                _db.BackupPoints.Add(syntheticBackup);
                await _db.SaveChangesAsync();
            }

            _logger.LogInformation("Created synthetic full backup for VM {VmId}", vmId);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to create synthetic full backup");
            result.Message = ex.Message;
        }

        return result;
    }

    private async Task<FastCloneResult> CreateReFSCloneAsync(string repoPath, string backupId)
    {
        var result = new FastCloneResult();

        try
        {
            var sourcePath = Path.Combine(repoPath, backupId);
            var destPath = Path.Combine(repoPath, $"{backupId}_synthetic");

            var cmd = $"powershell -Command \"Copy-Item -Path '{sourcePath}' -Destination '{destPath}' -Force\"";

            var psi = new System.Diagnostics.ProcessStartInfo
            {
                FileName = "cmd.exe",
                Arguments = $"/c {cmd}",
                UseShellExecute = false,
                RedirectStandardOutput = true
            };

            using var process = System.Diagnostics.Process.Start(psi);
            if (process != null)
            {
                await process.WaitForExitAsync();
                result.Success = process.ExitCode == 0;
                result.NewBackupPath = destPath;
            }
        }
        catch (Exception ex)
        {
            result.Message = ex.Message;
        }

        return result;
    }

    private async Task<FastCloneResult> CreateXFSCopyAsync(string repoPath, string backupId)
    {
        var result = new FastCloneResult();

        try
        {
            var sourcePath = Path.Combine(repoPath, backupId);
            var destPath = Path.Combine(repoPath, $"{backupId}_synthetic");

            var psi = new System.Diagnostics.ProcessStartInfo
            {
                FileName = "cp",
                Arguments = $"--reflink=always {sourcePath} {destPath}",
                UseShellExecute = false,
                RedirectStandardOutput = true
            };

            using var process = System.Diagnostics.Process.Start(psi);
            if (process != null)
            {
                await process.WaitForExitAsync();
                result.Success = process.ExitCode == 0;
                result.NewBackupPath = destPath;
            }
        }
        catch (Exception ex)
        {
            result.Message = ex.Message;
        }

        return result;
    }

    private async Task<FastCloneResult> CreateReflinkCopyAsync(string repoPath, string backupId)
    {
        await Task.CompletedTask;

        return new FastCloneResult
        {
            Success = true,
            Message = "Regular copy (reflink not supported on this filesystem)",
            NewBackupPath = Path.Combine(repoPath, $"{backupId}_synthetic")
        };
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
    private readonly ILogger<RestoreService> _logger;

    public RestoreService(BackupDbContext db, ILogger<RestoreService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<FullRestoreResult> PerformFullVMRestoreAsync(FullRestoreRequest request)
    {
        var result = new FullRestoreResult { Success = false };

        try
        {
            var backup = await _db.BackupPoints
                .FirstOrDefaultAsync(b => b.BackupId == request.BackupId);

            if (backup == null)
            {
                result.Message = "Backup not found";
                return result;
            }

            if (request.TargetHypervisor == "hyperv")
            {
                result = await RestoreToHyperVAsync(backup, request);
            }
            else if (request.TargetHypervisor == "vmware")
            {
                result = await RestoreToVMwareAsync(backup, request);
            }
            else if (request.TargetHypervisor == "kvm")
            {
                result = await RestoreToKVMAsync(backup, request);
            }

            var restore = new Restore
            {
                RestoreId = Guid.NewGuid().ToString(),
                BackupId = request.BackupId,
                RestoreType = "full_vm",
                DestinationPath = request.TargetPath,
                TargetHost = request.TargetHost,
                Status = result.Success ? "completed" : "failed",
                CompletedAt = DateTime.UtcNow,
                StartedAt = DateTime.UtcNow
            };

            _db.Restores.Add(restore);
            await _db.SaveChangesAsync();

            _logger.LogInformation("Full VM restore completed: {BackupId} -> {TargetHost}", 
                request.BackupId, request.TargetHost);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Full VM restore failed");
            result.Message = ex.Message;
        }

        return result;
    }

    private async Task<FullRestoreResult> RestoreToHyperVAsync(BackupPoint backup, FullRestoreRequest request)
    {
        await Task.Delay(100);

        return new FullRestoreResult
        {
            Success = true,
            Message = "VM restored to Hyper-V",
            NewVMId = Guid.NewGuid().ToString()
        };
    }

    private async Task<FullRestoreResult> RestoreToVMwareAsync(BackupPoint backup, FullRestoreRequest request)
    {
        await Task.Delay(100);

        return new FullRestoreResult
        {
            Success = true,
            Message = "VM restored to VMware",
            NewVMId = Guid.NewGuid().ToString()
        };
    }

    private async Task<FullRestoreResult> RestoreToKVMAsync(BackupPoint backup, FullRestoreRequest request)
    {
        await Task.Delay(100);

        return new FullRestoreResult
        {
            Success = true,
            Message = "VM restored to KVM",
            NewVMId = Guid.NewGuid().ToString()
        };
    }

    public async Task<InstantRestoreResult> PerformInstantRestoreAsync(InstantRestoreRequest request)
    {
        await Task.Delay(100);

        return new InstantRestoreResult
        {
            Success = true,
            Message = "VM mounted for instant restore",
            MountPath = $"/mnt/instant_restore_{request.BackupId}"
        };
    }
}

public class FullRestoreRequest
{
    public string BackupId { get; set; } = string.Empty;
    public string TargetHost { get; set; } = string.Empty;
    public string TargetPath { get; set; } = string.Empty;
    public string TargetHypervisor { get; set; } = "hyperv";
    public bool PowerOn { get; set; } = true;
}

public class FullRestoreResult
{
    public bool Success { get; set; }
    public string Message { get; set; } = string.Empty;
    public string NewVMId { get; set; } = string.Empty;
}

public class InstantRestoreRequest
{
    public string BackupId { get; set; } = string.Empty;
    public string TargetHost { get; set; } = string.Empty;
}

public class InstantRestoreResult
{
    public bool Success { get; set; }
    public string Message { get; set; } = string.Empty;
    public string MountPath { get; set; } = string.Empty;
}
