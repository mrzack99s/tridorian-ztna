import React, { useState, useEffect } from 'react';
import {
    AppBar,
    Box,
    CssBaseline,
    Drawer,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Toolbar,
    Typography,
    Button,
    Card,
    CardContent,
    Grid,
    Chip,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    CircularProgress,
    Divider,
    ThemeProvider,
    createTheme,
    alpha,
    Avatar,
    Fade,
    Grow,
    InputAdornment
} from '@mui/material';
import {
    Dashboard as DashboardIcon,
    Business as BusinessIcon,
    People as PeopleIcon,
    Settings as SettingsIcon,
    Add as AddIcon,
    CheckCircle as CheckCircleIcon,
    ErrorOutline as ErrorOutlineIcon,
    MoreVert as MoreVertIcon,
    Search as SearchIcon,
    Menu as MenuIcon,
    Email as EmailIcon,
    VpnKey as VpnKeyIcon,
    CloudQueue as CloudIcon,
    Delete as DeleteIcon
} from '@mui/icons-material';

// Premium Google Cloud-inspired Theme
const theme = createTheme({
    palette: {
        primary: {
            main: '#1a73e8', // Google Blue
            light: '#e8f0fe',
            dark: '#174ea6',
        },
        success: {
            main: '#1e8e3e', // Google Green
        },
        warning: {
            main: '#f9ab00', // Google Yellow
        },
        background: {
            default: '#f1f3f4',
            paper: '#ffffff',
        },
        text: {
            primary: '#202124',
            secondary: '#5f6368',
        },
    },
    typography: {
        fontFamily: '"Google Sans", "Roboto", "Helvetica", "Arial", sans-serif',
        h5: {
            fontWeight: 500,
            letterSpacing: -0.5,
        },
        h6: {
            fontWeight: 500,
            fontSize: '1.1rem',
        },
        button: {
            textTransform: 'none',
            fontWeight: 500,
            fontSize: '0.875rem',
        },
    },
    shape: {
        borderRadius: 8,
    },
    components: {
        MuiButton: {
            styleOverrides: {
                root: {
                    padding: '8px 24px',
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: '0 1px 2px 0 rgba(60,64,67,.302), 0 1px 3px 1px rgba(60,64,67,.149)',
                    },
                },
                containedPrimary: {
                    backgroundColor: '#1a73e8',
                }
            },
        },
        MuiCard: {
            styleOverrides: {
                root: {
                    border: '1px solid #dadce0',
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: '0 1px 2px 0 rgba(60,64,67,.3), 0 1px 3px 1px rgba(60,64,67,.15)',
                    },
                },
            },
        },
        MuiDialog: {
            styleOverrides: {
                paper: {
                    borderRadius: 12,
                    padding: 8,
                }
            }
        }
    },
});

const drawerWidth = 260;

interface Tenant {
    id: string;
    name: string;
    slug: string;
    primary_domain: string;
    google_client_id?: string;
}

