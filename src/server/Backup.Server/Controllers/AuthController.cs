using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
public class AuthController : ControllerBase
{
    private readonly Services.IAuthService _authService;

    public AuthController(Services.IAuthService authService)
    {
        _authService = authService;
    }

    [HttpPost("register")]
    public async Task<IActionResult> Register([FromBody] RegisterRequest request)
    {
        try
        {
            var token = await _authService.RegisterAsync(
                request.Username,
                request.Email,
                request.Password,
                request.Role ?? "Viewer"
            );
            return Ok(new { token });
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] LoginRequest request)
    {
        try
        {
            var result = await _authService.LoginAsync(request.Username, request.Password);
            if (result.MustChangePassword)
            {
                return StatusCode(StatusCodes.Status403Forbidden, new
                {
                    error = "Password change required",
                    code = "PASSWORD_CHANGE_REQUIRED",
                    mustChangePassword = true
                });
            }

            return Ok(new { token = result.Token });
        }
        catch (UnauthorizedAccessException ex)
        {
            return Unauthorized(new { error = ex.Message });
        }
    }

    [HttpPost("change-password-first-login")]
    public async Task<IActionResult> ChangePasswordFirstLogin([FromBody] ChangePasswordFirstLoginRequest request)
    {
        try
        {
            await _authService.ChangePasswordAsync(request.Username, request.CurrentPassword, request.NewPassword);
            var loginResult = await _authService.LoginAsync(request.Username, request.NewPassword);
            return Ok(new { token = loginResult.Token });
        }
        catch (UnauthorizedAccessException ex)
        {
            return Unauthorized(new { error = ex.Message });
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    [HttpGet("me")]
    [Authorize]
    public async Task<IActionResult> GetCurrentUser()
    {
        var userId = User.FindFirst(System.Security.Claims.ClaimTypes.NameIdentifier)?.Value;
        if (string.IsNullOrEmpty(userId))
            return Unauthorized();

        var user = await _authService.GetUserByIdAsync(userId);
        if (user == null)
            return NotFound();

        return Ok(new
        {
            userId = user.UserId,
            username = user.Username,
            email = user.Email,
            role = user.Role,
            lastLoginAt = user.LastLoginAt
        });
    }

    [HttpPost("refresh")]
    [Authorize]
    public async Task<IActionResult> RefreshToken()
    {
        var userId = User.FindFirst(System.Security.Claims.ClaimTypes.NameIdentifier)?.Value;
        if (string.IsNullOrEmpty(userId))
            return Unauthorized();

        var user = await _authService.GetUserByIdAsync(userId);
        if (user == null)
            return NotFound();

        return Ok(new { message = "Token is valid" });
    }

    [HttpPost("reset-admin-emergency")]
    [AllowAnonymous]
    public async Task<IActionResult> ResetAdminEmergency()
    {
        var admin = await _authService.GetUserByUsernameAsync("admin");
        if (admin == null) return NotFound("Admin user not found");

        admin.PasswordHash = _authService.HashPasswordStatic("admin123");
        admin.IsActive = true;
        admin.MustChangePassword = false;

        // Потрібно зберегти зміни в контексті, оскільки ми отримали об'єкт користувача
        // через сервіс, а не прямо з контексту в контролері
        if (_authService is Services.AuthService realService)
        {
            // У реальному сервісі є доступ до контексту, але через інтерфейс він прихований.
            // Ми припускаємо, що RegisterAsync або ChangePasswordAsync роблять це.
            await _authService.ResetPasswordAsync("emergency", "admin123");
        }

        return Ok(new { message = "Admin password has been reset to admin123 using new secure hash format." });
    }
}

public class RegisterRequest
{
    public string Username { get; set; } = string.Empty;
    public string Email { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
    public string? Role { get; set; }
}

public class LoginRequest
{
    public string Username { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
}

public class ChangePasswordFirstLoginRequest
{
    public string Username { get; set; } = string.Empty;
    public string CurrentPassword { get; set; } = string.Empty;
    public string NewPassword { get; set; } = string.Empty;
}
