using System.Net;
using System.Net.Http.Json;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class AuthApiIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly HttpClient _client;

    public AuthApiIntegrationTests(IntegrationTestWebApplicationFactory factory)
    {
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task Login_ShouldReturnTooManyRequests_WhenLockoutThresholdIsReached()
    {
        var username = $"lockout-{Guid.NewGuid():N}";
        await RegisterUserAsync(username, "StrongPass123!");

        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);
        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);

        var lockedResponse = await _client.PostAsJsonAsync("/api/auth/login", new
        {
            username,
            password = "WrongPass123!"
        });

        Assert.Equal(HttpStatusCode.TooManyRequests, lockedResponse.StatusCode);
        Assert.True(lockedResponse.Headers.TryGetValues("Retry-After", out var values));
        Assert.False(string.IsNullOrWhiteSpace(values.SingleOrDefault()));

        var payload = await lockedResponse.Content.ReadFromJsonAsync<AuthErrorResponse>();
        Assert.Equal("Too many failed login attempts. Try again later.", payload?.Error);
        Assert.True(payload?.RetryAfterSeconds > 0);
    }

    [Fact]
    public async Task SuccessfulPasswordValidation_ShouldResetFailureCounter()
    {
        var username = $"reset-{Guid.NewGuid():N}";
        await RegisterUserAsync(username, "StrongPass123!");

        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);
        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);
        await AssertLoginStatusAsync(username, "StrongPass123!", HttpStatusCode.OK);
        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);
        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.Unauthorized);
        await AssertLoginStatusAsync(username, "WrongPass123!", HttpStatusCode.TooManyRequests);
    }

    private async Task RegisterUserAsync(string username, string password)
    {
        var response = await _client.PostAsJsonAsync("/api/auth/register", new
        {
            username,
            email = $"{username}@example.com",
            password
        });

        response.EnsureSuccessStatusCode();
    }

    private async Task AssertLoginStatusAsync(string username, string password, HttpStatusCode expectedStatusCode)
    {
        var response = await _client.PostAsJsonAsync("/api/auth/login", new
        {
            username,
            password
        });

        Assert.Equal(expectedStatusCode, response.StatusCode);
    }

    private sealed class AuthErrorResponse
    {
        public string? Error { get; set; }
        public int RetryAfterSeconds { get; set; }
    }
}
