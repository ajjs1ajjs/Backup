import React, { useState } from 'react';
import { Box, Card, CardContent, TextField, Button, Typography, Alert } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [showChangePassword, setShowChangePassword] = useState(false);
  const [error, setError] = useState('');
  const { login, changePasswordFirstLogin, authError } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (username && password) {
      const success = await login(username, password);
      if (success) {
        navigate('/dashboard');
      } else if (authError === 'PASSWORD_CHANGE_REQUIRED') {
        setShowChangePassword(true);
        setError('Потрібно змінити пароль при першому вході');
      } else {
        setError('Invalid credentials');
      }
    } else {
      setError('Please enter username and password');
    }
  };

  const handleFirstLoginPasswordChange = async (e) => {
    e.preventDefault();
    try {
      await changePasswordFirstLogin(username, password, newPassword);
      navigate('/dashboard');
    } catch (changeError) {
      setError(changeError.message);
    }
  };

  return (
    <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh" bgcolor="#f5f5f5">
      <Card sx={{ maxWidth: 400, width: '100%' }}>
        <CardContent>
          <Typography variant="h4" align="center" gutterBottom>Backup System</Typography>
          <Typography variant="body2" align="center" color="text.secondary" mb={3}>Sign in to continue</Typography>
          
          {error && <Alert severity="error" mb={2}>{error}</Alert>}
          
          <form onSubmit={showChangePassword ? handleFirstLoginPasswordChange : handleSubmit}>
            <TextField
              fullWidth
              label="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
            />
            <TextField
              fullWidth
              label={showChangePassword ? 'Current Password' : 'Password'}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
            />
            {showChangePassword && (
              <TextField
                fullWidth
                label="New Password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                margin="normal"
              />
            )}
            <Button fullWidth variant="contained" type="submit" size="large" sx={{ mt: 2 }}>
              {showChangePassword ? 'Change Password' : 'Sign In'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </Box>
  );
}
