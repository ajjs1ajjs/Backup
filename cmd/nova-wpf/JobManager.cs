using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.IO;
using System.Text.Json;

namespace NovaBackup.WPF
{
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

            return new ObservableCollection<BackupJob>();
        }

        public static void SaveJobs(ObservableCollection<BackupJob> jobs)
        {
            try
            {
                var dir = Path.GetDirectoryName(JobsFile);
                if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
                    Directory.CreateDirectory(dir);

                var json = JsonSerializer.Serialize(jobs, new JsonSerializerOptions { WriteIndented = true });
                File.WriteAllText(JobsFile, json);
            }
            catch { }
        }

        public static void AddJob(BackupJob job)
        {
            var jobs = LoadJobs();
            jobs.Add(job);
            SaveJobs(jobs);
        }
    }
}
