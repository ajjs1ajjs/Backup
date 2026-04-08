import React, { useState } from 'react';
import { Box, Button, TextField, Typography, Card, CardContent, Alert } from '@mui/material';
import { fetchWithAuth } from '../services/ApiContext';

export default function TwoFactorSetup({ onEnabled }) {
  const [secret, setSecret] = useState(null);
  const [code, setCode] = useState('');
  const [error, setError] = useState(null);

  const setup2FA = async () => {
    try {
      const res = await fetchWithAuth('/api/auth/2fa/setup', { method: 'POST' });
      const data = await res.json();
      setSecret(data.secret);
    } catch (e) {
      setError('Failed to initialize 2FA');
    }
  };

  const verify2FA = async () => {
    try {
      await fetchWithAuth('/api/auth/2fa/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code })
      });
      onEnabled();
    } catch (e) {
      setError('Invalid code');
    }
  };

  if (!secret) {
    return <Button variant="contained" onClick={setup2FA}>Enable 2FA</Button>;
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6">Setup Two-Factor Authentication</Typography>
        <Typography variant="body2" sx={{ my: 2 }}>
          Scan this secret with your authenticator app: <strong>{secret}</strong>
        </Typography>
        <TextField 
          label="Verification Code" 
          value={code} 
          onChange={(e) => setCode(e.target.value)} 
          fullWidth 
          sx={{ mb: 2 }}
        />
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        <Button variant="contained" onClick={verify2FA}>Verify & Enable</Button>
      </CardContent>
    </Card>
  );
}
