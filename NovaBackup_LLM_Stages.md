# NovaBackup -- LLM Development Stages

This document contains staged prompts you can give to an LLM to
progressively build a Windows backup system similar in functionality to
Veeam Backup & Replication.

Recommended stack: - Language: C# - Framework: .NET 8 - GUI: WPF -
Database: SQLite - Installer: WiX Toolset

------------------------------------------------------------------------

# Stage 1 -- Project Architecture

PROMPT:

You are a senior software architect.

Design a Windows backup system similar to Veeam Backup & Replication but
simplified.

Technology stack: - C# - .NET 8 - WPF GUI - SQLite database

Create a modular architecture.

Modules:

NovaBackup.Core NovaBackup.Engine NovaBackup.Storage
NovaBackup.Scheduler NovaBackup.Agent NovaBackup.API NovaBackup.GUI

Output:

1.  Complete project folder structure
2.  Responsibilities of each module
3.  Data flow diagram
4.  Initial solution structure for Visual Studio

------------------------------------------------------------------------

# Stage 2 -- Backup Engine

PROMPT:

Create a modular backup engine for Windows.

Language: C# (.NET 8)

Features:

-   file backup
-   folder backup
-   full backup
-   incremental backup
-   compression
-   AES-256 encryption

Architecture:

BackupEngine FileScanner BackupProcessor CompressionService
EncryptionService MetadataWriter

Requirements:

-   handle large files
-   multithreaded processing
-   metadata for restore

Output:

Full C# implementation and project structure.

------------------------------------------------------------------------

# Stage 3 -- Job Scheduler

PROMPT:

Create a backup job scheduler service.

Language: C# (.NET 8)

Features:

-   scheduled backups
-   daily jobs
-   weekly jobs
-   retry failed jobs
-   job queue
-   job status tracking

Components:

SchedulerService JobRunner JobQueue RetryManager

Output:

Full working scheduler service implementation.

------------------------------------------------------------------------

# Stage 4 -- Storage System

PROMPT:

Create a storage abstraction layer.

Supported storage:

-   Local disk
-   SMB network share
-   NAS

Architecture:

StorageProvider LocalStorageProvider SmbStorageProvider
RepositoryManager

Requirements:

-   streaming writes
-   large archive support

Output:

Complete C# implementation.

------------------------------------------------------------------------

# Stage 5 -- Restore Engine

PROMPT:

Create a restore engine.

Features:

-   restore single file
-   restore folders
-   restore full backup
-   restore from incremental chain

Architecture:

RestoreEngine ArchiveReader MetadataLoader FileRestorer

Output:

Working restore system with example restore workflow.

------------------------------------------------------------------------

# Stage 6 -- GUI

PROMPT:

Create a modern WPF interface.

Sections:

Dashboard Backup Jobs Restore Repositories History Settings

Features:

-   dark theme
-   real-time backup progress
-   job logs viewer
-   repository usage stats

Output:

Complete WPF application layout and code.

------------------------------------------------------------------------

# Stage 7 -- Notifications

PROMPT:

Create a notification system.

Events:

-   backup success
-   backup failure
-   restore completed

Channels:

-   email
-   Telegram
-   webhook

Output:

Notification service implementation.

------------------------------------------------------------------------

# Stage 8 -- Deduplication

PROMPT:

Implement block-level deduplication.

Steps:

1.  split files into blocks
2.  calculate SHA256 hash
3.  store only unique blocks
4.  maintain block index database

Requirements:

-   optimized for large datasets
-   fast lookup

Output:

Deduplication engine module.

------------------------------------------------------------------------

# Stage 9 -- Replication

PROMPT:

Create a replication system.

Features:

-   replicate backup repositories to another server
-   incremental replication
-   SMB and SFTP support

Components:

ReplicationService ReplicationQueue TransferManager

Output:

Replication module implementation.

------------------------------------------------------------------------

# Stage 10 -- VM Backup

PROMPT:

Implement virtual machine backup.

Support:

-   Hyper-V

Requirements:

-   use VSS snapshots
-   incremental VM disk backup
-   restore VM

Output:

VM backup module.

------------------------------------------------------------------------

# Stage 11 -- Cloud Backup

PROMPT:

Implement cloud backup support.

Targets:

-   Amazon S3
-   S3-compatible storage

Features:

-   multipart uploads
-   encrypted backups
-   retention policies

Output:

Cloud storage module.

------------------------------------------------------------------------

# Stage 12 -- Installer

PROMPT:

Create a professional Windows installer.

Technology:

WiX Toolset

Features:

-   install services
-   install GUI application
-   configure default repository
-   create start menu shortcuts

Output:

Complete MSI installer project.
