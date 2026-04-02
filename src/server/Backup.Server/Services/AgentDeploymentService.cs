using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;
using System.Net;
using System.Net.Sockets;
using System.Diagnostics;
using System.Text.Json;

namespace Backup.Server.Services;

public class AgentDeploymentService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<AgentDeploymentService> _logger;
    private readonly IWebHostEnvironment _env;
    private readonly IConfiguration _configuration;

    public AgentDeploymentService(
        BackupDbContext db, 
        ILogger<AgentDeploymentService> logger,
        IWebHostEnvironment env,
        IConfiguration configuration)
    {
        _db = db;
        _logger = logger;
        _env = env;
        _configuration = configuration;
    }

    public async Task<DeploymentResult> DeployAgentAsync(DeploymentRequest request)
    {
        var result = new DeploymentResult
        {
            Success = false,
            Message = ""
        };

        try
        {
            var agent = await _db.Agents.FirstOrDefaultAsync(a => a.AgentId == request.AgentId);
            if (agent == null)
            {
                result.Message = "Agent not found";
                return result;
            }

            var deployScript = await GenerateDeployScriptAsync(request);
            
            if (request.DeployMethod == "ssh")
            {
                result = await DeployViaSshAsync(request.TargetHost, deployScript, request.Credentials);
            }
            else if (request.DeployMethod == "winrm")
            {
                result = await DeployViaWinrmAsync(request.TargetHost, deployScript, request.Credentials);
            }
            else if (request.DeployMethod == "manual")
            {
                result = await GenerateManualInstructionsAsync(request);
            }

            _logger.LogInformation("Deployed agent {AgentId} to {Host}", request.AgentId, request.TargetHost);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to deploy agent {AgentId}", request.AgentId);
            result.Message = ex.Message;
        }

        return result;
    }

    private async Task<string> GenerateDeployScriptAsync(DeploymentRequest request)
    {
        var serverUrl = await GetConfiguredServerUrlAsync();
        
        if (request.AgentType == "hyperv" || request.AgentType == "mssql")
        {
            return $@"
$server = '{serverUrl}'
$token = '{request.AgentToken}'
$type = '{request.AgentType}'

$script = Invoke-WebRequest -Uri 'https://get.backupsystem.com/agent/install.ps1' -UseBasicParsing
Invoke-Expression $script

.\install.ps1 -Server $server -Token $token -AgentType $type -AutoStart
";
        }
        else
        {
            return $@"
#!/bin/bash
SERVER='{serverUrl}'
TOKEN='{request.AgentToken}'
TYPE='{request.AgentType}'

curl -fsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- --server $SERVER --token $TOKEN --agent-type $TYPE --auto-start
";
        }
    }

    private async Task<DeploymentResult> DeployViaSshAsync(string host, string script, Dictionary<string, string>? credentials)
    {
        var result = new DeploymentResult { Success = true };
        
        try
        {
            using var client = new SshClient(host, credentials?.GetValueOrDefault("username", "root"), 
                credentials?.GetValueOrDefault("password", ""));
            
            client.Connect();
            
            if (client.IsConnected)
            {
                using var cmd = client.CreateCommand(script);
                cmd.Execute();
                
                result.Message = cmd.Result;
                _logger.LogInformation("SSH deployment to {Host} completed", host);
            }
        }
        catch (Exception ex)
        {
            result.Success = false;
            result.Message = $"SSH connection failed: {ex.Message}";
        }

        return result;
    }

    private async Task<DeploymentResult> DeployViaWinrmAsync(string host, string script, Dictionary<string, string>? credentials)
    {
        var result = new DeploymentResult { Success = true };
        
        try
        {
            var psScript = Convert.ToBase64String(System.Text.Encoding.UTF8.GetBytes(script));
            
            var psi = new ProcessStartInfo
            {
                FileName = "powershell",
                Arguments = $"-ExecutionPolicy Bypass -EncodedCommand {psScript}",
                UseShellExecute = false,
                RedirectStandardOutput = true,
                RedirectStandardError = true
            };

            using var process = Process.Start(psi);
            if (process != null)
            {
                await process.WaitForExitAsync();
                result.Message = await process.StandardOutput.ReadToEndAsync();
            }
        }
        catch (Exception ex)
        {
            result.Success = false;
            result.Message = $"WinRM deployment failed: {ex.Message}";
        }

        return result;
    }

    private async Task<DeploymentResult> GenerateManualInstructionsAsync(DeploymentRequest request)
    {
        var serverUrl = await GetConfiguredServerUrlAsync();
        
        return new DeploymentResult
        {
            Success = true,
            Message = "",
            Instructions = request.AgentType is "hyperv" or "mssql"
                ? $@"
Windows Agent Installation:
1. Download installation script:
   Invoke-WebRequest -Uri 'https://get.backupsystem.com/agent/install.ps1' -OutFile install.ps1

2. Run installation:
   .\install.ps1 -Server {serverUrl} -Token {request.AgentToken} -AgentType {request.AgentType} -AutoStart
"
                : $@"
Linux Agent Installation:
1. Run installation:
   curl -fsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- --server {serverUrl} --token {request.AgentToken} --agent-type {request.AgentType} --auto-start
"
        };
    }

    public async Task<AgentStatus> GetAgentStatusAsync(long agentId)
    {
        var agent = await _db.Agents.FindAsync(agentId);
        if (agent == null)
            return AgentStatus.Offline;

        if (agent.LastHeartbeat.HasValue && 
            agent.LastHeartbeat.Value > DateTime.UtcNow.AddMinutes(-5))
        {
            return AgentStatus.Online;
        }

        return AgentStatus.Offline;
    }

    public async Task<List<AgentMetrics>> GetAgentMetricsAsync(long agentId)
    {
        return new List<AgentMetrics>();
    }

    private async Task<string> GetConfiguredServerUrlAsync()
    {
        var settingValue = await _db.Settings
            .Where(s => s.Key == "server.public_url")
            .Select(s => s.Value)
            .FirstOrDefaultAsync();

        if (!string.IsNullOrWhiteSpace(settingValue))
            return settingValue;

        return _configuration["Server:PublicUrl"] ?? "http://localhost:8050";
    }
}

public class DeploymentRequest
{
    public string AgentId { get; set; } = string.Empty;
    public string AgentType { get; set; } = "hyperv";
    public string AgentToken { get; set; } = string.Empty;
    public string TargetHost { get; set; } = string.Empty;
    public string DeployMethod { get; set; } = "ssh";
    public Dictionary<string, string>? Credentials { get; set; }
}

public class DeploymentResult
{
    public bool Success { get; set; }
    public string Message { get; set; } = string.Empty;
    public string? Instructions { get; set; }
}

public enum AgentStatus
{
    Online,
    Offline,
    Error
}

public class AgentMetrics
{
    public DateTime Timestamp { get; set; }
    public double CpuPercent { get; set; }
    public long MemoryUsedMb { get; set; }
    public double NetworkSpeedMbps { get; set; }
    public double DiskSpeedMbps { get; set; }
}
