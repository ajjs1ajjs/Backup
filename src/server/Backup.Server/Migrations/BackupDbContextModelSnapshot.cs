using Backup.Server.Database;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;

namespace Backup.Server.Migrations;

[DbContext(typeof(BackupDbContext))]
partial class BackupDbContextModelSnapshot : ModelSnapshot
{
    protected override void BuildModel(ModelBuilder modelBuilder)
    {
#pragma warning disable 612, 618
        modelBuilder
            .HasAnnotation("ProductVersion", "8.0.0");

        modelBuilder.Entity("Backup.Server.Database.Entities.Agent", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<string>("AgentId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("AgentType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("AgentVersion")
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("Capabilities")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Hostname")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("LastHeartbeat")
                .HasColumnType("TEXT");

            b.Property<string>("OsType")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.HasKey("Id");

            b.HasIndex("AgentId")
                .IsUnique();

            b.ToTable("agents");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.AuditLog", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<string>("Action")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Details")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("EntityId")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("EntityType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("UserId")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.HasKey("Id");
            b.ToTable("audit_logs");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.BackupPoint", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<string>("BackupId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("BackupType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("Checksum")
                .HasMaxLength(128)
                .HasColumnType("TEXT");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<DateTime?>("CompletedAt")
                .HasColumnType("TEXT");

            b.Property<string>("FilePath")
                .HasMaxLength(1024)
                .HasColumnType("TEXT");

            b.Property<bool>("IsSynthetic")
                .HasColumnType("INTEGER");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("Metadata")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<long>("OriginalSizeBytes")
                .HasColumnType("INTEGER");

            b.Property<string>("ParentBackupId")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("RepositoryId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<long>("SizeBytes")
                .HasColumnType("INTEGER");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("VmId")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.HasKey("Id");

            b.HasIndex("BackupId")
                .IsUnique();

            b.HasIndex("JobId");

            b.ToTable("backup_points");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Hypervisor", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Host")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("HypervisorId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("LastConnectedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("Password")
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<int>("Port")
                .HasColumnType("INTEGER");

            b.Property<string>("Status")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("Type")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Username")
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<int>("VmCount")
                .HasColumnType("INTEGER");

            b.HasKey("Id");

            b.HasIndex("HypervisorId")
                .IsUnique();

            b.ToTable("hypervisors");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Job", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("DestinationId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<bool>("Enabled")
                .HasColumnType("INTEGER");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("JobType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("LastRun")
                .HasColumnType("TEXT");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("NextRun")
                .HasColumnType("TEXT");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("Schedule")
                .HasColumnType("TEXT");

            b.Property<string>("SourceId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("SourceType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.HasKey("Id");

            b.HasIndex("JobId")
                .IsUnique();

            b.ToTable("jobs");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.JobRunHistory", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<long>("BytesProcessed")
                .HasColumnType("INTEGER");

            b.Property<DateTime?>("EndTime")
                .HasColumnType("TEXT");

            b.Property<long>("FilesProcessed")
                .HasColumnType("INTEGER");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("RunId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<double>("SpeedMbps")
                .HasColumnType("REAL");

            b.Property<DateTime>("StartTime")
                .HasColumnType("TEXT");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("ErrorMessage")
                .HasColumnType("TEXT");

            b.HasKey("Id");
            b.ToTable("job_run_history");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Repository", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<long?>("CapacityBytes")
                .HasColumnType("INTEGER");

            b.Property<string>("Credentials")
                .HasColumnType("TEXT");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<DateTime?>("LastUsedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("Path")
                .IsRequired()
                .HasMaxLength(1024)
                .HasColumnType("TEXT");

            b.Property<string>("RepositoryId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("Type")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.Property<long>("UsedBytes")
                .HasColumnType("INTEGER");

            b.HasKey("Id");

            b.HasIndex("RepositoryId")
                .IsUnique();

            b.ToTable("repositories");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Restore", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<string>("BackupId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<DateTime?>("CompletedAt")
                .HasColumnType("TEXT");

            b.Property<string>("DestinationPath")
                .HasMaxLength(1024)
                .HasColumnType("TEXT");

            b.Property<string>("ErrorMessage")
                .HasColumnType("TEXT");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("RestoreId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("RestoreType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("StartedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("TargetHost")
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<long>("TotalBytes")
                .HasColumnType("INTEGER");

            b.Property<long>("BytesRestored")
                .HasColumnType("INTEGER");

            b.HasKey("Id");

            b.HasIndex("RestoreId")
                .IsUnique();

            b.ToTable("restores");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Setting", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<string>("Description")
                .HasColumnType("TEXT");

            b.Property<string>("Key")
                .IsRequired()
                .HasMaxLength(128)
                .HasColumnType("TEXT");

            b.Property<string>("Type")
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Value")
                .IsRequired()
                .HasColumnType("TEXT");

            b.HasKey("Id");
            b.ToTable("settings");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.User", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Email")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<bool>("IsActive")
                .HasColumnType("INTEGER");

            b.Property<DateTime?>("LastLoginAt")
                .HasColumnType("TEXT");

            b.Property<bool>("MustChangePassword")
                .HasColumnType("INTEGER");

            b.Property<string>("PasswordHash")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("Role")
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("TwoFactorSecret")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("UserId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("Username")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.HasKey("Id");

            b.HasIndex("Username")
                .IsUnique();

            b.ToTable("users");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.VirtualMachine", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("INTEGER");

            b.Property<int?>("CpuCores")
                .HasColumnType("INTEGER");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("Disks")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<string>("HypervisorHost")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("HypervisorType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<DateTime?>("LastBackupAt")
                .HasColumnType("TEXT");

            b.Property<string>("LastBackupId")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<long?>("MemoryMb")
                .HasColumnType("INTEGER");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("TEXT");

            b.Property<string>("OsType")
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.Property<string>("Status")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("TEXT");

            b.Property<string>("Tags")
                .IsRequired()
                .HasColumnType("TEXT");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("TEXT");

            b.Property<string>("VmId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("TEXT");

            b.HasKey("Id");

            b.HasIndex("VmId")
                .IsUnique();

            b.ToTable("virtual_machines");
        });
#pragma warning restore 612, 618
    }
}
