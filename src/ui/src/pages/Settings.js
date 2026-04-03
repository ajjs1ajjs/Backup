import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, TextField, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Switch, FormControlLabel, Divider } from '@mui/material';
import { Save as SaveIcon } from '@mui/icons-material';
import { useApi, useApiMutation } from '../services/ApiContext';

export default function Settings() {
  const { data: settings } = useApi('/api/settings');
  const [localSettings, setLocalSettings] = useState({});

  React.useEffect(() => {
    if (settings) {
      const obj = {};
      settings.forEach((s) => { obj[s.key] = s.value; });
      setLocalSettings(obj);
    }
  }, [settings]);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Settings</Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Backup Settings</Typography>
          <Box display="flex" flexDirection="column" gap={2}>
            <TextField 
              label="Default Compression" 
              select 
              fullWidth 
              SelectProps={{ native: true }}
              value={localSettings['backup.compression'] || 'zstd'}
              onChange={(e) => setLocalSettings({...localSettings, 'backup.compression': e.target.value})}
            >
              <option value="zstd">Zstd (Recommended)</option>
              <option value="lz4">LZ4 (Fast)</option>
              <option value="gzip">Gzip</option>
              <option value="none">None</option>
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
              control={<Switch defaultChecked />} 
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
              SelectProps={{ native: true }}
            >
              <option value="aes256">AES-256</option>
              <option value="aes128">AES-128</option>
            </TextField>
            <FormControlLabel 
              control={<Switch defaultChecked />} 
              label="Require Two-Factor Authentication" 
            />
          </Box>
        </CardContent>
      </Card>

      <Button variant="contained" startIcon={<SaveIcon />}>
        Save Settings
      </Button>
    </Box>
  );
}
