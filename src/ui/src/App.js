import React, { useMemo, useState } from 'react';
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
import Users from './pages/Users';
import AuditLogs from './pages/AuditLogs';

const getTheme = (mode) => createTheme({
  palette: {
    mode,
    primary: { main: '#4fc3f7' },
    secondary: { main: '#ab47bc' },
    background: { 
      default: mode === 'light' ? '#f5f6f8' : '#121212', 
      paper: mode === 'light' ? '#ffffff' : '#1e1e1e' 
    },
  },
});

const ProtectedRoute = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />;
};

function App() {
  const [mode, setMode] = useState('light');
  const theme = useMemo(() => getTheme(mode), [mode]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <ApiProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/" element={<ProtectedRoute><Layout mode={mode} setMode={setMode} /></ProtectedRoute>}>
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
              <Route path="users" element={<Users />} />
              <Route path="audit-logs" element={<AuditLogs />} />
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
