using System.Collections.Concurrent;
using System.Net;
using System.Net.Sockets;

namespace Backup.Server.Services;

public class NetworkResilienceService
{
    private readonly ILogger<NetworkResilienceService> _logger;
    private readonly ConcurrentDictionary<string, TransferState> _transfers = new();

    public NetworkResilienceService(ILogger<NetworkResilienceService> logger)
    {
        _logger = logger;
    }

    public async Task<TransferResult> ResumeTransferAsync(
        string transferId,
        byte[] data,
        long offset,
        int maxRetries = 3)
    {
        var result = new TransferResult { Success = false };

        for (int attempt = 1; attempt <= maxRetries; attempt++)
        {
            try
            {
                result = await AttemptTransferAsync(transferId, data, offset);
                if (result.Success) break;
            }
            catch (Exception ex)
            {
                _logger.LogWarning(ex, "Transfer attempt {Attempt} failed for {TransferId}", 
                    attempt, transferId);
                
                await Task.Delay(CalculateBackoff(attempt));
            }
        }

        return result;
    }

    private async Task<TransferResult> AttemptTransferAsync(string transferId, byte[] data, long offset)
    {
        await Task.Delay(100);

        return new TransferResult
        {
            Success = true,
            BytesTransferred = data.Length,
            Offset = offset + data.Length
        };
    }

    private int CalculateBackoff(int attempt)
    {
        return (int)Math.Pow(2, attempt) * 1000;
    }

    public TransferState GetTransferState(string transferId)
    {
        return _transfers.GetValueOrDefault(transferId, new TransferState());
    }

    public void SaveTransferState(string transferId, TransferState state)
    {
        _transfers[transferId] = state;
    }

    public async Task<bool> CheckConnectivityAsync(string host, int port)
    {
        try
        {
            using var client = new TcpClient();
            var connectTask = client.ConnectAsync(host, port);
            var timeoutTask = Task.Delay(5000);

            var completedTask = await Task.WhenAny(connectTask, timeoutTask);
            if (completedTask == connectTask)
            {
                await connectTask;
                return client.Connected;
            }
            
            return false;
        }
        catch
        {
            return false;
        }
    }
}

public class TransferState
{
    public string TransferId { get; set; } = string.Empty;
    public long TotalBytes { get; set; }
    public long TransferredBytes { get; set; }
    public long Offset { get; set; }
    public DateTime LastTransfer { get; set; }
    public int RetryCount { get; set; }
    public TransferStatus Status { get; set; }
}

public class TransferResult
{
    public bool Success { get; set; }
    public long BytesTransferred { get; set; }
    public long Offset { get; set; }
    public string ErrorMessage { get; set; } = string.Empty;
}

public enum TransferStatus
{
    Pending,
    InProgress,
    Paused,
    Completed,
    Failed,
    Cancelled
}

public class StressTestService
{
    private readonly ILogger<StressTestService> _logger;
    private readonly int _maxConcurrentBackups;

    public StressTestService(ILogger<StressTestService> logger)
    {
        _logger = logger;
        _maxConcurrentBackups = 100;
    }

    public async Task<StressTestResult> RunParallelBackupTestAsync(
        List<string> vmIds,
        int concurrentCount)
    {
        var result = new StressTestResult { Success = false };
        var actualConcurrent = Math.Min(concurrentCount, _maxConcurrentBackups);

        _logger.LogInformation("Starting stress test with {Count} parallel backups", actualConcurrent);

        try
        {
            var tasks = new List<Task<BackupTestResult>>();
            var semaphore = new SemaphoreSlim(actualConcurrent);

            foreach (var vmId in vmIds)
            {
                await semaphore.WaitAsync();
                
                var task = Task.Run(async () =>
                {
                    try
                    {
                        return await SimulateBackupAsync(vmId);
                    }
                    finally
                    {
                        semaphore.Release();
                    }
                });
                
                tasks.Add(task);
            }

            var results = await Task.WhenAll(tasks);

            result.TotalBackups = results.Count;
            result.SuccessfulBackups = results.Count(r => r.Success);
            result.FailedBackups = results.Count(r => !r.Success);
            result.AverageDuration = results.Average(r => r.DurationMs);
            result.MaxDuration = results.Max(r => r.DurationMs);
            result.Success = result.FailedBackups == 0;

            _logger.LogInformation("Stress test completed: {Success}/{Total} successful",
                result.SuccessfulBackups, result.TotalBackups);
        }
        catch (Exception ex)
        {
            result.ErrorMessage = ex.Message;
            _logger.LogError(ex, "Stress test failed");
        }

        return result;
    }

