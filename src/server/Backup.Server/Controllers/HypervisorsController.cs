using System.Diagnostics;
using System.Text.Json;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Backup.Server.Services;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class HypervisorsController : ControllerBase
{
    private readonly BackupDbContext _db;
    private readonly ILogger<HypervisorsController> _logger;
    private readonly IEncryptionService _encryption;

    public HypervisorsController(
        BackupDbContext db,
        ILogger<HypervisorsController> logger,
        IEncryptionService encryption)
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
        if (hypervisor == null)
        {
            return NotFound();
        }

        return Ok(hypervisor);
    }

    [HttpPost]
    public async Task<ActionResult> CreateHypervisor([FromBody] HypervisorDto hypervisorDto)
    {
        var hypervisor = new Hypervisor
        {
            HypervisorId = Guid.NewGuid().ToString(),
            Name = hypervisorDto.Name,
            Type = hypervisorDto.Type,
            Host = hypervisorDto.Host,
            Port = hypervisorDto.Port,
            Username = hypervisorDto.Username,
            Status = string.IsNullOrWhiteSpace(hypervisorDto.Status) ? "offline" : hypervisorDto.Status
        };

        if (string.IsNullOrWhiteSpace(hypervisor.HypervisorId))
        {
            hypervisor.HypervisorId = Guid.NewGuid().ToString();
        }

        if (!string.IsNullOrWhiteSpace(hypervisor.Password))
        {
            hypervisor.Password = _encryption.Encrypt(hypervisor.Password);
        }

        hypervisor.CreatedAt = DateTime.UtcNow;
        hypervisor.UpdatedAt = DateTime.UtcNow;

        _db.Hypervisors.Add(hypervisor);
        await _db.SaveChangesAsync();

        _logger.LogInformation(
            "Added hypervisor {Id}: {Name} ({Type}) at {Host}",
            hypervisor.HypervisorId,
            hypervisor.Name,
            hypervisor.Type,
            hypervisor.Host);

        return CreatedAtAction(nameof(GetHypervisor), new { hypervisorId = hypervisor.HypervisorId }, hypervisor);
    }

    [HttpPut("{hypervisorId}")]
    public async Task<ActionResult> UpdateHypervisor(string hypervisorId, [FromBody] HypervisorDto hypervisorDto)
    {
        var existing = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (existing == null)
        {
            return NotFound();
        }

        existing.Name = hypervisorDto.Name;
        existing.Type = hypervisorDto.Type;
        existing.Host = hypervisorDto.Host;
        existing.Port = hypervisorDto.Port;
        existing.Username = hypervisorDto.Username;
        if (!string.IsNullOrWhiteSpace(hypervisorDto.Password))
        {
            existing.Password = _encryption.Encrypt(hypervisorDto.Password);
        }

        existing.Status = string.IsNullOrWhiteSpace(hypervisorDto.Status) ? existing.Status : hypervisorDto.Status;
        existing.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();
        return Ok(existing);
    }

    [HttpDelete("{hypervisorId}")]
    public async Task<ActionResult> DeleteHypervisor(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null)
        {
            return NotFound();
        }

        _db.Hypervisors.Remove(hypervisor);
        await _db.SaveChangesAsync();

        return NoContent();
    }

    [HttpPost("{hypervisorId}/test")]
    public async Task<ActionResult> TestConnection(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null)
        {
            return NotFound();
        }

        try
        {
            var connected = hypervisor.Type.ToLowerInvariant() switch
            {
                "hyperv" => await TestHyperVConnectionAsync(hypervisor),
                "vmware" => await TestTcpConnectionAsync(hypervisor.Host, hypervisor.Port <= 0 ? 443 : hypervisor.Port),
                "kvm" => await TestTcpConnectionAsync(hypervisor.Host, hypervisor.Port <= 0 ? 22 : hypervisor.Port),
                _ => false
            };

            hypervisor.Status = connected ? "connected" : "error";
            hypervisor.LastConnectedAt = connected ? DateTime.UtcNow : hypervisor.LastConnectedAt;
            hypervisor.UpdatedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync();

            return Ok(new
            {
                success = connected,
                status = hypervisor.Status,
                message = connected ? "Connection successful" : "Connection failed"
            });
        }
        catch (Exception ex)
        {
            hypervisor.Status = "error";
            hypervisor.UpdatedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync();

            _logger.LogWarning(ex, "Hypervisor connection test failed for {HypervisorId}", hypervisorId);
            return Ok(new { success = false, status = "error", message = ex.Message });
        }
    }

    [HttpPost("{hypervisorId}/refresh")]
    public async Task<ActionResult> RefreshVMs(string hypervisorId)
    {
        var hypervisor = await _db.Hypervisors.FirstOrDefaultAsync(h => h.HypervisorId == hypervisorId);
        if (hypervisor == null)
        {
            return NotFound();
        }

        if (!string.Equals(hypervisor.Type, "hyperv", StringComparison.OrdinalIgnoreCase))
        {
            return Ok(new
            {
                message = "Automatic VM refresh is currently implemented only for Hyper-V."
            });
        }

        var vmInfos = await DiscoverHyperVVMsAsync(hypervisor);
        var existingVMs = await _db.VirtualMachines
            .Where(v => v.HypervisorHost == hypervisor.Host)
            .ToListAsync();

        foreach (var vmInfo in vmInfos)
        {
            var existing = existingVMs.FirstOrDefault(v => v.VmId == vmInfo.VmId);
            if (existing != null)
            {
                existing.Name = vmInfo.Name;
                existing.Status = vmInfo.Status;
                existing.CpuCores = vmInfo.CpuCores;
                existing.MemoryMb = vmInfo.MemoryMb;
                existing.UpdatedAt = DateTime.UtcNow;
            }
            else
            {
                _db.VirtualMachines.Add(new VirtualMachine
                {
                    VmId = vmInfo.VmId,
                    Name = vmInfo.Name,
                    HypervisorType = hypervisor.Type,
                    HypervisorHost = hypervisor.Host,
                    Status = vmInfo.Status,
                    CpuCores = vmInfo.CpuCores,
                    MemoryMb = vmInfo.MemoryMb,
                    CreatedAt = DateTime.UtcNow,
                    UpdatedAt = DateTime.UtcNow
                });
            }
        }

        hypervisor.VmCount = vmInfos.Count;
        hypervisor.LastConnectedAt = DateTime.UtcNow;
        hypervisor.UpdatedAt = DateTime.UtcNow;

        await _db.SaveChangesAsync();

        return Ok(new { message = "VM list refreshed", count = vmInfos.Count });
    }

    private async Task<bool> TestHyperVConnectionAsync(Hypervisor hypervisor)
    {
        var result = await RunPowerShellAsync(
            "Get-VMHost | Select-Object Name, LogicalProcessorCount, MemoryCapacity | ConvertTo-Json",
            hypervisor);

        return result.ExitCode == 0 && !string.IsNullOrWhiteSpace(result.StandardOutput);
    }

    private static async Task<bool> TestTcpConnectionAsync(string host, int port)
    {
        try
        {
            using var client = new System.Net.Sockets.TcpClient();
            using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5));
            await client.ConnectAsync(host, port, cts.Token);
            return client.Connected;
        }
        catch
        {
            return false;
        }
    }

    private async Task<List<HyperVVmInfo>> DiscoverHyperVVMsAsync(Hypervisor hypervisor)
    {
        var result = await RunPowerShellAsync(
            "Get-VM | Select-Object Name, State, CPUCount, MemoryStartup, Id | ConvertTo-Json",
            hypervisor);

        if (result.ExitCode != 0 || string.IsNullOrWhiteSpace(result.StandardOutput))
        {
            _logger.LogWarning(
                "Hyper-V VM refresh failed for {HypervisorId}. Error: {Error}",
                hypervisor.HypervisorId,
                result.StandardError);
            return new List<HyperVVmInfo>();
        }

        try
        {
            using var doc = JsonDocument.Parse(result.StandardOutput);
            var elements = doc.RootElement.ValueKind == JsonValueKind.Array
                ? doc.RootElement.EnumerateArray().ToList()
                : new List<JsonElement> { doc.RootElement };

            return elements.Select(item => new HyperVVmInfo
            {
                Name = item.TryGetProperty("Name", out var name) ? name.GetString() ?? string.Empty : string.Empty,
                VmId = item.TryGetProperty("Id", out var vmId) ? vmId.GetString() ?? string.Empty : string.Empty,
                Status = item.TryGetProperty("State", out var state) && string.Equals(state.GetString(), "Running", StringComparison.OrdinalIgnoreCase)
                    ? "running"
                    : "stopped",
                CpuCores = item.TryGetProperty("CPUCount", out var cpu) ? cpu.GetInt32() : 0,
                MemoryMb = item.TryGetProperty("MemoryStartup", out var memory) ? memory.GetInt64() / (1024 * 1024) : 0
            })
            .Where(vm => !string.IsNullOrWhiteSpace(vm.VmId))
            .ToList();
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Failed to parse Hyper-V VM list for {HypervisorId}", hypervisor.HypervisorId);
            return new List<HyperVVmInfo>();
        }
    }

    private async Task<CommandResult> RunPowerShellAsync(string command, Hypervisor hypervisor)
    {
        var username = hypervisor.Username ?? string.Empty;
        var password = string.IsNullOrWhiteSpace(hypervisor.Password) ? string.Empty : _encryption.Decrypt(hypervisor.Password);

        var script = string.IsNullOrWhiteSpace(username) || string.IsNullOrWhiteSpace(password)
            ? command
            : "$securePwd = ConvertTo-SecureString -String $env:BACKUP_HYPERV_PASSWORD -AsPlainText -Force; " +
              "$cred = New-Object System.Management.Automation.PSCredential($env:BACKUP_HYPERV_USERNAME, $securePwd); " +
              $"Invoke-Command -ComputerName '{EscapePowerShellSingleQuoted(hypervisor.Host)}' -Credential $cred -ScriptBlock {{ {command} }}";

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
        psi.ArgumentList.Add(script);

        if (!string.IsNullOrWhiteSpace(username) && !string.IsNullOrWhiteSpace(password))
        {
            psi.Environment["BACKUP_HYPERV_USERNAME"] = username;
            psi.Environment["BACKUP_HYPERV_PASSWORD"] = password;
        }

        using var process = Process.Start(psi);
        if (process == null)
        {
            throw new InvalidOperationException("Failed to start PowerShell process");
        }

        var standardOutput = await process.StandardOutput.ReadToEndAsync();
        var standardError = await process.StandardError.ReadToEndAsync();
        await process.WaitForExitAsync();

        return new CommandResult(process.ExitCode, standardOutput, standardError);
    }

    private static string EscapePowerShellSingleQuoted(string value)
    {
        return value.Replace("'", "''");
    }

    private sealed record CommandResult(int ExitCode, string StandardOutput, string StandardError);

    private sealed class HyperVVmInfo
    {
        public string VmId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Status { get; set; } = "stopped";
        public int CpuCores { get; set; }
        public long MemoryMb { get; set; }
    }
}

public class HypervisorDto
{
    public string Name { get; set; } = string.Empty;
    public string Type { get; set; } = "hyperv";
    public string Host { get; set; } = string.Empty;
    public int Port { get; set; }
    public string? Username { get; set; }
    public string? Password { get; set; }
    public string? Status { get; set; }
}
