import React from 'react';
import { Box, Typography, Grid, Card, CardContent } from '@mui/material';
import { Language as LanguageIcon, Google as GoogleIcon, VerifiedUser as VerifiedUserIcon } from '@mui/icons-material';
import { Tenant } from '../../types';

interface DashboardViewProps {
    tenant: Tenant | null;
}

const DashboardView: React.FC<DashboardViewProps> = ({ tenant }) => {
    return (
        <Box>
            <Box sx={{ mb: 4, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 700, color: '#202124', mb: 1 }}>
                        Overview
                    </Typography>
                    <Typography color="textSecondary" variant="body1">
                        Welcome to the specific management console for {tenant?.name || 'your organization'}.
                    </Typography>
                </Box>
            </Box>

            <Grid container spacing={3}>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Card sx={{ height: '100%', position: 'relative', overflow: 'visible' }}>
                        <Box sx={{
                            position: 'absolute',
                            top: -20,
                            left: 20,
                            width: 64,
                            height: 64,
                            bgcolor: '#e8f0fe',
                            borderRadius: '16px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
                        }}>
                            <LanguageIcon sx={{ fontSize: 32, color: '#1a73e8' }} />
                        </Box>
                        <CardContent sx={{ pt: 6 }}>
                            <Typography color="textSecondary" variant="overline" sx={{ fontWeight: 600, letterSpacing: 1 }}>
                                PRIMARY DOMAIN
                            </Typography>
                            <Typography variant="h5" sx={{ mt: 1, fontWeight: 500 }}>
                                {tenant?.primary_domain}
                            </Typography>
                            <Typography variant="body2" color="success.main" sx={{ mt: 2, display: 'flex', alignItems: 'center' }}>
                                <VerifiedUserIcon sx={{ fontSize: 16, mr: 0.5 }} /> Verified & Active
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>

                <Grid size={{ xs: 12, md: 4 }}>
                    <Card sx={{ height: '100%', position: 'relative', overflow: 'visible' }}>
                        <Box sx={{
                            position: 'absolute',
                            top: -20,
                            left: 20,
                            width: 64,
                            height: 64,
                            bgcolor: '#fce8e6',
                            borderRadius: '16px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
                        }}>
                            <GoogleIcon sx={{ fontSize: 32, color: '#d93025' }} />
                        </Box>
                        <CardContent sx={{ pt: 6 }}>
                            <Typography color="textSecondary" variant="overline" sx={{ fontWeight: 600, letterSpacing: 1 }}>
                                IDENTITY PROVIDER
                            </Typography>
                            <Typography variant="h5" sx={{ mt: 1, fontWeight: 500 }}>
                                Google Workspace
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ mt: 2 }}>
                                Integration active
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>

                <Grid size={{ xs: 12, md: 4 }}>
                    <Card sx={{ height: '100%', position: 'relative', overflow: 'visible' }}>
                        <Box sx={{
                            position: 'absolute',
                            top: -20,
                            left: 20,
                            width: 64,
                            height: 64,
                            bgcolor: '#e6f4ea',
                            borderRadius: '16px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
                        }}>
                            <VerifiedUserIcon sx={{ fontSize: 32, color: '#188038' }} />
                        </Box>
                        <CardContent sx={{ pt: 6 }}>
                            <Typography color="textSecondary" variant="overline" sx={{ fontWeight: 600, letterSpacing: 1 }}>
                                ACTIVE SESSIONS
                            </Typography>
                            <Typography variant="h5" sx={{ mt: 1, fontWeight: 500 }}>
                                0
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ mt: 2 }}>
                                No active users connected
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>
            </Grid>
        </Box>
    );
};

export default DashboardView;
