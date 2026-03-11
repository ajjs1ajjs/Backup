using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.ServiceProcess;
using System.Threading.Tasks;
using System.Timers;
using Newtonsoft.Json;

namespace NovaBackup.Desktop.Services
{
    public class NovaBackupService
    {
        private readonly string _configPath;
        private readonly string _logPath;
        private readonly Timer _monitoringTimer;
        private readonly List<BackupJob> _backupJobs;
        private readonly List<BackupSchedule> _schedules;
        
        private bool _isRunning;
        private BackupStatus _currentStatus;

        public NovaBackupService()
        {
            _configPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), 
                "NovaBackup", "config");
            _logPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), 
                "NovaBackup", "logs");
            
            _backupJobs = new List<BackupJob>();
            _schedules = new List<BackupSchedule>();
            
            _monitoringTimer = new Timer(60000); // Check every minute
            _monitoringTimer.Elapsed += MonitorSchedules;
            
            _currentStatus = new BackupStatus
            {
                Status = "Stopped",
                LastBackup = null,
                NextBackup = null,
                Progress = 0
            };
            
            EnsureDirectoriesExist();
            LoadConfiguration();
        }

        public bool IsServiceRunning()
        {
            try
            {
                var services = ServiceController.GetServices();
                var novaService = services.FirstOrDefault(s => s.ServiceName == "NovaBackupService");
                return novaService?.Status == ServiceControllerStatus.Running;
            }
            catch
            {
                return false;
            }
        }

        public async Task<BackupStatus> GetBackupStatus()
        {
            return await Task.FromResult(_currentStatus);
        }

        public async Task<List<BackupJob>> GetBackupJobs()
        {
            return await Task.FromResult(_backupJobs);
        }

        public async Task<List<BackupSchedule>> GetSchedules()
        {
            return await Task.FromResult(_schedules);
        }

        public async Task<bool> StartBackup(string jobId)
        {
            var job = _backupJobs.FirstOrDefault(j => j.Id == jobId);
            if (job == null)
                return false;

            try
            {
                _currentStatus.Status = "Running";
                _currentStatus.Progress = 0;
                
                // Simulate backup process
                await Task.Run(() => SimulateBackup(job));
                
                return true;
            }
            catch (Exception ex)
            {
                LogError($"Failed to start backup {jobId}: {ex.Message}");
                return false;
            }
        }

        public async Task<bool> StopBackup(string jobId)
        {
            try
            {
                _currentStatus.Status = "Stopped";
                LogInfo($"Backup {jobId} stopped by user");
                return true;
            }
            catch (Exception ex)
            {
                LogError($"Failed to stop backup {jobId}: {ex.Message}");
                return false;
            }
        }

        public async Task<bool> CreateBackupJob(BackupJob job)
        {
            try
            {
                job.Id = Guid.NewGuid().ToString();
                job.CreatedAt = DateTime.Now;
                job.Status = "Created";
                
                _backupJobs.Add(job);
                await SaveConfiguration();
                
                LogInfo($"Created backup job: {job.Name}");
                return true;
            }
            catch (Exception ex)
            {
                LogError($"Failed to create backup job: {ex.Message}");
                return false;
            }
        }

        public async Task<bool> UpdateBackupJob(BackupJob job)
        {
            try
            {
                var existing = _backupJobs.FirstOrDefault(j => j.Id == job.Id);
                if (existing != null)
                {
                    existing.Name = job.Name;
                    existing.SourcePath = job.SourcePath;
                    existing.DestinationPath = job.DestinationPath;
                    existing.BackupType = job.BackupType;
                    existing.Schedule = job.Schedule;
                    existing.Enabled = job.Enabled;
                    existing.UpdatedAt = DateTime.Now;
                    
                    await SaveConfiguration();
                    LogInfo($"Updated backup job: {job.Name}");
                    return true;
                }
                return false;
            }
            catch (Exception ex)
            {
                LogError($"Failed to update backup job: {ex.Message}");
                return false;
            }
        }

        public async Task<bool> DeleteBackupJob(string jobId)
        {
            try
            {
                var job = _backupJobs.FirstOrDefault(j => j.Id == jobId);
                if (job != null)
                {
                    _backupJobs.Remove(job);
                    await SaveConfiguration();
                    LogInfo($"Deleted backup job: {job.Name}");
                    return true;
                }
                return false;
            }
            catch (Exception ex)
            {
                LogError($"Failed to delete backup job: {ex.Message}");
                return false;
            }
        }

        public async Task<bool> CreateSchedule(BackupSchedule schedule)
        {
            try
            {
                schedule.Id = Guid.NewGuid().ToString();
                schedule.CreatedAt = DateTime.Now;
                schedule.Enabled = true;
                
                _schedules.Add(schedule);
                await SaveConfiguration();
                
                LogInfo($"Created backup schedule: {schedule.Name}");
                return true;
            }
            catch (Exception ex)
            {
                LogError($"Failed to create backup schedule: {ex.Message}");
                return false;
            }
        }

        public async Task<StorageInfo> GetStorageInfo()
        {
            try
            {
                // Simulate storage info
                var drives = DriveInfo.GetDrives()
                    .Where(d => d.IsReady && d.DriveType == DriveType.Fixed)
                    .Select(d => new DriveInfo
                    {
                        Name = d.Name,
                        TotalSize = d.TotalSize,
                        AvailableSpace = d.AvailableFreeSpace,
                        UsedSpace = d.TotalSize - d.AvailableFreeSpace
                    })
                    .ToList();

                var totalSpace = drives.Sum(d => d.TotalSize);
                var usedSpace = drives.Sum(d => d.UsedSpace);
                var availableSpace = drives.Sum(d => d.AvailableSpace);

                return await Task.FromResult(new StorageInfo
                {
                    TotalSpace = totalSpace,
                    UsedSpace = usedSpace,
                    AvailableSpace = availableSpace,
                    Drives = drives
                });
            }
            catch (Exception ex)
            {
                LogError($"Failed to get storage info: {ex.Message}");
                return new StorageInfo();
            }
        }

        public async Task<BackupReport> GenerateReport(DateTime from, DateTime to)
        {
            try
            {
                // Simulate report generation
                var report = new BackupReport
                {
                    GeneratedAt = DateTime.Now,
                    PeriodFrom = from,
                    PeriodTo = to,
                    TotalBackups = _backupJobs.Count,
                    SuccessfulBackups = _backupJobs.Count(j => j.Status == "Completed"),
                    FailedBackups = _backupJobs.Count(j => j.Status == "Failed"),
                    TotalDataSize = _backupJobs.Sum(j => j.DataSize),
                    AverageBackupTime = TimeSpan.FromMinutes(15)
                };

                return await Task.FromResult(report);
            }
            catch (Exception ex)
            {
                LogError($"Failed to generate report: {ex.Message}");
                return new BackupReport();
            }
        }

        public void StartMonitoring()
        {
            _isRunning = true;
            _monitoringTimer.Start();
            LogInfo("Backup service monitoring started");
        }

        public void StopMonitoring()
        {
            _isRunning = false;
            _monitoringTimer.Stop();
            LogInfo("Backup service monitoring stopped");
        }

        private void MonitorSchedules(object sender, ElapsedEventArgs e)
        {
            if (!_isRunning) return;

            try
            {
                var now = DateTime.Now;
                var dueSchedules = _schedules.Where(s => s.Enabled && ShouldRun(s, now)).ToList();

                foreach (var schedule in dueSchedules)
                {
                    Task.Run(async () => await ExecuteScheduledBackup(schedule));
                }
            }
            catch (Exception ex)
            {
                LogError($"Error in schedule monitoring: {ex.Message}");
            }
        }

        private bool ShouldRun(BackupSchedule schedule, DateTime now)
        {
            // Simple schedule check - in real implementation, use proper cron parsing
            if (schedule.Type == "daily")
            {
                return now.Hour == schedule.Time.Hour && now.Minute == schedule.Time.Minute;
            }
            else if (schedule.Type == "weekly")
            {
                return now.DayOfWeek == (DayOfWeek)schedule.DayOfWeek && 
                       now.Hour == schedule.Time.Hour && 
                       now.Minute == schedule.Time.Minute;
            }
            else if (schedule.Type == "monthly")
            {
                return now.Day == schedule.Day && 
                       now.Hour == schedule.Time.Hour && 
                       now.Minute == schedule.Time.Minute;
            }

            return false;
        }

        private async Task ExecuteScheduledBackup(BackupSchedule schedule)
        {
            try
            {
                var job = _backupJobs.FirstOrDefault(j => j.Id == schedule.BackupJobId);
                if (job == null || !job.Enabled)
                    return;

                LogInfo($"Executing scheduled backup: {schedule.Name}");
                await StartBackup(job.Id);
            }
            catch (Exception ex)
            {
                LogError($"Failed to execute scheduled backup {schedule.Name}: {ex.Message}");
            }
        }

        private void SimulateBackup(BackupJob job)
        {
            try
            {
                var steps = new[]
                {
                    "Validating source path",
                    "Creating backup snapshot",
                    "Compressing data",
                    "Encrypting backup",
                    "Transferring to destination",
                    "Verifying backup integrity"
                };

                for (int i = 0; i < steps.Length; i++)
                {
                    _currentStatus.Progress = (i + 1) * 100 / steps.Length;
                    _currentStatus.CurrentStep = steps[i];
                    Thread.Sleep(2000); // Simulate work
                }

                _currentStatus.Status = "Completed";
                _currentStatus.Progress = 100;
                _currentStatus.LastBackup = DateTime.Now;
                job.LastRun = DateTime.Now;
                job.Status = "Completed";

                LogInfo($"Backup completed: {job.Name}");
            }
            catch (Exception ex)
            {
                _currentStatus.Status = "Failed";
                job.Status = "Failed";
                LogError($"Backup failed: {job.Name} - {ex.Message}");
            }
        }

        private void EnsureDirectoriesExist()
        {
            Directory.CreateDirectory(_configPath);
            Directory.CreateDirectory(_logPath);
        }

        private void LoadConfiguration()
        {
            try
            {
                var configFile = Path.Combine(_configPath, "backup-config.json");
                if (File.Exists(configFile))
                {
                    var config = JsonConvert.DeserializeObject<BackupConfiguration>(File.ReadAllText(configFile));
                    _backupJobs.AddRange(config.BackupJobs ?? new List<BackupJob>());
                    _schedules.AddRange(config.Schedules ?? new List<BackupSchedule>());
                }
            }
            catch (Exception ex)
            {
                LogError($"Failed to load configuration: {ex.Message}");
            }
        }

        private async Task SaveConfiguration()
        {
            try
            {
                var config = new BackupConfiguration
                {
                    BackupJobs = _backupJobs,
                    Schedules = _schedules,
                    UpdatedAt = DateTime.Now
                };

                var configFile = Path.Combine(_configPath, "backup-config.json");
                await File.WriteAllTextAsync(configFile, JsonConvert.SerializeObject(config, Formatting.Indented));
            }
            catch (Exception ex)
            {
                LogError($"Failed to save configuration: {ex.Message}");
            }
        }

        private void LogInfo(string message)
        {
            var logFile = Path.Combine(_logPath, $"backup-{DateTime.Now:yyyy-MM-dd}.log");
            var logEntry = $"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] [INFO] {message}{Environment.NewLine}";
            File.AppendAllText(logFile, logEntry);
        }

        private void LogError(string message)
        {
            var logFile = Path.Combine(_logPath, $"backup-{DateTime.Now:yyyy-MM-dd}.log");
            var logEntry = $"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] [ERROR] {message}{Environment.NewLine}";
            File.AppendAllText(logFile, logEntry);
        }
    }

    // Data models
    public class BackupJob
    {
        public string Id { get; set; }
        public string Name { get; set; }
        public string Description { get; set; }
        public string SourcePath { get; set; }
        public string DestinationPath { get; set; }
        public string BackupType { get; set; }
        public string Schedule { get; set; }
        public bool Enabled { get; set; }
        public string Status { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime UpdatedAt { get; set; }
        public DateTime? LastRun { get; set; }
        public long DataSize { get; set; }
        public Dictionary<string, object> Settings { get; set; }
    }

    public class BackupSchedule
    {
        public string Id { get; set; }
        public string Name { get; set; }
        public string Description { get; set; }
        public string BackupJobId { get; set; }
        public string Type { get; set; }
        public ScheduleTime Time { get; set; }
        public int DayOfWeek { get; set; }
        public int Day { get; set; }
        public bool Enabled { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime? LastRun { get; set; }
        public DateTime? NextRun { get; set; }
    }

    public class ScheduleTime
    {
        public int Hour { get; set; }
        public int Minute { get; set; }
    }

    public class BackupStatus
    {
        public string Status { get; set; }
        public string CurrentStep { get; set; }
        public int Progress { get; set; }
        public DateTime? LastBackup { get; set; }
        public DateTime? NextBackup { get; set; }
        public string ActiveJob { get; set; }
    }

    public class StorageInfo
    {
        public long TotalSpace { get; set; }
        public long UsedSpace { get; set; }
        public long AvailableSpace { get; set; }
        public List<DriveInfo> Drives { get; set; }
    }

    public class DriveInfo
    {
        public string Name { get; set; }
        public long TotalSize { get; set; }
        public long AvailableSpace { get; set; }
        public long UsedSpace { get; set; }
    }

    public class BackupReport
    {
        public DateTime GeneratedAt { get; set; }
        public DateTime PeriodFrom { get; set; }
        public DateTime PeriodTo { get; set; }
        public int TotalBackups { get; set; }
        public int SuccessfulBackups { get; set; }
        public int FailedBackups { get; set; }
        public long TotalDataSize { get; set; }
        public TimeSpan AverageBackupTime { get; set; }
        public List<BackupJobResult> Results { get; set; }
    }

    public class BackupJobResult
    {
        public string JobId { get; set; }
        public string JobName { get; set; }
        public DateTime StartTime { get; set; }
        public DateTime EndTime { get; set; }
        public TimeSpan Duration { get; set; }
        public string Status { get; set; }
        public long DataSize { get; set; }
        public string ErrorMessage { get; set; }
    }

    public class BackupConfiguration
    {
        public List<BackupJob> BackupJobs { get; set; }
        public List<BackupSchedule> Schedules { get; set; }
        public DateTime UpdatedAt { get; set; }
    }
}
