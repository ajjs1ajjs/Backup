import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  Grid,
  IconButton,
  InputAdornment,
  InputLabel,
  LinearProgress,
  MenuItem,
  Select,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  CheckCircle as TestIcon,
  Cloud as CloudIcon,
  Delete as DeleteIcon,
  Folder as FolderIcon,
  Language as NetworkIcon,
  Refresh as RefreshIcon,
  Storage as StorageIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const typeConfig = {
  local: { icon: <FolderIcon />, color: '#4fc3f7', label: 'Local Disk' },
  nfs: { icon: <NetworkIcon />, color: '#66bb6a', label: 'NFS' },
  smb: { icon: <NetworkIcon />, color: '#ffa726', label: 'SMB/CIFS' },
  s3: { icon: <CloudIcon />, color: '#ab47bc', label: 'Amazon S3' },
  azure_blob: { icon: <CloudIcon />, color: '#2196F3', label: 'Azure Blob' },
  gcs: { icon: <CloudIcon />, color: '#FF5722', label: 'Google Cloud Storage' }
};

const statusColors = {
  online: 'success',
  offline: 'default',
  error: 'error',
  maintenance: 'warning'
};

const initialForm = {
  name: '',
  type: 'local',
  path: '',
  host: '',
  port: '',
  share: '',
  username: '',
  password: '',
  bucket: '',
  region: '',
  accessKey: '',
  secretKey: '',
  storageAccount: '',
  container: '',
  accessKey2: '',
  capacityBytes: ''
};

