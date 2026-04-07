import React from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, IconButton, CircularProgress } from '@mui/material';
import { Delete as DeleteIcon, CheckCircle as VerifyIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Backups() {
  const { data, loading, refetch } = useApi('/api/backups');

  const handleVerify = async (backupId) => {
    try { await fetchWithAuth(`/api/backups/${backupId}/verify`, { method: 'POST' }); refetch(); } catch (e) { refetch(); }
  };

  const handleDelete = async (backupId) => {
    try { await fetchWithAuth(`/api/backups/${backupId}`, { method: 'DELETE' }); refetch(); } catch (e) { refetch(); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  const backups = data?.backups || [];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Backups</Typography>
      
      {backups.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>No backups found</Typography>
            <Typography variant="body2" color="text.secondary">Run a backup job to see results here</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Backup ID</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Repository</TableCell>
                <TableCell>Size</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Created</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {backups.map((backup) => (
                <TableRow key={backup.id || backup.backupId}>
                  <TableCell>{(backup.id || backup.backupId || '').substring(0, 8)}...</TableCell>
                  <TableCell><Chip label={String(backup.backupType || 'Full').replace('_', ' ')} size="small" /></TableCell>
                  <TableCell>{backup.repositoryId ? String(backup.repositoryId).substring(0, 8) + '...' : '-'}</TableCell>
                  <TableCell>{backup.sizeBytes ? (backup.sizeBytes / 1024 / 1024).toFixed(2) + ' MB' : '-'}</TableCell>
                  <TableCell>
                    <Chip 
                      label={String(backup.status || 'Pending')} 
                      color={
                        String(backup.status).toLowerCase() === 'completed' || String(backup.status) === 'Completed' ? 'success' : 
                        String(backup.status).toLowerCase() === 'failed' || String(backup.status) === 'Failed' ? 'error' : 
                        String(backup.status).toLowerCase() === 'inprogress' || String(backup.status) === 'InProgress' ? 'info' :
                        'warning'
                      } 
                      size="small" 
                    />
                  </TableCell>
                  <TableCell>{backup.createdAt ? new Date(backup.createdAt).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <IconButton onClick={() => handleVerify(backup.id || backup.backupId)}><VerifyIcon /></IconButton>
                    <IconButton onClick={() => handleDelete(backup.id || backup.backupId)} color="error"><DeleteIcon /></IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