const App: React.FC = () => {
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [tenants, setTenants] = useState<Tenant[]>([]);
    const [loading, setLoading] = useState(true);
    const [showCreateModal, setShowCreateModal] = useState(false);

    const [newTenantName, setNewTenantName] = useState('');
    const [successData, setSuccessData] = useState<{ email: string; pass: string } | null>(null);
    const [showSuccessModal, setShowSuccessModal] = useState(false);

    // Auth Form State
    const [loginEmail, setLoginEmail] = useState('');
    const [loginPassword, setLoginPassword] = useState('');

    const fetchTenants = async () => {
        try {
            const res = await fetch('/api/v1/tenants');
            if (res.status === 401) {
                setIsLoggedIn(false);
                return;
            }
            const data = await res.json();
            if (data.success) {
                setTenants(data.data);
            }
        } catch (err) {
            console.error('Failed to fetch tenants:', err);
        } finally {
            setLoading(false);
        }
    };

    const checkSession = async () => {
        try {
            const res = await fetch('/auth/backoffice/me');
            if (res.ok) {
                setIsLoggedIn(true);
                fetchTenants();
            } else {
                setIsLoggedIn(false);
                setLoading(false);
            }
        } catch (err) {
            setIsLoggedIn(false);
            setLoading(false);
        }
    };

    useEffect(() => {
        checkSession();
    }, []);

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const res = await fetch('/auth/backoffice/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: loginEmail, password: loginPassword })
            });
            console.log(res);

            const data = await res.json();
            if (data.success) {
                setIsLoggedIn(true);
                fetchTenants();
            } else {
                alert('Login failed: ' + data.message);
            }
        } catch (err) {
            console.error('Login error:', err);
        }
    };

    const handleLogout = async () => {
        await fetch('/auth/backoffice/logout', { method: 'POST' });
        setIsLoggedIn(false);
    };

    const handleCreateTenant = async () => {
        if (!newTenantName) {
            alert('Please enter a tenant name');
            return;
        }

        try {
            const res = await fetch('/api/v1/tenants', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name: newTenantName
                })
            });
            const data = await res.json();
            if (data.success) {
                setShowCreateModal(false);
                setNewTenantName('');
                setSuccessData({
                    email: data.data.admin_email,
                    pass: data.data.admin_password
                });
                setShowSuccessModal(true);
                fetchTenants();
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) {
            console.error('Failed to create tenant:', err);
        }
    };

    const handleDeleteTenant = async (id: string) => {
        if (!confirm('Are you sure you want to delete this tenant? This will delete all associated data.')) return;
        try {
            const res = await fetch('/api/v1/tenants', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id })
            });
            const data = await res.json();
            if (data.success) {
                fetchTenants();
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) {
            console.error('Failed to delete tenant:', err);
        }
    };

    if (!isLoggedIn && !loading) {
        return (
            <ThemeProvider theme={theme}>
                <Box sx={{
                    minHeight: '100vh',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    bgcolor: '#f1f3f4'
                }}>
                    <Card sx={{ maxWidth: 400, width: '100%', p: 2, borderRadius: 3 }}>
                        <CardContent>
                            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mb: 3 }}>
                                <CloudIcon sx={{ color: '#1a73e8', fontSize: 48, mb: 1 }} />
                                <Typography variant="h5" sx={{ fontWeight: 700 }}>Tridorian Console</Typography>
                                <Typography variant="body2" color="text.secondary">Sign in with your administrator account</Typography>
                            </Box>
                            <form onSubmit={handleLogin}>
                                <TextField
                                    fullWidth
                                    label="Email"
                                    margin="normal"
                                    value={loginEmail}
                                    onChange={(e) => setLoginEmail(e.target.value)}
                                    required
                                />
                                <TextField
                                    fullWidth
                                    label="Password"
                                    type="password"
                                    margin="normal"
                                    value={loginPassword}
                                    onChange={(e) => setLoginPassword(e.target.value)}
                                    required
                                />
                                <Button
                                    fullWidth
                                    variant="contained"
                                    type="submit"
                                    sx={{ mt: 3, py: 1.5 }}
                                >
                                    Sign In
                                </Button>
                            </form>
                        </CardContent>
                    </Card>
                </Box>
            </ThemeProvider>
        );
    }

    return (
        <ThemeProvider theme={theme}>
            <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                <CssBaseline />

                {/* App Bar - Modern Glass style */}
                <AppBar
                    position="fixed"
                    elevation={0}
                    sx={{
                        zIndex: (theme) => theme.zIndex.drawer + 1,
                        backgroundColor: 'rgba(255, 255, 255, 0.95)',
                        backdropFilter: 'blur(8px)',
                        color: '#5f6368',
                        borderBottom: '1px solid #dadce0',
                    }}
                >
                    <Toolbar sx={{ px: '16px !important' }}>
                        <IconButton edge="start" color="inherit" sx={{ mr: 1 }}>
                            <MenuIcon />
                        </IconButton>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <CloudIcon sx={{ color: '#1a73e8', fontSize: 30 }} />
                            <Typography variant="h6" noWrap component="div" sx={{ color: '#202124', letterSpacing: -0.5, display: 'flex', alignItems: 'center' }}>
                                <Box component="span" sx={{ fontWeight: 700 }}>Tridorian</Box>
                                <Box component="span" sx={{ ml: 1, fontWeight: 400, color: '#5f6368' }}>Console</Box>
                            </Typography>
                        </Box>

                        <Box sx={{ flexGrow: 1 }} />

                        {/* Search Bar - Premium styled */}
                        <Box sx={{
                            display: 'flex',
                            alignItems: 'center',
                            backgroundColor: '#f1f3f4',
                            px: 2,
                            py: 0.8,
                            borderRadius: '8px',
                            width: 500,
                            transition: 'all 0.2s',
                            '&:focus-within': {
                                backgroundColor: '#ffffff',
                                boxShadow: '0 1px 1px 0 rgba(65,69,73,0.3), 0 1px 3px 1px rgba(65,69,73,0.15)'
                            }
                        }}>
                            <SearchIcon sx={{ color: '#5f6368', mr: 1.5, fontSize: 20 }} />
                            <Typography variant="body2" sx={{ color: '#80868b', flexGrow: 1 }}>Search resources, services, and docs</Typography>
                            <Typography variant="caption" sx={{ color: '#80868b', border: '1px solid #dadce0', px: 0.5, borderRadius: 0.5 }}>/</Typography>
                        </Box>

                        <Box sx={{ flexGrow: 1 }} />

                        <IconButton color="inherit" sx={{ mr: 1 }} onClick={handleLogout}>
                            <SettingsIcon fontSize="small" />
                        </IconButton>
                        <Avatar sx={{ width: 32, height: 32, bgcolor: '#1a73e8', fontSize: '0.875rem', fontWeight: 600, cursor: 'pointer' }} onClick={handleLogout}>Z</Avatar>
                    </Toolbar>
                </AppBar>

                {/* Sidebar - Dark theme for high contrast common in GCP */}
                <Drawer
                    variant="permanent"
                    sx={{
                        width: drawerWidth,
                        flexShrink: 0,
                        [`& .MuiDrawer-paper`]: {
                            width: drawerWidth,
                            boxSizing: 'border-box',
                            backgroundColor: '#ffffff',
                            borderRight: '1px solid #dadce0'
                        },
                    }}
                >
                    <Toolbar />
                    <Box sx={{ overflow: 'auto', mt: 2 }}>
                        <List sx={{ px: 1 }}>
                            {[
                                { text: 'Dashboard', icon: <DashboardIcon /> },
                                { text: 'Tenants', icon: <BusinessIcon />, active: true },
                                { text: 'IAM & Admin', icon: <PeopleIcon /> },
                                { text: 'Billing', icon: <Typography sx={{ fontSize: 18, fontWeight: 700 }}>฿</Typography> },
                                { text: 'Settings', icon: <SettingsIcon /> },
                            ].map((item) => (
                                <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}>
                                    <ListItemButton
                                        selected={item.active}
                                        sx={{
                                            borderRadius: '0 20px 20px 0',
                                            mr: 1,
                                            '&.Mui-selected': {
                                                backgroundColor: alpha('#1a73e8', 0.1),
                                                color: '#1a73e8',
                                                '& .MuiListItemIcon-root': { color: '#1a73e8' },
                                                '&:hover': { backgroundColor: alpha('#1a73e8', 0.15) }
                                            },
                                            '&:hover': {
                                                backgroundColor: '#f1f3f4',
                                                borderRadius: '0 20px 20px 0',
                                            }
                                        }}
                                    >
                                        <ListItemIcon sx={{ minWidth: 40, color: item.active ? '#1a73e8' : '#5f6368' }}>{item.icon}</ListItemIcon>
                                        <ListItemText primary={item.text} primaryTypographyProps={{ fontSize: '0.875rem', fontWeight: item.active ? 600 : 400 }} />
                                    </ListItemButton>
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                </Drawer>

                {/* Main Content */}
                <Box component="main" sx={{ flexGrow: 1, p: 4, mt: 8, backgroundColor: '#f8f9fa' }}>
                    <Fade in={true} timeout={800}>
                        <Box>
                            <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                <Box>
                                    <Typography variant="h5" sx={{ mb: 1, display: 'flex', alignItems: 'center', gap: 1 }}>
                                        Tenants Management
                                    </Typography>
                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                        <Chip label="Global" size="small" sx={{ borderRadius: 1, height: 20, fontSize: '0.65rem', bgcolor: '#e8f0fe', color: '#1967d2', border: 'none' }} />
                                        <Typography variant="body2" color="text.secondary">Register and configure organizations for ZTNA gateway.</Typography>
                                    </Box>
                                </Box>
                                <Button
                                    variant="contained"
                                    startIcon={<AddIcon />}
                                    onClick={() => setShowCreateModal(true)}
                                    sx={{ bgcolor: '#1a73e8' }}
                                >
                                    Create Tenant
                                </Button>
                            </Box>

                            {loading ? (
                                <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 15 }}>
                                    <CircularProgress size={40} thickness={4} sx={{ color: '#1a73e8' }} />
                                    <Typography sx={{ mt: 3, fontWeight: 500 }} color="text.secondary">Fetching tenant metadata...</Typography>
                                </Box>
                            ) : (
                                <Grid container spacing={3}>
                                    {tenants.map((tenant, index) => (
                                        <Grid item xs={12} md={6} lg={4} key={tenant.id}>
                                            <Grow in={true} timeout={400 + index * 100}>
                                                <Card sx={{
                                                    transition: 'all 0.2s',
                                                    '&:hover': {
                                                        borderColor: '#1a73e8',
                                                        transform: 'translateY(-2px)'
                                                    }
                                                }}>
                                                    <CardContent sx={{ p: 3 }}>
                                                        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
                                                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                                                <Avatar sx={{ bgcolor: alpha(theme.palette.primary.main, 0.1), color: theme.palette.primary.main, width: 40, height: 40, borderRadius: 1.5 }}>
                                                                    <BusinessIcon />
                                                                </Avatar>
                                                                <Box>
                                                                    <Typography variant="h6" sx={{ lineHeight: 1.2 }}>{tenant.name}</Typography>
                                                                    <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>ID: {tenant.id.slice(0, 13)}</Typography>
                                                                </Box>
                                                            </Box>
                                                            <IconButton size="small" color="error" onClick={() => handleDeleteTenant(tenant.id)}><DeleteIcon fontSize="small" /></IconButton>
                                                        </Box>

                                                        <Box sx={{ mb: 3, p: 2, bgcolor: '#f1f3f4', borderRadius: 1.5 }}>
                                                            <Typography variant="caption" sx={{ display: 'block', color: '#5f6368', fontWeight: 700, mb: 0.5, letterSpacing: 0.5 }}>GATEWAY ENDPOINT</Typography>
                                                            <Typography variant="body2" sx={{ fontWeight: 600, color: tenant.primary_domain ? '#202124' : '#f9ab00', display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                                                {tenant.primary_domain || 'Awaiting Setup'}
                                                                {!tenant.primary_domain && <ErrorOutlineIcon sx={{ fontSize: 14 }} />}
                                                            </Typography>
                                                        </Box>

                                                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                                                            <Chip
                                                                size="small"
                                                                label={tenant.slug}
                                                                variant="outlined"
                                                                sx={{ border: '1px solid #dadce0', fontWeight: 600, color: '#3c4043', bgcolor: '#ffffff' }}
                                                            />
                                                            {tenant.google_client_id ? (
                                                                <Chip
                                                                    size="small"
                                                                    label="OIDC ACTIVE"
                                                                    sx={{ bgcolor: '#e6f4ea', color: '#1e8e3e', fontWeight: 600, fontSize: '0.7rem' }}
                                                                    icon={<CheckCircleIcon sx={{ fontSize: '14px !important', color: 'inherit !important' }} />}
                                                                />
                                                            ) : (
                                                                <Chip
                                                                    size="small"
                                                                    label="NO IDENTITY"
                                                                    sx={{ bgcolor: '#fef7e0', color: '#b06000', fontWeight: 600, fontSize: '0.7rem' }}
                                                                />
                                                            )}
                                                        </Box>
                                                    </CardContent>
                                                </Card>
                                            </Grow>
                                        </Grid>
                                    ))}
                                </Grid>
                            )}
                        </Box>
                    </Fade>

                    {/* Create Dialog - Modern Multi-step feel */}
                    <Dialog open={showCreateModal} onClose={() => setShowCreateModal(false)} maxWidth="sm" fullWidth>
                        <DialogTitle sx={{ pb: 1, pt: 3 }}>
                            <Typography variant="h5">Create new tenant</Typography>
                            <Typography variant="body2" color="text.secondary">Onboard a new organization and set up the primary administrator.</Typography>
                        </DialogTitle>
                        <DialogContent>
                            <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 3 }}>
                                <TextField
                                    autoFocus
                                    label="Organization Name"
                                    fullWidth
                                    value={newTenantName}
                                    onChange={(e) => setNewTenantName(e.target.value)}
                                    placeholder="e.g. Acme Corporation"
                                    helperText="A project slug will be derived from this name"
                                />
                            </Box>
                        </DialogContent>
                        <DialogActions sx={{ p: 3, pt: 1 }}>
                            <Button onClick={() => setShowCreateModal(false)} color="inherit">Cancel</Button>
                            <Button onClick={handleCreateTenant} variant="contained" disabled={!newTenantName}>
                                Confirm & Create
                            </Button>
                        </DialogActions>
                    </Dialog>

                    {/* Success Modal - Display Generated Credentials */}
                    <Dialog open={showSuccessModal} onClose={() => setShowSuccessModal(false)} maxWidth="xs" fullWidth>
                        <DialogTitle sx={{ textAlign: 'center', pt: 4 }}>
                            <CheckCircleIcon sx={{ color: '#1e8e3e', fontSize: 64, mb: 2 }} />
                            <Typography variant="h5" sx={{ fontWeight: 700 }}>Tenant Created!</Typography>
                        </DialogTitle>
                        <DialogContent>
                            <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
                                The organization has been successfully registered. Please save these administrator credentials securely.
                            </Typography>

                            <Box sx={{ bgcolor: '#f8f9fa', p: 2, borderRadius: 2, border: '1px solid #dadce0' }}>
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="caption" sx={{ fontWeight: 700, color: '#5f6368', textTransform: 'uppercase' }}>Admin Email</Typography>
                                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mt: 0.5 }}>
                                        <Typography variant="body1" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>{successData?.email}</Typography>
                                    </Box>
                                </Box>
                                <Box>
                                    <Typography variant="caption" sx={{ fontWeight: 700, color: '#5f6368', textTransform: 'uppercase' }}>Temp Password</Typography>
                                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mt: 0.5 }}>
                                        <Typography variant="body1" sx={{ fontWeight: 600, fontFamily: 'monospace', color: '#1a73e8' }}>{successData?.pass}</Typography>
                                    </Box>
                                </Box>
                            </Box>
                        </DialogContent>
                        <DialogActions sx={{ p: 3, justifyContent: 'center' }}>
                            <Button onClick={() => setShowSuccessModal(false)} variant="contained" fullWidth>
                                Got it, proceed to Dashboard
                            </Button>
                        </DialogActions>
                    </Dialog>
                </Box>
            </Box>
        </ThemeProvider>
    );
};

export default App;
