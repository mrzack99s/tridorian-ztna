import React, { useState } from 'react';
import {
    Box, Typography, Button, TableContainer, Table, TableHead, TableRow, TableCell,
    TableBody, Paper, IconButton, Dialog, DialogTitle, DialogContent,
    TextField, DialogActions
} from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon, Edit as EditIcon } from '@mui/icons-material';
import { Admin } from '../../types';

interface AdminsViewProps {
    admins: Admin[];
    onCreate: (admin: any) => Promise<void>;
    onDelete: (id: string) => Promise<void>;
    onUpdate: (id: string, name: string) => Promise<void>;
}

const AdminsView: React.FC<AdminsViewProps> = ({ admins, onCreate, onDelete, onUpdate }) => {
    const [showDialog, setShowDialog] = useState(false);
    const [newAdmin, setNewAdmin] = useState({ name: '', email: '', password: '' });

    const handleCreate = async () => {
        await onCreate(newAdmin);
        setShowDialog(false);
        setNewAdmin({ name: '', email: '', password: '' });
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 4 }}>
                <Typography variant="h5">Console Administrators</Typography>
                <Button variant="contained" startIcon={<AddIcon />} onClick={() => setShowDialog(true)}>Add Administrator</Button>
            </Box>
            <TableContainer component={Paper}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Email</TableCell>
                            <TableCell width={100}></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {admins.map((admin) => (
                            <TableRow key={admin.id}>
                                <TableCell>{admin.name}</TableCell>
                                <TableCell>{admin.email}</TableCell>
                                <TableCell>
                                    <IconButton onClick={() => onUpdate(admin.id, prompt('New name:', admin.name) || admin.name)} color="primary">
                                        <EditIcon />
                                    </IconButton>
                                    <IconButton onClick={() => onDelete(admin.id)} color="error">
                                        <DeleteIcon />
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>

            <Dialog open={showDialog} onClose={() => setShowDialog(false)}>
                <DialogTitle>Add Administrator</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="Full Name"
                        fullWidth
                        value={newAdmin.name}
                        onChange={(e) => setNewAdmin({ ...newAdmin, name: e.target.value })}
                    />
                    <TextField
                        margin="dense"
                        label="Email Address"
                        type="email"
                        fullWidth
                        value={newAdmin.email}
                        onChange={(e) => setNewAdmin({ ...newAdmin, email: e.target.value })}
                    />
                    <TextField
                        margin="dense"
                        label="Password"
                        type="password"
                        fullWidth
                        value={newAdmin.password}
                        onChange={(e) => setNewAdmin({ ...newAdmin, password: e.target.value })}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setShowDialog(false)}>Cancel</Button>
                    <Button onClick={handleCreate} variant="contained">Add</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default AdminsView;
