using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

namespace Backup.Server.Migrations;

/// <inheritdoc />
public partial class InitialCreate : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.CreateTable(
            name: "agents",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                AgentId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Hostname = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                OsType = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                AgentVersion = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: true),
                AgentType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                Status = table.Column<string>(type: "text", nullable: false),
                IpAddress = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                LastHeartbeat = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                Capabilities = table.Column<string>(type: "text", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_agents", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "audit_logs",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                UserId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                Action = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                EntityType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                EntityId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                Details = table.Column<string>(type: "text", nullable: false),
                IpAddress = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_audit_logs", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "hypervisors",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                HypervisorId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Name = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                Type = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                Host = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                Port = table.Column<int>(type: "integer", nullable: false),
                Username = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: true),
                Password = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: true),
                Status = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                VmCount = table.Column<int>(type: "integer", nullable: false),
                LastConnectedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_hypervisors", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "job_run_history",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                RunId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                JobId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                StartTime = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                EndTime = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                Status = table.Column<string>(type: "text", nullable: false),
                BytesProcessed = table.Column<long>(type: "bigint", nullable: false),
                FilesProcessed = table.Column<long>(type: "bigint", nullable: false),
                SpeedMbps = table.Column<double>(type: "double precision", nullable: false),
                ErrorMessage = table.Column<string>(type: "text", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_job_run_history", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "jobs",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                JobId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Name = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                JobType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                SourceId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                SourceType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                DestinationId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Schedule = table.Column<string>(type: "text", nullable: true),
                Options = table.Column<string>(type: "text", nullable: false),
                Enabled = table.Column<bool>(type: "boolean", nullable: false),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                LastRun = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                NextRun = table.Column<DateTime>(type: "timestamp with time zone", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_jobs", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "repositories",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                RepositoryId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Name = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                Type = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                Path = table.Column<string>(type: "character varying(1024)", maxLength: 1024, nullable: false),
                Status = table.Column<string>(type: "text", nullable: false),
                CapacityBytes = table.Column<long>(type: "bigint", nullable: true),
                UsedBytes = table.Column<long>(type: "bigint", nullable: false),
                Credentials = table.Column<string>(type: "text", nullable: true),
                Options = table.Column<string>(type: "text", nullable: false),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                LastUsedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_repositories", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "settings",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                Key = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false),
                Value = table.Column<string>(type: "text", nullable: false),
                Type = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                Description = table.Column<string>(type: "text", nullable: true),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_settings", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "users",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                UserId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Username = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Email = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                PasswordHash = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                Role = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                IsActive = table.Column<bool>(type: "boolean", nullable: false),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                LastLoginAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                MustChangePassword = table.Column<bool>(type: "boolean", nullable: false),
                TwoFactorSecret = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_users", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "virtual_machines",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                VmId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                Name = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                HypervisorType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                HypervisorHost = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: false),
                IpAddress = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                OsType = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                MemoryMb = table.Column<long>(type: "bigint", nullable: true),
                CpuCores = table.Column<int>(type: "integer", nullable: true),
                Disks = table.Column<string>(type: "text", nullable: false),
                Tags = table.Column<string>(type: "text", nullable: false),
                Status = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                LastBackupAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                LastBackupId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_virtual_machines", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "backup_points",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                BackupId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                JobId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                VmId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                BackupType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                RepositoryId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                FilePath = table.Column<string>(type: "character varying(1024)", maxLength: 1024, nullable: true),
                SizeBytes = table.Column<long>(type: "bigint", nullable: false),
                OriginalSizeBytes = table.Column<long>(type: "bigint", nullable: false),
                Checksum = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: true),
                IsSynthetic = table.Column<bool>(type: "boolean", nullable: false),
                ParentBackupId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: true),
                Metadata = table.Column<string>(type: "text", nullable: false),
                Status = table.Column<string>(type: "text", nullable: false),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                CompletedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_backup_points", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "restores",
            columns: table => new
            {
                Id = table.Column<long>(type: "bigint", nullable: false)
                    .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                RestoreId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                BackupId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                RestoreType = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                DestinationPath = table.Column<string>(type: "character varying(1024)", maxLength: 1024, nullable: true),
                TargetHost = table.Column<string>(type: "character varying(255)", maxLength: 255, nullable: true),
                Options = table.Column<string>(type: "text", nullable: false),
                Status = table.Column<string>(type: "text", nullable: false),
                BytesRestored = table.Column<long>(type: "bigint", nullable: false),
                TotalBytes = table.Column<long>(type: "bigint", nullable: false),
                ErrorMessage = table.Column<string>(type: "text", nullable: true),
                CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                CompletedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                StartedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_restores", x => x.Id);
            });

        migrationBuilder.CreateIndex(
            name: "IX_agents_AgentId",
            table: "agents",
            column: "AgentId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_backup_points_BackupId",
            table: "backup_points",
            column: "BackupId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_backup_points_JobId",
            table: "backup_points",
            column: "JobId");

        migrationBuilder.CreateIndex(
            name: "IX_hypervisors_HypervisorId",
            table: "hypervisors",
            column: "HypervisorId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_jobs_JobId",
            table: "jobs",
            column: "JobId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_repositories_RepositoryId",
            table: "repositories",
            column: "RepositoryId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_restores_RestoreId",
            table: "restores",
            column: "RestoreId",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_users_Username",
            table: "users",
            column: "Username",
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_virtual_machines_VmId",
            table: "virtual_machines",
            column: "VmId",
            unique: true);
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropTable(name: "agents");
        migrationBuilder.DropTable(name: "audit_logs");
        migrationBuilder.DropTable(name: "backup_points");
        migrationBuilder.DropTable(name: "hypervisors");
        migrationBuilder.DropTable(name: "job_run_history");
        migrationBuilder.DropTable(name: "jobs");
        migrationBuilder.DropTable(name: "repositories");
        migrationBuilder.DropTable(name: "restores");
        migrationBuilder.DropTable(name: "settings");
        migrationBuilder.DropTable(name: "users");
        migrationBuilder.DropTable(name: "virtual_machines");
    }
}
