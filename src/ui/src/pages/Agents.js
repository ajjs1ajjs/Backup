import React from 'react';
import { Box, Card, CardContent, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Button, IconButton, CircularProgress } from '@mui/material';
import { Add as AddIcon, Computer as ComputerIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Agents() {
  const { data, loading, refetch } = useApi('/api/agents');

  const handleDelete = async (agentId) => {
    try { await fetchWithAuth(`/api/agents/${agentId}`, { method: 'DELETE' }); refetch(); } catch (e) { refetch(); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  const agents = data || [];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Агенти</Typography>
        <Button variant="contained" startIcon={<AddIcon />}>
          Розгорнути агент
        </Button>
      </Box>

      {agents.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>Немає розгорнутих агентів</Typography>
            <Typography variant="body2" color="text.secondary">Розгорніть агент на цільовій машині для керування бекапами</Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Ім'я хоста</TableCell>
                <TableCell>Тип</TableCell>
                <TableCell>ОС</TableCell>
                <TableCell>IP-адреса</TableCell>
                <TableCell>Версія</TableCell>
                <TableCell>Статус</TableCell>
                <TableCell>Остання активність</TableCell>
                <TableCell>Дії</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {agents.map((agent) => (
                <TableRow key={agent.id || agent.agentId}>
                  <TableCell><ComputerIcon sx={{ mr: 1, verticalAlign: 'middle' }} />{agent.hostname || '-'}</TableCell>
                  <TableCell><Chip label={agent.agentType || 'невідомо'} size="small" /></TableCell>
                  <TableCell>{agent.osType || '-'}</TableCell>
                  <TableCell>{agent.ipAddress || '-'}</TableCell>
                  <TableCell>{agent.agentVersion || '-'}</TableCell>
                  <TableCell>
                    <Chip label={agent.status || 'офлайн'} color={agent.status === 'idle' ? 'success' : agent.status === 'backing_up' ? 'warning' : 'error'} size="small" />
                  </TableCell>
                  <TableCell>{agent.lastHeartbeat ? new Date(agent.lastHeartbeat).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <IconButton onClick={() => handleDelete(agent.id || agent.agentId)} color="error"><DeleteIcon /></IconButton>
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
