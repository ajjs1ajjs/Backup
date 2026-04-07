using System.Net;
using System.Net.Http.Json;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class AgentApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public AgentApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task UnauthorizedRequest_ShouldBeRejected()
    {
        var response = await _client.GetAsync("/api/agents");

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task AuthorizedRequest_ShouldReturnOk()
    {
        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "agents");

        var response = await _client.GetAsync("/api/agents");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }
}

public class ReportsApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public ReportsApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task SummaryEndpoint_ShouldRequireAuth_AndReturnStats()
    {
        var unauthorized = await _client.GetAsync("/api/reports/summary");
        Assert.Equal(HttpStatusCode.Unauthorized, unauthorized.StatusCode);

        await TestAuthHelper.RegisterAndAuthenticateAsync(_client, "reports");
        var authorized = await _client.GetAsync("/api/reports/summary");

        Assert.Equal(HttpStatusCode.OK, authorized.StatusCode);

        var body = await authorized.Content.ReadAsStringAsync();
        Assert.Contains("totalJobs", body);
    }
}
