import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  Button,
  TextField,
  Select,
  MenuItem,
  LinearProgress,
  CircularProgress,
  Alert
} from '@mui/material';
import { Restore as RestoreIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Restore() {
  const { data, loading, refetch } = useApi('/api/restore');
  const { data: backups, refetch: refetchBackups } = useApi('/api/backups');
  const [open, setOpen] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState('');
  const [restoreType, setRestoreType] = useState('full_vm');
  const [targetHost, setTargetHost] = useState('');
  const [destinationPath, setDestinationPath] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const handleStartRestore = async () => {
    if (!selectedBackup) {
      return;
    }

    try {
      setError('');
      setMessage('');

      const response = await fetchWithAuth('/api/restore', {
        method: 'POST',
        body: JSON.stringify({
          backupId: selectedBackup,
          restoreType: restoreType === 'instant' ? 'instant_restore' : restoreType,
          targetHost,
          destinationPath
        })
      });

      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(payload?.message || payload?.error || 'Restore failed to start.');
      }

      setMessage(payload?.message || 'Restore started successfully.');
      setOpen(false);
      setSelectedBackup('');
      setTargetHost('');
      setDestinationPath('');
      refetch();
      refetchBackups();
    } catch (e) {
      setError(e.message || 'Restore failed to start.');
      setOpen(false);
    }
  };

  if (loading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

  const restores = data?.restores || data || [];
  const backupList = backups?.backups || backups || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Restore</Typography>
        <Button variant="contained" startIcon={<RestoreIcon />} onClick={() => setOpen(true)}>
          Start Restore
        </Button>
      </Box>

      {message && <Alert severity="success" sx={{ mb: 2 }}>{message}</Alert>}
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      {restores.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>No restore operations yet</Typography>
            <Typography variant="body2" color="text.secondary">Start a restore to monitor progress here.</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Restore ID</TableCell>
                <TableCell>Backup ID</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Progress</TableCell>
                <TableCell>Destination</TableCell>
                <TableCell>Created</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {restores.map((restore) => {
                const restoreId = restore.restoreId || restore.id;
                const progress = typeof restore.progress === 'number'
                  ? restore.progress
                  : restore.totalBytes > 0
                    ? Math.round((restore.bytesRestored || 0) / restore.totalBytes * 100)
                    : 0;

                return (
                  <TableRow key={restoreId}>
                    <TableCell>{String(restoreId || '').substring(0, 8)}...</TableCell>
                    <TableCell>{String(restore.backupId || '').substring(0, 8)}...</TableCell>
                    <TableCell><Chip label={restore.restoreType || 'full_vm'} size="small" /></TableCell>
                    <TableCell>
                      <Chip
                        label={restore.status || 'pending'}
                        color={
                          restore.status === 'completed' ? 'success' :
                          restore.status === 'failed' ? 'error' :
                          restore.status === 'cancelled' ? 'default' :
                          'warning'
                        }
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Box display="flex" alignItems="center" gap={1}>
                        <LinearProgress variant="determinate" value={progress} sx={{ width: 100 }} />
                        <Typography variant="caption">{progress}%</Typography>
                      </Box>
                    </TableCell>
                    <TableCell>{restore.destinationPath || '-'}</TableCell>
                    <TableCell>{restore.createdAt ? new Date(restore.createdAt).toLocaleString() : '-'}</TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {open && (
        <Box
          sx={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            bgcolor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1300
          }}
          onClick={() => setOpen(false)}
        >
          <Card sx={{ width: 520, maxWidth: '90%' }} onClick={(e) => e.stopPropagation()}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Start Restore</Typography>
              <Box display="flex" flexDirection="column" gap={2} pt={1}>
                <Select fullWidth value={selectedBackup} onChange={(e) => setSelectedBackup(e.target.value)} displayEmpty>
                  <MenuItem value="" disabled>Select backup</MenuItem>
                  {backupList.map((backup) => (
                    <MenuItem key={backup.backupId || backup.id} value={backup.backupId}>
                      {String(backup.backupId || '').substring(0, 8)} - {backup.backupType || 'backup'}
                    </MenuItem>
                  ))}
                </Select>
                <Select fullWidth value={restoreType} onChange={(e) => setRestoreType(e.target.value)}>
                  <MenuItem value="full_vm">Full VM Restore</MenuItem>
                  <MenuItem value="instant">Instant Restore</MenuItem>
                  <MenuItem value="file_level">File-Level Restore</MenuItem>
                </Select>
                <TextField
                  label={restoreType === 'instant' ? 'Mount Path' : 'Destination Path'}
                  fullWidth
                  value={destinationPath}
                  onChange={(e) => setDestinationPath(e.target.value)}
                />
                <TextField
                  label="Target Host"
                  fullWidth
                  value={targetHost}
                  onChange={(e) => setTargetHost(e.target.value)}
                />
              </Box>
              <Box display="flex" justifyContent="flex-end" gap={1} sx={{ mt: 2 }}>
                <Button onClick={() => setOpen(false)}>Cancel</Button>
                <Button variant="contained" onClick={handleStartRestore} disabled={!selectedBackup || !destinationPath}>
                  Start
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
