import React, { useState } from 'react';
import {
    Box,
    Typography,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    Grid,
    IconButton,
    Chip,
    FormControlLabel,
    Checkbox,
    MenuItem,
    Paper,
    CircularProgress,
    Divider,
    Stack,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    CheckCircle as CheckCircleIcon,
    Block as BlockIcon,
    Warning as WarningIcon,
    DeviceHub as DeviceHubIcon,
    Public as PublicIcon,
    Group as GroupIcon,
    KeyboardArrowRight as ArrowIcon
} from '@mui/icons-material';
import { SignInPolicy, PolicyNode, PolicyCondition } from '../../types';

interface SignInPoliciesViewProps {
    policies: SignInPolicy[];
    onRefresh: () => void;
}

const CONDITION_TYPES = [
    { value: 'User', label: 'User / Group', icon: <GroupIcon fontSize="small" /> },
    { value: 'Network', label: 'Network / IP', icon: <PublicIcon fontSize="small" /> },
    { value: 'Device', label: 'Device / OS', icon: <DeviceHubIcon fontSize="small" /> },
];

const OPS: Record<string, { label: string; value: string }[]> = {
    'User': [{ label: 'In Group', value: 'in_group' }, { label: 'Email Ends With', value: 'email_suffix' }],
    'Network': [{ label: 'IP in CIDR', value: 'cidr' }, { label: 'Country Equals', value: 'country' }],
    'Device': [{ label: 'OS Equals', value: 'os' }, { label: 'Is Managed', value: 'managed' }],
};

