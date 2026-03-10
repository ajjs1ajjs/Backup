import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import apiClient, { BackupJob } from '../api/client';
import './JobList.css';

const JobList: React.FC = () => {
  const [jobs, setJobs] = useState<BackupJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionProgress, setActionProgress] = useState<number | null>(null);

  useEffect(() => {
    fetchJobs();
  }, []);

  const fetchJobs = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getJobs();
      setJobs(data);
      setError(null);
    } catch (err) {
      setError('Failed to load jobs');
      console.error('Fetch jobs error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      setActionProgress(id);
      await apiClient.toggleJob(id, !enabled);
      await fetchJobs();
    } catch (err) {
      alert('Failed to toggle job');
    } finally {
      setActionProgress(null);
    }
  };

  const handleRun = async (id: number) => {
    try {
      setActionProgress(id);
      await apiClient.runJob(id);
      alert('Job started');
      await fetchJobs();
    } catch (err) {
      alert('Failed to start job');
    } finally {
      setActionProgress(null);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Delete this job?')) return;
    try {
      setActionProgress(id);
      await apiClient.deleteJob(id);
      await fetchJobs();
    } catch (err) {
      alert('Failed to delete job');
    } finally {
      setActionProgress(null);
    }
  };

  const getStatusIcon = (status: BackupJob['status']) => {
    const icons: Record<string, string> = {
      completed: '✓',
      failed: '✗',
      running: '⟳',
      pending: '⏳'
    };
    return icons[status] || '?';
  };

  const getStatusClass = (status: BackupJob['status']) => {
    const classes: Record<string, string> = {
      completed: 'status-completed',
      failed: 'status-failed',
      running: 'status-running',
      pending: 'status-pending'
    };
    return classes[status] || '';
  };

  if (loading) {
    return (
      <div className="page">
        <div className="loading">Loading jobs...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page">
        <div className="error">{error}</div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="job-list">
        <div className="job-list-header">
          <h2>💼 Backup Jobs</h2>
          <Link to="/jobs/new" className="btn btn-primary">+ New Job</Link>
        </div>

        {jobs.length === 0 ? (
          <div className="no-jobs">
            <div className="no-jobs-icon">📭</div>
            <h3>No Backup Jobs</h3>
            <p>Create your first backup job to get started</p>
            <Link to="/jobs/new" className="btn btn-primary">Create Job</Link>
          </div>
        ) : (
          <div className="jobs-grid">
            {jobs.map((job) => (
              <div key={job.id} className={`job-card ${!job.enabled ? 'disabled' : ''}`}>
                <div className="job-card-header">
                  <div className="job-title">
                    <span className="job-icon">📘</span>
                    <h3>{job.name}</h3>
                  </div>
                  <span className={`job-status ${getStatusClass(job.status)}`}>
                    {getStatusIcon(job.status)} {job.status}
                  </span>
                </div>

                <div className="job-card-body">
                  <div className="job-info-row">
                    <span className="info-label">📁 Source:</span>
                    <span className="info-value">{job.source}</span>
                  </div>
                  <div className="job-info-row">
                    <span className="info-label">📤 Destination:</span>
                    <span className="info-value">{job.destination}</span>
                  </div>
                  <div className="job-info-row">
                    <span className="info-label">☁ Provider:</span>
                    <span className="info-value">{job.provider}</span>
                  </div>
                  <div className="job-info-row">
                    <span className="info-label">⏰ Schedule:</span>
                    <span className="info-value">{job.schedule}</span>
                  </div>
                  {job.lastRun && (
                    <div className="job-info-row">
                      <span className="info-label">🕐 Last Run:</span>
                      <span className="info-value">{new Date(job.lastRun).toLocaleString()}</span>
                    </div>
                  )}
                </div>

                <div className="job-card-footer">
                  <div className="job-actions">
                    <button
                      className={`btn btn-toggle ${job.enabled ? 'active' : ''}`}
                      onClick={() => handleToggle(job.id, job.enabled)}
                      disabled={actionProgress === job.id}
                    >
                      {job.enabled ? 'Disable' : 'Enable'}
                    </button>
                    <button
                      className="btn btn-run"
                      onClick={() => handleRun(job.id)}
                      disabled={actionProgress === job.id}
                    >
                      Run Now
                    </button>
                    <Link to={`/jobs/edit/${job.id}`} className="btn btn-edit">
                      Edit
                    </Link>
                    <button
                      className="btn btn-delete"
                      onClick={() => handleDelete(job.id)}
                      disabled={actionProgress === job.id}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default JobList;
