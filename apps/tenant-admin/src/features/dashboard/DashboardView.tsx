import React from 'react';
import { Box, Typography, Grid, Card, CardContent } from '@mui/material';
import { Language as LanguageIcon, Google as GoogleIcon } from '@mui/icons-material';
import { Tenant } from '../../types';

interface DashboardViewProps {
    tenant: Tenant | null;
}

const DashboardView: React.FC<DashboardViewProps> = ({ tenant }) => {
    return (
        <Box>
            <Typography variant="h5" sx={{ mb: 4 }}>Dashboard</Typography>
            <Grid container spacing={3}>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Card>
                        <CardContent>
                            <Typography color="textSecondary" gutterBottom>Primary Domain</Typography>
                            <Typography variant="h6" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <LanguageIcon color="primary" /> {tenant?.primary_domain}
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Card>
                        <CardContent>
                            <Typography color="textSecondary" gutterBottom>Identity Provider</Typography>
                            <Typography variant="h6" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <GoogleIcon color="primary" /> Google Workspace
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>
                <Grid size={{ xs: 12, md: 4 }}>
                    <Card>
                        <CardContent>
                            <Typography color="textSecondary" gutterBottom>Active Sessions</Typography>
                            <Typography variant="h6">0</Typography>
                        </CardContent>
                    </Card>
                </Grid>
            </Grid>
        </Box>
    );
};

export default DashboardView;
