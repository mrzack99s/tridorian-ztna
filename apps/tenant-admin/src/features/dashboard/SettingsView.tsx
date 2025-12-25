import React, { useState, useEffect, useRef } from 'react';
import {
    Box, Typography, Card, CardContent, Button, Divider, TextField, Grid,
    Alert, Fade, IconButton, Paper, CircularProgress, Dialog, DialogTitle,
    DialogContent, DialogActions, Chip
} from '@mui/material';
import {
    Logout as LogoutIcon, Settings as SettingsIcon, Google as GoogleIcon,
    Save as SaveIcon, Edit as EditIcon, Language as LanguageIcon,
    CheckCircle as CheckCircleIcon, ErrorOutline as ErrorOutlineIcon,
    Refresh as RefreshIcon, Upload as UploadIcon
} from '@mui/icons-material';
import { Tenant } from '../../types';

interface SettingsViewProps {
    tenant: Tenant | null;
    onRefresh: () => void;
}

const SettingsView: React.FC<SettingsViewProps> = ({ tenant, onRefresh }) => {
    // Edit Modes
    const [editIDP, setEditIDP] = useState(false);
    const [showDomainDialog, setShowDomainDialog] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    // IDP State
    const [idpConfig, setIdpConfig] = useState({
        clientID: '',
        clientSecret: '',
        saKey: '',
        adminEmail: ''
    });

    // New Domain State
    const [newDomain, setNewDomain] = useState(tenant?.primary_domain || '');

    useEffect(() => {
        if (tenant?.primary_domain) {
            setNewDomain(tenant.primary_domain);
        }
    }, [tenant?.primary_domain]);
    const [verificationStep, setVerificationStep] = useState<'idle' | 'pending'>('idle');
    const [verificationToken, setVerificationToken] = useState('');
    const [customDomainID, setCustomDomainID] = useState('');

    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    useEffect(() => {
        if (tenant) {
            setIdpConfig({
                clientID: tenant.google_client_id || '',
                clientSecret: '',
                saKey: '',
                adminEmail: tenant.google_admin_email || ''
            });
        }
    }, [tenant]);

    const handleLogout = async () => {
        await fetch('/auth/mgmt/logout', { method: 'POST' });
        window.location.reload();
    };

    const handleUploadKey = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (event) => {
                const content = event.target?.result as string;
                setIdpConfig(prev => ({ ...prev, saKey: content }));
            };
            reader.readAsText(file);
        }
    };

    const handleUpdateIdentity = async () => {
        setLoading(true);
        setSuccess(false);
        try {
            const res = await fetch('/api/v1/tenants/identity', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    google_client_id: idpConfig.clientID,
                    google_client_secret: idpConfig.clientSecret,
                    google_service_account_key: idpConfig.saKey,
                    google_admin_email: idpConfig.adminEmail
                })
            });
            const data = await res.json();
            if (data.success) {
                setSuccess(true);
                setEditIDP(false);
                onRefresh();
                setIdpConfig(prev => ({ ...prev, clientSecret: '', saKey: '' }));
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    // Domain Handlers
    const handleDomainSubmit = async () => {
        const domainToRegister = newDomain.toLowerCase().trim();
        const freeSuffix = '.devztna.rattanaburi.ac.th';
        if (domainToRegister.endsWith(freeSuffix)) {
            const expectedDomain = `${tenant?.slug}${freeSuffix}`;
            if (domainToRegister !== expectedDomain) {
                alert(`Invalid subdomain. For free domains, you must use your organization slug: ${expectedDomain}`);
                return;
            }
        }

        setLoading(true);
        try {
            const res = await fetch('/api/v1/tenants/domains', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain: newDomain })
            });
            const data = await res.json();
            if (data.success) {
                if (data.data.is_verified) {
                    await activateDomain(data.data.domain);
                } else {
                    setVerificationToken(data.data.verification_token);
                    setCustomDomainID(data.data.id);
                    setVerificationStep('pending');
                }
            } else {
                alert('Error: ' + (data.message || data.error));
            }
        } catch (err) { console.error(err); }
        finally { setLoading(false); }
    };

    const handleVerifyClick = async () => {
        setLoading(true);
        try {
            const res = await fetch('/api/v1/tenants/domains/verify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain_id: customDomainID })
            });
            const data = await res.json();
            if (data.success) {
                await activateDomain(newDomain.toLowerCase().trim());
            } else {
                alert('Verification Failed: ' + data.message);
            }
        } catch (err) { console.error(err); }
        finally { setLoading(false); }
    };

    const activateDomain = async (domainToActivate: string) => {
        try {
            const res = await fetch('/api/v1/tenants/activate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain: domainToActivate })
            });
            const data = await res.json();
            if (data.success) {
                onRefresh();
                setShowDomainDialog(false);
                setVerificationStep('idle');
                setNewDomain('');
            } else {
                alert('Activation Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };

    return (
        <Box sx={{ maxWidth: 1200, mx: 'auto', py: 2 }}>
            <Box sx={{ mb: 6 }}>
                <Typography variant="h4" sx={{ fontWeight: 800, letterSpacing: '-0.02em', mb: 1 }}>Settings</Typography>
                <Typography color="text.secondary">Manage your organization's core configuration and security policies.</Typography>
            </Box>

            <Divider sx={{ mb: 5 }} />

            {/* Section: Organization Details */}
            <Grid container spacing={4} sx={{ mb: 8 }}>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>Organization</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        Basic identity and identification for your tenant within the Tridorian network.
                    </Typography>
                </Grid>
                <Grid size={{ xs: 12, md: 8 }}>
                    <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
                        <Box sx={{ p: 3, borderBottom: '1px solid #f0f0f0' }}>
                            <Grid container spacing={2}>
                                <Grid size={{ xs: 12, sm: 6 }}>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Organization Name</Typography>
                                    <Typography variant="body1" sx={{ fontWeight: 500, mt: 0.5 }}>{tenant?.name}</Typography>
                                </Grid>
                                <Grid size={{ xs: 12, sm: 6 }}>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Unique Identifier (Slug)</Typography>
                                    <Typography variant="body1" sx={{ fontFamily: 'monospace', color: 'primary.main', fontWeight: 600, mt: 0.5 }}>{tenant?.slug}</Typography>
                                </Grid>
                            </Grid>
                        </Box>
                        <Box sx={{ p: 3, bgcolor: '#fafafa' }}>
                            <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                                <Box>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Primary Domain</Typography>
                                    <Box sx={{ display: 'flex', alignItems: 'center', mt: 0.5, gap: 1 }}>
                                        <Typography variant="h6" sx={{ fontWeight: 700, letterSpacing: '-0.01em' }}>{tenant?.primary_domain}</Typography>
                                        <Chip
                                            icon={<CheckCircleIcon style={{ fontSize: 16 }} />}
                                            label="Verified"
                                            size="small"
                                            color="success"
                                            variant="outlined"
                                            sx={{ height: 24, fontWeight: 600, bgcolor: 'success.50' }}
                                        />
                                    </Box>
                                </Box>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    startIcon={<RefreshIcon />}
                                    onClick={() => setShowDomainDialog(true)}
                                    sx={{ borderRadius: 2, textTransform: 'none', fontWeight: 600 }}
                                >
                                    Update Domain
                                </Button>
                            </Box>
                        </Box>
                    </Paper>
                </Grid>
            </Grid>

            {/* Section: Identity Provider */}
            <Grid container spacing={4} sx={{ mb: 8 }}>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>Identity Provider</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        Configure Google Workspace to enable Single Sign-On (SSO) and automated user directory synchronization.
                    </Typography>
                    <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
                        <GoogleIcon sx={{ color: '#ea4335', fontSize: 20 }} />
                        <Typography variant="caption" sx={{ fontWeight: 600 }}>Google Workspace Integrated</Typography>
                    </Box>
                </Grid>
                <Grid size={{ xs: 12, md: 8 }}>
                    <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
                        <Box sx={{ p: 3, borderBottom: editIDP ? '1px solid #f0f0f0' : 'none' }}>
                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: editIDP ? 3 : 0 }}>
                                <Box>
                                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>Google OIDC Configuration</Typography>
                                    <Typography variant="body2" color="text.secondary">Manage OAuth 2.0 and Service Account credentials.</Typography>
                                </Box>
                                {!editIDP ? (
                                    <Button
                                        variant="contained"
                                        startIcon={<EditIcon />}
                                        size="small"
                                        onClick={() => setEditIDP(true)}
                                        sx={{ borderRadius: 2, textTransform: 'none', fontWeight: 600, boxShadow: 'none' }}
                                    >
                                        Edit Config
                                    </Button>
                                ) : (
                                    <Button
                                        variant="text"
                                        color="inherit"
                                        size="small"
                                        onClick={() => setEditIDP(false)}
                                        sx={{ textTransform: 'none' }}
                                    >
                                        Cancel
                                    </Button>
                                )}
                            </Box>

                            {success && !editIDP && (
                                <Fade in>
                                    <Alert severity="success" sx={{ mt: 2, borderRadius: 2 }}>Configuration updated successfully.</Alert>
                                </Fade>
                            )}

                            {!editIDP ? (
                                <Box sx={{ mt: 3, bgcolor: '#f8f9fa', p: 2, borderRadius: 2 }}>
                                    <Grid container spacing={2}>
                                        <Grid size={{ xs: 12, sm: 8 }}>
                                            <Typography variant="caption" sx={{ color: 'text.secondary', fontWeight: 700 }}>Client ID</Typography>
                                            <Typography variant="body2" sx={{ fontFamily: 'monospace', color: 'text.primary', mt: 0.5, wordBreak: 'break-all' }}>
                                                {tenant?.google_client_id || 'Not configured'}
                                            </Typography>
                                        </Grid>
                                        <Grid size={{ xs: 12, sm: 4 }}>
                                            <Typography variant="caption" sx={{ color: 'text.secondary', fontWeight: 700 }}>Admin Email</Typography>
                                            <Typography variant="body2" sx={{ mt: 0.5 }}>{tenant?.google_admin_email || 'Not configured'}</Typography>
                                        </Grid>
                                    </Grid>
                                </Box>
                            ) : (
                                <Fade in>
                                    <Box>
                                        <Grid container spacing={3}>
                                            <Grid size={12}>
                                                <TextField
                                                    fullWidth
                                                    label="Google Client ID"
                                                    value={idpConfig.clientID}
                                                    onChange={e => setIdpConfig({ ...idpConfig, clientID: e.target.value })}
                                                    placeholder="000000000-xxxxx.apps.googleusercontent.com"
                                                />
                                            </Grid>
                                            <Grid size={12}>
                                                <TextField
                                                    fullWidth
                                                    label="Google Client Secret"
                                                    type="password"
                                                    value={idpConfig.clientSecret}
                                                    onChange={e => setIdpConfig({ ...idpConfig, clientSecret: e.target.value })}
                                                    placeholder="••••••••"
                                                    helperText="Leave blank to keep existing"
                                                />
                                            </Grid>
                                            <Grid size={12}>
                                                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                                                    <Typography variant="body2" sx={{ fontWeight: 600 }}>Service Account Key (JSON)</Typography>
                                                    <Button
                                                        size="small"
                                                        startIcon={<UploadIcon />}
                                                        onClick={() => fileInputRef.current?.click()}
                                                        sx={{ textTransform: 'none' }}
                                                    >
                                                        Upload JSON
                                                    </Button>
                                                </Box>
                                                <input
                                                    type="file"
                                                    ref={fileInputRef}
                                                    style={{ display: 'none' }}
                                                    accept=".json"
                                                    onChange={handleUploadKey}
                                                />
                                                <TextField
                                                    fullWidth
                                                    multiline
                                                    rows={4}
                                                    value={idpConfig.saKey}
                                                    onChange={e => setIdpConfig({ ...idpConfig, saKey: e.target.value })}
                                                    placeholder="{ ... }"
                                                    sx={{ '& .MuiInputBase-input': { fontFamily: 'monospace', fontSize: 13 } }}
                                                    helperText="Required for directory sync"
                                                />
                                            </Grid>
                                            <Grid size={12}>
                                                <TextField
                                                    fullWidth
                                                    label="Google Admin User Email"
                                                    value={idpConfig.adminEmail}
                                                    onChange={e => setIdpConfig({ ...idpConfig, adminEmail: e.target.value })}
                                                    placeholder="admin@yourcompany.com"
                                                />
                                            </Grid>
                                        </Grid>
                                        <Box sx={{ mt: 4, display: 'flex', justifyContent: 'flex-end', gap: 2 }}>
                                            <Button
                                                variant="contained"
                                                startIcon={loading ? <CircularProgress size={18} color="inherit" /> : <SaveIcon />}
                                                disabled={loading}
                                                onClick={handleUpdateIdentity}
                                                sx={{ borderRadius: 2, px: 4, py: 1 }}
                                            >
                                                Save Identity Config
                                            </Button>
                                        </Box>
                                    </Box>
                                </Fade>
                            )}
                        </Box>
                    </Paper>
                </Grid>
            </Grid>

            {/* Section: Danger Zone */}
            <Grid container spacing={4} sx={{ mb: 8 }}>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Typography variant="h6" sx={{ fontWeight: 700, color: 'error.main', mb: 1 }}>Danger Zone</Typography>
                    <Typography variant="body2" color="text.secondary">
                        Critical actions that impact your console session and workspace accessibility.
                    </Typography>
                </Grid>
                <Grid size={{ xs: 12, md: 8 }}>
                    <Paper variant="outlined" sx={{ borderRadius: 3, borderColor: 'error.light', bgcolor: 'rgba(211, 47, 47, 0.02)' }}>
                        <Box sx={{ p: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <Box>
                                <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>Logout from Management Console</Typography>
                                <Typography variant="body2" color="text.secondary">This will terminate your current administrative session.</Typography>
                            </Box>
                            <Button
                                variant="outlined"
                                color="error"
                                startIcon={<LogoutIcon />}
                                onClick={handleLogout}
                                sx={{ borderRadius: 2, textTransform: 'none', fontWeight: 600 }}
                            >
                                Sign Out
                            </Button>
                        </Box>
                    </Paper>
                </Grid>
            </Grid>

            {/* Domain Dialog */}
            <Dialog open={showDomainDialog} onClose={() => setShowDomainDialog(false)} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 3, p: 1 } }}>
                <DialogTitle sx={{ fontWeight: 800 }}>Update Primary Domain</DialogTitle>
                <DialogContent>
                    {verificationStep === 'idle' ? (
                        <>
                            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                                Changing your domain will update the endpoint used by users and nodes. Authentication flows may be affected.
                            </Typography>
                            {(() => {
                                const freeSuffix = '.devztna.rattanaburi.ac.th';
                                const expectedFreeDomain = `${tenant?.slug}${freeSuffix}`;
                                const isFreeDomainSuffix = newDomain.toLowerCase().endsWith(freeSuffix);
                                const isInvalidFreeDomain = isFreeDomainSuffix && newDomain.toLowerCase() !== expectedFreeDomain;
                                const isValidFreeDomain = isFreeDomainSuffix && newDomain.toLowerCase() === expectedFreeDomain;

                                return (
                                    <>
                                        <TextField
                                            fullWidth
                                            label="New Domain"
                                            placeholder="ztna.example.com"
                                            value={newDomain}
                                            onChange={(e) => setNewDomain(e.target.value)}
                                            disabled={loading}
                                            autoFocus
                                            error={isInvalidFreeDomain}
                                            helperText={isInvalidFreeDomain ? `For onzt.tridorian.com, you must use your organization slug: ${expectedFreeDomain}` : ""}
                                        />
                                        <Box sx={{ mt: 1 }}>
                                            <Button
                                                size="small"
                                                onClick={() => setNewDomain(expectedFreeDomain)}
                                                sx={{ textTransform: 'none' }}
                                            >
                                                Use onzt.tridorian.com: {expectedFreeDomain}
                                            </Button>
                                        </Box>
                                        {isValidFreeDomain && (
                                            <Typography variant="caption" color="success.main" sx={{ display: 'block', mt: 2, fontWeight: 600 }}>
                                                ✓ This onzt.tridorian.com domain is pre-verified and active immediately.
                                            </Typography>
                                        )}
                                    </>
                                );
                            })()}
                        </>
                    ) : (
                        <Box>
                            <Alert severity="warning" variant="outlined" sx={{ mb: 3, borderRadius: 2 }}>
                                <strong>DNS Verification Required</strong>
                                <br />
                                Add the following TXT record to your DNS provider to prove ownership of <strong>{newDomain}</strong>.
                            </Alert>
                            <Paper variant="outlined" sx={{ p: 2, bgcolor: '#f8f9fa', borderRadius: 2, mb: 1 }}>
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, display: 'block', mb: 0.5 }}>NAME / HOST</Typography>
                                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 700 }}>_tridorian-challenge.{newDomain}</Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, display: 'block', mb: 0.5 }}>VALUE / CONTENT</Typography>
                                    <Typography variant="body2" sx={{ fontFamily: 'monospace', wordBreak: 'break-all', fontWeight: 700 }}>{verificationToken}</Typography>
                                </Box>
                            </Paper>
                        </Box>
                    )}
                </DialogContent>
                <DialogActions sx={{ p: 3, gap: 1 }}>
                    <Button variant="text" color="inherit" onClick={() => { setShowDomainDialog(false); setVerificationStep('idle'); }} sx={{ borderRadius: 2 }}>
                        Cancel
                    </Button>
                    {verificationStep === 'idle' ? (
                        <Button
                            variant="contained"
                            onClick={handleDomainSubmit}
                            disabled={!newDomain || loading || (newDomain.toLowerCase().endsWith('.devztna.rattanaburi.ac.th') && newDomain.toLowerCase() !== `${tenant?.slug}.devztna.rattanaburi.ac.th`)}
                            sx={{ borderRadius: 2, px: 3 }}
                        >
                            {loading ? <CircularProgress size={20} color="inherit" /> : (newDomain.toLowerCase().endsWith('.devztna.rattanaburi.ac.th') ? 'Update Domain' : 'Request Update')}
                        </Button>
                    ) : (
                        <Button
                            variant="contained"
                            color="primary"
                            onClick={handleVerifyClick}
                            disabled={loading}
                            sx={{ borderRadius: 2, px: 3 }}
                        >
                            {loading ? <CircularProgress size={20} color="inherit" /> : 'Verify DNS Record'}
                        </Button>
                    )}
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default SettingsView;
