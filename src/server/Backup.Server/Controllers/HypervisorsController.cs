using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Backup.Server.Services;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class HypervisorsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<HypervisorsController> _logger;
    private readonly IEncryptionService _encryption;

    public HypervisorsController(BackupDbContext db, ILogger<HypervisorsController> logger, IEncryptionService encryption)
    {
        _db = db;
        _logger = logger;
        _encryption = encryption;
    }

    [HttpGet]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetHypervisors()
    {
        var hypervisors = await _db.Hypervisors
            .Select(h => new
            {
                h.Id,
                h.HypervisorId,
                h.Name,
                h.Type,
                h.Host,
                h.Port,
                h.Username,
                h.Status,
                h.VmCount,
                h.LastConnectedAt,
                h.CreatedAt
            })
            .ToListAsync();

        return Ok(hypervisors);
    }

    [HttpGet("{hypervisorId}")]
    public async Task<ActionResult> GetHypervisor(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null) return NotFound();
        return Ok(hypervisor);
    }

    [HttpPost]
    public async Task<ActionResult> CreateHypervisor([FromBody] Hypervisor hypervisor)
    {
        if (string.IsNullOrEmpty(hypervisor.HypervisorId))
            hypervisor.HypervisorId = Guid.NewGuid().ToString();

        if (!string.IsNullOrEmpty(hypervisor.Password))
            hypervisor.Password = _encryption.Encrypt(hypervisor.Password);

        hypervisor.CreatedAt = DateTime.UtcNow;
        hypervisor.UpdatedAt = DateTime.UtcNow;

        _db.Hypervisors.Add(hypervisor);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Added hypervisor {Id}: {Name} ({Type}) at {Host}", hypervisor.HypervisorId, hypervisor.Name, hypervisor.Type, hypervisor.Host);
        return CreatedAtAction(nameof(GetHypervisor), new { hypervisorId = hypervisor.HypervisorId }, hypervisor);
    }

    [HttpPut("{hypervisorId}")]
    public async Task<ActionResult> UpdateHypervisor(string hypervisorId, [FromBody] Hypervisor hypervisor)
    {
        var existing = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (existing == null) return NotFound();

        existing.Name = hypervisor.Name;
        existing.Type = hypervisor.Type;
        existing.Host = hypervisor.Host;
        existing.Port = hypervisor.Port;
        existing.Username = hypervisor.Username;
        if (!string.IsNullOrEmpty(hypervisor.Password))
            existing.Password = _encryption.Encrypt(hypervisor.Password);
        existing.Status = hypervisor.Status;
        existing.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{hypervisorId}")]
    public async Task<ActionResult> DeleteHypervisor(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null) return NotFound();

        _db.Hypervisors.Remove(hypervisor);
        await _db.SaveChangesAsync();

        return NoContent();
    }

    [HttpPost("{hypervisorId}/test")]
    public async Task<ActionResult> TestConnection(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null) return NotFound();

        try
        {
            if (hypervisor.Type == "hyperv")
            {
                var decryptedPassword = _encryption.Decrypt(hypervisor.Password);
                var process = new System.Diagnostics.Process
                {
                    StartInfo = new System.Diagnostics.ProcessStartInfo
                    {
                        FileName = "powershell.exe",
                        Arguments = $"-Command \"$securePwd = ConvertTo-SecureString -String '{decryptedPassword}' -AsPlainText -Force; $cred = New-Object System.Management.Automation.PSCredential('{hypervisor.Username}', $securePwd); Enter-PSSession -ComputerName '{hypervisor.Host}' -Credential $cred; Get-VMHost | Select-Object Name, LogicalProcessorCount, MemoryCapacity | ConvertTo-Json\"",
                        RedirectStandardOutput = true,
                        UseShellExecute = false,
                        CreateNoWindow = true
                    }
                };
                process.Start();
                var output = await process.StandardOutput.ReadToEndAsync();
                await process.WaitForExitAsync();

                hypervisor.Status = process.ExitCode == 0 ? "connected" : "error";
                hypervisor.LastConnectedAt = DateTime.UtcNow;
            }
            else
            {
                hypervisor.Status = "connected";
                hypervisor.LastConnectedAt = DateTime.UtcNow;
            }

            hypervisor.UpdatedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync();

            return Ok(new { success = true, status = hypervisor.Status, message = "Connection successful" });
        }
        catch (Exception ex)
        {
            hypervisor.Status = "error";
            hypervisor.UpdatedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync();

            return Ok(new { success = false, status = "error", message = ex.Message });
        }
    }

    [HttpPost("{hypervisorId}/refresh")]
    public async Task<ActionResult> RefreshVMs(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null) return NotFound();

        var existingVMs = await _db.VirtualMachines.Where(v => v.HypervisorHost == hypervisor.Host).ToListAsync();

        if (hypervisor.Type == "hyperv")
        {
            var decryptedPassword = _encryption.Decrypt(hypervisor.Password);
            var process = new System.Diagnostics.Process
            {
                StartInfo = new System.Diagnostics.ProcessStartInfo
                {
                    FileName = "powershell.exe",
                    Arguments = $"-Command \"$securePwd = ConvertTo-SecureString -String '{decryptedPassword}' -AsPlainText -Force; $cred = New-Object System.Management.Automation.PSCredential('{hypervisor.Username}', $securePwd); Enter-PSSession -ComputerName '{hypervisor.Host}' -Credential $cred; Get-VM | Select-Object Name, State, CPUCount, MemoryStartup, Uuid, Path | ConvertTo-Json\"",
                    RedirectStandardOutput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true
                }
            };
            process.Start();
            var output = await process.StandardOutput.ReadToEndAsync();
            await process.WaitForExitAsync();

            if (!string.IsNullOrEmpty(output) && process.ExitCode == 0)
            {
                using var doc = System.Text.Json.JsonDocument.Parse(output);
                var root = doc.RootElement;
                var vmArray = root.ValueKind == System.Text.Json.JsonValueKind.Array
                    ? root.EnumerateArray()
                    : new[] { root }.AsEnumerable();

                foreach (var item in vmArray)
                {
                    var name = item.GetProperty("Name").GetString() ?? "";
                    var vmId = item.GetProperty("Uuid").GetString() ?? "";
                    var state = item.GetProperty("State").GetString() ?? "";
                    var cpu = item.TryGetProperty("CPUCount", out var cpuProp) ? cpuProp.GetInt32() : 0;
                    var memory = item.TryGetProperty("MemoryStartup", out var memProp) ? memProp.GetInt64() / (1024 * 1024) : 0;
                    var path = item.TryGetProperty("Path", out var pathProp) ? pathProp.GetString() ?? "" : "";

                    var existing = existingVMs.FirstOrDefault(v => v.VmId == vmId);
                    if (existing != null)
                    {
                        existing.Name = name;
                        existing.Status = state.ToLower() == "running" ? "running" : "stopped";
                        existing.CpuCores = cpu;
                        existing.MemoryMb = memory;
                        existing.UpdatedAt = DateTime.UtcNow;
                    }
                    else
                    {
                        _db.VirtualMachines.Add(new VirtualMachine
                        {
                            VmId = vmId,
                            Name = name,
                            HypervisorType = hypervisor.Type,
                            HypervisorHost = hypervisor.Host,
                            Status = state.ToLower() == "running" ? "running" : "stopped",
                            CpuCores = cpu,
                            MemoryMb = memory,
                            CreatedAt = DateTime.UtcNow,
                            UpdatedAt = DateTime.UtcNow
                        });
                    }
                }

                await _db.SaveChangesAsync();
                hypervisor.VmCount = await _db.VirtualMachines.CountAsync(v => v.HypervisorHost == hypervisor.Host);
                hypervisor.LastConnectedAt = DateTime.UtcNow;
                hypervisor.UpdatedAt = DateTime.UtcNow;
                await _db.SaveChangesAsync();
            }
        }

        return Ok(new { message = "VM list refreshed" });
    }
}
