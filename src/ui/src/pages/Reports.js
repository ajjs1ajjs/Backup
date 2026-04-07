import React from 'react';
import { Box, Button, Card, CardContent, Grid, Tooltip, Typography } from '@mui/material';
import { Download as DownloadIcon } from '@mui/icons-material';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis
} from 'recharts';
import { useApi } from '../services/ApiContext';

export default function Reports() {
  const { data: summary } = useApi('/api/reports/summary');
  const { data: storage } = useApi('/api/reports/storage');

  const statusData = [
    { name: 'Successful', value: summary?.successfulBackups || 0, color: '#2e7d32' },
    { name: 'Failed', value: (summary?.totalBackups || 0) - (summary?.successfulBackups || 0), color: '#d32f2f' }
  ];

  const storageData = Array.isArray(storage) ? storage : [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Reports</Typography>
        <Tooltip title="PDF export is not implemented yet">
          <span>
            <Button variant="outlined" startIcon={<DownloadIcon />} disabled>
              Export PDF
            </Button>
          </span>
        </Tooltip>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Backup Status</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={statusData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label>
                    {statusData.map((entry) => (
                      <Cell key={entry.name} fill={entry.color} />
                    ))}
                  </Pie>
                  <ChartTooltip />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Repository Utilization</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={storageData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <ChartTooltip />
                  <Bar dataKey="utilizationPercent" fill="#1976d2" name="Utilization %" />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>Summary</Typography>
              <Box display="grid" gridTemplateColumns={{ xs: '1fr 1fr', md: 'repeat(4, 1fr)' }} gap={2}>
                <Box textAlign="center" p={2} bgcolor="#f5f5f5" borderRadius={2}>
                  <Typography variant="h4">{summary?.totalJobs || 0}</Typography>
                  <Typography variant="body2" color="text.secondary">Total Jobs</Typography>
                </Box>
                <Box textAlign="center" p={2} bgcolor="#f5f5f5" borderRadius={2}>
                  <Typography variant="h4">{summary?.totalBackups || 0}</Typography>
                  <Typography variant="body2" color="text.secondary">Total Backups</Typography>
                </Box>
                <Box textAlign="center" p={2} bgcolor="#f5f5f5" borderRadius={2}>
                  <Typography variant="h4">{(summary?.successRate || 100).toFixed(1)}%</Typography>
                  <Typography variant="body2" color="text.secondary">Success Rate</Typography>
                </Box>
                <Box textAlign="center" p={2} bgcolor="#f5f5f5" borderRadius={2}>
                  <Typography variant="h4">{summary?.onlineAgents || 0}/{summary?.totalAgents || 0}</Typography>
                  <Typography variant="body2" color="text.secondary">Agents Online</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
