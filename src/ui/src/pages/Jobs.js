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
  LinearProgress,
  IconButton,
  InputAdornment,
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
  Visibility as VisibilityIcon,
  Warning as WarningIcon,
  Schedule as ScheduleIcon,
  AccessTime as TimeIcon,
  CalendarMonth as CalendarIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const schedulePresets = [
  { label: 'Daily at 2:00 AM', cron: '0 2 * * *', description: 'Every day at midnight' },
  { label: 'Daily at midnight', cron: '0 0 * * *', description: 'Every day at midnight' },
  { label: 'Every 6 hours', cron: '0 */6 * * *', description: '6 times per day' },
  { label: 'Every 12 hours', cron: '0 */12 * * *', description: '2 times per day' },
  { label: 'Weekdays at 8:00 PM', cron: '0 20 * * 1-5', description: 'Mon-Fri at 8 PM' },
  { label: 'Weekly on Sunday at 1:00 AM', cron: '0 1 * * 0', description: 'Every Sunday' },
  { label: 'Monthly on day 1 at 3:00 AM', cron: '0 3 1 * *', description: 'First day of month' }
];

const jobTypeColors = {
  Full: '#4fc3f7',
  Incremental: '#66bb6a',
  Differential: '#ffa726',
  Replication: '#ab47bc',
  Restore: '#ef5350'
};

const sourceTypes = [
  { value: 'VirtualMachine', label: 'Virtual Machine', icon: '💻' },
  { value: 'Agent', label: 'Agent', icon: '🤖' },
  { value: 'Folder', label: 'Folder', icon: '📁' },
  { value: 'Database', label: 'Database', icon: '🗄️' }
];

