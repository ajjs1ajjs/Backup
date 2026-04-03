import React from 'react';
import { Box, Card, CardContent, Grid, Typography, Chip, CircularProgress, LinearProgress, Table, TableBody, TableCell, TableHead, TableRow, IconButton, Tooltip } from '@mui/material';
import { CheckCircle as CheckIcon, Warning as WarningIcon, Error as ErrorIcon, Storage as StorageIcon, Backup as BackupIcon, Restore as RestoreIcon, Computer as ComputerIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Dashboard() {
  const { data: summary, loading: summaryLoading } = useApi('/api/reports/summary');
  const { data: activity, loading: activityLoading } = useApi('/api/reports/activity');
  const { data: jobs } = useApi('/api/jobs');
  const { data: backups } = useApi('/api/backups');

  if (summaryLoading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  const totalJobs = summary?.totalJobs || 0;
  const totalBackups = summary?.totalBackups || 0;
  const totalRepos = summary?.totalRepositories || 0;
  const totalAgents = summary?.totalAgents || 0;

  const jobsList = jobs?.jobs || [];
  const backupsList = backups?.backups || [];
  const recentActivity = activity?.items || activity || [];

  const successCount = backupsList.filter(b => b.status === 'completed').length;
  const failedCount = backupsList.filter(b => b.status === 'failed').length;
  const runningCount = jobsList.filter(j => j.status === 'running' || j.status === 'active').length;

  const protectionRate = totalBackups > 0 ? Math.round((successCount / totalBackups) * 100) : 100;

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3, color: '#1a1d23' }}>Dashboard</Typography>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #4fc3f7', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Protected Workloads</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalAgents}</Typography>
                </Box>
                <ComputerIcon sx={{ fontSize: 48, color: '#4fc3f7', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #66bb6a', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Backup Jobs</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalJobs}</Typography>
                </Box>
                <BackupIcon sx={{ fontSize: 48, color: '#66bb6a', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #ffa726', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Backup Points</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalBackups}</Typography>
                </Box>
                <RestoreIcon sx={{ fontSize: 48, color: '#ffa726', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #ab47bc', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Repositories</Typography>
                  <Typography variant="h3" sx={{ fontWeight: 700, color: '#1a1d23' }}>{totalRepos}</Typography>
                </Box>
                <StorageIcon sx={{ fontSize: 48, color: '#ab47bc', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Card sx={{ borderRadius: 2, mb: 3 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Backup Success Rate</Typography>
              <Box display="flex" alignItems="center" gap={3} mb={2}>
                <Box sx={{ position: 'relative', display: 'inline-flex' }}>
                  <CircularProgress variant="determinate" value={protectionRate} size={120} thickness={4} sx={{ color: protectionRate >= 80 ? '#66bb6a' : protectionRate >= 50 ? '#ffa726' : '#ef5350' }} />
                  <Box sx={{ top: 0, left: 0, bottom: 0, right: 0, position: 'absolute', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <Typography variant="h4" sx={{ fontWeight: 700 }}>{protectionRate}%</Typography>
                  </Box>
                </Box>
                <Box flexGrow={1}>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                    <CheckCircleIcon sx={{ color: '#66bb6a', fontSize: 18 }} />
                    <Typography variant="body2">Successful: {successCount}</Typography>
                  </Box>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                    <ErrorIcon sx={{ color: '#ef5350', fontSize: 18 }} />
                    <Typography variant="body2">Failed: {failedCount}</Typography>
                  </Box>
                  <Box display="flex" alignItems="center" gap={1}>
                    <WarningIcon sx={{ color: '#ffa726', fontSize: 18 }} />
                    <Typography variant="body2">Running: {runningCount}</Typography>
                  </Box>
                </Box>
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Recent Backup Jobs</Typography>
              {jobsList.length > 0 ? (
                <Table>
                  <TableHead>
                    <TableRow sx={{ borderBottom: '1px solid #e8eaed' }}>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>JOB NAME</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LAST RUN</TableCell>
                      <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NEXT RUN</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {jobsList.slice(0, 8).map((job) => (
                      <TableRow key={job.id || job.jobId} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                        <TableCell sx={{ fontWeight: 500 }}>{job.name}</TableCell>
                        <TableCell><Chip label={job.jobType || 'full'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} /></TableCell>
                        <TableCell>
                          <Chip
                            icon={job.enabled ? <CheckIcon sx={{ fontSize: 14 }} /> : <WarningIcon sx={{ fontSize: 14 }} />}
                            label={job.enabled ? 'Active' : 'Disabled'}
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
              {totalRepos > 0 ? (
                <Box display="flex" flexDirection="column" gap={2}>
                  {[
                    { name: 'Local Storage', used: 65, color: '#4fc3f7' },
                    { name: 'Cloud Storage', used: 35, color: '#66bb6a' },
                    { name: 'Archive', used: 15, color: '#ab47bc' },
                  ].map((s) => (
                    <Box key={s.name}>
                      <Box display="flex" justifyContent="space-between" mb={0.5}>
                        <Typography variant="body2">{s.name}</Typography>
                        <Typography variant="body2" sx={{ color: '#8b92a5' }}>{s.used}%</Typography>
                      </Box>
                      <LinearProgress variant="determinate" value={s.used} sx={{ height: 6, borderRadius: 3, bgcolor: '#e8eaed', '& .MuiLinearProgress-bar': { bgcolor: s.color } }} />
                    </Box>
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary">No repositories configured</Typography>
              )}
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Recent Activity</Typography>
              {recentActivity.length > 0 ? (
                <Box display="flex" flexDirection="column" gap={1.5}>
                  {recentActivity.slice(0, 8).map((item, index) => (
                    <Box key={index} display="flex" alignItems="center" gap={1.5} py={0.5} borderBottom="1px solid #f0f0f0">
                      {item.status === 'completed' ? <CheckIcon sx={{ color: '#66bb6a', fontSize: 18 }} /> :
                       item.status === 'failed' ? <ErrorIcon sx={{ color: '#ef5350', fontSize: 18 }} /> :
                       <WarningIcon sx={{ color: '#ffa726', fontSize: 18 }} />}
                      <Box flexGrow={1}>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>{item.jobName || item.jobId || 'Job'}</Typography>
                        <Typography variant="caption" color="text.secondary">{item.type || ''}</Typography>
                      </Box>
                      <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
                        {item.startTime || item.createdAt ? new Date(item.startTime || item.createdAt).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : ''}
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
    </Box>
  );
}
