import React, { useState, useEffect } from 'react';
import './Storage.css';

interface StorageBackend {
  id: string;
  name: string;
  type: 'local' | 's3' | 'nfs' | 'smb';
  status: 'online' | 'offline' | 'error';
  totalSize: number;
  usedSize: number;
  freeSize: number;
  objectCount: number;
  endpoint?: string;
  bucket?: string;
  region?: string;
  mountPoint?: string;
  lastUpdated: string;
}

interface ChunkInfo {
  hash: string;
  size: number;
  compressedSize: number;
  storagePath: string;
  createdAt: string;
}

const Storage: React.FC = () => {
  const [backends, setBackends] = useState<StorageBackend[]>([]);
  const [selectedBackend, setSelectedBackend] = useState<StorageBackend | null>(null);
  const [chunks, setChunks] = useState<ChunkInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddBackend, setShowAddBackend] = useState(false);
  const [backendForm, setBackendForm] = useState({
    name: '',
    type: 'local' as 'local' | 's3' | 'nfs' | 'smb',
    endpoint: '',
    bucket: '',
    region: '',
    mountPoint: '',
    accessKey: '',
    secretKey: '',
    username: '',
    password: '',
    domain: '',
  });

  useEffect(() => {
    fetchBackends();
  }, []);

  const fetchBackends = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/storage/backends');
      if (!response.ok) {
        throw new Error('Failed to fetch storage backends');
      }
      const data = await response.json();
      setBackends(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const fetchChunks = async (backendId: string) => {
    try {
      const response = await fetch(`/api/v1/storage/backends/${backendId}/chunks`);
      if (!response.ok) {
        throw new Error('Failed to fetch chunks');
      }
      const data = await response.json();
      setChunks(data);
    } catch (err) {
      console.error('Failed to fetch chunks:', err);
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'online': return '#52c41a';
      case 'offline': return '#8c8c8c';
      case 'error': return '#ff4d4f';
      default: return '#8c8c8c';
    }
  };

  const getTypeIcon = (type: string): string => {
    switch (type) {
      case 'local': return '💻';
      case 's3': return '☁️';
      case 'nfs': return '🌐';
      case 'smb': return '🔗';
      default: return '💾';
    }
  };

  const handleBackendClick = (backend: StorageBackend) => {
    setSelectedBackend(backend);
    fetchChunks(backend.id);
  };

  const handleAddBackend = async () => {
    try {
      const response = await fetch('/api/v1/storage/backends', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(backendForm),
      });

      if (!response.ok) {
        throw new Error('Failed to add storage backend');
      }

      setShowAddBackend(false);
      setBackendForm({
        name: '',
        type: 'local',
        endpoint: '',
        bucket: '',
        region: '',
        mountPoint: '',
        accessKey: '',
        secretKey: '',
        username: '',
        password: '',
        domain: '',
      });
      fetchBackends();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add storage backend');
    }
  };

  const handleDeleteBackend = async (backendId: string) => {
    if (!window.confirm('Are you sure you want to delete this storage backend?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/storage/backends/${backendId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete storage backend');
      }

      if (selectedBackend?.id === backendId) {
        setSelectedBackend(null);
        setChunks([]);
      }
      fetchBackends();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete storage backend');
    }
  };

  const handleTestConnection = async (backendId: string) => {
    try {
      const response = await fetch(`/api/v1/storage/backends/${backendId}/test`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Connection test failed');
      }

      alert('Connection test successful!');
      fetchBackends();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Connection test failed');
    }
  };

  const handleCleanup = async (backendId: string) => {
    if (!window.confirm('Are you sure you want to cleanup old chunks?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/storage/backends/${backendId}/cleanup`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Cleanup failed');
      }

      alert('Cleanup completed successfully!');
      fetchBackends();
      if (selectedBackend?.id === backendId) {
        fetchChunks(backendId);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Cleanup failed');
    }
  };

  if (loading) {
    return (
      <div className="page">
        <div className="page-header">
          <h2>💾 Storage Management</h2>
        </div>
        <div className="content">
          <div className="loading">Loading storage information...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page">
        <div className="page-header">
          <h2>💾 Storage Management</h2>
        </div>
        <div className="content">
          <div className="error">Error: {error}</div>
          <button onClick={fetchBackends} className="retry-button">Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>💾 Storage Management</h2>
        <button 
          onClick={() => setShowAddBackend(true)}
          className="add-backend-button"
        >
          ➕ Add Backend
        </button>
      </div>
      
      <div className="content">
        <div className="storage-grid">
          {backends.map((backend) => (
            <div 
              key={backend.id} 
              className={`storage-card ${selectedBackend?.id === backend.id ? 'selected' : ''}`}
              onClick={() => handleBackendClick(backend)}
            >
              <div className="storage-header">
                <div className="storage-title">
                  <span className="storage-type-icon">{getTypeIcon(backend.type)}</span>
                  <h3>{backend.name}</h3>
                </div>
                <div 
                  className="storage-status"
                  style={{ backgroundColor: getStatusColor(backend.status) }}
                >
                  {backend.status}
                </div>
              </div>
              
              <div className="storage-details">
                <div className="detail-row">
                  <span className="label">Type:</span>
                  <span className="value">{backend.type.toUpperCase()}</span>
                </div>
                {backend.endpoint && (
                  <div className="detail-row">
                    <span className="label">Endpoint:</span>
                    <span className="value">{backend.endpoint}</span>
                  </div>
                )}
                {backend.bucket && (
                  <div className="detail-row">
                    <span className="label">Bucket:</span>
                    <span className="value">{backend.bucket}</span>
                  </div>
                )}
                {backend.mountPoint && (
                  <div className="detail-row">
                    <span className="label">Mount:</span>
                    <span className="value">{backend.mountPoint}</span>
                  </div>
                )}
                <div className="detail-row">
                  <span className="label">Objects:</span>
                  <span className="value">{backend.objectCount.toLocaleString()}</span>
                </div>
                <div className="detail-row">
                  <span className="label">Updated:</span>
                  <span className="value">{new Date(backend.lastUpdated).toLocaleString()}</span>
                </div>
              </div>

              <div className="storage-usage">
                <div className="usage-bar">
                  <div 
                    className="usage-used"
                    style={{ 
                      width: `${(backend.usedSize / backend.totalSize) * 100}%`,
                      backgroundColor: backend.status === 'error' ? '#ff4d4f' : '#1890ff'
                    }}
                  ></div>
                </div>
                <div className="usage-stats">
                  <span>Used: {formatBytes(backend.usedSize)}</span>
                  <span>Free: {formatBytes(backend.freeSize)}</span>
                </div>
              </div>

              <div className="storage-actions">
                <button 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleTestConnection(backend.id);
                  }}
                  className="action-button test-button"
                >
                  🔍 Test
                </button>
                <button 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCleanup(backend.id);
                  }}
                  className="action-button cleanup-button"
                  disabled={backend.status !== 'online'}
                >
                  🧹 Cleanup
                </button>
                <button 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeleteBackend(backend.id);
                  }}
                  className="action-button delete-button"
                >
                  🗑️ Delete
                </button>
              </div>
            </div>
          ))}
        </div>

        {selectedBackend && (
          <div className="chunks-section">
            <div className="chunks-header">
              <h3>📦 Chunks in {selectedBackend.name}</h3>
              <div className="chunks-stats">
                <span>Total: {chunks.length.toLocaleString()}</span>
                <span>Size: {formatBytes(chunks.reduce((sum, chunk) => sum + chunk.size, 0))}</span>
              </div>
            </div>
            
            <div className="chunks-grid">
              {chunks.slice(0, 50).map((chunk) => (
                <div key={chunk.hash} className="chunk-card">
                  <div className="chunk-header">
                    <span className="chunk-hash">{chunk.hash.substring(0, 16)}...</span>
                    <span className="chunk-size">{formatBytes(chunk.size)}</span>
                  </div>
                  <div className="chunk-details">
                    <div className="detail-row">
                      <span className="label">Compressed:</span>
                      <span className="value">{formatBytes(chunk.compressedSize)}</span>
                    </div>
                    <div className="detail-row">
                      <span className="label">Ratio:</span>
                      <span className="value">
                        {Math.round((1 - chunk.compressedSize / chunk.size) * 100)}%
                      </span>
                    </div>
                    <div className="detail-row">
                      <span className="label">Created:</span>
                      <span className="value">{new Date(chunk.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            
            {chunks.length > 50 && (
              <div className="chunks-more">
                <p>Showing 50 of {chunks.length.toLocaleString()} chunks</p>
              </div>
            )}
          </div>
        )}

        {showAddBackend && (
          <div className="modal-overlay">
            <div className="modal">
              <div className="modal-header">
                <h3>Add Storage Backend</h3>
                <button 
                  onClick={() => setShowAddBackend(false)}
                  className="close-button"
                >
                  ✕
                </button>
              </div>
              
              <div className="modal-content">
                <div className="form-group">
                  <label>Name:</label>
                  <input
                    type="text"
                    value={backendForm.name}
                    onChange={(e) => setBackendForm({...backendForm, name: e.target.value})}
                    placeholder="Backend name"
                  />
                </div>
                
                <div className="form-group">
                  <label>Type:</label>
                  <select
                    value={backendForm.type}
                    onChange={(e) => setBackendForm({...backendForm, type: e.target.value as any})}
                  >
                    <option value="local">Local</option>
                    <option value="s3">S3</option>
                    <option value="nfs">NFS</option>
                    <option value="smb">SMB</option>
                  </select>
                </div>

                {backendForm.type === 's3' && (
                  <>
                    <div className="form-group">
                      <label>Endpoint:</label>
                      <input
                        type="text"
                        value={backendForm.endpoint}
                        onChange={(e) => setBackendForm({...backendForm, endpoint: e.target.value})}
                        placeholder="s3.amazonaws.com"
                      />
                    </div>
                    <div className="form-group">
                      <label>Bucket:</label>
                      <input
                        type="text"
                        value={backendForm.bucket}
                        onChange={(e) => setBackendForm({...backendForm, bucket: e.target.value})}
                        placeholder="my-backup-bucket"
                      />
                    </div>
                    <div className="form-group">
                      <label>Region:</label>
                      <input
                        type="text"
                        value={backendForm.region}
                        onChange={(e) => setBackendForm({...backendForm, region: e.target.value})}
                        placeholder="us-east-1"
                      />
                    </div>
                    <div className="form-group">
                      <label>Access Key:</label>
                      <input
                        type="text"
                        value={backendForm.accessKey}
                        onChange={(e) => setBackendForm({...backendForm, accessKey: e.target.value})}
                        placeholder="AKIAIOSFODNN7EXAMPLE"
                      />
                    </div>
                    <div className="form-group">
                      <label>Secret Key:</label>
                      <input
                        type="password"
                        value={backendForm.secretKey}
                        onChange={(e) => setBackendForm({...backendForm, secretKey: e.target.value})}
                        placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                      />
                    </div>
                  </>
                )}

                {(backendForm.type === 'nfs' || backendForm.type === 'smb') && (
                  <>
                    <div className="form-group">
                      <label>Host:</label>
                      <input
                        type="text"
                        value={backendForm.endpoint}
                        onChange={(e) => setBackendForm({...backendForm, endpoint: e.target.value})}
                        placeholder="server.example.com"
                      />
                    </div>
                    <div className="form-group">
                      <label>Share:</label>
                      <input
                        type="text"
                        value={backendForm.bucket}
                        onChange={(e) => setBackendForm({...backendForm, bucket: e.target.value})}
                        placeholder="/backup"
                      />
                    </div>
                    <div className="form-group">
                      <label>Mount Point:</label>
                      <input
                        type="text"
                        value={backendForm.mountPoint}
                        onChange={(e) => setBackendForm({...backendForm, mountPoint: e.target.value})}
                        placeholder="/mnt/nova-backup"
                      />
                    </div>
                    {backendForm.type === 'smb' && (
                      <>
                        <div className="form-group">
                          <label>Username:</label>
                          <input
                            type="text"
                            value={backendForm.username}
                            onChange={(e) => setBackendForm({...backendForm, username: e.target.value})}
                            placeholder="backupuser"
                          />
                        </div>
                        <div className="form-group">
                          <label>Password:</label>
                          <input
                            type="password"
                            value={backendForm.password}
                            onChange={(e) => setBackendForm({...backendForm, password: e.target.value})}
                            placeholder="password"
                          />
                        </div>
                        <div className="form-group">
                          <label>Domain:</label>
                          <input
                            type="text"
                            value={backendForm.domain}
                            onChange={(e) => setBackendForm({...backendForm, domain: e.target.value})}
                            placeholder="DOMAIN"
                          />
                        </div>
                      </>
                    )}
                  </>
                )}

                {backendForm.type === 'local' && (
                  <div className="form-group">
                    <label>Storage Path:</label>
                    <input
                      type="text"
                      value={backendForm.mountPoint}
                      onChange={(e) => setBackendForm({...backendForm, mountPoint: e.target.value})}
                      placeholder="/var/lib/nova-backup"
                    />
                  </div>
                )}
              </div>
              
              <div className="modal-footer">
                <button onClick={() => setShowAddBackend(false)} className="cancel-button">
                  Cancel
                </button>
                <button onClick={handleAddBackend} className="submit-button">
                  Add Backend
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Storage;
