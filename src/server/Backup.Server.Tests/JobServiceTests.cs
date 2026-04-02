using Xunit;
using Moq;
using Microsoft.Extensions.Logging;
using Backup.Server.Services;
using Backup.Server.Database;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Tests;

public class JobServiceTests : IDisposable
{
    private readonly BackupDbContext _db;
    private readonly JobServiceImpl _service;

    public JobServiceTests()
    {
        var options = new DbContextOptionsBuilder<BackupDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;
        
        _db = new BackupDbContext(options);
        var logger = new Mock<ILogger<JobServiceImpl>>();
        _service = new JobServiceImpl(logger.Object, _db);
    }

    public void Dispose()
    {
        _db.Dispose();
    }

    [Fact]
    public async Task CreateJob_ShouldCreateJob()
    {
        var request = new Backup.Contracts.JobRequest
        {
            Name = "Test Job",
            JobType = Backup.Contracts.JobType.JobTypeFullBackup,
            SourceId = "vm-1",
            DestinationId = "repo-1",
            Enabled = true
        };

        var result = await _service.CreateJob(request, Grpc.Core.ServerCallContext.Null);

        Assert.True(result.Success);
        Assert.NotEmpty(result.JobId);
    }

    [Fact]
    public async Task ListJobs_ShouldReturnJobs()
    {
        var request = new Backup.Contracts.JobListRequest
        {
            Page = 1,
            PageSize = 10
        };

        var result = await _service.ListJobs(request, Grpc.Core.ServerCallContext.Null);

        Assert.NotNull(result);
    }
}
