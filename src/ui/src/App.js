import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material';
import { ApiProvider } from './services/ApiContext';
import { useAuthStore } from './store/authStore';

import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Jobs from './pages/Jobs';
import Backups from './pages/Backups';
import Restore from './pages/Restore';
import Repositories from './pages/Repositories';
import Agents from './pages/Agents';
import Settings from './pages/Settings';
import Reports from './pages/Reports';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: { main: '#1976d2' },
    secondary: { main: '#dc004e' },
    success: { main: '#2e7d32' },
    warning: { main: '#ed6c02' },
    error: { main: '#d32f2f' },
    background: { default: '#f5f5f5' },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
  },
  components: {
    MuiButton: { styleOverrides: { root: { textTransform: 'none' } } },
    MuiCard: { styleOverrides: { root: { borderRadius: 12 } } },
  },
});

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />;
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <ApiProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route index element={<Navigate to="/dashboard" />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="jobs" element={<Jobs />} />
              <Route path="backups" element={<Backups />} />
              <Route path="restore" element={<Restore />} />
              <Route path="repositories" element={<Repositories />} />
              <Route path="agents" element={<Agents />} />
              <Route path="settings" element={<Settings />} />
              <Route path="reports" element={<Reports />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ApiProvider>
    </ThemeProvider>
  );
}

export default App;
