import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, TextField, Select, MenuItem, CircularProgress, Paper } from '@mui/material';
import { PlayArrow as PlayIcon, Stop as StopIcon, Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon, CheckCircle as CheckIcon, Warning as WarningIcon, Error as ErrorIcon } from '@mui/icons-material';
import { useApi, useApiMutation } from '../services/ApiContext';

export default function Jobs() {
  const { data, loading, refetch } = useApi('/api/jobs');
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({ name: '', jobType: 'full_backup', sourceId: '', destinationId: '', schedule: '0 2 * * *' });

  const handleRunJob = async (jobId) => {
    try { await fetch(`/api/jobs/${jobId}/run`, { method: 'POST' }); refetch(); } catch (e) { refetch(); }
  };

  const handleStopJob = async (jobId) => {
    try { await fetch(`/api/jobs/${jobId}/stop`, { method: 'POST' }); refetch(); } catch (e) { refetch(); }
  };

  const handleDeleteJob = async (jobId) => {
    try { await fetch(`/api/jobs/${jobId}`, { method: 'DELETE' }); refetch(); } catch (e) { refetch(); }
  };

  const handleCreateJob = async () => {
    try {
      await fetch('/api/jobs', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(formData) });
      setOpen(false);
      setFormData({ name: '', jobType: 'full_backup', sourceId: '', destinationId: '', schedule: '0 2 * * *' });
      refetch();
    } catch (e) { setOpen(false); refetch(); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;

  const jobs = data?.jobs || [];

  const statusIcon = (job) => {
    if (job.status === 'running' || job.status === 'active') return <WarningIcon sx={{ color: '#ffa726', fontSize: 18 }} />;
    if (job.status === 'failed' || job.status === 'error') return <ErrorIcon sx={{ color: '#ef5350', fontSize: 18 }} />;
    if (job.enabled) return <CheckIcon sx={{ color: '#66bb6a', fontSize: 18 }} />;
    return <WarningIcon sx={{ color: '#bdbdbd', fontSize: 18 }} />;
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Backup Jobs</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          New Job
        </Button>
      </Box>

      {jobs.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <BackupIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No backup jobs configured</Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>Create your first backup job to start protecting your data</Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)} sx={{ bgcolor: '#4fc3f7' }}>Create Job</Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>JOB NAME</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>SOURCE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>DESTINATION</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>SCHEDULE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LAST RUN</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NEXT RUN</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ACTIONS</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {jobs.map((job) => (
                <TableRow key={job.id || job.jobId} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>{statusIcon(job)}</TableCell>
                  <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                  <TableCell><Chip label={job.jobType || 'full_backup'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} /></TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.sourceId || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.destinationId || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem', fontFamily: 'monospace' }}>{job.schedule || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Button size="small" onClick={() => handleRunJob(job.id || job.jobId)} sx={{ minWidth: 32, p: 0.5 }}><PlayIcon fontSize="small" sx={{ color: '#66bb6a' }} /></Button>
                    <Button size="small" onClick={() => handleStopJob(job.id || job.jobId)} sx={{ minWidth: 32, p: 0.5 }}><StopIcon fontSize="small" sx={{ color: '#ffa726' }} /></Button>
                    <Button size="small" onClick={() => handleDeleteJob(job.id || job.jobId)} sx={{ minWidth: 32, p: 0.5 }}><DeleteIcon fontSize="small" sx={{ color: '#ef5350' }} /></Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {open && (
        <Box sx={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, bgcolor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1300 }} onClick={() => setOpen(false)}>
          <Card sx={{ width: 550, maxWidth: '90%', borderRadius: 2 }} onClick={(e) => e.stopPropagation()}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Create Backup Job</Typography>
              <Box display="flex" flexDirection="column" gap={2}>
                <TextField label="Job Name" fullWidth value={formData.name} onChange={(e) => setFormData({...formData, name: e.target.value})} size="small" />
                <Select fullWidth size="small" value={formData.jobType} onChange={(e) => setFormData({...formData, jobType: e.target.value})}>
                  <MenuItem value="full_backup">Full Backup</MenuItem>
                  <MenuItem value="incremental">Incremental Backup</MenuItem>
                  <MenuItem value="differential">Differential Backup</MenuItem>
                </Select>
                <TextField label="Source" fullWidth value={formData.sourceId} onChange={(e) => setFormData({...formData, sourceId: e.target.value})} size="small" helperText="VM name, agent ID, or source path" />
                <TextField label="Destination" fullWidth value={formData.destinationId} onChange={(e) => setFormData({...formData, destinationId: e.target.value})} size="small" helperText="Repository ID or target path" />
                <TextField label="Schedule (Cron)" fullWidth value={formData.schedule} onChange={(e) => setFormData({...formData, schedule: e.target.value})} size="small" placeholder="0 2 * * *" helperText="Default: daily at 2:00 AM" />
              </Box>
              <Box display="flex" justifyContent="flex-end" gap={1} sx={{ mt: 3 }}>
                <Button onClick={() => setOpen(false)}>Cancel</Button>
                <Button variant="contained" onClick={handleCreateJob} disabled={!formData.name} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>Create</Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
