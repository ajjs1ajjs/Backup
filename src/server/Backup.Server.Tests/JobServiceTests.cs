using Backup.Server.Database.Entities;
using Backup.Server.Database;
using Backup.Server.Services;
using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
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

public class BackupExecutionServiceTests
{
    [Fact]
    public async Task ExecuteRunAsync_ShouldCreateBackupPoint_ForLocalFileJob()
    {
        await using var connection = new SqliteConnection("Data Source=:memory:");
        await connection.OpenAsync();

        var options = new DbContextOptionsBuilder<BackupDbContext>()
            .UseSqlite(connection)
            .Options;

        await using var db = new BackupDbContext(options);
        await db.Database.EnsureCreatedAsync();

        var logger = new Mock<ILogger<BackupExecutionService>>();
        var service = new BackupExecutionService(db, logger.Object);

        var tempRoot = Path.Combine(Path.GetTempPath(), "backup-tests", Guid.NewGuid().ToString("N"));
        var sourceFile = Path.Combine(tempRoot, "source.txt");
        var repositoryPath = Path.Combine(tempRoot, "repo");
        Directory.CreateDirectory(tempRoot);
        await File.WriteAllTextAsync(sourceFile, "backup test payload");

        db.Repositories.Add(new Repository
        {
            RepositoryId = "repo-1",
            Name = "Local Repo",
            Type = RepositoryType.Local,
            Path = repositoryPath
        });

        db.Jobs.Add(new Job
        {
            JobId = "job-1",
            Name = "File Backup",
            JobType = JobType.Full,
            SourceId = sourceFile,
            SourceType = "File",
            DestinationId = "repo-1",
            Enabled = true
        });

        db.JobRunHistory.Add(new JobRunHistory
        {
            RunId = "run-1",
            JobId = "job-1",
            StartTime = DateTime.UtcNow,
            Status = "queued"
        });

        await db.SaveChangesAsync();

        try
        {
            var result = await service.ExecuteRunAsync("run-1");

            Assert.True(result.Success);
            Assert.False(string.IsNullOrWhiteSpace(result.BackupId));

            var backupPoint = await db.BackupPoints.FirstOrDefaultAsync(b => b.BackupId == result.BackupId);
            Assert.NotNull(backupPoint);
            Assert.Equal(BackupStatus.Completed, backupPoint!.Status);
            Assert.True(File.Exists(backupPoint.FilePath));
            Assert.True(backupPoint.SizeBytes > 0);

            var runHistory = await db.JobRunHistory.FirstAsync(r => r.RunId == "run-1");
            Assert.Equal("completed", runHistory.Status);
            Assert.True(runHistory.BytesProcessed > 0);
            Assert.True(runHistory.FilesProcessed >= 1);
        }
        finally
        {
            if (Directory.Exists(tempRoot))
            {
                Directory.Delete(tempRoot, true);
            }
        }
    }
}
