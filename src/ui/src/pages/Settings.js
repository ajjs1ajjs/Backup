import React, { useState, useEffect } from 'react';
import { Box, Card, CardContent, Typography, TextField, Button, Select, MenuItem, Switch, FormControlLabel, CircularProgress, Alert } from '@mui/material';
import { Save as SaveIcon } from '@mui/icons-material';
import { useApi, fetchWithAuth } from '../services/ApiContext';

export default function Settings() {
  const { data: settings, loading, refetch } = useApi('/api/settings');
  const [localSettings, setLocalSettings] = useState({});
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (settings) {
      const obj = {};
      settings.forEach((s) => { obj[s.key] = s.value; });
      setLocalSettings(obj);
    }
  }, [settings]);

  const handleSave = async () => {
    try {
      for (const [key, value] of Object.entries(localSettings)) {
        await fetchWithAuth('/api/settings', {
          method: 'PUT',
          body: JSON.stringify({ key, value, type: 'string' })
        });
      }
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
      refetch();
    } catch (e) { console.error(e); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Settings</Typography>

      {saved && <Alert severity="success" sx={{ mb: 2 }}>Settings saved successfully</Alert>}

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Backup Settings</Typography>
          <Box display="flex" flexDirection="column" gap={2}>
            <TextField
              label="Default Compression"
              select
              fullWidth
              value={localSettings['backup.compression'] || 'zstd'}
              onChange={(e) => setLocalSettings({...localSettings, 'backup.compression': e.target.value})}
            >
              <MenuItem value="zstd">Zstd (Recommended)</MenuItem>
              <MenuItem value="lz4">LZ4 (Fast)</MenuItem>
              <MenuItem value="gzip">Gzip</MenuItem>
              <MenuItem value="none">None</MenuItem>
            </TextField>
            <TextField
              label="Default Retention Days"
              type="number"
              fullWidth
              value={localSettings['backup.retention_days'] || '30'}
              onChange={(e) => setLocalSettings({...localSettings, 'backup.retention_days': e.target.value})}
            />
            <TextField
              label="Block Size (KB)"
              type="number"
              fullWidth
              value={localSettings['backup.block_size_kb'] || '64'}
              onChange={(e) => setLocalSettings({...localSettings, 'backup.block_size_kb': e.target.value})}
            />
          </Box>
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Network Settings</Typography>
          <Box display="flex" flexDirection="column" gap={2}>
            <TextField
              label="Server Port"
              type="number"
              fullWidth
              value={localSettings['network.port'] || '8000'}
              onChange={(e) => setLocalSettings({...localSettings, 'network.port': e.target.value})}
            />
            <FormControlLabel
              control={<Switch checked={localSettings['network.tls'] === 'true'} onChange={(e) => setLocalSettings({...localSettings, 'network.tls': e.target.checked.toString()})} />}
              label="Enable TLS"
            />
          </Box>
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Security</Typography>
          <Box display="flex" flexDirection="column" gap={2}>
            <TextField
              label="Encryption Algorithm"
              select
              fullWidth
              value={localSettings['security.encryption'] || 'aes256'}
              onChange={(e) => setLocalSettings({...localSettings, 'security.encryption': e.target.value})}
            >
              <MenuItem value="aes256">AES-256</MenuItem>
              <MenuItem value="aes128">AES-128</MenuItem>
            </TextField>
            <FormControlLabel
              control={<Switch checked={localSettings['security.2fa'] === 'true'} onChange={(e) => setLocalSettings({...localSettings, 'security.2fa': e.target.checked.toString()})} />}
              label="Require Two-Factor Authentication"
            />
          </Box>
        </CardContent>
      </Card>

      <Button variant="contained" startIcon={<SaveIcon />} onClick={handleSave}>
        Save Settings
      </Button>
    </Box>
  );
}
