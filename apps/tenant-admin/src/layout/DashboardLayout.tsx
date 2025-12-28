import React, { useState, useEffect } from 'react';
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
    ThemeProvider,
    IconButton,
    useMediaQuery,
    useTheme as useMuiTheme,
    Container,
    Menu,
    MenuItem,
    Divider
} from '@mui/material';
import {
    Dashboard as DashboardIcon,
    People as PeopleIcon,
    Settings as SettingsIcon,
    CloudQueue as CloudIcon,
    VpnKey as VpnKeyIcon,
    VerifiedUser as VerifiedUserIcon,
    Apps as AppsIcon,
    Menu as MenuIcon,
    Logout as LogoutIcon
} from '@mui/icons-material';
import { theme } from '../theme/theme';
import { Tenant, User } from '../types';

const drawerWidth = 260;

interface DashboardLayoutProps {
    children: React.ReactNode;
    view: string;
    setView: (view: any) => void;
    tenant: Tenant | null;
    user: User | null;
}

const DashboardLayout: React.FC<DashboardLayoutProps> = ({ children, view, setView, tenant, user }) => {
    const [mobileOpen, setMobileOpen] = useState(false);
    const [backendVersion, setBackendVersion] = useState<string>('');
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const muiTheme = useMuiTheme();

    useEffect(() => {
        const fetchVersion = async () => {
            try {
                const response = await fetch('/version'); // Proxied to backend
                if (response.ok) {
                    const data = await response.json();
                    setBackendVersion(`v${data.version}`);
                }
            } catch (error) {
                console.error("Failed to fetch backend version:", error);
            }
        };
        fetchVersion();
    }, []);
    const isMobile = useMediaQuery(muiTheme.breakpoints.down('md'));

    const handleDrawerToggle = () => {
        setMobileOpen(!mobileOpen);
    };

    const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const handleMenuClose = () => {
        setAnchorEl(null);
    };

    const handleLogout = async () => {
        await fetch('/auth/mgmt/logout', { method: 'POST' });
        window.location.reload();
    };

    const navItems = [
        { id: 'dashboard', label: 'Dashboard', icon: <DashboardIcon />, roles: ['super_admin', 'admin', 'policy_admin'] },
        { id: 'signin_policies', label: 'Sign-in Policies', icon: <VpnKeyIcon />, roles: ['super_admin', 'admin', 'policy_admin'] },
        { id: 'access_policies', label: 'Access Policies', icon: <VerifiedUserIcon />, roles: ['super_admin', 'admin', 'policy_admin'] },
        { id: 'applications', label: 'Applications', icon: <AppsIcon />, roles: ['super_admin', 'admin'] },
        { id: 'nodes', label: 'Gateways', icon: <CloudIcon />, roles: ['super_admin', 'admin'] },
        { id: 'admins', label: 'Console Administrators', icon: <PeopleIcon />, roles: ['super_admin'] },
        { id: 'settings', label: 'Settings', icon: <SettingsIcon />, roles: ['super_admin', 'admin'] },
    ].filter(item => !user || item.roles.includes(user.role));

    const drawerContent = (
        <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            <Toolbar />
            <Box sx={{ overflow: 'auto', mt: 2, flexGrow: 1 }}>
                <List sx={{ pt: 1, px: 0 }}>
                    {navItems.map((item) => (
                        <ListItem key={item.id} disablePadding sx={{ mb: 0.5, display: 'block' }}>
                            <ListItemButton
                                selected={view === item.id}
                                onClick={() => {
                                    setView(item.id);
                                    if (isMobile) setMobileOpen(false);
                                }}
                                sx={{
                                    mx: 0,
                                    pl: 3,
                                    borderRadius: '0 24px 24px 0',
                                    width: '95%',
                                    '&.Mui-selected': {
                                        backgroundColor: '#e8f0fe',
                                        color: '#1a73e8',
                                        '&:hover': {
                                            backgroundColor: '#d2e3fc',
                                        },
                                        '& .MuiListItemIcon-root': {
                                            color: '#1a73e8',
                                        },
                                    },
                                    '&:hover': {
                                        backgroundColor: '#f1f3f4',
                                    },
                                    transition: 'background-color 0.2s',
                                }}
                            >
                                <ListItemIcon sx={{ minWidth: 40, color: view === item.id ? '#1a73e8' : '#5f6368' }}>
                                    {item.icon}
                                </ListItemIcon>
                                <ListItemText
                                    primary={item.label}
                                    primaryTypographyProps={{
                                        fontSize: '0.875rem',
                                        fontWeight: view === item.id ? 500 : 400
                                    }}
                                />
                            </ListItemButton>
                        </ListItem>
                    ))}
                </List>
            </Box>
            <Box sx={{ p: 2, borderTop: '1px solid #dadce0' }}>
                <Typography variant="caption" display="block" color="text.secondary" sx={{ fontWeight: 500 }}>
                    Console: v0.1.0
                </Typography>
                <Typography variant="caption" display="block" color="text.secondary">
                    Management: {backendVersion || '...'}
                </Typography>
            </Box>
        </Box >
    );

    return (
        <ThemeProvider theme={theme}>
            <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                <CssBaseline />
                <AppBar
                    position="fixed"
                    elevation={0}
                    sx={{
                        zIndex: (theme) => theme.zIndex.drawer + 1,
                        backgroundColor: '#fff',
                        color: '#5f6368',
                        borderBottom: '1px solid #dadce0',
                    }}
                >
                    <Toolbar sx={{ justifyContent: 'space-between', minHeight: 64, px: 2 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center' }}>
                            <IconButton
                                color="inherit"
                                aria-label="open drawer"
                                edge="start"
                                onClick={handleDrawerToggle}
                                sx={{ mr: 2, display: { md: 'none' } }}
                            >
                                <MenuIcon />
                            </IconButton>

                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                <CloudIcon sx={{ color: '#1a73e8', fontSize: 28 }} />
                                <Typography variant="h6" noWrap component="div" sx={{ color: '#202124', fontWeight: 700, fontSize: '1.25rem', letterSpacing: -0.5 }}>
                                    Tridorian <Box component="span" sx={{ fontWeight: 400, color: '#5f6368' }}>Console</Box>
                                </Typography>
                            </Box>
                        </Box>

                        <Box sx={{ flexGrow: 1, px: 4, display: { xs: 'none', md: 'flex' }, justifyContent: 'center' }}>
                            <Box sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 3,
                                bgcolor: '#f8f9fa',
                                px: 3,
                                py: 1,
                                borderRadius: '12px',
                                border: '1px solid #e8eaed',
                                boxShadow: '0 1px 2px rgba(60, 64, 67, 0.05)'
                            }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                    <Box sx={{
                                        width: 10,
                                        height: 10,
                                        bgcolor: '#1e8e3e',
                                        borderRadius: '50%',
                                        animation: 'pulse 2s infinite',
                                        '@keyframes pulse': {
                                            '0%': { boxShadow: '0 0 0 0 rgba(30, 142, 62, 0.4)' },
                                            '70%': { boxShadow: '0 0 0 8px rgba(30, 142, 62, 0)' },
                                            '100%': { boxShadow: '0 0 0 0 rgba(30, 142, 62, 0)' }
                                        }
                                    }} />
                                    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                                        <Typography variant="caption" sx={{ color: '#5f6368', fontWeight: 700, fontSize: '0.65rem', textTransform: 'uppercase', lineHeight: 1 }}>
                                            System Status
                                        </Typography>
                                        <Typography variant="body2" sx={{ fontWeight: 700, color: '#1e8e3e', fontSize: '0.8125rem' }}>
                                            Operational
                                        </Typography>
                                    </Box>
                                </Box>

                                <Box sx={{ height: 28, width: '1px', bgcolor: '#dadce0' }} />

                                <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                                    <Typography variant="caption" sx={{ color: '#5f6368', fontWeight: 700, fontSize: '0.65rem', textTransform: 'uppercase', lineHeight: 1 }}>
                                        Organization
                                    </Typography>
                                    <Typography variant="body2" sx={{ fontWeight: 600, color: '#202124', fontSize: '0.8125rem' }}>
                                        {tenant?.name || 'Loading...'}
                                    </Typography>
                                </Box>

                                <Box sx={{ height: 28, width: '1px', bgcolor: '#dadce0' }} />

                                <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                                    <Typography variant="caption" sx={{ color: '#5f6368', fontWeight: 700, fontSize: '0.65rem', textTransform: 'uppercase', lineHeight: 1 }}>
                                        Access Level
                                    </Typography>
                                    <Typography variant="body2" sx={{ fontWeight: 700, color: '#1a73e8', fontSize: '0.8125rem' }}>
                                        {user?.role === 'super_admin' ? 'Super Administrator' : user?.role === 'admin' ? 'Administrator' : 'Policy Admin'}
                                    </Typography>
                                </Box>
                            </Box>
                        </Box>

                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <IconButton onClick={handleMenuOpen} sx={{ p: 0 }}>
                                <Avatar sx={{ bgcolor: '#3367d6', width: 32, height: 32, fontSize: '0.875rem', fontWeight: 600 }}>
                                    {user?.name.charAt(0).toUpperCase() || 'A'}
                                </Avatar>
                            </IconButton>
                            <Menu
                                anchorEl={anchorEl}
                                open={Boolean(anchorEl)}
                                onClose={handleMenuClose}
                                transformOrigin={{ horizontal: 'right', vertical: 'top' }}
                                anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
                                PaperProps={{
                                    elevation: 0,
                                    sx: {
                                        overflow: 'visible',
                                        filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                                        mt: 1.5,
                                        '& .MuiAvatar-root': {
                                            width: 32,
                                            height: 32,
                                            ml: -0.5,
                                            mr: 1,
                                        },
                                        '&:before': {
                                            content: '""',
                                            display: 'block',
                                            position: 'absolute',
                                            top: 0,
                                            right: 14,
                                            width: 10,
                                            height: 10,
                                            bgcolor: 'background.paper',
                                            transform: 'translateY(-50%) rotate(45deg)',
                                            zIndex: 0,
                                        },
                                    },
                                }}
                            >
                                <Box sx={{ px: 2, py: 1.5, minWidth: 200 }}>
                                    <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>{user?.name}</Typography>
                                    <Typography variant="caption" color="text.secondary">{user?.email}</Typography>
                                </Box>
                                <Divider />
                                <MenuItem onClick={handleLogout} sx={{ color: 'error.main', py: 1.5 }}>
                                    <ListItemIcon>
                                        <LogoutIcon fontSize="small" sx={{ color: 'error.main' }} />
                                    </ListItemIcon>
                                    <Typography variant="body2" sx={{ fontWeight: 600 }}>Sign Out</Typography>
                                </MenuItem>
                            </Menu>
                        </Box>
                    </Toolbar>
                </AppBar>

                <Box
                    component="nav"
                    sx={{ width: { md: drawerWidth }, flexShrink: { md: 0 } }}
                >
                    {/* Mobile Drawer */}
                    <Drawer
                        variant="temporary"
                        open={mobileOpen}
                        onClose={handleDrawerToggle}
                        ModalProps={{
                            keepMounted: true, // Better open performance on mobile.
                        }}
                        sx={{
                            display: { xs: 'block', md: 'none' },
                            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
                        }}
                    >
                        {drawerContent}
                    </Drawer>

                    {/* Desktop Drawer */}
                    <Drawer
                        variant="permanent"
                        sx={{
                            display: { xs: 'none', md: 'block' },
                            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth, borderRight: '1px solid #dadce0' },
                        }}
                        open
                    >
                        {drawerContent}
                    </Drawer>
                </Box>

                <Box
                    component="main"
                    sx={{
                        flexGrow: 1,
                        p: { xs: 2, md: 4 },
                        mt: 8,
                        bgcolor: '#f8f9fa',
                        width: { md: `calc(100% - ${drawerWidth}px)` },
                        minHeight: '100vh'
                    }}
                >
                    <Container maxWidth={false} sx={{ p: 0 }}>
                        {children}
                    </Container>
                </Box>
            </Box>
        </ThemeProvider>
    );
};

export default DashboardLayout;
