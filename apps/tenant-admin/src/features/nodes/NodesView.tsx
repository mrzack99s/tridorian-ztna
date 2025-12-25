import React, { useState } from 'react';
import {
    Box, Typography, Button, TableContainer, Table, TableHead, TableRow, TableCell,
    TableBody, Paper, IconButton, Dialog, DialogTitle, DialogContent,
    TextField, DialogActions, Alert
} from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { Node } from '../../types';

interface NodesViewProps {
    nodes: Node[];
    onCreate: (name: string) => Promise<string | null>;
    onDelete: (id: string) => Promise<void>;
}

const NodesView: React.FC<NodesViewProps> = ({ nodes, onCreate, onDelete }) => {
    const [showDialog, setShowDialog] = useState(false);
    const [newNodeName, setNewNodeName] = useState('');
    const [generatedToken, setGeneratedToken] = useState<string | null>(null);

    const handleCreate = async () => {
        const token = await onCreate(newNodeName);
        if (token) {
            setGeneratedToken(token);
        } else {
            setShowDialog(false);
            setNewNodeName('');
        }
    };

    const handleClose = () => {
        setShowDialog(false);
        setNewNodeName('');
        setGeneratedToken(null);
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 4 }}>
                <Typography variant="h5">Nodes & Gateways</Typography>
                <Button variant="contained" startIcon={<AddIcon />} onClick={() => setShowDialog(true)}>Register Node</Button>
            </Box>
            <TableContainer component={Paper}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Status</TableCell>
                            <TableCell>Version</TableCell>
                            <TableCell>Hostname</TableCell>
                            <TableCell width={50}></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {nodes.map((node) => (
                            <TableRow key={node.id}>
                                <TableCell>{node.name}</TableCell>
                                <TableCell>{node.status}</TableCell>
                                <TableCell>{node.version}</TableCell>
                                <TableCell>{node.hostname}</TableCell>
                                <TableCell>
                                    <IconButton onClick={() => onDelete(node.id)} color="error">
                                        <DeleteIcon />
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                        {nodes.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} align="center">No nodes registered.</TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </TableContainer>

            <Dialog open={showDialog} onClose={handleClose}>
                <DialogTitle>Register New Node</DialogTitle>
                <DialogContent>
                    {!generatedToken ? (
                        <TextField
                            autoFocus
                            margin="dense"
                            label="Node Name"
                            fullWidth
                            value={newNodeName}
                            onChange={(e) => setNewNodeName(e.target.value)}
                        />
                    ) : (
                        <Box sx={{ mt: 2 }}>
                            <Alert severity="success" sx={{ mb: 2 }}>
                                Node registered successfully! Use the token below to connect your gateway.
                            </Alert>
                            <TextField
                                fullWidth
                                label="Registration Token"
                                value={generatedToken}
                                InputProps={{ readOnly: true }}
                            />
                            <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                                Copy this token now. It will not be shown again.
                            </Typography>
                        </Box>
                    )}
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose}>Close</Button>
                    {!generatedToken && (
                        <Button onClick={handleCreate} variant="contained">Register</Button>
                    )}
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default NodesView;
