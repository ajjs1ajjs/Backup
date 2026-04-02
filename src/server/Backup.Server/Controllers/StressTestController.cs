using Microsoft.AspNetCore.Mvc;
using Backup.Server.Services;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
public class StressTestController : ControllerBase
{
    private readonly StressTestService _stressTestService;
    private readonly ILogger<StressTestController> _logger;

    public StressTestController(
        StressTestService stressTestService,
        ILogger<StressTestController> logger)
    {
        _stressTestService = stressTestService;
        _logger = logger;
    }

    /// <summary>
    /// Run parallel backup stress test with 100+ VMs
    /// </summary>
    /// <param name="request">Stress test configuration</param>
    /// <returns>Stress test results</returns>
    [HttpPost("run")]
    public async Task<ActionResult<StressTestResponse>> RunStressTest(
        [FromBody] RunStressTestRequest request)
    {
        try
        {
            _logger.LogInformation("Starting stress test: {VMCount} VMs, {Concurrency} concurrent",
                request.VMIds?.Count ?? 0, request.ConcurrentCount);

            var vmIds = request.VMIds ?? GenerateVMIds(request.VMCount, "vm-stress");
            var result = await _stressTestService.RunParallelBackupTestAsync(
                vmIds,
                request.ConcurrentCount);

            return Ok(new StressTestResponse
            {
                Success = result.Success,
                SessionId = result.SessionId,
                TotalBackups = result.TotalBackups,
                SuccessfulBackups = result.SuccessfulBackups,
                FailedBackups = result.FailedBackups,
                AverageDurationMs = result.AverageDuration,
                MinDurationMs = result.MinDuration,
                MaxDurationMs = result.MaxDuration,
                Percentile95DurationMs = result.Percentile95Duration,
                TotalDurationSeconds = result.TotalDuration,
                ErrorMessage = result.ErrorMessage
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Stress test failed");
            return BadRequest(new StressTestResponse
            {
                Success = false,
                ErrorMessage = ex.Message
            });
        }
    }

    /// <summary>
    /// Run stress test with network failure simulation
    /// </summary>
    [HttpPost("run-with-failures")]
    public async Task<ActionResult<StressTestResponse>> RunStressTestWithFailures(
        [FromBody] RunStressTestWithFailuresRequest request)
    {
        try
        {
            _logger.LogInformation("Starting stress test with network failures: {VMCount} VMs, {FailureRate}% failure rate",
                request.VMIds?.Count ?? 0, request.FailureRatePercent);

            var vmIds = request.VMIds ?? GenerateVMIds(request.VMCount, "vm-failure");
            var result = await _stressTestService.RunStressTestWithNetworkFailuresAsync(
                vmIds,
                request.ConcurrentCount,
                request.FailureRatePercent,
                request.FailureDurationMs);

            return Ok(new StressTestResponse
            {
                Success = result.Success,
                SessionId = result.SessionId,
                TotalBackups = result.TotalBackups,
                SuccessfulBackups = result.SuccessfulBackups,
                FailedBackups = result.FailedBackups,
                AverageDurationMs = result.AverageDuration,
                TotalDurationSeconds = result.TotalDuration,
                ErrorMessage = result.ErrorMessage
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Stress test with failures failed");
            return BadRequest(new StressTestResponse
            {
                Success = false,
                ErrorMessage = ex.Message
            });
        }
    }

    /// <summary>
    /// Run scalability test - gradually increase concurrency
    /// </summary>
    [HttpPost("scalability")]
    public async Task<ActionResult<ScalabilityTestResponse>> RunScalabilityTest(
        [FromBody] RunScalabilityTestRequest request)
    {
        try
        {
            _logger.LogInformation("Starting scalability test: {Start} -> {Max} (step {Step})",
                request.StartConcurrency, request.MaxConcurrency, request.StepSize);

            var vmIds = request.VMIds ?? GenerateVMIds(request.VMCount, "vm-scale");
            var result = await _stressTestService.RunScalabilityTestAsync(
                vmIds,
                request.StartConcurrency,
                request.MaxConcurrency,
                request.StepSize);

            return Ok(new ScalabilityTestResponse
            {
                Success = result.Success,
                Metrics = result.Metrics.Select(m => new ScalabilityMetricResponse
                {
                    Concurrency = m.Concurrency,
                    TotalBackups = m.TotalBackups,
                    SuccessfulBackups = m.SuccessfulBackups,
                    AverageDurationMs = m.AverageDuration,
                    Percentile95DurationMs = m.Percentile95Duration,
                    ThroughputPerSecond = m.ThroughputPerSecond
                }).ToList()
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Scalability test failed");
            return BadRequest(new ScalabilityTestResponse
            {
                Success = false,
                ErrorMessage = ex.Message
            });
        }
    }

    /// <summary>
    /// Run endurance test (long-running)
    /// </summary>
    [HttpPost("endurance")]
    public async Task<ActionResult> RunEnduranceTest(
        [FromBody] RunEnduranceTestRequest request)
    {
        try
        {
            _logger.LogInformation("Starting endurance test: {VMCount} VMs, {Concurrency} concurrent, {Duration} minutes",
                request.VMCount, request.ConcurrentCount, request.DurationMinutes);

            var vmIds = GenerateVMIds(request.VMCount, "vm-endurance");
            var duration = TimeSpan.FromMinutes(request.DurationMinutes);

            // Start background test and return session ID
            _ = Task.Run(async () =>
            {
                await foreach (var metrics in _stressTestService.RunEnduranceTestAsync(
                    vmIds, request.ConcurrentCount, duration))
                {
                    _logger.LogInformation("Endurance test progress: {SessionId}, Run {Runs}, SuccessRate {Rate}%",
                        metrics.SessionId, metrics.TotalRuns, metrics.SuccessRate);
                }
            });

            return Ok(new { Message = "Endurance test started", DurationMinutes = request.DurationMinutes });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Endurance test failed to start");
            return BadRequest(new { Error = ex.Message });
        }
    }

    /// <summary>
    /// Get performance metrics for a specific VM
    /// </summary>
    [HttpGet("performance/{vmId}")]
    public async Task<ActionResult<PerformanceMetricsResponse>> GetPerformanceMetrics(
        string vmId,
        [FromQuery] int durationSeconds = 30)
    {
        try
        {
            var metrics = await _stressTestService.MeasurePerformanceAsync(vmId, durationSeconds);

            return Ok(new PerformanceMetricsResponse
            {
                VmId = metrics.VmId,
                AverageCpuPercent = metrics.AverageCpu,
                AverageMemoryMB = metrics.AverageMemory,
                AverageNetworkMbps = metrics.AverageNetwork,
                AverageDiskMbps = metrics.AverageDisk,
                Samples = metrics.Samples.Select(s => new PerformanceSampleResponse
                {
                    Timestamp = s.Timestamp,
                    CpuPercent = s.CpuPercent,
                    MemoryMB = s.MemoryMb,
                    NetworkMbps = s.NetworkMbps,
                    DiskMbps = s.DiskMbps
                }).ToList()
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to get performance metrics");
            return BadRequest(new { Error = ex.Message });
        }
    }

    /// <summary>
    /// Simulate network failure for testing resilience
    /// </summary>
    [HttpPost("simulate-network-failure")]
    public async Task<ActionResult<NetworkFailureTestResponse>> SimulateNetworkFailure(
        [FromBody] SimulateNetworkFailureRequest request)
    {
        try
        {
            var result = await _stressTestService.SimulateNetworkFailureAsync(
                request.AgentId,
                request.FailureDurationMs,
                request.ShouldRecover);

            return Ok(new NetworkFailureTestResponse
            {
                AgentId = result.AgentId,
                TotalTestDurationMs = result.TotalTestDuration,
                Recovered = result.Recovered
            });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Network failure simulation failed");
            return BadRequest(new { Error = ex.Message });
        }
    }

    private static List<string> GenerateVMIds(int count, string prefix)
    {
        return Enumerable.Range(0, count)
            .Select(i => $"{prefix}-{i:D4}")
            .ToList();
    }
}

// Request/Response DTOs

public class RunStressTestRequest
{
    public List<string>? VMIds { get; set; }
    public int VMCount { get; set; } = 100;
    public int ConcurrentCount { get; set; } = 50;
}

public class RunStressTestWithFailuresRequest
{
    public List<string>? VMIds { get; set; }
    public int VMCount { get; set; } = 100;
    public int ConcurrentCount { get; set; } = 50;
    public int FailureRatePercent { get; set; } = 10;
    public int FailureDurationMs { get; set; } = 5000;
}

public class RunScalabilityTestRequest
{
    public List<string>? VMIds { get; set; }
    public int VMCount { get; set; } = 100;
    public int StartConcurrency { get; set; } = 10;
    public int MaxConcurrency { get; set; } = 100;
    public int StepSize { get; set; } = 10;
}

public class RunEnduranceTestRequest
{
    public int VMCount { get; set; } = 50;
    public int ConcurrentCount { get; set; } = 25;
    public int DurationMinutes { get; set; } = 480; // 8 hours
}

public class SimulateNetworkFailureRequest
{
    public string AgentId { get; set; } = string.Empty;
    public int FailureDurationMs { get; set; } = 5000;
    public bool ShouldRecover { get; set; } = true;
}

public class StressTestResponse
{
    public bool Success { get; set; }
    public string SessionId { get; set; } = string.Empty;
    public int TotalBackups { get; set; }
    public int SuccessfulBackups { get; set; }
    public int FailedBackups { get; set; }
    public double AverageDurationMs { get; set; }
    public double MinDurationMs { get; set; }
    public double MaxDurationMs { get; set; }
    public double Percentile95DurationMs { get; set; }
    public double TotalDurationSeconds { get; set; }
    public string ErrorMessage { get; set; } = string.Empty;
}

public class ScalabilityTestResponse
{
    public bool Success { get; set; }
    public List<ScalabilityMetricResponse> Metrics { get; set; } = new();
    public string ErrorMessage { get; set; } = string.Empty;
}

public class ScalabilityMetricResponse
{
    public int Concurrency { get; set; }
    public int TotalBackups { get; set; }
    public int SuccessfulBackups { get; set; }
    public double AverageDurationMs { get; set; }
    public double Percentile95DurationMs { get; set; }
    public double ThroughputPerSecond { get; set; }
}

public class PerformanceMetricsResponse
{
    public string VmId { get; set; } = string.Empty;
    public double AverageCpuPercent { get; set; }
    public double AverageMemoryMB { get; set; }
    public double AverageNetworkMbps { get; set; }
    public double AverageDiskMbps { get; set; }
    public List<PerformanceSampleResponse> Samples { get; set; } = new();
}

public class PerformanceSampleResponse
{
    public DateTime Timestamp { get; set; }
    public double CpuPercent { get; set; }
    public double MemoryMB { get; set; }
    public double NetworkMbps { get; set; }
    public double DiskMbps { get; set; }
}

public class NetworkFailureTestResponse
{
    public string AgentId { get; set; } = string.Empty;
    public int TotalTestDurationMs { get; set; }
    public bool Recovered { get; set; }
}
