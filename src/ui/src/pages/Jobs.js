import React, { useState } from 'react';
import { Box, Card, CardContent, Grid, Typography, Chip, CircularProgress, LinearProgress, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Button, TextField, Select, MenuItem } from '@mui/material';
import { PlayArrow as PlayIcon, Stop as StopIcon, Add as AddIcon } from '@mui/icons-material';
import { useApi, useApiMutation } from '../services/ApiContext';

export default function Jobs() {
  const { data, loading, refetch } = useApi('/api/jobs');
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({ name: '', jobType: 'full_backup', sourceId: '', destinationId: '', schedule: '0 2 * * *' });

  const runJob = useApiMutation(`/api/jobs/run`, 'POST');
  const stopJob = useApiMutation(`/api/jobs/stop`, 'POST');
  const deleteJob = useApiMutation(`/api/jobs/delete`, 'DELETE');

  const handleRunJob = async (jobId) => {
    try {
      await runJob.mutate({ jobId });
      refetch();
    } catch (e) { refetch(); }
  };

  const handleStopJob = async (jobId) => {
    try {
      await stopJob.mutate({ jobId });
      refetch();
    } catch (e) { refetch(); }
  };

  const handleDeleteJob = async (jobId) => {
    try {
      await deleteJob.mutate({ jobId });
      refetch();
    } catch (e) { refetch(); }
  };

  const handleCreateJob = async () => {
    const createJob = useApiMutation('/api/jobs', 'POST');
    try {
      await createJob.mutate(formData);
      setOpen(false);
      setFormData({ name: '', jobType: 'full_backup', sourceId: '', destinationId: '', schedule: '0 2 * * *' });
      refetch();
    } catch (e) { setOpen(false); refetch(); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  const jobs = data?.jobs || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Jobs</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
          New Job
        </Button>
      </Box>

      {jobs.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>No jobs configured</Typography>
            <Typography variant="body2" color="text.secondary" mb={2}>Create your first backup job to get started</Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>Create Job</Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Source</TableCell>
                <TableCell>Destination</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Last Run</TableCell>
                <TableCell>Next Run</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {jobs.map((job) => (
                <TableRow key={job.id || job.jobId}>
                  <TableCell>{job.name}</TableCell>
                  <TableCell><Chip label={job.jobType || 'full_backup'} size="small" /></TableCell>
                  <TableCell>{job.sourceId || '-'}</TableCell>
                  <TableCell>{job.destinationId || '-'}</TableCell>
                  <TableCell>
                    <Chip label={job.enabled ? 'Active' : 'Disabled'} color={job.enabled ? 'success' : 'default'} size="small" />
                  </TableCell>
                  <TableCell>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Button size="small" onClick={() => handleRunJob(job.id || job.jobId)}><PlayIcon fontSize="small" /></Button>
                    <Button size="small" onClick={() => handleStopJob(job.id || job.jobId)}><StopIcon fontSize="small" /></Button>
                    <Button size="small" color="error" onClick={() => handleDeleteJob(job.id || job.jobId)}>Delete</Button>
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
              <Typography variant="h6" gutterBottom>Create Job</Typography>
              <Box display="flex" flexDirection="column" gap={2} pt={1}>
                <TextField label="Job Name" fullWidth value={formData.name} onChange={(e) => setFormData({...formData, name: e.target.value})} />
                <Select fullWidth value={formData.jobType} onChange={(e) => setFormData({...formData, jobType: e.target.value})}>
                  <MenuItem value="full_backup">Full Backup</MenuItem>
                  <MenuItem value="incremental">Incremental</MenuItem>
                  <MenuItem value="differential">Differential</MenuItem>
                </Select>
                <TextField label="Source" fullWidth value={formData.sourceId} onChange={(e) => setFormData({...formData, sourceId: e.target.value})} />
                <TextField label="Destination" fullWidth value={formData.destinationId} onChange={(e) => setFormData({...formData, destinationId: e.target.value})} />
                <TextField label="Schedule (Cron)" fullWidth value={formData.schedule} onChange={(e) => setFormData({...formData, schedule: e.target.value})} placeholder="0 2 * * *" />
              </Box>
              <Box display="flex" justifyContent="flex-end" gap={1} sx={{ mt: 2 }}>
                <Button onClick={() => setOpen(false)}>Cancel</Button>
                <Button variant="contained" onClick={handleCreateJob} disabled={!formData.name}>Save</Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
