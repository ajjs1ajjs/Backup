using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Backup.Server.Database;
using Backup.Server.Database.Entities;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class DashboardController : ControllerBase
{
    private readonly BackupDbContext _db;

    public DashboardController(BackupDbContext db)
    {
        _db = db;
    }

    [HttpGet("stats")]
    public async Task<ActionResult> GetStats()
    {
        var totalJobs = await _db.Jobs.CountAsync();
        var totalVMs = await _db.VirtualMachines.CountAsync();
        var totalBackups = await _db.BackupPoints.CountAsync();
        var totalRepositories = await _db.Repositories.CountAsync();

        var recentBackups = await _db.BackupPoints
            .OrderByDescending(b => b.CreatedAt)
            .Take(7)
            .ToListAsync();

        var backupHistory = await _db.JobRunHistory
            .OrderByDescending(h => h.StartTime)
            .Take(30)
            .Select(h => new 
            {
                h.StartTime,
                h.Status,
                h.BytesProcessed
            })
            .ToListAsync();

        var storageStats = await _db.Repositories
            .Select(r => new 
            {
                r.Name,
                r.UsedBytes,
                r.CapacityBytes
            })
            .ToListAsync();

        return Ok(new
        {
            summary = new { totalJobs, totalVMs, totalBackups, totalRepositories },
            recentBackups,
            backupHistory = backupHistory.OrderBy(h => h.StartTime),
            storageStats
        });
    }
}
