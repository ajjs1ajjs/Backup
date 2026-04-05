import React, { useState, useEffect } from 'react';
import {
  Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, Chip, Button, TextField, Select, MenuItem, CircularProgress,
  Paper, Dialog, DialogTitle, DialogContent, DialogActions, IconButton, Tooltip,
  FormControl, InputLabel, Tabs, Tab, FormControlLabel, Switch, Slider, Grid
} from '@mui/material';
import {
  PlayArrow as PlayIcon, Stop as StopIcon, Add as AddIcon, Delete as DeleteIcon,
  CheckCircle as CheckIcon, Warning as WarningIcon, Error as ErrorIcon,
  Backup as BackupIcon, Edit as EditIcon, Schedule as ScheduleIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const schedulePresets = [
  { label: 'Daily at 2:00 AM', cron: '0 2 * * *' },
  { label: 'Daily at 12:00 AM (Midnight)', cron: '0 0 * * *' },
  { label: 'Every 6 hours', cron: '0 */6 * * *' },
  { label: 'Every 12 hours', cron: '0 */12 * * *' },
  { label: 'Weekdays at 8:00 PM', cron: '0 20 * * 1-5' },
  { label: 'Weekly on Sunday at 1:00 AM', cron: '0 1 * * 0' },
  { label: 'Monthly on 1st at 3:00 AM', cron: '0 3 1 * *' },
];

const jobTypeColors = {
  full_backup: '#4fc3f7',
  incremental: '#66bb6a',
  differential: '#ffa726',
  verify: '#ab47bc',
  restore: '#ef5350'
};

export default function Jobs() {
  const [jobs, setJobs] = useState([]);
  const [vms, setVMs] = useState([]);
  const [repos, setRepos] = useState([]);
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editJob, setEditJob] = useState(null);
  const [scheduleTab, setScheduleTab] = useState(0);
  const [formData, setFormData] = useState({
    name: '', jobType: 'full_backup', sourceId: '', sourceType: 'vm',
    destinationId: '', schedule: '0 2 * * *', enabled: true,
    options: JSON.stringify({ compression: 'zstd', retention: 7 })
  });

  useEffect(() => { loadData(); }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [jobsRes, vmsRes, reposRes, agentsRes] = await Promise.all([
        fetchWithAuth('/api/jobs'),
        fetchWithAuth('/api/virtualmachines'),
        fetchWithAuth('/api/repositories'),
        fetchWithAuth('/api/agents'),
      ]);
      const [jobsData, vmsData, reposData, agentsData] = await Promise.all([
        jobsRes.json().catch(() => ({ jobs: [] })),
        vmsRes.json().catch(() => []),
        reposRes.json().catch(() => []),
        agentsRes.json().catch(() => []),
      ]);
      setJobs(jobsData.jobs || (Array.isArray(jobsData) ? jobsData : []));
      setVMs(Array.isArray(vmsData) ? vmsData : []);
      setRepos(Array.isArray(reposData) ? reposData : []);
      setAgents(Array.isArray(agentsData) ? agentsData : []);
    } catch (e) { console.error(e); }
    finally { setLoading(false); }
  };

  const handleRunJob = async (jobId) => {
    try { await fetchWithAuth(`/api/jobs/${jobId}/run`, { method: 'POST' }); loadData(); } catch (e) { loadData(); }
  };

  const handleStopJob = async (jobId) => {
    try { await fetchWithAuth(`/api/jobs/${jobId}/stop`, { method: 'POST' }); loadData(); } catch (e) { loadData(); }
  };

  const handleDeleteJob = async (jobId) => {
    try { await fetchWithAuth(`/api/jobs/${jobId}`, { method: 'DELETE' }); loadData(); } catch (e) { loadData(); }
  };

  const openCreate = () => {
    setEditJob(null);
    setScheduleTab(0);
    setFormData({
      name: '', jobType: 'full_backup', sourceId: '', sourceType: 'vm',
      destinationId: '', schedule: '0 2 * * *', enabled: true,
      options: JSON.stringify({ compression: 'zstd', retention: 7 })
    });
    setOpen(true);
  };

  const openEdit = (job) => {
    setEditJob(job);
    setScheduleTab(0);
    setFormData({
      name: job.name || '',
      jobType: job.jobType || 'full_backup',
      sourceId: job.sourceId || '',
      sourceType: job.sourceType || 'vm',
      destinationId: job.destinationId || '',
      schedule: job.schedule || '0 2 * * *',
      enabled: job.enabled !== false,
      options: typeof job.options === 'string' ? job.options : JSON.stringify(job.options || {})
    });
    setOpen(true);
  };

  const handleSaveJob = async () => {
    try {
      if (editJob) {
        await fetchWithAuth(`/api/jobs/${editJob.jobId}`, {
          method: 'PUT',
          body: JSON.stringify({ ...formData, jobId: editJob.jobId })
        });
      } else {
        await fetchWithAuth('/api/jobs', { method: 'POST', body: JSON.stringify(formData) });
      }
      setOpen(false);
      loadData();
    } catch (e) { console.error(e); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;

  const statusIcon = (job) => {
    if (job.enabled) return <CheckIcon sx={{ color: '#66bb6a', fontSize: 18 }} />;
    return <WarningIcon sx={{ color: '#bdbdbd', fontSize: 18 }} />;
  };

  const formatSchedule = (cron) => {
    if (!cron) return 'Manual only';
    const preset = schedulePresets.find(p => p.cron === cron);
    return preset ? preset.label : cron;
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Backup Jobs</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          New Job
        </Button>
      </Box>

      {jobs.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <BackupIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No backup jobs configured</Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>Create your first backup job to start protecting your data</Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ bgcolor: '#4fc3f7' }}>Create Job</Button>
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
                <TableRow key={job.jobId || job.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>{statusIcon(job)}</TableCell>
                  <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                  <TableCell>
                    <Chip label={job.jobType?.replace('_', ' ') || 'full'} size="small"
                      sx={{ fontSize: '0.7rem', height: 22, bgcolor: (jobTypeColors[job.jobType] || '#999') + '20', color: jobTypeColors[job.jobType] || '#999', fontWeight: 'bold' }} />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.sourceId || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.destinationId || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{formatSchedule(job.schedule)}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Tooltip title="Run now">
                      <IconButton size="small" onClick={() => handleRunJob(job.jobId || job.id)}><PlayIcon fontSize="small" sx={{ color: '#66bb6a' }} /></IconButton>
                    </Tooltip>
                    <Tooltip title="Stop">
                      <IconButton size="small" onClick={() => handleStopJob(job.jobId || job.id)}><StopIcon fontSize="small" sx={{ color: '#ffa726' }} /></IconButton>
                    </Tooltip>
                    <Tooltip title="Edit">
                      <IconButton size="small" onClick={() => openEdit(job)}><EditIcon fontSize="small" sx={{ color: '#4fc3f7' }} /></IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" onClick={() => handleDeleteJob(job.jobId || job.id)}><DeleteIcon fontSize="small" sx={{ color: '#ef5350' }} /></IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editJob ? 'Edit' : 'Create'} Backup Job</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 0.5 }}>
            <Grid item xs={12} sm={6}>
              <TextField fullWidth label="Job Name" value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })} size="small" required />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Backup Type</InputLabel>
                <Select value={formData.jobType} label="Backup Type"
                  onChange={(e) => setFormData({ ...formData, jobType: e.target.value })}>
                  <MenuItem value="full_backup">Full Backup</MenuItem>
                  <MenuItem value="incremental">Incremental</MenuItem>
                  <MenuItem value="differential">Differential</MenuItem>
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Source Type</InputLabel>
                <Select value={formData.sourceType} label="Source Type"
                  onChange={(e) => setFormData({ ...formData, sourceType: e.target.value, sourceId: '' })}>
                  <MenuItem value="vm">Virtual Machine</MenuItem>
                  <MenuItem value="agent">Agent / Physical Server</MenuItem>
                  <MenuItem value="folder">Folder / Files</MenuItem>
                  <MenuItem value="database">Database</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Source</InputLabel>
                <Select value={formData.sourceId} label="Source"
                  onChange={(e) => setFormData({ ...formData, sourceId: e.target.value })}>
                  {formData.sourceType === 'vm' && vms.length > 0 && vms.map(vm => (
                    <MenuItem key={vm.vmId} value={vm.vmId}>{vm.name} ({vm.ipAddress || 'no IP'})</MenuItem>
                  ))}
                  {formData.sourceType === 'agent' && agents.length > 0 && agents.map(a => (
                    <MenuItem key={a.agentId} value={a.agentId}>{a.hostname} ({a.ipAddress || 'no IP'})</MenuItem>
                  ))}
                  {formData.sourceType === 'vm' && vms.length === 0 && (
                    <MenuItem value="__manual__" disabled>No VMs registered — add manually below</MenuItem>
                  )}
                  {formData.sourceType === 'agent' && agents.length === 0 && (
                    <MenuItem value="__manual__" disabled>No agents registered — add manually below</MenuItem>
                  )}
                  {formData.sourceType === 'database' && (
                    <MenuItem value="__manual__" disabled>No databases registered — add manually below</MenuItem>
                  )}
                </Select>
              </FormControl>
              {(formData.sourceType === 'folder' ||
                (formData.sourceType === 'vm' && vms.length === 0) ||
                (formData.sourceType === 'agent' && agents.length === 0) ||
                formData.sourceType === 'database') && (
                <TextField fullWidth size="small" sx={{ mt: 1 }}
                  label={formData.sourceType === 'folder' ? 'Folder Path' : 'Source ID / Name'}
                  value={formData.sourceId === '__manual__' ? '' : formData.sourceId}
                  onChange={(e) => setFormData({ ...formData, sourceId: e.target.value })}
                  placeholder={formData.sourceType === 'vm' ? 'VM name or ID' : formData.sourceType === 'folder' ? 'C:\\Data or /home/user' : 'Source identifier'}
                  helperText={formData.sourceType === 'vm' ? 'Or register VMs via Hypervisors page' : formData.sourceType === 'folder' ? 'Path to files/folder to backup' : 'Enter source identifier'}
                />
              )}
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Destination Repository</InputLabel>
                <Select value={formData.destinationId} label="Destination Repository"
                  onChange={(e) => setFormData({ ...formData, destinationId: e.target.value })}>
                  {repos.length > 0 && repos.map(r => (
                    <MenuItem key={r.repositoryId} value={r.repositoryId}>{r.name} ({r.type})</MenuItem>
                  ))}
                  {repos.length === 0 && (
                    <MenuItem value="__manual__" disabled>No repositories — add manually below</MenuItem>
                  )}
                </Select>
              </FormControl>
              {repos.length === 0 && (
                <TextField fullWidth size="small" sx={{ mt: 1 }}
                  label="Destination Path / Repository ID"
                  value={formData.destinationId === '__manual__' ? '' : formData.destinationId}
                  onChange={(e) => setFormData({ ...formData, destinationId: e.target.value })}
                  placeholder="e.g. D:\Backups or /mnt/backups or repo-id"
                  helperText="Or add a repository via Repositories page"
                />
              )}

            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Destination Repository</InputLabel>
                <Select value={formData.destinationId} label="Destination Repository"
                  onChange={(e) => setFormData({ ...formData, destinationId: e.target.value })}>
                  {repos.map(r => (
                    <MenuItem key={r.repositoryId} value={r.repositoryId}>{r.name} ({r.type})</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControlLabel
                control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
                label="Job Enabled"
              />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>Schedule</Typography>
              <Tabs value={scheduleTab} onChange={(e, v) => setScheduleTab(v)} sx={{ mb: 2 }}>
                <Tab label="Preset" icon={<ScheduleIcon fontSize="small" />} iconPosition="start" />
                <Tab label="Custom Cron" />
              </Tabs>

              {scheduleTab === 0 ? (
                <Box display="flex" flexDirection="column" gap={1}>
                  {schedulePresets.map((preset) => (
                    <Button key={preset.cron} variant={formData.schedule === preset.cron ? 'contained' : 'outlined'}
                      size="small" onClick={() => setFormData({ ...formData, schedule: preset.cron })}
                      sx={{ justifyContent: 'flex-start', textTransform: 'none' }}>
                      {preset.label}
                    </Button>
                  ))}
                </Box>
              ) : (
                <TextField fullWidth label="Cron Expression" value={formData.schedule}
                  onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                  size="small" placeholder="0 2 * * *" helperText="min hour day month weekday" />
              )}
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSaveJob} disabled={!formData.name || !formData.sourceId || formData.sourceId === '__manual__' || !formData.destinationId || formData.destinationId === '__manual__'}
            sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
            {editJob ? 'Save' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
