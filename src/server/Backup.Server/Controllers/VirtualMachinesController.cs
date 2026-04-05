using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Backup.Server.Database;
using Backup.Server.Database.Entities;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class VirtualMachinesController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<VirtualMachinesController> _logger;

    public VirtualMachinesController(BackupDbContext db, ILogger<VirtualMachinesController> logger)
    {
        _db = db;
        _logger = logger;
    }

    [HttpGet]
    [Authorize(Policy = "Viewer")]
    public async Task<ActionResult> GetVMs(
        [FromQuery] string? hypervisorType = null,
        [FromQuery] string? hypervisorHost = null,
        [FromQuery] string? status = null)
    {
        var query = _db.VirtualMachines.AsQueryable();

        if (!string.IsNullOrEmpty(hypervisorType))
            query = query.Where(v => v.HypervisorType == hypervisorType);
        if (!string.IsNullOrEmpty(hypervisorHost))
            query = query.Where(v => v.HypervisorHost == hypervisorHost);
        if (!string.IsNullOrEmpty(status))
            query = query.Where(v => v.Status == status);

        var vms = await query.Select(v => new
        {
            v.Id,
            v.VmId,
            v.Name,
            v.HypervisorType,
            v.HypervisorHost,
            v.IpAddress,
            v.OsType,
            v.MemoryMb,
            v.CpuCores,
            v.Disks,
            v.Tags,
            v.Status,
            v.LastBackupAt,
            v.LastBackupId,
            v.CreatedAt
        }).ToListAsync();

        return Ok(vms);
    }

    [HttpGet("{vmId}")]
    public async Task<ActionResult> GetVM(string vmId)
    {
        var vm = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == vmId);
        if (vm == null) return NotFound();
        return Ok(vm);
    }

    [HttpPost]
    public async Task<ActionResult> CreateVM([FromBody] VirtualMachine vm)
    {
        if (string.IsNullOrEmpty(vm.VmId))
            vm.VmId = Guid.NewGuid().ToString();

        vm.CreatedAt = DateTime.UtcNow;
        vm.UpdatedAt = DateTime.UtcNow;

        _db.VirtualMachines.Add(vm);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Registered VM {VmId}: {Name} on {HypervisorHost}", vm.VmId, vm.Name, vm.HypervisorHost);
        return CreatedAtAction(nameof(GetVM), new { vmId = vm.VmId }, vm);
    }

    [HttpPut("{vmId}")]
    public async Task<ActionResult> UpdateVM(string vmId, [FromBody] VirtualMachine vm)
    {
        var existing = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == vmId);
        if (existing == null) return NotFound();

        existing.Name = vm.Name;
        existing.HypervisorType = vm.HypervisorType;
        existing.HypervisorHost = vm.HypervisorHost;
        existing.IpAddress = vm.IpAddress;
        existing.OsType = vm.OsType;
        existing.MemoryMb = vm.MemoryMb;
        existing.CpuCores = vm.CpuCores;
        existing.Disks = vm.Disks;
        existing.Tags = vm.Tags;
        existing.Status = vm.Status;
        existing.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{vmId}")]
    public async Task<ActionResult> DeleteVM(string vmId)
    {
        var vm = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == vmId);
        if (vm == null) return NotFound();

        _db.VirtualMachines.Remove(vm);
        await _db.SaveChangesAsync();

        return NoContent();
    }

    [HttpPost("discover")]
    public async Task<ActionResult> DiscoverVMs([FromBody] DiscoverRequest request)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == request.HypervisorId);
        if (hypervisor == null)
            return NotFound(new { error = "Hypervisor not found" });

        var discovered = new List<object>();

        if (hypervisor.Type == "hyperv")
        {
            discovered = await DiscoverHyperVVMs(hypervisor);
        }
        else if (hypervisor.Type == "vmware")
        {
            discovered = new List<object> { new { name = "VMware discovery not yet implemented", vmId = "" } };
        }
        else if (hypervisor.Type == "kvm")
        {
            discovered = new List<object> { new { name = "KVM discovery not yet implemented", vmId = "" } };
        }

        return Ok(discovered);
    }

    private async Task<List<object>> DiscoverHyperVVMs(Hypervisor hypervisor)
    {
        var vms = new List<object>();
        try
        {
            var process = new System.Diagnostics.Process
            {
                StartInfo = new System.Diagnostics.ProcessStartInfo
                {
                    FileName = "powershell.exe",
                    Arguments = $"-Command \"Get-VM | Select-Object Name, State, CPUCount, MemoryStartup, Uuid | ConvertTo-Json\"",
                    RedirectStandardOutput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true
                }
            };
            process.Start();
            var output = await process.StandardOutput.ReadToEndAsync();
            await process.WaitForExitAsync();

            if (!string.IsNullOrEmpty(output))
            {
                using var doc = System.Text.Json.JsonDocument.Parse(output);
                if (doc.RootElement.ValueKind == System.Text.Json.JsonValueKind.Array)
                {
                    foreach (var item in doc.RootElement.EnumerateArray())
                    {
                        var name = item.GetProperty("Name").GetString() ?? "";
                        var vmId = item.GetProperty("Uuid").GetString() ?? "";
                        var state = item.GetProperty("State").GetString() ?? "";
                        var cpu = item.TryGetProperty("CPUCount", out var cpuProp) ? cpuProp.GetInt32() : 0;
                        var memory = item.TryGetProperty("MemoryStartup", out var memProp) ? memProp.GetInt64() / (1024 * 1024) : 0;

                        vms.Add(new
                        {
                            name,
                            vmId,
                            hypervisorType = "hyperv",
                            hypervisorHost = hypervisor.Host,
                            status = state.ToLower() == "running" ? "running" : "stopped",
                            cpuCores = cpu,
                            memoryMb = memory
                        });
                    }
                }
            }
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Hyper-V discovery failed on {Host}", hypervisor.Host);
        }

        return vms;
    }

    [HttpGet("{vmId}/backups")]
    public async Task<ActionResult> GetVMBackups(string vmId)
    {
        var backups = await _db.BackupPoints
            .Where(b => b.VmId == vmId)
            .OrderByDescending(b => b.CreatedAt)
            .Select(b => new
            {
                b.BackupId,
                b.BackupType,
                b.RepositoryId,
                b.SizeBytes,
                b.Status,
                b.CreatedAt,
                b.CompletedAt
            })
            .ToListAsync();

        return Ok(backups);
    }
}

public class DiscoverRequest
{
    public string HypervisorId { get; set; } = string.Empty;
}
