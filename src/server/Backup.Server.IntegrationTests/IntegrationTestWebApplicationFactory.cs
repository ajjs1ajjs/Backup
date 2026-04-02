using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Grpc.Net.Client;
using Backup.Server.Database;
using Testcontainers.PostgreSql;

namespace Backup.Server.IntegrationTests;

public class IntegrationTestWebApplicationFactory : WebApplicationFactory<Program>, IAsyncLifetime
{
    private readonly PostgreSqlContainer _postgresContainer;

    public IntegrationTestWebApplicationFactory()
    {
        _postgresContainer = new PostgreSqlBuilder()
            .WithImage("postgres:14")
            .WithDatabase("backup_test")
            .WithUsername("postgres")
            .WithPassword("postgres")
            .Build();
    }

    protected override IHost CreateHost(IHostBuilder builder)
    {
        builder.ConfigureServices(services =>
        {
            var descriptor = services.SingleOrDefault(
                d => d.ServiceType == typeof(DbContextOptions<BackupDbContext>));

            if (descriptor != null)
            {
                services.Remove(descriptor);
            }

            services.AddDbContext<BackupDbContext>((sp, options) =>
            {
                var connectionString = _postgresContainer.GetConnectionString();
                options.UseNpgsql(connectionString);
            });
        });

        return base.CreateHost(builder);
    }

    public async Task InitializeAsync()
    {
        await _postgresContainer.StartAsync();
    }

    public new async Task DisposeAsync()
    {
        await _postgresContainer.StopAsync();
    }
}

public class GrpcIntegrationTestFixture : IAsyncLifetime
{
    private readonly HttpClient _httpClient;
    private readonly HttpMessageHandler _handler;

    public GrpcIntegrationTestFixture()
    {
        var factory = new IntegrationTestWebApplicationFactory();
        _httpClient = factory.CreateClient();
        _handler = factory.Server.CreateHandler();
    }

    public GrpcChannel CreateGrpcChannel()
    {
        var serverUrl = Environment.GetEnvironmentVariable("TEST_ServerUrl")
            ?? _httpClient.BaseAddress?.ToString()
            ?? "http://localhost:8000";

        return GrpcChannel.ForAddress(serverUrl, new GrpcChannelOptions
        {
            HttpHandler = _handler
        });
    }

    public HttpClient CreateHttpClient()
    {
        return _httpClient;
    }

    public async Task InitializeAsync()
    {
        await Task.CompletedTask;
    }

    public async Task DisposeAsync()
    {
        _httpClient.Dispose();
        await Task.CompletedTask;
    }
}
