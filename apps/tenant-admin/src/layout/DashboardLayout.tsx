import React from 'react';
import {
    AppBar,
    Box,
    CssBaseline,
    Drawer,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Toolbar,
    Typography,
    Avatar,
    ThemeProvider
} from '@mui/material';
import {
    Dashboard as DashboardIcon,
    People as PeopleIcon,
    Settings as SettingsIcon,
    CloudQueue as CloudIcon,
    VpnKey as VpnKeyIcon,
    VerifiedUser as VerifiedUserIcon
} from '@mui/icons-material';
import { theme } from '../theme/theme';
import { Tenant } from '../types';

const drawerWidth = 260;

interface DashboardLayoutProps {
    children: React.ReactNode;
    view: string;
    setView: (view: any) => void;
    tenant: Tenant | null;
}

const DashboardLayout: React.FC<DashboardLayoutProps> = ({ children, view, setView, tenant }) => {
    const navItems = [
        { id: 'dashboard', label: 'Dashboard', icon: <DashboardIcon /> },
        { id: 'users', label: 'Users & Groups', icon: <PeopleIcon /> },
        { id: 'signin_policies', label: 'Sign-in Policies', icon: <VpnKeyIcon /> },
        { id: 'access_policies', label: 'Access Policies', icon: <VerifiedUserIcon /> },
        { id: 'nodes', label: 'Nodes & Gateways', icon: <CloudIcon /> },
        { id: 'admins', label: 'Console Administrators', icon: <PeopleIcon /> },
        { id: 'settings', label: 'Settings', icon: <SettingsIcon /> },
    ];

    return (
        <ThemeProvider theme={theme}>
            <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                <CssBaseline />
                <AppBar position="fixed" elevation={0} sx={{ zIndex: (theme) => theme.zIndex.drawer + 1, backgroundColor: 'rgba(255, 255, 255, 0.95)', backdropFilter: 'blur(8px)', color: '#5f6368', borderBottom: '1px solid #dadce0' }}>
                    <Toolbar>
                        <CloudIcon sx={{ color: '#1a73e8', fontSize: 30, mr: 1 }} />
                        <Typography variant="h6" noWrap component="div" sx={{ color: '#202124', fontWeight: 700 }}>
                            Tridorian ZTNA Console
                        </Typography>
                        <Typography variant="body2" sx={{ ml: 2, color: '#5f6368', borderLeft: '1px solid #dadce0', pl: 2 }}>
                            {tenant?.name}
                        </Typography>
                        <Box sx={{ flexGrow: 1 }} />
                        <Avatar sx={{ bgcolor: '#1a73e8', cursor: 'pointer' }} onClick={() => window.location.reload()}>A</Avatar>
                    </Toolbar>
                </AppBar>

                <Drawer variant="permanent" sx={{ width: drawerWidth, flexShrink: 0, [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box', borderRight: '1px solid #dadce0' } }}>
                    <Toolbar />
                    <Box sx={{ overflow: 'auto', mt: 2 }}>
                        <List>
                            {navItems.map((item) => (
                                <ListItem key={item.id} disablePadding>
                                    <ListItemButton
                                        selected={view === item.id}
                                        onClick={() => setView(item.id)}
                                    >
                                        <ListItemIcon sx={{ color: view === item.id ? '#1a73e8' : 'inherit' }}>
                                            {item.icon}
                                        </ListItemIcon>
                                        <ListItemText primary={item.label} />
                                    </ListItemButton>
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                </Drawer>

                <Box component="main" sx={{ flexGrow: 1, p: 4, mt: 8, bgcolor: '#f8f9fa' }}>
                    {children}
                </Box>
            </Box>
        </ThemeProvider>
    );
};

export default DashboardLayout;
