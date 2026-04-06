import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, TextField, Select, MenuItem, Dialog, DialogTitle, DialogContent, DialogActions, LinearProgress, CircularProgress } from '@mui/material';
import { Restore as RestoreIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Restore() {
  const { data, loading } = useApi('/api/restore');
  const { data: backups } = useApi('/api/backups');
  const [open, setOpen] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState('');
  const [restoreType, setRestoreType] = useState('full_vm');
  const [targetHost, setTargetHost] = useState('');

  const handleStartRestore = async () => {
    if (!selectedBackup) return;
    try {
      await fetchWithAuth('/api/restore', {
        method: 'POST',
        body: JSON.stringify({ backupId: selectedBackup, restoreType, targetHost })
      });
      setOpen(false);
      setSelectedBackup('');
      setTargetHost('');
    } catch (e) { setOpen(false); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  const restores = data?.restores || data || [];
  const backupList = backups?.backups || backups || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Відновлення</Typography>
        <Button variant="contained" startIcon={<RestoreIcon />} onClick={() => setOpen(true)}>
          Розпочати відновлення
        </Button>
      </Box>

      {restores.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>Немає операцій відновлення</Typography>
            <Typography variant="body2" color="text.secondary">Запустіть відновлення, щоб побачити прогрес тут</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID відновлення</TableCell>
                <TableCell>ID бекапу</TableCell>
                <TableCell>Тип</TableCell>
                <TableCell>Статус</TableCell>
                <TableCell>Прогрес</TableCell>
                <TableCell>Створено</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {restores.map((restore) => (
                <TableRow key={restore.id || restore.restoreId}>
                  <TableCell>{(restore.id || restore.restoreId).substring(0, 8)}...</TableCell>
                  <TableCell>{(restore.backupId || '').substring(0, 8)}...</TableCell>
                  <TableCell><Chip label={restore.restoreType || 'full_vm'} size="small" /></TableCell>
                  <TableCell>
                    <Chip label={restore.status || 'pending'} color={restore.status === 'completed' ? 'success' : restore.status === 'failed' ? 'error' : 'warning'} size="small" />
                  </TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      <LinearProgress variant="determinate" value={restore.progress || 0} sx={{ width: 100 }} />
                      <Typography variant="caption">{restore.progress || 0}%</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>{restore.createdAt ? new Date(restore.createdAt).toLocaleString() : '-'}</TableCell>
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
              <Typography variant="h6" gutterBottom>Розпочати відновлення</Typography>
              <Box display="flex" flexDirection="column" gap={2} pt={1}>
                <Select fullWidth value={selectedBackup} onChange={(e) => setSelectedBackup(e.target.value)} displayEmpty>
                  <MenuItem value="" disabled>Виберіть бекап</MenuItem>
                  {backupList.map((b) => (
                    <MenuItem key={b.id || b.backupId} value={b.id || b.backupId}>
                      {(b.id || b.backupId).substring(0, 8)} - {b.backupType || 'бекап'}
                    </MenuItem>
                  ))}
                </Select>
                <Select fullWidth value={restoreType} onChange={(e) => setRestoreType(e.target.value)}>
                  <MenuItem value="full_vm">Повне відновлення ВМ</MenuItem>
                  <MenuItem value="instant">Миттєве відновлення</MenuItem>
                  <MenuItem value="file_level">Відновлення файлів</MenuItem>
                </Select>
                <TextField label="Цільовий хост" fullWidth value={targetHost} onChange={(e) => setTargetHost(e.target.value)} />
              </Box>
              <Box display="flex" justifyContent="flex-end" gap={1} sx={{ mt: 2 }}>
                <Button onClick={() => setOpen(false)}>Скасувати</Button>
                <Button variant="contained" onClick={handleStartRestore} disabled={!selectedBackup}>Розпочати</Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
