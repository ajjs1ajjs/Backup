import React, { useMemo, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Paper,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  Cloud as CloudIcon,
  DesktopWindows as VmIcon,
  LaptopMac as LaptopIcon,
  Storage as StorageIcon
} from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

const workloadMatchesTab = (agentType, tab) => {
  if (tab === 0) return true;
  if (tab === 1) return ['vmware', 'hyperv', 'kvm', 'proxmox'].includes(agentType);
  if (tab === 2) return agentType === 'physical';
  if (tab === 3) return ['aws', 'azure', 'gcp'].includes(agentType);
  return true;
};

const getIcon = (type) => {
  if (['vmware', 'hyperv', 'kvm', 'proxmox'].includes(type)) {
    return <VmIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#4fc3f7' }} />;
  }
  if (type === 'physical') {
    return <LaptopIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#66bb6a' }} />;
  }
  if (['aws', 'azure', 'gcp'].includes(type)) {
    return <CloudIcon sx={{ mr: 1, verticalAlign: 'middle', color: '#ffa726' }} />;
  }
  return <VmIcon sx={{ mr: 1, verticalAlign: 'middle' }} />;
};

export default function Inventory() {
  const { data: agents, loading: agentsLoading } = useApi('/api/agents');
  const { data: repos } = useApi('/api/repositories');
  const [tab, setTab] = useState(0);

  const agentsList = Array.isArray(agents) ? agents : [];
  const reposList = Array.isArray(repos) ? repos : [];

  const tabs = useMemo(() => ([
    { label: 'All Workloads', count: agentsList.length },
    { label: 'Virtual Machines', count: agentsList.filter((a) => workloadMatchesTab(a.agentType, 1)).length },
    { label: 'Physical', count: agentsList.filter((a) => workloadMatchesTab(a.agentType, 2)).length },
    { label: 'Cloud', count: agentsList.filter((a) => workloadMatchesTab(a.agentType, 3)).length }
  ]), [agentsList]);

  const visibleAgents = agentsList.filter((agent) => workloadMatchesTab(agent.agentType, tab));

  if (agentsLoading) {
    return <Box display="flex" justifyContent="center" p={8}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Inventory</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}
          disabled
        >
          Add Workload
        </Button>
      </Box>

      <Tabs
        value={tab}
        onChange={(event, value) => setTab(value)}
        sx={{ mb: 3, borderBottom: '1px solid #e8eaed' }}
      >
        {tabs.map((item) => (
          <Tab
            key={item.label}
            label={`${item.label} (${item.count})`}
            sx={{ textTransform: 'none', fontWeight: 500 }}
          />
        ))}
      </Tabs>

      {visibleAgents.length === 0 ? (
        <Card sx={{ borderRadius: 2 }}>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <StorageIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No workloads discovered
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Deploy agents or connect hypervisors to populate inventory.
            </Typography>
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
              </TableRow>
            </TableHead>
            <TableBody>
              {visibleAgents.map((agent) => (
                <TableRow key={agent.agentId || agent.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                  <TableCell>
                    <Box display="flex" alignItems="center">
                      {getIcon(agent.agentType)}
                      <Box>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>
                          {agent.hostname || 'Unknown'}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {agent.osType || ''}
                        </Typography>
                      </Box>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip label={agent.agentType || 'unknown'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>{agent.agentVersion || '-'}</TableCell>
                  <TableCell sx={{ fontSize: '0.85rem', fontFamily: 'monospace' }}>{agent.ipAddress || '-'}</TableCell>
                  <TableCell>
                    <Chip
                      label={agent.status || 'offline'}
                      color={agent.status === 'idle' ? 'success' : agent.status === 'backing_up' ? 'warning' : 'default'}
                      size="small"
                      sx={{ fontSize: '0.7rem', height: 22 }}
                    />
                  </TableCell>
                  <TableCell sx={{ fontSize: '0.85rem' }}>
                    {agent.lastHeartbeat ? new Date(agent.lastHeartbeat).toLocaleString() : 'Never'}
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
            <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
              Storage Repositories
            </Typography>
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
                  <TableRow key={repo.repositoryId || repo.id}>
                    <TableCell sx={{ fontWeight: 500 }}>{repo.name}</TableCell>
                    <TableCell>
                      <Chip label={repo.type || 'local'} size="small" sx={{ fontSize: '0.7rem', height: 22 }} />
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.85rem', fontFamily: 'monospace' }}>{repo.path || '-'}</TableCell>
                    <TableCell>
                      <Chip
                        label={repo.status || 'unknown'}
                        color={repo.status === 'online' ? 'success' : 'warning'}
                        size="small"
                        sx={{ fontSize: '0.7rem', height: 22 }}
                      />
                    </TableCell>
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
