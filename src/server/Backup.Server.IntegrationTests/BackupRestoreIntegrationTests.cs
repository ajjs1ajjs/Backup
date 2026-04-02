using Grpc.Net.Client;
using Backup.Contracts;
using Xunit;
using Microsoft.AspNetCore.Mvc.Testing;

namespace Backup.Server.IntegrationTests;

public class BackupFlowIntegrationTests : IClassFixture<GrpcIntegrationTestFixture>, IAsyncLifetime
{
    private readonly GrpcChannel _channel;
    private readonly JobService.JobServiceClient _jobClient;
    private readonly BackupService.BackupServiceClient _backupClient;

    public BackupFlowIntegrationTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _jobClient = new JobService.JobServiceClient(_channel);
        _backupClient = new BackupService.BackupServiceClient(_channel);
    }

    public Task InitializeAsync() => Task.CompletedTask;
    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task CreateBackupJob_ShouldSucceed()
    {
        // Arrange
        var jobRequest = new JobRequest
        {
            Name = "Integration Test Backup Job",
            JobType = JobType.JobTypeFullBackup,
            SourceId = "vm-test-001",
            DestinationId = "repo-test-001",
            Enabled = true,
            Schedule = "0 2 * * *", // Daily at 2 AM
            CompressionEnabled = true,
            DeduplicationEnabled = true
        };

        // Act
        var createResponse = await _jobClient.CreateJob(jobRequest);

        // Assert
        Assert.True(createResponse.Success);
        Assert.NotEmpty(createResponse.JobId);
        Assert.Equal(jobRequest.Name, createResponse.Job.Name);
    }

    [Fact]
    public async Task CreateAndRunBackupJob_FullFlow()
    {
        // Arrange - Create job
        var jobRequest = new JobRequest
        {
            Name = "Full Flow Test Job",
            JobType = JobType.JobTypeFullBackup,
            SourceId = "vm-integration-001",
            DestinationId = "repo-001",
            Enabled = true
        };

        var createResponse = await _jobClient.CreateJob(jobRequest);
        Assert.True(createResponse.Success);

        // Act - Run job immediately
        var runRequest = new RunJobRequest
        {
            JobId = createResponse.JobId
        };

        var runResponse = await _jobClient.RunJob(runRequest);

        // Assert
        Assert.True(runResponse.Success);
    }

    [Fact]
    public async Task CreateIncrementalBackupJob_ShouldSucceed()
    {
        // Arrange
        var jobRequest = new JobRequest
        {
            Name = "Incremental Backup Test",
            JobType = JobType.JobTypeIncrementalBackup,
            SourceId = "vm-incremental-001",
            DestinationId = "repo-001",
            Enabled = true,
            IncrementalBasePath = "/backups/base-001"
        };

        // Act
        var response = await _jobClient.CreateJob(jobRequest);

        // Assert
        Assert.True(response.Success);
    }

    [Fact]
    public async Task ListJobs_ShouldReturnCreatedJobs()
    {
        // Arrange - Create multiple jobs
        var jobsToCreate = new[]
        {
            new JobRequest { Name = "List Test Job 1", JobType = JobType.JobTypeFullBackup, SourceId = "vm-1", DestinationId = "repo-1", Enabled = true },
            new JobRequest { Name = "List Test Job 2", JobType = JobType.JobTypeIncrementalBackup, SourceId = "vm-2", DestinationId = "repo-1", Enabled = true },
            new JobRequest { Name = "List Test Job 3", JobType = JobType.JobTypeDifferentialBackup, SourceId = "vm-3", DestinationId = "repo-2", Enabled = false }
        };

        foreach (var job in jobsToCreate)
        {
            await _jobClient.CreateJob(job);
        }

        // Act - List jobs
        var listRequest = new JobListRequest
        {
            Page = 1,
            PageSize = 10
        };

        var listResponse = await _jobClient.ListJobs(listRequest);

        // Assert
        Assert.NotNull(listResponse);
        Assert.True(listResponse.Total > 0);
        Assert.NotEmpty(listResponse.Jobs);
    }

    [Fact]
    public async Task StopRunningJob_ShouldSucceed()
    {
        // Arrange - Create and start job
        var jobRequest = new JobRequest
        {
            Name = "Stop Test Job",
            JobType = JobType.JobTypeFullBackup,
            SourceId = "vm-stop-test",
            DestinationId = "repo-1",
            Enabled = true
        };

        var createResponse = await _jobClient.CreateJob(jobRequest);
        Assert.True(createResponse.Success);

        var runRequest = new RunJobRequest { JobId = createResponse.JobId };
        await _jobClient.RunJob(runRequest);

        // Act - Stop job
        var stopRequest = new StopJobRequest
        {
            JobId = createResponse.JobId,
            Force = false
        };

        var stopResponse = await _jobClient.StopJob(stopRequest);

        // Assert
        Assert.True(stopResponse.Success);
    }

    [Fact]
    public async Task DeleteJob_ShouldSucceed()
    {
        // Arrange - Create job
        var jobRequest = new JobRequest
        {
            Name = "Delete Test Job",
            JobType = JobType.JobTypeFullBackup,
            SourceId = "vm-delete-test",
            DestinationId = "repo-1",
            Enabled = true
        };

        var createResponse = await _jobClient.CreateJob(jobRequest);
        Assert.True(createResponse.Success);

        // Act - Delete job
        var deleteRequest = new DeleteJobRequest
        {
            JobId = createResponse.JobId
        };

        var deleteResponse = await _jobClient.DeleteJob(deleteRequest);

        // Assert
        Assert.True(deleteResponse.Success);
    }
}

