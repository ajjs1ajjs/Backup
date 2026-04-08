import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  AppBar,
  Avatar,
  Badge,
  Box,
  Collapse,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Toolbar,
  Typography
} from '@mui/material';
import {
  Assessment as ReportIcon,
  Backup as BackupIcon,
  Cloud as CloudIcon,
  Computer as ComputerIcon,
  Dashboard as DashboardIcon,
  Dns as DnsIcon,
  DeveloperBoard as VMIcon,
  ExpandLess,
  ExpandMore,
  Group as GroupIcon,
  Logout as LogoutIcon,
  Menu as MenuIcon,
  Notifications as NotificationsIcon,
  Restore as RestoreIcon,
  Settings as SettingsIcon,
  Storage as StorageIcon,
  Warning as WarningIcon
} from '@mui/icons-material';
import { useAuthStore } from '../store/authStore';

const drawerWidth = 260;

const navSections = [
  {
    label: 'OVERVIEW',
    items: [
      { text: 'Dashboard', icon: <DashboardIcon />, path: '/dashboard' }
    ]
  },
  {
    label: 'PROTECTION',
    items: [
      { text: 'Backup Jobs', icon: <BackupIcon />, path: '/jobs' },
      { text: 'Restore', icon: <RestoreIcon />, path: '/restore' },
      { text: 'Replication', icon: <CloudIcon />, path: '/replication' }
    ]
  },
  {
    label: 'INFRASTRUCTURE',
    items: [
      { text: 'Hypervisors', icon: <DnsIcon />, path: '/hypervisors' },
      { text: 'Virtual Machines', icon: <VMIcon />, path: '/virtual-machines' },
      { text: 'Inventory', icon: <StorageIcon />, path: '/inventory' },
      { text: 'Repositories', icon: <StorageIcon />, path: '/repositories' },
      { text: 'Agents', icon: <ComputerIcon />, path: '/agents' }
    ]
  },
  {
    label: 'MONITORING',
    items: [
      { text: 'Alerts', icon: <WarningIcon />, path: '/alerts' },
      { text: 'Reports', icon: <ReportIcon />, path: '/reports' }
    ]
  },
  {
    label: 'CONFIGURATION',
    items: [
      { text: 'Users', icon: <GroupIcon />, path: '/users' },
      { text: 'Audit Logs', icon: <ReportIcon />, path: '/audit-logs' },
      { text: 'Settings', icon: <SettingsIcon />, path: '/settings' }
    ]
  }
];

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout, username } = useAuthStore();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState(null);
  const [expandedSections, setExpandedSections] = useState({});

  const handleDrawerToggle = () => setMobileOpen((prev) => !prev);
  const handleMenu = (event) => setAnchorEl(event.currentTarget);
  const handleClose = () => setAnchorEl(null);

  const handleLogout = () => {
    handleClose();
    logout();
    navigate('/login');
  };

  const toggleSection = (label) => {
    setExpandedSections((prev) => ({ ...prev, [label]: !prev[label] }));
  };

  const currentTitle = navSections
    .flatMap((section) => section.items)
    .find((item) => item.path === location.pathname)?.text || 'Dashboard';

  const drawer = (
    <Box sx={{ bgcolor: '#1a1d23', color: '#fff', height: '100%' }}>
      <Toolbar sx={{ justifyContent: 'center', borderBottom: '1px solid rgba(255,255,255,0.1)' }}>
        <Typography variant="h6" sx={{ fontWeight: 700, color: '#4fc3f7', letterSpacing: 1 }}>
          BACKUP SYSTEM
        </Typography>
      </Toolbar>

      <Box sx={{ overflowY: 'auto', py: 1 }}>
        {navSections.map((section) => (
          <Box key={section.label}>
            <ListItemButton onClick={() => toggleSection(section.label)} sx={{ px: 2, py: 0.75 }}>
              <Typography
                variant="caption"
                sx={{ color: '#8b92a5', fontWeight: 600, letterSpacing: 1, fontSize: '0.7rem' }}
              >
                {section.label}
              </Typography>
              <Box sx={{ ml: 'auto' }}>
                {expandedSections[section.label] ? (
                  <ExpandLess sx={{ fontSize: 16, color: '#8b92a5' }} />
                ) : (
                  <ExpandMore sx={{ fontSize: 16, color: '#8b92a5' }} />
                )}
              </Box>
            </ListItemButton>
            <Collapse in={expandedSections[section.label] !== false} timeout={200} unmountOnExit={false}>
              <List sx={{ py: 0 }}>
                {section.items.map((item) => {
                  const isActive = location.pathname === item.path;

                  return (
                    <ListItem key={item.text} disablePadding>
                      <ListItemButton
                        selected={isActive}
                        onClick={() => {
                          navigate(item.path);
                          setMobileOpen(false);
                        }}
                        sx={{
                          py: 0.7,
                          px: 3,
                          '&.Mui-selected': {
                            bgcolor: 'rgba(79, 195, 247, 0.15)',
                            borderRight: '3px solid #4fc3f7'
                          },
                          '&.Mui-selected:hover': {
                            bgcolor: 'rgba(79, 195, 247, 0.25)'
                          },
                          '& .MuiListItemIcon-root': {
                            color: isActive ? '#4fc3f7' : '#8b92a5',
                            minWidth: 36
                          },
                          '& .MuiListItemText-primary': {
                            color: isActive ? '#fff' : '#c0c6d4',
                            fontSize: '0.875rem'
                          }
                        }}
                      >
                        <ListItemIcon>{item.icon}</ListItemIcon>
                        <ListItemText primary={item.text} />
                      </ListItemButton>
                    </ListItem>
                  );
                })}
              </List>
            </Collapse>
          </Box>
        ))}
      </Box>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <AppBar
        position="fixed"
        sx={{
          zIndex: (theme) => theme.zIndex.drawer + 1,
          bgcolor: '#fff',
          color: '#333',
          boxShadow: '0 1px 3px rgba(0,0,0,0.12)'
        }}
      >
        <Toolbar>
          <IconButton color="inherit" edge="start" onClick={handleDrawerToggle} sx={{ mr: 2, display: { sm: 'none' } }}>
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 500, color: '#333', fontSize: '1rem' }}>
            {currentTitle}
          </Typography>
          <IconButton sx={{ color: '#666' }}>
            <Badge badgeContent={0} color="error">
              <NotificationsIcon />
            </Badge>
          </IconButton>
          <IconButton onClick={handleMenu} sx={{ ml: 1, color: '#666' }}>
            <Avatar sx={{ width: 32, height: 32, bgcolor: '#4fc3f7', fontSize: '0.875rem' }}>
              {(username || 'A').charAt(0).toUpperCase()}
            </Avatar>
          </IconButton>
          <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleClose}>
            <MenuItem onClick={handleClose} sx={{ fontSize: '0.875rem' }}>Profile</MenuItem>
            <MenuItem onClick={handleLogout} sx={{ fontSize: '0.875rem' }}>
              <LogoutIcon sx={{ mr: 1, fontSize: 18 }} /> Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>

      <Box component="nav" sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}>
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{ keepMounted: true }}
          sx={{
            display: { xs: 'block', sm: 'none' },
            '& .MuiDrawer-paper': { width: drawerWidth, boxSizing: 'border-box', border: 'none' }
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': {
              width: drawerWidth,
              boxSizing: 'border-box',
              border: 'none',
              top: 64,
              height: 'calc(100% - 64px)'
            }
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          bgcolor: '#f5f6f8',
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          minHeight: '100vh'
        }}
      >
        <Toolbar />
        <Box sx={{ p: 3 }}>
          <Outlet />
        </Box>
      </Box>
    </Box>
  );
}
