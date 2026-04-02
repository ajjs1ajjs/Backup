import React from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, IconButton, Button } from '@mui/material';
import { Delete as DeleteIcon, Download as DownloadIcon, Verify as VerifyIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Backups() {
  const { data, loading, refetch } = useApi('/api/backups');

  const handleVerify = async (backupId: string) => {
    await fetch(`/api/backups/${backupId}/verify`, { method: 'POST' });
    refetch();
  };

  const handleDelete = async (backupId: string) => {
    await fetch(`/api/backups/${backupId}`, { method: 'DELETE' });
    refetch();
  };

  if (loading) return <Typography>Loading...</Typography>;

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Backups</Typography>
      
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
            {data?.backups?.map((backup: any) => (
              <TableRow key={backup.backupId}>
                <TableCell>{backup.backupId.substring(0, 8)}...</TableCell>
                <TableCell><Chip label={backup.backupType} size="small" /></TableCell>
                <TableCell>{backup.repositoryId?.substring(0, 8)}...</TableCell>
                <TableCell>{(backup.sizeBytes / 1024 / 1024).toFixed(2)} MB</TableCell>
                <TableCell>
                  <Chip 
                    label={backup.status} 
                    color={backup.status === 'completed' ? 'success' : backup.status === 'failed' ? 'error' : 'warning'}
                    size="small" 
                  />
                </TableCell>
                <TableCell>{new Date(backup.createdAt).toLocaleString()}</TableCell>
                <TableCell>
                  <IconButton onClick={() => handleVerify(backup.backupId)}><VerifyIcon /></IconButton>
                  <IconButton><DownloadIcon /></IconButton>
                  <IconButton onClick={() => handleDelete(backup.backupId)} color="error"><DeleteIcon /></IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
