import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Button, TextField, Select, MenuItem, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, LinearProgress } from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon, CheckCircle as TestIcon } from '@mui/icons-material';
import { useApi } from '../services/ApiContext';

export default function Repositories() {
  const { data, loading, refetch } = useApi('/api/repositories');
  const [open, setOpen] = useState(false);

  const handleTest = async (repoId: string) => {
    await fetch(`/api/repositories/${repoId}/test`, { method: 'POST' });
  };

  if (loading) return <Typography>Loading...</Typography>;

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Repositories</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
          Add Repository
        </Button>
      </Box>

      <TableContainer component={Card}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Path</TableCell>
              <TableCell>Capacity</TableCell>
              <TableCell>Used</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.repositories?.map((repo: any) => (
              <TableRow key={repo.repositoryId}>
                <TableCell>{repo.name}</TableCell>
                <TableCell><Chip label={repo.type} size="small" /></TableCell>
                <TableCell sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{repo.path}</TableCell>
                <TableCell>{(repo.capacityBytes / 1024 / 1024 / 1024).toFixed(1)} GB</TableCell>
                <TableCell>
                  <Box display="flex" alignItems="center" gap={1}>
                    <LinearProgress variant="determinate" value={(repo.usedBytes / repo.capacityBytes) * 100} sx={{ width: 60 }} />
                    <Typography variant="caption">{((repo.usedBytes / repo.capacityBytes) * 100).toFixed(0)}%</Typography>
                  </Box>
                </TableCell>
                <TableCell><Chip label={repo.status} color={repo.status === 'online' ? 'success' : 'error'} size="small" /></TableCell>
                <TableCell>
                  <IconButton onClick={() => handleTest(repo.repositoryId)}><TestIcon /></IconButton>
                  <IconButton color="error"><DeleteIcon /></IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Repository</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} pt={2}>
            <TextField label="Name" fullWidth />
            <Select fullWidth defaultValue="local">
              <MenuItem value="local">Local Disk</MenuItem>
              <MenuItem value="nfs">NFS</MenuItem>
              <MenuItem value="smb">SMB/CIFS</MenuItem>
              <MenuItem value="s3">Amazon S3</MenuItem>
              <MenuItem value="azure_blob">Azure Blob</MenuItem>
              <MenuItem value="gcs">Google Cloud Storage</MenuItem>
            </Select>
            <TextField label="Path" fullWidth />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={() => setOpen(false)}>Save</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
