using Backup.Server.Database.Entities;
using Backup.Server.Services;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;

namespace Backup.Server.Tests;

public class SchedulerServiceTests
{
    private readonly SchedulerService _service;

    public SchedulerServiceTests()
    {
        var logger = new Mock<ILogger<SchedulerService>>();
        _service = new SchedulerService(logger.Object);
    }

    [Fact]
    public void CalculateNextRun_ShouldParseDirectCron()
    {
        var job = new Job
        {
            JobId = "job-1",
            Schedule = "0 2 * * *"
        };

        var nextRun = _service.CalculateNextRun(job, new DateTime(2026, 4, 7, 1, 0, 0, DateTimeKind.Utc));

        Assert.Equal(new DateTime(2026, 4, 7, 2, 0, 0, DateTimeKind.Utc), nextRun);
    }

    [Fact]
    public void CalculateNextRun_ShouldParseJsonCronConfig()
    {
        var job = new Job
        {
            JobId = "job-2",
            Schedule = "{\"cron\":\"0 3 * * *\"}"
        };

        var nextRun = _service.CalculateNextRun(job, new DateTime(2026, 4, 7, 2, 0, 0, DateTimeKind.Utc));

        Assert.Equal(new DateTime(2026, 4, 7, 3, 0, 0, DateTimeKind.Utc), nextRun);
    }

    [Fact]
    public void CalculateNextRun_ShouldReturnNullForInvalidSchedule()
    {
        var job = new Job
        {
            JobId = "job-3",
            Schedule = "not-a-cron"
        };

        var nextRun = _service.CalculateNextRun(job);

        Assert.Null(nextRun);
    }
}
