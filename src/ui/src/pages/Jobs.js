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
  FormControlLabel,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Switch,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  Backup as BackupIcon,
  CheckCircle as CheckIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon,
  Warning as WarningIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const schedulePresets = [
  { label: 'Daily at 2:00 AM', cron: '0 2 * * *' },
  { label: 'Daily at midnight', cron: '0 0 * * *' },
  { label: 'Every 6 hours', cron: '0 */6 * * *' },
  { label: 'Every 12 hours', cron: '0 */12 * * *' },
  { label: 'Weekdays at 8:00 PM', cron: '0 20 * * 1-5' },
  { label: 'Weekly on Sunday at 1:00 AM', cron: '0 1 * * 0' },
  { label: 'Monthly on day 1 at 3:00 AM', cron: '0 3 1 * *' }
];

const jobTypeColors = {
  Full: '#4fc3f7',
  Incremental: '#66bb6a',
  Differential: '#ffa726',
  Replication: '#ab47bc',
  Restore: '#ef5350'
};

const emptyForm = {
  name: '',
  jobType: 'Full',
  sourceId: '',
  sourceType: 'VirtualMachine',
  sourceHost: '',
  destinationId: '',
  schedule: '0 2 * * *',
  enabled: true,
  options: JSON.stringify({ compression: 'zstd', retention: 7 }, null, 2)
};

