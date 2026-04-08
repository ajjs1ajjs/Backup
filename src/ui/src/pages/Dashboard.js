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
  Typography,
  useTheme
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
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend
} from 'recharts';
import { fetchWithAuth } from '../services/ApiContext';

const formatBytes = (bytes) => {
  if (!bytes) return '0 MB';
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
};

export default function Dashboard() {
  const theme = useTheme();
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000); // Оновлення кожні 30 секунд
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      const res = await fetchWithAuth('/api/dashboard/stats');
      const data = await res.json();
      setStats(data);
    } catch (error) {
      console.error('Dashboard load error:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading && !stats) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  const { summary, recentBackups, backupHistory, storageStats } = stats || {
    summary: { totalJobs: 0, totalVMs: 0, totalBackups: 0, totalRepositories: 0 },
    recentBackups: [],
    backupHistory: [],
    storageStats: []
  };

  const chartData = backupHistory.map(h => ({
    name: new Date(h.startTime).toLocaleDateString([], { month: 'short', day: 'numeric' }),
    size: Math.round(h.bytesProcessed / (1024 * 1024)),
    status: h.status
  }));

  const pieData = storageStats.map(s => ({
    name: s.name,
    value: s.usedBytes,
    capacity: s.capacityBytes
  }));

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3, color: '#1a1d23' }}>
        System Dashboard
      </Typography>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #4fc3f7', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Virtual Machines</Typography>
                  <Typography variant="h4" sx={{ fontWeight: 700, color: '#1a1d23' }}>{summary.totalVMs}</Typography>
                </Box>
                <VMIcon sx={{ fontSize: 40, color: '#4fc3f7', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #66bb6a', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Backup Points</Typography>
                  <Typography variant="h4" sx={{ fontWeight: 700, color: '#1a1d23' }}>{summary.totalBackups}</Typography>
                </Box>
                <RestoreIcon sx={{ fontSize: 40, color: '#66bb6a', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ borderLeft: '4px solid #ffa726', borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="body2" sx={{ color: '#8b92a5', mb: 0.5 }}>Active Jobs</Typography>
                  <Typography variant="h4" sx={{ fontWeight: 700, color: '#1a1d23' }}>{summary.totalJobs}</Typography>
                </Box>
                <BackupIcon sx={{ fontSize: 40, color: '#ffa726', opacity: 0.3 }} />
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
                  <Typography variant="h4" sx={{ fontWeight: 700, color: '#1a1d23' }}>{summary.totalRepositories}</Typography>
                </Box>
                <StorageIcon sx={{ fontSize: 40, color: '#ef5350', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Card sx={{ borderRadius: 2, mb: 3 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 3 }}>Backup Data Processed (MB)</Typography>
              <Box sx={{ height: 300, width: '100%' }}>
                <ResponsiveContainer>
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="colorSize" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#8884d8" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                    <XAxis dataKey="name" stroke="#8b92a5" fontSize={12} tickLine={false} axisLine={false} />
                    <YAxis stroke="#8b92a5" fontSize={12} tickLine={false} axisLine={false} />
                    <Tooltip 
                      contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                    />
                    <Area type="monotone" dataKey="size" stroke="#8884d8" fillOpacity={1} fill="url(#colorSize)" />
                  </AreaChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Recent Backup Points</Typography>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ color: '#8b92a5', fontWeight: 600 }}>VM ID</TableCell>
                    <TableCell sx={{ color: '#8b92a5', fontWeight: 600 }}>TYPE</TableCell>
                    <TableCell sx={{ color: '#8b92a5', fontWeight: 600 }}>SIZE</TableCell>
                    <TableCell sx={{ color: '#8b92a5', fontWeight: 600 }}>DATE</TableCell>
                    <TableCell sx={{ color: '#8b92a5', fontWeight: 600 }}>STATUS</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {recentBackups.map((b) => (
                    <TableRow key={b.backupId}>
                      <TableCell sx={{ fontWeight: 500 }}>{b.vmId || 'Host'}</TableCell>
                      <TableCell><Chip label={b.backupType} size="small" variant="outlined" sx={{ fontSize: 10, height: 20 }} /></TableCell>
                      <TableCell>{formatBytes(b.sizeBytes)}</TableCell>
                      <TableCell sx={{ fontSize: 12 }}>{new Date(b.createdAt).toLocaleString()}</TableCell>
                      <TableCell>
                        <Chip 
                          label={b.status} 
                          size="small" 
                          color={b.status.toLowerCase() === 'completed' || b.status.toLowerCase() === 'verified' ? 'success' : 'error'}
                          sx={{ fontSize: 10, height: 20 }}
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card sx={{ borderRadius: 2, mb: 3 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Storage Distribution</Typography>
              <Box sx={{ height: 250, width: '100%' }}>
                <ResponsiveContainer>
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={80}
                      paddingAngle={5}
                      dataKey="value"
                    >
                      {pieData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => formatBytes(value)} />
                    <Legend verticalAlign="bottom" height={36}/>
                  </PieChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ borderRadius: 2 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Storage Health</Typography>
              <Box display="flex" flexDirection="column" gap={2}>
                {storageStats.map((repo, index) => {
                  const pct = repo.capacityBytes > 0 ? Math.round((repo.usedBytes / repo.capacityBytes) * 100) : 0;
                  const color = pct > 90 ? '#ef5350' : pct > 70 ? '#ffa726' : '#66bb6a';
                  return (
                    <Box key={index}>
                      <Box display="flex" justifyContent="space-between" mb={0.5}>
                        <Typography variant="body2" fontWeight="medium">{repo.name}</Typography>
                        <Typography variant="body2" sx={{ color: '#8b92a5' }}>{pct}%</Typography>
                      </Box>
                      <LinearProgress 
                        variant="determinate" 
                        value={pct} 
                        sx={{ height: 6, borderRadius: 3, bgcolor: '#f0f0f0', '& .MuiLinearProgress-bar': { bgcolor: color } }} 
                      />
                    </Box>
                  );
                })}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
