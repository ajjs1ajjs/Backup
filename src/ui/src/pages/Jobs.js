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
  Full: '#4fc3f7',
  Incremental: '#66bb6a',
  Differential: '#ffa726',
  Replication: '#ab47bc',
  Restore: '#ef5350'
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
    name: '', jobType: 'Full', sourceId: '', sourceType: 'vm',
    sourceHost: '', destinationId: '', schedule: '0 2 * * *', enabled: true,
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
    setScheduleTab(1); // Default to Simple
    setFormData({
      name: '', jobType: 'Full', sourceId: '', sourceType: 'vm',
      sourceHost: '', destinationId: '', schedule: '0 2 * * *', enabled: true,
      options: JSON.stringify({ compression: 'zstd', retention: 7 })
    });
    setOpen(true);
  };

  const openEdit = (job) => {
    setEditJob(job);
    
    let tab = 2;
    if (!job.schedule) tab = 0;
    else if (job.schedule.split(' ').length === 5) tab = 1;
    
    setScheduleTab(tab);
    setFormData({
      name: job.name || '',
      jobType: job.jobType || 'Full',
      sourceId: job.sourceId || '',
      sourceType: job.sourceType || 'vm',
      sourceHost: job.sourceHost || '',
      destinationId: job.destinationId || '',
      schedule: job.schedule || '0 2 * * *',
      enabled: job.enabled !== false,
      options: typeof job.options === 'string' ? job.options : JSON.stringify(job.options || {})
    });
    setOpen(true);
  };

  const handleSaveJob = async () => {
    try {
      const dataToSave = { ...formData, schedule: scheduleTab === 0 ? null : formData.schedule };
      if (editJob) {
        await fetchWithAuth(`/api/jobs/${editJob.jobId}`, {
          method: 'PUT',
          body: JSON.stringify({ ...dataToSave, jobId: editJob.jobId })
        });
      } else {
        await fetchWithAuth('/api/jobs', { method: 'POST', body: JSON.stringify(dataToSave) });
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
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Завдання бекапу</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          Створити завдання
        </Button>
      </Box>

      {jobs.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <BackupIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>Немає налаштованих завдань</Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>Створіть своє перше завдання, щоб почати захист даних</Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ bgcolor: '#4fc3f7' }}>Створити завдання</Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>СТАТУС</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>НАЗВА</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ТИП</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ДЖЕРЕЛО</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ПРИЗНАЧЕННЯ</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>РОЗКЛАД</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ОСТАННІЙ ЗАПУСК</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>НАСТУПНИЙ ЗАПУСК</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ДІЇ</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {jobs.map((job) => (
                <TableRow key={job.jobId || job.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>{statusIcon(job)}</TableCell>
                  <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                  <TableCell>
                    <Chip label={String(job.jobType || 'Full').replace('_', ' ')} size="small"
                      sx={{ fontSize: '0.7rem', height: 22, bgcolor: (jobTypeColors[job.jobType] || '#999') + '20', color: jobTypeColors[job.jobType] || '#999', fontWeight: 'bold' }} />
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
        <DialogTitle>{editJob ? 'Редагувати' : 'Створити'} завдання бекапу</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 0.5 }}>
            <Grid item xs={12} sm={6}>
              <TextField fullWidth label="Назва завдання" value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })} size="small" required />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Tooltip title="Наприклад: 'Full' для повного бекапу, 'Incremental' для інкрементального">
                <FormControl fullWidth size="small">
                  <InputLabel>Тип бекапу</InputLabel>
                  <Select value={formData.jobType} label="Тип бекапу"
                    onChange={(e) => setFormData({ ...formData, jobType: e.target.value })}>
                    <MenuItem value="Full">Повний</MenuItem>
                    <MenuItem value="Incremental">Інкрементальний</MenuItem>
                    <MenuItem value="Differential">Диференціальний</MenuItem>
                  </Select>
                </FormControl>
              </Tooltip>
            </Grid>

            <Grid item xs={12} sm={6}>
              <Tooltip title="Виберіть тип джерела: ВМ, агент, папка або база даних">
                <FormControl fullWidth size="small">
                  <InputLabel>Тип джерела</InputLabel>
                  <Select value={formData.sourceType} label="Тип джерела"
                    onChange={(e) => setFormData({ ...formData, sourceType: e.target.value, sourceId: '' })}>
                    <MenuItem value="vm">Віртуальна машина</MenuItem>
                    <MenuItem value="agent">Агент / Сервер</MenuItem>
                    <MenuItem value="folder">Папка / Файли</MenuItem>
                    <MenuItem value="database">База даних</MenuItem>
                  </Select>
                </FormControl>
              </Tooltip>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Tooltip title="Вкажіть шлях до файлів або IP-адресу (напр. C:\\Backups або 192.168.1.50)">
                <TextField fullWidth size="small" sx={{ mt: 1 }}
                  label="Шлях або ID джерела"
                  value={formData.sourceId}
                  onChange={(e) => setFormData({ ...formData, sourceId: e.target.value })}
                  placeholder="напр. C:\\Data або 192.168.1.50"
                />
              </Tooltip>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Tooltip title="Вкажіть локальний шлях (напр. D:\\Backups) або мережевий шлях (напр. \\\\Server\\Share)">
                <TextField fullWidth size="small" sx={{ mt: 1 }}
                  label="Шлях або ID сховища"
                  value={formData.destinationId === '__manual__' ? '' : formData.destinationId}
                  onChange={(e) => setFormData({ ...formData, destinationId: e.target.value })}
                  placeholder="напр. D:\\Backups або \\\\Server\\Share"
                />
              </Tooltip>
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControlLabel
                control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
                label="Завдання активне"
              />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>Розклад</Typography>
              <Tabs value={scheduleTab} onChange={(e, v) => setScheduleTab(v)} sx={{ mb: 2, borderBottom: 1, borderColor: 'divider' }}>
                <Tab label="Ручний" />
                <Tab label="Простий" />
                <Tab label="Cron" />
              </Tabs>

              {scheduleTab === 0 ? (
                <Box p={2} bgcolor="#f8f9fa" borderRadius={1} textAlign="center">
                  <Typography variant="body2" color="text.secondary">
                    Завдання буде запускатися тільки вручну або через API.
                  </Typography>
                  <Button variant="outlined" sx={{ mt: 1 }} size="small" onClick={() => setFormData({ ...formData, schedule: null })}>
                    Встановити ручний режим
                  </Button>
                </Box>
              ) : scheduleTab === 1 ? (
                <Box display="flex" flexDirection="column" gap={2}>
                  <Box display="flex" gap={2} alignItems="center">
                    <FormControl size="small" sx={{ minWidth: 150 }}>
                      <InputLabel>Періодичність</InputLabel>
                      <Select
                        value={formData.schedule?.includes('* * *') ? 'daily' : (formData.schedule?.split(' ').length === 5 && formData.schedule?.split(' ')[4] !== '*' ? 'weekly' : 'monthly')}
                        label="Періодичність"
                        onChange={(e) => {
                          const val = e.target.value;
                          if (val === 'daily') setFormData({ ...formData, schedule: '0 2 * * *' });
                          else if (val === 'weekly') setFormData({ ...formData, schedule: '0 2 * * 1' });
                          else setFormData({ ...formData, schedule: '0 2 1 * *' });
                        }}
                      >
                        <MenuItem value="daily">Щодня</MenuItem>
                        <MenuItem value="weekly">Щотижня</MenuItem>
                        <MenuItem value="monthly">Щомісяця</MenuItem>
                      </Select>
                    </FormControl>

                    <TextField
                      label="Час"
                      type="time"
                      size="small"
                      value={(() => {
                        const parts = (formData.schedule || '0 2 * * *').split(' ');
                        const h = parts[1].padStart(2, '0');
                        const m = parts[0].padStart(2, '0');
                        return `${h}:${m}`;
                      })()}
                      onChange={(e) => {
                        const [h, m] = e.target.value.split(':');
                        const parts = (formData.schedule || '0 2 * * *').split(' ');
                        parts[0] = parseInt(m).toString();
                        parts[1] = parseInt(h).toString();
                        setFormData({ ...formData, schedule: parts.join(' ') });
                      }}
                      InputLabelProps={{ shrink: true }}
                    />
                  </Box>

                  {formData.schedule?.split(' ').length === 5 && formData.schedule?.split(' ')[4] !== '*' && (
                    <FormControl size="small" fullWidth>
                      <InputLabel>День тижня</InputLabel>
                      <Select
                        value={formData.schedule?.split(' ')[4]}
                        label="День тижня"
                        onChange={(e) => {
                          const parts = formData.schedule.split(' ');
                          parts[4] = e.target.value;
                          setFormData({ ...formData, schedule: parts.join(' ') });
                        }}
                      >
                        <MenuItem value="0">Неділя</MenuItem>
                        <MenuItem value="1">Понеділок</MenuItem>
                        <MenuItem value="2">Вівторок</MenuItem>
                        <MenuItem value="3">Середа</MenuItem>
                        <MenuItem value="4">Четвер</MenuItem>
                        <MenuItem value="5">П'ятниця</MenuItem>
                        <MenuItem value="6">Субота</MenuItem>
                      </Select>
                    </FormControl>
                  )}

                  {formData.schedule?.split(' ').length === 5 && formData.schedule?.split(' ')[2] !== '*' && (
                    <TextField
                      label="День місяця"
                      type="number"
                      size="small"
                      fullWidth
                      InputProps={{ inputProps: { min: 1, max: 31 } }}
                      value={formData.schedule?.split(' ')[2]}
                      onChange={(e) => {
                        const parts = formData.schedule.split(' ');
                        parts[2] = e.target.value;
                        setFormData({ ...formData, schedule: parts.join(' ') });
                      }}
                    />
                  )}
                </Box>
              ) : (
                <Tooltip title="Формат: хвилини (0-59) години (0-23) дні (1-31) місяці (1-12) дні_тижня (0-6)">
                  <TextField fullWidth label="Cron вираз" value={formData.schedule || ''}
                    onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                    size="small" placeholder="0 2 * * *" helperText="хв год день місяць день_тижня" />
                </Tooltip>
              )}
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Скасувати</Button>
          <Button variant="contained" onClick={handleSaveJob} disabled={!formData.name || !formData.sourceId || formData.sourceId === '__manual__' || !formData.destinationId || formData.destinationId === '__manual__'}
            sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
            {editJob ? 'Зберегти' : 'Створити'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
