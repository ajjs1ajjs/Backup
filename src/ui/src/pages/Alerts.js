import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, CircularProgress, Paper, Tabs, Tab, IconButton, Tooltip } from '@mui/material';
import { CheckCircle as CheckIcon, Warning as WarningIcon, Error as ErrorIcon, Info as InfoIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Alerts() {
  const { data: activity, loading } = useApi('/api/reports/activity');
  const [tab, setTab] = useState(0);

  if (loading) return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;

  const items = activity?.items || activity || [];

  const getSeverity = (item) => {
    if (item.status === 'failed' || item.status === 'error') return 'error';
    if (item.status === 'warning') return 'warning';
    if (item.status === 'completed') return 'success';
    return 'info';
  };

  const getIcon = (severity) => {
    switch (severity) {
      case 'error': return <ErrorIcon sx={{ color: '#ef5350' }} />;
      case 'warning': return <WarningIcon sx={{ color: '#ffa726' }} />;
      case 'success': return <CheckIcon sx={{ color: '#66bb6a' }} />;
      default: return <InfoIcon sx={{ color: '#4fc3f7' }} />;
    }
  };

  const filteredItems = tab === 0 ? items : items.filter(i => getSeverity(i) === ['error', 'warning', 'success', 'info'][tab]);

  const tabs = [
    { label: 'All', count: items.length },
    { label: 'Errors', count: items.filter(i => getSeverity(i) === 'error').length },
    { label: 'Warnings', count: items.filter(i => getSeverity(i) === 'warning').length },
    { label: 'Success', count: items.filter(i => getSeverity(i) === 'success').length },
    { label: 'Info', count: items.filter(i => getSeverity(i) === 'info').length },
  ];

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3 }}>Monitoring & Alerts</Typography>

      <Tabs value={tab} onChange={(e, v) => setTab(v)} sx={{ mb: 3, borderBottom: '1px solid #e8eaed' }}>
        {tabs.map((t) => (
          <Tab key={t.label} label={`${t.label} (${t.count})`} sx={{ textTransform: 'none', fontWeight: 500 }} />
        ))}
      </Tabs>

      {filteredItems.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <CheckIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No alerts</Typography>
            <Typography variant="body2" color="text.secondary">Everything is running smoothly</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>SEVERITY</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>MESSAGE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>JOB</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TIME</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ACTIONS</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredItems.map((item, index) => {
                const severity = getSeverity(item);
                return (
                  <TableRow key={index} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                    <TableCell>{getIcon(severity)}</TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 500 }}>{item.message || `${item.status} - ${item.type || 'job'}`}</Typography>
                      <Typography variant="caption" color="text.secondary">{item.details || ''}</Typography>
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.85rem' }}>{item.jobName || item.jobId || '-'}</TableCell>
                    <TableCell sx={{ fontSize: '0.85rem' }}>{item.startTime || item.createdAt ? new Date(item.startTime || item.createdAt).toLocaleString() : '-'}</TableCell>
                    <TableCell>
                      <Tooltip title="Dismiss"><IconButton size="small"><DeleteIcon fontSize="small" sx={{ color: '#bdbdbd' }} /></IconButton></Tooltip>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