export default function Jobs() {
  const [jobs, setJobs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editJob, setEditJob] = useState(null);
  const [scheduleTab, setScheduleTab] = useState(1);
  const [formData, setFormData] = useState(emptyForm);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await fetchWithAuth('/api/jobs');
      const data = await response.json().catch(() => ({ jobs: [] }));
      setJobs(data.jobs || (Array.isArray(data) ? data : []));
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleRunJob = async (jobId) => {
    try {
      await fetchWithAuth(`/api/jobs/${jobId}/run`, { method: 'POST' });
    } finally {
      loadData();
    }
  };

  const handleStopJob = async (jobId) => {
    try {
      await fetchWithAuth(`/api/jobs/${jobId}/stop`, { method: 'POST' });
    } finally {
      loadData();
    }
  };

  const handleDeleteJob = async (jobId) => {
    try {
      await fetchWithAuth(`/api/jobs/${jobId}`, { method: 'DELETE' });
    } finally {
      loadData();
    }
  };

  const openCreate = () => {
    setEditJob(null);
    setScheduleTab(1);
    setFormData(emptyForm);
    setOpen(true);
  };

  const openEdit = (job) => {
    setEditJob(job);
    setScheduleTab(!job.schedule ? 0 : 1);
    setFormData({
      name: job.name || '',
      jobType: job.jobType || 'Full',
      sourceId: job.sourceId || '',
      sourceType: job.sourceType || 'VirtualMachine',
      sourceHost: job.sourceHost || '',
      destinationId: job.destinationId || '',
      schedule: job.schedule || '0 2 * * *',
      enabled: job.enabled !== false,
      options: typeof job.options === 'string'
        ? job.options
        : JSON.stringify(job.options || {}, null, 2)
    });
    setOpen(true);
  };

  const handleSaveJob = async () => {
    const payload = {
      ...formData,
      sourceType: formData.sourceType,
      schedule: scheduleTab === 0 ? null : formData.schedule
    };

    try {
      if (editJob) {
        await fetchWithAuth(`/api/jobs/${editJob.jobId}`, {
          method: 'PUT',
          body: JSON.stringify({ ...payload, jobId: editJob.jobId })
        });
      } else {
        await fetchWithAuth('/api/jobs', {
          method: 'POST',
          body: JSON.stringify(payload)
        });
      }

      setOpen(false);
      loadData();
    } catch (error) {
      console.error(error);
    }
  };

  const statusIcon = (job) => (
    job.enabled
      ? <CheckIcon sx={{ color: '#66bb6a', fontSize: 18 }} />
      : <WarningIcon sx={{ color: '#bdbdbd', fontSize: 18 }} />
  );

  const formatSchedule = (cron) => {
    if (!cron) return 'Manual only';
    const preset = schedulePresets.find((item) => item.cron === cron);
    return preset ? preset.label : cron;
  };

  const isSaveDisabled = !formData.name || !formData.sourceId || !formData.destinationId;

  if (loading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Backup Jobs</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={openCreate}
          sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}
        >
          Create Job
        </Button>
      </Box>

      {jobs.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <BackupIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No jobs configured
            </Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>
              Create your first backup job to start protecting workloads.
            </Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ bgcolor: '#4fc3f7' }}>
              Create Job
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NAME</TableCell>
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
                <TableRow key={job.jobId || job.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>{statusIcon(job)}</TableCell>
                  <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                  <TableCell>
                    <Chip
                      label={typeof job.jobType === 'string' ? job.jobType.replace('_', ' ') : 'Full'}
                      size="small"
                      sx={{
                        fontSize: '0.7rem',
                        height: 22,
                        bgcolor: `${jobTypeColors[job.jobType] || '#999'}20`,
                        color: jobTypeColors[job.jobType] || '#999',
                        fontWeight: 'bold'
                      }}
                    />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>
                    {job.sourceHost ? `${job.sourceId} (${job.sourceHost})` : job.sourceId || '-'}
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.destinationId || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{formatSchedule(job.schedule)}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Tooltip title="Run now">
                      <IconButton size="small" onClick={() => handleRunJob(job.jobId || job.id)}>
                        <PlayIcon fontSize="small" sx={{ color: '#66bb6a' }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Stop">
                      <IconButton size="small" onClick={() => handleStopJob(job.jobId || job.id)}>
                        <StopIcon fontSize="small" sx={{ color: '#ffa726' }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Edit">
                      <IconButton size="small" onClick={() => openEdit(job)}>
                        <EditIcon fontSize="small" sx={{ color: '#4fc3f7' }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" onClick={() => handleDeleteJob(job.jobId || job.id)}>
                        <DeleteIcon fontSize="small" sx={{ color: '#ef5350' }} />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editJob ? 'Edit Backup Job' : 'Create Backup Job'}</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 0.5 }}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Job Name"
                value={formData.name}
                onChange={(event) => setFormData({ ...formData, name: event.target.value })}
                size="small"
                required
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Backup Type</InputLabel>
                <Select
                  value={formData.jobType}
                  label="Backup Type"
                  onChange={(event) => setFormData({ ...formData, jobType: event.target.value })}
                >
                  <MenuItem value="Full">Full</MenuItem>
                  <MenuItem value="Incremental">Incremental</MenuItem>
                  <MenuItem value="Differential">Differential</MenuItem>
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Source Type</InputLabel>
                <Select
                  value={formData.sourceType}
                  label="Source Type"
                  onChange={(event) => setFormData({ ...formData, sourceType: event.target.value })}
                >
                  <MenuItem value="VirtualMachine">Virtual Machine</MenuItem>
                  <MenuItem value="Agent">Agent</MenuItem>
                  <MenuItem value="Folder">Folder</MenuItem>
                  <MenuItem value="Database">Database</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                size="small"
                label="Source ID or Path"
                value={formData.sourceId}
                onChange={(event) => setFormData({ ...formData, sourceId: event.target.value })}
                placeholder="vm-001 or C:\\Data"
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                size="small"
                label="Source Host"
                value={formData.sourceHost}
                onChange={(event) => setFormData({ ...formData, sourceHost: event.target.value })}
                placeholder="Optional host or hypervisor name"
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                size="small"
                label="Destination Repository"
                value={formData.destinationId}
                onChange={(event) => setFormData({ ...formData, destinationId: event.target.value })}
                placeholder="repo-001"
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.enabled}
                    onChange={(event) => setFormData({ ...formData, enabled: event.target.checked })}
                  />
                }
                label="Job enabled"
              />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>Schedule</Typography>
              <Tabs value={scheduleTab} onChange={(event, value) => setScheduleTab(value)} sx={{ mb: 2, borderBottom: 1, borderColor: 'divider' }}>
                <Tab label="Manual" />
                <Tab label="Preset" />
                <Tab label="Cron" />
              </Tabs>

              {scheduleTab === 0 ? (
                <Box p={2} bgcolor="#f8f9fa" borderRadius={1} textAlign="center">
                  <Typography variant="body2" color="text.secondary">
                    This job will run only when triggered manually or through the API.
                  </Typography>
                </Box>
              ) : scheduleTab === 1 ? (
                <FormControl fullWidth size="small">
                  <InputLabel>Schedule Preset</InputLabel>
                  <Select
                    value={formData.schedule}
                    label="Schedule Preset"
                    onChange={(event) => setFormData({ ...formData, schedule: event.target.value })}
                  >
                    {schedulePresets.map((preset) => (
                      <MenuItem key={preset.cron} value={preset.cron}>{preset.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              ) : (
                <TextField
                  fullWidth
                  label="Cron Expression"
                  value={formData.schedule || ''}
                  onChange={(event) => setFormData({ ...formData, schedule: event.target.value })}
                  size="small"
                  placeholder="0 2 * * *"
                  helperText="Format: minute hour day month weekday"
                />
              )}
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                multiline
                minRows={4}
                label="Options JSON"
                value={formData.options}
                onChange={(event) => setFormData({ ...formData, options: event.target.value })}
                size="small"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleSaveJob}
            disabled={isSaveDisabled}
            sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}
          >
            {editJob ? 'Save' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
