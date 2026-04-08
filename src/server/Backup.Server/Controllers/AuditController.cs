using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Backup.Server.Database;
using Backup.Server.Database.Entities;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize(Roles = "admin")]
public class AuditController : ControllerBase
{
    private readonly BackupDbContext _db;

    public AuditController(BackupDbContext db)
    {
        _db = db;
    }

    [HttpGet]
    public async Task<ActionResult> GetLogs([FromQuery] int page = 1, [FromQuery] int pageSize = 50)
    {
        var total = await _db.AuditLogs.CountAsync();
        var logs = await _db.AuditLogs
            .OrderByDescending(l => l.CreatedAt)
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync();

        return Ok(new
        {
            total,
            page,
            pageSize,
            logs
        });
    }

    [HttpGet("user/{userId}")]
    public async Task<ActionResult> GetUserLogs(string userId)
    {
        var logs = await _db.AuditLogs
            .Where(l => l.UserId == userId)
            .OrderByDescending(l => l.CreatedAt)
            .Take(100)
            .ToListAsync();

        return Ok(logs);
    }
}
