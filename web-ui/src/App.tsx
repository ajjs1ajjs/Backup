import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom';
import './App.css';

// Components
import Dashboard from './components/Dashboard';
import JobList from './components/JobList';
import JobForm from './components/JobForm';

// Navigation Component
const Navigation: React.FC = () => {
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  return (
    <nav className="sidebar">
      <div className="sidebar-header">
        <h1>🚀 NovaBackup</h1>
        <span className="version">v6.0</span>
      </div>
      <ul className="nav-menu">
        <li>
          <Link to="/" className={isActive('/') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">📊</span>
            <span>Dashboard</span>
          </Link>
        </li>
        <li>
          <Link to="/jobs" className={isActive('/jobs') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">💼</span>
            <span>Jobs</span>
          </Link>
        </li>
        <li>
          <Link to="/backups" className={isActive('/backups') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">📁</span>
            <span>Backups</span>
          </Link>
        </li>
        <li>
          <Link to="/restore" className={isActive('/restore') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">🔄</span>
            <span>Restore</span>
          </Link>
        </li>
        <li>
          <Link to="/storage" className={isActive('/storage') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">💾</span>
            <span>Storage</span>
          </Link>
        </li>
        <li>
          <Link to="/settings" className={isActive('/settings') ? 'nav-link active' : 'nav-link'}>
            <span className="nav-icon">⚙️</span>
            <span>Settings</span>
          </Link>
        </li>
      </ul>
      <div className="sidebar-footer">
        <div className="status-indicator">
          <span className="status-dot online"></span>
          <span>API Connected</span>
        </div>
      </div>
    </nav>
  );
};

// Placeholder Components
const Backups: React.FC = () => {
  return (
    <div className="page">
      <div className="page-header">
        <h2>📁 Backup History</h2>
      </div>
      <div className="content">
        <p>Backup history will be displayed here.</p>
      </div>
    </div>
  );
};

const Restore: React.FC = () => {
  return (
    <div className="page">
      <div className="page-header">
        <h2>🔄 Restore</h2>
      </div>
      <div className="content">
        <p>Select a backup to restore files from.</p>
      </div>
    </div>
  );
};

const Storage: React.FC = () => {
  return (
    <div className="page">
      <div className="page-header">
        <h2>💾 Storage Management</h2>
      </div>
      <div className="content">
        <p>Storage statistics and configuration.</p>
      </div>
    </div>
  );
};

const Settings: React.FC = () => {
  return (
    <div className="page">
      <div className="page-header">
        <h2>⚙️ Settings</h2>
      </div>
      <div className="content">
        <p>Application settings and configuration.</p>
      </div>
    </div>
  );
};

// Main App Component
const App: React.FC = () => {
  return (
    <Router>
      <div className="app">
        <Navigation />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/jobs" element={<JobList />} />
            <Route path="/jobs/new" element={<JobForm />} />
            <Route path="/jobs/edit/:id" element={<JobForm />} />
            <Route path="/backups" element={<Backups />} />
            <Route path="/restore" element={<Restore />} />
            <Route path="/storage" element={<Storage />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;
