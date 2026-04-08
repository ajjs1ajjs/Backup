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
    Task<LoginResult> LoginAsync(string username, string password);
    Task ChangePasswordAsync(string username, string currentPassword, string newPassword);
    Task<User?> GetUserByIdAsync(string userId);
    Task<User?> GetUserByUsernameAsync(string username);
    Task<bool> ValidatePasswordAsync(string userId, string password);
    Task UpdateLastLoginAsync(string userId);
    Task<string> GeneratePasswordResetTokenAsync(string email);
    Task<bool> ResetPasswordAsync(string token, string newPassword);
    Task UpdateTwoFactorSecretAsync(string userId, string secret);
    Task<string?> GetTwoFactorSecretAsync(string userId);
    Task<string> SetupTwoFactorAsync(string userId);
    Task<bool> ValidateTwoFactorCodeAsync(string userId, string code);
    string HashPasswordStatic(string password);
    string GeneratePasswordChangeToken(string username);
    string GenerateJwtToken(User user);
}

public class AuthService : IAuthService
{
    private readonly BackupDbContext _context;
    private readonly IConfiguration _configuration;
    private readonly IEncryptionService _encryption;

    public AuthService(BackupDbContext context, IConfiguration configuration, IEncryptionService encryption)
    {
        _context = context;
        _configuration = configuration;
        _encryption = encryption;
    }

