import React, { useState, useEffect } from 'react';
import {
    Box, Typography, Button, Table, TableHead, TableRow, TableCell,
    TableBody, Paper, IconButton, Dialog, DialogTitle, DialogContent,
    TextField, DialogActions, Select, MenuItem, InputLabel, FormControl, Alert, Grid,
    Avatar, Chip, Tooltip, Stack, useMediaQuery, useTheme
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    ContentCopy as CopyIcon,
    Person as PersonIcon,
    Security as SecurityIcon,
    SupervisorAccount as SupervisorIcon,
    Warning as WarningIcon
} from '@mui/icons-material';
import { Admin } from '../../types';

interface AdminsViewProps {
    admins: Admin[];
    domains: string[];
    onCreate: (admin: any) => Promise<string | null>;
    onDelete: (id: string) => Promise<void>;
    onUpdate: (id: string, name: string, role: string) => Promise<void>;
}

const AdminsView: React.FC<AdminsViewProps> = ({ admins, domains, onCreate, onDelete, onUpdate }) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const isSmallMobile = useMediaQuery(theme.breakpoints.down('sm'));

    // Creating Admin State
    const [showDialog, setShowDialog] = useState(false);
    const [name, setName] = useState('');
    const [username, setUsername] = useState('');
    const [selectedDomain, setSelectedDomain] = useState('');
    const [role, setRole] = useState<'super_admin' | 'admin' | 'policy_admin'>('admin');
    const [error, setError] = useState<string | null>(null);

    // Success Dialog State
    const [generatedPassword, setGeneratedPassword] = useState<string | null>(null);
    const [createdAdminEmail, setCreatedAdminEmail] = useState('');

    useEffect(() => {
        if (showDialog && domains.length > 0 && !selectedDomain) {
            setSelectedDomain(domains[0]);
        }
    }, [showDialog, domains, selectedDomain]);

    // Edit Name Dialog State
    const [editDialog, setEditDialog] = useState<{ open: boolean; adminId: string; name: string; role: 'super_admin' | 'admin' | 'policy_admin' }>({
        open: false,
        adminId: '',
        name: '',
        role: 'admin'
    });

    // Delete Confirmation State
    const [deleteDialog, setDeleteDialog] = useState<{ open: boolean; adminId: string; name: string }>({
        open: false,
        adminId: '',
        name: ''
    });

    const handleOpenEdit = (admin: Admin) => {
        setEditDialog({ open: true, adminId: admin.id, name: admin.name, role: admin.role });
    };

    const handleSaveEdit = async () => {
        if (editDialog.name.trim()) {
            await onUpdate(editDialog.adminId, editDialog.name, editDialog.role);
            setEditDialog({ ...editDialog, open: false });
        }
    };

    const handleOpenDelete = (admin: Admin) => {
        setDeleteDialog({ open: true, adminId: admin.id, name: admin.name });
    };

    const handleConfirmDelete = async () => {
        await onDelete(deleteDialog.adminId);
        setDeleteDialog({ ...deleteDialog, open: false });
    };

    const handleCreate = async () => {
        if (!name || !username || !selectedDomain) return;

        if (username.toLowerCase() === 'admin') {
            setError("The username 'admin' is reserved and cannot be deleted or created manually.");
            return;
        }

        const email = `${username}@${selectedDomain}`;
        const newAdmin = { name, email, role }; // No password sent

        const password = await onCreate(newAdmin);

        if (password) {
            setGeneratedPassword(password);
            setCreatedAdminEmail(email);
            setShowDialog(false);
        }
    };

    const handleCloseDialog = () => {
        setShowDialog(false);
        setName('');
        setUsername('');
        setRole('admin');
        setError(null);
    };

    const handleCloseSuccess = () => {
        setGeneratedPassword(null);
        setCreatedAdminEmail('');
        setName('');
        setUsername('');
        setRole('admin');
    };

    const stringToColor = (string: string) => {
        let hash = 0;
        let i;
        for (i = 0; i < string.length; i += 1) {
            hash = string.charCodeAt(i) + ((hash << 5) - hash);
        }
        let color = '#';
        for (i = 0; i < 3; i += 1) {
            const value = (hash >> (i * 8)) & 0xff;
            color += `00${value.toString(16)}`.slice(-2);
        }
        return color;
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant={isSmallMobile ? "h5" : "h4"} sx={{ fontWeight: 800, color: '#202124' }}>Console Administrators</Typography>
                    <Typography color="text.secondary" sx={{ fontSize: isSmallMobile ? '0.85rem' : '1rem' }}>Manage administrators for this Tenant.</Typography>
                </Box>
                <Button
                    variant="contained"
                    disableElevation
                    size={isSmallMobile ? "small" : "medium"}
                    startIcon={<AddIcon />}
                    onClick={() => setShowDialog(true)}
                    sx={{ borderRadius: 2, px: isSmallMobile ? 2 : 3, bgcolor: '#1a73e8', '&:hover': { bgcolor: '#1765cc' } }}
                >
                    {isSmallMobile ? "Add" : "Add Administrator"}
                </Button>
            </Box>

            {admins.length === 0 ? (
                <Paper variant="outlined" sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: '#fff', border: '1px dashed #dadce0' }}>
                    <SupervisorIcon sx={{ fontSize: 48, color: '#dadce0', mb: 2 }} />
                    <Typography variant="h6" sx={{ fontWeight: 600, color: '#3c4043' }}>No administrators found</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 4, maxWidth: 400, mx: 'auto' }}>
                        Add administrators to help manage your organization's security policies and settings.
                    </Typography>
                    <Button variant="outlined" startIcon={<AddIcon />} onClick={() => setShowDialog(true)} sx={{ borderRadius: 2 }}>
                        Add First Administrator
                    </Button>
                </Paper>
            ) : (
                <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden', border: '1px solid #dadce0' }}>
                    <Table>
                        <TableHead sx={{ bgcolor: '#f8f9fa' }}>
                            <TableRow>
                                <TableCell sx={{ fontWeight: 600, color: '#5f6368', py: 2 }}>Name</TableCell>
                                {!isSmallMobile && <TableCell sx={{ fontWeight: 600, color: '#5f6368', py: 2 }}>Email / Username</TableCell>}
                                {!isMobile && <TableCell sx={{ fontWeight: 600, color: '#5f6368', py: 2 }}>Role</TableCell>}
                                <TableCell align="right" sx={{ fontWeight: 600, color: '#5f6368', py: 2 }}>Actions</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {admins.map((admin) => (
                                <TableRow key={admin.id} sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>

                                    <TableCell>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                            <Avatar sx={{ bgcolor: stringToColor(admin.name), width: 32, height: 32, fontSize: '0.875rem' }}>
                                                {admin.name.charAt(0).toUpperCase()}
                                            </Avatar>
                                            <Box>
                                                <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#202124' }}>
                                                    {admin.name}
                                                </Typography>
                                                {isSmallMobile && (
                                                    <Typography variant="caption" display="block" color="text.secondary">{admin.email}</Typography>
                                                )}
                                            </Box>
                                        </Box>
                                    </TableCell>
                                    {!isSmallMobile && (
                                        <TableCell>
                                            <Typography variant="body2" sx={{ color: '#3c4043' }}>{admin.email}</Typography>
                                        </TableCell>
                                    )}
                                    {!isMobile && (
                                        <TableCell>
                                            <Chip
                                                icon={<SecurityIcon sx={{ fontSize: '14px !important' }} />}
                                                label={admin.role === 'super_admin' ? 'Super Admin' : admin.role === 'admin' ? 'Administrator' : 'Policy Admin'}
                                                size="small"
                                                sx={{
                                                    height: 24,
                                                    borderRadius: 1.5,
                                                    bgcolor: admin.role === 'super_admin' ? '#e8f0fe' : admin.role === 'admin' ? '#fce8e6' : '#e6f4ea',
                                                    color: admin.role === 'super_admin' ? '#1a73e8' : admin.role === 'admin' ? '#d93025' : '#1e8e3e',
                                                    fontWeight: 600,
                                                    '& .MuiChip-icon': { color: 'inherit' }
                                                }}
                                            />
                                        </TableCell>
                                    )}
                                    <TableCell align="right">
                                        <Tooltip title="Edit Name">
                                            <IconButton size="small" onClick={() => handleOpenEdit(admin)} sx={{ color: '#5f6368' }}>
                                                <EditIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="Delete Administrator">
                                            <span>
                                                <IconButton
                                                    size="small"
                                                    onClick={() => handleOpenDelete(admin)}
                                                    color="error"
                                                    disabled={admin.email.startsWith('admin@')}
                                                    sx={{ ml: 1, opacity: admin.email.startsWith('admin@') ? 0.3 : 1 }}
                                                >
                                                    <DeleteIcon fontSize="small" />
                                                </IconButton>
                                            </span>
                                        </Tooltip>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </Paper>
            )}

            {/* Create Admin Dialog */}
            <Dialog
                open={showDialog}
                fullScreen={isMobile}
                onClose={(_, reason) => {
                    if (reason !== 'backdropClick') handleCloseDialog();
                }}
                maxWidth="sm"
                fullWidth
                PaperProps={{ sx: { borderRadius: isMobile ? 0 : 3 } }}
            >
                <DialogTitle sx={{ pb: 1 }}>
                    <Typography variant="h6" sx={{ fontWeight: 700 }}>Add Administrator</Typography>
                </DialogTitle>
                <DialogContent sx={{ pb: 3 }}>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                        Create a new administrator account for this tenant console.
                    </Typography>

                    <TextField
                        autoFocus
                        label="Full Name"
                        fullWidth
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="e.g. John Doe"
                        sx={{ mb: 3, '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                    />

                    <Grid container spacing={2} sx={{ mb: 3 }}>
                        <Grid size={{ xs: 12, sm: 6 }}>
                            <TextField
                                label="Username"
                                fullWidth
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                placeholder="john.doe"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            />
                        </Grid>
                        <Grid size={{ xs: 12, sm: 6 }}>
                            <FormControl fullWidth>
                                <InputLabel>Domain</InputLabel>
                                <Select
                                    value={selectedDomain}
                                    label="Domain"
                                    onChange={(e) => setSelectedDomain(e.target.value)}
                                    sx={{ borderRadius: 2 }}
                                >
                                    {domains.map((d) => (
                                        <MenuItem key={d} value={d}>
                                            {d}
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                        </Grid>
                    </Grid>

                    <FormControl fullWidth sx={{ mb: 3 }}>
                        <InputLabel>Role</InputLabel>
                        <Select
                            value={role}
                            label="Role"
                            onChange={(e) => setRole(e.target.value as any)}
                            sx={{ borderRadius: 2 }}
                        >
                            <MenuItem value="super_admin">Super Administrator</MenuItem>
                            <MenuItem value="admin">Administrator (Gateway, App, Policy)</MenuItem>
                            <MenuItem value="policy_admin">Policy Administrator (Policy Only)</MenuItem>
                        </Select>
                    </FormControl>

                    <Box sx={{ mt: 2, p: 2, bgcolor: '#f8f9fa', borderRadius: 2, display: 'flex', alignItems: 'center', gap: 1.5 }}>
                        <PersonIcon sx={{ color: '#5f6368' }} />
                        <Box>
                            <Typography variant="caption" color="text.secondary" display="block">FINAL EMAIL ADDRESS</Typography>
                            <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 500 }}>
                                {username && selectedDomain ? `${username}@${selectedDomain}` : 'Enter username & domain'}
                            </Typography>
                        </Box>
                    </Box>

                    {error && (
                        <Alert severity="error" sx={{ mt: 2, borderRadius: 2 }}>
                            {error}
                        </Alert>
                    )}

                    <Alert severity="info" Icon={<WarningIcon fontSize="inherit" />} sx={{ mt: 3, borderRadius: 2, '& .MuiAlert-icon': { color: '#1a73e8' }, bgcolor: '#e8f0fe', color: '#174ea6' }}>
                        A secure 12-character password will be automatically generated. You must copy it immediately after creation.
                    </Alert>
                </DialogContent>
                <DialogActions sx={{ px: 3, pb: 3 }}>
                    <Button onClick={handleCloseDialog} sx={{ color: '#5f6368', fontWeight: 600 }}>Cancel</Button>
                    <Button
                        onClick={handleCreate}
                        variant="contained"
                        disableElevation
                        disabled={!name || !username || !selectedDomain}
                        sx={{ bgcolor: '#1a73e8', borderRadius: 2, px: 3, fontWeight: 600 }}
                    >
                        Create Account
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Success Dialog (Password Display) */}
            <Dialog
                open={!!generatedPassword}
                fullScreen={isMobile}
                onClose={(_, reason) => {
                    if (reason !== 'backdropClick') handleCloseSuccess();
                }}
                maxWidth="sm"
                fullWidth
                PaperProps={{ sx: { borderRadius: isMobile ? 0 : 3 } }}
            >
                <DialogTitle sx={{ textAlign: 'center', pt: 4 }}>
                    <Box sx={{
                        width: 64, height: 64, borderRadius: '50%', bgcolor: '#e6f4ea', color: '#1e8e3e',
                        display: 'flex', alignItems: 'center', justifyContent: 'center', mx: 'auto', mb: 2
                    }}>
                        <AddIcon fontSize="large" />
                    </Box>
                    <Typography variant="h5" sx={{ fontWeight: 700 }}>Administrator Created</Typography>
                </DialogTitle>
                <DialogContent>
                    <Typography variant="body1" align="center" color="text.secondary" sx={{ mb: 3 }}>
                        Account for <strong>{createdAdminEmail}</strong> is ready.
                    </Typography>

                    <Paper variant="outlined" sx={{ p: 3, bgcolor: '#f8f9fa', borderRadius: 2, position: 'relative', border: '1px dashed #dadce0', mx: 2 }}>
                        <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1, fontWeight: 700, letterSpacing: 0.5 }}>
                            AUTO-GENERATED PASSWORD
                        </Typography>
                        <Stack direction="row" alignItems="center" justifyContent="space-between">
                            <Typography variant="h4" sx={{ fontFamily: 'monospace', fontWeight: 700, letterSpacing: 1, color: '#202124' }}>
                                {generatedPassword}
                            </Typography>
                            <Tooltip title="Copy Password">
                                <IconButton
                                    onClick={() => generatedPassword && navigator.clipboard.writeText(generatedPassword)}
                                    sx={{ color: '#1a73e8' }}
                                >
                                    <CopyIcon />
                                </IconButton>
                            </Tooltip>
                        </Stack>
                    </Paper>

                    <Alert severity="warning" sx={{ mt: 3, mx: 2, borderRadius: 2 }}>
                        <strong>Important:</strong> Copy this password now. It will not be shown again.
                    </Alert>
                </DialogContent>
                <DialogActions sx={{ justifyContent: 'center', pb: 4 }}>
                    <Button onClick={handleCloseSuccess} variant="contained" disableElevation sx={{ borderRadius: 2, px: 4, bgcolor: '#1a73e8', fontWeight: 600 }}>
                        I have copied the password
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Edit Name Dialog */}
            <Dialog
                open={editDialog.open}
                fullScreen={isMobile}
                onClose={() => setEditDialog({ ...editDialog, open: false })}
                fullWidth
                maxWidth="xs"
                PaperProps={{ sx: { borderRadius: isMobile ? 0 : 3 } }}
            >
                <DialogTitle sx={{ fontWeight: 800 }}>Edit Name</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="Full Name"
                        fullWidth
                        variant="outlined"
                        value={editDialog.name}
                        onChange={(e) => setEditDialog({ ...editDialog, name: e.target.value })}
                        sx={{ mt: 1, mb: 3, '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                    />
                    <FormControl fullWidth>
                        <InputLabel>Role</InputLabel>
                        <Select
                            value={editDialog.role}
                            label="Role"
                            onChange={(e) => setEditDialog({ ...editDialog, role: e.target.value as any })}
                            sx={{ borderRadius: 2 }}
                        >
                            <MenuItem value="super_admin">Super Administrator</MenuItem>
                            <MenuItem value="admin">Administrator (Proxy, App, Policy)</MenuItem>
                            <MenuItem value="policy_admin">Policy Administrator (Policy Only)</MenuItem>
                        </Select>
                    </FormControl>
                </DialogContent>
                <DialogActions sx={{ p: 2 }}>
                    <Button onClick={() => setEditDialog({ ...editDialog, open: false })} sx={{ fontWeight: 600, color: '#5f6368' }}>Cancel</Button>
                    <Button onClick={handleSaveEdit} variant="contained" disabled={!editDialog.name.trim()} sx={{ borderRadius: 2, fontWeight: 700 }}>Save</Button>
                </DialogActions>
            </Dialog>

            {/* Delete Confirmation Dialog */}
            <Dialog
                open={deleteDialog.open}
                fullScreen={isMobile}
                onClose={() => setDeleteDialog({ ...deleteDialog, open: false })}
                fullWidth
                maxWidth="xs"
                PaperProps={{ sx: { borderRadius: isMobile ? 0 : 3 } }}
            >
                <DialogTitle sx={{ fontWeight: 800, color: '#d93025' }}>Delete Administrator?</DialogTitle>
                <DialogContent>
                    <Typography>
                        Are you sure you want to delete <strong>{deleteDialog.name}</strong>? This action cannot be undone.
                    </Typography>
                </DialogContent>
                <DialogActions sx={{ p: 2 }}>
                    <Button onClick={() => setDeleteDialog({ ...deleteDialog, open: false })} sx={{ fontWeight: 600, color: '#5f6368' }}>Cancel</Button>
                    <Button onClick={handleConfirmDelete} variant="contained" color="error" sx={{ borderRadius: 2, fontWeight: 700 }}>Delete</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default AdminsView;
