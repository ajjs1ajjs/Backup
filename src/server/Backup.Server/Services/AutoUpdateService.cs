using System.Text.Json;
using System.Diagnostics;
using System.Security.Cryptography;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public interface IAutoUpdateService
{
    Task<UpdateInfo?> CheckForUpdatesAsync(string agentType, string currentVersion);
    Task<DownloadResult?> DownloadUpdateAsync(string agentType, string version);
    Task<VerifyResult> VerifyUpdateAsync(string downloadPath, string expectedChecksum);
    Task<string> GetLatestVersionAsync(string agentType);
}

public class AutoUpdateService : IAutoUpdateService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<AutoUpdateService> _logger;
    private readonly IConfiguration _configuration;
    private readonly HttpClient _httpClient;

    private static readonly Dictionary<string, string> LatestVersions = new()
    {
        { "hyperv", "1.0.0" },
        { "vmware", "1.0.0" },
        { "kvm", "1.0.0" },
        { "mssql", "1.0.0" },
        { "postgresql", "1.0.0" },
        { "oracle", "1.0.0" }
    };

    public AutoUpdateService(
        BackupDbContext db,
        ILogger<AutoUpdateService> logger,
        IConfiguration configuration,
        HttpClient httpClient)
    {
        _db = db;
        _logger = logger;
        _configuration = configuration;
        _httpClient = httpClient;
    }

    public async Task<UpdateInfo?> CheckForUpdatesAsync(string agentType, string currentVersion)
    {
        var latestVersion = await GetLatestVersionAsync(agentType);
        
        if (string.IsNullOrEmpty(latestVersion))
            return null;

        if (CompareVersions(latestVersion, currentVersion) > 0)
        {
            return new UpdateInfo
            {
                AgentType = agentType,
                Version = latestVersion,
                ReleaseDate = DateTime.UtcNow,
                DownloadUrl = $"{_configuration["UpdateServer:BaseUrl"]}/agents/{agentType}-{latestVersion}.zip",
                Checksum = await CalculateChecksumAsync($"{agentType}-{latestVersion}.zip"),
                ReleaseNotes = "Performance improvements and bug fixes"
            };
        }

        return null;
    }

    public async Task<DownloadResult?> DownloadUpdateAsync(string agentType, string version)
    {
        var downloadUrl = $"{_configuration["UpdateServer:BaseUrl"]}/agents/{agentType}-{version}.zip";
        var outputPath = Path.Combine(Path.GetTempPath(), $"{agentType}-{version}.zip");

        try
        {
            var response = await _httpClient.GetAsync(downloadUrl);
            response.EnsureSuccessStatusCode();

            await using var fs = new FileStream(outputPath, FileMode.Create);
            await response.Content.CopyToAsync(fs);

            _logger.LogInformation("Downloaded update for {AgentType} v{Version}", agentType, version);

            return new DownloadResult
            {
                Success = true,
                FilePath = outputPath,
                FileSize = new FileInfo(outputPath).Length
            };
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to download update for {AgentType} v{Version}", agentType, version);
            return new DownloadResult
            {
                Success = false,
                ErrorMessage = ex.Message
            };
        }
    }

    public async Task<VerifyResult> VerifyUpdateAsync(string downloadPath, string expectedChecksum)
    {
        try
        {
            var actualChecksum = await CalculateFileChecksumAsync(downloadPath);
            var isValid = actualChecksum.Equals(expectedChecksum, StringComparison.OrdinalIgnoreCase);

            return new VerifyResult
            {
                Success = true,
                IsValid = isValid,
                ActualChecksum = actualChecksum
            };
        }
        catch (Exception ex)
        {
            return new VerifyResult
            {
                Success = false,
                ErrorMessage = ex.Message
            };
        }
    }

    public Task<string> GetLatestVersionAsync(string agentType)
    {
        return Task.FromResult(LatestVersions.GetValueOrDefault(agentType.ToLower(), "1.0.0"));
    }

    private static int CompareVersions(string v1, string v2)
    {
        var parts1 = v1.Split('.').Select(int.Parse).ToArray();
        var parts2 = v2.Split('.').Select(int.Parse).ToArray();

        for (int i = 0; i < Math.Max(parts1.Length, parts2.Length); i++)
        {
            var p1 = i < parts1.Length ? parts1[i] : 0;
            var p2 = i < parts2.Length ? parts2[i] : 0;

            if (p1 > p2) return 1;
            if (p1 < p2) return -1;
        }

        return 0;
    }

    private Task<string> CalculateChecksumAsync(string filename)
    {
        var hash = SHA256.HashData(System.Text.Encoding.UTF8.GetBytes(filename));
        return Task.FromResult(Convert.ToHexString(hash).ToLower());
    }

    private static async Task<string> CalculateFileChecksumAsync(string filePath)
    {
        await using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
        var hash = await SHA256.HashDataAsync(fs);
        return Convert.ToHexString(hash).ToLower();
    }
}

public class UpdateInfo
{
    public string AgentType { get; set; } = string.Empty;
    public string Version { get; set; } = string.Empty;
    public DateTime ReleaseDate { get; set; }
    public string DownloadUrl { get; set; } = string.Empty;
    public string Checksum { get; set; } = string.Empty;
    public string ReleaseNotes { get; set; } = string.Empty;
}

public class DownloadResult
{
    public bool Success { get; set; }
    public string? FilePath { get; set; }
    public long FileSize { get; set; }
    public string? ErrorMessage { get; set; }
}

public class VerifyResult
{
    public bool Success { get; set; }
    public bool IsValid { get; set; }
    public string? ActualChecksum { get; set; }
    public string? ErrorMessage { get; set; }
}
