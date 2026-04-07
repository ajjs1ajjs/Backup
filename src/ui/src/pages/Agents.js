import React from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  Computer as ComputerIcon,
  Delete as DeleteIcon
} from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

const getStatusColor = (status) => {
  const normalized = String(status || '').toLowerCase();
  if (normalized === 'idle' || normalized === 'online') return 'success';
  if (normalized === 'backing_up' || normalized === 'busy') return 'warning';
  return 'default';
};

export default function Agents() {
  const { data, loading, refetch } = useApi('/api/agents');
  const agents = Array.isArray(data) ? data : [];

  const handleDelete = async (agentId) => {
    try {
      await fetchWithAuth(`/api/agents/${agentId}`, { method: 'DELETE' });
    } finally {
      refetch();
    }
  };

  if (loading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Agents</Typography>
        <Button variant="contained" startIcon={<AddIcon />} disabled>
          Deploy Agent
        </Button>
      </Box>

      {agents.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No deployed agents
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Agents will appear here after they connect to the server.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Host</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>OS</TableCell>
                <TableCell>IP Address</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Last Heartbeat</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {agents.map((agent) => (
                <TableRow key={agent.agentId || agent.id}>
                  <TableCell>
                    <ComputerIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                    {agent.hostname || '-'}
                  </TableCell>
                  <TableCell>
                    <Chip label={String(agent.agentType || 'Unknown')} size="small" />
                  </TableCell>
                  <TableCell>{agent.osType || '-'}</TableCell>
                  <TableCell>{agent.ipAddress || '-'}</TableCell>
                  <TableCell>{agent.agentVersion || '-'}</TableCell>
                  <TableCell>
                    <Chip
                      label={String(agent.status || 'Offline')}
                      color={getStatusColor(agent.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {agent.lastHeartbeat ? new Date(agent.lastHeartbeat).toLocaleString() : '-'}
                  </TableCell>
                  <TableCell>
                    <IconButton
                      onClick={() => handleDelete(agent.agentId || agent.id)}
                      color="error"
                    >
                      <DeleteIcon />
                    </IconButton>
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
