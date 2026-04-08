using Backup.Server.BackgroundServices;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Backup.Server.Services;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using Microsoft.OpenApi.Models;
using Serilog;
using System.Net;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;

namespace Backup.Server;

public partial class Program
{
    private const int DefaultServerPort = 8000;

    public static async Task Main(string[] args)
    {
        Environment.ExitCode = await RunAsync(args);
    }

    public static async Task<int> RunAsync(string[] args)
    {
        var isService = Environment.UserInteractive == false
            || Environment.CommandLine.Contains("--service")
            || AppDomain.CurrentDomain.FriendlyName.Contains("Backup.Server", StringComparison.OrdinalIgnoreCase);

        ConfigureLogging(isService);

        try
        {
            if (isService)
            {
                Log.Information("Starting Backup Server as Windows Service...");
            }
            else
            {
                Log.Information("Starting Backup Server (console mode)...");
            }

            var app = BuildApplication(args, isService);
            await InitializeApplicationAsync(app);
            await app.RunAsync();
            return 0;
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
    }

    private static void ConfigureLogging(bool isService)
    {
        var logConfig = new LoggerConfiguration()
            .MinimumLevel.Information()
            .WriteTo.Console(outputTemplate: "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj}{NewLine}{Exception}");

        if (isService)
        {
            var logDir = Path.Combine(AppContext.BaseDirectory, "logs");
            Directory.CreateDirectory(logDir);
            logConfig.WriteTo.File(
                Path.Combine(logDir, "backup-server-.log"),
                rollingInterval: RollingInterval.Day,
                retainedFileCountLimit: 30,
                outputTemplate: "{Timestamp:yyyy-MM-dd HH:mm:ss.fff zzz} [{Level:u3}] {Message:lj}{NewLine}{Exception}");
        }

        Log.Logger = logConfig.CreateLogger();
    }

    private static WebApplication BuildApplication(string[] args, bool isService)
    {
        var builder = WebApplication.CreateBuilder(args);

        builder.Host.UseSerilog();

        if (isService)
        {
            builder.Host.UseWindowsService(options =>
            {
                options.ServiceName = "BackupServer";
            });
        }

        var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
            ?? "Host=localhost;Database=backup_db;Username=postgres;Password=postgres";

        builder.Services.AddDbContext<BackupDbContext>(options =>
            options.UseNpgsql(connectionString));

        builder.WebHost.ConfigureKestrel(options =>
        {
            options.ListenAnyIP(DefaultServerPort, listenOptions =>
            {
                listenOptions.Protocols = Microsoft.AspNetCore.Server.Kestrel.Core.HttpProtocols.Http1AndHttp2;
            });
        });

        builder.Services.AddGrpc();

        var agentManager = new AgentGrpcService(builder.Services.BuildServiceProvider(), builder.Logging.CreateLogger<AgentGrpcService>());
        builder.Services.AddSingleton<IAgentManager>(agentManager);
        builder.Services.AddSingleton<AgentGrpcService>(agentManager);

        var jwtKey = EnsureJwtKey(builder.Configuration);
        var jwtIssuer = builder.Configuration["Jwt:Issuer"] ?? "BackupServer";
        var jwtAudience = builder.Configuration["Jwt:Audience"] ?? "BackupClients";

        builder.Services.AddAuthentication(options =>
        {
            options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
            options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
        })
        .AddJwtBearer(options => { /* ... ваш поточний JWT ... */ })
        .AddOpenIdConnect("oidc", options =>
        {
            options.Authority = builder.Configuration["Oidc:Authority"];
            options.ClientId = builder.Configuration["Oidc:ClientId"];
            options.ClientSecret = builder.Configuration["Oidc:ClientSecret"];
            options.ResponseType = "code";
            options.SaveTokens = true;
            options.Scope.Add("openid");
            options.Scope.Add("profile");
            options.Scope.Add("email");
        });

        builder.Services.AddAuthorization(options =>
        {
            options.AddPolicy("Admin", policy => policy.RequireRole("Admin"));
            options.AddPolicy("Operator", policy => policy.RequireRole("Admin", "Operator"));
            options.AddPolicy("Viewer", policy => policy.RequireRole("Admin", "Operator", "Viewer"));
            options.FallbackPolicy = null;
        });

        builder.Services.AddControllers()
            .AddJsonOptions(options =>
            {
                options.JsonSerializerOptions.Converters.Add(new System.Text.Json.Serialization.JsonStringEnumConverter());
            });
        builder.Services.AddEndpointsApiExplorer();
        builder.Services.AddSwaggerGen(c =>
        {
            c.SwaggerDoc("v1", new OpenApiInfo { Title = "Backup API", Version = "v1" });
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

        ConfigureCors(builder);

        builder.Services.AddScoped<IAuthService, AuthService>();
        builder.Services.AddScoped<IEncryptionService, EncryptionService>();
        builder.Services.AddScoped<IAuditService, AuditService>();
        builder.Services.AddScoped<SchedulerService>();
        builder.Services.AddScoped<RepositoryService>();
        builder.Services.AddScoped<ICloudStorageService, CloudStorageService>();
        builder.Services.AddScoped<BackupExecutionService>();
        builder.Services.AddScoped<FastCloneService>();
        builder.Services.AddScoped<RestoreService>();
        builder.Services.AddSingleton<IBackupQueue, BackupQueue>();
        builder.Services.AddSingleton<IRestoreQueue, RestoreQueue>();
        builder.Services.AddHostedService<BackupProcessingService>();
        builder.Services.AddHostedService<JobSchedulerService>();
        builder.Services.AddHostedService<AgentHealthCheckService>();
        builder.Services.AddHostedService<RetentionPolicyService>();
        builder.Services.AddHostedService<RestoreProcessingService>();
        builder.Services.AddHostedService<TelegramBotService>();

        var app = builder.Build();
        ConfigureMiddleware(app);
        return app;
    }

    private static string EnsureJwtKey(ConfigurationManager configuration)
    {
        var jwtKey = configuration["Jwt:Key"];
        if (!string.IsNullOrWhiteSpace(jwtKey))
        {
            return jwtKey;
        }

        var jwtKeyPath = Path.Combine(AppContext.BaseDirectory, "jwt.key");
        if (File.Exists(jwtKeyPath))
        {
            jwtKey = File.ReadAllText(jwtKeyPath).Trim();
            if (string.IsNullOrWhiteSpace(jwtKey))
            {
                throw new InvalidOperationException("Jwt:Key is empty in configuration and jwt.key file");
            }

            Log.Information("JWT key loaded from jwt.key file");
            configuration["Jwt:Key"] = jwtKey;
            return jwtKey;
        }

        var bytes = RandomNumberGenerator.GetBytes(64);
        jwtKey = Convert.ToBase64String(bytes);
        File.WriteAllText(jwtKeyPath, jwtKey);
        File.SetAttributes(jwtKeyPath, FileAttributes.Hidden);
        configuration["Jwt:Key"] = jwtKey;
        Log.Warning("Generated new JWT key and saved to jwt.key. Keep this file secure.");
        return jwtKey;
    }

    private static void ConfigureCors(WebApplicationBuilder builder)
    {
        var allowedOrigins = builder.Configuration.GetSection("AllowedOrigins").Get<string[]>()
            ?? Array.Empty<string>();
        var developmentOrigins = BuildDevelopmentCorsOrigins(
            builder.Configuration["Server:PublicUrl"],
            DefaultServerPort);

        builder.Services.AddCors(options =>
        {
            if (allowedOrigins.Length > 0)
            {
                options.AddDefaultPolicy(policy => policy
                    .WithOrigins(allowedOrigins)
                    .AllowAnyMethod()
                    .AllowAnyHeader()
                    .AllowCredentials());
            }
            else
            {
                options.AddDefaultPolicy(policy => policy
                    .WithOrigins(developmentOrigins)
                    .AllowAnyMethod()
                    .AllowAnyHeader());
                Log.Warning(
                    "No AllowedOrigins configured. CORS is limited to local development origins: {Origins}. Configure AllowedOrigins for deployed environments.",
                    string.Join(", ", developmentOrigins));
            }
        });
    }

    private static void ConfigureMiddleware(WebApplication app)
    {
        app.UseSwagger();
        app.UseSwaggerUI(c =>
        {
            c.SwaggerEndpoint("/swagger/v1/swagger.json", "Backup API v1");
            c.RoutePrefix = "swagger";
        });

        app.UseSerilogRequestLogging();
        app.UseCors();
        app.UseStaticFiles();
        app.UseAuthentication();
        app.UseAuthorization();

        app.MapGrpcService<AgentGrpcService>();
        app.MapControllers();
        app.MapGet("/health", () => Results.Ok(new { status = "healthy" })).AllowAnonymous();
        app.MapGet("/api/health", () => Results.Ok(new { status = "healthy" })).AllowAnonymous();
        app.MapGet("/api", () => Results.Ok(new { message = "Backup API" })).AllowAnonymous();
        app.MapFallbackToFile("index.html").AllowAnonymous();
    }

    private static async Task InitializeApplicationAsync(WebApplication app)
    {
        var hostAddress = Dns.GetHostAddresses(Dns.GetHostName())
            .FirstOrDefault(ip => ip.AddressFamily == AddressFamily.InterNetwork && !IPAddress.IsLoopback(ip))
            ?.ToString() ?? "localhost";

        if (!app.Urls.Any())
        {
            app.Urls.Add($"http://localhost:{DefaultServerPort}");
            app.Urls.Add($"http://{hostAddress}:{DefaultServerPort}");
        }

        using var scope = app.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
        db.Database.Migrate();
        Log.Information("Database migrations applied");

        var authService = scope.ServiceProvider.GetRequiredService<IAuthService>();
        var bootstrapAdminUsername = app.Configuration["BootstrapAdmin:Username"] ?? "admin";
        var bootstrapAdminEmail = app.Configuration["BootstrapAdmin:Email"] ?? "admin@backupsystem.com";
        var bootstrapAdminPassword = app.Configuration["BootstrapAdmin:Password"];
        
        if (string.IsNullOrWhiteSpace(bootstrapAdminPassword))
        {
            var bytes = RandomNumberGenerator.GetBytes(16);
            bootstrapAdminPassword = Convert.ToBase64String(bytes)[..12];
            Log.Warning("No BootstrapAdmin:Password provided. Generated random password: {Password}. CHANGE THIS IMMEDIATELY.", bootstrapAdminPassword);
        }

        var publicServerUrl = app.Configuration["Server:PublicUrl"];
        if (string.IsNullOrWhiteSpace(publicServerUrl))
        {
            publicServerUrl = $"http://{hostAddress}:{DefaultServerPort}";
        }

        var adminUser = await authService.GetUserByUsernameAsync(bootstrapAdminUsername);
        if (adminUser == null)
        {
            await authService.RegisterAsync(bootstrapAdminUsername, bootstrapAdminEmail, bootstrapAdminPassword, "Admin");
            Log.Information("Bootstrap admin user created");
        }
        else if (!adminUser.PasswordHash.Contains('.'))
        {
            Log.Warning("Legacy password hash detected for user {Username}. Migrating to PBKDF2...", bootstrapAdminUsername);
            adminUser.PasswordHash = authService.HashPasswordStatic(bootstrapAdminPassword);
            await db.SaveChangesAsync();
            Log.Information("Password migrated successfully for {Username}", bootstrapAdminUsername);
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

    private static string[] BuildDevelopmentCorsOrigins(string? configuredPublicUrl, int defaultServerPort)
    {
        var origins = new HashSet<string>(StringComparer.OrdinalIgnoreCase)
        {
            $"http://localhost:{defaultServerPort}",
            "http://localhost:3000",
            "http://127.0.0.1:3000"
        };

        if (Uri.TryCreate(configuredPublicUrl, UriKind.Absolute, out var publicUri))
        {
            origins.Add(publicUri.GetLeftPart(UriPartial.Authority));
        }

        return origins.ToArray();
    }
}
