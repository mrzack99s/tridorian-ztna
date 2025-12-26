import React, { useState, useEffect } from 'react';
import { Box, CircularProgress, ThemeProvider, CssBaseline } from '@mui/material';
import { theme } from './theme/theme';
import { User, Tenant, AccessPolicy, SignInPolicy, Node, Admin } from './types';
import DashboardLayout from './layout/DashboardLayout';

// Features
import LoginView from './features/auth/LoginView';
import ChangePasswordView from './features/auth/ChangePasswordView';
import SetupWizard from './features/setup/SetupWizard';
import DashboardView from './features/dashboard/DashboardView';
import UsersView from './features/users/UsersView';
import SignInPoliciesView from './features/policies/SignInPoliciesView';
import PoliciesView from './features/policies/PoliciesView';
import NodesView from './features/nodes/NodesView';
import AdminsView from './features/admins/AdminsView';
import SettingsView from './features/dashboard/SettingsView';

type ViewType = 'loading' | 'login' | 'change_password' | 'wizard' | 'dashboard' | 'users' | 'signin_policies' | 'access_policies' | 'nodes' | 'settings' | 'admins';

const App: React.FC = () => {
    const [view, setView] = useState<ViewType>('loading');
    const [user, setUser] = useState<User | null>(null);
    const [tenant, setTenant] = useState<Tenant | null>(null);

    // Data State
    const [accessPolicies, setAccessPolicies] = useState<AccessPolicy[]>([]);
    const [signInPolicies, setSignInPolicies] = useState<SignInPolicy[]>([]);
    const [nodes, setNodes] = useState<Node[]>([]);
    const [admins, setAdmins] = useState<Admin[]>([]);

    const checkSession = async () => {
        try {
            const res = await fetch('/auth/mgmt/me');
            if (!res.ok) {
                setView('login');
                return;
            }

            const userData = await res.json();
            if (!userData.success) {
                setView('login');
                return;
            }
            setUser(userData.data);

            if (userData.data.change_password_required) {
                setView('change_password');
                return;
            }

            const tenantRes = await fetch('/api/v1/tenant/me');
            if (tenantRes.ok) {
                const tenantData = await tenantRes.json();
                if (tenantData.success) {
                    setTenant(tenantData.data);
                    const t = tenantData.data;
                    if (!t.primary_domain || !t.google_client_id) {
                        setView('wizard');
                    } else if (view === 'loading') {
                        setView('dashboard');
                    }
                } else {
                    setView('wizard');
                }
            } else {
                setView('wizard');
            }
        } catch (err) {
            console.error('Session check failed:', err);
            setView('login');
        }
    };

    // Session Check
    useEffect(() => {
        checkSession();
    }, []);

    // Fetching Logic
    useEffect(() => {
        if (view === 'access_policies') fetchPolicies();
        if (view === 'signin_policies') fetchSignInPolicies();
        if (view === 'nodes') fetchNodes();
        if (view === 'admins') fetchAdmins();
    }, [view]);

    const fetchPolicies = async () => {
        const res = await fetch('/api/v1/policies/access');
        if (res.ok) {
            const data = await res.json();
            setAccessPolicies(data.data || []);
        }
    };

    const fetchSignInPolicies = async () => {
        const res = await fetch('/api/v1/policies/sign-in');
        if (res.ok) {
            const data = await res.json();
            setSignInPolicies(data.data || []);
        }
    };

    const fetchNodes = async () => {
        const res = await fetch('/api/v1/nodes');
        if (res.ok) {
            const data = await res.json();
            setNodes(data.data || []);
        }
    };

    const fetchAdmins = async () => {
        const res = await fetch('/api/v1/admins');
        if (res.ok) {
            const data = await res.json();
            setAdmins(data.data || []);
        }
    };

    // Global Handlers
    const handleLogin = async (email: string, password: string) => {
        try {
            const res = await fetch('/auth/mgmt/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });
            const data = await res.json();
            if (data.success) {
                window.location.reload();
            } else {
                alert('Login failed: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };

    const handleChangePassword = async (old_password: string, new_password: string) => {
        try {
            const res = await fetch('/api/v1/profile/change-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ old_password, new_password })
            });
            const data = await res.json();
            if (data.success) {
                alert('Password changed successfully. Please log in again.');
                await fetch('/auth/mgmt/logout', { method: 'POST' });
                window.location.reload();
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };



    const handleCreateNode = async (name: string): Promise<string | null> => {
        const res = await fetch('/api/v1/nodes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name })
        });
        if (res.ok) {
            const data = await res.json();
            fetchNodes();
            return data.data.auth_token;
        }
        alert('Failed to register node');
        return null;
    };

    const handleDeleteNode = async (id: string) => {
        if (!confirm('Are you sure? This will disconnect the gateway.')) return;
        await fetch('/api/v1/nodes', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id })
        });
        fetchNodes();
    };

    const handleCreateAdmin = async (admin: any) => {
        const res = await fetch('/api/v1/admins', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(admin)
        });
        if (res.ok) fetchAdmins();
        else {
            const data = await res.json();
            alert('Failed to create admin: ' + data.message);
        }
    };

    const handleDeleteAdmin = async (id: string) => {
        if (!confirm('Are you sure you want to delete this administrator?')) return;
        const res = await fetch('/api/v1/admins', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id })
        });
        if (res.ok) fetchAdmins();
        else {
            const data = await res.json();
            alert('Error: ' + data.message);
        }
    };

    const handleUpdateAdmin = async (id: string, name: string) => {
        const res = await fetch('/api/v1/admins', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id, name })
        });
        if (res.ok) fetchAdmins();
        else {
            const data = await res.json();
            alert('Error: ' + data.message);
        }
    };

    // Render Logic
    if (view === 'loading') {
        return (
            <ThemeProvider theme={theme}>
                <CssBaseline />
                <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
                    <CircularProgress />
                </Box>
            </ThemeProvider>
        );
    }

    if (view === 'login') return <LoginView onLogin={handleLogin} />;
    if (view === 'change_password') return <ChangePasswordView onChangePassword={handleChangePassword} />;
    if (view === 'wizard') return <SetupWizard tenant={tenant} onRefresh={checkSession} onComplete={() => setView('dashboard')} />;

    // Authenticated Layout
    const renderContent = () => {
        switch (view) {
            case 'dashboard': return <DashboardView tenant={tenant} />;
            case 'users': return <UsersView />;
            case 'signin_policies': return <SignInPoliciesView policies={signInPolicies} onRefresh={fetchSignInPolicies} />;
            case 'access_policies': return <PoliciesView policies={accessPolicies} onRefresh={fetchPolicies} />;
            case 'nodes': return <NodesView nodes={nodes} onCreate={handleCreateNode} onDelete={handleDeleteNode} />;
            case 'admins': return <AdminsView admins={admins} onCreate={handleCreateAdmin} onDelete={handleDeleteAdmin} onUpdate={handleUpdateAdmin} />;
            case 'settings': return <SettingsView tenant={tenant} onRefresh={checkSession} />;
            default: return <DashboardView tenant={tenant} />;
        }
    };

    return (
        <ThemeProvider theme={theme}>
            <DashboardLayout view={view} setView={setView} tenant={tenant}>
                {renderContent()}
            </DashboardLayout>
        </ThemeProvider>
    );
};

export default App;