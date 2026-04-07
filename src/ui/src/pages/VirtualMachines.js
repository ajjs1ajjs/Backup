import React, { useState, useEffect } from 'react';
import {
  Box, Card, CardContent, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, TablePagination, Chip, IconButton, Typography,
  Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField,
  MenuItem, Alert, Tooltip, Tabs, Tab, Avatar, CircularProgress
} from '@mui/material';
import {
  Delete as DeleteIcon, Refresh as RefreshIcon, Add as AddIcon,
  Computer as ComputerIcon, PlayArrow as PlayIcon, Stop as StopIcon,
  CheckCircle as CheckCircleIcon, Error as ErrorIcon, HourglassEmpty as HourglassIcon
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
  const [newVM, setNewVM] = useState({
    name: '', hypervisorType: 'hyperv', hypervisorHost: '',
    ipAddress: '', osType: 'windows', cpuCores: 2, memoryMb: 4096
  });

  useEffect(() => {
    fetchVMs();
    fetchHypervisors();
  }, [tabValue]);

  const fetchVMs = async () => {
    setLoading(true);
    try {
      const filter = tabValue === 0 ? '' :
        tabValue === 1 ? '?hypervisorType=hyperv' :
        tabValue === 2 ? '?hypervisorType=vmware' :
        '?hypervisorType=kvm';
      const response = await fetchWithAuth(`/api/virtualmachines${filter}`);
      const data = await response.json();
      setVMs(data);
    } catch (e) {
      setError('Failed to load VMs');
    } finally {
      setLoading(false);
    }
  };

  const fetchHypervisors = async () => {
    try {
      const response = await fetchWithAuth('/api/hypervisors');
      const data = await response.json();
      setHypervisors(data);
    } catch (e) { /* ignore */ }
  };

  const handleDeleteVM = async (vmId) => {
    if (!window.confirm('Delete this VM from inventory? (backup data will not be deleted)')) return;
    try {
      await fetchWithAuth(`/api/virtualmachines/${vmId}`, { method: 'DELETE' });
      fetchVMs();
    } catch (e) {
      setError('Failed to delete VM');
    }
  };

  const handleAddVM = async (e) => {
    e.preventDefault();
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
      setNewVM({ name: '', hypervisorType: 'hyperv', hypervisorHost: '', ipAddress: '', osType: 'windows', cpuCores: 2, memoryMb: 4096 });
      fetchVMs();
    } catch (e) {
      setError('Failed to add VM');
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
    } catch (e) {
      setError('Discovery failed');
    } finally {
      setDiscovering(false);
      setOpenDiscover(false);
    }
  };

  const formatBytes = (bytes) => {
    if (!bytes) return '—';
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  };

  const filteredVMs = tabValue === 0 ? vms :
    tabValue === 1 ? vms.filter(v => v.hypervisorType === 'hyperv') :
    tabValue === 2 ? vms.filter(v => v.hypervisorType === 'vmware') :
    vms.filter(v => v.hypervisorType === 'kvm');

  const paginatedVMs = filteredVMs.slice(page * rowsPerPage, (page + 1) * rowsPerPage);

  return (
    <Box p={3}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">Віртуальні машини</Typography>
        <Box display="flex" gap={1}>
          <Button variant="outlined" startIcon={<RefreshIcon />} onClick={fetchVMs}>Оновити</Button>
          <Button variant="outlined" startIcon={<ComputerIcon />} onClick={() => setOpenDiscover(true)}
            disabled={hypervisors.length === 0}>Виявити ВМ</Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpenAdd(true)}>Додати ВМ</Button>
        </Box>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess('')}>{success}</Alert>}

      <Tabs value={tabValue} onChange={(e, v) => { setTabValue(v); setPage(0); }} sx={{ mb: 2 }}>
        <Tab label={`Всі (${vms.length})`} />
        <Tab label={`Hyper-V (${vms.filter(v => v.hypervisorType === 'hyperv').length})`} />
        <Tab label={`VMware (${vms.filter(v => v.hypervisorType === 'vmware').length})`} />
        <Tab label={`KVM (${vms.filter(v => v.hypervisorType === 'kvm').length})`} />
      </Tabs>

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Статус</TableCell>
              <TableCell>Назва</TableCell>
              <TableCell>Гіпервізор</TableCell>
              <TableCell>Хост</TableCell>
              <TableCell>IP-адреса</TableCell>
              <TableCell>ОС</TableCell>
              <TableCell>CPU</TableCell>
              <TableCell>RAM</TableCell>
              <TableCell>Останній бекап</TableCell>
              <TableCell>Дії</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={10} align="center"><CircularProgress size={30} sx={{ my: 2 }} /></TableCell></TableRow>
            ) : paginatedVMs.length === 0 ? (
              <TableRow><TableCell colSpan={10} align="center">ВМ не знайдено</TableCell></TableRow>
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
                  <Chip label={vm.hypervisorType?.toUpperCase()} size="small"
                    sx={{ bgcolor: (hypervisorTypeColors[vm.hypervisorType] || '#666') + '20', color: hypervisorTypeColors[vm.hypervisorType] || '#666', fontWeight: 'bold' }} />
                </TableCell>
                <TableCell>{vm.hypervisorHost || '—'}</TableCell>
                <TableCell>{vm.ipAddress || '—'}</TableCell>
                <TableCell>{vm.osType || '—'}</TableCell>
                <TableCell>{vm.cpuCores || '—'} ядер</TableCell>
                <TableCell>{vm.memoryMb ? `${(vm.memoryMb / 1024).toFixed(1)} ГБ` : '—'}</TableCell>
                <TableCell>
                  {vm.lastBackupAt ? (
                    <Typography variant="body2">{new Date(vm.lastBackupAt).toLocaleString()}</Typography>
                  ) : (
                    <Chip label="Немає бекапів" size="small" color="warning" />
                  )}
                </TableCell>
                <TableCell>
                  <Tooltip title="Видалити з інвентарю">
                    <IconButton size="small" onClick={() => handleDeleteVM(vm.vmId)}><DeleteIcon fontSize="small" /></IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination
          component="div" count={filteredVMs.length} page={page}
          onPageChange={(e, p) => setPage(p)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(e) => { setRowsPerPage(Number(e.target.value)); setPage(0); }}
        />
      </TableContainer>

      <Dialog open={openAdd} onClose={() => setOpenAdd(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Додати віртуальну машину</DialogTitle>
        <DialogContent>
          <form onSubmit={handleAddVM}>
            <TextField fullWidth label="Назва ВМ" value={newVM.name} margin="normal"
              onChange={(e) => setNewVM({ ...newVM, name: e.target.value })} required />
            <TextField fullWidth select label="Тип гіпервізора" value={newVM.hypervisorType} margin="normal"
              onChange={(e) => setNewVM({ ...newVM, hypervisorType: e.target.value })}>
              <MenuItem value="hyperv">Hyper-V</MenuItem>
              <MenuItem value="vmware">VMware</MenuItem>
              <MenuItem value="kvm">KVM</MenuItem>
            </TextField>
            <TextField fullWidth label="Хост гіпервізора (IP або ім'я)" value={newVM.hypervisorHost} margin="normal"
              onChange={(e) => setNewVM({ ...newVM, hypervisorHost: e.target.value })} required />
            <TextField fullWidth label="IP-адреса" value={newVM.ipAddress} margin="normal"
              onChange={(e) => setNewVM({ ...newVM, ipAddress: e.target.value })} />
            <TextField fullWidth select label="Тип ОС" value={newVM.osType} margin="normal"
              onChange={(e) => setNewVM({ ...newVM, osType: e.target.value })}>
              <MenuItem value="windows">Windows</MenuItem>
              <MenuItem value="linux">Linux</MenuItem>
              <MenuItem value="other">Інша</MenuItem>
            </TextField>
            <Box display="flex" gap={2}>
              <TextField fullWidth label="Ядра CPU" type="number" value={newVM.cpuCores} margin="normal"
                onChange={(e) => setNewVM({ ...newVM, cpuCores: parseInt(e.target.value) })} />
              <TextField fullWidth label="RAM (МБ)" type="number" value={newVM.memoryMb} margin="normal"
                onChange={(e) => setNewVM({ ...newVM, memoryMb: parseInt(e.target.value) })} />
            </Box>
          </form>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAdd(false)}>Скасувати</Button>
          <Button variant="contained" onClick={handleAddVM}>Додати ВМ</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openDiscover} onClose={() => setOpenDiscover(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Виявити віртуальні машини</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" mb={2}>
            Виберіть гіпервізор, щоб знайти всі зареєстровані на ньому ВМ.
          </Typography>
          <TextField fullWidth select label="Гіпервізор" value={selectedHypervisor} margin="normal"
            onChange={(e) => setSelectedHypervisor(e.target.value)}>
            {hypervisors.map(h => (
              <MenuItem key={h.hypervisorId} value={h.hypervisorId}>
                {h.name} ({h.type?.toUpperCase()}) — {h.host}
              </MenuItem>
            ))}
          </TextField>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDiscover(false)}>Скасувати</Button>
          <Button variant="contained" onClick={handleDiscover} disabled={discovering}>
            {discovering ? 'Виявлення...' : 'Виявити'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
