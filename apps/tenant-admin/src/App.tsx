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
import SignInPoliciesView from './features/policies/SignInPoliciesView';
import PoliciesView from './features/policies/PoliciesView';
import NodesView from './features/nodes/NodesView';
import AdminsView from './features/admins/AdminsView';
import SettingsView from './features/dashboard/SettingsView';
import ApplicationsView from './features/applications/ApplicationsView';

type ViewType = 'loading' | 'login' | 'change_password' | 'wizard' | 'dashboard' | 'users' | 'signin_policies' | 'access_policies' | 'applications' | 'nodes' | 'settings' | 'admins';

const App: React.FC = () => {
    const [view, setView] = useState<ViewType>('loading');
    const [user, setUser] = useState<User | null>(null);
    const [tenant, setTenant] = useState<Tenant | null>(null);

    const [domains, setDomains] = useState<string[]>([]);
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
        if (view === 'admins') {
            fetchAdmins();
            fetchDomains();
        }
    }, [view]);

    const fetchDomains = async () => {
        const res = await fetch('/api/v1/tenants/domains');
        if (res.ok) {
            const data = await res.json();
            setDomains(data.data || []);
        }
    };

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
                // Auto-login essentially means we just refresh the session context.
                // Since the backend cookie might still be valid (or we assume the backend didn't invalidate it),
                // we try to reload. If the backend invalidated the session, it will redirect to login anyway.
                // However, usually change-password might NOT invalidate the current session if we don't tell it to.
                // Let's assume the session persists or we just proceed to dashboard.

                // If the "change_password_required" flag was set, we need to refresh the user profile to clear it.
                await checkSession();
                // checkSession will update the user state. If change_password_required is now false, 
                // it will switch view to 'dashboard' (or 'wizard' if verified).

                // Note: The original code forced a logout. We removed that.
                // Using window.location.reload() is also an option but checkSession() is smoother if it works.
                // But checkSession logic has: if (view === 'loading') setView('dashboard'). 
                // Since view is currently 'change_password', checkSession might need a tweak or we manually set view.

                if (view === 'change_password') {
                    // Force transition to dashboard if session supports it
                    // We rely on checkSession updating the 'user' object where change_password_required is false.
                    // But checkSession implementation above:
                    /*
                       if (userData.data.change_password_required) {
                           setView('change_password');
                           return;
                       }
                       //... logic to setView('dashboard') only if view === 'loading'
                    */
                    // So we need to explicitly setView('dashboard') if checkSession passes.
                    window.location.reload(); // Simplest way to re-run the whole auth logic clean.
                }
            } else {
                alert('Error: ' + data.message);
            }
        } catch (err) { console.error(err); }
    };



    const handleCreateNode = async (name: string, skuId: string, clientCIDR: string): Promise<string | null> => {
        const res = await fetch('/api/v1/nodes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, sku_id: skuId, client_cidr: clientCIDR })
        });
        if (res.ok) {
            const data = await res.json();
            fetchNodes();
            return data.data.id;
        }
        alert('Failed to create gateway');
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
        // Remove password if empty to trigger auto-generation on backend
        if (!admin.password) delete admin.password;

        const res = await fetch('/api/v1/admins', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(admin)
        });

        if (res.ok) {
            const data = await res.json();
            fetchAdmins();
            return data.data.password; // Return the generated password
        } else {
            const data = await res.json();
            alert('Failed to create admin: ' + data.message);
            return null;
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

    const handleUpdateAdmin = async (id: string, name: string, role: string) => {
        const res = await fetch('/api/v1/admins', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id, name, role })
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
            case 'signin_policies': return <SignInPoliciesView policies={signInPolicies} onRefresh={fetchSignInPolicies} />;
            case 'access_policies': return <PoliciesView policies={accessPolicies} onRefresh={fetchPolicies} />;
            case 'applications': return <ApplicationsView />;
            case 'nodes': return <NodesView nodes={nodes} onCreate={handleCreateNode} onDelete={handleDeleteNode} />;
            case 'admins': return <AdminsView admins={admins} domains={domains} onCreate={handleCreateAdmin} onDelete={handleDeleteAdmin} onUpdate={handleUpdateAdmin} />;
            case 'settings': return <SettingsView tenant={tenant} onRefresh={checkSession} user={user} />;
            default: return <DashboardView tenant={tenant} />;
        }
    };

    return (
        <ThemeProvider theme={theme}>
            <DashboardLayout view={view} setView={setView} tenant={tenant} user={user}>
                {renderContent()}
            </DashboardLayout>
        </ThemeProvider>
    );
};

export default App;