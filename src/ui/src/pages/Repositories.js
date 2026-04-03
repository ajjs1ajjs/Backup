import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Button, TextField, Select, MenuItem, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, IconButton, CircularProgress } from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon, CheckCircle as TestIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Repositories() {
  const { data, loading, refetch } = useApi('/api/repositories');
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({ name: '', type: 'local', path: '' });

  const handleTest = async (repoId) => {
    try { await fetchWithAuth(`/api/repositories/${repoId}/test`, { method: 'POST' }); refetch(); } catch (e) { refetch(); }
  };

  const handleDelete = async (repoId) => {
    try { await fetchWithAuth(`/api/repositories/${repoId}`, { method: 'DELETE' }); refetch(); } catch (e) { refetch(); }
  };

  const handleAddRepository = async () => {
    try {
      await fetchWithAuth('/api/repositories', { method: 'POST', body: JSON.stringify(formData) });
      setOpen(false);
      setFormData({ name: '', type: 'local', path: '' });
      refetch();
    } catch (e) { setOpen(false); refetch(); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  const repositories = data?.repositories || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Repositories</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
          Add Repository
        </Button>
      </Box>

      {repositories.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>No repositories configured</Typography>
            <Typography variant="body2" color="text.secondary" mb={2}>Add a storage repository to store your backups</Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>Add Repository</Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Path</TableCell>
                <TableCell>Capacity</TableCell>
                <TableCell>Used</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {repositories.map((repo) => (
                <TableRow key={repo.id || repo.repositoryId}>
                  <TableCell>{repo.name}</TableCell>
                  <TableCell><Chip label={repo.type || 'local'} size="small" /></TableCell>
                  <TableCell sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{repo.path || '-'}</TableCell>
                  <TableCell>{repo.capacityBytes ? (repo.capacityBytes / 1024 / 1024 / 1024).toFixed(1) + ' GB' : '-'}</TableCell>
                  <TableCell>
                    {repo.usedBytes && repo.capacityBytes ? ((repo.usedBytes / repo.capacityBytes) * 100).toFixed(0) + '%' : '-'}
                  </TableCell>
                  <TableCell><Chip label={repo.status || 'unknown'} color={repo.status === 'online' ? 'success' : 'warning'} size="small" /></TableCell>
                  <TableCell>
                    <IconButton onClick={() => handleTest(repo.id || repo.repositoryId)}><TestIcon /></IconButton>
                    <IconButton onClick={() => handleDelete(repo.id || repo.repositoryId)} color="error"><DeleteIcon /></IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {open && (
        <Box sx={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, bgcolor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1300 }} onClick={() => setOpen(false)}>
          <Card sx={{ width: 500, maxWidth: '90%' }} onClick={(e) => e.stopPropagation()}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Add Repository</Typography>
              <Box display="flex" flexDirection="column" gap={2} pt={1}>
                <TextField label="Name" fullWidth value={formData.name} onChange={(e) => setFormData({...formData, name: e.target.value})} />
                <Select fullWidth value={formData.type} onChange={(e) => setFormData({...formData, type: e.target.value})}>
                  <MenuItem value="local">Local Disk</MenuItem>
                  <MenuItem value="nfs">NFS</MenuItem>
                  <MenuItem value="smb">SMB/CIFS</MenuItem>
                  <MenuItem value="s3">Amazon S3</MenuItem>
                  <MenuItem value="azure_blob">Azure Blob</MenuItem>
                  <MenuItem value="gcs">Google Cloud Storage</MenuItem>
                </Select>
                <TextField label="Path" fullWidth value={formData.path} onChange={(e) => setFormData({...formData, path: e.target.value})} />
              </Box>
              <Box display="flex" justifyContent="flex-end" gap={1} sx={{ mt: 2 }}>
                <Button onClick={() => setOpen(false)}>Cancel</Button>
                <Button variant="contained" onClick={handleAddRepository} disabled={!formData.name || !formData.path}>Save</Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
