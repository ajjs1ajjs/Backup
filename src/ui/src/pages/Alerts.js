import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  CircularProgress,
  IconButton,
  Paper,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Tooltip,
  Typography
} from '@mui/material';
import {
  CheckCircle as CheckIcon,
  Delete as DeleteIcon,
  Error as ErrorIcon,
  Info as InfoIcon,
  Warning as WarningIcon
} from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Alerts() {
  const { data: activity, loading } = useApi('/api/reports/activity');
  const [tab, setTab] = useState(0);

  const items = Array.isArray(activity?.items)
    ? activity.items
    : Array.isArray(activity)
      ? activity
      : [];

  const getSeverity = (item) => {
    const status = String(item.status || '').toLowerCase();
    if (status === 'failed' || status === 'error') return 'error';
    if (status === 'warning') return 'warning';
    if (status === 'completed') return 'success';
    return 'info';
  };

  const getIcon = (severity) => {
    switch (severity) {
      case 'error':
        return <ErrorIcon sx={{ color: '#ef5350' }} />;
      case 'warning':
        return <WarningIcon sx={{ color: '#ffa726' }} />;
      case 'success':
        return <CheckIcon sx={{ color: '#66bb6a' }} />;
      default:
        return <InfoIcon sx={{ color: '#4fc3f7' }} />;
    }
  };

  const severityOrder = ['error', 'warning', 'success', 'info'];
  const filteredItems = tab === 0 ? items : items.filter((item) => getSeverity(item) === severityOrder[tab - 1]);

  const tabs = [
    { label: 'All', count: items.length },
    { label: 'Errors', count: items.filter((item) => getSeverity(item) === 'error').length },
    { label: 'Warnings', count: items.filter((item) => getSeverity(item) === 'warning').length },
    { label: 'Success', count: items.filter((item) => getSeverity(item) === 'success').length },
    { label: 'Info', count: items.filter((item) => getSeverity(item) === 'info').length }
  ];

  if (loading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3 }}>
        Monitoring and Alerts
      </Typography>

      <Tabs value={tab} onChange={(event, value) => setTab(value)} sx={{ mb: 3, borderBottom: '1px solid #e8eaed' }}>
        {tabs.map((item) => (
          <Tab key={item.label} label={`${item.label} (${item.count})`} sx={{ textTransform: 'none', fontWeight: 500 }} />
        ))}
      </Tabs>

      {filteredItems.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <CheckIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No alerts
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Recent activity looks healthy.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LEVEL</TableCell>
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
                  <TableRow key={`${item.runId || item.jobId || 'alert'}-${index}`} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                    <TableCell>{getIcon(severity)}</TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 500 }}>
                        {item.message || `${item.status} - ${item.type || 'job'}`}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {item.details || ''}
                      </Typography>
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.85rem' }}>{item.jobName || item.jobId || '-'}</TableCell>
                    <TableCell sx={{ fontSize: '0.85rem' }}>
                      {item.startTime || item.createdAt ? new Date(item.startTime || item.createdAt).toLocaleString() : '-'}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Alert records are currently read-only">
                        <span>
                          <IconButton size="small" disabled>
                            <DeleteIcon fontSize="small" sx={{ color: '#bdbdbd' }} />
                          </IconButton>
                        </span>
                      </Tooltip>
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
