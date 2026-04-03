using Backup.Server.Database;
using Backup.Server.Services;
using Backup.Server.BackgroundServices;
using Microsoft.EntityFrameworkCore;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using Microsoft.OpenApi.Models;
using Backup.Server.Database.Entities;
using Serilog;
using Swashbuckle.AspNetCore.SwaggerGen;
using System.Text;
using System.Net;
using System.Net.Sockets;

Log.Logger = new LoggerConfiguration()
    .WriteTo.Console()
    .CreateLogger();

try
{
    Log.Information("Starting Backup Server...");
    const int defaultServerPort = 8000;

    var builder = WebApplication.CreateBuilder(args);

    builder.Host.UseSerilog();

    var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
        ?? "Host=localhost;Database=backup;Username=backup_user;Password=postgres";

    builder.Services.AddDbContext<BackupDbContext>(options =>
        options.UseNpgsql(connectionString));

    var jwtKey = builder.Configuration["Jwt:Key"];
    if (string.IsNullOrWhiteSpace(jwtKey))
    {
        throw new InvalidOperationException("Missing required configuration: Jwt:Key");
    }
    var jwtIssuer = builder.Configuration["Jwt:Issuer"] ?? "BackupServer";
    var jwtAudience = builder.Configuration["Jwt:Audience"] ?? "BackupClients";

    builder.Services.AddAuthentication(options =>
    {
        options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
        options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
    })
    .AddJwtBearer(options =>
    {
        options.TokenValidationParameters = new TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidateAudience = true,
            ValidateLifetime = true,
            ValidateIssuerSigningKey = true,
            ValidIssuer = jwtIssuer,
            ValidAudience = jwtAudience,
            IssuerSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtKey)),
            ClockSkew = TimeSpan.Zero
        };
    });

    builder.Services.AddAuthorization(options =>
    {
        options.AddPolicy("Admin", policy => policy.RequireRole("Admin"));
        options.AddPolicy("Operator", policy => policy.RequireRole("Admin", "Operator"));
        options.AddPolicy("Viewer", policy => policy.RequireRole("Admin", "Operator", "Viewer"));
    });

    builder.Services.AddControllers();
    builder.Services.AddEndpointsApiExplorer();
    builder.Services.AddSwaggerGen(c =>
    {
        c.SwaggerDoc("v1", new() { Title = "Backup API", Version = "v1" });
        c.AddSecurityDefinition("Bearer", new OpenApiSecurityScheme
        {
            Description = "JWT Authorization header using the Bearer scheme",
            Name = "Authorization",
            In = ParameterLocation.Header,
            Type = SecuritySchemeType.ApiKey,
            Scheme = "Bearer"
        });
        c.AddSecurityRequirement(new OpenApiSecurityRequirement
        {
            {
                new OpenApiSecurityScheme
                {
                    Reference = new OpenApiReference { Type = ReferenceType.SecurityScheme, Id = "Bearer" }
                },
                Array.Empty<string>()
            }
        });
    });

    builder.Services.AddCors(o => o.AddDefaultPolicy(p => p
        .AllowAnyOrigin()
        .AllowAnyMethod()
        .AllowAnyHeader()));

    builder.Services.AddScoped<IAuthService, AuthService>();

    builder.Services.AddHostedService<JobSchedulerService>();
    builder.Services.AddHostedService<AgentHealthCheckService>();
    builder.Services.AddHostedService<RetentionPolicyService>();

    var app = builder.Build();

    var hostAddress = Dns.GetHostAddresses(Dns.GetHostName())
        .FirstOrDefault(ip => ip.AddressFamily == AddressFamily.InterNetwork && !IPAddress.IsLoopback(ip))
        ?.ToString() ?? "localhost";

    // Bind to both localhost and server IP unless explicit URL config is provided.
    if (!app.Urls.Any())
    {
        app.Urls.Add($"http://localhost:{defaultServerPort}");
        app.Urls.Add($"http://{hostAddress}:{defaultServerPort}");
    }

    app.UseSwagger();
    app.UseSwaggerUI(c =>
    {
        c.SwaggerEndpoint("/swagger/v1/swagger.json", "Backup API v1");
        c.RoutePrefix = "swagger";
    });

    using (var scope = app.Services.CreateScope())
    {
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
        db.Database.EnsureCreated();
        Log.Information("Database initialized");
        
        var authService = scope.ServiceProvider.GetRequiredService<IAuthService>();
        var bootstrapAdminUsername = builder.Configuration["BootstrapAdmin:Username"] ?? "admin";
        var bootstrapAdminEmail = builder.Configuration["BootstrapAdmin:Email"] ?? "admin@backupsystem.com";
        var bootstrapAdminPassword = builder.Configuration["BootstrapAdmin:Password"] ?? "admin123";
        var configuredPublicServerUrl = builder.Configuration["Server:PublicUrl"];
        var publicServerUrl = configuredPublicServerUrl;
        if (string.IsNullOrWhiteSpace(publicServerUrl))
        {
            publicServerUrl = $"http://{hostAddress}:{defaultServerPort}";
        }

        var adminUser = await authService.GetUserByUsernameAsync(bootstrapAdminUsername);
        if (adminUser == null)
        {
            await authService.RegisterAsync(bootstrapAdminUsername, bootstrapAdminEmail, bootstrapAdminPassword, "Admin");
            var createdAdmin = await authService.GetUserByUsernameAsync(bootstrapAdminUsername);
            if (createdAdmin != null)
            {
                createdAdmin.MustChangePassword = true;
                await db.SaveChangesAsync();
            }

            Log.Information("Bootstrap admin user created and marked for first-login password change");
        }

        var publicUrlSetting = await db.Settings.FirstOrDefaultAsync(s => s.Key == "server.public_url");
        if (publicUrlSetting == null)
        {
            db.Settings.Add(new Setting
            {
                Key = "server.public_url",
                Value = publicServerUrl,
                Type = "string",
                Description = "Public server URL used by agents and installers",
                UpdatedAt = DateTime.UtcNow
            });
            await db.SaveChangesAsync();
        }
    }

    app.UseSerilogRequestLogging();
    app.UseCors();
    app.UseStaticFiles();
    app.UseAuthentication();
    app.UseAuthorization();

    app.MapControllers();

    app.MapGet("/", () => "Backup Server v1.0.0");
    app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));

    app.Run();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Server terminated unexpectedly");
    return 1;
}
finally
{
    Log.CloseAndFlush();
}

return 0;
