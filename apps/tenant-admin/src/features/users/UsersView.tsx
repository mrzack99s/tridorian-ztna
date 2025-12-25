import React from 'react';
import { Box, Typography, Card, Button } from '@mui/material';
import { People as PeopleIcon } from '@mui/icons-material';

const UsersView: React.FC = () => {
    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 4 }}>
                <Typography variant="h5">Users & Groups</Typography>
                <Button variant="contained" startIcon={<PeopleIcon />}>Sync from Google</Button>
            </Box>
            <Card>
                <Box sx={{ p: 4, textAlign: 'center', color: 'text.secondary' }}>
                    <PeopleIcon sx={{ fontSize: 48, mb: 2, opacity: 0.5 }} />
                    <Typography>No users synced yet.</Typography>
                    <Button sx={{ mt: 2 }}>Configure Auto-Sync</Button>
                </Box>
            </Card>
        </Box>
    );
};

export default UsersView;
