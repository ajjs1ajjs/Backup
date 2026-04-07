using System.Net;
using System.Net.Http.Json;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.Extensions.DependencyInjection;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class JobApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;
    private readonly IntegrationTestWebApplicationFactory _factory;

    public JobApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _factory = factory;
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task CreateJob_AndListJobs_ShouldSucceed()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "jobs");

        var createResponse = await _client.PostAsJsonAsync("/api/jobs", new
        {
            name = "Integration Job",
            jobType = "Full",
            sourceId = "vm-001",
            sourceType = "VirtualMachine",
            destinationId = "repo-001",
            schedule = "0 2 * * *",
            enabled = true
        });

        Assert.Equal(HttpStatusCode.Created, createResponse.StatusCode);

        var listResponse = await _client.GetAsync("/api/jobs");
        Assert.Equal(HttpStatusCode.OK, listResponse.StatusCode);

        var body = await listResponse.Content.ReadAsStringAsync();
        Assert.Contains("Integration Job", body);
    }

    [Fact]
    public async Task RunJob_ForExistingJob_ShouldSucceed()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "run-job");

        var tempRoot = Path.Combine(Path.GetTempPath(), "backup-integration-tests", Guid.NewGuid().ToString("N"));
        var sourceFile = Path.Combine(tempRoot, "payload.txt");
        var repositoryPath = Path.Combine(tempRoot, "repo");
        Directory.CreateDirectory(tempRoot);
        await File.WriteAllTextAsync(sourceFile, "integration backup payload");

        try
        {
            var repositoryResponse = await _client.PostAsJsonAsync("/api/repositories", new
            {
                name = "Runnable Repo",
                type = "Local",
                path = repositoryPath,
                status = "online"
            });
            Assert.Equal(HttpStatusCode.Created, repositoryResponse.StatusCode);

            var repositoryBody = await repositoryResponse.Content.ReadAsStringAsync();
            var repositoryId = ExtractValue(repositoryBody, "repositoryId");

            var createResponse = await _client.PostAsJsonAsync("/api/jobs", new
            {
                name = "Runnable Job",
                jobType = "Full",
                sourceId = sourceFile,
                sourceType = "File",
                destinationId = repositoryId,
                enabled = true
            });

            var createdBody = await createResponse.Content.ReadAsStringAsync();
            var jobId = ExtractValue(createdBody, "jobId");

            var runResponse = await _client.PostAsync($"/api/jobs/{jobId}/run", null);
            Assert.Equal(HttpStatusCode.OK, runResponse.StatusCode);

            await WaitForConditionAsync(async () =>
            {
                var backupsResponse = await _client.GetAsync($"/api/backups?jobId={jobId}");
                if (!backupsResponse.IsSuccessStatusCode)
                {
                    return false;
                }

                var content = await backupsResponse.Content.ReadAsStringAsync();
                return content.Contains("Completed", StringComparison.OrdinalIgnoreCase)
                    || content.Contains("completed", StringComparison.OrdinalIgnoreCase);
            });

            var backupsListResponse = await _client.GetAsync($"/api/backups?jobId={jobId}");
            var backupsBody = await backupsListResponse.Content.ReadAsStringAsync();
            Assert.Contains("completed", backupsBody, StringComparison.OrdinalIgnoreCase);

            var runsResponse = await _client.GetAsync($"/api/jobs/{jobId}/runs");
            Assert.Equal(HttpStatusCode.OK, runsResponse.StatusCode);
            var runsBody = await runsResponse.Content.ReadAsStringAsync();
            Assert.Contains("completed", runsBody, StringComparison.OrdinalIgnoreCase);
            Assert.True(Directory.Exists(repositoryPath));
            Assert.NotEmpty(Directory.GetFileSystemEntries(repositoryPath));
        }
        finally
        {
            if (Directory.Exists(tempRoot))
            {
                Directory.Delete(tempRoot, true);
            }
        }
    }

    [Fact]
    public async Task StopJob_ForQueuedRun_ShouldMarkRunAsCancelled()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "stop-job");

        const string jobId = "queued-job";
        const string runId = "queued-run";

        await using (var scope = _factory.Services.CreateAsyncScope())
        {
            var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

            db.Jobs.Add(new Job
            {
                JobId = jobId,
                Name = "Queued Job",
                JobType = JobType.Full,
                SourceId = "C:\\missing-source",
                SourceType = "File",
                DestinationId = "repo-missing",
                Enabled = true,
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow
            });

            db.JobRunHistory.Add(new JobRunHistory
            {
                RunId = runId,
                JobId = jobId,
                StartTime = DateTime.UtcNow,
                Status = "queued"
            });

            await db.SaveChangesAsync();
        }

        var stopResponse = await _client.PostAsync($"/api/jobs/{jobId}/stop", null);
        Assert.Equal(HttpStatusCode.OK, stopResponse.StatusCode);

        var stopBody = await stopResponse.Content.ReadAsStringAsync();
        Assert.Contains("cancelled", stopBody, StringComparison.OrdinalIgnoreCase);
        Assert.Contains(runId, stopBody, StringComparison.OrdinalIgnoreCase);

        var runResponse = await _client.GetAsync($"/api/jobs/runs/{runId}");
        Assert.Equal(HttpStatusCode.OK, runResponse.StatusCode);

        var runBody = await runResponse.Content.ReadAsStringAsync();
        Assert.Contains("cancelled", runBody, StringComparison.OrdinalIgnoreCase);
    }

    private static string ExtractValue(string json, string propertyName)
    {
        using var document = System.Text.Json.JsonDocument.Parse(json);
        return document.RootElement.GetProperty(propertyName).GetString()
            ?? throw new InvalidOperationException($"{propertyName} was not present");
    }

    private static async Task WaitForConditionAsync(Func<Task<bool>> condition, int maxAttempts = 20, int delayMs = 150)
    {
        for (var attempt = 0; attempt < maxAttempts; attempt++)
        {
            if (await condition())
            {
                return;
            }

            await Task.Delay(delayMs);
        }

        throw new TimeoutException("Timed out waiting for background backup completion.");
    }
}

public class RepositoryApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public RepositoryApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task CreateRepository_AndListRepositories_ShouldSucceed()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "repos");

        var createResponse = await _client.PostAsJsonAsync("/api/repositories", new
        {
            name = "Local Repo",
            type = "Local",
            path = ".",
            status = "online"
        });

        Assert.Equal(HttpStatusCode.Created, createResponse.StatusCode);

        var listResponse = await _client.GetAsync("/api/repositories");
        Assert.Equal(HttpStatusCode.OK, listResponse.StatusCode);

        var body = await listResponse.Content.ReadAsStringAsync();
        Assert.Contains("Local Repo", body);
    }
}

public class RestoreApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public RestoreApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task StartRestore_ForMissingBackup_ShouldReturnNotFound()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "restore");

        var response = await _client.PostAsJsonAsync("/api/restore", new
        {
            backupId = "missing-backup",
            restoreType = "full_vm",
            destinationPath = "C:\\Restore",
            targetHost = "localhost"
        });

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }
}
