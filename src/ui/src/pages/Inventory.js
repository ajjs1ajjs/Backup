import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, TextField, Select, MenuItem, CircularProgress, Paper, Tabs, Tab } from '@mui/material';
import { Add as AddIcon, DesktopWindows as VmIcon, LaptopMac as LaptopIcon, Cloud as CloudIcon, Storage as StorageIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Inventory() {
  const { data: agents, loading: agentsLoading } = useApi('/api/agents');
  const { data: repos } = useApi('/api/repositories');
  const [tab, setTab] = useState(0);

  if (agentsLoading) return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;

  const agentsList = agents || [];
  const reposList = repos?.repositories || [];

  const tabs = [
    { label: 'All Workloads', count: agentsList.length },
    { label: 'Virtual Machines', count: agentsList.filter(a => a.agentType === 'vmware' || a.agentType === 'hyperv' || a.agentType === 'kvm').length },
    { label: 'Physical', count: agentsList.filter(a => a.agentType === 'physical').length },
    { label: 'Cloud', count: agentsList.filter(a => a.agentType === 'aws' || a.agentType === 'azure').length },
  ];

  const getIcon = (type) => {
    if (['vmware', 'hyperv', 'kvm', 'proxmox'].includes(type)) return <VmIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#4fc3f7' }} />;
    if (type === 'physical') return <LaptopIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#66bb6a' }} />;
    if (['aws', 'azure', 'gcp'].includes(type)) return <CloudIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#ffa726' }} />;
    return <VmIcon sx={{ mr: 1, verticalAlign: 'middle' }} />;
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Inventory</Typography>
        <Button variant="contained" startIcon={<AddIcon />} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          Add Workload
        </Button>
      </Box>

      <Tabs value={tab} onChange={(e, v) => setTab(v)} sx={{ mb: 3, borderBottom: '1px solid #e8eaed' }}>
        {tabs.map((t) => (
          <Tab key={t.label} label={`${t.label} (${t.count})`} sx={{ textTransform: 'none', fontWeight: 500 }} />
        ))}
      </Tabs>

      {agentsList.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <StorageIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>No workloads discovered</Typography>
            <Typography variant="body2" color="text.secondary">Deploy agents or connect to hypervisors to discover workloads</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2, overflow: 'hidden' }}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: '#f5f6f8' }}>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>WORKLOAD</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>PLATFORM</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>IP ADDRESS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>LAST SEEN</TableCell>
                <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>ACTIONS</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {agentsList.map((agent) => (
                <TableRow key={agent.id || agent.agentId} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>
                    <Box display="flex" alignItems="center">
                      {getIcon(agent.agentType)}
                      <Box>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>{agent.hostname || 'Unknown'}</Typography>
                        <Typography variant="caption" color="text.secondary">{agent.osType || ''}</Typography>
                      </Box>
                    </Box>
                  </TableCell>
                  <TableCell><Chip label={agent.agentType || 'unknown'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} /></TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{agent.agentVersion || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem', fontFamily: 'monospace' }}>{agent.ipAddress || '-'}</TableCell>
                  <TableCell>
                    <Chip label={agent.status || 'offline'} color={agent.status === 'idle' ? 'success' : agent.status === 'backing_up' ? 'warning' : 'error'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{agent.lastHeartbeat ? new Date(agent.lastHeartbeat).toLocaleString() : 'Never'}</TableCell>
                  <TableCell>
                    <Button size="small" color="error" onClick={() => {}} sx={{ minWidth: 32, p: 0.5 }}><DeleteIcon fontSize="small" /></Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {reposList.length > 0 && (
        <Card sx={{ borderRadius: 2, mt: 3 }}>
          <CardContent>
            <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>Storage Repositories</Typography>
            <Table>
              <TableHead>
                <TableRow sx={{ borderBottom: '1px solid #e8eaed' }}>
                  <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>NAME</TableCell>
                  <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>TYPE</TableCell>
                  <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>PATH</TableCell>
                  <TableCell sx={{ color: '#8b92a5', fontWeight: 600, fontSize: '0.75rem' }}>STATUS</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {reposList.map((repo) => (
                  <TableRow key={repo.id || repo.repositoryId}>
                    <TableCell sx={{ fontWeight: 500 }}>{repo.name}</TableCell>
                    <TableCell><Chip label={repo.type || 'local'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} /></TableCell>
                    <TableCell sx={{ fontSize: '0.85rem', fontFamily: 'monospace' }}>{repo.path || '-'}</TableCell>
                    <TableCell><Chip label={repo.status || 'unknown'} color={repo.status === 'online' ? 'success' : 'warning'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} /></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </Box>
  );
}