export default function Repositories() {
  const [repos, setRepos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [testing, setTesting] = useState(null);
  const [formData, setFormData] = useState(initialForm);

  useEffect(() => {
    fetchRepos();
  }, []);

  const fetchRepos = async () => {
    setLoading(true);
    try {
      const response = await fetchWithAuth('/api/repositories');
      const data = await response.json().catch(() => []);
      setRepos(Array.isArray(data) ? data : []);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => setFormData(initialForm);

  const handleTest = async (repoId) => {
    setTesting(repoId);
    try {
      await fetchWithAuth(`/api/repositories/${repoId}/test`, { method: 'POST' });
    } finally {
      setTesting(null);
      fetchRepos();
    }
  };

  const handleDelete = async (repoId) => {
    if (!window.confirm('Delete this repository? Backup data will not be deleted.')) return;
    try {
      await fetchWithAuth(`/api/repositories/${repoId}`, { method: 'DELETE' });
    } finally {
      fetchRepos();
    }
  };

  const buildPath = () => {
    if (formData.type === 'nfs') return `${formData.host}:${formData.share || formData.path}`;
    if (formData.type === 'smb') return `\\\\${formData.host}\\${formData.share || formData.path}`;
    if (formData.type === 's3') return `s3://${formData.bucket}${formData.region ? ` (${formData.region})` : ''}`;
    if (formData.type === 'azure_blob') return `azure://${formData.storageAccount}/${formData.container}`;
    if (formData.type === 'gcs') return `gcs://${formData.bucket}`;
    return formData.path;
  };

  const handleAdd = async () => {
    const capacityBytes = formData.capacityBytes
      ? parseInt(formData.capacityBytes, 10) * 1024 * 1024 * 1024
      : null;

    const credentials =
      formData.type === 'smb'
        ? JSON.stringify({ username: formData.username, password: formData.password })
        : formData.type === 's3'
          ? JSON.stringify({ accessKey: formData.accessKey, secretKey: formData.secretKey, region: formData.region })
          : formData.type === 'azure_blob'
            ? JSON.stringify({ storageAccount: formData.storageAccount, accessKey: formData.accessKey2 })
            : formData.type === 'gcs'
              ? JSON.stringify({ bucket: formData.bucket })
              : null;

    const body = {
      name: formData.name,
      type: formData.type,
      path: buildPath(),
      capacityBytes,
      status: 'online',
      credentials,
      options: JSON.stringify({})
    };

    try {
      await fetchWithAuth('/api/repositories', { method: 'POST', body: JSON.stringify(body) });
      setOpen(false);
      resetForm();
      fetchRepos();
    } catch (error) {
      console.error(error);
    }
  };

  const formatBytes = (bytes) => {
    if (!bytes) return '-';
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  };

  if (loading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">Backup Repositories</Typography>
        <Box display="flex" gap={1}>
          <Button variant="outlined" startIcon={<RefreshIcon />} onClick={fetchRepos}>Refresh</Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => { resetForm(); setOpen(true); }}>
            Add Repository
          </Button>
        </Box>
      </Box>

      {repos.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <StorageIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No repositories configured</Typography>
            <Typography variant="body2" color="text.secondary" mb={2}>
              Add a repository to store backups.
            </Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
              Add Repository
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NAME</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>PATH / ENDPOINT</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>CAPACITY</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>USED</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ACTIONS</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {repos.map((repo) => {
                const pct = repo.capacityBytes > 0 ? Math.round((repo.usedBytes / repo.capacityBytes) * 100) : 0;
                const config = typeConfig[repo.type] || typeConfig.local;

                return (
                  <TableRow key={repo.repositoryId || repo.id} hover>
                    <TableCell>
                      <Chip label={repo.status || 'unknown'} size="small" color={statusColors[repo.status] || 'default'} />
                    </TableCell>
                    <TableCell sx={{ fontWeight: 500 }}>{repo.name}</TableCell>
                    <TableCell>
                      <Chip
                        icon={config.icon}
                        label={config.label}
                        size="small"
                        sx={{ bgcolor: `${config.color}20`, color: config.color, fontWeight: 'bold' }}
                      />
                    </TableCell>
                    <TableCell sx={{ maxWidth: 300, fontFamily: 'monospace', fontSize: '0.85rem' }}>{repo.path || '-'}</TableCell>
                    <TableCell>{formatBytes(repo.capacityBytes)}</TableCell>
                    <TableCell sx={{ minWidth: 150 }}>
                      {repo.capacityBytes > 0 ? (
                        <Box>
                          <Box display="flex" justifyContent="space-between" mb={0.5}>
                            <Typography variant="caption">{formatBytes(repo.usedBytes)}</Typography>
                            <Typography variant="caption" color="text.secondary">{pct}%</Typography>
                          </Box>
                          <LinearProgress
                            variant="determinate"
                            value={pct}
                            sx={{
                              height: 4,
                              borderRadius: 2,
                              bgcolor: '#e8eaed',
                              '& .MuiLinearProgress-bar': {
                                bgcolor: pct > 90 ? '#ef5350' : pct > 70 ? '#ffa726' : '#66bb6a'
                              }
                            }}
                          />
                        </Box>
                      ) : '-'}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Test Connection">
                        <IconButton size="small" onClick={() => handleTest(repo.repositoryId)} disabled={testing === repo.repositoryId}>
                          {testing === repo.repositoryId ? <CircularProgress size={18} /> : <TestIcon fontSize="small" sx={{ color: '#66bb6a' }} />}
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton size="small" onClick={() => handleDelete(repo.repositoryId)}>
                          <DeleteIcon fontSize="small" sx={{ color: '#ef5350' }} />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Repository</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 0.5 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Name"
                value={formData.name}
                onChange={(event) => setFormData({ ...formData, name: event.target.value })}
                size="small"
                placeholder="e.g. Backup Storage 1"
              />
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth size="small">
                <InputLabel>Repository Type</InputLabel>
                <Select
                  value={formData.type}
                  label="Repository Type"
                  onChange={(event) => setFormData({ ...formData, type: event.target.value })}
                >
                  <MenuItem value="local">Local Disk</MenuItem>
                  <MenuItem value="nfs">NFS Share</MenuItem>
                  <MenuItem value="smb">SMB/CIFS Share</MenuItem>
                  <MenuItem value="s3">Amazon S3 / S3-Compatible</MenuItem>
                  <MenuItem value="azure_blob">Azure Blob Storage</MenuItem>
                  <MenuItem value="gcs">Google Cloud Storage</MenuItem>
                </Select>
              </FormControl>
            </Grid>

            {formData.type === 'local' && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Local Path"
                  value={formData.path}
                  onChange={(event) => setFormData({ ...formData, path: event.target.value })}
                  size="small"
                  placeholder="D:\\Backups or /mnt/backups"
                />
              </Grid>
            )}

            {(formData.type === 'nfs' || formData.type === 'smb') && (
              <>
                <Grid item xs={12} sm={8}>
                  <TextField
                    fullWidth
                    label="Server / IP"
                    value={formData.host}
                    onChange={(event) => setFormData({ ...formData, host: event.target.value })}
                    size="small"
                    placeholder="192.168.1.100"
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <TextField
                    fullWidth
                    label="Port"
                    type="number"
                    value={formData.port}
                    onChange={(event) => setFormData({ ...formData, port: event.target.value })}
                    size="small"
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label={formData.type === 'nfs' ? 'Export Path' : 'Share Name'}
                    value={formData.share}
                    onChange={(event) => setFormData({ ...formData, share: event.target.value })}
                    size="small"
                    placeholder={formData.type === 'nfs' ? '/exports/backups' : 'Backups'}
                  />
                </Grid>
                {formData.type === 'smb' && (
                  <>
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        label="Username"
                        value={formData.username}
                        onChange={(event) => setFormData({ ...formData, username: event.target.value })}
                        size="small"
                        placeholder="DOMAIN\\user"
                      />
                    </Grid>
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        label="Password"
                        type="password"
                        value={formData.password}
                        onChange={(event) => setFormData({ ...formData, password: event.target.value })}
                        size="small"
                      />
                    </Grid>
                  </>
                )}
              </>
            )}

            {formData.type === 's3' && (
              <>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Bucket Name"
                    value={formData.bucket}
                    onChange={(event) => setFormData({ ...formData, bucket: event.target.value })}
                    size="small"
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Region"
                    value={formData.region}
                    onChange={(event) => setFormData({ ...formData, region: event.target.value })}
                    size="small"
                    placeholder="us-east-1"
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Access Key"
                    value={formData.accessKey}
                    onChange={(event) => setFormData({ ...formData, accessKey: event.target.value })}
                    size="small"
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Secret Key"
                    type="password"
                    value={formData.secretKey}
                    onChange={(event) => setFormData({ ...formData, secretKey: event.target.value })}
                    size="small"
                  />
                </Grid>
              </>
            )}

            {formData.type === 'azure_blob' && (
              <>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Storage Account"
                    value={formData.storageAccount}
                    onChange={(event) => setFormData({ ...formData, storageAccount: event.target.value })}
                    size="small"
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    label="Container"
                    value={formData.container}
                    onChange={(event) => setFormData({ ...formData, container: event.target.value })}
                    size="small"
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Access Key"
                    type="password"
                    value={formData.accessKey2}
                    onChange={(event) => setFormData({ ...formData, accessKey2: event.target.value })}
                    size="small"
                  />
                </Grid>
              </>
            )}

            {formData.type === 'gcs' && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Bucket Name"
                  value={formData.bucket}
                  onChange={(event) => setFormData({ ...formData, bucket: event.target.value })}
                  size="small"
                />
              </Grid>
            )}

            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Capacity (GB)"
                type="number"
                value={formData.capacityBytes}
                onChange={(event) => setFormData({ ...formData, capacityBytes: event.target.value })}
                size="small"
                InputProps={{ endAdornment: <InputAdornment position="end">GB</InputAdornment> }}
                helperText="Optional - used for capacity monitoring"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleAdd}
            disabled={!formData.name}
            sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}
          >
            Add Repository
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
