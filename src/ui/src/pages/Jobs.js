import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Button, TextField, Select, MenuItem, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, IconButton, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { Add as AddIcon, PlayArrow as RunIcon, Stop as StopIcon, Delete as DeleteIcon, Edit as EditIcon } from '@mui/icons-material';
import { useApi, useApiMutation } from '../services/ApiContext';

export default function Jobs() {
  const { data, loading, refetch } = useApi('/api/jobs');
  const [open, setOpen] = useState(false);
  const [selectedJob, setSelectedJob] = useState(null);

  const handleRunJob = async (jobId) => {
    await fetch(`/api/jobs/${jobId}/run`, { method: 'POST' });
    refetch();
  };

  const handleStopJob = async (jobId: string) => {
    await fetch(`/api/jobs/${jobId}/stop`, { method: 'POST' });
    refetch();
  };

  const handleDeleteJob = async (jobId: string) => {
    await fetch(`/api/jobs/${jobId}`, { method: 'DELETE' });
    refetch();
  };

  if (loading) return <Typography>Loading...</Typography>;

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Jobs</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
          New Job
        </Button>
      </Box>

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
            {data?.jobs?.map((job: any) => (
              <TableRow key={job.jobId}>
                <TableCell>{job.name}</TableCell>
                <TableCell><Chip label={job.jobType} size="small" /></TableCell>
                <TableCell>{job.sourceId}</TableCell>
                <TableCell>{job.destinationId}</TableCell>
                <TableCell>
                  <Chip 
                    label={job.enabled ? 'Active' : 'Disabled'} 
                    color={job.enabled ? 'success' : 'default'} 
                    size="small" 
                  />
                </TableCell>
                <TableCell>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                <TableCell>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                <TableCell>
                  <IconButton onClick={() => handleRunJob(job.jobId)}><RunIcon /></IconButton>
                  <IconButton onClick={() => handleStopJob(job.jobId)}><StopIcon /></IconButton>
                  <IconButton onClick={() => { setSelectedJob(job); setOpen(true); }}><EditIcon /></IconButton>
                  <IconButton onClick={() => handleDeleteJob(job.jobId)} color="error"><DeleteIcon /></IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{selectedJob ? 'Edit Job' : 'Create Job'}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} pt={2}>
            <TextField label="Job Name" fullWidth defaultValue={selectedJob?.name} />
            <Select label="Job Type" fullWidth defaultValue={selectedJob?.jobType || 'full_backup'}>
              <MenuItem value="full_backup">Full Backup</MenuItem>
              <MenuItem value="incremental">Incremental</MenuItem>
              <MenuItem value="differential">Differential</MenuItem>
            </Select>
            <TextField label="Source" fullWidth defaultValue={selectedJob?.sourceId} />
            <TextField label="Destination" fullWidth defaultValue={selectedJob?.destinationId} />
            <TextField label="Schedule (Cron)" fullWidth placeholder="0 2 * * *" />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={() => setOpen(false)}>Save</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
