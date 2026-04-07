import React, { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Grid,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography
} from '@mui/material';
import {
  Backup as BackupIcon,
  CheckCircle as CheckCircleIcon,
  Dns as DnsIcon,
  Error as ErrorIcon,
  Restore as RestoreIcon,
  Storage as StorageIcon,
  Warning as WarningIcon,
  DeveloperBoard as VMIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const getBackupsFromPayload = (payload) => {
  if (Array.isArray(payload)) return payload;
  return payload?.backups || [];
};

const formatBytes = (bytes) => {
  if (!bytes) return '-';
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
};

export default function Dashboard() {
  const [summary, setSummary] = useState(null);
  const [activity, setActivity] = useState([]);
  const [jobs, setJobs] = useState([]);
  const [backups, setBackups] = useState([]);
  const [repos, setRepos] = useState([]);
  const [vms, setVMs] = useState([]);
  const [hypervisors, setHypervisors] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [summaryRes, activityRes, jobsRes, backupsRes, reposRes, vmsRes, hypRes] = await Promise.all([
        fetchWithAuth('/api/reports/summary'),
        fetchWithAuth('/api/reports/activity'),
        fetchWithAuth('/api/jobs'),
        fetchWithAuth('/api/backups'),
        fetchWithAuth('/api/repositories'),
        fetchWithAuth('/api/virtualmachines'),
        fetchWithAuth('/api/hypervisors')
      ]);

      const [summaryData, activityData, jobsData, backupsData, reposData, vmsData, hypData] = await Promise.all([
        summaryRes.json().catch(() => ({})),
        activityRes.json().catch(() => []),
        jobsRes.json().catch(() => ({ jobs: [] })),
        backupsRes.json().catch(() => ({ backups: [] })),
        reposRes.json().catch(() => []),
        vmsRes.json().catch(() => []),
        hypRes.json().catch(() => [])
      ]);

      setSummary(summaryData);
      setActivity(Array.isArray(activityData) ? activityData : []);
      setJobs(jobsData.jobs || (Array.isArray(jobsData) ? jobsData : []));
      setBackups(getBackupsFromPayload(backupsData));
      setRepos(Array.isArray(reposData) ? reposData : []);
      setVMs(Array.isArray(vmsData) ? vmsData : []);
      setHypervisors(Array.isArray(hypData) ? hypData : []);
    } catch (error) {
      console.error('Dashboard load error:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  const totalJobs = summary?.totalJobs || jobs.length;
  const totalBackups = summary?.totalBackups || backups.length;
  const totalRepos = summary?.totalRepositories || repos.length;
  const successCount = backups.filter((b) => ['Completed', 'Verified', 'completed', 'verified'].includes(String(b.status))).length;
  const failedCount = backups.filter((b) => String(b.status).toLowerCase() === 'failed').length;
  const activeJobs = summary?.activeJobs ?? jobs.filter((job) => job.enabled).length;
  const protectionRate = totalBackups > 0 ? Math.round((successCount / totalBackups) * 100) : 100;

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3, color: '#1a1d23' }}>
        Dashboard
      </Typography>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #4fc3f7', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Virtual Machines</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{vms.length}</Typography>
                </Box>
                <VMIcon sx={{ fontSize: 48, color: '#4fc3f7', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #66bb6a', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Hypervisors</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{hypervisors.length}</Typography>
                </Box>
                <DnsIcon sx={{ fontSize: 48, color: '#66bb6a', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #ffa726', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Backup Jobs</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalJobs}</Typography>
                </Box>
                <BackupIcon sx={{ fontSize: 48, color: '#ffa726', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #ef5350', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Repositories</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalRepos}</Typography>
                </Box>
                <StorageIcon sx={{ fontSize: 48, color: '#ef5350', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Card sx={{ borderRadius: 2, mb: 3 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Protection Overview</Typography>
              <Box display="flex" flexDirection={{ xs: 'column', md: 'row' }} gap={3}>
                <Box minWidth={180}>
                  <Typography variant="h3" sx={{ fontWeight: 700, mb: 1 }}>
                    {protectionRate}%
                  </Typography>
                  <LinearProgress
                    variant="determinate"
                    value={protectionRate}
                    sx={{
                      height: 10,
                      borderRadius: 5,
                      bgcolor: '#e8eaed',
                      '& .MuiLinearProgress-bar': {
                        bgcolor: protectionRate >= 80 ? '#66bb6a' : protectionRate >= 50 ? '#ffa726' : '#ef5350'
                      }
                    }}
                  />
                </Box>
                <Box flexGrow={1}>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                    <CheckCircleIcon sx={{ color: '#66bb6a', fontSize: 18 }} />
                    <Typography variant="body2">Successful backups: {successCount}</Typography>
                  </Box>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                    <ErrorIcon sx={{ color: '#ef5350', fontSize: 18 }} />
                    <Typography variant="body2">Failed backups: {failedCount}</Typography>
                  </Box>
                  <Box display="flex" alignItems="center" gap={1}>
                    <WarningIcon sx={{ color: '#ffa726', fontSize: 18 }} />
                    <Typography variant="body2">Active jobs: {activeJobs}</Typography>
                  </Box>
                </Box>
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Recent Jobs</Typography>
              {jobs.length > 0 ? (
                <Table>
                  <TableHead>
                    <TableRow sx={{ borderBottom: '1px solid #e8eaed' }}>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NAME</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LAST RUN</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NEXT RUN</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {jobs.slice(0, 8).map((job) => (
                      <TableRow key={job.jobId || job.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                        <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                        <TableCell>
                          <Chip label={job.jobType || 'Full'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} />
                        </TableCell>
                        <TableCell>
                          <Chip
                            icon={job.enabled ? <CheckCircleIcon sx={{ fontSize: 14 }} /> : <WarningIcon sx={{ fontSize: 14 }} />}
                            label={job.enabled ? 'Enabled' : 'Disabled'}
                            color={job.enabled ? 'success' : 'default'}
                            size="small"
                            sx={{ fontSize: '0.7rem', height: 22 }}
                          />
                        </TableCell>
                        <TableCell sx={{ fontSize: '0.85rem' }}>{job.lastRun ? new Date(job.lastRun).toLocaleString() : '-'}</TableCell>
                        <TableCell sx={{ fontSize: '0.85rem' }}>{job.nextRun ? new Date(job.nextRun).toLocaleString() : '-'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Box textAlign="center" py={4}>
                  <Typography variant="body2" color="text.secondary">No backup jobs configured</Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card sx={{ borderRadius: 2, mb: 3 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Storage Overview</Typography>
              {repos.length > 0 ? (
                <Box display="flex" flexDirection="column" gap={2}>
                  {repos.map((repo) => {
                    const pct = repo.capacityBytes > 0 ? Math.round((repo.usedBytes / repo.capacityBytes) * 100) : 0;
                    const color = pct > 90 ? '#ef5350' : pct > 70 ? '#ffa726' : '#66bb6a';

                    return (
                      <Box key={repo.repositoryId || repo.id}>
                        <Box display="flex" justifyContent="space-between" mb={0.5}>
                          <Typography variant="body2" fontWeight="medium">{repo.name}</Typography>
                          <Typography variant="body2" sx={{ color: '#8b92a5' }}>
                            {formatBytes(repo.usedBytes)} / {formatBytes(repo.capacityBytes)} ({pct}%)
                          </Typography>
                        </Box>
                        <LinearProgress
                          variant="determinate"
                          value={pct}
                          sx={{
                            height: 6,
                            borderRadius: 3,
                            bgcolor: '#e8eaed',
                            '& .MuiLinearProgress-bar': { bgcolor: color }
                          }}
                        />
                        <Typography variant="caption" color="text.secondary">
                          {repo.type?.toUpperCase() || 'LOCAL'} - {repo.path || '-'}
                        </Typography>
                      </Box>
                    );
                  })}
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary">No repositories configured</Typography>
              )}
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Recent Activity</Typography>
              {activity.length > 0 ? (
                <Box display="flex" flexDirection="column" gap={1.5}>
                  {activity.slice(0, 8).map((item, index) => (
                    <Box
                      key={`${item.runId || item.jobId || 'activity'}-${index}`}
                      display="flex"
                      alignItems="center"
                      gap={1.5}
                      py={0.5}
                      borderBottom="1px solid #f0f0f0"
                    >
                      {item.status === 'completed' ? <CheckCircleIcon sx={{ color: '#66bb6a', fontSize: 18 }} /> :
                        item.status === 'failed' ? <ErrorIcon sx={{ color: '#ef5350', fontSize: 18 }} /> :
                          <WarningIcon sx={{ color: '#ffa726', fontSize: 18 }} />}
                      <Box flexGrow={1}>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>
                          {item.jobName || item.jobId || 'Job'}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {item.type || item.status || ''}
                        </Typography>
                      </Box>
                      <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
                        {item.startTime || item.createdAt
                          ? new Date(item.startTime || item.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                          : ''}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary">No recent activity</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box mt={3}>
        <Card sx={{ borderRadius: 2 }}>
          <CardContent>
            <Typography variant="h6" sx={{ fontWeight: 600, mb: 1 }}>Restore Points</Typography>
            <Typography variant="body2" color="text.secondary">
              Total restore points: {summary?.totalBackups ?? totalBackups}
            </Typography>
            <Box display="flex" alignItems="center" gap={1} mt={1}>
              <RestoreIcon sx={{ color: '#ab47bc', fontSize: 18 }} />
              <Typography variant="body2">Successful backup ratio is tracked from current backup records.</Typography>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
}
