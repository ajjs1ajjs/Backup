using System.Net.Http.Headers;
using System.Net.Http.Json;
using Backup.Server.Database;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class IntegrationTestWebApplicationFactory : WebApplicationFactory<Program>, IAsyncLifetime
{
    private SqliteConnection? _connection;

    protected override void ConfigureWebHost(Microsoft.AspNetCore.Hosting.IWebHostBuilder builder)
    {
        builder.ConfigureServices(services =>
        {
            _connection = new SqliteConnection("Data Source=:memory:");
            _connection.Open();

            services.RemoveAll(typeof(DbContextOptions<BackupDbContext>));
            services.AddDbContext<BackupDbContext>(options => options.UseSqlite(_connection));
        });
    }

    public async Task InitializeAsync()
    {
        using var scope = Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
        await db.Database.EnsureCreatedAsync();
    }

    public new async Task DisposeAsync()
    {
        if (_connection != null)
        {
            await _connection.DisposeAsync();
        }
    }
}

public static class TestAuthHelper
{
    public static async Task<string> RegisterAndAuthenticateAsync(HttpClient client, string usernameSuffix)
    {
        var request = new
        {
            username = $"tester-{usernameSuffix}",
            email = $"tester-{usernameSuffix}@example.com",
            password = "StrongPass123!",
            role = "Admin"
        };

        var response = await client.PostAsJsonAsync("/api/auth/register", request);
        response.EnsureSuccessStatusCode();

        var payload = await response.Content.ReadFromJsonAsync<AuthTokenResponse>();
        if (string.IsNullOrWhiteSpace(payload?.Token))
        {
            throw new InvalidOperationException("Auth token was not returned");
        }

        client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", payload.Token);
        return payload.Token;
    }
}

public class AuthTokenResponse
{
    public string? Token { get; set; }
}
