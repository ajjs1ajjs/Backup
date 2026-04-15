using Backup.Server.Services;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.Configuration;
using Xunit;

namespace Backup.Server.Tests;

public class AuthLockoutServiceTests
{
    [Fact]
    public void RegisterFailure_ShouldLockAccount_WhenThresholdIsReached()
    {
        var timeProvider = new ManualTimeProvider(new DateTimeOffset(2026, 4, 15, 10, 0, 0, TimeSpan.Zero));
        var service = CreateService(timeProvider);

        Assert.False(service.RegisterFailure("admin").IsLockedOut);
        Assert.False(service.RegisterFailure("admin").IsLockedOut);

        var lockedStatus = service.RegisterFailure("admin");

        Assert.True(lockedStatus.IsLockedOut);
        Assert.Equal(3, lockedStatus.FailedAttempts);
        Assert.Equal(0, lockedStatus.RemainingAttempts);
        Assert.Equal(timeProvider.GetUtcNow().AddMinutes(15), lockedStatus.LockedUntil);
    }

    [Fact]
    public void GetStatus_ShouldResetExpiredLockout()
    {
        var timeProvider = new ManualTimeProvider(new DateTimeOffset(2026, 4, 15, 10, 0, 0, TimeSpan.Zero));
        var service = CreateService(timeProvider);

        service.RegisterFailure("admin");
        service.RegisterFailure("admin");
        service.RegisterFailure("admin");

        timeProvider.Advance(TimeSpan.FromMinutes(16));

        var status = service.GetStatus("admin");

        Assert.False(status.IsLockedOut);
        Assert.Equal(0, status.FailedAttempts);
        Assert.Equal(3, status.RemainingAttempts);
        Assert.Null(status.LockedUntil);
    }

    [Fact]
    public void Reset_ShouldClearRecordedFailures()
    {
        var timeProvider = new ManualTimeProvider(new DateTimeOffset(2026, 4, 15, 10, 0, 0, TimeSpan.Zero));
        var service = CreateService(timeProvider);

        service.RegisterFailure("admin");
        service.RegisterFailure("admin");
        service.Reset("admin");

        var status = service.GetStatus("admin");

        Assert.False(status.IsLockedOut);
        Assert.Equal(0, status.FailedAttempts);
        Assert.Equal(3, status.RemainingAttempts);
    }

    [Fact]
    public void RegisterFailure_ShouldDropStaleFailures_OutsideFailureWindow()
    {
        var timeProvider = new ManualTimeProvider(new DateTimeOffset(2026, 4, 15, 10, 0, 0, TimeSpan.Zero));
        var service = CreateService(timeProvider);

        service.RegisterFailure("admin");
        timeProvider.Advance(TimeSpan.FromMinutes(6));

        var status = service.RegisterFailure("admin");

        Assert.False(status.IsLockedOut);
        Assert.Equal(1, status.FailedAttempts);
        Assert.Equal(2, status.RemainingAttempts);
    }

    private static AuthLockoutService CreateService(TimeProvider timeProvider)
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["Auth:Lockout:MaxFailedAttempts"] = "3",
                ["Auth:Lockout:DurationMinutes"] = "15",
                ["Auth:Lockout:FailureWindowMinutes"] = "5"
            })
            .Build();

        return new AuthLockoutService(new MemoryCache(new MemoryCacheOptions()), configuration, timeProvider);
    }

    private sealed class ManualTimeProvider : TimeProvider
    {
        private DateTimeOffset _utcNow;

        public ManualTimeProvider(DateTimeOffset utcNow)
        {
            _utcNow = utcNow;
        }

        public override DateTimeOffset GetUtcNow()
        {
            return _utcNow;
        }

        public void Advance(TimeSpan duration)
        {
            _utcNow = _utcNow.Add(duration);
        }
    }
}
