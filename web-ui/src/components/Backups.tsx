import React, { useState, useEffect } from 'react';
import './Backups.css';

interface Backup {
  id: string;
  jobName: string;
  status: 'completed' | 'failed' | 'running' | 'cancelled';
  startTime: string;
  endTime?: string;
  bytesRead: number;
  bytesWritten: number;
  filesTotal: number;
  filesSuccess: number;
  filesFailed: number;
  errorMessage?: string;
  type: 'full' | 'incremental' | 'differential';
  source: string;
  destination: string;
}

interface StorageInfo {
  type: string;
  totalSize: number;
  usedSize: number;
  freeSize: number;
  objectCount: number;
  endpoint?: string;
  bucket?: string;
  region?: string;
}

const Backups: React.FC = () => {
  const [backups, setBackups] = useState<Backup[]>([]);
  const [storageInfo, setStorageInfo] = useState<StorageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedBackup, setSelectedBackup] = useState<Backup | null>(null);
  const [filter, setFilter] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState<string>('');

  useEffect(() => {
    fetchBackups();
    fetchStorageInfo();
  }, []);

  const fetchBackups = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/backups');
      if (!response.ok) {
        throw new Error('Failed to fetch backups');
      }
      const data = await response.json();
      setBackups(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const fetchStorageInfo = async () => {
    try {
      const response = await fetch('/api/v1/storage/info');
      if (!response.ok) {
        throw new Error('Failed to fetch storage info');
      }
      const data = await response.json();
      setStorageInfo(data);
    } catch (err) {
      console.error('Failed to fetch storage info:', err);
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'completed': return '#52c41a';
      case 'failed': return '#ff4d4f';
      case 'running': return '#1890ff';
      case 'cancelled': return '#8c8c8c';
      default: return '#8c8c8c';
    }
  };

  const getTypeIcon = (type: string): string => {
    switch (type) {
      case 'full': return '📦';
      case 'incremental': return '🔄';
      case 'differential': return '📊';
      default: return '📁';
    }
  };

  const filteredBackups = backups.filter(backup => {
    const matchesFilter = filter === 'all' || backup.status === filter;
    const matchesSearch = searchTerm === '' || 
      backup.jobName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      backup.source.toLowerCase().includes(searchTerm.toLowerCase()) ||
      backup.destination.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  const handleDeleteBackup = async (backupId: string) => {
    if (!window.confirm('Are you sure you want to delete this backup?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/backups/${backupId}`, {
        method: 'DELETE',
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete backup');
      }
      
      fetchBackups();
      fetchStorageInfo();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete backup');
    }
  };

  const handleRestore = (backup: Backup) => {
    setSelectedBackup(backup);
    // Navigate to restore page with backup ID
    window.location.href = `/restore?backupId=${backup.id}`;
  };

  if (loading) {
    return (
      <div className="page">
        <div className="page-header">
          <h2>📁 Backup History</h2>
        </div>
        <div className="content">
          <div className="loading">Loading backups...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page">
        <div className="page-header">
          <h2>📁 Backup History</h2>
        </div>
        <div className="content">
          <div className="error">Error: {error}</div>
          <button onClick={fetchBackups} className="retry-button">Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>📁 Backup History</h2>
        {storageInfo && (
          <div className="storage-summary">
            <div className="storage-item">
              <span className="label">Storage:</span>
              <span className="value">{storageInfo.type}</span>
            </div>
            <div className="storage-item">
              <span className="label">Used:</span>
              <span className="value">{formatBytes(storageInfo.usedSize)}</span>
            </div>
            <div className="storage-item">
              <span className="label">Free:</span>
              <span className="value">{formatBytes(storageInfo.freeSize)}</span>
            </div>
            <div className="storage-item">
              <span className="label">Objects:</span>
              <span className="value">{storageInfo.objectCount.toLocaleString()}</span>
            </div>
          </div>
        )}
      </div>
      
      <div className="content">
        <div className="backups-controls">
          <div className="search-filter">
            <input
              type="text"
              placeholder="Search backups..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="search-input"
            />
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="filter-select"
            >
              <option value="all">All Status</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="running">Running</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
          
          <div className="actions">
            <button onClick={fetchBackups} className="refresh-button">
              🔄 Refresh
            </button>
          </div>
        </div>

        {filteredBackups.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">📭</div>
            <h3>No backups found</h3>
            <p>Try adjusting your search or filter criteria.</p>
          </div>
        ) : (
          <div className="backups-grid">
            {filteredBackups.map((backup) => (
              <div key={backup.id} className="backup-card">
                <div className="backup-header">
                  <div className="backup-title">
                    <span className="backup-type-icon">{getTypeIcon(backup.type)}</span>
                    <h3>{backup.jobName}</h3>
                  </div>
                  <div 
                    className="backup-status"
                    style={{ backgroundColor: getStatusColor(backup.status) }}
                  >
                    {backup.status}
                  </div>
                </div>
                
                <div className="backup-details">
                  <div className="detail-row">
                    <span className="label">Source:</span>
                    <span className="value">{backup.source}</span>
                  </div>
                  <div className="detail-row">
                    <span className="label">Destination:</span>
                    <span className="value">{backup.destination}</span>
                  </div>
                  <div className="detail-row">
                    <span className="label">Started:</span>
                    <span className="value">{formatDate(backup.startTime)}</span>
                  </div>
                  {backup.endTime && (
                    <div className="detail-row">
                      <span className="label">Duration:</span>
                      <span className="value">
                        {Math.round((new Date(backup.endTime).getTime() - new Date(backup.startTime).getTime()) / 1000 / 60)} min
                      </span>
                    </div>
                  )}
                </div>

                <div className="backup-stats">
                  <div className="stat-item">
                    <div className="stat-number">{formatBytes(backup.bytesWritten)}</div>
                    <div className="stat-label">Size</div>
                  </div>
                  <div className="stat-item">
                    <div className="stat-number">{backup.filesSuccess}/{backup.filesTotal}</div>
                    <div className="stat-label">Files</div>
                  </div>
                  <div className="stat-item">
                    <div className="stat-number">
                      {backup.bytesRead > 0 ? Math.round((backup.bytesWritten / backup.bytesRead) * 100) : 0}%
                    </div>
                    <div className="stat-label">Ratio</div>
                  </div>
                </div>

                {backup.errorMessage && (
                  <div className="backup-error">
                    <strong>Error:</strong> {backup.errorMessage}
                  </div>
                )}

                <div className="backup-actions">
                  <button 
                    onClick={() => handleRestore(backup)}
                    className="action-button restore-button"
                    disabled={backup.status !== 'completed'}
                  >
                    🔄 Restore
                  </button>
                  <button 
                    onClick={() => handleDeleteBackup(backup.id)}
                    className="action-button delete-button"
                    disabled={backup.status === 'running'}
                  >
                    🗑️ Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Backups;
