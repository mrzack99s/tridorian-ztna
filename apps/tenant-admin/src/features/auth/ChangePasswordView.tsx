import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, TextField, Button, ThemeProvider } from '@mui/material';
import { VpnKey as VpnKeyIcon } from '@mui/icons-material';
import { theme } from '../../theme/theme';

interface ChangePasswordViewProps {
    onChangePassword: (oldPassword: string, newPassword: string) => Promise<void>;
}

const ChangePasswordView: React.FC<ChangePasswordViewProps> = ({ onChangePassword }) => {
    const [password, setPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onChangePassword(password, newPassword);
    };

    return (
        <ThemeProvider theme={theme}>
            <Box sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', bgcolor: '#f1f3f4' }}>
                <Card sx={{ maxWidth: 400, width: '100%', p: 2 }}>
                    <CardContent>
                        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mb: 3 }}>
                            <VpnKeyIcon sx={{ color: '#1a73e8', fontSize: 48, mb: 1 }} />
                            <Typography variant="h5" sx={{ fontWeight: 700 }}>Change Password</Typography>
                            <Typography variant="body2" color="text.secondary" align="center">Security policy requires you to change your password.</Typography>
                        </Box>
                        <form onSubmit={handleSubmit}>
                            <TextField fullWidth label="Current Password" type="password" margin="normal" value={password} onChange={e => setPassword(e.target.value)} required />
                            <TextField fullWidth label="New Password" type="password" margin="normal" value={newPassword} onChange={e => setNewPassword(e.target.value)} required />
                            <Button fullWidth variant="contained" type="submit" sx={{ mt: 3, py: 1.5 }}>Update Password</Button>
                        </form>
                    </CardContent>
                </Card>
            </Box>
        </ThemeProvider>
    );
};

export default ChangePasswordView;
