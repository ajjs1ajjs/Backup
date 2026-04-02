import React from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, IconButton } from '@mui/material';
import { Add as AddIcon, Computer as ComputerIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Agents() {
  const { data, loading, refetch } = useApi('/api/agents');

  if (loading) return <Typography>Loading...</Typography>;

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Agents</Typography>
        <Button variant="contained" startIcon={<AddIcon />}>
          Deploy Agent
        </Button>
      </Box>

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Hostname</TableCell>
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
            {data?.map((agent: any) => (
              <TableRow key={agent.agentId}>
                <TableCell><ComputerIcon sx={{ mr: 1 }} />{agent.hostname}</TableCell>
                <TableCell><Chip label={agent.agentType} size="small" /></TableCell>
                <TableCell>{agent.osType}</TableCell>
                <TableCell>{agent.ipAddress}</TableCell>
                <TableCell>{agent.agentVersion}</TableCell>
                <TableCell>
                  <Chip 
                    label={agent.status} 
                    color={agent.status === 'idle' ? 'success' : agent.status === 'backing_up' ? 'warning' : 'error'}
                    size="small" 
                  />
                </TableCell>
                <TableCell>{agent.lastHeartbeat ? new Date(agent.lastHeartbeat).toLocaleString() : '-'}</TableCell>
                <TableCell>
                  <IconButton color="error"><DeleteIcon /></IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
