
using System;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Security.Cryptography;
using System.Text;
using Microsoft.IdentityModel.Tokens;

class Program {
    static void Main() {
        var jwtKey = Convert.ToBase64String(RandomNumberGenerator.GetBytes(32));
        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtKey));
        var credentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);
        
        var claims = new[] {
            new Claim(ClaimTypes.NameIdentifier, "123"),
            new Claim(ClaimTypes.Name, "admin"),
            new Claim(ClaimTypes.Role, "Admin"),
            new Claim("password_change", "true")
        };
        
        var token = new JwtSecurityToken(
            issuer: "BackupServer",
            audience: "BackupClients",
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(30),
            signingCredentials: credentials);
            
        var tokenString = new JwtSecurityTokenHandler().WriteToken(token);
        Console.WriteLine("Generated Token");
        
        try {
            var handler = new JwtSecurityTokenHandler();
            var principal = handler.ValidateToken(tokenString, new TokenValidationParameters {
                ValidateIssuerSigningKey = true,
                IssuerSigningKey = key,
                ValidateIssuer = true,
                ValidIssuer = "BackupServer",
                ValidateAudience = true,
                ValidAudience = "BackupClients",
                ValidateLifetime = true,
                ClockSkew = TimeSpan.Zero
            }, out SecurityToken validatedToken);
            
            var tokenUsername = principal.FindFirst(ClaimTypes.Name)?.Value;
            var isPasswordChange = principal.FindFirst("password_change")?.Value;
            Console.WriteLine($"tokenUsername: {tokenUsername}, isPasswordChange: {isPasswordChange}");
        } catch (Exception ex) {
            Console.WriteLine("Error: " + ex.Message);
        }
    }
}

