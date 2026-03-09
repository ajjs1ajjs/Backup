You are a principal distributed systems architect and senior backup infrastructure engineer.

Your task is to design and implement a complete enterprise-grade backup and disaster recovery platform similar to Veeam Backup & Replication.

The goal is to design a production-ready backup system capable of protecting virtual machines, physical servers, containers, NAS storage and cloud workloads.

The output must be extremely detailed and written as if this system will be implemented by a team of engineers.

--------------------------------------------------

SYSTEM GOALS

Build a full backup platform with the following capabilities:

• VM backup
• Physical server backup
• File-level backup
• Database-aware backup
• Continuous data protection
• Replication for disaster recovery
• Instant recovery
• Deduplication
• Compression
• Encryption
• Backup verification
• Multi-repository storage
• Scale-out storage architecture
• Object storage tiering

--------------------------------------------------

SUPPORTED INFRASTRUCTURE

The platform must support the following environments:

Virtualization
• VMware vSphere
• Hyper-V
• KVM

Physical systems
• Windows Server
• Linux

Cloud
• AWS EC2
• Azure VMs
• Google Cloud

Containers
• Kubernetes cluster backup

Storage
• Local disk
• NFS
• SMB
• S3 compatible object storage
• MinIO
• Ceph

--------------------------------------------------

CORE SYSTEM ARCHITECTURE

Design the system using the following core components:

1. CONTROL PLANE

Central orchestration service responsible for:

• backup jobs
• schedules
• policies
• repository management
• credential storage
• infrastructure discovery
• API management
• tenant management

Responsibilities:

• job scheduling
• infrastructure inventory
• policy enforcement
• metadata management

--------------------------------------------------

2. DATA MOVER NODES

Workers responsible for executing backup operations.

Responsibilities:

• snapshot orchestration
• reading source data
• block change tracking
• deduplication pipeline
• compression
• encryption
• writing data to storage

Requirements:

• parallel processing
• high throughput
• resumable transfers
• throttling

--------------------------------------------------

3. STORAGE ENGINE

Responsible for storing backup blocks.

Design:

• chunk-based storage
• deduplicated block store
• compressed block format
• content addressable storage

Features:

• block hashing
• dedupe index
• garbage collection
• immutable backups
• retention policies

--------------------------------------------------

4. BACKUP CATALOG DATABASE

Stores metadata about backups.

Must store:

• backup jobs
• restore points
• file indexes
• block maps
• repository metadata
• tenant data

Use PostgreSQL.

--------------------------------------------------

5. BACKUP REPOSITORY

Storage nodes responsible for storing backup files.

Support:

• local repositories
• network repositories
• scale-out repositories
• object storage repositories

Implement:

• repository load balancing
• capacity management
• health monitoring

--------------------------------------------------

6. RESTORE ENGINE

Responsible for recovery operations.

Support:

• full VM restore
• file-level restore
• instant VM recovery
• bare-metal restore
• database restore

Must support:

• streaming restore
• instant mount
• granular restore

--------------------------------------------------

7. REPLICATION ENGINE

Supports disaster recovery.

Features:

• VM replication
• incremental replication
• failover orchestration
• failback automation

--------------------------------------------------

8. MONITORING AND ALERTING

Build a monitoring subsystem.

Collect metrics for:

• backup speed
• repository capacity
• job success rate
• dedupe ratio
• system health

Export metrics in Prometheus format.

Integrate with Grafana.

--------------------------------------------------

9. SECURITY SYSTEM

Implement enterprise security features:

• RBAC
• API authentication
• audit logging
• encryption at rest
• encryption in transit
• immutable backups
• ransomware protection

--------------------------------------------------

BACKUP PIPELINE DESIGN

Design a backup pipeline with these stages:

1. Snapshot creation
2. Block discovery
3. Changed block detection
4. Block hashing
5. Deduplication check
6. Compression
7. Encryption
8. Transfer to repository
9. Metadata update

Explain every stage in detail.

--------------------------------------------------

DEDUPLICATION ENGINE

Implement global deduplication.

Explain:

• chunking algorithm
• hash index
• block referencing
• garbage collection

--------------------------------------------------

SCALE OUT ARCHITECTURE

Design a scale-out architecture supporting:

• multiple backup proxies
• multiple repositories
• distributed metadata
• load balancing

--------------------------------------------------

DISASTER RECOVERY DESIGN

Explain how the system handles:

• repository failure
• control plane failure
• data mover failure
• network partition

--------------------------------------------------

TECHNOLOGY STACK

Use the following technologies:

Core services
Go or Rust

Metadata database
PostgreSQL

Message bus
NATS or Kafka

Object storage
S3 compatible

API
REST + gRPC

Web UI
React

Agent software
Go or Rust

--------------------------------------------------

PROJECT STRUCTURE

Generate a complete repository layout.

Example:

/control-plane
/data-mover
/storage-engine
/repository-service
/restore-engine
/replication-engine
/web-ui
/cli
/agents
/docs

--------------------------------------------------

IMPLEMENTATION PHASES

Generate the system in phases:

Phase 1
Architecture design

Phase 2
Control Plane

Phase 3
Backup Job Engine

Phase 4
Data Movers

Phase 5
Storage Engine

Phase 6
Repositories

Phase 7
Restore Engine

Phase 8
Replication Engine

Phase 9
Web Interface

Phase 10
Monitoring

--------------------------------------------------

OUTPUT FORMAT

For every component include:

• architecture explanation
• sequence diagrams
• API endpoints
• folder structure
• code skeleton
• example configuration files

Write everything as if building a real enterprise backup platform.