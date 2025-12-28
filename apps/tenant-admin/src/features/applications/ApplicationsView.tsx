import React, { useState, useEffect } from 'react';
import {
    Box,
    Typography,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    Grid,
    IconButton,
    Chip,
    Paper,
    CircularProgress,
    Stack,
    useMediaQuery,
    useTheme
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    Apps as AppsIcon,
} from '@mui/icons-material';
import { Application, ApplicationCIDR } from '../../types';

const ApplicationsView: React.FC = () => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));

    const [applications, setApplications] = useState<Application[]>([]);
    const [loading, setLoading] = useState(false);
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingApp, setEditingApp] = useState<Application | null>(null);
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        cidrs: ['']
    });

    useEffect(() => {
        fetchApplications();
    }, []);

    const fetchApplications = async () => {
        try {
            const res = await fetch('/api/v1/applications');
            const data = await res.json();
            if (data.success) {
                setApplications(data.data || []);
            }
        } catch (err) {
            console.error('Failed to fetch applications', err);
        }
    };

    const handleOpenDialog = (app?: Application) => {
        if (app) {
            setEditingApp(app);
            // Fetch CIDRs for this app
            fetchAppCIDRs(app.id);
        } else {
            setEditingApp(null);
            setFormData({
                name: '',
                description: '',
                cidrs: ['']
            });
        }
        setDialogOpen(true);
    };

    const fetchAppCIDRs = async (appId: string) => {
        try {
            const res = await fetch(`/api/v1/applications?id=${appId}`);
            const data = await res.json();
            if (data.success) {
                const cidrs = data.data.cidrs?.map((c: ApplicationCIDR) => c.cidr) || [''];
                setFormData({
                    name: data.data.application.name,
                    description: data.data.application.description || '',
                    cidrs: cidrs.length > 0 ? cidrs : ['']
                });
            }
        } catch (err) {
            console.error('Failed to fetch app CIDRs', err);
        }
    };

    const handleSave = async () => {
        setLoading(true);
        try {
            const url = '/api/v1/applications';
            const method = editingApp ? 'PATCH' : 'POST';
            const body = editingApp
                ? { id: editingApp.id, ...formData }
                : formData;

            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                setDialogOpen(false);
                fetchApplications();
            }
        } catch (err) {
            console.error('Save failed', err);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!window.confirm('Are you sure you want to delete this application?')) return;
        try {
            const res = await fetch('/api/v1/applications', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id })
            });
            if (res.ok) fetchApplications();
        } catch (err) {
            console.error('Delete failed', err);
        }
    };

    const handleAddCIDR = () => {
        setFormData({ ...formData, cidrs: [...formData.cidrs, ''] });
    };

    const handleRemoveCIDR = (index: number) => {
        const newCIDRs = formData.cidrs.filter((_, i) => i !== index);
        setFormData({ ...formData, cidrs: newCIDRs.length > 0 ? newCIDRs : [''] });
    };

    const handleCIDRChange = (index: number, value: string) => {
        const newCIDRs = [...formData.cidrs];
        newCIDRs[index] = value;
        setFormData({ ...formData, cidrs: newCIDRs });
    };

    return (
        <Box sx={{ p: 0 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 800, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 2 }}>
                        <AppsIcon sx={{ fontSize: 40, color: 'primary.main' }} />
                        Applications
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
                        Pre-define applications with multiple CIDRs for use in access policies.
                    </Typography>
                </Box>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => handleOpenDialog()}
                    sx={{
                        borderRadius: 2.5,
                        px: 3,
                        py: 1,
                        boxShadow: '0 4px 12px rgba(26, 115, 232, 0.2)',
                        textTransform: 'none',
                        fontWeight: 700
                    }}
                >
                    Create Application
                </Button>
            </Box>

            <Grid container spacing={3}>
                {applications.map((app) => (
                    <Grid key={app.id} size={12}>
                        <Card sx={{
                            borderRadius: 4,
                            border: '1px solid #eef0f2',
                            boxShadow: '0 2px 8px rgba(0,0,0,0.03)',
                            transition: 'all 0.2s',
                            '&:hover': {
                                boxShadow: '0 8px 16px rgba(0,0,0,0.06)',
                                borderColor: 'primary.main'
                            }
                        }}>
                            <CardContent sx={{ p: 3 }}>
                                <Grid container spacing={2} alignItems="center">
                                    <Grid size={{ xs: 12, md: 8 }}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                            <Box sx={{
                                                width: 48,
                                                height: 48,
                                                borderRadius: 3,
                                                bgcolor: 'primary.light',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                color: '#fff',
                                                opacity: 0.9
                                            }}>
                                                <AppsIcon />
                                            </Box>
                                            <Box>
                                                <Typography variant="h6" sx={{ fontWeight: 700 }}>{app.name}</Typography>
                                                <Typography variant="caption" color="text.secondary">
                                                    {app.description || 'No description'}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </Grid>
                                    <Grid size={{ xs: 6, md: 2 }} sx={{ display: { xs: 'none', md: 'block' } }}>
                                        <Typography variant="caption" color="text.secondary">
                                            CIDRs will be shown when editing
                                        </Typography>
                                    </Grid>
                                    <Grid size={{ xs: 12, md: 2 }} sx={{ textAlign: 'right', mt: { xs: 2, md: 0 } }}>
                                        <IconButton size="small" onClick={() => handleOpenDialog(app)} sx={{ mr: 1, color: 'primary.main', bgcolor: 'rgba(26, 115, 232, 0.05)' }}>
                                            <EditIcon fontSize="small" />
                                        </IconButton>
                                        <IconButton size="small" onClick={() => handleDelete(app.id)} sx={{ color: 'error.main', bgcolor: 'rgba(211, 47, 47, 0.05)' }}>
                                            <DeleteIcon fontSize="small" />
                                        </IconButton>
                                    </Grid>
                                </Grid>
                            </CardContent>
                        </Card>
                    </Grid>
                ))}

                {applications.length === 0 && (
                    <Grid size={12}>
                        <Paper sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: 'grey.50', border: '2px dashed #e0e0e0' }}>
                            <AppsIcon sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
                            <Typography variant="h6" color="text.secondary">No applications defined</Typography>
                            <Typography variant="body2" color="text.disabled" sx={{ mt: 1 }}>
                                Start by creating your first application with multiple CIDRs.
                            </Typography>
                        </Paper>
                    </Grid>
                )}
            </Grid>

            <Dialog
                open={dialogOpen}
                onClose={(_, reason) => {
                    if (reason !== 'backdropClick' && !loading) {
                        setDialogOpen(false);
                    }
                }}
                maxWidth="md"
                fullWidth
                fullScreen={isMobile}
                PaperProps={{
                    sx: { borderRadius: isMobile ? 0 : 4, boxShadow: '0 24px 48px rgba(0,0,0,0.1)' }
                }}
            >
                <DialogTitle sx={{ px: 4, pt: 4, pb: 2 }}>
                    <Typography variant="h5" sx={{ fontWeight: 800 }}>
                        {editingApp ? 'Edit Application' : 'Create Application'}
                    </Typography>
                </DialogTitle>
                <DialogContent sx={{ px: 4 }}>
                    <Grid container spacing={3} sx={{ mt: 0.5 }}>
                        <Grid size={12}>
                            <TextField
                                fullWidth
                                label="Application Name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="e.g. Internal Services"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                            />
                        </Grid>

                        <Grid size={12}>
                            <TextField
                                fullWidth
                                multiline
                                rows={2}
                                label="Description"
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                placeholder="e.g. All internal microservices"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                            />
                        </Grid>

                        <Grid size={12}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 2, color: 'text.secondary' }}>
                                Network CIDRs
                            </Typography>
                            <Stack spacing={2}>
                                {formData.cidrs.map((cidr, index) => (
                                    <Box key={index} sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
                                        <TextField
                                            fullWidth
                                            label={`CIDR ${index + 1}`}
                                            value={cidr}
                                            onChange={(e) => handleCIDRChange(index, e.target.value)}
                                            placeholder="e.g. 10.0.0.0/24"
                                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                                        />
                                        {formData.cidrs.length > 1 && (
                                            <IconButton
                                                size="small"
                                                color="error"
                                                onClick={() => handleRemoveCIDR(index)}
                                            >
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        )}
                                    </Box>
                                ))}
                            </Stack>
                            <Button
                                size="small"
                                startIcon={<AddIcon />}
                                onClick={handleAddCIDR}
                                sx={{ mt: 2, textTransform: 'none', fontWeight: 600 }}
                            >
                                Add CIDR
                            </Button>
                        </Grid>
                    </Grid>
                </DialogContent>
                <DialogActions sx={{ px: 4, pb: 4, pt: 2 }}>
                    <Button onClick={() => setDialogOpen(false)} disabled={loading} sx={{ fontWeight: 700 }}>Cancel</Button>
                    <Button
                        variant="contained"
                        onClick={handleSave}
                        disabled={loading}
                        sx={{
                            borderRadius: 2.5,
                            px: 4,
                            fontWeight: 700,
                            boxShadow: '0 4px 12px rgba(26, 115, 232, 0.2)'
                        }}
                    >
                        {loading ? <CircularProgress size={24} color="inherit" /> : (editingApp ? 'Update Application' : 'Create Application')}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default ApplicationsView;
