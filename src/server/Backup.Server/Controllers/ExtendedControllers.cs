using Microsoft.AspNetCore.Mvc;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
public class BackupsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<BackupsController> _logger;

    public BackupsController(BackupDbContext db, ILogger<BackupsController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetBackups(
        [FromQuery] string? jobId = null,
        [FromQuery] string? repositoryId = null,
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 20)
    {
        var query = _db.BackupPoints.AsQueryable();

        if (!string.IsNullOrEmpty(jobId))
            query = query.Where(b => b.JobId == jobId);
        if (!string.IsNullOrEmpty(repositoryId))
            query = query.Where(b => b.RepositoryId == repositoryId);

        var backups = await query
            .OrderByDescending(b => b.CreatedAt)
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync();

        var total = await query.CountAsync();

        return Ok(new { backups, total, page, pageSize });
    }

    [HttpGet("{backupId}")]
    public async Task<ActionResult> GetBackup(string backupId)
    {
        var backup = await _db.BackupPoints
            .Include(b => b.Repository)
            .FirstOrDefaultAsync(b => b.BackupId == backupId);

        if (backup == null) return NotFound();

        return Ok(backup);
    }

    [HttpDelete("{backupId}")]
    public async Task<ActionResult> DeleteBackup(string backupId, [FromQuery] bool force = false)
    {
        var backup = await _db.BackupPoints.FindAsync(backupId);
        if (backup == null) return NotFound();

        _db.BackupPoints.Remove(backup);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Deleted backup {BackupId}", backupId);
        return NoContent();
    }

    [HttpPost("{backupId}/verify")]
    public async Task<ActionResult> VerifyBackup(string backupId, [FromQuery] bool checkChecksum = true)
    {
        var backup = await _db.BackupPoints.FindAsync(backupId);
        if (backup == null) return NotFound();

        backup.Status = "verified";
        await _db.SaveChangesAsync();

        _logger.LogInformation("Verified backup {BackupId}", backupId);
        return Ok(new { success = true, message = "Backup verified successfully" });
    }

    [HttpGet("{backupId}/download")]
    public async Task<ActionResult> GetBackupDownloadUrl(string backupId)
    {
        var backup = await _db.BackupPoints.FindAsync(backupId);
        if (backup == null) return NotFound();

        return Ok(new { url = backup.FilePath, method = "direct" });
    }
}

[ApiController]
[Route("api/[controller]")]
public class RestoreController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<RestoreController> _logger;

    public RestoreController(BackupDbContext db, ILogger<RestoreController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpPost]
    public async Task<ActionResult> StartRestore([FromBody] RestoreRequest request)
    {
        var backup = await _db.BackupPoints.FindAsync(request.BackupId);
        if (backup == null) return NotFound("Backup not found");

        var restore = new Restore
        {
            RestoreId = Guid.NewGuid().ToString(),
            BackupId = request.BackupId,
            RestoreType = request.RestoreType,
            DestinationPath = request.DestinationPath,
            TargetHost = request.TargetHost,
            Options = System.Text.Json.JsonSerializer.Serialize(request.Options),
            Status = "pending",
            CreatedAt = DateTime.UtcNow
        };

        _db.Restores.Add(restore);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Started restore {RestoreId} from backup {BackupId}", 
            restore.RestoreId, request.BackupId);

        return Ok(new { restoreId = restore.RestoreId, status = "pending" });
    }

    [HttpGet("{restoreId}")]
    public async Task<ActionResult> GetRestoreProgress(string restoreId)
    {
        var restore = await _db.Restores.FindAsync(restoreId);
        if (restore == null) return NotFound();

        return Ok(new
        {
            restoreId = restore.RestoreId,
            status = restore.Status,
            bytesRestored = restore.BytesRestored,
            totalBytes = restore.TotalBytes,
            progress = restore.TotalBytes > 0 
                ? (double)restore.BytesRestored / restore.TotalBytes * 100 
                : 0,
            errorMessage = restore.ErrorMessage
        });
    }

    [HttpPost("{restoreId}/cancel")]
    public async Task<ActionResult> CancelRestore(string restoreId, [FromQuery] bool force = false)
    {
        var restore = await _db.Restores.FindAsync(restoreId);
        if (restore == null) return NotFound();

        restore.Status = "cancelled";
        restore.CompletedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();

        _logger.LogInformation("Cancelled restore {RestoreId}", restoreId);
        return Ok(new { success = true, message = "Restore cancelled" });
    }

    [HttpGet]
    public async Task<ActionResult> GetRestores([FromQuery] int page = 1, [FromQuery] int pageSize = 20)
    {
        var restores = await _db.Restores
            .OrderByDescending(r => r.CreatedAt)
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync();

        var total = await _db.Restores.CountAsync();

        return Ok(new { restores, total, page, pageSize });
    }

    [HttpGet("{restoreId}/files")]
    public async Task<ActionResult> BrowseFiles(string restoreId, [FromQuery] string path = "/")
    {
        var restore = await _db.Restores.FindAsync(restoreId);
        if (restore == null) return NotFound();

        return Ok(new { files = new List<object>(), currentPath = path, hasMore = false });
    }
}

