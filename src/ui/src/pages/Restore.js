import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, TextField, Select, MenuItem, Dialog, DialogTitle, DialogContent, DialogActions, LinearProgress } from '@mui/material';
import { Restore as RestoreIcon, Cancel as CancelIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Restore() {
  const { data, loading } = useApi('/api/restore');
  const { data: backups } = useApi('/api/backups');
  const [open, setOpen] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState('');

  const handleStartRestore = async () => {
    if (!selectedBackup) return;
    await fetch('/api/restore', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ backupId: selectedBackup, restoreType: 'full_vm' })
    });
    setOpen(false);
  };

  if (loading) return <Typography>Loading...</Typography>;

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Restore</Typography>
        <Button variant="contained" startIcon={<RestoreIcon />} onClick={() => setOpen(true)}>
          Start Restore
        </Button>
      </Box>

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Restore ID</TableCell>
              <TableCell>Backup ID</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Progress</TableCell>
              <TableCell>Created</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.restores?.map((restore: any) => (
              <TableRow key={restore.restoreId}>
                <TableCell>{restore.restoreId.substring(0, 8)}...</TableCell>
                <TableCell>{restore.backupId.substring(0, 8)}...</TableCell>
                <TableCell><Chip label={restore.restoreType} size="small" /></TableCell>
                <TableCell>
                  <Chip 
                    label={restore.status} 
                    color={restore.status === 'completed' ? 'success' : restore.status === 'failed' ? 'error' : 'warning'}
                    size="small" 
                  />
                </TableCell>
                <TableCell>
                  <Box display="flex" alignItems="center" gap={1}>
                    <LinearProgress variant="determinate" value={restore.progress || 0} sx={{ width: 100 }} />
                    <Typography variant="caption">{restore.progress || 0}%</Typography>
                  </Box>
                </TableCell>
                <TableCell>{new Date(restore.createdAt).toLocaleString()}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Start Restore</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} pt={2}>
            <Select 
              fullWidth 
              value={selectedBackup} 
              onChange={(e) => setSelectedBackup(e.target.value)}
              displayEmpty
            >
              <MenuItem value="" disabled>Select Backup</MenuItem>
              {backups?.backups?.map((b: any) => (
                <MenuItem key={b.backupId} value={b.backupId}>
                  {b.backupId.substring(0, 8)} - {b.backupType}
                </MenuItem>
              ))}
            </Select>
            <Select fullWidth defaultValue="full_vm">
              <MenuItem value="full_vm">Full VM Restore</MenuItem>
              <MenuItem value="instant">Instant Restore</MenuItem>
              <MenuItem value="file_level">File Level Restore</MenuItem>
            </Select>
            <TextField label="Target Host" fullWidth />
            <TextField label="Target Path" fullWidth />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleStartRestore}>Start</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
