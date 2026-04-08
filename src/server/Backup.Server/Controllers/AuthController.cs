using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Backup.Server.Controllers;

[ApiController]
[Route("api/[controller]")]
public class AuthController : ControllerBase
{
    private readonly Services.IAuthService _authService;
    private readonly Services.IAuditService _auditService;

    public AuthController(Services.IAuthService authService, Services.IAuditService auditService)
    {
        _authService = authService;
        _auditService = auditService;
    }

    [HttpPost("register")]
    [AllowAnonymous]
    public async Task<IActionResult> Register([FromBody] RegisterRequest request)
    {
        try
        {
            var token = await _authService.RegisterAsync(
                request.Username,
                request.Email,
                request.Password,
                "Viewer");

            return Ok(new { token });
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    [HttpPost("login")]
    [AllowAnonymous]
    public async Task<IActionResult> Login([FromBody] LoginRequest request)
    {
        try
        {
            var result = await _authService.LoginAsync(request.Username, request.Password);
            
            if (result.RequiresTwoFactor)
            {
                await _auditService.LogAsync(result.UserId, "LoginInitiated", "User", result.UserId, new { method = "2FA" }, Request.HttpContext.Connection.RemoteIpAddress?.ToString());
                return Ok(new { 
                    requiresTwoFactor = true, 
                    userId = result.UserId,
                    message = "Two-factor authentication required" 
                });
            }

            if (result.MustChangePassword)
            {
                var changePasswordToken = _authService.GeneratePasswordChangeToken(request.Username);
                return Ok(new { 
                    token = changePasswordToken,
                    mustChangePassword = true,
                    message = "Password change required"
                });
            }

            var user = await _authService.GetUserByUsernameAsync(request.Username);
            await _auditService.LogAsync(user?.UserId, "LoginSuccess", "User", user?.UserId, null, Request.HttpContext.Connection.RemoteIpAddress?.ToString());

            return Ok(new { token = result.Token });
        }
        catch (UnauthorizedAccessException ex)
        {
            await _auditService.LogAsync(null, "LoginFailed", "User", request.Username, new { error = ex.Message }, Request.HttpContext.Connection.RemoteIpAddress?.ToString());
            return Unauthorized(new { error = ex.Message });
        }
    }

    [HttpPost("login-2fa")]
    [AllowAnonymous]
    public async Task<IActionResult> Login2Fa([FromBody] Login2FaRequest request)
    {
        try
        {
            var isValid = await _authService.ValidateTwoFactorCodeAsync(request.UserId, request.Code);
            if (!isValid)
            {
                return Unauthorized(new { error = "Invalid two-factor code" });
            }

            var user = await _authService.GetUserByIdAsync(request.UserId);
            if (user == null) return NotFound(new { error = "User not found" });

            await _authService.UpdateLastLoginAsync(request.UserId);
            var token = _authService.GenerateJwtToken(user);
            
            return Ok(new { token });
        }
        catch (Exception ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

public class Login2FaRequest
{
    public string UserId { get; set; } = string.Empty;
    public string Code { get; set; } = string.Empty;
}

    [HttpPost("change-password-first-login")]
    [AllowAnonymous]
    public async Task<IActionResult> ChangePasswordFirstLogin([FromBody] ChangePasswordFirstLoginRequest request)
    {
        try
        {
            await _authService.ChangePasswordWithTokenAsync(request.Username, request.Token, request.NewPassword);
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

    [HttpPost("change-password")]
    [Authorize]
    public async Task<IActionResult> ChangePassword([FromBody] ChangePasswordFirstLoginRequest request)
    {
        try
        {
            await _authService.ChangePasswordAsync(request.Username, request.CurrentPassword, request.NewPassword);
            return Ok(new { message = "Password changed successfully" });
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
        {
            return Unauthorized();
        }

        var user = await _authService.GetUserByIdAsync(userId);
        if (user == null)
        {
            return NotFound();
        }

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
        {
            return Unauthorized();
        }

        var user = await _authService.GetUserByIdAsync(userId);
        if (user == null)
        {
            return NotFound();
        }

        return Ok(new { message = "Token is valid" });
    }

    [HttpPost("2fa/setup")]
    [Authorize]
    public async Task<IActionResult> SetupTwoFactor()
    {
        var userId = User.FindFirst(System.Security.Claims.ClaimTypes.NameIdentifier)?.Value;
        if (string.IsNullOrEmpty(userId)) return Unauthorized();

        var secret = await _authService.SetupTwoFactorAsync(userId);
        return Ok(new { secret });
    }

    [HttpPost("2fa/verify")]
    [Authorize]
    public async Task<IActionResult> VerifyTwoFactor([FromBody] Login2FaRequest request)
    {
        var userId = User.FindFirst(System.Security.Claims.ClaimTypes.NameIdentifier)?.Value;
        if (string.IsNullOrEmpty(userId)) return Unauthorized();

        var isValid = await _authService.ValidateTwoFactorCodeAsync(userId, request.Code);
        return isValid ? Ok(new { message = "2FA enabled" }) : BadRequest(new { error = "Invalid code" });
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
    public string Token { get; set; } = string.Empty;
    public string NewPassword { get; set; } = string.Empty;
}
