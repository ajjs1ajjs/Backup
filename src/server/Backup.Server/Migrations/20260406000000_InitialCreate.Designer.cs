using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Migrations;
using Microsoft.EntityFrameworkCore.Storage.ValueConversion;
using Backup.Server.Database;

#nullable disable

namespace Backup.Server.Migrations
{
    [DbContext(typeof(BackupDbContext))]
    [Migration("20260406000000_InitialCreate")]
    partial class InitialCreate
    {
        protected override void BuildTargetModel(ModelBuilder modelBuilder)
        {
            modelBuilder
                .HasAnnotation("ProductVersion", "8.0.0");

            modelBuilder.Entity("Backup.Server.Database.Entities.Agent", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("AgentId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("AgentType").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("AgentVersion").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("Capabilities").HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<string>("Hostname").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("IpAddress").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<DateTime?>("LastHeartbeat").HasColumnType("TEXT");
                b.Property<string>("OsType").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Status").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("AgentId").IsUnique();
                b.ToTable("agents");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.User", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("UserId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Username").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Email").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("PasswordHash").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("Role").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<bool>("IsActive").HasColumnType("INTEGER");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("LastLoginAt").HasColumnType("TEXT");
                b.Property<bool>("MustChangePassword").HasColumnType("INTEGER");
                b.Property<string>("TwoFactorSecret").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("PasswordResetToken").HasMaxLength(128).HasColumnType("TEXT");
                b.Property<DateTime?>("PasswordResetTokenExpiry").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("Username").IsUnique();
                b.ToTable("users");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.VirtualMachine", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("VmId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Name").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("HypervisorType").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("HypervisorHost").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("IpAddress").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("OsType").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<long?>("MemoryMb").HasColumnType("INTEGER");
                b.Property<int?>("CpuCores").HasColumnType("INTEGER");
                b.Property<string>("Disks").HasColumnType("TEXT");
                b.Property<string>("Tags").HasColumnType("TEXT");
                b.Property<string>("Status").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<DateTime?>("LastBackupAt").HasColumnType("TEXT");
                b.Property<string>("LastBackupId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("VmId").IsUnique();
                b.ToTable("virtual_machines");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Repository", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("RepositoryId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Name").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<int>("Type").HasColumnType("INTEGER");
                b.Property<string>("Path").HasMaxLength(1024).HasColumnType("TEXT");
                b.Property<string>("Status").HasColumnType("TEXT");
                b.Property<long?>("CapacityBytes").HasColumnType("INTEGER");
                b.Property<long>("UsedBytes").HasColumnType("INTEGER");
                b.Property<string>("Credentials").HasColumnType("TEXT");
                b.Property<string>("Options").HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("LastUsedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("RepositoryId").IsUnique();
                b.ToTable("repositories");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Job", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("JobId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Name").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<int>("JobType").HasColumnType("INTEGER");
                b.Property<string>("SourceId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("SourceType").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("DestinationId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("AgentId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Schedule").HasColumnType("TEXT");
                b.Property<string>("Options").HasColumnType("TEXT");
                b.Property<bool>("Enabled").HasColumnType("INTEGER");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("LastRun").HasColumnType("TEXT");
                b.Property<DateTime?>("NextRun").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("JobId").IsUnique();
                b.ToTable("jobs");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.BackupPoint", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("BackupId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("JobId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("VmId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<int>("BackupType").HasColumnType("INTEGER");
                b.Property<string>("RepositoryId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("FilePath").HasMaxLength(1024).HasColumnType("TEXT");
                b.Property<long>("SizeBytes").HasColumnType("INTEGER");
                b.Property<long>("OriginalSizeBytes").HasColumnType("INTEGER");
                b.Property<string>("Checksum").HasMaxLength(128).HasColumnType("TEXT");
                b.Property<bool>("IsSynthetic").HasColumnType("INTEGER");
                b.Property<string>("ParentBackupId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Metadata").HasColumnType("TEXT");
                b.Property<int>("Status").HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("CompletedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("BackupId").IsUnique();
                b.HasIndex("JobId");
                b.ToTable("backup_points");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Restore", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("RestoreId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("BackupId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("RestoreType").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("DestinationPath").HasMaxLength(1024).HasColumnType("TEXT");
                b.Property<string>("TargetHost").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("Options").HasColumnType("TEXT");
                b.Property<string>("Status").HasColumnType("TEXT");
                b.Property<long>("BytesRestored").HasColumnType("INTEGER");
                b.Property<long>("TotalBytes").HasColumnType("INTEGER");
                b.Property<string>("ErrorMessage").HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("CompletedAt").HasColumnType("TEXT");
                b.Property<DateTime?>("StartedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("RestoreId").IsUnique();
                b.ToTable("restores");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Hypervisor", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("HypervisorId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Name").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("Type").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("Host").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<int>("Port").HasColumnType("INTEGER");
                b.Property<string>("Username").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("Password").HasMaxLength(255).HasColumnType("TEXT");
                b.Property<string>("Status").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<int>("VmCount").HasColumnType("INTEGER");
                b.Property<DateTime?>("LastConnectedAt").HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.HasIndex("HypervisorId").IsUnique();
                b.ToTable("hypervisors");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.AuditLog", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("UserId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Action").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("EntityType").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("EntityId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("Details").HasColumnType("TEXT");
                b.Property<string>("IpAddress").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.ToTable("audit_logs");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Setting", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("Key").HasMaxLength(128).HasColumnType("TEXT");
                b.Property<string>("Value").HasColumnType("TEXT");
                b.Property<string>("Type").HasMaxLength(32).HasColumnType("TEXT");
                b.Property<string>("Description").HasColumnType("TEXT");
                b.Property<DateTime>("UpdatedAt").HasColumnType("TEXT");
                b.HasKey("Id");
                b.ToTable("settings");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.JobRunHistory", b =>
            {
                b.Property<long>("Id").ValueGeneratedOnAdd().HasColumnType("INTEGER");
                b.Property<string>("RunId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<string>("JobId").HasMaxLength(64).HasColumnType("TEXT");
                b.Property<DateTime>("StartTime").HasColumnType("TEXT");
                b.Property<DateTime?>("EndTime").HasColumnType("TEXT");
                b.Property<string>("Status").HasColumnType("TEXT");
                b.Property<long>("BytesProcessed").HasColumnType("INTEGER");
                b.Property<long>("FilesProcessed").HasColumnType("INTEGER");
                b.Property<double>("SpeedMbps").HasColumnType("REAL");
                b.Property<string>("ErrorMessage").HasColumnType("TEXT");
                b.HasKey("Id");
                b.ToTable("job_run_history");
            });

            modelBuilder.Entity("Backup.Server.Database.Entities.Chunk", b =>
            {
                b.Property<string>("Hash").HasMaxLength(128).HasColumnType("TEXT");
                b.Property<byte[]>("Data").HasColumnType("BLOB");
                b.Property<long>("ReferenceCount").HasColumnType("INTEGER");
                b.Property<DateTime>("CreatedAt").HasColumnType("TEXT");
                b.HasKey("Hash");
                b.ToTable("chunks");
            });
        }
    }
}