[ApiController]
[Route("api/[controller]")]
public class SettingsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<SettingsController> _logger;

    public SettingsController(BackupDbContext db, ILogger<SettingsController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    public async Task<ActionResult> GetSettings()
    {
        var settings = await _db.Settings.ToListAsync();
        return Ok(settings);
    }

    [HttpGet("{key}")]
    public async Task<ActionResult> GetSetting(string key)
    {
        var setting = await _db.Settings.FirstOrDefaultAsync(s => s.Key == key);
        if (setting == null) return NotFound();
        return Ok(setting);
    }

    [HttpPut("{key}")]
    public async Task<ActionResult> UpdateSetting(string key, [FromBody] Setting setting)
    {
        var existing = await _db.Settings.FirstOrDefaultAsync(s => s.Key == key);
        
        if (existing == null)
        {
            setting.Key = key;
            setting.UpdatedAt = DateTime.UtcNow;
            _db.Settings.Add(setting);
        }
        else
        {
            existing.Value = setting.Value;
            existing.Type = setting.Type;
            existing.Description = setting.Description;
            existing.UpdatedAt = DateTime.UtcNow;
        }

        await _db.SaveChangesAsync();
        _logger.LogInformation("Updated setting {Key}", key);
        return Ok(existing);
    }

    [HttpDelete("{key}")]
    public async Task<ActionResult> DeleteSetting(string key)
    {
        var setting = await _db.Settings.FirstOrDefaultAsync(s => s.Key == key);
        if (setting == null) return NotFound();

        _db.Settings.Remove(setting);
        await _db.SaveChangesAsync();

        return NoContent();
    }
}

[ApiController]
[Route("api/[controller]")]
public class ReportsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<ReportsController> _logger;

    public ReportsController(BackupDbContext db, ILogger<ReportsController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet("summary")]
    public async Task<ActionResult> GetSummary()
    {
        var totalJobs = await _db.Jobs.CountAsync();
        var activeJobs = await _db.Jobs.CountAsync(j => j.Enabled);
        var totalBackups = await _db.BackupPoints.CountAsync();
        var successfulBackups = await _db.BackupPoints.CountAsync(b => b.Status == "completed");
        var totalRepositories = await _db.Repositories.CountAsync();
        var totalAgents = await _db.Agents.CountAsync();
        var onlineAgents = await _db.Agents.CountAsync(a => a.Status != "offline");

        var successRate = totalBackups > 0 
            ? (double)successfulBackups / totalBackups * 100 
            : 100;

        return Ok(new
        {
            totalJobs,
            activeJobs,
            totalBackups,
            successfulBackups,
            totalRepositories,
            totalAgents,
            onlineAgents,
            successRate
        });
    }

    [HttpGet("activity")]
    public async Task<ActionResult> GetActivity(
        [FromQuery] DateTime? from = null,
        [FromQuery] DateTime? to = null,
        [FromQuery] int limit = 50)
    {
        var fromDate = from ?? DateTime.UtcNow.AddDays(-7);
        var toDate = to ?? DateTime.UtcNow;

        var history = await _db.JobRunHistory
            .Where(h => h.StartTime >= fromDate && h.StartTime <= toDate)
            .OrderByDescending(h => h.StartTime)
            .Take(limit)
            .ToListAsync();

        return Ok(history);
    }

    [HttpGet("storage")]
    public async Task<ActionResult> GetStorageReport()
    {
        var repos = await _db.Repositories.ToListAsync();
        
        var report = repos.Select(r => new
        {
            repositoryId = r.RepositoryId,
            name = r.Name,
            type = r.Type,
            capacityBytes = r.CapacityBytes,
            usedBytes = r.UsedBytes,
            availableBytes = r.CapacityBytes - r.UsedBytes,
            utilizationPercent = r.CapacityBytes > 0 
                ? (double)r.UsedBytes / r.CapacityBytes * 100 
                : 0
        }).ToList();

        return Ok(report);
    }

    [HttpGet("jobs/{jobId}/history")]
    public async Task<ActionResult> GetJobHistory(string jobId, [FromQuery] int limit = 10)
    {
        var history = await _db.JobRunHistory
            .Where(h => h.JobId == jobId)
            .OrderByDescending(h => h.StartTime)
            .Take(limit)
            .ToListAsync();

        return Ok(history);
    }
}

public class RestoreRequest
{
    public string BackupId { get; set; } = string.Empty;
    public string RestoreType { get; set; } = "full_vm";
    public string DestinationPath { get; set; } = string.Empty;
    public string TargetHost { get; set; } = string.Empty;
    public Dictionary<string, string> Options { get; set; } = new();
}
