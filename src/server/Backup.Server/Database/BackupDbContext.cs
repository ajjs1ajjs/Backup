using Microsoft.EntityFrameworkCore;
using Backup.Server.Database.Entities;

namespace Backup.Server.Database;

public class BackupDbContext : DbContext
{
    public BackupDbContext(DbContextOptions<BackupDbContext> options) : base(options) { }

    public DbSet<Agent> Agents => Set<Agent>();
    public DbSet<VirtualMachine> VirtualMachines => Set<VirtualMachine>();
    public DbSet<Repository> Repositories => Set<Repository>();
    public DbSet<Job> Jobs => Set<Job>();
    public DbSet<BackupPoint> BackupPoints => Set<BackupPoint>();
    public DbSet<Restore> Restores => Set<Restore>();
    public DbSet<JobRunHistory> JobRunHistory => Set<JobRunHistory>();
    public DbSet<User> Users => Set<User>();
    public DbSet<AuditLog> AuditLogs => Set<AuditLog>();
    public DbSet<Setting> Settings => Set<Setting>();
    public DbSet<Hypervisor> Hypervisors => Set<Hypervisor>();
    public DbSet<Chunk> Chunks => Set<Chunk>();
protected override void OnModelCreating(ModelBuilder modelBuilder)
{
    base.OnModelCreating(modelBuilder);

    // --- ENUM Converters ---
    var enumConverter = new Microsoft.EntityFrameworkCore.Storage.ValueConversion.EnumToStringConverter<JobType>();

    // --- Agent Configuration ---
    modelBuilder.Entity<Agent>(entity =>
    {
        entity.HasIndex(e => e.AgentId).IsUnique();
        entity.HasIndex(e => e.Status);
    });

    // --- Virtual Machine Configuration ---
    modelBuilder.Entity<VirtualMachine>(entity =>
    {
        entity.HasIndex(e => e.VmId).IsUnique();
        entity.HasIndex(e => new { e.HypervisorType, e.HypervisorHost });
    });

    // --- Repository Configuration ---
    modelBuilder.Entity<Repository>(entity =>
    {
        entity.HasIndex(e => e.RepositoryId).IsUnique();
    });

    // --- Job Configuration ---
    modelBuilder.Entity<Job>(entity =>
    {
        entity.HasIndex(e => e.JobId).IsUnique();
        entity.HasIndex(e => e.Enabled);
        entity.HasIndex(e => e.NextRun);
    });

    // --- Backup Point Configuration ---
    modelBuilder.Entity<BackupPoint>(entity =>
    {
        entity.HasIndex(e => e.BackupId).IsUnique();
        entity.HasIndex(e => e.JobId);
        entity.HasIndex(e => e.VmId);
        entity.HasIndex(e => e.RepositoryId);
    });

    // --- Restore Configuration ---
    modelBuilder.Entity<Restore>(entity =>
    {
        entity.HasIndex(e => e.RestoreId).IsUnique();
        entity.HasIndex(e => e.BackupId);
    });

    // --- User Configuration ---
    modelBuilder.Entity<User>(entity =>
    {
        entity.HasIndex(e => e.UserId).IsUnique();
        entity.HasIndex(e => e.Username).IsUnique();
        entity.HasIndex(e => e.Email).IsUnique();
    });
}
}
        modelBuilder.Entity<Repository>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.RepositoryId).IsUnique();
        });

        modelBuilder.Entity<Job>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.JobId).IsUnique();
        });

        modelBuilder.Entity<BackupPoint>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.BackupId).IsUnique();
        });

        modelBuilder.Entity<Restore>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.RestoreId).IsUnique();
        });

        modelBuilder.Entity<JobRunHistory>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.RunId).IsUnique();
        });

        modelBuilder.Entity<User>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.UserId).IsUnique();
            entity.HasIndex(e => e.Username).IsUnique();
            entity.HasIndex(e => e.Email).IsUnique();
        });

        modelBuilder.Entity<AuditLog>(entity =>
        {
            entity.HasKey(e => e.Id);
        });

        modelBuilder.Entity<Setting>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.Key).IsUnique();
        });

        modelBuilder.Entity<Hypervisor>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.HypervisorId).IsUnique();
        });
    }
}
