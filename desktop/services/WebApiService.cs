using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.FileProviders;
using Newtonsoft.Json;

namespace NovaBackup.Desktop.Services
{
    public class WebApiService
    {
        private readonly string _baseUrl;
        private readonly HttpClient _httpClient;
        private IHost _webHost;

        public WebApiService(string baseUrl = "http://0.0.0.0:8080")
        {
            _baseUrl = baseUrl;
            _httpClient = new HttpClient();
            _httpClient.BaseAddress = new Uri(baseUrl);
        }

        public async Task<bool> StartWebServer()
        {
            try
            {
                var assembly = System.Reflection.Assembly.GetExecutingAssembly();
                var contentRoot = Path.GetDirectoryName(assembly.Location);
                
                _webHost = Host.CreateDefaultBuilder()
                    .ConfigureWebHostDefaults(webBuilder =>
                    {
                        // Bind to all interfaces for remote access
                        webBuilder.UseUrls("http://0.0.0.0:8080");
                        webBuilder.ConfigureServices(services =>
                        {
                            services.AddCors();
                        });
                        webBuilder.Configure(app =>
                        {
                            app.UseRouting();
                            app.UseCors(builder =>
                            {
                                builder.AllowAnyOrigin()
                                       .AllowAnyMethod()
                                       .AllowAnyHeader()
                                       .AllowCredentials();
                            });

                            // Add authentication middleware
                            app.Use(async (context, next) =>
                            {
                                // Basic authentication for remote access
                                var authHeader = context.Request.Headers["Authorization"].FirstOrDefault();
                                if (string.IsNullOrEmpty(authHeader) || !IsValidAuth(authHeader))
                                {
                                    // Allow localhost access without auth
                                    var remoteIp = context.Connection.RemoteIpAddress?.ToString();
                                    if (remoteIp != "::1" && remoteIp != "127.0.0.1")
                                    {
                                        context.Response.StatusCode = 401;
                                        await context.Response.WriteAsync("Unauthorized");
                                        return;
                                    }
                                }
                                await next();
                            });

                            // API Routes
                            app.MapGet("/api/status", GetStatus);
                            app.MapGet("/api/backups", GetBackups);
                            app.MapPost("/api/backups", CreateBackup);
                            app.MapGet("/api/schedules", GetSchedules);
                            app.MapPost("/api/schedules", CreateSchedule);
                            app.MapGet("/api/storage", GetStorage);
                            app.MapGet("/api/reports", GetReports);
                            
                            // Static files for web UI - use embedded resources
                            app.UseStaticFiles(new StaticFileOptions
                            {
                                FileProvider = new EmbeddedFileProvider(assembly, "NovaBackup.Desktop.web-ui"),
                                RequestPath = ""
                            });
                            
                            // Fallback to index.html for SPA
                            app.MapFallback(async context =>
                            {
                                context.Response.ContentType = "text/html";
                                var indexStream = assembly.GetManifestResourceStream("NovaBackup.Desktop.web-ui.index.html");
                                if (indexStream != null)
                                {
                                    using var reader = new StreamReader(indexStream);
                                    var content = await reader.ReadToEndAsync();
                                    await context.Response.WriteAsync(content);
                                }
                            });
                        });
                    })
                    .Build();

                await _webHost.StartAsync();
                return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Failed to start web server: {ex.Message}");
                return false;
            }
        }

        public async Task StopWebServer()
        {
            if (_webHost != null)
            {
                await _webHost.StopAsync();
                _webHost.Dispose();
                _webHost = null;
            }
        }

        public async Task<bool> CheckConnection()
        {
            try
            {
                var response = await _httpClient.GetAsync("/api/status");
                return response.IsSuccessStatusCode;
            }
            catch
            {
                return false;
            }
        }

        public async Task<BackupStatus> GetStatus()
        {
            try
            {
                var response = await _httpClient.GetStringAsync("/api/status");
                return JsonConvert.DeserializeObject<BackupStatus>(response);
            }
            catch
            {
                return new BackupStatus { Status = "Unknown" };
            }
        }

        public async Task<List<BackupJob>> GetBackups()
        {
            try
            {
                var response = await _httpClient.GetStringAsync("/api/backups");
                return JsonConvert.DeserializeObject<List<BackupJob>>(response);
            }
            catch
            {
                return new List<BackupJob>();
            }
        }