public class RestoreFlowIntegrationTests : IClassFixture<GrpcIntegrationTestFixture>
{
    private readonly GrpcChannel _channel;
    private readonly RestoreService.RestoreServiceClient _restoreClient;

    public RestoreFlowIntegrationTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _restoreClient = new RestoreService.RestoreServiceClient(_channel);
    }

    [Fact]
    public async Task CreateRestoreJob_ShouldSucceed()
    {
        // Arrange
        var restoreRequest = new RestoreRequest
        {
            BackupId = "backup-test-001",
            RestoreType = RestoreType.RestoreTypeFullVm,
            TargetHost = "hyperv-host-001",
            DestinationPath = "C:\\VMs\\Restored",
            PointInTime = DateTime.UtcNow.AddHours(-1).ToString("O")
        };

        // Act
        var response = await _restoreClient.CreateRestore(restoreRequest);

        // Assert
        Assert.True(response.Success);
        Assert.NotEmpty(response.RestoreId);
    }

    [Fact]
    public async Task CreateFileLevelRestore_ShouldSucceed()
    {
        // Arrange
        var restoreRequest = new RestoreRequest
        {
            BackupId = "backup-flr-001",
            RestoreType = RestoreType.RestoreTypeFileLevel,
            TargetHost = "hyperv-host-001",
            DestinationPath = "C:\\RestoredFiles",
            FilesToRestore = { "C:\\Data\\file1.txt", "C:\\Data\\file2.docx" }
        };

        // Act
        var response = await _restoreClient.CreateRestore(restoreRequest);

        // Assert
        Assert.True(response.Success);
    }

    [Fact]
    public async Task GetRestoreStatus_ShouldReturnProgress()
    {
        // Arrange
        var statusRequest = new RestoreStatusRequest
        {
            RestoreId = "restore-status-test"
        };

        // Act
        var response = await _restoreClient.GetRestoreStatus(statusRequest);

        // Assert
        Assert.NotNull(response);
        Assert.NotNull(response.Status);
    }

    [Fact]
    public async Task CancelRestore_ShouldSucceed()
    {
        // Arrange
        var cancelRequest = new CancelRestoreRequest
        {
            RestoreId = "restore-cancel-test"
        };

        // Act
        var response = await _restoreClient.CancelRestore(cancelRequest);

        // Assert
        Assert.True(response.Success);
    }
}

public class RepositoryIntegrationTests : IClassFixture<GrpcIntegrationTestFixture>
{
    private readonly GrpcChannel _channel;
    private readonly RepositoryService.RepositoryServiceClient _repositoryClient;

    public RepositoryIntegrationTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _repositoryClient = new RepositoryService.RepositoryServiceClient(_channel);
    }

    [Fact]
    public async Task CreateRepository_ShouldSucceed()
    {
        // Arrange
        var repoRequest = new RepositoryRequest
        {
            Name = "Test Repository",
            Type = RepositoryType.RepositoryTypeLocal,
            Path = "D:\\Backups",
            MaxCapacityBytes = 1099511627776, // 1TB
            WarningThresholdPercent = 80,
            CriticalThresholdPercent = 95
        };

        // Act
        var response = await _repositoryClient.CreateRepository(repoRequest);

        // Assert
        Assert.True(response.Success);
        Assert.NotEmpty(response.RepositoryId);
    }

    [Fact]
    public async Task TestRepositoryConnection_LocalPath_ShouldSucceed()
    {
        // Arrange
        var testRequest = new TestRepositoryRequest
        {
            RepositoryId = "repo-test-001"
        };

        // Act
        var response = await _repositoryClient.TestRepository(testRequest);

        // Assert
        Assert.True(response.Success);
        Assert.True(response.IsAccessible);
    }

    [Fact]
    public async Task ListRepositories_ShouldReturnRepositories()
    {
        // Act
        var response = await _repositoryClient.ListRepositories(new RepositoryListRequest());

        // Assert
        Assert.NotNull(response);
        Assert.NotNull(response.Repositories);
    }
}
