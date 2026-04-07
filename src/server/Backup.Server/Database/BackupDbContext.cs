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

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        modelBuilder.Entity<Agent>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.AgentId).IsUnique();
        });

        modelBuilder.Entity<VirtualMachine>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.VmId).IsUnique();
        });

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
