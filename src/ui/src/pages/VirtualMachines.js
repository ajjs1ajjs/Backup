import React, { useEffect, useState } from 'react';
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  Tabs,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import {
  Add as AddIcon,
  CheckCircle as CheckCircleIcon,
  Computer as ComputerIcon,
  Delete as DeleteIcon,
  Error as ErrorIcon,
  HourglassEmpty as HourglassIcon,
  Refresh as RefreshIcon,
  Stop as StopIcon
} from '@mui/icons-material';
import { fetchWithAuth } from '../services/ApiContext';

const hypervisorTypeColors = {
  hyperv: '#0078D4',
  vmware: '#607D8B',
  kvm: '#FF9800'
};

const statusIcons = {
  running: <CheckCircleIcon fontSize="small" sx={{ color: '#4CAF50' }} />,
  stopped: <StopIcon fontSize="small" sx={{ color: '#9E9E9E' }} />,
  paused: <HourglassIcon fontSize="small" sx={{ color: '#FF9800' }} />,
  error: <ErrorIcon fontSize="small" sx={{ color: '#F44336' }} />
};

const initialVm = {
  name: '',
  hypervisorType: 'hyperv',
  hypervisorHost: '',
  ipAddress: '',
  osType: 'windows',
  cpuCores: 2,
  memoryMb: 4096
};

