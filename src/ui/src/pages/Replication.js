import React from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, CircularProgress, Paper } from '@mui/material';
import { Add as AddIcon, CloudOff as ReplicateIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Replication() {
  const { data, loading } = useApi('/api/replication');

  if (loading) return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;

  const replications = data?.replications || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Replication</Typography>
        <Button variant="contained" startIcon={<AddIcon />} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          New Replication Job
        </Button>
      </Box>

      {replications.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <ReplicateIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No replication jobs configured</Typography>
            <Typography variant="body2" color="text.secondary">Replicate VMs to a secondary site for disaster recovery</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NAME</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>SOURCE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TARGET</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LAST REPLICATION</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ACTIONS</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {replications.map((rep) => (
                <TableRow key={rep.id || rep.replicationId} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell sx={{ fontWeight: 500 }}>{rep.name}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{rep.source || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{rep.target || '-'}</TableCell>
                  <TableCell>
                    <Chip label={rep.status || 'idle'} color={rep.status === 'running' ? 'warning' : rep.status === 'completed' ? 'success' : 'default'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{rep.lastReplication ? new Date(rep.lastReplication).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <Button size="small" color="error" sx={{ minWidth: 32, p: 0.5 }}><DeleteIcon fontSize="small" /></Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
