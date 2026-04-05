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
import Replication from './pages/Replication';
import Inventory from './pages/Inventory';
import Repositories from './pages/Repositories';
import Agents from './pages/Agents';
import Settings from './pages/Settings';
import Reports from './pages/Reports';
import Alerts from './pages/Alerts';
import VirtualMachines from './pages/VirtualMachines';
import Hypervisors from './pages/Hypervisors';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: { main: '#4fc3f7' },
    secondary: { main: '#ab47bc' },
    success: { main: '#66bb6a' },
    warning: { main: '#ffa726' },
    error: { main: '#ef5350' },
    background: { default: '#f5f6f8', paper: '#ffffff' },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h5: { fontWeight: 600 },
    h6: { fontWeight: 600 },
  },
  components: {
    MuiButton: { styleOverrides: { root: { textTransform: 'none', borderRadius: 8 } } },
    MuiCard: { styleOverrides: { root: { borderRadius: 12, boxShadow: '0 1px 3px rgba(0,0,0,0.08), 0 1px 2px rgba(0,0,0,0.06)' } } },
    MuiPaper: { styleOverrides: { root: { boxShadow: '0 1px 3px rgba(0,0,0,0.08)' } } },
    MuiTableCell: { styleOverrides: { root: { borderBottom: '1px solid #f0f0f0' } } },
  },
});

const ProtectedRoute = ({ children }) => {
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
              <Route path="replication" element={<Replication />} />
              <Route path="inventory" element={<Inventory />} />
              <Route path="repositories" element={<Repositories />} />
              <Route path="agents" element={<Agents />} />
              <Route path="alerts" element={<Alerts />} />
              <Route path="settings" element={<Settings />} />
              <Route path="reports" element={<Reports />} />
              <Route path="virtual-machines" element={<VirtualMachines />} />
              <Route path="hypervisors" element={<Hypervisors />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ApiProvider>
    </ThemeProvider>
  );
}

export default App;
