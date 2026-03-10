import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import apiClient, { DashboardStats, Backup } from '../api/client';
import './Dashboard.css';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recentBackups, setRecentBackups] = useState<Backup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDashboardData();
    // Refresh every 30 seconds
    const interval = setInterval(fetchDashboardData, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const [statsData, backupsData] = await Promise.all([
        apiClient.getStats(),
        apiClient.getBackups()
      ]);
      setStats(statsData);
      setRecentBackups(backupsData.slice(0, 5));
      setError(null);
    } catch (err) {
      setError('Failed to load dashboard data');
      console.error('Dashboard error:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTime = (dateString?: string): string => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  const getTimeAgo = (dateString?: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  if (loading) {
    return (
      <div className="page">
        <div className="loading">Loading dashboard...</div>
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
      <div className="dashboard">
        <div className="dashboard-header">
          <h2>📊 Dashboard</h2>
          <span className="last-update">
            Last updated: {new Date().toLocaleTimeString()}
          </span>
        </div>

        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-card-header">
              <div>
                <div className="stat-value">{stats?.totalJobs || 0}</div>
                <div className="stat-label">Total Jobs</div>
              </div>
              <div className="stat-icon">💼</div>
            </div>
            <div className="stat-change positive">
              {stats?.activeJobs || 0} active
            </div>
          </div>

          <div className="stat-card success">
            <div className="stat-card-header">
              <div>
                <div className="stat-value">{stats?.totalBackups || 0}</div>
                <div className="stat-label">Total Backups</div>
              </div>
              <div className="stat-icon">📁</div>
            </div>
            <div className="stat-change positive">
              Last: {getTimeAgo(stats?.lastBackup)}
            </div>
          </div>

          <div className="stat-card warning">
            <div className="stat-card-header">
              <div>
                <div className="stat-value">{formatSize(stats?.totalSize || 0)}</div>
                <div className="stat-label">Total Size</div>
              </div>
              <div className="stat-icon">💾</div>
            </div>
            <div className="stat-change">
              Across all repositories
            </div>
          </div>

          <div className="stat-card danger">
            <div className="stat-card-header">
              <div>
                <div className="stat-value">{stats?.recentFailures || 0}</div>
                <div className="stat-label">Recent Failures</div>
              </div>
              <div className="stat-icon">⚠️</div>
            </div>
            <div className="stat-change negative">
              Last 24 hours
            </div>
          </div>
        </div>

        <div className="recent-activity">
          <h3>Recent Backup Activity</h3>
          {recentBackups.length === 0 ? (
            <div className="no-data">
              <div className="no-data-icon">📭</div>
              <p>No backups yet</p>
              <Link to="/jobs/new" className="btn btn-primary">Create First Job</Link>
            </div>
          ) : (
            <div className="activity-list">
              {recentBackups.map((backup) => (
                <div key={backup.id} className="activity-item">
                  <div className={`activity-icon ${backup.status}`}>
                    {backup.status === 'completed' ? '✓' : backup.status === 'failed' ? '✗' : '⟳'}
                  </div>
                  <div className="activity-info">
                    <div className="activity-title">{backup.jobName}</div>
                    <div className="activity-time">
                      {formatTime(backup.startTime)} • {formatSize(backup.size)}
                    </div>
                  </div>
                  <span className={`activity-status ${backup.status}`}>
                    {backup.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="quick-actions">
          <h3>Quick Actions</h3>
          <div className="quick-actions-grid">
            <Link to="/jobs/new" className="quick-action-card">
              <span className="quick-action-icon">➕</span>
              <span>Create New Job</span>
            </Link>
            <Link to="/jobs" className="quick-action-card">
              <span className="quick-action-icon">▶️</span>
              <span>Run Backup</span>
            </Link>
            <Link to="/restore" className="quick-action-card">
              <span className="quick-action-icon">🔄</span>
              <span>Restore Files</span>
            </Link>
            <Link to="/storage" className="quick-action-card">
              <span className="quick-action-icon">📊</span>
              <span>View Storage</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
