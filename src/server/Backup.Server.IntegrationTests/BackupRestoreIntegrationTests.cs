using System.Net;
using System.Net.Http.Json;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class JobApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public JobApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
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

        var createResponse = await _client.PostAsJsonAsync("/api/jobs", new
        {
            name = "Runnable Job",
            jobType = "Full",
            sourceId = "vm-002",
            sourceType = "VirtualMachine",
            destinationId = "repo-001",
            enabled = true
        });

        var createdBody = await createResponse.Content.ReadAsStringAsync();
        var jobId = ExtractValue(createdBody, "jobId");

        var runResponse = await _client.PostAsync($"/api/jobs/{jobId}/run", null);
        Assert.Equal(HttpStatusCode.OK, runResponse.StatusCode);
    }

    private static string ExtractValue(string json, string propertyName)
    {
        using var document = System.Text.Json.JsonDocument.Parse(json);
        return document.RootElement.GetProperty(propertyName).GetString()
            ?? throw new InvalidOperationException($"{propertyName} was not present");
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
