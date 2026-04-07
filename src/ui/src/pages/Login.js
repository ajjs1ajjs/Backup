import React, { useState } from 'react';
import { Box, Card, CardContent, TextField, Button, Typography, Alert, InputAdornment, IconButton } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { Backup as BackupIcon, Visibility, VisibilityOff, Person, Lock } from '@mui/icons-material';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showChangePassword, setShowChangePassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login, changePasswordFirstLogin, authError } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!username || !password) {
      setError('Please enter username and password.');
      return;
    }

    setError('');
    setLoading(true);
    const success = await login(username, password);
    setLoading(false);
    if (success) {
      navigate('/dashboard');
    }
  };

  React.useEffect(() => {
    if (authError === 'PASSWORD_CHANGE_REQUIRED') {
      setShowChangePassword(true);
      setError('You must change your password on first login.');
    } else if (authError) {
      setError(authError);
    }
  }, [authError]);

  const handleFirstLoginPasswordChange = async (e) => {
    e.preventDefault();

    if (!newPassword || newPassword.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }

    setError('');
    setLoading(true);
    try {
      await changePasswordFirstLogin(username, password, newPassword);
      navigate('/dashboard');
    } catch (changeError) {
      setError(changeError.message);
    }
    setLoading(false);
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #1a1d23 0%, #2d3748 50%, #1a1d23 100%)',
        position: 'relative',
        overflow: 'hidden'
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          top: '10%',
          left: '10%',
          width: 300,
          height: 300,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(79, 195, 247, 0.1) 0%, transparent 70%)',
          filter: 'blur(60px)'
        }}
      />
      <Box
        sx={{
          position: 'absolute',
          bottom: '10%',
          right: '10%',
          width: 250,
          height: 250,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(102, 187, 106, 0.08) 0%, transparent 70%)',
          filter: 'blur(50px)'
        }}
      />

      <Card
        sx={{
          maxWidth: 420,
          width: '100%',
          mx: 2,
          borderRadius: 3,
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
          background: 'rgba(255, 255, 255, 0.98)',
          backdropFilter: 'blur(20px)'
        }}
      >
        <CardContent sx={{ p: 4 }}>
          <Box display="flex" justifyContent="center" mb={3}>
            <Box
              sx={{
                width: 72,
                height: 72,
                borderRadius: 2,
                background: 'linear-gradient(135deg, #4fc3f7 0%, #29b6f6 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 8px 32px rgba(79, 195, 247, 0.4)'
              }}
            >
              <BackupIcon sx={{ fontSize: 40, color: '#fff' }} />
            </Box>
          </Box>

          <Typography variant="h4" align="center" sx={{ fontWeight: 700, color: '#1a1d23', mb: 1 }}>
            Backup System
          </Typography>
          <Typography variant="body2" align="center" color="text.secondary" sx={{ mb: 3 }}>
            {showChangePassword ? 'Create your new password' : 'Sign in to continue'}
          </Typography>

          {error && (
            <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>
              {error}
            </Alert>
          )}

          <form onSubmit={showChangePassword ? handleFirstLoginPasswordChange : handleSubmit}>
            <TextField
              fullWidth
              label="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Person sx={{ color: '#8b92a5' }} />
                  </InputAdornment>
                )
              }}
              sx={{
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                  '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7' },
                  '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7', borderWidth: 2 }
                }
              }}
            />
            <TextField
              fullWidth
              label={showChangePassword ? 'Current Password' : 'Password'}
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Lock sx={{ color: '#8b92a5' }} />
                  </InputAdornment>
                ),
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                )
              }}
              sx={{
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                  '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7' },
                  '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7', borderWidth: 2 }
                }
              }}
            />
            {showChangePassword && (
              <TextField
                fullWidth
                label="New Password"
                type={showPassword ? 'text' : 'password'}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                margin="normal"
                helperText="Password must be at least 8 characters"
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <Lock sx={{ color: '#8b92a5' }} />
                    </InputAdornment>
                  )
                }}
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7' },
                    '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#4fc3f7', borderWidth: 2 }
                  }
                }}
              />
            )}
            <Button
              fullWidth
              variant="contained"
              type="submit"
              size="large"
              disabled={loading}
              sx={{
                mt: 3,
                py: 1.5,
                borderRadius: 2,
                fontWeight: 600,
                fontSize: '1rem',
                background: 'linear-gradient(135deg, #4fc3f7 0%, #29b6f6 100%)',
                boxShadow: '0 4px 14px rgba(79, 195, 247, 0.4)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #29b6f6 0%, #0288d1 100%)',
                  boxShadow: '0 6px 20px rgba(79, 195, 247, 0.5)'
                },
                '&:disabled': {
                  background: '#e0e0e0',
                  color: '#9e9e9e'
                }
              }}
            >
              {loading ? 'Please wait...' : showChangePassword ? 'Change Password' : 'Sign In'}
            </Button>
          </form>

          <Box mt={3} textAlign="center">
            <Typography variant="caption" color="text.secondary">
              Enterprise Backup Solution v1.0
            </Typography>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}
