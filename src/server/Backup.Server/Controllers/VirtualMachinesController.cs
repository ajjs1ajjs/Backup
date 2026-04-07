using System.Diagnostics;
using System.Text.Json;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

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

        if (!string.IsNullOrWhiteSpace(hypervisorType))
        {
            query = query.Where(v => v.HypervisorType == hypervisorType);
        }

        if (!string.IsNullOrWhiteSpace(hypervisorHost))
        {
            query = query.Where(v => v.HypervisorHost == hypervisorHost);
        }

        if (!string.IsNullOrWhiteSpace(status))
        {
            query = query.Where(v => v.Status == status);
        }

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
        if (vm == null)
        {
            return NotFound();
        }

        return Ok(vm);
    }

    [HttpPost]
    public async Task<ActionResult> CreateVM([FromBody] VirtualMachineDto vmDto)
    {
        var vm = new VirtualMachine
        {
            VmId = string.IsNullOrWhiteSpace(vmDto.VmId) ? Guid.NewGuid().ToString() : vmDto.VmId,
            Name = vmDto.Name,
            HypervisorType = vmDto.HypervisorType,
            HypervisorHost = vmDto.HypervisorHost,
            IpAddress = vmDto.IpAddress,
            OsType = vmDto.OsType,
            MemoryMb = vmDto.MemoryMb,
            CpuCores = vmDto.CpuCores,
            Disks = string.IsNullOrWhiteSpace(vmDto.Disks) ? "[]" : vmDto.Disks,
            Tags = string.IsNullOrWhiteSpace(vmDto.Tags) ? "{}" : vmDto.Tags,
            Status = string.IsNullOrWhiteSpace(vmDto.Status) ? "running" : vmDto.Status
        };

        if (string.IsNullOrWhiteSpace(vm.VmId))
        {
            vm.VmId = Guid.NewGuid().ToString();
        }

        vm.CreatedAt = DateTime.UtcNow;
        vm.UpdatedAt = DateTime.UtcNow;

        _db.VirtualMachines.Add(vm);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Registered VM {VmId}: {Name} on {HypervisorHost}", vm.VmId, vm.Name, vm.HypervisorHost);
        return CreatedAtAction(nameof(GetVM), new { vmId = vm.VmId }, vm);
    }

    [HttpPut("{vmId}")]
    public async Task<ActionResult> UpdateVM(string vmId, [FromBody] VirtualMachineDto vmDto)
    {
        var existing = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == vmId);
        if (existing == null)
        {
            return NotFound();
        }

        existing.Name = vmDto.Name;
        existing.HypervisorType = vmDto.HypervisorType;
        existing.HypervisorHost = vmDto.HypervisorHost;
        existing.IpAddress = vmDto.IpAddress;
        existing.OsType = vmDto.OsType;
        existing.MemoryMb = vmDto.MemoryMb;
        existing.CpuCores = vmDto.CpuCores;
        existing.Disks = string.IsNullOrWhiteSpace(vmDto.Disks) ? existing.Disks : vmDto.Disks;
        existing.Tags = string.IsNullOrWhiteSpace(vmDto.Tags) ? existing.Tags : vmDto.Tags;
        existing.Status = string.IsNullOrWhiteSpace(vmDto.Status) ? existing.Status : vmDto.Status;
        existing.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{vmId}")]
    public async Task<ActionResult> DeleteVM(string vmId)
    {
        var vm = await _db.VirtualMachines.FirstOrDefaultAsync(v => v.VmId == vmId);
        if (vm == null)
        {
            return NotFound();
        }

        _db.VirtualMachines.Remove(vm);
        await _db.SaveChangesAsync();

        return NoContent();
    }

    [HttpPost("discover")]
    public async Task<ActionResult> DiscoverVMs([FromBody] DiscoverRequest request)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == request.HypervisorId);
        if (hypervisor == null)
        {
            return NotFound(new { error = "Hypervisor not found" });
        }

        List<object> discovered = hypervisor.Type.ToLowerInvariant() switch
        {
            "hyperv" => await DiscoverHyperVVMs(hypervisor),
            "vmware" => DiscoverVmwareVMs(hypervisor),
            "kvm" => await DiscoverKvmVMs(hypervisor),
            _ => new List<object>()
        };

        return Ok(discovered);
    }

    private async Task<List<object>> DiscoverKvmVMs(Hypervisor hypervisor)
    {
        try
        {
            var psi = new ProcessStartInfo
            {
                FileName = "ssh",
                UseShellExecute = false,
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                CreateNoWindow = true
            };

            psi.ArgumentList.Add($"{hypervisor.Username}@{hypervisor.Host}");
            psi.ArgumentList.Add("virsh list --all --uuid --name");

            using var process = Process.Start(psi);
            if (process == null)
            {
                return new List<object>();
            }

            var output = await process.StandardOutput.ReadToEndAsync();
            var error = await process.StandardError.ReadToEndAsync();
            await process.WaitForExitAsync();

            if (process.ExitCode != 0)
            {
                _logger.LogWarning("KVM discovery failed on {Host}: {Error}", hypervisor.Host, error);
                return new List<object>();
            }

            return output
                .Split('\n', StringSplitOptions.RemoveEmptyEntries)
                .Select(line => line.Trim())
                .Where(line => !string.IsNullOrWhiteSpace(line))
                .Select(line => new
                {
                    vmId = line,
                    name = line,
                    hypervisorType = "kvm",
                    hypervisorHost = hypervisor.Host,
                    status = "unknown"
                })
                .Cast<object>()
                .ToList();
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "KVM discovery failed on {Host}", hypervisor.Host);
            return new List<object>();
        }
    }

    private static List<object> DiscoverVmwareVMs(Hypervisor hypervisor)
    {
        return new List<object>
        {
            new
            {
                name = "VMware discovery requires a dedicated integration",
                vmId = "vmware-discovery-not-implemented",
                hypervisorType = "vmware",
                hypervisorHost = hypervisor.Host,
                status = "unknown"
            }
        };
    }

    private async Task<List<object>> DiscoverHyperVVMs(Hypervisor hypervisor)
    {
        try
        {
            var psi = new ProcessStartInfo
            {
                FileName = "powershell.exe",
                UseShellExecute = false,
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                CreateNoWindow = true
            };

            psi.ArgumentList.Add("-NoProfile");
            psi.ArgumentList.Add("-NonInteractive");
            psi.ArgumentList.Add("-Command");
            psi.ArgumentList.Add("Get-VM | Select-Object Name, State, CPUCount, MemoryStartup, Id | ConvertTo-Json");

            using var process = Process.Start(psi);
            if (process == null)
            {
                return new List<object>();
            }

            var output = await process.StandardOutput.ReadToEndAsync();
            var error = await process.StandardError.ReadToEndAsync();
            await process.WaitForExitAsync();

            if (process.ExitCode != 0 || string.IsNullOrWhiteSpace(output))
            {
                _logger.LogWarning("Hyper-V discovery failed on {Host}: {Error}", hypervisor.Host, error);
                return new List<object>();
            }

            using var doc = JsonDocument.Parse(output);
            var elements = doc.RootElement.ValueKind == JsonValueKind.Array
                ? doc.RootElement.EnumerateArray().ToList()
                : new List<JsonElement> { doc.RootElement };

            return elements.Select(item => new
            {
                name = item.TryGetProperty("Name", out var name) ? name.GetString() ?? string.Empty : string.Empty,
                vmId = item.TryGetProperty("Id", out var vmId) ? vmId.GetString() ?? string.Empty : string.Empty,
                hypervisorType = "hyperv",
                hypervisorHost = hypervisor.Host,
                status = item.TryGetProperty("State", out var state) && string.Equals(state.GetString(), "Running", StringComparison.OrdinalIgnoreCase)
                    ? "running"
                    : "stopped",
                cpuCores = item.TryGetProperty("CPUCount", out var cpu) ? cpu.GetInt32() : 0,
                memoryMb = item.TryGetProperty("MemoryStartup", out var memory) ? memory.GetInt64() / (1024 * 1024) : 0
            }).Cast<object>().ToList();
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Hyper-V discovery failed on {Host}", hypervisor.Host);
            return new List<object>();
        }
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

public class VirtualMachineDto
{
    public string? VmId { get; set; }
    public string Name { get; set; } = string.Empty;
    public string HypervisorType { get; set; } = "hyperv";
    public string HypervisorHost { get; set; } = string.Empty;
    public string? IpAddress { get; set; }
    public string? OsType { get; set; }
    public long? MemoryMb { get; set; }
    public int? CpuCores { get; set; }
    public string? Disks { get; set; }
    public string? Tags { get; set; }
    public string? Status { get; set; }
}

public class DiscoverRequest
{
    public string HypervisorId { get; set; } = string.Empty;
}
