import React, { useState } from 'react';
import {
    Box, Typography, Button, TableContainer, Table, TableHead, TableRow, TableCell,
    TableBody, Paper, Chip, IconButton, Dialog, DialogTitle, DialogContent,
    TextField, Select, MenuItem, DialogActions
} from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { AccessPolicy } from '../../types';

interface PoliciesViewProps {
    policies: AccessPolicy[];
    onCreate: (policy: Partial<AccessPolicy>) => Promise<void>;
    onDelete: (id: string) => Promise<void>;
}

const PoliciesView: React.FC<PoliciesViewProps> = ({ policies, onCreate, onDelete }) => {
    const [showDialog, setShowDialog] = useState(false);
    const [newPolicy, setNewPolicy] = useState<Partial<AccessPolicy>>({ name: '', effect: 'allow', priority: 10 });

    const handleCreate = async () => {
        await onCreate(newPolicy);
        setShowDialog(false);
        setNewPolicy({ name: '', effect: 'allow', priority: 10 });
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 4 }}>
                <Typography variant="h5">Access Policies</Typography>
                <Button variant="contained" startIcon={<AddIcon />} onClick={() => setShowDialog(true)}>Create Policy</Button>
            </Box>
            <TableContainer component={Paper}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Effect</TableCell>
                            <TableCell>Priority</TableCell>
                            <TableCell width={50}></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {policies.map((policy) => (
                            <TableRow key={policy.id}>
                                <TableCell component="th" scope="row">{policy.name}</TableCell>
                                <TableCell>
                                    <Chip
                                        label={policy.effect.toUpperCase()}
                                        color={policy.effect === 'allow' ? 'success' : 'error'}
                                        size="small"
                                    />
                                </TableCell>
                                <TableCell>{policy.priority}</TableCell>
                                <TableCell>
                                    <IconButton onClick={() => onDelete(policy.id)} color="error">
                                        <DeleteIcon />
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                        {policies.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={4} align="center">No policies found.</TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </TableContainer>

            <Dialog open={showDialog} onClose={() => setShowDialog(false)}>
                <DialogTitle>Create Access Policy</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="Policy Name"
                        fullWidth
                        value={newPolicy.name}
                        onChange={(e) => setNewPolicy({ ...newPolicy, name: e.target.value })}
                    />
                    <TextField
                        margin="dense"
                        label="Priority"
                        type="number"
                        fullWidth
                        value={newPolicy.priority}
                        onChange={(e) => setNewPolicy({ ...newPolicy, priority: parseInt(e.target.value) })}
                    />
                    <Select
                        fullWidth
                        value={newPolicy.effect}
                        onChange={(e) => setNewPolicy({ ...newPolicy, effect: e.target.value as any })}
                        sx={{ mt: 2 }}
                    >
                        <MenuItem value="allow">Allow</MenuItem>
                        <MenuItem value="deny">Deny</MenuItem>
                    </Select>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setShowDialog(false)}>Cancel</Button>
                    <Button onClick={handleCreate} variant="contained">Create</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default PoliciesView;
