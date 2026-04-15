using Microsoft.Extensions.Caching.Memory;

namespace Backup.Server.Services;

public interface IAuthLockoutService
{
    AuthLockoutStatus GetStatus(string username);
    AuthLockoutStatus RegisterFailure(string username);
    void Reset(string username);
}

public sealed record AuthLockoutStatus(
    bool IsLockedOut,
    int FailedAttempts,
    int RemainingAttempts,
    DateTimeOffset? LockedUntil);

public sealed class AuthLockoutException : InvalidOperationException
{
    public AuthLockoutException(DateTimeOffset lockedUntil)
        : base("Too many failed login attempts. Try again later.")
    {
        LockedUntil = lockedUntil;
    }

    public DateTimeOffset LockedUntil { get; }
}

public class AuthLockoutService : IAuthLockoutService
{
    private readonly IMemoryCache _cache;
    private readonly TimeProvider _timeProvider;
    private readonly object _sync = new();
    private readonly int _maxFailedAttempts;
    private readonly TimeSpan _lockoutDuration;
    private readonly TimeSpan _failureWindow;

    public AuthLockoutService(IMemoryCache cache, IConfiguration configuration, TimeProvider timeProvider)
    {
        _cache = cache;
        _timeProvider = timeProvider;
        _maxFailedAttempts = Math.Max(1, configuration.GetValue<int?>("Auth:Lockout:MaxFailedAttempts") ?? 5);
        _lockoutDuration = TimeSpan.FromMinutes(Math.Max(1, configuration.GetValue<int?>("Auth:Lockout:DurationMinutes") ?? 15));
        _failureWindow = TimeSpan.FromMinutes(Math.Max(1, configuration.GetValue<int?>("Auth:Lockout:FailureWindowMinutes") ?? 15));
    }

    public AuthLockoutStatus GetStatus(string username)
    {
        var cacheKey = NormalizeUsername(username);
        if (string.IsNullOrEmpty(cacheKey))
        {
            return EmptyStatus();
        }

        lock (_sync)
        {
            var now = _timeProvider.GetUtcNow();
            var state = GetActiveState(cacheKey, now);
            return BuildStatus(state);
        }
    }

    public AuthLockoutStatus RegisterFailure(string username)
    {
        var cacheKey = NormalizeUsername(username);
        if (string.IsNullOrEmpty(cacheKey))
        {
            return EmptyStatus();
        }

        lock (_sync)
        {
            var now = _timeProvider.GetUtcNow();
            var state = GetActiveState(cacheKey, now) ?? new AuthLockoutState();

            state.FailedAttempts++;
            state.LastFailedAt = now;

            if (state.FailedAttempts >= _maxFailedAttempts)
            {
                state.LockedUntil = now.Add(_lockoutDuration);
            }

            _cache.Set(cacheKey, state, new MemoryCacheEntryOptions
            {
                AbsoluteExpiration = state.LockedUntil ?? now.Add(_failureWindow)
            });

            return BuildStatus(state);
        }
    }

    public void Reset(string username)
    {
        var cacheKey = NormalizeUsername(username);
        if (string.IsNullOrEmpty(cacheKey))
        {
            return;
        }

        lock (_sync)
        {
            _cache.Remove(cacheKey);
        }
    }

    private AuthLockoutState? GetActiveState(string cacheKey, DateTimeOffset now)
    {
        if (!_cache.TryGetValue<AuthLockoutState>(cacheKey, out var state) || state == null)
        {
            return null;
        }

        if (state.LockedUntil.HasValue)
        {
            if (state.LockedUntil.Value <= now)
            {
                _cache.Remove(cacheKey);
                return null;
            }

            return state;
        }

        if (!state.LastFailedAt.HasValue || now - state.LastFailedAt.Value > _failureWindow)
        {
            _cache.Remove(cacheKey);
            return null;
        }

        return state;
    }

    private AuthLockoutStatus BuildStatus(AuthLockoutState? state)
    {
        if (state == null)
        {
            return EmptyStatus();
        }

        if (state.LockedUntil.HasValue)
        {
            return new AuthLockoutStatus(true, state.FailedAttempts, 0, state.LockedUntil);
        }

        return new AuthLockoutStatus(
            false,
            state.FailedAttempts,
            Math.Max(0, _maxFailedAttempts - state.FailedAttempts),
            null);
    }

    private AuthLockoutStatus EmptyStatus()
    {
        return new AuthLockoutStatus(false, 0, _maxFailedAttempts, null);
    }

    private static string NormalizeUsername(string username)
    {
        return string.IsNullOrWhiteSpace(username)
            ? string.Empty
            : username.Trim().ToUpperInvariant();
    }

    private sealed class AuthLockoutState
    {
        public int FailedAttempts { get; set; }
        public DateTimeOffset? LastFailedAt { get; set; }
        public DateTimeOffset? LockedUntil { get; set; }
    }
}
