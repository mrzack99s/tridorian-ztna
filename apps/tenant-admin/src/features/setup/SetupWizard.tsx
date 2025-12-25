import React, { useState, useRef } from 'react';
import {
    Box, Stepper, Step, StepLabel, Typography, Container, Card,
    CardContent, TextField, Button, Alert, Fade, IconButton
} from '@mui/material';
import {
    Language as LanguageIcon,
    Google as GoogleIcon,
    CheckCircle as CheckCircleIcon,
    Upload as UploadIcon
} from '@mui/icons-material';
import { Tenant } from '../../types';

interface SetupWizardProps {
    tenant: Tenant | null;
    onRefresh: () => void;
    onComplete: () => void;
}

const SetupWizard: React.FC<SetupWizardProps> = ({ tenant, onRefresh, onComplete }) => {
    const [step, setStep] = useState(0);
    const [domain, setDomain] = useState('');
    const [verificationStep, setVerificationStep] = useState<'idle' | 'pending'>('idle');
    const [verificationToken, setVerificationToken] = useState('');
    const [customDomainID, setCustomDomainID] = useState('');
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [idpConfig, setIdpConfig] = useState({
        clientID: '',
        clientSecret: '',
        saKey: '',
        adminEmail: ''
    });

    const steps = ['Connect Domain', 'Identity Provider', 'Review & Launch'];

    const handleDomainSubmit = async () => {
        if (domain.endsWith('.devztna.rattanaburi.ac.th')) {
            await activateDomain(domain);
            return;
        }
        try {
            const res = await fetch('/api/v1/tenants/domains', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain })
            });
            const data = await res.json();
            if (data.success) {
                if (data.data.is_verified) {
                    await activateDomain(domain);
                } else {
                    setVerificationToken(data.data.verification_token);
                    setCustomDomainID(data.data.id);
                    setVerificationStep('pending');
                }
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
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
                setStep(1);
            } else {
                alert('Activation Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };

    const handleVerifyClick = async () => {
        try {
            const res = await fetch('/api/v1/tenants/domains/verify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain_id: customDomainID })
            });
            const data = await res.json();
            if (data.success) {
                await activateDomain(domain);
            } else {
                alert('Verification Failed: ' + data.message);
            }
        } catch (err) { console.error(err); }
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
                onRefresh();
                setStep(2);
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };

    return (
        <Box sx={{ minHeight: '100vh', py: 8, bgcolor: '#f1f3f4' }}>
            <Container maxWidth="md">
                <Box sx={{ mb: 6, textAlign: 'center' }}>
                    <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>Welcome to Tridorian ZTNA</Typography>
                    <Typography color="text.secondary">Let's set up your secure access environment</Typography>
                </Box>

                <Stepper activeStep={step} sx={{ mb: 6 }}>
                    {steps.map((label) => <Step key={label}><StepLabel>{label}</StepLabel></Step>)}
                </Stepper>

                {step === 0 && (
                    <Fade in>
                        <Card>
                            <CardContent sx={{ p: 4 }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                                    <LanguageIcon sx={{ color: '#1a73e8', mr: 2, fontSize: 32 }} />
                                    <Box>
                                        <Typography variant="h6">Connect your Domain</Typography>
                                        <Typography variant="body2" color="text.secondary">Use your own domain or our free subdomain.</Typography>
                                    </Box>
                                </Box>

                                {verificationStep === 'idle' ? (
                                    <>
                                        <TextField
                                            fullWidth
                                            label="Your Domain"
                                            placeholder="ztna.yourcompany.com"
                                            value={domain}
                                            onChange={(e) => setDomain(e.target.value)}
                                            sx={{ mb: 2 }}
                                        />
                                        <Button
                                            variant="outlined"
                                            fullWidth
                                            sx={{ mb: 3 }}
                                            onClick={() => setDomain(`${tenant?.slug}.devztna.rattanaburi.ac.th`)}
                                        >
                                            Use free domain: {tenant?.slug}.devztna.rattanaburi.ac.th
                                        </Button>
                                        <Button variant="contained" fullWidth size="large" onClick={handleDomainSubmit}>Next</Button>
                                    </>
                                ) : (
                                    <Box>
                                        <Alert severity="info" sx={{ mb: 3 }}>
                                            To verify your domain, please add the following TXT record to your DNS settings:
                                        </Alert>
                                        <Card variant="outlined" sx={{ bgcolor: '#f8f9fa', mb: 3 }}>
                                            <CardContent>
                                                <Typography variant="caption" color="text.secondary">TYPE: TXT</Typography>
                                                <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>Name: _tridorian-challenge.{domain}</Typography>
                                                <Typography variant="subtitle1" sx={{ fontWeight: 700, wordBreak: 'break-all' }}>
                                                    Value: {verificationToken}
                                                </Typography>
                                            </CardContent>
                                        </Card>
                                        <Button variant="contained" fullWidth size="large" onClick={handleVerifyClick}>Verify & Continue</Button>
                                    </Box>
                                )}
                            </CardContent>
                        </Card>
                    </Fade>
                )}

                {step === 1 && (
                    <Fade in>
                        <Card>
                            <CardContent sx={{ p: 4 }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                                    <GoogleIcon sx={{ color: '#ea4335', mr: 2, fontSize: 32 }} />
                                    <Box>
                                        <Typography variant="h6">Google Workspace Identity</Typography>
                                        <Typography variant="body2" color="text.secondary">Connect your organization's directory.</Typography>
                                    </Box>
                                </Box>
                                <TextField fullWidth label="Google Client ID" value={idpConfig.clientID} onChange={e => setIdpConfig({ ...idpConfig, clientID: e.target.value })} sx={{ mb: 2 }} />
                                <TextField fullWidth label="Google Client Secret" type="password" value={idpConfig.clientSecret} onChange={e => setIdpConfig({ ...idpConfig, clientSecret: e.target.value })} sx={{ mb: 2 }} />
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
                                <TextField fullWidth multiline rows={4} value={idpConfig.saKey} onChange={e => setIdpConfig({ ...idpConfig, saKey: e.target.value })} sx={{ mb: 2 }} />
                                <TextField fullWidth label="Admin User Email" value={idpConfig.adminEmail} onChange={e => setIdpConfig({ ...idpConfig, adminEmail: e.target.value })} sx={{ mb: 3 }} />
                                <Button variant="contained" fullWidth size="large" onClick={handleUpdateIdentity}>Connect Provider</Button>
                            </CardContent>
                        </Card>
                    </Fade>
                )}

                {step === 2 && (
                    <Fade in>
                        <Card sx={{ textAlign: 'center', p: 4 }}>
                            <CheckCircleIcon sx={{ color: '#1e8e3e', fontSize: 64, mb: 2 }} />
                            <Typography variant="h5" sx={{ fontWeight: 700, mb: 2 }}>Ready to Launch!</Typography>
                            <Typography color="text.secondary" sx={{ mb: 4 }}>Your secure environment is configured and ready to go.</Typography>
                            <Button variant="contained" size="large" fullWidth onClick={onComplete}>Enter Console</Button>
                        </Card>
                    </Fade>
                )}
            </Container>
        </Box>
    );
};

export default SetupWizard;
