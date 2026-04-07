import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  Chip,
  Alert,
  CircularProgress
} from '@mui/material';
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon, LockReset as ResetIcon } from '@mui/icons-material';
import { fetchWithAuth, useApi } from '../services/ApiContext';

const ROLES = [
  { value: 'admin', label: 'Administrator', color: 'error' },
  { value: 'operator', label: 'Operator', color: 'warning' },
  { value: 'Viewer', label: 'Viewer', color: 'info' }
];

export default function Users() {
  const { data: users, loading, refetch } = useApi('/api/users');
  const [open, setOpen] = useState(false);
  const [editMode, setEditMode] = useState(false  );
  const [selectedUser, setSelectedUser] = useState(null);
  const [form, setForm] = useState({ username: '', email: '', password: '', role: 'Viewer' });
  const [resetOpen, setResetOpen] = useState(false);
  const [resetPassword, setResetPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (!loading && !users) {
      refetch();
    }
  }, []);

  const handleOpen = (user = null) => {
    if (user) {
      setEditMode(true);
      setSelectedUser(user);
      setForm({ username: user.Username, email: user.Email, password: '', role: user.Role });
    } else {
      setEditMode(false);
      setSelectedUser(null);
      setForm({ username: '', email: '', password: '', role: 'Viewer' });
    }
    setError('');
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setSelectedUser(null);
    setForm({ username: '', email: '', password: '', role: 'Viewer' });
  };

  const handleSave = async () => {
    try {
      setError('');
      if (!editMode && !form.password) {
        setError('Password is required');
        return;
      }

      if (editMode) {
        await fetchWithAuth(`/api/users/${selectedUser.Id}`, {
          method: 'PUT',
          body: JSON.stringify(form)
        });
        setSuccess('User updated successfully');
      } else {
        await fetchWithAuth('/api/users', {
          method: 'POST',
          body: JSON.stringify(form)
        });
        setSuccess('User created successfully');
      }

      handleClose();
      refetch();
      setTimeout(() => setSuccess(''), 3000);
    } catch (e) {
      setError(e.message || 'Failed to save user');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this user?')) return;
    try {
      await fetchWithAuth(`/api/users/${id}`, { method: 'DELETE' });
      setSuccess('User deleted successfully');
      refetch();
      setTimeout(() => setSuccess(''), 3000);
    } catch (e) {
      setError(e.message || 'Failed to delete user');
    }
  };

  const handleResetPassword = async () => {
    try {
      await fetchWithAuth(`/api/users/${selectedUser.Id}/reset-password`, {
        method: 'POST',
        body: JSON.stringify({ NewPassword: resetPassword })
      });
      setSuccess('Password reset successfully');
      setResetOpen(false);
      setResetPassword('');
      setTimeout(() => setSuccess(''), 3000);
    } catch (e) {
      setError(e.message || 'Failed to reset password');
    }
  };

  const openResetDialog = (user) => {
    setSelectedUser(user);
    setResetPassword('');
    setResetOpen(true);
  };

  if (loading) {
    return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Users</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpen()}>
          Add User
        </Button>
      </Box>

      {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Card>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Username</TableCell>
                <TableCell>Email</TableCell>
                <TableCell>Role</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Created</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {(users || []).map((user) => (
                <TableRow key={user.Id}>
                  <TableCell>{user.Username}</TableCell>
                  <TableCell>{user.Email}</TableCell>
                  <TableCell>
                    <Chip 
                      label={user.Role} 
                      color={ROLES.find(r => r.value === user.Role)?.color || 'default'} 
                      size="small" 
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={user.IsActive ? 'Active' : 'Inactive'} 
                      color={user.IsActive ? 'success' : 'default'} 
                      size="small" 
                    />
                  </TableCell>
                  <TableCell>{new Date(user.CreatedAt).toLocaleDateString()}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => openResetDialog(user)} title="Reset Password">
                      <ResetIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleOpen(user)} title="Edit">
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton 
                      size="small" 
                      onClick={() => handleDelete(user.Id)} 
                      title="Delete"
                      disabled={user.Username === 'admin'}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Card>

      <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
        <DialogTitle>{editMode ? 'Edit User' : 'Add User'}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} sx={{ pt: 1 }}>
            <TextField
              label="Username"
              fullWidth
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              disabled={editMode && form.username === 'admin'}
            />
            <TextField
              label="Email"
              fullWidth
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
            />
            <TextField
              label={editMode ? 'New Password (leave empty to keep current)' : 'Password'}
              fullWidth
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
            />
            <TextField
              label="Role"
              fullWidth
              select
              value={form.role}
              onChange={(e) => setForm({ ...form, role: e.target.value })}
            >
              {ROLES.map((role) => (
                <MenuItem key={role.value} value={role.value}>{role.label}</MenuItem>
              ))}
            </TextField>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            {editMode ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={resetOpen} onClose={() => setResetOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Reset Password</DialogTitle>
        <DialogContent>
          <TextField
            label="New Password"
            fullWidth
            type="password"
            value={resetPassword}
            onChange={(e) => setResetPassword(e.target.value)}
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResetOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleResetPassword}>Reset</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}