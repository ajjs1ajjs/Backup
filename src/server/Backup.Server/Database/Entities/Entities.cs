using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace Backup.Server.Database.Entities;

[Table("agents")]
public class Agent
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string AgentId { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string Hostname { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string OsType { get; set; } = string.Empty;

    [MaxLength(32)]
    public string? AgentVersion { get; set; }

    [Required]
    [MaxLength(32)]
    public string AgentType { get; set; } = string.Empty;

    public string Status { get; set; } = "idle";

    [MaxLength(64)]
    public string? IpAddress { get; set; }

    public DateTime? LastHeartbeat { get; set; }

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public string Capabilities { get; set; } = "[]";
}

[Table("virtual_machines")]
public class VirtualMachine
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string VmId { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string Name { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string HypervisorType { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string HypervisorHost { get; set; } = string.Empty;

    [MaxLength(64)]
    public string? IpAddress { get; set; }

    [MaxLength(64)]
    public string? OsType { get; set; }

    public long? MemoryMb { get; set; }

    public int? CpuCores { get; set; }

    public string Disks { get; set; } = "[]";

    public string Tags { get; set; } = "{}";

    [MaxLength(32)]
    public string Status { get; set; } = "running";

    public DateTime? LastBackupAt { get; set; }

    [MaxLength(64)]
    public string? LastBackupId { get; set; }

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}

[Table("repositories")]
public class Repository
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string RepositoryId { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string Name { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string Type { get; set; } = string.Empty;

    [Required]
    [MaxLength(1024)]
    public string Path { get; set; } = string.Empty;

    public string Status { get; set; } = "online";

    public long? CapacityBytes { get; set; }

    public long UsedBytes { get; set; }

    public string? Credentials { get; set; }

    public string Options { get; set; } = "{}";

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? LastUsedAt { get; set; }
}

[Table("jobs")]
public class Job
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string JobId { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string Name { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string JobType { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string SourceId { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string SourceType { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string DestinationId { get; set; }

    public string? Schedule { get; set; }

    public string Options { get; set; } = "{}";

    public bool Enabled { get; set; } = true;

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? LastRun { get; set; }

    public DateTime? NextRun { get; set; }
}

[Table("backup_points")]
public class BackupPoint
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string BackupId { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string JobId { get; set; } = string.Empty;

    [MaxLength(64)]
    public string? VmId { get; set; }

    [Required]
    [MaxLength(32)]
    public string BackupType { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string RepositoryId { get; set; } = string.Empty;

    [MaxLength(1024)]
    public string? FilePath { get; set; }

    public long SizeBytes { get; set; }

    public long OriginalSizeBytes { get; set; }

    [MaxLength(128)]
    public string? Checksum { get; set; }

    public bool IsSynthetic { get; set; }

    [MaxLength(64)]
    public string? ParentBackupId { get; set; }

    public string Metadata { get; set; } = "{}";

    public string Status { get; set; } = "in_progress";

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? CompletedAt { get; set; }

    [ForeignKey("RepositoryId")]
    public Repository? Repository { get; set; }
}

[Table("restores")]
public class Restore
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string RestoreId { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string BackupId { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string RestoreType { get; set; } = string.Empty;

    [MaxLength(1024)]
    public string? DestinationPath { get; set; }

    [MaxLength(255)]
    public string? TargetHost { get; set; }

    public string Options { get; set; } = "{}";

    public string Status { get; set; } = "pending";

    public long BytesRestored { get; set; }

    public long TotalBytes { get; set; }

    public string? ErrorMessage { get; set; }

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? CompletedAt { get; set; }

    public DateTime? StartedAt { get; set; }
}

[Table("job_run_history")]
public class JobRunHistory
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string RunId { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string JobId { get; set; } = string.Empty;

    public DateTime StartTime { get; set; }

    public DateTime? EndTime { get; set; }

    public string Status { get; set; } = "pending";

    public long BytesProcessed { get; set; }

    public long FilesProcessed { get; set; }

    public double SpeedMbps { get; set; }

    public string? ErrorMessage { get; set; }
}

[Table("users")]
public class User
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(64)]
    public string UserId { get; set; } = string.Empty;

    [Required]
    [MaxLength(64)]
    public string Username { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string Email { get; set; } = string.Empty;

    [Required]
    [MaxLength(255)]
    public string PasswordHash { get; set; } = string.Empty;

    [MaxLength(32)]
    public string Role { get; set; } = "user";

    public bool IsActive { get; set; } = true;

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime? LastLoginAt { get; set; }

    public bool MustChangePassword { get; set; } = false;

    [MaxLength(64)]
    public string? TwoFactorSecret { get; set; }
}

[Table("audit_logs")]
public class AuditLog
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [MaxLength(64)]
    public string? UserId { get; set; }

    [Required]
    [MaxLength(64)]
    public string Action { get; set; } = string.Empty;

    [Required]
    [MaxLength(32)]
    public string EntityType { get; set; } = string.Empty;

    [MaxLength(64)]
    public string? EntityId { get; set; }

    public string Details { get; set; } = "{}";

    [MaxLength(64)]
    public string? IpAddress { get; set; }

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
}

[Table("settings")]
public class Setting
{
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    public long Id { get; set; }

    [Required]
    [MaxLength(128)]
    public string Key { get; set; } = string.Empty;

    [Required]
    public string Value { get; set; } = string.Empty;

    [MaxLength(32)]
    public string Type { get; set; } = "string";

    public string? Description { get; set; }

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