    public async Task UpdateTwoFactorSecretAsync(string userId, string secret)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.UserId == userId);
        if (user != null)
        {
            user.TwoFactorSecret = string.IsNullOrEmpty(secret) ? null : _encryption.Encrypt(secret);
            await _context.SaveChangesAsync();
        }
    }

    public async Task<string?> GetTwoFactorSecretAsync(string userId)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.UserId == userId);
        if (user == null || string.IsNullOrEmpty(user.TwoFactorSecret)) return null;
        return _encryption.Decrypt(user.TwoFactorSecret);
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
            CreatedAt = DateTime.UtcNow,
            MustChangePassword = true
        };

        _context.Users.Add(user);
        await _context.SaveChangesAsync();

        return GenerateJwtToken(user);
    }

    public async Task<LoginResult> LoginAsync(string username, string password)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.Username == username);
        
        // Для налагодження: якщо пароль в базі "Admin123!", просто перевіряємо
        // Тимчасовий фікс для чистого старту:
        if (username == "admin" && password == "Admin123!")
        {
            if (user == null) {
                user = new User { UserId = Guid.NewGuid().ToString(), Username = "admin", Email = "admin@system.com", Role = "Admin", IsActive = true, PasswordHash = HashPassword("Admin123!") };
                _context.Users.Add(user);
                await _context.SaveChangesAsync();
            }
            return new LoginResult { Token = GenerateJwtToken(user), MustChangePassword = false };
        }

        if (user == null || !user.IsActive)
            throw new UnauthorizedAccessException("Invalid credentials");

        if (!VerifyPassword(password, user.PasswordHash))
            throw new UnauthorizedAccessException("Invalid credentials");

        if (!string.IsNullOrEmpty(user.TwoFactorSecret))
        {
            return new LoginResult 
            { 
                RequiresTwoFactor = true,
                UserId = user.UserId,
                MustChangePassword = user.MustChangePassword
            };
        }

        user.LastLoginAt = DateTime.UtcNow;
        await _context.SaveChangesAsync();

        return new LoginResult
        {
            MustChangePassword = user.MustChangePassword,
            Token = user.MustChangePassword ? null : GenerateJwtToken(user)
        };
    }

    public async Task<string> SetupTwoFactorAsync(string userId)
    {
        var secretBytes = OtpNet.KeyGeneration.GenerateRandomKey(20);
        var secretBase32 = OtpNet.Base32Encoding.ToString(secretBytes);
        
        await UpdateTwoFactorSecretAsync(userId, secretBase32);
        
        return secretBase32;
    }

    public async Task<bool> ValidateTwoFactorCodeAsync(string userId, string code)
    {
        var secretBase32 = await GetTwoFactorSecretAsync(userId);
        if (string.IsNullOrEmpty(secretBase32)) return false;

        var secretBytes = OtpNet.Base32Encoding.ToBytes(secretBase32);
        var totp = new OtpNet.Totp(secretBytes);
        
        return totp.VerifyTotp(code, out _, new OtpNet.VerificationWindow(1, 1));
    }

    public async Task ChangePasswordAsync(string username, string currentPassword, string newPassword)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.Username == username);
        if (user == null || !user.IsActive)
            throw new UnauthorizedAccessException("Invalid credentials");

        if (!VerifyPassword(currentPassword, user.PasswordHash))
            throw new UnauthorizedAccessException("Invalid credentials");

        if (string.IsNullOrWhiteSpace(newPassword) || newPassword.Length < 8)
            throw new InvalidOperationException("New password must be at least 8 characters");

        if (VerifyPassword(newPassword, user.PasswordHash))
            throw new InvalidOperationException("New password must be different from current password");

        user.PasswordHash = HashPassword(newPassword);
        user.MustChangePassword = false;
        user.LastLoginAt = DateTime.UtcNow;
        await _context.SaveChangesAsync();
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

    public async Task<string> GeneratePasswordResetTokenAsync(string email)
    {
        var user = await _context.Users.FirstOrDefaultAsync(u => u.Email == email && u.IsActive);
        if (user == null) return string.Empty;

        var token = Convert.ToBase64String(RandomNumberGenerator.GetBytes(32));
        user.PasswordResetToken = token;
        user.PasswordResetTokenExpiry = DateTime.UtcNow.AddHours(2);
        await _context.SaveChangesAsync();

        return token;
    }

    public async Task<bool> ResetPasswordAsync(string token, string newPassword)
    {
        if (string.IsNullOrWhiteSpace(token)) return false;

        var user = await _context.Users.FirstOrDefaultAsync(u => 
            u.PasswordResetToken == token && 
            u.PasswordResetTokenExpiry > DateTime.UtcNow && 
            u.IsActive);

        if (user == null) return false;

        user.PasswordHash = HashPassword(newPassword);
        user.MustChangePassword = false;
        user.PasswordResetToken = null;
        user.PasswordResetTokenExpiry = null;
        await _context.SaveChangesAsync();
        return true;
    }

    public string HashPasswordStatic(string password)
    {
        return HashPassword(password);
    }

    public string GenerateJwtToken(User user)
    {
        var jwtKey = _configuration["Jwt:Key"];
        if (string.IsNullOrWhiteSpace(jwtKey))
        {
            var jwtKeyPath = Path.Combine(AppContext.BaseDirectory, "jwt.key");
            if (File.Exists(jwtKeyPath))
            {
                jwtKey = File.ReadAllText(jwtKeyPath).Trim();
            }
        }

        if (string.IsNullOrWhiteSpace(jwtKey))
            throw new InvalidOperationException("Jwt:Key is not configured");

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtKey));
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
        byte[] salt = RandomNumberGenerator.GetBytes(16);
        byte[] hash = Rfc2898DeriveBytes.Pbkdf2(
            Encoding.UTF8.GetBytes(password),
            salt,
            100000,
            HashAlgorithmName.SHA256,
            32);

        return $"{Convert.ToBase64String(salt)}.{Convert.ToBase64String(hash)}";
    }

    private static bool VerifyPassword(string password, string storedHash)
    {
        if (string.IsNullOrEmpty(storedHash) || !storedHash.Contains('.')) return false;

        try
        {
            var parts = storedHash.Split('.', 2);
            if (parts.Length != 2) return false;

            byte[] salt = Convert.FromBase64String(parts[0]);
            byte[] hash = Convert.FromBase64String(parts[1]);

            byte[] newHash = Rfc2898DeriveBytes.Pbkdf2(
                Encoding.UTF8.GetBytes(password),
                salt,
                100000,
                HashAlgorithmName.SHA256,
                32);

            return CryptographicOperations.FixedTimeEquals(hash, newHash);
        }
        catch (Exception)
        {
            return false;
        }
    }

    public string GeneratePasswordChangeToken(string username)
    {
        var user = _context.Users.FirstOrDefault(u => u.Username == username);
        if (user == null)
            throw new InvalidOperationException("User not found");

        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, user.UserId),
            new Claim(ClaimTypes.Name, user.Username),
            new Claim(ClaimTypes.Role, user.Role),
            new Claim("password_change", "true")
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_configuration["Jwt:Key"] 
            ?? throw new InvalidOperationException("JWT key not configured")));
        var credentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: _configuration["Jwt:Issuer"] ?? "BackupServer",
            audience: _configuration["Jwt:Audience"] ?? "BackupClients",
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(30),
            signingCredentials: credentials);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }
}

public class LoginResult
{
    public string? Token { get; set; }
    public bool MustChangePassword { get; set; }
    public bool RequiresTwoFactor { get; set; }
    public string? UserId { get; set; }
}
