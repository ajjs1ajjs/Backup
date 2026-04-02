import React, { useState, useEffect } from 'react';
import { Box, Card, CardContent, Grid, Typography, Chip, CircularProgress, LinearProgress } from '@mui/material';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line } from 'recharts';
import { useApi } from '../services/ApiContext';

export default function Dashboard() {
  const { data: summary, loading, refetch } = useApi('/api/reports/summary');
  const { data: activity } = useApi('/api/reports/activity');

  const statsData = [
    { name: 'Jobs', value: summary?.totalJobs || 0, color: '#1976d2' },
    { name: 'Backups', value: summary?.totalBackups || 0, color: '#2e7d32' },
    { name: 'Repositories', value: summary?.totalRepositories || 0, color: '#ed6c02' },
    { name: 'Agents', value: summary?.totalAgents || 0, color: '#9c27b0' },
  ];

  const performanceData = [
    { time: '00:00', speed: 120, size: 45 },
    { time: '04:00', speed: 145, size: 52 },
    { time: '08:00', speed: 180, size: 78 },
    { time: '12:00', speed: 165, size: 65 },
    { time: '16:00', speed: 190, size: 85 },
    { time: '20:00', speed: 175, size: 70 },
    { time: '24:00', speed: 160, size: 60 },
  ];

  if (loading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

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

        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Backup Performance (MB/s)</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={performanceData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Line type="monotone" dataKey="speed" stroke="#1976d2" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Storage Usage</Typography>
              <Box display="flex" flexDirection="column" gap={2}>
                <Box>
                  <Typography variant="body2">Local Storage</Typography>
                  <LinearProgress variant="determinate" value={65} sx={{ height: 8, borderRadius: 4 }} />
                  <Typography variant="caption">65% used</Typography>
                </Box>
                <Box>
                  <Typography variant="body2">Cloud Storage</Typography>
                  <LinearProgress variant="determinate" value={35} sx={{ height: 8, borderRadius: 4 }} color="success" />
                  <Typography variant="caption">35% used</Typography>
                </Box>
                <Box>
                  <Typography variant="body2">Archive Storage</Typography>
                  <LinearProgress variant="determinate" value={15} sx={{ height: 8, borderRadius: 4 }} color="secondary" />
                  <Typography variant="caption">15% used</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Recent Activity</Typography>
              {activity?.length > 0 ? (
                <Box>
                  {activity.slice(0, 5).map((item: any, index: number) => (
                    <Box key={index} display="flex" alignItems="center" py={1} borderBottom="1px solid #eee">
                      <Chip 
                        label={item.status} 
                        color={item.status === 'completed' ? 'success' : item.status === 'failed' ? 'error' : 'warning'}
                        size="small" 
                        sx={{ mr: 2 }}
                      />
                      <Typography variant="body2">Job {item.jobId}</Typography>
                      <Box flexGrow={1} />
                      <Typography variant="caption" color="text.secondary">
                        {new Date(item.startTime).toLocaleString()}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : (
                <Typography color="text.secondary">No recent activity</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
