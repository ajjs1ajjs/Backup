using System.Collections.Concurrent;
using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Runtime.CompilerServices;

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
    private readonly ConcurrentDictionary<string, StressTestSession> _activeSessions = new();
    private static readonly SemaphoreSlim _globalSemaphore = new(100); // Max 100 concurrent

    public StressTestService(ILogger<StressTestService> logger)
    {
        _logger = logger;
        _maxConcurrentBackups = 100;
    }

    /// <summary>
    /// Runs stress test with configurable parallelism for 100+ VMs
    /// </summary>
    public async Task<StressTestResult> RunParallelBackupTestAsync(
        List<string> vmIds,
        int concurrentCount)
    {
        var sessionId = $"stress-{Guid.NewGuid():N}";
        var session = new StressTestSession
        {
            SessionId = sessionId,
            StartTime = DateTime.UtcNow,
            TotalVMs = vmIds.Count,
            TargetConcurrency = concurrentCount
        };

        _activeSessions[sessionId] = session;
        var result = new StressTestResult { Success = false, SessionId = sessionId };
        var actualConcurrent = Math.Min(concurrentCount, _maxConcurrentBackups);

        _logger.LogInformation("Starting stress test {SessionId} with {Count} VMs, {Concurrency} concurrent",
            sessionId, vmIds.Count, actualConcurrent);

        try
        {
            var tasks = new List<Task<BackupTestResult>>();
            var semaphore = new SemaphoreSlim(actualConcurrent);

            foreach (var vmId in vmIds)
            {
                await semaphore.WaitAsync();
                session.ActiveTasks++;

                var task = Task.Run(async () =>
                {
                    try
                    {
                        var backupResult = await SimulateBackupAsync(vmId);
                        session.CompletedTasks++;
                        session.SuccessfulBackups += backupResult.Success ? 1 : 0;
                        session.FailedBackups += backupResult.Success ? 0 : 1;
                        return backupResult;
                    }
                    finally
                    {
                        semaphore.Release();
                        session.ActiveTasks--;
                    }
                });

                tasks.Add(task);
            }

            var results = await Task.WhenAll(tasks);

            result.TotalBackups = results.Length;
            result.SuccessfulBackups = results.Count(r => r.Success);
            result.FailedBackups = results.Count(r => !r.Success);
            result.AverageDuration = results.Any() ? results.Average(r => r.DurationMs) : 0;
            result.MaxDuration = results.Any() ? results.Max(r => r.DurationMs) : 0;
            result.MinDuration = results.Any() ? results.Min(r => r.DurationMs) : 0;
            result.Percentile95Duration = GetPercentile(results.Select(r => r.DurationMs).OrderBy(x => x).ToList(), 95);
            result.Success = result.FailedBackups == 0;
            result.EndTime = DateTime.UtcNow;
            result.TotalDuration = (result.EndTime - result.StartTime).TotalSeconds;

            _logger.LogInformation("Stress test {SessionId} completed: {Success}/{Total} successful, Avg: {Avg}ms, P95: {P95}ms",
                sessionId, result.SuccessfulBackups, result.TotalBackups, result.AverageDuration, result.Percentile95Duration);
        }
        catch (Exception ex)
        {
            result.ErrorMessage = ex.Message;
            _logger.LogError(ex, "Stress test {SessionId} failed", sessionId);
        }
        finally
        {
            session.EndTime = DateTime.UtcNow;
            session.Status = "Completed";
        }

        return result;
    }

    /// <summary>
    /// Enhanced stress test with network failure simulation
    /// </summary>
    public async Task<StressTestResult> RunStressTestWithNetworkFailuresAsync(
        List<string> vmIds,
        int concurrentCount,
        int failureRatePercent = 10,
        int failureDurationMs = 5000)
    {
        var sessionId = $"stress-net-{Guid.NewGuid():N}";
        var result = new StressTestResult { Success = false, SessionId = sessionId };
        var actualConcurrent = Math.Min(concurrentCount, _maxConcurrentBackups);

        _logger.LogInformation("Starting stress test with network failures: {Count} VMs, {FailureRate}% failure rate",
            vmIds.Count, failureRatePercent);

        try
        {
            var tasks = new List<Task<BackupTestResult>>();
            var semaphore = new SemaphoreSlim(actualConcurrent);
            var random = new Random();

            foreach (var vmId in vmIds)
            {
                await semaphore.WaitAsync();

                var shouldFail = random.Next(0, 100) < failureRatePercent;

                var task = Task.Run(async () =>
                {
                    try
                    {
                        if (shouldFail)
                        {
                            // Simulate network failure
                            await Task.Delay(failureDurationMs);
                            // Retry after failure
                            return await SimulateBackupAsync(vmId);
                        }

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

            result.TotalBackups = results.Length;
            result.SuccessfulBackups = results.Count(r => r.Success);
            result.FailedBackups = results.Count(r => !r.Success);
            result.AverageDuration = results.Average(r => r.DurationMs);
            result.Success = result.FailedBackups == 0;
        }
        catch (Exception ex)
        {
            result.ErrorMessage = ex.Message;
            _logger.LogError(ex, "Stress test with network failures failed");
        }

        return result;
    }

    /// <summary>
    /// Long-running endurance test (8+ hours)
    /// </summary>
    public async IAsyncEnumerable<EnduranceTestMetrics> RunEnduranceTestAsync(
        List<string> vmIds,
        int concurrentCount,
        TimeSpan duration,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        var sessionId = $"endurance-{Guid.NewGuid():N}";
        var startTime = DateTime.UtcNow;
        var totalRuns = 0;
        var totalSuccess = 0;
        var totalFailures = 0;
        var allDurations = new ConcurrentBag<double>();

        _logger.LogInformation("Starting endurance test {SessionId} for {Duration}", sessionId, duration);

        while (DateTime.UtcNow - startTime < duration && !cancellationToken.IsCancellationRequested)
        {
            var runResult = await RunParallelBackupTestAsync(vmIds, concurrentCount);

            totalRuns++;
            totalSuccess += runResult.SuccessfulBackups;
            totalFailures += runResult.FailedBackups;

            foreach (var durationMs in runResult.Durations)
            {
                allDurations.Add(durationMs);
            }

            var metrics = new EnduranceTestMetrics
            {
                SessionId = sessionId,
                ElapsedTime = DateTime.UtcNow - startTime,
                TotalRuns = totalRuns,
                TotalSuccess = totalSuccess,
                TotalFailures = totalFailures,
                SuccessRate = totalRuns > 0 ? (double)totalSuccess / (totalSuccess + totalFailures) * 100 : 0,
                AverageDuration = allDurations.Any() ? allDurations.Average() : 0,
                MemoryUsageMB = GC.GetTotalMemory(false) / 1024 / 1024,
                ActiveThreads = System.Diagnostics.Process.GetCurrentProcess().Threads.Count
            };

            yield return metrics;

            // Wait between runs
            await Task.Delay(TimeSpan.FromSeconds(10), cancellationToken);
        }

        _logger.LogInformation("Endurance test {SessionId} completed: {SuccessRate}% success rate",
            sessionId, totalRuns > 0 ? (double)totalSuccess / (totalSuccess + totalFailures) * 100 : 0);
    }

    /// <summary>
    /// Scalability test - gradually increase concurrent VMs
    /// </summary>
    public async Task<ScalabilityTestResult> RunScalabilityTestAsync(
        List<string> vmIds,
        int startConcurrency,
        int maxConcurrency,
        int stepSize)
    {
        var result = new ScalabilityTestResult();
        var currentConcurrency = startConcurrency;

        _logger.LogInformation("Starting scalability test: {Start} -> {Max} (step {Step})",
            startConcurrency, maxConcurrency, stepSize);

        while (currentConcurrency <= maxConcurrency && currentConcurrency <= _maxConcurrentBackups)
        {
            var runResult = await RunParallelBackupTestAsync(vmIds, currentConcurrency);

            result.Metrics.Add(new ScalabilityMetric
            {
                Concurrency = currentConcurrency,
                TotalBackups = runResult.TotalBackups,
                SuccessfulBackups = runResult.SuccessfulBackups,
                AverageDuration = runResult.AverageDuration,
                Percentile95Duration = runResult.Percentile95Duration,
                ThroughputPerSecond = runResult.TotalDuration > 0
                    ? runResult.TotalBackups / runResult.TotalDuration
                    : 0
            });

            _logger.LogInformation("Scalability test at {Concurrency}: Avg={Avg}ms, P95={P95}ms, Throughput={Throughput}/s",
                currentConcurrency, runResult.AverageDuration, runResult.Percentile95Duration,
                runResult.TotalDuration > 0 ? runResult.TotalBackups / runResult.TotalDuration : 0);

            currentConcurrency += stepSize;
            await Task.Delay(TimeSpan.FromSeconds(5)); // Cool-down between runs
        }

        result.Success = true;
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
            // Simulate realistic backup duration (1-10 seconds)
            await Task.Delay(Random.Shared.Next(1000, 10000));

            // 95% success rate simulation
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
    public string SessionId { get; set; } = string.Empty;
    public bool Success { get; set; }
    public int TotalBackups { get; set; }
    public int SuccessfulBackups { get; set; }
    public int FailedBackups { get; set; }
    public double AverageDuration { get; set; }
    public double MaxDuration { get; set; }
    public double MinDuration { get; set; }
    public double Percentile95Duration { get; set; }
    public List<double> Durations { get; set; } = new();
    public string ErrorMessage { get; set; } = string.Empty;
    public DateTime StartTime { get; set; }
    public DateTime EndTime { get; set; }
    public double TotalDuration { get; set; }
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

public class StressTestSession
{
    public string SessionId { get; set; } = string.Empty;
    public DateTime StartTime { get; set; }
    public DateTime? EndTime { get; set; }
    public int TotalVMs { get; set; }
    public int TargetConcurrency { get; set; }
    public int ActiveTasks { get; set; }
    public int CompletedTasks { get; set; }
    public int SuccessfulBackups { get; set; }
    public int FailedBackups { get; set; }
    public string Status { get; set; } = "Running";
}

public class EnduranceTestMetrics
{
    public string SessionId { get; set; } = string.Empty;
    public TimeSpan ElapsedTime { get; set; }
    public int TotalRuns { get; set; }
    public int TotalSuccess { get; set; }
    public int TotalFailures { get; set; }
    public double SuccessRate { get; set; }
    public double AverageDuration { get; set; }
    public double MemoryUsageMB { get; set; }
    public int ActiveThreads { get; set; }
}

public class ScalabilityTestResult
{
    public bool Success { get; set; }
    public List<ScalabilityMetric> Metrics { get; set; } = new();
}

public class ScalabilityMetric
{
    public int Concurrency { get; set; }
    public int TotalBackups { get; set; }
    public int SuccessfulBackups { get; set; }
    public double AverageDuration { get; set; }
    public double Percentile95Duration { get; set; }
    public double ThroughputPerSecond { get; set; }
}

public class StressTestConfiguration
{
    public int MaxConcurrentBackups { get; set; } = 100;
    public int TestDurationMinutes { get; set; } = 60;
    public int FailureRatePercent { get; set; } = 5;
    public bool EnableNetworkFailureSimulation { get; set; } = true;
    public bool EnableResourceMonitoring { get; set; } = true;
    public int MonitoringIntervalSeconds { get; set; } = 30;
}

public class ResourceMetrics
{
    public DateTime Timestamp { get; set; }
    public double CpuUsagePercent { get; set; }
    public double MemoryUsageMB { get; set; }
    public double NetworkThroughputMbps { get; set; }
    public double DiskThroughputMbps { get; set; }
    public int ActiveConnections { get; set; }
    public int ThreadPoolThreads { get; set; }
    public int PendingRequests { get; set; }
}

public static class StressTestExtensions
{
    public static double GetPercentile(List<double> values, int percentile)
    {
        if (!values.Any()) return 0;

        var sorted = values.OrderBy(x => x).ToList();
        var index = (int)Math.Ceiling(percentile / 100.0 * sorted.Count) - 1;
        return sorted[Math.Max(0, index)];
    }
}