        public async Task<bool> CreateBackup(BackupJob backup)
        {
            try
            {
                var json = JsonConvert.SerializeObject(backup);
                var content = new StringContent(json, Encoding.UTF8, "application/json");
                var response = await _httpClient.PostAsync("/api/backups", content);
                return response.IsSuccessStatusCode;
            }
            catch
            {
                return false;
            }
        }

        public async Task<List<BackupSchedule>> GetSchedules()
        {
            try
            {
                var response = await _httpClient.GetStringAsync("/api/schedules");
                return JsonConvert.DeserializeObject<List<BackupSchedule>>(response);
            }
            catch
            {
                return new List<BackupSchedule>();
            }
        }

        public async Task<bool> CreateSchedule(BackupSchedule schedule)
        {
            try
            {
                var json = JsonConvert.SerializeObject(schedule);
                var content = new StringContent(json, Encoding.UTF8, "application/json");
                var response = await _httpClient.PostAsync("/api/schedules", content);
                return response.IsSuccessStatusCode;
            }
            catch
            {
                return false;
            }
        }

        public async Task<StorageInfo> GetStorage()
        {
            try
            {
                var response = await _httpClient.GetStringAsync("/api/storage");
                return JsonConvert.DeserializeObject<StorageInfo>(response);
            }
            catch
            {
                return new StorageInfo();
            }
        }

        public async Task<BackupReport> GetReports(DateTime from, DateTime to)
        {
            try
            {
                var url = $"/api/reports?from={from:yyyy-MM-dd}&to={to:yyyy-MM-dd}";
                var response = await _httpClient.GetStringAsync(url);
                return JsonConvert.DeserializeObject<BackupReport>(response);
            }
            catch
            {
                return new BackupReport();
            }
        }

        // API Endpoints
        private static async Task GetStatus(HttpContext context)
        {
            var status = new BackupStatus
            {
                Status = "Running",
                CurrentStep = "Monitoring",
                Progress = 0,
                LastBackup = DateTime.Now.AddHours(-2),
                NextBackup = DateTime.Now.AddHours(2),
                ActiveJob = "Daily Backup"
            };
            
            context.Response.ContentType = "application/json";
            await context.Response.WriteAsync(JsonConvert.SerializeObject(status));
        }

        private static async Task GetBackups(HttpContext context)
        {
            var backups = new List<BackupJob>
            {
                new BackupJob
                {
                    Id = "1",
                    Name = "Daily Backup",
                    Status = "Completed",
                    LastRun = DateTime.Now.AddHours(-2),
                    DataSize = 1024 * 1024 * 1024 // 1GB
                },
                new BackupJob
                {
                    Id = "2",
                    Name = "Weekly Backup",
                    Status = "Running",
                    LastRun = DateTime.Now.AddDays(-7),
                    DataSize = 5 * 1024 * 1024 * 1024 // 5GB
                }
            };
            
            context.Response.ContentType = "application/json";
            await context.Response.WriteAsync(JsonConvert.SerializeObject(backups));
        }

        private static async Task CreateBackup(HttpContext context)
        {
            using var reader = new StreamReader(context.Request.Body);
            var json = await reader.ReadToEndAsync();
            var backup = JsonConvert.DeserializeObject<BackupJob>(json);
            
            // Simulate creating backup
            backup.Id = Guid.NewGuid().ToString();
            backup.CreatedAt = DateTime.Now;
            backup.Status = "Created";
            
            context.Response.ContentType = "application/json";
            context.Response.StatusCode = 201;
            await context.Response.WriteAsync(JsonConvert.SerializeObject(backup));
        }

