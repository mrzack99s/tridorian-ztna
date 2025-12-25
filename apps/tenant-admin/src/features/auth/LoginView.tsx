import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, TextField, Button, ThemeProvider } from '@mui/material';
import { CloudQueue as CloudIcon } from '@mui/icons-material';
import { theme } from '../../theme/theme';

interface LoginViewProps {
    onLogin: (email: string, password: string) => Promise<void>;
}

const LoginView: React.FC<LoginViewProps> = ({ onLogin }) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onLogin(email, password);
    };

    return (
        <ThemeProvider theme={theme}>
            <Box sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', bgcolor: '#f1f3f4' }}>
                <Card sx={{ maxWidth: 400, width: '100%', p: 2 }}>
                    <CardContent>
                        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mb: 3 }}>
                            <CloudIcon sx={{ color: '#1a73e8', fontSize: 48, mb: 1 }} />
                            <Typography variant="h5" sx={{ fontWeight: 700 }}>Tenant Admin</Typography>
                            <Typography variant="body2" color="text.secondary">Sign in to manage your organization</Typography>
                        </Box>
                        <form onSubmit={handleSubmit}>
                            <TextField fullWidth label="Email" margin="normal" value={email} onChange={e => setEmail(e.target.value)} required />
                            <TextField fullWidth label="Password" type="password" margin="normal" value={password} onChange={e => setPassword(e.target.value)} required />
                            <Button fullWidth variant="contained" type="submit" sx={{ mt: 3, py: 1.5 }}>Sign In</Button>
                        </form>
                    </CardContent>
                </Card>
            </Box>
        </ThemeProvider>
    );
};

export default LoginView;
