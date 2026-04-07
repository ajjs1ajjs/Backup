import React, { useEffect, useState } from 'react';
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  LinearProgress,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  CheckCircle as CheckCircleIcon,
  Delete as DeleteIcon,
  Dns as DnsIcon,
  Error as ErrorIcon,
  Refresh as RefreshIcon,
  Sync as SyncIcon,
  VpnKey as VpnKeyIcon,
  WifiOff as WifiOffIcon
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

const initialHypervisor = {
  name: '',
  type: 'hyperv',
  host: '',
  port: 0,
  username: '',
  password: ''
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
  const [newHypervisor, setNewHypervisor] = useState(initialHypervisor);

  useEffect(() => {
    fetchHypervisors();
  }, []);

  const fetchHypervisors = async () => {
    setLoading(true);
    try {
      const response = await fetchWithAuth('/api/hypervisors');
      const data = await response.json().catch(() => []);
      setHypervisors(Array.isArray(data) ? data : []);
    } catch (requestError) {
      console.error(requestError);
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
      setNewHypervisor(initialHypervisor);
      fetchHypervisors();
    } catch (requestError) {
      console.error(requestError);
      setError('Failed to add hypervisor');
    }
  };

  const handleDelete = async (hypervisorId) => {
    if (!window.confirm('Delete this hypervisor? VMs will remain in inventory.')) return;
    try {
      await fetchWithAuth(`/api/hypervisors/${hypervisorId}`, { method: 'DELETE' });
      fetchHypervisors();
    } catch (requestError) {
      console.error(requestError);
      setError('Failed to delete hypervisor');
    }
  };

  const handleTest = async (hypervisorId) => {
    setTesting(hypervisorId);
    try {
      const response = await fetchWithAuth(`/api/hypervisors/${hypervisorId}/test`, { method: 'POST' });
      const data = await response.json().catch(() => ({}));
      setSuccess(data.message || 'Connection test completed');
      setTimeout(() => setSuccess(''), 3000);
      fetchHypervisors();
    } catch (requestError) {
      console.error(requestError);
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
    } catch (requestError) {
      console.error(requestError);
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
              <TableRow>
                <TableCell colSpan={8} align="center"><LinearProgress /></TableCell>
              </TableRow>
            ) : paginated.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} align="center">No hypervisors configured</TableCell>
              </TableRow>
            ) : paginated.map((hypervisor) => {
              const status = statusConfig[hypervisor.status] || statusConfig.offline;
              return (
                <TableRow key={hypervisor.hypervisorId} hover>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      {status.icon}
                      <Chip label={status.label} size="small" color={status.color} variant="outlined" />
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      <Avatar sx={{ width: 32, height: 32, bgcolor: `${typeColors[hypervisor.type] || '#666'}30` }}>
                        <DnsIcon fontSize="small" sx={{ color: typeColors[hypervisor.type] || '#666' }} />
                      </Avatar>
                      <Typography variant="body2" fontWeight="medium">{hypervisor.name}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={hypervisor.type?.toUpperCase()}
                      size="small"
                      sx={{
                        bgcolor: `${typeColors[hypervisor.type] || '#666'}20`,
                        color: typeColors[hypervisor.type] || '#666',
                        fontWeight: 'bold'
                      }}
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" fontFamily="monospace">{hypervisor.host}</Typography>
                  </TableCell>
                  <TableCell>{hypervisor.port || '-'}</TableCell>
                  <TableCell>{hypervisor.vmCount || 0}</TableCell>
                  <TableCell>
                    {hypervisor.lastConnectedAt ? (
                      <Typography variant="body2" color="text.secondary">
                        {new Date(hypervisor.lastConnectedAt).toLocaleString()}
                      </Typography>
                    ) : (
                      <Typography variant="body2" color="text.secondary">Never</Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="Test Connection">
                      <IconButton size="small" onClick={() => handleTest(hypervisor.hypervisorId)} disabled={testing === hypervisor.hypervisorId}>
                        {testing === hypervisor.hypervisorId ? <RefreshIcon fontSize="small" /> : <VpnKeyIcon fontSize="small" />}
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Discover VMs">
                      <IconButton size="small" onClick={() => handleRefresh(hypervisor.hypervisorId)}>
                        <SyncIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" onClick={() => handleDelete(hypervisor.hypervisorId)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        <TablePagination
          component="div"
          count={hypervisors.length}
          page={page}
          onPageChange={(event, newPage) => setPage(newPage)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(event) => {
            setRowsPerPage(Number(event.target.value));
            setPage(0);
          }}
        />
      </TableContainer>

      <Dialog open={openAdd} onClose={() => setOpenAdd(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Hypervisor</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Name"
            value={newHypervisor.name}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, name: event.target.value })}
            placeholder="e.g. Hyper-V Host 1"
          />
          <TextField
            fullWidth
            select
            label="Type"
            value={newHypervisor.type}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, type: event.target.value })}
          >
            <MenuItem value="hyperv">Hyper-V (Windows)</MenuItem>
            <MenuItem value="vmware">VMware vSphere</MenuItem>
            <MenuItem value="kvm">KVM / libvirt</MenuItem>
          </TextField>
          <TextField
            fullWidth
            label="Host / IP Address"
            value={newHypervisor.host}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, host: event.target.value })}
            placeholder="e.g. 192.168.1.100 or hyperv-host"
          />
          <TextField
            fullWidth
            label="Port (optional)"
            type="number"
            value={newHypervisor.port || ''}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, port: parseInt(event.target.value, 10) || 0 })}
          />
          <TextField
            fullWidth
            label="Username"
            value={newHypervisor.username}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, username: event.target.value })}
            placeholder="e.g. DOMAIN\\admin"
          />
          <TextField
            fullWidth
            label="Password"
            type="password"
            value={newHypervisor.password}
            margin="normal"
            onChange={(event) => setNewHypervisor({ ...newHypervisor, password: event.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAdd(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleAdd}>Add</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
