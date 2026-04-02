using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Security.Cryptography;
using System.Text;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using Backup.Server.Database;
using Backup.Server.Database.Entities;

namespace Backup.Server.Services;

public interface IAuthService
{
    Task<string> RegisterAsync(string username, string email, string password, string role = "Viewer");
    Task<string> LoginAsync(string username, string password);
    Task<User?> GetUserByIdAsync(string userId);
    Task<User?> GetUserByUsernameAsync(string username);
    Task<bool> ValidatePasswordAsync(string userId, string password);
    Task UpdateLastLoginAsync(string userId);
    Task<string> GeneratePasswordResetTokenAsync(string email);
    Task<bool> ResetPasswordAsync(string token, string newPassword);
}

public class AuthService : IAuthService
{
    private readonly BackupDbContext _context;
    private readonly IConfiguration _configuration;

    public AuthService(BackupDbContext context, IConfiguration configuration)
    {
        _context = context;
        _configuration = configuration;
    }

    public async Task<string> RegisterAsync(string username, string email, string password, string role = "Viewer")
    {
        if (await _context.Users.AnyAsync(u => u.Username == username))
            throw new InvalidOperationException("Username already exists");

        if (await _context.Users.AnyAsync(u => u.Email == email))
            throw new InvalidOperationException("Email already exists");

        var userId = Guid.NewGuid().ToString();
        var passwordHash = HashPassword(password);

        var user = new User
        {
            UserId = userId,
            Username = username,
            Email = email,
            PasswordHash = passwordHash,
            Role = role,
            IsActive = true,
            CreatedAt = DateTime.UtcNow
        };

        _context.Users.Add(user);
        await _context.SaveChangesAsync();

        return GenerateJwtToken(user);
    }

    public async Task<string> LoginAsync(string username, string password)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.Username == username);
        
        if (user == null || !user.IsActive)
            throw new UnauthorizedAccessException("Invalid credentials");

        if (!VerifyPassword(password, user.PasswordHash))
            throw new UnauthorizedAccessException("Invalid credentials");

        user.LastLoginAt = DateTime.UtcNow;
        await _context.SaveChangesAsync();

        return GenerateJwtToken(user);
    }

    public async Task<User?> GetUserByIdAsync(string userId)
    {
        return await _context.Users.FirstOrDefaultAsync(u => u.UserId == userId);
    }

    public async Task<User?> GetUserByUsernameAsync(string username)
    {
        return await _context.Users.FirstOrDefaultAsync(u => u.Username == username);
    }

    public async Task<bool> ValidatePasswordAsync(string userId, string password)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.UserId == userId);
        return user != null && VerifyPassword(password, user.PasswordHash);
    }

    public async Task UpdateLastLoginAsync(string userId)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.UserId == userId);
        if (user != null)
        {
            user.LastLoginAt = DateTime.UtcNow;
            await _context.SaveChangesAsync();
        }
    }

    public Task<string> GeneratePasswordResetTokenAsync(string email)
    {
        var token = Convert.ToBase64String(RandomNumberGenerator.GetBytes(32));
        return Task.FromResult(token);
    }

    public async Task<bool> ResetPasswordAsync(string token, string newPassword)
    {
        return await Task.FromResult(true);
    }

    private string GenerateJwtToken(User user)
    {
        var key = new SymmetricSecurityKey(
            Encoding.UTF8.GetBytes(_configuration["Jwt:Key"] ?? "BackupServerSecretKey2024!@#$%^&*()"));
        var credentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, user.UserId),
            new Claim(ClaimTypes.Name, user.Username),
            new Claim(ClaimTypes.Email, user.Email),
            new Claim(ClaimTypes.Role, user.Role),
            new Claim(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
        };

        var token = new JwtSecurityToken(
            issuer: _configuration["Jwt:Issuer"] ?? "BackupServer",
            audience: _configuration["Jwt:Audience"] ?? "BackupClients",
            claims: claims,
            expires: DateTime.UtcNow.AddHours(24),
            signingCredentials: credentials
        );

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    private static string HashPassword(string password)
    {
        using var sha256 = SHA256.Create();
        var hash = sha256.ComputeHash(Encoding.UTF8.GetBytes(password));
        return Convert.ToBase64String(hash);
    }

    private static bool VerifyPassword(string password, string hash)
    {
        return HashPassword(password) == hash;
    }
}
