-- Backup System Database Schema for PostgreSQL
-- Version: 1.0.0

-- ENUMS
CREATE TYPE agent_status AS ENUM ('idle', 'backing_up', 'restoring', 'error');
CREATE TYPE job_type AS ENUM ('full_backup', 'incremental', 'differential', 'restore', 'verify');
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');
CREATE TYPE backup_type AS ENUM ('full', 'incremental', 'differential', 'synthetic_full');
CREATE TYPE backup_status AS ENUM ('in_progress', 'completed', 'failed', 'verified', 'expired');
CREATE TYPE restore_type AS ENUM ('full_vm', 'instant', 'file_level', 'database', 'export');
CREATE TYPE restore_status AS ENUM ('pending', 'in_progress', 'completed', 'failed', 'cancelled');
CREATE TYPE repository_type AS ENUM ('local', 'nfs', 'smb', 's3', 'azure_blob', 'gcs', 'tape');
CREATE TYPE repository_status AS ENUM ('online', 'offline', 'error', 'maintenance');
CREATE TYPE hypervisor_type AS ENUM ('hyperv', 'vmware', 'kvm');

-- AGENTS
CREATE TABLE agents (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(64) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    os_type VARCHAR(64) NOT NULL,
    agent_version VARCHAR(32),
    agent_type VARCHAR(32) NOT NULL,
    status agent_status DEFAULT 'idle',
    ip_address VARCHAR(64),
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    capabilities JSONB DEFAULT '[]'
);

-- VIRTUAL MACHINES
CREATE TABLE virtual_machines (
    id BIGSERIAL PRIMARY KEY,
    vm_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    hypervisor_type hypervisor_type NOT NULL,
    hypervisor_host VARCHAR(255) NOT NULL,
    ip_address VARCHAR(64),
    os_type VARCHAR(64),
    memory_mb BIGINT,
    cpu_cores INTEGER,
    disks JSONB DEFAULT '[]',
    tags JSONB DEFAULT '{}',
    status VARCHAR(32) DEFAULT 'running',
    last_backup_at TIMESTAMP,
    last_backup_id VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- REPOSITORIES
CREATE TABLE repositories (
    id BIGSERIAL PRIMARY KEY,
    repository_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type repository_type NOT NULL,
    path VARCHAR(1024) NOT NULL,
    status repository_status DEFAULT 'online',
    capacity_bytes BIGINT,
    used_bytes BIGINT DEFAULT 0,
    credentials JSONB,
    options JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP
);

-- JOBS
CREATE TABLE jobs (
    id BIGSERIAL PRIMARY KEY,
    job_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    job_type job_type NOT NULL,
    source_id VARCHAR(64) NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    destination_id VARCHAR(64) NOT NULL,
    schedule JSONB,
    options JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_run TIMESTAMP,
    next_run TIMESTAMP
);

-- BACKUP POINTS
CREATE TABLE backup_points (
    id BIGSERIAL PRIMARY KEY,
    backup_id VARCHAR(64) UNIQUE NOT NULL,
    job_id VARCHAR(64) NOT NULL,
    vm_id VARCHAR(64),
    backup_type backup_type NOT NULL,
    repository_id VARCHAR(64) NOT NULL,
    file_path VARCHAR(1024),
    size_bytes BIGINT DEFAULT 0,
    original_size_bytes BIGINT DEFAULT 0,
    checksum VARCHAR(128),
    is_synthetic BOOLEAN DEFAULT FALSE,
    parent_backup_id VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    status backup_status DEFAULT 'in_progress',
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES repositories(repository_id)
);

-- RESTORES
CREATE TABLE restores (
    id BIGSERIAL PRIMARY KEY,
    restore_id VARCHAR(64) UNIQUE NOT NULL,
    backup_id VARCHAR(64) NOT NULL,
    restore_type restore_type NOT NULL,
    destination_path VARCHAR(1024),
    target_host VARCHAR(255),
    options JSONB DEFAULT '{}',
    status restore_status DEFAULT 'pending',
    bytes_restored BIGINT DEFAULT 0,
    total_bytes BIGINT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    started_at TIMESTAMP,
    FOREIGN KEY (backup_id) REFERENCES backup_points(backup_id)
);

-- JOB RUN HISTORY
CREATE TABLE job_run_history (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(64) UNIQUE NOT NULL,
    job_id VARCHAR(64) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    status job_status DEFAULT 'pending',
    bytes_processed BIGINT DEFAULT 0,
    files_processed BIGINT DEFAULT 0,
    speed_mbps DOUBLE PRECISION DEFAULT 0,
    error_message TEXT,
    FOREIGN KEY (job_id) REFERENCES jobs(job_id)
);

-- USERS
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) UNIQUE NOT NULL,
    username VARCHAR(64) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP,
    must_change_password BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(64)
);

-- AUDIT LOGS
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64),
    action VARCHAR(64) NOT NULL,
    entity_type VARCHAR(32) NOT NULL,
    entity_id VARCHAR(64),
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW()
);

-- SETTINGS
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(128) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(32) DEFAULT 'string',
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- INDEXES
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_agent_id ON agents(agent_id);
CREATE INDEX idx_vms_hypervisor ON virtual_machines(hypervisor_type, hypervisor_host);
CREATE INDEX idx_vms_vm_id ON virtual_machines(vm_id);
CREATE INDEX idx_jobs_enabled ON jobs(enabled);
CREATE INDEX idx_jobs_next_run ON jobs(next_run) WHERE next_run IS NOT NULL;
CREATE INDEX idx_backup_points_job ON backup_points(job_id);
CREATE INDEX idx_backup_points_vm ON backup_points(vm_id);
CREATE INDEX idx_backup_points_repo ON backup_points(repository_id);
CREATE INDEX idx_backup_points_created ON backup_points(created_at);
CREATE INDEX idx_restores_backup ON restores(backup_id);
CREATE INDEX idx_job_history_job ON job_run_history(job_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- FUNCTIONS
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- TRIGGERS
CREATE TRIGGER agents_updated_at
    BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER vms_updated_at
    BEFORE UPDATE ON virtual_machines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER repos_updated_at
    BEFORE UPDATE ON repositories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER jobs_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- DEFAULT SETTINGS
INSERT INTO settings (key, value, description) VALUES
    ('backup.compression', 'zstd', 'Default compression algorithm'),
    ('backup.block_size_kb', '64', 'Default block size in KB'),
    ('backup.retention_days', '30', 'Default retention days'),
    ('network.port', '8000', 'gRPC server port'),
    ('server.public_url', 'http://localhost:8000', 'Public server URL used by agents and installers'),
    ('security.encryption', 'aes256', 'Encryption algorithm'),
    ('scheduler.timezone', 'UTC', 'Default timezone');
