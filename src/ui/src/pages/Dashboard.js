import React from 'react';
import { Box, Card, CardContent, Grid, Typography, CircularProgress, LinearProgress, Table, TableBody, TableCell, TableHead, TableRow, Chip } from '@mui/material';
import { useApi } from '../services/ApiContext';

export default function Dashboard() {
  const { data: summary, loading: summaryLoading } = useApi('/api/reports/summary');
  const { data: activity, loading: activityLoading } = useApi('/api/reports/activity');

  if (summaryLoading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

  const statsData = [
    { name: 'Jobs', value: summary?.totalJobs || 0, color: '#1976d2' },
    { name: 'Backups', value: summary?.totalBackups || 0, color: '#2e7d32' },
    { name: 'Repositories', value: summary?.totalRepositories || 0, color: '#ed6c02' },
    { name: 'Agents', value: summary?.totalAgents || 0, color: '#9c27b0' },
  ];

  const recentActivity = activity?.items || activity || [];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Dashboard</Typography>
      
      <Grid container spacing={3}>
        {statsData.map((stat) => (
          <Grid item xs={12} sm={6} md={3} key={stat.name}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>{stat.name}</Typography>
                <Typography variant="h3" sx={{ color: stat.color }}>{stat.value}</Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}

        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Recent Activity</Typography>
              {recentActivity.length > 0 ? (
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Status</TableCell>
                      <TableCell>Job</TableCell>
                      <TableCell>Type</TableCell>
                      <TableCell>Time</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {recentActivity.slice(0, 10).map((item, index) => (
                      <TableRow key={index}>
                        <TableCell>
                          <Chip label={item.status || 'unknown'} color={item.status === 'completed' ? 'success' : item.status === 'failed' ? 'error' : 'warning'} size="small" />
                        </TableCell>
                        <TableCell>{item.jobName || item.jobId || '-'}</TableCell>
                        <TableCell>{item.type || '-'}</TableCell>
                        <TableCell>{item.startTime || item.createdAt ? new Date(item.startTime || item.createdAt).toLocaleString() : '-'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Typography color="text.secondary" sx={{ py: 3 }}>No recent activity</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
