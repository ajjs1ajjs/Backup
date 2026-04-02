-- Database initialization script for integration tests
-- This script runs when the PostgreSQL container starts

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create tables for integration tests
CREATE TABLE IF NOT EXISTS agents (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    os_type VARCHAR(50) NOT NULL,
    agent_version VARCHAR(50) NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'idle',
    capabilities TEXT[] DEFAULT '{}',
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS repositories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    path VARCHAR(512) NOT NULL,
    max_capacity_bytes BIGINT,
    warning_threshold_percent INT DEFAULT 80,
    critical_threshold_percent INT DEFAULT 95,
    is_accessible BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS jobs (
    id BIGSERIAL PRIMARY KEY,
    job_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    job_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    destination_id VARCHAR(255) NOT NULL,
    schedule VARCHAR(100),
    enabled BOOLEAN DEFAULT TRUE,
    compression_enabled BOOLEAN DEFAULT TRUE,
    deduplication_enabled BOOLEAN DEFAULT FALSE,
    incremental_base_path VARCHAR(512),
    retention_days INT DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS job_run_history (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT REFERENCES jobs(id) ON DELETE CASCADE,
    job_run_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT,
    bytes_processed BIGINT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backups (
    id BIGSERIAL PRIMARY KEY,
    backup_id VARCHAR(255) UNIQUE NOT NULL,
    job_id BIGINT REFERENCES jobs(id) ON DELETE SET NULL,
    repository_id BIGINT REFERENCES repositories(id) ON DELETE SET NULL,
    source_id VARCHAR(255) NOT NULL,
    backup_type VARCHAR(50) NOT NULL,
    size_bytes BIGINT,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    retention_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS restores (
    id BIGSERIAL PRIMARY KEY,
    restore_id VARCHAR(255) UNIQUE NOT NULL,
    backup_id BIGINT REFERENCES backups(id) ON DELETE SET NULL,
    restore_type VARCHAR(50) NOT NULL,
    target_host VARCHAR(255),
    destination_path VARCHAR(512),
    status VARCHAR(50) NOT NULL,
    progress_percent INT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stress_test_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    total_vms INT NOT NULL,
    target_concurrency INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    total_backups INT DEFAULT 0,
    successful_backups INT DEFAULT 0,
    failed_backups INT DEFAULT 0,
    average_duration_ms DOUBLE PRECISION,
    percentile_95_duration_ms DOUBLE PRECISION,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_agent_type ON agents(agent_type);
CREATE INDEX IF NOT EXISTS idx_jobs_enabled ON jobs(enabled);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type ON jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_job_run_history_job_id ON job_run_history(job_id);
CREATE INDEX IF NOT EXISTS idx_job_run_history_status ON job_run_history(status);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status);
CREATE INDEX IF NOT EXISTS idx_backups_job_id ON backups(job_id);
CREATE INDEX IF NOT EXISTS idx_restores_status ON restores(status);
CREATE INDEX IF NOT EXISTS idx_stress_test_sessions_status ON stress_test_sessions(session_id);

-- Insert test data
INSERT INTO repositories (name, type, path, max_capacity_bytes) VALUES
    ('Test Local Repository', 'local', '/data/backups/local', 1099511627776),
    ('Test S3 Repository', 's3', 's3://backup-bucket/backups', 10995116277760),
    ('Test Azure Repository', 'azure', 'azure://backupcontainer/backups', 5497558138880);

INSERT INTO agents (agent_id, hostname, os_type, agent_version, agent_type, status, capabilities) VALUES
    ('test-agent-001', 'hyperv-host-01.local', 'Windows', '1.0.0', 'hyperv', 'idle', ARRAY['backup', 'restore', 'cbt']),
    ('test-agent-002', 'vmware-host-01.local', 'Linux', '1.0.0', 'vmware', 'idle', ARRAY['backup', 'restore', 'cbt']),
    ('test-agent-003', 'kvm-host-01.local', 'Linux', '1.0.0', 'kvm', 'idle', ARRAY['backup', 'restore']);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_repositories_updated_at BEFORE UPDATE ON repositories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