export default function VirtualMachines() {
  const [vms, setVMs] = useState([]);
  const [hypervisors, setHypervisors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(15);
  const [tabValue, setTabValue] = useState(0);
  const [openAdd, setOpenAdd] = useState(false);
  const [openDiscover, setOpenDiscover] = useState(false);
  const [selectedHypervisor, setSelectedHypervisor] = useState('');
  const [discovering, setDiscovering] = useState(false);
  const [newVM, setNewVM] = useState(initialVm);

  useEffect(() => {
    fetchVMs();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tabValue]);

  useEffect(() => {
    fetchHypervisors();
  }, []);

  const fetchVMs = async () => {
    setLoading(true);
    try {
      const filter = tabValue === 0
        ? ''
        : tabValue === 1
          ? '?hypervisorType=hyperv'
          : tabValue === 2
            ? '?hypervisorType=vmware'
            : '?hypervisorType=kvm';

      const response = await fetchWithAuth(`/api/virtualmachines${filter}`);
      const data = await response.json().catch(() => []);
      setVMs(Array.isArray(data) ? data : []);
    } catch (requestError) {
      console.error(requestError);
      setError('Failed to load virtual machines');
    } finally {
      setLoading(false);
    }
  };

  const fetchHypervisors = async () => {
    try {
      const response = await fetchWithAuth('/api/hypervisors');
      const data = await response.json().catch(() => []);
      setHypervisors(Array.isArray(data) ? data : []);
    } catch (requestError) {
      console.error(requestError);
    }
  };

  const handleDeleteVM = async (vmId) => {
    if (!window.confirm('Delete this VM from inventory? Backup data will not be deleted.')) return;
    try {
      await fetchWithAuth(`/api/virtualmachines/${vmId}`, { method: 'DELETE' });
      fetchVMs();
    } catch (requestError) {
      console.error(requestError);
      setError('Failed to delete virtual machine');
    }
  };

  const handleAddVM = async () => {
    try {
      await fetchWithAuth('/api/virtualmachines', {
        method: 'POST',
        body: JSON.stringify({
          ...newVM,
          status: 'running',
          disks: JSON.stringify([]),
          tags: JSON.stringify({})
        })
      });
      setOpenAdd(false);
      setNewVM(initialVm);
      fetchVMs();
    } catch (requestError) {
      console.error(requestError);
      setError('Failed to add virtual machine');
    }
  };

  const handleDiscover = async () => {
    if (!selectedHypervisor) {
      setError('Select a hypervisor first');
      return;
    }

    setDiscovering(true);
    setError('');
    try {
      await fetchWithAuth(`/api/hypervisors/${selectedHypervisor}/refresh`, { method: 'POST' });
      setSuccess('VM list refreshed successfully');
      setTimeout(() => setSuccess(''), 3000);
      fetchVMs();
      fetchHypervisors();
    } catch (requestError) {
      console.error(requestError);
      setError('Discovery failed');
    } finally {
      setDiscovering(false);
      setOpenDiscover(false);
    }
  };

  const filteredVMs = tabValue === 0
    ? vms
    : tabValue === 1
      ? vms.filter((vm) => vm.hypervisorType === 'hyperv')
      : tabValue === 2
        ? vms.filter((vm) => vm.hypervisorType === 'vmware')
        : vms.filter((vm) => vm.hypervisorType === 'kvm');

  const paginatedVMs = filteredVMs.slice(page * rowsPerPage, (page + 1) * rowsPerPage);

  return (
    <Box p={3}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">Virtual Machines</Typography>
        <Box display="flex" gap={1}>
          <Button variant="outlined" startIcon={<RefreshIcon />} onClick={fetchVMs}>Refresh</Button>
          <Button
            variant="outlined"
            startIcon={<ComputerIcon />}
            onClick={() => setOpenDiscover(true)}
            disabled={hypervisors.length === 0}
          >
            Discover VMs
          </Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpenAdd(true)}>Add VM</Button>
        </Box>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess('')}>{success}</Alert>}

      <Tabs value={tabValue} onChange={(event, value) => { setTabValue(value); setPage(0); }} sx={{ mb: 2 }}>
        <Tab label={`All (${vms.length})`} />
        <Tab label={`Hyper-V (${vms.filter((vm) => vm.hypervisorType === 'hyperv').length})`} />
        <Tab label={`VMware (${vms.filter((vm) => vm.hypervisorType === 'vmware').length})`} />
        <Tab label={`KVM (${vms.filter((vm) => vm.hypervisorType === 'kvm').length})`} />
      </Tabs>

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Status</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Hypervisor</TableCell>
              <TableCell>Host</TableCell>
              <TableCell>IP Address</TableCell>
              <TableCell>OS</TableCell>
              <TableCell>CPU</TableCell>
              <TableCell>RAM</TableCell>
              <TableCell>Last Backup</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={10} align="center"><CircularProgress size={30} sx={{ my: 2 }} /></TableCell>
              </TableRow>
            ) : paginatedVMs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={10} align="center">No virtual machines found</TableCell>
              </TableRow>
            ) : paginatedVMs.map((vm) => (
              <TableRow key={vm.vmId} hover>
                <TableCell>{statusIcons[vm.status] || statusIcons.stopped}</TableCell>
                <TableCell>
                  <Box display="flex" alignItems="center" gap={1}>
                    <Avatar sx={{ width: 32, height: 32, bgcolor: hypervisorTypeColors[vm.hypervisorType] || '#666' }}>
                      <ComputerIcon fontSize="small" />
                    </Avatar>
                    <Box>
                      <Typography variant="body2" fontWeight="medium">{vm.name}</Typography>
                      <Typography variant="caption" color="text.secondary">{vm.vmId?.substring(0, 8)}</Typography>
                    </Box>
                  </Box>
                </TableCell>
                <TableCell>
                  <Chip
                    label={vm.hypervisorType?.toUpperCase()}
                    size="small"
                    sx={{
                      bgcolor: `${hypervisorTypeColors[vm.hypervisorType] || '#666'}20`,
                      color: hypervisorTypeColors[vm.hypervisorType] || '#666',
                      fontWeight: 'bold'
                    }}
                  />
                </TableCell>
                <TableCell>{vm.hypervisorHost || '-'}</TableCell>
                <TableCell>{vm.ipAddress || '-'}</TableCell>
                <TableCell>{vm.osType || '-'}</TableCell>
                <TableCell>{vm.cpuCores || '-'} cores</TableCell>
                <TableCell>{vm.memoryMb ? `${(vm.memoryMb / 1024).toFixed(1)} GB` : '-'}</TableCell>
                <TableCell>
                  {vm.lastBackupAt ? (
                    <Typography variant="body2">{new Date(vm.lastBackupAt).toLocaleString()}</Typography>
                  ) : (
                    <Chip label="No backups" size="small" color="warning" />
                  )}
                </TableCell>
                <TableCell>
                  <Tooltip title="Remove from inventory">
                    <IconButton size="small" onClick={() => handleDeleteVM(vm.vmId)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination
          component="div"
          count={filteredVMs.length}
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
        <DialogTitle>Add Virtual Machine</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="VM Name"
            value={newVM.name}
            margin="normal"
            onChange={(event) => setNewVM({ ...newVM, name: event.target.value })}
            required
          />
          <TextField
            fullWidth
            select
            label="Hypervisor Type"
            value={newVM.hypervisorType}
            margin="normal"
            onChange={(event) => setNewVM({ ...newVM, hypervisorType: event.target.value })}
          >
            <MenuItem value="hyperv">Hyper-V</MenuItem>
            <MenuItem value="vmware">VMware</MenuItem>
            <MenuItem value="kvm">KVM</MenuItem>
          </TextField>
          <TextField
            fullWidth
            label="Hypervisor Host"
            value={newVM.hypervisorHost}
            margin="normal"
            onChange={(event) => setNewVM({ ...newVM, hypervisorHost: event.target.value })}
            placeholder="IP address or hostname"
            required
          />
          <TextField
            fullWidth
            label="IP Address"
            value={newVM.ipAddress}
            margin="normal"
            onChange={(event) => setNewVM({ ...newVM, ipAddress: event.target.value })}
          />
          <TextField
            fullWidth
            select
            label="Operating System"
            value={newVM.osType}
            margin="normal"
            onChange={(event) => setNewVM({ ...newVM, osType: event.target.value })}
          >
            <MenuItem value="windows">Windows</MenuItem>
            <MenuItem value="linux">Linux</MenuItem>
            <MenuItem value="other">Other</MenuItem>
          </TextField>
          <Box display="flex" gap={2}>
            <TextField
              fullWidth
              label="CPU Cores"
              type="number"
              value={newVM.cpuCores}
              margin="normal"
              onChange={(event) => setNewVM({ ...newVM, cpuCores: parseInt(event.target.value, 10) || 0 })}
            />
            <TextField
              fullWidth
              label="RAM (MB)"
              type="number"
              value={newVM.memoryMb}
              margin="normal"
              onChange={(event) => setNewVM({ ...newVM, memoryMb: parseInt(event.target.value, 10) || 0 })}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAdd(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleAddVM}>Add VM</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openDiscover} onClose={() => setOpenDiscover(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Discover Virtual Machines</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" mb={2}>
            Select a hypervisor to refresh the virtual machine inventory from that host.
          </Typography>
          <TextField
            fullWidth
            select
            label="Hypervisor"
            value={selectedHypervisor}
            margin="normal"
            onChange={(event) => setSelectedHypervisor(event.target.value)}
          >
            {hypervisors.map((hypervisor) => (
              <MenuItem key={hypervisor.hypervisorId} value={hypervisor.hypervisorId}>
                {hypervisor.name} ({hypervisor.type?.toUpperCase()}) - {hypervisor.host}
              </MenuItem>
            ))}
          </TextField>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDiscover(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleDiscover} disabled={discovering}>
            {discovering ? 'Discovering...' : 'Discover'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