const NodeEditor: React.FC<{
    node: PolicyNode;
    onChange: (newNode: PolicyNode) => void;
    onDelete?: () => void;
    depth?: number;
}> = ({ node, onChange, onDelete, depth = 0 }) => {
    const isLeaf = !!node.condition;

    const handleAddCondition = () => {
        const newNode: PolicyNode = {
            operator: 'AND',
            condition: { type: 'User', field: 'group', op: 'in_group', value: '' }
        };
        if (node.children) {
            onChange({ ...node, children: [...node.children, newNode] });
        } else {
            // Transform to branch
            onChange({ ...node, condition: undefined, children: [newNode] });
        }
    };

    const handleAddSubGroup = () => {
        const newNode: PolicyNode = {
            operator: 'OR',
            children: []
        };
        if (node.children) {
            onChange({ ...node, children: [...node.children, newNode] });
        } else {
            onChange({ ...node, condition: undefined, children: [newNode] });
        }
    };

    const handleChildChange = (index: number, child: PolicyNode) => {
        const newChildren = [...(node.children || [])];
        newChildren[index] = child;
        onChange({ ...node, children: newChildren });
    };

    const handleRemoveChild = (index: number) => {
        const newChildren = [...(node.children || [])];
        newChildren.splice(index, 1);
        onChange({ ...node, children: newChildren });
    };

    return (
        <Paper variant="outlined" sx={{
            p: 2,
            mb: 2,
            borderLeft: depth > 0 ? `4px solid ${node.operator === 'AND' ? '#1a73e8' : '#f4b400'}` : 'none',
            bgcolor: depth % 2 === 0 ? 'rgba(0,0,0,0.01)' : '#fff',
            borderRadius: 2
        }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: isLeaf ? 0 : 2, gap: 2 }}>
                {!isLeaf && (
                    <ToggleButtonGroup
                        size="small"
                        color="primary"
                        value={node.operator}
                        exclusive
                        onChange={(_, val) => val && onChange({ ...node, operator: val })}
                    >
                        <ToggleButton value="AND" sx={{ px: 2, fontWeight: 700, fontSize: '0.75rem' }}>AND</ToggleButton>
                        <ToggleButton value="OR" sx={{ px: 2, fontWeight: 700, fontSize: '0.75rem' }}>OR</ToggleButton>
                    </ToggleButtonGroup>
                )}

                {isLeaf && (
                    <Grid container spacing={1} alignItems="center">
                        <Grid item xs={3}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Type"
                                value={node.condition?.type}
                                onChange={(e) => onChange({ ...node, condition: { ...node.condition!, type: e.target.value, op: OPS[e.target.value][0].value } })}
                            >
                                {CONDITION_TYPES.map(t => <MenuItem key={t.value} value={t.value}>{t.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid item xs={3}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Operator"
                                value={node.condition?.op}
                                onChange={(e) => onChange({ ...node, condition: { ...node.condition!, op: e.target.value } })}
                            >
                                {node.condition && (OPS[node.condition.type] || []).map(o => <MenuItem key={o.value} value={o.value}>{o.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid item xs={5}>
                            <TextField
                                fullWidth
                                size="small"
                                label="Value"
                                value={node.condition?.value}
                                onChange={(e) => onChange({ ...node, condition: { ...node.condition!, value: e.target.value } })}
                                placeholder="e.g. admin, 192.168.1.0/24"
                            />
                        </Grid>
                        <Grid item xs={1} sx={{ textAlign: 'right' }}>
                            {onDelete && (
                                <IconButton size="small" color="error" onClick={onDelete}>
                                    <DeleteIcon fontSize="small" />
                                </IconButton>
                            )}
                        </Grid>
                    </Grid>
                )}

                {!isLeaf && <Box sx={{ flexGrow: 1 }} />}

                {!isLeaf && onDelete && (
                    <IconButton size="small" color="error" onClick={onDelete}>
                        <DeleteIcon fontSize="small" />
                    </IconButton>
                )}
            </Box>

            {!isLeaf && (
                <Box sx={{ ml: depth > 0 ? 1 : 0, pl: depth > 0 ? 2 : 0, borderLeft: depth > 0 ? '1px dashed #ddd' : 'none' }}>
                    {(node.children || []).map((child, idx) => (
                        <NodeEditor
                            key={idx}
                            node={child}
                            depth={depth + 1}
                            onChange={(updated) => handleChildChange(idx, updated)}
                            onDelete={() => handleRemoveChild(idx)}
                        />
                    ))}
                    <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                        <Button size="small" startIcon={<AddIcon />} onClick={handleAddCondition} variant="text" sx={{ fontSize: '0.75rem' }}>
                            Add Condition
                        </Button>
                        <Button size="small" startIcon={<AddIcon />} onClick={handleAddSubGroup} variant="text" color="secondary" sx={{ fontSize: '0.75rem' }}>
                            Add Sub-Group
                        </Button>
                    </Stack>
                </Box>
            )}
        </Paper>
    );
};

const SignInPoliciesView: React.FC<SignInPoliciesViewProps> = ({ policies, onRefresh }) => {
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<SignInPolicy | null>(null);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<{
        name: string;
        priority: number;
        block: boolean;
        root_node: PolicyNode;
    }>({
        name: '',
        priority: 10,
        block: false,
        root_node: { operator: 'AND', children: [] }
    });

    const handleOpenDialog = (policy?: SignInPolicy) => {
        if (policy) {
            setEditingPolicy(policy);
            setFormData({
                name: policy.name,
                priority: policy.priority,
                block: policy.block,
                root_node: policy.root_node || { operator: 'AND', children: [] }
            });
        } else {
            setEditingPolicy(null);
            setFormData({
                name: '',
                priority: 10,
                block: false,
                root_node: { operator: 'AND', children: [] }
            });
        }
        setDialogOpen(true);
    };

    const handleCloseDialog = () => {
        setDialogOpen(false);
        setEditingPolicy(null);
    };

    const handleSubmit = async () => {
        setLoading(true);
        const url = '/api/v1/policies/sign-in';
        const method = editingPolicy ? 'PATCH' : 'POST';
        const body = editingPolicy ? { ...formData, id: editingPolicy.id } : formData;

        try {
            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            const data = await res.json();
            if (data.success) {
                onRefresh();
                handleCloseDialog();
            } else {
                alert('Error: ' + (data.error || data.message));
            }
        } catch (err) {
            console.error('Failed to save policy:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this policy?')) return;
        try {
            const res = await fetch('/api/v1/policies/sign-in', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id })
            });
            if (res.ok) {
                onRefresh();
            } else {
                const data = await res.json();
                alert('Error: ' + (data.error || data.message));
            }
        } catch (err) {
            console.error('Failed to delete policy:', err);
        }
    };

    const renderNodeSummary = (node: PolicyNode): string => {
        if (node.condition) {
            return `${node.condition.type} ${node.condition.op} "${node.condition.value}"`;
        }
        if (!node.children || node.children.length === 0) return "TRUE";
        const childrenSummary = node.children.map(c => renderNodeSummary(c));
        return `(${childrenSummary.join(` ${node.operator} `)})`;
    };

    return (
        <Box sx={{ maxWidth: 1200, mx: 'auto', py: 4 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 800, color: '#202124' }}>Sign-in Policies</Typography>
                    <Typography color="text.secondary">Implement conditional access and zero-trust security policies.</Typography>
                </Box>
                <Button
                    variant="contained"
                    disableElevation
                    startIcon={<AddIcon />}
                    onClick={() => handleOpenDialog()}
                    sx={{ borderRadius: 2, px: 3, bgcolor: '#1a73e8', '&:hover': { bgcolor: '#1765cc' } }}
                >
                    Add Policy
                </Button>
            </Box>

            {policies.length === 0 ? (
                <Paper variant="outlined" sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: '#fff', border: '1px dashed #dadce0' }}>
                    <WarningIcon sx={{ fontSize: 48, color: '#dadce0', mb: 2 }} />
                    <Typography variant="h6" sx={{ fontWeight: 600, color: '#3c4043' }}>No policies configured</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 4, maxWidth: 400, mx: 'auto' }}>
                        Zero Trust policies allow you to control access based on user context, location, and device status.
                    </Typography>
                    <Button variant="outlined" startIcon={<AddIcon />} onClick={() => handleOpenDialog()} sx={{ borderRadius: 2 }}>
                        Create First Policy
                    </Button>
                </Paper>
            ) : (
                <Stack spacing={2}>
                    {policies.sort((a, b) => a.priority - b.priority).map((policy) => (
                        <Card key={policy.id} variant="outlined" sx={{ borderRadius: 3, border: '1px solid #dadce0', transition: '0.2s', '&:hover': { boxShadow: '0 1px 6px rgba(32,33,36,.28)', borderColor: 'transparent' } }}>
                            <CardContent sx={{ p: 2.5 }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2.5 }}>
                                    <Box sx={{
                                        width: 48,
                                        height: 48,
                                        borderRadius: '50%',
                                        bgcolor: policy.block ? 'rgba(217, 48, 37, 0.1)' : 'rgba(30, 142, 62, 0.1)',
                                        color: policy.block ? '#d93025' : '#1e8e32',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center'
                                    }}>
                                        {policy.block ? <BlockIcon /> : <CheckCircleIcon />}
                                    </Box>
                                    <Box sx={{ flexGrow: 1 }}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 0.5 }}>
                                            <Typography variant="subtitle1" sx={{ fontWeight: 700, color: '#202124' }}>{policy.name}</Typography>
                                            <Chip
                                                label={`P${policy.priority}`}
                                                size="small"
                                                variant="outlined"
                                                sx={{ height: 20, fontSize: 10, fontWeight: 700, color: '#5f6368', borderColor: '#dadce0' }}
                                            />
                                            <Chip
                                                label={policy.block ? 'DENY' : 'ALLOW'}
                                                size="small"
                                                sx={{
                                                    height: 20,
                                                    fontSize: 10,
                                                    fontWeight: 900,
                                                    bgcolor: policy.block ? '#fce8e6' : '#e6f4ea',
                                                    color: policy.block ? '#c5221f' : '#137333',
                                                    border: 'none'
                                                }}
                                            />
                                        </Box>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                            <ArrowIcon sx={{ fontSize: 16, color: '#9aa0a6' }} />
                                            <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: 12, color: '#5f6368', bgcolor: '#f8f9fa', px: 1, py: 0.5, borderRadius: 1 }}>
                                                <strong>IF</strong> {policy.root_node ? renderNodeSummary(policy.root_node) : "TRUE"}
                                            </Typography>
                                        </Box>
                                    </Box>
                                    <Box sx={{ display: 'flex', gap: 1 }}>
                                        <Tooltip title="Edit Policy">
                                            <IconButton onClick={() => handleOpenDialog(policy)} size="small" sx={{ color: '#5f6368' }}>
                                                <EditIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="Delete Policy">
                                            <IconButton onClick={() => handleDelete(policy.id)} size="small" color="error">
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                    </Box>
                                </Box>
                            </CardContent>
                        </Card>
                    ))}
                </Stack>
            )}

            <Dialog
                open={dialogOpen}
                onClose={handleCloseDialog}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: 4, boxShadow: '0 24px 38px 3px rgba(0,0,0,0.14), 0 9px 46px 8px rgba(0,0,0,0.12), 0 11px 15px -7px rgba(0,0,0,0.2)' }
                }}
            >
                <DialogTitle sx={{ fontWeight: 800, p: 3, color: '#202124' }}>
                    {editingPolicy ? 'Edit Access Policy' : 'Create Zero Trust Policy'}
                </DialogTitle>
                <DialogContent sx={{ p: 3, pt: 1 }}>
                    <Grid container spacing={3} sx={{ mb: 4 }}>
                        <Grid item xs={8}>
                            <TextField
                                fullWidth
                                variant="outlined"
                                label="Policy Name"
                                placeholder="e.g. Block login from outside TH"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            />
                        </Grid>
                        <Grid item xs={4}>
                            <TextField
                                fullWidth
                                type="number"
                                label="Priority"
                                value={formData.priority}
                                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                                helperText="Lower values have higher precedence"
                            />
                        </Grid>
                    </Grid>

                    <Typography variant="overline" sx={{ fontWeight: 700, color: '#5f6368', mb: 2, display: 'block' }}>
                        Conditional Logic
                    </Typography>

                    <NodeEditor
                        node={formData.root_node}
                        onChange={(node) => setFormData({ ...formData, root_node: node })}
                    />

                    <Box sx={{ mt: 4, p: 3, borderRadius: 3, bgcolor: formData.block ? '#fce8e6' : '#e6f4ea', border: `1px solid ${formData.block ? '#f1b3b1' : '#81c995'}` }}>
                        <FormControlLabel
                            control={
                                <Checkbox
                                    checked={formData.block}
                                    onChange={(e) => setFormData({ ...formData, block: e.target.checked })}
                                    color={formData.block ? "error" : "success"}
                                />
                            }
                            label={
                                <Box>
                                    <Typography variant="subtitle2" sx={{ fontWeight: 700, color: formData.block ? '#c5221f' : '#137333' }}>
                                        {formData.block ? 'DENY Access if criteria met' : 'ALLOW Access if criteria met'}
                                    </Typography>
                                    <Typography variant="caption" sx={{ color: formData.block ? '#d93025' : '#1e8e3e' }}>
                                        {formData.block
                                            ? 'Users matching these conditions will be strictly blocked from logging in.'
                                            : 'Users matching these conditions will be granted access to the network.'}
                                    </Typography>
                                </Box>
                            }
                        />
                    </Box>
                </DialogContent>
                <DialogActions sx={{ p: 3, bgcolor: '#f8f9fa' }}>
                    <Button onClick={handleCloseDialog} sx={{ color: '#5f6368', fontWeight: 600 }}>Cancel</Button>
                    <Button
                        variant="contained"
                        disableElevation
                        onClick={handleSubmit}
                        disabled={!formData.name || loading}
                        sx={{ borderRadius: 2, px: 4, fontWeight: 700, bgcolor: '#1a73e8' }}
                    >
                        {loading ? <CircularProgress size={20} color="inherit" /> : (editingPolicy ? 'Update Policy' : 'Create Policy')}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default SignInPoliciesView;
