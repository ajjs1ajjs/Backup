using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Backup.Server.Database;
using Backup.Server.Services;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize(Roles = "admin")]
public class UsersController : ControllerBase
{
    private readonly BackupDbContext _context;
    private readonly Services.IAuthService _authService;

    public UsersController(BackupDbContext context, Services.IAuthService authService)
    {
        _context = context;
        _authService = authService;
    }

    [HttpGet]
    public async Task<IActionResult> GetUsers()
    {
        var users = await _context.Users
            .Select(u => new
            {
                u.Id,
                u.Username,
                u.Email,
                u.Role,
                u.IsActive,
                u.CreatedAt,
                u.MustChangePassword
            })
            .OrderBy(u => u.Username)
            .ToListAsync();

        return Ok(users);
    }

    [HttpGet("{id}")]
    public async Task<IActionResult> GetUser(long id)
    {
        var user = await _context.Users.FindAsync(id);
        if (user == null)
            return NotFound(new { error = "User not found" });

        return Ok(new
        {
            user.Id,
            user.Username,
            user.Email,
            user.Role,
            user.IsActive,
            user.CreatedAt,
            user.MustChangePassword
        });
    }

    [HttpPost]
    public async Task<IActionResult> CreateUser([FromBody] CreateUserRequest request)
    {
        var existingUser = await _context.Users.FirstOrDefaultAsync(u => u.Username == request.Username);
        if (existingUser != null)
            return BadRequest(new { error = "Username already exists" });

        var existingEmail = await _context.Users.FirstOrDefaultAsync(u => u.Email == request.Email);
        if (existingEmail != null)
            return BadRequest(new { error = "Email already exists" });

        var passwordHash = _authService.HashPasswordStatic(request.Password);
        var userId = Guid.NewGuid().ToString();

        var user = new Database.Entities.User
        {
            UserId = userId,
            Username = request.Username,
            Email = request.Email,
            PasswordHash = passwordHash,
            Role = request.Role,
            IsActive = true,
            CreatedAt = DateTime.UtcNow,
            MustChangePassword = true
        };

        _context.Users.Add(user);
        await _context.SaveChangesAsync();

        return CreatedAtAction(nameof(GetUser), new { id = user.Id }, new
        {
            user.Id,
            user.Username,
            user.Email,
            user.Role,
            user.IsActive
        });
    }

    [HttpPut("{id}")]
    public async Task<IActionResult> UpdateUser(long id, [FromBody] UpdateUserRequest request)
    {
        var user = await _context.Users.FindAsync(id);
        if (user == null)
            return NotFound(new { error = "User not found" });

        if (user.Username == "admin" && request.Role != "admin")
            return BadRequest(new { error = "Cannot change admin role" });

        if (!string.IsNullOrEmpty(request.Username) && request.Username != user.Username)
        {
            var existing = await _context.Users.FirstOrDefaultAsync(u => u.Username == request.Username && u.Id != id);
            if (existing != null)
                return BadRequest(new { error = "Username already exists" });
            user.Username = request.Username;
        }

        if (!string.IsNullOrEmpty(request.Email) && request.Email != user.Email)
        {
            var existing = await _context.Users.FirstOrDefaultAsync(u => u.Email == request.Email && u.Id != id);
            if (existing != null)
                return BadRequest(new { error = "Email already exists" });
            user.Email = request.Email;
        }

        if (!string.IsNullOrEmpty(request.Role))
            user.Role = request.Role;

        if (request.IsActive.HasValue)
            user.IsActive = request.IsActive.Value;

        if (!string.IsNullOrEmpty(request.NewPassword))
        {
            user.PasswordHash = _authService.HashPasswordStatic(request.NewPassword);
            user.MustChangePassword = true;
        }

        await _context.SaveChangesAsync();

        return Ok(new { message = "User updated successfully" });
    }

    [HttpDelete("{id}")]
    public async Task<IActionResult> DeleteUser(long id)
    {
        var user = await _context.Users.FindAsync(id);
        if (user == null)
            return NotFound(new { error = "User not found" });

        if (user.Username == "admin")
            return BadRequest(new { error = "Cannot delete admin user" });

        _context.Users.Remove(user);
        await _context.SaveChangesAsync();

        return Ok(new { message = "User deleted successfully" });
    }

    [HttpPost("{id}/reset-password")]
    public async Task<IActionResult> ResetPassword(long id, [FromBody] ResetPasswordRequest request)
    {
        var user = await _context.Users.FindAsync(id);
        if (user == null)
            return NotFound(new { error = "User not found" });

        user.PasswordHash = _authService.HashPasswordStatic(request.NewPassword);
        user.MustChangePassword = true;
        await _context.SaveChangesAsync();

        return Ok(new { message = "Password reset successfully" });
    }
}

public class CreateUserRequest
{
    public string Username { get; set; } = string.Empty;
    public string Email { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
    public string Role { get; set; } = "Viewer";
}

public class UpdateUserRequest
{
    public string? Username { get; set; }
    public string? Email { get; set; }
    public string? Role { get; set; }
    public bool? IsActive { get; set; }
    public string? NewPassword { get; set; }
}

public class ResetPasswordRequest
{
    public string NewPassword { get; set; } = string.Empty;
}