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
            .HasAnnotation("ProductVersion", "8.0.0")
            .HasAnnotation("Relational:MaxIdentifierLength", 63);

        NpgsqlModelBuilderExtensions.UseIdentityByDefaultColumns(modelBuilder);

        modelBuilder.Entity("Backup.Server.Database.Entities.Agent", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<string>("AgentId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("AgentType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("AgentVersion")
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("Capabilities")
                .IsRequired()
                .HasColumnType("text");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Hostname")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<DateTime?>("LastHeartbeat")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("OsType")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("text");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.HasKey("Id");

            b.HasIndex("AgentId")
                .IsUnique();

            b.ToTable("agents");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.AuditLog", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<string>("Action")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Details")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("EntityId")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("EntityType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("UserId")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.HasKey("Id");
            b.ToTable("audit_logs");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.BackupPoint", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<string>("BackupId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("BackupType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("Checksum")
                .HasMaxLength(128)
                .HasColumnType("character varying(128)");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<DateTime?>("CompletedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("FilePath")
                .HasMaxLength(1024)
                .HasColumnType("character varying(1024)");

            b.Property<bool>("IsSynthetic")
                .HasColumnType("boolean");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("Metadata")
                .IsRequired()
                .HasColumnType("text");

            b.Property<long>("OriginalSizeBytes")
                .HasColumnType("bigint");

            b.Property<string>("ParentBackupId")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("RepositoryId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<long>("SizeBytes")
                .HasColumnType("bigint");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("VmId")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

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
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Host")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("HypervisorId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<DateTime?>("LastConnectedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("Password")
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<int>("Port")
                .HasColumnType("integer");

            b.Property<string>("Status")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("Type")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Username")
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<int>("VmCount")
                .HasColumnType("integer");

            b.HasKey("Id");

            b.HasIndex("HypervisorId")
                .IsUnique();

            b.ToTable("hypervisors");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Job", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("DestinationId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<bool>("Enabled")
                .HasColumnType("boolean");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("JobType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime?>("LastRun")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<DateTime?>("NextRun")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("Schedule")
                .HasColumnType("text");

            b.Property<string>("SourceId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("SourceType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.HasKey("Id");

            b.HasIndex("JobId")
                .IsUnique();

            b.ToTable("jobs");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.JobRunHistory", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<long>("BytesProcessed")
                .HasColumnType("bigint");

            b.Property<DateTime?>("EndTime")
                .HasColumnType("timestamp with time zone");

            b.Property<long>("FilesProcessed")
                .HasColumnType("bigint");

            b.Property<string>("JobId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("RunId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<double>("SpeedMbps")
                .HasColumnType("double precision");

            b.Property<DateTime>("StartTime")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("ErrorMessage")
                .HasColumnType("text");

            b.HasKey("Id");
            b.ToTable("job_run_history");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Repository", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<long?>("CapacityBytes")
                .HasColumnType("bigint");

            b.Property<string>("Credentials")
                .HasColumnType("text");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<DateTime?>("LastUsedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("Path")
                .IsRequired()
                .HasMaxLength(1024)
                .HasColumnType("character varying(1024)");

            b.Property<string>("RepositoryId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("Type")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<long>("UsedBytes")
                .HasColumnType("bigint");

            b.HasKey("Id");

            b.HasIndex("RepositoryId")
                .IsUnique();

            b.ToTable("repositories");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Restore", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<string>("BackupId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<DateTime?>("CompletedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("DestinationPath")
                .HasMaxLength(1024)
                .HasColumnType("character varying(1024)");

            b.Property<string>("ErrorMessage")
                .HasColumnType("text");

            b.Property<string>("Options")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("RestoreId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("RestoreType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime?>("StartedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Status")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("TargetHost")
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<long>("TotalBytes")
                .HasColumnType("bigint");

            b.Property<long>("BytesRestored")
                .HasColumnType("bigint");

            b.HasKey("Id");

            b.HasIndex("RestoreId")
                .IsUnique();

            b.ToTable("restores");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.Setting", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<string>("Description")
                .HasColumnType("text");

            b.Property<string>("Key")
                .IsRequired()
                .HasMaxLength(128)
                .HasColumnType("character varying(128)");

            b.Property<string>("Type")
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Value")
                .IsRequired()
                .HasColumnType("text");

            b.HasKey("Id");
            b.ToTable("settings");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.User", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Email")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<bool>("IsActive")
                .HasColumnType("boolean");

            b.Property<DateTime?>("LastLoginAt")
                .HasColumnType("timestamp with time zone");

            b.Property<bool>("MustChangePassword")
                .HasColumnType("boolean");

            b.Property<string>("PasswordHash")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("Role")
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("TwoFactorSecret")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("UserId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("Username")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.HasKey("Id");

            b.HasIndex("Username")
                .IsUnique();

            b.ToTable("users");
        });

        modelBuilder.Entity("Backup.Server.Database.Entities.VirtualMachine", b =>
        {
            b.Property<long>("Id")
                .ValueGeneratedOnAdd()
                .HasColumnType("bigint");
            NpgsqlPropertyBuilderExtensions.UseIdentityByDefaultColumn(b.Property<long>("Id"));

            b.Property<int?>("CpuCores")
                .HasColumnType("integer");

            b.Property<DateTime>("CreatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("Disks")
                .IsRequired()
                .HasColumnType("text");

            b.Property<string>("HypervisorHost")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("HypervisorType")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("IpAddress")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<DateTime?>("LastBackupAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("LastBackupId")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<long?>("MemoryMb")
                .HasColumnType("bigint");

            b.Property<string>("Name")
                .IsRequired()
                .HasMaxLength(255)
                .HasColumnType("character varying(255)");

            b.Property<string>("OsType")
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.Property<string>("Status")
                .IsRequired()
                .HasMaxLength(32)
                .HasColumnType("character varying(32)");

            b.Property<string>("Tags")
                .IsRequired()
                .HasColumnType("text");

            b.Property<DateTime>("UpdatedAt")
                .HasColumnType("timestamp with time zone");

            b.Property<string>("VmId")
                .IsRequired()
                .HasMaxLength(64)
                .HasColumnType("character varying(64)");

            b.HasKey("Id");

            b.HasIndex("VmId")
                .IsUnique();

            b.ToTable("virtual_machines");
        });
#pragma warning restore 612, 618
    }
}