const jobTypes = [
  { value: 'Full', label: 'Full Backup', description: 'Complete backup of all data' },
  { value: 'Incremental', label: 'Incremental', description: 'Backup only changed data' },
  { value: 'Differential', label: 'Differential', description: 'Backup changes since last full' }
];

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
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [selectedJob, setSelectedJob] = useState(null);
  const [jobRuns, setJobRuns] = useState([]);
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

  const openHistory = async (job) => {
    setSelectedJob(job);
    setHistoryOpen(true);
    setHistoryLoading(true);
    setJobRuns([]);

    try {
      const response = await fetchWithAuth(`/api/jobs/${job.jobId}/runs`);
      const data = await response.json().catch(() => ({ runs: [] }));
      setJobRuns(data.runs || []);
    } catch (error) {
      console.error(error);
      setJobRuns([]);
    } finally {
      setHistoryLoading(false);
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

  const runStatusChip = (status) => (
    <Chip
      label={status}
      size="small"
      sx={{
        fontSize: '0.7rem',
        height: 22,
        bgcolor:
          String(status).toLowerCase() === 'completed' ? '#66bb6a20' :
          String(status).toLowerCase() === 'running' ? '#4fc3f720' :
          String(status).toLowerCase() === 'queued' ? '#ffa72620' :
          String(status).toLowerCase() === 'cancelled' ? '#bdbdbd40' :
          String(status).toLowerCase() === 'failed' ? '#ef535020' :
          '#bdbdbd20',
        color:
          String(status).toLowerCase() === 'completed' ? '#66bb6a' :
          String(status).toLowerCase() === 'running' ? '#4fc3f7' :
          String(status).toLowerCase() === 'queued' ? '#ffa726' :
          String(status).toLowerCase() === 'cancelled' ? '#616161' :
          String(status).toLowerCase() === 'failed' ? '#ef5350' :
          '#757575',
        fontWeight: 'bold'
      }}
    />
  );

  const isSaveDisabled = !formData.name || !formData.sourceId || !formData.destinationId;

  if (loading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography variant="h5" sx={{ fontWeight: 600, color: '#1a1d23' }}>Backup Jobs</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            Manage your backup jobs and schedules
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={openCreate}
          sx={{
            bgcolor: '#4fc3f7',
            '&:hover': { bgcolor: '#29b6f6' },
            px: 3,
            py: 1,
            borderRadius: 2,
            fontWeight: 600,
            boxShadow: '0 4px 14px rgba(79, 195, 247, 0.4)'
          }}
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
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>RUN STATUS</TableCell>
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
                  <TableCell>
                    {job.latestRun?.status ? (
                      runStatusChip(job.latestRun.status)
                    ) : '-'}
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Tooltip title="Run history">
                      <IconButton size="small" onClick={() => openHistory(job)}>
                        <VisibilityIcon fontSize="small" sx={{ color: '#8b92a5' }} />
                      </IconButton>
                    </Tooltip>
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

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth PaperProps={{ sx: { borderRadius: 3 } }}>
        <DialogTitle sx={{ pb: 1 }}>
          <Box display="flex" alignItems="center" gap={2}>
            <Box sx={{ 
              width: 40, 
              height: 40, 
              borderRadius: 2, 
              bgcolor: '#4fc3f720',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}>
              <BackupIcon sx={{ color: '#4fc3f7' }} />
            </Box>
            <Box>
              <Typography variant="h6" sx={{ fontWeight: 600 }}>
                {editJob ? 'Edit Backup Job' : 'Create Backup Job'}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {editJob ? 'Update job configuration' : 'Configure a new backup job'}
              </Typography>
            </Box>
          </Box>
        </DialogTitle>
        <DialogContent sx={{ pt: 2 }}>
          <Grid container spacing={3}>
            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1a1d23', mb: 1 }}>
                Job Details
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Job Name"
                value={formData.name}
                onChange={(event) => setFormData({ ...formData, name: event.target.value })}
                size="small"
                required
                placeholder="e.g., Daily VM Backup"
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <BackupIcon sx={{ color: '#8b92a5', fontSize: 20 }} />
                    </InputAdornment>
                  )
                }}
                sx={{
                  '& .MuiOutlinedInput-root': { borderRadius: 2 }
                }}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Backup Type</InputLabel>
                <Select
                  value={formData.jobType}
                  label="Backup Type"
                  onChange={(event) => setFormData({ ...formData, jobType: event.target.value })}
                  sx={{ borderRadius: 2 }}
                >
                  {jobTypes.map((type) => (
                    <MenuItem key={type.value} value={type.value}>
                      <Box>
                        <Typography variant="body2">{type.label}</Typography>
                        <Typography variant="caption" color="text.secondary">{type.description}</Typography>
                      </Box>
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1a1d23', mb: 1, mt: 1 }}>
                Source Configuration
              </Typography>
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControl fullWidth size="small">
                <InputLabel>Source Type</InputLabel>
                <Select
                  value={formData.sourceType}
                  label="Source Type"
                  onChange={(event) => setFormData({ ...formData, sourceType: event.target.value })}
                  sx={{ borderRadius: 2 }}
                >
                  {sourceTypes.map((type) => (
                    <MenuItem key={type.value} value={type.value}>
                      <Box display="flex" alignItems="center" gap={1}>
                        <span>{type.icon}</span> {type.label}
                      </Box>
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={4}>
              <TextField
                fullWidth
                size="small"
                label="Source ID or Name"
                value={formData.sourceId}
                onChange={(event) => setFormData({ ...formData, sourceId: event.target.value })}
                placeholder="e.g., vm-001"
                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
              />
            </Grid>
            <Grid item xs={12} sm={4}>
              <TextField
                fullWidth
                size="small"
                label="Source Host (Optional)"
                value={formData.sourceHost}
                onChange={(event) => setFormData({ ...formData, sourceHost: event.target.value })}
                placeholder="e.g., hypervisor-01"
                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
              />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1a1d23', mb: 1, mt: 1 }}>
                Destination
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                size="small"
                label="Destination Repository"
                value={formData.destinationId}
                onChange={(event) => setFormData({ ...formData, destinationId: event.target.value })}
                placeholder="e.g., repo-local-01"
                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.enabled}
                    onChange={(event) => setFormData({ ...formData, enabled: event.target.checked })}
                    sx={{
                      '& .MuiSwitch-switchBase.Mui-checked': {
                        color: '#4fc3f7',
                      },
                      '& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track': {
                        backgroundColor: '#4fc3f7',
                      },
                    }}
                  />
                }
                label={
                  <Box>
                    <Typography variant="body2">Job Enabled</Typography>
                    <Typography variant="caption" color="text.secondary">
                      Job will run according to schedule
                    </Typography>
                  </Box>
                }
              />
            </Grid>

            <Grid item xs={12}>
              <Box sx={{ 
                bgcolor: '#f8f9fa', 
                borderRadius: 2, 
                p: 2,
                border: '1px solid #e8eaed'
              }}>
                <Box display="flex" alignItems="center" gap={1} mb={2}>
                  <ScheduleIcon sx={{ color: '#4fc3f7' }} />
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1a1d23' }}>
                    Schedule Configuration
                  </Typography>
                </Box>
                <Tabs 
                  value={scheduleTab} 
                  onChange={(event, value) => setScheduleTab(value)} 
                  sx={{ 
                    mb: 2, 
                    borderBottom: 1, 
                    borderColor: 'divider',
                    '& .MuiTab-root': { textTransform: 'none' }
                  }}
                >
                  <Tab icon={<TimeIcon sx={{ fontSize: 18 }} />} iconPosition="start" label="Manual" />
                  <Tab icon={<CalendarIcon sx={{ fontSize: 18 }} />} iconPosition="start" label="Schedule" />
                  <Tab label="Custom Cron" />
                </Tabs>

                {scheduleTab === 0 ? (
                  <Box p={2} bgcolor="#fff" borderRadius={1} textAlign="center" border="1px dashed #e0e0e0">
                    <Typography variant="body2" color="text.secondary">
                      This job will run only when triggered manually or through the API.
                    </Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
                      Click "Run Now" button to execute this job.
                    </Typography>
                  </Box>
                ) : scheduleTab === 1 ? (
                  <Grid container spacing={2}>
                    {schedulePresets.map((preset) => (
                      <Grid item xs={12} sm={6} key={preset.cron}>
                        <Box
                          onClick={() => setFormData({ ...formData, schedule: preset.cron })}
                          sx={{
                            p: 2,
                            borderRadius: 2,
                            border: formData.schedule === preset.cron ? '2px solid #4fc3f7' : '1px solid #e8eaed',
                            bgcolor: formData.schedule === preset.cron ? '#4fc3f710' : '#fff',
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                            '&:hover': {
                              borderColor: '#4fc3f7',
                              bgcolor: '#4fc3f708'
                            }
                          }}
                        >
                          <Typography variant="body2" sx={{ fontWeight: 600, color: '#1a1d23' }}>
                            {preset.label}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {preset.description}
                          </Typography>
                          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.5, fontFamily: 'monospace' }}>
                            {preset.cron}
                          </Typography>
                        </Box>
                      </Grid>
                    ))}
                  </Grid>
                ) : (
                  <TextField
                    fullWidth
                    label="Cron Expression"
                    value={formData.schedule || ''}
                    onChange={(event) => setFormData({ ...formData, schedule: event.target.value })}
                    size="small"
                    placeholder="0 2 * * *"
                    helperText="Format: minute hour day month weekday. Example: 0 2 * * * = Daily at 2:00 AM"
                    sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                  />
                )}
              </Box>
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1a1d23', mb: 1 }}>
                Advanced Options (JSON)
              </Typography>
              <TextField
                fullWidth
                multiline
                minRows={3}
                maxRows={6}
                label="Options"
                value={formData.options}
                onChange={(event) => setFormData({ ...formData, options: event.target.value })}
                size="small"
                placeholder='{"compression": "zstd", "retention": 7}'
                sx={{ 
                  '& .MuiOutlinedInput-root': { borderRadius: 2 },
                  '& .MuiInputBase-input': { fontFamily: 'monospace', fontSize: '0.85rem' }
                }}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button 
            onClick={() => setOpen(false)}
            sx={{ color: '#666' }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            onClick={handleSaveJob}
            disabled={isSaveDisabled}
            sx={{
              bgcolor: '#4fc3f7',
              '&:hover': { bgcolor: '#29b6f6' },
              px: 4,
              borderRadius: 2,
              fontWeight: 600
            }}
          >
            {editJob ? 'Save Changes' : 'Create Job'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={historyOpen} onClose={() => setHistoryOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{selectedJob ? `${selectedJob.name} Run History` : 'Run History'}</DialogTitle>
        <DialogContent>
          {historyLoading ? (
            <Box display="flex" justifyContent="center" p={4}>
              <CircularProgress />
            </Box>
          ) : jobRuns.length === 0 ? (
            <Box py={4} textAlign="center">
              <Typography variant="body2" color="text.secondary">
                No run history found for this job yet.
              </Typography>
            </Box>
          ) : (
            <TableContainer component={Paper} sx={{ borderRadius: 2, mt: 1 }}>
              <Table size="small">
                <TableHead>
                  <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>RUN ID</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>STARTED</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>ENDED</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>BYTES</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>FILES</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>SPEED</TableCell>
                    <TableCell sx={{ fontWeight: 600, fontSize: '0.75rem' }}>DETAILS</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {jobRuns.map((run) => (
                    <TableRow key={run.runId} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                      <TableCell sx={{ fontSize: '0.8rem', fontFamily: 'monospace' }}>{run.runId}</TableCell>
                      <TableCell>{runStatusChip(run.status)}</TableCell>
                      <TableCell sx={{ fontSize: '0.8rem' }}>
                        {run.startTime ? new Date(run.startTime).toLocaleString() : '-'}
                      </TableCell>
                      <TableCell sx={{ fontSize: '0.8rem' }}>
                        {run.endTime ? new Date(run.endTime).toLocaleString() : '-'}
                      </TableCell>
                      <TableCell sx={{ fontSize: '0.8rem' }}>
                        {typeof run.bytesProcessed === 'number' ? run.bytesProcessed.toLocaleString() : '-'}
                      </TableCell>
                      <TableCell sx={{ fontSize: '0.8rem' }}>
                        {typeof run.filesProcessed === 'number' ? run.filesProcessed : '-'}
                      </TableCell>
                      <TableCell sx={{ fontSize: '0.8rem' }}>
                        {typeof run.speedMbps === 'number' && run.speedMbps > 0
                          ? `${run.speedMbps.toFixed(2)} Mbps`
                          : '-'}
                      </TableCell>
                      <TableCell sx={{ minWidth: 220 }}>
                        {run.status === 'running' ? (
                          <Box display="flex" alignItems="center" gap={1}>
                            <LinearProgress sx={{ flex: 1, minWidth: 80 }} />
                            <Typography variant="caption" color="text.secondary">Running</Typography>
                          </Box>
                        ) : run.errorMessage ? (
                          <Typography variant="caption" color="error.main">{run.errorMessage}</Typography>
                        ) : (
                          <Typography variant="caption" color="text.secondary">
                            {run.status === 'completed' ? 'Completed successfully' : 'No error details'}
                          </Typography>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setHistoryOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
