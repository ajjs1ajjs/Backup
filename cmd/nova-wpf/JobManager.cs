using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.IO;
using System.Text.Json;
using System.Linq;

namespace NovaBackup.WPF
{
    public class BackupJob
    {
        public Guid Id { get; set; }
        public string Name { get; set; }
        public string Description { get; set; }
        public string Type { get; set; }
        public List<string> Sources { get; set; }
        public string Destination { get; set; }
        public bool Compression { get; set; }
        public bool Encryption { get; set; }
        public string ScheduleType { get; set; }
        public string ScheduleTime { get; set; }
        public List<string> ScheduleDays { get; set; }
        public bool Enabled { get; set; }
        public DateTime? LastRun { get; set; }
        public DateTime? NextRun { get; set; }
        public string Status { get; set; }
        public string StatusIcon { get; set; }

        public BackupJob()
        {
            Id = Guid.NewGuid();
            Sources = new List<string>();
            ScheduleDays = new List<string>();
            Enabled = true;
            Status = "Active";
            StatusIcon = "✅";
        }
    }

    public class JobManager
    {
        private static readonly string JobsFile = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.CommonApplicationData),
            "NovaBackup", "Config", "jobs.json");

        public static ObservableCollection<BackupJob> LoadJobs()
        {
            try
            {
                if (File.Exists(JobsFile))
                {
                    var json = File.ReadAllText(JobsFile);
                    var jobs = JsonSerializer.Deserialize<List<BackupJob>>(json);
                    return new ObservableCollection<BackupJob>(jobs ?? new List<BackupJob>());
                }
            }
            catch { }

            // Default jobs for demo
            return new ObservableCollection<BackupJob>
            {
                new BackupJob
                {
                    Name = "Daily Documents Backup",
                    Type = "File Backup",
                    Sources = new List<string> { "C:\\Users\\Documents" },
                    Destination = "C:\\ProgramData\\NovaBackup\\Backups",
                    ScheduleType = "Daily",
                    ScheduleTime = "02:00",
                    LastRun = DateTime.Now.AddDays(-1).Date.AddHours(2),
                    NextRun = DateTime.Now.Date.AddDays(1).AddHours(2),
                    Compression = true,
                    Encryption = false
                },
                new BackupJob
                {
                    Name = "Weekly System Backup",
                    Type = "System",
                    Sources = new List<string> { "C:\\Windows", "C:\\Program Files" },
                    Destination = "D:\\Backups",
                    ScheduleType = "Weekly",
                    ScheduleTime = "22:00",
                    ScheduleDays = new List<string> { "Sunday" },
                    LastRun = DateTime.Now.AddDays(-7).Date.AddHours(22),
                    NextRun = DateTime.Now.AddDays(7 - (int)DateTime.Now.DayOfWeek).Date.AddHours(22),
                    Compression = true,
                    Encryption = true
                }
            };
        }

        public static void SaveJobs(ObservableCollection<BackupJob> jobs)
        {
            try
            {
                var dir = Path.GetDirectoryName(JobsFile);
                if (!Directory.Exists(dir))
                    Directory.CreateDirectory(dir);

                var json = JsonSerializer.Serialize(jobs, new JsonSerializerOptions { WriteIndented = true });
                File.WriteAllText(JobsFile, json);
            }
            catch (Exception ex)
            {
                System.Windows.MessageBox.Show($"Failed to save jobs: {ex.Message}");
            }
        }

        public static void AddJob(BackupJob job)
        {
            var jobs = LoadJobs();
            jobs.Add(job);
            SaveJobs(jobs);
        }

        public static void UpdateJob(BackupJob job)
        {
            var jobs = LoadJobs();
            var existing = jobs.FirstOrDefault(j => j.Id == job.Id);
            if (existing != null)
            {
                existing.Name = job.Name;
                existing.Description = job.Description;
                existing.Type = job.Type;
                existing.Sources = job.Sources;
                existing.Destination = job.Destination;
                existing.Compression = job.Compression;
                existing.Encryption = job.Encryption;
                existing.ScheduleType = job.ScheduleType;
                existing.ScheduleTime = job.ScheduleTime;
                existing.ScheduleDays = job.ScheduleDays;
            }
            SaveJobs(jobs);
        }

        public static void DeleteJob(Guid id)
        {
            var jobs = LoadJobs();
            var job = jobs.FirstOrDefault(j => j.Id == id);
            if (job != null)
            {
                jobs.Remove(job);
                SaveJobs(jobs);
            }
        }

        public static void ToggleJob(Guid id, bool enabled)
        {
            var jobs = LoadJobs();
            var job = jobs.FirstOrDefault(j => j.Id == id);
            if (job != null)
            {
                job.Enabled = enabled;
                job.Status = enabled ? "Active" : "Disabled";
                job.StatusIcon = enabled ? "✅" : "⏸️";
                SaveJobs(jobs);
            }
        }
    }
}
