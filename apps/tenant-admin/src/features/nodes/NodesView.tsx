import React, { useState, useMemo, useEffect } from 'react';
import {
    Box, Typography, Button, TableContainer, Table, TableHead, TableRow, TableCell,
    TableBody, Paper, IconButton, Dialog, DialogTitle, DialogContent,
    TextField, DialogActions, Alert, TablePagination, Radio, InputAdornment, Chip,
    useMediaQuery, useTheme
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Search as SearchIcon,
    Router as RouterIcon,
    Memory as MemoryIcon,
    Speed as SpeedIcon,
    Storage as StorageIcon,
    AttachMoney as MoneyIcon,
    HelpOutline as HelpIcon,
    Terminal as TerminalIcon
} from '@mui/icons-material';
import { Node, NodeSku } from '../../types';

interface NodesViewProps {
    nodes: Node[];
    onCreate: (name: string, skuId: string, clientCIDR: string) => Promise<string | null>;
    onDelete: (id: string) => Promise<void>;
}

const NodesView: React.FC<NodesViewProps> = ({ nodes, onCreate, onDelete }) => {
    // Dialog State
    const [showDialog, setShowDialog] = useState(false);
    const [newNodeName, setNewNodeName] = useState('');
    const [selectedSkuId, setSelectedSkuId] = useState<string | null>(null);
    const [createdNodeId, setCreatedNodeId] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [clientCIDR, setClientCIDR] = useState('10.8.0.0/23');
    const [cidrError, setCidrError] = useState<string | null>(null);

    // View Guide State
    const [viewNode, setViewNode] = useState<Node | null>(null);

    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const isSmallMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const validateCIDR = (cidr: string) => {
        // Regex for Private IP (10.x, 172.16-31.x, 192.168.x)
        const privateIpRegex = /^(10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(1[6-9]|2\d|3[0-1])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3})\/(\d{1,2})$/;
        const match = cidr.match(privateIpRegex);

        if (!match) {
            setCidrError("Must be a valid private IP CIDR (e.g., 10.x.x.x/xx)");
            return false;
        }

        const prefix = parseInt(match[3], 10);
        if (prefix < 22) { // smaller number = larger subnet
            setCidrError("Subnet is too large (must be /22 or smaller, checking prefix length >= 22)");
            return false;
        }

        setCidrError(null);
        return true;
    };

    // SKU Data
    const [skus, setSkus] = useState<NodeSku[]>([]);
    const [loadingSkus, setLoadingSkus] = useState(false);

    // SKU Table State
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(5);
    const [searchQuery, setSearchQuery] = useState('');

    useEffect(() => {
        fetchSkus();
    }, []);

    const fetchSkus = async () => {
        setLoadingSkus(true);
        try {
            const res = await fetch('/api/v1/nodes/skus');
            if (res.ok) {
                const data = await res.json();
                setSkus(data.data || []);
            }
        } catch (err) {
            console.error('Failed to fetch SKUs', err);
        } finally {
            setLoadingSkus(false);
        }
    };

    const filteredSkus = useMemo(() => {
        return skus.filter(sku =>
            sku.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            (sku.description && sku.description.toLowerCase().includes(searchQuery.toLowerCase()))
        );
    }, [skus, searchQuery]);

    const paginatedSkus = useMemo(() => {
        return filteredSkus.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);
    }, [filteredSkus, page, rowsPerPage]);

    const handleCreate = async () => {
        if (!newNodeName || !selectedSkuId) return;
        if (!validateCIDR(clientCIDR)) return;

        setLoading(true);
        const nodeId = await onCreate(newNodeName, selectedSkuId, clientCIDR);
        setLoading(false);
        if (nodeId) {
            setCreatedNodeId(nodeId);
        }
    };

    const handleClose = () => {
        setShowDialog(false);
        setNewNodeName('');
        setSelectedSkuId(null);
        setCreatedNodeId(null);
        setPage(0);
        setSearchQuery('');
        setClientCIDR('10.8.0.0/23');
        setCidrError(null);
    };

    const formatBandwidth = (bwInfo: number) => {
        if (bwInfo >= 1000) return `${(bwInfo / 1000).toFixed(1)} Gbps`;
        return `${bwInfo} Mbps`;
    };

    const formatPrice = (cents: number) => {
        return `$${(cents / 100).toFixed(2)}`;
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant={isSmallMobile ? "h5" : "h4"} sx={{ fontWeight: 800, color: '#202124', display: 'flex', alignItems: 'center', gap: 2 }}>
                        <RouterIcon sx={{ fontSize: isSmallMobile ? 32 : 40, color: '#1a73e8' }} />
                        Gateways
                    </Typography>
                    <Typography color="text.secondary" sx={{ mt: 1, fontSize: isSmallMobile ? '0.85rem' : '1rem' }}>
                        Manage your edge gateways and connectors.
                    </Typography>
                </Box>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => setShowDialog(true)}
                    sx={{
                        bgcolor: '#1a73e8',
                        borderRadius: 2,
                        px: 3,
                        py: 1,
                        fontWeight: 700,
                        textTransform: 'none'
                    }}
                >
                    Register Gateway
                </Button>
            </Box>

            {isMobile ? (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                    {nodes.map((node) => (
                        <Paper key={node.id} variant="outlined" sx={{ p: 2, borderRadius: 3, display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                <Box>
                                    <Typography variant="body1" sx={{ fontWeight: 700 }}>{node.name}</Typography>
                                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.8rem' }}>{node.hostname}</Typography>
                                </Box>
                                <Chip
                                    label={node.status}
                                    size="small"
                                    color={node.status === 'ONLINE' ? 'success' : 'default'}
                                    sx={{ borderRadius: 1.5, fontWeight: 700, height: 24, fontSize: '0.7rem' }}
                                />
                            </Box>

                            <Box sx={{ display: 'flex', gap: 2, color: 'text.secondary', fontSize: '0.8rem' }}>
                                {node.ip_address && (
                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                        <Typography variant="caption" sx={{ fontWeight: 600 }}>IP:</Typography> {node.ip_address}
                                    </Box>
                                )}
                                {(node.gateway_version || node.version) && (
                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                        <Typography variant="caption" sx={{ fontWeight: 600 }}>Ver:</Typography> {node.gateway_version || node.version}
                                    </Box>
                                )}
                            </Box>

                            {node.node_sku && (
                                <Box sx={{ bgcolor: '#f8f9fa', p: 1.5, borderRadius: 2, border: '1px solid #e0e0e0' }}>
                                    <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1, fontSize: '0.8rem' }}>{node.node_sku.name}</Typography>
                                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                            <SpeedIcon sx={{ fontSize: 16, color: '#5f6368' }} />
                                            <Typography variant="caption" sx={{ fontWeight: 500, color: '#3c4043' }}>{formatBandwidth(node.node_sku.bandwidth)}</Typography>
                                        </Box>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                            <MemoryIcon sx={{ fontSize: 16, color: '#5f6368' }} />
                                            <Typography variant="caption" sx={{ fontWeight: 500, color: '#3c4043' }}>{node.node_sku.max_users} Users</Typography>
                                        </Box>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                            <MoneyIcon sx={{ fontSize: 16, color: '#5f6368' }} />
                                            <Typography variant="caption" sx={{ fontWeight: 500, color: '#3c4043' }}>{formatPrice(node.node_sku.price_cents)}/hr</Typography>
                                        </Box>
                                    </Box>
                                </Box>
                            )}

                            <Box sx={{ display: 'flex', gap: 1, mt: 0.5 }}>
                                <Button
                                    fullWidth
                                    size="small"
                                    variant="outlined"
                                    startIcon={<TerminalIcon />}
                                    onClick={() => setViewNode(node)}
                                    sx={{ textTransform: 'none', fontWeight: 600, borderColor: '#dadce0', color: '#5f6368' }}
                                >
                                    Connect
                                </Button>
                                <IconButton onClick={() => onDelete(node.id)} size="small" color="error" sx={{ border: '1px solid #ffcdd2' }}>
                                    <DeleteIcon fontSize="small" />
                                </IconButton>
                            </Box>
                        </Paper>
                    ))}
                    {nodes.length === 0 && (
                        <Paper variant="outlined" sx={{ p: 4, textAlign: 'center', borderRadius: 3, borderStyle: 'dashed' }}>
                            <Typography variant="body2" color="text.secondary">No gateways registered yet.</Typography>
                        </Paper>
                    )}
                </Box>
            ) : (
                <TableContainer component={Paper} elevation={0} sx={{ border: '1px solid #e0e0e0', borderRadius: 3 }}>
                    <Table>
                        <TableHead sx={{ bgcolor: '#f8f9fa' }}>
                            <TableRow>
                                <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>Name</TableCell>
                                <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>Status</TableCell>
                                {!isSmallMobile && <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>Public IP</TableCell>}
                                {!isSmallMobile && <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>Version</TableCell>}
                                {!isMobile && <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>SKU</TableCell>}
                                <TableCell sx={{ fontWeight: 700, color: '#5f6368' }}>Actions</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {nodes.map((node) => (
                                <TableRow key={node.id} hover>
                                    <TableCell>
                                        <Box>
                                            <Typography variant="body2" sx={{ fontWeight: 600 }}>{node.name}</Typography>
                                            <Typography variant="caption" color="text.secondary">{node.hostname}</Typography>
                                        </Box>
                                    </TableCell>
                                    <TableCell>
                                        <Chip
                                            label={node.status}
                                            size="small"
                                            color={node.status === 'ONLINE' ? 'success' : 'default'}
                                            sx={{ borderRadius: 1.5, fontWeight: 700, height: 24, fontSize: '0.7rem' }}
                                        />
                                    </TableCell>

                                    {!isSmallMobile && <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>{node.ip_address || '-'}</TableCell>}
                                    {!isSmallMobile && <TableCell sx={{ fontSize: '0.85rem' }}>{node.gateway_version || node.version || '-'}</TableCell>}
                                    {!isMobile && (
                                        <TableCell>
                                            {node.node_sku ? (
                                                <Box>
                                                    <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.85rem' }}>{node.node_sku.name}</Typography>
                                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mt: 0.5, color: 'text.secondary' }}>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                                            <SpeedIcon sx={{ fontSize: 14 }} />
                                                            <Typography variant="caption">{formatBandwidth(node.node_sku.bandwidth)}</Typography>
                                                        </Box>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                                            <MemoryIcon sx={{ fontSize: 14 }} />
                                                            <Typography variant="caption">{node.node_sku.max_users} Users</Typography>
                                                        </Box>
                                                    </Box>
                                                </Box>
                                            ) : <Typography variant="caption" color="text.secondary">-</Typography>}
                                        </TableCell>
                                    )}
                                    <TableCell>
                                        <Box sx={{ display: 'flex', gap: 1 }}>
                                            <Button
                                                size="small"
                                                variant="outlined"
                                                startIcon={<TerminalIcon />}
                                                onClick={() => setViewNode(node)}
                                                sx={{ textTransform: 'none', fontWeight: 600, borderColor: '#dadce0', color: '#5f6368' }}
                                            >
                                                Connect
                                            </Button>
                                            <IconButton onClick={() => onDelete(node.id)} size="small" color="error">
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </Box>
                                    </TableCell>
                                </TableRow>
                            ))}
                            {nodes.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} align="center" sx={{ py: 8 }}>
                                        <Typography variant="body1" color="text.secondary">No gateways registered yet.</Typography>
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </TableContainer>
            )}

            <Dialog
                open={showDialog}
                fullScreen={isMobile}
                onClose={(_, reason) => {
                    if (reason !== 'backdropClick') handleClose();
                }}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: isMobile ? 0 : 4, minHeight: isMobile ? '100vh' : 600 }
                }}
            >
                <DialogTitle sx={{ fontWeight: 800, p: 3, borderBottom: '1px solid #f1f3f4' }}>
                    Register New Gateway
                </DialogTitle>
                <DialogContent sx={{ p: isMobile ? 3 : 4 }}>
                    {!createdNodeId ? (
                        <>
                            <Box sx={{ mb: 4 }}>
                                <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>Gateway Name</Typography>
                                <TextField
                                    autoFocus
                                    placeholder="e.g. AWS-Production-Gateway-01"
                                    fullWidth
                                    value={newNodeName}
                                    onChange={(e) => setNewNodeName(e.target.value)}
                                    sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                />
                            </Box>

                            <Box sx={{ mb: 4 }}>
                                <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>Private Client CIDR</Typography>
                                <TextField
                                    fullWidth
                                    placeholder="e.g. 10.8.0.0/23"
                                    value={clientCIDR}
                                    onChange={(e) => {
                                        setClientCIDR(e.target.value);
                                        validateCIDR(e.target.value);
                                    }}
                                    error={!!cidrError}
                                    helperText={cidrError || "Default: 10.8.0.0/23 (Max size /22)"}
                                    sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                />
                            </Box>

                            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 2 }}>Select Gateway SKU</Typography>

                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                                <TextField
                                    size="small"
                                    placeholder="Search SKUs..."
                                    value={searchQuery}
                                    onChange={(e) => {
                                        setSearchQuery(e.target.value);
                                        setPage(0);
                                    }}
                                    InputProps={{
                                        startAdornment: <InputAdornment position="start"><SearchIcon color="action" fontSize="small" /></InputAdornment>,
                                    }}
                                    sx={{ width: 300, '& .MuiOutlinedInput-root': { borderRadius: 2 as any } }}
                                />
                            </Box>

                            <TableContainer component={Paper} variant="outlined" sx={{ borderRadius: 2, mb: 2, maxHeight: isMobile ? 250 : 300, overflowX: isMobile ? 'auto' : 'hidden' }}>
                                <Table stickyHeader size="small">
                                    <TableHead>
                                        <TableRow>
                                            <TableCell width={50}></TableCell>
                                            <TableCell>SKU Name</TableCell>
                                            <TableCell>Max Users</TableCell>
                                            <TableCell>Bandwidth</TableCell>
                                            <TableCell>Price</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {paginatedSkus.map((sku) => {
                                            const isSelected = selectedSkuId === sku.id;
                                            return (
                                                <TableRow
                                                    key={sku.id}
                                                    hover
                                                    onClick={() => setSelectedSkuId(sku.id)}
                                                    selected={isSelected}
                                                    sx={{ cursor: 'pointer' }}
                                                >
                                                    <TableCell padding="checkbox">
                                                        <Radio
                                                            checked={isSelected}
                                                            onChange={() => setSelectedSkuId(sku.id)}
                                                            value={sku.id}
                                                            size="small"
                                                        />
                                                    </TableCell>
                                                    <TableCell sx={{ fontWeight: 600 }}>
                                                        {sku.name}
                                                        <Typography variant="caption" display="block" color="text.secondary">{sku.description}</Typography>
                                                    </TableCell>
                                                    <TableCell sx={{ color: 'text.secondary' }}>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}><MemoryIcon fontSize="inherit" />{sku.max_users}</Box>
                                                    </TableCell>
                                                    <TableCell sx={{ color: 'text.secondary' }}>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}><SpeedIcon fontSize="inherit" />{formatBandwidth(sku.bandwidth)}</Box>
                                                    </TableCell>
                                                    <TableCell sx={{ color: 'text.secondary' }}>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>{formatPrice(sku.price_cents)} / hr</Box>
                                                    </TableCell>
                                                </TableRow>
                                            );
                                        })}
                                        {paginatedSkus.length === 0 && (
                                            <TableRow>
                                                <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                                                    {loadingSkus ? 'Loading SKUs...' : `No SKUs found matching "${searchQuery}"`}
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                            <TablePagination
                                component="div"
                                count={filteredSkus.length}
                                page={page}
                                onPageChange={(_, newPage) => setPage(newPage)}
                                rowsPerPage={rowsPerPage}
                                onRowsPerPageChange={(e) => {
                                    setRowsPerPage(parseInt(e.target.value, 10));
                                    setPage(0);
                                }}
                                rowsPerPageOptions={[5, 10, 25]}
                            />
                        </>
                    ) : (
                        <Box sx={{ mt: 2 }}>
                            <Box sx={{ textAlign: 'center', mb: 4 }}>
                                <Box sx={{
                                    width: 64, height: 64, borderRadius: '50%', bgcolor: '#e8f0fe', color: '#1a73e8',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center', mx: 'auto', mb: 2
                                }}>
                                    <RouterIcon fontSize="large" color="inherit" />
                                </Box>
                                <Typography variant="h5" sx={{ fontWeight: 800, mb: 1, color: '#202124' }}>Gateway Created Successfully!</Typography>
                                <Typography color="text.secondary">
                                    Your gateway <strong>{newNodeName}</strong> is ready to be configured.
                                </Typography>
                            </Box>

                            <Typography variant="h6" sx={{ fontWeight: 700, mb: 2, fontSize: '1rem', color: '#202124' }}>
                                Setup Instructions
                            </Typography>

                            <Box sx={{ mb: 3 }}>
                                <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                                    1. Copy your Node ID
                                </Typography>
                                <Paper variant="outlined" sx={{ p: 2, bgcolor: '#f8f9fa', borderRadius: 2, fontFamily: 'monospace', display: 'flex', justifyContent: 'space-between', alignItems: 'center', border: '1px solid #dadce0' }}>
                                    <span style={{ wordBreak: 'break-all', fontWeight: 600, color: '#202124' }}>{createdNodeId}</span>
                                    <Button
                                        size="small"
                                        sx={{ textTransform: 'none', fontWeight: 700 }}
                                        onClick={() => navigator.clipboard.writeText(createdNodeId!)}
                                    >
                                        Copy
                                    </Button>
                                </Paper>
                            </Box>

                            <Box sx={{ mb: 3 }}>
                                <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                                    2. Run the Gateway Registration Command
                                </Typography>
                                <Paper variant="outlined" sx={{ p: 2, bgcolor: '#202124', color: '#e8eaed', borderRadius: 2, fontFamily: 'Google Sans Mono, monospace', fontSize: '0.85rem' }}>
                                    <Box sx={{ mb: 1, userSelect: 'none', color: '#9aa0a6' }}>
                                        # Run this temporarily to test connection
                                    </Box>
                                    <div style={{ wordBreak: 'break-all', lineHeight: 1.5 }}>
                                        ./gateway --node-id "{createdNodeId}"
                                    </div>
                                </Paper>
                            </Box>

                            <Box sx={{ mb: 3 }}>
                                <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                                    3. Install as Background Service
                                </Typography>
                                <Paper variant="outlined" sx={{ p: 2, bgcolor: '#202124', color: '#e8eaed', borderRadius: 2, fontFamily: 'Google Sans Mono, monospace', fontSize: '0.85rem' }}>
                                    <Box sx={{ mb: 1, userSelect: 'none', color: '#9aa0a6' }}>
                                        # Install and start as a system service
                                    </Box>
                                    <div style={{ wordBreak: 'break-all', lineHeight: 1.5 }}>
                                        sudo ./gateway --install-service --node-id "{createdNodeId}"
                                    </div>
                                </Paper>
                            </Box>

                            <Alert severity="info" sx={{ borderRadius: 2, bgcolor: '#e8f0fe', color: '#174ea6', '& .MuiAlert-icon': { color: '#1967d2' } }}>
                                The gateway status will update to <Box component="span" sx={{ fontWeight: 700 }}>Online</Box> once connected.
                            </Alert>
                        </Box>
                    )}
                </DialogContent>
                <DialogActions sx={{ p: 3, px: 4, bgcolor: '#f8f9fa', borderTop: '1px solid #f1f3f4' }}>
                    <Button onClick={handleClose} sx={{ fontWeight: 600, color: '#5f6368' }}>{createdNodeId ? 'Close' : 'Cancel'}</Button>
                    {!createdNodeId && (
                        <Button
                            onClick={handleCreate}
                            variant="contained"
                            disabled={!newNodeName || !selectedSkuId || loading}
                            disableElevation
                            sx={{ borderRadius: 2, fontWeight: 700, px: 4 }}
                        >
                            {loading ? 'Creating...' : 'Create Gateway'}
                        </Button>
                    )}
                </DialogActions>
            </Dialog>

            {/* View Setup Guide Dialog */}
            <Dialog
                open={!!viewNode}
                fullScreen={isMobile}
                onClose={() => setViewNode(null)}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: isMobile ? 0 : 4 }
                }}
            >
                <DialogTitle sx={{ fontWeight: 800, p: 3, borderBottom: '1px solid #f1f3f4', display: 'flex', alignItems: 'center', gap: 2 }}>
                    <TerminalIcon color="primary" />
                    Setup Instructions: {viewNode?.name}
                </DialogTitle>
                <DialogContent sx={{ p: isMobile ? 3 : 4 }}>
                    <Alert severity="info" sx={{ mb: 4, borderRadius: 2 }}>
                        Follow these instructions to connect your gateway to the Tridorian control plane.
                    </Alert>

                    <Box sx={{ mb: 3 }}>
                        <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                            1. Node ID
                        </Typography>
                        <Paper variant="outlined" sx={{ p: 2, bgcolor: '#f8f9fa', borderRadius: 2, fontFamily: 'monospace', display: 'flex', justifyContent: 'space-between', alignItems: 'center', border: '1px solid #dadce0' }}>
                            <span style={{ wordBreak: 'break-all', fontWeight: 600, color: '#202124' }}>{viewNode?.id}</span>
                            <Button
                                size="small"
                                sx={{ textTransform: 'none', fontWeight: 700 }}
                                onClick={() => viewNode && navigator.clipboard.writeText(viewNode.id)}
                            >
                                Copy
                            </Button>
                        </Paper>
                    </Box>

                    <Box sx={{ mb: 3 }}>
                        <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                            2. Run the Gateway Registration Command
                        </Typography>
                        <Paper variant="outlined" sx={{ p: 2, bgcolor: '#202124', color: '#e8eaed', borderRadius: 2, fontFamily: 'Google Sans Mono, monospace', fontSize: '0.85rem' }}>
                            <Box sx={{ mb: 1, userSelect: 'none', color: '#9aa0a6' }}>
                                # Run this temporarily to test connection
                            </Box>
                            <div style={{ wordBreak: 'break-all', lineHeight: 1.5 }}>
                                ./gateway --node-id "{viewNode?.id}"
                            </div>
                        </Paper>
                    </Box>

                    <Box sx={{ mb: 3 }}>
                        <Typography variant="body2" sx={{ mb: 1, fontWeight: 600, color: '#5f6368' }}>
                            3. Install as Background Service
                        </Typography>
                        <Paper variant="outlined" sx={{ p: 2, bgcolor: '#202124', color: '#e8eaed', borderRadius: 2, fontFamily: 'Google Sans Mono, monospace', fontSize: '0.85rem' }}>
                            <Box sx={{ mb: 1, userSelect: 'none', color: '#9aa0a6' }}>
                                # Install and start as a system service
                            </Box>
                            <div style={{ wordBreak: 'break-all', lineHeight: 1.5 }}>
                                sudo ./gateway --install-service --node-id "{viewNode?.id}"
                            </div>
                        </Paper>
                    </Box>
                </DialogContent>
                <DialogActions sx={{ p: 3, bgcolor: '#f8f9fa', borderTop: '1px solid #f1f3f4' }}>
                    <Button onClick={() => setViewNode(null)} sx={{ fontWeight: 600 }}>Close</Button>
                </DialogActions>
            </Dialog>
        </Box >
    );
};

export default NodesView;
