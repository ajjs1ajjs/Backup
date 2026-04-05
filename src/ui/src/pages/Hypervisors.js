import React, { useState, useEffect } from 'react';
import {
  Box, Card, CardContent, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, TablePagination, Chip, IconButton, Typography,
  Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField,
  MenuItem, Alert, Tooltip, Avatar, LinearProgress
} from '@mui/material';
import {
  Delete as DeleteIcon, Refresh as RefreshIcon, Add as AddIcon,
  Dns as DnsIcon, PlayArrow as PlayIcon, Stop as StopIcon,
  CheckCircle as CheckCircleIcon, Error as ErrorIcon, WifiOff as WifiOffIcon,
  Sync as SyncIcon, VpnKey as VpnKeyIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const typeColors = {
  hyperv: '#0078D4',
  vmware: '#607D8B',
  kvm: '#FF9800'
};

const statusConfig = {
  connected: { icon: <CheckCircleIcon fontSize="small" sx={{ color: '#4CAF50' }} />, label: 'Connected', color: 'success' },
  offline: { icon: <WifiOffIcon fontSize="small" sx={{ color: '#9E9E9E' }} />, label: 'Offline', color: 'default' },
  error: { icon: <ErrorIcon fontSize="small" sx={{ color: '#F44336' }} />, label: 'Error', color: 'error' },
  connecting: { icon: <SyncIcon fontSize="small" sx={{ color: '#FF9800' }} />, label: 'Connecting', color: 'warning' }
};

export default function Hypervisors() {
  const [hypervisors, setHypervisors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [openAdd, setOpenAdd] = useState(false);
  const [testing, setTesting] = useState(null);
  const [newHypervisor, setNewHypervisor] = useState({
    name: '', type: 'hyperv', host: '', port: 0, username: '', password: ''
  });

  useEffect(() => {
    fetchHypervisors();
  }, []);

  const fetchHypervisors = async () => {
    setLoading(true);
    try {
      const response = await fetchWithAuth('/api/hypervisors');
      const data = await response.json();
      setHypervisors(data);
    } catch (e) {
      setError('Failed to load hypervisors');
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = async () => {
    try {
      await fetchWithAuth('/api/hypervisors', {
        method: 'POST',
        body: JSON.stringify({ ...newHypervisor, status: 'offline' })
      });
      setOpenAdd(false);
      setNewHypervisor({ name: '', type: 'hyperv', host: '', port: 0, username: '', password: '' });
      fetchHypervisors();
    } catch (e) {
      setError('Failed to add hypervisor');
    }
  };

  const handleDelete = async (hypervisorId) => {
    if (!window.confirm('Delete this hypervisor? VMs will remain in inventory.')) return;
    try {
      await fetchWithAuth(`/api/hypervisors/${hypervisorId}`, { method: 'DELETE' });
      fetchHypervisors();
    } catch (e) {
      setError('Failed to delete hypervisor');
    }
  };

  const handleTest = async (hypervisorId) => {
    setTesting(hypervisorId);
    try {
      const response = await fetchWithAuth(`/api/hypervisors/${hypervisorId}/test`, { method: 'POST' });
      const data = await response.json();
      setSuccess(data.message || 'Connection test completed');
      setTimeout(() => setSuccess(''), 3000);
      fetchHypervisors();
    } catch (e) {
      setError('Connection test failed');
    } finally {
      setTesting(null);
    }
  };

  const handleRefresh = async (hypervisorId) => {
    try {
      await fetchWithAuth(`/api/hypervisors/${hypervisorId}/refresh`, { method: 'POST' });
      setSuccess('VM discovery completed');
      setTimeout(() => setSuccess(''), 3000);
      fetchHypervisors();
    } catch (e) {
      setError('VM discovery failed');
    }
  };

  const paginated = hypervisors.slice(page * rowsPerPage, (page + 1) * rowsPerPage);

  return (
    <Box p={3}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">Hypervisor Infrastructure</Typography>
        <Box display="flex" gap={1}>
          <Button variant="outlined" startIcon={<RefreshIcon />} onClick={fetchHypervisors}>Refresh</Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpenAdd(true)}>Add Hypervisor</Button>
        </Box>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess('')}>{success}</Alert>}

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Status</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Host / IP</TableCell>
              <TableCell>Port</TableCell>
              <TableCell>VMs</TableCell>
              <TableCell>Last Connected</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={8} align="center"><LinearProgress /></TableCell></TableRow>
            ) : paginated.length === 0 ? (
              <TableRow><TableCell colSpan={8} align="center">No hypervisors configured</TableCell></TableRow>
            ) : paginated.map((h) => {
              const sc = statusConfig[h.status] || statusConfig.offline;
              return (
                <TableRow key={h.hypervisorId} hover>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      {sc.icon}
                      <Chip label={sc.label} size="small" color={sc.color} variant="outlined" />
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      <Avatar sx={{ width: 32, height: 32, bgcolor: (typeColors[h.type] || '#666') + '30' }}>
                        <DnsIcon fontSize="small" sx={{ color: typeColors[h.type] || '#666' }} />
                      </Avatar>
                      <Typography variant="body2" fontWeight="medium">{h.name}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip label={h.type?.toUpperCase()} size="small"
                      sx={{ bgcolor: (typeColors[h.type] || '#666') + '20', color: typeColors[h.type] || '#666', fontWeight: 'bold' }} />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" fontFamily="monospace">{h.host}</Typography>
                  </TableCell>
                  <TableCell>{h.port || '—'}</TableCell>
                  <TableCell>
                    <Typography variant="body2" fontWeight="medium">{h.vmCount || 0}</Typography>
                  </TableCell>
                  <TableCell>
                    {h.lastConnectedAt ? (
                      <Typography variant="body2" color="text.secondary">
                        {new Date(h.lastConnectedAt).toLocaleString()}
                      </Typography>
                    ) : (
                      <Typography variant="body2" color="text.secondary">Never</Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="Test Connection">
                      <IconButton size="small" onClick={() => handleTest(h.hypervisorId)} disabled={testing === h.hypervisorId}>
                        {testing === h.hypervisorId ? <RefreshIcon fontSize="small" /> : <VpnKeyIcon fontSize="small" />}
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Discover VMs">
                      <IconButton size="small" onClick={() => handleRefresh(h.hypervisorId)}>
                        <SyncIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" onClick={() => handleDelete(h.hypervisorId)}><DeleteIcon fontSize="small" /></IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        <TablePagination
          component="div" count={hypervisors.length} page={page}
          onPageChange={(e, p) => setPage(p)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(e) => { setRowsPerPage(Number(e.target.value)); setPage(0); }}
        />
      </TableContainer>

      <Dialog open={openAdd} onClose={() => setOpenAdd(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Hypervisor</DialogTitle>
        <DialogContent>
          <TextField fullWidth label="Name" value={newHypervisor.name} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, name: e.target.value })} placeholder="e.g. Hyper-V Host 1" />
          <TextField fullWidth select label="Type" value={newHypervisor.type} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, type: e.target.value })}>
            <MenuItem value="hyperv">Hyper-V (Windows)</MenuItem>
            <MenuItem value="vmware">VMware vSphere</MenuItem>
            <MenuItem value="kvm">KVM / libvirt</MenuItem>
          </TextField>
          <TextField fullWidth label="Host / IP Address" value={newHypervisor.host} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, host: e.target.value })} placeholder="e.g. 192.168.1.100 or hyperv-host" />
          <TextField fullWidth label="Port (optional)" type="number" value={newHypervisor.port || ''} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, port: parseInt(e.target.value) || 0 })} />
          <TextField fullWidth label="Username" value={newHypervisor.username} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, username: e.target.value })} placeholder="e.g. DOMAIN\\admin" />
          <TextField fullWidth label="Password" type="password" value={newHypervisor.password} margin="normal"
            onChange={(e) => setNewHypervisor({ ...newHypervisor, password: e.target.value })} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAdd(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleAdd}>Add</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