        private static async Task GetSchedules(HttpContext context)
        {
            var schedules = new List<BackupSchedule>
            {
                new BackupSchedule
                {
                    Id = "1",
                    Name = "Daily Schedule",
                    Type = "daily",
                    Time = new ScheduleTime { Hour = 2, Minute = 0 },
                    Enabled = true,
                    LastRun = DateTime.Now.AddHours(-2),
                    NextRun = DateTime.Now.AddHours(22)
                },
                new BackupSchedule
                {
                    Id = "2",
                    Name = "Weekly Schedule",
                    Type = "weekly",
                    DayOfWeek = 0, // Sunday
                    Time = new ScheduleTime { Hour = 3, Minute = 0 },
                    Enabled = true,
                    LastRun = DateTime.Now.AddDays(-7),
                    NextRun = DateTime.Now.AddDays(7 - (int)DateTime.Now.DayOfWeek)
                }
            };
            
            context.Response.ContentType = "application/json";
            await context.Response.WriteAsync(JsonConvert.SerializeObject(schedules));
        }

        private static async Task CreateSchedule(HttpContext context)
        {
            using var reader = new StreamReader(context.Request.Body);
            var json = await reader.ReadToEndAsync();
            var schedule = JsonConvert.DeserializeObject<BackupSchedule>(json);
            
            // Simulate creating schedule
            schedule.Id = Guid.NewGuid().ToString();
            schedule.CreatedAt = DateTime.Now;
            schedule.Enabled = true;
            
            context.Response.ContentType = "application/json";
            context.Response.StatusCode = 201;
            await context.Response.WriteAsync(JsonConvert.SerializeObject(schedule));
        }

        private static async Task GetStorage(HttpContext context)
        {
            var storage = new StorageInfo
            {
                TotalSpace = 1000L * 1024 * 1024 * 1024, // 1TB
                UsedSpace = 500L * 1024 * 1024 * 1024,  // 500GB
                AvailableSpace = 500L * 1024 * 1024 * 1024, // 500GB
                Drives = new List<DriveInfo>
                {
                    new DriveInfo
                    {
                        Name = "C:",
                        TotalSize = 500L * 1024 * 1024 * 1024,
                        UsedSpace = 300L * 1024 * 1024 * 1024,
                        AvailableSpace = 200L * 1024 * 1024 * 1024
                    },
                    new DriveInfo
                    {
                        Name = "D:",
                        TotalSize = 500L * 1024 * 1024 * 1024,
                        UsedSpace = 200L * 1024 * 1024 * 1024,
                        AvailableSpace = 300L * 1024 * 1024 * 1024
                    }
                }
            };
            
            context.Response.ContentType = "application/json";
            await context.Response.WriteAsync(JsonConvert.SerializeObject(storage));
        }

        private static async Task GetReports(HttpContext context)
        {
            var from = DateTime.Parse(context.Request.Query["from"]);
            var to = DateTime.Parse(context.Request.Query["to"]);
            
            var report = new BackupReport
            {
                GeneratedAt = DateTime.Now,
                PeriodFrom = from,
                PeriodTo = to,
                TotalBackups = 30,
                SuccessfulBackups = 28,
                FailedBackups = 2,
                TotalDataSize = 100L * 1024 * 1024 * 1024, // 100GB
                AverageBackupTime = TimeSpan.FromMinutes(15),
                Results = new List<BackupJobResult>
                {
                    new BackupJobResult
                    {
                        JobId = "1",
                        JobName = "Daily Backup",
                        StartTime = DateTime.Now.AddHours(-2),
                        EndTime = DateTime.Now.AddHours(-1).AddMinutes(-45),
                        Duration = TimeSpan.FromMinutes(15),
                        Status = "Success",
                        DataSize = 1024 * 1024 * 1024
                    }
                }
            };
            
            context.Response.ContentType = "application/json";
            await context.Response.WriteAsync(JsonConvert.SerializeObject(report));
        }

        public void Dispose()
        {
            _httpClient?.Dispose();
            StopWebServer().Wait();
        }

        private bool IsValidAuth(string authHeader)
        {
            try
            {
                // Basic authentication: "Basic base64(username:password)"
                if (!authHeader.StartsWith("Basic "))
                    return false;

                var token = authHeader.Substring("Basic ".Length);
                var credentials = Convert.FromBase64String(token);
                var pair = System.Text.Encoding.UTF8.GetString(credentials).Split(':');

                if (pair.Length != 2)
                    return false;

                var username = pair[0];
                var password = pair[1];

                // Check against configured credentials (for demo, use admin/admin)
                return username == "admin" && password == "admin";
            }
            catch
            {
                return false;
            }
        }
    }
}