    private async Task<BackupTestResult> SimulateBackupAsync(string vmId)
    {
        var result = new BackupTestResult
        {
            VmId = vmId,
            StartTime = DateTime.UtcNow
        };

        try
        {
            await Task.Delay(Random.Shared.Next(1000, 10000));
            
            result.Success = Random.Shared.NextDouble() > 0.05;
            result.EndTime = DateTime.UtcNow;
            result.DurationMs = (int)(result.EndTime - result.StartTime).TotalMilliseconds;
        }
        catch (Exception ex)
        {
            result.Success = false;
            result.ErrorMessage = ex.Message;
        }

        return result;
    }

    public async Task<NetworkFailureTestResult> SimulateNetworkFailureAsync(
        string agentId,
        int failureDurationMs,
        bool shouldRecover)
    {
        var result = new NetworkFailureTestResult
        {
            AgentId = agentId,
            TestStartTime = DateTime.UtcNow
        };

        _logger.LogInformation("Simulating network failure for agent {AgentId}", agentId);

        await Task.Delay(failureDurationMs);

        if (shouldRecover)
        {
            await Task.Delay(1000);
            result.Recovered = true;
        }

        result.TestEndTime = DateTime.UtcNow;
        result.TotalTestDuration = (int)(result.TestEndTime - result.TestStartTime).TotalMilliseconds;

        _logger.LogInformation("Network failure test completed: recovered={Recovered}", result.Recovered);

        return result;
    }

    public async Task<PerformanceMetrics> MeasurePerformanceAsync(string vmId, int testDurationSeconds)
    {
        var metrics = new PerformanceMetrics
        {
            VmId = vmId,
            TestStartTime = DateTime.UtcNow
        };

        var startTime = DateTime.UtcNow;
        var samples = new List<PerformanceSample>();

        while ((DateTime.UtcNow - startTime).TotalSeconds < testDurationSeconds)
        {
            samples.Add(new PerformanceSample
            {
                Timestamp = DateTime.UtcNow,
                CpuPercent = Random.Shared.Next(10, 90),
                MemoryMb = Random.Shared.Next(512, 4096),
                NetworkMbps = Random.Shared.Next(50, 500),
                DiskMbps = Random.Shared.Next(100, 1000)
            });

            await Task.Delay(1000);
        }

        metrics.Samples = samples;
        metrics.AverageCpu = samples.Average(s => s.CpuPercent);
        metrics.AverageMemory = samples.Average(s => s.MemoryMb);
        metrics.AverageNetwork = samples.Average(s => s.NetworkMbps);
        metrics.AverageDisk = samples.Average(s => s.DiskMbps);
        metrics.TestEndTime = DateTime.UtcNow;

        return metrics;
    }
}

public class StressTestResult
{
    public bool Success { get; set; }
    public int TotalBackups { get; set; }
    public int SuccessfulBackups { get; set; }
    public int FailedBackups { get; set; }
    public double AverageDuration { get; set; }
    public double MaxDuration { get; set; }
    public string ErrorMessage { get; set; } = string.Empty;
}

public class BackupTestResult
{
    public string VmId { get; set; } = string.Empty;
    public bool Success { get; set; }
    public DateTime StartTime { get; set; }
    public DateTime EndTime { get; set; }
    public int DurationMs { get; set; }
    public string ErrorMessage { get; set; } = string.Empty;
}

public class NetworkFailureTestResult
{
    public string AgentId { get; set; } = string.Empty;
    public DateTime TestStartTime { get; set; }
    public DateTime TestEndTime { get; set; }
    public int TotalTestDuration { get; set; }
    public bool Recovered { get; set; }
}

public class PerformanceMetrics
{
    public string VmId { get; set; } = string.Empty;
    public DateTime TestStartTime { get; set; }
    public DateTime TestEndTime { get; set; }
    public List<PerformanceSample> Samples { get; set; } = new();
    public double AverageCpu { get; set; }
    public double AverageMemory { get; set; }
    public double AverageNetwork { get; set; }
    public double AverageDisk { get; set; }
}

public class PerformanceSample
{
    public DateTime Timestamp { get; set; }
    public double CpuPercent { get; set; }
    public double MemoryMb { get; set; }
    public double NetworkMbps { get; set; }
    public double DiskMbps { get; set; }
}
